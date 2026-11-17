// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          data/types.go
// @id            cog.data.types
// @layer         Layer 0: Pure Invariant Data Substrate
// @archetype     pure_types | stateless
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Sovereign Identity & Truth-Space Vocabulary
// @role          Pure Contract & Vocabulary: NodeID, RungLevel, QuadState (T4)
// @maturity      Rev 5 Baseline | Wave 1 Seed
// @chassis       zero-heap | Zero stateful globals; scalar value types (1..8B)
// @scale         O(1) range predicates & formatting; zero dynamic allocations
// @seams         in: none (pure leaf) | uses: fmt | exports: NodeID, Rung, T4
// @invariants    1. Stratum Baseline: ALLOC_ZERO_HEAP, Downward Dependency Law,
//                   Universal 5-Block Lifecycle Protocol (Rev 5).
//                2. Pure stateless types invariant: zero package-level mutable
//                   state and zero OS or network imports.
//                3. NodeID is 8-byte uint64 — substrate-independent identity.
//                4. QuadState faithfully encodes OmniCode T4 (-1, 0, +1, NULL).
//                5. RungLevel strictly bounds the 4-Rung Adaptive Ladder (1..4).
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Provides the shared, standalone Layer 0 scalar vocabulary for FaithNet so
//   every operational engine, adaptive orchestrator, physical transceiver, and
//   legacy interpreter shares identical identity, survival-rung, and T4 truth
//   semantics without circular imports.
//
//   ### For Architects, Operators & Non-Programmers:
//   Defines the shared dictionary of sovereign device identities (NodeID),
//   the four survival ladder rungs (from Morse/radio floor up to global), and
//   OmniCode's four-state truth verdicts (+1, 0, -1, NULL).
//
//   ### For Systems Programmers & Engine Architects:
//   - Ownership & Reentrancy: Pure caller-owned scalar value types; 100%
//     stateless and safe across concurrent goroutines without locks.
//   - Extension Seam: Add canonical rungs or T4 metadata in one leaf file
//     inherited cleanly across Layers 1..4.
//   - Explicit Non-Goal: Holds zero stateful buffers, I/O handles, or wire
//     serialization loops (owned by Layer 1 `engine/`).
// ==============================================================================

package data

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Shared Pure Structural Vocabulary, Value Bounds & Stateless Predicates )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 0: Pure Invariant Data Substrate
// Module:        cog.data.types
// Mission:       Provide shared sovereign identity, rung, and T4 scalar types
//                with pure O(1) stateless range predicates.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import "fmt"

// --- Tier 2: External Library Dependencies (None — 100% Go Stdlib Invariant) ---

// --- Tier 3: Internal FaithNet Foundation Packages (None — Pure Leaf Root) ---

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Pre-Conditions, Physical Bounds & Scalar Limits
// ------------------------------------------------------------------------------
// Fixed wire-width bounds shared across all 4 adaptive rungs.
//
// ### For Architects, Operators & Non-Programmers:
// Locks the exact range of survival rungs (1 through 4) and the four truth
// states so corrupted or out-of-range numbers are rejected immediately.
//
// ### For Systems Programmers & Engine Architects:
// Compile-time untyped/typed integer limits guaranteeing 1-byte wire encodings
// for RungLevel (low nibble) and QuadState (1 byte), and 8-byte NodeIDs.
// ------------------------------------------------------------------------------

const (
	// NodeIDNull represents an uninitialized or absent node identity.
	NodeIDNull NodeID = 0x0000000000000000

	// NodeIDBroadcast addresses all listening peers within immediate reach.
	NodeIDBroadcast NodeID = 0xFFFFFFFFFFFFFFFF

	// RungMinValid is the lowest valid survival tier (Civilizational Floor).
	RungMinValid RungLevel = Rung1Floor

	// RungMaxValid is the highest valid survival tier (Federated Backbone).
	RungMaxValid RungLevel = Rung4Global
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.2: Pure-Types Logical Anchor
// ------------------------------------------------------------------------------
// Movement-resilient identifier shared by diagnostic and audit subsystems.
// ------------------------------------------------------------------------------

const (
	// ModuleIDTypes is the canonical logical ID for this pure_types leaf.
	ModuleIDTypes = "cog.data.types"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.3: Public Tagged Ontology (Identity, Ladder & T4 Truth Space)
// ------------------------------------------------------------------------------
// Canonical scalar value types forming the bedrock vocabulary of FaithNet.
// ------------------------------------------------------------------------------

// NodeID is a 64-bit sovereign node identifier independent of IPv4/IPv6.
//
// ### For Architects, Operators & Non-Programmers:
// A permanent digital name tag for a device that works offline in a room,
// across a street over radio, or on a global network without asking an ISP.
//
// ### For Systems Programmers & Engine Architects:
// Fixed-width uint64 serialized big-endian at header offsets [8..15] and
// [16..23]. Zero heap allocation; directly comparable with ==.
type NodeID uint64

// RungLevel identifies a physical or logical tier in the 4-Rung Ladder.
//
// ### For Architects, Operators & Non-Programmers:
// Tracks which transport tier a node is currently riding—climbing up when
// modern Wi-Fi or fiber is healthy, and falling down gracefully to radio or
// copper wire when infrastructure degrades.
//
// ### For Systems Programmers & Engine Architects:
// Stored in the lower 4 bits of header byte [2] (OffsetVerRung). Valid range
// is [Rung1Floor (1) .. Rung4Global (4)].
type RungLevel uint8

const (
	// RungNone indicates no active transport (floor collapsed / uninitialized).
	RungNone RungLevel = 0

	// Rung1Floor is the civilizational floor (Serial, AFSK Audio, Morse CW).
	Rung1Floor RungLevel = 1

	// Rung2Host is the host-local substrate (In-Memory channels, Unix IPC).
	Rung2Host RungLevel = 2

	// Rung3Local is the intimate room & neighborhood airspace (UDP, Wi-Fi AP).
	Rung3Local RungLevel = 3

	// Rung4Global is optional modern backbone compatibility (TCP, Airlock).
	Rung4Global RungLevel = 4
)

// QuadState represents the OmniCode Balanced Ternary Quad-State truth space.
//
// ### For Architects, Operators & Non-Programmers:
// Captures whether a piece of information is verified and aligned (+1),
// neutral data waiting quietly for review (0), rejected as harmful (-1), or
// structurally broken and halted (NULL).
//
// ### For Systems Programmers & Engine Architects:
// Encoded in header byte [5] (OffsetQuadState) as [0x00..0x03] so every wire
// frame carries explicit epistemic provenance without string allocations.
type QuadState uint8

const (
	// QuadNull (NULL) indicates boundary halt, corruption, or uninitialized.
	QuadNull QuadState = 0x00

	// QuadNegative (-1) indicates rejection, rollback, or covenant violation.
	QuadNegative QuadState = 0x01

	// QuadAwait (0) indicates neutral equilibrium, empirical data, or standby.
	QuadAwait QuadState = 0x02

	// QuadPositive (+1) indicates verified truth, alignment, and acceptance.
	QuadPositive QuadState = 0x03
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.4: Pre-Conditional Helper Cogs & Stateless Value Predicates
// ------------------------------------------------------------------------------
// Pure O(1) stateless range guards and symbolic stringifiers safe across all
// goroutines and layers.
// ------------------------------------------------------------------------------

// IsValid reports whether the NodeID is non-null.
func (id NodeID) IsValid() bool {
	return id != NodeIDNull
}

// IsBroadcast reports whether the NodeID targets all reachable peers.
func (id NodeID) IsBroadcast() bool {
	return id == NodeIDBroadcast
}

// String renders NodeID as a canonical 16-digit hexadecimal sovereign address.
func (id NodeID) String() string {
	return fmt.Sprintf("0x%016X", uint64(id))
}

// IsValid reports whether r lies within the canonical 4-Rung Ladder (1..4).
func (r RungLevel) IsValid() bool {
	return r >= RungMinValid && r <= RungMaxValid
}

// String returns the human-readable name of the adaptive ladder rung.
func (r RungLevel) String() string {
	switch r {
	case Rung1Floor:
		return "RUNG_1_FLOOR"
	case Rung2Host:
		return "RUNG_2_HOST"
	case Rung3Local:
		return "RUNG_3_LOCAL"
	case Rung4Global:
		return "RUNG_4_GLOBAL"
	default:
		return "RUNG_NONE"
	}
}

// IsValid reports whether q is one of the four canonical T4 states.
func (q QuadState) IsValid() bool {
	return q <= QuadPositive
}

// String renders the QuadState using OmniCode canonical T4 notation.
func (q QuadState) String() string {
	switch q {
	case QuadPositive:
		return "+1:ALIGNED"
	case QuadAwait:
		return "0:AWAIT"
	case QuadNegative:
		return "-1:REJECT"
	default:
		return "NULL:HALT"
	}
}

// ==============================================================================
// BLOCK 3: BODY BLOCK
// ( Statutory Pure-Types Notice & Structural Boundary )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 0: Pure Invariant Data Substrate
// Module:        cog.data.types
// Mission:       Pure-types vocabulary module: define purely static structures,
//                immutable enums, and pure O(1) predicates; runtime dispatch,
//                binary codecs, and adaptive state machines reside in Layer 1
//                (`engine/`) and Layer 2 (`orchestrator/`).
// ==============================================================================

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Statutory Pure-Types Audit Cross-Reference )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 0: Pure Invariant Data Substrate
// Module:        cog.data.types
// Mission:       Functional audits, layout tripwires, and pre-flight routines
//                verifying these scalar bounds are declared in Layer 1
//                (`engine/codec.go`: `AuditCodec`, `PreflightCodec`).
// ==============================================================================

// ==============================================================================
// BLOCK 5: CLOSING BLOCK / STATUTORY EXEMPTION & LIFECYCLE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 0: Pure Invariant Data Substrate
// Module:        cog.data.types
// Mission:       Operational inspection cards, subsystem lifecycle init, and
//                teardown routines are declared in Layer 1 (`engine/`) and
//                Layer 2 (`orchestrator/`).
// Stateless Exemption Notice: None. (Stateless module / pure stack-allocated execution).
// ==============================================================================

// End of data/types.go
