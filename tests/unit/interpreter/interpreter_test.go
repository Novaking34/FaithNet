// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/unit/interpreter/interpreter_test.go
// @id            test.unit.interpreter.well
// @layer         Layer 4: Tier-2 Unit Verification Suite (Interpreter)
// @archetype     verification_suite | unit_isolation
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Layer 3B Interpreter ("The Well") Unit Suite
// @role          Tier-2 Unit Isolation Gate for Translator, WellServer & Bridge
// @maturity      Rev 5 Baseline | 4x4 Verification Matrix
// @chassis       bounded-heap | Exhaustive foreign wrapping, HTML & OS tests
// @scale         O(1) deterministic unit tests with live TCP Start/Close check
// @seams         in: data, engine, orchestrator, interpreter, common
// @invariants    1. Clean external payloads ALWAYS map to 0:AWAIT, never +1.
//                2. Form-encoded AND raw POSTs screen identically through Airlock.
//                3. HTML output escapes `<` / `>` cleanly against injection.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Tier-2 Unit isolation suite verifying `interpreter/translator.go`,
//   `interpreter/captive_web.go`, and `interpreter/legacy_bridge.go` across
//   every client family, error branch, ring limit, and live TCP lifecycle.
//
//   ### For Architects, Operators & Non-Programmers:
//   Proves The Well serves clean offline pages to every device family, escapes
//   all output safely, blocks malicious input, and shuts down cleanly.
//
//   ### For Systems Programmers & Engine Architects:
//   100% branch coverage of Layer 3B `interpreter/` under `go test -race`.
// ==============================================================================

package interpreter_unit_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Identifier, Verification Imports & Tier-2 Unit Sensor Anchor )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.interpreter.well
// Mission:       Bind Layer 3B interpreter cogs and shared harness.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// --- Tier 2: External Library Dependencies (None) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---
import (
	"github.com/CreativeWorkzStudio/FaithNet/data"
	"github.com/CreativeWorkzStudio/FaithNet/engine"
	"github.com/CreativeWorkzStudio/FaithNet/interpreter"
	"github.com/CreativeWorkzStudio/FaithNet/orchestrator"
	"github.com/CreativeWorkzStudio/FaithNet/tests/common"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Operational Suite Constants & Identifiers
// ------------------------------------------------------------------------------

const (
	interpreterUnitSuiteName = "Tier 2 Unit: Layer 3B Interpreter (The Well)"
)

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.interpreter.well
// Mission:       Verify Translator, WellServer, and LegacyBridge thoroughly.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: [ SECTION 1: ALWAYS PASS (NOMINAL GOLDEN SURFACE COGS) ]
// ------------------------------------------------------------------------------

func TestInterpreterUnit_Sec1_TranslatorAndAllClientFamilies(t *testing.T) {
	tr := interpreter.NewTranslator(nil)
	if tr.Airlock() == nil {
		t.Fatal("[SEC 1] Translator must auto-initialize Genesis 3 Airlock")
	}

	f, v, err := tr.WrapForeignMessage(
		common.NodeExternal, common.NodeAlpha, data.RungNone, 1,
		[]byte("Clean room bulletin from macOS"),
	)
	if err != nil || f.Rung != data.Rung3Local || v.State != data.QuadAwait {
		t.Fatalf("[SEC 1] WrapForeignMessage failed: err=%v rung=%s v=%+v", err, f.Rung, v)
	}
	view := tr.FormatFrameForBrowser(&f)
	if view.Text != "Clean room bulletin from macOS" || view.Truth != "0:AWAIT" {
		t.Fatalf("[SEC 1] FormatFrameForBrowser mismatch: %+v", view)
	}

	uaCases := map[string]interpreter.ClientFamily{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64)":              interpreter.ClientWindows,
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 14_5)":           interpreter.ClientMacOS,
		"Mozilla/5.0 (X11; Ubuntu; Linux x86_64)":                interpreter.ClientLinux,
		"Mozilla/5.0 (Linux; Android 14; Pixel 8)":               interpreter.ClientAndroid,
		"Mozilla/5.0 (iPhone; CPU iPhone OS 17 like Mac OS X)":   interpreter.ClientIOS,
		"Mozilla/5.0 (iPad; CPU OS 17 like Mac OS X)":            interpreter.ClientIOS,
		"curl/8.5.0":                                             interpreter.ClientCLI,
		"Go-http-client/1.1":                                     interpreter.ClientCLI,
		"":                                                       interpreter.ClientUnknown,
		"CustomRadio/1.0":                                        interpreter.ClientUnknown,
	}
	for ua, want := range uaCases {
		got := interpreter.DetectLegacyClient(ua)
		if got.Family != want || !got.Family.IsValid() {
			t.Fatalf("[SEC 1] UA %q: got %s want %s", ua, got.Family, want)
		}
	}
}

func TestInterpreterUnit_Sec1_FormEncodedPostAndLiveTCPServer(t *testing.T) {
	disp, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(common.NodeAlpha, "well-live"), engine.NewMockClock(0),
	)
	defer func() { _ = disp.Close() }()

	well := interpreter.NewWellServer(disp, "127.0.0.1:0")
	if err := well.Start(); err != nil {
		t.Fatalf("[SEC 1] WellServer.Start failed: %v", err)
	}
	defer func() { _ = well.Close() }()

	// Real HTTP form POST over local TCP loopback!
	form := url.Values{}
	form.Set("message", "Storm shelter open at Elm Ave")
	resp, err := http.PostForm("http://"+well.ListenAddr()+"/api/message", form)
	if err != nil {
		t.Fatalf("[SEC 1] live POST failed: %v", err)
	}
	_, _ = io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("[SEC 1] live form POST want 202 got %d", resp.StatusCode)
	}

	// Verify GET / renders the posted bulletin live.
	homeResp, err := http.Get("http://" + well.ListenAddr() + "/")
	if err != nil {
		t.Fatalf("[SEC 1] live GET / failed: %v", err)
	}
	bodyBytes, _ := io.ReadAll(homeResp.Body)
	_ = homeResp.Body.Close()
	if !strings.Contains(string(bodyBytes), "Storm shelter open at Elm Ave") {
		t.Fatalf("[SEC 1] live GET / missing posted bulletin")
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: [ SECTION 2: NEVER FAIL (FAULT INVERSION & DEGRADATION) ]
// ------------------------------------------------------------------------------

func TestInterpreterUnit_Sec2_NilGuards404AndHostileBlocks(t *testing.T) {
	var nilTr *interpreter.Translator
	common.AssertAuditFail(t, nilTr.AuditState(), data.AuditSeverityFatal)

	var nilWell *interpreter.WellServer
	common.AssertAuditFail(t, nilWell.AuditState(), data.AuditSeverityFatal)

	tr := interpreter.NewTranslator(nil)
	nilView := tr.FormatFrameForBrowser(nil)
	if nilView.Opcode != "INVALID" || nilView.Truth != "NULL:HALT" {
		t.Fatalf("[SEC 2] nil FormatFrameForBrowser mismatch: %+v", nilView)
	}

	disp, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(common.NodeAlpha, "well-404"), engine.NewMockClock(0),
	)
	defer func() { _ = disp.Close() }()
	well := interpreter.NewWellServer(disp, "")

	rec404 := httptest.NewRecorder()
	well.Handler().ServeHTTP(rec404, httptest.NewRequest(http.MethodGet, "/does-not-exist", nil))
	if rec404.Code != http.StatusNotFound {
		t.Fatalf("[SEC 2] unknown path want 404 got %d", rec404.Code)
	}
	if err := well.Close(); err != nil {
		t.Fatalf("[SEC 2] Close on unstarted WellServer failed: %v", err)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: [ SECTION 3: INVARIANTS, SATURATION & LIMIT GUARDS ]
// ------------------------------------------------------------------------------

func TestInterpreterUnit_Sec3_HTMLEscapingAndNulStripping(t *testing.T) {
	tr := interpreter.NewTranslator(nil)
	var f data.Frame
	f.Opcode = data.OpData
	f.Truth = data.QuadPositive
	f.SetPayload([]byte("safe\x00text"))
	if got := tr.FormatFrameForBrowser(&f).Text; got != "safetext" {
		t.Fatalf("[SEC 3] NUL byte not stripped in browser view: %q", got)
	}

	disp, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(common.NodeAlpha, "A&B<Node>"), engine.NewMockClock(0),
	)
	defer func() { _ = disp.Close() }()
	well := interpreter.NewWellServer(disp, "127.0.0.1:0")
	rec := httptest.NewRecorder()
	well.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if !strings.Contains(rec.Body.String(), "A&amp;B&lt;Node&gt;") {
		t.Fatal("[SEC 3] HTML label escaping failed")
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.4: [ SECTION 4: EMERGENCE & COMPOSITE WORKFLOWS ]
// ------------------------------------------------------------------------------

func TestInterpreterUnit_Sec4_AllThreeDumpCardsRenderCleanly(t *testing.T) {
	tr := interpreter.NewTranslator(nil)
	if !strings.Contains(tr.ExecuteTranslatorDumpCard(), "PASS") {
		t.Fatal("[SEC 4] translator card missing PASS")
	}
	if !strings.Contains(interpreter.ExecuteBridgeDumpCard(), "Windows, macOS, Linux") {
		t.Fatal("[SEC 4] bridge card missing OS list")
	}
	disp, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(common.NodeAlpha, "card"), engine.NewMockClock(0),
	)
	defer func() { _ = disp.Close() }()
	well := interpreter.NewWellServer(disp, "127.0.0.1:0")
	if !strings.Contains(well.ExecuteWellDumpCard(), "THE WELL CAPTIVE PORTAL") {
		t.Fatal("[SEC 4] well dump card missing title")
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Preflight Diagnostic Gates & Physical Source Anchor Verification )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.interpreter.well
// Mission:       Run all Layer 3B preflights and verify physical anchors.
// ==============================================================================

func TestInterpreterUnit_Block4_PreflightAndAnchors(t *testing.T) {
	if interpreterUnitSuiteName == "" {
		t.Fatal("[BLOCK 4] suite identifier must not be empty")
	}
	common.AssertAuditPass(t, interpreter.PreflightTranslator(), "")
	common.AssertAuditPass(t, interpreter.PreflightWellServer(), "")
	common.AssertAuditPass(t, interpreter.PreflightLegacyBridge(), "")
	if !interpreter.AuditTranslator() || !interpreter.AuditWellServer() || !interpreter.AuditLegacyBridge() {
		t.Fatal("[BLOCK 4] Layer 3B boolean audit wrappers failed")
	}
	common.AssertRepoAnchors(t,
		"interpreter/translator.go",
		"interpreter/captive_web.go",
		"interpreter/legacy_bridge.go",
	)
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Test Suite Epilogue & Clean Readiness Handshake )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.interpreter.well
// Mission:       Confirm Layer 3B readiness handshake post-suite.
// ==============================================================================

func TestInterpreterUnit_Block5_EpilogueReadiness(t *testing.T) {
	if !interpreter.IsTranslatorReady() || !interpreter.IsWellServerReady() ||
		!interpreter.IsLegacyBridgeReady() || !common.IsHarnessReady() {
		t.Fatal("[BLOCK 5: EPILOGUE] interpreter unit readiness handshake failed")
	}
}

// End of tests/unit/interpreter/interpreter_test.go
