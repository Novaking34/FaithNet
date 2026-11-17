// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          transport/rung1_floor/morse.go
// @id            cog.transport.rung1_morse
// @layer         Layer 3A: Physical & Virtual Transport Substrate (Rung 1 Floor)
// @archetype     operational | stateful
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Rung 1 Civilizational Floor — International Morse CW
// @role          1844 Telegraph / CW Pulse-Keyed Lossless Frame Transceiver
// @maturity      Rev 5 Baseline | Wave 5 Civilizational Floor
// @chassis       bounded-heap | Lossless hex-over-ITU-Morse & CW pulse timeline
// @scale         O(N_frame) lossless binary <-> Morse <-> CW pulse conversion
// @seams         in: data, transport | uses: encoding/hex, strings, sync/atomic
// @invariants    1. Stratum Baseline: Downward Dependency Law, Rev 5 5-Block.
//                2. Authentic ITU-R M.1677-1 Morse alphabet for hex digits 0-9,
//                   A-F plus canonical timing (dit=1, dah=3, intra=1, inter=3).
//                3. Losslessly carries ANY arbitrary 512-byte FaithNet binary
//                   frame (`0xFA17...`) across raw on/off CW pulses.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Implements the absolute civilizational floor of FaithNet (pre-dial-up,
//   1844 telegraph / CW radio / flashing LED / single-wire on-off keying!).
//   Encodes any 512-byte FaithNet binary frame as hexadecimal digits in genuine
//   International Morse Code (`0-9`, `A-F`), keys it into discrete boolean CW
//   carrier pulses (dit=1 unit ON, dah=3 units ON, char gap=3 units OFF), and
//   decodes those raw pulses losslessly back into exact wire bytes.
//
//   ### For Architects, Operators & Non-Programmers:
//   Even if every internet server, Wi-Fi chip, and modem vanishes—leaving only
//   a single telegraph key, CW radio beacon, or blinking flashlight—FaithNet's
//   same 512-byte sovereign envelopes still cross the street intact!
//
//   ### For Systems Programmers & Engine Architects:
//   - Ownership & Reentrancy: Pure `MorseEncodeBytes`/`MorseDecodeBytes` and
//     `MorseToPulses`/`PulsesToMorse` plus `MorseTransceiver` implementing Cog 3
//     `transport.Transceiver` on `data.Rung1Floor`.
// ==============================================================================

package rung1_floor

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( ITU Morse Hex Alphabet, Timing Units, Sentinel Errors & Morse Chassis )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3A: Physical & Virtual Transport Substrate (Rung 1 Floor)
// Module:        cog.transport.rung1_morse
// Mission:       Define ITU Morse tables, CW pulse timing, and Rung 1
//                MorseTransceiver Cog 3 state & accessors.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
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
// Sub-Assembly 2.1: Pre-Conditions, ITU Morse Tables & Sentinel Errors
// ------------------------------------------------------------------------------

const (
	// DefaultMorseQueueDepth bounds buffered CW pulse streams per receiver.
	DefaultMorseQueueDepth = 16
)

var (
	// ErrMorseInvalidSymbol indicates an unrecognized Morse dit/dah sequence.
	ErrMorseInvalidSymbol = errors.New("faithnet/morse: unrecognized ITU Morse symbol")

	// ErrMorseEmpty indicates an empty byte input or empty Morse CW stream.
	ErrMorseEmpty = errors.New("faithnet/morse: empty Morse payload")
)

// Authentic ITU-R M.1677-1 International Morse Code for hex alphabet [0-9, A-F].
var hexToMorse = map[byte]string{
	'0': "-----",
	'1': ".----",
	'2': "..---",
	'3': "...--",
	'4': "....-",
	'5': ".....",
	'6': "-....",
	'7': "--...",
	'8': "---..",
	'9': "----.",
	'A': ".-",
	'B': "-...",
	'C': "-.-.",
	'D': "-..",
	'E': ".",
	'F': "..-.",
}

var morseToHex = map[string]byte{
	"-----": '0',
	".----": '1',
	"..---": '2',
	"...--": '3',
	"....-": '4',
	".....": '5',
	"-....": '6',
	"--...": '7',
	"---..": '8',
	"----.": '9',
	".-":    'A',
	"-...":  'B',
	"-.-.":  'C',
	"-..":   'D',
	".":     'E',
	"..-.":  'F',
}

// ------------------------------------------------------------------------------
// Sub-Assembly 2.2: Operational Logical Anchor
// ------------------------------------------------------------------------------

const (
	// ModuleIDMorseFloor is the canonical logical ID for Rung 1 Morse CW.
	ModuleIDMorseFloor = "cog.transport.rung1_morse"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.3: Public Tagged Ontology (MorseMedium & MorseTransceiver)
// ------------------------------------------------------------------------------

type morseBurst struct {
	senderID data.NodeID
	pulses   []bool // Discrete CW carrier key-down (true) / key-up (false) units
}

// MorseMedium simulates a shared single-frequency CW telegraph / radio carrier.
type MorseMedium struct {
	mu    sync.RWMutex
	peers map[data.NodeID]*MorseTransceiver
}

// MorseTransceiver implements `transport.Transceiver` on `data.Rung1Floor`,
// keying every FaithNet wire frame into real ITU Morse CW pulses and decoding
// inbound CW pulse streams losslessly back into wire envelopes.
type MorseTransceiver struct {
	medium  *MorseMedium
	selfID  data.NodeID
	inbox   chan morseBurst
	healthy atomic.Bool
	closed  atomic.Bool
	sentCnt atomic.Uint64
	recvCnt atomic.Uint64
}

// ------------------------------------------------------------------------------
// Sub-Assembly 2.4: Pre-Conditional Helper Cogs & O(1) Accessors
// ------------------------------------------------------------------------------

// Name returns the canonical Rung 1 Morse CW transport identifier.
//
// ### For Architects, Operators & Non-Programmers:
// Names this telegraph / Morse CW channel clearly in status cards and logs.
//
// ### For Systems Programmers & Engine Architects:
// O(1) static getter implementing `transport.Transceiver`.
func (m *MorseTransceiver) Name() string {
	return "rung1:morse-cw"
}

// Rung returns data.Rung1Floor.
//
// ### For Architects, Operators & Non-Programmers:
// Identifies Morse CW as belonging to Rung 1 (Civilizational Floor).
//
// ### For Systems Programmers & Engine Architects:
// O(1) constant accessor implementing `transport.Transceiver`.
func (m *MorseTransceiver) Rung() data.RungLevel {
	return data.Rung1Floor
}

// Healthy reports whether the Morse CW transceiver is open and active.
//
// ### For Architects, Operators & Non-Programmers:
// Checks whether the Morse CW keyer/receiver is currently operational.
//
// ### For Systems Programmers & Engine Architects:
// Lock-free atomic check verifying `!closed && healthy`.
func (m *MorseTransceiver) Healthy() bool {
	return !m.closed.Load() && m.healthy.Load()
}

// SetHealthy toggles live health state for deterministic cascade testing.
//
// ### For Architects, Operators & Non-Programmers:
// Simulates disconnecting or restoring the Morse telegraph/CW carrier cleanly.
//
// ### For Systems Programmers & Engine Architects:
// Lock-free atomic store gating `Healthy()` and outbound `Send()`.
func (m *MorseTransceiver) SetHealthy(ok bool) {
	m.healthy.Store(ok)
}

// ==============================================================================
// BLOCK 3: BODY BLOCK
// ( Composable Business Logic: ITU Morse Codec, CW Pulse Keying & Transceiver )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3A: Physical & Virtual Transport Substrate (Rung 1 Floor)
// Module:        cog.transport.rung1_morse
// Mission:       Convert binary FaithNet frames <-> ITU Morse <-> discrete CW
//                on/off carrier pulses losslessly over Rung 1 Floor.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: Pure Binary <-> ITU Morse <-> CW Pulse Keying
// ------------------------------------------------------------------------------

// MorseEncodeBytes encodes arbitrary wire bytes (<= 512B) into space-separated
// ITU International Morse Code symbols representing uppercase hexadecimal nibbles.
//
// ### For Architects, Operators & Non-Programmers:
// Translates a binary FaithNet packet into genuine dots and dashes (`.- -...`)
// that can be tapped over a telegraph wire, CW radio, or flashing light.
//
// ### For Systems Programmers & Engine Architects:
// Uppercase hex encoding guarantees 100% lossless transport of all 256 byte
// values using strictly the 16 canonical ITU Morse symbols `0-9` and `A-F`.
func MorseEncodeBytes(raw []byte) (string, error) {
	if len(raw) == 0 {
		return "", ErrMorseEmpty
	}
	if len(raw) > data.FloorMaxFrameSize {
		return "", transport.ErrWireFrameOversize
	}
	hexStr := strings.ToUpper(hex.EncodeToString(raw))
	var sb strings.Builder
	sb.Grow(len(hexStr) * 5)
	for i := 0; i < len(hexStr); i++ {
		if i > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(hexToMorse[hexStr[i]])
	}
	return sb.String(), nil
}

// MorseDecodeBytes decodes space-separated ITU Morse Code hex symbols back into
// the exact original binary FaithNet wire bytes.
//
// ### For Architects, Operators & Non-Programmers:
// Reads a stream of Morse dots and dashes and reconstructs the exact original
// FaithNet binary envelope bit-for-bit.
//
// ### For Systems Programmers & Engine Architects:
// Rejects unknown Morse tokens, odd hex lengths, and >512B decoded frames.
func MorseDecodeBytes(morse string) ([]byte, error) {
	tokens := strings.Fields(morse)
	if len(tokens) == 0 {
		return nil, ErrMorseEmpty
	}
	if len(tokens)%2 != 0 {
		return nil, ErrMorseInvalidSymbol
	}
	if len(tokens)/2 > data.FloorMaxFrameSize {
		return nil, transport.ErrWireFrameOversize
	}

	hexBuf := make([]byte, len(tokens))
	for i, tok := range tokens {
		ch, ok := morseToHex[tok]
		if !ok {
			return nil, ErrMorseInvalidSymbol
		}
		hexBuf[i] = ch
	}
	return hex.DecodeString(string(hexBuf))
}

// MorseToPulses keys a space-separated Morse string into discrete CW carrier
// time-units (dit=1 ON, dah=3 ON, intra-symbol gap=1 OFF, inter-char gap=3 OFF).
//
// ### For Architects, Operators & Non-Programmers:
// Converts written dots and dashes into the actual ON/OFF electrical or radio
// signal timeline sent down a copper wire or CW transmitter.
//
// ### For Systems Programmers & Engine Architects:
// Adheres to exact ITU 1:3:1:3 unit timing ratio so `PulsesToMorse` recovers
// every symbol unambiguously.
func MorseToPulses(morse string) []bool {
	tokens := strings.Fields(morse)
	pulses := make([]bool, 0, len(morse)*4)
	for ti, tok := range tokens {
		if ti > 0 {
			// Inter-character gap: 3 units OFF.
			pulses = append(pulses, false, false, false)
		}
		for si := 0; si < len(tok); si++ {
			if si > 0 {
				// Intra-character element gap: 1 unit OFF.
				pulses = append(pulses, false)
			}
			switch tok[si] {
			case '.':
				pulses = append(pulses, true)
			case '-':
				pulses = append(pulses, true, true, true)
			}
		}
	}
	return pulses
}

// PulsesToMorse demodulates a discrete CW on/off carrier pulse timeline back
// into space-separated ITU Morse dot/dash symbols.
//
// ### For Architects, Operators & Non-Programmers:
// Measures the length of each ON pulse and OFF pause on a CW line and turns
// them cleanly back into Morse dots and dashes.
//
// ### For Systems Programmers & Engine Architects:
// Run-length decoder classifying ON runs >= 2 units as dah (`-`), 1 unit as
// dit (`.`), and OFF gaps >= 2 units as character boundaries (` `).
func PulsesToMorse(pulses []bool) string {
	var sb strings.Builder
	n := len(pulses)
	i := 0
	for i < n {
		val := pulses[i]
		start := i
		for i < n && pulses[i] == val {
			i++
		}
		runLen := i - start
		if val {
			if runLen >= 2 {
				sb.WriteByte('-')
			} else {
				sb.WriteByte('.')
			}
		} else if sb.Len() > 0 && i < n && runLen >= 2 {
			sb.WriteByte(' ')
		}
	}
	return sb.String()
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: Transceiver Send & Receive Over CW Pulse Carrier
// ------------------------------------------------------------------------------

// Send encodes raw wire bytes to ITU Morse -> discrete CW pulses and keys them
// across MorseMedium to target (or all listening CW stations).
//
// ### For Architects, Operators & Non-Programmers:
// Taps a FaithNet packet out in genuine Morse code pulses over the floor wire.
//
// ### For Systems Programmers & Engine Architects:
// Non-blocking pulse burst delivery enforcing 512B ceiling & health guards.
func (m *MorseTransceiver) Send(target data.NodeID, raw []byte) error {
	if m.closed.Load() {
		return transport.ErrTransceiverClosed
	}
	if !m.healthy.Load() {
		return transport.ErrTransceiverOffline
	}
	morseStr, err := MorseEncodeBytes(raw)
	if err != nil {
		return err
	}
	burst := morseBurst{
		senderID: m.selfID,
		pulses:   MorseToPulses(morseStr),
	}

	m.medium.mu.RLock()
	defer m.medium.mu.RUnlock()

	if target.IsBroadcast() {
		for id, peer := range m.medium.peers {
			if id == m.selfID || !peer.Healthy() {
				continue
			}
			select {
			case peer.inbox <- burst:
			default:
			}
		}
		m.sentCnt.Add(1)
		return nil
	}

	peer, ok := m.medium.peers[target]
	if !ok || !peer.Healthy() {
		return transport.ErrPeerUnreachable
	}
	select {
	case peer.inbox <- burst:
		m.sentCnt.Add(1)
		return nil
	default:
		return transport.ErrPeerUnreachable
	}
}

// Receive waits for an incoming CW pulse burst, demodulates Pulses -> Morse ->
// wire bytes, and returns the exact FaithNet WireEnvelope.
//
// ### For Architects, Operators & Non-Programmers:
// Listens to incoming Morse code pulses and reassembles the exact binary
// FaithNet envelope cleanly.
//
// ### For Systems Programmers & Engine Architects:
// Context-aware reception honoring cancellation with zero goroutine leaks.
func (m *MorseTransceiver) Receive(ctx context.Context) (transport.WireEnvelope, error) {
	if m.closed.Load() {
		return transport.WireEnvelope{}, transport.ErrTransceiverClosed
	}
	for {
		select {
		case <-ctx.Done():
			return transport.WireEnvelope{}, ctx.Err()
		case burst, ok := <-m.inbox:
			if !ok {
				return transport.WireEnvelope{}, transport.ErrTransceiverClosed
			}
			morseStr := PulsesToMorse(burst.pulses)
			raw, err := MorseDecodeBytes(morseStr)
			if err != nil {
				continue
			}
			var env transport.WireEnvelope
			env.SenderID = burst.senderID
			env.TargetID = m.selfID
			env.Length = len(raw)
			copy(env.Bytes[:], raw)
			m.recvCnt.Add(1)
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
// Module:        cog.transport.rung1_morse
// Mission:       Audit live MorseTransceiver state and run a 4-section CW
//                pre-flight verifying bytes -> ITU Morse -> CW pulses -> bytes.
// @triggers
//   - Block 4 Audits : (*MorseTransceiver).AuditState(),
//                      PreflightMorseFloor(), AuditMorseFloor()
//   - Layer 4 Tests  : tests/unit/transport/floor_transports_test.go,
//                      tests/system/sovereign_mesh_e2e_test.go
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Contract, Input & Live-Instance Verification Cogs
// ------------------------------------------------------------------------------

// AuditState inspects a live MorseTransceiver non-destructively to confirm
// medium linkage, valid NodeID, and closed/healthy consistency.
//
// ### For Architects, Operators & Non-Programmers:
// Checks a live Morse CW station endpoint to ensure it is properly connected.
//
// ### For Systems Programmers & Engine Architects:
// Non-mutating O(1) live probe returning structured `data.AuditResult`.
func (m *MorseTransceiver) AuditState() data.AuditResult {
	const check = "rung1_morse_live_state_contract"

	if m == nil || m.medium == nil || m.inbox == nil {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDMorseFloor, check,
			"MorseTransceiver is nil or detached from MorseMedium.")
	}
	if !m.selfID.IsValid() || m.selfID.IsBroadcast() {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDMorseFloor, check,
			"MorseTransceiver selfID is null or broadcast.")
	}
	if m.closed.Load() && m.healthy.Load() {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDMorseFloor, check,
			"Closed MorseTransceiver erroneously reports Healthy()==true.")
	}
	return data.AuditPass(ModuleIDMorseFloor, check,
		"Live Rung 1 Morse CW Transceiver medium and identity verified.")
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Machine Self-Audit & Pre-Flight Validation Tripwires
// ------------------------------------------------------------------------------

// PreflightMorseFloor executes a 4-section pre-flight audit covering
// (1) binary -> ITU Morse -> CW pulses -> binary parity, (2) invalid symbol
// & empty fault inversion, (3) 512B ceiling enforcement, and (4) live CW ping.
//
// ### For Architects, Operators & Non-Programmers:
// Bench-tests FaithNet's 1844-grade Morse code floor to guarantee even full
// binary packets survive pure on/off pulse keying without a single bit lost.
//
// ### For Systems Programmers & Engine Architects:
// Isolated 4-section pre-flight gate returning structured `data.AuditResult`.
func PreflightMorseFloor() data.AuditResult {
	const check = "rung1_morse_preflight_4section"

	// Micro Check 1 (Always Pass): Full hex nibble range 0x00..0xFF round-trip.
	sample := []byte{0xFA, 0x17, 0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF}
	mStr, err := MorseEncodeBytes(sample)
	if err != nil {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDMorseFloor, check,
			"Sec 1 MorseEncodeBytes failed on full hex sample.")
	}
	pulses := MorseToPulses(mStr)
	mBack := PulsesToMorse(pulses)
	dec, err := MorseDecodeBytes(mBack)
	if err != nil || string(dec) != string(sample) {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDMorseFloor, check,
			"Sec 1 CW pulse -> Morse -> binary recovery mismatch.")
	}

	// Micro Check 2 (Never Fail): Bad symbol, odd nibble count, empty rejected.
	if _, err := MorseDecodeBytes("...... ......"); err != ErrMorseInvalidSymbol {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDMorseFloor, check,
			"Sec 2 invalid Morse symbol was not rejected.")
	}
	if _, err := MorseDecodeBytes(".-"); err != ErrMorseInvalidSymbol {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDMorseFloor, check,
			"Sec 2 odd nibble Morse stream was not rejected.")
	}
	if _, err := MorseEncodeBytes(nil); err != ErrMorseEmpty {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDMorseFloor, check,
			"Sec 2 empty MorseEncodeBytes was not rejected.")
	}

	// Micro Check 3 (Invariants): 513B rejected at 512B Floppy Sector ceiling.
	var over [data.FloorMaxFrameSize + 1]byte
	if _, err := MorseEncodeBytes(over[:]); err != transport.ErrWireFrameOversize {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDMorseFloor, check,
			"Sec 3 513B payload bypassed 512B Floppy Sector ceiling.")
	}

	// Micro Check 4 (Emergence): Two-node MorseMedium CW pulse exchange.
	med := NewMorseMedium()
	txA := med.Attach(0x01)
	txB := med.Attach(0x02)
	defer func() {
		_ = txA.Close()
		_ = txB.Close()
	}()
	if live := txA.AuditState(); !live.IsPass() {
		return live
	}
	if err := txA.Send(0x02, sample); err != nil {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDMorseFloor, check,
			"Sec 4 MorseTransceiver Send failed.")
	}
	env, err := txB.Receive(context.Background())
	if err != nil || string(env.Slice()) != string(sample) {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDMorseFloor, check,
			"Sec 4 MorseTransceiver Receive mismatch.")
	}
	return data.AuditPass(ModuleIDMorseFloor, check,
		"Rung 1 International Morse CW Floor passed all 4 micro pre-flight gates.")
}

// AuditMorseFloor returns true iff PreflightMorseFloor() passes cleanly.
//
// ### For Architects, Operators & Non-Programmers:
// Quick pass/fail check confirming the Morse CW survival floor works.
//
// ### For Systems Programmers & Engine Architects:
// Convenience boolean bridge delegating to `PreflightMorseFloor().IsPass()`.
func AuditMorseFloor() bool {
	return PreflightMorseFloor().IsPass()
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK / EXECUTABLE PROCEDURES & LIFECYCLE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3A: Physical & Virtual Transport Substrate (Rung 1 Floor)
// Module:        cog.transport.rung1_morse
// Mission:       Expose epilogue readiness, 60-col CW cards, and MorseMedium
//                lifecycle creation, attachment, and Close teardown.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Executable Procedures, Inspection, Readiness & Cards
// ------------------------------------------------------------------------------

// IsMorseFloorReady reports whether the Rung 1 Morse CW floor passed preflight.
//
// ### For Architects, Operators & Non-Programmers:
// Epilogue confirmation that Morse telegraph/CW networking is healthy and ready.
//
// ### For Systems Programmers & Engine Architects:
// Matches OmniCode's Block 5 epilogue readiness predicate (`is_*_ready`).
func IsMorseFloorReady() bool {
	return PreflightMorseFloor().IsPass()
}

// ExecuteMorseDumpCard renders a 60-column inner-width Rung 1 Morse CW card.
//
// ### For Architects, Operators & Non-Programmers:
// Displays the Morse CW station ID, ITU pulse timing, health, and traffic count.
//
// ### For Systems Programmers & Engine Architects:
// Non-mutating diagnostic card builder using lock-free atomic telemetry.
func (m *MorseTransceiver) ExecuteMorseDumpCard() string {
	return fmt.Sprintf(
		"+------------------------------------------------------------+\n"+
			"| %-58s |\n"+
			"+------------------------------------------------------------+\n"+
			"| Module:   %-48s |\n"+
			"| NodeID:   %-48s |\n"+
			"| Standard: %-48s |\n"+
			"| Healthy:  %-48t |\n"+
			"| Traffic:  %-48s |\n"+
			"+------------------------------------------------------------+",
		"FAITHNET RUNG 1 INTERNATIONAL MORSE CW CARD",
		ModuleIDMorseFloor,
		m.selfID.String(),
		"ITU-R M.1677-1 Hex CW (dit=1, dah=3, gap=3)",
		m.Healthy(),
		fmt.Sprintf("%d sent / %d recv", m.sentCnt.Load(), m.recvCnt.Load()),
	)
}

// ------------------------------------------------------------------------------
// Sub-Assembly 5.2: Subsystem Lifecycle Initialization & Teardown
// ------------------------------------------------------------------------------

// NewMorseMedium creates an empty shared single-carrier CW Morse channel.
//
// ### For Architects, Operators & Non-Programmers:
// Opens a shared telegraph/CW frequency for nearby FaithNet stations.
//
// ### For Systems Programmers & Engine Architects:
// Allocates mutex-guarded `NodeID -> *MorseTransceiver` peer map.
func NewMorseMedium() *MorseMedium {
	return &MorseMedium{
		peers: make(map[data.NodeID]*MorseTransceiver),
	}
}

// Attach registers nodeID onto MorseMedium and returns its MorseTransceiver.
//
// ### For Architects, Operators & Non-Programmers:
// Connects a node's Morse keyer/receiver to the shared CW channel.
//
// ### For Systems Programmers & Engine Architects:
// Allocates bounded `morseBurst` channel (`DefaultMorseQueueDepth`) under lock.
func (med *MorseMedium) Attach(nodeID data.NodeID) *MorseTransceiver {
	tx := &MorseTransceiver{
		medium: med,
		selfID: nodeID,
		inbox:  make(chan morseBurst, DefaultMorseQueueDepth),
	}
	tx.healthy.Store(true)

	med.mu.Lock()
	med.peers[nodeID] = tx
	med.mu.Unlock()
	return tx
}

// Close detaches this MorseTransceiver from MorseMedium and closes its inbox.
//
// ### For Architects, Operators & Non-Programmers:
// Disconnects this Morse CW station cleanly.
//
// ### For Systems Programmers & Engine Architects:
// Idempotent via atomic CompareAndSwap; safe for concurrent Close() calls.
func (m *MorseTransceiver) Close() error {
	if !m.closed.CompareAndSwap(false, true) {
		return nil
	}
	m.healthy.Store(false)
	m.medium.mu.Lock()
	delete(m.medium.peers, m.selfID)
	m.medium.mu.Unlock()
	close(m.inbox)
	return nil
}

// End of transport/rung1_floor/morse.go
