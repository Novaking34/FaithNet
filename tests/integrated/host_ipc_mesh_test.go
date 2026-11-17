// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/integrated/host_ipc_mesh_test.go
// @id            test.integrated.host_ipc_mesh
// @layer         Layer 4: Tier-3 Integrated Verification Suite
// @archetype     verification_suite | integrated_host_mesh
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Rung 2 Host (MemoryBus + Unix IPC) + Orchestrator Suite
// @role          Tier-3 Integrated Sensor: Dual Rung 2 Host IPC/RAM Mesh
// @maturity      Rev 5 Baseline | Wave 6 4x4 Symmetry Expansion
// @chassis       bounded-heap | Real Unixgram IPC + RAM Bus + Dispatcher Mesh
// @scale         O(1) multi-peer host IPC & RAM dispatch under -race
// @seams         in: data, engine, orchestrator, transport/rung2_host, common
// @invariants    1. Macro Tier 3 (Integrated): binds Rung 2 real Unixgram IPC
//                   sockets and Rung 2 MemoryBus inside live Dispatchers.
//                2. Zero leaked Unix sockets (`t.TempDir()`) and zero races.
//                3. Proves dual-slot Rung 2 resilience and concurrent pings.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Tier-3 Integrated verification suite combining Rung 2 Host transports
//   (`IPCTransceiver` real OS Unix domain datagrams + `MemoryBus` in-process
//   RAM fabric) with `orchestrator.Dispatcher`, `AdaptiveLadder`, `Router`,
//   and `PingEngine` for multi-node local host mesh and concurrent ping tests.
//
//   ### For Architects, Operators & Non-Programmers:
//   Proves multiple FaithNet services running on the same computer can talk
//   over real local OS sockets or RAM channels, handle simultaneous traffic,
//   and survive cleanly if one local channel goes down.
//
//   ### For Systems Programmers & Engine Architects:
//   Exercises `IPCTransceiver` + `MemoryBus` + `Dispatcher` under `-race`,
//   including concurrent `PingPeer` goroutines, broadcast `OpBeacon` learning,
//   and automatic teardown of temporary Unix sockets.
// ==============================================================================

package integrated_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Identifier, Verification Imports & Tier-3 Host Mesh Anchor )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-3 Integrated Verification Suite
// Module:        test.integrated.host_ipc_mesh
// Mission:       Bind Rung 2 Host IPC/MemoryBus with live Orchestrator nodes.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"context"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// --- Tier 2: External Library Dependencies (None — 100% Go Stdlib Invariant) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---
import (
	"github.com/CreativeWorkzStudio/FaithNet/data"
	"github.com/CreativeWorkzStudio/FaithNet/engine"
	"github.com/CreativeWorkzStudio/FaithNet/orchestrator"
	"github.com/CreativeWorkzStudio/FaithNet/tests/common"
	"github.com/CreativeWorkzStudio/FaithNet/transport/rung2_host"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Operational Suite Constants & Identifiers
// ------------------------------------------------------------------------------

const (
	hostIPCMeshSuiteName = "Tier 3 Integrated: Rung 2 Host (IPC + MemoryBus) Mesh"
)

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-3 Integrated Verification Suite
// Module:        test.integrated.host_ipc_mesh
// Mission:       Verify live Unixgram IPC pings, fault resilience, invariants,
//                and concurrent 3-node host mesh emergence under `-race`.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: [ SECTION 1: ALWAYS PASS (REAL UNIX IPC DISPATCHER PING) ]
// ------------------------------------------------------------------------------

func TestHostIPCMesh_Sec1_LiveUnixgramIPCOrchestratorPing(t *testing.T) {
	tmp := t.TempDir()
	ipcA, err := rung2_host.NewIPCTransceiver(common.NodeAlpha, filepath.Join(tmp, "alpha.sock"))
	if err != nil {
		t.Fatalf("[SEC 1] NewIPCTransceiver alpha failed: %v", err)
	}
	ipcB, err := rung2_host.NewIPCTransceiver(common.NodeBeta, filepath.Join(tmp, "beta.sock"))
	if err != nil {
		_ = ipcA.Close()
		t.Fatalf("[SEC 1] NewIPCTransceiver beta failed: %v", err)
	}
	ipcA.RegisterPeer(common.NodeBeta, ipcB.SocketPath())
	ipcB.RegisterPeer(common.NodeAlpha, ipcA.SocketPath())

	dA, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(common.NodeAlpha, "ipc-alpha"), engine.NewMonotonicClock(),
	)
	dB, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(common.NodeBeta, "ipc-beta"), engine.NewMonotonicClock(),
	)
	defer func() {
		_ = dA.Close()
		_ = dB.Close()
	}()

	_ = dA.Ladder().RegisterRung(ipcA)
	_ = dB.Ladder().RegisterRung(ipcB)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = dA.Start(ctx)
	_ = dB.Start(ctx)

	rtt, rung, err := dA.PingPeer(ctx, common.NodeBeta, []byte("HOST-IPC-PING"))
	if err != nil || rung != data.Rung2Host || rtt < 0 {
		t.Fatalf("[SEC 1] live IPC PingPeer failed: rtt=%d rung=%s err=%v", rtt, rung, err)
	}
	if _, learned := dB.Router().Lookup(common.NodeAlpha); !learned {
		t.Fatal("[SEC 1] NodeBeta Router did not learn NodeAlpha over Unix IPC")
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: [ SECTION 2: NEVER FAIL (IPC FAILURE -> RAM FAILOVER) ]
// ------------------------------------------------------------------------------

func TestHostIPCMesh_Sec2_IPCDegradationFailsOverCleanlyToMemoryBus(t *testing.T) {
	tmp := t.TempDir()
	bus := rung2_host.NewMemoryBus()
	ramA := bus.Attach(common.NodeAlpha)
	ramB := bus.Attach(common.NodeBeta)

	ipcA, err := rung2_host.NewIPCTransceiver(common.NodeAlpha, filepath.Join(tmp, "f_a.sock"))
	if err != nil {
		t.Fatalf("[SEC 2] bind IPC A failed: %v", err)
	}
	ipcB, err := rung2_host.NewIPCTransceiver(common.NodeBeta, filepath.Join(tmp, "f_b.sock"))
	if err != nil {
		_ = ipcA.Close()
		t.Fatalf("[SEC 2] bind IPC B failed: %v", err)
	}
	ipcA.RegisterPeer(common.NodeBeta, ipcB.SocketPath())
	ipcB.RegisterPeer(common.NodeAlpha, ipcA.SocketPath())

	dA, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(common.NodeAlpha, "dual-a"), engine.NewMonotonicClock(),
	)
	dB, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(common.NodeBeta, "dual-b"), engine.NewMonotonicClock(),
	)
	defer func() {
		_ = dA.Close()
		_ = dB.Close()
	}()

	// Register RAM first, then IPC (both Rung2Host); cutting IPC must drop onto RAM!
	_ = dA.Ladder().RegisterRung(ramA)
	_ = dA.Ladder().RegisterRung(ipcA)
	_ = dB.Ladder().RegisterRung(ramB)
	_ = dB.Ladder().RegisterRung(ipcB)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = dA.Start(ctx)
	_ = dB.Start(ctx)

	// Mark IPC unhealthy; Ladder must seamlessly route through sibling RAM transceiver.
	ipcA.SetHealthy(false)
	if _, rung, err := dA.PingPeer(ctx, common.NodeBeta, []byte("RAM-RESCUE")); err != nil || rung != data.Rung2Host {
		t.Fatalf("[SEC 2] failover from degraded IPC to RAM failed: rung=%s err=%v", rung, err)
	}
	if dA.Ping().HealthSnapshot().LossPercent() != 0 {
		t.Fatalf("[SEC 2] expected 0%% loss during IPC->RAM failover, got %+v", dA.Ping().HealthSnapshot())
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: [ SECTION 3: INVARIANTS (CONCURRENT -RACE PING BURST) ]
// ------------------------------------------------------------------------------

func TestHostIPCMesh_Sec3_ConcurrentHostPingsUnderRaceDetector(t *testing.T) {
	bus := rung2_host.NewMemoryBus()
	dA, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(common.NodeAlpha, "race-a"), engine.NewMonotonicClock(),
	)
	dB, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(common.NodeBeta, "race-b"), engine.NewMonotonicClock(),
	)
	defer func() {
		_ = dA.Close()
		_ = dB.Close()
	}()
	_ = dA.Ladder().RegisterRung(bus.Attach(common.NodeAlpha))
	_ = dB.Ladder().RegisterRung(bus.Attach(common.NodeBeta))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = dA.Start(ctx)
	_ = dB.Start(ctx)

	const workers = 8
	var wg sync.WaitGroup
	errCh := make(chan error, workers)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, err := dA.PingPeer(ctx, common.NodeBeta, []byte("CONCURRENT-HOST"))
			if err != nil {
				errCh <- err
			}
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Fatalf("[SEC 3] concurrent host ping failed: %v", err)
	}

	h := dA.Ping().HealthSnapshot()
	if h.PingsSent != workers || h.PongsRecv != workers || h.LossPercent() != 0 {
		t.Fatalf("[SEC 3] expected %d/%d concurrent host pings, got %+v", workers, workers, h)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.4: [ SECTION 4: EMERGENCE (3-NODE BROADCAST DISCOVERY MESH) ]
// ------------------------------------------------------------------------------

func TestHostIPCMesh_Sec4_ThreeNodeBroadcastBeaconDiscoveryOnMemoryBus(t *testing.T) {
	bus := rung2_host.NewMemoryBus()
	dA, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(common.NodeAlpha, "mesh-a"), engine.NewMonotonicClock(),
	)
	dB, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(common.NodeBeta, "mesh-b"), engine.NewMonotonicClock(),
	)
	dR, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(common.NodeRelay, "mesh-relay"), engine.NewMonotonicClock(),
	)
	defer func() {
		_ = dA.Close()
		_ = dB.Close()
		_ = dR.Close()
	}()
	_ = dA.Ladder().RegisterRung(bus.Attach(common.NodeAlpha))
	_ = dB.Ladder().RegisterRung(bus.Attach(common.NodeBeta))
	_ = dR.Ladder().RegisterRung(bus.Attach(common.NodeRelay))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = dA.Start(ctx)
	_ = dB.Start(ctx)
	_ = dR.Start(ctx)

	// Broadcast beacon from NodeAlpha -> BOTH NodeBeta AND NodeRelay learn Alpha!
	beacon := common.MakeSovereignFrame(
		common.NodeAlpha, data.NodeIDBroadcast, data.Rung2Host, data.OpBeacon, 1, []byte("MESH-HELLO"),
	)
	if _, err := dA.SendFrame(&beacon); err != nil {
		t.Fatalf("[SEC 4] broadcast beacon failed: %v", err)
	}

	// Ping both Beta and Relay to confirm full 3-node mesh convergence.
	if _, _, err := dA.PingPeer(ctx, common.NodeBeta, []byte("SYNC-B")); err != nil {
		t.Fatalf("[SEC 4] ping Beta failed: %v", err)
	}
	if _, _, err := dA.PingPeer(ctx, common.NodeRelay, []byte("SYNC-R")); err != nil {
		t.Fatalf("[SEC 4] ping Relay failed: %v", err)
	}
	if dA.Router().ActiveCount() < 2 {
		t.Fatalf("[SEC 4] expected Alpha Router to track >= 2 peers, got %d", dA.Router().ActiveCount())
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Preflight Diagnostic Gates & Physical Source Anchor Verification )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-3 Integrated Verification Suite
// Module:        test.integrated.host_ipc_mesh
// Mission:       Assert Rung 2 Host & Orchestrator preflights and source anchors.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Block 4 Pre-Flight Contract Gates
// ------------------------------------------------------------------------------

func TestHostIPCMesh_Block4_PreflightGates(t *testing.T) {
	if hostIPCMeshSuiteName == "" {
		t.Fatal("[BLOCK 4] suite identifier must not be empty")
	}
	common.AssertAuditPass(t, rung2_host.PreflightMemoryBus(), "")
	common.AssertAuditPass(t, rung2_host.PreflightHostIPC(), "")
	common.AssertAuditPass(t, orchestrator.PreflightLadder(), "")
	common.AssertAuditPass(t, orchestrator.PreflightDispatcher(), "")
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Physical Source Anchor Verification
// ------------------------------------------------------------------------------

func TestHostIPCMesh_Block4_SourceAnchors(t *testing.T) {
	common.AssertRepoAnchors(t,
		"transport/rung2_host/memory.go",
		"transport/rung2_host/ipc.go",
		"orchestrator/ladder.go",
		"orchestrator/dispatcher.go",
	)
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Test Suite Epilogue & Clean Readiness Handshake )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-3 Integrated Verification Suite
// Module:        test.integrated.host_ipc_mesh
// Mission:       Confirm Rung 2 Host & Orchestrator epilogue readiness.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Epilogue Readiness Verification
// ------------------------------------------------------------------------------

func TestHostIPCMesh_Block5_EpilogueReadiness(t *testing.T) {
	if !rung2_host.IsMemoryBusReady() || !rung2_host.IsHostIPCReady() ||
		!orchestrator.IsLadderReady() || !orchestrator.IsDispatcherReady() ||
		!common.IsHarnessReady() {
		t.Fatal("[BLOCK 5: EPILOGUE] Rung 2 Host mesh integrated readiness failed")
	}
}

// End of tests/integrated/host_ipc_mesh_test.go
