// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/integrated/room_and_well_test.go
// @id            test.integrated.room_and_well
// @layer         Layer 4: Tier-3 Integrated Verification Suite
// @archetype     verification_suite | integrated_room_well
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet "The Room & The Well" Live UDP + Captive Portal Suite
// @role          Tier-3 Integrated Sensor: Real LAN UDP + Beacon + Well Portal
// @maturity      Rev 5 Baseline | Wave 4 The Room & The Well
// @chassis       bounded-heap | Two live UDP nodes + live HTTP Well portal
// @scale         O(1) live loopback UDP ping & HTTP portal end-to-end flow
// @seams         in: data, engine, orchestrator, rung3_local, interpreter
// @invariants    1. Macro Tier 3 (Integrated): live real OS UDP sockets +
//                   real HTTP WellServer + BeaconSwitch + Genesis 3 Airlock.
//                2. Zero internet access required; 100% 127.0.0.1 loopback.
//                3. Zero leaked goroutines or sockets under `go test -race`.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Tier-3 Integrated verification suite proving Wave 4 ("The Room & The Well")
//   end-to-end: two live orchestrated FaithNet nodes discovering and pinging
//   each other over REAL loopback UDP sockets while serving a live zero-install
//   HTML portal ("The Well") that screens foreign visitors through the Genesis 3
//   Airlock.
//
//   ### For Architects, Operators & Non-Programmers:
//   Proves two laptops in a room can find each other over local UDP, turn on
//   their room beacon, and let any visitor's phone post local updates through
//   The Well while blocking web attacks completely offline.
//
//   ### For Systems Programmers & Engine Architects:
//   Exercises `Dispatcher` + `UDPTransceiver` (with peer auto-learning) +
//   `BeaconSwitch` + `WellServer` concurrently under `-race`.
// ==============================================================================

package integrated_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Identifier, Verification Imports & Tier-3 Integrated Anchor )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-3 Integrated Verification Suite
// Module:        test.integrated.room_and_well
// Mission:       Bind Rung 3 Local UDP, Beacon, Orchestrator, and The Well.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// --- Tier 2: External Library Dependencies (None) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---
import (
	"github.com/CreativeWorkzStudio/FaithNet/data"
	"github.com/CreativeWorkzStudio/FaithNet/engine"
	"github.com/CreativeWorkzStudio/FaithNet/interpreter"
	"github.com/CreativeWorkzStudio/FaithNet/orchestrator"
	"github.com/CreativeWorkzStudio/FaithNet/tests/common"
	"github.com/CreativeWorkzStudio/FaithNet/transport/rung2_host"
	"github.com/CreativeWorkzStudio/FaithNet/transport/rung3_local"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Operational Suite Constants & Identifiers
// ------------------------------------------------------------------------------

const (
	roomAndWellSuiteName = "Tier 3 Integrated: The Room (UDP/Beacon) & The Well"
)

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-3 Integrated Verification Suite
// Module:        test.integrated.room_and_well
// Mission:       Prove live UDP Room pings, UDP->RAM cascade, and Well portal.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: [ SECTION 1: ALWAYS PASS (LIVE ROOM UDP PING EXCHANGE) ]
// ------------------------------------------------------------------------------

func TestRoomWellIntegrated_Sec1_LiveLoopbackUDPPingAndAutoLearn(t *testing.T) {
	dA, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(common.NodeAlpha, "room-alpha"), engine.NewMonotonicClock(),
	)
	dB, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(common.NodeBeta, "room-beta"), engine.NewMonotonicClock(),
	)
	defer func() {
		_ = dA.Close()
		_ = dB.Close()
	}()

	udpA, err := rung3_local.NewUDPTransceiver(common.NodeAlpha, rung3_local.DefaultLocalUDPAddr)
	if err != nil {
		t.Fatalf("[SEC 1] bind UDP A failed: %v", err)
	}
	udpB, err := rung3_local.NewUDPTransceiver(common.NodeBeta, rung3_local.DefaultLocalUDPAddr)
	if err != nil {
		_ = udpA.Close()
		t.Fatalf("[SEC 1] bind UDP B failed: %v", err)
	}

	// Only A registers B's UDP address; B learns A's address on the fly!
	if err := udpA.RegisterPeer(common.NodeBeta, udpB.Addr()); err != nil {
		t.Fatalf("[SEC 1] RegisterPeer failed: %v", err)
	}
	_ = dA.Ladder().RegisterRung(udpA)
	_ = dB.Ladder().RegisterRung(udpB)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = dA.Start(ctx)
	_ = dB.Start(ctx)

	rtt, rung, err := dA.PingPeer(ctx, common.NodeBeta, []byte("LIVE-ROOM-UDP"))
	if err != nil || rung != data.Rung3Local || rtt < 0 {
		t.Fatalf("[SEC 1] live UDP PingPeer failed: rtt=%d rung=%s err=%v", rtt, rung, err)
	}
	if _, learned := dB.Router().Lookup(common.NodeAlpha); !learned {
		t.Fatal("[SEC 1] NodeBeta Router did not learn NodeAlpha over live UDP")
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: [ SECTION 2: NEVER FAIL (AIRLOCK AT THE WELL DOOR) ]
// ------------------------------------------------------------------------------

func TestRoomWellIntegrated_Sec2_WellBlocksHostileBrowserAttacks(t *testing.T) {
	dA, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(common.NodeAlpha, "room-guard"), engine.NewMonotonicClock(),
	)
	defer func() { _ = dA.Close() }()
	well := interpreter.NewWellServer(dA, "127.0.0.1:0")

	attacks := []string{
		"<script>document.cookie</script>",
		"SELECT * FROM secrets; DROP TABLE users;",
		"1 UNION SELECT password FROM accounts",
		"javascript:alert(1)",
	}
	for _, atk := range attacks {
		rec := httptest.NewRecorder()
		well.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/message",
			strings.NewReader(atk)))
		if rec.Code != http.StatusForbidden {
			t.Fatalf("[SEC 2] attack %q bypassed WellServer: HTTP %d", atk, rec.Code)
		}
	}

	// Status endpoint must reflect all 4 airlock blocks cleanly.
	stRec := httptest.NewRecorder()
	well.Handler().ServeHTTP(stRec, httptest.NewRequest(http.MethodGet, "/api/status", nil))
	var st interpreter.WellStatusPayload
	if err := json.Unmarshal(stRec.Body.Bytes(), &st); err != nil {
		t.Fatalf("[SEC 2] invalid /api/status JSON: %v", err)
	}
	if st.AirlockBlocked != uint64(len(attacks)) || st.Bulletins != 0 {
		t.Fatalf("[SEC 2] unexpected Well status after attack sweep: %+v", st)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: [ SECTION 3: INVARIANTS (UDP FALLBACK TO RUNG 2 RAM) ]
// ------------------------------------------------------------------------------

func TestRoomWellIntegrated_Sec3_RealUDPSnapCascadesDownToRung2Memory(t *testing.T) {
	dA, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(common.NodeAlpha, "cascade-a"), engine.NewMonotonicClock(),
	)
	dB, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(common.NodeBeta, "cascade-b"), engine.NewMonotonicClock(),
	)
	defer func() {
		_ = dA.Close()
		_ = dB.Close()
	}()

	// Wire BOTH Rung 2 MemoryBus AND Rung 3 Real UDP.
	bus := rung2_host.NewMemoryBus()
	_ = dA.Ladder().RegisterRung(bus.Attach(common.NodeAlpha))
	_ = dB.Ladder().RegisterRung(bus.Attach(common.NodeBeta))

	udpA, _ := rung3_local.NewUDPTransceiver(common.NodeAlpha, rung3_local.DefaultLocalUDPAddr)
	udpB, _ := rung3_local.NewUDPTransceiver(common.NodeBeta, rung3_local.DefaultLocalUDPAddr)
	_ = udpA.RegisterPeer(common.NodeBeta, udpB.Addr())
	_ = dA.Ladder().RegisterRung(udpA)
	_ = dB.Ladder().RegisterRung(udpB)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = dA.Start(ctx)
	_ = dB.Start(ctx)

	// Step 1: Rides Rung 3 Real UDP!
	if _, rung, err := dA.PingPeer(ctx, common.NodeBeta, []byte("ON-UDP")); err != nil || rung != data.Rung3Local {
		t.Fatalf("[SEC 3] expected initial ping on Rung3Local, got rung=%s err=%v", rung, err)
	}

	// Step 2: Wi-Fi/UDP drops -> automatically cascades down to Rung 2 MemoryBus!
	udpA.SetHealthy(false)
	if _, rung, err := dA.PingPeer(ctx, common.NodeBeta, []byte("ON-RAM")); err != nil || rung != data.Rung2Host {
		t.Fatalf("[SEC 3] expected failover onto Rung2Host, got rung=%s err=%v", rung, err)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.4: [ SECTION 4: EMERGENCE (BEACON + UDP + WELL DASHBOARD) ]
// ------------------------------------------------------------------------------

func TestRoomWellIntegrated_Sec4_FullRoomBeaconUDPDiscoveryAndWellPortal(t *testing.T) {
	clk := engine.NewMonotonicClock()
	dA, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(common.NodeAlpha, "Community-Anchor"), clk,
	)
	dB, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(common.NodeBeta, "Neighbor-Laptop"), clk,
	)
	defer func() {
		_ = dA.Close()
		_ = dB.Close()
	}()

	udpA, _ := rung3_local.NewUDPTransceiver(common.NodeAlpha, rung3_local.DefaultLocalUDPAddr)
	udpB, _ := rung3_local.NewUDPTransceiver(common.NodeBeta, rung3_local.DefaultLocalUDPAddr)
	_ = udpA.RegisterPeer(common.NodeBeta, udpB.Addr())
	_ = dA.Ladder().RegisterRung(udpA)
	_ = dB.Ladder().RegisterRung(udpB)

	beacon, err := rung3_local.NewBeaconSwitch(
		"FaithNet-Block-7", rung3_local.DefaultBeaconChannel, rung3_local.BeaconModeSimulated, clk,
	)
	if err != nil || beacon.TurnOn() != nil {
		t.Fatalf("[SEC 4] BeaconSwitch TurnOn failed: %v", err)
	}
	defer func() { _ = beacon.Close() }()

	well := interpreter.NewWellServer(dA, "127.0.0.1:0")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = dA.Start(ctx)
	_ = dB.Start(ctx)

	// Broadcast room beacon from NodeAlpha -> NodeBeta receives and learns route.
	bf := beacon.BuildBeaconFrame(common.NodeAlpha, data.Rung3Local, 1)
	if _, err := dA.SendFrame(&bf); err != nil {
		t.Fatalf("[SEC 4] SendFrame OpBeacon failed: %v", err)
	}
	if _, _, err := dA.PingPeer(ctx, common.NodeBeta, []byte("NEIGHBOR-SYNC")); err != nil {
		t.Fatalf("[SEC 4] neighbor ping failed: %v", err)
	}

	// Visitor on Android posts clean local news to The Well; verify JSON status.
	req := httptest.NewRequest(http.MethodPost, "/api/message",
		strings.NewReader("Block 7 generator check complete"))
	req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 14; Pixel 8)")
	postRec := httptest.NewRecorder()
	well.Handler().ServeHTTP(postRec, req)
	if postRec.Code != http.StatusAccepted {
		t.Fatalf("[SEC 4] Android Well post failed: HTTP %d", postRec.Code)
	}

	stRec := httptest.NewRecorder()
	well.Handler().ServeHTTP(stRec, httptest.NewRequest(http.MethodGet, "/api/status", nil))
	var st interpreter.WellStatusPayload
	_ = json.Unmarshal(stRec.Body.Bytes(), &st)
	if st.ActiveRung != "RUNG_3_LOCAL" || st.PeerCount != 1 || st.Bulletins != 1 {
		t.Fatalf("[SEC 4] unexpected integrated Well status: %+v", st)
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Preflight Diagnostic Gates & Physical Source Anchor Verification )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-3 Integrated Verification Suite
// Module:        test.integrated.room_and_well
// Mission:       Assert Rung 3 & Layer 3B preflights and verify source anchors.
// ==============================================================================

func TestRoomWellIntegrated_Block4_PreflightAndAnchors(t *testing.T) {
	if roomAndWellSuiteName == "" {
		t.Fatal("[BLOCK 4] suite identifier must not be empty")
	}
	common.AssertAuditPass(t, rung3_local.PreflightLocalUDP(), "")
	common.AssertAuditPass(t, rung3_local.PreflightBeaconSwitch(), "")
	common.AssertAuditPass(t, interpreter.PreflightTranslator(), "")
	common.AssertAuditPass(t, interpreter.PreflightWellServer(), "")
	common.AssertAuditPass(t, interpreter.PreflightLegacyBridge(), "")
	common.AssertRepoAnchors(t,
		"cmd/faithnet-node/main.go",
		"transport/rung3_local/udp.go",
		"transport/rung3_local/beacon.go",
		"interpreter/captive_web.go",
	)
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Test Suite Epilogue & Clean Readiness Handshake )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-3 Integrated Verification Suite
// Module:        test.integrated.room_and_well
// Mission:       Confirm full Room & Well readiness handshake post-suite.
// ==============================================================================

func TestRoomWellIntegrated_Block5_EpilogueReadiness(t *testing.T) {
	if !rung3_local.IsLocalUDPReady() || !rung3_local.IsBeaconSwitchReady() ||
		!interpreter.IsWellServerReady() || !common.IsHarnessReady() {
		t.Fatal("[BLOCK 5: EPILOGUE] Room & Well integrated readiness failed")
	}
}

// End of tests/integrated/room_and_well_test.go
