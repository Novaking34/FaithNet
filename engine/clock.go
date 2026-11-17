// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          engine/clock.go
// @id            cog.engine.clock
// @layer         Layer 1: Pure Functional Engine Substrate
// @archetype     operational | stateful
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Monotonic Temporal Pulse Substrate (Cog 4)
// @role          Runtime Engine: NTP-Independent Monotonic Clock & Mock Clock
// @maturity      Rev 5 Baseline | Wave 1 Seed
// @chassis       zero-heap | Interface + fixed value anchors (no NTP I/O)
// @scale         O(1) nanosecond elapsed queries; lock-free atomic MockClock
// @seams         in: stdlib time, sync/atomic, fmt | exports: Clock, mocks
// @invariants    1. Stratum Baseline: ALLOC_ZERO_HEAP, Downward Dependency Law,
//                   Universal 5-Block Lifecycle Protocol (Rev 5).
//                2. Zero dependence on external NTP or wall-clock calendars.
//                3. MonotonicClock measures strictly non-decreasing deltas.
//                4. MockClock enables 100% deterministic RTT test replay.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Provides Cog 4 of the FaithNet Ping machine: a pure monotonic temporal
//   pulse. Go's time.Since(anchor) reads the hardware monotonic clock directly
//   without consulting wall-clock adjustments or external NTP servers, so
//   round-trip latency measurements remain truthful on air-gapped machines.
//
//   ### For Architects, Operators & Non-Programmers:
//   Acts like an internal stopwatch built right into the CPU. It never asks
//   Google or the internet what time it is—it only measures how long an echo
//   took to travel and return.
//
//   ### For Systems Programmers & Engine Architects:
//   - Controlled Memory Chassis: Fixed 24-byte time.Time anchor or 8-byte
//     atomic.Int64 tick register; zero heap allocations on NowNano().
//   - Probe & Config Wiring: Block 4 verifies non-decreasing monotonicity via
//     AuditClock (4.1) and PreflightClock (4.2).
// ==============================================================================

package engine

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Clock Bounds, Logical Anchor, Interface & Concrete Chassis Structs )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 1: Pure Functional Engine Substrate
// Module:        cog.engine.clock
// Mission:       Declare the monotonic Clock interface, concrete hardware
//                clock anchor, and deterministic atomic MockClock chassis.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"fmt"
	"sync/atomic"
	"time"
)

// --- Tier 2: External Library Dependencies (None — 100% Go Stdlib Invariant) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---
import "github.com/CreativeWorkzStudio/FaithNet/data"

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Pre-Conditions & Temporal Bounds
// ------------------------------------------------------------------------------
// Explicit nanosecond floor guarding against negative clock skew.
// ------------------------------------------------------------------------------

const (
	// ClockMinTickNano is the non-negative monotonic floor for relative ticks.
	ClockMinTickNano int64 = 0
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.2: Operational Logical Anchor & Interface Contract
// ------------------------------------------------------------------------------
// Movement-resilient logical ID and minimal Clock contract interface.
// ------------------------------------------------------------------------------

const (
	// ModuleIDClock is the canonical logical ID for the temporal clock cog.
	ModuleIDClock = "cog.engine.clock"
)

// Clock provides strictly monotonic nanosecond ticks relative to local boot.
type Clock interface {
	NowNano() int64
}

// ------------------------------------------------------------------------------
// Sub-Assembly 2.3: Public Tagged Ontology (Concrete Clock Chassis)
// ------------------------------------------------------------------------------
// Fixed-size monotonic and mock clock structs.
// ------------------------------------------------------------------------------

// MonotonicClock anchors an epoch at creation and returns elapsed nanoseconds
// using Go's monotonic runtime clock (immune to NTP wall-clock jumps).
type MonotonicClock struct {
	anchor time.Time // Hardware-backed monotonic boot anchor captured at init
}

// MockClock is a deterministic, thread-safe manual clock for verification tests.
type MockClock struct {
	ticks atomic.Int64 // Lock-free deterministic nanosecond counter
}

// ------------------------------------------------------------------------------
// Sub-Assembly 2.4: Pre-Conditional Helper Cogs & O(1) Delta Clamps
// ------------------------------------------------------------------------------
// Pure O(1) non-negative latency clamp guarding against clock anomalies.
// ------------------------------------------------------------------------------

// ClampDeltaNano clamps elapsed duration (end - start) to >= 0 nanoseconds.
//
// ### For Architects, Operators & Non-Programmers:
// Prevents negative travel times if a system clock ever jitters backward.
//
// ### For Systems Programmers & Engine Architects:
// Pure O(1) saturating subtraction clamped at ClockMinTickNano (0).
func ClampDeltaNano(startNano, endNano int64) int64 {
	delta := endNano - startNano
	if delta < ClockMinTickNano {
		return ClockMinTickNano
	}
	return delta
}

// ==============================================================================
// BLOCK 3: BODY BLOCK
// ( Composable Business Logic: Monotonic Tick Queries & Manual Stepping )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 1: Pure Functional Engine Substrate
// Module:        cog.engine.clock
// Mission:       Serve non-decreasing monotonic ticks and deterministic mock
//                advances cleanly without lifecycle or diagnostic coupling.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: Production Monotonic Clock Sampling
// ------------------------------------------------------------------------------

// NowNano returns monotonic nanoseconds elapsed since the clock was anchored.
//
// ### For Architects, Operators & Non-Programmers:
// Reads the computer's built-in stopwatch without asking the internet for time.
//
// ### For Systems Programmers & Engine Architects:
// O(1) zero-allocation call to time.Since(c.anchor).Nanoseconds().
func (c *MonotonicClock) NowNano() int64 {
	return ClampDeltaNano(0, time.Since(c.anchor).Nanoseconds())
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: Deterministic Mock Clock Operations
// ------------------------------------------------------------------------------

// NowNano returns the current deterministic tick value.
//
// ### For Architects, Operators & Non-Programmers:
// Reads the simulated stopwatch used during automated laboratory tests.
//
// ### For Systems Programmers & Engine Architects:
// O(1) lock-free atomic int64 load; safe for concurrent test goroutines.
func (m *MockClock) NowNano() int64 {
	return m.ticks.Load()
}

// Advance steps the deterministic clock forward by deltaNano nanoseconds.
//
// ### For Architects, Operators & Non-Programmers:
// Moves the test stopwatch forward by an exact amount to verify speed math.
//
// ### For Systems Programmers & Engine Architects:
// O(1) atomic addition returning the new monotonic nanosecond reading.
func (m *MockClock) Advance(deltaNano int64) int64 {
	return m.ticks.Add(deltaNano)
}

// Set assigns an exact nanosecond tick value.
//
// ### For Architects, Operators & Non-Programmers:
// Sets the test stopwatch to an exact starting position for replay.
//
// ### For Systems Programmers & Engine Architects:
// O(1) atomic store; strictly intended for deterministic test setup.
func (m *MockClock) Set(valNano int64) {
	m.ticks.Store(valNano)
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Contract Verification Cogs & Pre-Flight Validation Tripwires )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 1: Pure Functional Engine Substrate
// Module:        cog.engine.clock
// Mission:       Prove monotonic non-decreasing behavior, live clock contracts,
//                and 4-section pre-flight temporal stability without NTP.
// @triggers
//   - Block 4 Audits : ValidateClockContract(), PreflightClock(), AuditClock()
//   - Layer 4 Tests  : tests/functional/engine_functional_test.go,
//                      tests/unit/engine/clock_test.go
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Contract, Input & Live-Instance Verification Cogs
// ------------------------------------------------------------------------------

// ValidateClockContract verifies non-destructively that clock c is non-nil and
// returns non-decreasing, non-negative monotonic readings across two probes.
//
// ### For Architects, Operators & Non-Programmers:
// Checks a live stopwatch to confirm it is connected and never ticks backwards.
//
// ### For Systems Programmers & Engine Architects:
// Read-only O(1) live-instance audit returning structured `data.AuditResult`.
func ValidateClockContract(c Clock) data.AuditResult {
	const check = "clock_live_monotonic_contract"

	if c == nil {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDClock, check,
			"Clock instance is nil.")
	}
	t1 := c.NowNano()
	t2 := c.NowNano()
	if t1 < ClockMinTickNano || t2 < ClockMinTickNano {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDClock, check,
			"Clock returned negative monotonic tick reading.")
	}
	if t2 < t1 {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDClock, check,
			"Clock regressed monotonically between consecutive samples.")
	}
	return data.AuditPass(ModuleIDClock, check,
		"Live clock is non-nil, non-negative, and monotonic.")
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Machine Self-Audit & Pre-Flight Validation Tripwires
// ------------------------------------------------------------------------------

// PreflightClock executes a 4-section non-destructive pre-flight self-audit
// covering (1) nominal ticking, (2) negative-skew fault clamping, (3) zero
// baseline/monotonic invariants, and (4) Set/Reset lifecycle cleanliness.
//
// ### For Architects, Operators & Non-Programmers:
// Bench-tests both the hardware stopwatch and laboratory mock stopwatch before
// FaithNet relies on them for sonar latency measurements.
//
// ### For Systems Programmers & Engine Architects:
// Stack-pure 4-section pre-flight probe returning structured `data.AuditResult`.
func PreflightClock() data.AuditResult {
	const check = "clock_preflight_4section"

	// Micro Check 1 (Always Pass): Nominal mock advance and hardware sampling.
	mc := NewMockClock(1000)
	if mc.NowNano() != 1000 || mc.Advance(250) != 1250 {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDClock, check,
			"Sec 1 MockClock nominal tick/advance math failed.")
	}
	rc := NewMonotonicClock()
	if res := ValidateClockContract(rc); !res.IsPass() {
		return res
	}

	// Micro Check 2 (Never Fail): Negative skew and nil contract inversion.
	if ClampDeltaNano(500, 100) != 0 || ClampDeltaNano(100, 100) != 0 {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDClock, check,
			"Sec 2 ClampDeltaNano failed to clamp negative or zero skew.")
	}
	if ValidateClockContract(nil).IsPass() {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDClock, check,
			"Sec 2 ValidateClockContract failed to reject nil Clock.")
	}

	// Micro Check 3 (Invariants & Limits): Exact positive delta & Set bounds.
	if ClampDeltaNano(100, 500) != 400 {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDClock, check,
			"Sec 3 ClampDeltaNano positive interval math violated.")
	}
	mc.Set(999_999)
	if mc.NowNano() != 999_999 {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDClock, check,
			"Sec 3 MockClock.Set() did not update atomic tick accurately.")
	}

	// Micro Check 4 (Emergence): Clean Reset() restoration to ClockMinTickNano.
	mc.Reset()
	if mc.NowNano() != ClockMinTickNano {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDClock, check,
			"Sec 4 MockClock.Reset() left non-zero residual tick state.")
	}
	rc.Reset()
	if res := ValidateClockContract(rc); !res.IsPass() {
		return res
	}

	return data.AuditPass(ModuleIDClock, check,
		"Clock passed all 4 micro pre-flight gates (nominal/skew/bounds/reset).")
}

// AuditClock returns true iff PreflightClock() completes with AuditSeverityPass.
//
// ### For Architects, Operators & Non-Programmers:
// Quick pass/fail check confirming FaithNet's offline stopwatches are healthy.
//
// ### For Systems Programmers & Engine Architects:
// Convenience boolean bridge delegating to `PreflightClock().IsPass()`.
func AuditClock() bool {
	return PreflightClock().IsPass()
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK / EXECUTABLE PROCEDURES & LIFECYCLE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 1: Pure Functional Engine Substrate
// Module:        cog.engine.clock
// Mission:       Render 60-column temporal diagnostic cards, expose epilogue
//                readiness, and own MonotonicClock/MockClock lifecycles.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Executable Procedures, Inspection, Readiness & Cards
// ------------------------------------------------------------------------------

// IsClockReady reports whether Cog 4 passed its full pre-flight audit.
//
// ### For Architects, Operators & Non-Programmers:
// Epilogue confirmation that offline timing is trustworthy and ready.
//
// ### For Systems Programmers & Engine Architects:
// Matches OmniCode's Block 5 epilogue readiness predicate (`is_*_ready`).
func IsClockReady() bool {
	return PreflightClock().IsPass()
}

// ExecuteClockDumpCard formats a 60-column diagnostic card for clock c.
//
// ### For Architects, Operators & Non-Programmers:
// Displays which stopwatch mode is active and its current nanosecond reading.
//
// ### For Systems Programmers & Engine Architects:
// Non-mutating diagnostic formatter safe to invoke on nil or live Clock values.
func ExecuteClockDumpCard(c Clock) string {
	var now int64
	mode := "NIL"
	if c != nil {
		now = c.NowNano()
		switch c.(type) {
		case *MonotonicClock:
			mode = "HARDWARE_MONOTONIC (NTP-INDEPENDENT)"
		case *MockClock:
			mode = "DETERMINISTIC_MOCK_ATOMIC"
		default:
			mode = "CUSTOM_CLOCK_IMPL"
		}
	}
	return fmt.Sprintf(
		"+------------------------------------------------------------+\n"+
			"| %-58s |\n"+
			"+------------------------------------------------------------+\n"+
			"| Module:    %-47s |\n"+
			"| Mode:      %-47s |\n"+
			"| Tick (ns): %-47d |\n"+
			"+------------------------------------------------------------+",
		"FAITHNET MONOTONIC CLOCK DIAGNOSTIC CARD",
		ModuleIDClock,
		mode,
		now,
	)
}

// ------------------------------------------------------------------------------
// Sub-Assembly 5.2: Subsystem Lifecycle Initialization & Teardown
// ------------------------------------------------------------------------------

// NewMonotonicClock initializes a monotonic clock anchored at startup.
//
// ### For Architects, Operators & Non-Programmers:
// Starts a fresh offline hardware stopwatch anchored to the current instant.
//
// ### For Systems Programmers & Engine Architects:
// Allocates a single MonotonicClock whose Go `time.Time` carries monotonic ticks.
func NewMonotonicClock() *MonotonicClock {
	return &MonotonicClock{anchor: time.Now()}
}

// Reset re-anchors the monotonic clock epoch cleanly.
//
// ### For Architects, Operators & Non-Programmers:
// Zeroes the hardware stopwatch back to the current moment.
//
// ### For Systems Programmers & Engine Architects:
// Re-samples `time.Now()` to reset relative nanosecond elapsed origin.
func (c *MonotonicClock) Reset() {
	c.anchor = time.Now()
}

// NewMockClock initializes a deterministic atomic mock clock at startNano.
//
// ### For Architects, Operators & Non-Programmers:
// Creates a controllable test stopwatch starting at an exact nanosecond mark.
//
// ### For Systems Programmers & Engine Architects:
// Initializes lock-free `atomic.Int64` tick storage for zero-flake tests.
func NewMockClock(startNano int64) *MockClock {
	mc := &MockClock{}
	mc.ticks.Store(startNano)
	return mc
}

// Reset zeroes the mock clock tick counter cleanly.
//
// ### For Architects, Operators & Non-Programmers:
// Rewinds the test stopwatch back to zero nanoseconds.
//
// ### For Systems Programmers & Engine Architects:
// Atomic store of `ClockMinTickNano` (0).
func (m *MockClock) Reset() {
	m.ticks.Store(ClockMinTickNano)
}

// End of engine/clock.go
