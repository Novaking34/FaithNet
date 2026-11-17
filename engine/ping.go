// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          engine/ping.go
// @id            cog.engine.ping
// @layer         Layer 1: Pure Functional Engine Substrate
// @archetype     operational | stateful
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet 5-Cog Sovereign Ping/Pong Reactor Engine
// @role          Runtime Engine: Zero-Alloc Ping Builder, Reflector & RTT Math
// @maturity      Rev 5 Baseline | Wave 1 Seed
// @chassis       zero-heap | Fixed 64-slot ring array; 0 allocs/op hot path
// @scale         O(1) ping build, O(1) echo reflection, O(1) ring lookup
// @seams         in: data.Frame, engine.Clock | exports: PingEngine, Reflect
// @invariants    1. Stratum Baseline: ALLOC_ZERO_HEAP, Downward Dependency Law,
//                   Universal 5-Block Lifecycle Protocol (Rev 5).
//                2. BuildPingInto, ReflectPingInto, and ProcessPong execute
//                   with strictly 0 dynamic heap allocations.
//                3. Echo payload ceiling is FloorMaxPayload - 8 (472 bytes)
//                   so timestamp + echo never breaches the 512B Floor MTU.
//                4. Reflection preserves exact SeqID and payload bytes.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Implements Cog 5 of FaithNet: the irreducible sonar heartbeat machine.
//   Constructs outgoing OpPing pulses stamped with local monotonic ticks,
//   reflects incoming OpPing frames instantaneously into OpPong replies, and
//   measures exact round-trip latency, min/max/avg RTT, and packet health
//   without allocating a single byte on the heap.
//
//   ### For Architects, Operators & Non-Programmers:
//   This is the heartbeat of every FaithNet device. Just like submarine sonar
//   sending a pulse and listening for the echo, it confirms another node is
//   alive and measures how fast the link is—whether over RAM, Wi-Fi, or Morse.
//
//   ### For Systems Programmers & Engine Architects:
//   - Controlled Memory Chassis: Fixed power-of-two ring (`MaxInFlightPings = 64`)
//     indexed by `seq & (MaxInFlightPings - 1)` for O(1) lookup with zero map
//     heap growth or GC scanning.
//   - Probe & Config Wiring: Block 4 verifies exact RTT math via MockClock;
//     Block 5 owns lifecycle (`NewPingEngine`, `Reset`) and 60-col dump cards.
// ==============================================================================

package engine

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Bounded Ring Limits, Sentinel Errors, Chassis Struct & O(1) Indexing )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 1: Pure Functional Engine Substrate
// Module:        cog.engine.ping
// Mission:       Declare fixed-capacity ping tracking bounds, sentinel errors,
//                PingEngine chassis struct, and O(1) ring index helpers.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"encoding/binary"
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
// Explicit power-of-two ring sizing and pre-allocated sentinel error values.
// ------------------------------------------------------------------------------

const (
	// MaxInFlightPings is the power-of-two fixed ring capacity for RTT tracking.
	MaxInFlightPings = 64

	// MaxPingEchoBytes reserves 8 bytes for monotonic int64 timestamp prefix.
	MaxPingEchoBytes = data.FloorMaxPayload - data.PingTimestampBytes
)

var (
	ErrPingNotRequest    = errors.New("faithnet/ping: frame is not OpPing")
	ErrPingEchoTooLarge  = errors.New("faithnet/ping: echo exceeds 472 bytes")
	ErrPingInvalidTarget = errors.New("faithnet/ping: invalid target node ID")
	ErrPingNilFrame      = errors.New("faithnet/ping: nil frame pointer")
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.2: Operational Logical Anchor & File-Private Slot Struct
// ------------------------------------------------------------------------------
// Logical identity and internal fixed-size ring slot layout.
// ------------------------------------------------------------------------------

const (
	// ModuleIDPing is the canonical logical ID for the 5-Cog Ping reactor.
	ModuleIDPing = "cog.engine.ping"
)

type inFlightSlot struct {
	seqID    uint32
	targetID data.NodeID
	rung     data.RungLevel
	sentNano int64
	active   bool
}

// ------------------------------------------------------------------------------
// Sub-Assembly 2.3: Public Tagged Ontology (PingEngine State Chassis)
// ------------------------------------------------------------------------------
// Bounded lock-protected heartbeat chassis with zero dynamic allocations.
// ------------------------------------------------------------------------------

// PingEngine manages sovereign heartbeat emission, reflection, and RTT math.
type PingEngine struct {
	mu       sync.Mutex
	selfID   data.NodeID
	clock    Clock
	nextSeq  uint32
	ring     [MaxInFlightPings]inFlightSlot
	health   data.LinkHealth
	rttTotal int64
}

// ------------------------------------------------------------------------------
// Sub-Assembly 2.4: Pre-Conditional Helper Cogs & O(1) Ring Predicates
// ------------------------------------------------------------------------------
// Pure O(1) bitmask slot indexing over the power-of-two ring chassis.
// ------------------------------------------------------------------------------

func pingSlotIndex(seq uint32) uint32 {
	return seq & (MaxInFlightPings - 1)
}

// ==============================================================================
// BLOCK 3: BODY BLOCK
// ( Composable Business Logic: Ping Emission, Echo Reflection & Pong Math )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 1: Pure Functional Engine Substrate
// Module:        cog.engine.ping
// Mission:       Build outgoing OpPing pulses, reflect incoming OpPing pulses
//                statelessly into OpPong, and correlate RTT with 0 allocs/op.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: Zero-Allocation Ping Emission
// ------------------------------------------------------------------------------

// BuildPingInto populates out with an OpPing frame targeting target on rung.
// Payload layout: [0..7] big-endian int64 sentNano, [8..] echoPayload.
//
// ### For Architects, Operators & Non-Programmers:
// Stamps a fresh sonar call with the exact stopwatch reading and optional echo
// text so the sender can verify both presence and round-trip speed.
//
// ### For Systems Programmers & Engine Architects:
// Writes directly into caller-owned out (*data.Frame) and records the in-flight
// sequence inside the fixed 64-slot ring with strictly 0 heap allocations.
func (p *PingEngine) BuildPingInto(
	out *data.Frame,
	target data.NodeID,
	rung data.RungLevel,
	echoPayload []byte,
) error {
	if out == nil {
		return ErrPingNilFrame
	}
	if !target.IsValid() {
		return ErrPingInvalidTarget
	}
	if !rung.IsValid() {
		return ErrInvalidRung
	}
	if len(echoPayload) > MaxPingEchoBytes {
		return ErrPingEchoTooLarge
	}

	p.mu.Lock()
	p.nextSeq++
	seq := p.nextSeq
	now := p.clock.NowNano()

	p.ring[pingSlotIndex(seq)] = inFlightSlot{
		seqID:    seq,
		targetID: target,
		rung:     rung,
		sentNano: now,
		active:   true,
	}
	p.health.PeerID = target
	p.health.ActiveRung = rung
	p.health.PingsSent++
	p.mu.Unlock()

	out.Version = data.ProtocolVersion
	out.Rung = rung
	out.Opcode = data.OpPing
	out.Flags = data.FlagFloorCompact | data.FlagCovenantVerified
	out.Truth = data.QuadPositive
	out.SourceID = p.selfID
	out.TargetID = target
	out.SeqID = seq
	out.PayloadLen = uint16(data.PingTimestampBytes + len(echoPayload))

	binary.BigEndian.PutUint64(out.Payload[0:8], uint64(now))
	if len(echoPayload) > 0 {
		copy(out.Payload[8:8+len(echoPayload)], echoPayload)
	}
	return nil
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: Stateless Echo Reflection & Pong RTT Correlation
// ------------------------------------------------------------------------------

// ReflectPingInto transforms an incoming OpPing frame into an OpPong reply
// inside caller-owned resp with strictly zero heap allocations.
//
// ### For Architects, Operators & Non-Programmers:
// Acts as an instant mirror: the moment a node hears a PING, it flips the verb
// to PONG, addresses it back to the caller, and echoes the exact message back.
//
// ### For Systems Programmers & Engine Architects:
// 100% stateless, lock-free, 0 allocs/op. Safe even in bare-metal or
// interrupt-constrained survival floors.
func ReflectPingInto(req *data.Frame, resp *data.Frame, selfID data.NodeID) error {
	if req == nil || resp == nil {
		return ErrPingNilFrame
	}
	if req.Opcode != data.OpPing {
		return ErrPingNotRequest
	}
	if int(req.PayloadLen) > data.FloorMaxPayload {
		return ErrPayloadTooLarge
	}

	resp.Version = data.ProtocolVersion
	resp.Rung = req.Rung
	resp.Opcode = data.OpPong
	resp.Flags = req.Flags | data.FlagFloorCompact
	resp.Truth = data.QuadPositive
	resp.SourceID = selfID
	resp.TargetID = req.SourceID
	resp.SeqID = req.SeqID
	resp.PayloadLen = req.PayloadLen
	if req.PayloadLen > 0 {
		copy(resp.Payload[:req.PayloadLen], req.Payload[:req.PayloadLen])
	}
	return nil
}

// ProcessPong matches an incoming OpPong against the in-flight ring and
// updates cumulative LinkHealth telemetry. Returns (rttNano, true) on match.
//
// ### For Architects, Operators & Non-Programmers:
// Checks an arriving echo against the pings we sent, calculates the exact
// round-trip travel time, and updates the link's live health scorecard.
//
// ### For Systems Programmers & Engine Architects:
// O(1) indexed ring lookup under mutex; rejects duplicate or foreign-target
// pongs cleanly with 0 dynamic heap allocations.
func (p *PingEngine) ProcessPong(pong *data.Frame) (int64, bool) {
	if pong == nil || pong.Opcode != data.OpPong {
		return 0, false
	}
	if pong.TargetID != p.selfID {
		return 0, false
	}

	now := p.clock.NowNano()
	slotIdx := pingSlotIndex(pong.SeqID)

	p.mu.Lock()
	defer p.mu.Unlock()

	slot := &p.ring[slotIdx]
	if !slot.active || slot.seqID != pong.SeqID {
		return 0, false
	}

	rtt := ClampDeltaNano(slot.sentNano, now)
	slot.active = false

	p.health.PeerID = pong.SourceID
	p.health.ActiveRung = pong.Rung
	p.health.PongsRecv++
	p.health.LastRTTNano = rtt
	p.health.LastSeenNs = now
	if p.health.MinRTTNano < 0 || rtt < p.health.MinRTTNano {
		p.health.MinRTTNano = rtt
	}
	if rtt > p.health.MaxRTTNano {
		p.health.MaxRTTNano = rtt
	}
	p.rttTotal += rtt
	p.health.AvgRTTNano = p.rttTotal / int64(p.health.PongsRecv)
	return rtt, true
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Contract Verification Cogs & Pre-Flight Validation Tripwires )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 1: Pure Functional Engine Substrate
// Module:        cog.engine.ping
// Mission:       Verify deterministic Ping -> Pong -> RTT parity and ring
//                power-of-two capacity invariants before node boot.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Contract, Input & Cog Verification Cogs
// ------------------------------------------------------------------------------

// AuditPingEngine executes a non-destructive deterministic Ping->Pong->RTT test.
func AuditPingEngine() bool {
	if (MaxInFlightPings & (MaxInFlightPings - 1)) != 0 {
		return false
	}
	mc := NewMockClock(10_000)
	pe := NewPingEngine(0x0A, mc)
	var req, resp data.Frame

	if err := pe.BuildPingInto(&req, 0x0B, data.Rung2Host, []byte("ECHO")); err != nil {
		return false
	}
	mc.Advance(450)
	if err := ReflectPingInto(&req, &resp, 0x0B); err != nil {
		return false
	}
	rtt, ok := pe.ProcessPong(&resp)
	return ok && rtt == 450 && resp.Opcode == data.OpPong
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Machine Self-Audit & Pre-Flight Validation Tripwires
// ------------------------------------------------------------------------------

// PreflightPingEngine performs complete pre-flight clearance for Cog 5.
func PreflightPingEngine() bool {
	return PreflightClock() && AuditPingEngine()
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK / EXECUTABLE PROCEDURES & LIFECYCLE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 1: Pure Functional Engine Substrate
// Module:        cog.engine.ping
// Mission:       Provide read-only health inspection, 60-column dump cards,
//                and lifecycle initialization/reset for PingEngine.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Executable Procedures, Inspection & Diagnostic Cards
// ------------------------------------------------------------------------------

// HealthSnapshot returns a thread-safe value copy of current LinkHealth metrics.
func (p *PingEngine) HealthSnapshot() data.LinkHealth {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.health
}

// ExecutePingDumpCard renders a 60-column inner-width diagnostic inspection card.
func (p *PingEngine) ExecutePingDumpCard() string {
	h := p.HealthSnapshot()
	return fmt.Sprintf(
		"+------------------------------------------------------------+\n"+
			"| %-58s |\n"+
			"+------------------------------------------------------------+\n"+
			"| Module:     %-46s |\n"+
			"| SelfID:     %-46s |\n"+
			"| ActiveRung: %-46s |\n"+
			"| Sent/Recv:  %-46s |\n"+
			"| RTT (ns):   %-46s |\n"+
			"+------------------------------------------------------------+",
		"FAITHNET 5-COG PING ENGINE DIAGNOSTIC CARD",
		ModuleIDPing,
		p.selfID.String(),
		h.ActiveRung.String(),
		fmt.Sprintf("%d sent / %d pongs", h.PingsSent, h.PongsRecv),
		fmt.Sprintf("last=%d min=%d avg=%d max=%d",
			h.LastRTTNano, h.MinRTTNano, h.AvgRTTNano, h.MaxRTTNano),
	)
}

// ------------------------------------------------------------------------------
// Sub-Assembly 5.2: Subsystem Lifecycle Initialization & Teardown
// ------------------------------------------------------------------------------

// NewPingEngine initializes a bounded PingEngine for selfID using clock.
func NewPingEngine(selfID data.NodeID, clock Clock) *PingEngine {
	if clock == nil {
		clock = NewMonotonicClock()
	}
	return &PingEngine{
		selfID: selfID,
		clock:  clock,
		health: data.LinkHealth{
			MinRTTNano: -1,
		},
	}
}

// Reset clears the in-flight ring and resets cumulative telemetry cleanly.
func (p *PingEngine) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.nextSeq = 0
	p.ring = [MaxInFlightPings]inFlightSlot{}
	p.rttTotal = 0
	p.health = data.LinkHealth{MinRTTNano: -1}
}

// End of engine/ping.go
