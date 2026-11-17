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
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
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

// WellServer hosts FaithNet's zero-install local HTML captive portal.
//
// ### For Architects, Operators & Non-Programmers:
// Runs a lightweight offline webpage right inside a FaithNet node so any nearby
// phone or laptop browser can read local news and post safe messages.
//
// ### For Systems Programmers & Engine Architects:
// Mutex-guarded `[32]BulletinEntry` ring + stdlib `http.ServeMux` / `http.Server`.
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

	srvMu    sync.Mutex
	server   *http.Server
	listener net.Listener
	mux      *http.ServeMux
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
	sb.WriteString("</ul></body></html>")

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
//   - Layer 4 Tests  : tests/unit/interpreter/interpreter_test.go,
//                      tests/integrated/room_and_well_test.go
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
	w.mux.HandleFunc("/generate_204", w.handleCaptiveRedirect)
	w.mux.HandleFunc("/hotspot-detect.html", w.handleCaptiveRedirect)
	w.mux.HandleFunc("/connecttest.txt", w.handleCaptiveRedirect)
	return w
}

// Start binds the TCP listener and serves The Well HTTP portal in the background.
//
// ### For Architects, Operators & Non-Programmers:
// Opens The Well's local webpage so phones and laptops can connect immediately.
//
// ### For Systems Programmers & Engine Architects:
// Synchronously binds `net.Listen("tcp", listenAddr)` so `ListenAddr()` returns
// the exact ephemeral port immediately upon return.
func (w *WellServer) Start() error {
	ln, err := net.Listen("tcp", w.listenAddr)
	if err != nil {
		return err
	}
	srv := &http.Server{
		Handler:           w.mux,
		ReadHeaderTimeout: 3 * time.Second,
	}
	w.srvMu.Lock()
	w.listener = ln
	w.server = srv
	w.srvMu.Unlock()

	go func() {
		_ = srv.Serve(ln)
	}()
	return nil
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
