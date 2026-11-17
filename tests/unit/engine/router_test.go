// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/unit/engine/router_test.go
// @id            test.unit.engine.router
// @layer         Layer 4: Tier-2 Unit Verification Suite (Engine Sub-Stratum)
// @archetype     verification_suite | unit_isolation
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Layer 1 Bounded 128-Peer Sovereign Router Unit Suite
// @role          Tier-2 Unit Isolation Gate for cog.engine.router
// @maturity      Rev 5 Baseline | 4x4 Verification Matrix
// @chassis       zero-heap | Verifies fixed [128]PeerRoute array & zero maps
// @scale         O(MaxPeers) deterministic upsert, lookup, remove & snapshot
// @seams         in: data, engine, tests/common | target: engine/router.go
// @invariants    1. Hard MaxPeers (128) ceiling returns ErrRouteTableFull.
//                2. Updating an existing peer never double-counts ActiveCount.
//                3. SnapshotInto truncates cleanly to caller dst length.
//                4. Concurrent read/write access is 100% -race clean.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Tier-2 Unit isolation suite strictly verifying `engine/router.go`:
//   direct and multi-hop ("Across the Street") routes, in-place upsert updates,
//   error sentinels, 128-peer capacity guard, bounded snapshots, live
//   `AuditState`, and concurrent `-race` safety.
//
//   ### For Architects, Operators & Non-Programmers:
//   Confirms FaithNet's neighborhood route directory tracks direct neighbors
//   and multi-hop relays cleanly without ever overflowing memory.
//
//   ### For Systems Programmers & Engine Architects:
//   100% branch coverage of `UpsertRoute`, `Lookup`, `Remove`, `ActiveCount`,
//   `Capacity`, `SnapshotInto`, `AuditState`, `Reset`, and dump card.
// ==============================================================================

package engine_unit_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Identifier, Verification Imports & Tier-2 Unit Sensor Anchor )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.engine.router
// Mission:       Bind Router symbols and shared fixtures for isolated tests.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"strings"
	"sync"
	"testing"
)

// --- Tier 2: External Library Dependencies (None) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---
import (
	"github.com/CreativeWorkzStudio/FaithNet/data"
	"github.com/CreativeWorkzStudio/FaithNet/engine"
	"github.com/CreativeWorkzStudio/FaithNet/tests/common"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Operational Suite Constants & Identifiers
// ------------------------------------------------------------------------------

const (
	routerUnitSuiteName = "Tier 2 Unit: cog.engine.router (128-Peer Table)"
)

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.engine.router
// Mission:       Verify Router upsert, lookup, removal, 128-cap, and -race.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: [ SECTION 1: ALWAYS PASS (NOMINAL GOLDEN SURFACE COGS) ]
// ------------------------------------------------------------------------------

func TestRouterUnit_Sec1_DirectMultiHopAndInPlaceUpsert(t *testing.T) {
	rt := engine.NewRouter(common.NodeAlpha)
	if rt.Capacity() != engine.MaxPeers || rt.ActiveCount() != 0 {
		t.Fatalf("[SEC 1] initial router capacity/count mismatch: cap=%d cnt=%d",
			rt.Capacity(), rt.ActiveCount())
	}

	// Direct peer without NextHopID defaults NextHopID = PeerID.
	if err := rt.UpsertRoute(engine.PeerRoute{
		PeerID: common.NodeBeta, Rung: data.Rung3Local, HopCount: 1, LastRTTNano: 500,
	}); err != nil {
		t.Fatalf("[SEC 1] direct upsert failed: %v", err)
	}
	got, ok := rt.Lookup(common.NodeBeta)
	if !ok || got.NextHopID != common.NodeBeta || got.HopCount != 1 {
		t.Fatalf("[SEC 1] direct lookup mismatch: %+v", got)
	}

	// In-place update improves RTT without incrementing ActiveCount.
	if err := rt.UpsertRoute(engine.PeerRoute{
		PeerID: common.NodeBeta, Rung: data.Rung2Host, HopCount: 1, LastRTTNano: 90,
	}); err != nil {
		t.Fatalf("[SEC 1] in-place update failed: %v", err)
	}
	if rt.ActiveCount() != 1 {
		t.Fatalf("[SEC 1] in-place upsert double-counted peer: count=%d", rt.ActiveCount())
	}
	updated, _ := rt.Lookup(common.NodeBeta)
	if updated.Rung != data.Rung2Host || updated.LastRTTNano != 90 {
		t.Fatalf("[SEC 1] in-place update did not apply: %+v", updated)
	}

	// Across-the-street 2-hop relay via NodeBeta.
	if err := rt.UpsertRoute(engine.PeerRoute{
		PeerID: common.NodeRelay, NextHopID: common.NodeBeta,
		Rung: data.Rung3Local, HopCount: 2,
	}); err != nil {
		t.Fatalf("[SEC 1] relay upsert failed: %v", err)
	}
	rel, ok := rt.Lookup(common.NodeRelay)
	if !ok || rel.NextHopID != common.NodeBeta || rel.HopCount != 2 {
		t.Fatalf("[SEC 1] relay lookup mismatch: %+v", rel)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: [ SECTION 2: NEVER FAIL (FAULT INVERSION & DEGRADATION) ]
// ------------------------------------------------------------------------------

func TestRouterUnit_Sec2_RejectsInvalidRoutesAndMissingRemoval(t *testing.T) {
	rt := engine.NewRouter(common.NodeAlpha)

	if err := rt.UpsertRoute(engine.PeerRoute{
		PeerID: data.NodeIDNull, Rung: data.Rung2Host, HopCount: 1,
	}); err != engine.ErrRouteInvalidID {
		t.Fatalf("[SEC 2] null PeerID: got %v", err)
	}
	if err := rt.UpsertRoute(engine.PeerRoute{
		PeerID: common.NodeAlpha, Rung: data.Rung2Host, HopCount: 1,
	}); err != engine.ErrRouteInvalidID {
		t.Fatalf("[SEC 2] self-route: got %v", err)
	}
	if err := rt.UpsertRoute(engine.PeerRoute{
		PeerID: common.NodeBeta, Rung: data.RungNone, HopCount: 1,
	}); err != engine.ErrInvalidRung {
		t.Fatalf("[SEC 2] RungNone: got %v", err)
	}
	if err := rt.UpsertRoute(engine.PeerRoute{
		PeerID: common.NodeBeta, Rung: data.Rung2Host, HopCount: 0,
	}); err != engine.ErrRouteInvalidHop {
		t.Fatalf("[SEC 2] 0 hops: got %v", err)
	}
	if _, found := rt.Lookup(common.NodeBeta); found {
		t.Fatal("[SEC 2] Lookup on empty table returned found=true")
	}
	if rt.Remove(common.NodeBeta) {
		t.Fatal("[SEC 2] Remove of absent peer returned true")
	}

	var nilRT *engine.Router
	common.AssertAuditFail(t, nilRT.AuditState(), data.AuditSeverityCritical)
	badRT := engine.NewRouter(data.NodeIDNull)
	common.AssertAuditFail(t, badRT.AuditState(), data.AuditSeverityCritical)
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: [ SECTION 3: INVARIANTS, SATURATION & LIMIT GUARDS ]
// ------------------------------------------------------------------------------

func TestRouterUnit_Sec3_Hard128CapacitySnapshotBoundsAndConcurrency(t *testing.T) {
	rt := engine.NewRouter(common.NodeAlpha)
	for i := 0; i < engine.MaxPeers; i++ {
		if err := rt.UpsertRoute(engine.PeerRoute{
			PeerID: data.NodeID(0x5000 + i), Rung: data.Rung2Host, HopCount: 1,
		}); err != nil {
			t.Fatalf("[SEC 3] fill %d failed: %v", i, err)
		}
	}
	if rt.ActiveCount() != engine.MaxPeers {
		t.Fatalf("[SEC 3] expected count=%d got %d", engine.MaxPeers, rt.ActiveCount())
	}
	if err := rt.UpsertRoute(engine.PeerRoute{
		PeerID: 0x9999, Rung: data.Rung2Host, HopCount: 1,
	}); err != engine.ErrRouteTableFull {
		t.Fatalf("[SEC 3] 129th peer: want ErrRouteTableFull got %v", err)
	}

	// Bounded SnapshotInto truncation & nil slice safety.
	if rt.SnapshotInto(nil) != 0 {
		t.Fatal("[SEC 3] SnapshotInto(nil) must return 0")
	}
	var small [5]engine.PeerRoute
	if n := rt.SnapshotInto(small[:]); n != 5 {
		t.Fatalf("[SEC 3] SnapshotInto([5]) want 5 got %d", n)
	}

	// Remove 1 slot and confirm new peer reuses freed slot cleanly.
	if !rt.Remove(0x5000) || rt.ActiveCount() != engine.MaxPeers-1 {
		t.Fatal("[SEC 3] Remove did not free slot cleanly")
	}
	if err := rt.UpsertRoute(engine.PeerRoute{
		PeerID: 0x9999, Rung: data.Rung1Floor, HopCount: 1,
	}); err != nil || rt.ActiveCount() != engine.MaxPeers {
		t.Fatalf("[SEC 3] slot reuse failed: err=%v count=%d", err, rt.ActiveCount())
	}

	// Concurrent -race safety proof.
	var wg sync.WaitGroup
	for g := 0; g < 4; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			peer := data.NodeID(0x5010 + id)
			for i := 0; i < 40; i++ {
				_ = rt.UpsertRoute(engine.PeerRoute{
					PeerID: peer, Rung: data.Rung3Local, HopCount: 1, LastRTTNano: int64(i),
				})
				_, _ = rt.Lookup(peer)
				_ = rt.ActiveCount()
			}
		}(g)
	}
	wg.Wait()
	common.AssertAuditPass(t, rt.AuditState(), rt.ExecuteRouterDumpCard())
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.4: [ SECTION 4: EMERGENCE & COMPOSITE WORKFLOWS ]
// ------------------------------------------------------------------------------

func TestRouterUnit_Sec4_DumpCardAndCleanReset(t *testing.T) {
	rt := engine.NewRouter(common.NodeAlpha)
	_ = rt.UpsertRoute(engine.PeerRoute{
		PeerID: common.NodeRelay, NextHopID: common.NodeBeta, Rung: data.Rung3Local, HopCount: 2,
	})
	card := rt.ExecuteRouterDumpCard()
	if !strings.Contains(card, "1 active / 128 max capacity") {
		t.Fatalf("[SEC 4] unexpected router dump card:\n%s", card)
	}
	rt.Reset()
	if rt.ActiveCount() != 0 {
		t.Fatal("[SEC 4] Router.Reset() left active routes behind")
	}
	common.AssertAuditPass(t, rt.AuditState(), rt.ExecuteRouterDumpCard())
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Preflight Diagnostic Gates & Physical Source Anchor Verification )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.engine.router
// Mission:       Run PreflightRouter and verify physical source anchor.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Block 4 Pre-Flight Contract Gates
// ------------------------------------------------------------------------------

func TestRouterUnit_Block4_PreflightGates(t *testing.T) {
	if routerUnitSuiteName == "" {
		t.Fatal("[BLOCK 4] suite identifier must not be empty")
	}
	common.AssertAuditPass(t, engine.PreflightRouter(), "")
	if !engine.AuditRouter() {
		t.Fatal("[BLOCK 4] AuditRouter returned false")
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Physical Source Anchor Verification
// ------------------------------------------------------------------------------

func TestRouterUnit_Block4_SourceAnchors(t *testing.T) {
	common.AssertRepoAnchors(t, "engine/router.go")
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Test Suite Epilogue & Clean Readiness Handshake )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.engine.router
// Mission:       Confirm Router and harness epilogue readiness post-suite.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Epilogue Readiness Verification
// ------------------------------------------------------------------------------

func TestRouterUnit_Block5_EpilogueReadiness(t *testing.T) {
	if !engine.IsRouterReady() || !common.IsHarnessReady() {
		t.Fatal("[BLOCK 5: EPILOGUE] router unit readiness handshake failed")
	}
}

// End of tests/unit/engine/router_test.go
