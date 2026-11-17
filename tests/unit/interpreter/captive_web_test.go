// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/unit/interpreter/captive_web_test.go
// @id            test.unit.interpreter.captive_web
// @layer         Layer 4: Tier-2 Unit Verification Suite (Interpreter)
// @archetype     verification_suite | unit_isolation
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Layer 3B "The Well" Zero-Install Captive Portal Suite
// @role          Tier-2 Unit Isolation Gate for cog.interpreter.captive_web
// @maturity      Rev 5 Baseline | 4x4 Verification Matrix
// @chassis       bounded-heap | Verifies 32-bulletin ring, HTML escaping & TCP
// @scale         O(1) HTTP handler & live loopback Start/Close verification
// @seams         in: data, engine, orchestrator, interpreter, common
// @invariants    1. Fixed [32]BulletinEntry ring wraps cleanly without leak.
//                2. HTML output escapes `<`, `>`, `&` cleanly.
//                3. Live TCP Start()/Close() shuts down cleanly under `-race`.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Dedicated 1:1 Tier-2 Unit isolation suite for `interpreter/captive_web.go`.
//
//   ### For Architects, Operators & Non-Programmers:
//   Verifies The Well's offline web dashboard, form submission, captive portal
//   redirects, airlock blocking, and clean live server startup/shutdown.
//
//   ### For Systems Programmers & Engine Architects:
//   100% branch coverage of `WellServer` routes, ring buffer, `Start`, `Close`.
// ==============================================================================

package interpreter_unit_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Identifier, Verification Imports & Tier-2 Unit Sensor Anchor )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.interpreter.captive_web
// Mission:       Bind WellServer, Dispatcher, net/http, and shared harness.
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
	captiveWebUnitSuiteName = "Tier 2 Unit: cog.interpreter.captive_web (The Well)"
)

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.interpreter.captive_web
// Mission:       Verify HTML rendering, form POST, 32-slot ring, and live TCP.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: [ SECTION 1: ALWAYS PASS (NOMINAL GOLDEN SURFACE COGS) ]
// ------------------------------------------------------------------------------

func TestCaptiveWebUnit_Sec1_LiveTCPServerFormPostAndStatus(t *testing.T) {
	disp, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(common.NodeAlpha, "well-unit"), engine.NewMockClock(0),
	)
	defer func() { _ = disp.Close() }()

	well := interpreter.NewWellServer(disp, "127.0.0.1:0")
	if err := well.Start(); err != nil {
		t.Fatalf("[SEC 1] WellServer.Start failed: %v", err)
	}
	defer func() { _ = well.Close() }()

	form := url.Values{}
	form.Set("message", "Community well water check passed")
	resp, err := http.PostForm("http://"+well.ListenAddr()+"/api/message", form)
	if err != nil {
		t.Fatalf("[SEC 1] live POST failed: %v", err)
	}
	_, _ = io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("[SEC 1] live form POST want 202 got %d", resp.StatusCode)
	}

	home, err := http.Get("http://" + well.ListenAddr() + "/")
	if err != nil {
		t.Fatalf("[SEC 1] live GET / failed: %v", err)
	}
	body, _ := io.ReadAll(home.Body)
	_ = home.Body.Close()
	if !strings.Contains(string(body), "Community well water check passed") {
		t.Fatal("[SEC 1] GET / missing posted bulletin")
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: [ SECTION 2: NEVER FAIL (FAULT INVERSION & DEGRADATION) ]
// ------------------------------------------------------------------------------

func TestCaptiveWebUnit_Sec2_404MethodNotAllowedXSSAndNilAudit(t *testing.T) {
	var nilWell *interpreter.WellServer
	common.AssertAuditFail(t, nilWell.AuditState(), data.AuditSeverityFatal)

	disp, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(common.NodeAlpha, "well-fault"), engine.NewMockClock(0),
	)
	defer func() { _ = disp.Close() }()
	well := interpreter.NewWellServer(disp, "")

	rec404 := httptest.NewRecorder()
	well.Handler().ServeHTTP(rec404, httptest.NewRequest(http.MethodGet, "/missing", nil))
	if rec404.Code != http.StatusNotFound {
		t.Fatalf("[SEC 2] unknown route want 404 got %d", rec404.Code)
	}
	rec405 := httptest.NewRecorder()
	well.Handler().ServeHTTP(rec405, httptest.NewRequest(http.MethodDelete, "/api/message", nil))
	if rec405.Code != http.StatusMethodNotAllowed {
		t.Fatalf("[SEC 2] DELETE /api/message want 405 got %d", rec405.Code)
	}
	rec403 := httptest.NewRecorder()
	well.Handler().ServeHTTP(rec403, httptest.NewRequest(http.MethodPost, "/api/message",
		strings.NewReader("<script>bad()</script>")))
	if rec403.Code != http.StatusForbidden {
		t.Fatalf("[SEC 2] XSS POST want 403 got %d", rec403.Code)
	}
	if err := well.Close(); err != nil {
		t.Fatalf("[SEC 2] Close on unstarted server failed: %v", err)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: [ SECTION 3: INVARIANTS, SATURATION & LIMIT GUARDS ]
// ------------------------------------------------------------------------------

func TestCaptiveWebUnit_Sec3_HTMLEscapingAnd32SlotRingWrap(t *testing.T) {
	disp, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(common.NodeAlpha, "A&B<Node>"), engine.NewMockClock(0),
	)
	defer func() { _ = disp.Close() }()
	well := interpreter.NewWellServer(disp, "127.0.0.1:0")

	for i := 1; i <= interpreter.MaxWellBulletins+8; i++ {
		well.RecordBulletin(interpreter.BulletinEntry{Seq: uint32(i), Message: "bulletin"})
	}
	var snap [64]interpreter.BulletinEntry
	n := well.BulletinsSnapshotInto(snap[:])
	if n != interpreter.MaxWellBulletins || snap[0].Seq != 9 || snap[n-1].Seq != 40 {
		t.Fatalf("[SEC 3] ring wrap mismatch: n=%d first=%d last=%d", n, snap[0].Seq, snap[n-1].Seq)
	}

	rec := httptest.NewRecorder()
	well.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if !strings.Contains(rec.Body.String(), "A&amp;B&lt;Node&gt;") {
		t.Fatal("[SEC 3] HTML label escaping failed")
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.4: [ SECTION 4: EMERGENCE & COMPOSITE WORKFLOWS ]
// ------------------------------------------------------------------------------

func TestCaptiveWebUnit_Sec4_CaptiveProbesAndDumpCard(t *testing.T) {
	disp, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(common.NodeAlpha, "well-cap"), engine.NewMockClock(0),
	)
	defer func() { _ = disp.Close() }()
	well := interpreter.NewWellServer(disp, "127.0.0.1:0")

	for _, p := range []string{"/generate_204", "/hotspot-detect.html", "/connecttest.txt"} {
		rec := httptest.NewRecorder()
		well.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, p, nil))
		if rec.Code != http.StatusFound || rec.Header().Get("Location") != "/" {
			t.Fatalf("[SEC 4] captive probe %s failed: %d", p, rec.Code)
		}
	}
	card := well.ExecuteWellDumpCard()
	common.AssertAuditPass(t, well.AuditState(), card)
	if !strings.Contains(card, "THE WELL CAPTIVE PORTAL") {
		t.Fatalf("[SEC 4] unexpected Well dump card:\n%s", card)
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Preflight Diagnostic Gates & Physical Source Anchor Verification )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.interpreter.captive_web
// Mission:       Run PreflightWellServer and verify physical source anchor.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Preflight Diagnostic Gates
// ------------------------------------------------------------------------------

func TestCaptiveWebUnit_Block4_PreflightGate(t *testing.T) {
	common.AssertAuditPass(t, interpreter.PreflightWellServer(), "")
	if !interpreter.AuditWellServer() {
		t.Fatal("[BLOCK 4] AuditWellServer returned false")
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Suite Identifier & Physical Source Anchor Gate
// ------------------------------------------------------------------------------

func TestCaptiveWebUnit_Block4_SourceAnchor(t *testing.T) {
	if captiveWebUnitSuiteName == "" {
		t.Fatal("[BLOCK 4] suite identifier must not be empty")
	}
	common.AssertRepoAnchors(t, "interpreter/captive_web.go")
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Test Suite Epilogue & Clean Readiness Handshake )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.interpreter.captive_web
// Mission:       Confirm WellServer epilogue readiness handshake post-suite.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Epilogue & Clean State Handshake
// ------------------------------------------------------------------------------

func TestCaptiveWebUnit_Block5_EpilogueReadiness(t *testing.T) {
	if !interpreter.IsWellServerReady() || !common.IsHarnessReady() {
		t.Fatal("[BLOCK 5: EPILOGUE] captive_web unit readiness handshake failed")
	}
}

// End of tests/unit/interpreter/captive_web_test.go
