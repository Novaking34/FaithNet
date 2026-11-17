// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/integrated/global_airlock_bridge_test.go
// @id            test.integrated.global_airlock_bridge
// @layer         Layer 4: Tier-3 Integrated Verification Suite
// @archetype     verification_suite | integrated_global_bridge
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Rung 4 Global (TCP + Egress) + Airlock + Well Suite
// @role          Tier-3 Integrated Sensor: Global TCP/Egress -> Airlock -> Well
// @maturity      Rev 5 Baseline | Wave 6 4x4 Symmetry Expansion
// @chassis       bounded-heap | Local TCP + httptest Egress + WellServer Bridge
// @scale         O(1) bounded airlock fetch, TCP->UDP failover & Well bulletin
// @seams         in: data, engine, orchestrator, rung3_local, rung4_global, interpreter
// @invariants    1. Macro Tier 3 (Integrated): bridges Rung 4 Global TCP/Egress
//                   with Genesis 3 Airlock, Dispatcher, Rung 3 UDP, and The Well.
//                2. Zero real external internet traffic (100% 127.0.0.1 httptest).
//                3. External empirical facts enter strictly as 0:AWAIT, never +1.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Tier-3 Integrated verification suite proving the complete Rung 4 Global
//   boundary: live TCP peering (`TCPTransceiver`), opt-in Airlock-gated HTTP
//   fetching (`EgressGateway`), quarantine of external facts (`0:AWAIT`) routed
//   to The Well (`WellServer`) for local neighborhood display, hostile origin
//   rejection (`-1:NEGATIVE`), and automatic Rung 4 TCP -> Rung 3 UDP cascade.
//
//   ### For Architects, Operators & Non-Programmers:
//   Proves an anchor laptop can safely pull a short weather/news fact from an
//   outside web server when opted in, inspect it at the Genesis 3 Airlock,
//   share it cleanly on the local Well bulletin board for nearby phones, block
//   malicious websites, and drop cleanly to local UDP if TCP disconnects.
//
//   ### For Systems Programmers & Engine Architects:
//   Wires `EgressGateway` (`httptest.Server`), `TCPTransceiver`, `UDPTransceiver`,
//   `Dispatcher`, `Translator`, and `WellServer` concurrently under `-race`.
// ==============================================================================

package integrated_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Identifier, Verification Imports & Tier-3 Global Bridge Anchor )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-3 Integrated Verification Suite
// Module:        test.integrated.global_airlock_bridge
// Mission:       Bind Rung 4 Global TCP/Egress with Airlock, Dispatcher & Well.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// --- Tier 2: External Library Dependencies (None — 100% Go Stdlib Invariant) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---
import (
	"github.com/CreativeWorkzStudio/FaithNet/data"
	"github.com/CreativeWorkzStudio/FaithNet/engine"
	"github.com/CreativeWorkzStudio/FaithNet/engine/covenant"
	"github.com/CreativeWorkzStudio/FaithNet/interpreter"
	"github.com/CreativeWorkzStudio/FaithNet/orchestrator"
	"github.com/CreativeWorkzStudio/FaithNet/tests/common"
	"github.com/CreativeWorkzStudio/FaithNet/transport/rung3_local"
	"github.com/CreativeWorkzStudio/FaithNet/transport/rung4_global"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Operational Suite Constants & Identifiers
// ------------------------------------------------------------------------------

const (
	globalBridgeSuiteName = "Tier 3 Integrated: Rung 4 Global (TCP/Egress) + Airlock + Well"
)

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-3 Integrated Verification Suite
// Module:        test.integrated.global_airlock_bridge
// Mission:       Verify Egress -> Airlock -> Well pipeline, hostile remote
//                blocking, Rung 4 -> Rung 3 failover, and opt-in isolation.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: [ SECTION 1: ALWAYS PASS (EGRESS -> AIRLOCK -> THE WELL) ]
// ------------------------------------------------------------------------------

func TestGlobalAirlockBridge_Sec1_CleanEgressFactBridgesToWellAsAwait(t *testing.T) {
	// Local loopback mock web origin serving a short factual weather bulletin.
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("NOAA Local Advisory: Clear skies, 72F"))
	}))
	defer origin.Close()

	m := data.DefaultNodeManifest(common.NodeAlpha, "global-anchor")
	m.AllowGlobalEgress = true
	m.PreferredMaxRung = data.Rung4Global

	disp, err := orchestrator.NewDispatcher(m, engine.NewMonotonicClock())
	if err != nil {
		t.Fatalf("[SEC 1] NewDispatcher failed: %v", err)
	}
	defer func() { _ = disp.Close() }()

	gw := rung4_global.NewEgressGateway(m, disp.Airlock())
	well := interpreter.NewWellServer(disp, "127.0.0.1:0")
	tr := interpreter.NewTranslator(disp.Airlock())

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	frame, verdict, err := gw.FetchEmpiricalFrame(ctx, origin.URL, 1)
	if err != nil || verdict.State != data.QuadAwait || frame.Truth != data.QuadAwait {
		t.Fatalf("[SEC 1] clean egress fetch failed: verdict=%+v truth=%s err=%v",
			verdict, frame.Truth, err)
	}

	// Render through Translator and publish onto The Well local board.
	view := tr.FormatFrameForBrowser(&frame)
	well.RecordBulletin(interpreter.BulletinEntry{
		Seq:      view.SeqID,
		ClientOS: "Rung4-Egress",
		Truth:    view.Truth,
		Rung:     view.Rung,
		Message:  view.Text,
	})

	rec := httptest.NewRecorder()
	well.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "NOAA Local Advisory") {
		t.Fatalf("[SEC 1] Well HTML missing bridged empirical fact: HTTP %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "0:AWAIT") {
		t.Fatal("[SEC 1] bridged external fact must display 0:AWAIT quarantine badge")
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: [ SECTION 2: NEVER FAIL (HOSTILE & OVERSIZE REMOTE BLOCKED) ]
// ------------------------------------------------------------------------------

func TestGlobalAirlockBridge_Sec2_HostileAndOversizeRemoteOriginsBlocked(t *testing.T) {
	xssOrigin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("<script>window.location='http://evil'</script>"))
	}))
	defer xssOrigin.Close()

	oversizeOrigin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("Z", data.FloorMaxPayload+64)))
	}))
	defer oversizeOrigin.Close()

	// 1. Default offline-first manifest MUST refuse egress before dialing.
	offlineManifest := data.DefaultNodeManifest(common.NodeAlpha, "offline-strict")
	offlineGW := rung4_global.NewEgressGateway(offlineManifest, covenant.NewFilter())
	if _, _, err := offlineGW.FetchEmpiricalFrame(context.Background(), xssOrigin.URL, 1); !errors.Is(err, rung4_global.ErrGlobalEgressDisabled) {
		t.Fatalf("[SEC 2] offline manifest allowed egress: %v", err)
	}

	// 2. Opt-in manifest MUST still block XSS and >480B OOM responses at Airlock.
	optIn := offlineManifest
	optIn.AllowGlobalEgress = true
	gw := rung4_global.NewEgressGateway(optIn, covenant.NewFilter())

	if _, v, err := gw.FetchEmpiricalFrame(context.Background(), xssOrigin.URL, 2); !errors.Is(err, rung4_global.ErrEgressAirlockBlocked) || !v.IsBlocked() {
		t.Fatalf("[SEC 2] hostile XSS remote response bypassed Egress Airlock: v=%+v err=%v", v, err)
	}
	if _, v, err := gw.FetchEmpiricalFrame(context.Background(), oversizeOrigin.URL, 3); !errors.Is(err, rung4_global.ErrEgressAirlockBlocked) || !v.IsBlocked() {
		t.Fatalf("[SEC 2] oversize remote response bypassed Egress 480B cap: v=%+v err=%v", v, err)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: [ SECTION 3: INVARIANTS (TCP->UDP CASCADE UNDER DISPATCHER) ]
// ------------------------------------------------------------------------------

func TestGlobalAirlockBridge_Sec3_LiveTCPCutCascadesCleanlyToLocalUDP(t *testing.T) {
	mA := data.DefaultNodeManifest(common.NodeAlpha, "tcp-udp-a")
	mA.AllowGlobalEgress = true
	mA.PreferredMaxRung = data.Rung4Global
	mB := data.DefaultNodeManifest(common.NodeBeta, "tcp-udp-b")
	mB.AllowGlobalEgress = true
	mB.PreferredMaxRung = data.Rung4Global

	dA, _ := orchestrator.NewDispatcher(mA, engine.NewMonotonicClock())
	dB, _ := orchestrator.NewDispatcher(mB, engine.NewMonotonicClock())
	defer func() {
		_ = dA.Close()
		_ = dB.Close()
	}()

	udpA, _ := rung3_local.NewUDPTransceiver(common.NodeAlpha, rung3_local.DefaultLocalUDPAddr)
	udpB, _ := rung3_local.NewUDPTransceiver(common.NodeBeta, rung3_local.DefaultLocalUDPAddr)
	_ = udpA.RegisterPeer(common.NodeBeta, udpB.Addr())
	_ = udpB.RegisterPeer(common.NodeAlpha, udpA.Addr())

	tcpA, _ := rung4_global.NewTCPTransceiver(common.NodeAlpha, rung4_global.DefaultGlobalTCPAddr)
	tcpB, _ := rung4_global.NewTCPTransceiver(common.NodeBeta, rung4_global.DefaultGlobalTCPAddr)
	tcpA.RegisterPeer(common.NodeBeta, tcpB.Addr())
	tcpB.RegisterPeer(common.NodeAlpha, tcpA.Addr())

	_ = dA.Ladder().RegisterRung(udpA)
	_ = dA.Ladder().RegisterRung(tcpA)
	_ = dB.Ladder().RegisterRung(udpB)
	_ = dB.Ladder().RegisterRung(tcpB)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = dA.Start(ctx)
	_ = dB.Start(ctx)

	// 1. Starts on Rung 4 Global TCP.
	if _, r, err := dA.PingPeer(ctx, common.NodeBeta, []byte("GLOBAL-TCP")); err != nil || r != data.Rung4Global {
		t.Fatalf("[SEC 3] expected Rung4Global TCP ping, got rung=%s err=%v", r, err)
	}
	// 2. TCP snaps -> seamlessly drops to Rung 3 Local UDP with 0% loss.
	tcpA.SetHealthy(false)
	if _, r, err := dA.PingPeer(ctx, common.NodeBeta, []byte("LOCAL-UDP")); err != nil || r != data.Rung3Local {
		t.Fatalf("[SEC 3] expected fallback to Rung3Local UDP, got rung=%s err=%v", r, err)
	}
	if dA.Ping().HealthSnapshot().LossPercent() != 0 {
		t.Fatalf("[SEC 3] expected 0%% loss across TCP->UDP failover, got %+v", dA.Ping().HealthSnapshot())
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.4: [ SECTION 4: EMERGENCE (EGRESS -> TCP -> WELL JSON TELEMETRY) ]
// ------------------------------------------------------------------------------

func TestGlobalAirlockBridge_Sec4_EgressFactRelayedOverTCPToNeighborWell(t *testing.T) {
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("Grid Status: Substation B restored"))
	}))
	defer origin.Close()

	mA := data.DefaultNodeManifest(common.NodeAlpha, "relay-origin")
	mA.AllowGlobalEgress = true
	mA.PreferredMaxRung = data.Rung4Global
	mB := data.DefaultNodeManifest(common.NodeBeta, "neighbor-well")
	mB.AllowGlobalEgress = true
	mB.PreferredMaxRung = data.Rung4Global

	dA, _ := orchestrator.NewDispatcher(mA, engine.NewMonotonicClock())
	dB, _ := orchestrator.NewDispatcher(mB, engine.NewMonotonicClock())
	defer func() {
		_ = dA.Close()
		_ = dB.Close()
	}()

	wellB := interpreter.NewWellServer(dB, "127.0.0.1:0")
	trB := interpreter.NewTranslator(dB.Airlock())
	gotNotice := make(chan struct{}, 1)
	dB.SetDeliverHandler(func(f *data.Frame, _ covenant.Verdict) {
		if f.Opcode == data.OpLegacyWrap {
			v := trB.FormatFrameForBrowser(f)
			wellB.RecordBulletin(interpreter.BulletinEntry{
				Seq: v.SeqID, ClientOS: "Rung4-TCP", Truth: v.Truth, Rung: v.Rung, Message: v.Text,
			})
			select {
			case gotNotice <- struct{}{}:
			default:
			}
		}
	})

	tcpA, _ := rung4_global.NewTCPTransceiver(common.NodeAlpha, rung4_global.DefaultGlobalTCPAddr)
	tcpB, _ := rung4_global.NewTCPTransceiver(common.NodeBeta, rung4_global.DefaultGlobalTCPAddr)
	tcpA.RegisterPeer(common.NodeBeta, tcpB.Addr())
	tcpB.RegisterPeer(common.NodeAlpha, tcpA.Addr())
	_ = dA.Ladder().RegisterRung(tcpA)
	_ = dB.Ladder().RegisterRung(tcpB)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = dA.Start(ctx)
	_ = dB.Start(ctx)

	// NodeAlpha fetches clean fact via Egress and relays it over TCP to NodeBeta's Well!
	gw := rung4_global.NewEgressGateway(mA, dA.Airlock())
	frame, _, err := gw.FetchEmpiricalFrame(ctx, origin.URL, 55)
	if err != nil {
		t.Fatalf("[SEC 4] FetchEmpiricalFrame failed: %v", err)
	}
	frame.TargetID = common.NodeBeta
	if _, err := dA.SendFrame(&frame); err != nil {
		t.Fatalf("[SEC 4] SendFrame over TCP failed: %v", err)
	}

	select {
	case <-gotNotice:
	case <-ctx.Done():
		t.Fatal("[SEC 4] timed out waiting for relayed Egress frame at NodeBeta Well")
	}

	stRec := httptest.NewRecorder()
	wellB.Handler().ServeHTTP(stRec, httptest.NewRequest(http.MethodGet, "/api/status", nil))
	var st interpreter.WellStatusPayload
	if err := json.Unmarshal(stRec.Body.Bytes(), &st); err != nil {
		t.Fatalf("[SEC 4] /api/status JSON decode failed: %v", err)
	}
	if st.Bulletins != 1 || st.ActiveRung != "RUNG_4_GLOBAL" {
		t.Fatalf("[SEC 4] unexpected neighbor Well status: %+v", st)
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Preflight Diagnostic Gates & Physical Source Anchor Verification )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-3 Integrated Verification Suite
// Module:        test.integrated.global_airlock_bridge
// Mission:       Assert Rung 4 Global & Interpreter preflights and anchors.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Block 4 Pre-Flight Contract Gates
// ------------------------------------------------------------------------------

func TestGlobalAirlockBridge_Block4_PreflightGates(t *testing.T) {
	if globalBridgeSuiteName == "" {
		t.Fatal("[BLOCK 4] suite identifier must not be empty")
	}
	common.AssertAuditPass(t, rung4_global.PreflightGlobalTCP(), "")
	common.AssertAuditPass(t, rung4_global.PreflightGlobalEgress(), "")
	common.AssertAuditPass(t, interpreter.PreflightTranslator(), "")
	common.AssertAuditPass(t, interpreter.PreflightWellServer(), "")
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Physical Source Anchor Verification
// ------------------------------------------------------------------------------

func TestGlobalAirlockBridge_Block4_SourceAnchors(t *testing.T) {
	common.AssertRepoAnchors(t,
		"transport/rung4_global/tcp.go",
		"transport/rung4_global/egress.go",
		"interpreter/translator.go",
		"interpreter/captive_web.go",
	)
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Test Suite Epilogue & Clean Readiness Handshake )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-3 Integrated Verification Suite
// Module:        test.integrated.global_airlock_bridge
// Mission:       Confirm Rung 4 Global & Interpreter epilogue readiness.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Epilogue Readiness Verification
// ------------------------------------------------------------------------------

func TestGlobalAirlockBridge_Block5_EpilogueReadiness(t *testing.T) {
	if !rung4_global.IsGlobalTCPReady() || !rung4_global.IsGlobalEgressReady() ||
		!interpreter.IsTranslatorReady() || !interpreter.IsWellServerReady() ||
		!common.IsHarnessReady() {
		t.Fatal("[BLOCK 5: EPILOGUE] Rung 4 Global + Well bridge readiness failed")
	}
}

// End of tests/integrated/global_airlock_bridge_test.go
