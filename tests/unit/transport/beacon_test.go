// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/unit/transport/beacon_test.go
// @id            test.unit.transport.beacon
// @layer         Layer 4: Tier-2 Unit Verification Suite (Transport Sub-Stratum)
// @archetype     verification_suite | unit_isolation
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Rung 3 Room Wi-Fi AP Beacon Switch Unit Suite
// @role          Tier-2 Unit Isolation Gate for cog.transport.rung3_beacon
// @maturity      Rev 5 Baseline | 4x4 Verification Matrix
// @chassis       zero-heap | Safe simulated Wi-Fi AP On/Off state machine
// @scale         O(1) toggle, IEEE 802.11 bounds & OpBeacon frame tests
// @seams         in: data, engine, rung3_local, common | target: beacon.go
// @invariants    1. BeaconModeSimulated NEVER touches host OS Wi-Fi state.
//                2. Enforces IEEE 802.11 SSID (1..32B) and channel (1..14).
//                3. TurnOn/TurnOff/Close are idempotent and thread-safe.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Dedicated 1:1 Tier-2 Unit isolation suite for `transport/rung3_local/beacon.go`.
//
//   ### For Architects, Operators & Non-Programmers:
//   Verifies FaithNet's local room beacon On/Off switch enforces legal Wi-Fi
//   names/channels and builds clean room announcement envelopes.
//
//   ### For Systems Programmers & Engine Architects:
//   100% branch coverage of `BeaconSwitch`, `BeaconMode`, `BuildBeaconFrame`.
// ==============================================================================

package transport_unit_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Identifier, Verification Imports & Tier-2 Unit Sensor Anchor )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.beacon
// Mission:       Bind Rung 3 BeaconSwitch and shared test harness.
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
	"github.com/CreativeWorkzStudio/FaithNet/engine"
	"github.com/CreativeWorkzStudio/FaithNet/tests/common"
	"github.com/CreativeWorkzStudio/FaithNet/transport/rung3_local"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Operational Suite Constants & Identifiers
// ------------------------------------------------------------------------------

const (
	beaconUnitSuiteName = "Tier 2 Unit: cog.transport.rung3_beacon (Room Beacon)"
)

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.beacon
// Mission:       Verify BeaconSwitch toggle lifecycle, IEEE limits & frames.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: [ SECTION 1: ALWAYS PASS (NOMINAL GOLDEN SURFACE COGS) ]
// ------------------------------------------------------------------------------

func TestBeaconUnit_Sec1_NominalToggleIdempotenceAndOpBeacon(t *testing.T) {
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
	_ = sw.TurnOn() // Idempotent TurnOn.
	st := sw.StatusSnapshot()
	if st.ToggleCount != 1 || st.StartedAtNano != 500 ||
		st.Mode.String() != "SIMULATED_SAFE" ||
		rung3_local.BeaconModeSystem.String() != "SYSTEM_HOST_AP" {
		t.Fatalf("[SEC 1] unexpected BeaconStatus: %+v", st)
	}

	bf := sw.BuildBeaconFrame(common.NodeAlpha, data.RungNone, 7)
	if bf.Opcode != data.OpBeacon || bf.Rung != data.Rung3Local || !bf.TargetID.IsBroadcast() ||
		!strings.Contains(string(bf.PayloadSlice()), "LivingRoom-Mesh:CH11") {
		t.Fatalf("[SEC 1] unexpected OpBeacon frame: %+v", bf)
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

func TestBeaconUnit_Sec2_RejectsInvalidSSIDChannelAndNilAudit(t *testing.T) {
	var nilB *rung3_local.BeaconSwitch
	common.AssertAuditFail(t, nilB.AuditState(), data.AuditSeverityCritical)

	if _, err := rung3_local.NewBeaconSwitch("", 6, rung3_local.BeaconModeSimulated, nil); err != rung3_local.ErrInvalidBeaconSSID {
		t.Fatalf("[SEC 2] empty SSID got %v", err)
	}
	if _, err := rung3_local.NewBeaconSwitch("OK", 0, rung3_local.BeaconModeSimulated, nil); err != rung3_local.ErrInvalidBeaconChannel {
		t.Fatalf("[SEC 2] channel 0 got %v", err)
	}
	if _, err := rung3_local.NewBeaconSwitch("OK", 15, rung3_local.BeaconModeSimulated, nil); err != rung3_local.ErrInvalidBeaconChannel {
		t.Fatalf("[SEC 2] channel 15 got %v", err)
	}
	if rung3_local.BeaconMode(99).String() != "UNKNOWN_MODE" {
		t.Fatal("[SEC 2] OOB BeaconMode did not return UNKNOWN_MODE")
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: [ SECTION 3: INVARIANTS, SATURATION & LIMIT GUARDS ]
// ------------------------------------------------------------------------------

func TestBeaconUnit_Sec3_IEEE80211BoundaryExact32AndChannels1To14(t *testing.T) {
	if _, err := rung3_local.NewBeaconSwitch(strings.Repeat("B", 32), 1, rung3_local.BeaconModeSimulated, nil); err != nil {
		t.Fatalf("[SEC 3] exact 32B SSID ch=1 rejected: %v", err)
	}
	if _, err := rung3_local.NewBeaconSwitch("X", 14, rung3_local.BeaconModeSimulated, nil); err != nil {
		t.Fatalf("[SEC 3] 1B SSID ch=14 rejected: %v", err)
	}
	if _, err := rung3_local.NewBeaconSwitch(strings.Repeat("B", 33), 6, rung3_local.BeaconModeSimulated, nil); err != rung3_local.ErrInvalidBeaconSSID {
		t.Fatalf("[SEC 3] 33B SSID bypassed IEEE ceiling: %v", err)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.4: [ SECTION 4: EMERGENCE & COMPOSITE WORKFLOWS ]
// ------------------------------------------------------------------------------

func TestBeaconUnit_Sec4_LiveAuditAndDumpCard(t *testing.T) {
	sw, _ := rung3_local.NewBeaconSwitch("Block7-Anchor", 6, rung3_local.BeaconModeSimulated, nil)
	_ = sw.TurnOn()
	card := sw.ExecuteBeaconDumpCard()
	common.AssertAuditPass(t, sw.AuditState(), card)
	if !strings.Contains(card, "ON (BROADCASTING)") || !strings.Contains(card, "Block7-Anchor") {
		t.Fatalf("[SEC 4] unexpected beacon card:\n%s", card)
	}
	_ = sw.Close()
	if !strings.Contains(sw.ExecuteBeaconDumpCard(), "OFF") {
		t.Fatal("[SEC 4] post-Close beacon card did not render OFF")
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Preflight Diagnostic Gates & Physical Source Anchor Verification )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.beacon
// Mission:       Run PreflightBeaconSwitch and verify physical source anchor.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Preflight Diagnostic Gates
// ------------------------------------------------------------------------------

func TestBeaconUnit_Block4_PreflightGate(t *testing.T) {
	common.AssertAuditPass(t, rung3_local.PreflightBeaconSwitch(), "")
	if !rung3_local.AuditBeaconSwitch() {
		t.Fatal("[BLOCK 4] AuditBeaconSwitch returned false")
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Suite Identifier & Physical Source Anchor Gate
// ------------------------------------------------------------------------------

func TestBeaconUnit_Block4_SourceAnchor(t *testing.T) {
	if beaconUnitSuiteName == "" {
		t.Fatal("[BLOCK 4] suite identifier must not be empty")
	}
	common.AssertRepoAnchors(t, "transport/rung3_local/beacon.go")
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Test Suite Epilogue & Clean Readiness Handshake )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.beacon
// Mission:       Confirm Rung 3 Beacon readiness handshake post-suite.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Epilogue & Clean State Handshake
// ------------------------------------------------------------------------------

func TestBeaconUnit_Block5_EpilogueReadiness(t *testing.T) {
	if !rung3_local.IsBeaconSwitchReady() || !common.IsHarnessReady() {
		t.Fatal("[BLOCK 5: EPILOGUE] beacon unit readiness handshake failed")
	}
}

// End of tests/unit/transport/beacon_test.go
