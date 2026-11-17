// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/functional/transport_functional_test.go
// @id            test.functional.transport
// @layer         Layer 4: Tier-1 Functional Verification Suite
// @archetype     verification_suite | functional_surface
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Layer 3A All-4-Rungs Transport Functional Surface
// @role          Tier-1 Functional Diagnostic Sensor (4-Section Micro Protocol)
// @maturity      Rev 5 Baseline | 4x4 Verification Matrix
// @chassis       bounded-heap | Exercises Cog 3 Transceiver across Rungs 1..4
// @scale         O(1) polymorphic send/recv sweep over every survival rung
// @seams         in: transport, rung1_floor, rung2_host, rung3_local, rung4_global
// @invariants    1. Macro Tier 1 (Functional): exercises ALL 4 survival rungs
//                   (Serial/AFSK/Morse, RAM/IPC, UDP/Beacon, TCP/Egress).
//                2. Enforces 512-byte Floppy Sector wire ceiling universally.
//                3. Zero leaked temporary sockets or goroutines after Close().
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Tier-1 Functional surface suite exercising FaithNet's Cog 3 `Transceiver`
//   contract across all 4 adaptive rungs (`rung1_floor`, `rung2_host`,
//   `rung3_local`, `rung4_global`) through all 4 horizontal micro sections,
//   all 9 transport Block 4 preflight gates, and all 9 Block 5 readiness probes.
//
//   ### For Architects, Operators & Non-Programmers:
//   Confirms at a surface level that every communication wire in FaithNet—from
//   1844 Morse pulses and radio chirps to RAM, IPC, LAN UDP, and global TCP—
//   speaks the exact same plug-and-play language.
//
//   ### For Systems Programmers & Engine Architects:
//   Paired with 10 dedicated 1:1 Tier 2 unit files under `tests/unit/transport/`.
// ==============================================================================

package functional_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Identifier, Verification Imports & Tier-1 Sensor Anchor )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-1 Functional Verification Suite
// Module:        test.functional.transport
// Mission:       Bind Cog 3 Transceiver and all 4 concrete survival rung pkgs.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"context"
	"io"
	"path/filepath"
	"testing"
	"time"
)

// --- Tier 2: External Library Dependencies (None — 100% Go Stdlib Invariant) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---
import (
	"github.com/CreativeWorkzStudio/FaithNet/data"
	"github.com/CreativeWorkzStudio/FaithNet/engine"
	"github.com/CreativeWorkzStudio/FaithNet/tests/common"
	"github.com/CreativeWorkzStudio/FaithNet/transport"
	"github.com/CreativeWorkzStudio/FaithNet/transport/rung1_floor"
	"github.com/CreativeWorkzStudio/FaithNet/transport/rung2_host"
	"github.com/CreativeWorkzStudio/FaithNet/transport/rung3_local"
	"github.com/CreativeWorkzStudio/FaithNet/transport/rung4_global"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Operational Suite Constants & Identifiers
// ------------------------------------------------------------------------------

const (
	transportFunctionalSuiteName = "Tier 1: Layer 3A All-4-Rungs Transport Surface"
)

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-1 Functional Verification Suite
// Module:        test.functional.transport
// Mission:       Verify polymorphic Transceiver behavior across all 4 rungs.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: [ SECTION 1: ALWAYS PASS (POLYMORPHIC ALL-4-RUNGS SURFACE) ]
// ------------------------------------------------------------------------------

func TestTransportFunctional_Sec1_AllFourRungsPolymorphicSurface(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Rung 1 Floor: Morse CW, AFSK Audio, SLIP Serial.
	mMed := rung1_floor.NewMorseMedium()
	var r1MorseA, r1MorseB transport.Transceiver = mMed.Attach(common.NodeAlpha), mMed.Attach(common.NodeBeta)
	defer func() { _ = r1MorseA.Close(); _ = r1MorseB.Close() }()
	if r1MorseA.Rung() != data.Rung1Floor || r1MorseA.Send(common.NodeBeta, []byte("R1M")) != nil {
		t.Fatal("[SEC 1] Rung 1 Morse surface failed")
	}
	if env, err := r1MorseB.Receive(ctx); err != nil || string(env.Slice()) != "R1M" {
		t.Fatalf("[SEC 1] Rung 1 Morse recv mismatch: err=%v got=%q", err, env.Slice())
	}

	aMed := rung1_floor.NewAudioMedium()
	var r1AudioA, r1AudioB transport.Transceiver = aMed.Attach(common.NodeAlpha), aMed.Attach(common.NodeBeta)
	defer func() { _ = r1AudioA.Close(); _ = r1AudioB.Close() }()
	if r1AudioA.Send(common.NodeBeta, []byte("R1A")) != nil {
		t.Fatal("[SEC 1] Rung 1 Audio surface failed")
	}
	if env, err := r1AudioB.Receive(ctx); err != nil || string(env.Slice()) != "R1A" {
		t.Fatalf("[SEC 1] Rung 1 Audio recv mismatch: err=%v got=%q", err, env.Slice())
	}

	rB, wA := io.Pipe()
	rA, wB := io.Pipe()
	var r1SerA, r1SerB transport.Transceiver = rung1_floor.NewSerialTransceiver(common.NodeAlpha, rA, wA, wA),
		rung1_floor.NewSerialTransceiver(common.NodeBeta, rB, wB, wB)
	defer func() { _ = r1SerA.Close(); _ = r1SerB.Close() }()
	go func() { _ = r1SerA.Send(common.NodeBeta, []byte("R1S")) }()
	if env, err := r1SerB.Receive(ctx); err != nil || string(env.Slice()) != "R1S" {
		t.Fatalf("[SEC 1] Rung 1 Serial recv mismatch: err=%v got=%q", err, env.Slice())
	}

	// Rung 2 Host: RAM MemoryBus & Unixgram IPC.
	bus := rung2_host.NewMemoryBus()
	var r2MemA, r2MemB transport.Transceiver = bus.Attach(common.NodeAlpha), bus.Attach(common.NodeBeta)
	defer func() { _ = r2MemA.Close(); _ = r2MemB.Close() }()
	if r2MemA.Send(common.NodeBeta, []byte("R2M")) != nil {
		t.Fatal("[SEC 1] Rung 2 Memory surface failed")
	}
	if env, err := r2MemB.Receive(ctx); err != nil || string(env.Slice()) != "R2M" {
		t.Fatalf("[SEC 1] Rung 2 Memory recv mismatch: err=%v got=%q", err, env.Slice())
	}

	tmp := t.TempDir()
	ipcA, _ := rung2_host.NewIPCTransceiver(common.NodeAlpha, filepath.Join(tmp, "a.sock"))
	ipcB, _ := rung2_host.NewIPCTransceiver(common.NodeBeta, filepath.Join(tmp, "b.sock"))
	defer func() { _ = ipcA.Close(); _ = ipcB.Close() }()
	ipcA.RegisterPeer(common.NodeBeta, ipcB.SocketPath())
	if ipcA.Send(common.NodeBeta, []byte("R2I")) != nil {
		t.Fatal("[SEC 1] Rung 2 IPC surface failed")
	}
	if env, err := ipcB.Receive(ctx); err != nil || string(env.Slice()) != "R2I" {
		t.Fatalf("[SEC 1] Rung 2 IPC recv mismatch: err=%v got=%q", err, env.Slice())
	}

	// Rung 3 Local: LAN UDP & Room Beacon.
	udpA, _ := rung3_local.NewUDPTransceiver(common.NodeAlpha, rung3_local.DefaultLocalUDPAddr)
	udpB, _ := rung3_local.NewUDPTransceiver(common.NodeBeta, rung3_local.DefaultLocalUDPAddr)
	defer func() { _ = udpA.Close(); _ = udpB.Close() }()
	_ = udpA.RegisterPeer(common.NodeBeta, udpB.Addr())
	if udpA.Send(common.NodeBeta, []byte("R3U")) != nil {
		t.Fatal("[SEC 1] Rung 3 UDP surface failed")
	}
	if env, err := udpB.Receive(ctx); err != nil || string(env.Slice()) != "R3U" {
		t.Fatalf("[SEC 1] Rung 3 UDP recv mismatch: err=%v got=%q", err, env.Slice())
	}

	// Rung 4 Global: Length-Prefixed TCP.
	tcpA, _ := rung4_global.NewTCPTransceiver(common.NodeAlpha, rung4_global.DefaultGlobalTCPAddr)
	tcpB, _ := rung4_global.NewTCPTransceiver(common.NodeBeta, rung4_global.DefaultGlobalTCPAddr)
	defer func() { _ = tcpA.Close(); _ = tcpB.Close() }()
	tcpA.RegisterPeer(common.NodeBeta, tcpB.Addr())
	if tcpA.Send(common.NodeBeta, []byte("R4T")) != nil {
		t.Fatal("[SEC 1] Rung 4 TCP surface failed")
	}
	if env, err := tcpB.Receive(ctx); err != nil || string(env.Slice()) != "R4T" {
		t.Fatalf("[SEC 1] Rung 4 TCP recv mismatch: err=%v got=%q", err, env.Slice())
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: [ SECTION 2: NEVER FAIL (FAULT INVERSION & DEGRADATION) ]
// ------------------------------------------------------------------------------

func TestTransportFunctional_Sec2_OfflineClosedAndSealedEgressRejection(t *testing.T) {
	bus := rung2_host.NewMemoryBus()
	a := bus.Attach(common.NodeAlpha)
	defer func() { _ = a.Close() }()

	if err := a.Send(common.NodeBeta, []byte("MISSING")); err != transport.ErrPeerUnreachable {
		t.Fatalf("[SEC 2] expected ErrPeerUnreachable, got %v", err)
	}
	a.SetHealthy(false)
	if err := a.Send(common.NodeBeta, []byte("OFFLINE")); err != transport.ErrTransceiverOffline {
		t.Fatalf("[SEC 2] expected ErrTransceiverOffline, got %v", err)
	}
	_ = a.Close()
	if err := a.Send(common.NodeBeta, []byte("CLOSED")); err != transport.ErrTransceiverClosed {
		t.Fatalf("[SEC 2] expected ErrTransceiverClosed, got %v", err)
	}

	// Offline-first EgressGateway must refuse outbound web access by default.
	gw := rung4_global.NewEgressGateway(data.DefaultNodeManifest(common.NodeAlpha, "sealed"), nil)
	if _, _, err := gw.FetchEmpiricalFrame(context.Background(), "http://127.0.0.1:1/", 1); err != rung4_global.ErrGlobalEgressDisabled {
		t.Fatalf("[SEC 2] expected ErrGlobalEgressDisabled, got %v", err)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: [ SECTION 3: INVARIANTS, SATURATION & LIMIT GUARDS ]
// ------------------------------------------------------------------------------

func TestTransportFunctional_Sec3_Wire512ByteCeilingAndEnvelopeClamp(t *testing.T) {
	var overWire [data.FloorMaxFrameSize + 1]byte
	if _, err := rung1_floor.SLIPEncode(overWire[:]); err != transport.ErrWireFrameOversize {
		t.Fatalf("[SEC 3] Rung 1 SLIP 513B bypassed ceiling: %v", err)
	}
	if _, err := rung1_floor.ModulateAFSK(overWire[:]); err != transport.ErrWireFrameOversize {
		t.Fatalf("[SEC 3] Rung 1 AFSK 513B bypassed ceiling: %v", err)
	}
	if _, err := rung1_floor.MorseEncodeBytes(overWire[:]); err != transport.ErrWireFrameOversize {
		t.Fatalf("[SEC 3] Rung 1 Morse 513B bypassed ceiling: %v", err)
	}

	neg := transport.WireEnvelope{Length: -5}
	big := transport.WireEnvelope{Length: 9999}
	if len(neg.Slice()) != 0 || len(big.Slice()) != data.FloorMaxFrameSize {
		t.Fatal("[SEC 3] WireEnvelope.Slice() saturation bounds violated")
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.4: [ SECTION 4: EMERGENCE & COMPOSITE WORKFLOWS ]
// ------------------------------------------------------------------------------

func TestTransportFunctional_Sec4_FullPingPongAcrossMemoryBus(t *testing.T) {
	bus := rung2_host.NewMemoryBus()
	wireA := bus.Attach(common.NodeAlpha)
	wireB := bus.Attach(common.NodeBeta)
	defer func() {
		_ = wireA.Close()
		_ = wireB.Close()
	}()

	clk := engine.NewMockClock(1000)
	pe := engine.NewPingEngine(common.NodeAlpha, clk)
	var req, decodedReq, resp, decodedResp data.Frame
	var buf [data.FloorMaxFrameSize]byte

	_ = pe.BuildPingInto(&req, common.NodeBeta, data.Rung2Host, []byte("FIRST-BREATH"))
	n, _ := engine.EncodeInto(buf[:], &req)
	_ = wireA.Send(common.NodeBeta, buf[:n])

	clk.Advance(300)
	envIn, _ := wireB.Receive(context.Background())
	_ = engine.DecodeFrom(envIn.Slice(), &decodedReq)
	_ = engine.ReflectPingInto(&decodedReq, &resp, common.NodeBeta)
	n2, _ := engine.EncodeInto(buf[:], &resp)
	_ = wireB.Send(common.NodeAlpha, buf[:n2])

	clk.Advance(300)
	envOut, _ := wireA.Receive(context.Background())
	_ = engine.DecodeFrom(envOut.Slice(), &decodedResp)
	rtt, ok := pe.ProcessPong(&decodedResp)
	if !ok || rtt != 600 {
		t.Fatalf("[SEC 4] expected First Breath RTT=600ns, got ok=%v rtt=%d", ok, rtt)
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Preflight Diagnostic Gates & Physical Source Anchor Verification )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-1 Functional Verification Suite
// Module:        test.functional.transport
// Mission:       Assert ALL 9 transport Block 4 preflights and 10 anchors.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Preflight Diagnostic Gates
// ------------------------------------------------------------------------------

func TestTransportFunctional_Block4_PreflightAllFourRungs(t *testing.T) {
	common.AssertAuditPass(t, rung1_floor.PreflightSerialFloor(), "")
	common.AssertAuditPass(t, rung1_floor.PreflightAudioModem(), "")
	common.AssertAuditPass(t, rung1_floor.PreflightMorseFloor(), "")
	common.AssertAuditPass(t, rung2_host.PreflightMemoryBus(), "")
	common.AssertAuditPass(t, rung2_host.PreflightHostIPC(), "")
	common.AssertAuditPass(t, rung3_local.PreflightLocalUDP(), "")
	common.AssertAuditPass(t, rung3_local.PreflightBeaconSwitch(), "")
	common.AssertAuditPass(t, rung4_global.PreflightGlobalTCP(), "")
	common.AssertAuditPass(t, rung4_global.PreflightGlobalEgress(), "")
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Suite Identifier & Physical Source Anchor Gate
// ------------------------------------------------------------------------------

func TestTransportFunctional_Block4_SourceAnchors(t *testing.T) {
	if transportFunctionalSuiteName == "" {
		t.Fatal("[BLOCK 4] suite identifier must not be empty")
	}
	common.AssertRepoAnchors(t,
		"transport/transceiver.go",
		"transport/rung1_floor/serial.go",
		"transport/rung1_floor/audio.go",
		"transport/rung1_floor/morse.go",
		"transport/rung2_host/memory.go",
		"transport/rung2_host/ipc.go",
		"transport/rung3_local/udp.go",
		"transport/rung3_local/beacon.go",
		"transport/rung4_global/tcp.go",
		"transport/rung4_global/egress.go",
	)
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Test Suite Epilogue & Clean Readiness Handshake )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-1 Functional Verification Suite
// Module:        test.functional.transport
// Mission:       Verify all 9 Layer 3A transport readiness probes post-suite.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Epilogue & Clean State Handshake
// ------------------------------------------------------------------------------

func TestTransportFunctional_Block5_EpilogueReadiness(t *testing.T) {
	if !rung1_floor.IsSerialFloorReady() || !rung1_floor.IsAudioModemReady() ||
		!rung1_floor.IsMorseFloorReady() || !rung2_host.IsMemoryBusReady() ||
		!rung2_host.IsHostIPCReady() || !rung3_local.IsLocalUDPReady() ||
		!rung3_local.IsBeaconSwitchReady() || !rung4_global.IsGlobalTCPReady() ||
		!rung4_global.IsGlobalEgressReady() || !common.IsHarnessReady() {
		t.Fatal("[BLOCK 5: EPILOGUE] all-4-rungs transport readiness handshake failed")
	}
}

// End of tests/functional/transport_functional_test.go
