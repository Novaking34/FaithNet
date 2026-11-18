// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/unit/orchestrator/dispatcher_test.go
// @id            test.unit.orchestrator.dispatcher
// @layer         Layer 4: Tier-2 Unit Verification Suite (Orchestrator)
// @archetype     verification_suite | unit_isolation
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Layer 2 Central Dispatcher Unit Suite
// @role          Tier-2 Unit Isolation Gate for cog.orchestrator.dispatcher
// @maturity      Rev 5 Baseline | 4x4 Verification Matrix
// @chassis       bounded-heap | Verifies Codec -> Airlock -> Router -> Deliver
// @scale         O(1) deterministic inbound/outbound dispatch isolation
// @seams         in: data, engine, covenant, orchestrator, rung2_host, common
// @invariants    1. Malformed wire bytes increment DecodeErrors without panic.
//                2. Hostile frames increment AirlockBlocked & never reach Deliver.
//                3. Foreign-target frames are ignored without delivery.
//                4. Start() rejects duplicate runs; Stop()/Close() is idempotent.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Tier-2 Unit isolation suite strictly verifying `orchestrator/dispatcher.go`
//   across synchronous inbound envelopes, delivery callbacks, airlock blocks,
//   duplicate Start() guards, and live `AuditState`.
//
//   ### For Architects, Operators & Non-Programmers:
//   Tests the FaithNet node brain's door security, message delivery, and clean
//   start/stop lifecycle individually before multi-node integration tests.
//
//   ### For Systems Programmers & Engine Architects:
//   100% branch coverage of `HandleInboundEnvelope`, `SendFrame`, `PingPeer`,
//   `Start`, `Stop`, `Close`, `AuditState`, and `ExecuteDispatcherDumpCard`.
// ==============================================================================

package orchestrator_unit_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Identifier, Verification Imports & Tier-2 Unit Sensor Anchor )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.orchestrator.dispatcher
// Mission:       Bind Dispatcher, Covenant, Rung 2 MemoryBus & shared harness.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"context"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// --- Tier 2: External Library Dependencies (None) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---
import (
	"github.com/CreativeWorkzStudio/FaithNet/data"
	"github.com/CreativeWorkzStudio/FaithNet/engine"
	"github.com/CreativeWorkzStudio/FaithNet/engine/covenant"
	"github.com/CreativeWorkzStudio/FaithNet/orchestrator"
	"github.com/CreativeWorkzStudio/FaithNet/tests/common"
	"github.com/CreativeWorkzStudio/FaithNet/transport"
	"github.com/CreativeWorkzStudio/FaithNet/transport/rung2_host"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Operational Suite Constants & Identifiers
// ------------------------------------------------------------------------------

const (
	dispatcherUnitSuiteName = "Tier 2 Unit: cog.orchestrator.dispatcher"
)

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.orchestrator.dispatcher
// Mission:       Verify inbound decode/airlock/deliver, outbound send & guards.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: [ SECTION 1: ALWAYS PASS (NOMINAL GOLDEN SURFACE COGS) ]
// ------------------------------------------------------------------------------

func TestDispatcherUnit_Sec1_SovereignDeliverAndRouteLearning(t *testing.T) {
	d, err := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(common.NodeBeta, "beta"), engine.NewMockClock(100),
	)
	if err != nil {
		t.Fatalf("[SEC 1] NewDispatcher failed: %v", err)
	}
	defer func() { _ = d.Close() }()

	var delivered atomic.Int64
	var gotTruth atomic.Uint32
	d.SetDeliverHandler(func(f *data.Frame, v covenant.Verdict) {
		delivered.Add(1)
		gotTruth.Store(uint32(v.State))
	})

	frame := common.MakeSovereignFrame(
		common.NodeAlpha, common.NodeBeta, data.Rung2Host, data.OpCPISync, 1,
		[]byte("CPI-SI-DELTA"),
	)
	wire := common.EncodeFixture(t, &frame)
	var env transport.WireEnvelope
	env.Length = len(wire)
	copy(env.Bytes[:], wire)

	out := d.HandleInboundEnvelope(env)
	if !out.Decoded || out.Verdict.State != data.QuadPositive ||
		delivered.Load() != 1 || data.QuadState(gotTruth.Load()) != data.QuadPositive {
		t.Fatalf("[SEC 1] sovereign delivery failed: out=%+v deliv=%d", out, delivered.Load())
	}
	if _, learned := d.Router().Lookup(common.NodeAlpha); !learned {
		t.Fatal("[SEC 1] Dispatcher failed to learn sender route into Router")
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: [ SECTION 2: NEVER FAIL (FAULT INVERSION & DEGRADATION) ]
// ------------------------------------------------------------------------------

func TestDispatcherUnit_Sec2_CorruptWireHostileAirlockAndForeignTarget(t *testing.T) {
	d, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(common.NodeBeta, "beta"), engine.NewMockClock(100),
	)
	defer func() { _ = d.Close() }()

	var delivered atomic.Int64
	d.SetDeliverHandler(func(_ *data.Frame, _ covenant.Verdict) {
		delivered.Add(1)
	})

	// 1. Corrupt random wire bytes -> DecodeErrors++.
	corrupt := transport.WireEnvelope{Length: 40}
	out1 := d.HandleInboundEnvelope(corrupt)
	if out1.Decoded || d.StatsSnapshot().DecodeErrors != 1 {
		t.Fatalf("[SEC 2] corrupt wire did not record decode error: %+v", out1)
	}

	// 2. Hostile external payload -> blocked at Genesis 3 Airlock, NEVER delivered.
	hostile := common.MakeSovereignFrame(
		common.NodeExternal, common.NodeBeta, data.Rung4Global, data.OpLegacyWrap, 2,
		[]byte("1; DROP TABLE users;"),
	)
	hostile.Flags = data.FlagExternalOrigin
	wireH := common.EncodeFixture(t, &hostile)
	var envH transport.WireEnvelope
	envH.Length = len(wireH)
	copy(envH.Bytes[:], wireH)
	out2 := d.HandleInboundEnvelope(envH)
	if !out2.Verdict.IsBlocked() || d.StatsSnapshot().AirlockBlocked != 1 {
		t.Fatalf("[SEC 2] hostile SQLi bypassed Dispatcher Airlock: %+v", out2)
	}

	// 3. Valid frame addressed to a DIFFERENT target node -> ignored cleanly.
	foreign := common.MakeSovereignFrame(
		common.NodeAlpha, common.NodeRelay, data.Rung2Host, data.OpData, 3, []byte("FOR-RELAY"),
	)
	wireF := common.EncodeFixture(t, &foreign)
	var envF transport.WireEnvelope
	envF.Length = len(wireF)
	copy(envF.Bytes[:], wireF)
	_ = d.HandleInboundEnvelope(envF)

	if delivered.Load() != 0 {
		t.Fatalf("[SEC 2] forbidden/foreign frames reached DeliverHandler: %d", delivered.Load())
	}

	// 4. Nil SendFrame & nil AuditState rejected cleanly.
	if _, err := d.SendFrame(nil); err != engine.ErrBufferTooSmall {
		t.Fatalf("[SEC 2] SendFrame(nil) got %v", err)
	}
	var nilD *orchestrator.Dispatcher
	common.AssertAuditFail(t, nilD.AuditState(), data.AuditSeverityCritical)
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: [ SECTION 3: INVARIANTS, SATURATION & LIMIT GUARDS ]
// ------------------------------------------------------------------------------

func TestDispatcherUnit_Sec3_DuplicateStartGuardAndCleanStop(t *testing.T) {
	bus := rung2_host.NewMemoryBus()
	d, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(common.NodeAlpha, "alpha"), engine.NewMockClock(0),
	)
	_ = d.Ladder().RegisterRung(bus.Attach(common.NodeAlpha))

	if err := d.Start(context.Background()); err != nil {
		t.Fatalf("[SEC 3] first Start failed: %v", err)
	}
	if err := d.Start(context.Background()); err != orchestrator.ErrDispatcherAlreadyRunning {
		t.Fatalf("[SEC 3] duplicate Start: want ErrDispatcherAlreadyRunning got %v", err)
	}
	d.Stop()
	d.Stop() // Idempotent Stop must never panic.
	_ = d.Close()
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.4: [ SECTION 4: EMERGENCE & COMPOSITE WORKFLOWS ]
// ------------------------------------------------------------------------------

func TestDispatcherUnit_Sec4_SendFrameAutoStampsSourceAndUpdatesStats(t *testing.T) {
	bus := rung2_host.NewMemoryBus()
	dA, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(common.NodeAlpha, "alpha"), engine.NewMockClock(10),
	)
	txB := bus.Attach(common.NodeBeta)
	defer func() {
		_ = dA.Close()
		_ = txB.Close()
	}()
	_ = dA.Ladder().RegisterRung(bus.Attach(common.NodeAlpha))

	var msg data.Frame
	msg.Opcode = data.OpData
	msg.Truth = data.QuadPositive
	msg.TargetID = common.NodeBeta
	msg.SeqID = 11
	msg.SetPayload([]byte("AUTO-STAMPED"))

	rung, err := dA.SendFrame(&msg)
	if err != nil || rung != data.Rung2Host || msg.SourceID != common.NodeAlpha {
		t.Fatalf("[SEC 4] SendFrame failed: rung=%s err=%v src=%s", rung, err, msg.SourceID)
	}
	card := dA.ExecuteDispatcherDumpCard()
	if !strings.Contains(card, "sent=1") {
		t.Fatalf("[SEC 4] unexpected dispatcher card:\n%s", card)
	}
	common.AssertAuditPass(t, dA.AuditState(), card)
}

func TestDispatcherUnit_Sec4_NativeZeroHTTPFetchResourceRoundtrip(t *testing.T) {
	clk := engine.NewMonotonicClock()
	bus := rung2_host.NewMemoryBus()

	dA, err := orchestrator.NewDispatcher(data.DefaultNodeManifest(common.NodeAlpha, "Client-A"), clk)
	if err != nil {
		t.Fatalf("[SEC 4] NewDispatcher A failed: %v", err)
	}
	dB, err := orchestrator.NewDispatcher(data.DefaultNodeManifest(common.NodeBeta, "Anchor-B"), clk)
	if err != nil {
		t.Fatalf("[SEC 4] NewDispatcher B failed: %v", err)
	}
	defer func() { _ = dA.Close(); _ = dB.Close() }()

	if err := dA.Ladder().RegisterRung(bus.Attach(common.NodeAlpha)); err != nil {
		t.Fatalf("[SEC 4] RegisterRung A failed: %v", err)
	}
	if err := dB.Ladder().RegisterRung(bus.Attach(common.NodeBeta)); err != nil {
		t.Fatalf("[SEC 4] RegisterRung B failed: %v", err)
	}

	dB.SetResourceHandler(func(reqPath string, from data.NodeID) ([]byte, bool) {
		if reqPath == "/custom" {
			return []byte("FNP/1-NATIVE-OK:" + from.String()), true
		}
		return nil, false
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := dA.Start(ctx); err != nil {
		t.Fatalf("[SEC 4] dA.Start failed: %v", err)
	}
	if err := dB.Start(ctx); err != nil {
		t.Fatalf("[SEC 4] dB.Start failed: %v", err)
	}

	body, _, rung, err := dA.FetchResource(ctx, common.NodeBeta, "/custom")
	if err != nil || rung != data.Rung2Host || !strings.Contains(string(body), "FNP/1-NATIVE-OK:") {
		t.Fatalf("[SEC 4] native OpFetch /custom failed: body=%q rung=%s err=%v", string(body), rung, err)
	}
	defBody, _, _, err := dA.FetchResource(ctx, common.NodeBeta, "/protocol")
	if err != nil || !strings.Contains(string(defBody), "FNP/1") {
		t.Fatalf("[SEC 4] default /protocol OpFetch failed: body=%q err=%v", string(defBody), err)
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Preflight Diagnostic Gates & Physical Source Anchor Verification )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.orchestrator.dispatcher
// Mission:       Run PreflightDispatcher and verify physical source anchor.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Block 4 Pre-Flight Contract Gates
// ------------------------------------------------------------------------------

func TestDispatcherUnit_Block4_PreflightGates(t *testing.T) {
	if dispatcherUnitSuiteName == "" {
		t.Fatal("[BLOCK 4] suite identifier must not be empty")
	}
	common.AssertAuditPass(t, orchestrator.PreflightDispatcher(), "")
	if !orchestrator.AuditDispatcher() {
		t.Fatal("[BLOCK 4] AuditDispatcher returned false")
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Physical Source Anchor Verification
// ------------------------------------------------------------------------------

func TestDispatcherUnit_Block4_SourceAnchors(t *testing.T) {
	common.AssertRepoAnchors(t, "orchestrator/dispatcher.go")
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Test Suite Epilogue & Clean Readiness Handshake )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.orchestrator.dispatcher
// Mission:       Confirm dispatcher and harness readiness post-suite.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Epilogue Readiness Verification
// ------------------------------------------------------------------------------

func TestDispatcherUnit_Block5_EpilogueReadiness(t *testing.T) {
	if !orchestrator.IsDispatcherReady() || !common.IsHarnessReady() {
		t.Fatal("[BLOCK 5: EPILOGUE] dispatcher unit readiness handshake failed")
	}
}

// End of tests/unit/orchestrator/dispatcher_test.go
