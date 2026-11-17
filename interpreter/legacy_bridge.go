// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          interpreter/legacy_bridge.go
// @id            cog.interpreter.legacy_bridge
// @layer         Layer 3B: Universal OS & Legacy Interpreter ("The Well")
// @archetype     operational | stateless
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Universal Foreign OS & Legacy Client Classifier
// @role          Data-Driven Client Profile Detection (Windows/Mac/Linux/Mobile)
// @maturity      Rev 5 Baseline | Wave 4 The Room & The Well
// @chassis       zero-heap | Pure O(N) User-Agent classification into value struct
// @scale         O(N_ua) stateless client body classification
// @seams         in: data | uses: fmt, strings | exports: DetectLegacyClient
// @invariants    1. Stratum Baseline: Downward Dependency Law, Rev 5 5-Block.
//                2. Multi-OS awareness lives STRICTLY inside Layer 3B data
//                   records—never leaking OS-specific branches into Layer 0/1.
//                3. Unknown or empty User-Agent degrades cleanly to ClientUnknown.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Implements FaithNet's universal legacy OS classifier ("Compatibility
//   Without Dependency"). Identifies visiting Windows, macOS, Linux, iOS,
//   Android, or raw CLI clients into a pure `LegacyClientProfile` value struct
//   so The Well can render appropriate captive hints and telemetry without
//   contaminating MillenniumOS's native protocol core.
//
//   ### For Architects, Operators & Non-Programmers:
//   Recognizes whether a visitor at The Well is holding an iPhone, Android,
//   Windows PC, Mac, or Linux laptop so FaithNet can greet them cleanly without
//   ever forcing them to install proprietary apps.
//
//   ### For Systems Programmers & Engine Architects:
//   - Ownership & Reentrancy: 100% stateless pure functions returning value
//     `LegacyClientProfile` structs; zero locks, safe across goroutines.
// ==============================================================================

package interpreter

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( ClientFamily Enum, LegacyClientProfile Record & O(1) Stringifiers )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3B: Universal OS & Legacy Interpreter ("The Well")
// Module:        cog.interpreter.legacy_bridge
// Mission:       Define foreign OS client ontology and data-driven profile tags.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"fmt"
	"strings"
)

// --- Tier 2: External Library Dependencies (None — 100% Go Stdlib Invariant) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---
import "github.com/CreativeWorkzStudio/FaithNet/data"

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Pre-Conditions & Bounds
// ------------------------------------------------------------------------------

const (
	// ClientFamilyMaxValid is the highest canonical legacy client family tag.
	ClientFamilyMaxValid ClientFamily = ClientCLI
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.2: Operational Logical Anchor
// ------------------------------------------------------------------------------

const (
	// ModuleIDLegacyBridge is the canonical logical ID for legacy OS bridging.
	ModuleIDLegacyBridge = "cog.interpreter.legacy_bridge"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.3: Public Tagged Ontology (ClientFamily & LegacyClientProfile)
// ------------------------------------------------------------------------------

// ClientFamily identifies the foreign operating system visiting The Well.
type ClientFamily uint8

const (
	ClientUnknown ClientFamily = iota
	ClientWindows
	ClientMacOS
	ClientLinux
	ClientAndroid
	ClientIOS
	ClientCLI
)

// LegacyClientProfile describes a foreign visiting device purely as data.
type LegacyClientProfile struct {
	Family          ClientFamily // Detected foreign OS family
	SupportsCaptive bool         // Whether OS natively triggers captive probes
	PrefersPlain    bool         // True for curl/CLI/terminal clients
}

// ------------------------------------------------------------------------------
// Sub-Assembly 2.4: Pre-Conditional Helper Cogs & O(1) Predicates
// ------------------------------------------------------------------------------

// IsValid reports whether c lies within canonical ClientFamily bounds.
//
// ### For Architects, Operators & Non-Programmers:
// Confirms a visitor's OS tag is a recognized client category.
//
// ### For Systems Programmers & Engine Architects:
// O(1) unsigned upper-bound check against `ClientFamilyMaxValid`.
func (c ClientFamily) IsValid() bool {
	return c <= ClientFamilyMaxValid
}

// String renders ClientFamily as a canonical uppercase label.
//
// ### For Architects, Operators & Non-Programmers:
// Displays the visiting device type cleanly (WINDOWS, MACOS, ANDROID, IOS, etc.).
//
// ### For Systems Programmers & Engine Architects:
// Pure O(1) static string switch (0 heap allocations).
func (c ClientFamily) String() string {
	switch c {
	case ClientWindows:
		return "WINDOWS"
	case ClientMacOS:
		return "MACOS"
	case ClientLinux:
		return "LINUX"
	case ClientAndroid:
		return "ANDROID"
	case ClientIOS:
		return "IOS"
	case ClientCLI:
		return "CLI_HTTP"
	default:
		return "UNKNOWN_OS"
	}
}

// ==============================================================================
// BLOCK 3: BODY BLOCK
// ( Composable Business Logic: Data-Driven Legacy Client Classification )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3B: Universal OS & Legacy Interpreter ("The Well")
// Module:        cog.interpreter.legacy_bridge
// Mission:       Classify arbitrary HTTP User-Agent strings into clean
//                LegacyClientProfile value records without OS coupling.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: User-Agent -> LegacyClientProfile Classification
// ------------------------------------------------------------------------------

// DetectLegacyClient classifies an HTTP User-Agent string into a pure
// `LegacyClientProfile` value record cleanly handling empty or unknown strings.
//
// ### For Architects, Operators & Non-Programmers:
// Looks gently at a visiting browser's greeting header so FaithNet knows
// whether it is speaking to a phone, desktop, or terminal tool.
//
// ### For Systems Programmers & Engine Architects:
// Case-insensitive ordered match: iOS/Android evaluated prior to macOS/Linux
// so mobile browsers carrying "Mac OS X" or "Linux" tokens classify accurately.
func DetectLegacyClient(userAgent string) LegacyClientProfile {
	ua := strings.ToLower(strings.TrimSpace(userAgent))
	switch {
	case ua == "":
		return LegacyClientProfile{Family: ClientUnknown}
	case strings.Contains(ua, "curl/") || strings.Contains(ua, "wget/") || strings.Contains(ua, "go-http-client"):
		return LegacyClientProfile{Family: ClientCLI, PrefersPlain: true}
	case strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad") || strings.Contains(ua, "ios"):
		return LegacyClientProfile{Family: ClientIOS, SupportsCaptive: true}
	case strings.Contains(ua, "android"):
		return LegacyClientProfile{Family: ClientAndroid, SupportsCaptive: true}
	case strings.Contains(ua, "windows"):
		return LegacyClientProfile{Family: ClientWindows, SupportsCaptive: true}
	case strings.Contains(ua, "macintosh") || strings.Contains(ua, "mac os x"):
		return LegacyClientProfile{Family: ClientMacOS, SupportsCaptive: true}
	case strings.Contains(ua, "linux"):
		return LegacyClientProfile{Family: ClientLinux, SupportsCaptive: true}
	default:
		return LegacyClientProfile{Family: ClientUnknown}
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Contract Verification Cogs & Pre-Flight Validation Tripwires )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3B: Universal OS & Legacy Interpreter ("The Well")
// Module:        cog.interpreter.legacy_bridge
// Mission:       Audit LegacyClientProfile contracts and run a 4-section
//                pre-flight check across Windows, Mac, Linux, iOS, Android, CLI.
// @triggers
//   - Block 4 Audits : AuditLegacyClientProfile(), PreflightLegacyBridge(),
//                      AuditLegacyBridge()
//   - Layer 4 Tests  : tests/functional/interpreter_functional_test.go,
//                      tests/unit/interpreter/legacy_bridge_test.go,
//                      tests/integrated/room_and_well_test.go
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Contract, Input & Live-Instance Verification Cogs
// ------------------------------------------------------------------------------

// AuditLegacyClientProfile verifies a LegacyClientProfile carries a valid
// ClientFamily enum value.
//
// ### For Architects, Operators & Non-Programmers:
// Checks that a classified visitor device record has a legal OS category.
//
// ### For Systems Programmers & Engine Architects:
// O(1) pure contract check returning structured `data.AuditResult`.
func AuditLegacyClientProfile(p LegacyClientProfile) data.AuditResult {
	const check = "legacy_client_profile_contract"
	if !p.Family.IsValid() {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDLegacyBridge, check,
			"LegacyClientProfile carries out-of-range ClientFamily.")
	}
	return data.AuditPass(ModuleIDLegacyBridge, check,
		"LegacyClientProfile family enum verified.")
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Machine Self-Audit & Pre-Flight Validation Tripwires
// ------------------------------------------------------------------------------

// PreflightLegacyBridge runs a 4-section pre-flight self-audit verifying
// (1) Windows/Mac/Linux/iOS/Android/CLI classification, (2) empty/unknown safe
// degradation, (3) mobile-over-desktop token precedence, and (4) audit gate.
//
// ### For Architects, Operators & Non-Programmers:
// Confirms The Well correctly recognizes all major phones and computers without
// mislabeling mobile browsers as desktops.
//
// ### For Systems Programmers & Engine Architects:
// Stack-pure 4-section pre-flight gate returning structured `data.AuditResult`.
func PreflightLegacyBridge() data.AuditResult {
	const check = "legacy_bridge_preflight_4section"

	// Micro Check 1 (Always Pass): Standard desktop & CLI clients.
	if DetectLegacyClient("Mozilla/5.0 (Windows NT 10.0)").Family != ClientWindows ||
		DetectLegacyClient("Mozilla/5.0 (Macintosh; Intel Mac OS X 14_0)").Family != ClientMacOS ||
		DetectLegacyClient("Mozilla/5.0 (X11; Linux x86_64)").Family != ClientLinux ||
		!DetectLegacyClient("curl/8.5.0").PrefersPlain {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDLegacyBridge, check,
			"Sec 1 desktop/CLI client classification failed.")
	}

	// Micro Check 2 (Never Fail): Blank & unrecognizable User-Agent degrade safely.
	if DetectLegacyClient("").Family != ClientUnknown ||
		DetectLegacyClient("ExoticToaster/0.1").Family != ClientUnknown {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDLegacyBridge, check,
			"Sec 2 unknown User-Agent did not degrade to ClientUnknown.")
	}

	// Micro Check 3 (Invariants): iPhone containing "Mac OS X" & Android containing "Linux".
	if DetectLegacyClient("Mozilla/5.0 (iPhone; CPU iPhone OS 17 like Mac OS X)").Family != ClientIOS ||
		DetectLegacyClient("Mozilla/5.0 (Linux; Android 14; Pixel 8)").Family != ClientAndroid {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDLegacyBridge, check,
			"Sec 3 mobile-over-desktop precedence failed.")
	}

	// Micro Check 4 (Emergence): Profile audit gate rejects invalid enum.
	if AuditLegacyClientProfile(LegacyClientProfile{Family: ClientFamily(99)}).IsPass() {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDLegacyBridge, check,
			"Sec 4 AuditLegacyClientProfile accepted invalid enum.")
	}
	return data.AuditPass(ModuleIDLegacyBridge, check,
		"LegacyBridge passed all 4 micro pre-flight gates.")
}

// AuditLegacyBridge returns true iff PreflightLegacyBridge() passes cleanly.
//
// ### For Architects, Operators & Non-Programmers:
// Quick pass/fail shortcut confirming multi-OS client bridging is healthy.
//
// ### For Systems Programmers & Engine Architects:
// Convenience boolean bridge delegating to `PreflightLegacyBridge().IsPass()`.
func AuditLegacyBridge() bool {
	return PreflightLegacyBridge().IsPass()
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK / EXECUTABLE PROCEDURES & LIFECYCLE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3B: Universal OS & Legacy Interpreter ("The Well")
// Module:        cog.interpreter.legacy_bridge
// Mission:       Expose epilogue readiness, 60-col bridge cards, and declare
//                stateless lifecycle compliance.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Executable Procedures, Inspection, Readiness & Cards
// ------------------------------------------------------------------------------

// IsLegacyBridgeReady reports whether the multi-OS LegacyBridge passed preflight.
//
// ### For Architects, Operators & Non-Programmers:
// Epilogue confirmation that universal OS recognition is ready.
//
// ### For Systems Programmers & Engine Architects:
// Matches OmniCode's Block 5 epilogue readiness predicate (`is_*_ready`).
func IsLegacyBridgeReady() bool {
	return PreflightLegacyBridge().IsPass()
}

// ExecuteBridgeDumpCard renders a 60-column inner-width LegacyBridge card.
//
// ### For Architects, Operators & Non-Programmers:
// Summarizes the supported visiting device families and pre-flight status.
//
// ### For Systems Programmers & Engine Architects:
// Non-mutating diagnostic card builder invoking `PreflightLegacyBridge()`.
func ExecuteBridgeDumpCard() string {
	pf := PreflightLegacyBridge()
	return fmt.Sprintf(
		"+------------------------------------------------------------+\n"+
			"| %-58s |\n"+
			"+------------------------------------------------------------+\n"+
			"| Module:    %-47s |\n"+
			"| Supported: %-47s |\n"+
			"| Isolation: %-47s |\n"+
			"| Preflight: %-47s |\n"+
			"+------------------------------------------------------------+",
		"FAITHNET UNIVERSAL LEGACY OS BRIDGE CARD",
		ModuleIDLegacyBridge,
		"Windows, macOS, Linux, Android, iOS, CLI",
		"Data-Driven Layer 3B (Zero Core OS Coupling)",
		pf.Severity.String(),
	)
}

// ------------------------------------------------------------------------------
// Sub-Assembly 5.2: Subsystem Lifecycle Initialization & Teardown
// ------------------------------------------------------------------------------
// Stateless Exemption Notice: None. (Stateless module / pure stack-allocated execution).
// ------------------------------------------------------------------------------

// End of interpreter/legacy_bridge.go
