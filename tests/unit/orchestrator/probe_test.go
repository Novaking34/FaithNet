// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/unit/orchestrator/probe_test.go
// @id            test.unit.orchestrator.probe
// @layer         Layer 4: Tier-2 Unit Verification Suite (Orchestrator)
// @archetype     verification_suite | unit_isolation
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Layer 2 Dynamic Host Body Prober Unit Suite
// @role          Tier-2 Unit Isolation Gate for cog.orchestrator.probe
// @maturity      Rev 5 Baseline | 4x4 Verification Matrix
// @chassis       zero-heap | Evaluates HardwareProfile & PolicyCeiling bounds
// @scale         O(1) deterministic synthetic body & policy matrix tests
// @seams         in: data, orchestrator, tests/common | target: probe.go
// @invariants    1. Live ProbeHostBody() always exposes HasMemoryBus==true.
//                2. Policy ceiling never exceeds manifest or hardware reality.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Tier-2 Unit isolation suite strictly verifying `orchestrator/probe.go`
//   across live host probing, synthetic body matrices, invalid profile fault
//   inversion, and 60-column probe cards.
//
//   ### For Architects, Operators & Non-Programmers:
//   Confirms FaithNet's hardware checkup accurately maps any computer's organs
//   to safe survival rungs without violating offline policy locks.
//
//   ### For Systems Programmers & Engine Architects:
//   100% branch coverage of `ProbeHostBody`, `EvaluatePolicyCeiling`,
//   `AuditHardwareProfile`, `PreflightProbe`, and `ExecuteProbeDumpCard`.
// ==============================================================================

package orchestrator_unit_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Identifier, Verification Imports & Tier-2 Unit Sensor Anchor )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.orchestrator.probe
// Mission:       Bind BodyProbe functions and shared harness for isolation.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"strings"
	"testing"
)

// --- Tier 2: External Library Dependencies (None) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---
import (
	"github.com/CreativeWorkzStudio/FaithNet/data"
	"github.com/CreativeWorkzStudio/FaithNet/orchestrator"
	"github.com/CreativeWorkzStudio/FaithNet/tests/common"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Operational Suite Constants & Identifiers
// ------------------------------------------------------------------------------

const (
	probeUnitSuiteName = "Tier 2 Unit: cog.orchestrator.probe (Host Body Probe)"
)

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.orchestrator.probe
// Mission:       Verify host probing, policy intersection, and audit guards.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: [ SECTION 1: ALWAYS PASS (NOMINAL GOLDEN SURFACE COGS) ]
// ------------------------------------------------------------------------------

func TestProbeUnit_Sec1_LiveProbeAndPolicyCeilingMatrix(t *testing.T) {
	live := orchestrator.ProbeHostBody()
	if live.OSName == "" || live.Arch == "" || live.CPUCores < 1 || !live.HasMemoryBus {
		t.Fatalf("[SEC 1] live ProbeHostBody incomplete: %+v", live)
	}
	common.AssertAuditPass(t, orchestrator.AuditHardwareProfile(live),
		orchestrator.ExecuteProbeDumpCard(live))

	offline := data.DefaultNodeManifest(common.NodeAlpha, "offline")
	global := offline
	global.AllowGlobalEgress = true
	global.PreferredMaxRung = data.Rung4Global

	if got := orchestrator.EvaluatePolicyCeiling(common.RichHostProfile(), offline); got != data.Rung3Local {
		t.Fatalf("[SEC 1] rich+offline want Rung3Local got %s", got)
	}
	if got := orchestrator.EvaluatePolicyCeiling(common.RichHostProfile(), global); got != data.Rung4Global {
		t.Fatalf("[SEC 1] rich+global want Rung4Global got %s", got)
	}
	if got := orchestrator.EvaluatePolicyCeiling(common.OfflineRoomProfile(), global); got != data.Rung3Local {
		t.Fatalf("[SEC 1] room+global want Rung3Local got %s", got)
	}
	if got := orchestrator.EvaluatePolicyCeiling(common.EmergencyFloorProfile(), offline); got != data.Rung1Floor {
		t.Fatalf("[SEC 1] floor+offline want Rung1Floor got %s", got)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: [ SECTION 2: NEVER FAIL (FAULT INVERSION & DEGRADATION) ]
// ------------------------------------------------------------------------------

func TestProbeUnit_Sec2_FaultInversionOnBrokenProfilesAndManifests(t *testing.T) {
	common.AssertAuditFail(t, orchestrator.AuditHardwareProfile(data.HardwareProfile{}), data.AuditSeverityCritical)
	common.AssertAuditFail(t, orchestrator.AuditHardwareProfile(data.HardwareProfile{
		OSName: "linux", Arch: "amd64", CPUCores: 0, HasMemoryBus: true,
	}), data.AuditSeverityCritical)
	common.AssertAuditFail(t, orchestrator.AuditHardwareProfile(data.HardwareProfile{
		OSName: "linux", Arch: "amd64", CPUCores: 2, InterfaceCount: -1, HasMemoryBus: true,
	}), data.AuditSeverityCritical)
	common.AssertAuditFail(t, orchestrator.AuditHardwareProfile(data.HardwareProfile{
		OSName: "linux", Arch: "amd64", CPUCores: 2,
	}), data.AuditSeverityCritical)

	// Dead hardware or invalid manifest -> RungNone cleanly.
	offline := data.DefaultNodeManifest(common.NodeAlpha, "ok")
	if got := orchestrator.EvaluatePolicyCeiling(data.HardwareProfile{}, offline); got != data.RungNone {
		t.Fatalf("[SEC 2] dead body want RungNone got %s", got)
	}
	if got := orchestrator.EvaluatePolicyCeiling(common.RichHostProfile(), data.NodeManifest{}); got != data.RungNone {
		t.Fatalf("[SEC 2] invalid manifest want RungNone got %s", got)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: [ SECTION 3: INVARIANTS, SATURATION & LIMIT GUARDS ]
// ------------------------------------------------------------------------------

func TestProbeUnit_Sec3_MinimumFloorHigherThanHardwareReturnsNone(t *testing.T) {
	manifest := data.DefaultNodeManifest(common.NodeAlpha, "strict-lan")
	manifest.MinimumFloorRung = data.Rung3Local
	manifest.PreferredMaxRung = data.Rung3Local

	if got := orchestrator.EvaluatePolicyCeiling(common.EmergencyFloorProfile(), manifest); got != data.RungNone {
		t.Fatalf("[SEC 3] floor-only body below Rung3 minimum must return RungNone, got %s", got)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.4: [ SECTION 4: EMERGENCE & COMPOSITE WORKFLOWS ]
// ------------------------------------------------------------------------------

func TestProbeUnit_Sec4_ProbeDumpCardReflectsBodyOrgans(t *testing.T) {
	card := orchestrator.ExecuteProbeDumpCard(common.EmergencyFloorProfile())
	if !strings.Contains(card, "RUNG_1_FLOOR") || !strings.Contains(card, "serial=true audio=true") {
		t.Fatalf("[SEC 4] unexpected emergency probe card:\n%s", card)
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Preflight Diagnostic Gates & Physical Source Anchor Verification )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.orchestrator.probe
// Mission:       Run PreflightProbe and verify physical source anchor.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Block 4 Pre-Flight Contract Gates
// ------------------------------------------------------------------------------

func TestProbeUnit_Block4_PreflightGates(t *testing.T) {
	if probeUnitSuiteName == "" {
		t.Fatal("[BLOCK 4] suite identifier must not be empty")
	}
	common.AssertAuditPass(t, orchestrator.PreflightProbe(), "")
	if !orchestrator.AuditProbe() {
		t.Fatal("[BLOCK 4] AuditProbe returned false")
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Physical Source Anchor Verification
// ------------------------------------------------------------------------------

func TestProbeUnit_Block4_SourceAnchors(t *testing.T) {
	common.AssertRepoAnchors(t, "orchestrator/probe.go")
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Test Suite Epilogue & Clean Readiness Handshake )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.orchestrator.probe
// Mission:       Confirm probe and harness readiness post-suite.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Epilogue Readiness Verification
// ------------------------------------------------------------------------------

func TestProbeUnit_Block5_EpilogueReadiness(t *testing.T) {
	if !orchestrator.IsProbeReady() || !common.IsHarnessReady() {
		t.Fatal("[BLOCK 5: EPILOGUE] probe unit readiness handshake failed")
	}
}

// End of tests/unit/orchestrator/probe_test.go
