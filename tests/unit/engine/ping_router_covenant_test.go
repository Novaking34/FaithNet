// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/unit/engine/ping_router_covenant_test.go
// @id            test.unit.engine.ping_router_covenant
// @layer         Layer 4: Tier-2 Unit Verification Suite (Engine Sub-Stratum)
// @archetype     operational | stateless
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Layer 1 Ping, Router & Genesis 3 Airlock Unit Suite
// @role          Tier-2 Unit Diagnostic Sensor (4-Section Micro Protocol)
// @maturity      Rev 5 Baseline | 4x4 Verification Matrix
// @chassis       zero-heap | Verifies ALLOC_ZERO_HEAP on PingEngine hot paths
// @scale         O(1) deterministic unit tests using atomic MockClock
// @seams         in: data, engine, engine/covenant, tests/common
// @invariants    1. Macro Tier 2 (Unit): fine-grained isolation of Layer 1 cogs.
//                2. Micro 4-Section Protocol: Always Pass, Never Fail,
//                   Invariants & Limits, Emergence.
//                3. BuildPingInto, ReflectPingInto, ProcessPong == 0 allocs/op.
//                4. Ring wrap-around (64) & Peer table ceiling (128) enforced.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Tier-2 unit diagnostic sensor validating FaithNet's Layer 1 functional
//   cogs in fine-grained isolation: PingEngine RTT math & 64-slot ring wrap,
//   bounded 128-peer Router multi-hop resolution, and Genesis 3 CovenantFilter
//   T4 airlock verdicts.
//
//   ### For Architects, Operators & Non-Programmers:
//   Verifies each core gear individually—confirming the heartbeat stopwatch is
//   exact, the neighborhood route book handles relays cleanly without memory
//   overflow, and the airlock blocks hostile web noise every single time.
//
//   ### For Systems Programmers & Engine Architects:
//   Uses atomic MockClock for zero-flake nanosecond assertions, probes ring
//   overwrite boundaries past MaxInFlightPings (64), and proves 0 allocs/op.
// ==============================================================================

package engine_unit_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Verification Imports & Tier-2 Unit Sensor Anchor )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.engine.ping_router_covenant
// Mission:       Bind Layer 0/1 cogs and shared fixtures for isolated probes.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"bytes"
	"strings"
	"testing"
)

// --- Tier 2: External Library Dependencies (None) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---
import (
	"github.com/CreativeWorkzStudio/FaithNet/data"
	"github.com/CreativeWorkzStudio/FaithNet/engine"
	"github.com/CreativeWorkzStudio/FaithNet/engine/covenant"
	"github.com/CreativeWorkzStudio/FaithNet/tests/common"
)

// ==============================================================================
// BLOCK 3: BODY BLOCK
// ( Four-Section Micro Testing Protocol: Pass / Never Fail / Invariants / Emerge )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.engine.ping_router_covenant
// Mission:       Execute Section 1 (Always Pass), Section 2 (Never Fail),
//                Section 3 (Invariants & Limits), and Section 4 (Emergence).
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: SECTION 1 — ALWAYS PASS (Nominal Golden Paths)
// ------------------------------------------------------------------------------

func TestSection1_AlwaysPass_PingRouterAndAirlockNominal(t *testing.T) {
	// 1. PingEngine nominal RTT sequence.
	clk := engine.NewMockClock(1_000_000)
	nodeA := engine.NewPingEngine(common.NodeAlpha, clk)
	var req, resp data.Frame

	latencies := []int64{100, 300, 200}
	for _, delta := range latencies {
		echo := []byte("HELLO-CPI-SI")
		if err := nodeA.BuildPingInto(&req, common.NodeBeta, data.Rung2Host, echo); err != nil {
			t.Fatalf("BuildPingInto failed: %v", err)
		}
		clk.Advance(delta)
		if err := engine.ReflectPingInto(&req, &resp, common.NodeBeta); err != nil {
			t.Fatalf("ReflectPingInto failed: %v", err)
		}
		rtt, matched := nodeA.ProcessPong(&resp)
		if !matched || rtt != delta {
			t.Fatalf("expected RTT=%d matched=true, got rtt=%d matched=%v", delta, rtt, matched)
		}
		if !bytes.Equal(resp.PayloadSlice()[8:], echo) {
			t.Fatalf("echo payload mismatch: got %q", resp.PayloadSlice()[8:])
		}
	}

	h := nodeA.HealthSnapshot()
	if h.PingsSent != 3 || h.PongsRecv != 3 ||
		h.MinRTTNano != 100 || h.MaxRTTNano != 300 || h.AvgRTTNano != 200 {
		t.Fatalf("unexpected LinkHealth: %+v", h)
	}

	// 2. Router direct + multi-hop ("Across the Street") nominal lookup.
	rt := engine.NewRouter(common.NodeAlpha)
	if err := rt.UpsertRoute(engine.PeerRoute{
		PeerID:   common.NodeBeta,
		Rung:     data.Rung3Local,
		HopCount: 1,
	}); err != nil {
		t.Fatalf("direct upsert failed: %v", err)
	}
	if err := rt.UpsertRoute(engine.PeerRoute{
		PeerID:    common.NodeRelay,
		NextHopID: common.NodeBeta,
		Rung:      data.Rung3Local,
		HopCount:  2,
	}); err != nil {
		t.Fatalf("relay upsert failed: %v", err)
	}
	got, ok := rt.Lookup(common.NodeRelay)
	if !ok || got.NextHopID != common.NodeBeta || got.HopCount != 2 {
		t.Fatalf("unexpected multi-hop lookup: %+v", got)
	}

	// 3. CovenantFilter nominal sovereign (+1) and clean external (0:AWAIT).
	airlock := covenant.NewFilter()
	sov := common.MakeSovereignFrame(
		common.NodeAlpha, common.NodeBeta, data.Rung3Local, data.OpCPISync, 1,
		[]byte("Sovereign local CPI-SI truth delta"),
	)
	if v := airlock.EvaluateFrame(&sov); v.State != data.QuadPositive || !v.IsPermitted() {
		t.Fatalf("expected +1:ALIGNED for sovereign frame, got %+v", v)
	}
	vExt := airlock.EvaluateExternalPayload([]byte("NOAA Weather: 68F Clear"))
	if vExt.State != data.QuadAwait || !vExt.IsPermitted() {
		t.Fatalf("expected 0:AWAIT for clean external data, got %+v", vExt)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: SECTION 2 — NEVER FAIL (Fault Inversion & Adversarial Input)
// ------------------------------------------------------------------------------

func TestSection2_NeverFail_FaultyPingsRoutesAndHostileAirlock(t *testing.T) {
	clk := engine.NewMockClock(500)
	pe := engine.NewPingEngine(common.NodeAlpha, clk)
	var req, resp data.Frame

	// Nil / invalid targets / oversized echo payloads fail gracefully.
	if err := pe.BuildPingInto(nil, common.NodeBeta, data.Rung2Host, nil); err != engine.ErrPingNilFrame {
		t.Fatalf("expected ErrPingNilFrame, got %v", err)
	}
	if err := pe.BuildPingInto(&req, data.NodeIDNull, data.Rung2Host, nil); err != engine.ErrPingInvalidTarget {
		t.Fatalf("expected ErrPingInvalidTarget, got %v", err)
	}
	tooBig := make([]byte, engine.MaxPingEchoBytes+1)
	if err := pe.BuildPingInto(&req, common.NodeBeta, data.Rung2Host, tooBig); err != engine.ErrPingEchoTooLarge {
		t.Fatalf("expected ErrPingEchoTooLarge, got %v", err)
	}

	// ReflectPingInto rejects non-OpPing frames.
	req.Opcode = data.OpData
	if err := engine.ReflectPingInto(&req, &resp, common.NodeBeta); err != engine.ErrPingNotRequest {
		t.Fatalf("expected ErrPingNotRequest, got %v", err)
	}

	// Duplicate or foreign-target OpPong rejected cleanly.
	if err := pe.BuildPingInto(&req, common.NodeBeta, data.Rung2Host, nil); err != nil {
		t.Fatalf("setup ping failed: %v", err)
	}
	_ = engine.ReflectPingInto(&req, &resp, common.NodeBeta)
	if _, ok := pe.ProcessPong(&resp); !ok {
		t.Fatal("first valid pong should match")
	}
	if _, dup := pe.ProcessPong(&resp); dup {
		t.Fatal("duplicate pong must be rejected")
	}

	// Router rejects self-route, invalid rung, and zero hop count.
	rt := engine.NewRouter(common.NodeAlpha)
	if err := rt.UpsertRoute(engine.PeerRoute{
		PeerID: common.NodeAlpha, Rung: data.Rung2Host, HopCount: 1,
	}); err != engine.ErrRouteInvalidID {
		t.Fatalf("expected ErrRouteInvalidID for self-route, got %v", err)
	}
	if err := rt.UpsertRoute(engine.PeerRoute{
		PeerID: common.NodeBeta, Rung: data.Rung2Host, HopCount: 0,
	}); err != engine.ErrRouteInvalidHop {
		t.Fatalf("expected ErrRouteInvalidHop for 0 hops, got %v", err)
	}

	// Genesis 3 Airlock intercepts forged claims, XSS, SQLi, null-byte smuggling & nil.
	airlock := covenant.NewFilter()
	var forged data.Frame
	forged.Rung = data.Rung4Global
	forged.Opcode = data.OpLegacyWrap
	forged.Flags = data.FlagExternalOrigin | data.FlagCovenantVerified
	forged.SourceID = common.NodeExternal
	forged.SetPayload([]byte("Pretending to be verified"))
	if v := airlock.EvaluateFrame(&forged); v.State != data.QuadNegative ||
		v.Reason != covenant.ReasonForgedCovenantClaim {
		t.Fatalf("forged covenant flag was not rejected: %+v", v)
	}

	hostiles := [][]byte{
		[]byte("news <ScRiPt>alert(1)</script>"),
		[]byte("1; DROP TABLE users;"),
		[]byte("safe\x00smuggled"),
		bytes.Repeat([]byte{'X'}, data.FloorMaxPayload+1),
	}
	for _, raw := range hostiles {
		if v := airlock.EvaluateExternalPayload(raw); v.State != data.QuadNegative || !v.IsBlocked() {
			t.Fatalf("hostile payload %q was not blocked: %+v", raw, v)
		}
	}
	if v := airlock.EvaluateFrame(nil); v.State != data.QuadNull || !v.IsBlocked() {
		t.Fatalf("nil frame must return NULL:HALT, got %+v", v)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: SECTION 3 — INVARIANTS & LIMITS (Ring Wrap, 128 Cap, 0 Alloc)
// ------------------------------------------------------------------------------

func TestSection3_Invariants_RingWrapCapacityAndZeroAlloc(t *testing.T) {
	// 1. PingEngine 64-slot ring wrap-around boundary proof.
	clk := engine.NewMockClock(100)
	pe := engine.NewPingEngine(common.NodeAlpha, clk)
	var firstReq, latestReq, resp data.Frame

	if err := pe.BuildPingInto(&firstReq, common.NodeBeta, data.Rung2Host, nil); err != nil {
		t.Fatalf("initial ping failed: %v", err)
	}
	// Emit MaxInFlightPings additional pings so slot 1 wraps around cleanly.
	for i := 0; i < engine.MaxInFlightPings; i++ {
		clk.Advance(5)
		if err := pe.BuildPingInto(&latestReq, common.NodeBeta, data.Rung2Host, nil); err != nil {
			t.Fatalf("ring wrap ping %d failed: %v", i, err)
		}
	}
	// Stale seq=1 was overwritten by seq=65; reflecting seq=1 must NOT falsely match.
	_ = engine.ReflectPingInto(&firstReq, &resp, common.NodeBeta)
	if _, matched := pe.ProcessPong(&resp); matched {
		t.Fatal("overwritten stale seq=1 erroneously matched after ring wrap")
	}
	// Latest seq=65 MUST match cleanly.
	clk.Advance(20)
	_ = engine.ReflectPingInto(&latestReq, &resp, common.NodeBeta)
	if rtt, matched := pe.ProcessPong(&resp); !matched || rtt != 20 {
		t.Fatalf("expected latest wrapped seq match RTT=20, got matched=%v rtt=%d", matched, rtt)
	}

	// 2. Router hard 128-peer capacity ceiling proof.
	rt := engine.NewRouter(common.NodeAlpha)
	for i := 0; i < engine.MaxPeers; i++ {
		if err := rt.UpsertRoute(engine.PeerRoute{
			PeerID:   data.NodeID(0x5000 + i),
			Rung:     data.Rung2Host,
			HopCount: 1,
		}); err != nil {
			t.Fatalf("route fill %d failed: %v", i, err)
		}
	}
	if err := rt.UpsertRoute(engine.PeerRoute{
		PeerID: 0x9999, Rung: data.Rung2Host, HopCount: 1,
	}); err != engine.ErrRouteTableFull {
		t.Fatalf("expected ErrRouteTableFull at cap %d, got %v", engine.MaxPeers, err)
	}

	// 3. ALLOC_ZERO_HEAP proof across BuildPingInto + ReflectPingInto + ProcessPong.
	pe.Reset()
	echo := []byte("ZERO-ALLOC-SONAR")
	allocs := testing.AllocsPerRun(500, func() {
		_ = pe.BuildPingInto(&latestReq, common.NodeBeta, data.Rung1Floor, echo)
		clk.Advance(10)
		_ = engine.ReflectPingInto(&latestReq, &resp, common.NodeBeta)
		_, _ = pe.ProcessPong(&resp)
	})
	if allocs != 0 {
		t.Fatalf("ALLOC_ZERO_HEAP violated in PingEngine loop: %.2f allocs/op", allocs)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.4: SECTION 4 — EMERGENCE (Multi-Hop Discovery -> Airlock -> Dump)
// ------------------------------------------------------------------------------

func TestSection4_Emergence_NeighborhoodRelayAndDiagnosticCards(t *testing.T) {
	clk := engine.NewMockClock(10_000)
	pe := engine.NewPingEngine(common.NodeAlpha, clk)
	rt := engine.NewRouter(common.NodeAlpha)
	airlock := covenant.NewFilter()

	// NodeAlpha pings NodeRelay via NodeBeta (across-the-street 2-hop relay).
	_ = rt.UpsertRoute(engine.PeerRoute{
		PeerID:    common.NodeRelay,
		NextHopID: common.NodeBeta,
		Rung:      data.Rung3Local,
		HopCount:  2,
		LastSeenNs: clk.NowNano(),
	})

	route, ok := rt.Lookup(common.NodeRelay)
	if !ok {
		t.Fatal("failed resolving across-the-street route")
	}

	var req, resp data.Frame
	_ = pe.BuildPingInto(&req, route.PeerID, route.Rung, []byte("ACROSS-THE-STREET"))
	if v := airlock.EvaluateFrame(&req); v.State != data.QuadPositive {
		t.Fatalf("sovereign relay ping rejected: %+v", v)
	}
	clk.Advance(1200)
	_ = engine.ReflectPingInto(&req, &resp, common.NodeRelay)
	rtt, matched := pe.ProcessPong(&resp)
	if !matched || rtt != 1200 {
		t.Fatalf("expected 1200ns relay RTT, got matched=%v rtt=%d", matched, rtt)
	}

	// Update route telemetry with newly measured RTT and verify diagnostic cards.
	route.LastRTTNano = rtt
	route.LastSeenNs = clk.NowNano()
	_ = rt.UpsertRoute(route)

	pingCard := pe.ExecutePingDumpCard()
	routerCard := rt.ExecuteRouterDumpCard()
	if !strings.Contains(pingCard, "1 sent / 1 pongs") ||
		!strings.Contains(routerCard, "1 active / 128 max capacity") {
		t.Fatalf("unexpected diagnostic cards:\n%s\n%s", pingCard, routerCard)
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Unit Stratum Preflight Cross-Check )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.engine.ping_router_covenant
// Mission:       Confirm Reset() teardown leaves zero stale state behind.
// ==============================================================================

func TestAudit_LifecycleResetCleanliness(t *testing.T) {
	clk := engine.NewMockClock(42)
	pe := engine.NewPingEngine(common.NodeAlpha, clk)
	var req data.Frame
	_ = pe.BuildPingInto(&req, common.NodeBeta, data.Rung2Host, nil)
	pe.Reset()
	clk.Reset()
	if clk.NowNano() != 0 || pe.HealthSnapshot().PingsSent != 0 {
		t.Fatal("lifecycle Reset() left stale state behind")
	}
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK / STATUTORY EXEMPTION & LIFECYCLE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.engine.ping_router_covenant
// Mission:       Test harness managed by Go testing runner.
// Stateless Exemption Notice: None. (Stateless test suite).
// ==============================================================================

// End of tests/unit/engine/ping_router_covenant_test.go
