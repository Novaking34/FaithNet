// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/unit/data/audit_test.go
// @id            test.unit.data.audit
// @layer         Layer 4: Tier-2 Unit Verification Suite (Data Sub-Stratum)
// @archetype     verification_suite | unit_isolation
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Layer 0 AuditResult & AuditSeverity Unit Suite
// @role          Tier-2 Unit Isolation Gate for cog.data.audit
// @maturity      Rev 5 Baseline | 4x4 Verification Matrix
// @chassis       zero-heap | Verifies structured Block 4 diagnostic verdicts
// @scale         O(1) deterministic verdict & severity bounds tests
// @seams         in: data, engine, tests/common | target: data/audit.go
// @invariants    1. AuditPass ALWAYS produces Passed==true & Severity==Pass.
//                2. AuditFail NEVER masquerades as AuditSeverityPass.
//                3. IsPass() enforces dual consistency (Passed && Pass sev).
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Tier-2 Unit isolation suite verifying `AuditSeverity` and `AuditResult`
//   declared in `data/audit.go` across all 4 horizontal micro sections.
//
//   ### For Architects, Operators & Non-Programmers:
//   Verifies FaithNet's diagnostic report cards cannot lie—a failed safety
//   check can never accidentally disguise itself as "PASS".
//
//   ### For Systems Programmers & Engine Architects:
//   100% branch coverage of `AuditSeverity`, `AuditPass`, `AuditFail`,
//   `IsPass`, and `String`.
// ==============================================================================

package data_unit_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Identifier, Verification Imports & Tier-2 Unit Sensor Anchor )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.data.audit
// Mission:       Bind Layer 0 audit verdict primitives and shared harness.
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
	"github.com/CreativeWorkzStudio/FaithNet/engine"
	"github.com/CreativeWorkzStudio/FaithNet/tests/common"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Operational Suite Constants & Identifiers
// ------------------------------------------------------------------------------

const (
	auditUnitSuiteName = "Tier 2 Unit: cog.data.audit (AuditSeverity, AuditResult)"
)

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.data.audit
// Mission:       Verify AuditSeverity scale and AuditResult consistency guards.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: [ SECTION 1: ALWAYS PASS (NOMINAL GOLDEN SURFACE COGS) ]
// ------------------------------------------------------------------------------

func TestAuditUnit_Sec1_AllSeveritiesAndNominalConstructors(t *testing.T) {
	labels := map[data.AuditSeverity]string{
		data.AuditSeverityPass:     "PASS",
		data.AuditSeverityWarn:     "WARN",
		data.AuditSeverityCritical: "CRITICAL",
		data.AuditSeverityFatal:    "FATAL",
	}
	for sev, want := range labels {
		if !sev.IsValid() || sev.String() != want {
			t.Fatalf("[SEC 1] severity %d: valid=%v str=%q want=%q", sev, sev.IsValid(), sev.String(), want)
		}
	}

	ok := data.AuditPass("cog.test", "check_ok", "all good")
	if !ok.IsPass() || !ok.Passed || ok.Severity != data.AuditSeverityPass {
		t.Fatalf("[SEC 1] AuditPass mismatch: %+v", ok)
	}
	if !strings.Contains(ok.String(), "[PASS]") || !strings.Contains(ok.String(), "cog.test") {
		t.Fatalf("[SEC 1] AuditPass.String() formatting mismatch: %q", ok.String())
	}

	fail := data.AuditFail(data.AuditSeverityFatal, "cog.test", "check_fail", "boom")
	if fail.IsPass() || fail.Passed || fail.Severity != data.AuditSeverityFatal {
		t.Fatalf("[SEC 1] AuditFail mismatch: %+v", fail)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: [ SECTION 2: NEVER FAIL (FAULT INVERSION & DEGRADATION) ]
// ------------------------------------------------------------------------------

func TestAuditUnit_Sec2_RejectsMisleadingFailPassAndOOBSeverity(t *testing.T) {
	// AuditFail passed AuditSeverityPass must auto-escalate to Critical.
	escalated := data.AuditFail(data.AuditSeverityPass, "cog.test", "bad_sev", "clamped")
	if escalated.IsPass() || escalated.Severity != data.AuditSeverityCritical {
		t.Fatalf("[SEC 2] AuditFail(Pass) must clamp to Critical, got %+v", escalated)
	}

	// AuditFail passed OOB severity must clamp to Critical.
	oobFail := data.AuditFail(data.AuditSeverity(99), "cog.test", "oob", "clamped")
	if oobFail.Severity != data.AuditSeverityCritical {
		t.Fatalf("[SEC 2] AuditFail(99) must clamp to Critical, got %+v", oobFail)
	}

	// Hand-constructed inconsistent struct (Passed=true with Critical) fails IsPass().
	inconsistent := data.AuditResult{Passed: true, Severity: data.AuditSeverityCritical}
	if inconsistent.IsPass() {
		t.Fatal("[SEC 2] IsPass() must reject Passed==true when Severity!=Pass")
	}
	if data.AuditSeverity(4).IsValid() || data.AuditSeverity(4).String() != "UNKNOWN" {
		t.Fatal("[SEC 2] OOB AuditSeverity did not degrade to UNKNOWN")
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: [ SECTION 3: INVARIANTS, SATURATION & LIMIT GUARDS ]
// ------------------------------------------------------------------------------

func TestAuditUnit_Sec3_ZeroHeapAuditPassAndFailConstruction(t *testing.T) {
	if data.AuditSeverityMinValid != data.AuditSeverityPass ||
		data.AuditSeverityMaxValid != data.AuditSeverityFatal {
		t.Fatal("[SEC 3] AuditSeverity min/max bounds drifted")
	}
	allocs := testing.AllocsPerRun(500, func() {
		r1 := data.AuditPass("cog.a", "c1", "ok")
		r2 := data.AuditFail(data.AuditSeverityCritical, "cog.a", "c2", "bad")
		_ = r1.IsPass()
		_ = r2.IsPass()
	})
	if allocs != 0 {
		t.Fatalf("[SEC 3] ALLOC_ZERO_HEAP violated on AuditPass/AuditFail: %.2f", allocs)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.4: [ SECTION 4: EMERGENCE & COMPOSITE WORKFLOWS ]
// ------------------------------------------------------------------------------

func TestAuditUnit_Sec4_ForensicFormattingAcrossAllSeverities(t *testing.T) {
	r := data.AuditFail(data.AuditSeverityWarn, "cog.net", "rtt_drift", "latency > 50ms")
	s := r.String()
	if !strings.Contains(s, "[WARN]") || !strings.Contains(s, "rtt_drift") ||
		!strings.Contains(s, "latency > 50ms") {
		t.Fatalf("[SEC 4] unexpected forensic string: %q", s)
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Preflight Diagnostic Gates & Physical Source Anchor Verification )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.data.audit
// Mission:       Run vocabulary preflight gate and verify physical anchor.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Block 4 Pre-Flight Contract Gates
// ------------------------------------------------------------------------------

func TestAuditUnit_Block4_PreflightGates(t *testing.T) {
	if auditUnitSuiteName == "" {
		t.Fatal("[BLOCK 4] suite identifier must not be empty")
	}
	common.AssertAuditPass(t, engine.AuditDataVocabulary(), engine.ExecuteCodecDumpCard())
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Physical Source Anchor Verification
// ------------------------------------------------------------------------------

func TestAuditUnit_Block4_SourceAnchors(t *testing.T) {
	common.AssertRepoAnchors(t, "data/audit.go")
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Test Suite Epilogue & Clean Readiness Handshake )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.data.audit
// Mission:       Confirm readiness handshake post-suite.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Epilogue Readiness Verification
// ------------------------------------------------------------------------------

func TestAuditUnit_Block5_EpilogueReadiness(t *testing.T) {
	if !common.IsHarnessReady() || !engine.IsCodecReady() {
		t.Fatal("[BLOCK 5: EPILOGUE] audit unit readiness handshake failed")
	}
}

// End of tests/unit/data/audit_test.go
