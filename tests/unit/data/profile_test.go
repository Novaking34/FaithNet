// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/unit/data/profile_test.go
// @id            test.unit.data.profile
// @layer         Layer 4: Tier-2 Unit Verification Suite (Data Sub-Stratum)
// @archetype     verification_suite | unit_isolation
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Layer 0 HardwareProfile, Manifest & LinkHealth Suite
// @role          Tier-2 Unit Isolation Gate for cog.data.profile
// @maturity      Rev 5 Baseline | 4x4 Verification Matrix
// @chassis       zero-heap | Pure O(1) capability & loss-percentage predicates
// @scale         O(1) deterministic host-body & manifest policy sweep
// @seams         in: data, engine, tests/common | target: data/profile.go
// @invariants    1. DefaultNodeManifest enforces AllowGlobalEgress == false.
//                2. LinkHealth.LossPercent clamps safely in [0..100] with 0 div.
//                3. HardwareProfile.HighestViableRung degrades 4->3->2->1->None.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Tier-2 Unit isolation suite verifying `HardwareProfile`, `NodeManifest`,
//   `LinkHealth`, and `CascadeEvent` declared in `data/profile.go`.
//
//   ### For Architects, Operators & Non-Programmers:
//   Verifies FaithNet's capability medical chart picks the right survival rung
//   for any computer, never opens internet access without permission, and
//   calculates packet loss accurately without divide-by-zero crashes.
//
//   ### For Systems Programmers & Engine Architects:
//   100% branch coverage of `HighestViableRung()`, `IsValid()`, `LossPercent()`,
//   and `DefaultNodeManifest()`.
// ==============================================================================

package data_unit_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Identifier, Verification Imports & Tier-2 Unit Sensor Anchor )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.data.profile
// Mission:       Bind Layer 0 profile schemas and shared harness.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import "testing"

// --- Tier 2: External Library Dependencies (None) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---
import (
	"github.com/CreativeWorkzStudio/FaithNet/data"
	"github.com/CreativeWorkzStudio/FaithNet/engine"
	"github.com/CreativeWorkzStudio/FaithNet/tests/common"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Operational Suite Constants & Identifiers
// ------------------------------------------------------------------------------

const (
	profileUnitSuiteName = "Tier 2 Unit: cog.data.profile (Profile, Manifest, Health)"
)

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.data.profile
// Mission:       Verify hardware rung cascading, manifest safety, and LossPercent.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: [ SECTION 1: ALWAYS PASS (NOMINAL GOLDEN SURFACE COGS) ]
// ------------------------------------------------------------------------------

func TestProfileUnit_Sec1_HighestViableRungAllTiers(t *testing.T) {
	cases := []struct {
		name string
		prof data.HardwareProfile
		want data.RungLevel
	}{
		{"global_tcp", common.RichHostProfile(), data.Rung4Global},
		{"local_room", common.OfflineRoomProfile(), data.Rung3Local},
		{"wifi_only", data.HardwareProfile{HasWiFiHardware: true}, data.Rung3Local},
		{"host_ram", data.HardwareProfile{HasMemoryBus: true}, data.Rung2Host},
		{"host_ipc", data.HardwareProfile{HasLocalIPC: true}, data.Rung2Host},
		{"floor_serial", data.HardwareProfile{HasSerialPorts: true}, data.Rung1Floor},
		{"floor_audio", data.HardwareProfile{HasAudioDevice: true}, data.Rung1Floor},
		{"no_hardware", data.HardwareProfile{}, data.RungNone},
	}
	for _, tc := range cases {
		if got := tc.prof.HighestViableRung(); got != tc.want {
			t.Fatalf("[SEC 1] %s: got %s want %s", tc.name, got, tc.want)
		}
	}
}

func TestProfileUnit_Sec1_DefaultManifestAndNominalLossPercent(t *testing.T) {
	m := data.DefaultNodeManifest(common.NodeAlpha, "anchor")
	if !m.IsValid() || m.AllowGlobalEgress ||
		m.PreferredMaxRung != data.Rung3Local ||
		m.MinimumFloorRung != data.Rung1Floor ||
		m.WellListenAddr != data.DefaultWellListenAddr ||
		m.BeaconSSID != data.DefaultBeaconSSID {
		t.Fatalf("[SEC 1] unexpected DefaultNodeManifest: %+v", m)
	}

	h := data.LinkHealth{PingsSent: 10, PongsRecv: 7}
	if pct := h.LossPercent(); pct != 30 {
		t.Fatalf("[SEC 1] expected 30%% loss, got %d%%", pct)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: [ SECTION 2: NEVER FAIL (FAULT INVERSION & DEGRADATION) ]
// ------------------------------------------------------------------------------

func TestProfileUnit_Sec2_RejectsUnsafeOrMalformedManifests(t *testing.T) {
	// 1. Null or broadcast NodeID rejected.
	if data.DefaultNodeManifest(data.NodeIDNull, "bad").IsValid() {
		t.Fatal("[SEC 2] NodeIDNull manifest must be invalid")
	}
	if data.DefaultNodeManifest(data.NodeIDBroadcast, "bad").IsValid() {
		t.Fatal("[SEC 2] NodeIDBroadcast manifest must be invalid")
	}

	// 2. MinimumFloorRung > PreferredMaxRung rejected.
	inverted := data.DefaultNodeManifest(common.NodeAlpha, "inv")
	inverted.MinimumFloorRung = data.Rung3Local
	inverted.PreferredMaxRung = data.Rung1Floor
	if inverted.IsValid() {
		t.Fatal("[SEC 2] inverted min/max rung manifest must be rejected")
	}

	// 3. Invalid rung bounds rejected.
	badFloor := data.DefaultNodeManifest(common.NodeAlpha, "bad-floor")
	badFloor.MinimumFloorRung = data.RungNone
	if badFloor.IsValid() {
		t.Fatal("[SEC 2] RungNone minimum floor must be rejected")
	}

	// 4. Rung4Global without explicit AllowGlobalEgress rejected.
	sneaky := data.DefaultNodeManifest(common.NodeAlpha, "sneaky")
	sneaky.PreferredMaxRung = data.Rung4Global
	sneaky.AllowGlobalEgress = false
	if sneaky.IsValid() {
		t.Fatal("[SEC 2] Rung4Global without opt-in must be rejected")
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: [ SECTION 3: INVARIANTS, SATURATION & LIMIT GUARDS ]
// ------------------------------------------------------------------------------

func TestProfileUnit_Sec3_LossPercentBoundaryGuards(t *testing.T) {
	if (data.LinkHealth{PingsSent: 0, PongsRecv: 0}).LossPercent() != 0 {
		t.Fatal("[SEC 3] zero-sent LossPercent must be 0 (no div-by-zero)")
	}
	if (data.LinkHealth{PingsSent: 5, PongsRecv: 9}).LossPercent() != 0 {
		t.Fatal("[SEC 3] PongsRecv > PingsSent must clamp LossPercent to 0")
	}
	if (data.LinkHealth{PingsSent: 5, PongsRecv: 0}).LossPercent() != 100 {
		t.Fatal("[SEC 3] 0 pongs of 5 sent must yield 100% loss")
	}
	if (data.LinkHealth{PingsSent: 5, PongsRecv: 5}).LossPercent() != 0 {
		t.Fatal("[SEC 3] 5/5 pongs must yield 0% loss")
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.4: [ SECTION 4: EMERGENCE & COMPOSITE WORKFLOWS ]
// ------------------------------------------------------------------------------

func TestProfileUnit_Sec4_ExplicitGlobalOptInAndCascadeRecord(t *testing.T) {
	m := data.DefaultNodeManifest(common.NodeAlpha, "gateway")
	m.AllowGlobalEgress = true
	m.PreferredMaxRung = data.Rung4Global
	if !m.IsValid() {
		t.Fatal("[SEC 4] explicit AllowGlobalEgress=true manifest should be valid")
	}

	evt := data.CascadeEvent{
		TimestampNano: 42_000,
		FromRung:      data.Rung4Global,
		ToRung:        data.Rung3Local,
		Direction:     data.OpCascadeDown,
		Reason:        "upstream fiber severed",
	}
	if !evt.Direction.IsControl() || evt.FromRung.StepDown() != evt.ToRung {
		t.Fatalf("[SEC 4] malformed CascadeEvent record: %+v", evt)
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Preflight Diagnostic Gates & Physical Source Anchor Verification )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.data.profile
// Mission:       Run vocabulary preflight gate and verify physical anchor.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Block 4 Pre-Flight Contract Gates
// ------------------------------------------------------------------------------

func TestProfileUnit_Block4_PreflightGates(t *testing.T) {
	if profileUnitSuiteName == "" {
		t.Fatal("[BLOCK 4] suite identifier must not be empty")
	}
	common.AssertAuditPass(t, engine.AuditDataVocabulary(), engine.ExecuteCodecDumpCard())
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Physical Source Anchor Verification
// ------------------------------------------------------------------------------

func TestProfileUnit_Block4_SourceAnchors(t *testing.T) {
	common.AssertRepoAnchors(t, "data/profile.go")
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Test Suite Epilogue & Clean Readiness Handshake )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.data.profile
// Mission:       Confirm readiness handshake post-suite.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Epilogue Readiness Verification
// ------------------------------------------------------------------------------

func TestProfileUnit_Block5_EpilogueReadiness(t *testing.T) {
	if !common.IsHarnessReady() || !engine.IsCodecReady() {
		t.Fatal("[BLOCK 5: EPILOGUE] profile unit readiness handshake failed")
	}
}

// End of tests/unit/data/profile_test.go
