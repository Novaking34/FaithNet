// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/unit/engine/ping_test.go
// @id            test.unit.engine.ping
// @layer         Layer 4: Tier-2 Unit Verification Suite (Engine Sub-Stratum)
// @archetype     verification_suite | unit_isolation
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Layer 1 5-Cog Sonar Ping Engine Unit Suite
// @role          Tier-2 Unit Isolation Gate for cog.engine.ping
// @maturity      Rev 5 Baseline | 4x4 Verification Matrix
// @chassis       zero-heap | Verifies ALLOC_ZERO_HEAP on Build/Reflect/Process
// @scale         O(1) deterministic 64-slot ring & RTT statistics suite
// @seams         in: data, engine, tests/common | target: engine/ping.go
// @invariants    1. BuildPingInto, ReflectPingInto, ProcessPong == 0 allocs/op.
//                2. 64-slot ring wrap cleanly invalidates overwritten slots.
//                3. Nil clock defaults to hardware MonotonicClock safely.
//                4. Concurrent goroutine access is 100% -race clean.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Tier-2 Unit isolation suite strictly verifying Cog 5 (`engine/ping.go`):
//   deterministic min/avg/max RTT telemetry, every error sentinel, duplicate &
//   foreign-target rejection, 64-slot ring wrap, live `AuditState`, `-race`
//   concurrency, and zero-heap hot-path compliance.
//
//   ### For Architects, Operators & Non-Programmers:
//   Proves FaithNet's heartbeat sonar tracks round-trip latency accurately,
//   ignores duplicate or spoofed echoes, and never leaks memory over long runs.
//
//   ### For Systems Programmers & Engine Architects:
//   100% branch coverage of `BuildPingInto`, `ReflectPingInto`, `ProcessPong`,
//   `HealthSnapshot`, `ActiveInFlight`, `AuditState`, `Reset`, and dump card.
// ==============================================================================

package engine_unit_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Identifier, Verification Imports & Tier-2 Unit Sensor Anchor )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.engine.ping
// Mission:       Bind PingEngine, Clock, and shared fixtures for isolation.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"bytes"
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
	pingUnitSuiteName = "Tier 2 Unit: cog.engine.ping (5-Cog Sonar & Ring)"
)

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.engine.ping
// Mission:       Verify PingEngine RTT math, fault inversion, ring wrap, & race.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: [ SECTION 1: ALWAYS PASS (NOMINAL GOLDEN SURFACE COGS) ]
// ------------------------------------------------------------------------------

func TestPingUnit_Sec1_NominalRTTMinAvgMaxAndDefaultClock(t *testing.T) {
	clk := engine.NewMockClock(1_000_000)
	pe := engine.NewPingEngine(common.NodeAlpha, clk)
	var req, resp data.Frame

	latencies := []int64{100, 300, 200}
	for _, delta := range latencies {
		echo := []byte("HELLO-CPI-SI")
		if err := pe.BuildPingInto(&req, common.NodeBeta, data.Rung2Host, echo); err != nil {
			t.Fatalf("[SEC 1] BuildPingInto failed: %v", err)
		}
		if pe.ActiveInFlight() != 1 {
			t.Fatalf("[SEC 1] expected 1 in-flight ping, got %d", pe.ActiveInFlight())
		}
		clk.Advance(delta)
		if err := engine.ReflectPingInto(&req, &resp, common.NodeBeta); err != nil {
			t.Fatalf("[SEC 1] ReflectPingInto failed: %v", err)
		}
		rtt, matched := pe.ProcessPong(&resp)
		if !matched || rtt != delta {
			t.Fatalf("[SEC 1] expected RTT=%d matched=true, got rtt=%d matched=%v", delta, rtt, matched)
		}
		if !bytes.Equal(resp.PayloadSlice()[data.PingTimestampBytes:], echo) {
			t.Fatalf("[SEC 1] echo payload mismatch: %q", resp.PayloadSlice()[data.PingTimestampBytes:])
		}
	}

	h := pe.HealthSnapshot()
	if h.PingsSent != 3 || h.PongsRecv != 3 || h.LossPercent() != 0 ||
		h.MinRTTNano != 100 || h.MaxRTTNano != 300 || h.AvgRTTNano != 200 ||
		pe.ActiveInFlight() != 0 {
		t.Fatalf("[SEC 1] unexpected LinkHealth: %+v", h)
	}

	// Nil clock argument must default safely to hardware MonotonicClock.
	hw := engine.NewPingEngine(common.NodeAlpha, nil)
	common.AssertAuditPass(t, hw.AuditState(), hw.ExecutePingDumpCard())
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: [ SECTION 2: NEVER FAIL (FAULT INVERSION & DEGRADATION) ]
// ------------------------------------------------------------------------------

func TestPingUnit_Sec2_EveryBuildReflectAndProcessPongFaultBranch(t *testing.T) {
	clk := engine.NewMockClock(500)
	pe := engine.NewPingEngine(common.NodeAlpha, clk)
	var req, resp data.Frame

	if err := pe.BuildPingInto(nil, common.NodeBeta, data.Rung2Host, nil); err != engine.ErrPingNilFrame {
		t.Fatalf("[SEC 2] nil dst: want ErrPingNilFrame got %v", err)
	}
	if err := pe.BuildPingInto(&req, data.NodeIDNull, data.Rung2Host, nil); err != engine.ErrPingInvalidTarget {
		t.Fatalf("[SEC 2] null target: want ErrPingInvalidTarget got %v", err)
	}
	if err := pe.BuildPingInto(&req, common.NodeBeta, data.RungNone, nil); err != engine.ErrInvalidRung {
		t.Fatalf("[SEC 2] RungNone: want ErrInvalidRung got %v", err)
	}
	if err := pe.BuildPingInto(&req, common.NodeBeta, data.Rung2Host,
		make([]byte, engine.MaxPingEchoBytes+1)); err != engine.ErrPingEchoTooLarge {
		t.Fatalf("[SEC 2] oversize echo: want ErrPingEchoTooLarge got %v", err)
	}

	// ReflectPingInto error branches.
	if err := engine.ReflectPingInto(nil, &resp, common.NodeBeta); err != engine.ErrPingNilFrame {
		t.Fatalf("[SEC 2] nil src reflect: got %v", err)
	}
	if err := engine.ReflectPingInto(&req, nil, common.NodeBeta); err != engine.ErrPingNilFrame {
		t.Fatalf("[SEC 2] nil dst reflect: got %v", err)
	}
	req.Opcode = data.OpData
	if err := engine.ReflectPingInto(&req, &resp, common.NodeBeta); err != engine.ErrPingNotRequest {
		t.Fatalf("[SEC 2] non-OpPing reflect: got %v", err)
	}
	req.Opcode = data.OpPing
	if err := engine.ReflectPingInto(&req, &resp, data.NodeIDNull); err != engine.ErrPingInvalidTarget {
		t.Fatalf("[SEC 2] null responder reflect: got %v", err)
	}

	// ProcessPong rejections: nil, wrong opcode, foreign target, duplicate.
	_ = pe.BuildPingInto(&req, common.NodeBeta, data.Rung2Host, nil)
	_ = engine.ReflectPingInto(&req, &resp, common.NodeBeta)

	if _, ok := pe.ProcessPong(nil); ok {
		t.Fatal("[SEC 2] nil pong must not match")
	}
	wrongOp := resp
	wrongOp.Opcode = data.OpData
	if _, ok := pe.ProcessPong(&wrongOp); ok {
		t.Fatal("[SEC 2] non-OpPong must not match")
	}
	foreign := resp
	foreign.TargetID = common.NodeRelay
	if _, ok := pe.ProcessPong(&foreign); ok {
		t.Fatal("[SEC 2] foreign-target OpPong must not match")
	}
	if _, ok := pe.ProcessPong(&resp); !ok {
		t.Fatal("[SEC 2] first valid OpPong should match")
	}
	if _, dup := pe.ProcessPong(&resp); dup {
		t.Fatal("[SEC 2] duplicate OpPong must be rejected")
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: [ SECTION 3: INVARIANTS, SATURATION & LIMIT GUARDS ]
// ------------------------------------------------------------------------------

func TestPingUnit_Sec3_RingWrapZeroAllocAndConcurrentRaceSafety(t *testing.T) {
	clk := engine.NewMockClock(100)
	pe := engine.NewPingEngine(common.NodeAlpha, clk)
	var firstReq, latestReq, resp data.Frame

	_ = pe.BuildPingInto(&firstReq, common.NodeBeta, data.Rung2Host, nil)
	for i := 0; i < engine.MaxInFlightPings; i++ {
		clk.Advance(5)
		_ = pe.BuildPingInto(&latestReq, common.NodeBeta, data.Rung2Host, nil)
	}
	if pe.ActiveInFlight() != engine.MaxInFlightPings {
		t.Fatalf("[SEC 3] expected full 64-slot ring, got %d", pe.ActiveInFlight())
	}

	// Overwritten seq=1 must NOT match; latest seq=65 MUST match.
	_ = engine.ReflectPingInto(&firstReq, &resp, common.NodeBeta)
	if _, stale := pe.ProcessPong(&resp); stale {
		t.Fatal("[SEC 3] stale wrapped seq=1 erroneously matched")
	}
	clk.Advance(20)
	_ = engine.ReflectPingInto(&latestReq, &resp, common.NodeBeta)
	if rtt, ok := pe.ProcessPong(&resp); !ok || rtt != 20 {
		t.Fatalf("[SEC 3] latest wrapped seq failed: ok=%v rtt=%d", ok, rtt)
	}

	// ALLOC_ZERO_HEAP proof.
	pe.Reset()
	echo := []byte("ZERO-ALLOC-SONAR")
	allocs := testing.AllocsPerRun(500, func() {
		_ = pe.BuildPingInto(&latestReq, common.NodeBeta, data.Rung1Floor, echo)
		clk.Advance(10)
		_ = engine.ReflectPingInto(&latestReq, &resp, common.NodeBeta)
		_, _ = pe.ProcessPong(&resp)
	})
	if allocs != 0 {
		t.Fatalf("[SEC 3] ALLOC_ZERO_HEAP violated in PingEngine loop: %.2f", allocs)
	}

	// Concurrent -race safety proof.
	pe.Reset()
	var wg sync.WaitGroup
	for g := 0; g < 4; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var r1, r2 data.Frame
			for i := 0; i < 40; i++ {
				_ = pe.BuildPingInto(&r1, common.NodeBeta, data.Rung2Host, nil)
				clk.Advance(3)
				_ = engine.ReflectPingInto(&r1, &r2, common.NodeBeta)
				_, _ = pe.ProcessPong(&r2)
				_ = pe.HealthSnapshot()
			}
		}()
	}
	wg.Wait()
	common.AssertAuditPass(t, pe.AuditState(), pe.ExecutePingDumpCard())
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.4: [ SECTION 4: EMERGENCE & COMPOSITE WORKFLOWS ]
// ------------------------------------------------------------------------------

func TestPingUnit_Sec4_LiveAuditCardAndCleanResetLifecycle(t *testing.T) {
	clk := engine.NewMockClock(42)
	pe := engine.NewPingEngine(common.NodeAlpha, clk)
	var req, resp data.Frame

	_ = pe.BuildPingInto(&req, common.NodeBeta, data.Rung3Local, []byte("CARD"))
	clk.Advance(88)
	_ = engine.ReflectPingInto(&req, &resp, common.NodeBeta)
	_, _ = pe.ProcessPong(&resp)

	card := pe.ExecutePingDumpCard()
	if !strings.Contains(card, "1 sent / 1 pongs") || !strings.Contains(card, "RUNG_3_LOCAL") {
		t.Fatalf("[SEC 4] unexpected Ping dump card:\n%s", card)
	}
	common.AssertAuditPass(t, pe.AuditState(), card)

	// Invalid selfID & nil receiver fault inversion on AuditState.
	var nilPE *engine.PingEngine
	common.AssertAuditFail(t, nilPE.AuditState(), data.AuditSeverityCritical)
	badPE := engine.NewPingEngine(data.NodeIDNull, clk)
	common.AssertAuditFail(t, badPE.AuditState(), data.AuditSeverityCritical)

	pe.Reset()
	if pe.HealthSnapshot().PingsSent != 0 || pe.HealthSnapshot().MinRTTNano != -1 ||
		pe.ActiveInFlight() != 0 {
		t.Fatal("[SEC 4] PingEngine.Reset() left stale telemetry behind")
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Preflight Diagnostic Gates & Physical Source Anchor Verification )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.engine.ping
// Mission:       Run PreflightPingEngine and verify physical source anchor.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Block 4 Pre-Flight Contract Gates
// ------------------------------------------------------------------------------

func TestPingUnit_Block4_PreflightGates(t *testing.T) {
	if pingUnitSuiteName == "" {
		t.Fatal("[BLOCK 4] suite identifier must not be empty")
	}
	common.AssertAuditPass(t, engine.PreflightPingEngine(), "")
	if !engine.AuditPingEngine() {
		t.Fatal("[BLOCK 4] AuditPingEngine returned false")
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Physical Source Anchor Verification
// ------------------------------------------------------------------------------

func TestPingUnit_Block4_SourceAnchors(t *testing.T) {
	common.AssertRepoAnchors(t, "engine/ping.go")
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Test Suite Epilogue & Clean Readiness Handshake )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.engine.ping
// Mission:       Confirm PingEngine and harness epilogue readiness post-suite.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Epilogue Readiness Verification
// ------------------------------------------------------------------------------

func TestPingUnit_Block5_EpilogueReadiness(t *testing.T) {
	if !engine.IsPingEngineReady() || !common.IsHarnessReady() {
		t.Fatal("[BLOCK 5: EPILOGUE] ping unit readiness handshake failed")
	}
}

// End of tests/unit/engine/ping_test.go
