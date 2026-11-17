// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/functional/data_codec_test.go
// @id            test.functional.data_codec
// @layer         Layer 4: Tier-1 Functional Verification Suite
// @archetype     operational | stateless
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Layer 0/1 Data Layout & Zero-Alloc Codec Audit
// @role          Tier-1 Functional Diagnostic Sensor (4-Section Micro Protocol)
// @maturity      Rev 5 Baseline | 4x4 Verification Matrix
// @chassis       zero-heap | Verifies ALLOC_ZERO_HEAP via testing.AllocsPerRun
// @scale         O(1) deterministic functional tests & hot-path benchmarks
// @seams         in: data, engine, engine/covenant, tests/common
// @invariants    1. Macro Tier 1 (Functional): verifies public API surface.
//                2. Micro 4-Section Protocol: Always Pass, Never Fail,
//                   Invariants & Limits, Emergence.
//                3. FrameHeaderSize == 32 && FloorMaxFrameSize == 512.
//                4. EncodeInto and DecodeFrom execute with 0.0 allocs/op.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Tier-1 functional diagnostic sensor verifying FaithNet's Layer 0 data
//   vocabulary and Layer 1 binary wire codec across all 4 horizontal micro
//   sections: nominal golden paths, adversarial fault injection, hard sector
//   and zero-alloc ceilings, and multi-step encode/reflect/decode emergence.
//
//   ### For Architects, Operators & Non-Programmers:
//   Tests the postal envelope system from the outside in—confirming that good
//   envelopes pass cleanly, broken envelopes never crash the engine, 512-byte
//   floppy disk rules cannot be breached, and full ping conversations emerge
//   intact across the wire.
//
//   ### For Systems Programmers & Engine Architects:
//   Mirrors `engine/tests/functional/` from Cornerstone under Go's testing
//   runner with deterministic fixtures (`tests/common`) and zero-alloc audits.
// ==============================================================================

package functional_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Verification Imports & Tier-1 Sensor Anchor )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-1 Functional Verification Suite
// Module:        test.functional.data_codec
// Mission:       Bind Layer 0, Layer 1, and shared fixtures for Tier-1 probes.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"bytes"
	"strings"
	"testing"
)

// --- Tier 2: External Library Dependencies (None) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---
import (
	"github.com/CreativeWorkzStudio/FaithNet/data"
	"github.com/CreativeWorkzStudio/FaithNet/engine"
	"github.com/CreativeWorkzStudio/FaithNet/engine/covenant"
	"github.com/CreativeWorkzStudio/FaithNet/tests/common"
)

// ==============================================================================
// BLOCK 3: BODY BLOCK
// ( Four-Section Micro Testing Protocol: Pass / Never Fail / Invariants / Emerge )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-1 Functional Verification Suite
// Module:        test.functional.data_codec
// Mission:       Execute Section 1 (Always Pass), Section 2 (Never Fail),
//                Section 3 (Invariants & Limits), and Section 4 (Emergence).
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: SECTION 1 — ALWAYS PASS (Nominal Golden Path)
// ------------------------------------------------------------------------------

func TestSection1_AlwaysPass_CodecRoundTripAllRungs(t *testing.T) {
	if !common.AuditFixtures() {
		t.Fatal("shared fixtures failed self-audit")
	}

	rungs := []data.RungLevel{
		data.Rung1Floor,
		data.Rung2Host,
		data.Rung3Local,
		data.Rung4Global,
	}

	var wire [data.FloorMaxFrameSize]byte
	for idx, rung := range rungs {
		msg := []byte("MillenniumOS CPI-SI Sovereign Sync Payload")
		in := common.MakeSovereignFrame(
			common.NodeAlpha,
			common.NodeBeta,
			rung,
			data.OpCPISync,
			uint32(100+idx),
			msg,
		)

		n, err := engine.EncodeInto(wire[:], &in)
		if err != nil {
			t.Fatalf("rung %s encode failed: %v", rung, err)
		}
		if n != data.FrameHeaderSize+len(msg) {
			t.Fatalf("expected %d wire bytes, got %d", data.FrameHeaderSize+len(msg), n)
		}

		var out data.Frame
		if err := engine.DecodeFrom(wire[:n], &out); err != nil {
			t.Fatalf("rung %s decode failed: %v", rung, err)
		}
		common.AssertFramesEqual(t, &in, &out)
		if out.Checksum != in.Checksum {
			t.Fatalf("checksum mismatch: want 0x%08X got 0x%08X", in.Checksum, out.Checksum)
		}
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: SECTION 2 — NEVER FAIL (Fault Inversion & Adversarial Input)
// ------------------------------------------------------------------------------

func TestSection2_NeverFail_CorruptAndAdversarialWireInputs(t *testing.T) {
	var wire [data.FloorMaxFrameSize]byte
	var out data.Frame

	// Nil pointers & undersized buffers must fail cleanly without panicking.
	if _, err := engine.EncodeInto(wire[:], nil); err != engine.ErrBufferTooSmall {
		t.Fatalf("expected ErrBufferTooSmall on nil encode, got %v", err)
	}
	if err := engine.DecodeFrom(wire[:10], &out); err != engine.ErrBufferTooSmall {
		t.Fatalf("expected ErrBufferTooSmall on short wire, got %v", err)
	}
	if err := engine.DecodeFrom(wire[:], nil); err != engine.ErrBufferTooSmall {
		t.Fatalf("expected ErrBufferTooSmall on nil decode target, got %v", err)
	}

	// Encode valid baseline frame, then invert every header/body invariant.
	in := common.MakeSovereignFrame(
		common.NodeAlpha, common.NodeBeta, data.Rung1Floor, data.OpPing, 42,
		[]byte("SONAR-ECHO-01"),
	)
	n, err := engine.EncodeInto(wire[:], &in)
	if err != nil {
		t.Fatalf("baseline encode failed: %v", err)
	}

	// 1. Single-bit payload flip -> ErrChecksumMismatch.
	wire[data.FrameHeaderSize+3] ^= 0x01
	if err := engine.DecodeFrom(wire[:n], &out); err != engine.ErrChecksumMismatch {
		t.Fatalf("expected ErrChecksumMismatch on payload flip, got %v", err)
	}
	wire[data.FrameHeaderSize+3] ^= 0x01

	// 2. Truncated payload slice -> ErrTruncatedPayload.
	if err := engine.DecodeFrom(wire[:n-2], &out); err != engine.ErrTruncatedPayload {
		t.Fatalf("expected ErrTruncatedPayload on clipped wire, got %v", err)
	}

	// 3. Invalid magic byte -> ErrBadMagic.
	badMagic := append([]byte(nil), wire[:n]...)
	badMagic[0] = 0x00
	if err := engine.DecodeFrom(badMagic, &out); err != engine.ErrBadMagic {
		t.Fatalf("expected ErrBadMagic, got %v", err)
	}

	// 4. Unsupported version nibble -> ErrUnsupportedVer.
	badVer := append([]byte(nil), wire[:n]...)
	badVer[data.OffsetVerRung] = (0x09 << 4) | uint8(data.Rung1Floor)
	if err := engine.DecodeFrom(badVer, &out); err != engine.ErrUnsupportedVer {
		t.Fatalf("expected ErrUnsupportedVer, got %v", err)
	}

	// 5. Invalid rung nibble -> ErrInvalidRung.
	badRung := append([]byte(nil), wire[:n]...)
	badRung[data.OffsetVerRung] = (data.ProtocolVersion << 4) | 0x07
	if err := engine.DecodeFrom(badRung, &out); err != engine.ErrInvalidRung {
		t.Fatalf("expected ErrInvalidRung, got %v", err)
	}

	// 6. Invalid opcode -> ErrInvalidOpcode.
	badOp := append([]byte(nil), wire[:n]...)
	badOp[data.OffsetOpcode] = 0xFF
	if err := engine.DecodeFrom(badOp, &out); err != engine.ErrInvalidOpcode {
		t.Fatalf("expected ErrInvalidOpcode, got %v", err)
	}

	// 7. Invalid QuadState -> ErrInvalidQuadState.
	badT4 := append([]byte(nil), wire[:n]...)
	badT4[data.OffsetQuadState] = 0x09
	if err := engine.DecodeFrom(badT4, &out); err != engine.ErrInvalidQuadState {
		t.Fatalf("expected ErrInvalidQuadState, got %v", err)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: SECTION 3 — INVARIANTS & LIMITS (512B Ceiling & 0 Allocs)
// ------------------------------------------------------------------------------

func TestSection3_Invariants_FloppySectorAndZeroHeap(t *testing.T) {
	if data.FrameHeaderSize != 32 || data.FloorMaxPayload != 480 ||
		data.FloorMaxFrameSize != 512 {
		t.Fatalf("512B Floppy Sector geometry violated: hdr=%d pay=%d tot=%d",
			data.FrameHeaderSize, data.FloorMaxPayload, data.FloorMaxFrameSize)
	}

	// Exact 480-byte ceiling succeeds; 481-byte payload is rejected.
	var f data.Frame
	fullPayload := bytes.Repeat([]byte{0x5A}, data.FloorMaxPayload)
	if !f.SetPayload(fullPayload) {
		t.Fatal("SetPayload rejected exact 480-byte floor ceiling")
	}
	if f.WireSize() != 512 {
		t.Fatalf("expected exact 512-byte WireSize, got %d", f.WireSize())
	}
	oversize := make([]byte, data.FloorMaxPayload+1)
	if f.SetPayload(oversize) {
		t.Fatal("SetPayload accepted 481-byte payload violating 512B invariant")
	}

	// ALLOC_ZERO_HEAP proof: EncodeInto and DecodeFrom must do 0.0 allocs/op.
	var wire [data.FloorMaxFrameSize]byte
	var out data.Frame
	in := common.MakeSovereignFrame(
		common.NodeAlpha, common.NodeBeta, data.Rung2Host, data.OpPing, 1,
		[]byte("ZERO-HEAP-PROOF"),
	)

	encAllocs := testing.AllocsPerRun(1000, func() {
		_, _ = engine.EncodeInto(wire[:], &in)
	})
	if encAllocs != 0 {
		t.Fatalf("ALLOC_ZERO_HEAP violated in EncodeInto: %.2f allocs/op", encAllocs)
	}

	n, _ := engine.EncodeInto(wire[:], &in)
	decAllocs := testing.AllocsPerRun(1000, func() {
		_ = engine.DecodeFrom(wire[:n], &out)
	})
	if decAllocs != 0 {
		t.Fatalf("ALLOC_ZERO_HEAP violated in DecodeFrom: %.2f allocs/op", decAllocs)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.4: SECTION 4 — EMERGENCE (Full Wire Ping->Reflect->Pong Loop)
// ------------------------------------------------------------------------------

func TestSection4_Emergence_FullWireHeartbeatConversation(t *testing.T) {
	clk := engine.NewMockClock(50_000)
	nodeA := engine.NewPingEngine(common.NodeAlpha, clk)
	airlock := covenant.NewFilter()

	var req, wireIn, resp, wireOut data.Frame
	var bufA, bufB [data.FloorMaxFrameSize]byte

	// Step 1: Node A builds Ping -> encodes into wire bytes.
	if err := nodeA.BuildPingInto(&req, common.NodeBeta, data.Rung1Floor,
		[]byte("ROOM-PRESENCE")); err != nil {
		t.Fatalf("BuildPingInto failed: %v", err)
	}
	nA, err := engine.EncodeInto(bufA[:], &req)
	if err != nil {
		t.Fatalf("EncodeInto A->B failed: %v", err)
	}

	// Step 2: Travel delay -> Node B decodes, audits via Airlock, reflects Pong.
	clk.Advance(750)
	if err := engine.DecodeFrom(bufA[:nA], &wireIn); err != nil {
		t.Fatalf("DecodeFrom at B failed: %v", err)
	}
	if v := airlock.EvaluateFrame(&wireIn); v.State != data.QuadPositive {
		t.Fatalf("sovereign ping rejected by airlock: %+v", v)
	}
	if err := engine.ReflectPingInto(&wireIn, &resp, common.NodeBeta); err != nil {
		t.Fatalf("ReflectPingInto failed: %v", err)
	}
	nB, err := engine.EncodeInto(bufB[:], &resp)
	if err != nil {
		t.Fatalf("EncodeInto B->A failed: %v", err)
	}

	// Step 3: Return delay -> Node A decodes Pong and computes exact 1500ns RTT.
	clk.Advance(750)
	if err := engine.DecodeFrom(bufB[:nB], &wireOut); err != nil {
		t.Fatalf("DecodeFrom at A failed: %v", err)
	}
	rtt, ok := nodeA.ProcessPong(&wireOut)
	if !ok || rtt != 1500 {
		t.Fatalf("expected emergent round-trip RTT=1500ns, got ok=%v rtt=%d", ok, rtt)
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Block 4 Preflight Gates, 60-Col Dump Cards & Throughput Benchmarks )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-1 Functional Verification Suite
// Module:        test.functional.data_codec
// Mission:       Confirm every Layer 0/1 Preflight gate passes and benchmark
//                hot-path zero-alloc wire throughput.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Block 4 Preflight Gate & Diagnostic Card Audit
// ------------------------------------------------------------------------------

func TestAudit_Layer0And1PreflightAndDumpCards(t *testing.T) {
	if !engine.AuditDataVocabulary() || !engine.PreflightCodec() ||
		!engine.PreflightClock() || !engine.PreflightPingEngine() ||
		!engine.PreflightRouter() || !covenant.PreflightCovenantFilter() {
		t.Fatal("one or more Layer 0/1 Block 4 Preflight gates failed")
	}

	codecCard := engine.ExecuteCodecDumpCard()
	if !strings.Contains(codecCard, "PASS") || !strings.Contains(codecCard, "512 bytes") {
		t.Fatalf("unexpected codec dump card:\n%s", codecCard)
	}
	clkCard := engine.ExecuteClockDumpCard(engine.NewMockClock(42))
	if !strings.Contains(clkCard, "DETERMINISTIC_MOCK_ATOMIC") {
		t.Fatalf("unexpected clock dump card:\n%s", clkCard)
	}
	airlockCard := covenant.NewFilter().ExecuteAirlockDumpCard()
	if !strings.Contains(airlockCard, "PASS") {
		t.Fatalf("unexpected airlock dump card:\n%s", airlockCard)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Zero-Allocation Wire Benchmarks
// ------------------------------------------------------------------------------

func BenchmarkEncodeInto(b *testing.B) {
	in := common.MakeSovereignFrame(
		common.NodeAlpha, common.NodeBeta, data.Rung2Host, data.OpPing, 1,
		[]byte("FAITHNET-BENCH"),
	)
	var wire [data.FloorMaxFrameSize]byte

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.EncodeInto(wire[:], &in)
	}
}

func BenchmarkDecodeFrom(b *testing.B) {
	in := common.MakeSovereignFrame(
		common.NodeAlpha, common.NodeBeta, data.Rung2Host, data.OpPing, 1,
		[]byte("FAITHNET-BENCH"),
	)
	var out data.Frame
	var wire [data.FloorMaxFrameSize]byte
	n, _ := engine.EncodeInto(wire[:], &in)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = engine.DecodeFrom(wire[:n], &out)
	}
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK / STATUTORY EXEMPTION & LIFECYCLE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-1 Functional Verification Suite
// Module:        test.functional.data_codec
// Mission:       Test harness managed by Go testing runner.
// Stateless Exemption Notice: None. (Stateless test suite).
// ==============================================================================

// End of tests/functional/data_codec_test.go
