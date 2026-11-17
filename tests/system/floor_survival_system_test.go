// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/system/floor_survival_system_test.go
// @id            test.system.floor_survival
// @layer         Layer 4: Tier-4 System End-to-End & Stress Verification Suite
// @archetype     verification_suite | system_floor_survival
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Civilizational Blackout & Rung 1 Survival Proving Ground
// @role          Tier-4 System Gate: Total IP/Wi-Fi Blackout Survival Over Floor
// @maturity      Rev 5 Baseline | Wave 6 4x4 Symmetry Expansion
// @chassis       bounded-heap | Morse CW + Noisy AFSK PCM + SLIP Serial Chain
// @scale         Sustained blackout stream & 3-node cross-medium floor relay
// @seams         in: data, engine, orchestrator, transport/rung1_floor, common
// @invariants    1. Macro Tier 4 (System): zero IP, zero UDP, zero TCP, zero
//                   Wi-Fi, zero DNS—100% Rung 1 Civilizational Floor survival.
//                2. Multi-hop relay bridges SLIP Serial -> AFSK Audio -> Morse
//                   CW end-to-end with 100% CRC-32C bit fidelity.
//                3. EmergencyFloorProfile caps ladder ceiling cleanly at Rung 1.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Tier-4 System proving ground simulating a total modern infrastructure
//   blackout where only bare Rung 1 Floor physics survive: RFC 1055 SLIP copper
//   serial cables, 1200/2400 Hz AFSK radio audio tones (with simulated static),
//   and 1844 International Morse CW keying pulses. Proves a 3-node neighborhood
//   chain can still route local news and sovereign pings with 0% corruption.
//
//   ### For Architects, Operators & Non-Programmers:
//   Answers Seanje's founding question—"what is the oldest tech this can
//   support when our block is the only block still getting local news?"—by
//   proving live 3-node communication across copper serial, noisy radio chirps,
//   and Morse code pulses with zero modern networking active.
//
//   ### For Systems Programmers & Engine Architects:
//   Exercises `EmergencyFloorProfile`, `SerialTransceiver`, `AudioTransceiver`,
//   `MorseTransceiver`, `AdaptiveLadder`, and `Dispatcher` in sustained streams
//   and cross-medium relays under `go test -race`.
// ==============================================================================

package system_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Identifier, Verification Imports & Blackout Survival Anchor )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-4 System Verification Suite
// Module:        test.system.floor_survival
// Mission:       Bind Rung 1 Floor mediums and Orchestrator for blackout stress.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"context"
	"fmt"
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
	floorSurvivalSuiteName = "Tier 4 System: Civilizational Blackout Floor Survival"
)

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-4 System Verification Suite
// Module:        test.system.floor_survival
// Mission:       Prove sustained blackout streams, corruption rejection, floor
//                ceiling invariants, and 3-node cross-medium blackout relay.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: [ SECTION 1: ALWAYS PASS (SUSTAINED SLIP BLACKOUT STREAM) ]
// ------------------------------------------------------------------------------

func TestFloorSurvivalSystem_Sec1_SustainedBlackoutStreamOverSLIPSerial(t *testing.T) {
	rB, wA := io.Pipe()
	rA, wB := io.Pipe()
	defer func() {
		_ = rA.Close()
		_ = rB.Close()
	}()

	dA, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(common.NodeAlpha, "blackout-alpha"), engine.NewMonotonicClock(),
	)
	dB, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(common.NodeBeta, "blackout-beta"), engine.NewMonotonicClock(),
	)
	defer func() {
		_ = dA.Close()
		_ = dB.Close()
	}()

	_ = dA.Ladder().RegisterRung(rung1_floor.NewSerialTransceiver(common.NodeAlpha, rA, wA, wA))
	_ = dB.Ladder().RegisterRung(rung1_floor.NewSerialTransceiver(common.NodeBeta, rB, wB, wB))

	const streamCount = 16
	received := make(chan string, streamCount)
	dB.SetDeliverHandler(func(f *data.Frame, _ covenant.Verdict) {
		if f.Opcode == data.OpData && f.Rung == data.Rung1Floor {
			received <- string(f.PayloadSlice())
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = dA.Start(ctx)
	_ = dB.Start(ctx)

	for i := 1; i <= streamCount; i++ {
		body := fmt.Sprintf("BLOCK-NEWS-%02d:Generator+Water+OK", i)
		f := common.MakeSovereignFrame(
			common.NodeAlpha, common.NodeBeta, data.Rung1Floor, data.OpData, uint32(i), []byte(body),
		)
		if rung, err := dA.SendFrame(&f); err != nil || rung != data.Rung1Floor {
			t.Fatalf("[SEC 1] SLIP stream frame %d failed: rung=%s err=%v", i, rung, err)
		}
	}

	for i := 1; i <= streamCount; i++ {
		want := fmt.Sprintf("BLOCK-NEWS-%02d:Generator+Water+OK", i)
		select {
		case got := <-received:
			if got != want {
				t.Fatalf("[SEC 1] stream frame %d mismatch: want %q got %q", i, want, got)
			}
		case <-ctx.Done():
			t.Fatalf("[SEC 1] timed out waiting for SLIP blackout bulletin %d/%d", i, streamCount)
		}
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: [ SECTION 2: NEVER FAIL (CORRUPTED FLOOR SIGNALS DROPPED) ]
// ------------------------------------------------------------------------------

func TestFloorSurvivalSystem_Sec2_CorruptedMorseAFSKAndSLIPSignalsRejected(t *testing.T) {
	// 1. Truncated/invalid Morse CW stream must return clean error, never panic.
	if _, err := rung1_floor.MorseDecodeBytes("........ -------- / ???"); err == nil {
		t.Fatal("[SEC 2] garbled Morse CW stream unexpectedly decoded")
	}

	// 2. Flat dead-air PCM buffer must fail AFSK demodulation cleanly.
	deadAir := make([]int16, 4096)
	if _, err := rung1_floor.DemodulateAFSK(deadAir); err == nil {
		t.Fatal("[SEC 2] dead-air PCM unexpectedly demodulated as AFSK")
	}

	// 3. Corrupted SLIP escape byte (0xDB followed by invalid 0x00) must fail.
	badSLIP := []byte{rung1_floor.SLIPEND, rung1_floor.SLIPESC, 0x00, rung1_floor.SLIPEND}
	if _, err := rung1_floor.SLIPDecode(badSLIP); err == nil {
		t.Fatal("[SEC 2] invalid SLIP escape sequence unexpectedly decoded")
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: [ SECTION 3: INVARIANTS (EMERGENCY BODY CEILING AT RUNG 1) ]
// ------------------------------------------------------------------------------

func TestFloorSurvivalSystem_Sec3_EmergencyBodyProfileCapsLadderAtRung1(t *testing.T) {
	profile := common.EmergencyFloorProfile()
	if profile.HighestViableRung() != data.Rung1Floor {
		t.Fatalf("[SEC 3] EmergencyFloorProfile ceiling mismatch: got %s", profile.HighestViableRung())
	}

	m := data.DefaultNodeManifest(common.NodeAlpha, "emergency-radio")
	m.PreferredMaxRung = profile.HighestViableRung()
	ladder := orchestrator.NewAdaptiveLadder(m, engine.NewMonotonicClock())
	defer func() { _ = ladder.Close() }()

	morseMed := rung1_floor.NewMorseMedium()
	audioMed := rung1_floor.NewAudioMedium()
	_ = ladder.RegisterRung(morseMed.Attach(common.NodeAlpha))
	_ = ladder.RegisterRung(audioMed.Attach(common.NodeAlpha))

	active := ladder.EvaluateAndAdapt("blackout_boot")
	if active != data.Rung1Floor {
		t.Fatalf("[SEC 3] expected EmergencyFloorProfile ladder to lock onto Rung1Floor, got %s", active)
	}
	common.AssertAuditPass(t, ladder.AuditState(), ladder.ExecuteLadderDumpCard())
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.4: [ SECTION 4: EMERGENCE (3-NODE SLIP -> AFSK -> MORSE RELAY) ]
// ------------------------------------------------------------------------------

func TestFloorSurvivalSystem_Sec4_ThreeNodeBlackoutRelayAcrossSLIPAFSKAndMorse(t *testing.T) {
	// NodeAlpha ---(SLIP Copper)---> NodeRelay ---(Noisy AFSK + Morse CW)---> NodeBeta
	rRelay, wAlpha := io.Pipe()
	rAlpha, wRelay := io.Pipe()
	defer func() {
		_ = rAlpha.Close()
		_ = rRelay.Close()
	}()

	slipAlpha := rung1_floor.NewSerialTransceiver(common.NodeAlpha, rAlpha, wAlpha, wAlpha)
	slipRelay := rung1_floor.NewSerialTransceiver(common.NodeRelay, rRelay, wRelay, wRelay)

	afskMed := rung1_floor.NewAudioMedium()
	afskRelay := afskMed.Attach(common.NodeRelay)
	afskBeta := afskMed.Attach(common.NodeBeta)

	morseMed := rung1_floor.NewMorseMedium()
	morseRelay := morseMed.Attach(common.NodeRelay)
	morseBeta := morseMed.Attach(common.NodeBeta)

	dAlpha, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(common.NodeAlpha, "house-1"), engine.NewMonotonicClock(),
	)
	dRelay, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(common.NodeRelay, "block-relay"), engine.NewMonotonicClock(),
	)
	dBeta, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(common.NodeBeta, "house-3"), engine.NewMonotonicClock(),
	)
	defer func() {
		_ = dAlpha.Close()
		_ = dRelay.Close()
		_ = dBeta.Close()
	}()

	_ = dAlpha.Ladder().RegisterRung(slipAlpha)
	_ = dRelay.Ladder().RegisterRung(slipRelay)
	_ = dRelay.Ladder().RegisterRung(morseRelay)
	_ = dRelay.Ladder().RegisterRung(afskRelay)
	_ = dBeta.Ladder().RegisterRung(morseBeta)
	_ = dBeta.Ladder().RegisterRung(afskBeta)

	// Relay forwards any OpData addressed to NodeBeta over AFSK (or Morse fallback).
	dRelay.SetDeliverHandler(func(f *data.Frame, _ covenant.Verdict) {
		if f.Opcode == data.OpData && f.SourceID == common.NodeAlpha {
			fwd := common.MakeSovereignFrame(
				common.NodeRelay, common.NodeBeta, data.Rung1Floor, data.OpData, f.SeqID+100, f.PayloadSlice(),
			)
			// Send directly over AFSK first, then over Morse CW to prove both hops!
			wire, err := encodeForRelay(&fwd)
			if err == nil {
				_ = afskRelay.Send(common.NodeBeta, wire)
				_ = morseRelay.Send(common.NodeBeta, wire)
			}
		}
	})

	arrivals := make(chan string, 4)
	dBeta.SetDeliverHandler(func(f *data.Frame, _ covenant.Verdict) {
		if f.Opcode == data.OpData {
			arrivals <- string(f.PayloadSlice())
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	_ = dAlpha.Start(ctx)
	_ = dRelay.Start(ctx)
	_ = dBeta.Start(ctx)

	bulletin := common.MakeSovereignFrame(
		common.NodeAlpha, common.NodeRelay, data.Rung1Floor, data.OpData, 1,
		[]byte("BLACKOUT-BULLETIN:SHELTER-OPEN"),
	)
	if _, err := dAlpha.SendFrame(&bulletin); err != nil {
		t.Fatalf("[SEC 4] Alpha SLIP SendFrame failed: %v", err)
	}

	// Expect BOTH AFSK and Morse CW copies to arrive cleanly at NodeBeta!
	for i := 0; i < 2; i++ {
		select {
		case msg := <-arrivals:
			if msg != "BLACKOUT-BULLETIN:SHELTER-OPEN" {
				t.Fatalf("[SEC 4] relayed blackout bulletin corrupted: %q", msg)
			}
		case <-ctx.Done():
			t.Fatalf("[SEC 4] timed out waiting for floor relay arrival %d/2", i+1)
		}
	}
}

func encodeForRelay(f *data.Frame) ([]byte, error) {
	var buf [data.FloorMaxFrameSize]byte
	n, err := engine.EncodeInto(buf[:], f)
	if err != nil {
		return nil, err
	}
	out := make([]byte, n)
	copy(out, buf[:n])
	return out, nil
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Preflight Diagnostic Gates & Physical Source Anchor Verification )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-4 System Verification Suite
// Module:        test.system.floor_survival
// Mission:       Assert all Rung 1 Floor preflights and verify source anchors.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Block 4 Pre-Flight Contract Gates
// ------------------------------------------------------------------------------

func TestFloorSurvivalSystem_Block4_PreflightGates(t *testing.T) {
	if floorSurvivalSuiteName == "" {
		t.Fatal("[BLOCK 4] suite identifier must not be empty")
	}
	common.AssertAuditPass(t, rung1_floor.PreflightSerialFloor(), "")
	common.AssertAuditPass(t, rung1_floor.PreflightAudioModem(), "")
	common.AssertAuditPass(t, rung1_floor.PreflightMorseFloor(), "")
	common.AssertAuditPass(t, orchestrator.PreflightProbe(), "")
	common.AssertAuditPass(t, orchestrator.PreflightDispatcher(), "")
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Physical Source Anchor Verification
// ------------------------------------------------------------------------------

func TestFloorSurvivalSystem_Block4_SourceAnchors(t *testing.T) {
	common.AssertRepoAnchors(t,
		"transport/rung1_floor/serial.go",
		"transport/rung1_floor/audio.go",
		"transport/rung1_floor/morse.go",
		"orchestrator/ladder.go",
	)
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Test Suite Epilogue & Clean Readiness Handshake )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-4 System Verification Suite
// Module:        test.system.floor_survival
// Mission:       Confirm Rung 1 Floor system readiness handshake post-suite.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Epilogue Readiness Verification
// ------------------------------------------------------------------------------

func TestFloorSurvivalSystem_Block5_EpilogueReadiness(t *testing.T) {
	if !rung1_floor.IsSerialFloorReady() || !rung1_floor.IsAudioModemReady() ||
		!rung1_floor.IsMorseFloorReady() || !common.IsHarnessReady() {
		t.Fatal("[BLOCK 5: EPILOGUE] Rung 1 Floor survival readiness failed")
	}
}

// End of tests/system/floor_survival_system_test.go
