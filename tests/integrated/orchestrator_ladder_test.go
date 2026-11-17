// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/integrated/orchestrator_ladder_test.go
// @id            test.integrated.orchestrator_ladder
// @layer         Layer 4: Tier-3 Integrated Verification Suite
// @archetype     verification_suite | integrated_cascade
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Full 4-Rung Bidirectional Cascade Integration Suite
// @role          Tier-3 Integrated Sensor: Rung 4->3->2->1 Fall & 1->4 Rise
// @maturity      Rev 5 Baseline | Wave 3 Orchestrator Integration
// @chassis       bounded-heap | Two full live Dispatchers across 4 survival rungs
// @scale         O(N_rungs) live mid-flight link severing & recovery verification
// @seams         in: data, engine, covenant, orchestrator, rung2_host, common
// @invariants    1. Macro Tier 3 (Integrated): composes Layers 0+1+2+3 live.
//                2. Severing Rung 4 -> 3 -> 2 mid-flight cascades cleanly to
//                   Rung 1 Floor without dropping sovereign Ping/Pong dialogue.
//                3. Healing Rung 3 and Rung 4 triggers deterministic Rising Up.
//                4. Zero race conditions or goroutine leaks under `-race`.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Tier-3 Integrated verification suite proving FaithNet's MillenniumOS-style
//   Falling-Down (`EVT_CASCADE_DOWN`) and Rising-Up (`EVT_CASCADE_UP`) pattern
//   across TWO full live orchestrated nodes bridged over all 4 survival rungs.
//
//   ### For Architects, Operators & Non-Programmers:
//   Simulates unplugging a node's internet, then Wi-Fi, then local host bus
//   right in the middle of a live conversation—proving the two FaithNet nodes
//   automatically keep talking all the way down to the Rung 1 Floor, and climb
//   right back up when links heal.
//
//   ### For Systems Programmers & Engine Architects:
//   Uses tagged Rung 1..4 `MemoryBus` wrappers (`rungTaggedTx`) so real encoded
//   512-byte wire frames traverse separate physical buses per rung under `-race`.
// ==============================================================================

package integrated_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Identifier, Imports & Rung-Tagged Multi-Bus Bridge Adapter )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-3 Integrated Verification Suite
// Module:        test.integrated.orchestrator_ladder
// Mission:       Bridge two live Dispatchers across 4 independent survival buses.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"context"
	"testing"
	"time"
)

// --- Tier 2: External Library Dependencies (None) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---
import (
	"github.com/CreativeWorkzStudio/FaithNet/data"
	"github.com/CreativeWorkzStudio/FaithNet/engine"
	"github.com/CreativeWorkzStudio/FaithNet/orchestrator"
	"github.com/CreativeWorkzStudio/FaithNet/tests/common"
	"github.com/CreativeWorkzStudio/FaithNet/transport"
	"github.com/CreativeWorkzStudio/FaithNet/transport/rung2_host"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Operational Suite Constants & Identifiers
// ------------------------------------------------------------------------------

const (
	orchestratorLadderSuiteName = "Tier 3 Integrated: 4-Rung Bidirectional Cascade"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.2: Rung-Tagged MemoryBus Wrapper (4 Independent Survival Buses)
// ------------------------------------------------------------------------------

type rungTaggedTx struct {
	inner *rung2_host.MemoryTransceiver
	rung  data.RungLevel
	label string
}

func (r *rungTaggedTx) Name() string          { return r.label }
func (r *rungTaggedTx) Rung() data.RungLevel  { return r.rung }
func (r *rungTaggedTx) Healthy() bool         { return r.inner.Healthy() }
func (r *rungTaggedTx) SetHealthy(ok bool)    { r.inner.SetHealthy(ok) }
func (r *rungTaggedTx) Close() error          { return r.inner.Close() }
func (r *rungTaggedTx) Send(t data.NodeID, raw []byte) error {
	return r.inner.Send(t, raw)
}
func (r *rungTaggedTx) Receive(ctx context.Context) (transport.WireEnvelope, error) {
	return r.inner.Receive(ctx)
}

func wireFourRungPair(
	t *testing.T,
	dA, dB *orchestrator.Dispatcher,
) [5]*rungTaggedTx {
	t.Helper()
	var aTxs [5]*rungTaggedTx
	rungs := []struct {
		level data.RungLevel
		name  string
	}{
		{data.Rung1Floor, "rung1:floor-sim"},
		{data.Rung2Host, "rung2:host-sim"},
		{data.Rung3Local, "rung3:local-sim"},
		{data.Rung4Global, "rung4:global-sim"},
	}
	for _, r := range rungs {
		bus := rung2_host.NewMemoryBus()
		txA := &rungTaggedTx{inner: bus.Attach(dA.Manifest().ID), rung: r.level, label: r.name}
		txB := &rungTaggedTx{inner: bus.Attach(dB.Manifest().ID), rung: r.level, label: r.name}
		if err := dA.Ladder().RegisterRung(txA); err != nil {
			t.Fatalf("RegisterRung A %s failed: %v", r.name, err)
		}
		if err := dB.Ladder().RegisterRung(txB); err != nil {
			t.Fatalf("RegisterRung B %s failed: %v", r.name, err)
		}
		aTxs[r.level] = txA
	}
	return aTxs
}

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-3 Integrated Verification Suite
// Module:        test.integrated.orchestrator_ladder
// Mission:       Prove full 4->3->2->1 Defensive Cascade and 1->4 Rising Up.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: [ SECTION 1: ALWAYS PASS (NOMINAL 4-RUNG TOP-TIER PING) ]
// ------------------------------------------------------------------------------

func TestOrchestratorIntegrated_Sec1_NominalFourRungSelection(t *testing.T) {
	mA := data.DefaultNodeManifest(common.NodeAlpha, "alpha")
	mA.AllowGlobalEgress = true
	mA.PreferredMaxRung = data.Rung4Global

	mB := data.DefaultNodeManifest(common.NodeBeta, "beta")
	mB.AllowGlobalEgress = true
	mB.PreferredMaxRung = data.Rung4Global

	dA, _ := orchestrator.NewDispatcher(mA, engine.NewMonotonicClock())
	dB, _ := orchestrator.NewDispatcher(mB, engine.NewMonotonicClock())
	defer func() {
		_ = dA.Close()
		_ = dB.Close()
	}()

	_ = wireFourRungPair(t, dA, dB)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = dA.Start(ctx)
	_ = dB.Start(ctx)

	_, rung, err := dA.PingPeer(ctx, common.NodeBeta, []byte("TOP-RUNG"))
	if err != nil || rung != data.Rung4Global {
		t.Fatalf("[SEC 1] expected Rung4Global ping, got rung=%s err=%v", rung, err)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: [ SECTION 2: NEVER FAIL (FULL FLOOR COLLAPSE RECOVERY) ]
// ------------------------------------------------------------------------------

func TestOrchestratorIntegrated_Sec2_CompleteFloorCollapseAndRecovery(t *testing.T) {
	mA := data.DefaultNodeManifest(common.NodeAlpha, "alpha")
	mA.AllowGlobalEgress = true
	mA.PreferredMaxRung = data.Rung4Global
	mB := data.DefaultNodeManifest(common.NodeBeta, "beta")
	mB.AllowGlobalEgress = true
	mB.PreferredMaxRung = data.Rung4Global

	dA, _ := orchestrator.NewDispatcher(mA, engine.NewMonotonicClock())
	dB, _ := orchestrator.NewDispatcher(mB, engine.NewMonotonicClock())
	defer func() {
		_ = dA.Close()
		_ = dB.Close()
	}()

	aTxs := wireFourRungPair(t, dA, dB)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = dA.Start(ctx)
	_ = dB.Start(ctx)

	// Sever ALL 4 rungs -> must return ErrFloorCollapsed without panicking.
	for r := data.Rung1Floor; r <= data.Rung4Global; r++ {
		aTxs[r].SetHealthy(false)
	}
	if _, _, err := dA.PingPeer(ctx, common.NodeBeta, []byte("DARK")); err != orchestrator.ErrFloorCollapsed {
		t.Fatalf("[SEC 2] expected ErrFloorCollapsed with all rungs cut, got %v", err)
	}

	// Restore ONLY Rung 1 Floor -> live ping recovers immediately!
	aTxs[data.Rung1Floor].SetHealthy(true)
	_, rung, err := dA.PingPeer(ctx, common.NodeBeta, []byte("FLOOR-ALIVE"))
	if err != nil || rung != data.Rung1Floor {
		t.Fatalf("[SEC 2] expected recovery on Rung1Floor, got rung=%s err=%v", rung, err)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: [ SECTION 3: INVARIANTS, SATURATION & LIMIT GUARDS ]
// ------------------------------------------------------------------------------

func TestOrchestratorIntegrated_Sec3_OfflineNodeNeverUsesRung4EvenIfWired(t *testing.T) {
	// Default offline manifest (AllowGlobalEgress=false) wired with all 4 rungs.
	dA, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(common.NodeAlpha, "sovereign-offline"),
		engine.NewMonotonicClock(),
	)
	dB, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(common.NodeBeta, "sovereign-beta"),
		engine.NewMonotonicClock(),
	)
	defer func() {
		_ = dA.Close()
		_ = dB.Close()
	}()

	_ = wireFourRungPair(t, dA, dB)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = dA.Start(ctx)
	_ = dB.Start(ctx)

	_, rung, err := dA.PingPeer(ctx, common.NodeBeta, []byte("OFFLINE-CAP"))
	if err != nil || rung != data.Rung3Local {
		t.Fatalf("[SEC 3] offline node must top out at Rung3Local, got rung=%s err=%v", rung, err)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.4: [ SECTION 4: EMERGENCE (4->3->2->1 FALL & 1->3->4 RISE) ]
// ------------------------------------------------------------------------------

func TestOrchestratorIntegrated_Sec4_FullBidirectionalCascade4To1AndBackTo4(t *testing.T) {
	mA := data.DefaultNodeManifest(common.NodeAlpha, "alpha-mesh")
	mA.AllowGlobalEgress = true
	mA.PreferredMaxRung = data.Rung4Global
	mB := data.DefaultNodeManifest(common.NodeBeta, "beta-mesh")
	mB.AllowGlobalEgress = true
	mB.PreferredMaxRung = data.Rung4Global

	dA, _ := orchestrator.NewDispatcher(mA, engine.NewMonotonicClock())
	dB, _ := orchestrator.NewDispatcher(mB, engine.NewMonotonicClock())
	defer func() {
		_ = dA.Close()
		_ = dB.Close()
	}()

	aTxs := wireFourRungPair(t, dA, dB)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = dA.Start(ctx)
	_ = dB.Start(ctx)

	// Step 1: Healthy at Rung 4 Global.
	if _, rung, err := dA.PingPeer(ctx, common.NodeBeta, []byte("STEP-R4")); err != nil || rung != data.Rung4Global {
		t.Fatalf("[SEC 4] step R4 failed: rung=%s err=%v", rung, err)
	}

	// Step 2: Sever Rung 4 -> EVT_CASCADE_DOWN to Rung 3 Local.
	aTxs[data.Rung4Global].SetHealthy(false)
	if _, rung, err := dA.PingPeer(ctx, common.NodeBeta, []byte("STEP-R3")); err != nil || rung != data.Rung3Local {
		t.Fatalf("[SEC 4] cascade 4->3 failed: rung=%s err=%v", rung, err)
	}

	// Step 3: Sever Rung 3 -> EVT_CASCADE_DOWN to Rung 2 Host.
	aTxs[data.Rung3Local].SetHealthy(false)
	if _, rung, err := dA.PingPeer(ctx, common.NodeBeta, []byte("STEP-R2")); err != nil || rung != data.Rung2Host {
		t.Fatalf("[SEC 4] cascade 3->2 failed: rung=%s err=%v", rung, err)
	}

	// Step 4: Sever Rung 2 -> EVT_CASCADE_DOWN to Rung 1 Civilizational Floor!
	aTxs[data.Rung2Host].SetHealthy(false)
	if _, rung, err := dA.PingPeer(ctx, common.NodeBeta, []byte("STEP-R1")); err != nil || rung != data.Rung1Floor {
		t.Fatalf("[SEC 4] cascade 2->1 failed: rung=%s err=%v", rung, err)
	}

	// Step 5: Heal Rung 3 -> Rising Up to Rung 3 Local!
	aTxs[data.Rung3Local].SetHealthy(true)
	if _, rung, err := dA.PingPeer(ctx, common.NodeBeta, []byte("RISE-R3")); err != nil || rung != data.Rung3Local {
		t.Fatalf("[SEC 4] rising up 1->3 failed: rung=%s err=%v", rung, err)
	}

	// Step 6: Heal Rung 4 -> Rising Up to Rung 4 Global!
	aTxs[data.Rung4Global].SetHealthy(true)
	if _, rung, err := dA.PingPeer(ctx, common.NodeBeta, []byte("RISE-R4")); err != nil || rung != data.Rung4Global {
		t.Fatalf("[SEC 4] rising up 3->4 failed: rung=%s err=%v", rung, err)
	}

	// Confirm 6/6 pings succeeded with 0% loss across the entire 4->1->4 storm!
	h := dA.Ping().HealthSnapshot()
	if h.PingsSent != 6 || h.PongsRecv != 6 || h.LossPercent() != 0 {
		t.Fatalf("[SEC 4] expected 6/6 pings (0%% loss) across full cascade, got %+v", h)
	}

	// Verify flight-recorder captured both OpCascadeDown and OpCascadeUp events.
	var hist [32]data.CascadeEvent
	n := dA.Ladder().HistorySnapshotInto(hist[:])
	downs, ups := 0, 0
	for i := 0; i < n; i++ {
		if hist[i].Direction == data.OpCascadeDown {
			downs++
		}
		if hist[i].Direction == data.OpCascadeUp {
			ups++
		}
	}
	if downs < 3 || ups < 2 {
		t.Fatalf("[SEC 4] expected >=3 cascade downs and >=2 ups, got downs=%d ups=%d", downs, ups)
	}
	common.AssertAuditPass(t, dA.AuditState(), dA.ExecuteDispatcherDumpCard())
	common.AssertAuditPass(t, dB.AuditState(), dB.ExecuteDispatcherDumpCard())
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Preflight Diagnostic Gates & Physical Source Anchor Verification )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-3 Integrated Verification Suite
// Module:        test.integrated.orchestrator_ladder
// Mission:       Assert Layer 2 preflights and verify source anchors.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Block 4 Pre-Flight Contract Gates
// ------------------------------------------------------------------------------

func TestOrchestratorIntegrated_Block4_PreflightGates(t *testing.T) {
	if orchestratorLadderSuiteName == "" {
		t.Fatal("[BLOCK 4] suite identifier must not be empty")
	}
	common.AssertAuditPass(t, orchestrator.PreflightProbe(), "")
	common.AssertAuditPass(t, orchestrator.PreflightLadder(), "")
	common.AssertAuditPass(t, orchestrator.PreflightDispatcher(), "")
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Physical Source Anchor Verification
// ------------------------------------------------------------------------------

func TestOrchestratorIntegrated_Block4_SourceAnchors(t *testing.T) {
	common.AssertRepoAnchors(t,
		"orchestrator/probe.go",
		"orchestrator/ladder.go",
		"orchestrator/dispatcher.go",
	)
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Test Suite Epilogue & Clean Readiness Handshake )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-3 Integrated Verification Suite
// Module:        test.integrated.orchestrator_ladder
// Mission:       Confirm full Layer 2 readiness handshake post-cascade suite.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Epilogue Readiness Verification
// ------------------------------------------------------------------------------

func TestOrchestratorIntegrated_Block5_EpilogueReadiness(t *testing.T) {
	if !orchestrator.IsProbeReady() || !orchestrator.IsLadderReady() ||
		!orchestrator.IsDispatcherReady() || !common.IsHarnessReady() {
		t.Fatal("[BLOCK 5: EPILOGUE] integrated orchestrator readiness failed")
	}
}

// End of tests/integrated/orchestrator_ladder_test.go
