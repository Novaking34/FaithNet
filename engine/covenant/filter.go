// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          engine/covenant/filter.go
// @id            cog.covenant.airlock
// @layer         Layer 1: Pure Functional Engine Substrate (Covenant Airlock)
// @archetype     operational | stateless
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Genesis 3 Airlock & T4 Epistemic Filtration Engine
// @role          Runtime Engine: Deterministic Ingress Evaluator & Core Guard
// @maturity      Rev 5 Baseline | Wave 1 Seed
// @chassis       zero-heap | Pure slice inspection over bounded frame buffers
// @scale         O(N) linear scan over bounded ingress bytes (N <= FloorMaxPayload)
// @seams         in: data.Frame, []byte | uses: bytes, fmt | exports: Filter
// @invariants    1. Stratum Baseline: ALLOC_ZERO_HEAP, Downward Dependency Law,
//                   Universal 5-Block Lifecycle Protocol (Rev 5).
//                2. Untrusted external origin (FlagExternalOrigin) NEVER receives
//                   automatic QuadPositive (+1); it enters at most as 0:AWAIT.
//                3. Malformed structure triggers NULL:HALT; hostile signatures
//                   or forged verification flags trigger -1:REJECT.
//                4. Core CPI-SI weights remain insulated from raw web noise.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Implements the Genesis 3 Airlock for FaithNet. Rather than ingesting the
//   chaotic, fallen commercial web indiscriminately into core state, every
//   external payload and wire frame passes through deterministic structural,
//   provenance, and hygiene invariants anchored in OmniCode's T4 truth space.
//
//   ### For Architects, Operators & Non-Programmers:
//   Acts like an immune system and airlock door at the edge of the node.
//   Friendly local peers pass cleanly (+1), clean outside facts wait politely
//   in quarantine for inspection (0:AWAIT), and poisoned or deceptive streams
//   are shut out at the gate (-1:REJECT or NULL:HALT).
//
//   ### For Systems Programmers & Engine Architects:
//   - Controlled Memory Chassis: Immutable signature table and caller-owned
//     byte slice scanning with zero regex or lowercase-copy heap allocations.
//   - Probe & Config Wiring: Block 4 verifies sovereign pass, external await,
//     forged claim rejection, and hostile payload interception.
// ==============================================================================

package covenant

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Bounds, Signature Table, Verdict Ontology, Predicates & Monitored Support )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 1: Pure Functional Engine Substrate (Covenant Airlock)
// Module:        cog.covenant.airlock
// Mission:       Declare hostile signature tables, Verdict/Filter ontology,
//                O(1) verdict predicates, and monitored ASCII-fold support.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"bytes"
	"fmt"

	"github.com/CreativeWorkzStudio/FaithNet/data" // --- Tier 2: External Library Dependencies (None — 100% Go Stdlib Invariant) ---
)

// --- Tier 3: Internal FaithNet Foundation Packages ---

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Pre-Conditions, Physical Bounds & Hostile Signature Table
// ------------------------------------------------------------------------------
// Immutable byte ceilings and static hostile web/command injection signatures.
// ------------------------------------------------------------------------------

const (
	// DefaultAirlockMaxBytes caps external payloads to the 512B Floor MTU.
	DefaultAirlockMaxBytes = data.FloorMaxPayload
)

var hostileSignatures = [...][]byte{
	[]byte("<script"),
	[]byte("javascript:"),
	[]byte("onerror="),
	[]byte("onload="),
	[]byte("eval("),
	[]byte("rm -rf /"),
	[]byte("DROP TABLE"),
	[]byte("UNION SELECT"),
	[]byte("<?php"),
}

// ------------------------------------------------------------------------------
// Sub-Assembly 2.2: Operational Logical Anchor
// ------------------------------------------------------------------------------
// Movement-resilient identity shared across audits and diagnostic cards.
// ------------------------------------------------------------------------------

const (
	// ModuleIDAirlock is the canonical logical ID for the Genesis 3 Airlock.
	ModuleIDAirlock = "cog.covenant.airlock"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.3: Public Tagged Ontology (ReasonCode, Verdict & Filter)
// ------------------------------------------------------------------------------
// Deterministic T4 verdict descriptors and airlock policy configuration.
// ------------------------------------------------------------------------------

// ReasonCode identifies why the Genesis 3 Airlock issued a specific verdict.
type ReasonCode uint8

const (
	ReasonAlignedSovereign ReasonCode = iota
	ReasonEmpiricalExternalAwait
	ReasonMalformedStructure
	ReasonForgedCovenantClaim
	ReasonHostilePayloadSignature
	ReasonNullByteSmuggling
	ReasonPayloadOversize
)

// Verdict captures the deterministic T4 decision of the Covenant Airlock.
type Verdict struct {
	State  data.QuadState // T4 truth verdict (+1, 0, -1, NULL)
	Reason ReasonCode     // Exact structural or provenance cause
}

// Filter configures the Genesis 3 Airlock evaluation policy.
type Filter struct {
	MaxExternalBytes int  // Upper byte ceiling on external commercial streams
	StrictTextCheck  bool // Reject embedded 0x00 null-byte smuggling
}

// ------------------------------------------------------------------------------
// Sub-Assembly 2.4: Pre-Conditional Helper Cogs & O(1) Verdict Predicates
// ------------------------------------------------------------------------------
// Pure O(1) verdict classification predicates and symbolic labels.
// ------------------------------------------------------------------------------

// IsPermitted reports whether the verdict allows delivery (+1 or 0:AWAIT).
//
// ### For Architects, Operators & Non-Programmers:
// Returns true when an incoming message is safe to deliver—either verified
// sovereign truth (+1) or clean empirical outside data waiting in quarantine (0).
//
// ### For Systems Programmers & Engine Architects:
// Pure O(1) T4 predicate matching `QuadPositive` or `QuadAwait`.
func (v Verdict) IsPermitted() bool {
	return v.State == data.QuadPositive || v.State == data.QuadAwait
}

// IsBlocked reports whether the verdict denies delivery (-1 or NULL).
//
// ### For Architects, Operators & Non-Programmers:
// Returns true when the airlock stopped a hostile payload (-1) or structurally
// broken envelope (NULL) at the door.
//
// ### For Systems Programmers & Engine Architects:
// Pure O(1) T4 predicate matching `QuadNegative` or `QuadNull`.
func (v Verdict) IsBlocked() bool {
	return v.State == data.QuadNegative || v.State == data.QuadNull
}

// String returns the canonical symbolic label of ReasonCode.
//
// ### For Architects, Operators & Non-Programmers:
// Renders the numeric airlock reason code into a clear diagnostic tag.
//
// ### For Systems Programmers & Engine Architects:
// Pure O(1) switch returning static string literals (0 heap allocations).
func (r ReasonCode) String() string {
	switch r {
	case ReasonAlignedSovereign:
		return "ALIGNED_SOVEREIGN"
	case ReasonEmpiricalExternalAwait:
		return "EMPIRICAL_EXTERNAL_AWAIT"
	case ReasonMalformedStructure:
		return "MALFORMED_STRUCTURE"
	case ReasonForgedCovenantClaim:
		return "FORGED_COVENANT_CLAIM"
	case ReasonHostilePayloadSignature:
		return "HOSTILE_PAYLOAD_SIGNATURE"
	case ReasonNullByteSmuggling:
		return "NULL_BYTE_SMUGGLING"
	case ReasonPayloadOversize:
		return "PAYLOAD_OVERSIZE"
	default:
		return "UNKNOWN_REASON"
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 2.5: Monitored Support Infrastructure
// ------------------------------------------------------------------------------
// Support Role:  Zero-allocation ASCII case-insensitive substring matcher.
// Pattern Watch: If interpreter/ or future CPI-SI lexical engines also require
//                zero-heap ASCII folding, elevate into a shared text utility cog.
// ------------------------------------------------------------------------------

func containsFoldASCII(haystack, needle []byte) bool {
	nLen := len(needle)
	hLen := len(haystack)
	if nLen == 0 || nLen > hLen {
		return false
	}
	for i := 0; i <= hLen-nLen; i++ {
		match := true
		for j := 0; j < nLen; j++ {
			a := haystack[i+j]
			b := needle[j]
			if a >= 'A' && a <= 'Z' {
				a += 'a' - 'A'
			}
			if b >= 'A' && b <= 'Z' {
				b += 'a' - 'A'
			}
			if a != b {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// ==============================================================================
// BLOCK 3: BODY BLOCK
// ( Composable Business Logic: Frame & External Stream Airlock Evaluation )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 1: Pure Functional Engine Substrate (Covenant Airlock)
// Module:        cog.covenant.airlock
// Mission:       Evaluate sovereign frames and untrusted external payloads
//                against T4 invariants without mutating core cognitive state.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: Sovereign & External Wire Frame Ingress Evaluation
// ------------------------------------------------------------------------------

// EvaluateFrame audits a decoded Frame before the Dispatcher routes it.
//
// ### For Architects, Operators & Non-Programmers:
// Checks every arriving FaithNet envelope at the door: verifies it is intact,
// ensures external packets cannot pretend to be trusted local peers, and
// inspects foreign payloads before letting them inside.
//
// ### For Systems Programmers & Engine Architects:
// Pure O(N) read-only inspection (0 allocs/op). Rejects forged
// FlagCovenantVerified on FlagExternalOrigin frames with QuadNegative (-1).
func (f *Filter) EvaluateFrame(frame *data.Frame) Verdict {
	if frame == nil || !frame.Opcode.IsValid() || !frame.Rung.IsValid() ||
		!frame.SourceID.IsValid() {
		return Verdict{State: data.QuadNull, Reason: ReasonMalformedStructure}
	}
	if int(frame.PayloadLen) > data.FloorMaxPayload {
		return Verdict{State: data.QuadNull, Reason: ReasonPayloadOversize}
	}

	isExternal := frame.Flags.Has(data.FlagExternalOrigin)

	// An untrusted external packet may NEVER arrive pre-stamped as verified.
	if isExternal && frame.Flags.Has(data.FlagCovenantVerified) {
		return Verdict{State: data.QuadNegative, Reason: ReasonForgedCovenantClaim}
	}

	// Binary control pulses (PING/PONG/CASCADE) need only structural integrity.
	if frame.Opcode == data.OpPing || frame.Opcode == data.OpPong ||
		frame.Opcode == data.OpCascadeDown || frame.Opcode == data.OpCascadeUp {
		if isExternal {
			return Verdict{State: data.QuadAwait, Reason: ReasonEmpiricalExternalAwait}
		}
		return Verdict{State: data.QuadPositive, Reason: ReasonAlignedSovereign}
	}

	payload := frame.PayloadSlice()
	if isExternal || frame.Opcode == data.OpLegacyWrap {
		return f.EvaluateExternalPayload(payload)
	}

	return Verdict{State: data.QuadPositive, Reason: ReasonAlignedSovereign}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: Untrusted Commercial Web Payload Evaluation
// ------------------------------------------------------------------------------

// EvaluateExternalPayload audits raw bytes arriving from the legacy internet
// or foreign OS bridges. Clean empirical bytes enter strictly as 0:AWAIT.
//
// ### For Architects, Operators & Non-Programmers:
// Inspects raw data fetched from the outside internet. Dangerous scripts,
// database attacks, and hidden null bytes are turned away immediately; clean
// facts are marked "0:AWAIT" so the core can read them without being polluted.
//
// ### For Systems Programmers & Engine Architects:
// Zero-heap bounded scan enforcing MaxExternalBytes, optional null-byte guard,
// and ASCII-folded hostile signature screening.
func (f *Filter) EvaluateExternalPayload(raw []byte) Verdict {
	limit := f.MaxExternalBytes
	if limit <= 0 {
		limit = DefaultAirlockMaxBytes
	}
	if len(raw) > limit {
		return Verdict{State: data.QuadNegative, Reason: ReasonPayloadOversize}
	}
	if f.StrictTextCheck && bytes.IndexByte(raw, 0x00) >= 0 {
		return Verdict{State: data.QuadNegative, Reason: ReasonNullByteSmuggling}
	}
	for i := 0; i < len(hostileSignatures); i++ {
		if containsFoldASCII(raw, hostileSignatures[i]) {
			return Verdict{
				State:  data.QuadNegative,
				Reason: ReasonHostilePayloadSignature,
			}
		}
	}
	return Verdict{State: data.QuadAwait, Reason: ReasonEmpiricalExternalAwait}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Contract Verification Cogs & Pre-Flight Validation Tripwires )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 1: Pure Functional Engine Substrate (Covenant Airlock)
// Module:        cog.covenant.airlock
// Mission:       Audit live Filter policy bounds and execute a 4-section
//                pre-flight self-audit verifying all 4 T4 airlock verdicts.
// @triggers
//   - Block 4 Audits : (*Filter).AuditPolicy(), PreflightCovenantFilter(),
//                      AuditCovenantFilter()
//   - Layer 4 Tests  : tests/functional/covenant_functional_test.go,
//                      tests/unit/engine/covenant_test.go
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Contract, Input & Live-Instance Verification Cogs
// ------------------------------------------------------------------------------

// AuditPolicy inspects a live Filter non-destructively to confirm its byte
// ceiling and immutable hostile signature table remain intact.
//
// ### For Architects, Operators & Non-Programmers:
// Checks the live airlock settings to ensure maximum packet limits are valid
// and the threat-signature list is armed and non-empty.
//
// ### For Systems Programmers & Engine Architects:
// Read-only O(N_sigs) live policy probe returning structured `data.AuditResult`.
func (f *Filter) AuditPolicy() data.AuditResult {
	const check = "airlock_live_policy_contract"

	if f == nil {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDAirlock, check,
			"Filter receiver is nil.")
	}
	if len(hostileSignatures) == 0 {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDAirlock, check,
			"Hostile signature table is empty; airlock unguarded.")
	}
	if f.MaxExternalBytes < 0 || f.MaxExternalBytes > data.FloorMaxPayload {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDAirlock, check,
			"Filter MaxExternalBytes breaches 480-byte Floor MTU ceiling.")
	}
	return data.AuditPass(ModuleIDAirlock, check,
		"Live Genesis 3 Airlock policy ceiling and threat table verified.")
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Machine Self-Audit & Pre-Flight Validation Tripwires
// ------------------------------------------------------------------------------

// PreflightCovenantFilter executes a 4-section non-destructive pre-flight audit
// verifying (1) sovereign +1 and clean external 0:AWAIT, (2) forged flags,
// XSS/SQLi/null-byte -1 rejection and nil NULL:HALT, (3) 480B ceiling limits,
// and (4) live policy contract parity.
//
// ### For Architects, Operators & Non-Programmers:
// Bench-tests the Genesis 3 Airlock against good messages, clean outside facts,
// forged credentials, and web attacks before allowing any node traffic.
//
// ### For Systems Programmers & Engine Architects:
// Stack-pure 4-section pre-flight gate exercising all 4 T4 outcomes (`+1`, `0`,
// `-1`, `NULL`); returns structured `data.AuditResult`.
func PreflightCovenantFilter() data.AuditResult {
	const check = "airlock_preflight_4section"

	flt := NewFilter()
	if pol := flt.AuditPolicy(); !pol.IsPass() {
		return pol
	}

	// Micro Check 1 (Always Pass): Sovereign +1:ALIGNED & clean external 0:AWAIT.
	var sov data.Frame
	sov.Rung = data.Rung3Local
	sov.Opcode = data.OpCPISync
	sov.SourceID = 0x01
	sov.SetPayload([]byte("Sovereign truth"))
	if v := flt.EvaluateFrame(&sov); v.State != data.QuadPositive || !v.IsPermitted() {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDAirlock, check,
			"Sec 1 sovereign frame was not admitted as +1:ALIGNED.")
	}
	vClean := flt.EvaluateExternalPayload([]byte("Temp: 72F, Clear Skies"))
	if vClean.State != data.QuadAwait || !vClean.IsPermitted() {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDAirlock, check,
			"Sec 1 clean external payload was not quarantined as 0:AWAIT.")
	}

	// Micro Check 2 (Never Fail): Forged claim, hostile XSS, null smuggle, nil.
	var forged data.Frame
	forged.Rung = data.Rung4Global
	forged.Opcode = data.OpLegacyWrap
	forged.Flags = data.FlagExternalOrigin | data.FlagCovenantVerified
	forged.SourceID = 0x99
	if v := flt.EvaluateFrame(&forged); v.State != data.QuadNegative ||
		v.Reason != ReasonForgedCovenantClaim {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDAirlock, check,
			"Sec 2 forged FlagCovenantVerified external frame bypassed airlock.")
	}
	if v := flt.EvaluateExternalPayload([]byte("<SCRIPT>alert(1)</script>")); v.State != data.QuadNegative || !v.IsBlocked() {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDAirlock, check,
			"Sec 2 case-folded script injection bypassed airlock.")
	}
	if v := flt.EvaluateExternalPayload([]byte("a\x00b")); v.State != data.QuadNegative ||
		v.Reason != ReasonNullByteSmuggling {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDAirlock, check,
			"Sec 2 null-byte smuggling bypassed strict airlock.")
	}
	if v := flt.EvaluateFrame(nil); v.State != data.QuadNull || !v.IsBlocked() {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDAirlock, check,
			"Sec 2 nil frame failed to emit NULL:HALT verdict.")
	}

	// Micro Check 3 (Invariants & Limits): Exact 480B ceiling vs 481B reject.
	var ceilingBuf [data.FloorMaxPayload + 1]byte
	for i := range ceilingBuf {
		ceilingBuf[i] = 'A'
	}
	if v := flt.EvaluateExternalPayload(ceilingBuf[:data.FloorMaxPayload]); v.State != data.QuadAwait {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDAirlock, check,
			"Sec 3 exact 480B clean external payload was erroneously rejected.")
	}
	if v := flt.EvaluateExternalPayload(ceilingBuf[:]); v.State != data.QuadNegative ||
		v.Reason != ReasonPayloadOversize {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDAirlock, check,
			"Sec 3 481B oversized external payload bypassed ceiling guard.")
	}

	// Micro Check 4 (Emergence): External control pulse enters as 0:AWAIT.
	var extPing data.Frame
	extPing.Rung = data.Rung4Global
	extPing.Opcode = data.OpPing
	extPing.Flags = data.FlagExternalOrigin
	extPing.SourceID = 0x77
	if v := flt.EvaluateFrame(&extPing); v.State != data.QuadAwait {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDAirlock, check,
			"Sec 4 external control pulse was not downgraded to 0:AWAIT.")
	}

	return data.AuditPass(ModuleIDAirlock, check,
		"Genesis 3 Airlock passed all 4 micro pre-flight gates across T4.")
}

// AuditCovenantFilter returns true iff PreflightCovenantFilter() passes cleanly.
//
// ### For Architects, Operators & Non-Programmers:
// Quick pass/fail shortcut asking whether the Genesis 3 Airlock is armed.
//
// ### For Systems Programmers & Engine Architects:
// Convenience boolean bridge delegating to `PreflightCovenantFilter().IsPass()`.
func AuditCovenantFilter() bool {
	return PreflightCovenantFilter().IsPass()
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK / EXECUTABLE PROCEDURES & LIFECYCLE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 1: Pure Functional Engine Substrate (Covenant Airlock)
// Module:        cog.covenant.airlock
// Mission:       Render 60-column Genesis 3 Airlock diagnostic cards, expose
//                epilogue readiness, and own Filter policy lifecycle creation.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Executable Procedures, Inspection, Readiness & Cards
// ------------------------------------------------------------------------------

// IsAirlockReady reports whether the Genesis 3 Airlock passed pre-flight audit.
//
// ### For Architects, Operators & Non-Programmers:
// Epilogue confirmation that the epistemic boundary airlock is healthy.
//
// ### For Systems Programmers & Engine Architects:
// Matches OmniCode's Block 5 epilogue readiness predicate (`is_*_ready`).
func IsAirlockReady() bool {
	return PreflightCovenantFilter().IsPass()
}

// ExecuteAirlockDumpCard renders a 60-column inner-width diagnostic card.
//
// ### For Architects, Operators & Non-Programmers:
// Prints a framed dashboard showing the airlock's byte limit, threat-signature
// count, null-byte strictness, and startup safety check verdict.
//
// ### For Systems Programmers & Engine Architects:
// Non-mutating diagnostic card builder invoking `PreflightCovenantFilter()`.
func (f *Filter) ExecuteAirlockDumpCard() string {
	pf := PreflightCovenantFilter()
	return fmt.Sprintf(
		"+------------------------------------------------------------+\n"+
			"| %-58s |\n"+
			"+------------------------------------------------------------+\n"+
			"| Module:     %-46s |\n"+
			"| MaxBytes:   %-46d |\n"+
			"| Signatures: %-46d |\n"+
			"| StrictNull: %-46t |\n"+
			"| Preflight:  %-46s |\n"+
			"+------------------------------------------------------------+",
		"FAITHNET GENESIS 3 COVENANT AIRLOCK CARD",
		ModuleIDAirlock,
		f.MaxExternalBytes,
		len(hostileSignatures),
		f.StrictTextCheck,
		pf.Severity.String(),
	)
}

// ------------------------------------------------------------------------------
// Sub-Assembly 5.2: Subsystem Lifecycle Initialization & Teardown
// ------------------------------------------------------------------------------

// NewFilter creates a Covenant Filter bounded to FaithNet's Floor MTU ceiling.
//
// ### For Architects, Operators & Non-Programmers:
// Builds a ready-to-use Genesis 3 Airlock with strict text and size guards on.
//
// ### For Systems Programmers & Engine Architects:
// Initializes default policy (`DefaultAirlockMaxBytes = 480`,
// `StrictTextCheck = true`).
func NewFilter() *Filter {
	return &Filter{
		MaxExternalBytes: DefaultAirlockMaxBytes,
		StrictTextCheck:  true,
	}
}

// End of engine/covenant/filter.go
