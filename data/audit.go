// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          data/audit.go
// @id            cog.data.audit
// @layer         Layer 0: Pure Invariant Data Substrate
// @archetype     pure_types | stateless
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Universal Block 4 Diagnostic Verdict Vocabulary
// @role          Pure Contract & Vocabulary: AuditSeverity & AuditResult
// @maturity      Rev 5 Baseline | Wave 1.5 Audit Elevation
// @chassis       zero-heap | Fixed scalar/string-literal value record
// @scale         O(1) construction, predicate evaluation & formatting
// @seams         in: none (pure leaf) | uses: fmt | exports: AuditResult
// @invariants    1. Stratum Baseline: ALLOC_ZERO_HEAP, Downward Dependency Law,
//                   Universal 5-Block Lifecycle Protocol (Rev 5).
//                2. Mirrors Cornerstone C `diag_result_t` (`block_4_audit.json`)
//                   so every Layer 1..4 Block 4 probe emits uniform verdicts.
//                3. Zero OS, network, or external package imports.
//                4. Passing AuditResult ALWAYS pairs Passed==true with
//                   Severity==AuditSeverityPass.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Defines the shared Layer 0 diagnostic verdict record (`AuditResult`) and
//   severity scale (`AuditSeverity`) returned by every Block 4 contract
//   validator and pre-flight self-audit across FaithNet. Eliminates opaque
//   boolean failures in favor of structured forensic telemetry.
//
//   ### For Architects, Operators & Non-Programmers:
//   Gives every safety check in FaithNet a clear medical report card—not just
//   "failed", but which subsystem checked what rule, how severe the issue is,
//   and a plain-English explanation of what passed or tripped.
//
//   ### For Systems Programmers & Engine Architects:
//   - Ownership & Reentrancy: Pure caller-owned value struct; zero locks,
//     zero mutable global state, zero heap allocation for literal messages.
//   - Extension Seam: Consumed by Layer 1..3 Block 4 probes, Layer 4 shared
//     test harnesses (`tests/common/`), and CLI `--audit` reporters.
//   - Explicit Non-Goal: Holds zero stateful probe registries or I/O writers.
// ==============================================================================

package data

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Shared Pure Audit Vocabulary, Severity Scale & Stateless Predicates )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 0: Pure Invariant Data Substrate
// Module:        cog.data.audit
// Mission:       Provide zero-heap AuditSeverity and AuditResult value records
//                powering every Block 4 Audit & Preflight cog across FaithNet.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import "fmt"

// --- Tier 2: External Library Dependencies (None — 100% Go Stdlib Invariant) ---

// --- Tier 3: Internal FaithNet Foundation Packages (None — Pure Leaf Root) ---

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Pre-Conditions, Physical Bounds & Severity Limits
// ------------------------------------------------------------------------------
// Pinned severity bounds mirroring Cornerstone `diag_severity_t` semantics.
// ------------------------------------------------------------------------------

const (
	// AuditSeverityMinValid is the lowest canonical diagnostic severity.
	AuditSeverityMinValid AuditSeverity = AuditSeverityPass

	// AuditSeverityMaxValid is the highest canonical diagnostic severity.
	AuditSeverityMaxValid AuditSeverity = AuditSeverityFatal
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.2: Pure-Types Logical Anchor
// ------------------------------------------------------------------------------
// Movement-resilient identifier shared by diagnostic and audit subsystems.
// ------------------------------------------------------------------------------

const (
	// ModuleIDAudit is the canonical logical ID for this pure_types leaf.
	ModuleIDAudit = "cog.data.audit"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.3: Public Tagged Ontology (AuditSeverity & AuditResult)
// ------------------------------------------------------------------------------
// Canonical diagnostic verdict structures returned by all Block 4 cogs.
// ------------------------------------------------------------------------------

// AuditSeverity grades the operational impact of a Block 4 audit or preflight.
//
// ### For Architects, Operators & Non-Programmers:
// Ranks diagnostic findings from clean clearance (PASS) through advisory
// warning (WARN) up to critical degradation (CRITICAL) and hard stop (FATAL).
//
// ### For Systems Programmers & Engine Architects:
// Compact 1-byte enum aligned directly with Cornerstone's `SEVERITY_PASS`,
// `SEVERITY_WARNING`, `SEVERITY_CRITICAL`, and `SEVERITY_FATAL`.
type AuditSeverity uint8

const (
	// AuditSeverityPass indicates all audited contracts and invariants held.
	AuditSeverityPass AuditSeverity = 0

	// AuditSeverityWarn indicates a non-fatal advisory or degraded posture.
	AuditSeverityWarn AuditSeverity = 1

	// AuditSeverityCritical indicates a breached contract requiring fallback.
	AuditSeverityCritical AuditSeverity = 2

	// AuditSeverityFatal indicates an unrecoverable structural invariant fault.
	AuditSeverityFatal AuditSeverity = 3
)

// AuditResult is the universal Block 4 diagnostic verdict record across layers.
//
// ### For Architects, Operators & Non-Programmers:
// Carries the full outcome of a pre-flight or live health check—whether it
// passed, its severity, which cog ran it, which rule was tested, and why.
//
// ### For Systems Programmers & Engine Architects:
// Stack-allocated Go value equivalent of Cornerstone C's `diag_result_t`.
// Returned by value with zero heap escapes when populated with string literals.
type AuditResult struct {
	Passed    bool          // True iff the audited contract/preflight passed
	Severity  AuditSeverity // Graded diagnostic severity (Pass..Fatal)
	Subsystem string        // Logical @id of the audited cog (e.g. cog.engine.codec)
	CheckName string        // Name of the specific contract or preflight check
	Details   string        // Human-readable forensic explanation of outcome
}

// ------------------------------------------------------------------------------
// Sub-Assembly 2.4: Pre-Conditional Helper Cogs & Stateless Verdict Constructors
// ------------------------------------------------------------------------------
// Pure O(1) verdict builders, consistency predicates, and canonical formatters.
// ------------------------------------------------------------------------------

// IsValid reports whether s lies within the canonical AuditSeverity range.
//
// ### For Architects, Operators & Non-Programmers:
// Confirms a severity tag is a real level (PASS..FATAL) rather than garbage.
//
// ### For Systems Programmers & Engine Architects:
// O(1) unsigned upper-bound check against AuditSeverityMaxValid.
func (s AuditSeverity) IsValid() bool {
	return s <= AuditSeverityMaxValid
}

// String renders AuditSeverity as a canonical uppercase diagnostic label.
//
// ### For Architects, Operators & Non-Programmers:
// Converts a numeric severity level into readable text (PASS, WARN, CRITICAL,
// or FATAL) for diagnostic cards and test logs.
//
// ### For Systems Programmers & Engine Architects:
// Pure O(1) switch returning static string literals with zero allocations.
func (s AuditSeverity) String() string {
	switch s {
	case AuditSeverityPass:
		return "PASS"
	case AuditSeverityWarn:
		return "WARN"
	case AuditSeverityCritical:
		return "CRITICAL"
	case AuditSeverityFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// AuditPass constructs a canonical passing AuditResult verdict.
//
// ### For Architects, Operators & Non-Programmers:
// Stamps a clean green-light diagnostic report for a verified cog check.
//
// ### For Systems Programmers & Engine Architects:
// Stack-pure value constructor enforcing Passed==true and AuditSeverityPass.
func AuditPass(subsystem, checkName, details string) AuditResult {
	return AuditResult{
		Passed:    true,
		Severity:  AuditSeverityPass,
		Subsystem: subsystem,
		CheckName: checkName,
		Details:   details,
	}
}

// AuditFail constructs a canonical failing AuditResult verdict with severity.
//
// ### For Architects, Operators & Non-Programmers:
// Records an exact diagnostic tripwire report explaining which invariant broke
// and how urgently the system must react.
//
// ### For Systems Programmers & Engine Architects:
// Clamps an invalid or Pass severity on a failure upward to AuditSeverityCritical
// so failing verdicts can never masquerade as SEVERITY_PASS.
func AuditFail(
	sev AuditSeverity,
	subsystem, checkName, details string,
) AuditResult {
	if sev == AuditSeverityPass || !sev.IsValid() {
		sev = AuditSeverityCritical
	}
	return AuditResult{
		Passed:    false,
		Severity:  sev,
		Subsystem: subsystem,
		CheckName: checkName,
		Details:   details,
	}
}

// IsPass reports whether r represents an unambiguous passing audit verdict.
//
// ### For Architects, Operators & Non-Programmers:
// Returns true only when the check passed AND carries no warning/critical flag.
//
// ### For Systems Programmers & Engine Architects:
// Enforces dual consistency (`r.Passed && r.Severity == AuditSeverityPass`).
func (r AuditResult) IsPass() bool {
	return r.Passed && r.Severity == AuditSeverityPass
}

// String renders AuditResult as a compact single-line forensic summary.
//
// ### For Architects, Operators & Non-Programmers:
// Formats a safety check result for terminal output, logs, or test failures.
//
// ### For Systems Programmers & Engine Architects:
// Intended solely for diagnostic and failure paths, never hot wire loops.
func (r AuditResult) String() string {
	return fmt.Sprintf("[%s] %s :: %s — %s",
		r.Severity.String(), r.Subsystem, r.CheckName, r.Details)
}

// ==============================================================================
// BLOCK 3: BODY BLOCK
// ( Statutory Pure-Types Notice & Structural Boundary )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 0: Pure Invariant Data Substrate
// Module:        cog.data.audit
// Mission:       Pure-types diagnostic vocabulary module: define immutable
//                AuditSeverity and AuditResult records and O(1) constructors;
//                operational verification probes reside in Layer 1..3 Block 4s.
// ==============================================================================

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Statutory Pure-Types Audit Cross-Reference )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 0: Pure Invariant Data Substrate
// Module:        cog.data.audit
// Mission:       Functional audits verifying AuditResult consistency and Layer 0
//                vocabulary bounds are declared in `engine/codec.go`
//                (`AuditDataVocabulary`, `PreflightCodec`) and verified in
//                `tests/unit/data/audit_test.go`.
// ==============================================================================

// ==============================================================================
// BLOCK 5: CLOSING BLOCK / STATUTORY EXEMPTION & LIFECYCLE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 0: Pure Invariant Data Substrate
// Module:        cog.data.audit
// Mission:       Operational inspection cards and readiness predicates reside in
//                Layer 1..3 Block 5 modules.
// Stateless Exemption Notice: None. (Stateless module / pure stack-allocated execution).
// ==============================================================================

// End of data/audit.go
