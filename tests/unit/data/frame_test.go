// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/unit/data/frame_test.go
// @id            test.unit.data.frame
// @layer         Layer 4: Tier-2 Unit Verification Suite (Data Sub-Stratum)
// @archetype     verification_suite | unit_isolation
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Layer 0 512-Byte Inline Frame Unit Suite
// @role          Tier-2 Unit Isolation Gate for cog.data.frame
// @maturity      Rev 5 Baseline | 4x4 Verification Matrix
// @chassis       zero-heap | Verifies 512B Floppy Sector layout & zero allocs
// @scale         O(1) fixed-size inline struct & offset layout tests
// @seams         in: data, engine, tests/common | target: data/frame.go
// @invariants    1. FrameHeaderSize(32) + FloorMaxPayload(480) == 512 bytes.
//                2. Header offsets partition [0..31] contiguously with zero gaps.
//                3. SetPayload, PayloadSlice, WireSize, Reset == 0 allocs/op.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Tier-2 Unit isolation suite strictly verifying `data/frame.go` wire layout
//   constants, inline payload bounds, saturation clamping, and zero-alloc reuse.
//
//   ### For Architects, Operators & Non-Programmers:
//   Proves that FaithNet's 512-byte postal envelope never overflows, never
//   crashes on malformed lengths, and wipes completely clean between uses.
//
//   ### For Systems Programmers & Engine Architects:
//   100% branch coverage of `SetPayload`, `PayloadSlice`, `WireSize`,
//   `IsHeaderBoundsValid`, and `Reset` plus contiguous header-offset proofs.
// ==============================================================================

package data_unit_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Identifier, Verification Imports & Tier-2 Unit Sensor Anchor )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.data.frame
// Mission:       Bind `data.Frame` and shared harness for isolated unit tests.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"bytes"
	"testing"
)

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
	frameUnitSuiteName = "Tier 2 Unit: cog.data.frame (512B Inline Frame)"
)

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.data.frame
// Mission:       Verify inline Frame geometry, setters, bounds, and Reset().
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: [ SECTION 1: ALWAYS PASS (NOMINAL GOLDEN SURFACE COGS) ]
// ------------------------------------------------------------------------------

func TestFrameUnit_Sec1_NominalPayloadViewWireSizeAndHeaderValid(t *testing.T) {
	f := common.MakeSovereignFrame(
		common.NodeAlpha, common.NodeBeta, data.Rung1Floor, data.OpPing, 1,
		[]byte("HELLO-FLOOR"),
	)
	if !f.IsHeaderBoundsValid() {
		t.Fatal("[SEC 1] sovereign frame failed IsHeaderBoundsValid()")
	}
	if !bytes.Equal(f.PayloadSlice(), []byte("HELLO-FLOOR")) {
		t.Fatalf("[SEC 1] PayloadSlice mismatch: %q", f.PayloadSlice())
	}
	if f.WireSize() != data.FrameHeaderSize+11 {
		t.Fatalf("[SEC 1] WireSize=%d want %d", f.WireSize(), data.FrameHeaderSize+11)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: [ SECTION 2: NEVER FAIL (FAULT INVERSION & DEGRADATION) ]
// ------------------------------------------------------------------------------

func TestFrameUnit_Sec2_OversizeRejectMalformedLenClampAndNilReceiver(t *testing.T) {
	var nilFrame *data.Frame
	if nilFrame.IsHeaderBoundsValid() {
		t.Fatal("[SEC 2] nil *Frame must fail IsHeaderBoundsValid()")
	}

	var f data.Frame
	f.SetPayload([]byte("KEEP-ME"))
	if f.SetPayload(make([]byte, data.FloorMaxPayload+1)) {
		t.Fatal("[SEC 2] SetPayload accepted 481B oversize payload")
	}
	if !bytes.Equal(f.PayloadSlice(), []byte("KEEP-ME")) {
		t.Fatal("[SEC 2] rejected SetPayload corrupted prior payload")
	}

	// Manually corrupt PayloadLen > 480; PayloadSlice must clamp safely without panic.
	f.PayloadLen = 9999
	if len(f.PayloadSlice()) != data.FloorMaxPayload {
		t.Fatalf("[SEC 2] malformed PayloadLen did not clamp to %d", data.FloorMaxPayload)
	}
	if f.IsHeaderBoundsValid() {
		t.Fatal("[SEC 2] malformed PayloadLen=9999 must fail IsHeaderBoundsValid()")
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: [ SECTION 3: INVARIANTS, SATURATION & LIMIT GUARDS ]
// ------------------------------------------------------------------------------

func TestFrameUnit_Sec3_Contiguous32ByteOffsetsAndZeroAllocReuse(t *testing.T) {
	if data.OffsetMagic != 0 ||
		data.OffsetVerRung != 2 ||
		data.OffsetOpcode != 3 ||
		data.OffsetFlags != 4 ||
		data.OffsetQuadState != 5 ||
		data.OffsetPayloadLen != 6 ||
		data.OffsetSourceID != 8 ||
		data.OffsetTargetID != 16 ||
		data.OffsetSeqID != 24 ||
		data.OffsetChecksum != 28 ||
		data.OffsetChecksum+4 != data.FrameHeaderSize {
		t.Fatal("[SEC 3] 32-byte wire header offset table has gap or misalignment")
	}
	if data.FrameMagicHi != 0xFA || data.FrameMagicLo != 0x17 || data.FrameMagic != 0xFA17 {
		t.Fatal("[SEC 3] FrameMagic 0xFA17 signature drifted")
	}

	var f data.Frame
	payload := []byte("ZERO-ALLOC-FRAME")
	allocs := testing.AllocsPerRun(500, func() {
		_ = f.SetPayload(payload)
		_ = f.PayloadSlice()
		_ = f.WireSize()
		f.Reset()
	})
	if allocs != 0 {
		t.Fatalf("[SEC 3] ALLOC_ZERO_HEAP violated on Frame methods: %.2f", allocs)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.4: [ SECTION 4: EMERGENCE & COMPOSITE WORKFLOWS ]
// ------------------------------------------------------------------------------

func TestFrameUnit_Sec4_ZeroLengthAndFullSectorResetCycle(t *testing.T) {
	var f data.Frame
	full := bytes.Repeat([]byte{0x7E}, data.FloorMaxPayload)
	if !f.SetPayload(full) || f.WireSize() != 512 {
		t.Fatal("[SEC 4] exact 480B fill failed")
	}
	f.Reset()
	if f.PayloadLen != 0 || f.WireSize() != data.FrameHeaderSize || f.Payload[0] != 0 {
		t.Fatal("[SEC 4] Reset() failed to clear inline payload backing array")
	}
	if !f.SetPayload(nil) || len(f.PayloadSlice()) != 0 {
		t.Fatal("[SEC 4] zero-length payload cycle failed")
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Preflight Diagnostic Gates & Physical Source Anchor Verification )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.data.frame
// Mission:       Run Frame contract validator, vocabulary audit, and anchor check.
// ==============================================================================

func TestFrameUnit_Block4_PreflightAndAnchor(t *testing.T) {
	if frameUnitSuiteName == "" {
		t.Fatal("[BLOCK 4] suite identifier must not be empty")
	}
	sample := common.MakeSovereignFrame(
		common.NodeAlpha, common.NodeBeta, data.Rung2Host, data.OpData, 1, []byte("OK"),
	)
	common.AssertAuditPass(t, engine.ValidateFrameContract(&sample), engine.ExecuteCodecDumpCard())
	common.AssertAuditFail(t, engine.ValidateFrameContract(nil), data.AuditSeverityCritical)
	common.AssertRepoAnchors(t, "data/frame.go")
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Test Suite Epilogue & Clean Readiness Handshake )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.data.frame
// Mission:       Confirm readiness handshake post-suite.
// ==============================================================================

func TestFrameUnit_Block5_EpilogueReadiness(t *testing.T) {
	if !common.IsHarnessReady() || !engine.IsCodecReady() {
		t.Fatal("[BLOCK 5: EPILOGUE] frame unit readiness handshake failed")
	}
}

// End of tests/unit/data/frame_test.go
