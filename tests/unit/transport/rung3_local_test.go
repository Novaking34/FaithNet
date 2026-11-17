// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/unit/transport/rung3_local_test.go
// @id            test.unit.transport.rung3_local
// @layer         Layer 4: Tier-2 Unit Verification Suite (Transport Sub-Stratum)
// @archetype     verification_suite | unit_isolation
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Rung 3 Local UDP & Room Beacon Switch Unit Suite
// @role          Tier-2 Unit Isolation Gate for cog.transport.rung3_{udp,beacon}
// @maturity      Rev 5 Baseline | 4x4 Verification Matrix
// @chassis       bounded-heap | Isolated loopback UDP & simulated BeaconSwitch
// @scale         O(1) deterministic loopback datagrams & beacon toggles
// @seams         in: data, engine, transport, transport/rung3_local, common
// @invariants    1. Auto-learning UDP peer endpoint works on valid 0xFA17 header.
//                2. BeaconSwitch never mutates host OS Wi-Fi in simulated mode.
//                3. All IEEE 802.11 SSID (1..32B) and channel (1..14) bounds hold.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Tier-2 Unit isolation suite strictly verifying `transport/rung3_local/udp.go`
//   and `transport/rung3_local/beacon.go` across unicast/broadcast UDP, auto-
//   learning return addresses, fault inversion, IEEE bounds, and dump cards.
//
//   ### For Architects, Operators & Non-Programmers:
//   Tests FaithNet's local room intercom (UDP) and room campfire switch (Beacon)
//   thoroughly on loopback without ever touching live system Wi-Fi settings.
//
//   ### For Systems Programmers & Engine Architects:
//   100% branch coverage of `UDPTransceiver` and `BeaconSwitch` under `-race`.
// ==============================================================================

package transport_unit_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Identifier, Verification Imports & Tier-2 Unit Sensor Anchor )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.rung3_local
// Mission:       Bind Rung 3 UDP/Beacon types and shared harness.
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
	"github.com/CreativeWorkzStudio/FaithNet/engine"
	"github.com/CreativeWorkzStudio/FaithNet/tests/common"
	"github.com/CreativeWorkzStudio/FaithNet/transport"
	"github.com/CreativeWorkzStudio/FaithNet/transport/rung3_local"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Operational Suite Constants & Identifiers
// ------------------------------------------------------------------------------

const (
	rung3LocalUnitSuiteName = "Tier 2 Unit: Rung 3 Local (UDP & Room Beacon)"
)

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.rung3_local
// Mission:       Verify UDP auto-learning, broadcast, fault bounds, and Beacon.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: [ SECTION 1: ALWAYS PASS (NOMINAL GOLDEN SURFACE COGS) ]
// ------------------------------------------------------------------------------

func TestRung3Unit_Sec1_UDPAutoLearnPeerAndBroadcast(t *testing.T) {
	uA, err := rung3_local.NewUDPTransceiver(common.NodeAlpha, rung3_local.DefaultLocalUDPAddr)
	if err != nil {
		t.Fatalf("[SEC 1] bind UDP A failed: %v", err)
	}
	defer func() { _ = uA.Close() }()

	uB, err := rung3_local.NewUDPTransceiver(common.NodeBeta, rung3_local.DefaultLocalUDPAddr)
	if err != nil {
		t.Fatalf("[SEC 1] bind UDP B failed: %v", err)
	}
	defer func() { _ = uB.Close() }()

	if uA.Name() != "rung3:udp" || uA.Rung() != data.Rung3Local || !uA.Healthy() {
		t.Fatal("[SEC 1] unexpected UDPTransceiver metadata")
	}

	// Only A knows B initially; B MUST auto-learn A's UDP address from 0xFA17 frame!
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
	// Now B can reply straight back to NodeAlpha without manual RegisterPeer!
	if err := uB.Send(common.NodeAlpha, []byte("AUTO-REPLY")); err != nil {
		t.Fatalf("[SEC 1] B reply to auto-learned NodeAlpha failed: %v", err)
	}
	envA, err := uA.Receive(ctx)
	if err != nil || string(envA.Slice()) != "AUTO-REPLY" {
		t.Fatalf("[SEC 1] A did not receive auto-learned reply: err=%v got=%q", err, envA.Slice())
	}
	common.AssertAuditPass(t, uA.AuditState(), uA.ExecuteUDPDumpCard())
}

func TestRung3Unit_Sec1_BeaconSwitchNominalToggleAndFrame(t *testing.T) {
	clk := engine.NewMockClock(500)
	sw, err := rung3_local.NewBeaconSwitch("LivingRoom-Mesh", 11, rung3_local.BeaconModeSimulated, clk)
	if err != nil {
		t.Fatalf("[SEC 1] NewBeaconSwitch failed: %v", err)
	}
	defer func() { _ = sw.Close() }()

	if sw.IsActive() {
		t.Fatal("[SEC 1] new BeaconSwitch must start OFF")
	}
	if err := sw.TurnOn(); err != nil || !sw.IsActive() {
		t.Fatalf("[SEC 1] TurnOn failed: %v", err)
	}
	// Idempotent TurnOn does not double-increment ToggleCount.
	_ = sw.TurnOn()
	st := sw.StatusSnapshot()
	if st.ToggleCount != 1 || st.StartedAtNano != 500 || st.Mode.String() != "SIMULATED_SAFE" {
		t.Fatalf("[SEC 1] unexpected BeaconStatus: %+v", st)
	}
	bf := sw.BuildBeaconFrame(common.NodeAlpha, data.Rung3Local, 7)
	if bf.Opcode != data.OpBeacon || !strings.Contains(string(bf.PayloadSlice()), "LivingRoom-Mesh") {
		t.Fatalf("[SEC 1] unexpected beacon frame: %+v", bf)
	}
	sw.TurnOff()
	sw.TurnOff() // Idempotent TurnOff.
	if sw.IsActive() || sw.StatusSnapshot().ToggleCount != 2 {
		t.Fatalf("[SEC 1] unexpected post-off status: %+v", sw.StatusSnapshot())
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: [ SECTION 2: NEVER FAIL (FAULT INVERSION & DEGRADATION) ]
// ------------------------------------------------------------------------------

func TestRung3Unit_Sec2_FaultInversionUDPAndBeacon(t *testing.T) {
	var nilU *rung3_local.UDPTransceiver
	common.AssertAuditFail(t, nilU.AuditState(), data.AuditSeverityCritical)

	u, _ := rung3_local.NewUDPTransceiver(common.NodeAlpha, rung3_local.DefaultLocalUDPAddr)
	if err := u.RegisterPeer(common.NodeBeta, "not-a-valid-udp-addr"); err == nil {
		t.Fatal("[SEC 2] invalid UDP address was not rejected")
	}
	u.SetHealthy(false)
	if err := u.Send(common.NodeBeta, []byte("OFFLINE")); err != transport.ErrTransceiverOffline {
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

	// Beacon invalid bounds & nil receiver.
	var nilB *rung3_local.BeaconSwitch
	common.AssertAuditFail(t, nilB.AuditState(), data.AuditSeverityCritical)
	if _, err := rung3_local.NewBeaconSwitch("", 6, rung3_local.BeaconModeSimulated, nil); err != rung3_local.ErrInvalidBeaconSSID {
		t.Fatalf("[SEC 2] empty SSID got %v", err)
	}
	if _, err := rung3_local.NewBeaconSwitch("OK", 0, rung3_local.BeaconModeSimulated, nil); err != rung3_local.ErrInvalidBeaconChannel {
		t.Fatalf("[SEC 2] channel 0 got %v", err)
	}
	if rung3_local.BeaconMode(9).String() != "UNKNOWN_MODE" {
		t.Fatal("[SEC 2] OOB BeaconMode did not return UNKNOWN_MODE")
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: [ SECTION 3: INVARIANTS, SATURATION & LIMIT GUARDS ]
// ------------------------------------------------------------------------------

func TestRung3Unit_Sec3_512ByteUDPCeilingAnd32ByteSSIDLimit(t *testing.T) {
	u, _ := rung3_local.NewUDPTransceiver(common.NodeAlpha, rung3_local.DefaultLocalUDPAddr)
	defer func() { _ = u.Close() }()

	var over [data.FloorMaxFrameSize + 1]byte
	if err := u.Send(common.NodeBeta, over[:]); err != transport.ErrWireFrameOversize {
		t.Fatalf("[SEC 3] 513B UDP datagram bypassed ceiling: %v", err)
	}
	if _, err := rung3_local.NewBeaconSwitch(strings.Repeat("A", 32), 1, rung3_local.BeaconModeSimulated, nil); err != nil {
		t.Fatalf("[SEC 3] exact 32B SSID rejected: %v", err)
	}
	if _, err := rung3_local.NewBeaconSwitch(strings.Repeat("A", 33), 1, rung3_local.BeaconModeSimulated, nil); err != rung3_local.ErrInvalidBeaconSSID {
		t.Fatalf("[SEC 3] 33B SSID bypassed IEEE limit: %v", err)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.4: [ SECTION 4: EMERGENCE & COMPOSITE WORKFLOWS ]
// ------------------------------------------------------------------------------

func TestRung3Unit_Sec4_DiagnosticDumpCardsRenderCleanly(t *testing.T) {
	u, _ := rung3_local.NewUDPTransceiver(common.NodeAlpha, "")
	defer func() { _ = u.Close() }()
	if !strings.Contains(u.ExecuteUDPDumpCard(), "RUNG 3 LOCAL UDP") {
		t.Fatal("[SEC 4] UDP dump card missing title")
	}
	sw, _ := rung3_local.NewBeaconSwitch("Room-Test", 6, rung3_local.BeaconModeSimulated, nil)
	_ = sw.TurnOn()
	if !strings.Contains(sw.ExecuteBeaconDumpCard(), "ON (BROADCASTING)") {
		t.Fatal("[SEC 4] active beacon dump card missing ON status")
	}
	_ = sw.Close()
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Preflight Diagnostic Gates & Physical Source Anchor Verification )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.rung3_local
// Mission:       Run Rung 3 preflight gates and verify physical anchors.
// ==============================================================================

func TestRung3Unit_Block4_PreflightAndAnchor(t *testing.T) {
	if rung3LocalUnitSuiteName == "" {
		t.Fatal("[BLOCK 4] suite identifier must not be empty")
	}
	common.AssertAuditPass(t, rung3_local.PreflightLocalUDP(), "")
	common.AssertAuditPass(t, rung3_local.PreflightBeaconSwitch(), "")
	if !rung3_local.AuditLocalUDP() || !rung3_local.AuditBeaconSwitch() {
		t.Fatal("[BLOCK 4] Rung 3 boolean audit wrappers failed")
	}
	common.AssertRepoAnchors(t,
		"transport/rung3_local/udp.go",
		"transport/rung3_local/beacon.go",
	)
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Test Suite Epilogue & Clean Readiness Handshake )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.rung3_local
// Mission:       Confirm Rung 3 readiness handshake post-suite.
// ==============================================================================

func TestRung3Unit_Block5_EpilogueReadiness(t *testing.T) {
	if !rung3_local.IsLocalUDPReady() || !rung3_local.IsBeaconSwitchReady() ||
		!common.IsHarnessReady() {
		t.Fatal("[BLOCK 5: EPILOGUE] Rung 3 unit readiness handshake failed")
	}
}

// End of tests/unit/transport/rung3_local_test.go
