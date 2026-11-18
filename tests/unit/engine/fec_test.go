// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ==============================================================================
// @project       FaithNet
// @file          tests/unit/engine/fec_test.go
// @id            test.unit.engine.fec
// @layer         Layer 4: Tier-2 Unit Verification Suite (Engine Sub-Stratum)
// @archetype     verification_suite | unit_isolation
// @version       0.2.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @system        FaithNet OSI L2 SECDED Hamming(8,4) FEC Unit Suite
// @role          Strict 1:1 Unit Isolation Gate for cog.engine.fec
// @chassis       zero-heap | Verifies 0.00 allocs/op on EncodeFECInto/DecodeFECFrom
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

	"github.com/CreativeWorkzStudio/FaithNet/engine"
	"github.com/CreativeWorkzStudio/FaithNet/tests/common"
)

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================

func TestFECUnit_Sec1_Exhaustive128SingleBitSelfHeal(t *testing.T) {
	for nib := byte(0); nib < 16; nib++ {
		cw := engine.EncodeNibbleFEC(nib)
		clean, corr, err := engine.DecodeNibbleFEC(cw)
		if err != nil || corr || clean != nib {
			t.Fatalf("[SEC 1] clean nibble %d decode failed: got=%d corr=%v err=%v", nib, clean, corr, err)
		}
		for bit := uint(0); bit < 8; bit++ {
			healed, wasCorr, err := engine.DecodeNibbleFEC(cw ^ (1 << bit))
			if err != nil || !wasCorr || healed != nib {
				t.Fatalf("[SEC 1] single-bit heal failed nib=%d bit=%d: got=%d corr=%v err=%v",
					nib, bit, healed, wasCorr, err)
			}
		}
	}
}

func TestFECUnit_Sec2_DetectsDoubleBitAndBoundsFaults(t *testing.T) {
	cw := engine.EncodeNibbleFEC(0x0D) ^ 0x05
	if _, _, err := engine.DecodeNibbleFEC(cw); !errors.Is(err, engine.ErrFECUncorrectable) {
		t.Fatalf("[SEC 2] expected ErrFECUncorrectable for 2-bit flip, got %v", err)
	}
	var dst [4]byte
	if _, err := engine.EncodeFECInto(dst[:1], []byte{0xAB}); !errors.Is(err, engine.ErrFECBufferTooSmall) {
		t.Fatalf("[SEC 2] expected ErrFECBufferTooSmall, got %v", err)
	}
	if _, _, err := engine.DecodeFECFrom(dst[:], []byte{0x00, 0x00, 0x00}); !errors.Is(err, engine.ErrFECBadLength) {
		t.Fatalf("[SEC 2] expected ErrFECBadLength for odd slice, got %v", err)
	}
}

func TestFECUnit_Sec3_ZeroHeapAllocationsAndSectorBurstHeal(t *testing.T) {
	plain := make([]byte, engine.MaxFECPlainBytes)
	for i := range plain {
		plain[i] = byte((i*17 + 3) & 0xFF)
	}
	enc := make([]byte, len(plain)*2)
	dec := make([]byte, len(plain))

	allocs := testing.AllocsPerRun(200, func() {
		_, _ = engine.EncodeFECInto(enc, plain)
		_, _, _ = engine.DecodeFECFrom(dec, enc)
	})
	if allocs != 0 {
		t.Fatalf("[SEC 3] ALLOC_ZERO_HEAP violated: got %.2f allocs/op", allocs)
	}

	// Flip 1 bit in EVERY SINGLE ONE of the 480 encoded bytes (480 simultaneous bit flips!)
	nEnc, _ := engine.EncodeFECInto(enc, plain)
	for i := 0; i < nEnc; i++ {
		enc[i] ^= 1 << uint(i%8)
	}
	nDec, healed, err := engine.DecodeFECFrom(dec, enc[:nEnc])
	if err != nil || nDec != len(plain) || healed != nEnc || !bytes.Equal(dec[:nDec], plain) {
		t.Fatalf("[SEC 3] 480-bit burst self-heal failed: healed=%d err=%v", healed, err)
	}
}

func TestFECUnit_Sec4_PreflightAndDumpCard(t *testing.T) {
	common.AssertAuditPass(t, engine.PreflightFEC(), "PreflightFEC")
	if !engine.IsFECReady() {
		t.Fatal("[SEC 4] IsFECReady returned false")
	}
	if !strings.Contains(engine.ExecuteFECDumpCard(), "Hamming(8,4)") {
		t.Fatal("[SEC 4] FEC dump card missing Hamming(8,4) marker")
	}
}

// ==============================================================================
// BLOCK 4 & 5: AUDIT & CLOSING
// ==============================================================================
// End of tests/unit/engine/fec_test.go
