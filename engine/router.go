// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          engine/router.go
// @id            cog.engine.router
// @layer         Layer 1: Pure Functional Engine Substrate
// @archetype     operational | stateful
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Data-Oriented Peer & Next-Hop Routing Table
// @role          Runtime Engine: Fixed-Capacity Sovereign Mesh Route Resolver
// @maturity      Rev 5 Baseline | Wave 1 Seed
// @chassis       zero-heap | Fixed [128]PeerRoute array (controlled memory)
// @scale         O(N) linear scan over N <= 128 cache-contiguous peer records
// @seams         in: data.NodeID, data.RungLevel | exports: Router, PeerRoute
// @invariants    1. Stratum Baseline: ALLOC_ZERO_HEAP, Downward Dependency Law,
//                   Universal 5-Block Lifecycle Protocol (Rev 5).
//                2. Bounded memory chassis: never allocates past MaxPeers (128).
//                3. Supports direct peers (hops=1) and multi-hop relays
//                   ("Across the Street" mesh forwarding).
//                4. Thread-safe via internal sync.RWMutex.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Maintains the sovereign node's knowledge of reachable peers, which
//   adaptive rung reaches each peer best, and which intermediate neighbor
//   acts as next-hop relay when a peer sits beyond direct single-hop range.
//
//   ### For Architects, Operators & Non-Programmers:
//   Acts like a neighborhood address book that remembers who is in the room,
//   who is across the street, and which neighbor in the middle can pass a
//   message down the block if you cannot reach someone directly.
//
//   ### For Systems Programmers & Engine Architects:
//   - Controlled Memory Chassis: Contiguous `[MaxPeers]PeerRoute` array keeps
//     the entire routing table inside CPU cache with zero hash-map heap churn.
//   - Probe & Config Wiring: Block 4 verifies direct + multi-hop route math;
//     Block 5 owns lifecycle (`NewRouter`, `Reset`), snapshots, and dump cards.
// ==============================================================================

package engine

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Bounded Peer Limits, Sentinel Errors, Route Ontology & O(1) Queries )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 1: Pure Functional Engine Substrate
// Module:        cog.engine.router
// Mission:       Declare MaxPeers bounds, sentinel errors, PeerRoute/Router
//                structs, and O(1) capacity predicates ahead of Block 3 logic.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"errors"
	"fmt"
	"sync"
)

// --- Tier 2: External Library Dependencies (None — 100% Go Stdlib Invariant) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---
import "github.com/CreativeWorkzStudio/FaithNet/data"

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Pre-Conditions, Physical Bounds & Sentinel Errors
// ------------------------------------------------------------------------------
// Explicit Controlled Memory Chassis limits for peer storage.
// ------------------------------------------------------------------------------

const (
	// MaxPeers is the fixed upper bound on tracked peers in one Router table.
	MaxPeers = 128
)

var (
	ErrRouteTableFull  = errors.New("faithnet/router: peer table at capacity")
	ErrRouteInvalidID  = errors.New("faithnet/router: invalid peer node ID")
	ErrRouteInvalidHop = errors.New("faithnet/router: hop count must be >= 1")
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.2: Operational Logical Anchor
// ------------------------------------------------------------------------------
// Movement-resilient identity shared across audits and diagnostic cards.
// ------------------------------------------------------------------------------

const (
	// ModuleIDRouter is the canonical logical ID for the peer routing cog.
	ModuleIDRouter = "cog.engine.router"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.3: Public Tagged Ontology (PeerRoute & Router Chassis)
// ------------------------------------------------------------------------------
// Contiguous value records mapping sovereign NodeIDs to next-hop relays.
// ------------------------------------------------------------------------------

// PeerRoute maps a destination sovereign NodeID to its best next hop and rung.
type PeerRoute struct {
	PeerID      data.NodeID    // Sovereign destination node identifier
	NextHopID   data.NodeID    // Immediate relay peer (equals PeerID if 1-hop)
	Rung        data.RungLevel // Best active survival rung reaching NextHopID
	HopCount    uint8          // Mesh hop distance (>= 1)
	LastSeenNs  int64          // Monotonic timestamp of most recent contact
	LastRTTNano int64          // Most recent measured round-trip latency (ns)
	Active      bool           // Slot occupancy flag in fixed array
}

// Router holds a cache-contiguous, fixed-capacity sovereign routing table.
type Router struct {
	mu     sync.RWMutex
	selfID data.NodeID
	routes [MaxPeers]PeerRoute
	count  int
}

// ------------------------------------------------------------------------------
// Sub-Assembly 2.4: Pre-Conditional Helper Cogs & O(1) Capacity Predicates
// ------------------------------------------------------------------------------
// Read-only O(1) table occupancy and capacity checks.
// ------------------------------------------------------------------------------

// ActiveCount returns the current number of active peer routes.
func (r *Router) ActiveCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.count
}

// Capacity returns the fixed Controlled Memory Chassis ceiling (MaxPeers).
func (r *Router) Capacity() int {
	return MaxPeers
}

// ==============================================================================
// BLOCK 3: BODY BLOCK
// ( Composable Business Logic: Route Upsert, Next-Hop Lookup & Removal )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 1: Pure Functional Engine Substrate
// Module:        cog.engine.router
// Mission:       Insert, update, resolve, and retire direct or multi-hop peer
//                routes deterministically with zero heap allocations.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: Peer Route Registration & Updating
// ------------------------------------------------------------------------------

// UpsertRoute inserts a new peer route or updates an existing peer's entry.
//
// ### For Architects, Operators & Non-Programmers:
// Adds a newly discovered neighbor to the local map or updates how to reach
// them if a faster rung or shorter hop path becomes available.
//
// ### For Systems Programmers & Engine Architects:
// O(MaxPeers) contiguous array scan under write lock; enforces non-self PeerID,
// valid RungLevel, and HopCount >= 1. Defaults NextHopID to PeerID when unset.
func (r *Router) UpsertRoute(entry PeerRoute) error {
	if !entry.PeerID.IsValid() || entry.PeerID == r.selfID {
		return ErrRouteInvalidID
	}
	if !entry.Rung.IsValid() {
		return ErrInvalidRung
	}
	if entry.HopCount == 0 {
		return ErrRouteInvalidHop
	}
	if !entry.NextHopID.IsValid() {
		entry.NextHopID = entry.PeerID
	}
	entry.Active = true

	r.mu.Lock()
	defer r.mu.Unlock()

	freeIdx := -1
	for i := 0; i < MaxPeers; i++ {
		if r.routes[i].Active && r.routes[i].PeerID == entry.PeerID {
			r.routes[i] = entry
			return nil
		}
		if !r.routes[i].Active && freeIdx == -1 {
			freeIdx = i
		}
	}
	if freeIdx == -1 {
		return ErrRouteTableFull
	}
	r.routes[freeIdx] = entry
	r.count++
	return nil
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: Next-Hop Resolution & Route Retirement
// ------------------------------------------------------------------------------

// Lookup retrieves the active route for peerID.
//
// ### For Architects, Operators & Non-Programmers:
// Checks whether we know how to reach a specific neighbor and tells the node
// which survival rung and intermediate relay should carry the packet.
//
// ### For Systems Programmers & Engine Architects:
// O(MaxPeers) read-locked scan returning a value copy (0 heap allocations).
func (r *Router) Lookup(peerID data.NodeID) (PeerRoute, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for i := 0; i < MaxPeers; i++ {
		if r.routes[i].Active && r.routes[i].PeerID == peerID {
			return r.routes[i], true
		}
	}
	return PeerRoute{}, false
}

// Remove deletes a peer route if present and frees its fixed slot.
//
// ### For Architects, Operators & Non-Programmers:
// Clears an unreachable or departed neighbor from the local routing table.
//
// ### For Systems Programmers & Engine Architects:
// Zeroes the slot inline and decrements active count atomically under lock.
func (r *Router) Remove(peerID data.NodeID) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := 0; i < MaxPeers; i++ {
		if r.routes[i].Active && r.routes[i].PeerID == peerID {
			r.routes[i] = PeerRoute{}
			r.count--
			return true
		}
	}
	return false
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Contract Verification Cogs & Pre-Flight Validation Tripwires )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 1: Pure Functional Engine Substrate
// Module:        cog.engine.router
// Mission:       Verify direct + multi-hop route resolution, self-route
//                rejection, and bounded slot accounting before dispatch boot.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Contract, Input & Cog Verification Cogs
// ------------------------------------------------------------------------------

// AuditRouter verifies direct and multi-hop insertion, lookup, and removal.
func AuditRouter() bool {
	rt := NewRouter(0x01)
	if rt.UpsertRoute(PeerRoute{PeerID: 0x01, Rung: data.Rung2Host, HopCount: 1}) != ErrRouteInvalidID {
		return false
	}
	err := rt.UpsertRoute(PeerRoute{
		PeerID:    0x03,
		NextHopID: 0x02, // Across-the-street relay via Node 0x02
		Rung:      data.Rung3Local,
		HopCount:  2,
	})
	if err != nil || rt.ActiveCount() != 1 {
		return false
	}
	got, ok := rt.Lookup(0x03)
	if !ok || got.NextHopID != 0x02 || got.HopCount != 2 {
		return false
	}
	return rt.Remove(0x03) && rt.ActiveCount() == 0
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Machine Self-Audit & Pre-Flight Validation Tripwires
// ------------------------------------------------------------------------------

// PreflightRouter executes complete non-destructive pre-flight clearance.
func PreflightRouter() bool {
	return MaxPeers > 0 && AuditRouter()
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK / EXECUTABLE PROCEDURES & LIFECYCLE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 1: Pure Functional Engine Substrate
// Module:        cog.engine.router
// Mission:       Provide caller-owned route snapshots, 60-column inspection
//                cards, and Router lifecycle initialization/reset.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Executable Procedures, Inspection & Diagnostic Cards
// ------------------------------------------------------------------------------

// SnapshotInto copies active routes into caller-owned dst and returns count.
func (r *Router) SnapshotInto(dst []PeerRoute) int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	written := 0
	for i := 0; i < MaxPeers && written < len(dst); i++ {
		if r.routes[i].Active {
			dst[written] = r.routes[i]
			written++
		}
	}
	return written
}

// ExecuteRouterDumpCard renders a 60-column inner-width diagnostic card.
func (r *Router) ExecuteRouterDumpCard() string {
	return fmt.Sprintf(
		"+------------------------------------------------------------+\n"+
			"| %-58s |\n"+
			"+------------------------------------------------------------+\n"+
			"| Module:   %-48s |\n"+
			"| SelfID:   %-48s |\n"+
			"| Routes:   %-48s |\n"+
			"| Chassis:  %-48s |\n"+
			"+------------------------------------------------------------+",
		"FAITHNET SOVEREIGN ROUTER DIAGNOSTIC CARD",
		ModuleIDRouter,
		r.selfID.String(),
		fmt.Sprintf("%d active / %d max capacity", r.ActiveCount(), MaxPeers),
		"Fixed [128]PeerRoute array (ALLOC_ZERO_HEAP)",
	)
}

// ------------------------------------------------------------------------------
// Sub-Assembly 5.2: Subsystem Lifecycle Initialization & Teardown
// ------------------------------------------------------------------------------

// NewRouter initializes a bounded sovereign Router belonging to selfID.
func NewRouter(selfID data.NodeID) *Router {
	return &Router{selfID: selfID}
}

// Reset clears all active peer routes cleanly without reallocating memory.
func (r *Router) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.routes = [MaxPeers]PeerRoute{}
	r.count = 0
}

// End of engine/router.go
