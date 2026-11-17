// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/unit/engine/codec_test.go
// @id            test.unit.engine.codec
// @layer         Layer 4: Tier-2 Unit Verification Suite (Engine Sub-Stratum)
// @archetype     verification_suite | unit_isolation
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Layer 1 Binary Wire Codec Unit Suite
// @role          Tier-2 Unit Isolation Gate for cog.engine.codec
// @maturity      Rev 5 Baseline | 4x4 Verification Matrix
// @chassis       zero-heap | Verifies ALLOC_ZERO_HEAP on EncodeInto/DecodeFrom
// @scale         O(N_frame) exhaustive error-branch & checksum sensitivity sweep
// @seams         in: data, engine, tests/common | target: engine/codec.go
// @invariants    1. Every single sentinel error in codec.go is individually
//                   triggered and verified.
//                2. Failing DecodeFrom NEVER mutates caller's output Frame.
//                3. EncodeInto and DecodeFrom execute with 0.0 allocs/op.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Tier-2 Unit isolation suite strictly verifying `engine/codec.go` across all
//   encode/decode branches, FNV-1a bit-flip sensitivity, non-mutation guards,
//   512-byte ceiling bounds, and structured Block 4 pre-flight diagnostics.
//
//   ### For Architects, Operators & Non-Programmers:
//   Tests every single failure mode of FaithNet's envelope packer and tamper
//   seal so malformed or tampered bytes from any wire are stopped cold.
//
//   ### For Systems Programmers & Engine Architects:
//   Exhaustive branch coverage of `EncodeInto`, `DecodeFrom`,
//   `ComputeFrameChecksum`, `ValidateFrameContract`, and `PreflightCodec`.
// ==============================================================================

package engine_unit_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Identifier, Verification Imports & Tier-2 Unit Sensor Anchor )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.engine.codec
// Mission:       Bind Layer 0/1 codec symbols and shared harness.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"bytes"
	"encoding/binary"
	"strings"
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
	codecUnitSuiteName = "Tier 2 Unit: cog.engine.codec (32B Wire & FNV-1a)"
)

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.engine.codec
// Mission:       Verify every Codec branch, error sentinel, and seal invariant.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: [ SECTION 1: ALWAYS PASS (NOMINAL GOLDEN SURFACE COGS) ]
// ------------------------------------------------------------------------------

func TestCodecUnit_Sec1_ZeroVersionDefaultsAndEmptyPayloadRoundTrip(t *testing.T) {
	var in, out data.Frame
	var wire [data.FloorMaxFrameSize]byte

	in.Version = 0 // Must default automatically to ProtocolVersion.
	in.Rung = data.Rung1Floor
	in.Opcode = data.OpPing
	in.Flags = data.FlagFloorCompact
	in.Truth = data.QuadPositive
	in.SourceID = common.NodeAlpha
	in.TargetID = common.NodeBeta
	in.SeqID = 1

	n, err := engine.EncodeInto(wire[:], &in)
	if err != nil || n != data.FrameHeaderSize {
		t.Fatalf("[SEC 1] empty-payload encode failed: n=%d err=%v", n, err)
	}
	if in.Version != data.ProtocolVersion {
		t.Fatalf("[SEC 1] Version=0 did not default to %d", data.ProtocolVersion)
	}
	if err := engine.DecodeFrom(wire[:n], &out); err != nil {
		t.Fatalf("[SEC 1] empty-payload decode failed: %v", err)
	}
	common.AssertFramesEqual(t, &in, &out)
}

func TestCodecUnit_Sec1_FNV1aDeterminismAndBitSensitivity(t *testing.T) {
	in := common.MakeSovereignFrame(
		common.NodeAlpha, common.NodeBeta, data.Rung3Local, data.OpData, 9,
		[]byte("DETERMINISTIC-SEAL"),
	)
	var w1, w2 [data.FloorMaxFrameSize]byte
	n1, _ := engine.EncodeInto(w1[:], &in)
	n2, _ := engine.EncodeInto(w2[:], &in)
	if n1 != n2 || !bytes.Equal(w1[:n1], w2[:n2]) {
		t.Fatal("[SEC 1] identical frames did not produce identical wire bytes")
	}
	sum1 := engine.ComputeFrameChecksum(w1[:n1], int(in.PayloadLen))
	w2[data.OffsetSeqID] ^= 0x01
	sum2 := engine.ComputeFrameChecksum(w2[:n2], int(in.PayloadLen))
	if sum1 == sum2 {
		t.Fatal("[SEC 1] single-bit header flip produced identical FNV-1a checksum")
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: [ SECTION 2: NEVER FAIL (FAULT INVERSION & DEGRADATION) ]
// ------------------------------------------------------------------------------

func TestCodecUnit_Sec2_EncodeIntoEveryErrorBranch(t *testing.T) {
	var wire [data.FloorMaxFrameSize]byte
	valid := common.MakeSovereignFrame(
		common.NodeAlpha, common.NodeBeta, data.Rung2Host, data.OpData, 1, []byte("ABCD"),
	)

	if _, err := engine.EncodeInto(wire[:], nil); err != engine.ErrBufferTooSmall {
		t.Fatalf("[SEC 2] nil frame: want ErrBufferTooSmall got %v", err)
	}
	if _, err := engine.EncodeInto(wire[:data.FrameHeaderSize+2], &valid); err != engine.ErrBufferTooSmall {
		t.Fatalf("[SEC 2] short dst: want ErrBufferTooSmall got %v", err)
	}

	over := valid
	over.PayloadLen = uint16(data.FloorMaxPayload + 1)
	if _, err := engine.EncodeInto(wire[:], &over); err != engine.ErrPayloadTooLarge {
		t.Fatalf("[SEC 2] oversize payload: want ErrPayloadTooLarge got %v", err)
	}

	badVer := valid
	badVer.Version = 0x10
	if _, err := engine.EncodeInto(wire[:], &badVer); err != engine.ErrUnsupportedVer {
		t.Fatalf("[SEC 2] version > 0x0F: want ErrUnsupportedVer got %v", err)
	}

	badRung := valid
	badRung.Rung = data.RungNone
	if _, err := engine.EncodeInto(wire[:], &badRung); err != engine.ErrInvalidRung {
		t.Fatalf("[SEC 2] bad rung: want ErrInvalidRung got %v", err)
	}

	badOp := valid
	badOp.Opcode = data.OpInvalid
	if _, err := engine.EncodeInto(wire[:], &badOp); err != engine.ErrInvalidOpcode {
		t.Fatalf("[SEC 2] bad opcode: want ErrInvalidOpcode got %v", err)
	}

	badT4 := valid
	badT4.Truth = data.QuadState(0x09)
	if _, err := engine.EncodeInto(wire[:], &badT4); err != engine.ErrInvalidQuadState {
		t.Fatalf("[SEC 2] bad QuadState: want ErrInvalidQuadState got %v", err)
	}
}

func TestCodecUnit_Sec2_DecodeFromEveryErrorBranchAndNonMutation(t *testing.T) {
	in := common.MakeSovereignFrame(
		common.NodeAlpha, common.NodeBeta, data.Rung1Floor, data.OpPing, 77,
		[]byte("PROTECT-OUT"),
	)
	var wire [data.FloorMaxFrameSize]byte
	n, err := engine.EncodeInto(wire[:], &in)
	if err != nil {
		t.Fatalf("setup encode failed: %v", err)
	}

	// Pre-populate out with sentinel values; none of the failing decodes may touch it.
	out := common.MakeSovereignFrame(
		common.NodeRelay, common.NodeAlpha, data.Rung3Local, data.OpBeacon, 555, []byte("SENTINEL"),
	)

	// 1. Short wire & nil out.
	if err := engine.DecodeFrom(wire[:12], &out); err != engine.ErrBufferTooSmall {
		t.Fatalf("[SEC 2] short wire: got %v", err)
	}
	if err := engine.DecodeFrom(wire[:n], nil); err != engine.ErrBufferTooSmall {
		t.Fatalf("[SEC 2] nil out: got %v", err)
	}

	// 2. Bad magic.
	w := append([]byte(nil), wire[:n]...)
	w[0] = 0x00
	if err := engine.DecodeFrom(w, &out); err != engine.ErrBadMagic {
		t.Fatalf("[SEC 2] bad magic: got %v", err)
	}

	// 3. Unsupported version nibble.
	w = append([]byte(nil), wire[:n]...)
	w[data.OffsetVerRung] = (0x09 << 4) | uint8(data.Rung1Floor)
	if err := engine.DecodeFrom(w, &out); err != engine.ErrUnsupportedVer {
		t.Fatalf("[SEC 2] bad version: got %v", err)
	}

	// 4. Invalid rung nibble.
	w = append([]byte(nil), wire[:n]...)
	w[data.OffsetVerRung] = (data.ProtocolVersion << 4) | 0x07
	if err := engine.DecodeFrom(w, &out); err != engine.ErrInvalidRung {
		t.Fatalf("[SEC 2] bad rung: got %v", err)
	}

	// 5. Invalid opcode.
	w = append([]byte(nil), wire[:n]...)
	w[data.OffsetOpcode] = 0xFF
	if err := engine.DecodeFrom(w, &out); err != engine.ErrInvalidOpcode {
		t.Fatalf("[SEC 2] bad opcode: got %v", err)
	}

	// 6. Invalid QuadState.
	w = append([]byte(nil), wire[:n]...)
	w[data.OffsetQuadState] = 0x09
	if err := engine.DecodeFrom(w, &out); err != engine.ErrInvalidQuadState {
		t.Fatalf("[SEC 2] bad QuadState: got %v", err)
	}

	// 7. Declared PayloadLen > 480 -> ErrPayloadTooLarge.
	w = append([]byte(nil), wire[:n]...)
	binary.BigEndian.PutUint16(w[data.OffsetPayloadLen:], uint16(data.FloorMaxPayload+1))
	if err := engine.DecodeFrom(w, &out); err != engine.ErrPayloadTooLarge {
		t.Fatalf("[SEC 2] oversize PayloadLen: got %v", err)
	}

	// 8. Truncated payload -> ErrTruncatedPayload.
	if err := engine.DecodeFrom(wire[:n-1], &out); err != engine.ErrTruncatedPayload {
		t.Fatalf("[SEC 2] truncated wire: got %v", err)
	}

	// 9. Single-bit payload corruption -> ErrChecksumMismatch.
	w = append([]byte(nil), wire[:n]...)
	w[data.FrameHeaderSize+2] ^= 0x80
	if err := engine.DecodeFrom(w, &out); err != engine.ErrChecksumMismatch {
		t.Fatalf("[SEC 2] checksum flip: got %v", err)
	}

	// Non-mutation guarantee: sentinel frame must remain untouched!
	if out.SeqID != 555 || string(out.PayloadSlice()) != "SENTINEL" {
		t.Fatalf("[SEC 2] failing DecodeFrom mutated caller output frame: %+v", out)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: [ SECTION 3: INVARIANTS, SATURATION & LIMIT GUARDS ]
// ------------------------------------------------------------------------------

func TestCodecUnit_Sec3_Exact512ByteSectorAndZeroAlloc(t *testing.T) {
	in := common.MakeSovereignFrame(
		common.NodeAlpha, common.NodeBeta, data.Rung4Global, data.OpData, 1024,
		bytes.Repeat([]byte{0x3C}, data.FloorMaxPayload),
	)
	var wire [data.FloorMaxFrameSize]byte
	var out data.Frame

	n, err := engine.EncodeInto(wire[:], &in)
	if err != nil || n != data.FloorMaxFrameSize {
		t.Fatalf("[SEC 3] exact 512B encode failed: n=%d err=%v", n, err)
	}
	if err := engine.DecodeFrom(wire[:n], &out); err != nil {
		t.Fatalf("[SEC 3] exact 512B decode failed: %v", err)
	}
	common.AssertFramesEqual(t, &in, &out)

	allocs := testing.AllocsPerRun(500, func() {
		wn, _ := engine.EncodeInto(wire[:], &in)
		_ = engine.DecodeFrom(wire[:wn], &out)
	})
	if allocs != 0 {
		t.Fatalf("[SEC 3] ALLOC_ZERO_HEAP violated: %.2f allocs/op", allocs)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.4: [ SECTION 4: EMERGENCE & COMPOSITE WORKFLOWS ]
// ------------------------------------------------------------------------------

func TestCodecUnit_Sec4_ValidateFrameContractEveryViolation(t *testing.T) {
	base := common.MakeSovereignFrame(
		common.NodeAlpha, common.NodeBeta, data.Rung2Host, data.OpCPISync, 1, []byte("OK"),
	)
	common.AssertAuditPass(t, engine.ValidateFrameContract(&base), engine.ExecuteCodecDumpCard())

	badVer := base
	badVer.Version = 9
	common.AssertAuditFail(t, engine.ValidateFrameContract(&badVer), data.AuditSeverityCritical)

	badRung := base
	badRung.Rung = data.RungNone
	common.AssertAuditFail(t, engine.ValidateFrameContract(&badRung), data.AuditSeverityCritical)

	badOp := base
	badOp.Opcode = data.OpInvalid
	common.AssertAuditFail(t, engine.ValidateFrameContract(&badOp), data.AuditSeverityCritical)

	badT4 := base
	badT4.Truth = data.QuadState(9)
	common.AssertAuditFail(t, engine.ValidateFrameContract(&badT4), data.AuditSeverityCritical)

	over := base
	over.PayloadLen = uint16(data.FloorMaxPayload + 1)
	common.AssertAuditFail(t, engine.ValidateFrameContract(&over), data.AuditSeverityFatal)
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Preflight Diagnostic Gates & Physical Source Anchor Verification )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.engine.codec
// Mission:       Run PreflightCodec, verify 60-col dump card & repo anchor.
// ==============================================================================

func TestCodecUnit_Block4_PreflightCardAndAnchor(t *testing.T) {
	if codecUnitSuiteName == "" {
		t.Fatal("[BLOCK 4] suite identifier must not be empty")
	}
	card := engine.ExecuteCodecDumpCard()
	common.AssertAuditPass(t, engine.PreflightCodec(), card)
	if !engine.AuditCodec() || !strings.Contains(card, "PASS") {
		t.Fatalf("[BLOCK 4] AuditCodec or dump card failed:\n%s", card)
	}
	common.AssertRepoAnchors(t, "engine/codec.go")
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Test Suite Epilogue & Clean Readiness Handshake )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.engine.codec
// Mission:       Confirm codec and harness epilogue readiness post-suite.
// ==============================================================================

func TestCodecUnit_Block5_EpilogueReadiness(t *testing.T) {
	if !engine.IsCodecReady() || !common.IsHarnessReady() {
		t.Fatal("[BLOCK 5: EPILOGUE] codec unit readiness handshake failed")
	}
}

// End of tests/unit/engine/codec_test.go
