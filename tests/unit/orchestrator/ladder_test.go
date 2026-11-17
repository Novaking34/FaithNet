// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/unit/orchestrator/ladder_test.go
// @id            test.unit.orchestrator.ladder
// @layer         Layer 4: Tier-2 Unit Verification Suite (Orchestrator)
// @archetype     verification_suite | unit_isolation
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Layer 2 4-Rung Adaptive Ladder Unit Suite
// @role          Tier-2 Unit Isolation Gate for cog.orchestrator.ladder
// @maturity      Rev 5 Baseline | 4x4 Verification Matrix
// @chassis       bounded-heap | Verifies 4-rung failover & 64-entry ring wrap
// @scale         O(1) deterministic cascade transitions & ring extraction
// @seams         in: data, engine, orchestrator, transport | target: ladder.go
// @invariants    1. Rising Up respects offline-first policy ceilings.
//                2. Falling Down executes EVT_CASCADE_DOWN deterministically.
//                3. Cascade flight-recorder ring wraps cleanly past 64 events.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Tier-2 Unit isolation suite strictly verifying `orchestrator/ladder.go`
//   with injectable mock survival rungs across Rung 1..4, history ring wrap
//   past `MaxCascadeHistory` (64), policy enforcement, and live `AuditState`.
//
//   ### For Architects, Operators & Non-Programmers:
//   Tests every climb-up and fall-down gear shift of the 4-Rung Adaptive Ladder
//   in isolation, proving the flight recorder never overflows memory.
//
//   ### For Systems Programmers & Engine Architects:
//   100% branch coverage of `RegisterRung`, `EvaluateAndAdapt`, `CascadeDown`,
//   `SendFrame`, `HistorySnapshotInto`, `AuditState`, and `Close`.
// ==============================================================================

package orchestrator_unit_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Identifier, Imports & Injectable Multi-Rung Test Transceiver )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.orchestrator.ladder
// Mission:       Provide a deterministic controllable mock Transceiver for
//                isolating every AdaptiveLadder gear shift across Rung 1..4.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
)

// --- Tier 2: External Library Dependencies (None) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---
import (
	"github.com/CreativeWorkzStudio/FaithNet/data"
	"github.com/CreativeWorkzStudio/FaithNet/engine"
	"github.com/CreativeWorkzStudio/FaithNet/orchestrator"
	"github.com/CreativeWorkzStudio/FaithNet/tests/common"
	"github.com/CreativeWorkzStudio/FaithNet/transport"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Operational Suite Constants & Identifiers
// ------------------------------------------------------------------------------

const (
	ladderUnitSuiteName = "Tier 2 Unit: cog.orchestrator.ladder (4-Rung Ladder)"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.2: Controllable Multi-Rung Stub Transceiver
// ------------------------------------------------------------------------------

type stubRungTx struct {
	name     string
	rung     data.RungLevel
	healthy  atomic.Bool
	failSend atomic.Bool
	sentCnt  atomic.Int64
	closed   atomic.Bool
}

func newStubRung(name string, rung data.RungLevel) *stubRungTx {
	s := &stubRungTx{name: name, rung: rung}
	s.healthy.Store(true)
	return s
}

func (s *stubRungTx) Name() string          { return s.name }
func (s *stubRungTx) Rung() data.RungLevel  { return s.rung }
func (s *stubRungTx) Healthy() bool         { return !s.closed.Load() && s.healthy.Load() }
func (s *stubRungTx) Close() error          { s.closed.Store(true); s.healthy.Store(false); return nil }
func (s *stubRungTx) Receive(ctx context.Context) (transport.WireEnvelope, error) {
	<-ctx.Done()
	return transport.WireEnvelope{}, ctx.Err()
}
func (s *stubRungTx) Send(_ data.NodeID, raw []byte) error {
	if s.closed.Load() {
		return transport.ErrTransceiverClosed
	}
	if !s.healthy.Load() {
		return transport.ErrTransceiverOffline
	}
	if len(raw) > data.FloorMaxFrameSize {
		return transport.ErrWireFrameOversize
	}
	if s.failSend.Load() {
		return errors.New("simulated mid-flight wire snap")
	}
	s.sentCnt.Add(1)
	return nil
}

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.orchestrator.ladder
// Mission:       Verify Rising Up, EVT_CASCADE_DOWN, mid-flight failover, and
//                64-slot flight-recorder wrap-around.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: [ SECTION 1: ALWAYS PASS (NOMINAL GOLDEN SURFACE COGS) ]
// ------------------------------------------------------------------------------

func TestLadderUnit_Sec1_RisingUpAndMidFlightCascadeDown(t *testing.T) {
	clk := engine.NewMockClock(100)
	manifest := data.DefaultNodeManifest(common.NodeAlpha, "adaptive")
	manifest.AllowGlobalEgress = true
	manifest.PreferredMaxRung = data.Rung4Global

	lad := orchestrator.NewAdaptiveLadder(manifest, clk)
	defer func() { _ = lad.Close() }()

	r1 := newStubRung("rung1:morse", data.Rung1Floor)
	r2 := newStubRung("rung2:memory", data.Rung2Host)
	r3 := newStubRung("rung3:udp", data.Rung3Local)
	r4 := newStubRung("rung4:tcp", data.Rung4Global)

	// Registering bottom-up climbs Rung 1 -> 2 -> 3 -> 4 automatically.
	for _, tx := range []*stubRungTx{r1, r2, r3, r4} {
		clk.Advance(10)
		if err := lad.RegisterRung(tx); err != nil {
			t.Fatalf("[SEC 1] RegisterRung %s failed: %v", tx.Name(), err)
		}
		if lad.ActiveRung() != tx.Rung() {
			t.Fatalf("[SEC 1] expected climb to %s, got %s", tx.Rung(), lad.ActiveRung())
		}
	}

	// Simulate Rung 4 snapping mid-flight: SendFrame cascades automatically to Rung 3!
	r4.failSend.Store(true)
	used, err := lad.SendFrame(common.NodeBeta, []byte("SURVIVE-FAILOVER"))
	if err != nil || used != data.Rung3Local || r3.sentCnt.Load() != 1 {
		t.Fatalf("[SEC 1] mid-flight cascade 4->3 failed: used=%s err=%v r3=%d",
			used, err, r3.sentCnt.Load())
	}
	common.AssertAuditPass(t, lad.AuditState(), lad.ExecuteLadderDumpCard())
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: [ SECTION 2: NEVER FAIL (FAULT INVERSION & DEGRADATION) ]
// ------------------------------------------------------------------------------

func TestLadderUnit_Sec2_RejectsInvalidRungOversizeAndCollapsedFloor(t *testing.T) {
	lad := orchestrator.NewAdaptiveLadder(
		data.DefaultNodeManifest(common.NodeAlpha, "fault"), engine.NewMockClock(0),
	)
	defer func() { _ = lad.Close() }()

	if err := lad.RegisterRung(nil); err != orchestrator.ErrInvalidLadderRung {
		t.Fatalf("[SEC 2] nil tx: got %v", err)
	}
	badRung := newStubRung("bad", data.RungNone)
	if err := lad.RegisterRung(badRung); err != orchestrator.ErrInvalidLadderRung {
		t.Fatalf("[SEC 2] RungNone tx: got %v", err)
	}
	var over [data.FloorMaxFrameSize + 1]byte
	if _, err := lad.SendFrame(common.NodeBeta, over[:]); err != transport.ErrWireFrameOversize {
		t.Fatalf("[SEC 2] 513B SendFrame: got %v", err)
	}
	if _, err := lad.CascadeDown("empty"); err != orchestrator.ErrFloorCollapsed {
		t.Fatalf("[SEC 2] empty CascadeDown: got %v", err)
	}

	var nilLad *orchestrator.AdaptiveLadder
	common.AssertAuditFail(t, nilLad.AuditState(), data.AuditSeverityCritical)
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: [ SECTION 3: INVARIANTS, SATURATION & LIMIT GUARDS ]
// ------------------------------------------------------------------------------

func TestLadderUnit_Sec3_OfflineCeilingAnd64SlotHistoryRingWrap(t *testing.T) {
	clk := engine.NewMockClock(0)
	// Offline manifest (AllowGlobalEgress=false) MUST ignore Rung 4 even when registered!
	offline := data.DefaultNodeManifest(common.NodeAlpha, "offline-cap")
	lad := orchestrator.NewAdaptiveLadder(offline, clk)
	defer func() { _ = lad.Close() }()

	r2 := newStubRung("rung2:memory", data.Rung2Host)
	r4 := newStubRung("rung4:tcp", data.Rung4Global)
	_ = lad.RegisterRung(r2)
	_ = lad.RegisterRung(r4)
	if lad.ActiveRung() != data.Rung2Host {
		t.Fatalf("[SEC 3] offline ladder climbed onto forbidden Rung 4: %s", lad.ActiveRung())
	}

	// Trigger 80 transitions (> MaxCascadeHistory=64) to verify bounded ring wrap.
	for i := 0; i < 40; i++ {
		clk.Advance(10)
		r2.healthy.Store(false)
		lad.EvaluateAndAdapt("drop")
		clk.Advance(10)
		r2.healthy.Store(true)
		lad.EvaluateAndAdapt("heal")
	}

	var all [128]data.CascadeEvent
	n := lad.HistorySnapshotInto(all[:])
	if n != orchestrator.MaxCascadeHistory {
		t.Fatalf("[SEC 3] wrapped history count=%d want %d", n, orchestrator.MaxCascadeHistory)
	}
	// Chronological ordering invariant across ring wrap.
	for i := 1; i < n; i++ {
		if all[i].TimestampNano < all[i-1].TimestampNano {
			t.Fatalf("[SEC 3] ring history out of chronological order at %d", i)
		}
	}
	common.AssertAuditPass(t, lad.AuditState(), lad.ExecuteLadderDumpCard())
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.4: [ SECTION 4: EMERGENCE & COMPOSITE WORKFLOWS ]
// ------------------------------------------------------------------------------

func TestLadderUnit_Sec4_RegisteredSnapshotAndDumpCard(t *testing.T) {
	lad := orchestrator.NewAdaptiveLadder(
		data.DefaultNodeManifest(common.NodeAlpha, "snap"), engine.NewMockClock(10),
	)
	_ = lad.RegisterRung(newStubRung("r1", data.Rung1Floor))
	_ = lad.RegisterRung(newStubRung("r2", data.Rung2Host))

	var txs [4]transport.Transceiver
	if got := lad.RegisteredTransceivers(txs[:]); got != 2 {
		t.Fatalf("[SEC 4] expected 2 registered transceivers, got %d", got)
	}
	card := lad.ExecuteLadderDumpCard()
	if !strings.Contains(card, "RUNG_2_HOST") || !strings.Contains(card, "2 / 4") {
		t.Fatalf("[SEC 4] unexpected ladder card:\n%s", card)
	}
	_ = lad.Close()
	if lad.ActiveRung() != data.RungNone || lad.ActiveTransceiver() != nil {
		t.Fatal("[SEC 4] Close() left active rung behind")
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Preflight Diagnostic Gates & Physical Source Anchor Verification )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.orchestrator.ladder
// Mission:       Run PreflightLadder and verify physical source anchor.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Block 4 Pre-Flight Contract Gates
// ------------------------------------------------------------------------------

func TestLadderUnit_Block4_PreflightGates(t *testing.T) {
	if ladderUnitSuiteName == "" {
		t.Fatal("[BLOCK 4] suite identifier must not be empty")
	}
	common.AssertAuditPass(t, orchestrator.PreflightLadder(), "")
	if !orchestrator.AuditLadder() {
		t.Fatal("[BLOCK 4] AuditLadder returned false")
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Physical Source Anchor Verification
// ------------------------------------------------------------------------------

func TestLadderUnit_Block4_SourceAnchors(t *testing.T) {
	common.AssertRepoAnchors(t, "orchestrator/ladder.go")
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Test Suite Epilogue & Clean Readiness Handshake )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.orchestrator.ladder
// Mission:       Confirm ladder and harness readiness post-suite.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Epilogue Readiness Verification
// ------------------------------------------------------------------------------

func TestLadderUnit_Block5_EpilogueReadiness(t *testing.T) {
	if !orchestrator.IsLadderReady() || !common.IsHarnessReady() {
		t.Fatal("[BLOCK 5: EPILOGUE] ladder unit readiness handshake failed")
	}
}

// End of tests/unit/orchestrator/ladder_test.go
