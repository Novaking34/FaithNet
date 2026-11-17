// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/unit/interpreter/legacy_bridge_test.go
// @id            test.unit.interpreter.legacy_bridge
// @layer         Layer 4: Tier-2 Unit Verification Suite (Interpreter)
// @archetype     verification_suite | unit_isolation
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Layer 3B Universal Foreign OS Classifier Unit Suite
// @role          Tier-2 Unit Isolation Gate for cog.interpreter.legacy_bridge
// @maturity      Rev 5 Baseline | 4x4 Verification Matrix
// @chassis       zero-heap | Pure O(N) User-Agent classification into data
// @scale         O(1) exhaustive foreign client OS family test matrix
// @seams         in: data, interpreter, tests/common | target: legacy_bridge.go
// @invariants    1. Every canonical ClientFamily label and predicate verified.
//                2. Mobile tokens (iOS/Android) take precedence over Mac/Linux.
//                3. Unknown/blank UAs degrade cleanly to ClientUnknown.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Dedicated 1:1 Tier-2 Unit isolation suite for `interpreter/legacy_bridge.go`.
//
//   ### For Architects, Operators & Non-Programmers:
//   Confirms FaithNet recognizes Windows, Mac, Linux, Android, iOS, and CLI
//   clients purely as data without contaminating core protocol logic.
//
//   ### For Systems Programmers & Engine Architects:
//   100% branch coverage of `DetectLegacyClient`, `ClientFamily`,
//   `AuditLegacyClientProfile`, `PreflightLegacyBridge`, and dump card.
// ==============================================================================

package interpreter_unit_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Identifier, Verification Imports & Tier-2 Unit Sensor Anchor )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.interpreter.legacy_bridge
// Mission:       Bind LegacyBridge functions and shared test harness.
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
	"github.com/CreativeWorkzStudio/FaithNet/interpreter"
	"github.com/CreativeWorkzStudio/FaithNet/tests/common"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Operational Suite Constants & Identifiers
// ------------------------------------------------------------------------------

const (
	legacyBridgeUnitSuiteName = "Tier 2 Unit: cog.interpreter.legacy_bridge"
)

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.interpreter.legacy_bridge
// Mission:       Exhaustively verify every ClientFamily and UA precedence rule.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: [ SECTION 1: ALWAYS PASS (NOMINAL GOLDEN SURFACE COGS) ]
// ------------------------------------------------------------------------------

func TestLegacyBridgeUnit_Sec1_AllCanonicalClientFamiliesAndStrings(t *testing.T) {
	labels := map[interpreter.ClientFamily]string{
		interpreter.ClientUnknown: "UNKNOWN_OS",
		interpreter.ClientWindows: "WINDOWS",
		interpreter.ClientMacOS:   "MACOS",
		interpreter.ClientLinux:   "LINUX",
		interpreter.ClientAndroid: "ANDROID",
		interpreter.ClientIOS:     "IOS",
		interpreter.ClientCLI:     "CLI_HTTP",
	}
	for fam, want := range labels {
		if !fam.IsValid() || fam.String() != want {
			t.Fatalf("[SEC 1] ClientFamily %d: valid=%v str=%q want %q", fam, fam.IsValid(), fam.String(), want)
		}
		common.AssertAuditPass(t, interpreter.AuditLegacyClientProfile(interpreter.LegacyClientProfile{Family: fam}), "")
	}

	uaCases := map[string]interpreter.ClientFamily{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64)":            interpreter.ClientWindows,
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 14_5)":         interpreter.ClientMacOS,
		"Mozilla/5.0 (X11; Ubuntu; Linux x86_64)":              interpreter.ClientLinux,
		"curl/8.5.0":                                           interpreter.ClientCLI,
		"Wget/1.21":                                            interpreter.ClientCLI,
		"Go-http-client/1.1":                                   interpreter.ClientCLI,
	}
	for ua, want := range uaCases {
		if got := interpreter.DetectLegacyClient(ua).Family; got != want {
			t.Fatalf("[SEC 1] UA %q: got %s want %s", ua, got, want)
		}
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: [ SECTION 2: NEVER FAIL (FAULT INVERSION & DEGRADATION) ]
// ------------------------------------------------------------------------------

func TestLegacyBridgeUnit_Sec2_BlankUnknownAndOOBFamilyDegradation(t *testing.T) {
	if interpreter.DetectLegacyClient("").Family != interpreter.ClientUnknown ||
		interpreter.DetectLegacyClient("   ").Family != interpreter.ClientUnknown ||
		interpreter.DetectLegacyClient("ExoticRadio/0.1").Family != interpreter.ClientUnknown {
		t.Fatal("[SEC 2] blank/unknown UA must degrade to ClientUnknown")
	}
	oob := interpreter.ClientFamily(99)
	if oob.IsValid() || oob.String() != "UNKNOWN_OS" {
		t.Fatal("[SEC 2] OOB ClientFamily did not degrade cleanly")
	}
	common.AssertAuditFail(t, interpreter.AuditLegacyClientProfile(
		interpreter.LegacyClientProfile{Family: oob}), data.AuditSeverityCritical)
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: [ SECTION 3: INVARIANTS, SATURATION & LIMIT GUARDS ]
// ------------------------------------------------------------------------------

func TestLegacyBridgeUnit_Sec3_MobileOverDesktopTokenPrecedence(t *testing.T) {
	ios := interpreter.DetectLegacyClient("Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X)")
	if ios.Family != interpreter.ClientIOS || !ios.SupportsCaptive {
		t.Fatalf("[SEC 3] iPhone UA misclassified as %s", ios.Family)
	}
	ipad := interpreter.DetectLegacyClient("Mozilla/5.0 (iPad; CPU OS 17 like Mac OS X)")
	if ipad.Family != interpreter.ClientIOS {
		t.Fatalf("[SEC 3] iPad UA misclassified as %s", ipad.Family)
	}
	android := interpreter.DetectLegacyClient("Mozilla/5.0 (Linux; Android 14; Pixel 8)")
	if android.Family != interpreter.ClientAndroid || !android.SupportsCaptive {
		t.Fatalf("[SEC 3] Android UA misclassified as %s", android.Family)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.4: [ SECTION 4: EMERGENCE & COMPOSITE WORKFLOWS ]
// ------------------------------------------------------------------------------

func TestLegacyBridgeUnit_Sec4_BridgeDumpCardVerifiesAllOS(t *testing.T) {
	card := interpreter.ExecuteBridgeDumpCard()
	if !strings.Contains(card, "PASS") || !strings.Contains(card, "Windows, macOS, Linux") {
		t.Fatalf("[SEC 4] unexpected LegacyBridge dump card:\n%s", card)
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Preflight Diagnostic Gates & Physical Source Anchor Verification )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.interpreter.legacy_bridge
// Mission:       Run PreflightLegacyBridge and verify physical source anchor.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Preflight Diagnostic Gates
// ------------------------------------------------------------------------------

func TestLegacyBridgeUnit_Block4_PreflightGate(t *testing.T) {
	common.AssertAuditPass(t, interpreter.PreflightLegacyBridge(), "")
	if !interpreter.AuditLegacyBridge() {
		t.Fatal("[BLOCK 4] AuditLegacyBridge returned false")
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Suite Identifier & Physical Source Anchor Gate
// ------------------------------------------------------------------------------

func TestLegacyBridgeUnit_Block4_SourceAnchor(t *testing.T) {
	if legacyBridgeUnitSuiteName == "" {
		t.Fatal("[BLOCK 4] suite identifier must not be empty")
	}
	common.AssertRepoAnchors(t, "interpreter/legacy_bridge.go")
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Test Suite Epilogue & Clean Readiness Handshake )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.interpreter.legacy_bridge
// Mission:       Confirm LegacyBridge epilogue readiness handshake post-suite.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Epilogue & Clean State Handshake
// ------------------------------------------------------------------------------

func TestLegacyBridgeUnit_Block5_EpilogueReadiness(t *testing.T) {
	if !interpreter.IsLegacyBridgeReady() || !common.IsHarnessReady() {
		t.Fatal("[BLOCK 5: EPILOGUE] legacy_bridge unit readiness handshake failed")
	}
}

// End of tests/unit/interpreter/legacy_bridge_test.go
