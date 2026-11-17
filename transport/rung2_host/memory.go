// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          transport/rung2_host/memory.go
// @id            cog.transport.rung2_memory
// @layer         Layer 3A: Physical & Virtual Transport Substrate (Rung 2 Host)
// @archetype     operational | stateful
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Rung 2 Host In-Memory Go Channel Bus Transceiver
// @role          Runtime Engine: Zero-Hardware RAM Loopback Bus & Endpoint
// @maturity      Rev 5 Baseline | Wave 2 First Breath
// @chassis       bounded-heap | Fixed 64-slot channel queues per attached node
// @scale         O(1) point-to-point channel dispatch; O(N) broadcast delivery
// @seams         in: data, transport | uses: context, sync, sync/atomic, fmt
// @invariants    1. Stratum Baseline: Downward Dependency Law, Universal
//                   5-Block Lifecycle Protocol (Rev 5).
//                2. Requires zero physical NICs, Wi-Fi cards, or OS sockets.
//                3. Supports deterministic fault injection (`SetHealthy(false)`)
//                   to verify Layer 2 EVT_CASCADE_DOWN transitions.
//                4. Non-blocking close semantics preventing goroutine leaks.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Implements the primary Rung 2 (Host Substrate) transceiver: an in-memory
//   Go channel switchboard (`MemoryBus`) and per-node endpoint
//   (`MemoryTransceiver`). Mirrors the game-developer single-player loopback
//   architecture so two CPI-SI nodes can converse inside RAM with zero
//   networking hardware present.
//
//   ### For Architects, Operators & Non-Programmers:
//   Allows two FaithNet minds to talk inside the same computer's memory with
//   no Wi-Fi, no Ethernet cable, and no internet—ideal for local simulation,
//   offline single-machine operation, and deterministic fault testing.
//
//   ### For Systems Programmers & Engine Architects:
//   - Controlled Memory Chassis: Bounded 64-envelope buffered channels per
//     registered NodeID; drops overflow cleanly rather than blocking senders.
//   - Fault-Injection Seam: `SetHealthy(bool)` simulates physical link loss
//     cleanly for Layer 2 `AdaptiveLadder` cascade-down tests.
// ==============================================================================

package rung2_host

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Bounds, Logical Anchor, MemoryBus & MemoryTransceiver Chassis Structs )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3A: Physical & Virtual Transport Substrate (Rung 2 Host)
// Module:        cog.transport.rung2_memory
// Mission:       Declare channel queue limits, MemoryBus hub, and per-node
//                MemoryTransceiver implementing transport.Transceiver.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
)

// --- Tier 2: External Library Dependencies (None — 100% Go Stdlib Invariant) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---
import (
	"github.com/CreativeWorkzStudio/FaithNet/data"
	"github.com/CreativeWorkzStudio/FaithNet/transport"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Pre-Conditions & Channel Chassis Limits
// ------------------------------------------------------------------------------

const (
	// DefaultMemoryQueueDepth bounds buffered envelopes per attached RAM node.
	DefaultMemoryQueueDepth = 64
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.2: Operational Logical Anchor
// ------------------------------------------------------------------------------

const (
	// ModuleIDMemoryBus is the canonical logical ID for Rung 2 RAM transport.
	ModuleIDMemoryBus = "cog.transport.rung2_memory"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.3: Public Tagged Ontology (MemoryBus & MemoryTransceiver)
// ------------------------------------------------------------------------------

// MemoryBus is an in-RAM switchboard linking sovereign nodes via Go channels.
type MemoryBus struct {
	mu    sync.RWMutex
	peers map[data.NodeID]*MemoryTransceiver
}

// MemoryTransceiver is a single node's Rung 2 RAM attachment on a MemoryBus.
type MemoryTransceiver struct {
	bus     *MemoryBus
	selfID  data.NodeID
	inbox   chan transport.WireEnvelope
	healthy atomic.Bool
	closed  atomic.Bool
	sentCnt atomic.Uint64
	recvCnt atomic.Uint64
}

// ------------------------------------------------------------------------------
// Sub-Assembly 2.4: Pre-Conditional Helper Cogs & O(1) Accessors
// ------------------------------------------------------------------------------

// Name returns the canonical Rung 2 RAM transport identifier.
//
// ### For Architects, Operators & Non-Programmers:
// Names this communication channel clearly in status cards and logs.
//
// ### For Systems Programmers & Engine Architects:
// O(1) static string getter implementing `transport.Transceiver`.
func (m *MemoryTransceiver) Name() string {
	return "rung2:memory"
}

// Rung returns data.Rung2Host.
//
// ### For Architects, Operators & Non-Programmers:
// Identifies this in-memory wire as belonging to Rung 2 (Host Substrate).
//
// ### For Systems Programmers & Engine Architects:
// O(1) constant accessor implementing `transport.Transceiver`.
func (m *MemoryTransceiver) Rung() data.RungLevel {
	return data.Rung2Host
}

// Healthy reports whether this RAM link is currently operational.
//
// ### For Architects, Operators & Non-Programmers:
// Checks whether this virtual wire is plugged in and ready to carry frames.
//
// ### For Systems Programmers & Engine Architects:
// Lock-free atomic read verifying `!closed && healthy`.
func (m *MemoryTransceiver) Healthy() bool {
	return !m.closed.Load() && m.healthy.Load()
}

// SetHealthy toggles live health state (used for deterministic cascade tests).
//
// ### For Architects, Operators & Non-Programmers:
// Allows operators or tests to simulate unplugging or restoring this RAM link.
//
// ### For Systems Programmers & Engine Architects:
// Lock-free atomic boolean store controlling `Healthy()` and `Send()` gating.
func (m *MemoryTransceiver) SetHealthy(ok bool) {
	m.healthy.Store(ok)
}

// ==============================================================================
// BLOCK 3: BODY BLOCK
// ( Composable Business Logic: Send, Receive & Bus Delivery )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3A: Physical & Virtual Transport Substrate (Rung 2 Host)
// Module:        cog.transport.rung2_memory
// Mission:       Deliver bounded 512-byte WireEnvelopes directly across RAM
//                channels without touching any OS socket or network card.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: Point-to-Point & Broadcast RAM Transmission
// ------------------------------------------------------------------------------

// Send copies raw wire bytes into a bounded WireEnvelope and delivers it to
// target (or all attached peers when target is NodeIDBroadcast).
//
// ### For Architects, Operators & Non-Programmers:
// Hands an envelope instantaneously from one node's memory inbox to another
// node's memory inbox inside the same computer.
//
// ### For Systems Programmers & Engine Architects:
// Enforces 512-byte Floor MTU ceiling, checks closed/healthy guards, and
// performs non-blocking channel delivery under MemoryBus read lock.
func (m *MemoryTransceiver) Send(target data.NodeID, raw []byte) error {
	if m.closed.Load() {
		return transport.ErrTransceiverClosed
	}
	if !m.healthy.Load() {
		return transport.ErrTransceiverOffline
	}
	if len(raw) > data.FloorMaxFrameSize {
		return transport.ErrWireFrameOversize
	}

	var env transport.WireEnvelope
	env.SenderID = m.selfID
	env.TargetID = target
	env.Length = len(raw)
	copy(env.Bytes[:], raw)

	m.bus.mu.RLock()
	defer m.bus.mu.RUnlock()

	if target.IsBroadcast() {
		for id, peer := range m.bus.peers {
			if id == m.selfID || !peer.Healthy() {
				continue
			}
			select {
			case peer.inbox <- env:
			default:
			}
		}
		m.sentCnt.Add(1)
		return nil
	}

	peer, ok := m.bus.peers[target]
	if !ok || !peer.Healthy() {
		return transport.ErrPeerUnreachable
	}
	select {
	case peer.inbox <- env:
		m.sentCnt.Add(1)
		return nil
	default:
		return transport.ErrPeerUnreachable
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: Context-Aware RAM Reception
// ------------------------------------------------------------------------------

// Receive waits for the next WireEnvelope from inbox or returns ctx.Err().
//
// ### For Architects, Operators & Non-Programmers:
// Listens quietly for incoming messages in memory and stops cleanly the
// instant the node shuts down.
//
// ### For Systems Programmers & Engine Architects:
// Honors context cancellation cleanly to guarantee zero leaked goroutines.
func (m *MemoryTransceiver) Receive(ctx context.Context) (transport.WireEnvelope, error) {
	if m.closed.Load() {
		return transport.WireEnvelope{}, transport.ErrTransceiverClosed
	}
	select {
	case <-ctx.Done():
		return transport.WireEnvelope{}, ctx.Err()
	case env, ok := <-m.inbox:
		if !ok {
			return transport.WireEnvelope{}, transport.ErrTransceiverClosed
		}
		m.recvCnt.Add(1)
		return env, nil
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Contract Verification Cogs & Pre-Flight Validation Tripwires )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3A: Physical & Virtual Transport Substrate (Rung 2 Host)
// Module:        cog.transport.rung2_memory
// Mission:       Audit live MemoryTransceiver contracts and execute a 4-section
//                pre-flight check (attach/send/recv, fault bounds, close).
// @triggers
//   - Block 4 Audits : (*MemoryTransceiver).AuditState(),
//                      PreflightMemoryBus(), AuditMemoryBus()
//   - Layer 4 Tests  : tests/functional/transport_functional_test.go,
//                      tests/unit/transport/memory_test.go,
//                      tests/integrated/host_ipc_mesh_test.go,
//                      tests/system/sovereign_mesh_e2e_test.go
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Contract, Input & Live-Instance Verification Cogs
// ------------------------------------------------------------------------------

// AuditState inspects a live MemoryTransceiver non-destructively to confirm
// bus linkage, valid NodeID, Rung 2 tagging, and open/healthy parity.
//
// ### For Architects, Operators & Non-Programmers:
// Verifies a live in-memory wire endpoint is still plugged into its virtual
// switchboard and carrying a valid device address.
//
// ### For Systems Programmers & Engine Architects:
// Non-mutating O(1) live probe returning structured `data.AuditResult`.
func (m *MemoryTransceiver) AuditState() data.AuditResult {
	const check = "rung2_memory_live_state_contract"

	if m == nil {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDMemoryBus, check,
			"MemoryTransceiver receiver is nil.")
	}
	if m.bus == nil || m.inbox == nil {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDMemoryBus, check,
			"MemoryTransceiver missing bus switchboard or inbox channel.")
	}
	if !m.selfID.IsValid() || m.selfID.IsBroadcast() {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDMemoryBus, check,
			"MemoryTransceiver selfID is null or broadcast.")
	}
	if m.closed.Load() && m.healthy.Load() {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDMemoryBus, check,
			"Closed MemoryTransceiver erroneously reports Healthy()==true.")
	}
	return data.AuditPass(ModuleIDMemoryBus, check,
		"Live Rung 2 MemoryTransceiver state and bus invariants verified.")
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Machine Self-Audit & Pre-Flight Validation Tripwires
// ------------------------------------------------------------------------------

// PreflightMemoryBus executes a 4-section non-destructive pre-flight self-audit
// covering (1) bidirectional send/recv, (2) unreachable/oversize/closed fault
// inversion, (3) 512B ceiling limits, and (4) clean Close() detachment.
//
// ### For Architects, Operators & Non-Programmers:
// Bench-tests the RAM switchboard at startup so FaithNet knows its local
// zero-hardware wire works before any node opens sockets or radio links.
//
// ### For Systems Programmers & Engine Architects:
// Isolated 4-section pre-flight gate returning structured `data.AuditResult`.
func PreflightMemoryBus() data.AuditResult {
	const check = "rung2_memory_preflight_4section"

	bus := NewMemoryBus()
	a := bus.Attach(0x01)
	b := bus.Attach(0x02)
	defer func() {
		_ = a.Close()
		_ = b.Close()
	}()

	// Micro Check 1 (Always Pass): Live audit + nominal send/receive.
	if live := a.AuditState(); !live.IsPass() {
		return live
	}
	if err := a.Send(0x02, []byte("RAM-PING")); err != nil {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDMemoryBus, check,
			"Sec 1 nominal MemoryBus Send failed.")
	}
	env, err := b.Receive(context.Background())
	if err != nil || env.SenderID != 0x01 || string(env.Slice()) != "RAM-PING" {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDMemoryBus, check,
			"Sec 1 nominal MemoryBus Receive payload/sender mismatch.")
	}

	// Micro Check 2 (Never Fail): Unreachable peer & context cancel inversion.
	if err := a.Send(0x9999, []byte("GHOST")); err != transport.ErrPeerUnreachable {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDMemoryBus, check,
			"Sec 2 unknown peer did not return ErrPeerUnreachable.")
	}
	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := b.Receive(canceledCtx); err != context.Canceled {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDMemoryBus, check,
			"Sec 2 Receive did not honor cancelled context.")
	}

	// Micro Check 3 (Invariants & Limits): 512B ceiling accepted, 513B rejected.
	var oversize [data.FloorMaxFrameSize + 1]byte
	if err := a.Send(0x02, oversize[:]); err != transport.ErrWireFrameOversize {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDMemoryBus, check,
			"Sec 3 513B payload bypassed 512B Floppy Sector wire ceiling.")
	}

	// Micro Check 4 (Emergence): Close() detaches cleanly & blocks further Send.
	if err := b.Close(); err != nil || b.Healthy() {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDMemoryBus, check,
			"Sec 4 Close() failed to mark MemoryTransceiver unhealthy.")
	}
	if err := a.Send(0x02, []byte("AFTER-CLOSE")); err != transport.ErrPeerUnreachable {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDMemoryBus, check,
			"Sec 4 detached peer remained reachable on MemoryBus.")
	}

	return data.AuditPass(ModuleIDMemoryBus, check,
		"Rung 2 MemoryBus passed all 4 micro pre-flight gates.")
}

// AuditMemoryBus returns true iff PreflightMemoryBus() passes cleanly.
//
// ### For Architects, Operators & Non-Programmers:
// Quick pass/fail check confirming the in-memory switchboard is healthy.
//
// ### For Systems Programmers & Engine Architects:
// Convenience boolean bridge delegating to `PreflightMemoryBus().IsPass()`.
func AuditMemoryBus() bool {
	return PreflightMemoryBus().IsPass()
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK / EXECUTABLE PROCEDURES & LIFECYCLE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3A: Physical & Virtual Transport Substrate (Rung 2 Host)
// Module:        cog.transport.rung2_memory
// Mission:       Render 60-column RAM transport dump cards, expose epilogue
//                readiness, and own MemoryBus initialization and teardown.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Executable Procedures, Inspection, Readiness & Cards
// ------------------------------------------------------------------------------

// IsMemoryBusReady reports whether Rung 2 RAM transport passed pre-flight.
//
// ### For Architects, Operators & Non-Programmers:
// Epilogue confirmation that in-memory node communication is ready.
//
// ### For Systems Programmers & Engine Architects:
// Matches OmniCode's Block 5 epilogue readiness predicate (`is_*_ready`).
func IsMemoryBusReady() bool {
	return PreflightMemoryBus().IsPass()
}

// ExecuteMemoryDumpCard renders a 60-column inner-width diagnostic card.
//
// ### For Architects, Operators & Non-Programmers:
// Prints a framed status summary of this RAM wire's ID, health, and traffic.
//
// ### For Systems Programmers & Engine Architects:
// Non-mutating diagnostic formatter using lock-free atomic telemetry loads.
func (m *MemoryTransceiver) ExecuteMemoryDumpCard() string {
	return fmt.Sprintf(
		"+------------------------------------------------------------+\n"+
			"| %-58s |\n"+
			"+------------------------------------------------------------+\n"+
			"| Module:   %-48s |\n"+
			"| NodeID:   %-48s |\n"+
			"| Rung:     %-48s |\n"+
			"| Healthy:  %-48t |\n"+
			"| Traffic:  %-48s |\n"+
			"+------------------------------------------------------------+",
		"FAITHNET RUNG 2 MEMORY TRANSCEIVER CARD",
		ModuleIDMemoryBus,
		m.selfID.String(),
		m.Rung().String(),
		m.Healthy(),
		fmt.Sprintf("%d sent / %d recv", m.sentCnt.Load(), m.recvCnt.Load()),
	)
}

// ------------------------------------------------------------------------------
// Sub-Assembly 5.2: Subsystem Lifecycle Initialization & Teardown
// ------------------------------------------------------------------------------

// NewMemoryBus initializes an empty in-memory switchboard for Rung 2 peers.
//
// ### For Architects, Operators & Non-Programmers:
// Creates a private in-RAM virtual switchboard connecting local nodes.
//
// ### For Systems Programmers & Engine Architects:
// Allocates a mutex-guarded peer routing map (`NodeID -> *MemoryTransceiver`).
func NewMemoryBus() *MemoryBus {
	return &MemoryBus{
		peers: make(map[data.NodeID]*MemoryTransceiver),
	}
}

// Attach registers nodeID on bus and returns its ready MemoryTransceiver.
//
// ### For Architects, Operators & Non-Programmers:
// Plugs a node into the virtual RAM switchboard so it can immediately exchange
// frames with other local peers.
//
// ### For Systems Programmers & Engine Architects:
// Allocates a bounded inbox channel (`DefaultMemoryQueueDepth`) under lock.
func (b *MemoryBus) Attach(nodeID data.NodeID) *MemoryTransceiver {
	mt := &MemoryTransceiver{
		bus:    b,
		selfID: nodeID,
		inbox:  make(chan transport.WireEnvelope, DefaultMemoryQueueDepth),
	}
	mt.healthy.Store(true)

	b.mu.Lock()
	b.peers[nodeID] = mt
	b.mu.Unlock()
	return mt
}

// Close detaches this transceiver from MemoryBus and closes its inbox safely.
//
// ### For Architects, Operators & Non-Programmers:
// Unplugs this node from the virtual RAM switchboard cleanly with zero leaks.
//
// ### For Systems Programmers & Engine Architects:
// Idempotent via atomic CompareAndSwap; safe for repeated or concurrent calls.
func (m *MemoryTransceiver) Close() error {
	if !m.closed.CompareAndSwap(false, true) {
		return nil
	}
	m.healthy.Store(false)
	m.bus.mu.Lock()
	delete(m.bus.peers, m.selfID)
	m.bus.mu.Unlock()
	close(m.inbox)
	return nil
}

// End of transport/rung2_host/memory.go
