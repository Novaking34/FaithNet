// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/functional/data_functional_test.go
// @id            test.functional.data
// @layer         Layer 4: Tier-1 Functional Verification Suite
// @archetype     verification_suite | functional_surface
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Layer 0 Data Vocabulary Functional Surface Suite
// @role          Tier-1 Functional Diagnostic Sensor (4-Section Micro Protocol)
// @maturity      Rev 5 Baseline | 4x4 Verification Matrix
// @chassis       zero-heap | Verifies pure Layer 0 scalar & inline value surface
// @scale         O(1) deterministic coarse API surface sweep
// @seams         in: data, engine, tests/common | probes: AuditDataVocabulary
// @invariants    1. Macro Tier 1 (Functional): exercises every public Layer 0
//                   vocabulary type before deeper Tier 2 unit isolation.
//                2. Micro 4-Section Protocol: Always Pass, Never Fail,
//                   Invariants & Limits, Emergence.
//                3. FrameHeaderSize == 32 && FloorMaxFrameSize == 512.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Tier-1 Functional surface suite exercising FaithNet's Layer 0 vocabulary
//   (`data/types.go`, `data/opcodes.go`, `data/frame.go`, `data/profile.go`,
//   `data/audit.go`) across all 4 ontological micro sections.
//
//   ### For Architects, Operators & Non-Programmers:
//   Confirms at a surface level that FaithNet's shared dictionary of device
//   IDs, survival rungs, truth verdicts, 512-byte envelopes, and diagnostic
//   reports all behave cleanly and safely reject invalid values.
//
//   ### For Systems Programmers & Engine Architects:
//   Mirrors `tests/functional/test_log_functional.c` and `tools/tests/functional/`
//   with full 5-Block symmetry (Preflight & Anchor gates in Block 4, Epilogue
//   readiness handshake in Block 5).
// ==============================================================================

package functional_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Identifier, Verification Imports & Pre-Condition Fixtures )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-1 Functional Verification Suite
// Module:        test.functional.data
// Mission:       Bind Layer 0 data types and shared harness for surface audits.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"bytes"
	"strings"
	"testing"
)

// --- Tier 2: External Library Dependencies (None — 100% Go Stdlib Invariant) ---

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
	dataFunctionalSuiteName = "Tier 1: Layer 0 Data Vocabulary Functional Surface"
)

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-1 Functional Verification Suite
// Module:        test.functional.data
// Mission:       Verify Always-Pass nominal surface, Never-Fail fault inversion,
//                512-byte Floppy Sector limits, and composite data emergence.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: [ SECTION 1: ALWAYS PASS (NOMINAL GOLDEN SURFACE COGS) ]
// ------------------------------------------------------------------------------

func TestDataFunctional_Sec1_NominalIdentityRungAndTruthSurface(t *testing.T) {
	if !common.NodeAlpha.IsValid() || common.NodeAlpha.IsBroadcast() {
		t.Fatalf("[SEC 1] NodeAlpha must be valid non-broadcast")
	}
	if !data.NodeIDBroadcast.IsBroadcast() {
		t.Fatalf("[SEC 1] NodeIDBroadcast predicate failed")
	}
	if got := data.Rung1Floor.String(); got != "RUNG_1_FLOOR" {
		t.Fatalf("[SEC 1] expected RUNG_1_FLOOR, got %q", got)
	}
	if got := data.QuadPositive.String(); got != "+1:ALIGNED" {
		t.Fatalf("[SEC 1] expected +1:ALIGNED, got %q", got)
	}
	if got := data.OpPing.String(); got != "PING" || !data.OpPing.IsControl() {
		t.Fatalf("[SEC 1] OpPing surface failed: name=%q ctrl=%v", got, data.OpPing.IsControl())
	}
}

func TestDataFunctional_Sec1_NominalProfileAndAuditSurface(t *testing.T) {
	manifest := data.DefaultNodeManifest(common.NodeAlpha, "room-anchor")
	if !manifest.IsValid() || manifest.AllowGlobalEgress {
		t.Fatalf("[SEC 1] DefaultNodeManifest must be valid and offline-first: %+v", manifest)
	}
	pass := data.AuditPass("cog.data.test", "nominal_surface", "verified")
	if !pass.IsPass() || !strings.Contains(pass.String(), "[PASS]") {
		t.Fatalf("[SEC 1] AuditPass surface mismatch: %s", pass.String())
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: [ SECTION 2: NEVER FAIL (FAULT INVERSION & DEGRADATION) ]
// ------------------------------------------------------------------------------

func TestDataFunctional_Sec2_RejectsInvalidScalarsAndOversizePayload(t *testing.T) {
	if data.NodeIDNull.IsValid() {
		t.Fatal("[SEC 2] NodeIDNull must be invalid")
	}
	if data.RungNone.IsValid() || data.RungLevel(99).IsValid() {
		t.Fatal("[SEC 2] out-of-range RungLevel must be invalid")
	}
	if data.QuadState(0x77).IsValid() || data.Opcode(0xFF).IsValid() {
		t.Fatal("[SEC 2] OOB QuadState/Opcode must be rejected")
	}
	if data.AuditSeverity(9).IsValid() || data.AuditSeverity(9).String() != "UNKNOWN" {
		t.Fatal("[SEC 2] OOB AuditSeverity must degrade to UNKNOWN")
	}

	var f data.Frame
	if f.SetPayload(make([]byte, data.FloorMaxPayload+1)) {
		t.Fatal("[SEC 2] SetPayload must reject >480B payloads cleanly")
	}
	if f.PayloadLen != 0 {
		t.Fatalf("[SEC 2] rejected SetPayload mutated PayloadLen=%d", f.PayloadLen)
	}

	badManifest := data.DefaultNodeManifest(common.NodeAlpha, "unsafe-egress")
	badManifest.PreferredMaxRung = data.Rung4Global
	badManifest.AllowGlobalEgress = false
	if badManifest.IsValid() {
		t.Fatal("[SEC 2] manifest enabling Rung4 without AllowGlobalEgress must be rejected")
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: [ SECTION 3: INVARIANTS, SATURATION & LIMIT GUARDS ]
// ------------------------------------------------------------------------------

func TestDataFunctional_Sec3_FloppySectorAndSaturationInvariants(t *testing.T) {
	if data.FrameHeaderSize != 32 || data.FloorMaxPayload != 480 ||
		data.FloorMaxFrameSize != 512 {
		t.Fatalf("[SEC 3] 512B Floppy Sector invariant broken: %d+%d=%d",
			data.FrameHeaderSize, data.FloorMaxPayload, data.FloorMaxFrameSize)
	}
	if data.Rung1Floor.StepDown() != data.Rung1Floor {
		t.Fatal("[SEC 3] Rung1Floor.StepDown() must saturate at floor")
	}
	if data.Rung4Global.StepUp() != data.Rung4Global {
		t.Fatal("[SEC 3] Rung4Global.StepUp() must saturate at ceiling")
	}
	var f data.Frame
	if !f.SetPayload(bytes.Repeat([]byte{0xA5}, data.FloorMaxPayload)) || f.WireSize() != 512 {
		t.Fatalf("[SEC 3] exact 480B payload failed 512B wire invariant: size=%d", f.WireSize())
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.4: [ SECTION 4: EMERGENCE & COMPOSITE WORKFLOWS ]
// ------------------------------------------------------------------------------

func TestDataFunctional_Sec4_ManifestProfileFrameCompositeWorkflow(t *testing.T) {
	body := common.OfflineRoomProfile()
	manifest := data.DefaultNodeManifest(common.NodeAlpha, "living-room")
	rung := body.HighestViableRung()
	if rung > manifest.PreferredMaxRung {
		rung = manifest.PreferredMaxRung
	}

	frame := common.MakeSovereignFrame(
		manifest.ID, common.NodeBeta, rung, data.OpBeacon, 1,
		[]byte(manifest.BeaconSSID),
	)
	if !frame.IsHeaderBoundsValid() || frame.Rung != data.Rung3Local {
		t.Fatalf("[SEC 4] composite manifest->body->frame beacon failed: %+v", frame)
	}
	frame.Reset()
	if frame.IsHeaderBoundsValid() || frame.PayloadLen != 0 {
		t.Fatal("[SEC 4] Frame.Reset() did not restore clean zero state")
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Preflight Diagnostic Gates & Physical Source Anchor Verification )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-1 Functional Verification Suite
// Module:        test.functional.data
// Mission:       Run Layer 0 vocabulary pre-flight audit and verify anchors.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Preflight Diagnostic Gates
// ------------------------------------------------------------------------------

func TestDataFunctional_Block4_PreflightVocabularyAndHarness(t *testing.T) {
	common.AssertAuditPass(t, common.SelfAudit(), common.ExecuteHarnessDumpCard())
	common.AssertAuditPass(t, engine.AuditDataVocabulary(), engine.ExecuteCodecDumpCard())
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Suite Identifier & Physical Source Anchor Gate
// ------------------------------------------------------------------------------

func TestDataFunctional_Block4_SourceAnchors(t *testing.T) {
	if dataFunctionalSuiteName == "" {
		t.Fatal("[BLOCK 4] suite identifier must not be empty")
	}
	common.AssertRepoAnchors(t,
		"data/types.go",
		"data/opcodes.go",
		"data/frame.go",
		"data/profile.go",
		"data/audit.go",
	)
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Test Suite Epilogue & Clean Readiness Handshake )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-1 Functional Verification Suite
// Module:        test.functional.data
// Mission:       Confirm clean readiness handshake after Layer 0 surface sweep.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Epilogue & Clean State Handshake
// ------------------------------------------------------------------------------

func TestDataFunctional_Block5_EpilogueReadiness(t *testing.T) {
	if !common.IsHarnessReady() || !engine.IsCodecReady() {
		t.Fatal("[BLOCK 5: EPILOGUE] Layer 0/1 readiness handshake failed")
	}
}

// End of tests/functional/data_functional_test.go
