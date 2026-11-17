// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/unit/interpreter/translator_test.go
// @id            test.unit.interpreter.translator
// @layer         Layer 4: Tier-2 Unit Verification Suite (Interpreter)
// @archetype     verification_suite | unit_isolation
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Layer 3B Foreign <-> Sovereign Frame Translator Suite
// @role          Tier-2 Unit Isolation Gate for cog.interpreter.translator
// @maturity      Rev 5 Baseline | 4x4 Verification Matrix
// @chassis       bounded-heap | Exercises Genesis 3 Airlock wrapping & browser view
// @scale         O(N_payload) deterministic foreign-to-sovereign translation
// @seams         in: data, engine/covenant, interpreter, common
// @invariants    1. Clean foreign bytes enter strictly as 0:AWAIT (never +1).
//                2. Hostile XSS/SQLi/null/oversize payloads return ErrForeignPayloadBlocked.
//                3. Browser views strip NUL bytes cleanly and handle nil Frame.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Dedicated 1:1 Tier-2 Unit isolation suite for `interpreter/translator.go`.
//
//   ### For Architects, Operators & Non-Programmers:
//   Verifies The Well's message translator checks every foreign message through
//   the security airlock and formats replies safely for any web browser.
//
//   ### For Systems Programmers & Engine Architects:
//   100% branch coverage of `WrapForeignMessage`, `FormatFrameForBrowser`,
//   `Airlock`, `AuditState`, `PreflightTranslator`, and `ExecuteTranslatorDumpCard`.
// ==============================================================================

package interpreter_unit_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Identifier, Verification Imports & Tier-2 Unit Sensor Anchor )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.interpreter.translator
// Mission:       Bind Translator, CovenantFilter, and shared test harness.
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
	"github.com/CreativeWorkzStudio/FaithNet/interpreter"
	"github.com/CreativeWorkzStudio/FaithNet/tests/common"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Operational Suite Constants & Identifiers
// ------------------------------------------------------------------------------

const (
	translatorUnitSuiteName = "Tier 2 Unit: cog.interpreter.translator"
)

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.interpreter.translator
// Mission:       Verify foreign wrapping, fault blocks, 480B cap, and views.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: [ SECTION 1: ALWAYS PASS (NOMINAL GOLDEN SURFACE COGS) ]
// ------------------------------------------------------------------------------

func TestTranslatorUnit_Sec1_NominalWrapAndBrowserView(t *testing.T) {
	tr := interpreter.NewTranslator(nil)
	if tr.Airlock() == nil {
		t.Fatal("[SEC 1] Translator must auto-initialize default Airlock")
	}
	f, v, err := tr.WrapForeignMessage(
		common.NodeExternal, common.NodeAlpha, data.RungNone, 1,
		[]byte("Clean room update from Linux laptop"),
	)
	if err != nil || f.Rung != data.Rung3Local || v.State != data.QuadAwait ||
		f.Truth != data.QuadAwait || f.Opcode != data.OpLegacyWrap ||
		f.Flags.Has(data.FlagCovenantVerified) {
		t.Fatalf("[SEC 1] WrapForeignMessage mismatch: err=%v f=%+v v=%+v", err, f, v)
	}
	view := tr.FormatFrameForBrowser(&f)
	if view.Text != "Clean room update from Linux laptop" || view.Truth != "0:AWAIT" ||
		view.Opcode != "LEGACY_WRAP" || view.SeqID != 1 {
		t.Fatalf("[SEC 1] FormatFrameForBrowser mismatch: %+v", view)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: [ SECTION 2: NEVER FAIL (FAULT INVERSION & DEGRADATION) ]
// ------------------------------------------------------------------------------

func TestTranslatorUnit_Sec2_RejectsEmptyHostileAndNilInputs(t *testing.T) {
	tr := interpreter.NewTranslator(covenant.NewFilter())

	if _, _, err := tr.WrapForeignMessage(common.NodeExternal, common.NodeAlpha, data.Rung3Local, 1, []byte("  \t\n ")); err != interpreter.ErrEmptyForeignPayload {
		t.Fatalf("[SEC 2] whitespace payload got %v", err)
	}
	for _, atk := range [][]byte{
		[]byte("<script>x</script>"),
		[]byte("1; DROP TABLE users;"),
		[]byte("bad\x00smuggle"),
	} {
		if _, v, err := tr.WrapForeignMessage(common.NodeExternal, common.NodeAlpha, data.Rung3Local, 2, atk); err != interpreter.ErrForeignPayloadBlocked || !v.IsBlocked() {
			t.Fatalf("[SEC 2] hostile %q bypassed Translator: err=%v v=%+v", atk, err, v)
		}
	}

	nilView := tr.FormatFrameForBrowser(nil)
	if nilView.Opcode != "INVALID" || nilView.Truth != "NULL:HALT" {
		t.Fatalf("[SEC 2] nil FormatFrameForBrowser mismatch: %+v", nilView)
	}
	var nilTr *interpreter.Translator
	common.AssertAuditFail(t, nilTr.AuditState(), data.AuditSeverityFatal)
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: [ SECTION 3: INVARIANTS, SATURATION & LIMIT GUARDS ]
// ------------------------------------------------------------------------------

func TestTranslatorUnit_Sec3_480ByteCeilingAndNulStrip(t *testing.T) {
	tr := interpreter.NewTranslator(nil)
	exact := bytes.Repeat([]byte{'A'}, data.FloorMaxPayload)
	over := bytes.Repeat([]byte{'A'}, data.FloorMaxPayload+1)

	if f, _, err := tr.WrapForeignMessage(common.NodeExternal, common.NodeAlpha, data.Rung3Local, 1, exact); err != nil || f.WireSize() != 512 {
		t.Fatalf("[SEC 3] exact 480B clean payload failed: err=%v size=%d", err, f.WireSize())
	}
	if _, _, err := tr.WrapForeignMessage(common.NodeExternal, common.NodeAlpha, data.Rung3Local, 2, over); err != interpreter.ErrForeignPayloadBlocked {
		t.Fatalf("[SEC 3] 481B foreign payload bypassed ceiling: %v", err)
	}

	var f data.Frame
	f.Opcode = data.OpData
	f.Truth = data.QuadPositive
	f.SetPayload([]byte("ab\x00cd"))
	if got := tr.FormatFrameForBrowser(&f).Text; got != "abcd" {
		t.Fatalf("[SEC 3] NUL strip failed: %q", got)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.4: [ SECTION 4: EMERGENCE & COMPOSITE WORKFLOWS ]
// ------------------------------------------------------------------------------

func TestTranslatorUnit_Sec4_LiveAuditAndDumpCard(t *testing.T) {
	tr := interpreter.NewTranslator(nil)
	card := tr.ExecuteTranslatorDumpCard()
	common.AssertAuditPass(t, tr.AuditState(), card)
	if !strings.Contains(card, "PASS") || !strings.Contains(card, "0:AWAIT") {
		t.Fatalf("[SEC 4] unexpected translator dump card:\n%s", card)
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Preflight Diagnostic Gates & Physical Source Anchor Verification )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.interpreter.translator
// Mission:       Run PreflightTranslator and verify physical source anchor.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Preflight Diagnostic Gates
// ------------------------------------------------------------------------------

func TestTranslatorUnit_Block4_PreflightGate(t *testing.T) {
	common.AssertAuditPass(t, interpreter.PreflightTranslator(), "")
	if !interpreter.AuditTranslator() {
		t.Fatal("[BLOCK 4] AuditTranslator returned false")
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Suite Identifier & Physical Source Anchor Gate
// ------------------------------------------------------------------------------

func TestTranslatorUnit_Block4_SourceAnchor(t *testing.T) {
	if translatorUnitSuiteName == "" {
		t.Fatal("[BLOCK 4] suite identifier must not be empty")
	}
	common.AssertRepoAnchors(t, "interpreter/translator.go")
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Test Suite Epilogue & Clean Readiness Handshake )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.interpreter.translator
// Mission:       Confirm Translator epilogue readiness handshake post-suite.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Epilogue & Clean State Handshake
// ------------------------------------------------------------------------------

func TestTranslatorUnit_Block5_EpilogueReadiness(t *testing.T) {
	if !interpreter.IsTranslatorReady() || !common.IsHarnessReady() {
		t.Fatal("[BLOCK 5: EPILOGUE] translator unit readiness handshake failed")
	}
}

// End of tests/unit/interpreter/translator_test.go
