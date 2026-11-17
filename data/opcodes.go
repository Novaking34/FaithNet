// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          data/opcodes.go
// @id            cog.data.opcodes
// @layer         Layer 0: Pure Invariant Data Substrate
// @archetype     pure_types | stateless
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Protocol Intent & Frame Flag Vocabulary
// @role          Pure Contract & Vocabulary: Opcode Verbs & FrameFlags Bitmasks
// @maturity      Rev 5 Baseline | Wave 1 Seed
// @chassis       zero-heap | Zero stateful globals; 1-byte scalar enums
// @scale         O(1) range predicates & bitwise flag checks; zero allocations
// @seams         in: none (pure leaf) | uses: none | exports: Opcode, FrameFlags
// @invariants    1. Stratum Baseline: ALLOC_ZERO_HEAP, Downward Dependency Law,
//                   Universal 5-Block Lifecycle Protocol (Rev 5).
//                2. Zero OS, network, or external package imports.
//                3. OpPing (0x01) and OpPong (0x02) are permanently pinned.
//                4. All bitmask flags fit cleanly within a single uint8 byte.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Provides the pure single-byte protocol verbs (`Opcode`) and provenance
//   flags (`FrameFlags`) carried in every FaithNet wire header. Cleanly
//   separates sonar heartbeats, CPI-SI cognitive sync, adaptive ladder
//   cascades, and legacy OS translations ("The Well").
//
//   ### For Architects, Operators & Non-Programmers:
//   Defines the shared action words of FaithNet—calling out (PING), echoing
//   back (PONG), announcing presence (BEACON), sharing knowledge (CPI_SYNC),
//   climbing or stepping down the survival ladder, or safely wrapping foreign
//   web/OS messages.
//
//   ### For Systems Programmers & Engine Architects:
//   - Ownership & Reentrancy: Pure 1-byte value types; completely stateless.
//   - Extension Seam: Add new protocol verbs cleanly inside Sub-Assembly 2.3
//     and update `OpcodeMaxValid` in Sub-Assembly 2.1.
//   - Explicit Non-Goal: Holds zero dispatch handlers or state tables (owned
//     by `engine/` and `orchestrator/dispatcher.go`).
// ==============================================================================

package data

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Shared Pure Structural Vocabulary, Value Bounds & Stateless Predicates )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 0: Pure Invariant Data Substrate
// Module:        cog.data.opcodes
// Mission:       Provide shared 1-byte Opcode and FrameFlags vocabularies with
//                O(1) validation and bitwise query predicates.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages (None — Pure Primitive Vocabulary) ---

// --- Tier 2: External Library Dependencies (None — 100% Go Stdlib Invariant) ---

// --- Tier 3: Internal FaithNet Foundation Packages (None — Pure Leaf Root) ---

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Pre-Conditions, Physical Bounds & Opcode Range Limits
// ------------------------------------------------------------------------------
// Pinned wire-byte boundaries guaranteeing strict ingress opcode validation.
// ------------------------------------------------------------------------------

const (
	// OpcodeMinValid is the lowest valid protocol verb (OpPing = 0x01).
	OpcodeMinValid Opcode = OpPing

	// OpcodeMaxValid is the highest recognized protocol verb (OpData = 0x08).
	OpcodeMaxValid Opcode = OpData
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.2: Pure-Types Logical Anchor
// ------------------------------------------------------------------------------
// Movement-resilient identifier shared by diagnostic and audit subsystems.
// ------------------------------------------------------------------------------

const (
	// ModuleIDOpcodes is the canonical logical ID for this pure_types leaf.
	ModuleIDOpcodes = "cog.data.opcodes"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.3: Public Tagged Ontology (Opcode Verbs & FrameFlags)
// ------------------------------------------------------------------------------
// Canonical 1-byte wire intent verbs and provenance bitmasks.
// ------------------------------------------------------------------------------

// Opcode identifies the semantic intent of a FaithNet wire frame.
//
// ### For Architects, Operators & Non-Programmers:
// Tells a receiving device what the sender wants to do (ping, echo, announce
// presence, sync cognitive state, cascade rungs, or bridge a legacy OS).
//
// ### For Systems Programmers & Engine Architects:
// Serialized directly at header byte [3] (OffsetOpcode). Invalid values (0x00
// or > OpcodeMaxValid) are rejected by `engine.DecodeFrom`.
type Opcode uint8

const (
	// OpInvalid marks an uninitialized or malformed opcode.
	OpInvalid Opcode = 0x00

	// OpPing is an echo request pulse ("Are you present? Reflect this back.").
	OpPing Opcode = 0x01

	// OpPong is an echo reply pulse ("I hear you; here is your echo.").
	OpPong Opcode = 0x02

	// OpBeacon announces a sovereign node's presence and active rung level.
	OpBeacon Opcode = 0x03

	// OpCPISync exchanges sovereign CPI-SI cognitive state or episodic deltas.
	OpCPISync Opcode = 0x04

	// OpLegacyWrap encapsulates foreign OS/Web payloads ("The Well" bridge).
	OpLegacyWrap Opcode = 0x05

	// OpCascadeDown signals EVT_CASCADE_DOWN onto a lower survival rung.
	OpCascadeDown Opcode = 0x06

	// OpCascadeUp signals graceful escalation onto a higher-capacity rung.
	OpCascadeUp Opcode = 0x07

	// OpData carries bounded general-purpose sovereign payload bytes.
	OpData Opcode = 0x08
)

// FrameFlags holds bitwise boolean attributes for a FaithNet wire frame.
//
// ### For Architects, Operators & Non-Programmers:
// Stamps each envelope with clear provenance tags—whether it came from a
// verified local peer, entered through the outside web airlock, or was
// translated from a Windows/Mac/Android device.
//
// ### For Systems Programmers & Engine Architects:
// Serialized at header byte [4] (OffsetFlags). Orthogonal bit positions permit
// O(1) mask checks without allocations.
type FrameFlags uint8

const (
	// FlagNone indicates default unflagged frame semantics.
	FlagNone FrameFlags = 0x00

	// FlagCovenantVerified marks a frame audited by the Genesis 3 Airlock.
	FlagCovenantVerified FrameFlags = 1 << 0

	// FlagExternalOrigin marks data originating from the legacy commercial web.
	FlagExternalOrigin FrameFlags = 1 << 1

	// FlagLegacyTranslated marks a frame bridged via The Well (interpreter/).
	FlagLegacyTranslated FrameFlags = 1 << 2

	// FlagFloorCompact marks a frame strictly bounded to the 512B Floor MTU.
	FlagFloorCompact FrameFlags = 1 << 3

	// FlagRequiresAck requests explicit delivery confirmation from receiver.
	FlagRequiresAck FrameFlags = 1 << 4
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.4: Pre-Conditional Helper Cogs & Stateless Value Predicates
// ------------------------------------------------------------------------------
// Pure O(1) opcode validation, flag mask predicates, and symbolic stringifiers.
// ------------------------------------------------------------------------------

// IsValid reports whether op is a recognized FaithNet protocol opcode.
//
// ### For Architects, Operators & Non-Programmers:
// Confirms a packet carries a known FaithNet action verb rather than noise.
//
// ### For Systems Programmers & Engine Architects:
// O(1) closed-interval check [OpcodeMinValid .. OpcodeMaxValid].
func (op Opcode) IsValid() bool {
	return op >= OpcodeMinValid && op <= OpcodeMaxValid
}

// IsControl reports whether op is a low-level link-control verb (PING, PONG,
// EVT_CASCADE_DOWN, or EVT_CASCADE_UP).
//
// ### For Architects, Operators & Non-Programmers:
// Distinguishes mechanical heartbeat and ladder-shift pulses from user data
// or cognitive sync messages.
//
// ### For Systems Programmers & Engine Architects:
// O(1) pure predicate used by the Genesis 3 Airlock and Dispatcher routing.
func (op Opcode) IsControl() bool {
	return op == OpPing || op == OpPong ||
		op == OpCascadeDown || op == OpCascadeUp
}

// String returns the canonical symbolic name of the opcode.
//
// ### For Architects, Operators & Non-Programmers:
// Renders an envelope's action byte as a clear word (PING, PONG, BEACON, etc.).
//
// ### For Systems Programmers & Engine Architects:
// O(1) switch returning static string literals (0 heap allocations).
func (op Opcode) String() string {
	switch op {
	case OpPing:
		return "PING"
	case OpPong:
		return "PONG"
	case OpBeacon:
		return "BEACON"
	case OpCPISync:
		return "CPI_SYNC"
	case OpLegacyWrap:
		return "LEGACY_WRAP"
	case OpCascadeDown:
		return "EVT_CASCADE_DOWN"
	case OpCascadeUp:
		return "EVT_CASCADE_UP"
	case OpData:
		return "DATA"
	default:
		return "INVALID"
	}
}

// Has reports whether all bits in mask are set on f.
//
// ### For Architects, Operators & Non-Programmers:
// Checks whether an envelope carries one or more provenance stamps (such as
// Covenant Verified or External Origin).
//
// ### For Systems Programmers & Engine Architects:
// O(1) bitwise AND equality test (`(f & mask) == mask`).
func (f FrameFlags) Has(mask FrameFlags) bool {
	return (f & mask) == mask
}

// ==============================================================================
// BLOCK 3: BODY BLOCK
// ( Statutory Pure-Types Notice & Structural Boundary )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 0: Pure Invariant Data Substrate
// Module:        cog.data.opcodes
// Mission:       Pure-types vocabulary module: define purely static opcodes,
//                bitmask flags, and pure O(1) predicates; operational frame
//                dispatch resides in `engine/` and `orchestrator/dispatcher.go`.
// ==============================================================================

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Statutory Pure-Types Audit Cross-Reference )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 0: Pure Invariant Data Substrate
// Module:        cog.data.opcodes
// Mission:       Functional audits, opcode pinning checks, and pre-flight
//                routines verifying these invariants are declared in Layer 1
//                (`engine/codec.go`: `AuditCodec`, `PreflightCodec`).
// ==============================================================================

// ==============================================================================
// BLOCK 5: CLOSING BLOCK / STATUTORY EXEMPTION & LIFECYCLE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 0: Pure Invariant Data Substrate
// Module:        cog.data.opcodes
// Mission:       Operational inspection cards, subsystem lifecycle init, and
//                teardown routines are declared in Layer 1 (`engine/`) and
//                Layer 2 (`orchestrator/`).
// Stateless Exemption Notice: None. (Stateless module / pure stack-allocated execution).
// ==============================================================================

// End of data/opcodes.go
