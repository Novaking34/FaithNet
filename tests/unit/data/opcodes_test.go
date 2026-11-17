// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/unit/data/opcodes_test.go
// @id            test.unit.data.opcodes
// @layer         Layer 4: Tier-2 Unit Verification Suite (Data Sub-Stratum)
// @archetype     verification_suite | unit_isolation
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Layer 0 Opcode & FrameFlags Unit Suite
// @role          Tier-2 Unit Isolation Gate for cog.data.opcodes
// @maturity      Rev 5 Baseline | 4x4 Verification Matrix
// @chassis       zero-heap | Exhaustive 1-byte enum & bitmask verification
// @scale         O(1) deterministic opcode & bitmask sweep
// @seams         in: data, engine, tests/common | target: data/opcodes.go
// @invariants    1. OpPing==0x01 and OpPong==0x02 permanently pinned.
//                2. Every FrameFlags bit is pairwise orthogonal.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Tier-2 Unit isolation suite verifying every `Opcode` verb and `FrameFlags`
//   bitmask predicate declared in `data/opcodes.go` across Sections 1-4.
//
//   ### For Architects, Operators & Non-Programmers:
//   Confirms every protocol action word (PING, PONG, BEACON, CPI_SYNC, etc.)
//   and envelope stamp (Covenant Verified, External Origin, etc.) works without
//   bit collisions or misclassifications.
//
//   ### For Systems Programmers & Engine Architects:
//   100% branch coverage of `IsValid()`, `IsControl()`, `String()`, and `Has()`.
// ==============================================================================

package data_unit_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Identifier, Verification Imports & Tier-2 Unit Sensor Anchor )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.data.opcodes
// Mission:       Bind Layer 0 opcodes and shared harness for unit isolation.
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
	opcodesUnitSuiteName = "Tier 2 Unit: cog.data.opcodes (Opcode, FrameFlags)"
)

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.data.opcodes
// Mission:       Exhaustively verify every Opcode and FrameFlags combination.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: [ SECTION 1: ALWAYS PASS (NOMINAL GOLDEN SURFACE COGS) ]
// ------------------------------------------------------------------------------

func TestOpcodesUnit_Sec1_AllValidOpcodesAndLabels(t *testing.T) {
	want := map[data.Opcode]struct {
		label string
		ctrl  bool
	}{
		data.OpPing:        {"PING", true},
		data.OpPong:        {"PONG", true},
		data.OpBeacon:      {"BEACON", false},
		data.OpCPISync:     {"CPI_SYNC", false},
		data.OpLegacyWrap:  {"LEGACY_WRAP", false},
		data.OpCascadeDown: {"EVT_CASCADE_DOWN", true},
		data.OpCascadeUp:   {"EVT_CASCADE_UP", true},
		data.OpData:        {"DATA", false},
	}
	for op, exp := range want {
		if !op.IsValid() {
			t.Fatalf("[SEC 1] opcode 0x%02X reported invalid", uint8(op))
		}
		if op.String() != exp.label {
			t.Fatalf("[SEC 1] opcode 0x%02X label=%q want %q", uint8(op), op.String(), exp.label)
		}
		if op.IsControl() != exp.ctrl {
			t.Fatalf("[SEC 1] opcode %s IsControl=%v want %v", exp.label, op.IsControl(), exp.ctrl)
		}
	}
}

func TestOpcodesUnit_Sec1_FrameFlagsSingleAndCombinedMasks(t *testing.T) {
	all := data.FlagCovenantVerified | data.FlagExternalOrigin |
		data.FlagLegacyTranslated | data.FlagFloorCompact | data.FlagRequiresAck

	if !all.Has(data.FlagCovenantVerified) ||
		!all.Has(data.FlagExternalOrigin) ||
		!all.Has(data.FlagLegacyTranslated) ||
		!all.Has(data.FlagFloorCompact) ||
		!all.Has(data.FlagRequiresAck) ||
		!all.Has(data.FlagFloorCompact|data.FlagRequiresAck) {
		t.Fatal("[SEC 1] combined FrameFlags.Has() failed")
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: [ SECTION 2: NEVER FAIL (FAULT INVERSION & DEGRADATION) ]
// ------------------------------------------------------------------------------

func TestOpcodesUnit_Sec2_InvalidOpcodesAndUnsetFlags(t *testing.T) {
	for _, bad := range []data.Opcode{data.OpInvalid, 0x09, 0x7F, 0xFF} {
		if bad.IsValid() || bad.IsControl() {
			t.Fatalf("[SEC 2] invalid opcode 0x%02X accepted", uint8(bad))
		}
		if bad.String() != "INVALID" {
			t.Fatalf("[SEC 2] invalid opcode 0x%02X string=%q want INVALID", uint8(bad), bad.String())
		}
	}
	if data.FlagNone.Has(data.FlagCovenantVerified) {
		t.Fatal("[SEC 2] FlagNone must not match non-zero mask")
	}
	partial := data.FlagFloorCompact
	if partial.Has(data.FlagFloorCompact | data.FlagRequiresAck) {
		t.Fatal("[SEC 2] partial flags must fail multi-bit Has() mask")
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: [ SECTION 3: INVARIANTS, SATURATION & LIMIT GUARDS ]
// ------------------------------------------------------------------------------

func TestOpcodesUnit_Sec3_WirePinningAndBitwiseOrthogonality(t *testing.T) {
	if uint8(data.OpPing) != 0x01 || uint8(data.OpPong) != 0x02 {
		t.Fatal("[SEC 3] OpPing (0x01) / OpPong (0x02) wire pinning violated")
	}
	if data.OpcodeMinValid != data.OpPing || data.OpcodeMaxValid != data.OpData {
		t.Fatal("[SEC 3] OpcodeMinValid/MaxValid bounds drifted")
	}

	flags := []data.FrameFlags{
		data.FlagCovenantVerified,
		data.FlagExternalOrigin,
		data.FlagLegacyTranslated,
		data.FlagFloorCompact,
		data.FlagRequiresAck,
	}
	for i := 0; i < len(flags); i++ {
		for j := i + 1; j < len(flags); j++ {
			if (flags[i] & flags[j]) != 0 {
				t.Fatalf("[SEC 3] FrameFlags collision between index %d and %d", i, j)
			}
		}
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.4: [ SECTION 4: EMERGENCE & COMPOSITE WORKFLOWS ]
// ------------------------------------------------------------------------------

func TestOpcodesUnit_Sec4_ExhaustiveByteSpacePartition(t *testing.T) {
	validCount := 0
	for b := 0; b <= 255; b++ {
		op := data.Opcode(b)
		if op.IsValid() {
			validCount++
			if op.String() == "INVALID" {
				t.Fatalf("[SEC 4] valid opcode 0x%02X returned INVALID string", b)
			}
		} else if op.String() != "INVALID" {
			t.Fatalf("[SEC 4] invalid opcode 0x%02X returned %q", b, op.String())
		}
	}
	if validCount != 8 {
		t.Fatalf("[SEC 4] expected exactly 8 valid opcodes, counted %d", validCount)
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Preflight Diagnostic Gates & Physical Source Anchor Verification )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.data.opcodes
// Mission:       Run vocabulary preflight gate and verify physical anchor.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Block 4 Pre-Flight Contract Gates
// ------------------------------------------------------------------------------

func TestOpcodesUnit_Block4_PreflightGates(t *testing.T) {
	if opcodesUnitSuiteName == "" {
		t.Fatal("[BLOCK 4] suite identifier must not be empty")
	}
	common.AssertAuditPass(t, engine.AuditDataVocabulary(), engine.ExecuteCodecDumpCard())
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Physical Source Anchor Verification
// ------------------------------------------------------------------------------

func TestOpcodesUnit_Block4_SourceAnchors(t *testing.T) {
	common.AssertRepoAnchors(t, "data/opcodes.go")
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Test Suite Epilogue & Clean Readiness Handshake )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.data.opcodes
// Mission:       Confirm readiness handshake post-suite.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Epilogue Readiness Verification
// ------------------------------------------------------------------------------

func TestOpcodesUnit_Block5_EpilogueReadiness(t *testing.T) {
	if !common.IsHarnessReady() || !engine.IsCodecReady() {
		t.Fatal("[BLOCK 5: EPILOGUE] opcodes unit readiness handshake failed")
	}
}

// End of tests/unit/data/opcodes_test.go
