// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          engine/fec.go
// @id            cog.engine.fec
// @layer         Layer 1: Pure Engine Core (OSI L2 Forward Error Correction)
// @archetype     pure_engine | secded_hamming_fec
// @version       0.2.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Sovereign Adaptive Networking Substrate (FNP/1)
// @role          Zero-Heap SECDED Hamming(8,4) Forward Error Correction Codec
// @maturity      Rev 5 Baseline | Wave 20 OSI L2 Blind-Spot Closure
// @chassis       ALLOC_ZERO_HEAP | Static 16-entry LUT, 0.00 allocs/op hot path
// @scale         O(N) nibble encode/decode (2x wire expansion); heals 1 bit/nibble
// @seams         in: data (AuditResult) | exports: EncodeFECInto, DecodeFECFrom
// @invariants    1. Downward Dependency Law: ZERO net, os, or external imports.
//                2. Corrects ANY single-bit error per 8-bit codeword (nibble).
//                3. Detects ANY double-bit error per codeword cleanly (ErrFECUncorrectable).
//                4. Zero heap allocations across EncodeFECInto / DecodeFECFrom.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Closes OSI Layer 2 Blind Spot #1: provides lightweight, zero-heap
//   Single-Error-Correcting / Double-Error-Detecting (SECDED) Hamming(8,4)
//   Forward Error Correction for noisy Rung 1 acoustic/walkie-talkie/copper
//   channels so static pops self-heal without forcing full frame retransmission.
//
//   ### For Architects, Operators & Non-Programmers:
//   Adds mathematical self-healing armor to radio and audio messages so a crackle
//   of static flipping a bit in mid-air gets automatically fixed on arrival.
//
//   ### For Systems Programmers & Engine Architects:
//   Systematic Hamming(8,4) over each 4-bit nibble using O(1) parity syndrome
//   decoding. Returns exact corrected-bit telemetry without heap allocations.
// ==============================================================================

package engine

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Bounds, Error Sentinels, Module Anchor & Precomputed Hamming(8,4) Table )
// ==============================================================================

import (
	"errors"
	"fmt"

	"github.com/CreativeWorkzStudio/FaithNet/data"
)

const (
	// ModuleIDFEC is the canonical movement-resilient ID for cog.engine.fec.
	ModuleIDFEC = "cog.engine.fec"

	// MaxFECPlainBytes is the maximum raw bytes that fit into one 480B sector
	// after 2:1 SECDED Hamming(8,4) encoding (240 raw -> 480 encoded bytes).
	MaxFECPlainBytes = data.FloorMaxPayload / 2
)

var (
	// ErrFECBufferTooSmall signals insufficient caller destination slice capacity.
	ErrFECBufferTooSmall = errors.New("fec: destination buffer too small")
	// ErrFECBadLength signals an odd-length encoded stream (must be 2 bytes/raw byte).
	ErrFECBadLength = errors.New("fec: encoded input length must be even")
	// ErrFECUncorrectable signals a detected double-bit error inside a codeword.
	ErrFECUncorrectable = errors.New("fec: uncorrectable double-bit error detected")
)

// hamming84LUT maps every 4-bit nibble (0..15) to its exact SECDED Hamming(8,4)
// codeword: bits [b7 b6 b5 b4 b3 b2 b1 b0] = [p0 d3 d2 d1 p3 d0 p2 p1]
// where:
//   p1 = d0 ^ d1 ^ d3  (covers positions 1,3,5,7 -> bits 0,2,4,6)
//   p2 = d0 ^ d2 ^ d3  (covers positions 2,3,6,7 -> bits 1,2,5,6)
//   p3 = d1 ^ d2 ^ d3  (covers positions 4,5,6,7 -> bits 3,4,5,6)
//   p0 = overall even parity across all 8 bits (bit 7).
var hamming84LUT = [16]byte{
	0x00, 0x87, 0x99, 0x1E, 0xAA, 0x2D, 0x33, 0xB4,
	0x4B, 0xCC, 0xD2, 0x55, 0xE1, 0x66, 0x78, 0xFF,
}

func init() {
	// Derive hamming84LUT deterministically at startup to guarantee exact bit-law parity.
	for nib := 0; nib < 16; nib++ {
		d0 := byte((nib >> 0) & 1)
		d1 := byte((nib >> 1) & 1)
		d2 := byte((nib >> 2) & 1)
		d3 := byte((nib >> 3) & 1)

		p1 := d0 ^ d1 ^ d3
		p2 := d0 ^ d2 ^ d3
		p3 := d1 ^ d2 ^ d3

		code7 := (p1 << 0) | (p2 << 1) | (d0 << 2) | (p3 << 3) | (d1 << 4) | (d2 << 5) | (d3 << 6)
		overall := (p1 ^ p2 ^ d0 ^ p3 ^ d1 ^ d2 ^ d3) & 1
		hamming84LUT[nib] = code7 | (overall << 7)
	}
}

// ==============================================================================
// BLOCK 3: BODY BLOCK
// ( Zero-Heap SECDED Hamming(8,4) Nibble & Slice Encode / Decode Cogs )
// ==============================================================================

// EncodeNibbleFEC encodes a 4-bit nibble (low 4 bits) into an 8-bit SECDED
// Hamming(8,4) codeword in O(1) time with 0 heap allocations.
func EncodeNibbleFEC(nibble byte) byte {
	return hamming84LUT[nibble&0x0F]
}

// DecodeNibbleFEC decodes an 8-bit SECDED Hamming(8,4) codeword back into its
// 4-bit nibble. Returns (nibble, correctedBit, err):
//   - syndrome == 0 && overall == 0 -> clean (0 errors)
//   - overall == 1                  -> single-bit error healed (corrected=true)
//   - syndrome != 0 && overall == 0 -> double-bit error detected (ErrFECUncorrectable)
func DecodeNibbleFEC(cw byte) (byte, bool, error) {
	b1 := (cw >> 0) & 1
	b2 := (cw >> 1) & 1
	b3 := (cw >> 2) & 1
	b4 := (cw >> 3) & 1
	b5 := (cw >> 4) & 1
	b6 := (cw >> 5) & 1
	b7 := (cw >> 6) & 1
	b8 := (cw >> 7) & 1

	s1 := b1 ^ b3 ^ b5 ^ b7
	s2 := b2 ^ b3 ^ b6 ^ b7
	s3 := b4 ^ b5 ^ b6 ^ b7
	syndrome := s1 | (s2 << 1) | (s3 << 2)
	overall := b1 ^ b2 ^ b3 ^ b4 ^ b5 ^ b6 ^ b7 ^ b8

	corrected := false
	if overall == 1 {
		// Single-bit flip: syndrome 1..7 locates bit (syndrome-1); syndrome 0 means bit 7 flipped.
		if syndrome > 0 {
			cw ^= 1 << (syndrome - 1)
		} else {
			cw ^= 1 << 7
		}
		corrected = true
	} else if syndrome != 0 {
		return 0, false, ErrFECUncorrectable
	}

	d0 := (cw >> 2) & 1
	d1 := (cw >> 4) & 1
	d2 := (cw >> 5) & 1
	d3 := (cw >> 6) & 1
	return d0 | (d1 << 1) | (d2 << 2) | (d3 << 3), corrected, nil
}

// EncodeFECInto encodes src into caller-owned dst (len(dst) >= 2*len(src))
// using zero-heap SECDED Hamming(8,4) coding.
func EncodeFECInto(dst, src []byte) (int, error) {
	needed := len(src) * 2
	if len(dst) < needed {
		return 0, ErrFECBufferTooSmall
	}
	for i, b := range src {
		dst[i*2] = hamming84LUT[(b>>4)&0x0F]
		dst[i*2+1] = hamming84LUT[b&0x0F]
	}
	return needed, nil
}

// DecodeFECFrom decodes Hamming(8,4)-protected src into caller-owned dst,
// returning (bytesWritten, bitsCorrected, err). Automatically heals up to 1
// bit-flip per nibble and rejects double-bit codeword corruption cleanly.
func DecodeFECFrom(dst, src []byte) (int, int, error) {
	if len(src)%2 != 0 {
		return 0, 0, ErrFECBadLength
	}
	needed := len(src) / 2
	if len(dst) < needed {
		return 0, 0, ErrFECBufferTooSmall
	}
	correctedTotal := 0
	for i := 0; i < needed; i++ {
		hi, cHi, err := DecodeNibbleFEC(src[i*2])
		if err != nil {
			return 0, correctedTotal, err
		}
		lo, cLo, err := DecodeNibbleFEC(src[i*2+1])
		if err != nil {
			return 0, correctedTotal, err
		}
		if cHi {
			correctedTotal++
		}
		if cLo {
			correctedTotal++
		}
		dst[i] = (hi << 4) | lo
	}
	return needed, correctedTotal, nil
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Exhaustive Single-Bit Self-Healing & Double-Bit Detection Pre-Flight )
// ==============================================================================

// PreflightFEC executes a 4-section pre-flight gate proving:
// (1) exhaustive 16-nibble clean roundtrip,
// (2) exhaustive single-bit flip self-healing across all 16*8 = 128 single-bit faults,
// (3) double-bit flip rejection (ErrFECUncorrectable), and
// (4) multi-byte slice burst recovery under heavy radio static.
//
// @triggers tests/unit/engine/fec_test.go
func PreflightFEC() data.AuditResult {
	const check = "fec_secded_preflight_4section"

	// Micro Check 1 & 2: Exhaustive 16 nibbles x 8 bit positions = 128 single-bit heal proofs!
	for nib := byte(0); nib < 16; nib++ {
		cw := EncodeNibbleFEC(nib)
		got, corr, err := DecodeNibbleFEC(cw)
		if err != nil || corr || got != nib {
			return data.AuditFail(data.AuditSeverityFatal, ModuleIDFEC, check,
				fmt.Sprintf("Sec 1 clean decode failed for nibble %d.", nib))
		}
		for bit := uint(0); bit < 8; bit++ {
			flipped := cw ^ (1 << bit)
			healed, wasCorr, err := DecodeNibbleFEC(flipped)
			if err != nil || !wasCorr || healed != nib {
				return data.AuditFail(data.AuditSeverityFatal, ModuleIDFEC, check,
					fmt.Sprintf("Sec 2 single-bit heal failed nib=%d bit=%d.", nib, bit))
			}
		}
	}

	// Micro Check 3: Double-bit flip inside one codeword is detected cleanly.
	cw := EncodeNibbleFEC(0x0A) ^ 0x03
	if _, _, err := DecodeNibbleFEC(cw); !errors.Is(err, ErrFECUncorrectable) {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDFEC, check,
			"Sec 3 double-bit corruption bypassed SECDED detection.")
	}

	// Micro Check 4: Full message burst with 1 bit-flip per codeword heals 100%.
	msg := []byte("FAITHNET-RUNG1-FEC-OK")
	var enc [64]byte
	var dec [32]byte
	nEnc, err := EncodeFECInto(enc[:], msg)
	if err != nil {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDFEC, check, err.Error())
	}
	for i := 0; i < nEnc; i++ {
		enc[i] ^= 1 << uint(i%8)
	}
	nDec, healedBits, err := DecodeFECFrom(dec[:], enc[:nEnc])
	if err != nil || nDec != len(msg) || healedBits != nEnc || string(dec[:nDec]) != string(msg) {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDFEC, check,
			fmt.Sprintf("Sec 4 burst heal mismatch: healed=%d err=%v.", healedBits, err))
	}

	return data.AuditPass(ModuleIDFEC, check,
		"SECDED Hamming(8,4) FEC passed all 4 micro gates (128/128 single-bit heals + 2-bit detect).")
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ==============================================================================

// IsFECReady reports whether Layer 1 SECDED Hamming(8,4) FEC is verified ready.
func IsFECReady() bool {
	return PreflightFEC().IsPass()
}

// ExecuteFECDumpCard renders a 60-column telemetry card for cog.engine.fec.
func ExecuteFECDumpCard() string {
	return fmt.Sprintf(
		"+------------------------------------------------------------+\n"+
			"| FAITHNET RUNG 1 FORWARD ERROR CORRECTION CARD (OSI L2)     |\n"+
			"+------------------------------------------------------------+\n"+
			"| Module:     %-46s |\n"+
			"| Scheme:     SECDED Hamming(8,4) (2:1, 0.00 allocs/op)      |\n"+
			"| Capability: Self-heals 1 bit/nibble, detects 2-bit bursts  |\n"+
			"| Sector Cap: 240B raw -> 480B encoded inside 512B sector    |\n"+
			"+------------------------------------------------------------+",
		ModuleIDFEC,
	)
}

// End of engine/fec.go
