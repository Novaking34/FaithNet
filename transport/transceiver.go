// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          transport/transceiver.go
// @id            cog.transport.transceiver
// @layer         Layer 3A: Physical & Virtual Transport Substrate
// @archetype     pure_types | stateless
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Universal Wire Abstraction Contract (Cog 3)
// @role          Pure Contract: Medium-Agnostic Transceiver Interface & Errors
// @maturity      Rev 5 Baseline | Wave 2 First Breath
// @chassis       zero-heap | Interface contract + fixed-value WireEnvelope
// @scale         O(1) interface dispatch across Rung 1..Rung 4 mediums
// @seams         in: data.NodeID, data.RungLevel | exports: Transceiver
// @invariants    1. Stratum Baseline: Downward Dependency Law, Universal
//                   5-Block Lifecycle Protocol (Rev 5).
//                2. Zero concrete OS socket imports in this contract leaf.
//                3. Every Rung 1..4 transceiver satisfies identical Send/Recv.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Defines Cog 3 of FaithNet: the universal wire boundary decoupling Layer 1
//   state machines and Layer 2 orchestration from any physical or virtual medium.
//   Whether a byte slice crosses Go channels in RAM, Unix IPC sockets, LAN UDP,
//   RS-232 copper wires, AFSK radio tones, or Morse CW pulses, the upper layers
//   see only `Transceiver`.
//
//   ### For Architects, Operators & Non-Programmers:
//   Acts like a universal wall plug for communication mediums. You can unplug
//   Wi-Fi and plug in a walkie-talkie or copper wire without changing a single
//   line of the FaithNet core engine.
//
//   ### For Systems Programmers & Engine Architects:
//   - Ownership & Reentrancy: Transceiver implementations must be safe for
//     concurrent Send(), Receive(), Healthy(), and Close() calls.
//   - Extension Seam: Implement `Transceiver` in any `transport/rung*_*/`
//     sub-package to register a new survival medium with AdaptiveLadder.
// ==============================================================================

package transport

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Sentinel Transport Errors, Logical Anchor & Transceiver Interface Contract )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3A: Physical & Virtual Transport Substrate
// Module:        cog.transport.transceiver
// Mission:       Declare the polymorphic Transceiver interface, WireEnvelope
//                value container, and shared transport sentinel errors.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"context"
	"errors"
)

// --- Tier 2: External Library Dependencies (None — 100% Go Stdlib Invariant) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---
import "github.com/CreativeWorkzStudio/FaithNet/data"

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Pre-Conditions & Transport Sentinel Errors
// ------------------------------------------------------------------------------
// Static sentinel errors returned uniformly across all 4 survival rungs.
// ------------------------------------------------------------------------------

var (
	ErrTransceiverClosed  = errors.New("faithnet/transport: transceiver closed")
	ErrTransceiverOffline = errors.New("faithnet/transport: medium offline")
	ErrPeerUnreachable    = errors.New("faithnet/transport: target peer unreachable")
	ErrWireFrameOversize  = errors.New("faithnet/transport: wire bytes exceed 512B floor")
	ErrReceiveTimeout     = errors.New("faithnet/transport: receive deadline expired")
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.2: Pure-Contract Logical Anchor
// ------------------------------------------------------------------------------
// Movement-resilient identity shared across orchestrator diagnostics.
// ------------------------------------------------------------------------------

const (
	// ModuleIDTransceiver is the canonical logical ID for Cog 3 contract.
	ModuleIDTransceiver = "cog.transport.transceiver"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.3: Public Tagged Ontology (WireEnvelope & Transceiver)
// ------------------------------------------------------------------------------
// Medium-agnostic wire envelope and universal 4-Rung transceiver contract.
// ------------------------------------------------------------------------------

// WireEnvelope carries one raw encoded FaithNet frame across a medium boundary.
type WireEnvelope struct {
	SenderID data.NodeID             // Physical/logical sender ID (if known)
	TargetID data.NodeID             // Target sovereign NodeID
	Length   int                     // Active valid byte count in Bytes (<= 512)
	Bytes    [data.FloorMaxFrameSize]byte // Inline zero-heap wire byte array
}

// Transceiver is the universal Cog 3 wire contract implemented across Rungs 1..4.
//
// ### For Architects, Operators & Non-Programmers:
// Any communication channel—RAM, cable, Wi-Fi, sound, or Morse—that can send
// bytes, receive bytes, report if it is healthy, and shut down cleanly can
// power FaithNet.
//
// ### For Systems Programmers & Engine Architects:
// Implementations own their internal buffers and OS handles cleanly. Receive
// honors context cancellation so Dispatcher shutdown never leaks goroutines.
type Transceiver interface {
	// Name returns a concise human-readable identifier (e.g. "rung2:memory").
	Name() string

	// Rung returns the survival tier (Rung1Floor..Rung4Global) of this medium.
	Rung() data.RungLevel

	// Healthy reports live link availability for Rising-Up / Falling-Down.
	Healthy() bool

	// Send transmits raw wire bytes toward target over this physical/virtual medium.
	Send(target data.NodeID, raw []byte) error

	// Receive blocks until one WireEnvelope arrives or ctx is cancelled.
	Receive(ctx context.Context) (WireEnvelope, error)

	// Close gracefully releases any OS sockets, channels, or hardware handles.
	Close() error
}

// ------------------------------------------------------------------------------
// Sub-Assembly 2.4: Pre-Conditional Helper Cogs & O(1) Envelope Views
// ------------------------------------------------------------------------------
// Pure zero-allocation slice view over active WireEnvelope bytes.
// ------------------------------------------------------------------------------

// Slice returns a zero-allocation view over the active wire bytes in e.
//
// ### For Architects, Operators & Non-Programmers:
// Reads only the live bytes inside a wire envelope without copying memory.
//
// ### For Systems Programmers & Engine Architects:
// Saturates Length cleanly into `[0 .. FloorMaxFrameSize]` to prevent panics.
func (e *WireEnvelope) Slice() []byte {
	n := e.Length
	if n < 0 {
		n = 0
	}
	if n > data.FloorMaxFrameSize {
		n = data.FloorMaxFrameSize
	}
	return e.Bytes[:n]
}

// ==============================================================================
// BLOCK 3: BODY BLOCK
// ( Statutory Pure-Types / Interface Contract Notice )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3A: Physical & Virtual Transport Substrate
// Module:        cog.transport.transceiver
// Mission:       Pure interface contract leaf: concrete transport engines
//                reside in `transport/rung1_floor/`, `transport/rung2_host/`,
//                `transport/rung3_local/`, and `transport/rung4_global/`.
// ==============================================================================

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Statutory Contract Audit Cross-Reference )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3A: Physical & Virtual Transport Substrate
// Module:        cog.transport.transceiver
// Mission:       Each concrete transceiver implementation verifies its own
//                Send/Receive/Healthy/Close contract in its Block 4 Audit &
//                Preflight pair and in `tests/integrated/` & `tests/unit/`.
// ==============================================================================

// ==============================================================================
// BLOCK 5: CLOSING BLOCK / STATUTORY EXEMPTION & LIFECYCLE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3A: Physical & Virtual Transport Substrate
// Module:        cog.transport.transceiver
// Mission:       Concrete lifecycle constructors (`New*`) and `Close()`
//                implementations reside in each Rung 1..4 transceiver package.
// Stateless Exemption Notice: None. (Stateless interface contract).
// ==============================================================================

// End of transport/transceiver.go
