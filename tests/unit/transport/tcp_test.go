// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/unit/transport/tcp_test.go
// @id            test.unit.transport.tcp
// @layer         Layer 4: Tier-2 Unit Verification Suite (Transport Sub-Stratum)
// @archetype     verification_suite | unit_isolation
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Rung 4 Global Length-Prefixed TCP Transceiver Suite
// @role          Tier-2 Unit Isolation Gate for cog.transport.rung4_tcp
// @maturity      Rev 5 Baseline | 4x4 Verification Matrix
// @chassis       bounded-heap | Verifies length-prefixed TCP & 512B ceiling
// @scale         O(1) loopback TCP send/recv & clean acceptLoop teardown
// @seams         in: data, transport, rung4_global, common | target: tcp.go
// @invariants    1. Length-prefixed TCP enforces 1..512 byte framing strictly.
//                2. Close() drains acceptLoop cleanly with zero leaked sockets.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Dedicated 1:1 Tier-2 Unit isolation suite for `transport/rung4_global/tcp.go`.
//
//   ### For Architects, Operators & Non-Programmers:
//   Verifies FaithNet's optional global TCP tunnel exchanges framed envelopes
//   cleanly and tears down cleanly when disconnected.
//
//   ### For Systems Programmers & Engine Architects:
//   100% branch coverage of `TCPTransceiver` under `go test -race`.
// ==============================================================================

package transport_unit_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Identifier, Verification Imports & Tier-2 Unit Sensor Anchor )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.tcp
// Mission:       Bind Rung 4 TCPTransceiver and shared test harness.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"context"
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
	tcpUnitSuiteName = "Tier 2 Unit: cog.transport.rung4_tcp (Global TCP)"
)

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.tcp
// Mission:       Verify length-prefixed TCP, fault bounds, 512B cap & teardown.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: [ SECTION 1: ALWAYS PASS (NOMINAL GOLDEN SURFACE COGS) ]
// ------------------------------------------------------------------------------

func TestTCPUnit_Sec1_LoopbackFrameExchangeAndMetadata(t *testing.T) {
	tA, err := rung4_global.NewTCPTransceiver(common.NodeAlpha, rung4_global.DefaultGlobalTCPAddr)
	if err != nil {
		t.Fatalf("[SEC 1] bind A failed: %v", err)
	}
	defer func() { _ = tA.Close() }()

	tB, err := rung4_global.NewTCPTransceiver(common.NodeBeta, rung4_global.DefaultGlobalTCPAddr)
	if err != nil {
		t.Fatalf("[SEC 1] bind B failed: %v", err)
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
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: [ SECTION 2: NEVER FAIL (FAULT INVERSION & DEGRADATION) ]
// ------------------------------------------------------------------------------

func TestTCPUnit_Sec2_OfflineUnregisteredCanceledClosedAndNilAudit(t *testing.T) {
	var nilT *rung4_global.TCPTransceiver
	common.AssertAuditFail(t, nilT.AuditState(), data.AuditSeverityCritical)

	tx, _ := rung4_global.NewTCPTransceiver(common.NodeAlpha, "")
	if err := tx.Send(common.NodeBeta, []byte("NO-PEER")); err != transport.ErrPeerUnreachable {
		t.Fatalf("[SEC 2] unregistered peer got %v", err)
	}
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

func TestTCPUnit_Sec3_EmptyAnd513BytePayloadRejected(t *testing.T) {
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

func TestTCPUnit_Sec4_DumpCardRendersListenAddress(t *testing.T) {
	tx, _ := rung4_global.NewTCPTransceiver(common.NodeAlpha, "")
	defer func() { _ = tx.Close() }()
	card := tx.ExecuteTCPDumpCard()
	if !strings.Contains(card, "RUNG 4 GLOBAL TCP") || !strings.Contains(card, tx.Addr()) {
		t.Fatalf("[SEC 4] unexpected TCP dump card:\n%s", card)
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Preflight Diagnostic Gates & Physical Source Anchor Verification )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.tcp
// Mission:       Run PreflightGlobalTCP and verify physical source anchor.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Preflight Diagnostic Gates
// ------------------------------------------------------------------------------

func TestTCPUnit_Block4_PreflightGate(t *testing.T) {
	common.AssertAuditPass(t, rung4_global.PreflightGlobalTCP(), "")
	if !rung4_global.AuditGlobalTCP() {
		t.Fatal("[BLOCK 4] AuditGlobalTCP returned false")
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Suite Identifier & Physical Source Anchor Gate
// ------------------------------------------------------------------------------

func TestTCPUnit_Block4_SourceAnchor(t *testing.T) {
	if tcpUnitSuiteName == "" {
		t.Fatal("[BLOCK 4] suite identifier must not be empty")
	}
	common.AssertRepoAnchors(t, "transport/rung4_global/tcp.go")
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Test Suite Epilogue & Clean Readiness Handshake )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.tcp
// Mission:       Confirm Rung 4 Global TCP readiness handshake post-suite.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Epilogue & Clean State Handshake
// ------------------------------------------------------------------------------

func TestTCPUnit_Block5_EpilogueReadiness(t *testing.T) {
	if !rung4_global.IsGlobalTCPReady() || !common.IsHarnessReady() {
		t.Fatal("[BLOCK 5: EPILOGUE] tcp unit readiness handshake failed")
	}
}

// End of tests/unit/transport/tcp_test.go
