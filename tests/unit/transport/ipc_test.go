// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/unit/transport/ipc_test.go
// @id            test.unit.transport.ipc
// @layer         Layer 4: Tier-2 Unit Verification Suite (Transport Sub-Stratum)
// @archetype     verification_suite | unit_isolation
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Rung 2 Host Unixgram IPC Transceiver Unit Suite
// @role          Tier-2 Unit Isolation Gate for cog.transport.rung2_ipc
// @maturity      Rev 5 Baseline | 4x4 Verification Matrix
// @chassis       bounded-heap | Verifies unixgram sockets & zero disk leaks
// @scale         O(1) bidirectional cross-process datagram tests
// @seams         in: data, transport, rung2_host, common | target: ipc.go
// @invariants    1. Bidirectional unixgram exchange runs cleanly in TempDir.
//                2. Close() unlinks socket file from disk cleanly (0 leaks).
//                3. Enforces 512-byte Floppy Sector wire ceiling strictly.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Dedicated 1:1 Tier-2 Unit isolation suite for `transport/rung2_host/ipc.go`.
//
//   ### For Architects, Operators & Non-Programmers:
//   Proves separate FaithNet processes on the same computer talk over local
//   IPC mailboxes cleanly and leave zero leftover socket files behind.
//
//   ### For Systems Programmers & Engine Architects:
//   100% branch coverage of `IPCTransceiver` under `go test -race`.
// ==============================================================================

package transport_unit_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Identifier, Verification Imports & Tier-2 Unit Sensor Anchor )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.ipc
// Mission:       Bind Rung 2 IPCTransceiver and shared test harness.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// --- Tier 2: External Library Dependencies (None) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---
import (
	"github.com/CreativeWorkzStudio/FaithNet/data"
	"github.com/CreativeWorkzStudio/FaithNet/tests/common"
	"github.com/CreativeWorkzStudio/FaithNet/transport"
	"github.com/CreativeWorkzStudio/FaithNet/transport/rung2_host"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Operational Suite Constants & Identifiers
// ------------------------------------------------------------------------------

const (
	ipcUnitSuiteName = "Tier 2 Unit: cog.transport.rung2_ipc (Unixgram IPC)"
)

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.ipc
// Mission:       Verify bidirectional IPC, fault bounds, 512B cap & cleanup.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: [ SECTION 1: ALWAYS PASS (NOMINAL GOLDEN SURFACE COGS) ]
// ------------------------------------------------------------------------------

func TestIPCUnit_Sec1_BidirectionalUnixgramExchangeAndMetadata(t *testing.T) {
	tmp := t.TempDir()
	a, err := rung2_host.NewIPCTransceiver(common.NodeAlpha, filepath.Join(tmp, "a.sock"))
	if err != nil {
		t.Fatalf("[SEC 1] bind A failed: %v", err)
	}
	defer func() { _ = a.Close() }()

	b, err := rung2_host.NewIPCTransceiver(common.NodeBeta, filepath.Join(tmp, "b.sock"))
	if err != nil {
		t.Fatalf("[SEC 1] bind B failed: %v", err)
	}
	defer func() { _ = b.Close() }()

	if a.Name() != "rung2:ipc" || a.Rung() != data.Rung2Host || !a.Healthy() {
		t.Fatal("[SEC 1] unexpected IPCTransceiver metadata")
	}
	a.RegisterPeer(common.NodeBeta, b.SocketPath())
	b.RegisterPeer(common.NodeAlpha, a.SocketPath())

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := a.Send(common.NodeBeta, []byte("IPC-PING")); err != nil {
		t.Fatalf("[SEC 1] A->B failed: %v", err)
	}
	envB, err := b.Receive(ctx)
	if err != nil || string(envB.Slice()) != "IPC-PING" {
		t.Fatalf("[SEC 1] B recv mismatch: err=%v got=%q", err, envB.Slice())
	}

	if err := b.Send(common.NodeAlpha, []byte("IPC-PONG")); err != nil {
		t.Fatalf("[SEC 1] B->A failed: %v", err)
	}
	envA, err := a.Receive(ctx)
	if err != nil || string(envA.Slice()) != "IPC-PONG" {
		t.Fatalf("[SEC 1] A recv mismatch: err=%v got=%q", err, envA.Slice())
	}
	common.AssertAuditPass(t, a.AuditState(), a.ExecuteIPCDumpCard())
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: [ SECTION 2: NEVER FAIL (FAULT INVERSION & DEGRADATION) ]
// ------------------------------------------------------------------------------

func TestIPCUnit_Sec2_UnregisteredOfflineCanceledClosedAndSocketCleanup(t *testing.T) {
	var nilIPC *rung2_host.IPCTransceiver
	common.AssertAuditFail(t, nilIPC.AuditState(), data.AuditSeverityCritical)

	sock := filepath.Join(t.TempDir(), "solo.sock")
	u, err := rung2_host.NewIPCTransceiver(common.NodeAlpha, sock)
	if err != nil {
		t.Fatalf("[SEC 2] bind solo IPC failed: %v", err)
	}
	if err := u.Send(common.NodeBeta, []byte("NO-PEER")); err != transport.ErrPeerUnreachable {
		t.Fatalf("[SEC 2] unregistered peer got %v", err)
	}
	u.SetHealthy(false)
	if err := u.Send(common.NodeBeta, []byte("OFF")); err != transport.ErrTransceiverOffline {
		t.Fatalf("[SEC 2] offline IPC got %v", err)
	}
	u.SetHealthy(true)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := u.Receive(ctx); err != context.Canceled {
		t.Fatalf("[SEC 2] canceled Receive got %v", err)
	}
	_ = u.Close()
	_ = u.Close()
	if err := u.Send(common.NodeBeta, []byte("CLOSED")); err != transport.ErrTransceiverClosed {
		t.Fatalf("[SEC 2] closed IPC Send got %v", err)
	}
	if _, statErr := os.Stat(sock); !os.IsNotExist(statErr) {
		t.Fatal("[SEC 2] IPC Close() leaked socket file on disk")
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: [ SECTION 3: INVARIANTS, SATURATION & LIMIT GUARDS ]
// ------------------------------------------------------------------------------

func TestIPCUnit_Sec3_512ByteCeilingEnforced(t *testing.T) {
	u, _ := rung2_host.NewIPCTransceiver(common.NodeAlpha, filepath.Join(t.TempDir(), "cap.sock"))
	defer func() { _ = u.Close() }()

	var over [data.FloorMaxFrameSize + 1]byte
	if err := u.Send(common.NodeBeta, over[:]); err != transport.ErrWireFrameOversize {
		t.Fatalf("[SEC 3] 513B datagram bypassed IPC ceiling: %v", err)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.4: [ SECTION 4: EMERGENCE & COMPOSITE WORKFLOWS ]
// ------------------------------------------------------------------------------

func TestIPCUnit_Sec4_BroadcastAndDumpCardTelemetry(t *testing.T) {
	tmp := t.TempDir()
	a, _ := rung2_host.NewIPCTransceiver(common.NodeAlpha, filepath.Join(tmp, "a.sock"))
	b, _ := rung2_host.NewIPCTransceiver(common.NodeBeta, filepath.Join(tmp, "b.sock"))
	defer func() {
		_ = a.Close()
		_ = b.Close()
	}()
	a.RegisterPeer(common.NodeBeta, b.SocketPath())
	if err := a.Send(data.NodeIDBroadcast, []byte("IPC-BCAST")); err != nil {
		t.Fatalf("[SEC 4] IPC broadcast failed: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if env, err := b.Receive(ctx); err != nil || string(env.Slice()) != "IPC-BCAST" {
		t.Fatalf("[SEC 4] IPC broadcast recv mismatch: err=%v", err)
	}
	if !strings.Contains(a.ExecuteIPCDumpCard(), "1 sent / 0 recv") {
		t.Fatalf("[SEC 4] unexpected IPC dump card:\n%s", a.ExecuteIPCDumpCard())
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Preflight Diagnostic Gates & Physical Source Anchor Verification )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.ipc
// Mission:       Run PreflightHostIPC and verify physical source anchor.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Preflight Diagnostic Gates
// ------------------------------------------------------------------------------

func TestIPCUnit_Block4_PreflightGate(t *testing.T) {
	common.AssertAuditPass(t, rung2_host.PreflightHostIPC(), "")
	if !rung2_host.AuditHostIPC() {
		t.Fatal("[BLOCK 4] AuditHostIPC returned false")
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Suite Identifier & Physical Source Anchor Gate
// ------------------------------------------------------------------------------

func TestIPCUnit_Block4_SourceAnchor(t *testing.T) {
	if ipcUnitSuiteName == "" {
		t.Fatal("[BLOCK 4] suite identifier must not be empty")
	}
	common.AssertRepoAnchors(t, "transport/rung2_host/ipc.go")
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Test Suite Epilogue & Clean Readiness Handshake )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.ipc
// Mission:       Confirm Rung 2 Host IPC readiness handshake post-suite.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Epilogue & Clean State Handshake
// ------------------------------------------------------------------------------

func TestIPCUnit_Block5_EpilogueReadiness(t *testing.T) {
	if !rung2_host.IsHostIPCReady() || !common.IsHarnessReady() {
		t.Fatal("[BLOCK 5: EPILOGUE] ipc unit readiness handshake failed")
	}
}

// End of tests/unit/transport/ipc_test.go
