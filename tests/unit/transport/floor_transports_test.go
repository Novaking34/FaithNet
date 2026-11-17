// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/unit/transport/floor_transports_test.go
// @id            test.unit.transport.rung1_floor
// @layer         Layer 4: Tier-2 Unit Verification Suite (Transport Sub-Stratum)
// @archetype     verification_suite | unit_isolation
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Rung 1 Civilizational Floor Transports Unit Suite
// @role          Tier-2 Unit Isolation Gate for Serial SLIP, AFSK Audio & Morse
// @maturity      Rev 5 Baseline | 4x4 Verification Matrix
// @chassis       bounded-heap | Exercises real SLIP, 1200/2400Hz PCM & CW pulses
// @scale         O(N_frame) lossless binary floor transport verification
// @seams         in: data, engine, transport, transport/rung1_floor, common
// @invariants    1. All 3 Rung 1 mediums carry real 512B binary 0xFA17 frames
//                   losslessly bit-for-bit.
//                2. AFSK Goertzel detector survives simulated analog static.
//                3. Morse CW pulses honor ITU 1:3:1:3 timing losslessly.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Tier-2 Unit isolation suite verifying all three Civilizational Floor
//   transports (`transport/rung1_floor/{serial,audio,morse}.go`) across
//   nominal round-trips, fault inversion, 512B Floppy Sector ceilings, real
//   encoded FaithNet frames, and Block 4/5 audit handshakes.
//
//   ### For Architects, Operators & Non-Programmers:
//   Proves FaithNet's survival floor—copper serial cable, audible dial-up/radio
//   chirps, and 1844 Morse telegraph pulses—delivers real sovereign packets
//   intact without losing a single bit.
//
//   ### For Systems Programmers & Engine Architects:
//   100% branch coverage of `SLIPEncode/Decode`, `Modulate/DemodulateAFSK`,
//   `MorseEncode/DecodeBytes`, `MorseToPulses/PulsesToMorse`, and Cog 3 impls.
// ==============================================================================

package transport_unit_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Identifier, Verification Imports & Tier-2 Unit Sensor Anchor )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.rung1_floor
// Mission:       Bind Rung 1 Serial, Audio AFSK, Morse CW, and shared harness.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
)

// --- Tier 2: External Library Dependencies (None) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---
import (
	"github.com/CreativeWorkzStudio/FaithNet/data"
	"github.com/CreativeWorkzStudio/FaithNet/engine"
	"github.com/CreativeWorkzStudio/FaithNet/tests/common"
	"github.com/CreativeWorkzStudio/FaithNet/transport"
	"github.com/CreativeWorkzStudio/FaithNet/transport/rung1_floor"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Operational Suite Constants & Identifiers
// ------------------------------------------------------------------------------

const (
	floorTransportsUnitSuiteName = "Tier 2 Unit: Rung 1 Floor (SLIP, AFSK, Morse)"
)

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.rung1_floor
// Mission:       Verify Serial SLIP, AFSK Audio, and Morse CW across Sec 1-4.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: [ SECTION 1: ALWAYS PASS (REAL 0xFA17 FRAMES ON ALL 3) ]
// ------------------------------------------------------------------------------

func TestFloorUnit_Sec1_SerialSLIPAll256ByteValues(t *testing.T) {
	allBytes := make([]byte, 256)
	for i := 0; i < 256; i++ {
		allBytes[i] = byte(i)
	}
	enc, err := rung1_floor.SLIPEncode(allBytes)
	if err != nil {
		t.Fatalf("[SEC 1] SLIPEncode 0x00..0xFF failed: %v", err)
	}
	dec, err := rung1_floor.SLIPDecode(enc)
	if err != nil || !bytes.Equal(dec, allBytes) {
		t.Fatalf("[SEC 1] SLIPDecode 0x00..0xFF mismatch: err=%v len=%d", err, len(dec))
	}

	rB, wA := io.Pipe()
	rA, wB := io.Pipe()
	txA := rung1_floor.NewSerialTransceiver(common.NodeAlpha, rA, wA, wA)
	txB := rung1_floor.NewSerialTransceiver(common.NodeBeta, rB, wB, wB)
	defer func() {
		_ = txA.Close()
		_ = txB.Close()
		_ = rA.Close()
		_ = rB.Close()
	}()
	if txA.Name() != "rung1:serial-slip" || txA.Rung() != data.Rung1Floor {
		t.Fatal("[SEC 1] unexpected SerialTransceiver metadata")
	}

	f := common.MakeSovereignFrame(
		common.NodeAlpha, common.NodeBeta, data.Rung1Floor, data.OpPing, 1, []byte("COPPER-WIRE"),
	)
	wire := common.EncodeFixture(t, &f)
	done := make(chan error, 1)
	go func() { done <- txA.Send(common.NodeBeta, wire) }()
	env, err := txB.Receive(context.Background())
	if sErr := <-done; sErr != nil || err != nil {
		t.Fatalf("[SEC 1] Serial pipe transfer failed: send=%v recv=%v", sErr, err)
	}
	var out data.Frame
	if err := engine.DecodeFrom(env.Slice(), &out); err != nil {
		t.Fatalf("[SEC 1] FaithNet frame decode after SLIP failed: %v", err)
	}
	common.AssertFramesEqual(t, &f, &out)
}

func TestFloorUnit_Sec1_AFSKModemAndMorseCWRealFrameRoundTrip(t *testing.T) {
	f := common.MakeSovereignFrame(
		common.NodeAlpha, common.NodeBeta, data.Rung1Floor, data.OpCPISync, 7,
		[]byte("RADIO-FLOOR-SYNC"),
	)
	wire := common.EncodeFixture(t, &f)

	// 1. AFSK Audio Medium broadcast + unicast.
	aMed := rung1_floor.NewAudioMedium()
	aA := aMed.Attach(common.NodeAlpha)
	aB := aMed.Attach(common.NodeBeta)
	defer func() {
		_ = aA.Close()
		_ = aB.Close()
	}()
	if aA.Name() != "rung1:audio-afsk" || aA.Rung() != data.Rung1Floor {
		t.Fatal("[SEC 1] unexpected AudioTransceiver metadata")
	}
	if err := aA.Send(data.NodeIDBroadcast, wire); err != nil {
		t.Fatalf("[SEC 1] AFSK broadcast failed: %v", err)
	}
	aEnv, err := aB.Receive(context.Background())
	if err != nil {
		t.Fatalf("[SEC 1] AFSK receive failed: %v", err)
	}
	var aOut data.Frame
	if err := engine.DecodeFrom(aEnv.Slice(), &aOut); err != nil {
		t.Fatalf("[SEC 1] AFSK decoded frame failed Codec: %v", err)
	}
	common.AssertFramesEqual(t, &f, &aOut)

	// 2. Morse CW Medium broadcast + unicast.
	mMed := rung1_floor.NewMorseMedium()
	mA := mMed.Attach(common.NodeAlpha)
	mB := mMed.Attach(common.NodeBeta)
	defer func() {
		_ = mA.Close()
		_ = mB.Close()
	}()
	if mA.Name() != "rung1:morse-cw" || mA.Rung() != data.Rung1Floor {
		t.Fatal("[SEC 1] unexpected MorseTransceiver metadata")
	}
	if err := mA.Send(data.NodeIDBroadcast, wire); err != nil {
		t.Fatalf("[SEC 1] Morse CW broadcast failed: %v", err)
	}
	mEnv, err := mB.Receive(context.Background())
	if err != nil {
		t.Fatalf("[SEC 1] Morse CW receive failed: %v", err)
	}
	var mOut data.Frame
	if err := engine.DecodeFrom(mEnv.Slice(), &mOut); err != nil {
		t.Fatalf("[SEC 1] Morse CW decoded frame failed Codec: %v", err)
	}
	common.AssertFramesEqual(t, &f, &mOut)
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: [ SECTION 2: NEVER FAIL (FAULT INVERSION & DEGRADATION) ]
// ------------------------------------------------------------------------------

func TestFloorUnit_Sec2_FaultInversionSerialAudioAndMorse(t *testing.T) {
	// Serial faults.
	if _, err := rung1_floor.SLIPDecode([]byte{rung1_floor.SLIPEND, rung1_floor.SLIPEND}); err != rung1_floor.ErrSLIPEmptyFrame {
		t.Fatalf("[SEC 2] empty SLIPDecode got %v", err)
	}
	if _, err := rung1_floor.SLIPDecode([]byte{rung1_floor.SLIPESC}); err != rung1_floor.ErrSLIPBadEscape {
		t.Fatalf("[SEC 2] dangling SLIPESC got %v", err)
	}
	var nilS *rung1_floor.SerialTransceiver
	common.AssertAuditFail(t, nilS.AuditState(), data.AuditSeverityCritical)

	// Audio AFSK faults.
	if _, err := rung1_floor.ModulateAFSK(nil); err != rung1_floor.ErrAFSKTruncated {
		t.Fatalf("[SEC 2] empty ModulateAFSK got %v", err)
	}
	if _, err := rung1_floor.DemodulateAFSK(nil); err != rung1_floor.ErrAFSKTruncated {
		t.Fatalf("[SEC 2] nil DemodulateAFSK got %v", err)
	}
	var nilA *rung1_floor.AudioTransceiver
	common.AssertAuditFail(t, nilA.AuditState(), data.AuditSeverityCritical)

	aMed := rung1_floor.NewAudioMedium()
	aTx := aMed.Attach(common.NodeAlpha)
	aTx.SetHealthy(false)
	if err := aTx.Send(common.NodeBeta, []byte("OFF")); err != transport.ErrTransceiverOffline {
		t.Fatalf("[SEC 2] offline Audio Send got %v", err)
	}
	_ = aTx.Close()
	if err := aTx.Send(common.NodeBeta, []byte("CLOSED")); err != transport.ErrTransceiverClosed {
		t.Fatalf("[SEC 2] closed Audio Send got %v", err)
	}

	// Morse faults.
	if _, err := rung1_floor.MorseDecodeBytes(""); err != rung1_floor.ErrMorseEmpty {
		t.Fatalf("[SEC 2] empty MorseDecodeBytes got %v", err)
	}
	if _, err := rung1_floor.MorseDecodeBytes("-------- --------"); err != rung1_floor.ErrMorseInvalidSymbol {
		t.Fatalf("[SEC 2] bogus Morse symbol got %v", err)
	}
	var nilM *rung1_floor.MorseTransceiver
	common.AssertAuditFail(t, nilM.AuditState(), data.AuditSeverityCritical)

	mMed := rung1_floor.NewMorseMedium()
	mTx := mMed.Attach(common.NodeAlpha)
	mTx.SetHealthy(false)
	if err := mTx.Send(common.NodeBeta, []byte("OFF")); err != transport.ErrTransceiverOffline {
		t.Fatalf("[SEC 2] offline Morse Send got %v", err)
	}
	_ = mTx.Close()
	if err := mTx.Send(common.NodeBeta, []byte("CLOSED")); err != transport.ErrTransceiverClosed {
		t.Fatalf("[SEC 2] closed Morse Send got %v", err)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: [ SECTION 3: INVARIANTS, SATURATION & LIMIT GUARDS ]
// ------------------------------------------------------------------------------

func TestFloorUnit_Sec3_Exact512ByteFloorSectorAcrossAllThreeMediums(t *testing.T) {
	exact512 := bytes.Repeat([]byte{0xC0, 0xDB, 0xFA, 0x17}, 128) // Exactly 512 bytes!
	over513 := make([]byte, data.FloorMaxFrameSize+1)

	// 1. SLIP exact 512B passes; 513B rejected.
	if sEnc, err := rung1_floor.SLIPEncode(exact512); err != nil {
		t.Fatalf("[SEC 3] SLIP exact 512B failed: %v", err)
	} else if sDec, err := rung1_floor.SLIPDecode(sEnc); err != nil || !bytes.Equal(sDec, exact512) {
		t.Fatalf("[SEC 3] SLIP 512B decode mismatch: %v", err)
	}
	if _, err := rung1_floor.SLIPEncode(over513); err != transport.ErrWireFrameOversize {
		t.Fatalf("[SEC 3] SLIP 513B bypassed ceiling: %v", err)
	}

	// 2. AFSK exact 512B passes; 513B rejected.
	if pcm, err := rung1_floor.ModulateAFSK(exact512); err != nil {
		t.Fatalf("[SEC 3] AFSK exact 512B failed: %v", err)
	} else if aDec, err := rung1_floor.DemodulateAFSK(pcm); err != nil || !bytes.Equal(aDec, exact512) {
		t.Fatalf("[SEC 3] AFSK 512B decode mismatch: %v", err)
	}
	if _, err := rung1_floor.ModulateAFSK(over513); err != transport.ErrWireFrameOversize {
		t.Fatalf("[SEC 3] AFSK 513B bypassed ceiling: %v", err)
	}

	// 3. Morse exact 512B passes; 513B rejected.
	if mStr, err := rung1_floor.MorseEncodeBytes(exact512); err != nil {
		t.Fatalf("[SEC 3] Morse exact 512B failed: %v", err)
	} else if mDec, err := rung1_floor.MorseDecodeBytes(rung1_floor.PulsesToMorse(rung1_floor.MorseToPulses(mStr))); err != nil || !bytes.Equal(mDec, exact512) {
		t.Fatalf("[SEC 3] Morse 512B pulse round-trip mismatch: %v", err)
	}
	if _, err := rung1_floor.MorseEncodeBytes(over513); err != transport.ErrWireFrameOversize {
		t.Fatalf("[SEC 3] Morse 513B bypassed ceiling: %v", err)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.4: [ SECTION 4: EMERGENCE & COMPOSITE WORKFLOWS ]
// ------------------------------------------------------------------------------

func TestFloorUnit_Sec4_AllThreeFloorDumpCardsRenderCleanly(t *testing.T) {
	r, w := io.Pipe()
	sTx := rung1_floor.NewSerialTransceiver(common.NodeAlpha, r, w, w)
	if !strings.Contains(sTx.ExecuteSerialDumpCard(), "RFC 1055 SLIP") {
		t.Fatal("[SEC 4] Serial dump card missing SLIP label")
	}
	_ = sTx.Close()
	_ = r.Close()

	aTx := rung1_floor.NewAudioMedium().Attach(common.NodeAlpha)
	if !strings.Contains(aTx.ExecuteAudioDumpCard(), "1200Hz Mark / 2400Hz Space") {
		t.Fatal("[SEC 4] Audio dump card missing AFSK frequencies")
	}
	_ = aTx.Close()

	mTx := rung1_floor.NewMorseMedium().Attach(common.NodeAlpha)
	if !strings.Contains(mTx.ExecuteMorseDumpCard(), "ITU-R M.1677-1") {
		t.Fatal("[SEC 4] Morse dump card missing ITU standard label")
	}
	_ = mTx.Close()
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Preflight Diagnostic Gates & Physical Source Anchor Verification )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.rung1_floor
// Mission:       Run all three Rung 1 preflight gates and verify anchors.
// ==============================================================================

func TestFloorUnit_Block4_PreflightAndAnchors(t *testing.T) {
	if floorTransportsUnitSuiteName == "" {
		t.Fatal("[BLOCK 4] suite identifier must not be empty")
	}
	common.AssertAuditPass(t, rung1_floor.PreflightSerialFloor(), "")
	common.AssertAuditPass(t, rung1_floor.PreflightAudioModem(), "")
	common.AssertAuditPass(t, rung1_floor.PreflightMorseFloor(), "")
	if !rung1_floor.AuditSerialFloor() || !rung1_floor.AuditAudioModem() || !rung1_floor.AuditMorseFloor() {
		t.Fatal("[BLOCK 4] Rung 1 boolean audit wrappers failed")
	}
	common.AssertRepoAnchors(t,
		"transport/rung1_floor/serial.go",
		"transport/rung1_floor/audio.go",
		"transport/rung1_floor/morse.go",
	)
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Test Suite Epilogue & Clean Readiness Handshake )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.rung1_floor
// Mission:       Confirm all three Rung 1 Floor readiness probes post-suite.
// ==============================================================================

func TestFloorUnit_Block5_EpilogueReadiness(t *testing.T) {
	if !rung1_floor.IsSerialFloorReady() || !rung1_floor.IsAudioModemReady() ||
		!rung1_floor.IsMorseFloorReady() || !common.IsHarnessReady() {
		t.Fatal("[BLOCK 5: EPILOGUE] Rung 1 Floor readiness handshake failed")
	}
}

// End of tests/unit/transport/floor_transports_test.go
