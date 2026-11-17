// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          data/frame.go
// @id            cog.data.frame
// @layer         Layer 0: Pure Invariant Data Substrate
// @archetype     pure_types | stateless
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Fixed-Header Wire Frame & Floppy Sector Invariant
// @role          Pure Contract & Vocabulary: 32B Header + 480B Inline Frame
// @maturity      Rev 5 Baseline | Wave 1 Seed
// @chassis       zero-heap | Inline 512-byte value layout (zero slice pointers)
// @scale         O(1) fixed-size frame; exact 512-byte Floppy Sector ceiling
// @seams         in: data/types.go, data/opcodes.go | exports: Frame, offsets
// @invariants    1. Stratum Baseline: ALLOC_ZERO_HEAP, Downward Dependency Law,
//                   Universal 5-Block Lifecycle Protocol (Rev 5).
//                2. FrameHeaderSize == 32 bytes (64-bit aligned wire layout).
//                3. FloorMaxPayload == 480 bytes; Header + Payload == 512 bytes.
//                4. Magic bytes permanently sealed as 0xFA, 0x17 ("FAIT[H]").
//                5. Zero dynamic heap allocations: Payload lives inline in struct.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Defines the canonical binary envelope (Cog 2) of FaithNet. Every packet
//   sent over Morse code, analog audio chirps, serial copper wire, RAM
//   channels, local Wi-Fi, or TCP/IP shares this exact 32-byte header and
//   512-byte Floppy Sector ceiling.
//
//   ### For Architects, Operators & Non-Programmers:
//   Think of Frame as a standardized, unbreakable postal envelope that fits
//   inside a single 512-byte floppy disk sector. Whether carried by sound
//   waves or fiber optics, every node knows exactly where the sender,
//   recipient, intent, and tamper-evident seal sit inside the envelope.
//
//   ### For Systems Programmers & Engine Architects:
//   - Ownership & Reentrancy: Value-typed struct with inline `[480]byte`
//     payload backing array eliminating heap escapes and GC scanning.
//   - Extension Seam: All header offsets live in Sub-Assembly 2.1 so codec
//     implementations stay bit-exact across languages (Go, Rust, C11, ASM).
//   - Explicit Non-Goal: Performs zero wire serialization or FNV-1a hashing
//     (owned by Layer 1 `engine/codec.go`).
// ==============================================================================

package data

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Shared Pure Structural Vocabulary, Value Bounds & Stateless Predicates )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 0: Pure Invariant Data Substrate
// Module:        cog.data.frame
// Mission:       Provide immutable 32-byte header offsets, 512-byte Floppy
//                Sector bounds, inline Frame struct, and O(1) payload accessors.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages (None — Pure Primitive Vocabulary) ---

// --- Tier 2: External Library Dependencies (None — 100% Go Stdlib Invariant) ---

// --- Tier 3: Internal FaithNet Foundation Packages (Same Package: data) ---

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Pre-Conditions, Physical Bounds & Wire Header Offsets
// ------------------------------------------------------------------------------
// Immutable frame geometry sealing the 32-byte header and 512-byte Floor MTU.
// ------------------------------------------------------------------------------

const (
	// FrameMagicHi is byte 0 of the FaithNet frame magic signature (0xFA).
	FrameMagicHi uint8 = 0xFA

	// FrameMagicLo is byte 1 of the FaithNet frame magic signature (0x17).
	FrameMagicLo uint8 = 0x17

	// FrameMagic is the big-endian 16-bit signature 0xFA17 ("FAIT[H]").
	FrameMagic uint16 = 0xFA17

	// ProtocolVersion is the current 4-bit wire protocol revision (1..15).
	ProtocolVersion uint8 = 1

	// FrameHeaderSize is the exact fixed header length in bytes.
	FrameHeaderSize int = 32

	// FloorMaxPayload is the maximum payload bytes in a Floppy Sector frame.
	FloorMaxPayload int = 480

	// FloorMaxFrameSize is the exact 512-byte Floppy Sector Invariant ceiling.
	FloorMaxFrameSize int = FrameHeaderSize + FloorMaxPayload

	// PingTimestampBytes is the 8-byte int64 monotonic timestamp prefix for
	// OpPing and OpPong payloads.
	PingTimestampBytes int = 8
)

// Canonical header byte offsets (strictly 32 bytes total).
const (
	OffsetMagic      = 0  // 2 bytes: [0..1]   0xFA17
	OffsetVerRung    = 2  // 1 byte:  [2]      High nibble=Version, Low=Rung
	OffsetOpcode     = 3  // 1 byte:  [3]      Opcode
	OffsetFlags      = 4  // 1 byte:  [4]      FrameFlags bitmask
	OffsetQuadState  = 5  // 1 byte:  [5]      QuadState (T4)
	OffsetPayloadLen = 6  // 2 bytes: [6..7]   PayloadLen (uint16 big-endian)
	OffsetSourceID   = 8  // 8 bytes: [8..15]  SourceID (uint64 big-endian)
	OffsetTargetID   = 16 // 8 bytes: [16..23] TargetID (uint64 big-endian)
	OffsetSeqID      = 24 // 4 bytes: [24..27] SeqID (uint32 big-endian)
	OffsetChecksum   = 28 // 4 bytes: [28..31] FNV-1a 32-bit checksum
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.2: Pure-Types Logical Anchor
// ------------------------------------------------------------------------------
// Movement-resilient identifier shared by diagnostic and audit subsystems.
// ------------------------------------------------------------------------------

const (
	// ModuleIDFrame is the canonical logical ID for this pure_types leaf.
	ModuleIDFrame = "cog.data.frame"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.3: Public Tagged Ontology (Inline Zero-Heap Frame Struct)
// ------------------------------------------------------------------------------
// Canonical caller-owned POD envelope carrying 32B header + 480B inline buffer.
// ------------------------------------------------------------------------------

// Frame is the canonical, zero-heap FaithNet wire envelope (Cog 2).
//
// ### For Architects, Operators & Non-Programmers:
// Represents one complete message packet—who sent it, who should receive it,
// which survival rung carried it, its truth status, and up to 480 bytes of
// content sealed inside a single 512-byte floppy sector.
//
// ### For Systems Programmers & Engine Architects:
// Zero-pointer POD struct (`sizeof(Frame) == ~512B`). Allocate on stack or
// inside fixed ring arrays; never allocates heap memory for payload storage.
type Frame struct {
	Version    uint8                 // 4-bit wire version (ProtocolVersion = 1)
	Rung       RungLevel             // Active survival rung (Rung1..Rung4)
	Opcode     Opcode                // Protocol verb (PING, PONG, BEACON, etc.)
	Flags      FrameFlags            // Provenance and airlock bitmask flags
	Truth      QuadState             // OmniCode T4 verdict (-1, 0, +1, NULL)
	PayloadLen uint16                // Active payload length (0..480)
	SourceID   NodeID                // 64-bit sovereign origin address
	TargetID   NodeID                // 64-bit sovereign destination address
	SeqID      uint32                // Monotonic per-origin sequence counter
	Checksum   uint32                // FNV-1a 32-bit wire integrity seal
	Payload    [FloorMaxPayload]byte // Inline zero-heap payload backing array
}

// ------------------------------------------------------------------------------
// Sub-Assembly 2.4: Pre-Conditional Helper Cogs & Stateless Value Accessors
// ------------------------------------------------------------------------------
// Pure O(1) inline payload bounds setters, zero-copy views, and value reset.
// ------------------------------------------------------------------------------

// SetPayload copies src into the inline Payload array and updates PayloadLen.
// Returns false without mutating Payload if src exceeds FloorMaxPayload.
//
// ### For Architects, Operators & Non-Programmers:
// Places message bytes safely inside the 512-byte envelope and refuses any
// oversized payload that would break the floppy-sector limit.
//
// ### For Systems Programmers & Engine Architects:
// Zero-heap inline copy into fixed `[480]byte` backing array; preserves prior
// state atomically when `len(src) > FloorMaxPayload`.
func (f *Frame) SetPayload(src []byte) bool {
	if len(src) > FloorMaxPayload {
		return false
	}
	copy(f.Payload[:], src)
	f.PayloadLen = uint16(len(src))
	return true
}

// PayloadSlice returns a zero-allocation view over the active payload bytes.
//
// ### For Architects, Operators & Non-Programmers:
// Reads only the live content bytes inside the envelope without copying memory.
//
// ### For Systems Programmers & Engine Architects:
// Saturates malformed `PayloadLen > FloorMaxPayload` cleanly so callers can
// never trigger an out-of-bounds slice panic.
func (f *Frame) PayloadSlice() []byte {
	n := int(f.PayloadLen)
	if n > FloorMaxPayload {
		n = FloorMaxPayload
	}
	return f.Payload[:n]
}

// WireSize returns the total encoded byte size of the frame (header + payload).
//
// ### For Architects, Operators & Non-Programmers:
// Reports the exact byte length this envelope will occupy on the wire.
//
// ### For Systems Programmers & Engine Architects:
// Pure O(1) sum (`FrameHeaderSize + int(f.PayloadLen)`).
func (f *Frame) WireSize() int {
	return FrameHeaderSize + int(f.PayloadLen)
}

// IsHeaderBoundsValid reports whether scalar header fields satisfy Layer 0
// domain bounds (Rung, Opcode, Truth, PayloadLen <= FloorMaxPayload).
//
// ### For Architects, Operators & Non-Programmers:
// Performs an instant sanity check on an envelope's header before transmission
// or routing.
//
// ### For Systems Programmers & Engine Architects:
// Pure O(1) stateless predicate with 0 heap allocations.
func (f *Frame) IsHeaderBoundsValid() bool {
	if f == nil {
		return false
	}
	return f.Rung.IsValid() &&
		f.Opcode.IsValid() &&
		f.Truth.IsValid() &&
		int(f.PayloadLen) <= FloorMaxPayload
}

// Reset zeroes the Frame cleanly for stack/ring reuse without heap allocation.
//
// ### For Architects, Operators & Non-Programmers:
// Wipes a reusable envelope completely blank so no prior data leaks forward.
//
// ### For Systems Programmers & Engine Architects:
// Value assignment `*f = Frame{}` compiles to an inline memclr with 0 allocs.
func (f *Frame) Reset() {
	*f = Frame{}
}

// ==============================================================================
// BLOCK 3: BODY BLOCK
// ( Statutory Pure-Types Notice & Structural Boundary )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 0: Pure Invariant Data Substrate
// Module:        cog.data.frame
// Mission:       Pure-types frame vocabulary module: define purely static
//                layout constants, inline Frame struct, and O(1) value helpers;
//                wire encoding, decoding, and FNV-1a hashing reside in Layer 1
//                (`engine/codec.go`).
// ==============================================================================

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Statutory Pure-Types Audit Cross-Reference )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 0: Pure Invariant Data Substrate
// Module:        cog.data.frame
// Mission:       Functional audits, 512-byte sector layout tripwires, and
//                pre-flight routines verifying Frame geometry are declared in
//                Layer 1 (`engine/codec.go`: `AuditCodec`, `PreflightCodec`).
// ==============================================================================

// ==============================================================================
// BLOCK 5: CLOSING BLOCK / STATUTORY EXEMPTION & LIFECYCLE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 0: Pure Invariant Data Substrate
// Module:        cog.data.frame
// Mission:       Operational inspection cards, subsystem lifecycle init, and
//                teardown routines are declared in Layer 1 (`engine/`) and
//                Layer 2 (`orchestrator/`).
// Stateless Exemption Notice: None. (Stateless module / pure stack-allocated execution).
// ==============================================================================

// End of data/frame.go
