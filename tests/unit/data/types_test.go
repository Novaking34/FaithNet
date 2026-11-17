// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/unit/data/types_test.go
// @id            test.unit.data.types
// @layer         Layer 4: Tier-2 Unit Verification Suite (Data Sub-Stratum)
// @archetype     verification_suite | unit_isolation
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Layer 0 Scalar Identity, Rung & T4 Unit Suite
// @role          Tier-2 Unit Isolation Gate for cog.data.types
// @maturity      Rev 5 Baseline | 4x4 Verification Matrix
// @chassis       zero-heap | Pure scalar value predicates & saturation bounds
// @scale         O(1) deterministic exhaustive enum & boundary sweep
// @seams         in: data, engine, tests/common | target: data/types.go
// @invariants    1. Exhaustively tests every RungLevel (0..5), QuadState (0..4),
//                   and NodeID sentinel (Null, Broadcast, Normal).
//                2. StepDown saturates at Rung1Floor; StepUp at Rung4Global.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Tier-2 Unit isolation suite strictly verifying `data/types.go` across all
//   4 horizontal micro sections, Block 4 preflight gates, and Block 5 epilogue.
//
//   ### For Architects, Operators & Non-Programmers:
//   Confirms every device address, survival ladder rung, and four-state truth
//   label behaves reliably at both normal and extreme boundary inputs.
//
//   ### For Systems Programmers & Engine Architects:
//   100% branch coverage of `NodeID`, `RungLevel`, and `QuadState` predicates.
// ==============================================================================

package data_unit_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Identifier, Verification Imports & Tier-2 Unit Sensor Anchor )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.data.types
// Mission:       Bind `data` types and shared harness for isolated unit tests.
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
	typesUnitSuiteName = "Tier 2 Unit: cog.data.types (NodeID, RungLevel, T4)"
)

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.data.types
// Mission:       Exhaustively verify NodeID, RungLevel, and QuadState cogs.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: [ SECTION 1: ALWAYS PASS (NOMINAL GOLDEN SURFACE COGS) ]
// ------------------------------------------------------------------------------

func TestTypesUnit_Sec1_NodeIDFormattingAndPredicates(t *testing.T) {
	if !common.NodeAlpha.IsValid() || common.NodeAlpha.IsBroadcast() {
		t.Fatal("[SEC 1] NodeAlpha validity/broadcast mismatch")
	}
	if want := "0xFA17000000000001"; common.NodeAlpha.String() != want {
		t.Fatalf("[SEC 1] NodeID.String() want %q got %q", want, common.NodeAlpha.String())
	}
	if !data.NodeIDBroadcast.IsValid() || !data.NodeIDBroadcast.IsBroadcast() {
		t.Fatal("[SEC 1] NodeIDBroadcast must be valid and broadcast")
	}
	if want := "0xFFFFFFFFFFFFFFFF"; data.NodeIDBroadcast.String() != want {
		t.Fatalf("[SEC 1] NodeIDBroadcast.String() want %q got %q", want, data.NodeIDBroadcast.String())
	}
}

func TestTypesUnit_Sec1_AllCanonicalRungsAndQuadStates(t *testing.T) {
	rungNames := map[data.RungLevel]string{
		data.Rung1Floor:  "RUNG_1_FLOOR",
		data.Rung2Host:   "RUNG_2_HOST",
		data.Rung3Local:  "RUNG_3_LOCAL",
		data.Rung4Global: "RUNG_4_GLOBAL",
	}
	for r, want := range rungNames {
		if !r.IsValid() || r.String() != want {
			t.Fatalf("[SEC 1] rung %d: valid=%v str=%q want=%q", r, r.IsValid(), r.String(), want)
		}
	}

	quadNames := map[data.QuadState]string{
		data.QuadNull:     "NULL:HALT",
		data.QuadNegative: "-1:REJECT",
		data.QuadAwait:    "0:AWAIT",
		data.QuadPositive: "+1:ALIGNED",
	}
	for q, want := range quadNames {
		if !q.IsValid() || q.String() != want {
			t.Fatalf("[SEC 1] quad %d: valid=%v str=%q want=%q", q, q.IsValid(), q.String(), want)
		}
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: [ SECTION 2: NEVER FAIL (FAULT INVERSION & DEGRADATION) ]
// ------------------------------------------------------------------------------

func TestTypesUnit_Sec2_OutOfBoundsRungQuadAndNullNodeID(t *testing.T) {
	if data.NodeIDNull.IsValid() || data.NodeIDNull.IsBroadcast() {
		t.Fatal("[SEC 2] NodeIDNull must be invalid and non-broadcast")
	}
	for _, bad := range []data.RungLevel{data.RungNone, 5, 15, 255} {
		if bad.IsValid() {
			t.Fatalf("[SEC 2] OOB rung %d reported valid", bad)
		}
		if bad.String() != "RUNG_NONE" {
			t.Fatalf("[SEC 2] OOB rung %d string=%q want RUNG_NONE", bad, bad.String())
		}
	}
	for _, bad := range []data.QuadState{0x04, 0x10, 0xFF} {
		if bad.IsValid() {
			t.Fatalf("[SEC 2] OOB QuadState %d reported valid", bad)
		}
		if bad.String() != "NULL:HALT" {
			t.Fatalf("[SEC 2] OOB QuadState %d string=%q want NULL:HALT", bad, bad.String())
		}
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: [ SECTION 3: INVARIANTS, SATURATION & LIMIT GUARDS ]
// ------------------------------------------------------------------------------

func TestTypesUnit_Sec3_LadderStepUpAndStepDownSaturation(t *testing.T) {
	// Nominal descending steps + floor saturation.
	if data.Rung4Global.StepDown() != data.Rung3Local ||
		data.Rung3Local.StepDown() != data.Rung2Host ||
		data.Rung2Host.StepDown() != data.Rung1Floor ||
		data.Rung1Floor.StepDown() != data.Rung1Floor ||
		data.RungNone.StepDown() != data.Rung1Floor ||
		data.RungLevel(99).StepDown() != data.Rung4Global {
		t.Fatal("[SEC 3] RungLevel.StepDown() saturation boundary failed")
	}

	// Nominal ascending steps + ceiling saturation.
	if data.RungNone.StepUp() != data.Rung1Floor ||
		data.Rung1Floor.StepUp() != data.Rung2Host ||
		data.Rung2Host.StepUp() != data.Rung3Local ||
		data.Rung3Local.StepUp() != data.Rung4Global ||
		data.Rung4Global.StepUp() != data.Rung4Global ||
		data.RungLevel(99).StepUp() != data.Rung4Global {
		t.Fatal("[SEC 3] RungLevel.StepUp() saturation boundary failed")
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.4: [ SECTION 4: EMERGENCE & COMPOSITE WORKFLOWS ]
// ------------------------------------------------------------------------------

func TestTypesUnit_Sec4_FullBidirectionalLadderSweep(t *testing.T) {
	// Climb Rung 1 -> 4, then cascade 4 -> 1 cleanly.
	cur := data.Rung1Floor
	for i := 0; i < 6; i++ {
		cur = cur.StepUp()
	}
	if cur != data.Rung4Global {
		t.Fatalf("[SEC 4] expected climb ceiling Rung4Global, got %s", cur)
	}
	for i := 0; i < 6; i++ {
		cur = cur.StepDown()
	}
	if cur != data.Rung1Floor {
		t.Fatalf("[SEC 4] expected cascade floor Rung1Floor, got %s", cur)
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Preflight Diagnostic Gates & Physical Source Anchor Verification )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.data.types
// Mission:       Run vocabulary preflight gate and verify physical source anchor.
// ==============================================================================

func TestTypesUnit_Block4_PreflightAndAnchor(t *testing.T) {
	if typesUnitSuiteName == "" {
		t.Fatal("[BLOCK 4] suite identifier must not be empty")
	}
	common.AssertAuditPass(t, common.SelfAudit(), common.ExecuteHarnessDumpCard())
	common.AssertAuditPass(t, engine.AuditDataVocabulary(), engine.ExecuteCodecDumpCard())
	common.AssertRepoAnchors(t, "data/types.go")
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Test Suite Epilogue & Clean Readiness Handshake )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.data.types
// Mission:       Confirm harness and codec readiness after scalar unit suite.
// ==============================================================================

func TestTypesUnit_Block5_EpilogueReadiness(t *testing.T) {
	if !common.IsHarnessReady() || !engine.IsCodecReady() {
		t.Fatal("[BLOCK 5: EPILOGUE] types unit readiness handshake failed")
	}
}

// End of tests/unit/data/types_test.go
