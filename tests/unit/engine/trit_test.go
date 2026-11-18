// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/unit/engine/trit_test.go
// @id            test.unit.engine.trit
// @layer         Layer 4: Tier-2 Unit Verification Suite (Engine Sub-Stratum)
// @archetype     verification_suite | unit_isolation
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Layer 1 Bit <-> Balanced-Trit <-> Bit Unit Suite
// @role          Tier-2 Unit Isolation Gate for cog.engine.trit
// @maturity      Rev 5 Baseline | Wave 15 OmniCode Submodule Ternary Bridge
// @chassis       zero-heap | Verifies 0.00 allocs/op on all ternary hot paths
// @scale         O(243 + 256) exhaustive bijection & zero-alloc verification
// @seams         in: data, engine, tests/common | target: engine/trit.go
// @invariants    1. Exhaustive 243-state Tryte5 <-> byte bijection verified.
//                2. Exhaustive 256-byte bit -> 6-trit -> bit lossless identity.
//                3. Zero heap allocations verified via testing.AllocsPerRun.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Strict 1:1 Tier-2 unit verification suite for `engine/trit.go`, proving
//   lossless binary <-> balanced-ternary roundtrips, sentinel rejection,
//   T4 QuadState NULL isolation, and zero-heap execution.
// ==============================================================================

package engine_unit_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ==============================================================================

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/CreativeWorkzStudio/FaithNet/data"
	"github.com/CreativeWorkzStudio/FaithNet/engine"
	"github.com/CreativeWorkzStudio/FaithNet/tests/common"
)

const tritUnitSuiteName = "Tier 2 Unit: cog.engine.trit (Bit <-> Balanced-Trit <-> Bit)"

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================

// --- Section 1: Always Pass (Nominal Golden Surface) ---

func TestTritUnit_Sec1_Exhaustive243Tryte5Bijection(t *testing.T) {
	for b := 0; b < engine.Tryte5StateCount; b++ {
		w, ok := engine.UnpackTryte5(byte(b))
		if !ok || !w.IsValid() {
			t.Fatalf("[SEC 1] UnpackTryte5(%d) failed", b)
		}
		got, ok := engine.PackTryte5(w)
		if !ok || got != byte(b) {
			t.Fatalf("[SEC 1] PackTryte5 mismatch for %d: got %d", b, got)
		}
	}
}

func TestTritUnit_Sec1_Exhaustive256ByteBitTritBitRoundtrip(t *testing.T) {
	var in [256]byte
	for i := range in {
		in[i] = byte(i)
	}
	var trits [256 * engine.TritsPerBinaryByte]engine.Trit
	var out [256]byte

	nT, err := engine.BytesToBalancedTrits(trits[:], in[:])
	if err != nil || nT != len(trits) {
		t.Fatalf("[SEC 1] BytesToBalancedTrits failed: n=%d err=%v", nT, err)
	}
	for i, tr := range trits {
		if !tr.IsValid() {
			t.Fatalf("[SEC 1] non-canonical trit %d at index %d", tr, i)
		}
	}
	nB, err := engine.BalancedTritsToBytes(out[:], trits[:])
	if err != nil || nB != len(in) {
		t.Fatalf("[SEC 1] BalancedTritsToBytes failed: n=%d err=%v", nB, err)
	}
	if !bytes.Equal(in[:], out[:]) {
		t.Fatalf("[SEC 1] bit -> balanced-trit -> bit roundtrip diverged")
	}
}

func TestTritUnit_Sec1_EncodeDecodeSliceRoundtrip(t *testing.T) {
	in := []engine.Trit{
		engine.TritNeg, engine.TritZero, engine.TritPos, engine.TritNeg, engine.TritPos,
		engine.TritPos, engine.TritPos, engine.TritZero, engine.TritNeg, engine.TritNeg,
	}
	var wire [2]byte
	n, err := engine.EncodeTritsInto(wire[:], in)
	if err != nil || n != 2 {
		t.Fatalf("[SEC 1] EncodeTritsInto failed: n=%d err=%v", n, err)
	}
	var decoded [10]engine.Trit
	m, err := engine.DecodeTritsFrom(decoded[:], wire[:n])
	if err != nil || m != 10 {
		t.Fatalf("[SEC 1] DecodeTritsFrom failed: m=%d err=%v", m, err)
	}
	for i := range in {
		if decoded[i] != in[i] {
			t.Fatalf("[SEC 1] trit mismatch at %d: got %v want %v", i, decoded[i], in[i])
		}
	}
}

// --- Section 2: Never Fail (Fault Rejection & Sentinels) ---

func TestTritUnit_Sec2_RejectsSentinelsInvalidTritsAndBadBounds(t *testing.T) {
	// All 13 reserved sentinels (0xF3..0xFF) rejected cleanly.
	for s := int(engine.Tryte5MaxEncodedByte) + 1; s <= 255; s++ {
		if !engine.IsTritSentinelByte(byte(s)) {
			t.Fatalf("[SEC 2] IsTritSentinelByte(0x%02X) expected true", s)
		}
		if _, ok := engine.UnpackTryte5(byte(s)); ok {
			t.Fatalf("[SEC 2] UnpackTryte5 accepted sentinel 0x%02X", s)
		}
	}
	var trits [5]engine.Trit
	if _, err := engine.DecodeTritsFrom(trits[:], []byte{engine.TritNullSentinel}); !errors.Is(err, engine.ErrTritSentinelByte) {
		t.Fatalf("[SEC 2] expected ErrTritSentinelByte, got %v", err)
	}
	var wire [1]byte
	if _, err := engine.EncodeTritsInto(wire[:], []engine.Trit{engine.Trit(2), 0, 0, 0, 0}); !errors.Is(err, engine.ErrTritInvalid) {
		t.Fatalf("[SEC 2] expected ErrTritInvalid, got %v", err)
	}
	if _, err := engine.EncodeTritsInto(wire[:], []engine.Trit{0, 0, 0}); !errors.Is(err, engine.ErrTritBadLength) {
		t.Fatalf("[SEC 2] expected ErrTritBadLength, got %v", err)
	}
	if _, err := engine.EncodeTritsInto(nil, []engine.Trit{0, 0, 0, 0, 0}); !errors.Is(err, engine.ErrTritBufferTooSmall) {
		t.Fatalf("[SEC 2] expected ErrTritBufferTooSmall, got %v", err)
	}
	// Extreme 6-trit value (+364 -> shifted +492 > 255) rejected cleanly.
	var out [1]byte
	maxTrits := []engine.Trit{1, 1, 1, 1, 1, 1}
	if _, err := engine.BalancedTritsToBytes(out[:], maxTrits); !errors.Is(err, engine.ErrTritByteRange) {
		t.Fatalf("[SEC 2] expected ErrTritByteRange for +364, got %v", err)
	}
}

// --- Section 3: Invariants & Limits (Zero-Heap ALLOC_ZERO_HEAP Proof) ---

func TestTritUnit_Sec3_ZeroHeapAllocationsOnHotPaths(t *testing.T) {
	payload := []byte("FaithNet OmniCode Bit<->Trit<->Bit Zero Heap!")
	tritBuf := make([]engine.Trit, len(payload)*engine.TritsPerBinaryByte)
	outBuf := make([]byte, len(payload))

	allocs := testing.AllocsPerRun(500, func() {
		_, _ = engine.BytesToBalancedTrits(tritBuf, payload)
		_, _ = engine.BalancedTritsToBytes(outBuf, tritBuf)
	})
	if allocs != 0 {
		t.Fatalf("[SEC 3] ALLOC_ZERO_HEAP violated: got %.2f allocs/op, want 0", allocs)
	}
}

// --- Section 4: Emergence (T4 Bridge, Preflight & Dump Card) ---

func TestTritUnit_Sec4_QuadStateBridgePreflightAndCard(t *testing.T) {
	if tr, ok := engine.QuadStateToTrit(data.QuadPositive); !ok || tr != engine.TritPos {
		t.Fatalf("[SEC 4] QuadPositive -> TritPos failed")
	}
	if tr, ok := engine.QuadStateToTrit(data.QuadAwait); !ok || tr != engine.TritZero {
		t.Fatalf("[SEC 4] QuadAwait -> TritZero failed")
	}
	if tr, ok := engine.QuadStateToTrit(data.QuadNegative); !ok || tr != engine.TritNeg {
		t.Fatalf("[SEC 4] QuadNegative -> TritNeg failed")
	}
	if _, ok := engine.QuadStateToTrit(data.QuadNull); ok {
		t.Fatalf("[SEC 4] QuadNull must stay isolated (ok=false)")
	}
	if engine.TritToQuadState(engine.Trit(9)) != data.QuadNull {
		t.Fatalf("[SEC 4] invalid trit must map defensively to QuadNull")
	}
	common.AssertAuditPass(t, engine.PreflightTritCodec(), "PreflightTritCodec")
	if !engine.IsTritCodecReady() {
		t.Fatalf("[SEC 4] IsTritCodecReady returned false")
	}
	card := engine.ExecuteTritCodecDumpCard()
	if !strings.Contains(card, "OMNICODE BRIDGE") {
		t.Fatalf("[SEC 4] dump card missing header:\n%s", card)
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ==============================================================================

func TestTritUnit_Block4_SuiteSummaryAudit(t *testing.T) {
	t.Logf("%s: 243-state Tryte5, 256-byte bit<->trit<->bit & 0-alloc verified.", tritUnitSuiteName)
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ==============================================================================
// End of tests/unit/engine/trit_test.go
