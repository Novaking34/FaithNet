// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          transport/rung1_floor/serial.go
// @id            cog.transport.rung1_serial
// @layer         Layer 3A: Physical & Virtual Transport Substrate (Rung 1 Floor)
// @archetype     operational | stateful
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Rung 1 Civilizational Floor — RFC 1055 SLIP Serial
// @role          Copper-Wire / RS-232 / UART Byte-Stuffed Transceiver (Cog 3)
// @maturity      Rev 5 Baseline | Wave 5 Civilizational Floor
// @chassis       bounded-heap | Fixed SLIP escaping over any io.Reader/io.Writer
// @scale         O(N_frame) single-pass SLIP byte stuffing (<= 1026B max wire)
// @seams         in: data, transport | uses: io, sync, sync/atomic
// @invariants    1. Stratum Baseline: Downward Dependency Law, Rev 5 5-Block.
//                2. Lossless RFC 1055 byte stuffing (`END=0xC0`, `ESC=0xDB`,
//                   `ESC_END=0xDC`, `ESC_ESC=0xDD`) for arbitrary binary bytes.
//                3. Enforces 512-byte Floppy Sector ceiling before and after
//                   SLIP decoding (rejects oversize serial streams cleanly).
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Implements FaithNet's Rung 1 copper-wire serial transceiver using RFC 1055
//   SLIP framing over any `io.Reader`/`io.Writer` pair (RS-232 COM/ttyS ports,
//   USB-to-TTL serial cables, null-modem copper pairs, or `io.Pipe()`).
//
//   ### For Architects, Operators & Non-Programmers:
//   Allows two machines connected by nothing more than a 3-wire serial cable
//   (TX, RX, GND) to exchange verified 512-byte FaithNet frames completely
//   offline with zero Ethernet, Wi-Fi, or commercial network gear.
//
//   ### For Systems Programmers & Engine Architects:
//   - Ownership & Reentrancy: Pure `SLIPEncode`/`SLIPDecode` functions plus
//     `SerialTransceiver` implementing `transport.Transceiver` on `Rung1Floor`.
//   - Extension Seam: Accepts any OS serial file descriptor (`*os.File`) or
//     in-memory `io.Pipe()` without platform-specific syscalls in core framing.
// ==============================================================================

package rung1_floor

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( RFC 1055 SLIP Constants, Sentinel Errors & SerialTransceiver Chassis )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3A: Physical & Virtual Transport Substrate (Rung 1 Floor)
// Module:        cog.transport.rung1_serial
// Mission:       Define RFC 1055 SLIP delimiters, sentinel errors, and the
//                Rung 1 SerialTransceiver state chassis.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"

	"github.com/CreativeWorkzStudio/FaithNet/data"
	"github.com/CreativeWorkzStudio/FaithNet/transport"
)

// --- Tier 2: External Library Dependencies (None — 100% Go Stdlib Invariant) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Pre-Conditions, RFC 1055 Constants & Sentinel Errors
// ------------------------------------------------------------------------------

const (
	// SLIPEND terminates and delimits every RFC 1055 SLIP frame (0xC0).
	SLIPEND byte = 0xC0

	// SLIPESC escapes literal END or ESC bytes inside payloads (0xDB).
	SLIPESC byte = 0xDB

	// SLIPESCEND follows SLIPESC to represent a literal 0xC0 payload byte.
	SLIPESCEND byte = 0xDC

	// SLIPESCESC follows SLIPESC to represent a literal 0xDB payload byte.
	SLIPESCESC byte = 0xDD
)

var (
	// ErrSLIPBadEscape indicates a corrupt RFC 1055 escape sequence on wire.
	ErrSLIPBadEscape = errors.New("faithnet/serial: malformed SLIP escape sequence")

	// ErrSLIPEmptyFrame indicates an empty or unterminated SLIP frame.
	ErrSLIPEmptyFrame = errors.New("faithnet/serial: empty SLIP frame")
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.2: Operational Logical Anchor
// ------------------------------------------------------------------------------

const (
	// ModuleIDSerialFloor is the canonical logical ID for Rung 1 SLIP serial.
	ModuleIDSerialFloor = "cog.transport.rung1_serial"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.3: Public Tagged Ontology (SerialTransceiver Chassis)
// ------------------------------------------------------------------------------

// SerialTransceiver implements `transport.Transceiver` on `data.Rung1Floor`
// using RFC 1055 SLIP framing over any byte stream (`io.Reader` + `io.Writer`).
//
// ### For Architects, Operators & Non-Programmers:
// Turns any serial cable or byte pipe into a full FaithNet survival-floor wire.
//
// ### For Systems Programmers & Engine Architects:
// Serializes outbound writes under `writeMu` and reads framed packets cleanly
// into bounded 512-byte `transport.WireEnvelope` structs.
type SerialTransceiver struct {
	selfID  data.NodeID
	reader  io.Reader
	writer  io.Writer
	closer  io.Closer
	writeMu sync.Mutex
	readMu  sync.Mutex
	healthy atomic.Bool
	closed  atomic.Bool
	sentCnt atomic.Uint64
	recvCnt atomic.Uint64
}

// ------------------------------------------------------------------------------
// Sub-Assembly 2.4: Pre-Conditional Helper Cogs & O(1) Accessors
// ------------------------------------------------------------------------------

// Name returns the canonical Rung 1 SLIP serial transport identifier.
//
// ### For Architects, Operators & Non-Programmers:
// Names this copper-wire serial channel clearly in status cards and logs.
//
// ### For Systems Programmers & Engine Architects:
// O(1) static getter implementing `transport.Transceiver`.
func (s *SerialTransceiver) Name() string {
	return "rung1:serial-slip"
}

// Rung returns data.Rung1Floor.
//
// ### For Architects, Operators & Non-Programmers:
// Identifies this serial cable as belonging to Rung 1 (Civilizational Floor).
//
// ### For Systems Programmers & Engine Architects:
// O(1) constant accessor implementing `transport.Transceiver`.
func (s *SerialTransceiver) Rung() data.RungLevel {
	return data.Rung1Floor
}

// Healthy reports whether the serial transceiver is open and operational.
//
// ### For Architects, Operators & Non-Programmers:
// Checks whether the serial cable link is currently active and ready.
//
// ### For Systems Programmers & Engine Architects:
// Lock-free atomic check verifying `!closed && healthy`.
func (s *SerialTransceiver) Healthy() bool {
	return !s.closed.Load() && s.healthy.Load()
}

// SetHealthy toggles live health state for deterministic cascade testing.
//
// ### For Architects, Operators & Non-Programmers:
// Simulates unplugging or reconnecting the serial copper line cleanly.
//
// ### For Systems Programmers & Engine Architects:
// Lock-free atomic store gating `Healthy()` and outbound `Send()`.
func (s *SerialTransceiver) SetHealthy(ok bool) {
	s.healthy.Store(ok)
}

// ==============================================================================
// BLOCK 3: BODY BLOCK
// ( Composable Business Logic: RFC 1055 SLIP Codec & Serial Stream Send/Recv )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3A: Physical & Virtual Transport Substrate (Rung 1 Floor)
// Module:        cog.transport.rung1_serial
// Mission:       Encode/decode arbitrary 512-byte FaithNet frames cleanly over
//                raw asynchronous byte streams via RFC 1055 SLIP framing.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: Pure RFC 1055 SLIP Frame Stuffing & Unstuffing
// ------------------------------------------------------------------------------

// SLIPEncode wraps raw wire bytes (<= 512B) in RFC 1055 END delimiters with
// deterministic byte stuffing for literal 0xC0 and 0xDB bytes.
//
// ### For Architects, Operators & Non-Programmers:
// Adds start/end markers around a FaithNet envelope so a simple serial wire
// knows exactly where each packet begins and ends—even if the packet contains
// every possible binary byte value.
//
// ### For Systems Programmers & Engine Architects:
// Emits leading + trailing `SLIPEND` bytes (`double-END` variant) to flush
// line noise prior to frame payload start.
func SLIPEncode(raw []byte) ([]byte, error) {
	if len(raw) == 0 {
		return nil, ErrSLIPEmptyFrame
	}
	if len(raw) > data.FloorMaxFrameSize {
		return nil, transport.ErrWireFrameOversize
	}
	out := make([]byte, 0, len(raw)+4)
	out = append(out, SLIPEND)
	for _, b := range raw {
		switch b {
		case SLIPEND:
			out = append(out, SLIPESC, SLIPESCEND)
		case SLIPESC:
			out = append(out, SLIPESC, SLIPESCESC)
		default:
			out = append(out, b)
		}
	}
	out = append(out, SLIPEND)
	return out, nil
}

// SLIPDecode unstuffs one RFC 1055 SLIP byte slice back into raw FaithNet bytes.
//
// ### For Architects, Operators & Non-Programmers:
// Strips serial start/end markers and restores escaped bytes back into the
// exact original 512-byte FaithNet envelope.
//
// ### For Systems Programmers & Engine Architects:
// Rejects unknown escape codes (`ErrSLIPBadEscape`), empty payloads, and
// decoded streams exceeding `data.FloorMaxFrameSize` (`ErrWireFrameOversize`).
func SLIPDecode(slip []byte) ([]byte, error) {
	out := make([]byte, 0, len(slip))
	escaped := false

	for _, b := range slip {
		if escaped {
			switch b {
			case SLIPESCEND:
				out = append(out, SLIPEND)
			case SLIPESCESC:
				out = append(out, SLIPESC)
			default:
				return nil, ErrSLIPBadEscape
			}
			escaped = false
			if len(out) > data.FloorMaxFrameSize {
				return nil, transport.ErrWireFrameOversize
			}
			continue
		}
		switch b {
		case SLIPEND:
			if len(out) > 0 {
				return out, nil
			}
		case SLIPESC:
			escaped = true
		default:
			out = append(out, b)
			if len(out) > data.FloorMaxFrameSize {
				return nil, transport.ErrWireFrameOversize
			}
		}
	}
	if escaped {
		return nil, ErrSLIPBadEscape
	}
	if len(out) == 0 {
		return nil, ErrSLIPEmptyFrame
	}
	return out, nil
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: Transceiver Send & Receive Over Serial Stream
// ------------------------------------------------------------------------------

// Send SLIP-encodes raw wire bytes and writes them onto the serial stream.
//
// ### For Architects, Operators & Non-Programmers:
// Pushes one sealed FaithNet envelope out over the serial copper wire.
//
// ### For Systems Programmers & Engine Architects:
// Mutex-guarded write ensuring concurrent senders never interleave SLIP bytes.
func (s *SerialTransceiver) Send(_ data.NodeID, raw []byte) error {
	if s.closed.Load() {
		return transport.ErrTransceiverClosed
	}
	if !s.healthy.Load() {
		return transport.ErrTransceiverOffline
	}
	framed, err := SLIPEncode(raw)
	if err != nil {
		return err
	}
	s.writeMu.Lock()
	_, wErr := s.writer.Write(framed)
	s.writeMu.Unlock()
	if wErr != nil {
		return transport.ErrPeerUnreachable
	}
	s.sentCnt.Add(1)
	return nil
}

// Receive reads bytes from the serial stream until a complete SLIP frame arrives
// or ctx cancels.
//
// ### For Architects, Operators & Non-Programmers:
// Listens byte-by-byte on the serial line and hands back each complete FaithNet
// envelope the instant its ending marker arrives.
//
// ### For Systems Programmers & Engine Architects:
// Reads into a fixed `[512]byte` envelope buffer enforcing `FloorMaxFrameSize`.
func (s *SerialTransceiver) Receive(ctx context.Context) (transport.WireEnvelope, error) {
	if s.closed.Load() {
		return transport.WireEnvelope{}, transport.ErrTransceiverClosed
	}
	if ctx.Err() != nil {
		return transport.WireEnvelope{}, ctx.Err()
	}

	s.readMu.Lock()
	defer s.readMu.Unlock()

	var env transport.WireEnvelope
	var one [1]byte
	escaped := false

	for {
		if ctx.Err() != nil {
			return transport.WireEnvelope{}, ctx.Err()
		}
		if s.closed.Load() {
			return transport.WireEnvelope{}, transport.ErrTransceiverClosed
		}
		n, err := s.reader.Read(one[:])
		if err != nil {
			if s.closed.Load() {
				return transport.WireEnvelope{}, transport.ErrTransceiverClosed
			}
			return transport.WireEnvelope{}, err
		}
		if n == 0 {
			continue
		}
		b := one[0]
		if escaped {
			var decoded byte
			switch b {
			case SLIPESCEND:
				decoded = SLIPEND
			case SLIPESCESC:
				decoded = SLIPESC
			default:
				env.Length = 0
				escaped = false
				continue
			}
			escaped = false
			if env.Length >= data.FloorMaxFrameSize {
				env.Length = 0
				continue
			}
			env.Bytes[env.Length] = decoded
			env.Length++
			continue
		}
		switch b {
		case SLIPEND:
			if env.Length > 0 {
				env.TargetID = s.selfID
				s.recvCnt.Add(1)
				return env, nil
			}
		case SLIPESC:
			escaped = true
		default:
			if env.Length >= data.FloorMaxFrameSize {
				env.Length = 0
				continue
			}
			env.Bytes[env.Length] = b
			env.Length++
		}
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Contract Verification Cogs & Pre-Flight Validation Tripwires )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3A: Physical & Virtual Transport Substrate (Rung 1 Floor)
// Module:        cog.transport.rung1_serial
// Mission:       Audit live SerialTransceiver state and run a 4-section SLIP
//                pre-flight over in-memory io.Pipe() verifying escape parity.
// @triggers
//   - Block 4 Audits : (*SerialTransceiver).AuditState(),
//                      PreflightSerialFloor(), AuditSerialFloor()
//   - Layer 4 Tests  : tests/functional/transport_functional_test.go,
//                      tests/unit/transport/serial_test.go,
//                      tests/integrated/floor_to_orchestrator_test.go,
//                      tests/system/floor_survival_system_test.go
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Contract, Input & Live-Instance Verification Cogs
// ------------------------------------------------------------------------------

// AuditState inspects a live SerialTransceiver non-destructively to confirm
// reader/writer handles, valid NodeID, and closed/healthy consistency.
//
// ### For Architects, Operators & Non-Programmers:
// Checks a live serial wire connection to confirm both TX and RX paths exist.
//
// ### For Systems Programmers & Engine Architects:
// Non-mutating O(1) live probe returning structured `data.AuditResult`.
func (s *SerialTransceiver) AuditState() data.AuditResult {
	const check = "rung1_serial_live_state_contract"

	if s == nil {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDSerialFloor, check,
			"SerialTransceiver receiver is nil.")
	}
	if s.reader == nil || s.writer == nil {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDSerialFloor, check,
			"SerialTransceiver missing io.Reader or io.Writer handle.")
	}
	if !s.selfID.IsValid() || s.selfID.IsBroadcast() {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDSerialFloor, check,
			"SerialTransceiver selfID is null or broadcast.")
	}
	if s.closed.Load() && s.healthy.Load() {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDSerialFloor, check,
			"Closed SerialTransceiver erroneously reports Healthy()==true.")
	}
	return data.AuditPass(ModuleIDSerialFloor, check,
		"Live Rung 1 SerialTransceiver stream handles and identity verified.")
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Machine Self-Audit & Pre-Flight Validation Tripwires
// ------------------------------------------------------------------------------

// PreflightSerialFloor executes a 4-section non-destructive pre-flight audit
// covering (1) literal END/ESC byte stuffing round-trip, (2) bad-escape & empty
// fault inversion, (3) 512B ceiling enforcement, and (4) full-duplex io.Pipe().
//
// ### For Architects, Operators & Non-Programmers:
// Bench-tests FaithNet's copper-wire serial framing in memory to guarantee even
// tricky control bytes pass intact without corrupting packet boundaries.
//
// ### For Systems Programmers & Engine Architects:
// Isolated 4-section pre-flight gate returning structured `data.AuditResult`.
func PreflightSerialFloor() data.AuditResult {
	const check = "rung1_serial_preflight_4section"

	// Micro Check 1 (Always Pass): Payload containing literal SLIPEND & SLIPESC.
	tricky := []byte{0xFA, 0x17, SLIPEND, SLIPESC, SLIPESCEND, SLIPESCESC, 0x00, 0xFF}
	enc, err := SLIPEncode(tricky)
	if err != nil {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDSerialFloor, check,
			"Sec 1 SLIPEncode failed on literal END/ESC bytes.")
	}
	dec, err := SLIPDecode(enc)
	if err != nil || len(dec) != len(tricky) {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDSerialFloor, check,
			"Sec 1 SLIPDecode failed round-trip parity.")
	}
	for i := range tricky {
		if dec[i] != tricky[i] {
			return data.AuditFail(data.AuditSeverityFatal, ModuleIDSerialFloor, check,
				"Sec 1 SLIP round-trip byte mismatch.")
		}
	}

	// Micro Check 2 (Never Fail): Malformed escape & empty frames rejected.
	if _, err := SLIPDecode([]byte{SLIPEND, SLIPESC, 0x99, SLIPEND}); err != ErrSLIPBadEscape {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDSerialFloor, check,
			"Sec 2 invalid SLIP escape was not rejected.")
	}
	if _, err := SLIPEncode(nil); err != ErrSLIPEmptyFrame {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDSerialFloor, check,
			"Sec 2 empty SLIPEncode was not rejected.")
	}

	// Micro Check 3 (Invariants & Limits): 513B payload rejected at 512B ceiling.
	var over [data.FloorMaxFrameSize + 1]byte
	if _, err := SLIPEncode(over[:]); err != transport.ErrWireFrameOversize {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDSerialFloor, check,
			"Sec 3 513B payload bypassed 512B Floppy Sector ceiling.")
	}

	// Micro Check 4 (Emergence): Full-duplex io.Pipe() SerialTransceiver test.
	rB, wA := io.Pipe()
	rA, wB := io.Pipe()
	txA := NewSerialTransceiver(0x01, rA, wA, wA)
	txB := NewSerialTransceiver(0x02, rB, wB, wB)
	defer func() {
		_ = txA.Close()
		_ = txB.Close()
		_ = rA.Close()
		_ = rB.Close()
	}()
	if live := txA.AuditState(); !live.IsPass() {
		return live
	}
	done := make(chan error, 1)
	go func() {
		done <- txA.Send(0x02, tricky)
	}()
	env, rErr := txB.Receive(context.Background())
	if sErr := <-done; sErr != nil || rErr != nil || env.Length != len(tricky) {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDSerialFloor, check,
			"Sec 4 live io.Pipe() SLIP transfer failed.")
	}
	return data.AuditPass(ModuleIDSerialFloor, check,
		"Rung 1 SLIP Serial Floor passed all 4 micro pre-flight gates.")
}

// AuditSerialFloor returns true iff PreflightSerialFloor() passes cleanly.
//
// ### For Architects, Operators & Non-Programmers:
// Quick pass/fail shortcut confirming copper-wire serial framing works.
//
// ### For Systems Programmers & Engine Architects:
// Convenience boolean bridge delegating to `PreflightSerialFloor().IsPass()`.
func AuditSerialFloor() bool {
	return PreflightSerialFloor().IsPass()
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK / EXECUTABLE PROCEDURES & LIFECYCLE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3A: Physical & Virtual Transport Substrate (Rung 1 Floor)
// Module:        cog.transport.rung1_serial
// Mission:       Expose epilogue readiness, 60-col serial cards, and lifecycle
//                construction/Close teardown for SerialTransceiver.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Executable Procedures, Inspection, Readiness & Cards
// ------------------------------------------------------------------------------

// IsSerialFloorReady reports whether Rung 1 SLIP serial passed pre-flight.
//
// ### For Architects, Operators & Non-Programmers:
// Epilogue confirmation that FaithNet's copper-wire serial floor is ready.
//
// ### For Systems Programmers & Engine Architects:
// Matches OmniCode's Block 5 epilogue readiness predicate (`is_*_ready`).
func IsSerialFloorReady() bool {
	return PreflightSerialFloor().IsPass()
}

// ExecuteSerialDumpCard renders a 60-column inner-width Rung 1 serial card.
//
// ### For Architects, Operators & Non-Programmers:
// Displays this serial link's ID, framing protocol, health, and packet tally.
//
// ### For Systems Programmers & Engine Architects:
// Non-mutating diagnostic card builder using lock-free atomic counters.
func (s *SerialTransceiver) ExecuteSerialDumpCard() string {
	return fmt.Sprintf(
		"+------------------------------------------------------------+\n"+
			"| %-58s |\n"+
			"+------------------------------------------------------------+\n"+
			"| Module:   %-48s |\n"+
			"| NodeID:   %-48s |\n"+
			"| Framing:  %-48s |\n"+
			"| Healthy:  %-48t |\n"+
			"| Traffic:  %-48s |\n"+
			"+------------------------------------------------------------+",
		"FAITHNET RUNG 1 SERIAL SLIP FLOOR CARD",
		ModuleIDSerialFloor,
		s.selfID.String(),
		"RFC 1055 SLIP (END=0xC0, ESC=0xDB, 512B cap)",
		s.Healthy(),
		fmt.Sprintf("%d sent / %d recv", s.sentCnt.Load(), s.recvCnt.Load()),
	)
}

// ------------------------------------------------------------------------------
// Sub-Assembly 5.2: Subsystem Lifecycle Initialization & Teardown
// ------------------------------------------------------------------------------

// NewSerialTransceiver wraps an io.Reader and io.Writer as a Rung 1 transceiver.
//
// ### For Architects, Operators & Non-Programmers:
// Connects a serial cable or raw byte stream into FaithNet's adaptive ladder.
//
// ### For Systems Programmers & Engine Architects:
// Accepts optional `closer` released once on `Close()`.
func NewSerialTransceiver(
	selfID data.NodeID,
	r io.Reader,
	w io.Writer,
	closer io.Closer,
) *SerialTransceiver {
	s := &SerialTransceiver{
		selfID: selfID,
		reader: r,
		writer: w,
		closer: closer,
	}
	s.healthy.Store(true)
	return s
}

// Close marks the serial link closed and closes its optional underlying handle
// plus any closable reader so blocked Receive() calls unblock immediately.
//
// ### For Architects, Operators & Non-Programmers:
// Releases the serial cable cleanly during shutdown.
//
// ### For Systems Programmers & Engine Architects:
// Idempotent via atomic CompareAndSwap; closes `s.reader` if it implements
// `io.Closer` so blocking `io.Pipe()` or TTY reads exit cleanly.
func (s *SerialTransceiver) Close() error {
	if !s.closed.CompareAndSwap(false, true) {
		return nil
	}
	s.healthy.Store(false)
	var firstErr error
	if s.closer != nil {
		firstErr = s.closer.Close()
	}
	if rc, ok := s.reader.(io.Closer); ok && rc != s.closer {
		if err := rc.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// End of transport/rung1_floor/serial.go
