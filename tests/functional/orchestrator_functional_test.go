// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/functional/orchestrator_functional_test.go
// @id            test.functional.orchestrator
// @layer         Layer 4: Tier-1 Functional Verification Suite
// @archetype     verification_suite | functional_surface
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Layer 2 Orchestrator Functional Surface Suite
// @role          Tier-1 Functional Diagnostic Sensor (4-Section Micro Protocol)
// @maturity      Rev 5 Baseline | 4x4 Verification Matrix
// @chassis       bounded-heap | Exercises Probe, Ladder & Dispatcher surfaces
// @scale         O(1) deterministic orchestration surface sweep
// @seams         in: data, engine, orchestrator, transport/rung2_host, common
// @invariants    1. Macro Tier 1 (Functional): verifies ProbeHostBody,
//                   AdaptiveLadder, and Dispatcher public APIs.
//                2. Offline-first policy capping enforced across all surfaces.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Tier-1 Functional surface suite exercising Layer 2 `orchestrator/`
//   (`probe.go`, `ladder.go`, `dispatcher.go`) across Sections 1-4 plus
//   structured Block 4 preflight gates and Block 5 epilogue readiness.
//
//   ### For Architects, Operators & Non-Programmers:
//   Confirms at a surface level that FaithNet's hardware body prober, adaptive
//   survival ladder, and central node dispatcher all work cleanly together.
//
//   ### For Systems Programmers & Engine Architects:
//   Paired with fine-grained unit suites in `tests/unit/orchestrator/` and
//   4-rung cascade integration testing in `tests/integrated/orchestrator_ladder_test.go`.
// ==============================================================================

package functional_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Identifier, Verification Imports & Tier-1 Sensor Anchor )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-1 Functional Verification Suite
// Module:        test.functional.orchestrator
// Mission:       Bind Layer 2 orchestrator cogs and shared test harness.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"context"
	"strings"
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
	"github.com/CreativeWorkzStudio/FaithNet/transport/rung2_host"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Operational Suite Constants & Identifiers
// ------------------------------------------------------------------------------

const (
	orchestratorFunctionalSuiteName = "Tier 1: Layer 2 Orchestrator Surface"
)

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-1 Functional Verification Suite
// Module:        test.functional.orchestrator
// Mission:       Exercise Probe, AdaptiveLadder, and Dispatcher across Sec 1-4.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: [ SECTION 1: ALWAYS PASS (NOMINAL GOLDEN SURFACE COGS) ]
// ------------------------------------------------------------------------------

func TestOrchestratorFunctional_Sec1_ProbeLadderAndDispatcherSurface(t *testing.T) {
	body := orchestrator.ProbeHostBody()
	common.AssertAuditPass(t, orchestrator.AuditHardwareProfile(body),
		orchestrator.ExecuteProbeDumpCard(body))

	manifest := data.DefaultNodeManifest(common.NodeAlpha, "orch-func")
	ceiling := orchestrator.EvaluatePolicyCeiling(body, manifest)
	if !ceiling.IsValid() || ceiling > data.Rung3Local {
		t.Fatalf("[SEC 1] unexpected offline policy ceiling: %s", ceiling)
	}

	bus := rung2_host.NewMemoryBus()
	dA, err := orchestrator.NewDispatcher(manifest, engine.NewMonotonicClock())
	if err != nil {
		t.Fatalf("[SEC 1] NewDispatcher A failed: %v", err)
	}
	defer func() { _ = dA.Close() }()

	dB, err := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(common.NodeBeta, "orch-beta"),
		engine.NewMonotonicClock(),
	)
	if err != nil {
		t.Fatalf("[SEC 1] NewDispatcher B failed: %v", err)
	}
	defer func() { _ = dB.Close() }()

	_ = dA.Ladder().RegisterRung(bus.Attach(common.NodeAlpha))
	_ = dB.Ladder().RegisterRung(bus.Attach(common.NodeBeta))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_ = dA.Start(ctx)
	_ = dB.Start(ctx)

	rtt, rung, err := dA.PingPeer(ctx, common.NodeBeta, []byte("ORCH-SURFACE"))
	if err != nil || rung != data.Rung2Host || rtt < 0 {
		t.Fatalf("[SEC 1] PingPeer failed: rtt=%d rung=%s err=%v", rtt, rung, err)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: [ SECTION 2: NEVER FAIL (FAULT INVERSION & DEGRADATION) ]
// ------------------------------------------------------------------------------

func TestOrchestratorFunctional_Sec2_RejectsInvalidManifestAndEmptyLadder(t *testing.T) {
	bad := data.DefaultNodeManifest(data.NodeIDNull, "bad")
	if _, err := orchestrator.NewDispatcher(bad, nil); err != orchestrator.ErrDispatcherInvalidManifest {
		t.Fatalf("[SEC 2] expected ErrDispatcherInvalidManifest, got %v", err)
	}
	lad := orchestrator.NewAdaptiveLadder(
		data.DefaultNodeManifest(common.NodeAlpha, "empty"), engine.NewMockClock(0),
	)
	defer func() { _ = lad.Close() }()

	if err := lad.RegisterRung(nil); err != orchestrator.ErrInvalidLadderRung {
		t.Fatalf("[SEC 2] RegisterRung(nil) got %v", err)
	}
	if _, err := lad.SendFrame(common.NodeBeta, []byte("X")); err != orchestrator.ErrFloorCollapsed {
		t.Fatalf("[SEC 2] empty ladder SendFrame got %v", err)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: [ SECTION 3: INVARIANTS, SATURATION & LIMIT GUARDS ]
// ------------------------------------------------------------------------------

func TestOrchestratorFunctional_Sec3_OfflinePolicyCeilingEnforcement(t *testing.T) {
	rich := common.RichHostProfile()
	offline := data.DefaultNodeManifest(common.NodeAlpha, "airgapped")
	if got := orchestrator.EvaluatePolicyCeiling(rich, offline); got != data.Rung3Local {
		t.Fatalf("[SEC 3] offline manifest must cap RichHost at Rung3Local, got %s", got)
	}
	globalOptIn := offline
	globalOptIn.AllowGlobalEgress = true
	globalOptIn.PreferredMaxRung = data.Rung4Global
	if got := orchestrator.EvaluatePolicyCeiling(rich, globalOptIn); got != data.Rung4Global {
		t.Fatalf("[SEC 3] opt-in manifest should permit Rung4Global, got %s", got)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.4: [ SECTION 4: EMERGENCE & COMPOSITE WORKFLOWS ]
// ------------------------------------------------------------------------------

func TestOrchestratorFunctional_Sec4_DumpCardsEmitCompleteTelemetry(t *testing.T) {
	body := orchestrator.ProbeHostBody()
	if !strings.Contains(orchestrator.ExecuteProbeDumpCard(body), "HOST BODY HARDWARE PROBE") {
		t.Fatal("[SEC 4] probe dump card missing title")
	}
	d, _ := orchestrator.NewDispatcher(data.DefaultNodeManifest(common.NodeAlpha, "card"), nil)
	defer func() { _ = d.Close() }()
	if !strings.Contains(d.Ladder().ExecuteLadderDumpCard(), "ADAPTIVE LADDER") {
		t.Fatal("[SEC 4] ladder dump card missing title")
	}
	if !strings.Contains(d.ExecuteDispatcherDumpCard(), "DISPATCHER DIAGNOSTIC") {
		t.Fatal("[SEC 4] dispatcher dump card missing title")
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Preflight Diagnostic Gates & Physical Source Anchor Verification )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-1 Functional Verification Suite
// Module:        test.functional.orchestrator
// Mission:       Run Probe, Ladder, and Dispatcher Block 4 preflights & anchors.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Block 4 Pre-Flight Contract Gates
// ------------------------------------------------------------------------------

func TestOrchestratorFunctional_Block4_PreflightGates(t *testing.T) {
	if orchestratorFunctionalSuiteName == "" {
		t.Fatal("[BLOCK 4] suite identifier must not be empty")
	}
	common.AssertAuditPass(t, orchestrator.PreflightProbe(), "")
	common.AssertAuditPass(t, orchestrator.PreflightLadder(), "")
	common.AssertAuditPass(t, orchestrator.PreflightDispatcher(), "")
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Physical Source Anchor Verification
// ------------------------------------------------------------------------------

func TestOrchestratorFunctional_Block4_SourceAnchors(t *testing.T) {
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
// Stratum:       Layer 4: Tier-1 Functional Verification Suite
// Module:        test.functional.orchestrator
// Mission:       Confirm Layer 2 orchestrator readiness post-suite.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Epilogue Readiness Verification
// ------------------------------------------------------------------------------

func TestOrchestratorFunctional_Block5_EpilogueReadiness(t *testing.T) {
	if !orchestrator.IsProbeReady() || !orchestrator.IsLadderReady() ||
		!orchestrator.IsDispatcherReady() || !common.IsHarnessReady() {
		t.Fatal("[BLOCK 5: EPILOGUE] Layer 2 orchestrator readiness handshake failed")
	}
}

// End of tests/functional/orchestrator_functional_test.go
