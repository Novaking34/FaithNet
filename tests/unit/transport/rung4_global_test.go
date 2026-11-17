// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/unit/transport/rung4_global_test.go
// @id            test.unit.transport.rung4_global
// @layer         Layer 4: Tier-2 Unit Verification Suite (Transport Sub-Stratum)
// @archetype     verification_suite | unit_isolation
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Rung 4 Global TCP & Airlock Egress Unit Suite
// @role          Tier-2 Unit Isolation Gate for cog.transport.rung4_{tcp,egress}
// @maturity      Rev 5 Baseline | 4x4 Verification Matrix
// @chassis       bounded-heap | Local loopback TCP & httptest Egress isolation
// @scale         O(1) deterministic loopback TCP & airlock egress verification
// @seams         in: data, transport, transport/rung4_global, common
// @invariants    1. EgressGateway ALWAYS refuses requests when AllowGlobalEgress
//                   is false.
//                2. TCPTransceiver enforces 512B length prefix bounds cleanly.
//                3. Zero external internet dependency in tests (100% local).
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Tier-2 Unit isolation suite verifying `transport/rung4_global/tcp.go` and
//   `transport/rung4_global/egress.go` across length-prefixed TCP framing,
//   opt-in egress locking, Genesis 3 Airlock screening, and clean teardown.
//
//   ### For Architects, Operators & Non-Programmers:
//   Verifies FaithNet's optional modern-internet compatibility layer stays
//   locked off by default and blocks every web attack when opened.
//
//   ### For Systems Programmers & Engine Architects:
//   100% branch coverage of `TCPTransceiver` and `EgressGateway` under `-race`.
// ==============================================================================

package transport_unit_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Identifier, Verification Imports & Tier-2 Unit Sensor Anchor )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.rung4_global
// Mission:       Bind Rung 4 TCP, EgressGateway, and shared harness.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// --- Tier 2: External Library Dependencies (None) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---
import (
	"github.com/CreativeWorkzStudio/FaithNet/data"
	"github.com/CreativeWorkzStudio/FaithNet/tests/common"
	"github.com/CreativeWorkzStudio/FaithNet/transport"
	"github.com/CreativeWorkzStudio/FaithNet/transport/rung4_global"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Operational Suite Constants & Identifiers
// ------------------------------------------------------------------------------

const (
	rung4GlobalUnitSuiteName = "Tier 2 Unit: Rung 4 Global (TCP & Egress Airlock)"
)

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.rung4_global
// Mission:       Verify TCPTransceiver and EgressGateway across Sections 1-4.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: [ SECTION 1: ALWAYS PASS (NOMINAL TCP & EGRESS AWAIT) ]
// ------------------------------------------------------------------------------

func TestRung4Unit_Sec1_TCPRealFrameAndOptInEgressCleanFact(t *testing.T) {
	tA, err := rung4_global.NewTCPTransceiver(common.NodeAlpha, rung4_global.DefaultGlobalTCPAddr)
	if err != nil {
		t.Fatalf("[SEC 1] TCP bind A failed: %v", err)
	}
	defer func() { _ = tA.Close() }()

	tB, err := rung4_global.NewTCPTransceiver(common.NodeBeta, rung4_global.DefaultGlobalTCPAddr)
	if err != nil {
		t.Fatalf("[SEC 1] TCP bind B failed: %v", err)
	}
	defer func() { _ = tB.Close() }()

	if tA.Name() != "rung4:tcp" || tA.Rung() != data.Rung4Global || !tA.Healthy() {
		t.Fatal("[SEC 1] unexpected TCPTransceiver metadata")
	}

	tA.RegisterPeer(common.NodeBeta, tB.Addr())
	f := common.MakeSovereignFrame(
		common.NodeAlpha, common.NodeBeta, data.Rung4Global, data.OpPing, 4, []byte("GLOBAL-TCP"),
	)
	wire := common.EncodeFixture(t, &f)
	if err := tA.Send(common.NodeBeta, wire); err != nil {
		t.Fatalf("[SEC 1] TCP Send failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	env, err := tB.Receive(ctx)
	if err != nil || env.SenderID != common.NodeAlpha || len(env.Slice()) != len(wire) {
		t.Fatalf("[SEC 1] TCP Receive mismatch: err=%v sender=%s len=%d", err, env.SenderID, len(env.Slice()))
	}
	common.AssertAuditPass(t, tA.AuditState(), tA.ExecuteTCPDumpCard())

	// Opt-in EgressGateway fetching clean local httptest fact -> 0:AWAIT.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("NWS Weather: 71F Partly Cloudy"))
	}))
	defer srv.Close()

	optIn := data.DefaultNodeManifest(common.NodeAlpha, "egress")
	optIn.AllowGlobalEgress = true
	optIn.PreferredMaxRung = data.Rung4Global
	gw := rung4_global.NewEgressGateway(optIn, nil)
	if !gw.IsEnabled() {
		t.Fatal("[SEC 1] opt-in EgressGateway reported disabled")
	}
	gotF, v, err := gw.FetchEmpiricalFrame(context.Background(), srv.URL, 1)
	if err != nil || v.State != data.QuadAwait || string(gotF.PayloadSlice()) != "NWS Weather: 71F Partly Cloudy" {
		t.Fatalf("[SEC 1] clean egress mismatch: err=%v v=%+v payload=%q", err, v, gotF.PayloadSlice())
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: [ SECTION 2: NEVER FAIL (OFFLINE LOCK, XSS & TCP FAULTS) ]
// ------------------------------------------------------------------------------

func TestRung4Unit_Sec2_OfflineEgressLockHostileBlockAndTCPFaults(t *testing.T) {
	// Offline default manifest MUST block egress without even dialing!
	offlineGw := rung4_global.NewEgressGateway(data.DefaultNodeManifest(common.NodeAlpha, "sealed"), nil)
	if offlineGw.IsEnabled() {
		t.Fatal("[SEC 2] offline default EgressGateway reported enabled")
	}
	if _, _, err := offlineGw.FetchEmpiricalFrame(context.Background(), "http://127.0.0.1:1/never", 1); err != rung4_global.ErrGlobalEgressDisabled {
		t.Fatalf("[SEC 2] offline egress did not return ErrGlobalEgressDisabled, got %v", err)
	}

	var nilGw *rung4_global.EgressGateway
	common.AssertAuditFail(t, nilGw.AuditState(), data.AuditSeverityFatal)

	var nilT *rung4_global.TCPTransceiver
	common.AssertAuditFail(t, nilT.AuditState(), data.AuditSeverityCritical)

	tx, _ := rung4_global.NewTCPTransceiver(common.NodeAlpha, "")
	tx.SetHealthy(false)
	if err := tx.Send(common.NodeBeta, []byte("OFF")); err != transport.ErrTransceiverOffline {
		t.Fatalf("[SEC 2] offline TCP Send got %v", err)
	}
	tx.SetHealthy(true)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := tx.Receive(ctx); err != context.Canceled {
		t.Fatalf("[SEC 2] canceled TCP Receive got %v", err)
	}
	_ = tx.Close()
	_ = tx.Close()
	if err := tx.Send(common.NodeBeta, []byte("CLOSED")); err != transport.ErrTransceiverClosed {
		t.Fatalf("[SEC 2] closed TCP Send got %v", err)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: [ SECTION 3: INVARIANTS, SATURATION & LIMIT GUARDS ]
// ------------------------------------------------------------------------------

func TestRung4Unit_Sec3_TCPAndEgressFloorMTUCeiling(t *testing.T) {
	tx, _ := rung4_global.NewTCPTransceiver(common.NodeAlpha, "")
	defer func() { _ = tx.Close() }()

	if err := tx.Send(common.NodeBeta, nil); err != transport.ErrWireFrameOversize {
		t.Fatalf("[SEC 3] empty TCP Send got %v", err)
	}
	var over [data.FloorMaxFrameSize + 1]byte
	if err := tx.Send(common.NodeBeta, over[:]); err != transport.ErrWireFrameOversize {
		t.Fatalf("[SEC 3] 513B TCP Send got %v", err)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.4: [ SECTION 4: EMERGENCE & COMPOSITE WORKFLOWS ]
// ------------------------------------------------------------------------------

func TestRung4Unit_Sec4_DumpCardsRenderSealedAndActiveStates(t *testing.T) {
	tx, _ := rung4_global.NewTCPTransceiver(common.NodeAlpha, "")
	defer func() { _ = tx.Close() }()
	if !strings.Contains(tx.ExecuteTCPDumpCard(), "RUNG 4 GLOBAL TCP") {
		t.Fatal("[SEC 4] TCP dump card missing title")
	}

	sealed := rung4_global.NewEgressGateway(data.DefaultNodeManifest(common.NodeAlpha, "sealed"), nil)
	if !strings.Contains(sealed.ExecuteEgressDumpCard(), "SEALED_OFFLINE") {
		t.Fatal("[SEC 4] sealed egress card missing SEALED_OFFLINE status")
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Preflight Diagnostic Gates & Physical Source Anchor Verification )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.rung4_global
// Mission:       Run Rung 4 TCP & Egress preflights and verify anchors.
// ==============================================================================

func TestRung4Unit_Block4_PreflightAndAnchors(t *testing.T) {
	if rung4GlobalUnitSuiteName == "" {
		t.Fatal("[BLOCK 4] suite identifier must not be empty")
	}
	common.AssertAuditPass(t, rung4_global.PreflightGlobalTCP(), "")
	common.AssertAuditPass(t, rung4_global.PreflightGlobalEgress(), "")
	if !rung4_global.AuditGlobalTCP() || !rung4_global.AuditGlobalEgress() {
		t.Fatal("[BLOCK 4] Rung 4 boolean audit wrappers failed")
	}
	common.AssertRepoAnchors(t,
		"transport/rung4_global/tcp.go",
		"transport/rung4_global/egress.go",
	)
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Test Suite Epilogue & Clean Readiness Handshake )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.rung4_global
// Mission:       Confirm Rung 4 Global readiness handshake post-suite.
// ==============================================================================

func TestRung4Unit_Block5_EpilogueReadiness(t *testing.T) {
	if !rung4_global.IsGlobalTCPReady() || !rung4_global.IsGlobalEgressReady() ||
		!common.IsHarnessReady() {
		t.Fatal("[BLOCK 5: EPILOGUE] Rung 4 unit readiness handshake failed")
	}
}

// End of tests/unit/transport/rung4_global_test.go
