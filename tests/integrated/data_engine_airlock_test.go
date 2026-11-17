// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/integrated/data_engine_airlock_test.go
// @id            test.integrated.data_engine_airlock
// @layer         Layer 4: Tier-3 Integrated Verification Suite
// @archetype     verification_suite | integrated_core_pipeline
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Layer 0 + Layer 1 + Genesis 3 Airlock Integration Suite
// @role          Tier-3 Integrated Sensor: Data + Codec + Clock + Ping + Router
// @maturity      Rev 5 Baseline | Wave 6 4x4 Symmetry Expansion
// @chassis       bounded-heap | Pure Layer 0 + Layer 1 zero-OS multi-hop flow
// @scale         O(1) deterministic multi-hop relay & airlock verification
// @seams         in: data, engine, engine/covenant, tests/common
// @invariants    1. Macro Tier 3 (Integrated): combines all Layer 0/1 cogs
//                   without any OS sockets (`net`/`os` remain zero-imported).
//                2. Proves 3-node multi-hop relay ("Across the Street") with
//                   deterministic MockClock RTT and Airlock vetting at every hop.
//                3. Zero corrupted frames or forged covenant claims survive.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Tier-3 Integrated verification suite wiring Layer 0 (`data/`) directly to
//   every Layer 1 operational cog (`Codec`, `MockClock`, `PingEngine`,
//   `Router`, and `covenant.Filter`) to prove end-to-end pure-engine message
//   flow, 3-node multi-hop routing, and strict Genesis 3 Airlock defense.
//
//   ### For Architects, Operators & Non-Programmers:
//   Proves the core brain of FaithNet—its message envelopes, packer, timer,
//   ping heartbeat, neighbor map, and security door—works smoothly together
//   across a 3-node chain before touching any network hardware.
//
//   ### For Systems Programmers & Engine Architects:
//   Exercises `EncodeInto` -> `DecodeFrom` -> `Filter.EvaluateFrame` ->
//   `Router.UpsertRoute`/`Lookup` -> `ReflectPingInto` -> `ProcessPong` in a
//   deterministic 3-node pipeline with zero heap churn on hot paths.
// ==============================================================================

package integrated_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Identifier, Verification Imports & Tier-3 Core Pipeline Anchor )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-3 Integrated Verification Suite
// Module:        test.integrated.data_engine_airlock
// Mission:       Bind Layer 0 data contracts and all Layer 1 engine/covenant cogs.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"testing"
)

// --- Tier 2: External Library Dependencies (None — 100% Go Stdlib Invariant) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---
import (
	"github.com/CreativeWorkzStudio/FaithNet/data"
	"github.com/CreativeWorkzStudio/FaithNet/engine"
	"github.com/CreativeWorkzStudio/FaithNet/engine/covenant"
	"github.com/CreativeWorkzStudio/FaithNet/tests/common"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Operational Suite Constants & Identifiers
// ------------------------------------------------------------------------------

const (
	corePipelineSuiteName = "Tier 3 Integrated: Layer 0 + Engine + Genesis 3 Airlock"
)

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-3 Integrated Verification Suite
// Module:        test.integrated.data_engine_airlock
// Mission:       Prove pure Layer 0+1 pipeline, airlock rejection, invariants,
//                and 3-node multi-hop ("Across the Street") relay emergence.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: [ SECTION 1: ALWAYS PASS (CORE CODEC + PING + AIRLOCK) ]
// ------------------------------------------------------------------------------

func TestDataEngineAirlock_Sec1_SovereignEncodeAirlockReflectPongFlow(t *testing.T) {
	clk := engine.NewMockClock(10_000)
	pingA := engine.NewPingEngine(common.NodeAlpha, clk)
	routerA := engine.NewRouter(common.NodeAlpha)
	routerB := engine.NewRouter(common.NodeBeta)
	airlock := covenant.NewFilter()

	var req, rxReq, resp, rxResp data.Frame
	var wireA, wireB [data.FloorMaxFrameSize]byte

	if err := pingA.BuildPingInto(&req, common.NodeBeta, data.Rung1Floor, []byte("CORE-SYNC")); err != nil {
		t.Fatalf("[SEC 1] BuildPingInto failed: %v", err)
	}
	nA, err := engine.EncodeInto(wireA[:], &req)
	if err != nil {
		t.Fatalf("[SEC 1] EncodeInto req failed: %v", err)
	}
	if err := engine.DecodeFrom(wireA[:nA], &rxReq); err != nil {
		t.Fatalf("[SEC 1] DecodeFrom req failed: %v", err)
	}
	if v := airlock.EvaluateFrame(&rxReq); v.State != data.QuadPositive {
		t.Fatalf("[SEC 1] Airlock rejected sovereign ping: %+v", v)
	}
	if err := routerB.UpsertRoute(engine.PeerRoute{
		PeerID:     rxReq.SourceID,
		NextHopID:  rxReq.SourceID,
		Rung:       rxReq.Rung,
		HopCount:   1,
		LastSeenNs: clk.NowNano(),
	}); err != nil {
		t.Fatalf("[SEC 1] routerB UpsertRoute failed: %v", err)
	}

	clk.Advance(750)
	if err := engine.ReflectPingInto(&rxReq, &resp, common.NodeBeta); err != nil {
		t.Fatalf("[SEC 1] ReflectPingInto failed: %v", err)
	}
	nB, err := engine.EncodeInto(wireB[:], &resp)
	if err != nil {
		t.Fatalf("[SEC 1] EncodeInto resp failed: %v", err)
	}
	if err := engine.DecodeFrom(wireB[:nB], &rxResp); err != nil {
		t.Fatalf("[SEC 1] DecodeFrom resp failed: %v", err)
	}
	if v := airlock.EvaluateFrame(&rxResp); v.State != data.QuadPositive {
		t.Fatalf("[SEC 1] Airlock rejected sovereign pong: %+v", v)
	}

	rtt, ok := pingA.ProcessPong(&rxResp)
	if !ok || rtt != 750 {
		t.Fatalf("[SEC 1] expected exact RTT=750ns, got ok=%v rtt=%d", ok, rtt)
	}
	if err := routerA.UpsertRoute(engine.PeerRoute{
		PeerID:      rxResp.SourceID,
		NextHopID:   rxResp.SourceID,
		Rung:        rxResp.Rung,
		HopCount:    1,
		LastSeenNs:  clk.NowNano(),
		LastRTTNano: rtt,
	}); err != nil {
		t.Fatalf("[SEC 1] routerA UpsertRoute failed: %v", err)
	}

	route, found := routerA.Lookup(common.NodeBeta)
	if !found || route.LastRTTNano != 750 {
		t.Fatalf("[SEC 1] routerA missing learned route with RTT 750: found=%v route=%+v", found, route)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: [ SECTION 2: NEVER FAIL (CORRUPTION & FORGED CLAIMS) ]
// ------------------------------------------------------------------------------

func TestDataEngineAirlock_Sec2_CorruptWireAndForgedClaimsNeverEnterRouter(t *testing.T) {
	airlock := covenant.NewFilter()
	router := engine.NewRouter(common.NodeAlpha)

	// 1. Single-bit flip anywhere on encoded wire must trip CRC-32C before Airlock.
	valid := common.MakeSovereignFrame(
		common.NodeBeta, common.NodeAlpha, data.Rung2Host, data.OpData, 1, []byte("CLEAN"),
	)
	wire := common.EncodeFixture(t, &valid)
	wire[len(wire)-1] ^= 0x01
	var decoded data.Frame
	if err := engine.DecodeFrom(wire, &decoded); err == nil {
		t.Fatal("[SEC 2] tampered wire frame bypassed Codec CRC-32C integrity check")
	}

	// 2. External frame forging FlagCovenantVerified must be blocked by Airlock (-1).
	forged := data.Frame{
		Version:  data.ProtocolVersion,
		Rung:     data.Rung4Global,
		Opcode:   data.OpLegacyWrap,
		Flags:    data.FlagExternalOrigin | data.FlagCovenantVerified,
		Truth:    data.QuadPositive,
		SourceID: common.NodeExternal,
		TargetID: common.NodeAlpha,
		SeqID:    2,
	}
	forged.SetPayload([]byte("TRUST ME"))
	if v := airlock.EvaluateFrame(&forged); !v.IsBlocked() || v.Reason != covenant.ReasonForgedCovenantClaim {
		t.Fatalf("[SEC 2] forged covenant claim bypassed Airlock: %+v", v)
	}

	// 3. Hostile script payload stamped as external must be blocked (-1).
	hostile := data.Frame{
		Version:  data.ProtocolVersion,
		Rung:     data.Rung4Global,
		Opcode:   data.OpLegacyWrap,
		Flags:    data.FlagExternalOrigin,
		Truth:    data.QuadAwait,
		SourceID: common.NodeExternal,
		TargetID: common.NodeAlpha,
		SeqID:    3,
	}
	hostile.SetPayload([]byte("<script>steal()</script>"))
	if v := airlock.EvaluateFrame(&hostile); !v.IsBlocked() {
		t.Fatalf("[SEC 2] hostile external payload bypassed Airlock: %+v", v)
	}

	if router.ActiveCount() != 0 {
		t.Fatalf("[SEC 2] Router must remain empty after rejected frames, got %d", router.ActiveCount())
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: [ SECTION 3: INVARIANTS (QUAD-STATE & ALLOCATION DISCIPLINE) ]
// ------------------------------------------------------------------------------

func TestDataEngineAirlock_Sec3_ExternalFramesStayAwaitAndLiveAuditsStayClean(t *testing.T) {
	clk := engine.NewMockClock(500)
	pe := engine.NewPingEngine(common.NodeAlpha, clk)
	rt := engine.NewRouter(common.NodeAlpha)
	airlock := covenant.NewFilter()

	// Clean external fact must remain strictly 0:AWAIT (never silently elevated to +1).
	ext := data.Frame{
		Version:  data.ProtocolVersion,
		Rung:     data.Rung4Global,
		Opcode:   data.OpLegacyWrap,
		Flags:    data.FlagExternalOrigin,
		Truth:    data.QuadAwait,
		SourceID: common.NodeExternal,
		TargetID: common.NodeAlpha,
		SeqID:    10,
	}
	ext.SetPayload([]byte("Local weather advisory: clear skies"))
	wire := common.EncodeFixture(t, &ext)

	var rx data.Frame
	if err := engine.DecodeFrom(wire, &rx); err != nil {
		t.Fatalf("[SEC 3] DecodeFrom external frame failed: %v", err)
	}
	v := airlock.EvaluateFrame(&rx)
	if v.State != data.QuadAwait || v.IsBlocked() {
		t.Fatalf("[SEC 3] clean external fact must stay 0:AWAIT, got %+v", v)
	}

	// Verify all live instance Block 4 audits pass together.
	common.AssertAuditPass(t, pe.AuditState(), pe.ExecutePingDumpCard())
	common.AssertAuditPass(t, rt.AuditState(), rt.ExecuteRouterDumpCard())
	common.AssertAuditPass(t, airlock.AuditPolicy(), airlock.ExecuteAirlockDumpCard())
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.4: [ SECTION 4: EMERGENCE (3-NODE MULTI-HOP RELAY CHAIN) ]
// ------------------------------------------------------------------------------

func TestDataEngineAirlock_Sec4_ThreeNodeAcrossTheStreetMultiHopRelay(t *testing.T) {
	clk := engine.NewMockClock(1_000)
	pingAlpha := engine.NewPingEngine(common.NodeAlpha, clk)
	routerAlpha := engine.NewRouter(common.NodeAlpha)
	routerRelay := engine.NewRouter(common.NodeRelay)
	routerBeta := engine.NewRouter(common.NodeBeta)
	airlock := covenant.NewFilter()

	// Topology: Alpha <-> Relay <-> Beta (2 hops; Alpha cannot reach Beta directly).
	_ = routerAlpha.UpsertRoute(engine.PeerRoute{
		PeerID: common.NodeBeta, NextHopID: common.NodeRelay,
		Rung: data.Rung3Local, HopCount: 2, LastSeenNs: clk.NowNano(),
	})
	_ = routerRelay.UpsertRoute(engine.PeerRoute{
		PeerID: common.NodeBeta, NextHopID: common.NodeBeta,
		Rung: data.Rung1Floor, HopCount: 1, LastSeenNs: clk.NowNano(),
	})
	_ = routerRelay.UpsertRoute(engine.PeerRoute{
		PeerID: common.NodeAlpha, NextHopID: common.NodeAlpha,
		Rung: data.Rung3Local, HopCount: 1, LastSeenNs: clk.NowNano(),
	})
	_ = routerBeta.UpsertRoute(engine.PeerRoute{
		PeerID: common.NodeAlpha, NextHopID: common.NodeRelay,
		Rung: data.Rung1Floor, HopCount: 2, LastSeenNs: clk.NowNano(),
	})

	// Hop 1: Alpha -> Relay (Rung 3 Local), Hop 2: Relay -> Beta (Rung 1 Floor).
	var req, atRelay, atBeta, pong, pongAtRelay, pongAtAlpha data.Frame
	var buf [data.FloorMaxFrameSize]byte

	rToBeta, ok := routerAlpha.Lookup(common.NodeBeta)
	if !ok || rToBeta.NextHopID != common.NodeRelay {
		t.Fatalf("[SEC 4] Alpha expected NextHop=NodeRelay for NodeBeta, got %+v", rToBeta)
	}
	if err := pingAlpha.BuildPingInto(&req, common.NodeBeta, rToBeta.Rung, []byte("ACROSS-THE-STREET")); err != nil {
		t.Fatalf("[SEC 4] Alpha BuildPingInto failed: %v", err)
	}

	n, _ := engine.EncodeInto(buf[:], &req)
	_ = engine.DecodeFrom(buf[:n], &atRelay)
	if v := airlock.EvaluateFrame(&atRelay); v.State != data.QuadPositive {
		t.Fatalf("[SEC 4] Relay airlock rejected forwarded ping: %+v", v)
	}
	clk.Advance(400) // Hop 1 flight time.

	// Relay forwards to Beta over Rung 1 Floor (downgrading hop rung cleanly).
	rRelayToBeta, ok := routerRelay.Lookup(atRelay.TargetID)
	if !ok || rRelayToBeta.Rung != data.Rung1Floor {
		t.Fatalf("[SEC 4] Relay missing Rung 1 route to Beta: %+v", rRelayToBeta)
	}
	atRelay.Rung = rRelayToBeta.Rung
	n, _ = engine.EncodeInto(buf[:], &atRelay)
	_ = engine.DecodeFrom(buf[:n], &atBeta)
	if v := airlock.EvaluateFrame(&atBeta); v.State != data.QuadPositive {
		t.Fatalf("[SEC 4] Beta airlock rejected multi-hop ping: %+v", v)
	}
	clk.Advance(600) // Hop 2 flight time.

	// Beta reflects Pong back via Relay to Alpha.
	if err := engine.ReflectPingInto(&atBeta, &pong, common.NodeBeta); err != nil {
		t.Fatalf("[SEC 4] Beta ReflectPingInto failed: %v", err)
	}
	n, _ = engine.EncodeInto(buf[:], &pong)
	_ = engine.DecodeFrom(buf[:n], &pongAtRelay)
	clk.Advance(600) // Return Hop 2.

	rRelayToAlpha, _ := routerRelay.Lookup(pongAtRelay.TargetID)
	pongAtRelay.Rung = rRelayToAlpha.Rung
	n, _ = engine.EncodeInto(buf[:], &pongAtRelay)
	_ = engine.DecodeFrom(buf[:n], &pongAtAlpha)
	clk.Advance(400) // Return Hop 1.

	rtt, matched := pingAlpha.ProcessPong(&pongAtAlpha)
	if !matched || rtt != 2_000 {
		t.Fatalf("[SEC 4] expected 2-hop total RTT=2000ns, got matched=%v rtt=%d", matched, rtt)
	}
	if len(pongAtAlpha.PayloadSlice()) < 8 || string(pongAtAlpha.PayloadSlice()[8:]) != "ACROSS-THE-STREET" {
		t.Fatalf("[SEC 4] payload altered during multi-hop relay: %q", pongAtAlpha.PayloadSlice())
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Preflight Diagnostic Gates & Physical Source Anchor Verification )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-3 Integrated Verification Suite
// Module:        test.integrated.data_engine_airlock
// Mission:       Assert Layer 0/1 preflights and verify core source anchors.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Block 4 Pre-Flight Contract Gates
// ------------------------------------------------------------------------------

func TestDataEngineAirlock_Block4_PreflightGates(t *testing.T) {
	if corePipelineSuiteName == "" {
		t.Fatal("[BLOCK 4] suite identifier must not be empty")
	}
	common.AssertAuditPass(t, engine.AuditDataVocabulary(), "")
	common.AssertAuditPass(t, engine.PreflightCodec(), "")
	common.AssertAuditPass(t, engine.PreflightClock(), "")
	common.AssertAuditPass(t, engine.PreflightPingEngine(), "")
	common.AssertAuditPass(t, engine.PreflightRouter(), "")
	common.AssertAuditPass(t, covenant.PreflightCovenantFilter(), "")
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Physical Source Anchor Verification
// ------------------------------------------------------------------------------

func TestDataEngineAirlock_Block4_SourceAnchors(t *testing.T) {
	common.AssertRepoAnchors(t,
		"data/types.go",
		"data/frame.go",
		"engine/codec.go",
		"engine/clock.go",
		"engine/ping.go",
		"engine/router.go",
		"engine/covenant/filter.go",
	)
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Test Suite Epilogue & Clean Readiness Handshake )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-3 Integrated Verification Suite
// Module:        test.integrated.data_engine_airlock
// Mission:       Confirm Layer 0/1 readiness handshake post-suite.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Epilogue Readiness Verification
// ------------------------------------------------------------------------------

func TestDataEngineAirlock_Block5_EpilogueReadiness(t *testing.T) {
	if !engine.IsCodecReady() || !engine.IsClockReady() ||
		!engine.IsPingEngineReady() || !engine.IsRouterReady() ||
		!covenant.IsAirlockReady() || !common.IsHarnessReady() {
		t.Fatal("[BLOCK 5: EPILOGUE] Layer 0+1+Airlock integrated readiness failed")
	}
}

// End of tests/integrated/data_engine_airlock_test.go
