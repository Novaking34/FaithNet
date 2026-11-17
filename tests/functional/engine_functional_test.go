// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/functional/engine_functional_test.go
// @id            test.functional.engine
// @layer         Layer 4: Tier-1 Functional Verification Suite
// @archetype     verification_suite | functional_surface
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Layer 1 Engine Functional Surface & Zero-Alloc Suite
// @role          Tier-1 Functional Diagnostic Sensor (4-Section Micro Protocol)
// @maturity      Rev 5 Baseline | 4x4 Verification Matrix
// @chassis       zero-heap | Verifies ALLOC_ZERO_HEAP via testing.AllocsPerRun
// @scale         O(1) deterministic functional tests & hot-path benchmarks
// @seams         in: data, engine, tests/common | probes: Preflight*
// @invariants    1. Macro Tier 1 (Functional): exercises Codec, Clock, Ping,
//                   and Router public API surfaces independently and together.
//                2. Micro 4-Section Protocol: Always Pass, Never Fail,
//                   Invariants & Limits, Emergence.
//                3. EncodeInto, DecodeFrom, and PingEngine loop == 0 allocs/op.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Tier-1 Functional surface suite validating FaithNet's Layer 1 engine cogs
//   (`engine/codec.go`, `engine/clock.go`, `engine/ping.go`, `engine/router.go`)
//   across nominal traces, fault inversion, zero-heap invariants, full-wire
//   heartbeat emergence, Block 4 preflight gates, and throughput benchmarks.
//
//   ### For Architects, Operators & Non-Programmers:
//   Exercises the entire core heartbeat engine from the outside—confirming
//   envelopes pack cleanly, stopwatches run offline, routes resolve neighbors,
//   and full ping echoes complete with zero wasted memory.
//
//   ### For Systems Programmers & Engine Architects:
//   Paired with fine-grained Tier 2 unit suites under `tests/unit/engine/`.
// ==============================================================================

package functional_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Identifier, Verification Imports & Tier-1 Sensor Anchor )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-1 Functional Verification Suite
// Module:        test.functional.engine
// Mission:       Bind Layer 0/1 cogs and shared harness for surface sweeps.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"strings"
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
	engineFunctionalSuiteName = "Tier 1: Layer 1 Engine Functional Surface"
)

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-1 Functional Verification Suite
// Module:        test.functional.engine
// Mission:       Exercise all Layer 1 surface APIs across Sections 1-4.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: [ SECTION 1: ALWAYS PASS (NOMINAL GOLDEN SURFACE COGS) ]
// ------------------------------------------------------------------------------

func TestEngineFunctional_Sec1_CodecRoundTripAllRungs(t *testing.T) {
	rungs := []data.RungLevel{
		data.Rung1Floor,
		data.Rung2Host,
		data.Rung3Local,
		data.Rung4Global,
	}

	var wire [data.FloorMaxFrameSize]byte
	for idx, rung := range rungs {
		msg := []byte("MillenniumOS CPI-SI Sovereign Sync Payload")
		in := common.MakeSovereignFrame(
			common.NodeAlpha, common.NodeBeta, rung, data.OpCPISync,
			uint32(100+idx), msg,
		)
		n, err := engine.EncodeInto(wire[:], &in)
		if err != nil {
			t.Fatalf("[SEC 1] rung %s encode failed: %v", rung, err)
		}
		if n != data.FrameHeaderSize+len(msg) {
			t.Fatalf("[SEC 1] want %d wire bytes, got %d", data.FrameHeaderSize+len(msg), n)
		}
		var out data.Frame
		if err := engine.DecodeFrom(wire[:n], &out); err != nil {
			t.Fatalf("[SEC 1] rung %s decode failed: %v", rung, err)
		}
		common.AssertFramesEqual(t, &in, &out)
		if out.Checksum != in.Checksum {
			t.Fatalf("[SEC 1] checksum mismatch: want 0x%08X got 0x%08X", in.Checksum, out.Checksum)
		}
	}
}

func TestEngineFunctional_Sec1_ClockPingAndRouterSurface(t *testing.T) {
	clk := engine.NewMockClock(1000)
	pe := engine.NewPingEngine(common.NodeAlpha, clk)
	rt := engine.NewRouter(common.NodeAlpha)

	if err := rt.UpsertRoute(engine.PeerRoute{
		PeerID: common.NodeBeta, Rung: data.Rung2Host, HopCount: 1,
	}); err != nil {
		t.Fatalf("[SEC 1] router upsert failed: %v", err)
	}
	if _, ok := rt.Lookup(common.NodeBeta); !ok {
		t.Fatal("[SEC 1] router lookup missed inserted peer")
	}

	var req, resp data.Frame
	if err := pe.BuildPingInto(&req, common.NodeBeta, data.Rung2Host, []byte("SONAR")); err != nil {
		t.Fatalf("[SEC 1] BuildPingInto failed: %v", err)
	}
	clk.Advance(250)
	if err := engine.ReflectPingInto(&req, &resp, common.NodeBeta); err != nil {
		t.Fatalf("[SEC 1] ReflectPingInto failed: %v", err)
	}
	if rtt, ok := pe.ProcessPong(&resp); !ok || rtt != 250 {
		t.Fatalf("[SEC 1] expected RTT=250, got ok=%v rtt=%d", ok, rtt)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: [ SECTION 2: NEVER FAIL (FAULT INVERSION & DEGRADATION) ]
// ------------------------------------------------------------------------------

func TestEngineFunctional_Sec2_CodecAndPingFaultInversion(t *testing.T) {
	var wire [data.FloorMaxFrameSize]byte
	var out data.Frame

	if _, err := engine.EncodeInto(wire[:], nil); err != engine.ErrBufferTooSmall {
		t.Fatalf("[SEC 2] expected ErrBufferTooSmall on nil frame, got %v", err)
	}
	if err := engine.DecodeFrom(wire[:10], &out); err != engine.ErrBufferTooSmall {
		t.Fatalf("[SEC 2] expected ErrBufferTooSmall on short wire, got %v", err)
	}
	if err := engine.DecodeFrom(wire[:], nil); err != engine.ErrBufferTooSmall {
		t.Fatalf("[SEC 2] expected ErrBufferTooSmall on nil target, got %v", err)
	}

	pe := engine.NewPingEngine(common.NodeAlpha, engine.NewMockClock(0))
	if _, ok := pe.ProcessPong(nil); ok {
		t.Fatal("[SEC 2] ProcessPong(nil) must return false")
	}
	rt := engine.NewRouter(common.NodeAlpha)
	if err := rt.UpsertRoute(engine.PeerRoute{
		PeerID: common.NodeAlpha, Rung: data.Rung2Host, HopCount: 1,
	}); err != engine.ErrRouteInvalidID {
		t.Fatalf("[SEC 2] expected self-route rejection, got %v", err)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: [ SECTION 3: INVARIANTS, SATURATION & LIMIT GUARDS ]
// ------------------------------------------------------------------------------

func TestEngineFunctional_Sec3_ZeroHeapHotPathProof(t *testing.T) {
	var wire [data.FloorMaxFrameSize]byte
	var out data.Frame
	in := common.MakeSovereignFrame(
		common.NodeAlpha, common.NodeBeta, data.Rung2Host, data.OpPing, 1,
		[]byte("ZERO-HEAP-PROOF"),
	)

	encAllocs := testing.AllocsPerRun(1000, func() {
		_, _ = engine.EncodeInto(wire[:], &in)
	})
	if encAllocs != 0 {
		t.Fatalf("[SEC 3] ALLOC_ZERO_HEAP violated in EncodeInto: %.2f allocs/op", encAllocs)
	}

	n, _ := engine.EncodeInto(wire[:], &in)
	decAllocs := testing.AllocsPerRun(1000, func() {
		_ = engine.DecodeFrom(wire[:n], &out)
	})
	if decAllocs != 0 {
		t.Fatalf("[SEC 3] ALLOC_ZERO_HEAP violated in DecodeFrom: %.2f allocs/op", decAllocs)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.4: [ SECTION 4: EMERGENCE & COMPOSITE WORKFLOWS ]
// ------------------------------------------------------------------------------

func TestEngineFunctional_Sec4_FullWireHeartbeatConversation(t *testing.T) {
	clk := engine.NewMockClock(50_000)
	nodeA := engine.NewPingEngine(common.NodeAlpha, clk)
	airlock := covenant.NewFilter()

	var req, wireIn, resp, wireOut data.Frame
	var bufA, bufB [data.FloorMaxFrameSize]byte

	if err := nodeA.BuildPingInto(&req, common.NodeBeta, data.Rung1Floor,
		[]byte("ROOM-PRESENCE")); err != nil {
		t.Fatalf("[SEC 4] BuildPingInto failed: %v", err)
	}
	nA, err := engine.EncodeInto(bufA[:], &req)
	if err != nil {
		t.Fatalf("[SEC 4] EncodeInto A->B failed: %v", err)
	}

	clk.Advance(750)
	if err := engine.DecodeFrom(bufA[:nA], &wireIn); err != nil {
		t.Fatalf("[SEC 4] DecodeFrom at B failed: %v", err)
	}
	if v := airlock.EvaluateFrame(&wireIn); v.State != data.QuadPositive {
		t.Fatalf("[SEC 4] sovereign ping rejected by airlock: %+v", v)
	}
	if err := engine.ReflectPingInto(&wireIn, &resp, common.NodeBeta); err != nil {
		t.Fatalf("[SEC 4] ReflectPingInto failed: %v", err)
	}
	nB, err := engine.EncodeInto(bufB[:], &resp)
	if err != nil {
		t.Fatalf("[SEC 4] EncodeInto B->A failed: %v", err)
	}

	clk.Advance(750)
	if err := engine.DecodeFrom(bufB[:nB], &wireOut); err != nil {
		t.Fatalf("[SEC 4] DecodeFrom at A failed: %v", err)
	}
	rtt, ok := nodeA.ProcessPong(&wireOut)
	if !ok || rtt != 1500 {
		t.Fatalf("[SEC 4] expected RTT=1500ns, got ok=%v rtt=%d", ok, rtt)
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Preflight Gates, 60-Col Dump Cards, Anchors & Zero-Alloc Benchmarks )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-1 Functional Verification Suite
// Module:        test.functional.engine
// Mission:       Assert structured Block 4 preflights for every Layer 1 cog
//                and benchmark ALLOC_ZERO_HEAP wire throughput.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Preflight Diagnostic Gates & 60-Col Cards
// ------------------------------------------------------------------------------

func TestEngineFunctional_Block4_PreflightAllLayer1Cogs(t *testing.T) {
	common.AssertAuditPass(t, engine.PreflightCodec(), engine.ExecuteCodecDumpCard())
	common.AssertAuditPass(t, engine.PreflightClock(),
		engine.ExecuteClockDumpCard(engine.NewMockClock(42)))
	common.AssertAuditPass(t, engine.PreflightPingEngine(),
		engine.NewPingEngine(common.NodeAlpha, engine.NewMockClock(0)).ExecutePingDumpCard())
	common.AssertAuditPass(t, engine.PreflightRouter(),
		engine.NewRouter(common.NodeAlpha).ExecuteRouterDumpCard())

	codecCard := engine.ExecuteCodecDumpCard()
	if !strings.Contains(codecCard, "PASS") || !strings.Contains(codecCard, "512 bytes") {
		t.Fatalf("[BLOCK 4] unexpected codec dump card:\n%s", codecCard)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Source Anchors & Zero-Allocation Benchmarks
// ------------------------------------------------------------------------------

func TestEngineFunctional_Block4_SourceAnchors(t *testing.T) {
	if engineFunctionalSuiteName == "" {
		t.Fatal("[BLOCK 4] suite name must not be empty")
	}
	common.AssertRepoAnchors(t,
		"engine/codec.go",
		"engine/clock.go",
		"engine/ping.go",
		"engine/router.go",
	)
}

func BenchmarkEncodeInto(b *testing.B) {
	in := common.MakeSovereignFrame(
		common.NodeAlpha, common.NodeBeta, data.Rung2Host, data.OpPing, 1,
		[]byte("FAITHNET-BENCH"),
	)
	var wire [data.FloorMaxFrameSize]byte

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.EncodeInto(wire[:], &in)
	}
}

func BenchmarkDecodeFrom(b *testing.B) {
	in := common.MakeSovereignFrame(
		common.NodeAlpha, common.NodeBeta, data.Rung2Host, data.OpPing, 1,
		[]byte("FAITHNET-BENCH"),
	)
	var out data.Frame
	var wire [data.FloorMaxFrameSize]byte
	n, _ := engine.EncodeInto(wire[:], &in)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = engine.DecodeFrom(wire[:n], &out)
	}
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Test Suite Epilogue & Clean Readiness Handshake )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-1 Functional Verification Suite
// Module:        test.functional.engine
// Mission:       Confirm every Layer 1 engine cog reports epilogue readiness.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Epilogue & Clean State Handshake
// ------------------------------------------------------------------------------

func TestEngineFunctional_Block5_EpilogueReadiness(t *testing.T) {
	if !engine.IsCodecReady() || !engine.IsClockReady() ||
		!engine.IsPingEngineReady() || !engine.IsRouterReady() ||
		!common.IsHarnessReady() {
		t.Fatal("[BLOCK 5: EPILOGUE] one or more Layer 1 readiness probes failed")
	}
}

// End of tests/functional/engine_functional_test.go
