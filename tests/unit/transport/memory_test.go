// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/unit/transport/memory_test.go
// @id            test.unit.transport.memory
// @layer         Layer 4: Tier-2 Unit Verification Suite (Transport Sub-Stratum)
// @archetype     verification_suite | unit_isolation
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Rung 2 Host In-Memory RAM Bus Unit Suite
// @role          Tier-2 Unit Isolation Gate for cog.transport.rung2_memory
// @maturity      Rev 5 Baseline | 4x4 Verification Matrix
// @chassis       bounded-heap | Verifies 64-slot inbox saturation & -race safety
// @scale         O(1) point-to-point & broadcast RAM channel verification
// @seams         in: data, transport, rung2_host, common | target: memory.go
// @invariants    1. Unicast & broadcast delivery operate cleanly under `-race`.
//                2. Full 64-slot inbox fails non-blockingly without deadlock.
//                3. Idempotent Close() detaches cleanly from MemoryBus.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Dedicated 1:1 Tier-2 Unit isolation suite for `transport/rung2_host/memory.go`.
//
//   ### For Architects, Operators & Non-Programmers:
//   Proves FaithNet's zero-NIC in-RAM virtual switchboard delivers packets
//   cleanly across local nodes and handles full queues or shutdowns gracefully.
//
//   ### For Systems Programmers & Engine Architects:
//   100% branch coverage of `MemoryBus` and `MemoryTransceiver` under `-race`.
// ==============================================================================

package transport_unit_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Identifier, Verification Imports & Tier-2 Unit Sensor Anchor )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.memory
// Mission:       Bind Rung 2 MemoryBus/MemoryTransceiver and shared harness.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"context"
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
	memoryUnitSuiteName = "Tier 2 Unit: cog.transport.rung2_memory (RAM Bus)"
)

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.memory
// Mission:       Verify unicast/broadcast, fault bounds, 64-cap, and -race storm.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: [ SECTION 1: ALWAYS PASS (NOMINAL GOLDEN SURFACE COGS) ]
// ------------------------------------------------------------------------------

func TestMemoryUnit_Sec1_UnicastBroadcastAndMetadata(t *testing.T) {
	bus := rung2_host.NewMemoryBus()
	a := bus.Attach(common.NodeAlpha)
	b := bus.Attach(common.NodeBeta)
	c := bus.Attach(common.NodeRelay)
	defer func() {
		_ = a.Close()
		_ = b.Close()
		_ = c.Close()
	}()

	if a.Name() != "rung2:memory" || a.Rung() != data.Rung2Host || !a.Healthy() {
		t.Fatal("[SEC 1] unexpected MemoryTransceiver metadata")
	}
	if err := a.Send(data.NodeIDBroadcast, []byte("BEACON-RAM")); err != nil {
		t.Fatalf("[SEC 1] broadcast failed: %v", err)
	}
	for _, rx := range []*rung2_host.MemoryTransceiver{b, c} {
		env, err := rx.Receive(context.Background())
		if err != nil || env.SenderID != common.NodeAlpha || string(env.Slice()) != "BEACON-RAM" {
			t.Fatalf("[SEC 1] broadcast recv mismatch: err=%v got=%q", err, env.Slice())
		}
	}
	common.AssertAuditPass(t, a.AuditState(), a.ExecuteMemoryDumpCard())
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: [ SECTION 2: NEVER FAIL (FAULT INVERSION & DEGRADATION) ]
// ------------------------------------------------------------------------------

func TestMemoryUnit_Sec2_OfflineUnhealthyClosedAndNilAudit(t *testing.T) {
	var nilMem *rung2_host.MemoryTransceiver
	common.AssertAuditFail(t, nilMem.AuditState(), data.AuditSeverityCritical)

	bus := rung2_host.NewMemoryBus()
	a := bus.Attach(common.NodeAlpha)
	b := bus.Attach(common.NodeBeta)
	defer func() { _ = b.Close() }()

	b.SetHealthy(false)
	if err := a.Send(common.NodeBeta, []byte("DROP")); err != transport.ErrPeerUnreachable {
		t.Fatalf("[SEC 2] unhealthy peer got %v", err)
	}
	a.SetHealthy(false)
	if err := a.Send(common.NodeBeta, []byte("OFF")); err != transport.ErrTransceiverOffline {
		t.Fatalf("[SEC 2] offline sender got %v", err)
	}
	a.SetHealthy(true)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := a.Receive(ctx); err != context.Canceled {
		t.Fatalf("[SEC 2] canceled Receive got %v", err)
	}
	_ = a.Close()
	_ = a.Close()
	if _, err := a.Receive(context.Background()); err != transport.ErrTransceiverClosed {
		t.Fatalf("[SEC 2] closed Receive got %v", err)
	}
	if err := a.Send(common.NodeBeta, []byte("CLOSED")); err != transport.ErrTransceiverClosed {
		t.Fatalf("[SEC 2] closed Send got %v", err)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: [ SECTION 3: INVARIANTS, SATURATION & LIMIT GUARDS ]
// ------------------------------------------------------------------------------

func TestMemoryUnit_Sec3_64SlotInboxSaturationAnd512ByteCeiling(t *testing.T) {
	bus := rung2_host.NewMemoryBus()
	a := bus.Attach(common.NodeAlpha)
	b := bus.Attach(common.NodeBeta)
	defer func() {
		_ = a.Close()
		_ = b.Close()
	}()

	for i := 0; i < rung2_host.DefaultMemoryQueueDepth; i++ {
		if err := a.Send(common.NodeBeta, []byte("FILL")); err != nil {
			t.Fatalf("[SEC 3] fill %d failed: %v", i, err)
		}
	}
	if err := a.Send(common.NodeBeta, []byte("OVER")); err != transport.ErrPeerUnreachable {
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

func TestMemoryUnit_Sec4_ConcurrentEchoStormAndDumpCard(t *testing.T) {
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
			t.Fatalf("[SEC 4] echo %d mismatch: err=%v", i, err)
		}
	}
	wg.Wait()
	if !strings.Contains(a.ExecuteMemoryDumpCard(), "50 sent / 50 recv") {
		t.Fatalf("[SEC 4] unexpected dump card:\n%s", a.ExecuteMemoryDumpCard())
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Preflight Diagnostic Gates & Physical Source Anchor Verification )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.memory
// Mission:       Run PreflightMemoryBus and verify physical source anchor.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Preflight Diagnostic Gates
// ------------------------------------------------------------------------------

func TestMemoryUnit_Block4_PreflightGate(t *testing.T) {
	common.AssertAuditPass(t, rung2_host.PreflightMemoryBus(), "")
	if !rung2_host.AuditMemoryBus() {
		t.Fatal("[BLOCK 4] AuditMemoryBus returned false")
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Suite Identifier & Physical Source Anchor Gate
// ------------------------------------------------------------------------------

func TestMemoryUnit_Block4_SourceAnchor(t *testing.T) {
	if memoryUnitSuiteName == "" {
		t.Fatal("[BLOCK 4] suite identifier must not be empty")
	}
	common.AssertRepoAnchors(t, "transport/rung2_host/memory.go")
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Test Suite Epilogue & Clean Readiness Handshake )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.memory
// Mission:       Confirm Rung 2 MemoryBus readiness handshake post-suite.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Epilogue & Clean State Handshake
// ------------------------------------------------------------------------------

func TestMemoryUnit_Block5_EpilogueReadiness(t *testing.T) {
	if !rung2_host.IsMemoryBusReady() || !common.IsHarnessReady() {
		t.Fatal("[BLOCK 5: EPILOGUE] memory unit readiness handshake failed")
	}
}

// End of tests/unit/transport/memory_test.go
