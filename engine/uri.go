// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          engine/uri.go
// @id            cog.engine.uri
// @layer         Layer 1: Pure Engine Core (FNP/1 Protocol & faithnet:// URI)
// @archetype     pure_engine | protocol_uri_specification
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Sovereign Adaptive Networking Substrate
// @role          Canonical FNP/1 Port 7777 & faithnet:// URI Parser/Formatter
// @maturity      Rev 5 Baseline | Wave 15 Official FaithNet Protocol & URI
// @chassis       zero-external-deps | Pure stdlib string/hex validation (no net/os)
// @scale         O(N_uri) bounded parsing (MaxFaithNetURILen = 256 bytes)
// @seams         in: data (NodeID) | exports: FaithNetURI, ParseFaithNetURI
// @invariants    1. Downward Dependency Law: ZERO net, os, or external imports.
//                2. Canonical protocol: FNP/1 on Port 7777 (UDP Rung 3 / TCP Rung 4);
//                   Legacy Well web bridge on Port 8077.
//                3. Strict anti-injection guard: rejects CR/LF, whitespace, NUL,
//                   and out-of-range ports (only 1..65535 permitted).
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Formalizes FaithNet's canonical protocol identity (FNP/1), default wire
//   port (7777), and human-readable `faithnet://` URI scheme inside Layer 1
//   without depending on OS networking libraries.
//
//   ### For Architects, Operators & Non-Programmers:
//   Just as web pages use `https://` on port 443, FaithNet nodes speak the
//   FaithNet Protocol (`FNP/1`) on port `7777` and can be addressed cleanly
//   using links like `faithnet://home-anchor/well` or
//   `faithnet://0xFA17000000000001/ping`.
//
//   ### For Systems Programmers & Engine Architects:
//   Pure bounded parser and canonical formatter for `faithnet://<host>[:port][/action[/target]]`,
//   recognizing sovereign 64-bit `0xFA17...` NodeIDs, DNS/LAN labels, and IPv4
//   hosts while rejecting control characters and malformed port specs.
// ==============================================================================

package engine

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Module Anchor, Protocol Constants, Public URI Ontology & O(1) Predicates )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 1: Pure Engine Core
// Module:        cog.engine.uri
// Mission:       Define canonical FNP/1 port 7777 and faithnet:// URI parsing.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// --- Tier 2: External Library Dependencies (None — 100% Go Stdlib Invariant) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---
import "github.com/CreativeWorkzStudio/FaithNet/data"

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Pre-Conditions, Canonical Protocol Constants & Sentinels
// ------------------------------------------------------------------------------

const (
	// FaithNetProtocolName is the canonical wire protocol identifier.
	FaithNetProtocolName = "FNP/1"

	// FaithNetURIScheme is the canonical URI scheme (`faithnet://`).
	FaithNetURIScheme = "faithnet"

	// FaithNetDefaultPort is the canonical native FNP/1 port (7777/udp & 7777/tcp).
	FaithNetDefaultPort uint16 = 7777

	// FaithNetWellDefaultPort is the legacy HTTP/HTTPS bridge port for The Well.
	FaithNetWellDefaultPort uint16 = 8077

	// FaithNetDefaultHost is the canonical unicast LAN anchor hostname.
	FaithNetDefaultHost = "faithnet.lan"

	// MaxFaithNetURILen bounds URI input length to prevent unbounded scans.
	MaxFaithNetURILen = 256
)

var (
	// ErrURIEmpty signals an empty URI input.
	ErrURIEmpty = errors.New("faithnet_uri: empty URI")
	// ErrURITooLong signals a URI exceeding MaxFaithNetURILen (256 bytes).
	ErrURITooLong = errors.New("faithnet_uri: exceeds 256-byte limit")
	// ErrURIUnsafeChars signals control characters, NUL, or whitespace inside URI.
	ErrURIUnsafeChars = errors.New("faithnet_uri: contains forbidden control or whitespace chars")
	// ErrURIBadScheme signals a scheme other than `faithnet://`.
	ErrURIBadScheme = errors.New("faithnet_uri: scheme must be faithnet://")
	// ErrURIMissingHost signals an empty host/authority component.
	ErrURIMissingHost = errors.New("faithnet_uri: missing host or sovereign NodeID")
	// ErrURIBadPort signals a non-numeric or out-of-range port (must be 1..65535).
	ErrURIBadPort = errors.New("faithnet_uri: port must be in 1..65535")
	// ErrURIBadAction signals an unsupported or malformed action segment.
	ErrURIBadAction = errors.New("faithnet_uri: invalid action path segment")
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.2: Operational Logical Anchor
// ------------------------------------------------------------------------------

const (
	// ModuleIDFaithNetURI is the movement-resilient logical ID for cog.engine.uri.
	ModuleIDFaithNetURI = "cog.engine.uri"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.3: Public Protocol URI Ontology
// ------------------------------------------------------------------------------

// FaithNetURI represents a parsed, validated `faithnet://` address.
type FaithNetURI struct {
	Scheme       string      // Always "faithnet"
	Host         string      // Hostname, IPv4, or canonical hex NodeID
	NodeID       data.NodeID // Populated when Host is a valid 0x... NodeID
	Port         uint16      // Defaults to FaithNetDefaultPort (7777)
	ExplicitPort bool        // True when caller explicitly wrote :port
	Action       string      // Optional verb: "well", "status", "ping", "pair", "say", or ""
	Target       string      // Optional trailing path argument
}

// ------------------------------------------------------------------------------
// Sub-Assembly 2.4: Pure O(1) Predicates
// ------------------------------------------------------------------------------

// HasNodeID reports whether the URI directly addresses a 64-bit sovereign NodeID.
func (u FaithNetURI) HasNodeID() bool {
	return u.NodeID.IsValid()
}

// IsWellAction reports whether the URI targets The Well portal surface.
func (u FaithNetURI) IsWellAction() bool {
	return u.Action == "" || u.Action == "well" || u.Action == "status" || u.Action == "pair"
}

// ==============================================================================
// BLOCK 3: BODY BLOCK
// ( Pure faithnet:// Parser, Formatter & Legacy Well URL Translator )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 1: Pure Engine Core
// Module:        cog.engine.uri
// Mission:       Parse and format canonical FNP/1 faithnet:// links safely.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: Canonical faithnet:// Parser
// ------------------------------------------------------------------------------

// ParseFaithNetURI validates and parses a canonical `faithnet://` URI string.
//
// Supported forms:
//   - `faithnet://faithnet.lan/well`
//   - `faithnet://home-anchor/status`
//   - `faithnet://0xFA17000000000001:7777/ping`
//   - `faithnet://10.77.77.1/pair`
//
// ### For Architects, Operators & Non-Programmers:
// Reads a `faithnet://` address and breaks it cleanly into destination node,
// port (defaulting to FaithNet's canonical port 7777), and requested action.
//
// ### For Systems Programmers & Engine Architects:
// Pure Layer 1 parser (zero net/os imports) with strict control-char rejection
// and automatic 64-bit hex NodeID recognition.
func ParseFaithNetURI(raw string) (FaithNetURI, error) {
	if raw == "" {
		return FaithNetURI{}, ErrURIEmpty
	}
	if len(raw) > MaxFaithNetURILen {
		return FaithNetURI{}, ErrURITooLong
	}
	for i := 0; i < len(raw); i++ {
		c := raw[i]
		if c <= 0x20 || c >= 0x7F {
			return FaithNetURI{}, ErrURIUnsafeChars
		}
	}

	const prefix = FaithNetURIScheme + "://"
	if len(raw) < len(prefix) || !strings.EqualFold(raw[:len(prefix)], prefix) {
		return FaithNetURI{}, ErrURIBadScheme
	}

	rest := raw[len(prefix):]
	authority := rest
	pathPart := ""
	if slash := strings.IndexByte(rest, '/'); slash >= 0 {
		authority = rest[:slash]
		pathPart = rest[slash+1:]
	}
	if authority == "" {
		return FaithNetURI{}, ErrURIMissingHost
	}

	host := authority
	port := FaithNetDefaultPort
	explicitPort := false
	if colon := strings.LastIndexByte(authority, ':'); colon >= 0 {
		host = authority[:colon]
		portStr := authority[colon+1:]
		if host == "" || portStr == "" {
			return FaithNetURI{}, ErrURIBadPort
		}
		p, err := strconv.Atoi(portStr)
		if err != nil || p < 1 || p > 65535 {
			return FaithNetURI{}, ErrURIBadPort
		}
		port = uint16(p)
		explicitPort = true
	}
	if host == "" {
		return FaithNetURI{}, ErrURIMissingHost
	}

	// Recognize sovereign 64-bit NodeID (e.g., 0xFA17000000000001).
	var nodeID data.NodeID
	if strings.HasPrefix(host, "0x") || strings.HasPrefix(host, "0X") {
		if parsed, err := strconv.ParseUint(host[2:], 16, 64); err == nil {
			cand := data.NodeID(parsed)
			if cand.IsValid() {
				nodeID = cand
				host = cand.String()
			}
		}
	}

	action := ""
	target := ""
	if pathPart != "" {
		if slash := strings.IndexByte(pathPart, '/'); slash >= 0 {
			action = strings.ToLower(pathPart[:slash])
			target = pathPart[slash+1:]
		} else {
			action = strings.ToLower(pathPart)
		}
		switch action {
		case "well", "status", "ping", "pair", "say", "ca":
			// Canonical FNP/1 URI actions.
		default:
			return FaithNetURI{}, ErrURIBadAction
		}
	}

	return FaithNetURI{
		Scheme:       FaithNetURIScheme,
		Host:         host,
		NodeID:       nodeID,
		Port:         port,
		ExplicitPort: explicitPort,
		Action:       action,
		Target:       target,
	}, nil
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: Canonical Formatters & Legacy Well Bridge URL
// ------------------------------------------------------------------------------

// String renders the canonical `faithnet://` URI string.
func (u FaithNetURI) String() string {
	host := u.Host
	if host == "" {
		host = FaithNetDefaultHost
	}
	port := u.Port
	if port == 0 {
		port = FaithNetDefaultPort
	}
	var sb strings.Builder
	sb.WriteString(FaithNetURIScheme)
	sb.WriteString("://")
	sb.WriteString(host)
	if u.ExplicitPort || port != FaithNetDefaultPort {
		sb.WriteByte(':')
		sb.WriteString(strconv.Itoa(int(port)))
	}
	if u.Action != "" {
		sb.WriteByte('/')
		sb.WriteString(u.Action)
		if u.Target != "" {
			sb.WriteByte('/')
			sb.WriteString(u.Target)
		}
	}
	return sb.String()
}

// WellBridgeURL projects a `faithnet://` URI into its corresponding legacy
// HTTP/HTTPS URL for The Well on port 8077 (or FaithNetDefaultHost if host is
// an abstract NodeID).
//
// ### For Architects, Operators & Non-Programmers:
// Translates a native `faithnet://` click into the right local Well address
// (`https://faithnet.lan:8077/...`) so everyday web browsers open it cleanly.
//
// ### For Systems Programmers & Engine Architects:
// Deterministic pure translator mapping action verbs (`well`, `status`, `pair`,
// `ca`) to canonical Well endpoints without performing any network I/O.
func (u FaithNetURI) WellBridgeURL(useTLS bool) string {
	proto := "http"
	if useTLS {
		proto = "https"
	}
	host := u.Host
	if host == "" || u.HasNodeID() || strings.EqualFold(host, "home-anchor") {
		host = FaithNetDefaultHost
	}
	wellPath := "/"
	switch u.Action {
	case "status":
		wellPath = "/api/status"
	case "pair":
		wellPath = "/pair.sh"
	case "ca":
		wellPath = "/faithnet-root-ca.crt"
	}
	return fmt.Sprintf("%s://%s:%d%s", proto, host, FaithNetWellDefaultPort, wellPath)
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Contract Verification Cogs & Pre-Flight Validation Tripwires )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 1: Pure Engine Core
// Module:        cog.engine.uri
// Mission:       4-section self-audit for canonical FNP/1 faithnet:// URIs.
// ==============================================================================

// PreflightFaithNetURI executes a 4-section pre-flight audit proving:
// (1) nominal parsing & default port 7777,
// (2) rejection of bad schemes, bad ports & CRLF injection,
// (3) 256-byte URI length ceiling enforcement, and
// (4) NodeID recognition + WellBridgeURL emergence.
//
// @triggers tests/unit/engine/uri_test.go
func PreflightFaithNetURI() data.AuditResult {
	const check = "faithnet_uri_preflight_4section"

	// Micro Check 1 (Always Pass): Nominal faithnet:// URI defaults to port 7777.
	u, err := ParseFaithNetURI("faithnet://faithnet.lan/well")
	if err != nil || u.Port != FaithNetDefaultPort || u.Action != "well" {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDFaithNetURI, check,
			fmt.Sprintf("Sec 1 nominal parse failed: u=%+v err=%v.", u, err))
	}

	// Micro Check 2 (Never Fail): Wrong scheme, port 0, and CRLF rejected.
	if _, err := ParseFaithNetURI("http://faithnet.lan/well"); err == nil {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDFaithNetURI, check,
			"Sec 2 non-faithnet scheme bypassed ParseFaithNetURI.")
	}
	if _, err := ParseFaithNetURI("faithnet://faithnet.lan:0/well"); err == nil {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDFaithNetURI, check,
			"Sec 2 port 0 bypassed ParseFaithNetURI.")
	}
	if _, err := ParseFaithNetURI("faithnet://faithnet.lan/well\r\nX-Bad:1"); err == nil {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDFaithNetURI, check,
			"Sec 2 CRLF injection bypassed ParseFaithNetURI.")
	}

	// Micro Check 3 (Invariants & Limits): >256-byte URI rejected cleanly.
	oversized := "faithnet://faithnet.lan/say/" + strings.Repeat("a", MaxFaithNetURILen)
	if _, err := ParseFaithNetURI(oversized); !errors.Is(err, ErrURITooLong) {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDFaithNetURI, check,
			"Sec 3 oversized URI did not return ErrURITooLong.")
	}

	// Micro Check 4 (Emergence): Sovereign NodeID & WellBridgeURL roundtrip.
	nu, err := ParseFaithNetURI("faithnet://0xFA17000000000001:7777/status")
	if err != nil || !nu.HasNodeID() || nu.WellBridgeURL(true) != "https://faithnet.lan:8077/api/status" {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDFaithNetURI, check,
			fmt.Sprintf("Sec 4 NodeID/WellBridge emergence failed: %+v (%v).", nu, err))
	}

	return data.AuditPass(ModuleIDFaithNetURI, check,
		"FNP/1 faithnet:// URI parser passed all 4 micro gates (Port 7777 & anti-injection).")
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Epilogue Readiness Handshake & 60-Column Protocol Spec Card )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 1: Pure Engine Core
// Module:        cog.engine.uri
// Mission:       Expose readiness handshake and 60-column FNP/1 spec card.
// ==============================================================================

// IsFaithNetURIReady verifies Layer 1 FNP/1 URI parser readiness.
func IsFaithNetURIReady() bool {
	return PreflightFaithNetURI().IsPass()
}

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: 60-Column Telemetry Card
// ------------------------------------------------------------------------------

// ExecuteProtocolSpecDumpCard renders a 60-column FNP/1 specification card.
func ExecuteProtocolSpecDumpCard() string {
	return fmt.Sprintf(
		"+------------------------------------------------------------+\n"+
			"| FAITHNET PROTOCOL SPECIFICATION CARD (FNP/1)               |\n"+
			"+------------------------------------------------------------+\n"+
			"| Protocol:   %-46s |\n"+
			"| URI Scheme: faithnet://<node-or-host>[:7777][/action]      |\n"+
			"| Wire Port:  7777/udp (Rung 3 LAN) & 7777/tcp (Rung 4 WAN)  |\n"+
			"| Well Port:  8077/tcp (Dual-Mode HTTP/HTTPS Legacy Bridge)  |\n"+
			"+------------------------------------------------------------+",
		FaithNetProtocolName+" ("+ModuleIDFaithNetURI+")",
	)
}

// End of engine/uri.go
