// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/unit/engine/clock_test.go
// @id            test.unit.engine.clock
// @layer         Layer 4: Tier-2 Unit Verification Suite (Engine Sub-Stratum)
// @archetype     verification_suite | unit_isolation
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Layer 1 NTP-Independent Monotonic Clock Unit Suite
// @role          Tier-2 Unit Isolation Gate for cog.engine.clock
// @maturity      Rev 5 Baseline | 4x4 Verification Matrix
// @chassis       zero-heap | Verifies lock-free atomic MockClock & MonotonicClock
// @scale         O(1) deterministic tick, clamp, and -race concurrency proofs
// @seams         in: data, engine, tests/common | target: engine/clock.go
// @invariants    1. MonotonicClock never regresses and requires zero NTP/internet.
//                2. ClampDeltaNano guarantees elapsed >= 0 on any skew.
//                3. MockClock is 100% race-free under concurrent goroutines.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Tier-2 Unit isolation suite strictly verifying Cog 4 (`engine/clock.go`):
//   hardware `MonotonicClock`, atomic `MockClock`, `ClampDeltaNano`, live
//   `ValidateClockContract`, all 4 dump-card modes, and `-race` safety.
//
//   ### For Architects, Operators & Non-Programmers:
//   Proves FaithNet's offline stopwatch ticks smoothly without internet time
//   servers and never produces corrupted negative travel-time readings.
//
//   ### For Systems Programmers & Engine Architects:
//   100% branch coverage of `engine/clock.go` including custom-clock contract
//   fault injection (regressing and negative synthetic clocks).
// ==============================================================================

package engine_unit_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Identifier, Verification Imports & Fault-Injection Test Clocks )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.engine.clock
// Mission:       Bind Clock types and define fault-injecting test doubles.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"strings"
	"sync"
	"sync/atomic"
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
	clockUnitSuiteName = "Tier 2 Unit: cog.engine.clock (Monotonic & Mock)"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.2: Fault-Injecting Custom Clock Doubles (Section 2 Testing)
// ------------------------------------------------------------------------------

type regressingClock struct {
	calls atomic.Int64
}

func (r *regressingClock) NowNano() int64 {
	if r.calls.Add(1) == 1 {
		return 1000
	}
	return 500
}

type negativeClock struct{}

func (negativeClock) NowNano() int64 { return -42 }

type steadyCustomClock struct{}

func (steadyCustomClock) NowNano() int64 { return 777 }

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.engine.clock
// Mission:       Verify hardware & mock clocks across Sections 1-4.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: [ SECTION 1: ALWAYS PASS (NOMINAL GOLDEN SURFACE COGS) ]
// ------------------------------------------------------------------------------

func TestClockUnit_Sec1_MonotonicAndMockClockNominal(t *testing.T) {
	rc := engine.NewMonotonicClock()
	r1 := rc.NowNano()
	r2 := rc.NowNano()
	if r1 < 0 || r2 < r1 {
		t.Fatalf("[SEC 1] hardware MonotonicClock regressed: r1=%d r2=%d", r1, r2)
	}
	rc.Reset()
	if rc.NowNano() < 0 {
		t.Fatal("[SEC 1] hardware MonotonicClock negative after Reset")
	}

	mc := engine.NewMockClock(5000)
	if mc.NowNano() != 5000 || mc.Advance(250) != 5250 || mc.NowNano() != 5250 {
		t.Fatalf("[SEC 1] MockClock advance mismatch: got %d", mc.NowNano())
	}
	mc.Set(9000)
	if mc.NowNano() != 9000 {
		t.Fatalf("[SEC 1] MockClock Set mismatch: got %d", mc.NowNano())
	}
	mc.Reset()
	if mc.NowNano() != engine.ClockMinTickNano {
		t.Fatalf("[SEC 1] MockClock Reset mismatch: got %d", mc.NowNano())
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: [ SECTION 2: NEVER FAIL (FAULT INVERSION & DEGRADATION) ]
// ------------------------------------------------------------------------------

func TestClockUnit_Sec2_ClampSkewAndValidateBrokenClocks(t *testing.T) {
	if engine.ClampDeltaNano(1000, 200) != 0 ||
		engine.ClampDeltaNano(500, 500) != 0 ||
		engine.ClampDeltaNano(-10, -50) != 0 {
		t.Fatal("[SEC 2] ClampDeltaNano failed to clamp non-positive delta to 0")
	}

	common.AssertAuditFail(t, engine.ValidateClockContract(nil), data.AuditSeverityCritical)
	common.AssertAuditFail(t, engine.ValidateClockContract(negativeClock{}), data.AuditSeverityCritical)
	common.AssertAuditFail(t, engine.ValidateClockContract(&regressingClock{}), data.AuditSeverityFatal)
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: [ SECTION 3: INVARIANTS, SATURATION & LIMIT GUARDS ]
// ------------------------------------------------------------------------------

func TestClockUnit_Sec3_ConcurrentMockClockRaceSafetyAndZeroAlloc(t *testing.T) {
	mc := engine.NewMockClock(0)
	var wg sync.WaitGroup
	const workers = 8
	const steps = 250

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < steps; i++ {
				mc.Advance(1)
				_ = mc.NowNano()
			}
		}()
	}
	wg.Wait()

	if want := int64(workers * steps); mc.NowNano() != want {
		t.Fatalf("[SEC 3] atomic race lost ticks: want %d got %d", want, mc.NowNano())
	}

	allocs := testing.AllocsPerRun(500, func() {
		_ = mc.NowNano()
		_ = mc.Advance(1)
		_ = engine.ClampDeltaNano(10, 50)
	})
	if allocs != 0 {
		t.Fatalf("[SEC 3] ALLOC_ZERO_HEAP violated on MockClock: %.2f", allocs)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.4: [ SECTION 4: EMERGENCE & COMPOSITE WORKFLOWS ]
// ------------------------------------------------------------------------------

func TestClockUnit_Sec4_AllFourDiagnosticDumpCardModes(t *testing.T) {
	if !strings.Contains(engine.ExecuteClockDumpCard(nil), "NIL") {
		t.Fatal("[SEC 4] nil clock dump card missing NIL label")
	}
	if !strings.Contains(engine.ExecuteClockDumpCard(engine.NewMonotonicClock()), "HARDWARE_MONOTONIC") {
		t.Fatal("[SEC 4] monotonic dump card missing HARDWARE_MONOTONIC label")
	}
	if !strings.Contains(engine.ExecuteClockDumpCard(engine.NewMockClock(12)), "DETERMINISTIC_MOCK_ATOMIC") {
		t.Fatal("[SEC 4] mock dump card missing DETERMINISTIC_MOCK_ATOMIC label")
	}
	if !strings.Contains(engine.ExecuteClockDumpCard(steadyCustomClock{}), "CUSTOM_CLOCK_IMPL") {
		t.Fatal("[SEC 4] custom clock dump card missing CUSTOM_CLOCK_IMPL label")
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Preflight Diagnostic Gates & Physical Source Anchor Verification )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.engine.clock
// Mission:       Run PreflightClock, ValidateClockContract, and anchor check.
// ==============================================================================

func TestClockUnit_Block4_PreflightAndAnchor(t *testing.T) {
	if clockUnitSuiteName == "" {
		t.Fatal("[BLOCK 4] suite identifier must not be empty")
	}
	common.AssertAuditPass(t, engine.PreflightClock(),
		engine.ExecuteClockDumpCard(engine.NewMonotonicClock()))
	common.AssertAuditPass(t, engine.ValidateClockContract(steadyCustomClock{}), "")
	if !engine.AuditClock() {
		t.Fatal("[BLOCK 4] AuditClock returned false")
	}
	common.AssertRepoAnchors(t, "engine/clock.go")
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Test Suite Epilogue & Clean Readiness Handshake )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.engine.clock
// Mission:       Confirm clock and harness epilogue readiness post-suite.
// ==============================================================================

func TestClockUnit_Block5_EpilogueReadiness(t *testing.T) {
	if !engine.IsClockReady() || !common.IsHarnessReady() {
		t.Fatal("[BLOCK 5: EPILOGUE] clock unit readiness handshake failed")
	}
}

// End of tests/unit/engine/clock_test.go
