// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          orchestrator/ladder.go
// @id            cog.orchestrator.ladder
// @layer         Layer 2: Adaptive Orchestration & Ladder Control
// @archetype     operational | stateful
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet 4-Rung Bidirectional Adaptive Survival Ladder
// @role          Rising-Up Escalation & EVT_CASCADE_DOWN Defensive Failover
// @maturity      Rev 5 Baseline | Wave 3 Orchestrator
// @chassis       bounded-heap | Fixed [5]Transceiver + [64]CascadeEvent ring
// @scale         O(4) rung selection; O(1) ring flight-recorder append
// @seams         in: data, engine.Clock, transport.Transceiver
// @invariants    1. Stratum Baseline: Downward Dependency Law, Rev 5 5-Block.
//                2. Defensive Cascade (`EVT_CASCADE_DOWN`) never panics and
//                   walks down Rung 4 -> 3 -> 2 -> 1 until a healthy floor holds.
//                3. Rising Up (`EVT_CASCADE_UP`) strictly respects
//                   `manifest.PreferredMaxRung` and `manifest.AllowGlobalEgress`.
//                4. Cascade flight-recorder history bounded to 64-slot ring.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Implements FaithNet's adaptive survival brain mirroring Track 0 of
//   MillenniumOS. Holds up to one Cog 3 `transport.Transceiver` per survival
//   rung (Rung 1 Floor through Rung 4 Global), automatically climbs to the
//   highest healthy permitted rung, falls down deterministically on link loss,
//   and records every transition in a fixed 64-entry flight-recorder ring.
//
//   ### For Architects, Operators & Non-Programmers:
//   When Wi-Fi drops or a cable is unplugged mid-conversation, the Adaptive
//   Ladder automatically shifts FaithNet down to the next working wire (or
//   radio/Morse floor) without losing control—and climbs right back up when
//   high-speed hardware heals.
//
//   ### For Systems Programmers & Engine Architects:
//   - Ownership & Reentrancy: Mutex-guarded `[5]transport.Transceiver` table
//     and `[MaxCascadeHistory]data.CascadeEvent` circular ring; 100% -race safe.
//   - Extension Seam: Any Rung 1..4 `Transceiver` registers via `RegisterRung`
//     without coupling the ladder to concrete socket or DSP implementations.
// ==============================================================================

package orchestrator

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Ladder Constants, Sentinel Errors, Logical Anchor & Chassis Struct )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 2: Adaptive Orchestration & Ladder Control
// Module:        cog.orchestrator.ladder
// Mission:       Define the fixed 4-rung transceiver chassis, 64-entry cascade
//                history ring, sentinel errors, and O(1) state accessors.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/CreativeWorkzStudio/FaithNet/data"
	"github.com/CreativeWorkzStudio/FaithNet/engine"
	"github.com/CreativeWorkzStudio/FaithNet/transport"
	"github.com/CreativeWorkzStudio/FaithNet/transport/rung2_host" // --- Tier 2: External Library Dependencies (None — 100% Go Stdlib Invariant) ---
)

// --- Tier 3: Internal FaithNet Foundation Packages ---

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Pre-Conditions, Physical Bounds & Sentinel Errors
// ------------------------------------------------------------------------------

const (
	// MaxCascadeHistory is the power-of-two capacity of the cascade ring log.
	MaxCascadeHistory = 64

	// MaxTransceiversPerRung bounds sibling mediums per survival rung (e.g.,
	// Rung 1 can hold SLIP Serial + AFSK Audio + Morse CW simultaneously).
	MaxTransceiversPerRung = 4

	// MaxTotalTransceivers is the fixed upper bound across all 4 rungs (16).
	MaxTotalTransceivers = 4 * MaxTransceiversPerRung

	// rungSlotCount holds indices 0..4 (index 0 unused; 1..4 = Rung1..Rung4).
	rungSlotCount = 5
)

var (
	// ErrFloorCollapsed indicates every registered rung down to floor failed.
	ErrFloorCollapsed = errors.New("faithnet/ladder: all registered survival rungs offline")

	// ErrInvalidLadderRung indicates registration with a nil transceiver or invalid rung.
	ErrInvalidLadderRung = errors.New("faithnet/ladder: invalid transceiver or rung level")
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.2: Operational Logical Anchor
// ------------------------------------------------------------------------------

const (
	// ModuleIDLadder is the canonical logical ID for the 4-Rung Adaptive Ladder.
	ModuleIDLadder = "cog.orchestrator.ladder"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.3: Public Tagged Ontology (AdaptiveLadder State Chassis)
// ------------------------------------------------------------------------------

// AdaptiveLadder orchestrates bidirectional Rising-Up and Falling-Down failover
// across registered Rung 1..4 transceivers.
//
// ### For Architects, Operators & Non-Programmers:
// Keeps all available communication wires ordered from survival floor to
// global backbone, shifting gears smoothly whenever hardware heals or breaks.
//
// ### For Systems Programmers & Engine Architects:
// Fixed-size `[5][4]transport.Transceiver` table + `[64]data.CascadeEvent` ring
// guarded by `sync.RWMutex`. Zero dynamic slice growth on transitions.
type AdaptiveLadder struct {
	mu            sync.RWMutex
	manifest      data.NodeManifest
	clock         engine.Clock
	slots         [rungSlotCount][MaxTransceiversPerRung]transport.Transceiver
	slotCounts    [rungSlotCount]int
	degradedUntil [rungSlotCount]int64
	activeRung    data.RungLevel
	history       [MaxCascadeHistory]data.CascadeEvent
	historyHead   uint64
	historyCount  uint64
	cascadeDowns  uint64
	cascadeUps    uint64
}

// ------------------------------------------------------------------------------
// Sub-Assembly 2.4: Pre-Conditional Helper Cogs & O(1) State Queries
// ------------------------------------------------------------------------------

// ActiveRung returns the currently selected survival rung level.
//
// ### For Architects, Operators & Non-Programmers:
// Reports which communication tier FaithNet is actively riding right now.
//
// ### For Systems Programmers & Engine Architects:
// Read-locked O(1) getter returning `data.RungNone` if floor collapsed.
func (l *AdaptiveLadder) ActiveRung() data.RungLevel {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.activeRung
}

// ActiveTransceiver returns the currently selected Cog 3 Transceiver, or nil.
//
// ### For Architects, Operators & Non-Programmers:
// Hands back the currently active wire plug so callers can inspect or use it.
//
// ### For Systems Programmers & Engine Architects:
// Read-locked O(1) slot lookup bounded strictly to `[RungMinValid..RungMaxValid]`.
func (l *AdaptiveLadder) ActiveTransceiver() transport.Transceiver {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if !l.activeRung.IsValid() {
		return nil
	}
	if tx := l.findHealthyOnRungLocked(l.activeRung); tx != nil {
		return tx
	}
	if n := l.slotCounts[l.activeRung]; n > 0 {
		return l.slots[l.activeRung][n-1]
	}
	return nil
}

// RegisteredTransceivers copies all non-nil registered transceivers into dst
// and returns the count written (at most MaxTotalTransceivers).
//
// ### For Architects, Operators & Non-Programmers:
// Lists every communication wire currently plugged into this node's ladder.
//
// ### For Systems Programmers & Engine Architects:
// Bounded zero-allocation read-locked copy into caller-owned buffer.
func (l *AdaptiveLadder) RegisteredTransceivers(dst []transport.Transceiver) int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	written := 0
	for r := data.RungMinValid; r <= data.RungMaxValid && written < len(dst); r++ {
		for i := 0; i < l.slotCounts[r] && written < len(dst); i++ {
			if l.slots[r][i] != nil {
				dst[written] = l.slots[r][i]
				written++
			}
		}
	}
	return written
}

// ------------------------------------------------------------------------------
// Sub-Assembly 2.5: Monitored Support Infrastructure
// ------------------------------------------------------------------------------
// Support Role:  Internal policy ceiling & fixed-ring history recorder (caller
//                must hold l.mu.Lock()).
// Pattern Watch: Keep strictly internal to AdaptiveLadder transitions.
// ------------------------------------------------------------------------------

func (l *AdaptiveLadder) effectiveMaxRungLocked() data.RungLevel {
	top := l.manifest.PreferredMaxRung
	if !top.IsValid() {
		top = data.Rung3Local
	}
	if !l.manifest.AllowGlobalEgress && top > data.Rung3Local {
		top = data.Rung3Local
	}
	return top
}

func (l *AdaptiveLadder) effectiveMinRungLocked() data.RungLevel {
	floor := l.manifest.MinimumFloorRung
	if !floor.IsValid() {
		return data.Rung1Floor
	}
	return floor
}

func (l *AdaptiveLadder) recordTransitionLocked(
	from, to data.RungLevel,
	dir data.Opcode,
	reason string,
) {
	evt := data.CascadeEvent{
		TimestampNano: l.clock.NowNano(),
		FromRung:      from,
		ToRung:        to,
		Direction:     dir,
		Reason:        reason,
	}
	idx := l.historyHead & (MaxCascadeHistory - 1)
	l.history[idx] = evt
	l.historyHead++
	if l.historyCount < MaxCascadeHistory {
		l.historyCount++
	}
	if dir == data.OpCascadeDown {
		l.cascadeDowns++
	} else if dir == data.OpCascadeUp {
		l.cascadeUps++
	}
}

func (l *AdaptiveLadder) findHealthyOnRungLocked(r data.RungLevel) transport.Transceiver {
	if !r.IsValid() {
		return nil
	}
	for i := l.slotCounts[r] - 1; i >= 0; i-- {
		if tx := l.slots[r][i]; tx != nil && tx.Healthy() {
			return tx
		}
	}
	return nil
}

func (l *AdaptiveLadder) isRungDegradedLocked(r data.RungLevel) bool {
	if !r.IsValid() {
		return false
	}
	until := l.degradedUntil[r]
	return until > 0 && l.clock.NowNano() < until
}

func (l *AdaptiveLadder) findHighestHealthyLocked(
	ceiling, floor data.RungLevel,
) data.RungLevel {
	if ceiling < floor {
		return data.RungNone
	}
	for r := ceiling; r >= floor; r-- {
		if !l.isRungDegradedLocked(r) && l.findHealthyOnRungLocked(r) != nil {
			return r
		}
		if r == data.RungMinValid {
			break
		}
	}
	return data.RungNone
}

// ==============================================================================
// BLOCK 3: BODY BLOCK
// ( Composable Business Logic: Registration, Rising-Up, Cascade-Down & Send )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 2: Adaptive Orchestration & Ladder Control
// Module:        cog.orchestrator.ladder
// Mission:       Register survival transceivers, execute bidirectional Rising-Up
//                and Defensive Cascade transitions, and transmit adaptively.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: Rung Registration & Bidirectional Adaptation
// ------------------------------------------------------------------------------

// RegisterRung attaches a Cog 3 Transceiver at its declared survival rung and
// immediately re-evaluates whether the ladder can rise up cleanly.
//
// ### For Architects, Operators & Non-Programmers:
// Plugs a communication wire (RAM, IPC, Wi-Fi, Serial, Radio) into the ladder;
// if it is healthy and faster than the current rung, FaithNet climbs up to it.
//
// ### For Systems Programmers & Engine Architects:
// Rejects nil transceivers or invalid rungs with `ErrInvalidLadderRung`.
func (l *AdaptiveLadder) RegisterRung(tx transport.Transceiver) error {
	if tx == nil || !tx.Rung().IsValid() {
		return ErrInvalidLadderRung
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	r := tx.Rung()
	replaced := false
	for i := 0; i < l.slotCounts[r]; i++ {
		if l.slots[r][i] != nil && l.slots[r][i].Name() == tx.Name() {
			l.slots[r][i] = tx
			replaced = true
			break
		}
	}
	if !replaced {
		if l.slotCounts[r] < MaxTransceiversPerRung {
			l.slots[r][l.slotCounts[r]] = tx
			l.slotCounts[r]++
		} else {
			l.slots[r][MaxTransceiversPerRung-1] = tx
		}
	}

	top := l.effectiveMaxRungLocked()
	floor := l.effectiveMinRungLocked()
	best := l.findHighestHealthyLocked(top, floor)
	if best != l.activeRung && best.IsValid() {
		prev := l.activeRung
		dir := data.OpCascadeUp
		if prev.IsValid() && best < prev {
			dir = data.OpCascadeDown
		}
		l.activeRung = best
		l.recordTransitionLocked(prev, best, dir, "rung registered: "+tx.Name())
	}
	return nil
}

// EvaluateAndAdapt inspects all registered rungs and either Rises Up to a
// newly healed higher rung or Cascades Down if the active rung degraded.
//
// ### For Architects, Operators & Non-Programmers:
// Checks the health of every attached wire—automatically climbing back up when
// Wi-Fi recovers, or stepping down to the survival floor if the active link dies.
//
// ### For Systems Programmers & Engine Architects:
// O(4) write-locked evaluation returning the newly selected `data.RungLevel`.
func (l *AdaptiveLadder) EvaluateAndAdapt(reason string) data.RungLevel {
	l.mu.Lock()
	defer l.mu.Unlock()

	top := l.effectiveMaxRungLocked()
	floor := l.effectiveMinRungLocked()
	best := l.findHighestHealthyLocked(top, floor)

	if best == l.activeRung {
		return l.activeRung
	}

	prev := l.activeRung
	dir := data.OpCascadeUp
	if !best.IsValid() || (prev.IsValid() && best < prev) {
		dir = data.OpCascadeDown
	}
	l.activeRung = best
	if reason == "" {
		if dir == data.OpCascadeUp {
			reason = "higher rung healthy"
		} else {
			reason = "active rung degraded"
		}
	}
	l.recordTransitionLocked(prev, best, dir, reason)
	return best
}

// CascadeDown forces an explicit Defensive Cascade (`EVT_CASCADE_DOWN`) below
// the current active rung onto the next highest healthy survival tier.
//
// ### For Architects, Operators & Non-Programmers:
// Immediately steps FaithNet down one or more rungs when a transmission fails
// mid-flight, logging the exact cause into the flight-recorder ring.
//
// ### For Systems Programmers & Engine Architects:
// Searches strictly below `l.activeRung` down to `MinimumFloorRung`; returns
// `ErrFloorCollapsed` only when zero healthy rungs remain.
func (l *AdaptiveLadder) CascadeDown(reason string) (data.RungLevel, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	floor := l.effectiveMinRungLocked()
	prev := l.activeRung
	var searchTop data.RungLevel
	if prev > floor {
		searchTop = prev - 1
	} else {
		// Already at floor; check if any sibling transceiver on floor is healthy.
		if floor.IsValid() && l.findHealthyOnRungLocked(floor) != nil {
			return floor, nil
		}
		l.activeRung = data.RungNone
		l.recordTransitionLocked(prev, data.RungNone, data.OpCascadeDown, reason)
		return data.RungNone, ErrFloorCollapsed
	}

	next := l.findHighestHealthyLocked(searchTop, floor)
	l.activeRung = next
	l.recordTransitionLocked(prev, next, data.OpCascadeDown, reason)
	if !next.IsValid() {
		return data.RungNone, ErrFloorCollapsed
	}
	return next, nil
}

// MarkRungDegraded places rung r into a monotonic cooldown quarantine after
// passive silent-blackhole detection and immediately triggers EVT_CASCADE_DOWN
// onto the highest remaining healthy non-degraded survival rung.
//
// ### For Architects, Operators & Non-Programmers:
// When a wireless or UDP link goes silently dead without reporting an OS error,
// this puts that link in a brief timeout box so FaithNet drops down to a backup
// wire or radio right away—and automatically retries climbing back up later.
//
// ### For Systems Programmers & Engine Architects:
// Zero-heap monotonic cooldown using `l.clock.NowNano()`; safe for deterministic
// `engine.MockClock` verification and `-race` concurrency.
func (l *AdaptiveLadder) MarkRungDegraded(
	r data.RungLevel,
	cooldown time.Duration,
	reason string,
) data.RungLevel {
	if !r.IsValid() {
		return l.ActiveRung()
	}
	if cooldown <= 0 {
		cooldown = 5 * time.Second
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	l.degradedUntil[r] = l.clock.NowNano() + cooldown.Nanoseconds()
	top := l.effectiveMaxRungLocked()
	floor := l.effectiveMinRungLocked()
	next := l.findHighestHealthyLocked(top, floor)
	prev := l.activeRung
	if !next.IsValid() {
		// Never quarantine the node's sole surviving link into a self-inflicted
		// blackout; keep it active to allow storm recovery while logging event.
		l.degradedUntil[r] = 0
		l.recordTransitionLocked(prev, prev, data.OpCascadeDown, reason+" (sole surviving link retained)")
		return prev
	}
	if next != prev {
		l.activeRung = next
		l.recordTransitionLocked(prev, next, data.OpCascadeDown, reason)
	}
	return next
}

// ClearRungDegraded lifts any passive silent-blackhole cooldown on rung r when
// live inbound or echo traffic proves the rung is healthy again.
//
// ### For Architects, Operators & Non-Programmers:
// Clears a link's quarantine immediately as soon as real traffic arrives on it.
//
// ### For Systems Programmers & Engine Architects:
// Write-locked O(1) clear on `l.degradedUntil[r]`; safe for concurrent calls.
func (l *AdaptiveLadder) ClearRungDegraded(r data.RungLevel) {
	if !r.IsValid() {
		return
	}
	l.mu.Lock()
	l.degradedUntil[r] = 0
	l.mu.Unlock()
}

// SendOnRung attempts to transmit raw wire bytes specifically on preferredRung
// (trying all healthy siblings on that rung) so OpPong echoes return on the
// exact survival rung that carried the OpPing—preventing asymmetric blackhole
// traps—and falls back seamlessly to SendFrame if preferredRung is unavailable.
//
// ### For Architects, Operators & Non-Programmers:
// Answers a neighbor on the exact same radio or wire they just called in on,
// instead of shouting back into a broken higher-rung link they already left.
//
// ### For Systems Programmers & Engine Architects:
// Zero-heap fixed-array sibling scan on `preferredRung` with `SendFrame` fallback.
func (l *AdaptiveLadder) SendOnRung(
	preferredRung data.RungLevel,
	target data.NodeID,
	raw []byte,
) (data.RungLevel, error) {
	if len(raw) > data.FloorMaxFrameSize {
		return data.RungNone, transport.ErrWireFrameOversize
	}
	if preferredRung.IsValid() {
		var siblings [MaxTransceiversPerRung]transport.Transceiver
		sibCount := 0
		l.mu.RLock()
		for i := l.slotCounts[preferredRung] - 1; i >= 0; i-- {
			if tx := l.slots[preferredRung][i]; tx != nil && tx.Healthy() {
				siblings[sibCount] = tx
				sibCount++
			}
		}
		l.mu.RUnlock()

		for i := 0; i < sibCount; i++ {
			if err := siblings[i].Send(target, raw); err == nil {
				return preferredRung, nil
			}
		}
	}
	return l.SendFrame(target, raw)
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: Adaptive Failover Transmission
// ------------------------------------------------------------------------------

// SendFrame transmits raw wire bytes over the active rung; if transmission
// fails mid-flight, it executes `EVT_CASCADE_DOWN` automatically across lower
// healthy survival rungs until delivery succeeds or the entire floor collapses.
//
// ### For Architects, Operators & Non-Programmers:
// Sends a FaithNet envelope on the fastest healthy link and seamlessly drops
// down to backup wires or radio if the primary link snaps mid-message.
//
// ### For Systems Programmers & Engine Architects:
// Returns `(deliveredRung, nil)` on success, or `(RungNone, ErrFloorCollapsed)`
// if every registered rung down to `MinimumFloorRung` failed delivery.
func (l *AdaptiveLadder) SendFrame(
	target data.NodeID,
	raw []byte,
) (data.RungLevel, error) {
	if len(raw) > data.FloorMaxFrameSize {
		return data.RungNone, transport.ErrWireFrameOversize
	}

	cur := l.EvaluateAndAdapt("pre-send health check")
	if !cur.IsValid() {
		return data.RungNone, ErrFloorCollapsed
	}

	for attempts := 0; attempts < rungSlotCount; attempts++ {
		rung := l.ActiveRung()
		if !rung.IsValid() {
			return data.RungNone, ErrFloorCollapsed
		}

		// Try every healthy sibling transceiver attached to the active rung
		// (highest priority first) before cascading down to a lower rung.
		var siblings [MaxTransceiversPerRung]transport.Transceiver
		sibCount := 0
		l.mu.RLock()
		for i := l.slotCounts[rung] - 1; i >= 0; i-- {
			if tx := l.slots[rung][i]; tx != nil && tx.Healthy() {
				siblings[sibCount] = tx
				sibCount++
			}
		}
		l.mu.RUnlock()

		var lastErr error
		var lastName string
		for i := 0; i < sibCount; i++ {
			tx := siblings[i]
			lastName = tx.Name()
			err := tx.Send(target, raw)
			if err == nil {
				return rung, nil
			}
			if err == transport.ErrWireFrameOversize {
				return data.RungNone, err
			}
			lastErr = err
		}
		if sibCount == 0 {
			return data.RungNone, ErrFloorCollapsed
		}

		// All healthy siblings on active rung failed mid-flight: cascade downward.
		next, cErr := l.CascadeDown(fmt.Sprintf("%s send failed: %v", lastName, lastErr))
		if cErr != nil || !next.IsValid() {
			return data.RungNone, ErrFloorCollapsed
		}
		// Avoid infinite retry if already standing on the lowest floor rung.
		if next == rung {
			return data.RungNone, lastErr
		}
	}
	return data.RungNone, ErrFloorCollapsed
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Contract Verification Cogs & Pre-Flight Validation Tripwires )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 2: Adaptive Orchestration & Ladder Control
// Module:        cog.orchestrator.ladder
// Mission:       Audit live AdaptiveLadder state and run a 4-section pre-flight
//                check proving Rising-Up, EVT_CASCADE_DOWN, and history bounds.
// @triggers
//   - Block 4 Audits : (*AdaptiveLadder).AuditState(), PreflightLadder(),
//                      AuditLadder()
//   - Layer 4 Tests  : tests/unit/orchestrator/ladder_test.go,
//                      tests/integrated/orchestrator_ladder_test.go
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Contract, Input & Live-Instance Verification Cogs
// ------------------------------------------------------------------------------

// AuditState inspects a live AdaptiveLadder non-destructively to confirm valid
// manifest policy, attached clock, history ring bounds, and active rung safety.
//
// ### For Architects, Operators & Non-Programmers:
// Verifies the live survival ladder obeys operator internet-egress rules and
// its flight-recorder memory bounds remain completely intact.
//
// ### For Systems Programmers & Engine Architects:
// Read-locked non-mutating live probe returning structured `data.AuditResult`.
func (l *AdaptiveLadder) AuditState() data.AuditResult {
	const check = "ladder_live_state_contract"

	if l == nil {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDLadder, check,
			"AdaptiveLadder receiver is nil.")
	}
	if !l.manifest.IsValid() {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDLadder, check,
			"AdaptiveLadder carries invalid NodeManifest.")
	}
	if clkRes := engine.ValidateClockContract(l.clock); !clkRes.IsPass() {
		return clkRes
	}

	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.historyCount > MaxCascadeHistory {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDLadder, check,
			"Cascade history ring count exceeded MaxCascadeHistory.")
	}
	if l.activeRung == data.Rung4Global && !l.manifest.AllowGlobalEgress {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDLadder, check,
			"Policy breach: activeRung is Rung4Global without AllowGlobalEgress.")
	}
	if l.activeRung.IsValid() && l.slotCounts[l.activeRung] == 0 {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDLadder, check,
			"Active rung points to nil transceiver slot.")
	}
	return data.AuditPass(ModuleIDLadder, check,
		"Live AdaptiveLadder policy, active slot, and ring bounds verified.")
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Machine Self-Audit & Pre-Flight Validation Tripwires
// ------------------------------------------------------------------------------

// PreflightLadder executes a 4-section non-destructive pre-flight self-audit
// covering (1) Rung 2 send/recv, (2) nil registration & collapsed floor fault
// inversion, (3) 64-entry history ring bounds, and (4) live AuditState & Close.
//
// ### For Architects, Operators & Non-Programmers:
// Bench-tests the adaptive survival ladder at startup so failover and recovery
// transitions are guaranteed before the node handles real traffic.
//
// ### For Systems Programmers & Engine Architects:
// Isolated 4-section pre-flight gate returning structured `data.AuditResult`.
func PreflightLadder() data.AuditResult {
	const check = "ladder_preflight_4section"

	if (MaxCascadeHistory & (MaxCascadeHistory - 1)) != 0 {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDLadder, check,
			"MaxCascadeHistory must be a power of two.")
	}

	clk := engine.NewMockClock(100)
	manifest := data.DefaultNodeManifest(0x01, "ladder-preflight")
	lad := NewAdaptiveLadder(manifest, clk)
	defer func() { _ = lad.Close() }()

	// Micro Check 2 (Never Fail) early check: empty ladder returns ErrFloorCollapsed.
	if err := lad.RegisterRung(nil); err != ErrInvalidLadderRung {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDLadder, check,
			"Sec 2 RegisterRung(nil) did not return ErrInvalidLadderRung.")
	}
	if _, err := lad.SendFrame(0x02, []byte("EMPTY")); err != ErrFloorCollapsed {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDLadder, check,
			"Sec 2 empty ladder SendFrame did not return ErrFloorCollapsed.")
	}

	// Micro Check 1 (Always Pass): Attach Rung 2 MemoryBus and deliver cleanly.
	bus := rung2_host.NewMemoryBus()
	txA := bus.Attach(0x01)
	txB := bus.Attach(0x02)
	defer func() { _ = txB.Close() }()

	if err := lad.RegisterRung(txA); err != nil || lad.ActiveRung() != data.Rung2Host {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDLadder, check,
			"Sec 1 RegisterRung failed to activate Rung2Host.")
	}
	usedRung, err := lad.SendFrame(0x02, []byte("LADDER-OK"))
	if err != nil || usedRung != data.Rung2Host {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDLadder, check,
			"Sec 1 SendFrame over Rung2Host failed.")
	}
	env, err := txB.Receive(context.Background())
	if err != nil || string(env.Slice()) != "LADDER-OK" {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDLadder, check,
			"Sec 1 peer did not receive ladder payload.")
	}

	// Micro Check 3 (Invariants & Limits): Severing Rung 2 triggers EVT_CASCADE_DOWN.
	txA.SetHealthy(false)
	clk.Advance(50)
	if r := lad.EvaluateAndAdapt("simulated cut"); r != data.RungNone {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDLadder, check,
			"Sec 3 degraded single-rung ladder failed to step to RungNone.")
	}
	txA.SetHealthy(true)
	clk.Advance(50)
	if r := lad.EvaluateAndAdapt("restored"); r != data.Rung2Host {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDLadder, check,
			"Sec 3 healed rung failed to rise back up to Rung2Host.")
	}

	// Micro Check 4 (Emergence): History snapshot & live AuditState pass cleanly.
	var hist [8]data.CascadeEvent
	if n := lad.HistorySnapshotInto(hist[:]); n < 3 {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDLadder, check,
			"Sec 4 cascade flight-recorder missed expected transitions.")
	}
	if live := lad.AuditState(); !live.IsPass() {
		return live
	}

	return data.AuditPass(ModuleIDLadder, check,
		"AdaptiveLadder passed all 4 micro pre-flight gates.")
}

// AuditLadder returns true iff PreflightLadder() completes cleanly.
//
// ### For Architects, Operators & Non-Programmers:
// Quick pass/fail check confirming the adaptive ladder passed safety checks.
//
// ### For Systems Programmers & Engine Architects:
// Convenience boolean bridge delegating to `PreflightLadder().IsPass()`.
func AuditLadder() bool {
	return PreflightLadder().IsPass()
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK / EXECUTABLE PROCEDURES & LIFECYCLE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 2: Adaptive Orchestration & Ladder Control
// Module:        cog.orchestrator.ladder
// Mission:       Expose flight-recorder snapshots, 60-col cards, epilogue
//                readiness, and AdaptiveLadder initialization & teardown.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Executable Procedures, Inspection, Readiness & Cards
// ------------------------------------------------------------------------------

// IsLadderReady reports whether the AdaptiveLadder passed pre-flight clearance.
//
// ### For Architects, Operators & Non-Programmers:
// Epilogue confirmation that FaithNet's adaptive failover engine is ready.
//
// ### For Systems Programmers & Engine Architects:
// Matches OmniCode's Block 5 epilogue readiness predicate (`is_*_ready`).
func IsLadderReady() bool {
	return PreflightLadder().IsPass()
}

// HistorySnapshotInto copies recorded CascadeEvents in chronological order
// into caller-owned dst and returns the number of entries copied.
//
// ### For Architects, Operators & Non-Programmers:
// Reads the ladder's flight-recorder log showing every time FaithNet climbed
// up or stepped down survival rungs.
//
// ### For Systems Programmers & Engine Architects:
// Zero-allocation chronological extraction from the fixed 64-entry circular ring.
func (l *AdaptiveLadder) HistorySnapshotInto(dst []data.CascadeEvent) int {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.historyCount == 0 || len(dst) == 0 {
		return 0
	}
	total := int(l.historyCount)
	start := uint64(0)
	if l.historyHead > l.historyCount {
		start = l.historyHead - l.historyCount
	}
	written := 0
	for i := 0; i < total && written < len(dst); i++ {
		idx := (start + uint64(i)) & (MaxCascadeHistory - 1)
		dst[written] = l.history[idx]
		written++
	}
	return written
}

// ExecuteLadderDumpCard renders a 60-column inner-width ladder status card.
//
// ### For Architects, Operators & Non-Programmers:
// Prints a neat framed dashboard showing the active survival rung, permitted
// policy bounds, and total Rising-Up vs Falling-Down transitions.
//
// ### For Systems Programmers & Engine Architects:
// Non-mutating read-locked diagnostic card builder.
func (l *AdaptiveLadder) ExecuteLadderDumpCard() string {
	l.mu.RLock()
	active := l.activeRung
	downs := l.cascadeDowns
	ups := l.cascadeUps
	top := l.effectiveMaxRungLocked()
	floor := l.effectiveMinRungLocked()
	registered := 0
	for r := data.RungMinValid; r <= data.RungMaxValid; r++ {
		if l.slotCounts[r] > 0 {
			registered++
		}
	}
	l.mu.RUnlock()

	return fmt.Sprintf(
		"+------------------------------------------------------------+\n"+
			"| %-58s |\n"+
			"+------------------------------------------------------------+\n"+
			"| Module:     %-46s |\n"+
			"| ActiveRung: %-46s |\n"+
			"| PolicyBand: %-46s |\n"+
			"| Registered: %-46s |\n"+
			"| Cascades:   %-46s |\n"+
			"+------------------------------------------------------------+",
		"FAITHNET 4-RUNG ADAPTIVE LADDER CARD",
		ModuleIDLadder,
		active.String(),
		fmt.Sprintf("%s .. %s", floor.String(), top.String()),
		fmt.Sprintf("%d / 4 survival rungs attached", registered),
		fmt.Sprintf("%d up / %d down (EVT_CASCADE_DOWN)", ups, downs),
	)
}

// ------------------------------------------------------------------------------
// Sub-Assembly 5.2: Subsystem Lifecycle Initialization & Teardown
// ------------------------------------------------------------------------------

// NewAdaptiveLadder constructs an AdaptiveLadder governed by manifest and clock.
//
// ### For Architects, Operators & Non-Programmers:
// Initializes a fresh 4-rung survival ladder for a node; automatically attaches
// an offline hardware clock if none was provided.
//
// ### For Systems Programmers & Engine Architects:
// Allocates a single fixed-chassis `AdaptiveLadder` with zero map overhead.
func NewAdaptiveLadder(
	manifest data.NodeManifest,
	clock engine.Clock,
) *AdaptiveLadder {
	if clock == nil {
		clock = engine.NewMonotonicClock()
	}
	return &AdaptiveLadder{
		manifest:   manifest,
		clock:      clock,
		activeRung: data.RungNone,
	}
}

// Close gracefully closes every registered transceiver on the ladder.
//
// ### For Architects, Operators & Non-Programmers:
// Shuts down every communication wire attached to this ladder cleanly.
//
// ### For Systems Programmers & Engine Architects:
// Closes all non-nil slots and resets `activeRung` to `RungNone` safely.
func (l *AdaptiveLadder) Close() error {
	l.mu.Lock()
	var toClose [MaxTotalTransceivers]transport.Transceiver
	closeCount := 0
	for r := data.RungMinValid; r <= data.RungMaxValid; r++ {
		for i := 0; i < l.slotCounts[r]; i++ {
			if l.slots[r][i] != nil {
				toClose[closeCount] = l.slots[r][i]
				closeCount++
				l.slots[r][i] = nil
			}
		}
		l.slotCounts[r] = 0
	}
	l.activeRung = data.RungNone
	l.mu.Unlock()

	var firstErr error
	for i := 0; i < closeCount; i++ {
		if err := toClose[i].Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// End of orchestrator/ladder.go
