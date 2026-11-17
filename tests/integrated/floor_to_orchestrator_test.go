// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/integrated/floor_to_orchestrator_test.go
// @id            test.integrated.floor_to_orchestrator
// @layer         Layer 4: Tier-3 Integrated Verification Suite
// @archetype     verification_suite | integrated_floor_orchestrator
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Rung 1 Floor (Morse/AFSK/SLIP) + Orchestrator Suite
// @role          Tier-3 Integrated Sensor: Rung 1 Floor Cogs + Live Dispatcher
// @maturity      Rev 5 Baseline | Wave 6 4x4 Symmetry Expansion
// @chassis       bounded-heap | Morse CW, AFSK 1200/2400Hz & SLIP Serial + Airlock
// @scale         O(1) live Rung 1 frame exchange & floor-to-floor failover
// @seams         in: data, engine, orchestrator, transport/rung1_floor, common
// @invariants    1. Macro Tier 3 (Integrated): wires all 3 Rung 1 survival floor
//                   mediums into live Orchestrator Dispatchers with Airlock.
//                2. Zero IP/TCP/UDP/Wi-Fi dependency across the entire suite.
//                3. Live hostile frames injected over Rung 1 are dropped cleanly.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Tier-3 Integrated verification suite proving all three Rung 1 Civilizational
//   Floor mediums (`MorseTransceiver` 1844 CW pulses, `AudioTransceiver`
//   1200/2400 Hz AFSK PCM modem, and `SerialTransceiver` RFC 1055 SLIP stream)
//   operate seamlessly under `orchestrator.Dispatcher`, `AdaptiveLadder`, and
//   the Genesis 3 Airlock—including sibling floor failover and hostile blocking.
//
//   ### For Architects, Operators & Non-Programmers:
//   Proves that even with zero internet, zero Wi-Fi, and zero Ethernet,
//   FaithNet's brain can still exchange pings, deliver local bulletins, block
//   tampered messages, and fail over between serial cable, radio audio, and
//   Morse code pulses automatically.
//
//   ### For Systems Programmers & Engine Architects:
//   Exercises `MorseMedium`, `AudioMedium`, `SerialTransceiver` (`io.Pipe`),
//   `Dispatcher.PingPeer`, `Dispatcher.SetDeliverHandler`, and live Airlock
//   screening concurrently under `-race`.
// ==============================================================================

package integrated_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Identifier, Verification Imports & Tier-3 Floor Integration Anchor )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-3 Integrated Verification Suite
// Module:        test.integrated.floor_to_orchestrator
// Mission:       Bind Rung 1 Floor mediums into live Orchestrator Dispatchers.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"context"
	"io"
	"testing"
	"time"
)

// --- Tier 2: External Library Dependencies (None — 100% Go Stdlib Invariant) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---
import (
	"github.com/CreativeWorkzStudio/FaithNet/data"
	"github.com/CreativeWorkzStudio/FaithNet/engine"
	"github.com/CreativeWorkzStudio/FaithNet/engine/covenant"
	"github.com/CreativeWorkzStudio/FaithNet/orchestrator"
	"github.com/CreativeWorkzStudio/FaithNet/tests/common"
	"github.com/CreativeWorkzStudio/FaithNet/transport/rung1_floor"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Operational Suite Constants & Identifiers
// ------------------------------------------------------------------------------

const (
	floorOrchSuiteName = "Tier 3 Integrated: Rung 1 Floor (Morse/AFSK/SLIP) + Orchestrator"
)

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-3 Integrated Verification Suite
// Module:        test.integrated.floor_to_orchestrator
// Mission:       Verify live SLIP/AFSK/Morse delivery, Airlock enforcement over
//                Rung 1, noisy AFSK demodulation, and floor sibling failover.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: [ SECTION 1: ALWAYS PASS (LIVE SLIP SERIAL DATA & PING) ]
// ------------------------------------------------------------------------------

func TestFloorToOrchestrator_Sec1_SLIPSerialLivePingAndBulletinDelivery(t *testing.T) {
	rB, wA := io.Pipe()
	rA, wB := io.Pipe()
	defer func() {
		_ = rA.Close()
		_ = rB.Close()
	}()

	sA := rung1_floor.NewSerialTransceiver(common.NodeAlpha, rA, wA, wA)
	sB := rung1_floor.NewSerialTransceiver(common.NodeBeta, rB, wB, wB)

	dA, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(common.NodeAlpha, "slip-a"), engine.NewMonotonicClock(),
	)
	dB, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(common.NodeBeta, "slip-b"), engine.NewMonotonicClock(),
	)
	defer func() {
		_ = dA.Close()
		_ = dB.Close()
	}()

	delivered := make(chan string, 1)
	dB.SetDeliverHandler(func(f *data.Frame, _ covenant.Verdict) {
		if f.Opcode == data.OpData {
			delivered <- string(f.PayloadSlice())
		}
	})

	_ = dA.Ladder().RegisterRung(sA)
	_ = dB.Ladder().RegisterRung(sB)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = dA.Start(ctx)
	_ = dB.Start(ctx)

	if _, rung, err := dA.PingPeer(ctx, common.NodeBeta, []byte("SLIP-CHECK")); err != nil || rung != data.Rung1Floor {
		t.Fatalf("[SEC 1] SLIP live ping failed: rung=%s err=%v", rung, err)
	}

	msg := common.MakeSovereignFrame(
		common.NodeAlpha, common.NodeBeta, data.Rung1Floor, data.OpData, 9, []byte("BLOCK-4-WATER-OK"),
	)
	if _, err := dA.SendFrame(&msg); err != nil {
		t.Fatalf("[SEC 1] SLIP SendFrame OpData failed: %v", err)
	}

	select {
	case got := <-delivered:
		if got != "BLOCK-4-WATER-OK" {
			t.Fatalf("[SEC 1] expected BLOCK-4-WATER-OK over SLIP, got %q", got)
		}
	case <-ctx.Done():
		t.Fatal("[SEC 1] timed out waiting for SLIP OpData delivery")
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: [ SECTION 2: NEVER FAIL (AIRLOCK BLOCKS HOSTILE OVER AFSK) ]
// ------------------------------------------------------------------------------

func TestFloorToOrchestrator_Sec2_AirlockBlocksHostileFrameOverAFSKAudio(t *testing.T) {
	med := rung1_floor.NewAudioMedium()
	txA := med.Attach(common.NodeAlpha)
	txB := med.Attach(common.NodeBeta)

	dB, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(common.NodeBeta, "afsk-guard"), engine.NewMonotonicClock(),
	)
	defer func() {
		_ = txA.Close()
		_ = dB.Close()
	}()

	delivered := make(chan string, 1)
	dB.SetDeliverHandler(func(f *data.Frame, _ covenant.Verdict) {
		delivered <- string(f.PayloadSlice())
	})
	_ = dB.Ladder().RegisterRung(txB)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = dB.Start(ctx)

	// Transmit an external-origin hostile script over 1200/2400 Hz AFSK PCM audio!
	hostile := data.Frame{
		Version:  data.ProtocolVersion,
		Rung:     data.Rung1Floor,
		Opcode:   data.OpLegacyWrap,
		Flags:    data.FlagExternalOrigin,
		Truth:    data.QuadAwait,
		SourceID: common.NodeExternal,
		TargetID: common.NodeBeta,
		SeqID:    44,
	}
	hostile.SetPayload([]byte("<script>alert('radio')</script>"))
	wire := common.EncodeFixture(t, &hostile)
	if err := txA.Send(common.NodeBeta, wire); err != nil {
		t.Fatalf("[SEC 2] AFSK Send failed: %v", err)
	}

	// Wait briefly for AFSK demodulation + Airlock drop, verify blocked counter increments.
	deadline := time.Now().Add(1 * time.Second)
	for time.Now().Before(deadline) {
		if dB.StatsSnapshot().AirlockBlocked >= 1 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if dB.StatsSnapshot().AirlockBlocked < 1 {
		t.Fatalf("[SEC 2] expected Dispatcher Airlock to block hostile AFSK frame, stats=%+v",
			dB.StatsSnapshot())
	}
	select {
	case leaked := <-delivered:
		t.Fatalf("[SEC 2] hostile AFSK frame leaked to deliver handler: %q", leaked)
	default:
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: [ SECTION 3: INVARIANTS (NOISY AFSK SIGNAL INTEGRITY) ]
// ------------------------------------------------------------------------------

func TestFloorToOrchestrator_Sec3_NoisyAFSKWaveformSurvivesCRCAndAirlock(t *testing.T) {
	frame := common.MakeSovereignFrame(
		common.NodeAlpha, common.NodeBeta, data.Rung1Floor, data.OpData, 77, []byte("NOISY-FM-OK"),
	)
	wire := common.EncodeFixture(t, &frame)
	pcm, err := rung1_floor.ModulateAFSK(wire)
	if err != nil {
		t.Fatalf("[SEC 3] ModulateAFSK failed: %v", err)
	}

	// Inject deterministic simulated radio static (+/- 1,500 amplitude on 16-bit PCM).
	for i := range pcm {
		static := int32(((i*37)%11)-5) * 300
		v := int32(pcm[i]) + static
		if v > 32767 {
			v = 32767
		} else if v < -32768 {
			v = -32768
		}
		pcm[i] = int16(v)
	}

	recovered, err := rung1_floor.DemodulateAFSK(pcm)
	if err != nil {
		t.Fatalf("[SEC 3] DemodulateAFSK failed under simulated static: %v", err)
	}
	var decoded data.Frame
	if err := engine.DecodeFrom(recovered, &decoded); err != nil {
		t.Fatalf("[SEC 3] Codec CRC failed after noisy AFSK demodulation: %v", err)
	}
	common.AssertFramesEqual(t, &frame, &decoded)
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.4: [ SECTION 4: EMERGENCE (AFSK AUDIO -> MORSE CW FLOOR DROP) ]
// ------------------------------------------------------------------------------

func TestFloorToOrchestrator_Sec4_AFSKDropCascadesTo1844MorseCWFloor(t *testing.T) {
	morseMed := rung1_floor.NewMorseMedium()
	morseA := morseMed.Attach(common.NodeAlpha)
	morseB := morseMed.Attach(common.NodeBeta)

	audioMed := rung1_floor.NewAudioMedium()
	audioA := audioMed.Attach(common.NodeAlpha)
	audioB := audioMed.Attach(common.NodeBeta)

	dA, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(common.NodeAlpha, "floor-a"), engine.NewMonotonicClock(),
	)
	dB, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(common.NodeBeta, "floor-b"), engine.NewMonotonicClock(),
	)
	defer func() {
		_ = dA.Close()
		_ = dB.Close()
	}()

	// Register Morse CW first, AFSK Audio second (both Rung1Floor).
	_ = dA.Ladder().RegisterRung(morseA)
	_ = dA.Ladder().RegisterRung(audioA)
	_ = dB.Ladder().RegisterRung(morseB)
	_ = dB.Ladder().RegisterRung(audioB)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = dA.Start(ctx)
	_ = dB.Start(ctx)

	// 1. Ping while AFSK Audio is healthy.
	if _, rung, err := dA.PingPeer(ctx, common.NodeBeta, []byte("TONE-1")); err != nil || rung != data.Rung1Floor {
		t.Fatalf("[SEC 4] initial Rung 1 ping failed: rung=%s err=%v", rung, err)
	}

	// 2. Cut AFSK Audio -> Ladder automatically survives on 1844 Morse CW pulses!
	audioA.SetHealthy(false)
	if _, rung, err := dA.PingPeer(ctx, common.NodeBeta, []byte("MORSE-2")); err != nil || rung != data.Rung1Floor {
		t.Fatalf("[SEC 4] Morse CW rescue ping failed: rung=%s err=%v", rung, err)
	}
	if dA.Ping().HealthSnapshot().PongsRecv != 2 {
		t.Fatalf("[SEC 4] expected 2/2 floor pings across AFSK->Morse, got %+v", dA.Ping().HealthSnapshot())
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Preflight Diagnostic Gates & Physical Source Anchor Verification )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-3 Integrated Verification Suite
// Module:        test.integrated.floor_to_orchestrator
// Mission:       Assert Rung 1 Floor & Orchestrator preflights and source anchors.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Block 4 Pre-Flight Contract Gates
// ------------------------------------------------------------------------------

func TestFloorToOrchestrator_Block4_PreflightGates(t *testing.T) {
	if floorOrchSuiteName == "" {
		t.Fatal("[BLOCK 4] suite identifier must not be empty")
	}
	common.AssertAuditPass(t, rung1_floor.PreflightSerialFloor(), "")
	common.AssertAuditPass(t, rung1_floor.PreflightAudioModem(), "")
	common.AssertAuditPass(t, rung1_floor.PreflightMorseFloor(), "")
	common.AssertAuditPass(t, orchestrator.PreflightDispatcher(), "")
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Physical Source Anchor Verification
// ------------------------------------------------------------------------------

func TestFloorToOrchestrator_Block4_SourceAnchors(t *testing.T) {
	common.AssertRepoAnchors(t,
		"transport/rung1_floor/serial.go",
		"transport/rung1_floor/audio.go",
		"transport/rung1_floor/morse.go",
		"orchestrator/dispatcher.go",
	)
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Test Suite Epilogue & Clean Readiness Handshake )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-3 Integrated Verification Suite
// Module:        test.integrated.floor_to_orchestrator
// Mission:       Confirm Rung 1 Floor & Orchestrator readiness post-suite.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Epilogue Readiness Verification
// ------------------------------------------------------------------------------

func TestFloorToOrchestrator_Block5_EpilogueReadiness(t *testing.T) {
	if !rung1_floor.IsSerialFloorReady() || !rung1_floor.IsAudioModemReady() ||
		!rung1_floor.IsMorseFloorReady() || !orchestrator.IsDispatcherReady() ||
		!common.IsHarnessReady() {
		t.Fatal("[BLOCK 5: EPILOGUE] Rung 1 Floor + Orchestrator readiness failed")
	}
}

// End of tests/integrated/floor_to_orchestrator_test.go
