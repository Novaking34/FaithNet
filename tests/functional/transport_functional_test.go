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
// @system        FaithNet Layer 3A Transport Functional Surface Suite
// @role          Tier-1 Functional Diagnostic Sensor (4-Section Micro Protocol)
// @maturity      Rev 5 Baseline | 4x4 Verification Matrix
// @chassis       bounded-heap | Exercises Cog 3 Transceiver & Rung 2 Host wires
// @scale         O(1) deterministic wire send/receive surface probes
// @seams         in: transport, transport/rung2_host, tests/common
// @invariants    1. Macro Tier 1 (Functional): verifies polymorphic Transceiver
//                   contract across Rung 2 RAM and Unix IPC mediums.
//                2. Zero leaked temporary sockets or goroutines after Close().
//                3. Enforces 512-byte Floppy Sector wire ceiling on every wire.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Tier-1 Functional surface suite exercising FaithNet's Cog 3 `Transceiver`
//   contract and Rung 2 Host mediums (`transport/transceiver.go`,
//   `transport/rung2_host/memory.go`, `transport/rung2_host/ipc.go`) across
//   all 4 ontological micro sections.
//
//   ### For Architects, Operators & Non-Programmers:
//   Verifies that two FaithNet nodes can talk cleanly over RAM channels and
//   local IPC mailboxes without any network card, and shut down cleanly.
//
//   ### For Systems Programmers & Engine Architects:
//   Exercises `transport.Transceiver` polymorphism, context cancellation,
//   512B oversize rejection, and Block 4 structured preflight gates.
// ==============================================================================

package functional_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Identifier, Verification Imports & Tier-1 Sensor Anchor )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-1 Functional Verification Suite
// Module:        test.functional.transport
// Mission:       Bind Cog 3 Transceiver contract and Rung 2 Host transports.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"context"
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
	"github.com/CreativeWorkzStudio/FaithNet/transport/rung2_host"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Operational Suite Constants & Identifiers
// ------------------------------------------------------------------------------

const (
	transportFunctionalSuiteName = "Tier 1: Layer 3A Transport Surface (Rung 2)"
)

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-1 Functional Verification Suite
// Module:        test.functional.transport
// Mission:       Verify polymorphic Transceiver behavior across Sections 1-4.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: [ SECTION 1: ALWAYS PASS (NOMINAL GOLDEN SURFACE COGS) ]
// ------------------------------------------------------------------------------

func TestTransportFunctional_Sec1_PolymorphicMemoryAndIPCSurface(t *testing.T) {
	// 1. Rung 2 MemoryBus nominal exchange via Transceiver interface.
	bus := rung2_host.NewMemoryBus()
	var txA transport.Transceiver = bus.Attach(common.NodeAlpha)
	var txB transport.Transceiver = bus.Attach(common.NodeBeta)
	defer func() {
		_ = txA.Close()
		_ = txB.Close()
	}()

	if txA.Rung() != data.Rung2Host || txA.Name() != "rung2:memory" || !txA.Healthy() {
		t.Fatalf("[SEC 1] unexpected MemoryTransceiver metadata: %s %s %v",
			txA.Name(), txA.Rung(), txA.Healthy())
	}
	if err := txA.Send(common.NodeBeta, []byte("MEM-HELLO")); err != nil {
		t.Fatalf("[SEC 1] MemoryTransceiver Send failed: %v", err)
	}
	env, err := txB.Receive(context.Background())
	if err != nil || string(env.Slice()) != "MEM-HELLO" {
		t.Fatalf("[SEC 1] MemoryTransceiver Receive mismatch: err=%v got=%q", err, env.Slice())
	}

	// 2. Rung 2 Unix IPC nominal exchange via Transceiver interface.
	tmp := t.TempDir()
	ipcA, err := rung2_host.NewIPCTransceiver(common.NodeAlpha, filepath.Join(tmp, "a.sock"))
	if err != nil {
		t.Fatalf("[SEC 1] NewIPCTransceiver A failed: %v", err)
	}
	defer func() { _ = ipcA.Close() }()

	ipcB, err := rung2_host.NewIPCTransceiver(common.NodeBeta, filepath.Join(tmp, "b.sock"))
	if err != nil {
		t.Fatalf("[SEC 1] NewIPCTransceiver B failed: %v", err)
	}
	defer func() { _ = ipcB.Close() }()

	ipcA.RegisterPeer(common.NodeBeta, ipcB.SocketPath())
	if err := ipcA.Send(common.NodeBeta, []byte("IPC-HELLO")); err != nil {
		t.Fatalf("[SEC 1] IPCTransceiver Send failed: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ipcEnv, err := ipcB.Receive(ctx)
	if err != nil || string(ipcEnv.Slice()) != "IPC-HELLO" {
		t.Fatalf("[SEC 1] IPCTransceiver Receive mismatch: err=%v got=%q", err, ipcEnv.Slice())
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: [ SECTION 2: NEVER FAIL (FAULT INVERSION & DEGRADATION) ]
// ------------------------------------------------------------------------------

func TestTransportFunctional_Sec2_OfflineClosedAndUnreachableRejection(t *testing.T) {
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
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: [ SECTION 3: INVARIANTS, SATURATION & LIMIT GUARDS ]
// ------------------------------------------------------------------------------

func TestTransportFunctional_Sec3_Wire512ByteCeilingAndEnvelopeClamp(t *testing.T) {
	bus := rung2_host.NewMemoryBus()
	a := bus.Attach(common.NodeAlpha)
	b := bus.Attach(common.NodeBeta)
	defer func() {
		_ = a.Close()
		_ = b.Close()
	}()

	var maxWire [data.FloorMaxFrameSize]byte
	if err := a.Send(common.NodeBeta, maxWire[:]); err != nil {
		t.Fatalf("[SEC 3] exact 512B wire frame rejected: %v", err)
	}
	var overWire [data.FloorMaxFrameSize + 1]byte
	if err := a.Send(common.NodeBeta, overWire[:]); err != transport.ErrWireFrameOversize {
		t.Fatalf("[SEC 3] 513B wire frame bypassed 512B ceiling: %v", err)
	}

	// WireEnvelope.Slice() saturates malformed negative or oversize Length safely.
	neg := transport.WireEnvelope{Length: -5}
	if len(neg.Slice()) != 0 {
		t.Fatalf("[SEC 3] negative WireEnvelope.Length did not clamp to 0")
	}
	big := transport.WireEnvelope{Length: 9999}
	if len(big.Slice()) != data.FloorMaxFrameSize {
		t.Fatalf("[SEC 3] oversize WireEnvelope.Length did not clamp to 512")
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
	if err := wireA.Send(common.NodeBeta, buf[:n]); err != nil {
		t.Fatalf("[SEC 4] Send ping failed: %v", err)
	}

	clk.Advance(300)
	envIn, err := wireB.Receive(context.Background())
	if err != nil || engine.DecodeFrom(envIn.Slice(), &decodedReq) != nil {
		t.Fatalf("[SEC 4] B receive/decode failed: %v", err)
	}
	_ = engine.ReflectPingInto(&decodedReq, &resp, common.NodeBeta)
	n2, _ := engine.EncodeInto(buf[:], &resp)
	if err := wireB.Send(common.NodeAlpha, buf[:n2]); err != nil {
		t.Fatalf("[SEC 4] Send pong failed: %v", err)
	}

	clk.Advance(300)
	envOut, err := wireA.Receive(context.Background())
	if err != nil || engine.DecodeFrom(envOut.Slice(), &decodedResp) != nil {
		t.Fatalf("[SEC 4] A receive/decode failed: %v", err)
	}
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
// Mission:       Assert Rung 2 Memory & IPC structured preflights and anchors.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Preflight Diagnostic Gates
// ------------------------------------------------------------------------------

func TestTransportFunctional_Block4_PreflightRung2Host(t *testing.T) {
	common.AssertAuditPass(t, rung2_host.PreflightMemoryBus(), "")
	common.AssertAuditPass(t, rung2_host.PreflightHostIPC(), "")
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
		"transport/rung2_host/memory.go",
		"transport/rung2_host/ipc.go",
	)
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Test Suite Epilogue & Clean Readiness Handshake )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-1 Functional Verification Suite
// Module:        test.functional.transport
// Mission:       Verify Rung 2 Host readiness probes post-suite.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Epilogue & Clean State Handshake
// ------------------------------------------------------------------------------

func TestTransportFunctional_Block5_EpilogueReadiness(t *testing.T) {
	if !rung2_host.IsMemoryBusReady() || !rung2_host.IsHostIPCReady() ||
		!common.IsHarnessReady() {
		t.Fatal("[BLOCK 5: EPILOGUE] Rung 2 transport readiness handshake failed")
	}
}

// End of tests/functional/transport_functional_test.go
