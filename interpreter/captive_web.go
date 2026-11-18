// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          interpreter/captive_web.go
// @id            cog.interpreter.captive_web
// @layer         Layer 3B: Universal OS & Legacy Interpreter ("The Well")
// @archetype     operational | stateful
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Zero-Install HTML Captive Portal ("The Well")
// @role          Local Offline HTTP Portal for Any Modern Phone or Computer
// @maturity      Rev 5 Baseline | Wave 4 The Room & The Well
// @chassis       bounded-heap | Fixed [32]BulletinEntry ring + stdlib net/http
// @scale         O(1) bounded bulletin ring; zero external CDN/JS/CSS deps
// @seams         in: data, orchestrator.Dispatcher, interpreter.Translator
// @invariants    1. Stratum Baseline: Downward Dependency Law, Rev 5 5-Block.
//                2. Zero external assets: 100% self-contained inline HTML/CSS
//                   working offline with zero internet access.
//                3. Every posted browser message passes Genesis 3 Airlock
//                   (`Translator.WrapForeignMessage`); hostile input -> HTTP 403.
//                4. Bounded [32]BulletinEntry ring prevents memory exhaustion.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Implements "The Well": FaithNet's zero-install captive HTML portal allowing
//   Windows, macOS, Linux, iOS, and Android devices joining a local room beacon
//   to view local sovereign bulletins, inspect active survival rung telemetry,
//   and submit airlock-screened messages using nothing but their built-in web
//   browser.
//
//   ### For Architects, Operators & Non-Programmers:
//   When someone walks into the room with an ordinary phone or laptop, they do
//   not need to install any app—opening any browser loads The Well cleanly off
//   this local FaithNet node, even if the entire commercial internet is gone.
//
//   ### For Systems Programmers & Engine Architects:
//   - Ownership & Reentrancy: Exposes `Handler() http.Handler` for direct
//     `httptest` verification plus `Start()`/`Close()` lifecycle over `net/http`.
//   - Security Boundary: All HTML output escaped via `html.EscapeString`; all
//     inbound posts bounded via `http.MaxBytesReader` + Genesis 3 Airlock.
// ==============================================================================

package interpreter

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Portal Constants, Bulletin Ring Ontology & O(1) Accessors )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3B: Universal OS & Legacy Interpreter ("The Well")
// Module:        cog.interpreter.captive_web
// Mission:       Define bounded 32-bulletin ring, portal configuration, and
//                stdlib HTTP handler wiring for The Well captive portal.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"bufio"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"html"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// --- Tier 2: External Library Dependencies (None — 100% Go Stdlib Invariant) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---
import (
	"github.com/CreativeWorkzStudio/FaithNet/data"
	"github.com/CreativeWorkzStudio/FaithNet/engine"
	"github.com/CreativeWorkzStudio/FaithNet/orchestrator"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Pre-Conditions, Physical Bounds & HTTP Safety Limits
// ------------------------------------------------------------------------------

const (
	// MaxWellBulletins is the power-of-two capacity of The Well's bulletin ring.
	MaxWellBulletins = 32

	// MaxWellRequestBytes caps inbound HTTP body reads cleanly at 2 KB.
	MaxWellRequestBytes = 2048

	// DefaultWellExternalNodeID identifies anonymous captive-portal visitors.
	DefaultWellExternalNodeID data.NodeID = 0xEEEE000000000001
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.2: Operational Logical Anchor
// ------------------------------------------------------------------------------

const (
	// ModuleIDWellPortal is the canonical logical ID for The Well portal.
	ModuleIDWellPortal = "cog.interpreter.captive_web"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.3: Public Tagged Ontology (BulletinEntry & WellServer)
// ------------------------------------------------------------------------------

// BulletinEntry stores one sovereign or airlock-quarantined room message.
type BulletinEntry struct {
	Seq      uint32 `json:"seq"`
	ClientOS string `json:"client_os"`
	Truth    string `json:"truth"`
	Rung     string `json:"rung"`
	Message  string `json:"message"`
}

// WellStatusPayload is the JSON status document served at `GET /api/status`.
type WellStatusPayload struct {
	NodeID         string `json:"node_id"`
	Label          string `json:"label"`
	ActiveRung     string `json:"active_rung"`
	PeerCount      int    `json:"peer_count"`
	Bulletins      int    `json:"bulletins"`
	AirlockBlocked uint64 `json:"airlock_blocked"`
}

// WellTLSOptions configures The Well's 3-Tier Adaptive TLS Trust Ladder:
//   - Tier 3 (Public Rollout): CertFile + KeyFile (browser-trusted PEMs).
//   - Tier 2 (Offline Sovereign Root CA): CADir (persistent local Root CA +
//     signed server leaf cert, downloadable at GET /faithnet-root-ca.crt).
//   - Tier 1 (Emergency Floor): Ephemeral in-memory ECDSA P-256 cert when
//     neither Tier 3 nor Tier 2 is supplied.
type WellTLSOptions struct {
	CertFile string // Tier 3: Operator-supplied PEM certificate chain
	KeyFile  string // Tier 3: Operator-supplied PEM private key
	CADir    string // Tier 2: Persistent local Sovereign Root CA directory
}

// WellPairOptions configures zero-config sovereign peer pairing served directly
// by The Well (`/api/pair`, `/pair.sh`, and `/download/faithnet-node`).
type WellPairOptions struct {
	AnchorUDPPort string // Local Rung 3 UDP port peers should join (default "7777")
	BinaryPath    string // Optional compiled faithnet-node path for self-replication
}

// WellPairPayload is the JSON handshake payload served at `GET /api/pair`.
type WellPairPayload struct {
	AnchorID       string `json:"anchor_id"`
	AnchorLabel    string `json:"anchor_label"`
	AnchorUDP      string `json:"anchor_udp"`
	AssignedNodeID string `json:"assigned_node_id"`
	BinaryURL      string `json:"binary_url"`
	PairScriptURL  string `json:"pair_script_url"`
	OneLiner       string `json:"one_liner"`
}

// WellServer hosts FaithNet's zero-install local HTML captive portal.
//
// ### For Architects, Operators & Non-Programmers:
// Runs a lightweight offline webpage right inside a FaithNet node so any nearby
// phone or laptop browser can read local news and post safe messages.
//
// ### For Systems Programmers & Engine Architects:
// Mutex-guarded `[32]BulletinEntry` ring + stdlib `http.ServeMux` / `http.Server`
// with a 3-Tier Adaptive TLS Trust Ladder (`dualProtoListener`).
type WellServer struct {
	dispatcher *orchestrator.Dispatcher
	translator *Translator
	listenAddr string

	mu            sync.RWMutex
	bulletins     [MaxWellBulletins]BulletinEntry
	bulletinHead  uint64
	bulletinCount uint64
	nextSeq       atomic.Uint32
	blockedPosts  atomic.Uint64

	srvMu      sync.Mutex
	server     *http.Server
	listener   net.Listener
	mux        *http.ServeMux
	tlsOpts    WellTLSOptions
	pairOpts   WellPairOptions
	nextPairID atomic.Uint64
	tlsTier    string
	rootCAPEM  []byte
}

// ------------------------------------------------------------------------------
// Sub-Assembly 2.4: Pre-Conditional Helper Cogs & O(1) Accessors
// ------------------------------------------------------------------------------

// Handler returns the stdlib `http.Handler` serving The Well routes directly.
//
// ### For Architects, Operators & Non-Programmers:
// Exposes The Well's web router for live serving or zero-socket lab testing.
//
// ### For Systems Programmers & Engine Architects:
// Direct getter compatible with `httptest.NewServer` and `httptest.NewRecorder`.
func (w *WellServer) Handler() http.Handler {
	return w.mux
}

// ListenAddr returns the active bound address (or configured address if unstarted).
//
// ### For Architects, Operators & Non-Programmers:
// Shows the local address (e.g. `127.0.0.1:8077`) where browsers reach The Well.
//
// ### For Systems Programmers & Engine Architects:
// Returns actual `listener.Addr().String()` once `Start()` binds an OS port.
func (w *WellServer) ListenAddr() string {
	w.srvMu.Lock()
	defer w.srvMu.Unlock()
	if w.listener != nil {
		return w.listener.Addr().String()
	}
	return w.listenAddr
}

// ==============================================================================
// BLOCK 3: BODY BLOCK
// ( Composable HTTP Routes: Home Dashboard, JSON Status, Airlock POST & Captive )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3B: Universal OS & Legacy Interpreter ("The Well")
// Module:        cog.interpreter.captive_web
// Mission:       Serve zero-dependency offline HTML, JSON telemetry, captive
//                portal redirects, and Genesis 3 Airlock-screened message posts.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: Bulletin Ring Ingestion & Snapshot
// ------------------------------------------------------------------------------

// RecordBulletin appends one verified or 0:AWAIT entry into the 32-slot ring.
//
// ### For Architects, Operators & Non-Programmers:
// Posts a clean message onto the local room bulletin board, automatically
// rolling over the oldest item once 32 bulletins are stored.
//
// ### For Systems Programmers & Engine Architects:
// O(1) power-of-two ring insertion under write lock (zero slice growth).
func (w *WellServer) RecordBulletin(entry BulletinEntry) {
	w.mu.Lock()
	idx := w.bulletinHead & (MaxWellBulletins - 1)
	w.bulletins[idx] = entry
	w.bulletinHead++
	if w.bulletinCount < MaxWellBulletins {
		w.bulletinCount++
	}
	w.mu.Unlock()
}

// BulletinsSnapshotInto copies active bulletins chronologically into dst.
//
// ### For Architects, Operators & Non-Programmers:
// Reads the current list of local room messages in order from oldest to newest.
//
// ### For Systems Programmers & Engine Architects:
// Bounded chronological copy from `[MaxWellBulletins]BulletinEntry`.
func (w *WellServer) BulletinsSnapshotInto(dst []BulletinEntry) int {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if w.bulletinCount == 0 || len(dst) == 0 {
		return 0
	}
	total := int(w.bulletinCount)
	start := uint64(0)
	if w.bulletinHead > w.bulletinCount {
		start = w.bulletinHead - w.bulletinCount
	}
	written := 0
	for i := 0; i < total && written < len(dst); i++ {
		idx := (start + uint64(i)) & (MaxWellBulletins - 1)
		dst[written] = w.bulletins[idx]
		written++
	}
	return written
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: HTTP Portal Route Handlers
// ------------------------------------------------------------------------------

func (w *WellServer) handleCaptiveRedirect(rw http.ResponseWriter, r *http.Request) {
	http.Redirect(rw, r, "/", http.StatusFound)
}

func (w *WellServer) handleStatus(rw http.ResponseWriter, _ *http.Request) {
	w.mu.RLock()
	cnt := int(w.bulletinCount)
	w.mu.RUnlock()

	st := WellStatusPayload{
		NodeID:         w.dispatcher.Manifest().ID.String(),
		Label:          w.dispatcher.Manifest().Label,
		ActiveRung:     w.dispatcher.Ladder().ActiveRung().String(),
		PeerCount:      w.dispatcher.Router().ActiveCount(),
		Bulletins:      cnt,
		AirlockBlocked: w.blockedPosts.Load() + w.dispatcher.StatsSnapshot().AirlockBlocked,
	}
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(rw).Encode(st)
}

func (w *WellServer) handlePostMessage(rw http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(rw, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(rw, r.Body, MaxWellRequestBytes)

	var rawMsg string
	ct := strings.ToLower(r.Header.Get("Content-Type"))
	if strings.Contains(ct, "application/x-www-form-urlencoded") {
		if err := r.ParseForm(); err == nil {
			rawMsg = r.FormValue("message")
		}
	} else {
		body, err := io.ReadAll(r.Body)
		if err == nil {
			rawMsg = string(body)
		}
	}

	seq := w.nextSeq.Add(1)
	rung := w.dispatcher.Ladder().ActiveRung()
	if !rung.IsValid() {
		rung = data.Rung3Local
	}

	frame, verdict, err := w.translator.WrapForeignMessage(
		DefaultWellExternalNodeID,
		w.dispatcher.Manifest().ID,
		rung,
		seq,
		[]byte(rawMsg),
	)
	if err != nil {
		w.blockedPosts.Add(1)
		rw.Header().Set("Content-Type", "application/json; charset=utf-8")
		rw.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(rw).Encode(map[string]string{
			"status":  "BLOCKED",
			"truth":   verdict.State.String(),
			"reason":  verdict.Reason.String(),
			"details": err.Error(),
		})
		return
	}

	clientProf := DetectLegacyClient(r.Header.Get("User-Agent"))
	entry := BulletinEntry{
		Seq:      frame.SeqID,
		ClientOS: clientProf.Family.String(),
		Truth:    frame.Truth.String(),
		Rung:     frame.Rung.String(),
		Message:  string(frame.PayloadSlice()),
	}
	w.RecordBulletin(entry)

	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(rw).Encode(map[string]any{
		"status": "QUARANTINED_AWAIT",
		"truth":  verdict.State.String(),
		"entry":  entry,
	})
}

// SetTLSOptions configures The Well's 3-Tier Adaptive TLS Trust Ladder prior
// to calling Start().
//
// ### For Architects, Operators & Non-Programmers:
// Lets an operator supply public rollout certificates (zero browser warnings)
// or a persistent local Sovereign Root CA directory (offline green padlock).
//
// ### For Systems Programmers & Engine Architects:
// Mutex-guarded configuration setter consumed by `WellServer.Start()`.
func (w *WellServer) SetTLSOptions(opts WellTLSOptions) {
	w.srvMu.Lock()
	w.tlsOpts = opts
	w.srvMu.Unlock()
}

// SetPairOptions enables one-command sovereign peer pairing and self-replicating
// binary download directly from The Well (`/api/pair`, `/pair.sh`, `/download/faithnet-node`).
//
// ### For Architects, Operators & Non-Programmers:
// Lets a running FaithNet laptop hand a copy of itself and auto-generated join
// configuration to any nearby laptop over Wi-Fi with one copy-paste command.
//
// ### For Systems Programmers & Engine Architects:
// Mutex-guarded pairing configuration setter with atomic unique NodeID allocation.
func (w *WellServer) SetPairOptions(opts WellPairOptions) {
	if opts.AnchorUDPPort == "" {
		opts.AnchorUDPPort = "7777"
	}
	w.srvMu.Lock()
	w.pairOpts = opts
	w.srvMu.Unlock()
}

func (w *WellServer) allocatePeerNodeID() data.NodeID {
	selfID := uint64(w.dispatcher.Manifest().ID)
	for {
		n := w.nextPairID.Add(1)
		candidate := uint64(0xFA17000000000001) + n
		if candidate != selfID {
			return data.NodeID(candidate)
		}
	}
}

func (w *WellServer) buildPairPayload(r *http.Request) WellPairPayload {
	w.srvMu.Lock()
	pOpts := w.pairOpts
	w.srvMu.Unlock()
	udpPort := pOpts.AnchorUDPPort
	if udpPort == "" {
		udpPort = "7777"
	}

	hostPort := r.Host
	if hostPort == "" {
		hostPort = w.ListenAddr()
	}
	hostOnly := hostPort
	if h, _, err := net.SplitHostPort(hostPort); err == nil && h != "" {
		hostOnly = h
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s", scheme, hostPort)
	m := w.dispatcher.Manifest()
	assigned := w.allocatePeerNodeID()

	pairURL := baseURL + "/pair.sh"
	oneLiner := fmt.Sprintf("curl -fsSL %s | sh", pairURL)
	if scheme == "https" {
		oneLiner = fmt.Sprintf("curl -fsSLk %s | sh", pairURL)
	}
	return WellPairPayload{
		AnchorID:       m.ID.String(),
		AnchorLabel:    m.Label,
		AnchorUDP:      net.JoinHostPort(hostOnly, udpPort),
		AssignedNodeID: assigned.String(),
		BinaryURL:      baseURL + "/download/faithnet-node",
		PairScriptURL:  pairURL,
		OneLiner:       oneLiner,
	}
}

func (w *WellServer) handlePairJSON(rw http.ResponseWriter, r *http.Request) {
	payload := w.buildPairPayload(r)
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(rw).Encode(payload)
}

func (w *WellServer) handlePairScript(rw http.ResponseWriter, r *http.Request) {
	p := w.buildPairPayload(r)
	curlFlags := "-fsSL"
	if strings.HasPrefix(p.BinaryURL, "https://") {
		curlFlags = "-fsSLk"
	}
	script := fmt.Sprintf(`#!/bin/sh
set -eu
echo "=== FaithNet Zero-Config Sovereign Node Pairing ==="
echo "Anchor : %s (%s) at %s"
echo "NodeID : %s"
DEST="/tmp/faithnet-node"
echo "[1/2] Downloading self-contained native faithnet-node binary..."
curl %s "%s" -o "${DEST}"
chmod +x "${DEST}"
LABEL="$(hostname 2>/dev/null || echo FaithNet-Peer)"
echo "[2/2] Pairing '${LABEL}' (%s) with Anchor %s..."
if [ -r /dev/tty ]; then
  exec "${DEST}" --id=%s --label="${LABEL}" --udp=0.0.0.0:7777 --join="%s" --join-id=%s --beacon --interactive </dev/tty
else
  exec "${DEST}" --id=%s --label="${LABEL}" --udp=0.0.0.0:7777 --join="%s" --join-id=%s --beacon
fi
`, p.AnchorLabel, p.AnchorID, p.AnchorUDP, p.AssignedNodeID,
		curlFlags, p.BinaryURL,
		p.AssignedNodeID, p.AnchorUDP,
		p.AssignedNodeID, p.AnchorUDP, p.AnchorID,
		p.AssignedNodeID, p.AnchorUDP, p.AnchorID)

	rw.Header().Set("Content-Type", "text/x-shellscript; charset=utf-8")
	_, _ = rw.Write([]byte(script))
}

func (w *WellServer) handleDownloadBinary(rw http.ResponseWriter, r *http.Request) {
	w.srvMu.Lock()
	binPath := w.pairOpts.BinaryPath
	w.srvMu.Unlock()

	if binPath == "" {
		http.Error(rw, "Self-replicating binary download not enabled on this node", http.StatusNotFound)
		return
	}
	if _, err := os.Stat(binPath); err != nil {
		http.Error(rw, "Binary unavailable on host", http.StatusNotFound)
		return
	}
	rw.Header().Set("Content-Type", "application/octet-stream")
	rw.Header().Set("Content-Disposition", `attachment; filename="faithnet-node"`)
	http.ServeFile(rw, r, binPath)
}

// TLSTrustTier reports the active TLS trust tier ("TIER_3_PUBLIC_PEM",
// "TIER_2_SOVEREIGN_CA", "TIER_1_EPHEMERAL", or "UNSTARTED").
//
// ### For Architects, Operators & Non-Programmers:
// Tells the operator which certificate trust level The Well is currently using.
//
// ### For Systems Programmers & Engine Architects:
// Mutex-guarded O(1) string telemetry getter.
func (w *WellServer) TLSTrustTier() string {
	w.srvMu.Lock()
	defer w.srvMu.Unlock()
	if w.tlsTier == "" {
		return "UNSTARTED"
	}
	return w.tlsTier
}

func (w *WellServer) handleRootCADownload(rw http.ResponseWriter, _ *http.Request) {
	w.srvMu.Lock()
	caPEM := make([]byte, len(w.rootCAPEM))
	copy(caPEM, w.rootCAPEM)
	w.srvMu.Unlock()

	if len(caPEM) == 0 {
		http.Error(rw, "Sovereign Local Root CA not enabled on this node (start with --tls-ca-dir)", http.StatusNotFound)
		return
	}
	rw.Header().Set("Content-Type", "application/x-x509-ca-cert")
	rw.Header().Set("Content-Disposition", `attachment; filename="faithnet-root-ca.crt"`)
	_, _ = rw.Write(caPEM)
}

func (w *WellServer) handleHomeHTML(rw http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(rw, r)
		return
	}
	var entries [MaxWellBulletins]BulletinEntry
	n := w.BulletinsSnapshotInto(entries[:])

	var sb strings.Builder
	sb.WriteString("<!doctype html><html lang=\"en\"><head><meta charset=\"utf-8\">")
	sb.WriteString("<title>FaithNet — The Well</title>")
	sb.WriteString("<style>body{font-family:monospace;max-width:720px;margin:2rem auto;padding:0 1rem;background:#111;color:#eee;}h1{color:#ffd166;}li{margin:.4rem 0;}input,button{font-family:monospace;padding:.4rem;}</style>")
	sb.WriteString("</head><body>")
	sb.WriteString("<h1>FaithNet :: The Well (Sovereign Local Portal)</h1>")
	sb.WriteString(fmt.Sprintf("<p>Node: <strong>%s</strong> (%s) | Active Rung: <strong>%s</strong> | Peers: <strong>%d</strong></p>",
		html.EscapeString(w.dispatcher.Manifest().Label),
		html.EscapeString(w.dispatcher.Manifest().ID.String()),
		html.EscapeString(w.dispatcher.Ladder().ActiveRung().String()),
		w.dispatcher.Router().ActiveCount(),
	))
	sb.WriteString("<form method=\"POST\" action=\"/api/message\">")
	sb.WriteString("<input type=\"text\" name=\"message\" maxlength=\"480\" placeholder=\"Share local bulletin (screened by Genesis 3 Airlock)...\" style=\"width:75%\"> ")
	sb.WriteString("<button type=\"submit\">Post</button></form>")
	sb.WriteString("<h2>Local Bulletins</h2><ul>")
	for i := n - 1; i >= 0; i-- {
		e := entries[i]
		sb.WriteString(fmt.Sprintf("<li>[#%d | %s | %s] %s</li>",
			e.Seq,
			html.EscapeString(e.Truth),
			html.EscapeString(e.ClientOS),
			html.EscapeString(e.Message),
		))
	}
	sb.WriteString("</ul>")
	w.srvMu.Lock()
	hasCA := len(w.rootCAPEM) > 0
	hasBin := w.pairOpts.BinaryPath != ""
	tier := w.tlsTier
	w.srvMu.Unlock()
	if hasBin {
		hostPort := r.Host
		if hostPort == "" {
			hostPort = w.ListenAddr()
		}
		sb.WriteString(`<hr><h3>Pair Another Linux Device (Zero-Config Auto-Join)</h3><pre>curl -fsSL http://`)
		sb.WriteString(html.EscapeString(hostPort))
		sb.WriteString(`/pair.sh | sh</pre>`)
	}
	if hasCA {
		sb.WriteString(`<hr><p><small>TLS Trust: `)
		sb.WriteString(html.EscapeString(tier))
		sb.WriteString(` | Offline Green Padlock: <a href="/faithnet-root-ca.crt">Download FaithNet Sovereign Root CA (.crt)</a></small></p>`)
	}
	sb.WriteString("</body></html>")

	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = rw.Write([]byte(sb.String()))
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Contract Verification Cogs & Pre-Flight Validation Tripwires )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3B: Universal OS & Legacy Interpreter ("The Well")
// Module:        cog.interpreter.captive_web
// Mission:       Audit live WellServer invariants and run an in-memory httptest
//                4-section pre-flight verifying HTML, JSON, 0:AWAIT & XSS 403.
// @triggers
//   - Block 4 Audits : (*WellServer).AuditState(), PreflightWellServer(),
//                      AuditWellServer()
//   - Layer 4 Tests  : tests/functional/interpreter_functional_test.go,
//                      tests/unit/interpreter/captive_web_test.go,
//                      tests/integrated/room_and_well_test.go,
//                      tests/system/airlock_hostility_system_test.go
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Contract, Input & Live-Instance Verification Cogs
// ------------------------------------------------------------------------------

// AuditState inspects a live WellServer non-destructively to confirm attached
// Dispatcher, armed Translator airlock, HTTP router, and 32-slot ring bounds.
//
// ### For Architects, Operators & Non-Programmers:
// Checks that The Well's offline portal is properly connected to the node brain
// and its security airlock before welcoming visitors.
//
// ### For Systems Programmers & Engine Architects:
// Non-mutating live probe returning structured `data.AuditResult`.
func (w *WellServer) AuditState() data.AuditResult {
	const check = "interpreter_well_live_contract"

	if w == nil || w.dispatcher == nil || w.translator == nil || w.mux == nil {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDWellPortal, check,
			"WellServer missing dispatcher, translator, or HTTP handler mux.")
	}
	if tr := w.translator.AuditState(); !tr.IsPass() {
		return tr
	}
	w.mu.RLock()
	cnt := w.bulletinCount
	w.mu.RUnlock()
	if cnt > MaxWellBulletins {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDWellPortal, check,
			"WellServer bulletin count breached MaxWellBulletins (32).")
	}
	return data.AuditPass(ModuleIDWellPortal, check,
		"Live WellServer dispatcher, airlock translator, and ring verified.")
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Machine Self-Audit & Pre-Flight Validation Tripwires
// ------------------------------------------------------------------------------

// PreflightWellServer executes a 4-section zero-socket `httptest` pre-flight
// covering (1) GET / HTML & GET /api/status JSON, (2) POST XSS -> HTTP 403,
// (3) clean POST -> HTTP 202 + 0:AWAIT bulletin storage, and (4) captive redirect.
//
// ### For Architects, Operators & Non-Programmers:
// Rehearses visitor requests against The Well inside memory to confirm normal
// browsers load the portal cleanly while hostile scripts get 403-blocked.
//
// ### For Systems Programmers & Engine Architects:
// Zero-port `httptest.NewRecorder()` 4-section pre-flight returning `AuditResult`.
func PreflightWellServer() data.AuditResult {
	const check = "interpreter_well_preflight_4section"

	disp, err := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(0x01, "well-preflight"), engine.NewMockClock(0),
	)
	if err != nil {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDWellPortal, check,
			"Sec 1 failed creating ephemeral Dispatcher.")
	}
	defer func() { _ = disp.Close() }()

	well := NewWellServer(disp, "127.0.0.1:0")
	if live := well.AuditState(); !live.IsPass() {
		return live
	}

	// Micro Check 1 (Always Pass): GET / renders HTML & GET /api/status returns JSON.
	rec1 := httptest.NewRecorder()
	well.Handler().ServeHTTP(rec1, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec1.Code != http.StatusOK || !strings.Contains(rec1.Body.String(), "The Well") {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDWellPortal, check,
			"Sec 1 GET / failed to render offline HTML portal.")
	}
	recStatus := httptest.NewRecorder()
	well.Handler().ServeHTTP(recStatus, httptest.NewRequest(http.MethodGet, "/api/status", nil))
	if recStatus.Code != http.StatusOK || !strings.Contains(recStatus.Body.String(), "well-preflight") {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDWellPortal, check,
			"Sec 1 GET /api/status JSON failed.")
	}

	// Micro Check 2 (Never Fail): Hostile XSS POST blocked with HTTP 403 Forbidden.
	badReq := httptest.NewRequest(http.MethodPost, "/api/message",
		strings.NewReader("<script>steal()</script>"))
	recBad := httptest.NewRecorder()
	well.Handler().ServeHTTP(recBad, badReq)
	if recBad.Code != http.StatusForbidden {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDWellPortal, check,
			"Sec 2 hostile XSS POST bypassed WellServer Genesis 3 Airlock.")
	}

	// Micro Check 3 (Invariants & Limits): Clean POST enters as 202 + 0:AWAIT.
	okReq := httptest.NewRequest(http.MethodPost, "/api/message",
		strings.NewReader("Local power restored on Oak Street"))
	recOK := httptest.NewRecorder()
	well.Handler().ServeHTTP(recOK, okReq)
	if recOK.Code != http.StatusAccepted || !strings.Contains(recOK.Body.String(), "0:AWAIT") {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDWellPortal, check,
			"Sec 3 clean POST did not return 202 QUARANTINED_AWAIT.")
	}

	// Micro Check 4 (Emergence): Captive portal probe redirects to / cleanly.
	recCap := httptest.NewRecorder()
	well.Handler().ServeHTTP(recCap, httptest.NewRequest(http.MethodGet, "/generate_204", nil))
	if recCap.Code != http.StatusFound || recCap.Header().Get("Location") != "/" {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDWellPortal, check,
			"Sec 4 captive portal probe did not redirect to /.")
	}

	return data.AuditPass(ModuleIDWellPortal, check,
		"WellServer passed all 4 micro pre-flight gates (HTML/403/AWAIT/captive).")
}

// AuditWellServer returns true iff PreflightWellServer() passes cleanly.
//
// ### For Architects, Operators & Non-Programmers:
// Quick pass/fail check confirming The Well offline web portal is healthy.
//
// ### For Systems Programmers & Engine Architects:
// Convenience boolean bridge delegating to `PreflightWellServer().IsPass()`.
func AuditWellServer() bool {
	return PreflightWellServer().IsPass()
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK / EXECUTABLE PROCEDURES & LIFECYCLE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3B: Universal OS & Legacy Interpreter ("The Well")
// Module:        cog.interpreter.captive_web
// Mission:       Expose epilogue readiness, 60-col cards, and HTTP server
//                initialization, Start(), and graceful Close() teardown.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Executable Procedures, Inspection, Readiness & Cards
// ------------------------------------------------------------------------------

// IsWellServerReady reports whether The Well portal passed pre-flight checks.
//
// ### For Architects, Operators & Non-Programmers:
// Epilogue confirmation that The Well offline captive portal is ready.
//
// ### For Systems Programmers & Engine Architects:
// Matches OmniCode's Block 5 epilogue readiness predicate (`is_*_ready`).
func IsWellServerReady() bool {
	return PreflightWellServer().IsPass()
}

// ExecuteWellDumpCard renders a 60-column inner-width WellServer status card.
//
// ### For Architects, Operators & Non-Programmers:
// Displays The Well's listen address, stored bulletins, and blocked attacks.
//
// ### For Systems Programmers & Engine Architects:
// Non-mutating diagnostic card builder.
func (w *WellServer) ExecuteWellDumpCard() string {
	w.mu.RLock()
	cnt := w.bulletinCount
	w.mu.RUnlock()
	return fmt.Sprintf(
		"+------------------------------------------------------------+\n"+
			"| %-58s |\n"+
			"+------------------------------------------------------------+\n"+
			"| Module:    %-47s |\n"+
			"| Listen:    %-47s |\n"+
			"| Bulletins: %-47s |\n"+
			"| Blocked:   %-47d |\n"+
			"+------------------------------------------------------------+",
		"FAITHNET THE WELL CAPTIVE PORTAL CARD",
		ModuleIDWellPortal,
		w.ListenAddr(),
		fmt.Sprintf("%d / %d bounded ring", cnt, MaxWellBulletins),
		w.blockedPosts.Load(),
	)
}

// ------------------------------------------------------------------------------
// Sub-Assembly 5.2: Subsystem Lifecycle Initialization & Teardown
// ------------------------------------------------------------------------------

// NewWellServer builds a WellServer attached to disp bound for listenAddr.
//
// ### For Architects, Operators & Non-Programmers:
// Prepares the zero-install offline webpage server wired to this FaithNet node.
//
// ### For Systems Programmers & Engine Architects:
// Registers `/`, `/api/status`, `/api/message`, and OS captive probe routes.
func NewWellServer(disp *orchestrator.Dispatcher, listenAddr string) *WellServer {
	if listenAddr == "" {
		listenAddr = data.DefaultWellListenAddr
	}
	var airlock = disp.Airlock()
	w := &WellServer{
		dispatcher: disp,
		translator: NewTranslator(airlock),
		listenAddr: listenAddr,
		mux:        http.NewServeMux(),
	}
	w.mux.HandleFunc("/", w.handleHomeHTML)
	w.mux.HandleFunc("/api/status", w.handleStatus)
	w.mux.HandleFunc("/api/message", w.handlePostMessage)
	w.mux.HandleFunc("/api/pair", w.handlePairJSON)
	w.mux.HandleFunc("/pair.sh", w.handlePairScript)
	w.mux.HandleFunc("/download/faithnet-node", w.handleDownloadBinary)
	w.mux.HandleFunc("/faithnet-root-ca.crt", w.handleRootCADownload)
	w.mux.HandleFunc("/generate_204", w.handleCaptiveRedirect)
	w.mux.HandleFunc("/hotspot-detect.html", w.handleCaptiveRedirect)
	w.mux.HandleFunc("/connecttest.txt", w.handleCaptiveRedirect)
	return w
}

// Start binds the TCP listener and serves The Well portal in dual-mode
// (accepting BOTH plain HTTP and TLS on the exact same port via first-byte
// 0x16 peeking), resolving TLS credentials across the 3-Tier Adaptive Trust
// Ladder: Tier 3 (Public Rollout PEMs) -> Tier 2 (Persistent Sovereign Local
// Root CA) -> Tier 1 (Ephemeral Emergency ECDSA Floor).
//
// ### For Architects, Operators & Non-Programmers:
// Opens The Well's local webpage so phones and laptops can connect cleanly via
// http:// or https://, using trusted rollout certs when supplied or an offline
// Sovereign Root CA so neighbors see a trusted padlock instead of scary warnings.
//
// ### For Systems Programmers & Engine Architects:
// Resolves Tier 3 -> Tier 2 -> Tier 1 via pure Go stdlib `crypto/x509` +
// `crypto/tls`, wraps `net.Listener` in `dualProtoListener`, and serves live.
func (w *WellServer) Start() error {
	rawLn, err := net.Listen("tcp", w.listenAddr)
	if err != nil {
		return err
	}
	w.srvMu.Lock()
	opts := w.tlsOpts
	w.srvMu.Unlock()

	tlsCfg, tier, caPEM, err := resolveAdaptiveWellTLS(opts)
	var ln net.Listener = rawLn
	if err == nil && tlsCfg != nil {
		ln = newDualProtoListener(rawLn, tlsCfg)
	}
	srv := &http.Server{
		Handler:           w.mux,
		ReadHeaderTimeout: 3 * time.Second,
	}
	w.srvMu.Lock()
	w.listener = ln
	w.server = srv
	w.tlsTier = tier
	w.rootCAPEM = caPEM
	w.srvMu.Unlock()

	go func() {
		_ = srv.Serve(ln)
	}()
	return nil
}

func resolveAdaptiveWellTLS(opts WellTLSOptions) (*tls.Config, string, []byte, error) {
	// Tier 3: Operator-supplied publicly trusted PEM cert & key (Rollout Phase).
	if opts.CertFile != "" && opts.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(opts.CertFile, opts.KeyFile)
		if err != nil {
			return nil, "", nil, err
		}
		return &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
		}, "TIER_3_PUBLIC_PEM", nil, nil
	}

	// Tier 2: Persistent Sovereign Local Root CA + Signed Leaf Cert (100% Offline).
	if opts.CADir != "" {
		cfg, caPEM, err := loadOrCreateSovereignLocalCA(opts.CADir)
		if err == nil {
			return cfg, "TIER_2_SOVEREIGN_CA", caPEM, nil
		}
	}

	// Tier 1: Ephemeral Emergency Floor (zero-config fallback).
	cfg, err := buildOfflineWellTLSConfig()
	if err != nil {
		return nil, "PLAIN_HTTP_ONLY", nil, err
	}
	return cfg, "TIER_1_EPHEMERAL", nil, nil
}

func collectLocalSANs() ([]net.IP, []string) {
	ips := []net.IP{
		net.ParseIP("127.0.0.1"),
		net.ParseIP("::1"),
		net.ParseIP("10.77.77.1"),
	}
	if addrs, err := net.InterfaceAddrs(); err == nil {
		for _, a := range addrs {
			if ipNet, ok := a.(*net.IPNet); ok && ipNet.IP != nil {
				ips = append(ips, ipNet.IP)
			}
		}
	}
	dnsNames := []string{"localhost", "faithnet.lan", "faithnet.home.arpa", "faithnet.local"}
	if host, err := os.Hostname(); err == nil && host != "" {
		dnsNames = append(dnsNames, host, host+".local")
	}
	return ips, dnsNames
}

// loadOrCreateSovereignLocalCA persists a FaithNet Sovereign Root CA (ca.crt +
// ca.key) in caDir and issues a fresh signed leaf certificate covering all
// current local IPs and hostnames so trusted community devices get an offline
// green padlock across reboots.
func loadOrCreateSovereignLocalCA(caDir string) (*tls.Config, []byte, error) {
	if err := os.MkdirAll(caDir, 0700); err != nil {
		return nil, nil, err
	}
	caCertPath := filepath.Join(caDir, "ca.crt")
	caKeyPath := filepath.Join(caDir, "ca.key")

	var caCert *x509.Certificate
	var caKey *ecdsa.PrivateKey
	var caPEM []byte

	if certBytes, err := os.ReadFile(caCertPath); err == nil {
		if keyBytes, kErr := os.ReadFile(caKeyPath); kErr == nil {
			cb, _ := pem.Decode(certBytes)
			kb, _ := pem.Decode(keyBytes)
			if cb != nil && kb != nil {
				if parsedCert, pErr := x509.ParseCertificate(cb.Bytes); pErr == nil {
					if parsedKey, pkErr := x509.ParseECPrivateKey(kb.Bytes); pkErr == nil {
						caCert = parsedCert
						caKey = parsedKey
						caPEM = certBytes
					}
				}
			}
		}
	}

	if caCert == nil || caKey == nil {
		priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return nil, nil, err
		}
		serialLimit := new(big.Int).Lsh(big.NewInt(1), 62)
		serial, err := rand.Int(rand.Reader, serialLimit)
		if err != nil {
			return nil, nil, err
		}
		now := time.Now()
		caTmpl := &x509.Certificate{
			SerialNumber: serial,
			Subject: pkix.Name{
				Organization: []string{"CreativeWorkzStudio LLC / FaithNet"},
				CommonName:   "FaithNet Sovereign Local Root CA",
			},
			NotBefore:             now.Add(-5 * time.Minute),
			NotAfter:              now.Add(10 * 365 * 24 * time.Hour),
			KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign | x509.KeyUsageDigitalSignature,
			BasicConstraintsValid: true,
			IsCA:                  true,
			MaxPathLen:            1,
		}
		caDER, err := x509.CreateCertificate(rand.Reader, caTmpl, caTmpl, &priv.PublicKey, priv)
		if err != nil {
			return nil, nil, err
		}
		caCert, err = x509.ParseCertificate(caDER)
		if err != nil {
			return nil, nil, err
		}
		caKey = priv
		caPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caDER})
		keyDER, err := x509.MarshalECPrivateKey(priv)
		if err != nil {
			return nil, nil, err
		}
		keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
		if err := os.WriteFile(caCertPath, caPEM, 0644); err != nil {
			return nil, nil, err
		}
		if err := os.WriteFile(caKeyPath, keyPEM, 0600); err != nil {
			return nil, nil, err
		}
	}

	// Sign leaf server certificate with FaithNet Sovereign Local Root CA.
	leafKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	serialLimit := new(big.Int).Lsh(big.NewInt(1), 62)
	serial, err := rand.Int(rand.Reader, serialLimit)
	if err != nil {
		return nil, nil, err
	}
	ips, dnsNames := collectLocalSANs()
	now := time.Now()
	leafTmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			Organization: []string{"CreativeWorkzStudio LLC / FaithNet"},
			CommonName:   "FaithNet Local Well",
		},
		NotBefore:             now.Add(-5 * time.Minute),
		NotAfter:              now.Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           ips,
		DNSNames:              dnsNames,
	}
	leafDER, err := x509.CreateCertificate(rand.Reader, leafTmpl, caCert, &leafKey.PublicKey, caKey)
	if err != nil {
		return nil, nil, err
	}
	leafPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: leafDER})
	leafKeyDER, err := x509.MarshalECPrivateKey(leafKey)
	if err != nil {
		return nil, nil, err
	}
	leafKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: leafKeyDER})
	_ = os.WriteFile(filepath.Join(caDir, "server.crt"), leafPEM, 0644)
	_ = os.WriteFile(filepath.Join(caDir, "server.key"), leafKeyPEM, 0600)

	tlsCert := tls.Certificate{
		Certificate: [][]byte{leafDER, caCert.Raw},
		PrivateKey:  leafKey,
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		MinVersion:   tls.VersionTLS12,
	}, caPEM, nil
}

// buildOfflineWellTLSConfig synthesizes an ephemeral offline ECDSA P-256 X.509
// certificate valid for localhost and all local IPv4/IPv6 addresses using 100%
// Go stdlib crypto—zero CA, zero internet, zero disk files.
func buildOfflineWellTLSConfig() (*tls.Config, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	serialLimit := new(big.Int).Lsh(big.NewInt(1), 62)
	serial, err := rand.Int(rand.Reader, serialLimit)
	if err != nil {
		return nil, err
	}

	ips := []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")}
	if addrs, err := net.InterfaceAddrs(); err == nil {
		for _, a := range addrs {
			if ipNet, ok := a.(*net.IPNet); ok && ipNet.IP != nil {
				ips = append(ips, ipNet.IP)
			}
		}
	}
	dnsNames := []string{"localhost", "faithnet.lan", "faithnet.home.arpa", "faithnet.local"}
	if host, err := os.Hostname(); err == nil && host != "" {
		dnsNames = append(dnsNames, host, host+".local")
	}

	now := time.Now()
	tmpl := x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			Organization: []string{"FaithNet Sovereign Offline Well"},
			CommonName:   "FaithNet Local Well",
		},
		NotBefore:             now.Add(-5 * time.Minute),
		NotAfter:              now.Add(30 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           ips,
		DNSNames:              dnsNames,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &priv.PublicKey, priv)
	if err != nil {
		return nil, err
	}
	cert := tls.Certificate{
		Certificate: [][]byte{derBytes},
		PrivateKey:  priv,
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}, nil
}

// peekConn replays buffered peek bytes before reading from the underlying conn.
type peekConn struct {
	net.Conn
	br *bufio.Reader
}

func (c *peekConn) Read(p []byte) (int, error) {
	return c.br.Read(p)
}

// dualProtoListener accepts TCP connections and inspects byte 0 concurrently:
// 0x16 (TLS ClientHello) is wrapped in tls.Server; all other traffic passes
// through as plain HTTP without blocking Accept().
type dualProtoListener struct {
	net.Listener
	tlsCfg  *tls.Config
	readyCh chan net.Conn
	errCh   chan error
	doneCh  chan struct{}
	once    sync.Once
}

func newDualProtoListener(inner net.Listener, tlsCfg *tls.Config) *dualProtoListener {
	dl := &dualProtoListener{
		Listener: inner,
		tlsCfg:   tlsCfg,
		readyCh:  make(chan net.Conn, 64),
		errCh:    make(chan error, 1),
		doneCh:   make(chan struct{}),
	}
	go dl.acceptLoop()
	return dl
}

func (dl *dualProtoListener) acceptLoop() {
	for {
		raw, err := dl.Listener.Accept()
		if err != nil {
			select {
			case dl.errCh <- err:
			default:
			}
			return
		}
		go func(c net.Conn) {
			_ = c.SetReadDeadline(time.Now().Add(2 * time.Second))
			br := bufio.NewReaderSize(c, 64)
			first, peekErr := br.Peek(1)
			_ = c.SetReadDeadline(time.Time{})

			wrapped := net.Conn(&peekConn{Conn: c, br: br})
			if peekErr == nil && len(first) == 1 && first[0] == 0x16 {
				wrapped = tls.Server(wrapped, dl.tlsCfg)
			}
			select {
			case dl.readyCh <- wrapped:
			case <-dl.doneCh:
				_ = wrapped.Close()
			}
		}(raw)
	}
}

func (dl *dualProtoListener) Accept() (net.Conn, error) {
	select {
	case c := <-dl.readyCh:
		return c, nil
	case err := <-dl.errCh:
		return nil, err
	case <-dl.doneCh:
		return nil, net.ErrClosed
	}
}

func (dl *dualProtoListener) Close() error {
	dl.once.Do(func() { close(dl.doneCh) })
	return dl.Listener.Close()
}

// Close gracefully shuts down the HTTP server and closes its TCP listener.
//
// ### For Architects, Operators & Non-Programmers:
// Shuts down The Well webpage cleanly without leaving open ports behind.
//
// ### For Systems Programmers & Engine Architects:
// Idempotent bounded shutdown via `http.Server.Shutdown` (1s timeout).
func (w *WellServer) Close() error {
	w.srvMu.Lock()
	srv := w.server
	w.server = nil
	w.listener = nil
	w.srvMu.Unlock()

	if srv == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return srv.Shutdown(ctx)
}

// End of interpreter/captive_web.go
