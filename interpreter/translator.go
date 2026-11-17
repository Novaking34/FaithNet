// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          interpreter/translator.go
// @id            cog.interpreter.translator
// @layer         Layer 3B: Universal OS & Legacy Interpreter ("The Well")
// @archetype     operational | stateless
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Foreign Payload <-> Sovereign Frame Translator
// @role          "The Well" Airlock-Gated Foreign Wrapping & Sovereign Formatting
// @maturity      Rev 5 Baseline | Wave 4 The Room & The Well
// @chassis       bounded-heap | Pure translation gated by Genesis 3 Airlock
// @scale         O(N_payload) bounded <= 480B Floor MTU screening & wrapping
// @seams         in: data, engine/covenant | exports: Translator, TranslatedView
// @invariants    1. Stratum Baseline: Downward Dependency Law, Rev 5 5-Block.
//                2. Foreign payloads NEVER receive FlagCovenantVerified or
//                   QuadPositive (+1); clean text enters strictly as 0:AWAIT.
//                3. Hostile foreign payloads are rejected (-1) prior to frame
//                   construction.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Translates untrusted text/JSON messages arriving from standard modern
//   devices (Windows, macOS, Linux, iOS, Android web browsers) into bounded
//   512-byte FaithNet `data.OpLegacyWrap` frames strictly screened by the
//   Genesis 3 Covenant Airlock, and renders sovereign frames back into safe
//   display records for standard browsers.
//
//   ### For Architects, Operators & Non-Programmers:
//   Acts as a bilingual ambassador at The Well: any ordinary phone or computer
//   can speak plain text to FaithNet, and the translator checks every word at
//   the airlock before wrapping it into a sovereign FaithNet envelope.
//
//   ### For Systems Programmers & Engine Architects:
//   - Ownership & Reentrancy: Holds an immutable `*covenant.Filter`; 100%
//     goroutine-safe across concurrent HTTP portal requests.
//   - Extension Seam: Used by `interpreter/captive_web.go` and `legacy_bridge.go`.
// ==============================================================================

package interpreter

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Constants, Sentinel Errors, Logical Anchor & Translator Chassis )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3B: Universal OS & Legacy Interpreter ("The Well")
// Module:        cog.interpreter.translator
// Mission:       Define the Airlock-gated Translator and safe browser view type.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"errors"
	"fmt"
	"strings"

	"github.com/CreativeWorkzStudio/FaithNet/data"
	"github.com/CreativeWorkzStudio/FaithNet/engine/covenant" // --- Tier 2: External Library Dependencies (None — 100% Go Stdlib Invariant) ---
)

// --- Tier 3: Internal FaithNet Foundation Packages ---

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Pre-Conditions & Sentinel Errors
// ------------------------------------------------------------------------------

var (
	// ErrForeignPayloadBlocked indicates the Genesis 3 Airlock rejected input.
	ErrForeignPayloadBlocked = errors.New("faithnet/interpreter: foreign payload blocked by Genesis 3 Airlock")

	// ErrEmptyForeignPayload indicates blank input with zero visible content.
	ErrEmptyForeignPayload = errors.New("faithnet/interpreter: foreign payload is empty")
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.2: Operational Logical Anchor
// ------------------------------------------------------------------------------

const (
	// ModuleIDTranslator is the canonical logical ID for The Well translator.
	ModuleIDTranslator = "cog.interpreter.translator"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.3: Public Tagged Ontology (TranslatedView & Translator)
// ------------------------------------------------------------------------------

// TranslatedView is a safe, browser-ready summary of a sovereign FaithNet Frame.
type TranslatedView struct {
	SourceID string `json:"source_id"`
	TargetID string `json:"target_id"`
	Rung     string `json:"rung"`
	Opcode   string `json:"opcode"`
	Truth    string `json:"truth"`
	SeqID    uint32 `json:"seq_id"`
	Text     string `json:"text"`
}

// Translator converts foreign OS/Web text into airlocked FaithNet frames and
// renders sovereign frames into safe browser-readable records.
//
// ### For Architects, Operators & Non-Programmers:
// Bridges standard web browsers and FaithNet nodes safely through the airlock.
//
// ### For Systems Programmers & Engine Architects:
// Stateless wrapper over an immutable `*covenant.Filter` policy.
type Translator struct {
	airlock *covenant.Filter
}

// ------------------------------------------------------------------------------
// Sub-Assembly 2.4: Pre-Conditional Helper Cogs & O(1) Accessors
// ------------------------------------------------------------------------------

// Airlock returns the underlying Genesis 3 Covenant Filter guarding translation.
//
// ### For Architects, Operators & Non-Programmers:
// Exposes the security filter attached to this translator for inspection.
//
// ### For Systems Programmers & Engine Architects:
// O(1) pointer getter.
func (tr *Translator) Airlock() *covenant.Filter {
	return tr.airlock
}

// ==============================================================================
// BLOCK 3: BODY BLOCK
// ( Composable Business Logic: Foreign->Frame Wrap & Frame->Browser View )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3B: Universal OS & Legacy Interpreter ("The Well")
// Module:        cog.interpreter.translator
// Mission:       Screen untrusted foreign payloads into 0:AWAIT OpLegacyWrap
//                frames and export clean browser-safe views from Frame values.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: Foreign OS / Browser Payload -> Airlocked FaithNet Frame
// ------------------------------------------------------------------------------

// WrapForeignMessage screens rawBytes through the Genesis 3 Airlock and wraps
// clean empirical input into a bounded `data.OpLegacyWrap` Frame stamped with
// `FlagExternalOrigin | FlagLegacyTranslated | FlagFloorCompact` and `0:AWAIT`.
//
// ### For Architects, Operators & Non-Programmers:
// Inspects a message typed by a visitor on a phone or laptop—turning away any
// harmful scripts immediately, and sealing safe text inside a FaithNet envelope
// marked "0:AWAIT" so the local node stays protected.
//
// ### For Systems Programmers & Engine Architects:
// Returns `(frame, verdict, nil)` on clean empirical text, or
// `(Frame{}, verdict, ErrForeignPayloadBlocked)` when `-1:REJECT` / `NULL:HALT`.
func (tr *Translator) WrapForeignMessage(
	src, dst data.NodeID,
	rung data.RungLevel,
	seq uint32,
	rawBytes []byte,
) (data.Frame, covenant.Verdict, error) {
	trimmed := strings.TrimSpace(string(rawBytes))
	if trimmed == "" {
		return data.Frame{}, covenant.Verdict{
			State:  data.QuadNull,
			Reason: covenant.ReasonMalformedStructure,
		}, ErrEmptyForeignPayload
	}

	verdict := tr.airlock.EvaluateExternalPayload(rawBytes)
	if verdict.IsBlocked() {
		return data.Frame{}, verdict, ErrForeignPayloadBlocked
	}

	if !rung.IsValid() {
		rung = data.Rung3Local
	}
	var f data.Frame
	f.Version = data.ProtocolVersion
	f.Rung = rung
	f.Opcode = data.OpLegacyWrap
	f.Flags = data.FlagExternalOrigin | data.FlagLegacyTranslated | data.FlagFloorCompact
	f.Truth = verdict.State // Strictly 0:AWAIT for clean external input.
	f.SourceID = src
	f.TargetID = dst
	f.SeqID = seq
	if !f.SetPayload(rawBytes) {
		return data.Frame{}, covenant.Verdict{
			State:  data.QuadNegative,
			Reason: covenant.ReasonPayloadOversize,
		}, ErrForeignPayloadBlocked
	}
	return f, verdict, nil
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: Sovereign FaithNet Frame -> Safe Browser View
// ------------------------------------------------------------------------------

// FormatFrameForBrowser converts a FaithNet `data.Frame` into a browser-safe
// `TranslatedView` record with explicit truth and rung badges.
//
// ### For Architects, Operators & Non-Programmers:
// Turns a binary FaithNet envelope into plain readable fields that any phone
// or desktop browser can show cleanly.
//
// ### For Systems Programmers & Engine Architects:
// Pure read-only formatter stripping raw NUL bytes for clean JSON/HTML export.
func (tr *Translator) FormatFrameForBrowser(f *data.Frame) TranslatedView {
	if f == nil {
		return TranslatedView{
			Opcode: data.OpInvalid.String(),
			Truth:  data.QuadNull.String(),
		}
	}
	cleanText := strings.ReplaceAll(string(f.PayloadSlice()), "\x00", "")
	return TranslatedView{
		SourceID: f.SourceID.String(),
		TargetID: f.TargetID.String(),
		Rung:     f.Rung.String(),
		Opcode:   f.Opcode.String(),
		Truth:    f.Truth.String(),
		SeqID:    f.SeqID,
		Text:     cleanText,
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Contract Verification Cogs & Pre-Flight Validation Tripwires )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3B: Universal OS & Legacy Interpreter ("The Well")
// Module:        cog.interpreter.translator
// Mission:       Audit Translator airlock coupling and run a 4-section preflight
//                verifying clean quarantine (0:AWAIT) and hostile blocks (-1).
// @triggers
//   - Block 4 Audits : (*Translator).AuditState(), PreflightTranslator(),
//                      AuditTranslator()
//   - Layer 4 Tests  : tests/functional/interpreter_functional_test.go,
//                      tests/unit/interpreter/translator_test.go,
//                      tests/integrated/room_and_well_test.go,
//                      tests/system/airlock_hostility_system_test.go
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Contract, Input & Live-Instance Verification Cogs
// ------------------------------------------------------------------------------

// AuditState inspects a live Translator non-destructively to confirm its
// Genesis 3 Airlock is attached and passes `AuditPolicy()`.
//
// ### For Architects, Operators & Non-Programmers:
// Verifies The Well's translator still has its security airlock properly armed.
//
// ### For Systems Programmers & Engine Architects:
// Non-mutating live probe returning structured `data.AuditResult`.
func (tr *Translator) AuditState() data.AuditResult {
	const check = "interpreter_translator_live_contract"

	if tr == nil || tr.airlock == nil {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDTranslator, check,
			"Translator is nil or unguarded (missing Genesis 3 Airlock).")
	}
	return tr.airlock.AuditPolicy()
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Machine Self-Audit & Pre-Flight Validation Tripwires
// ------------------------------------------------------------------------------

// PreflightTranslator executes a 4-section pre-flight check verifying
// (1) clean foreign wrap -> 0:AWAIT, (2) XSS/empty rejection, (3) 480B floor
// ceiling, and (4) browser view round-trip parity.
//
// ### For Architects, Operators & Non-Programmers:
// Bench-tests The Well's message translator to prove ordinary visitor text
// enters safely while hostile web scripts are stopped at the gate.
//
// ### For Systems Programmers & Engine Architects:
// Isolated 4-section pre-flight gate returning structured `data.AuditResult`.
func PreflightTranslator() data.AuditResult {
	const check = "interpreter_translator_preflight_4section"

	tr := NewTranslator(nil)
	if live := tr.AuditState(); !live.IsPass() {
		return live
	}

	// Micro Check 1 (Always Pass): Clean foreign text wraps into 0:AWAIT frame.
	f, v, err := tr.WrapForeignMessage(0xEE, 0x01, data.Rung3Local, 1, []byte("Hello from Android"))
	if err != nil || v.State != data.QuadAwait || f.Truth != data.QuadAwait ||
		!f.Flags.Has(data.FlagExternalOrigin|data.FlagLegacyTranslated) ||
		f.Flags.Has(data.FlagCovenantVerified) {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDTranslator, check,
			"Sec 1 clean foreign wrap violated 0:AWAIT or flag invariants.")
	}

	// Micro Check 2 (Never Fail): Hostile script & empty payload blocked cleanly.
	if _, hv, hErr := tr.WrapForeignMessage(0xEE, 0x01, data.Rung3Local, 2, []byte("<script>x</script>")); hErr != ErrForeignPayloadBlocked || hv.State != data.QuadNegative {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDTranslator, check,
			"Sec 2 hostile XSS payload bypassed Translator airlock.")
	}
	if _, _, eErr := tr.WrapForeignMessage(0xEE, 0x01, data.Rung3Local, 3, []byte("   ")); eErr != ErrEmptyForeignPayload {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDTranslator, check,
			"Sec 2 blank payload did not return ErrEmptyForeignPayload.")
	}

	// Micro Check 3 (Invariants & Limits): 481B foreign payload blocked.
	var big [data.FloorMaxPayload + 1]byte
	for i := range big {
		big[i] = 'Z'
	}
	if _, _, oErr := tr.WrapForeignMessage(0xEE, 0x01, data.Rung3Local, 4, big[:]); oErr != ErrForeignPayloadBlocked {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDTranslator, check,
			"Sec 3 481B foreign payload bypassed 480B Floor MTU ceiling.")
	}

	// Micro Check 4 (Emergence): Browser view formatting matches wrapped frame.
	view := tr.FormatFrameForBrowser(&f)
	if view.Text != "Hello from Android" || view.Truth != "0:AWAIT" || view.Opcode != "LEGACY_WRAP" {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDTranslator, check,
			"Sec 4 FormatFrameForBrowser output diverged from wrapped frame.")
	}
	return data.AuditPass(ModuleIDTranslator, check,
		"Translator passed all 4 micro pre-flight gates.")
}

// AuditTranslator returns true iff PreflightTranslator() passes cleanly.
//
// ### For Architects, Operators & Non-Programmers:
// Quick pass/fail check confirming The Well's translator is armed.
//
// ### For Systems Programmers & Engine Architects:
// Convenience boolean bridge delegating to `PreflightTranslator().IsPass()`.
func AuditTranslator() bool {
	return PreflightTranslator().IsPass()
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK / EXECUTABLE PROCEDURES & LIFECYCLE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3B: Universal OS & Legacy Interpreter ("The Well")
// Module:        cog.interpreter.translator
// Mission:       Expose epilogue readiness, 60-col cards, and Translator init.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Executable Procedures, Inspection, Readiness & Cards
// ------------------------------------------------------------------------------

// IsTranslatorReady reports whether The Well's Translator passed pre-flight.
//
// ### For Architects, Operators & Non-Programmers:
// Epilogue confirmation that foreign-to-sovereign translation is healthy.
//
// ### For Systems Programmers & Engine Architects:
// Matches OmniCode's Block 5 epilogue readiness predicate (`is_*_ready`).
func IsTranslatorReady() bool {
	return PreflightTranslator().IsPass()
}

// ExecuteTranslatorDumpCard renders a 60-column inner-width translator card.
//
// ### For Architects, Operators & Non-Programmers:
// Summarizes The Well translator's byte limit and Genesis 3 Airlock clearance.
//
// ### For Systems Programmers & Engine Architects:
// Non-mutating diagnostic card builder invoking `AuditState()`.
func (tr *Translator) ExecuteTranslatorDumpCard() string {
	st := tr.AuditState()
	return fmt.Sprintf(
		"+------------------------------------------------------------+\n"+
			"| %-58s |\n"+
			"+------------------------------------------------------------+\n"+
			"| Module:   %-48s |\n"+
			"| FloorMTU: %-48s |\n"+
			"| TruthOut: %-48s |\n"+
			"| Audit:    %-48s |\n"+
			"+------------------------------------------------------------+",
		"FAITHNET THE WELL TRANSLATOR DIAGNOSTIC CARD",
		ModuleIDTranslator,
		fmt.Sprintf("%d bytes (OpLegacyWrap)", data.FloorMaxPayload),
		"0:AWAIT Quarantine (Never +1 for external)",
		st.Severity.String(),
	)
}

// ------------------------------------------------------------------------------
// Sub-Assembly 5.2: Subsystem Lifecycle Initialization & Teardown
// ------------------------------------------------------------------------------

// NewTranslator initializes a Translator backed by airlock (or a fresh Filter).
//
// ### For Architects, Operators & Non-Programmers:
// Creates a guarded translator ready to serve visitors at The Well portal.
//
// ### For Systems Programmers & Engine Architects:
// Defaults nil `airlock` to `covenant.NewFilter()` so it can never run unguarded.
func NewTranslator(airlock *covenant.Filter) *Translator {
	if airlock == nil {
		airlock = covenant.NewFilter()
	}
	return &Translator{airlock: airlock}
}

// End of interpreter/translator.go
