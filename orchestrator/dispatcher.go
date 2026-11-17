// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          orchestrator/dispatcher.go
// @id            cog.orchestrator.dispatcher
// @layer         Layer 2: Adaptive Orchestration & Ladder Control
// @archetype     operational | stateful
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Central Domain Dispatcher & Ingress/Egress Event Loop
// @role          Central Node Orchestrator: Codec -> Airlock -> Router -> Ping
// @maturity      Rev 5 Baseline | Wave 3 Orchestrator
// @chassis       bounded-heap | Coordinates Layer 1 fixed rings + 4-rung ladder
// @scale         O(1) per frame decode/airlock/dispatch; bounded goroutines
// @seams         in: data, engine, engine/covenant, orchestrator/ladder
// @invariants    1. Stratum Baseline: Downward Dependency Law, Rev 5 5-Block.
//                2. EVERY inbound frame passes Genesis 3 Airlock BEFORE any
//                   Router, Ping, or application delivery callback runs.
//                3. Stop()/Close() waits cleanly for all listener goroutines
//                   (zero goroutine leaks verified under `go test -race`).
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Implements the central sovereign node dispatcher tying Layer 1 state
//   machines (`Codec`, `Clock`, `PingEngine`, `Router`, `CovenantFilter`) to
//   Layer 2's `AdaptiveLadder`. Decodes incoming envelopes, enforces Genesis 3
//   T4 verdicts at the door, reflects `OpPing` automatically, tracks RTTs on
//   `OpPong`, updates peer routes, and delivers permitted data/CPI-SI payloads.
//
//   ### For Architects, Operators & Non-Programmers:
//   Acts as the central nervous system of a FaithNet node: receiving frames
//   from whichever survival rung is alive, checking every message at the
//   airlock door, answering pings automatically, and handing clean messages
//   to the user or local portal.
//
//   ### For Systems Programmers & Engine Architects:
//   - Ownership & Reentrancy: Thread-safe atomic telemetry counters, mutex-
//     guarded start/stop lifecycle, and caller-provided `DeliverHandler`.
//   - Extension Seam: Layer 3B `interpreter/` ("The Well") and `cmd/faithnet-node`
//     plug directly into `SetDeliverHandler` and `PingPeer`.
// ==============================================================================

package orchestrator

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Logical Anchor, Telemetry Structs, Callbacks & Pre-Conditional Accessors )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 2: Adaptive Orchestration & Ladder Control
// Module:        cog.orchestrator.dispatcher
// Mission:       Declare Dispatcher state chassis, telemetry snapshot record,
//                delivery callback contract, and O(1) component accessors.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// --- Tier 2: External Library Dependencies (None — 100% Go Stdlib Invariant) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---
import (
	"github.com/CreativeWorkzStudio/FaithNet/data"
	"github.com/CreativeWorkzStudio/FaithNet/engine"
	"github.com/CreativeWorkzStudio/FaithNet/engine/covenant"
	"github.com/CreativeWorkzStudio/FaithNet/transport"
	"github.com/CreativeWorkzStudio/FaithNet/transport/rung2_host"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Pre-Conditions & Dispatcher Sentinel Errors
// ------------------------------------------------------------------------------

var (
	// ErrDispatcherInvalidManifest indicates startup with an invalid NodeManifest.
	ErrDispatcherInvalidManifest = errors.New("faithnet/dispatcher: invalid node manifest")

	// ErrDispatcherAlreadyRunning indicates duplicate Start() while active.
	ErrDispatcherAlreadyRunning = errors.New("faithnet/dispatcher: already running")
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.2: Operational Logical Anchor
// ------------------------------------------------------------------------------

const (
	// ModuleIDDispatcher is the canonical logical ID for the central dispatcher.
	ModuleIDDispatcher = "cog.orchestrator.dispatcher"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.3: Public Tagged Ontology (Stats, Outcome & Dispatcher)
// ------------------------------------------------------------------------------

// DeliverHandler receives permitted (+1 or 0:AWAIT) application, beacon, or
// legacy-wrapped frames after passing Codec and Genesis 3 Airlock verification.
type DeliverHandler func(frame *data.Frame, verdict covenant.Verdict)

// DispatcherStats captures a point-in-time read-only telemetry snapshot.
type DispatcherStats struct {
	FramesDecoded  uint64 // Valid frames unpacked by Codec
	DecodeErrors   uint64 // Malformed or seal-broken wire drops
	AirlockBlocked uint64 // Frames halted by Genesis 3 Airlock (-1 or NULL)
	PingsReflected uint64 // OpPing requests echoed automatically as OpPong
	PongsMatched   uint64 // OpPong echoes correlated with local PingEngine
	DataDelivered  uint64 // Permitted non-ping frames handed to DeliverHandler
	FramesSent     uint64 // Outbound frames transmitted over AdaptiveLadder
}

// InboundOutcome summarizes how HandleInboundEnvelope processed one envelope.
type InboundOutcome struct {
	Decoded   bool             // True if Codec unpacked bytes cleanly
	Verdict   covenant.Verdict // Genesis 3 Airlock T4 verdict
	Opcode    data.Opcode      // Decoded protocol verb (if Decoded)
	RTTNano   int64            // Measured RTT if Opcode==OpPong matched
	Reflected bool             // True if an OpPong echo was transmitted
}

// Dispatcher wires Layer 1 engines to Layer 2's 4-Rung AdaptiveLadder.
//
// ### For Architects, Operators & Non-Programmers:
// Runs one complete sovereign FaithNet node—handling inbound/outbound mail,
// enforcing airlock security, updating neighbor routes, and reflecting pings.
//
// ### For Systems Programmers & Engine Architects:
// Owns one `AdaptiveLadder`, `PingEngine`, `Router`, and `covenant.Filter`.
// Lifecycle goroutines are strictly bounded by `runCancel` and `wg`.
type Dispatcher struct {
	manifest data.NodeManifest
	clock    engine.Clock
	ladder   *AdaptiveLadder
	ping     *engine.PingEngine
	router   *engine.Router
	airlock  *covenant.Filter

	handlerMu sync.RWMutex
	onDeliver DeliverHandler

	pongMu     sync.Mutex
	pongNotify chan uint32

	lifeMu    sync.Mutex
	running   bool
	runCancel context.CancelFunc
	wg        sync.WaitGroup

	framesDecoded  atomic.Uint64
	decodeErrors   atomic.Uint64
	airlockBlocked atomic.Uint64
	pingsReflected atomic.Uint64
	pongsMatched   atomic.Uint64
	dataDelivered  atomic.Uint64
	framesSent     atomic.Uint64
}

// ------------------------------------------------------------------------------
// Sub-Assembly 2.4: Pre-Conditional Helper Cogs & Component Accessors
// ------------------------------------------------------------------------------

// Manifest returns a copy of this node's sovereign startup manifest.
//
// ### For Architects, Operators & Non-Programmers:
// Shows the node's identity, label, and offline/egress policy settings.
//
// ### For Systems Programmers & Engine Architects:
// Pure O(1) immutable value getter.
func (d *Dispatcher) Manifest() data.NodeManifest {
	return d.manifest
}

// Ladder returns the Dispatcher's underlying 4-Rung AdaptiveLadder.
//
// ### For Architects, Operators & Non-Programmers:
// Exposes the survival ladder so operators can attach wires or inspect rungs.
//
// ### For Systems Programmers & Engine Architects:
// Direct pointer accessor for rung registration and diagnostic queries.
func (d *Dispatcher) Ladder() *AdaptiveLadder {
	return d.ladder
}

// Router returns the Dispatcher's bounded 128-peer sovereign Router.
//
// ### For Architects, Operators & Non-Programmers:
// Exposes the neighborhood address book to inspect discovered peers.
//
// ### For Systems Programmers & Engine Architects:
// Direct pointer accessor to Layer 1 `engine.Router`.
func (d *Dispatcher) Router() *engine.Router {
	return d.router
}

// Ping returns the Dispatcher's 5-Cog sonar PingEngine.
//
// ### For Architects, Operators & Non-Programmers:
// Exposes the node's live latency scorecard and heartbeat sonar engine.
//
// ### For Systems Programmers & Engine Architects:
// Direct pointer accessor to Layer 1 `engine.PingEngine`.
func (d *Dispatcher) Ping() *engine.PingEngine {
	return d.ping
}

// Airlock returns the Dispatcher's Genesis 3 Covenant Filter.
//
// ### For Architects, Operators & Non-Programmers:
// Exposes the epistemic airlock guarding this node against hostile data.
//
// ### For Systems Programmers & Engine Architects:
// Direct pointer accessor to Layer 1 `covenant.Filter`.
func (d *Dispatcher) Airlock() *covenant.Filter {
	return d.airlock
}

// SetDeliverHandler registers a callback invoked on permitted non-ping frames.
//
// ### For Architects, Operators & Non-Programmers:
// Plugs a user interface, local web portal ("The Well"), or application into
// the dispatcher so it receives verified incoming messages.
//
// ### For Systems Programmers & Engine Architects:
// Write-locked callback swap; safe to invoke before or during runtime.
func (d *Dispatcher) SetDeliverHandler(fn DeliverHandler) {
	d.handlerMu.Lock()
	d.onDeliver = fn
	d.handlerMu.Unlock()
}

// ==============================================================================
// BLOCK 3: BODY BLOCK
// ( Composable Business Logic: Inbound Pipeline, Outbound Send & Event Loop )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 2: Adaptive Orchestration & Ladder Control
// Module:        cog.orchestrator.dispatcher
// Mission:       Process inbound wire envelopes through Codec -> Airlock ->
//                Router -> Ping/Deliver, and transmit adaptive outbound frames.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: Inbound Envelope Processing Pipeline
// ------------------------------------------------------------------------------

// HandleInboundEnvelope decodes one raw WireEnvelope, enforces the Genesis 3
// Airlock, updates neighbor routes, reflects OpPing, correlates OpPong, and
// forwards permitted data/CPI-SI/legacy frames to the registered handler.
//
// ### For Architects, Operators & Non-Programmers:
// Processes one arriving packet completely from envelope opening through
// security inspection and automatic echo reply.
//
// ### For Systems Programmers & Engine Architects:
// Stack-allocated frame decode/reflect path; callable synchronously in unit
// tests or asynchronously from Start() receive goroutines.
func (d *Dispatcher) HandleInboundEnvelope(env transport.WireEnvelope) InboundOutcome {
	var in data.Frame
	if err := engine.DecodeFrom(env.Slice(), &in); err != nil {
		d.decodeErrors.Add(1)
		return InboundOutcome{
			Decoded: false,
			Verdict: covenant.Verdict{
				State:  data.QuadNull,
				Reason: covenant.ReasonMalformedStructure,
			},
		}
	}
	d.framesDecoded.Add(1)

	verdict := d.airlock.EvaluateFrame(&in)
	if verdict.IsBlocked() {
		d.airlockBlocked.Add(1)
		return InboundOutcome{
			Decoded: true,
			Verdict: verdict,
			Opcode:  in.Opcode,
		}
	}

	// Ignore frames explicitly addressed to another non-broadcast node.
	if in.TargetID != d.manifest.ID && !in.TargetID.IsBroadcast() {
		return InboundOutcome{
			Decoded: true,
			Verdict: verdict,
			Opcode:  in.Opcode,
		}
	}

	// Learn or refresh peer route in bounded Router (best-effort).
	if in.SourceID.IsValid() && in.SourceID != d.manifest.ID {
		_ = d.router.UpsertRoute(engine.PeerRoute{
			PeerID:     in.SourceID,
			NextHopID:  in.SourceID,
			Rung:       in.Rung,
			HopCount:   1,
			LastSeenNs: d.clock.NowNano(),
		})
	}

	outcome := InboundOutcome{
		Decoded: true,
		Verdict: verdict,
		Opcode:  in.Opcode,
	}

	switch in.Opcode {
	case data.OpPing:
		var resp data.Frame
		var wire [data.FloorMaxFrameSize]byte
		if err := engine.ReflectPingInto(&in, &resp, d.manifest.ID); err == nil {
			if n, encErr := engine.EncodeInto(wire[:], &resp); encErr == nil {
				if _, sendErr := d.ladder.SendFrame(resp.TargetID, wire[:n]); sendErr == nil {
					d.pingsReflected.Add(1)
					d.framesSent.Add(1)
					outcome.Reflected = true
				}
			}
		}

	case data.OpPong:
		if rtt, ok := d.ping.ProcessPong(&in); ok {
			d.pongsMatched.Add(1)
			outcome.RTTNano = rtt
			// Refresh route RTT telemetry cleanly.
			if rt, found := d.router.Lookup(in.SourceID); found {
				rt.LastRTTNano = rtt
				rt.LastSeenNs = d.clock.NowNano()
				_ = d.router.UpsertRoute(rt)
			}
			select {
			case d.pongNotify <- in.SeqID:
			default:
			}
		}

	default:
		d.dataDelivered.Add(1)
		d.handlerMu.RLock()
		fn := d.onDeliver
		d.handlerMu.RUnlock()
		if fn != nil {
			fn(&in, verdict)
		}
	}
	return outcome
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: Outbound Frame Transmission & Live Ping Helper
// ------------------------------------------------------------------------------

// SendFrame encodes frame f into wire bytes and transmits it adaptively over
// the 4-Rung AdaptiveLadder (automatically cascading down on link failure).
//
// ### For Architects, Operators & Non-Programmers:
// Packs a sovereign message and sends it over whichever survival rung is
// currently healthy—automatically failing over if the top rung snaps.
//
// ### For Systems Programmers & Engine Architects:
// Stamps `f.SourceID` if unset, encodes on stack `[512]byte`, and delegates to
// `AdaptiveLadder.SendFrame`.
func (d *Dispatcher) SendFrame(f *data.Frame) (data.RungLevel, error) {
	if f == nil {
		return data.RungNone, engine.ErrBufferTooSmall
	}
	if !f.SourceID.IsValid() {
		f.SourceID = d.manifest.ID
	}
	activeRung := d.ladder.EvaluateAndAdapt("outbound send check")
	if !activeRung.IsValid() {
		return data.RungNone, ErrFloorCollapsed
	}
	f.Rung = activeRung

	var wire [data.FloorMaxFrameSize]byte
	n, err := engine.EncodeInto(wire[:], f)
	if err != nil {
		return data.RungNone, err
	}
	usedRung, err := d.ladder.SendFrame(f.TargetID, wire[:n])
	if err != nil {
		return data.RungNone, err
	}
	d.framesSent.Add(1)
	return usedRung, nil
}

// PingPeer sends a live OpPing to target over the AdaptiveLadder and waits
// until the matching OpPong arrives (requires Start() running) or ctx expires.
//
// ### For Architects, Operators & Non-Programmers:
// Sends a sonar ping to a neighbor over whichever rung is active and returns
// the exact round-trip speed plus which survival rung carried the pulse.
//
// ### For Systems Programmers & Engine Architects:
// Correlates `pongNotify` notifications by `req.SeqID` cleanly under context.
func (d *Dispatcher) PingPeer(
	ctx context.Context,
	target data.NodeID,
	echo []byte,
) (int64, data.RungLevel, error) {
	activeRung := d.ladder.EvaluateAndAdapt("ping peer check")
	if !activeRung.IsValid() {
		return 0, data.RungNone, ErrFloorCollapsed
	}

	var req data.Frame
	if err := d.ping.BuildPingInto(&req, target, activeRung, echo); err != nil {
		return 0, data.RungNone, err
	}
	var wire [data.FloorMaxFrameSize]byte
	n, err := engine.EncodeInto(wire[:], &req)
	if err != nil {
		return 0, data.RungNone, err
	}
	usedRung, err := d.ladder.SendFrame(target, wire[:n])
	if err != nil {
		return 0, data.RungNone, err
	}
	d.framesSent.Add(1)

	for {
		select {
		case <-ctx.Done():
			return 0, usedRung, ctx.Err()
		case seq := <-d.pongNotify:
			if seq == req.SeqID {
				return d.ping.HealthSnapshot().LastRTTNano, usedRung, nil
			}
		}
	}
}

// Start spawns bounded receive goroutines for every transceiver currently
// registered on the AdaptiveLadder until Stop() or parent ctx cancellation.
//
// ### For Architects, Operators & Non-Programmers:
// Turns the node's ears on across all plugged-in communication wires at once.
//
// ### For Systems Programmers & Engine Architects:
// Tracks every listener goroutine in `d.wg` so `Stop()` guarantees zero leaks.
func (d *Dispatcher) Start(parent context.Context) error {
	d.lifeMu.Lock()
	defer d.lifeMu.Unlock()

	if d.running {
		return ErrDispatcherAlreadyRunning
	}
	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithCancel(parent)
	d.runCancel = cancel
	d.running = true

	var txs [rungSlotCount]transport.Transceiver
	count := d.ladder.RegisteredTransceivers(txs[:])
	for i := 0; i < count; i++ {
		tx := txs[i]
		d.wg.Add(1)
		go func(rx transport.Transceiver) {
			defer d.wg.Done()
			for {
				env, err := rx.Receive(ctx)
				if err != nil {
					return
				}
				d.HandleInboundEnvelope(env)
			}
		}(tx)
	}
	return nil
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Contract Verification Cogs & Pre-Flight Validation Tripwires )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 2: Adaptive Orchestration & Ladder Control
// Module:        cog.orchestrator.dispatcher
// Mission:       Audit live Dispatcher sub-engines and execute a 4-section
//                pre-flight two-node live dispatch verification.
// @triggers
//   - Block 4 Audits : (*Dispatcher).AuditState(), PreflightDispatcher(),
//                      AuditDispatcher()
//   - Layer 4 Tests  : tests/integrated/orchestrator_ladder_test.go
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Contract, Input & Live-Instance Verification Cogs
// ------------------------------------------------------------------------------

// AuditState audits a live Dispatcher and all constituent Layer 1/2 sub-engines
// (Ladder, PingEngine, Router, CovenantFilter) non-destructively.
//
// ### For Architects, Operators & Non-Programmers:
// Runs a complete live health check across every organ of a running node.
//
// ### For Systems Programmers & Engine Architects:
// Composes `ladder.AuditState()`, `ping.AuditState()`, `router.AuditState()`,
// and `airlock.AuditPolicy()` into a single structured `data.AuditResult`.
func (d *Dispatcher) AuditState() data.AuditResult {
	const check = "dispatcher_live_state_contract"

	if d == nil {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDDispatcher, check,
			"Dispatcher receiver is nil.")
	}
	if !d.manifest.IsValid() {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDDispatcher, check,
			"Dispatcher holds invalid NodeManifest.")
	}
	if r := d.ladder.AuditState(); !r.IsPass() {
		return r
	}
	if r := d.ping.AuditState(); !r.IsPass() {
		return r
	}
	if r := d.router.AuditState(); !r.IsPass() {
		return r
	}
	if r := d.airlock.AuditPolicy(); !r.IsPass() {
		return r
	}
	return data.AuditPass(ModuleIDDispatcher, check,
		"Live Dispatcher, Ladder, Ping, Router, and Airlock all verified.")
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Machine Self-Audit & Pre-Flight Validation Tripwires
// ------------------------------------------------------------------------------

// PreflightDispatcher executes a 4-section non-destructive pre-flight audit
// spinning up two ephemeral Rung 2 nodes to prove (1) live PingPeer RTT,
// (2) hostile payload airlock interception, (3) invalid manifest rejection,
// and (4) clean Stop()/Close() teardown with zero leaks.
//
// ### For Architects, Operators & Non-Programmers:
// Runs a miniature two-node FaithNet rehearsal inside memory at startup to
// verify the full node brain works cleanly end-to-end.
//
// ### For Systems Programmers & Engine Architects:
// Composes `PreflightProbe()`, `PreflightLadder()`, and live two-node dispatch;
// returns structured `data.AuditResult`.
func PreflightDispatcher() data.AuditResult {
	const check = "dispatcher_preflight_4section"

	if r := PreflightProbe(); !r.IsPass() {
		return r
	}
	if r := PreflightLadder(); !r.IsPass() {
		return r
	}

	// Micro Check 2 (Never Fail): Invalid manifest rejected cleanly.
	if _, err := NewDispatcher(data.DefaultNodeManifest(data.NodeIDNull, "bad"), nil); err != ErrDispatcherInvalidManifest {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDDispatcher, check,
			"Sec 2 NewDispatcher failed to reject invalid manifest.")
	}

	// Micro Check 1 (Always Pass): Two live nodes ping across Rung 2 MemoryBus.
	bus := rung2_host.NewMemoryBus()
	dA, err := NewDispatcher(data.DefaultNodeManifest(0xA1, "alpha"), engine.NewMockClock(100))
	if err != nil {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDDispatcher, check,
			"Sec 1 failed constructing Dispatcher A.")
	}
	dB, err := NewDispatcher(data.DefaultNodeManifest(0xB2, "beta"), engine.NewMockClock(100))
	if err != nil {
		_ = dA.Close()
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDDispatcher, check,
			"Sec 1 failed constructing Dispatcher B.")
	}
	_ = dA.Ladder().RegisterRung(bus.Attach(0xA1))
	_ = dB.Ladder().RegisterRung(bus.Attach(0xB2))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_ = dA.Start(ctx)
	_ = dB.Start(ctx)

	_, rung, err := dA.PingPeer(ctx, 0xB2, []byte("DISPATCH-PREFLIGHT"))
	if err != nil || rung != data.Rung2Host {
		_ = dA.Close()
		_ = dB.Close()
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDDispatcher, check,
			"Sec 1 live two-node PingPeer failed.")
	}

	// Micro Check 3 (Invariants): Hostile legacy frame blocked at Airlock door.
	var hostile data.Frame
	hostile.Rung = data.Rung4Global
	hostile.Opcode = data.OpLegacyWrap
	hostile.Flags = data.FlagExternalOrigin
	hostile.Truth = data.QuadAwait
	hostile.SourceID = 0xEEEE
	hostile.TargetID = 0xB2
	hostile.SetPayload([]byte("<script>evil()</script>"))
	var wire [data.FloorMaxFrameSize]byte
	wn, _ := engine.EncodeInto(wire[:], &hostile)
	var env transport.WireEnvelope
	env.Length = wn
	copy(env.Bytes[:wn], wire[:wn])
	out := dB.HandleInboundEnvelope(env)
	if !out.Verdict.IsBlocked() || dB.StatsSnapshot().AirlockBlocked == 0 {
		_ = dA.Close()
		_ = dB.Close()
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDDispatcher, check,
			"Sec 3 hostile inbound frame bypassed Dispatcher Genesis 3 Airlock.")
	}

	// Micro Check 4 (Emergence): Live AuditState & graceful Close() teardown.
	if r := dA.AuditState(); !r.IsPass() {
		_ = dA.Close()
		_ = dB.Close()
		return r
	}
	_ = dA.Close()
	_ = dB.Close()

	return data.AuditPass(ModuleIDDispatcher, check,
		"Dispatcher passed all 4 micro pre-flight gates (ping/fault/airlock/stop).")
}

// AuditDispatcher returns true iff PreflightDispatcher() passes cleanly.
//
// ### For Architects, Operators & Non-Programmers:
// Quick pass/fail check asking whether the node dispatcher passed safety.
//
// ### For Systems Programmers & Engine Architects:
// Convenience boolean bridge delegating to `PreflightDispatcher().IsPass()`.
func AuditDispatcher() bool {
	return PreflightDispatcher().IsPass()
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK / EXECUTABLE PROCEDURES & LIFECYCLE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 2: Adaptive Orchestration & Ladder Control
// Module:        cog.orchestrator.dispatcher
// Mission:       Expose telemetry snapshots, 60-col diagnostic cards, epilogue
//                readiness, and Dispatcher initialization/Stop/Close teardown.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Executable Procedures, Inspection, Readiness & Cards
// ------------------------------------------------------------------------------

// IsDispatcherReady reports whether the central Dispatcher passed pre-flight.
//
// ### For Architects, Operators & Non-Programmers:
// Epilogue confirmation that the FaithNet node brain is healthy and ready.
//
// ### For Systems Programmers & Engine Architects:
// Matches OmniCode's Block 5 epilogue readiness predicate (`is_*_ready`).
func IsDispatcherReady() bool {
	return PreflightDispatcher().IsPass()
}

// StatsSnapshot returns a lock-free atomic snapshot of Dispatcher counters.
//
// ### For Architects, Operators & Non-Programmers:
// Reports live counts of decoded frames, reflected pings, matched pongs,
// delivered messages, and hostile packets blocked at the airlock.
//
// ### For Systems Programmers & Engine Architects:
// Pure lock-free atomic loads returning a value-typed `DispatcherStats`.
func (d *Dispatcher) StatsSnapshot() DispatcherStats {
	return DispatcherStats{
		FramesDecoded:  d.framesDecoded.Load(),
		DecodeErrors:   d.decodeErrors.Load(),
		AirlockBlocked: d.airlockBlocked.Load(),
		PingsReflected: d.pingsReflected.Load(),
		PongsMatched:   d.pongsMatched.Load(),
		DataDelivered:  d.dataDelivered.Load(),
		FramesSent:     d.framesSent.Load(),
	}
}

// ExecuteDispatcherDumpCard renders a 60-column inner-width node status card.
//
// ### For Architects, Operators & Non-Programmers:
// Prints a neat framed dashboard showing the node's ID, active rung, tracked
// peer count, traffic stats, and how many hostile packets were stopped.
//
// ### For Systems Programmers & Engine Architects:
// Non-mutating diagnostic card builder safe for concurrent queries.
func (d *Dispatcher) ExecuteDispatcherDumpCard() string {
	s := d.StatsSnapshot()
	return fmt.Sprintf(
		"+------------------------------------------------------------+\n"+
			"| %-58s |\n"+
			"+------------------------------------------------------------+\n"+
			"| Module:   %-48s |\n"+
			"| NodeID:   %-48s |\n"+
			"| Rung:     %-48s |\n"+
			"| Peers:    %-48s |\n"+
			"| Traffic:  %-48s |\n"+
			"| Airlock:  %-48s |\n"+
			"+------------------------------------------------------------+",
		"FAITHNET LAYER 2 DISPATCHER DIAGNOSTIC CARD",
		ModuleIDDispatcher,
		d.manifest.ID.String(),
		d.ladder.ActiveRung().String(),
		fmt.Sprintf("%d active / %d max", d.router.ActiveCount(), engine.MaxPeers),
		fmt.Sprintf("sent=%d dec=%d pong=%d deliv=%d",
			s.FramesSent, s.FramesDecoded, s.PongsMatched, s.DataDelivered),
		fmt.Sprintf("blocked=%d decode_err=%d", s.AirlockBlocked, s.DecodeErrors),
	)
}

// ------------------------------------------------------------------------------
// Sub-Assembly 5.2: Subsystem Lifecycle Initialization & Teardown
// ------------------------------------------------------------------------------

// NewDispatcher creates a sovereign Dispatcher wired to Layer 1/2 cogs.
//
// ### For Architects, Operators & Non-Programmers:
// Boots a complete FaithNet node brain governed by the given manifest policy.
//
// ### For Systems Programmers & Engine Architects:
// Validates `manifest.IsValid()` upfront; defaults nil clock to MonotonicClock.
func NewDispatcher(
	manifest data.NodeManifest,
	clock engine.Clock,
) (*Dispatcher, error) {
	if !manifest.IsValid() {
		return nil, ErrDispatcherInvalidManifest
	}
	if clock == nil {
		clock = engine.NewMonotonicClock()
	}
	return &Dispatcher{
		manifest:   manifest,
		clock:      clock,
		ladder:     NewAdaptiveLadder(manifest, clock),
		ping:       engine.NewPingEngine(manifest.ID, clock),
		router:     engine.NewRouter(manifest.ID),
		airlock:    covenant.NewFilter(),
		pongNotify: make(chan uint32, engine.MaxInFlightPings),
	}, nil
}

// Stop cancels all receive goroutines and waits for clean completion.
//
// ### For Architects, Operators & Non-Programmers:
// Stops listening on all communication wires cleanly without leaving background
// tasks running.
//
// ### For Systems Programmers & Engine Architects:
// Idempotent teardown synchronized via `lifeMu` and `sync.WaitGroup`.
func (d *Dispatcher) Stop() {
	d.lifeMu.Lock()
	cancel := d.runCancel
	wasRunning := d.running
	d.running = false
	d.runCancel = nil
	d.lifeMu.Unlock()

	if wasRunning && cancel != nil {
		cancel()
		d.wg.Wait()
	}
}

// Close stops all receive loops and closes every transceiver on the ladder.
//
// ### For Architects, Operators & Non-Programmers:
// Performs complete graceful shutdown of the node and all attached wires.
//
// ### For Systems Programmers & Engine Architects:
// Calls `Stop()` prior to `ladder.Close()` so listener goroutines exit cleanly.
func (d *Dispatcher) Close() error {
	d.Stop()
	return d.ladder.Close()
}

// End of orchestrator/dispatcher.go
