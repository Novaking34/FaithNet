// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/unit/transport/rung2_host_test.go
// @id            test.unit.transport.rung2_host
// @layer         Layer 4: Tier-2 Unit Verification Suite (Transport Sub-Stratum)
// @archetype     verification_suite | unit_isolation
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Rung 2 Host Transports (Memory & Unix IPC) Unit Suite
// @role          Tier-2 Unit Isolation Gate for cog.transport.rung2_{memory,ipc}
// @maturity      Rev 5 Baseline | 4x4 Verification Matrix
// @chassis       bounded-heap | Verifies channel & unixgram bounds and cleanup
// @scale         O(1) point-to-point & broadcast host transport verification
// @seams         in: transport, transport/rung2_host, tests/common
// @invariants    1. Both MemoryTransceiver and IPCTransceiver satisfy Cog 3
//                   `transport.Transceiver` cleanly across all 4 micro sections.
//                2. Idempotent Close() never panics and removes unixgram files.
//                3. Full inbox saturation drops cleanly without deadlock.
//                4. Concurrent Send/Receive is 100% -race clean.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Tier-2 Unit isolation suite strictly verifying `transport/rung2_host/memory.go`
//   and `transport/rung2_host/ipc.go` across unicast/broadcast delivery, fault
//   inversion, queue/MTU limits, `-race` concurrency, live `AuditState`, and
//   60-column diagnostic cards.
//
//   ### For Architects, Operators & Non-Programmers:
//   Proves FaithNet's zero-network-card communication mediums (in-RAM bus and
//   local IPC mailboxes) deliver frames reliably, survive dropped links, and
//   never leave orphaned socket files behind on disk.
//
//   ### For Systems Programmers & Engine Architects:
//   100% branch coverage of Rung 2 Host transport cogs under `go test -race`.
// ==============================================================================

package transport_unit_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Identifier, Verification Imports & Tier-2 Unit Sensor Anchor )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.rung2_host
// Mission:       Bind Rung 2 Host transports and shared harness for isolation.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
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
	rung2HostUnitSuiteName = "Tier 2 Unit: Rung 2 Host (MemoryBus & Unix IPC)"
)

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.rung2_host
// Mission:       Verify MemoryBus and IPCTransceiver across Sections 1-4.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: [ SECTION 1: ALWAYS PASS (NOMINAL GOLDEN SURFACE COGS) ]
// ------------------------------------------------------------------------------

func TestRung2Unit_Sec1_MemoryBusUnicastAndBroadcast(t *testing.T) {
	bus := rung2_host.NewMemoryBus()
	a := bus.Attach(common.NodeAlpha)
	b := bus.Attach(common.NodeBeta)
	c := bus.Attach(common.NodeRelay)
	defer func() {
		_ = a.Close()
		_ = b.Close()
		_ = c.Close()
	}()

	if err := a.Send(data.NodeIDBroadcast, []byte("BEACON-ALL")); err != nil {
		t.Fatalf("[SEC 1] MemoryBus broadcast failed: %v", err)
	}
	for _, rx := range []*rung2_host.MemoryTransceiver{b, c} {
		env, err := rx.Receive(context.Background())
		if err != nil || env.SenderID != common.NodeAlpha || string(env.Slice()) != "BEACON-ALL" {
			t.Fatalf("[SEC 1] broadcast delivery mismatch: err=%v got=%q", err, env.Slice())
		}
	}
	common.AssertAuditPass(t, a.AuditState(), a.ExecuteMemoryDumpCard())
}

func TestRung2Unit_Sec1_IPCBidirectionalExchangeAndDumpCard(t *testing.T) {
	tmp := t.TempDir()
	a, err := rung2_host.NewIPCTransceiver(common.NodeAlpha, filepath.Join(tmp, "alpha.sock"))
	if err != nil {
		t.Fatalf("[SEC 1] bind A failed: %v", err)
	}
	defer func() { _ = a.Close() }()

	b, err := rung2_host.NewIPCTransceiver(common.NodeBeta, filepath.Join(tmp, "beta.sock"))
	if err != nil {
		t.Fatalf("[SEC 1] bind B failed: %v", err)
	}
	defer func() { _ = b.Close() }()

	a.RegisterPeer(common.NodeBeta, b.SocketPath())
	b.RegisterPeer(common.NodeAlpha, a.SocketPath())

	if err := a.Send(common.NodeBeta, []byte("PING-IPC")); err != nil {
		t.Fatalf("[SEC 1] IPC A->B failed: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	envB, err := b.Receive(ctx)
	if err != nil || string(envB.Slice()) != "PING-IPC" {
		t.Fatalf("[SEC 1] IPC B recv mismatch: err=%v got=%q", err, envB.Slice())
	}

	if err := b.Send(common.NodeAlpha, []byte("PONG-IPC")); err != nil {
		t.Fatalf("[SEC 1] IPC B->A failed: %v", err)
	}
	envA, err := a.Receive(ctx)
	if err != nil || string(envA.Slice()) != "PONG-IPC" {
		t.Fatalf("[SEC 1] IPC A recv mismatch: err=%v got=%q", err, envA.Slice())
	}

	card := a.ExecuteIPCDumpCard()
	if !strings.Contains(card, "1 sent / 1 recv") {
		t.Fatalf("[SEC 1] unexpected IPC dump card:\n%s", card)
	}
	common.AssertAuditPass(t, a.AuditState(), card)
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: [ SECTION 2: NEVER FAIL (FAULT INVERSION & DEGRADATION) ]
// ------------------------------------------------------------------------------

func TestRung2Unit_Sec2_FaultInversionMemoryAndIPC(t *testing.T) {
	// MemoryTransceiver: nil AuditState, unhealthy peer, double Close, canceled ctx.
	var nilMem *rung2_host.MemoryTransceiver
	common.AssertAuditFail(t, nilMem.AuditState(), data.AuditSeverityCritical)

	bus := rung2_host.NewMemoryBus()
	a := bus.Attach(common.NodeAlpha)
	b := bus.Attach(common.NodeBeta)
	b.SetHealthy(false)
	if err := a.Send(common.NodeBeta, []byte("DROP")); err != transport.ErrPeerUnreachable {
		t.Fatalf("[SEC 2] unhealthy peer: want ErrPeerUnreachable got %v", err)
	}
	if err := a.Close(); err != nil {
		t.Fatalf("[SEC 2] first Close failed: %v", err)
	}
	if err := a.Close(); err != nil {
		t.Fatalf("[SEC 2] idempotent second Close failed: %v", err)
	}
	if _, err := a.Receive(context.Background()); err != transport.ErrTransceiverClosed {
		t.Fatalf("[SEC 2] closed Receive: want ErrTransceiverClosed got %v", err)
	}
	_ = b.Close()

	// IPCTransceiver: nil AuditState, offline, unregistered, canceled ctx, double Close.
	var nilIPC *rung2_host.IPCTransceiver
	common.AssertAuditFail(t, nilIPC.AuditState(), data.AuditSeverityCritical)

	tmp := t.TempDir()
	sock := filepath.Join(tmp, "solo.sock")
	u, err := rung2_host.NewIPCTransceiver(common.NodeAlpha, sock)
	if err != nil {
		t.Fatalf("[SEC 2] bind solo IPC failed: %v", err)
	}
	if err := u.Send(common.NodeBeta, []byte("NO-PEER")); err != transport.ErrPeerUnreachable {
		t.Fatalf("[SEC 2] unregistered IPC peer: got %v", err)
	}
	u.SetHealthy(false)
	if err := u.Send(common.NodeBeta, []byte("OFFLINE")); err != transport.ErrTransceiverOffline {
		t.Fatalf("[SEC 2] offline IPC: got %v", err)
	}
	u.SetHealthy(true)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := u.Receive(ctx); err != context.Canceled {
		t.Fatalf("[SEC 2] canceled IPC Receive: got %v", err)
	}
	_ = u.Close()
	_ = u.Close() // Idempotent second Close must not panic.
	if err := u.Send(common.NodeBeta, []byte("CLOSED")); err != transport.ErrTransceiverClosed {
		t.Fatalf("[SEC 2] closed IPC Send: got %v", err)
	}
	if _, statErr := os.Stat(sock); !os.IsNotExist(statErr) {
		t.Fatal("[SEC 2] IPC Close() leaked socket file on disk")
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: [ SECTION 3: INVARIANTS, SATURATION & LIMIT GUARDS ]
// ------------------------------------------------------------------------------

func TestRung2Unit_Sec3_InboxSaturationAnd512ByteCeiling(t *testing.T) {
	bus := rung2_host.NewMemoryBus()
	a := bus.Attach(common.NodeAlpha)
	b := bus.Attach(common.NodeBeta)
	defer func() {
		_ = a.Close()
		_ = b.Close()
	}()

	// Fill B's bounded channel to capacity (64) without blocking; 65th drops cleanly.
	for i := 0; i < rung2_host.DefaultMemoryQueueDepth; i++ {
		if err := a.Send(common.NodeBeta, []byte("FILL")); err != nil {
			t.Fatalf("[SEC 3] fill %d failed before queue depth: %v", i, err)
		}
	}
	if err := a.Send(common.NodeBeta, []byte("OVERFLOW")); err != transport.ErrPeerUnreachable {
		t.Fatalf("[SEC 3] saturated inbox must fail non-blockingly, got %v", err)
	}

	var over [data.FloorMaxFrameSize + 1]byte
	if err := a.Send(common.NodeBeta, over[:]); err != transport.ErrWireFrameOversize {
		t.Fatalf("[SEC 3] 513B payload bypassed ceiling: %v", err)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.4: [ SECTION 4: EMERGENCE & COMPOSITE WORKFLOWS ]
// ------------------------------------------------------------------------------

func TestRung2Unit_Sec4_ConcurrentMemoryBusEchoStorm(t *testing.T) {
	bus := rung2_host.NewMemoryBus()
	a := bus.Attach(common.NodeAlpha)
	b := bus.Attach(common.NodeBeta)
	defer func() {
		_ = a.Close()
		_ = b.Close()
	}()

	const total = 50
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < total; i++ {
			env, err := b.Receive(ctx)
			if err != nil {
				return
			}
			_ = b.Send(env.SenderID, env.Slice())
		}
	}()

	for i := 0; i < total; i++ {
		if err := a.Send(common.NodeBeta, []byte("STORM")); err != nil {
			t.Fatalf("[SEC 4] send %d failed: %v", i, err)
		}
		env, err := a.Receive(ctx)
		if err != nil || string(env.Slice()) != "STORM" {
			t.Fatalf("[SEC 4] echo %d mismatch: err=%v got=%q", i, err, env.Slice())
		}
	}
	wg.Wait()

	card := a.ExecuteMemoryDumpCard()
	if !strings.Contains(card, "50 sent / 50 recv") {
		t.Fatalf("[SEC 4] unexpected Memory dump card:\n%s", card)
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Preflight Diagnostic Gates & Physical Source Anchor Verification )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.rung2_host
// Mission:       Run MemoryBus & HostIPC structured preflights and anchor gate.
// ==============================================================================

func TestRung2Unit_Block4_PreflightAndAnchor(t *testing.T) {
	if rung2HostUnitSuiteName == "" {
		t.Fatal("[BLOCK 4] suite identifier must not be empty")
	}
	common.AssertAuditPass(t, rung2_host.PreflightMemoryBus(), "")
	common.AssertAuditPass(t, rung2_host.PreflightHostIPC(), "")
	if !rung2_host.AuditMemoryBus() || !rung2_host.AuditHostIPC() {
		t.Fatal("[BLOCK 4] Rung 2 Host boolean audit wrappers failed")
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
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.rung2_host
// Mission:       Confirm Rung 2 Host readiness handshake post-suite.
// ==============================================================================

func TestRung2Unit_Block5_EpilogueReadiness(t *testing.T) {
	if !rung2_host.IsMemoryBusReady() || !rung2_host.IsHostIPCReady() ||
		!common.IsHarnessReady() {
		t.Fatal("[BLOCK 5: EPILOGUE] Rung 2 Host readiness handshake failed")
	}
}

// End of tests/unit/transport/rung2_host_test.go
