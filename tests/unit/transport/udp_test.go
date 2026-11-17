// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/unit/transport/udp_test.go
// @id            test.unit.transport.udp
// @layer         Layer 4: Tier-2 Unit Verification Suite (Transport Sub-Stratum)
// @archetype     verification_suite | unit_isolation
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Rung 3 Local LAN UDP Transceiver Unit Suite
// @role          Tier-2 Unit Isolation Gate for cog.transport.rung3_udp
// @maturity      Rev 5 Baseline | 4x4 Verification Matrix
// @chassis       bounded-heap | Verifies loopback UDP & 0xFA17 auto-learning
// @scale         O(1) local room UDP unicast & broadcast verification
// @seams         in: data, transport, rung3_local, common | target: udp.go
// @invariants    1. Auto-learns sender UDP address from valid 0xFA17 headers.
//                2. Enforces 512-byte Floppy Sector datagram ceiling.
//                3. Context cancellation exits Receive() cleanly under `-race`.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Dedicated 1:1 Tier-2 Unit isolation suite for `transport/rung3_local/udp.go`.
//
//   ### For Architects, Operators & Non-Programmers:
//   Proves FaithNet's local room UDP intercom auto-discovers neighbors and
//   handles Wi-Fi dropouts cleanly with zero cloud servers.
//
//   ### For Systems Programmers & Engine Architects:
//   100% branch coverage of `UDPTransceiver` under `go test -race`.
// ==============================================================================

package transport_unit_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Identifier, Verification Imports & Tier-2 Unit Sensor Anchor )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.udp
// Mission:       Bind Rung 3 UDPTransceiver and shared test harness.
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
	"github.com/CreativeWorkzStudio/FaithNet/transport/rung3_local"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Operational Suite Constants & Identifiers
// ------------------------------------------------------------------------------

const (
	udpUnitSuiteName = "Tier 2 Unit: cog.transport.rung3_udp (Local LAN UDP)"
)

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.udp
// Mission:       Verify UDP auto-learning, broadcast, fault bounds & 512B cap.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: [ SECTION 1: ALWAYS PASS (NOMINAL GOLDEN SURFACE COGS) ]
// ------------------------------------------------------------------------------

func TestUDPUnit_Sec1_AutoLearnSenderAndBroadcast(t *testing.T) {
	uA, err := rung3_local.NewUDPTransceiver(common.NodeAlpha, rung3_local.DefaultLocalUDPAddr)
	if err != nil {
		t.Fatalf("[SEC 1] bind A failed: %v", err)
	}
	defer func() { _ = uA.Close() }()

	uB, err := rung3_local.NewUDPTransceiver(common.NodeBeta, rung3_local.DefaultLocalUDPAddr)
	if err != nil {
		t.Fatalf("[SEC 1] bind B failed: %v", err)
	}
	defer func() { _ = uB.Close() }()

	if uA.Name() != "rung3:udp" || uA.Rung() != data.Rung3Local || !uA.Healthy() {
		t.Fatal("[SEC 1] unexpected UDPTransceiver metadata")
	}
	if err := uA.RegisterPeer(common.NodeBeta, uB.Addr()); err != nil {
		t.Fatalf("[SEC 1] RegisterPeer failed: %v", err)
	}

	f := common.MakeSovereignFrame(
		common.NodeAlpha, common.NodeBeta, data.Rung3Local, data.OpPing, 1, []byte("HELLO-ROOM"),
	)
	wire := common.EncodeFixture(t, &f)
	if err := uA.Send(data.NodeIDBroadcast, wire); err != nil {
		t.Fatalf("[SEC 1] UDP broadcast failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	env, err := uB.Receive(ctx)
	if err != nil || env.SenderID != common.NodeAlpha {
		t.Fatalf("[SEC 1] B did not auto-learn NodeAlpha: err=%v sender=%s", err, env.SenderID)
	}
	if err := uB.Send(common.NodeAlpha, []byte("AUTO-REPLY")); err != nil {
		t.Fatalf("[SEC 1] B reply to auto-learned NodeAlpha failed: %v", err)
	}
	envA, err := uA.Receive(ctx)
	if err != nil || string(envA.Slice()) != "AUTO-REPLY" {
		t.Fatalf("[SEC 1] A did not receive reply: err=%v got=%q", err, envA.Slice())
	}
	common.AssertAuditPass(t, uA.AuditState(), uA.ExecuteUDPDumpCard())
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: [ SECTION 2: NEVER FAIL (FAULT INVERSION & DEGRADATION) ]
// ------------------------------------------------------------------------------

func TestUDPUnit_Sec2_BadAddrOfflineCanceledClosedAndNilAudit(t *testing.T) {
	var nilU *rung3_local.UDPTransceiver
	common.AssertAuditFail(t, nilU.AuditState(), data.AuditSeverityCritical)

	u, _ := rung3_local.NewUDPTransceiver(common.NodeAlpha, "")
	if err := u.RegisterPeer(common.NodeBeta, "not-a-valid-udp-addr"); err == nil {
		t.Fatal("[SEC 2] invalid UDP address was not rejected")
	}
	if err := u.Send(common.NodeBeta, []byte("UNKNOWN")); err != transport.ErrPeerUnreachable {
		t.Fatalf("[SEC 2] unknown peer got %v", err)
	}
	u.SetHealthy(false)
	if err := u.Send(common.NodeBeta, []byte("OFF")); err != transport.ErrTransceiverOffline {
		t.Fatalf("[SEC 2] offline UDP Send got %v", err)
	}
	u.SetHealthy(true)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := u.Receive(ctx); err != context.Canceled {
		t.Fatalf("[SEC 2] canceled UDP Receive got %v", err)
	}
	_ = u.Close()
	_ = u.Close()
	if err := u.Send(common.NodeBeta, []byte("CLOSED")); err != transport.ErrTransceiverClosed {
		t.Fatalf("[SEC 2] closed UDP Send got %v", err)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: [ SECTION 3: INVARIANTS, SATURATION & LIMIT GUARDS ]
// ------------------------------------------------------------------------------

func TestUDPUnit_Sec3_512ByteCeilingEnforced(t *testing.T) {
	u, _ := rung3_local.NewUDPTransceiver(common.NodeAlpha, "")
	defer func() { _ = u.Close() }()

	var over [data.FloorMaxFrameSize + 1]byte
	if err := u.Send(common.NodeBeta, over[:]); err != transport.ErrWireFrameOversize {
		t.Fatalf("[SEC 3] 513B datagram bypassed UDP ceiling: %v", err)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.4: [ SECTION 4: EMERGENCE & COMPOSITE WORKFLOWS ]
// ------------------------------------------------------------------------------

func TestUDPUnit_Sec4_DumpCardRendersListenAddress(t *testing.T) {
	u, _ := rung3_local.NewUDPTransceiver(common.NodeAlpha, "")
	defer func() { _ = u.Close() }()
	card := u.ExecuteUDPDumpCard()
	if !strings.Contains(card, "RUNG 3 LOCAL UDP") || !strings.Contains(card, u.Addr()) {
		t.Fatalf("[SEC 4] unexpected UDP dump card:\n%s", card)
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Preflight Diagnostic Gates & Physical Source Anchor Verification )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.udp
// Mission:       Run PreflightLocalUDP and verify physical source anchor.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Preflight Diagnostic Gates
// ------------------------------------------------------------------------------

func TestUDPUnit_Block4_PreflightGate(t *testing.T) {
	common.AssertAuditPass(t, rung3_local.PreflightLocalUDP(), "")
	if !rung3_local.AuditLocalUDP() {
		t.Fatal("[BLOCK 4] AuditLocalUDP returned false")
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Suite Identifier & Physical Source Anchor Gate
// ------------------------------------------------------------------------------

func TestUDPUnit_Block4_SourceAnchor(t *testing.T) {
	if udpUnitSuiteName == "" {
		t.Fatal("[BLOCK 4] suite identifier must not be empty")
	}
	common.AssertRepoAnchors(t, "transport/rung3_local/udp.go")
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Test Suite Epilogue & Clean Readiness Handshake )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.udp
// Mission:       Confirm Rung 3 Local UDP readiness handshake post-suite.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Epilogue & Clean State Handshake
// ------------------------------------------------------------------------------

func TestUDPUnit_Block5_EpilogueReadiness(t *testing.T) {
	if !rung3_local.IsLocalUDPReady() || !common.IsHarnessReady() {
		t.Fatal("[BLOCK 5: EPILOGUE] udp unit readiness handshake failed")
	}
}

// End of tests/unit/transport/udp_test.go
