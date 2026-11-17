// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/playtester/harness/audio_torture_bench/main.go
// @id            playtest.harness.audio_torture_bench
// @layer         Layer 4: Proving Ground (Rung 1 Acoustic AFSK & Morse WAV Lab)
// @archetype     playtest_harness | operational
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Rung 1 Acoustic Signal-to-Noise & WAV Torture Bench
// @role          Real 16-bit WAV Export, SNR Breaking Point, Offset & Codec Lab
// @maturity      Rev 5 Baseline | Wave 7A Playtester Proving Ground
// @chassis       bounded-heap | Pure Go RIFF/WAVE writer + optional ffmpeg probe
// @scale         Empirical SNR dB curve, volume invariance & mic-offset discovery
// @seams         in: data, engine, transport/rung1_floor
// @invariants    1. Writes real 16-bit PCM RIFF/WAVE files playable by `aplay`.
//                2. Reports exact empirical SNR breaking point AND honest
//                   microphone start-offset findings without hiding weaknesses.
//                3. Supports `--export-wav` and `--decode-wav` for real air-gap
//                   speaker-to-microphone field playtesting between 2 laptops!
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Pushes FaithNet's Rung 1 Civilizational Floor acoustic stack (`ModulateAFSK`,
//   `DemodulateAFSK`, and 1844 `MorseEncodeBytes`) beyond synthetic unit noise:
//   (1) generates genuine 16-bit PCM `.wav` files playable on real speakers,
//   (2) sweeps additive noise to locate the exact empirical SNR breaking point
//   in dB, (3) tests volume attenuation down to 1%, (4) tests leading-silence
//   microphone capture offset (crucial for real acoustic air-gap testing!), and
//   (5) runs real `ffmpeg` audio resampling/MP3 compression torture if installed.
//
//   ### For Architects, Operators & Non-Programmers:
//   Creates real audio files you can actually listen to or transmit across a
//   room from Laptop #1's speaker to Laptop #2's mic, and measures honestly
//   how much static, low volume, or recording offset breaks reception.
//
//   ### For Systems Programmers & Engine Architects:
//   Implements a zero-dependency standard 44-byte RIFF/WAVE reader/writer and
//   exposes `--export-wav=<path>` and `--decode-wav=<path>` CLI modes reused by
//   `tests/playtester/field_kits/02_acoustic_airgap_kit/`.
// ==============================================================================

package main

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( RIFF/WAVE PCM Helpers & Signal Analysis Math )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Playtester Proving Ground
// Module:        playtest.harness.audio_torture_bench
// Mission:       Provide pure-Go 16-bit PCM WAV read/write and SNR calculation.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// --- Tier 2: External Library Dependencies (None — 100% Go Stdlib Invariant) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---
import (
	"github.com/CreativeWorkzStudio/FaithNet/data"
	"github.com/CreativeWorkzStudio/FaithNet/engine"
	"github.com/CreativeWorkzStudio/FaithNet/transport/rung1_floor"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Operational Constants & RIFF/WAVE Codec
// ------------------------------------------------------------------------------

const (
	ModuleIDAudioTorture = "playtest.harness.audio_torture_bench"
	audioNodeAlpha       = data.NodeID(0xFA1700000000A501)
	audioNodeBeta        = data.NodeID(0xFA1700000000A502)
)

func writeWAV16Mono(path string, sampleRate uint32, pcm []int16) error {
	dataBytes := uint32(len(pcm) * 2)
	var hdr [44]byte
	copy(hdr[0:4], "RIFF")
	binary.LittleEndian.PutUint32(hdr[4:8], 36+dataBytes)
	copy(hdr[8:12], "WAVE")
	copy(hdr[12:16], "fmt ")
	binary.LittleEndian.PutUint32(hdr[16:20], 16) // PCM chunk size
	binary.LittleEndian.PutUint16(hdr[20:22], 1)  // PCM format
	binary.LittleEndian.PutUint16(hdr[22:24], 1)  // Mono
	binary.LittleEndian.PutUint32(hdr[24:28], sampleRate)
	binary.LittleEndian.PutUint32(hdr[28:32], sampleRate*2) // ByteRate
	binary.LittleEndian.PutUint16(hdr[32:34], 2)            // BlockAlign
	binary.LittleEndian.PutUint16(hdr[34:36], 16)           // BitsPerSample
	copy(hdr[36:40], "data")
	binary.LittleEndian.PutUint32(hdr[40:44], dataBytes)

	out := make([]byte, 44+dataBytes)
	copy(out[:44], hdr[:])
	for i, s := range pcm {
		binary.LittleEndian.PutUint16(out[44+i*2:], uint16(s))
	}
	return os.WriteFile(path, out, 0644)
}

func readWAV16Mono(path string) ([]int16, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(raw) < 44 || string(raw[0:4]) != "RIFF" || string(raw[8:12]) != "WAVE" {
		return nil, fmt.Errorf("not a valid RIFF/WAVE file")
	}
	// Locate "data" subchunk cleanly (ffmpeg may add LIST/INFO chunks before data).
	pos := 12
	for pos+8 <= len(raw) {
		chunkID := string(raw[pos : pos+4])
		chunkLen := int(binary.LittleEndian.Uint32(raw[pos+4 : pos+8]))
		pos += 8
		if chunkID == "data" {
			if pos+chunkLen > len(raw) {
				chunkLen = len(raw) - pos
			}
			samples := make([]int16, chunkLen/2)
			for i := range samples {
				samples[i] = int16(binary.LittleEndian.Uint16(raw[pos+i*2:]))
			}
			return samples, nil
		}
		pos += chunkLen
		if chunkLen%2 == 1 {
			pos++
		}
	}
	return nil, fmt.Errorf("WAVE file missing data chunk")
}

func computeSNRdB(clean, noisy []int16) float64 {
	var sigPow, noisePow float64
	for i := range clean {
		s := float64(clean[i])
		n := float64(noisy[i]) - s
		sigPow += s * s
		noisePow += n * n
	}
	if noisePow == 0 {
		return 99.9
	}
	return 10.0 * math.Log10(sigPow/noisePow)
}

// ==============================================================================
// BLOCK 3: BODY BLOCK
// ( Acoustic Torture Suite, WAV Export/Decode Modes & ffmpeg Distortion Probe )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Playtester Proving Ground
// Module:        playtest.harness.audio_torture_bench
// Mission:       Find exact empirical limits of Rung 1 AFSK & Morse waveforms.
// ==============================================================================

func buildTestWire(msg string) ([]byte, data.Frame) {
	var f data.Frame
	f.Version = data.ProtocolVersion
	f.Rung = data.Rung1Floor
	f.Opcode = data.OpData
	f.Flags = data.FlagFloorCompact | data.FlagCovenantVerified
	f.Truth = data.QuadPositive
	f.SourceID = audioNodeAlpha
	f.TargetID = audioNodeBeta
	f.SeqID = 7
	f.SetPayload([]byte(msg))

	var buf [data.FloorMaxFrameSize]byte
	n, err := engine.EncodeInto(buf[:], &f)
	if err != nil {
		panic(err)
	}
	out := make([]byte, n)
	copy(out, buf[:n])
	return out, f
}

func runDecodeWAV(path string) {
	pcm, err := readWAV16Mono(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[DECODE FAIL] readWAV: %v\n", err)
		os.Exit(1)
	}
	raw, err := rung1_floor.DemodulateAFSK(pcm)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[DEMOD FAIL] %v (samples=%d)\n", err, len(pcm))
		os.Exit(1)
	}
	var f data.Frame
	if err := engine.DecodeFrom(raw, &f); err != nil {
		fmt.Fprintf(os.Stderr, "[CODEC FAIL] %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("[AIRGAP DECODE OK] src=%s dst=%s seq=%d truth=%s payload=%q\n",
		f.SourceID, f.TargetID, f.SeqID, f.Truth, string(f.PayloadSlice()))
}

func main() {
	exportWAV := flag.String("export-wav", "", "write modulated AFSK FaithNet frame to WAV path and exit")
	decodeWAV := flag.String("decode-wav", "", "demodulate and decode FaithNet frame from WAV path and exit")
	message := flag.String("msg", "BLOCK-7-LOCAL-NEWS:AIRGAP-OK", "message payload to modulate")
	flag.Parse()

	if *decodeWAV != "" {
		runDecodeWAV(*decodeWAV)
		return
	}

	wire, _ := buildTestWire(*message)
	tMod := time.Now()
	cleanPCM, err := rung1_floor.ModulateAFSK(wire)
	modDur := time.Since(tMod)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ModulateAFSK: %v\n", err)
		os.Exit(1)
	}

	if *exportWAV != "" {
		if err := writeWAV16Mono(*exportWAV, rung1_floor.AFSKSampleRate, cleanPCM); err != nil {
			fmt.Fprintf(os.Stderr, "writeWAV: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("[EXPORTED WAV] %s | rate=%dHz | samples=%d (%.2fs) | wire=%dB\n",
			*exportWAV, rung1_floor.AFSKSampleRate, len(cleanPCM),
			float64(len(cleanPCM))/float64(rung1_floor.AFSKSampleRate), len(wire))
		return
	}

	tDemod := time.Now()
	rec, err := rung1_floor.DemodulateAFSK(cleanPCM)
	demodDur := time.Since(tDemod)
	if err != nil || !bytes.Equal(rec, wire) {
		fmt.Fprintf(os.Stderr, "baseline DemodulateAFSK failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== FAITHNET PLAYTESTER: RUNG 1 ACOUSTIC AFSK & MORSE TORTURE LAB ===")
	fmt.Printf("Wire Frame: %d bytes | SampleRate: %d Hz | PCM Samples: %d (%.2f s air time)\n",
		len(wire), rung1_floor.AFSKSampleRate, len(cleanPCM),
		float64(len(cleanPCM))/float64(rung1_floor.AFSKSampleRate))
	fmt.Printf("CPU Timing: Modulate=%.2fms | Demodulate=%.2fms\n\n",
		float64(modDur.Microseconds())/1000.0, float64(demodDur.Microseconds())/1000.0)

	// 1. Empirical Additive Noise / SNR Breaking-Point Sweep.
	fmt.Println("[1. ADDITIVE WHITE NOISE & SNR BREAKING-POINT SWEEP]")
	rng := rand.New(rand.NewSource(20261117))
	noiseLevels := []int32{0, 2000, 5000, 10000, 15000, 20000, 25000, 30000}
	for _, amp := range noiseLevels {
		noisy := make([]int16, len(cleanPCM))
		for i, s := range cleanPCM {
			n := int32(0)
			if amp > 0 {
				n = rng.Int31n(amp*2+1) - amp
			}
			v := int32(s) + n
			if v > 32767 {
				v = 32767
			} else if v < -32768 {
				v = -32768
			}
			noisy[i] = int16(v)
		}
		snr := computeSNRdB(cleanPCM, noisy)
		got, dErr := rung1_floor.DemodulateAFSK(noisy)
		status := "FAIL"
		if dErr == nil && bytes.Equal(got, wire) {
			status = "PASS"
		}
		fmt.Printf("  noiseAmp=+/-%-5d | SNR=%+6.2f dB -> %s\n", amp, snr, status)
	}

	// 2. Volume Attenuation Sweep (simulating distant speaker / weak mic).
	fmt.Println("\n[2. VOLUME ATTENUATION INVARIANCE SWEEP]")
	for _, gain := range []float64{1.0, 0.5, 0.25, 0.10, 0.05, 0.01} {
		att := make([]int16, len(cleanPCM))
		for i, s := range cleanPCM {
			att[i] = int16(float64(s) * gain)
		}
		got, dErr := rung1_floor.DemodulateAFSK(att)
		status := "FAIL"
		if dErr == nil && bytes.Equal(got, wire) {
			status = "PASS"
		}
		fmt.Printf("  volume=%3.0f%% -> %s\n", gain*100, status)
	}

	// 3. Honest Leading-Silence / Microphone Start-Offset Probe!
	fmt.Println("\n[3. MICROPHONE START-OFFSET (LEADING SILENCE) PROBE]")
	for _, padSamples := range []int{0, 5, 16, 40, 200, 4000} {
		padded := make([]int16, padSamples+len(cleanPCM)+padSamples)
		copy(padded[padSamples:], cleanPCM)
		got, dErr := rung1_floor.DemodulateAFSK(padded)
		status := fmt.Sprintf("FAIL (%v)", dErr)
		if dErr == nil && bytes.Equal(got, wire) {
			status = "PASS"
		}
		fmt.Printf("  leadSilence=%-4d samples (%.2fms) -> %s\n",
			padSamples, float64(padSamples)*1000.0/float64(rung1_floor.AFSKSampleRate), status)
	}

	// 4. Real ffmpeg codec torture (lossy MP3 32kbps + 300-3400Hz phone bandpass!).
	if _, err := exec.LookPath("ffmpeg"); err == nil {
		fmt.Println("\n[4. REAL FFMPEG AUDIO CODEC & PHONE-LINE BANDPASS TORTURE]")
		tmpDir, _ := os.MkdirTemp("", "faithnet-wav-torture-*")
		defer func() { _ = os.RemoveAll(tmpDir) }()
		srcWAV := filepath.Join(tmpDir, "clean.wav")
		_ = writeWAV16Mono(srcWAV, rung1_floor.AFSKSampleRate, cleanPCM)

		// 4a: Lossy MP3 32kbps round-trip!
		mp3Path := filepath.Join(tmpDir, "compressed.mp3")
		mp3Out := filepath.Join(tmpDir, "from_mp3.wav")
		_ = exec.Command("ffmpeg", "-y", "-loglevel", "error", "-i", srcWAV, "-b:a", "32k", mp3Path).Run()
		_ = exec.Command("ffmpeg", "-y", "-loglevel", "error", "-i", mp3Path, "-ar", "9600", "-ac", "1", mp3Out).Run()
		if pcm2, err := readWAV16Mono(mp3Out); err == nil {
			got, dErr := rung1_floor.DemodulateAFSK(pcm2)
			status := fmt.Sprintf("FAIL (%v)", dErr)
			if dErr == nil && bytes.Equal(got, wire) {
				status = "PASS"
			}
			fmt.Printf("  ffmpeg MP3 32kbps round-trip (samples %d -> %d) -> %s\n",
				len(cleanPCM), len(pcm2), status)
		}

		// 4b: Analog telephone bandpass (300Hz-3400Hz) at 9600Hz!
		phoneWAV := filepath.Join(tmpDir, "phone_bandpass.wav")
		_ = exec.Command("ffmpeg", "-y", "-loglevel", "error", "-i", srcWAV,
			"-af", "highpass=f=300,lowpass=f=3400", "-ar", "9600", "-ac", "1", phoneWAV).Run()
		if pcm3, err := readWAV16Mono(phoneWAV); err == nil {
			got, dErr := rung1_floor.DemodulateAFSK(pcm3)
			status := fmt.Sprintf("FAIL (%v)", dErr)
			if dErr == nil && bytes.Equal(got, wire) {
				status = "PASS"
			}
			fmt.Printf("  ffmpeg 300-3400Hz analog telephone bandpass -> %s\n", status)
		}
	}

	// 5. 1844 Morse CW baseline telemetry for the same 60-byte wire frame.
	morseStr, err := rung1_floor.MorseEncodeBytes(wire)
	if err == nil {
		pulses := rung1_floor.MorseToPulses(morseStr)
		backStr := rung1_floor.PulsesToMorse(pulses)
		backWire, decErr := rung1_floor.MorseDecodeBytes(backStr)
		ok := decErr == nil && bytes.Equal(backWire, wire)
		fmt.Printf("\n[5. 1844 ITU MORSE CW FLOOR TELEMETRY]\n")
		fmt.Printf("  wire=%dB -> morseChars=%d -> keyingPulses=%d | roundTripOK=%v\n",
			len(wire), len(morseStr), len(pulses), ok)
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Honest Empirical Discovery — Non-Destructive Read/Write )
// ==============================================================================

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( End of Rung 1 Acoustic Torture Lab )
// ==============================================================================

// End of tests/playtester/harness/audio_torture_bench/main.go
