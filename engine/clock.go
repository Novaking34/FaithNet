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

// --- Tier 3: Internal FaithNet Foundation Packages (None) ---

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
// Mission:       Prove monotonic non-decreasing behavior and clamp invariants
//                before heartbeat engines rely on relative RTT math.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Contract, Input & Cog Verification Cogs
// ------------------------------------------------------------------------------

// AuditClock verifies non-decreasing monotonic ticks on both mock and real
// clocks and confirms negative-delta clamping holds cleanly.
func AuditClock() bool {
	if ClampDeltaNano(500, 100) != 0 || ClampDeltaNano(100, 500) != 400 {
		return false
	}
	mc := NewMockClock(1000)
	t1 := mc.NowNano()
	t2 := mc.Advance(250)
	if t1 != 1000 || t2 != 1250 {
		return false
	}
	rc := NewMonotonicClock()
	r1 := rc.NowNano()
	r2 := rc.NowNano()
	return r2 >= r1
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Machine Self-Audit & Pre-Flight Validation Tripwires
// ------------------------------------------------------------------------------

// PreflightClock executes non-destructive pre-flight clearance for Cog 4.
func PreflightClock() bool {
	return AuditClock()
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK / EXECUTABLE PROCEDURES & LIFECYCLE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 1: Pure Functional Engine Substrate
// Module:        cog.engine.clock
// Mission:       Render 60-column temporal diagnostic cards and own lifecycle
//                constructors/reset for MonotonicClock and MockClock.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Executable Procedures, Inspection & Diagnostic Cards
// ------------------------------------------------------------------------------

// ExecuteClockDumpCard formats a 60-column diagnostic card for clock c.
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
func NewMonotonicClock() *MonotonicClock {
	return &MonotonicClock{anchor: time.Now()}
}

// Reset re-anchors the monotonic clock epoch cleanly.
func (c *MonotonicClock) Reset() {
	c.anchor = time.Now()
}

// NewMockClock initializes a deterministic atomic mock clock at startNano.
func NewMockClock(startNano int64) *MockClock {
	mc := &MockClock{}
	mc.ticks.Store(startNano)
	return mc
}

// Reset zeroes the mock clock tick counter cleanly.
func (m *MockClock) Reset() {
	m.ticks.Store(ClockMinTickNano)
}

// End of engine/clock.go
