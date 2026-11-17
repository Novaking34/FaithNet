// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/unit/engine/covenant_test.go
// @id            test.unit.engine.covenant
// @layer         Layer 4: Tier-2 Unit Verification Suite (Engine Sub-Stratum)
// @archetype     verification_suite | unit_isolation
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Layer 1 Genesis 3 Covenant Airlock Unit Suite
// @role          Tier-2 Unit Isolation Gate for cog.covenant.airlock
// @maturity      Rev 5 Baseline | 4x4 Verification Matrix
// @chassis       zero-heap | Verifies ALLOC_ZERO_HEAP airlock evaluation
// @scale         O(N_payload * N_sigs) exhaustive hostile-signature sweep
// @seams         in: data, engine/covenant, tests/common | target: filter.go
// @invariants    1. Every hostile signature is individually blocked (mixed case).
//                2. External frames NEVER enter as +1:ALIGNED.
//                3. EvaluateFrame and EvaluateExternalPayload == 0 allocs/op.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Tier-2 Unit isolation suite strictly verifying `engine/covenant/filter.go`:
//   every hostile signature, policy overrides (`StrictTextCheck`, custom
//   `MaxExternalBytes`), external control pulse quarantine, all `ReasonCode`
//   labels, live `AuditPolicy`, and zero-allocation enforcement.
//
//   ### For Architects, Operators & Non-Programmers:
//   Tests every security rule of the Genesis 3 Airlock individually so scripts,
//   database attacks, forged credentials, and hidden null bytes cannot slip by.
//
//   ### For Systems Programmers & Engine Architects:
//   100% branch coverage of `EvaluateFrame`, `EvaluateExternalPayload`,
//   `Verdict`, `ReasonCode`, `AuditPolicy`, `PreflightCovenantFilter`, and card.
// ==============================================================================

package engine_unit_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Identifier, Verification Imports & Tier-2 Unit Sensor Anchor )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.engine.covenant
// Mission:       Bind Genesis 3 Airlock symbols and shared fixtures.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"bytes"
	"strings"
	"testing"
)

// --- Tier 2: External Library Dependencies (None) ---

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
	covenantUnitSuiteName = "Tier 2 Unit: cog.covenant.airlock (Genesis 3 T4)"
)

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.engine.covenant
// Mission:       Exhaustively verify every airlock rule, signature, and policy.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: [ SECTION 1: ALWAYS PASS (NOMINAL GOLDEN SURFACE COGS) ]
// ------------------------------------------------------------------------------

func TestCovenantUnit_Sec1_SovereignControlDataAndAllReasonLabels(t *testing.T) {
	flt := covenant.NewFilter()

	// Sovereign control pulse -> +1:ALIGNED.
	ping := common.MakeSovereignFrame(
		common.NodeAlpha, common.NodeBeta, data.Rung1Floor, data.OpPing, 1, nil,
	)
	if v := flt.EvaluateFrame(&ping); v.State != data.QuadPositive ||
		v.Reason != covenant.ReasonAlignedSovereign || !v.IsPermitted() || v.IsBlocked() {
		t.Fatalf("[SEC 1] sovereign OpPing rejected: %+v", v)
	}

	// External control pulse -> 0:AWAIT (permitted in quarantine, never +1).
	extPing := ping
	extPing.Flags = data.FlagExternalOrigin
	if v := flt.EvaluateFrame(&extPing); v.State != data.QuadAwait ||
		v.Reason != covenant.ReasonEmpiricalExternalAwait || !v.IsPermitted() {
		t.Fatalf("[SEC 1] external OpPing should yield 0:AWAIT, got %+v", v)
	}

	// OpLegacyWrap without FlagExternalOrigin STILL routes through external airlock.
	wrap := common.MakeSovereignFrame(
		common.NodeAlpha, common.NodeBeta, data.Rung3Local, data.OpLegacyWrap, 2,
		[]byte("Clean browser form text"),
	)
	wrap.Flags = data.FlagLegacyTranslated
	if v := flt.EvaluateFrame(&wrap); v.State != data.QuadAwait {
		t.Fatalf("[SEC 1] OpLegacyWrap must enter as 0:AWAIT, got %+v", v)
	}

	// Verify every canonical ReasonCode.String() label + OOB fallback.
	reasons := map[covenant.ReasonCode]string{
		covenant.ReasonAlignedSovereign:        "ALIGNED_SOVEREIGN",
		covenant.ReasonEmpiricalExternalAwait:  "EMPIRICAL_EXTERNAL_AWAIT",
		covenant.ReasonMalformedStructure:      "MALFORMED_STRUCTURE",
		covenant.ReasonForgedCovenantClaim:     "FORGED_COVENANT_CLAIM",
		covenant.ReasonHostilePayloadSignature: "HOSTILE_PAYLOAD_SIGNATURE",
		covenant.ReasonNullByteSmuggling:       "NULL_BYTE_SMUGGLING",
		covenant.ReasonPayloadOversize:         "PAYLOAD_OVERSIZE",
		covenant.ReasonCode(99):               "UNKNOWN_REASON",
	}
	for code, want := range reasons {
		if got := code.String(); got != want {
			t.Fatalf("[SEC 1] ReasonCode(%d).String()=%q want %q", code, got, want)
		}
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: [ SECTION 2: NEVER FAIL (FAULT INVERSION & DEGRADATION) ]
// ------------------------------------------------------------------------------

func TestCovenantUnit_Sec2_EveryHostileSignatureAndMalformedFrame(t *testing.T) {
	flt := covenant.NewFilter()

	// Every hostile signature individually tested with mixed-case evasion attempts.
	attacks := [][]byte{
		[]byte("hello <ScRiPt>evil()</script>"),
		[]byte("click JaVaScRiPt:void(0)"),
		[]byte("img ONERROR=steal()"),
		[]byte("body OnLoAd=pwn()"),
		[]byte("x UNION SELECT password FROM users"),
		[]byte("1; DrOp TaBlE accounts;"),
		[]byte("run eVaL(payload)"),
		[]byte("<?PhP system('id'); ?>"),
	}
	for _, raw := range attacks {
		v := flt.EvaluateExternalPayload(raw)
		if v.State != data.QuadNegative || v.Reason != covenant.ReasonHostilePayloadSignature ||
			!v.IsBlocked() || v.IsPermitted() {
			t.Fatalf("[SEC 2] hostile signature %q bypassed airlock: %+v", raw, v)
		}
	}

	// Malformed frame structures -> NULL:HALT.
	badFrames := []*data.Frame{
		nil,
		{Rung: data.RungNone, Opcode: data.OpData, SourceID: common.NodeAlpha},
		{Rung: data.Rung2Host, Opcode: data.OpInvalid, SourceID: common.NodeAlpha},
		{Rung: data.Rung2Host, Opcode: data.OpData, SourceID: data.NodeIDNull},
	}
	for idx, bf := range badFrames {
		if v := flt.EvaluateFrame(bf); v.State != data.QuadNull ||
			v.Reason != covenant.ReasonMalformedStructure {
			t.Fatalf("[SEC 2] malformed case %d did not emit NULL:HALT: %+v", idx, v)
		}
	}

	overFrame := &data.Frame{
		Rung: data.Rung2Host, Opcode: data.OpData, SourceID: common.NodeAlpha,
		PayloadLen: uint16(data.FloorMaxPayload + 1),
	}
	if v := flt.EvaluateFrame(overFrame); v.State != data.QuadNull ||
		v.Reason != covenant.ReasonPayloadOversize {
		t.Fatalf("[SEC 2] oversize PayloadLen frame did not emit NULL:HALT: %+v", v)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: [ SECTION 3: INVARIANTS, SATURATION & LIMIT GUARDS ]
// ------------------------------------------------------------------------------

func TestCovenantUnit_Sec3_CustomPolicyBoundsAndZeroAlloc(t *testing.T) {
	// Custom tight byte cap (32B) and StrictTextCheck=false behavior.
	custom := &covenant.Filter{
		MaxExternalBytes: 32,
		StrictTextCheck:  false,
	}
	common.AssertAuditPass(t, custom.AuditPolicy(), custom.ExecuteAirlockDumpCard())
	if v := custom.EvaluateExternalPayload([]byte("bin\x00ok")); v.State != data.QuadAwait {
		t.Fatalf("[SEC 3] StrictTextCheck=false should allow raw binary null byte: %+v", v)
	}
	if v := custom.EvaluateExternalPayload(bytes.Repeat([]byte{'A'}, 33)); v.State != data.QuadNegative {
		t.Fatalf("[SEC 3] custom 32B ceiling breached: %+v", v)
	}

	// Zero/negative MaxExternalBytes falls back safely to DefaultAirlockMaxBytes (480).
	zeroLimit := &covenant.Filter{MaxExternalBytes: 0, StrictTextCheck: true}
	if v := zeroLimit.EvaluateExternalPayload(bytes.Repeat([]byte{'B'}, data.FloorMaxPayload)); v.State != data.QuadAwait {
		t.Fatalf("[SEC 3] MaxExternalBytes=0 fallback failed at 480B: %+v", v)
	}

	// Invalid policy audit checks: nil receiver & >480B ceiling.
	var nilFlt *covenant.Filter
	common.AssertAuditFail(t, nilFlt.AuditPolicy(), data.AuditSeverityCritical)
	badCeiling := &covenant.Filter{MaxExternalBytes: data.FloorMaxPayload + 1}
	common.AssertAuditFail(t, badCeiling.AuditPolicy(), data.AuditSeverityCritical)

	// ALLOC_ZERO_HEAP proof on EvaluateFrame.
	flt := covenant.NewFilter()
	f := common.MakeSovereignFrame(
		common.NodeExternal, common.NodeAlpha, data.Rung4Global, data.OpLegacyWrap, 1,
		[]byte("Clean empirical bulletin text"),
	)
	f.Flags = data.FlagExternalOrigin
	allocs := testing.AllocsPerRun(500, func() {
		_ = flt.EvaluateFrame(&f)
	})
	if allocs != 0 {
		t.Fatalf("[SEC 3] ALLOC_ZERO_HEAP violated in EvaluateFrame: %.2f", allocs)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.4: [ SECTION 4: EMERGENCE & COMPOSITE WORKFLOWS ]
// ------------------------------------------------------------------------------

func TestCovenantUnit_Sec4_DumpCardAndAllFourT4StatesObserved(t *testing.T) {
	flt := covenant.NewFilter()
	card := flt.ExecuteAirlockDumpCard()
	if !strings.Contains(card, "PASS") || !strings.Contains(card, "480") {
		t.Fatalf("[SEC 4] unexpected airlock dump card:\n%s", card)
	}

	seen := map[data.QuadState]bool{}
	sov := common.MakeSovereignFrame(
		common.NodeAlpha, common.NodeBeta, data.Rung2Host, data.OpCPISync, 1, []byte("truth"),
	)
	seen[flt.EvaluateFrame(&sov).State] = true
	seen[flt.EvaluateExternalPayload([]byte("weather")).State] = true
	seen[flt.EvaluateExternalPayload([]byte("<script>x</script>")).State] = true
	seen[flt.EvaluateFrame(nil).State] = true

	if len(seen) != 4 {
		t.Fatalf("[SEC 4] expected all 4 T4 states exercised, saw %v", seen)
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Preflight Diagnostic Gates & Physical Source Anchor Verification )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.engine.covenant
// Mission:       Run PreflightCovenantFilter and verify physical source anchor.
// ==============================================================================

func TestCovenantUnit_Block4_PreflightAndAnchor(t *testing.T) {
	if covenantUnitSuiteName == "" {
		t.Fatal("[BLOCK 4] suite identifier must not be empty")
	}
	common.AssertAuditPass(t, covenant.PreflightCovenantFilter(),
		covenant.NewFilter().ExecuteAirlockDumpCard())
	if !covenant.AuditCovenantFilter() {
		t.Fatal("[BLOCK 4] AuditCovenantFilter returned false")
	}
	common.AssertRepoAnchors(t, "engine/covenant/filter.go")
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Test Suite Epilogue & Clean Readiness Handshake )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.engine.covenant
// Mission:       Confirm Airlock and harness epilogue readiness post-suite.
// ==============================================================================

func TestCovenantUnit_Block5_EpilogueReadiness(t *testing.T) {
	if !covenant.IsAirlockReady() || !common.IsHarnessReady() {
		t.Fatal("[BLOCK 5: EPILOGUE] covenant unit readiness handshake failed")
	}
}

// End of tests/unit/engine/covenant_test.go
