// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          engine/trit.go
// @id            cog.engine.trit
// @layer         Layer 1: Pure Engine Core (Bit <-> Balanced-Trit <-> Bit Codec)
// @archetype     pure_engine | ternary_wire_bridge
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Sovereign Adaptive Networking Substrate
// @role          Zero-Heap Bit <-> Balanced-Trit <-> Bit OmniCode Submodule Bridge
// @maturity      Rev 5 Baseline | Wave 15 FNP/1 & OmniCode Ternary Bridge
// @chassis       ALLOC_ZERO_HEAP | Caller-owned slices, 0.00 allocs/op hot path
// @scale         O(N) deterministic linear packing/unpacking (5 trits/byte; 6 trits/byte)
// @seams         in: data (QuadState) | exports: Trit, Tryte5, Bit<->Trit codecs
// @invariants    1. Downward Dependency Law: ZERO net, os, or external imports.
//                2. 5 balanced trits = 3^5 = 243 states -> bytes 0x00..0xF2;
//                   bytes 0xF3..0xFF (13 values) reserved for hardware/NULL sentinels.
//                3. Lossless binary bridge: 1 byte (0..255) <-> 6 balanced trits
//                   (3^6 = 729 states, centered at +128) with exact bit-for-bit identity.
//                4. Zero heap allocations across Pack/Unpack/Encode/Decode.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Pure Layer 1 balanced-ternary codec preparing FaithNet to plug directly
//   into OmniCode as a native submodule. Bridges today's binary physical wires
//   (Wi-Fi, Ethernet, UART, PCM audio) with OmniCode's native balanced ternary
//   truth space (T3 = {-1, 0, +1} and T4 = {-1, 0, +1, NULL}) losslessly.
//
//   ### For Architects, Operators & Non-Programmers:
//   Modern computers speak in binary switches (0 and 1), while OmniCode thinks
//   in balanced three-way truth (-1 Contradict, 0 Await/Neutral, +1 Aligned).
//   This translator packs and unpacks between binary bytes and balanced trits
//   without losing a single signal or wasting memory.
//
//   ### For Systems Programmers & Engine Architects:
//   Provides O(1) PackTryte5/UnpackTryte5 (base-3 balanced radix [-121..+121]
//   mapped to [0..242] with [243..255] reserved sentinels), slice-oriented
//   EncodeTritsInto/DecodeTritsFrom, lossless BytesToBalancedTrits /
//   BalancedTritsToBytes (6 trits/byte), and T4 QuadState bridge helpers.
// ==============================================================================

package engine

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Module Anchor, Bounds, Public Ternary Ontology & Pure O(1) Predicates )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 1: Pure Engine Core
// Module:        cog.engine.trit
// Mission:       Zero-alloc binary <-> balanced-ternary wire bridge.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"errors"
	"fmt"
)

// --- Tier 2: External Library Dependencies (None — 100% Go Stdlib Invariant) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---
import "github.com/CreativeWorkzStudio/FaithNet/data"

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Pre-Conditions, Ternary Bounds & Error Sentinels
// ------------------------------------------------------------------------------

const (
	// TritsPerTryte5 is the number of balanced trits packed into a single byte
	// via PackTryte5 (3^5 = 243 states -> bytes 0x00..0xF2).
	TritsPerTryte5 = 5

	// Tryte5StateCount is 3^5 = 243 distinct balanced-ternary states.
	Tryte5StateCount = 243

	// Tryte5MaxEncodedByte is the highest valid encoded tryte byte (242 = 0xF2).
	Tryte5MaxEncodedByte byte = 242

	// Tryte5MinValue is the minimum signed integer value of 5 balanced trits (-121).
	Tryte5MinValue = -121

	// Tryte5MaxValue is the maximum signed integer value of 5 balanced trits (+121).
	Tryte5MaxValue = 121

	// TritNullSentinel is the reserved wire byte representing T4 NULL / HALT.
	TritNullSentinel byte = 0xFF

	// TritsPerBinaryByte is the exact number of balanced trits required to
	// represent any arbitrary 8-bit binary byte (0..255) losslessly (3^6 = 729).
	TritsPerBinaryByte = 6
)

var (
	// ErrTritInvalid signals a Trit value outside {-1, 0, +1}.
	ErrTritInvalid = errors.New("trit: value outside balanced ternary {-1, 0, +1}")
	// ErrTritBufferTooSmall signals insufficient caller destination capacity.
	ErrTritBufferTooSmall = errors.New("trit: destination buffer too small")
	// ErrTritBadLength signals input length not aligned to radix framing bounds.
	ErrTritBadLength = errors.New("trit: input length misaligned with framing radix")
	// ErrTritSentinelByte signals a reserved sentinel byte (0xF3..0xFF) in tryte stream.
	ErrTritSentinelByte = errors.New("trit: encountered reserved sentinel byte (0xF3..0xFF)")
	// ErrTritByteRange signals 6 balanced trits decoded outside [0..255] byte range.
	ErrTritByteRange = errors.New("trit: balanced trit word outside 0..255 byte domain")
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.2: Operational Logical Anchor
// ------------------------------------------------------------------------------

const (
	// ModuleIDTritCodec is the movement-resilient logical ID for cog.engine.trit.
	ModuleIDTritCodec = "cog.engine.trit"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.3: Public Ternary Ontology
// ------------------------------------------------------------------------------

// Trit represents a single balanced ternary digit in {-1, 0, +1}.
type Trit int8

const (
	// TritNeg represents -1 (Contradict / Negative polarity).
	TritNeg Trit = -1
	// TritZero represents 0 (Neutral / Await / Zero polarity).
	TritZero Trit = 0
	// TritPos represents +1 (Aligned / Positive polarity).
	TritPos Trit = +1
)

// Tryte5 is a fixed 5-trit balanced-ternary word (least-significant trit at
// index 0) packing cleanly into a single byte (0x00..0xF2) with 13 reserved
// hardware/sentinel bytes remaining (0xF3..0xFF).
type Tryte5 [TritsPerTryte5]Trit

// ------------------------------------------------------------------------------
// Sub-Assembly 2.4: Pure O(1) Predicates
// ------------------------------------------------------------------------------

// IsValid reports whether t is a canonical balanced trit (-1, 0, or +1).
func (t Trit) IsValid() bool {
	return t == TritNeg || t == TritZero || t == TritPos
}

// String renders a Trit using canonical balanced-ternary notation ("-", "0", "+").
func (t Trit) String() string {
	switch t {
	case TritNeg:
		return "-"
	case TritZero:
		return "0"
	case TritPos:
		return "+"
	default:
		return "?"
	}
}

// IsValid reports whether all 5 trits in the Tryte5 word are canonical.
func (w Tryte5) IsValid() bool {
	return w[0].IsValid() && w[1].IsValid() && w[2].IsValid() && w[3].IsValid() && w[4].IsValid()
}

// IsTritSentinelByte reports whether b belongs to the 13 reserved sentinel bytes
// (0xF3..0xFF) outside the 243-state Tryte5 space.
func IsTritSentinelByte(b byte) bool {
	return b > Tryte5MaxEncodedByte
}

// ==============================================================================
// BLOCK 3: BODY BLOCK
// ( Zero-Heap Bit <-> Balanced-Trit <-> Bit & T4 QuadState Conversion Cogs )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 1: Pure Engine Core
// Module:        cog.engine.trit
// Mission:       Deterministic zero-alloc ternary packing and binary bridging.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: Layer 0 QuadState (T4) <-> Balanced Trit Bridge
// ------------------------------------------------------------------------------

// QuadStateToTrit projects a Layer 0 QuadState onto a balanced Trit. Returns
// (TritZero, false) when q is QuadNull or invalid, honoring T4 NULL isolation.
//
// ### For Architects, Operators & Non-Programmers:
// Converts FaithNet's four-state truth label into a balanced trit (-1, 0, +1),
// flagging NULL explicitly so missing/halted state is never confused with zero.
//
// ### For Systems Programmers & Engine Architects:
// Branch-only O(1) total mapping from data.QuadState to (Trit, bool).
func QuadStateToTrit(q data.QuadState) (Trit, bool) {
	switch q {
	case data.QuadNegative:
		return TritNeg, true
	case data.QuadAwait:
		return TritZero, true
	case data.QuadPositive:
		return TritPos, true
	default:
		return TritZero, false
	}
}

// TritToQuadState lifts a canonical balanced Trit into Layer 0 QuadState.
// Any invalid non-ternary input maps defensively to QuadNull.
//
// ### For Architects, Operators & Non-Programmers:
// Translates a balanced ternary digit back into FaithNet's wire truth state.
//
// ### For Systems Programmers & Engine Architects:
// Total O(1) mapping returning data.QuadNull for out-of-domain int8 values.
func TritToQuadState(t Trit) data.QuadState {
	switch t {
	case TritNeg:
		return data.QuadNegative
	case TritZero:
		return data.QuadAwait
	case TritPos:
		return data.QuadPositive
	default:
		return data.QuadNull
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: 5-Trit Tryte5 <-> Single Byte Packing (3^5 = 243 States)
// ------------------------------------------------------------------------------

// PackTryte5 encodes a 5-trit balanced-ternary word into a single byte in
// [0x00..0xF2] using little-endian balanced radix-3 weighted sum shifted by +121.
//
// ### For Architects, Operators & Non-Programmers:
// Squeezes 5 balanced trits cleanly into 1 regular computer byte, leaving 13
// special values free for sync and safety signals.
//
// ### For Systems Programmers & Engine Architects:
// Pure O(1), 0-alloc arithmetic: val = sum(w[i] * 3^i) in [-121..+121],
// encoded = byte(val + 121). Returns (TritNullSentinel, false) on invalid trit.
func PackTryte5(w Tryte5) (byte, bool) {
	if !w.IsValid() {
		return TritNullSentinel, false
	}
	val := int(w[0]) +
		int(w[1])*3 +
		int(w[2])*9 +
		int(w[3])*27 +
		int(w[4])*81
	return byte(val + Tryte5MaxValue), true
}

// UnpackTryte5 decodes a byte in [0x00..0xF2] back into its exact Tryte5 word.
// Reserved sentinel bytes (0xF3..0xFF) return (Tryte5{}, false).
//
// ### For Architects, Operators & Non-Programmers:
// Restores 5 balanced trits from a single packed wire byte while rejecting
// reserved sentinel markers cleanly.
//
// ### For Systems Programmers & Engine Architects:
// Exact inverse of PackTryte5 using balanced radix-3 modulo normalization.
func UnpackTryte5(b byte) (Tryte5, bool) {
	if b > Tryte5MaxEncodedByte {
		return Tryte5{}, false
	}
	var w Tryte5
	v := int(b) - Tryte5MaxValue
	for i := 0; i < TritsPerTryte5; i++ {
		rem := v % 3
		v /= 3
		if rem == 2 {
			rem = -1
			v++
		} else if rem == -2 {
			rem = 1
			v--
		}
		w[i] = Trit(rem)
	}
	return w, true
}

// EncodeTritsInto packs a slice of balanced trits (whose length must be a
// multiple of 5) into caller-owned dst at 5 trits per byte (0 heap allocs).
//
// ### For Architects, Operators & Non-Programmers:
// Packs a stream of OmniCode trits into compact binary bytes ready to cross
// Wi-Fi, Ethernet, copper serial wire, or audio tones.
//
// ### For Systems Programmers & Engine Architects:
// Zero-allocation linear encoder validating divisibility by TritsPerTryte5.
func EncodeTritsInto(dst []byte, trits []Trit) (int, error) {
	if len(trits)%TritsPerTryte5 != 0 {
		return 0, ErrTritBadLength
	}
	needed := len(trits) / TritsPerTryte5
	if len(dst) < needed {
		return 0, ErrTritBufferTooSmall
	}
	for i := 0; i < needed; i++ {
		base := i * TritsPerTryte5
		w := Tryte5{trits[base], trits[base+1], trits[base+2], trits[base+3], trits[base+4]}
		b, ok := PackTryte5(w)
		if !ok {
			return 0, ErrTritInvalid
		}
		dst[i] = b
	}
	return needed, nil
}

// DecodeTritsFrom unpacks Tryte5-encoded wire bytes into caller-owned dst
// (5 trits per byte, 0 heap allocs).
//
// ### For Architects, Operators & Non-Programmers:
// Unpacks incoming wire bytes back into native OmniCode balanced trits, halting
// immediately if a reserved sentinel byte appears mid-stream.
//
// ### For Systems Programmers & Engine Architects:
// Zero-allocation linear decoder rejecting any byte > Tryte5MaxEncodedByte.
func DecodeTritsFrom(dst []Trit, src []byte) (int, error) {
	needed := len(src) * TritsPerTryte5
	if len(dst) < needed {
		return 0, ErrTritBufferTooSmall
	}
	for i, b := range src {
		w, ok := UnpackTryte5(b)
		if !ok {
			return 0, ErrTritSentinelByte
		}
		base := i * TritsPerTryte5
		dst[base] = w[0]
		dst[base+1] = w[1]
		dst[base+2] = w[2]
		dst[base+3] = w[3]
		dst[base+4] = w[4]
	}
	return needed, nil
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: Lossless Binary <-> Balanced-Trit Bridge (1 Byte <-> 6 Trits)
// ------------------------------------------------------------------------------

// BytesToBalancedTrits converts arbitrary binary bytes (0..255) into a lossless
// balanced-ternary stream at 6 trits per byte (3^6 = 729 > 256), with 0 allocs.
//
// ### For Architects, Operators & Non-Programmers:
// Turns any regular computer data (text, JSON, binary payloads) into balanced
// trits so OmniCode ternary pipelines can process legacy data natively.
//
// ### For Systems Programmers & Engine Architects:
// Centers each unsigned byte b in [0..255] via signed value v = int(b) - 128
// (range [-128..+127], strictly inside [-364..+364] of 6 balanced trits).
func BytesToBalancedTrits(dst []Trit, src []byte) (int, error) {
	needed := len(src) * TritsPerBinaryByte
	if len(dst) < needed {
		return 0, ErrTritBufferTooSmall
	}
	for i, b := range src {
		v := int(b) - 128
		base := i * TritsPerBinaryByte
		for k := 0; k < TritsPerBinaryByte; k++ {
			rem := v % 3
			v /= 3
			if rem == 2 {
				rem = -1
				v++
			} else if rem == -2 {
				rem = 1
				v--
			}
			dst[base+k] = Trit(rem)
		}
	}
	return needed, nil
}

// BalancedTritsToBytes converts a 6-trit-per-byte balanced-ternary stream back
// into exact original binary bytes (0..255) with 0 heap allocations.
//
// ### For Architects, Operators & Non-Programmers:
// Restores balanced trits back into exact original binary bytes for legacy
// hardware, browsers, and network sockets.
//
// ### For Systems Programmers & Engine Architects:
// Exact inverse of BytesToBalancedTrits; rejects non-canonical trits or decoded
// values outside [0..255].
func BalancedTritsToBytes(dst []byte, trits []Trit) (int, error) {
	if len(trits)%TritsPerBinaryByte != 0 {
		return 0, ErrTritBadLength
	}
	needed := len(trits) / TritsPerBinaryByte
	if len(dst) < needed {
		return 0, ErrTritBufferTooSmall
	}
	for i := 0; i < needed; i++ {
		base := i * TritsPerBinaryByte
		val := 0
		weight := 1
		for k := 0; k < TritsPerBinaryByte; k++ {
			t := trits[base+k]
			if !t.IsValid() {
				return 0, ErrTritInvalid
			}
			val += int(t) * weight
			weight *= 3
		}
		shifted := val + 128
		if shifted < 0 || shifted > 255 {
			return 0, ErrTritByteRange
		}
		dst[i] = byte(shifted)
	}
	return needed, nil
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Contract Verification Cogs & Pre-Flight Validation Tripwires )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 1: Pure Engine Core
// Module:        cog.engine.trit
// Mission:       Exhaustive 243-state & 256-byte mathematical self-verification.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Machine Self-Audit (4-Section Pre-Flight Validation)
// ------------------------------------------------------------------------------

// PreflightTritCodec runs a 4-section pre-flight audit proving:
// (1) exhaustive 243-state Tryte5 pack/unpack bijection,
// (2) reserved sentinel (0xF3..0xFF) and invalid trit rejection,
// (3) exhaustive 256-byte bit -> balanced-trit -> bit lossless identity, and
// (4) T4 QuadState <-> Trit emergence parity.
//
// @triggers tests/unit/engine/trit_test.go
func PreflightTritCodec() data.AuditResult {
	const check = "trit_codec_preflight_4section"

	// Micro Check 1 (Always Pass): Exhaustive 243-state Tryte5 roundtrip.
	for raw := 0; raw < Tryte5StateCount; raw++ {
		w, ok := UnpackTryte5(byte(raw))
		if !ok || !w.IsValid() {
			return data.AuditFail(data.AuditSeverityFatal, ModuleIDTritCodec, check,
				fmt.Sprintf("Sec 1 UnpackTryte5 failed for valid byte %d.", raw))
		}
		back, ok := PackTryte5(w)
		if !ok || back != byte(raw) {
			return data.AuditFail(data.AuditSeverityFatal, ModuleIDTritCodec, check,
				fmt.Sprintf("Sec 1 PackTryte5 mismatch: got %d want %d.", back, raw))
		}
	}

	// Micro Check 2 (Never Fail): Sentinel bytes and invalid trits rejected.
	for s := int(Tryte5MaxEncodedByte) + 1; s <= 255; s++ {
		if _, ok := UnpackTryte5(byte(s)); ok {
			return data.AuditFail(data.AuditSeverityCritical, ModuleIDTritCodec, check,
				fmt.Sprintf("Sec 2 sentinel byte 0x%02X bypassed UnpackTryte5.", s))
		}
	}
	if _, ok := PackTryte5(Tryte5{Trit(2), 0, 0, 0, 0}); ok {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDTritCodec, check,
			"Sec 2 invalid trit +2 bypassed PackTryte5.")
	}

	// Micro Check 3 (Invariants & Limits): All 256 binary byte values survive
	// bit -> 6 balanced trits -> bit losslessly.
	var allBytes [256]byte
	for i := 0; i < 256; i++ {
		allBytes[i] = byte(i)
	}
	var tritBuf [256 * TritsPerBinaryByte]Trit
	var outBytes [256]byte
	nT, err := BytesToBalancedTrits(tritBuf[:], allBytes[:])
	if err != nil || nT != len(tritBuf) {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDTritCodec, check,
			fmt.Sprintf("Sec 3 BytesToBalancedTrits failed: n=%d err=%v.", nT, err))
	}
	nB, err := BalancedTritsToBytes(outBytes[:], tritBuf[:])
	if err != nil || nB != 256 {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDTritCodec, check,
			fmt.Sprintf("Sec 3 BalancedTritsToBytes failed: n=%d err=%v.", nB, err))
	}
	for i := 0; i < 256; i++ {
		if outBytes[i] != allBytes[i] {
			return data.AuditFail(data.AuditSeverityFatal, ModuleIDTritCodec, check,
				fmt.Sprintf("Sec 3 bit->trit->bit mismatch at byte %d: got %d.", i, outBytes[i]))
		}
	}

	// Micro Check 4 (Emergence): QuadState T4 bridge preserves NULL isolation.
	if _, ok := QuadStateToTrit(data.QuadNull); ok {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDTritCodec, check,
			"Sec 4 QuadNull must not project to a valid balanced trit.")
	}
	if TritToQuadState(TritPos) != data.QuadPositive ||
		TritToQuadState(TritZero) != data.QuadAwait ||
		TritToQuadState(TritNeg) != data.QuadNegative {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDTritCodec, check,
			"Sec 4 TritToQuadState T4 truth mapping mismatch.")
	}

	return data.AuditPass(ModuleIDTritCodec, check,
		"Bit<->Trit<->Bit codec passed all 4 micro gates (243 trytes + 256 bytes + T4).")
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Epilogue Readiness Handshake & 60-Column Diagnostic Telemetry Card )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 1: Pure Engine Core
// Module:        cog.engine.trit
// Mission:       Expose readiness handshake and 60-column telemetry card.
// ==============================================================================

// IsTritCodecReady verifies Layer 1 balanced-ternary codec readiness.
func IsTritCodecReady() bool {
	return PreflightTritCodec().IsPass()
}

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: 60-Column Telemetry Card
// ------------------------------------------------------------------------------

// ExecuteTritCodecDumpCard renders a 60-column diagnostic card for cog.engine.trit.
func ExecuteTritCodecDumpCard() string {
	return fmt.Sprintf(
		"+------------------------------------------------------------+\n"+
			"| FAITHNET BALANCED TERNARY CODEC CARD (OMNICODE BRIDGE)     |\n"+
			"+------------------------------------------------------------+\n"+
			"| Module:     %-46s |\n"+
			"| Tryte5:     5 trits/byte (3^5=243 states, 0x00..0xF2)      |\n"+
			"| Sentinels:  13 reserved bytes (0xF3..0xFF, NULL=0xFF)      |\n"+
			"| Bit<->Trit: 1 byte (0..255) <-> 6 balanced trits (0 alloc) |\n"+
			"+------------------------------------------------------------+",
		ModuleIDTritCodec,
	)
}

// End of engine/trit.go
