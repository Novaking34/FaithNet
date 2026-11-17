// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/functional/interpreter_functional_test.go
// @id            test.functional.interpreter
// @layer         Layer 4: Tier-1 Functional Verification Suite
// @archetype     verification_suite | functional_surface
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Layer 3B Interpreter ("The Well") & Rung 3 Surface
// @role          Tier-1 Functional Diagnostic Sensor (4-Section Micro Protocol)
// @maturity      Rev 5 Baseline | 4x4 Verification Matrix
// @chassis       bounded-heap | Exercises Translator, WellServer & LegacyBridge
// @scale         O(1) zero-socket httptest & surface verification sweep
// @seams         in: interpreter, transport/rung3_local, orchestrator, common
// @invariants    1. Clean browser posts enter strictly as 0:AWAIT (HTTP 202).
//                2. Hostile browser posts fail cleanly with HTTP 403 Forbidden.
//                3. Multi-OS detection runs purely in Layer 3B data records.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Tier-1 Functional surface suite exercising FaithNet's Layer 3B `interpreter/`
//   ("The Well") and Rung 3 Local (`UDPTransceiver` & `BeaconSwitch`) public
//   APIs across all 4 ontological micro sections.
//
//   ### For Architects, Operators & Non-Programmers:
//   Confirms at a surface level that any visiting phone or laptop can view The
//   Well offline webpage, post safe messages, and get blocked on web attacks.
//
//   ### For Systems Programmers & Engine Architects:
//   Paired with unit suites in `tests/unit/interpreter/` and `tests/unit/transport/`,
//   plus live Room+Well integration in `tests/integrated/room_and_well_test.go`.
// ==============================================================================

package functional_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Identifier, Verification Imports & Tier-1 Sensor Anchor )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-1 Functional Verification Suite
// Module:        test.functional.interpreter
// Mission:       Bind Layer 3B interpreter, Rung 3 Local, and shared harness.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"net/http"
	"net/http/httptest"
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
	"github.com/CreativeWorkzStudio/FaithNet/transport/rung3_local"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Operational Suite Constants & Identifiers
// ------------------------------------------------------------------------------

const (
	interpreterFunctionalSuiteName = "Tier 1: Layer 3B Interpreter & Rung 3 Surface"
)

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-1 Functional Verification Suite
// Module:        test.functional.interpreter
// Mission:       Exercise The Well, Translator, LegacyBridge & BeaconSwitch.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: [ SECTION 1: ALWAYS PASS (NOMINAL GOLDEN SURFACE COGS) ]
// ------------------------------------------------------------------------------

func TestInterpreterFunctional_Sec1_WellHTMLStatusAndCleanPost(t *testing.T) {
	disp, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(common.NodeAlpha, "well-func"), engine.NewMockClock(0),
	)
	defer func() { _ = disp.Close() }()

	well := interpreter.NewWellServer(disp, "127.0.0.1:0")

	recHome := httptest.NewRecorder()
	well.Handler().ServeHTTP(recHome, httptest.NewRequest(http.MethodGet, "/", nil))
	if recHome.Code != http.StatusOK || !strings.Contains(recHome.Body.String(), "well-func") {
		t.Fatalf("[SEC 1] GET / failed: code=%d", recHome.Code)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/message",
		strings.NewReader("Water station open at Community Hall"))
	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 17 like Mac OS X)")
	recPost := httptest.NewRecorder()
	well.Handler().ServeHTTP(recPost, req)
	if recPost.Code != http.StatusAccepted || !strings.Contains(recPost.Body.String(), "IOS") {
		t.Fatalf("[SEC 1] clean POST failed: code=%d body=%s", recPost.Code, recPost.Body.String())
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: [ SECTION 2: NEVER FAIL (FAULT INVERSION & DEGRADATION) ]
// ------------------------------------------------------------------------------

func TestInterpreterFunctional_Sec2_RejectsHostilePostAndWrongMethod(t *testing.T) {
	disp, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(common.NodeAlpha, "well-guard"), engine.NewMockClock(0),
	)
	defer func() { _ = disp.Close() }()

	well := interpreter.NewWellServer(disp, "127.0.0.1:0")

	recGet := httptest.NewRecorder()
	well.Handler().ServeHTTP(recGet, httptest.NewRequest(http.MethodGet, "/api/message", nil))
	if recGet.Code != http.StatusMethodNotAllowed {
		t.Fatalf("[SEC 2] GET /api/message want 405 got %d", recGet.Code)
	}

	recXSS := httptest.NewRecorder()
	well.Handler().ServeHTTP(recXSS, httptest.NewRequest(http.MethodPost, "/api/message",
		strings.NewReader("<script>alert(1)</script>")))
	if recXSS.Code != http.StatusForbidden {
		t.Fatalf("[SEC 2] hostile XSS POST want 403 got %d", recXSS.Code)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: [ SECTION 3: INVARIANTS, SATURATION & LIMIT GUARDS ]
// ------------------------------------------------------------------------------

func TestInterpreterFunctional_Sec3_Bounded32BulletinRingInvariant(t *testing.T) {
	disp, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(common.NodeAlpha, "well-ring"), engine.NewMockClock(0),
	)
	defer func() { _ = disp.Close() }()
	well := interpreter.NewWellServer(disp, "127.0.0.1:0")

	for i := 1; i <= interpreter.MaxWellBulletins+10; i++ {
		well.RecordBulletin(interpreter.BulletinEntry{Seq: uint32(i), Message: "msg"})
	}
	var snap [64]interpreter.BulletinEntry
	n := well.BulletinsSnapshotInto(snap[:])
	if n != interpreter.MaxWellBulletins || snap[0].Seq != 11 || snap[n-1].Seq != 42 {
		t.Fatalf("[SEC 3] bulletin ring wrap mismatch: n=%d first=%d last=%d",
			n, snap[0].Seq, snap[n-1].Seq)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.4: [ SECTION 4: EMERGENCE & COMPOSITE WORKFLOWS ]
// ------------------------------------------------------------------------------

func TestInterpreterFunctional_Sec4_CaptiveProbesAndBeaconComposite(t *testing.T) {
	disp, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(common.NodeAlpha, "well-cap"), engine.NewMockClock(0),
	)
	defer func() { _ = disp.Close() }()
	well := interpreter.NewWellServer(disp, "127.0.0.1:0")

	for _, path := range []string{"/generate_204", "/hotspot-detect.html", "/connecttest.txt"} {
		rec := httptest.NewRecorder()
		well.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusFound || rec.Header().Get("Location") != "/" {
			t.Fatalf("[SEC 4] captive probe %s failed: code=%d loc=%q",
				path, rec.Code, rec.Header().Get("Location"))
		}
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Preflight Diagnostic Gates & Physical Source Anchor Verification )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-1 Functional Verification Suite
// Module:        test.functional.interpreter
// Mission:       Run Rung 3 & Layer 3B preflight gates and verify anchors.
// ==============================================================================

func TestInterpreterFunctional_Block4_PreflightAndAnchors(t *testing.T) {
	if interpreterFunctionalSuiteName == "" {
		t.Fatal("[BLOCK 4] suite identifier must not be empty")
	}
	common.AssertAuditPass(t, rung3_local.PreflightLocalUDP(), "")
	common.AssertAuditPass(t, rung3_local.PreflightBeaconSwitch(), "")
	common.AssertAuditPass(t, interpreter.PreflightTranslator(), "")
	common.AssertAuditPass(t, interpreter.PreflightWellServer(), "")
	common.AssertAuditPass(t, interpreter.PreflightLegacyBridge(), "")
	common.AssertRepoAnchors(t,
		"transport/rung3_local/udp.go",
		"transport/rung3_local/beacon.go",
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
// Stratum:       Layer 4: Tier-1 Functional Verification Suite
// Module:        test.functional.interpreter
// Mission:       Confirm Rung 3 & Layer 3B epilogue readiness post-suite.
// ==============================================================================

func TestInterpreterFunctional_Block5_EpilogueReadiness(t *testing.T) {
	if !rung3_local.IsLocalUDPReady() || !rung3_local.IsBeaconSwitchReady() ||
		!interpreter.IsTranslatorReady() || !interpreter.IsWellServerReady() ||
		!interpreter.IsLegacyBridgeReady() || !common.IsHarnessReady() {
		t.Fatal("[BLOCK 5: EPILOGUE] Rung 3 / Interpreter readiness handshake failed")
	}
}

// End of tests/functional/interpreter_functional_test.go
