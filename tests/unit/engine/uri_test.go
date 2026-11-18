// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/unit/engine/uri_test.go
// @id            test.unit.engine.uri
// @layer         Layer 4: Tier-2 Unit Verification Suite (Engine Sub-Stratum)
// @archetype     verification_suite | unit_isolation
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Layer 1 FNP/1 Protocol & faithnet:// URI Unit Suite
// @role          Tier-2 Unit Isolation Gate for cog.engine.uri
// @maturity      Rev 5 Baseline | Wave 15 Official FaithNet Protocol & Port 7777
// @chassis       zero-external-deps | Pure Layer 1 URI security & port checks
// @scale         O(N_uri) bounded validation
// @seams         in: data, engine, tests/common | target: engine/uri.go
// @invariants    1. Default canonical wire port is 7777.
//                2. Control chars / CRLF / whitespace / bad schemes rejected.
//                3. Sovereign 64-bit NodeIDs parse and map cleanly to Well URLs.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Strict 1:1 Tier-2 unit verification suite for `engine/uri.go`, verifying
//   canonical `FNP/1` `faithnet://` parsing, formatting, WellBridgeURL mapping,
//   anti-injection security boundaries, and Block 4 self-audit readiness.
// ==============================================================================

package engine_unit_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ==============================================================================

import (
	"errors"
	"strings"
	"testing"

	"github.com/CreativeWorkzStudio/FaithNet/data"
	"github.com/CreativeWorkzStudio/FaithNet/engine"
	"github.com/CreativeWorkzStudio/FaithNet/tests/common"
)

const uriUnitSuiteName = "Tier 2 Unit: cog.engine.uri (FNP/1 Port 7777 & faithnet://)"

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================

// --- Section 1: Always Pass (Nominal Golden Surface) ---

func TestURIUnit_Sec1_NominalParseAndFormat(t *testing.T) {
	u, err := engine.ParseFaithNetURI("faithnet://faithnet.lan/well")
	if err != nil {
		t.Fatalf("[SEC 1] ParseFaithNetURI failed: %v", err)
	}
	if u.Scheme != engine.FaithNetURIScheme || u.Host != "faithnet.lan" ||
		u.Port != engine.FaithNetDefaultPort || u.Action != "well" {
		t.Fatalf("[SEC 1] unexpected parsed URI: %+v", u)
	}
	if got := u.String(); got != "faithnet://faithnet.lan/well" {
		t.Fatalf("[SEC 1] String() mismatch: got %q", got)
	}

	u2, err := engine.ParseFaithNetURI("faithnet://0xFA17000000000001:7778/say/hello")
	if err != nil {
		t.Fatalf("[SEC 1] NodeID URI parse failed: %v", err)
	}
	if !u2.HasNodeID() || u2.NodeID != data.NodeID(0xFA17000000000001) ||
		u2.Port != 7778 || u2.Action != "say" || u2.Target != "hello" {
		t.Fatalf("[SEC 1] unexpected NodeID URI: %+v", u2)
	}
}

// --- Section 2: Never Fail (Anti-Injection & Error Sentinels) ---

func TestURIUnit_Sec2_RejectsInvalidAndHostileURIs(t *testing.T) {
	cases := []struct {
		raw  string
		want error
	}{
		{"", engine.ErrURIEmpty},
		{"http://faithnet.lan/well", engine.ErrURIBadScheme},
		{"faithnet://", engine.ErrURIMissingHost},
		{"faithnet://faithnet.lan:0/well", engine.ErrURIBadPort},
		{"faithnet://faithnet.lan:99999/well", engine.ErrURIBadPort},
		{"faithnet://faithnet.lan:abc/well", engine.ErrURIBadPort},
		{"faithnet://faithnet.lan/exec", engine.ErrURIBadAction},
		{"faithnet://faithnet.lan/well\r\nX-Inject:1", engine.ErrURIUnsafeChars},
		{"faithnet://faithnet.lan/well evil", engine.ErrURIUnsafeChars},
	}
	for _, tc := range cases {
		_, err := engine.ParseFaithNetURI(tc.raw)
		if !errors.Is(err, tc.want) {
			t.Fatalf("[SEC 2] raw=%q: got err=%v, want %v", tc.raw, err, tc.want)
		}
	}
}

// --- Section 3: Invariants & Limits (256-Byte URI Ceiling) ---

func TestURIUnit_Sec3_LengthCeilingEnforced(t *testing.T) {
	tooLong := "faithnet://faithnet.lan/say/" + strings.Repeat("x", engine.MaxFaithNetURILen)
	if _, err := engine.ParseFaithNetURI(tooLong); !errors.Is(err, engine.ErrURITooLong) {
		t.Fatalf("[SEC 3] expected ErrURITooLong, got %v", err)
	}
}

// --- Section 4: Emergence (Well Bridge URLs, Preflight & Spec Card) ---

func TestURIUnit_Sec4_WellBridgePreflightAndSpecCard(t *testing.T) {
	uStatus, err := engine.ParseFaithNetURI("faithnet://home-anchor/status")
	if err != nil {
		t.Fatalf("[SEC 4] parse status failed: %v", err)
	}
	if got := uStatus.WellBridgeURL(true); got != "https://faithnet.lan:8077/api/status" {
		t.Fatalf("[SEC 4] WellBridgeURL(true) mismatch: %q", got)
	}
	uCA, _ := engine.ParseFaithNetURI("faithnet://10.77.77.1/ca")
	if got := uCA.WellBridgeURL(false); got != "http://10.77.77.1:8077/faithnet-root-ca.crt" {
		t.Fatalf("[SEC 4] CA bridge URL mismatch: %q", got)
	}
	common.AssertAuditPass(t, engine.PreflightFaithNetURI(), "PreflightFaithNetURI")
	if !engine.IsFaithNetURIReady() {
		t.Fatalf("[SEC 4] IsFaithNetURIReady returned false")
	}
	card := engine.ExecuteProtocolSpecDumpCard()
	if !strings.Contains(card, "FNP/1") || !strings.Contains(card, "7777/udp") {
		t.Fatalf("[SEC 4] spec card missing FNP/1 or port 7777:\n%s", card)
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ==============================================================================

func TestURIUnit_Block4_SuiteSummaryAudit(t *testing.T) {
	t.Logf("%s: FNP/1 port 7777, faithnet:// parsing & anti-injection verified.", uriUnitSuiteName)
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ==============================================================================
// End of tests/unit/engine/uri_test.go
