// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          transport/rung1_floor/audio.go
// @id            cog.transport.rung1_audio
// @layer         Layer 3A: Physical & Virtual Transport Substrate (Rung 1 Floor)
// @archetype     operational | stateful
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Rung 1 Civilizational Floor — AFSK Audio Modem
// @role          1200/2400 Hz PCM Sine Modulator & Goertzel Demodulator (Cog 3)
// @maturity      Rev 5 Baseline | Wave 5 Civilizational Floor
// @chassis       bounded-heap | Pure Go stdlib math AFSK DSP (zero CGO/deps)
// @scale         O(N_bits * 16) real PCM sine synthesis & Goertzel detection
// @seams         in: data, transport | uses: encoding/binary, math, sync/atomic
// @invariants    1. Stratum Baseline: Downward Dependency Law, Rev 5 5-Block.
//                2. Exact DFT bin orthogonality: SampleRate=9600Hz, 16 samples
//                   per bit (600 baud), Mark=1200Hz (k=2), Space=2400Hz (k=4).
//                3. Enforces 512-byte Floppy Sector ceiling prior to modulation.
//                4. Zero third-party DSP libraries or CGO bindings.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Implements a real Audio Frequency-Shift Keying (AFSK) modem inspired by
//   1975 Kansas City Standard / acoustic-coupler / packet-radio tones
//   (1200 Hz Mark vs 2400 Hz Space at 9600 Hz 16-bit PCM). Modulates arbitrary
//   512-byte FaithNet frames into actual audio waveforms and demodulates them
//   via real Goertzel tone-energy detectors—enabling FaithNet communication over
//   dial-up phone lines, analog walkie-talkies, CB radios, or speaker-to-mic!
//
//   ### For Architects, Operators & Non-Programmers:
//   Turns any FaithNet envelope into real audible radio/modem chirps and
//   listens back for those tones mathematically—so even when every cell tower
//   and fiber cable is gone, two laptops with sound cards or walkie-talkies
//   can still ping and share news.
//
//   ### For Systems Programmers & Engine Architects:
//   - Signal Geometry: Preamble `[0xAA, 0x55]` + uint16 length + raw bytes;
//     Goertzel power comparison (`markPower >= spacePower`) recovers bits even
//     in the presence of significant additive analog channel noise.
// ==============================================================================

package rung1_floor

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( AFSK DSP Parameters, Sentinel Errors & AudioTransceiver Chassis )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3A: Physical & Virtual Transport Substrate (Rung 1 Floor)
// Module:        cog.transport.rung1_audio
// Mission:       Define orthogonal AFSK PCM parameters, audio medium switchboard,
//                and Rung 1 AudioTransceiver Cog 3 metadata accessors.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
)

// --- Tier 2: External Library Dependencies (None — 100% Go Stdlib Invariant) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---
import (
	"github.com/CreativeWorkzStudio/FaithNet/data"
	"github.com/CreativeWorkzStudio/FaithNet/transport"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Pre-Conditions, AFSK DSP Geometry & Sentinel Errors
// ------------------------------------------------------------------------------

const (
	// AFSKSampleRate is the PCM audio sample rate in Hz (9600 Hz).
	AFSKSampleRate = 9600

	// AFSKSamplesPerBit is the exact integer sample count per symbol (600 baud).
	AFSKSamplesPerBit = 16

	// AFSKFreqMarkHz is the logic-1 tone (1200 Hz, exact DFT bin k=2).
	AFSKFreqMarkHz = 1200.0

	// AFSKFreqSpaceHz is the logic-0 tone (2400 Hz, exact DFT bin k=4).
	AFSKFreqSpaceHz = 2400.0

	// afskAmplitude is the 16-bit signed PCM sine wave peak amplitude.
	afskAmplitude = 24000.0

	// afskPreambleHi / Lo frame every acoustic burst cleanly.
	afskPreambleHi byte = 0xAA
	afskPreambleLo byte = 0x55

	// afskMinToneEnergy rejects flat dead-air buffers during sliding sync while
	// staying >300x below even a 1% attenuated acoustic signal (~3.6e6 power).
	afskMinToneEnergy = 1e4

	// DefaultAudioQueueDepth bounds buffered PCM bursts per AudioTransceiver.
	DefaultAudioQueueDepth = 16
)

var (
	// ErrAFSKBadPreamble indicates acoustic stream lacked 0xAA55 sync header.
	ErrAFSKBadPreamble = errors.New("faithnet/audio: AFSK sync preamble not detected")

	// ErrAFSKTruncated indicates PCM samples ended before declared byte length.
	ErrAFSKTruncated = errors.New("faithnet/audio: AFSK PCM waveform truncated")
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.2: Operational Logical Anchor
// ------------------------------------------------------------------------------

const (
	// ModuleIDAudioModem is the canonical logical ID for Rung 1 AFSK audio.
	ModuleIDAudioModem = "cog.transport.rung1_audio"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.3: Public Tagged Ontology (AudioMedium & AudioTransceiver)
// ------------------------------------------------------------------------------

type audioBurst struct {
	senderID data.NodeID
	pcm      []int16
}

// AudioMedium simulates a shared acoustic / analog radio frequency channel
// carrying real 16-bit PCM AFSK waveforms between attached FaithNet nodes.
type AudioMedium struct {
	mu    sync.RWMutex
	peers map[data.NodeID]*AudioTransceiver
}

// AudioTransceiver implements `transport.Transceiver` on `data.Rung1Floor`
// modulating every outbound frame to real PCM tones and demodulating inbound
// PCM waveforms via Goertzel tone-energy detectors.
type AudioTransceiver struct {
	medium  *AudioMedium
	selfID  data.NodeID
	inbox   chan audioBurst
	healthy atomic.Bool
	closed  atomic.Bool
	sentCnt atomic.Uint64
	recvCnt atomic.Uint64
}

// ------------------------------------------------------------------------------
// Sub-Assembly 2.4: Pre-Conditional Helper Cogs & O(1) Accessors
// ------------------------------------------------------------------------------

// Name returns the canonical Rung 1 AFSK audio modem identifier.
//
// ### For Architects, Operators & Non-Programmers:
// Names this analog sound/radio channel clearly in status cards and logs.
//
// ### For Systems Programmers & Engine Architects:
// O(1) static getter implementing `transport.Transceiver`.
func (a *AudioTransceiver) Name() string {
	return "rung1:audio-afsk"
}

// Rung returns data.Rung1Floor.
//
// ### For Architects, Operators & Non-Programmers:
// Identifies this audio modem as belonging to Rung 1 (Civilizational Floor).
//
// ### For Systems Programmers & Engine Architects:
// O(1) constant accessor implementing `transport.Transceiver`.
func (a *AudioTransceiver) Rung() data.RungLevel {
	return data.Rung1Floor
}

// Healthy reports whether the audio modem transceiver is open and active.
//
// ### For Architects, Operators & Non-Programmers:
// Checks whether the sound-card / radio channel is currently operational.
//
// ### For Systems Programmers & Engine Architects:
// Lock-free atomic check verifying `!closed && healthy`.
func (a *AudioTransceiver) Healthy() bool {
	return !a.closed.Load() && a.healthy.Load()
}

// SetHealthy toggles live health state for deterministic cascade testing.
//
// ### For Architects, Operators & Non-Programmers:
// Simulates muting or restoring the analog radio / audio channel cleanly.
//
// ### For Systems Programmers & Engine Architects:
// Lock-free atomic store gating `Healthy()` and outbound `Send()`.
func (a *AudioTransceiver) SetHealthy(ok bool) {
	a.healthy.Store(ok)
}

// ------------------------------------------------------------------------------
// Sub-Assembly 2.5: Monitored Support Infrastructure (Goertzel Tone Detector)
// ------------------------------------------------------------------------------
// Support Role:  Single-frequency Goertzel DFT bin power estimator over PCM.
// Pattern Watch: Pure O(N) DSP kernel; zero external dependencies.
// ------------------------------------------------------------------------------

func goertzelPower(samples []int16, targetFreqHz float64, sampleRate float64) float64 {
	omega := (2.0 * math.Pi * targetFreqHz) / sampleRate
	coeff := 2.0 * math.Cos(omega)
	var sPrev, sPrev2 float64
	for _, x := range samples {
		s := float64(x) + coeff*sPrev - sPrev2
		sPrev2 = sPrev
		sPrev = s
	}
	return sPrev2*sPrev2 + sPrev*sPrev - coeff*sPrev*sPrev2
}

// ==============================================================================
// BLOCK 3: BODY BLOCK
// ( Composable Business Logic: AFSK PCM Modulator, Goertzel Demod & Transceiver )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3A: Physical & Virtual Transport Substrate (Rung 1 Floor)
// Module:        cog.transport.rung1_audio
// Mission:       Modulate 512B FaithNet frames into 1200/2400 Hz PCM sine waves
//                and recover exact bytes via Goertzel frequency detection.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: Real AFSK PCM Modulation & Goertzel Demodulation
// ------------------------------------------------------------------------------

// ModulateAFSK encodes raw FaithNet wire bytes (<= 512B) into signed 16-bit
// PCM audio samples using orthogonal 1200 Hz (Mark=1) and 2400 Hz (Space=0).
//
// ### For Architects, Operators & Non-Programmers:
// Converts a digital FaithNet envelope into an audible sequence of musical
// dial-up / walkie-talkie chirps ready for a speaker or radio transmitter.
//
// ### For Systems Programmers & Engine Architects:
// Prepends 4-byte sync header (`0xAA, 0x55, lenHi, lenLo`) and synthesizes
// `16` samples per bit (MSB first) at `9600 Hz`.
func ModulateAFSK(raw []byte) ([]int16, error) {
	if len(raw) == 0 {
		return nil, ErrAFSKTruncated
	}
	if len(raw) > data.FloorMaxFrameSize {
		return nil, transport.ErrWireFrameOversize
	}

	framed := make([]byte, 4+len(raw))
	framed[0] = afskPreambleHi
	framed[1] = afskPreambleLo
	binary.BigEndian.PutUint16(framed[2:4], uint16(len(raw)))
	copy(framed[4:], raw)

	totalSamples := len(framed) * 8 * AFSKSamplesPerBit
	pcm := make([]int16, 0, totalSamples)

	for _, b := range framed {
		for bit := 7; bit >= 0; bit-- {
			freq := AFSKFreqSpaceHz
			if ((b >> uint(bit)) & 1) == 1 {
				freq = AFSKFreqMarkHz
			}
			for s := 0; s < AFSKSamplesPerBit; s++ {
				phase := (2.0 * math.Pi * freq * float64(s)) / float64(AFSKSampleRate)
				pcm = append(pcm, int16(afskAmplitude*math.Sin(phase)))
			}
		}
	}
	return pcm, nil
}

func decodeByteAtOffset(pcm []int16, start, byteIdx int) (byte, float64, float64) {
	var val byte
	var totalEnergy, totalContrast float64
	baseBit := byteIdx * 8
	for bit := 0; bit < 8; bit++ {
		offset := start + (baseBit+bit)*AFSKSamplesPerBit
		window := pcm[offset : offset+AFSKSamplesPerBit]
		markE := goertzelPower(window, AFSKFreqMarkHz, float64(AFSKSampleRate))
		spaceE := goertzelPower(window, AFSKFreqSpaceHz, float64(AFSKSampleRate))
		totalEnergy += markE + spaceE
		if markE >= spaceE {
			val = (val << 1) | 1
			totalContrast += markE - spaceE
		} else {
			val <<= 1
			totalContrast += spaceE - markE
		}
	}
	return val, totalEnergy, totalContrast
}

func demodulateAtOffset(pcm []int16, start int) ([]byte, float64, error) {
	const headerSamples = 4 * 8 * AFSKSamplesPerBit
	if start < 0 || len(pcm)-start < headerSamples {
		return nil, 0, ErrAFSKTruncated
	}

	b0, e0, c0 := decodeByteAtOffset(pcm, start, 0)
	if e0 < afskMinToneEnergy || b0 != afskPreambleHi {
		return nil, 0, ErrAFSKBadPreamble
	}
	b1, e1, c1 := decodeByteAtOffset(pcm, start, 1)
	if e1 < afskMinToneEnergy || b1 != afskPreambleLo {
		return nil, 0, ErrAFSKBadPreamble
	}

	b2, _, _ := decodeByteAtOffset(pcm, start, 2)
	b3, _, _ := decodeByteAtOffset(pcm, start, 3)
	declaredLen := int(binary.BigEndian.Uint16([]byte{b2, b3}))
	if declaredLen == 0 || declaredLen > data.FloorMaxFrameSize {
		return nil, 0, transport.ErrWireFrameOversize
	}
	neededSamples := (4 + declaredLen) * 8 * AFSKSamplesPerBit
	if len(pcm)-start < neededSamples {
		return nil, 0, ErrAFSKTruncated
	}

	out := make([]byte, declaredLen)
	for i := 0; i < declaredLen; i++ {
		out[i], _, _ = decodeByteAtOffset(pcm, start, 4+i)
	}
	return out, c0 + c1, nil
}

// DemodulateAFSK recovers raw FaithNet wire bytes from 16-bit PCM audio samples
// by comparing Goertzel signal energy at 1200 Hz (Mark) vs 2400 Hz (Space),
// including an energy-gated sliding-window preamble correlator for raw
// asynchronous microphone captures with arbitrary leading silence.
//
// ### For Architects, Operators & Non-Programmers:
// Listens to an incoming stream of radio/modem chirps—even inside a microphone
// recording that starts with room silence—locks onto the preamble automatically,
// and translates the tones back into the exact original FaithNet envelope.
//
// ### For Systems Programmers & Engine Architects:
// Uses an O(1) fast path at sample offset 0, falling back to a quarter-symbol
// (4-sample) energy-gated sliding search + single-sample peak-contrast fine
// lock around `0xAA55` for asynchronous physical air-gap recordings.
func DemodulateAFSK(pcm []int16) ([]byte, error) {
	const headerSamples = 4 * 8 * AFSKSamplesPerBit
	if len(pcm) < headerSamples {
		return nil, ErrAFSKTruncated
	}

	// Fast path: aligned sample offset 0.
	out0, _, err0 := demodulateAtOffset(pcm, 0)
	if err0 == nil {
		return out0, nil
	}

	// Sliding-window energy-gated preamble lock for asynchronous mic captures.
	maxStart := len(pcm) - headerSamples
	const coarseStep = AFSKSamplesPerBit / 4 // 4 samples = 0.42ms at 9600Hz
	for s := 0; s <= maxStart; s += coarseStep {
		win := pcm[s : s+AFSKSamplesPerBit]
		energy := goertzelPower(win, AFSKFreqMarkHz, float64(AFSKSampleRate)) +
			goertzelPower(win, AFSKFreqSpaceHz, float64(AFSKSampleRate))
		if energy < afskMinToneEnergy {
			continue
		}
		candOut, candContrast, err := demodulateAtOffset(pcm, s)
		if err != nil {
			continue
		}
		// Fine-tune +/- 3 samples around coarse hit to center Goertzel windows.
		bestOut := candOut
		bestContrast := candContrast
		for delta := -3; delta <= 3; delta++ {
			if delta == 0 {
				continue
			}
			fineOut, fineContrast, fineErr := demodulateAtOffset(pcm, s+delta)
			if fineErr == nil && fineContrast > bestContrast {
				bestOut = fineOut
				bestContrast = fineContrast
			}
		}
		return bestOut, nil
	}

	return nil, err0
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: Transceiver Send & Receive Over Shared Audio Medium
// ------------------------------------------------------------------------------

// Send modulates raw wire bytes into real 1200/2400 Hz PCM tones and broadcasts
// them across the shared AudioMedium to target (or all listening radios).
//
// ### For Architects, Operators & Non-Programmers:
// Keys the radio/audio transmitter and sends a modulated FaithNet chirp burst.
//
// ### For Systems Programmers & Engine Architects:
// Non-blocking channel delivery carrying the synthesized `[]int16` PCM waveform.
func (a *AudioTransceiver) Send(target data.NodeID, raw []byte) error {
	if a.closed.Load() {
		return transport.ErrTransceiverClosed
	}
	if !a.healthy.Load() {
		return transport.ErrTransceiverOffline
	}
	pcm, err := ModulateAFSK(raw)
	if err != nil {
		return err
	}
	burst := audioBurst{senderID: a.selfID, pcm: pcm}

	a.medium.mu.RLock()
	defer a.medium.mu.RUnlock()

	if target.IsBroadcast() {
		for id, peer := range a.medium.peers {
			if id == a.selfID || !peer.Healthy() {
				continue
			}
			select {
			case peer.inbox <- burst:
			default:
			}
		}
		a.sentCnt.Add(1)
		return nil
	}

	peer, ok := a.medium.peers[target]
	if !ok || !peer.Healthy() {
		return transport.ErrPeerUnreachable
	}
	select {
	case peer.inbox <- burst:
		a.sentCnt.Add(1)
		return nil
	default:
		return transport.ErrPeerUnreachable
	}
}

// Receive waits for an incoming PCM audio burst, demodulates it via Goertzel,
// and returns the recovered FaithNet WireEnvelope cleanly.
//
// ### For Architects, Operators & Non-Programmers:
// Listens to the radio/sound channel, decodes arriving chirps back into a
// FaithNet envelope, and stops cleanly when the node shuts down.
//
// ### For Systems Programmers & Engine Architects:
// Honors `ctx.Done()` cleanly and copies demodulated bytes into `WireEnvelope`.
func (a *AudioTransceiver) Receive(ctx context.Context) (transport.WireEnvelope, error) {
	if a.closed.Load() {
		return transport.WireEnvelope{}, transport.ErrTransceiverClosed
	}
	for {
		select {
		case <-ctx.Done():
			return transport.WireEnvelope{}, ctx.Err()
		case burst, ok := <-a.inbox:
			if !ok {
				return transport.WireEnvelope{}, transport.ErrTransceiverClosed
			}
			raw, err := DemodulateAFSK(burst.pcm)
			if err != nil {
				continue
			}
			var env transport.WireEnvelope
			env.SenderID = burst.senderID
			env.TargetID = a.selfID
			env.Length = len(raw)
			copy(env.Bytes[:], raw)
			a.recvCnt.Add(1)
			return env, nil
		}
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Contract Verification Cogs & Pre-Flight Validation Tripwires )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3A: Physical & Virtual Transport Substrate (Rung 1 Floor)
// Module:        cog.transport.rung1_audio
// Mission:       Audit live AudioTransceiver state and run a 4-section AFSK
//                pre-flight including real additive noise immunity checks.
// @triggers
//   - Block 4 Audits : (*AudioTransceiver).AuditState(),
//                      PreflightAudioModem(), AuditAudioModem()
//   - Layer 4 Tests  : tests/functional/transport_functional_test.go,
//                      tests/unit/transport/audio_test.go,
//                      tests/integrated/floor_to_orchestrator_test.go,
//                      tests/system/floor_survival_system_test.go
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Contract, Input & Live-Instance Verification Cogs
// ------------------------------------------------------------------------------

// AuditState inspects a live AudioTransceiver non-destructively to confirm
// medium linkage, valid NodeID, and closed/healthy parity.
//
// ### For Architects, Operators & Non-Programmers:
// Verifies a live audio/radio modem endpoint is properly linked and open.
//
// ### For Systems Programmers & Engine Architects:
// Non-mutating O(1) live probe returning structured `data.AuditResult`.
func (a *AudioTransceiver) AuditState() data.AuditResult {
	const check = "rung1_audio_live_state_contract"

	if a == nil || a.medium == nil || a.inbox == nil {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDAudioModem, check,
			"AudioTransceiver is nil or detached from AudioMedium.")
	}
	if !a.selfID.IsValid() || a.selfID.IsBroadcast() {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDAudioModem, check,
			"AudioTransceiver selfID is null or broadcast.")
	}
	if a.closed.Load() && a.healthy.Load() {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDAudioModem, check,
			"Closed AudioTransceiver erroneously reports Healthy()==true.")
	}
	return data.AuditPass(ModuleIDAudioModem, check,
		"Live Rung 1 AFSK AudioTransceiver medium and identity verified.")
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Machine Self-Audit & Pre-Flight Validation Tripwires
// ------------------------------------------------------------------------------

// PreflightAudioModem executes a 4-section pre-flight check covering
// (1) 1200/2400 Hz PCM modulation/Goertzel demodulation, (2) silent/bad-sync
// fault inversion, (3) 512B ceiling & additive noise immunity, and (4) live
// two-node AudioMedium exchange.
//
// ### For Architects, Operators & Non-Programmers:
// Bench-tests FaithNet's audible modem chirps—even adding simulated radio
// static—to guarantee packets still arrive bit-perfect over analog audio.
//
// ### For Systems Programmers & Engine Architects:
// Pure stdlib 4-section DSP pre-flight gate returning structured `AuditResult`.
func PreflightAudioModem() data.AuditResult {
	const check = "rung1_audio_preflight_4section"

	// Micro Check 1 (Always Pass): Modulate -> Demodulate bit-exact recovery.
	msg := []byte{0xFA, 0x17, 0x11, 0x01, 'A', 'F', 'S', 'K'}
	pcm, err := ModulateAFSK(msg)
	if err != nil {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDAudioModem, check,
			"Sec 1 ModulateAFSK failed on nominal payload.")
	}
	dec, err := DemodulateAFSK(pcm)
	if err != nil || string(dec) != string(msg) {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDAudioModem, check,
			"Sec 1 Goertzel DemodulateAFSK failed bit-exact recovery.")
	}

	// Micro Check 2 (Never Fail): Silence/garbage rejected cleanly.
	if _, err := DemodulateAFSK(make([]int16, 16)); err != ErrAFSKTruncated {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDAudioModem, check,
			"Sec 2 short silence did not return ErrAFSKTruncated.")
	}
	if _, err := DemodulateAFSK(make([]int16, 1024)); err != ErrAFSKBadPreamble {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDAudioModem, check,
			"Sec 2 dead silence did not return ErrAFSKBadPreamble.")
	}

	// Micro Check 3 (Invariants): 513B rejected + additive analog noise resilience.
	var over [data.FloorMaxFrameSize + 1]byte
	if _, err := ModulateAFSK(over[:]); err != transport.ErrWireFrameOversize {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDAudioModem, check,
			"Sec 3 513B payload bypassed 512B Floppy Sector ceiling.")
	}
	noisy := append([]int16(nil), pcm...)
	for i := range noisy {
		// Deterministic +/- 4000 PCM static (~17% peak amplitude noise).
		if i%2 == 0 {
			noisy[i] += 4000
		} else {
			noisy[i] -= 4000
		}
	}
	noisyDec, nErr := DemodulateAFSK(noisy)
	if nErr != nil || string(noisyDec) != string(msg) {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDAudioModem, check,
			"Sec 3 Goertzel detector failed under simulated analog radio static.")
	}

	// Micro Check 4 (Emergence): Two-node AudioMedium exchange + clean Close().
	med := NewAudioMedium()
	txA := med.Attach(0x01)
	txB := med.Attach(0x02)
	defer func() {
		_ = txA.Close()
		_ = txB.Close()
	}()
	if live := txA.AuditState(); !live.IsPass() {
		return live
	}
	if err := txA.Send(0x02, msg); err != nil {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDAudioModem, check,
			"Sec 4 AudioTransceiver Send failed.")
	}
	env, err := txB.Receive(context.Background())
	if err != nil || string(env.Slice()) != string(msg) {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDAudioModem, check,
			"Sec 4 AudioTransceiver Receive mismatch.")
	}
	return data.AuditPass(ModuleIDAudioModem, check,
		"Rung 1 AFSK Audio Modem passed all 4 micro gates (incl. radio static).")
}

// AuditAudioModem returns true iff PreflightAudioModem() passes cleanly.
//
// ### For Architects, Operators & Non-Programmers:
// Quick pass/fail check confirming the analog AFSK audio modem works.
//
// ### For Systems Programmers & Engine Architects:
// Convenience boolean bridge delegating to `PreflightAudioModem().IsPass()`.
func AuditAudioModem() bool {
	return PreflightAudioModem().IsPass()
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK / EXECUTABLE PROCEDURES & LIFECYCLE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3A: Physical & Virtual Transport Substrate (Rung 1 Floor)
// Module:        cog.transport.rung1_audio
// Mission:       Expose epilogue readiness, 60-col AFSK cards, and AudioMedium
//                lifecycle creation, attachment, and Close teardown.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Executable Procedures, Inspection, Readiness & Cards
// ------------------------------------------------------------------------------

// IsAudioModemReady reports whether the Rung 1 AFSK Audio Modem passed preflight.
//
// ### For Architects, Operators & Non-Programmers:
// Epilogue confirmation that analog sound-wave networking is healthy and ready.
//
// ### For Systems Programmers & Engine Architects:
// Matches OmniCode's Block 5 epilogue readiness predicate (`is_*_ready`).
func IsAudioModemReady() bool {
	return PreflightAudioModem().IsPass()
}

// ExecuteAudioDumpCard renders a 60-column inner-width Rung 1 AFSK modem card.
//
// ### For Architects, Operators & Non-Programmers:
// Displays the AFSK tone frequencies, sample rate, health, and packet tally.
//
// ### For Systems Programmers & Engine Architects:
// Non-mutating diagnostic card builder using lock-free atomic telemetry.
func (a *AudioTransceiver) ExecuteAudioDumpCard() string {
	return fmt.Sprintf(
		"+------------------------------------------------------------+\n"+
			"| %-58s |\n"+
			"+------------------------------------------------------------+\n"+
			"| Module:   %-48s |\n"+
			"| NodeID:   %-48s |\n"+
			"| Modem:    %-48s |\n"+
			"| Healthy:  %-48t |\n"+
			"| Traffic:  %-48s |\n"+
			"+------------------------------------------------------------+",
		"FAITHNET RUNG 1 AFSK AUDIO MODEM CARD",
		ModuleIDAudioModem,
		a.selfID.String(),
		"1200Hz Mark / 2400Hz Space @ 9600Hz PCM (600 Bd)",
		a.Healthy(),
		fmt.Sprintf("%d sent / %d recv", a.sentCnt.Load(), a.recvCnt.Load()),
	)
}

// ------------------------------------------------------------------------------
// Sub-Assembly 5.2: Subsystem Lifecycle Initialization & Teardown
// ------------------------------------------------------------------------------

// NewAudioMedium creates an empty shared analog acoustic/radio channel.
//
// ### For Architects, Operators & Non-Programmers:
// Opens a shared acoustic channel where nearby FaithNet radios exchange chirps.
//
// ### For Systems Programmers & Engine Architects:
// Allocates mutex-guarded `NodeID -> *AudioTransceiver` peer map.
func NewAudioMedium() *AudioMedium {
	return &AudioMedium{
		peers: make(map[data.NodeID]*AudioTransceiver),
	}
}

// Attach tunes nodeID onto the AudioMedium and returns its AudioTransceiver.
//
// ### For Architects, Operators & Non-Programmers:
// Plugs a node's sound modem into the shared analog audio channel.
//
// ### For Systems Programmers & Engine Architects:
// Allocates bounded `audioBurst` channel (`DefaultAudioQueueDepth`) under lock.
func (m *AudioMedium) Attach(nodeID data.NodeID) *AudioTransceiver {
	tx := &AudioTransceiver{
		medium: m,
		selfID: nodeID,
		inbox:  make(chan audioBurst, DefaultAudioQueueDepth),
	}
	tx.healthy.Store(true)

	m.mu.Lock()
	m.peers[nodeID] = tx
	m.mu.Unlock()
	return tx
}

// Close detaches this AudioTransceiver from AudioMedium and closes its inbox.
//
// ### For Architects, Operators & Non-Programmers:
// Turns off this node's audio modem cleanly.
//
// ### For Systems Programmers & Engine Architects:
// Idempotent via atomic CompareAndSwap; safe for concurrent Close() calls.
func (a *AudioTransceiver) Close() error {
	if !a.closed.CompareAndSwap(false, true) {
		return nil
	}
	a.healthy.Store(false)
	a.medium.mu.Lock()
	delete(a.medium.peers, a.selfID)
	a.medium.mu.Unlock()
	close(a.inbox)
	return nil
}

// End of transport/rung1_floor/audio.go
