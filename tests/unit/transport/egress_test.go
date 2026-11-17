// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/unit/transport/egress_test.go
// @id            test.unit.transport.egress
// @layer         Layer 4: Tier-2 Unit Verification Suite (Transport Sub-Stratum)
// @archetype     verification_suite | unit_isolation
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Rung 4 Airlock-Gated EgressGateway Unit Suite
// @role          Tier-2 Unit Isolation Gate for cog.transport.rung4_egress
// @maturity      Rev 5 Baseline | 4x4 Verification Matrix
// @chassis       bounded-heap | Local httptest OOM-bounded & Airlock tests
// @scale         O(N_body <= 481B) deterministic outbound egress verification
// @seams         in: data, engine/covenant, rung4_global, common
// @invariants    1. AllowGlobalEgress==false ALWAYS returns ErrGlobalEgressDisabled.
//                2. Clean empirical responses enter strictly as 0:AWAIT.
//                3. Hostile or >480B remote pages are halted at Airlock.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Dedicated 1:1 Tier-2 Unit isolation suite for `transport/rung4_global/egress.go`.
//
//   ### For Architects, Operators & Non-Programmers:
//   Proves FaithNet's commercial-web airlock stays sealed by default and blocks
//   scripts, database attacks, and oversized web pages whenever opened.
//
//   ### For Systems Programmers & Engine Architects:
//   100% branch coverage of `EgressGateway` against local `httptest.NewServer`.
// ==============================================================================

package transport_unit_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Identifier, Verification Imports & Tier-2 Unit Sensor Anchor )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.egress
// Mission:       Bind Rung 4 EgressGateway and shared test harness.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// --- Tier 2: External Library Dependencies (None) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---
import (
	"github.com/CreativeWorkzStudio/FaithNet/data"
	"github.com/CreativeWorkzStudio/FaithNet/engine/covenant"
	"github.com/CreativeWorkzStudio/FaithNet/tests/common"
	"github.com/CreativeWorkzStudio/FaithNet/transport/rung4_global"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Operational Suite Constants & Identifiers
// ------------------------------------------------------------------------------

const (
	egressUnitSuiteName = "Tier 2 Unit: cog.transport.rung4_egress (Egress Airlock)"
)

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.egress
// Mission:       Verify opt-in lock, 0:AWAIT empirical fetch, XSS & OOM caps.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: [ SECTION 1: ALWAYS PASS (NOMINAL GOLDEN SURFACE COGS) ]
// ------------------------------------------------------------------------------

func TestEgressUnit_Sec1_OptInCleanEmpiricalFetchReturnsAwait(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("NWS Weather: 71F Clear"))
	}))
	defer srv.Close()

	m := data.DefaultNodeManifest(common.NodeAlpha, "optin")
	m.AllowGlobalEgress = true
	m.PreferredMaxRung = data.Rung4Global
	gw := rung4_global.NewEgressGateway(m, nil)

	if !gw.IsEnabled() {
		t.Fatal("[SEC 1] opt-in gateway reported disabled")
	}
	f, v, err := gw.FetchEmpiricalFrame(context.Background(), srv.URL, 9)
	if err != nil || v.State != data.QuadAwait || f.Truth != data.QuadAwait ||
		f.Rung != data.Rung4Global || f.Opcode != data.OpLegacyWrap ||
		string(f.PayloadSlice()) != "NWS Weather: 71F Clear" {
		t.Fatalf("[SEC 1] clean fetch mismatch: err=%v v=%+v f=%+v", err, v, f)
	}
	common.AssertAuditPass(t, gw.AuditState(), gw.ExecuteEgressDumpCard())
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: [ SECTION 2: NEVER FAIL (FAULT INVERSION & DEGRADATION) ]
// ------------------------------------------------------------------------------

func TestEgressUnit_Sec2_SealedOfflineLockHostileResponsesAndNilAudit(t *testing.T) {
	var nilGw *rung4_global.EgressGateway
	common.AssertAuditFail(t, nilGw.AuditState(), data.AuditSeverityFatal)

	sealed := rung4_global.NewEgressGateway(data.DefaultNodeManifest(common.NodeAlpha, "sealed"), nil)
	if sealed.IsEnabled() {
		t.Fatal("[SEC 2] offline default must be disabled")
	}
	if _, _, err := sealed.FetchEmpiricalFrame(context.Background(), "http://127.0.0.1:1/", 1); err != rung4_global.ErrGlobalEgressDisabled {
		t.Fatalf("[SEC 2] sealed egress got %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("<script>pwn()</script>"))
	}))
	defer srv.Close()

	optIn := data.DefaultNodeManifest(common.NodeAlpha, "open")
	optIn.AllowGlobalEgress = true
	optIn.PreferredMaxRung = data.Rung4Global
	gw := rung4_global.NewEgressGateway(optIn, nil)
	if _, v, err := gw.FetchEmpiricalFrame(context.Background(), srv.URL, 2); err != rung4_global.ErrEgressAirlockBlocked || v.State != data.QuadNegative {
		t.Fatalf("[SEC 2] hostile XSS bypassed EgressGateway: err=%v v=%+v", err, v)
	}
	if _, _, err := gw.FetchEmpiricalFrame(context.Background(), "://bad-url", 3); err == nil {
		t.Fatal("[SEC 2] malformed URL did not fail cleanly")
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: [ SECTION 3: INVARIANTS, SATURATION & LIMIT GUARDS ]
// ------------------------------------------------------------------------------

func TestEgressUnit_Sec3_480ByteCapAcceptedAndOversizeOOMBlocked(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/exact" {
			_, _ = w.Write(bytes.Repeat([]byte{'A'}, data.FloorMaxPayload))
			return
		}
		_, _ = w.Write(bytes.Repeat([]byte{'B'}, 4096))
	}))
	defer srv.Close()

	optIn := data.DefaultNodeManifest(common.NodeAlpha, "cap")
	optIn.AllowGlobalEgress = true
	optIn.PreferredMaxRung = data.Rung4Global
	gw := rung4_global.NewEgressGateway(optIn, nil)

	if f, v, err := gw.FetchEmpiricalFrame(context.Background(), srv.URL+"/exact", 1); err != nil || v.State != data.QuadAwait || f.WireSize() != 512 {
		t.Fatalf("[SEC 3] exact 480B response rejected: err=%v size=%d", err, f.WireSize())
	}
	if _, v, err := gw.FetchEmpiricalFrame(context.Background(), srv.URL+"/huge", 2); err != rung4_global.ErrEgressAirlockBlocked || v.Reason != covenant.ReasonPayloadOversize {
		t.Fatalf("[SEC 3] 4096B response bypassed 480B OOM cap: err=%v v=%+v", err, v)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.4: [ SECTION 4: EMERGENCE & COMPOSITE WORKFLOWS ]
// ------------------------------------------------------------------------------

func TestEgressUnit_Sec4_DumpCardReflectsSealedVsOptInState(t *testing.T) {
	sealed := rung4_global.NewEgressGateway(data.DefaultNodeManifest(common.NodeAlpha, "s"), nil)
	if !strings.Contains(sealed.ExecuteEgressDumpCard(), "SEALED_OFFLINE") {
		t.Fatal("[SEC 4] sealed egress card missing SEALED_OFFLINE")
	}
	optIn := data.DefaultNodeManifest(common.NodeAlpha, "o")
	optIn.AllowGlobalEgress = true
	optIn.PreferredMaxRung = data.Rung4Global
	open := rung4_global.NewEgressGateway(optIn, nil)
	if !strings.Contains(open.ExecuteEgressDumpCard(), "OPT_IN_AIRLOCKED") {
		t.Fatal("[SEC 4] opt-in egress card missing OPT_IN_AIRLOCKED")
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Preflight Diagnostic Gates & Physical Source Anchor Verification )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.egress
// Mission:       Run PreflightGlobalEgress and verify physical source anchor.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Preflight Diagnostic Gates
// ------------------------------------------------------------------------------

func TestEgressUnit_Block4_PreflightGate(t *testing.T) {
	common.AssertAuditPass(t, rung4_global.PreflightGlobalEgress(), "")
	if !rung4_global.AuditGlobalEgress() {
		t.Fatal("[BLOCK 4] AuditGlobalEgress returned false")
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Suite Identifier & Physical Source Anchor Gate
// ------------------------------------------------------------------------------

func TestEgressUnit_Block4_SourceAnchor(t *testing.T) {
	if egressUnitSuiteName == "" {
		t.Fatal("[BLOCK 4] suite identifier must not be empty")
	}
	common.AssertRepoAnchors(t, "transport/rung4_global/egress.go")
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Test Suite Epilogue & Clean Readiness Handshake )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.egress
// Mission:       Confirm Rung 4 Egress readiness handshake post-suite.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Epilogue & Clean State Handshake
// ------------------------------------------------------------------------------

func TestEgressUnit_Block5_EpilogueReadiness(t *testing.T) {
	if !rung4_global.IsGlobalEgressReady() || !common.IsHarnessReady() {
		t.Fatal("[BLOCK 5: EPILOGUE] egress unit readiness handshake failed")
	}
}

// End of tests/unit/transport/egress_test.go
