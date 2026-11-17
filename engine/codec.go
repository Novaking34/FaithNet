// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          engine/codec.go
// @id            cog.engine.codec
// @layer         Layer 1: Pure Functional Engine Substrate
// @archetype     operational | stateless
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Deterministic Wire Codec & FNV-1a Integrity Engine
// @role          Runtime Engine: Zero-Allocation Binary Encoder, Decoder & Seal
// @maturity      Rev 5 Baseline | Wave 1 Seed
// @chassis       zero-heap | Pure []byte operations (0 allocs/op hot path)
// @scale         O(N) over payload length (N <= 480 bytes Floor MTU)
// @seams         in: data.Frame | uses: encoding/binary, errors, fmt | exports: codec
// @invariants    1. Stratum Baseline: ALLOC_ZERO_HEAP, Downward Dependency Law,
//                   Universal 5-Block Lifecycle Protocol (Rev 5).
//                2. EncodeInto & DecodeFrom perform strictly 0 allocs/op.
//                3. Big-endian network byte order across all multi-byte fields.
//                4. FNV-1a 32-bit checksum covers bytes [0..27] + payload bytes.
//                5. Strict magic (0xFA17), version, opcode, rung, and T4 checks.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Transforms FaithNet Frame structs into deterministic wire byte streams
//   and reconstructs them on the receiving side while verifying tamper-evident
//   FNV-1a integrity seals. Operates purely on caller-owned byte buffers so
//   it compiles identically for RAM, serial, radio, audio, or fiber.
//
//   ### For Architects, Operators & Non-Programmers:
//   Seals every FaithNet envelope with a mathematical wax stamp (FNV-1a) before
//   transmission and checks that stamp upon arrival—catching radio static,
//   damaged copper wires, or forged packets immediately.
//
//   ### For Systems Programmers & Engine Architects:
//   - Controlled Memory Chassis: Zero package-level mutable state; zero heap
//     allocations in ComputeFrameChecksum, EncodeInto, and DecodeFrom.
//   - Bottleneck & Headroom: Inline FNV-1a arithmetic avoids heap-allocated
//     hash.Hash32 interfaces across all 512-byte floor frames.
//   - Probe & Config Wiring: Block 4 exposes both AuditCodec (4.1) and
//     PreflightCodec (4.2) validating Layer 0 vocabulary + round-trip parity.
// ==============================================================================

package engine

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Constants, Sentinel Errors, Module Anchor & Monitored Support Helpers )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 1: Pure Functional Engine Substrate
// Module:        cog.engine.codec
// Mission:       Declare FNV-1a constants, zero-alloc sentinel errors, and
//                monitored bit-packing helpers ahead of pure Block 3 logic.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"encoding/binary"
	"errors"
	"fmt"
)

// --- Tier 2: External Library Dependencies (None — 100% Go Stdlib Invariant) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---
import "github.com/CreativeWorkzStudio/FaithNet/data"

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Pre-Conditions, Physical Bounds & Sentinel Errors
// ------------------------------------------------------------------------------
// Immutable FNV-1a arithmetic seeds and pre-allocated sentinel error handles.
// ------------------------------------------------------------------------------

const (
	// fnv1aOffset32 is the canonical 32-bit FNV-1a offset basis.
	fnv1aOffset32 uint32 = 2166136261

	// fnv1aPrime32 is the canonical 32-bit FNV-1a prime multiplier.
	fnv1aPrime32 uint32 = 16777619
)

var (
	ErrBufferTooSmall   = errors.New("faithnet/codec: buffer too small")
	ErrBadMagic         = errors.New("faithnet/codec: invalid frame magic")
	ErrUnsupportedVer   = errors.New("faithnet/codec: unsupported version")
	ErrInvalidRung      = errors.New("faithnet/codec: invalid rung level")
	ErrInvalidOpcode    = errors.New("faithnet/codec: invalid opcode")
	ErrInvalidQuadState = errors.New("faithnet/codec: invalid quad state")
	ErrPayloadTooLarge  = errors.New("faithnet/codec: payload exceeds floor MTU")
	ErrTruncatedPayload = errors.New("faithnet/codec: truncated payload")
	ErrChecksumMismatch = errors.New("faithnet/codec: FNV-1a checksum mismatch")
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.2: Operational Logical Anchor
// ------------------------------------------------------------------------------
// Movement-resilient identity shared across audits and diagnostic cards.
// ------------------------------------------------------------------------------

const (
	// ModuleIDCodec is the canonical logical ID for the wire codec cog.
	ModuleIDCodec = "cog.engine.codec"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.3: State Chassis Declaration (Stateless Operational Cog)
// ------------------------------------------------------------------------------
// Zero mutable package-level buffers or globals; 100% stack-reentrant.
// ------------------------------------------------------------------------------

// ------------------------------------------------------------------------------
// Sub-Assembly 2.4: Pre-Conditional Helper Cogs & O(1) Bit-Nibble Predicates
// ------------------------------------------------------------------------------
// Pure O(1) bit packing/unpacking for the combined Version/Rung header byte.
// ------------------------------------------------------------------------------

func packVerRung(ver uint8, rung data.RungLevel) uint8 {
	return (ver << 4) | (uint8(rung) & 0x0F)
}

func unpackVerRung(b uint8) (uint8, data.RungLevel) {
	return (b >> 4) & 0x0F, data.RungLevel(b & 0x0F)
}

// ------------------------------------------------------------------------------
// Sub-Assembly 2.5: Monitored Support Infrastructure
// ------------------------------------------------------------------------------
// Support Role:  Inline zero-allocation FNV-1a accumulator over wire slices.
// Pattern Watch: If multiple Layer 1/3 modules require FNV-1a hashing beyond
//                frame headers (e.g., cognitive state digests), elevate into a
//                standalone `engine/hash.go` cog.
// ------------------------------------------------------------------------------

func fnv1aUpdate(h uint32, buf []byte) uint32 {
	for i := 0; i < len(buf); i++ {
		h ^= uint32(buf[i])
		h *= fnv1aPrime32
	}
	return h
}

// ==============================================================================
// BLOCK 3: BODY BLOCK
// ( Composable Business Logic: Binary Frame Encoding, Decoding & Sealing )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 1: Pure Functional Engine Substrate
// Module:        cog.engine.codec
// Mission:       Provide zero-allocation wire serialization, deserialization,
//                and FNV-1a integrity verification over caller-owned buffers.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: Deterministic FNV-1a Frame Integrity Seal
// ------------------------------------------------------------------------------

// ComputeFrameChecksum calculates the 32-bit FNV-1a seal over header[0..27]
// and payload[0..payloadLen].
//
// ### For Architects, Operators & Non-Programmers:
// Produces a unique fingerprint for every packet so static or tampering flips
// the fingerprint and is caught before the data enters the system.
//
// ### For Systems Programmers & Engine Architects:
// O(28 + payloadLen) pure inline integer loop with 0 heap allocations.
func ComputeFrameChecksum(wire []byte, payloadLen int) uint32 {
	h := fnv1aUpdate(fnv1aOffset32, wire[:data.OffsetChecksum])
	if payloadLen > 0 {
		end := data.FrameHeaderSize + payloadLen
		h = fnv1aUpdate(h, wire[data.FrameHeaderSize:end])
	}
	return h
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: Zero-Allocation Binary Encoder & Decoder Cogs
// ------------------------------------------------------------------------------

// EncodeInto writes f into dst in canonical big-endian wire format and stamps
// the computed FNV-1a checksum into both dst and f.Checksum.
//
// ### For Architects, Operators & Non-Programmers:
// Packs a FaithNet message into a compact, tamper-sealed binary envelope ready
// for RAM, serial wire, radio waves, or fiber optics.
//
// ### For Systems Programmers & Engine Architects:
// Caller owns dst (must be >= 32 + f.PayloadLen bytes). Performs 0 allocs/op
// and validates protocol version, rung, opcode, QuadState, and Floor MTU.
func EncodeInto(dst []byte, f *data.Frame) (int, error) {
	if f == nil {
		return 0, ErrBufferTooSmall
	}
	payloadLen := int(f.PayloadLen)
	if payloadLen > data.FloorMaxPayload {
		return 0, ErrPayloadTooLarge
	}
	total := data.FrameHeaderSize + payloadLen
	if len(dst) < total {
		return 0, ErrBufferTooSmall
	}
	ver := f.Version
	if ver == 0 {
		ver = data.ProtocolVersion
	}
	if ver > 0x0F {
		return 0, ErrUnsupportedVer
	}
	if !f.Rung.IsValid() {
		return 0, ErrInvalidRung
	}
	if !f.Opcode.IsValid() {
		return 0, ErrInvalidOpcode
	}
	if !f.Truth.IsValid() {
		return 0, ErrInvalidQuadState
	}

	dst[0] = data.FrameMagicHi
	dst[1] = data.FrameMagicLo
	dst[data.OffsetVerRung] = packVerRung(ver, f.Rung)
	dst[data.OffsetOpcode] = uint8(f.Opcode)
	dst[data.OffsetFlags] = uint8(f.Flags)
	dst[data.OffsetQuadState] = uint8(f.Truth)
	binary.BigEndian.PutUint16(dst[data.OffsetPayloadLen:], f.PayloadLen)
	binary.BigEndian.PutUint64(dst[data.OffsetSourceID:], uint64(f.SourceID))
	binary.BigEndian.PutUint64(dst[data.OffsetTargetID:], uint64(f.TargetID))
	binary.BigEndian.PutUint32(dst[data.OffsetSeqID:], f.SeqID)

	if payloadLen > 0 {
		copy(dst[data.FrameHeaderSize:total], f.Payload[:payloadLen])
	}

	sum := ComputeFrameChecksum(dst, payloadLen)
	binary.BigEndian.PutUint32(dst[data.OffsetChecksum:], sum)
	f.Version = ver
	f.Checksum = sum
	return total, nil
}

// DecodeFrom parses wire bytes from src into out and validates magic, version,
// rung, opcode, QuadState, length bounds, and FNV-1a checksum integrity.
//
// ### For Architects, Operators & Non-Programmers:
// Unpacks an incoming envelope and refuses to open it if the seal is broken,
// the format is unknown, or the message was truncated in transit.
//
// ### For Systems Programmers & Engine Architects:
// Caller owns out (*data.Frame). Performs 0 allocs/op; leaves out untouched if
// any header or FNV-1a verification check fails.
func DecodeFrom(src []byte, out *data.Frame) error {
	if out == nil || len(src) < data.FrameHeaderSize {
		return ErrBufferTooSmall
	}
	if src[0] != data.FrameMagicHi || src[1] != data.FrameMagicLo {
		return ErrBadMagic
	}

	ver, rung := unpackVerRung(src[data.OffsetVerRung])
	if ver != data.ProtocolVersion {
		return ErrUnsupportedVer
	}
	if !rung.IsValid() {
		return ErrInvalidRung
	}

	op := data.Opcode(src[data.OffsetOpcode])
	if !op.IsValid() {
		return ErrInvalidOpcode
	}

	truth := data.QuadState(src[data.OffsetQuadState])
	if !truth.IsValid() {
		return ErrInvalidQuadState
	}

	payloadLen := binary.BigEndian.Uint16(src[data.OffsetPayloadLen:])
	if int(payloadLen) > data.FloorMaxPayload {
		return ErrPayloadTooLarge
	}
	total := data.FrameHeaderSize + int(payloadLen)
	if len(src) < total {
		return ErrTruncatedPayload
	}

	wireSum := binary.BigEndian.Uint32(src[data.OffsetChecksum:])
	calcSum := ComputeFrameChecksum(src, int(payloadLen))
	if wireSum != calcSum {
		return ErrChecksumMismatch
	}

	out.Version = ver
	out.Rung = rung
	out.Opcode = op
	out.Flags = data.FrameFlags(src[data.OffsetFlags])
	out.Truth = truth
	out.PayloadLen = payloadLen
	out.SourceID = data.NodeID(binary.BigEndian.Uint64(src[data.OffsetSourceID:]))
	out.TargetID = data.NodeID(binary.BigEndian.Uint64(src[data.OffsetTargetID:]))
	out.SeqID = binary.BigEndian.Uint32(src[data.OffsetSeqID:])
	out.Checksum = wireSum

	if payloadLen > 0 {
		copy(out.Payload[:payloadLen], src[data.FrameHeaderSize:total])
	}
	return nil
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Contract Verification Cogs & Pre-Flight Validation Tripwires )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 1: Pure Functional Engine Substrate
// Module:        cog.engine.codec
// Mission:       Verify Layer 0 vocabulary invariants, 512-byte sector layout,
//                live Frame contracts, and 4-section pre-flight codec parity.
// @triggers
//   - Block 4 Audits : AuditDataVocabulary(), ValidateFrameContract(),
//                      PreflightCodec(), AuditCodec()
//   - Layer 4 Tests  : tests/functional/data_functional_test.go,
//                      tests/functional/engine_functional_test.go,
//                      tests/unit/engine/codec_test.go
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Contract, Input & Live-Instance Verification Cogs
// ------------------------------------------------------------------------------

// AuditDataVocabulary verifies Layer 0 pure_types invariants (NodeID, RungLevel,
// QuadState, Opcode pinning, AuditResult rules, 512-byte sector layout, and
// offline-first manifest defaults) in stack memory without heap allocation.
//
// ### For Architects, Operators & Non-Programmers:
// Double-checks FaithNet's core dictionary at startup to guarantee survival
// rungs, truth verdicts, and 512-byte envelope geometry have never drifted.
//
// ### For Systems Programmers & Engine Architects:
// Stack-pure O(1) contract gate returning structured `data.AuditResult` with
// `AuditSeverityFatal` on any violated Layer 0 structural invariant.
func AuditDataVocabulary() data.AuditResult {
	const check = "layer0_data_vocabulary_contract"

	if data.NodeIDNull.IsValid() || !data.NodeIDBroadcast.IsBroadcast() {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDCodec, check,
			"NodeID null/broadcast boundary predicates violated.")
	}
	if data.RungNone.IsValid() || !data.Rung1Floor.IsValid() ||
		!data.Rung4Global.IsValid() ||
		data.Rung1Floor.StepDown() != data.Rung1Floor ||
		data.Rung4Global.StepUp() != data.Rung4Global {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDCodec, check,
			"RungLevel bounds or saturating ladder step predicates violated.")
	}
	if !data.QuadNull.IsValid() || !data.QuadPositive.IsValid() ||
		data.QuadState(0x04).IsValid() {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDCodec, check,
			"QuadState T4 domain bounds [0x00..0x03] violated.")
	}
	if data.OpInvalid.IsValid() || uint8(data.OpPing) != 0x01 ||
		uint8(data.OpPong) != 0x02 || !data.OpPing.IsControl() ||
		data.OpData.IsControl() {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDCodec, check,
			"Opcode wire pinning or control-verb classification violated.")
	}
	if data.FrameHeaderSize != 32 || data.FloorMaxPayload != 480 ||
		data.FloorMaxFrameSize != 512 ||
		data.OffsetChecksum+4 != data.FrameHeaderSize {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDCodec, check,
			"512-byte Floppy Sector header/payload geometry violated.")
	}
	def := data.DefaultNodeManifest(0x01, "audit")
	if !def.IsValid() || def.AllowGlobalEgress ||
		def.MinimumFloorRung != data.Rung1Floor {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDCodec, check,
			"DefaultNodeManifest breached offline-first sovereignty invariant.")
	}
	if !data.AuditPass(ModuleIDCodec, check, "ok").IsPass() ||
		data.AuditFail(data.AuditSeverityPass, ModuleIDCodec, check, "err").IsPass() {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDCodec, check,
			"AuditResult pass/fail consistency guard violated.")
	}
	return data.AuditPass(ModuleIDCodec, check,
		"Layer 0 vocabulary, 512B sector geometry, and offline defaults verified.")
}

// ValidateFrameContract audits a caller-supplied Frame non-destructively
// against FaithNet's wire header and 512-byte Floppy Sector bounds.
//
// ### For Architects, Operators & Non-Programmers:
// Inspects an envelope before sending or processing to confirm its version,
// rung, opcode, truth badge, and size all fit safely inside FaithNet's rules.
//
// ### For Systems Programmers & Engine Architects:
// Read-only O(1) contract validator (0 allocs/op) returning structured
// `data.AuditResult` without mutating caller memory.
func ValidateFrameContract(f *data.Frame) data.AuditResult {
	const check = "frame_wire_contract"

	if f == nil {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDCodec, check,
			"Frame pointer is nil.")
	}
	if f.Version != 0 && f.Version != data.ProtocolVersion {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDCodec, check,
			"Frame Version exceeds supported ProtocolVersion.")
	}
	if !f.Rung.IsValid() {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDCodec, check,
			"Frame Rung lies outside canonical 4-Rung Ladder [1..4].")
	}
	if !f.Opcode.IsValid() {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDCodec, check,
			"Frame Opcode is invalid or unrecognized.")
	}
	if !f.Truth.IsValid() {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDCodec, check,
			"Frame QuadState lies outside canonical T4 space [0..3].")
	}
	if int(f.PayloadLen) > data.FloorMaxPayload {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDCodec, check,
			"Frame PayloadLen exceeds 480-byte Floppy Sector ceiling.")
	}
	return data.AuditPass(ModuleIDCodec, check,
		"Frame satisfies header, T4, and 512-byte Floppy Sector invariants.")
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Machine Self-Audit & Pre-Flight Validation Tripwires
// ------------------------------------------------------------------------------

// PreflightCodec executes a 4-section non-destructive pre-flight self-audit
// covering (1) nominal round-trip, (2) fault inversion & non-mutation,
// (3) 512B sector ceiling limits, and (4) emergent multi-rung parity.
//
// ### For Architects, Operators & Non-Programmers:
// Runs a complete startup bench test of the envelope packer and tamper seal so
// FaithNet never transmits or accepts corrupted frames.
//
// ### For Systems Programmers & Engine Architects:
// Stack-pure pre-flight gate mirroring Cornerstone's 4-Section micro protocol
// (`ALLOC_ZERO_HEAP`); returns `data.AuditResult` with exact fault telemetry.
func PreflightCodec() data.AuditResult {
	const check = "codec_preflight_4section"

	if vocab := AuditDataVocabulary(); !vocab.IsPass() {
		return vocab
	}

	// Micro Check 1 (Always Pass): Nominal round-trip encode & decode parity.
	var in, out data.Frame
	var buf [data.FloorMaxFrameSize]byte
	in.Version = data.ProtocolVersion
	in.Rung = data.Rung1Floor
	in.Opcode = data.OpPing
	in.Flags = data.FlagFloorCompact
	in.Truth = data.QuadPositive
	in.SourceID = 0xAA01
	in.TargetID = 0xBB02
	in.SeqID = 7
	in.SetPayload([]byte("FAITH"))

	if v := ValidateFrameContract(&in); !v.IsPass() {
		return v
	}
	n, err := EncodeInto(buf[:], &in)
	if err != nil || n != data.FrameHeaderSize+5 {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDCodec, check,
			"Sec 1 nominal EncodeInto failed or produced wrong wire length.")
	}
	if err := DecodeFrom(buf[:n], &out); err != nil {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDCodec, check,
			"Sec 1 nominal DecodeFrom rejected valid encoded frame.")
	}
	if out.SourceID != in.SourceID || out.TargetID != in.TargetID ||
		out.SeqID != in.SeqID || out.Checksum != in.Checksum ||
		out.PayloadLen != 5 {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDCodec, check,
			"Sec 1 decoded frame fields diverged from encoded source.")
	}

	// Micro Check 2 (Never Fail): Fault inversion & out-parameter non-mutation.
	if _, err := EncodeInto(buf[:4], &in); err != ErrBufferTooSmall {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDCodec, check,
			"Sec 2 undersized EncodeInto buffer was not rejected.")
	}
	if err := DecodeFrom(buf[:n-1], &out); err != ErrTruncatedPayload {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDCodec, check,
			"Sec 2 truncated wire buffer was not rejected.")
	}
	savedSeq := out.SeqID
	buf[data.FrameHeaderSize] ^= 0x01
	if err := DecodeFrom(buf[:n], &out); err != ErrChecksumMismatch {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDCodec, check,
			"Sec 2 single-bit payload tamper bypassed FNV-1a seal.")
	}
	if out.SeqID != savedSeq {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDCodec, check,
			"Sec 2 failing DecodeFrom mutated caller output frame.")
	}
	buf[data.FrameHeaderSize] ^= 0x01

	// Micro Check 3 (Invariants & Limits): Exact 512-byte ceiling & 481B reject.
	var maxF data.Frame
	maxF.Rung = data.Rung4Global
	maxF.Opcode = data.OpData
	maxF.Truth = data.QuadPositive
	maxF.PayloadLen = uint16(data.FloorMaxPayload)
	if fullN, err := EncodeInto(buf[:], &maxF); err != nil || fullN != data.FloorMaxFrameSize {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDCodec, check,
			"Sec 3 exact 512-byte Floppy Sector frame failed encoding.")
	}
	maxF.PayloadLen = uint16(data.FloorMaxPayload + 1)
	if _, err := EncodeInto(buf[:], &maxF); err != ErrPayloadTooLarge {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDCodec, check,
			"Sec 3 481-byte oversized payload bypassed Floppy Sector ceiling.")
	}

	// Micro Check 4 (Emergence): Reset cleanliness & multi-rung re-encode.
	out.Reset()
	if out.PayloadLen != 0 || out.SourceID != data.NodeIDNull {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDCodec, check,
			"Sec 4 Frame.Reset() left residual state behind.")
	}
	in.Rung = data.Rung3Local
	n2, err := EncodeInto(buf[:], &in)
	if err != nil || DecodeFrom(buf[:n2], &out) != nil || out.Rung != data.Rung3Local {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDCodec, check,
			"Sec 4 post-reset multi-rung re-encode parity failed.")
	}

	return data.AuditPass(ModuleIDCodec, check,
		"Codec passed all 4 micro pre-flight gates (nominal/fault/512B/reset).")
}

// AuditCodec returns true iff PreflightCodec() completes with AuditSeverityPass.
//
// ### For Architects, Operators & Non-Programmers:
// Quick pass/fail shortcut asking whether the binary wire codec passed safety.
//
// ### For Systems Programmers & Engine Architects:
// Convenience boolean bridge delegating to `PreflightCodec().IsPass()`.
func AuditCodec() bool {
	return PreflightCodec().IsPass()
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK / EXECUTABLE PROCEDURES & LIFECYCLE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 1: Pure Functional Engine Substrate
// Module:        cog.engine.codec
// Mission:       Emit 60-column diagnostic inspection cards, expose epilogue
//                readiness handshakes, and declare stateless lifecycle compliance.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Executable Procedures, Inspection, Readiness & Cards
// ------------------------------------------------------------------------------

// IsCodecReady reports whether Cog 2's wire codec passed full pre-flight audit.
//
// ### For Architects, Operators & Non-Programmers:
// Epilogue confirmation that FaithNet's envelope system is healthy and ready.
//
// ### For Systems Programmers & Engine Architects:
// Matches OmniCode's Block 5 epilogue readiness predicate (`is_*_ready`).
func IsCodecReady() bool {
	return PreflightCodec().IsPass()
}

// ExecuteCodecDumpCard renders a 60-column inner-width diagnostic card
// reporting FaithNet wire frame geometry and pre-flight audit status.
//
// ### For Architects, Operators & Non-Programmers:
// Prints a neat framed status card summarizing envelope size, seal type, and
// whether startup safety checks passed.
//
// ### For Systems Programmers & Engine Architects:
// Pure diagnostic formatter invoking `PreflightCodec()` non-destructively.
func ExecuteCodecDumpCard() string {
	pf := PreflightCodec()
	return fmt.Sprintf(
		"+------------------------------------------------------------+\n"+
			"| %-58s |\n"+
			"+------------------------------------------------------------+\n"+
			"| Module:    %-47s |\n"+
			"| Header:    %-47s |\n"+
			"| Floor MTU: %-47s |\n"+
			"| Integrity: %-47s |\n"+
			"| Preflight: %-47s |\n"+
			"+------------------------------------------------------------+",
		"FAITHNET WIRE CODEC DIAGNOSTIC CARD",
		ModuleIDCodec,
		fmt.Sprintf("%d bytes (Magic 0x%04X, Ver %d)",
			data.FrameHeaderSize, data.FrameMagic, data.ProtocolVersion),
		fmt.Sprintf("%d bytes payload (%d bytes Floppy Sector)",
			data.FloorMaxPayload, data.FloorMaxFrameSize),
		"FNV-1a 32-bit inline (ALLOC_ZERO_HEAP)",
		pf.Severity.String(),
	)
}

// ------------------------------------------------------------------------------
// Sub-Assembly 5.2: Subsystem Lifecycle Initialization & Teardown
// ------------------------------------------------------------------------------
// Stateless Exemption Notice: None. (Stateless module / pure stack-allocated execution).
// ------------------------------------------------------------------------------

// End of engine/codec.go
