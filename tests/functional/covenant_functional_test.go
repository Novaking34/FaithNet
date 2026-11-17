// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/functional/covenant_functional_test.go
// @id            test.functional.covenant
// @layer         Layer 4: Tier-1 Functional Verification Suite
// @archetype     verification_suite | functional_surface
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Genesis 3 Covenant Airlock Functional Surface Suite
// @role          Tier-1 Functional Diagnostic Sensor (4-Section Micro Protocol)
// @maturity      Rev 5 Baseline | 4x4 Verification Matrix
// @chassis       zero-heap | Verifies ALLOC_ZERO_HEAP airlock inspection
// @scale         O(N_payload * N_sigs) bounded deterministic screening
// @seams         in: data, engine/covenant, tests/common | probes: Preflight
// @invariants    1. Macro Tier 1 (Functional): exercises every public T4 airlock
//                   outcome (+1, 0:AWAIT, -1:REJECT, NULL:HALT).
//                2. External payloads NEVER enter as +1:ALIGNED.
//                3. Zero heap allocations during EvaluateFrame/ExternalPayload.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Tier-1 Functional surface suite exercising the Genesis 3 Covenant Airlock
//   (`engine/covenant/filter.go`) across nominal sovereign/external verdicts,
//   hostile fault inversion, floor MTU bounds, and Block 4/5 audit handshakes.
//
//   ### For Architects, Operators & Non-Programmers:
//   Verifies at the surface that FaithNet's epistemic door welcomes trusted
//   local peers, quarantines clean outside facts safely, and slams shut on
//   forged credentials or web attacks.
//
//   ### For Systems Programmers & Engine Architects:
//   Paired with exhaustive per-signature isolation in `tests/unit/engine/covenant_test.go`.
// ==============================================================================

package functional_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Identifier, Verification Imports & Tier-1 Sensor Anchor )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-1 Functional Verification Suite
// Module:        test.functional.covenant
// Mission:       Bind Layer 0 types, CovenantFilter, and shared harness.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"bytes"
	"strings"
	"testing"
)

// --- Tier 2: External Library Dependencies (None — 100% Go Stdlib Invariant) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---
import (
	"github.com/CreativeWorkzStudio/FaithNet/data"
	"github.com/CreativeWorkzStudio/FaithNet/engine/covenant"
	"github.com/CreativeWorkzStudio/FaithNet/tests/common"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Operational Suite Constants & Identifiers
// ------------------------------------------------------------------------------

const (
	covenantFunctionalSuiteName = "Tier 1: Genesis 3 Covenant Airlock Surface"
)

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-1 Functional Verification Suite
// Module:        test.functional.covenant
// Mission:       Exercise all 4 T4 airlock outcomes across Sections 1-4.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: [ SECTION 1: ALWAYS PASS (NOMINAL GOLDEN SURFACE COGS) ]
// ------------------------------------------------------------------------------

func TestCovenantFunctional_Sec1_NominalSovereignAndCleanExternal(t *testing.T) {
	flt := covenant.NewFilter()

	sov := common.MakeSovereignFrame(
		common.NodeAlpha, common.NodeBeta, data.Rung3Local, data.OpCPISync, 1,
		[]byte("Sovereign local CPI-SI truth delta"),
	)
	vSov := flt.EvaluateFrame(&sov)
	if vSov.State != data.QuadPositive || !vSov.IsPermitted() || vSov.IsBlocked() {
		t.Fatalf("[SEC 1] expected +1:ALIGNED for sovereign frame, got %+v", vSov)
	}

	vExt := flt.EvaluateExternalPayload([]byte("NOAA Weather: 68F Clear"))
	if vExt.State != data.QuadAwait || !vExt.IsPermitted() ||
		vExt.Reason != covenant.ReasonEmpiricalExternalAwait {
		t.Fatalf("[SEC 1] expected 0:AWAIT for clean external payload, got %+v", vExt)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: [ SECTION 2: NEVER FAIL (FAULT INVERSION & DEGRADATION) ]
// ------------------------------------------------------------------------------

func TestCovenantFunctional_Sec2_RejectsForgedHostileAndNilInputs(t *testing.T) {
	flt := covenant.NewFilter()

	if v := flt.EvaluateFrame(nil); v.State != data.QuadNull || !v.IsBlocked() {
		t.Fatalf("[SEC 2] nil frame must yield NULL:HALT, got %+v", v)
	}

	var forged data.Frame
	forged.Rung = data.Rung4Global
	forged.Opcode = data.OpLegacyWrap
	forged.Flags = data.FlagExternalOrigin | data.FlagCovenantVerified
	forged.SourceID = common.NodeExternal
	forged.SetPayload([]byte("forged"))
	if v := flt.EvaluateFrame(&forged); v.State != data.QuadNegative ||
		v.Reason != covenant.ReasonForgedCovenantClaim {
		t.Fatalf("[SEC 2] forged external claim bypassed airlock: %+v", v)
	}

	if v := flt.EvaluateExternalPayload([]byte("<ScRiPt>bad()</script>")); v.State != data.QuadNegative {
		t.Fatalf("[SEC 2] XSS payload bypassed airlock: %+v", v)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: [ SECTION 3: INVARIANTS, SATURATION & LIMIT GUARDS ]
// ------------------------------------------------------------------------------

func TestCovenantFunctional_Sec3_CeilingAndZeroHeapInvariants(t *testing.T) {
	flt := covenant.NewFilter()
	exact := bytes.Repeat([]byte{'A'}, data.FloorMaxPayload)
	over := bytes.Repeat([]byte{'A'}, data.FloorMaxPayload+1)

	if v := flt.EvaluateExternalPayload(exact); v.State != data.QuadAwait {
		t.Fatalf("[SEC 3] exact 480B clean payload rejected: %+v", v)
	}
	if v := flt.EvaluateExternalPayload(over); v.State != data.QuadNegative ||
		v.Reason != covenant.ReasonPayloadOversize {
		t.Fatalf("[SEC 3] 481B external payload bypassed ceiling: %+v", v)
	}

	allocs := testing.AllocsPerRun(500, func() {
		_ = flt.EvaluateExternalPayload(exact)
	})
	if allocs != 0 {
		t.Fatalf("[SEC 3] ALLOC_ZERO_HEAP violated in EvaluateExternalPayload: %.2f", allocs)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.4: [ SECTION 4: EMERGENCE & COMPOSITE WORKFLOWS ]
// ------------------------------------------------------------------------------

func TestCovenantFunctional_Sec4_LegacyBridgeQuarantineWorkflow(t *testing.T) {
	flt := covenant.NewFilter()
	wrapped := common.MakeSovereignFrame(
		common.NodeExternal, common.NodeAlpha, data.Rung4Global,
		data.OpLegacyWrap, 9, []byte("Public Bulletin: Storm Watch"),
	)
	wrapped.Flags = data.FlagExternalOrigin | data.FlagLegacyTranslated

	v := flt.EvaluateFrame(&wrapped)
	if v.State != data.QuadAwait || !v.IsPermitted() {
		t.Fatalf("[SEC 4] clean legacy-wrapped frame must enter as 0:AWAIT, got %+v", v)
	}
	if v.Reason.String() != "EMPIRICAL_EXTERNAL_AWAIT" {
		t.Fatalf("[SEC 4] unexpected reason label: %q", v.Reason.String())
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Preflight Diagnostic Gates & Physical Source Anchor Verification )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-1 Functional Verification Suite
// Module:        test.functional.covenant
// Mission:       Assert Genesis 3 Airlock preflight, policy audit, and anchors.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Preflight Diagnostic Gates
// ------------------------------------------------------------------------------

func TestCovenantFunctional_Block4_PreflightAndDumpCard(t *testing.T) {
	flt := covenant.NewFilter()
	card := flt.ExecuteAirlockDumpCard()
	common.AssertAuditPass(t, flt.AuditPolicy(), card)
	common.AssertAuditPass(t, covenant.PreflightCovenantFilter(), card)
	if !strings.Contains(card, "PASS") {
		t.Fatalf("[BLOCK 4] airlock dump card did not report PASS:\n%s", card)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Suite Identifier & Physical Source Anchor Gate
// ------------------------------------------------------------------------------

func TestCovenantFunctional_Block4_SourceAnchors(t *testing.T) {
	if covenantFunctionalSuiteName == "" {
		t.Fatal("[BLOCK 4] suite identifier must not be empty")
	}
	common.AssertRepoAnchors(t, "engine/covenant/filter.go")
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Test Suite Epilogue & Clean Readiness Handshake )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-1 Functional Verification Suite
// Module:        test.functional.covenant
// Mission:       Verify Genesis 3 Airlock epilogue readiness post-suite.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Epilogue & Clean State Handshake
// ------------------------------------------------------------------------------

func TestCovenantFunctional_Block5_EpilogueReadiness(t *testing.T) {
	if !covenant.IsAirlockReady() || !common.IsHarnessReady() {
		t.Fatal("[BLOCK 5: EPILOGUE] Genesis 3 Airlock readiness handshake failed")
	}
}

// End of tests/functional/covenant_functional_test.go
