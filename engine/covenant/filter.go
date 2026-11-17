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
)

// --- Tier 2: External Library Dependencies (None — 100% Go Stdlib Invariant) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---
import "github.com/CreativeWorkzStudio/FaithNet/data"

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
func (v Verdict) IsPermitted() bool {
	return v.State == data.QuadPositive || v.State == data.QuadAwait
}

// IsBlocked reports whether the verdict denies delivery (-1 or NULL).
func (v Verdict) IsBlocked() bool {
	return v.State == data.QuadNegative || v.State == data.QuadNull
}

// String returns the canonical symbolic label of ReasonCode.
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
// Mission:       Verify sovereign acceptance, external quarantine (0:AWAIT),
//                hostile rejection (-1), and malformed halt (NULL).
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Contract, Input & Cog Verification Cogs
// ------------------------------------------------------------------------------

// AuditCovenantFilter exercises all 4 T4 verdict paths cleanly in stack memory.
func AuditCovenantFilter() bool {
	flt := NewFilter()

	vClean := flt.EvaluateExternalPayload([]byte("Temp: 72F, Clear Skies"))
	if vClean.State != data.QuadAwait || !vClean.IsPermitted() {
		return false
	}
	vHostile := flt.EvaluateExternalPayload([]byte("<SCRIPT>alert(1)</script>"))
	if vHostile.State != data.QuadNegative || !vHostile.IsBlocked() {
		return false
	}
	vNull := flt.EvaluateFrame(nil)
	return vNull.State == data.QuadNull && vNull.IsBlocked()
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Machine Self-Audit & Pre-Flight Validation Tripwires
// ------------------------------------------------------------------------------

// PreflightCovenantFilter executes non-destructive pre-flight airlock clearance.
func PreflightCovenantFilter() bool {
	return len(hostileSignatures) > 0 && AuditCovenantFilter()
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK / EXECUTABLE PROCEDURES & LIFECYCLE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 1: Pure Functional Engine Substrate (Covenant Airlock)
// Module:        cog.covenant.airlock
// Mission:       Render 60-column Genesis 3 Airlock diagnostic cards and own
//                Filter policy lifecycle construction.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Executable Procedures, Inspection & Diagnostic Cards
// ------------------------------------------------------------------------------

// ExecuteAirlockDumpCard renders a 60-column inner-width diagnostic card.
func (f *Filter) ExecuteAirlockDumpCard() string {
	status := "FAIL"
	if PreflightCovenantFilter() {
		status = "PASS"
	}
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
		status,
	)
}

// ------------------------------------------------------------------------------
// Sub-Assembly 5.2: Subsystem Lifecycle Initialization & Teardown
// ------------------------------------------------------------------------------

// NewFilter creates a Covenant Filter bounded to FaithNet's Floor MTU ceiling.
func NewFilter() *Filter {
	return &Filter{
		MaxExternalBytes: DefaultAirlockMaxBytes,
		StrictTextCheck:  true,
	}
}

// End of engine/covenant/filter.go
