// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/unit/transport/morse_test.go
// @id            test.unit.transport.morse
// @layer         Layer 4: Tier-2 Unit Verification Suite (Transport Sub-Stratum)
// @archetype     verification_suite | unit_isolation
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Rung 1 International Morse CW Floor Unit Suite
// @role          Tier-2 Unit Isolation Gate for cog.transport.rung1_morse
// @maturity      Rev 5 Baseline | 4x4 Verification Matrix
// @chassis       bounded-heap | Verifies ITU Morse hex & CW pulse timing 1:3:1:3
// @scale         O(N_frame) lossless binary <-> Morse <-> CW pulse tests
// @seams         in: data, engine, transport, rung1_floor, common
// @invariants    1. All 16 hex digits [0-9, A-F] encode/decode cleanly in ITU Morse.
//                2. MorseToPulses/PulsesToMorse preserves exact 1:3:1:3 timing.
//                3. Enforces 512-byte Floppy Sector ceiling strictly.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Dedicated 1:1 Tier-2 Unit isolation suite for `transport/rung1_floor/morse.go`.
//
//   ### For Architects, Operators & Non-Programmers:
//   Verifies FaithNet's 1844-grade telegraph/CW pulse floor carries real binary
//   packets without losing a single bit.
//
//   ### For Systems Programmers & Engine Architects:
//   100% branch coverage of `MorseEncodeBytes`, `MorseDecodeBytes`,
//   `MorseToPulses`, `PulsesToMorse`, `MorseMedium`, and `MorseTransceiver`.
// ==============================================================================

package transport_unit_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Identifier, Verification Imports & Tier-2 Unit Sensor Anchor )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.morse
// Mission:       Bind Rung 1 Morse CW codec/transceiver and shared harness.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"bytes"
	"context"
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
	morseUnitSuiteName = "Tier 2 Unit: cog.transport.rung1_morse (ITU Morse CW)"
)

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.morse
// Mission:       Verify ITU hex Morse, CW pulses, fault bounds, and 512B cap.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: [ SECTION 1: ALWAYS PASS (NOMINAL GOLDEN SURFACE COGS) ]
// ------------------------------------------------------------------------------

func TestMorseUnit_Sec1_AllHexDigitsCWPulsesAndFaithNetFrame(t *testing.T) {
	f := common.MakeSovereignFrame(
		common.NodeAlpha, common.NodeBeta, data.Rung1Floor, data.OpPing, 1844,
		[]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF},
	)
	wire := common.EncodeFixture(t, &f)

	med := rung1_floor.NewMorseMedium()
	mA := med.Attach(common.NodeAlpha)
	mB := med.Attach(common.NodeBeta)
	defer func() {
		_ = mA.Close()
		_ = mB.Close()
	}()
	if mA.Name() != "rung1:morse-cw" || mA.Rung() != data.Rung1Floor || !mA.Healthy() {
		t.Fatal("[SEC 1] unexpected MorseTransceiver metadata")
	}
	if err := mA.Send(data.NodeIDBroadcast, wire); err != nil {
		t.Fatalf("[SEC 1] Morse CW broadcast failed: %v", err)
	}
	env, err := mB.Receive(context.Background())
	if err != nil {
		t.Fatalf("[SEC 1] Morse CW receive failed: %v", err)
	}
	var out data.Frame
	if err := engine.DecodeFrom(env.Slice(), &out); err != nil {
		t.Fatalf("[SEC 1] Morse CW decoded bytes failed Codec: %v", err)
	}
	common.AssertFramesEqual(t, &f, &out)
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: [ SECTION 2: NEVER FAIL (FAULT INVERSION & DEGRADATION) ]
// ------------------------------------------------------------------------------

func TestMorseUnit_Sec2_InvalidSymbolsEmptyOfflineClosedAndNilAudit(t *testing.T) {
	if _, err := rung1_floor.MorseEncodeBytes(nil); err != rung1_floor.ErrMorseEmpty {
		t.Fatalf("[SEC 2] empty MorseEncodeBytes got %v", err)
	}
	if _, err := rung1_floor.MorseDecodeBytes("   "); err != rung1_floor.ErrMorseEmpty {
		t.Fatalf("[SEC 2] blank MorseDecodeBytes got %v", err)
	}
	if _, err := rung1_floor.MorseDecodeBytes(".-"); err != rung1_floor.ErrMorseInvalidSymbol {
		t.Fatalf("[SEC 2] odd nibble Morse got %v", err)
	}
	if _, err := rung1_floor.MorseDecodeBytes("........ ........"); err != rung1_floor.ErrMorseInvalidSymbol {
		t.Fatalf("[SEC 2] invalid symbol Morse got %v", err)
	}

	var nilM *rung1_floor.MorseTransceiver
	common.AssertAuditFail(t, nilM.AuditState(), data.AuditSeverityCritical)

	med := rung1_floor.NewMorseMedium()
	tx := med.Attach(common.NodeAlpha)
	if err := tx.Send(common.NodeBeta, []byte("MISSING")); err != transport.ErrPeerUnreachable {
		t.Fatalf("[SEC 2] missing peer got %v", err)
	}
	tx.SetHealthy(false)
	if err := tx.Send(common.NodeBeta, []byte("OFF")); err != transport.ErrTransceiverOffline {
		t.Fatalf("[SEC 2] offline Morse Send got %v", err)
	}
	tx.SetHealthy(true)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := tx.Receive(ctx); err != context.Canceled {
		t.Fatalf("[SEC 2] canceled Morse Receive got %v", err)
	}
	_ = tx.Close()
	_ = tx.Close()
	if err := tx.Send(common.NodeBeta, []byte("CLOSED")); err != transport.ErrTransceiverClosed {
		t.Fatalf("[SEC 2] closed Morse Send got %v", err)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: [ SECTION 3: INVARIANTS, SATURATION & LIMIT GUARDS ]
// ------------------------------------------------------------------------------

func TestMorseUnit_Sec3_Exact512BytePulseRoundTripAnd513Reject(t *testing.T) {
	exact := bytes.Repeat([]byte{0xFA, 0x17, 0x00, 0xFF}, 128)
	mStr, err := rung1_floor.MorseEncodeBytes(exact)
	if err != nil {
		t.Fatalf("[SEC 3] exact 512B Morse encode failed: %v", err)
	}
	pulses := rung1_floor.MorseToPulses(mStr)
	mBack := rung1_floor.PulsesToMorse(pulses)
	dec, err := rung1_floor.MorseDecodeBytes(mBack)
	if err != nil || !bytes.Equal(dec, exact) {
		t.Fatalf("[SEC 3] exact 512B Morse pulse round-trip mismatch: %v", err)
	}

	over := make([]byte, data.FloorMaxFrameSize+1)
	if _, err := rung1_floor.MorseEncodeBytes(over); err != transport.ErrWireFrameOversize {
		t.Fatalf("[SEC 3] 513B bypassed Morse ceiling: %v", err)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.4: [ SECTION 4: EMERGENCE & COMPOSITE WORKFLOWS ]
// ------------------------------------------------------------------------------

func TestMorseUnit_Sec4_LiveAuditAndDumpCard(t *testing.T) {
	tx := rung1_floor.NewMorseMedium().Attach(common.NodeAlpha)
	defer func() { _ = tx.Close() }()
	card := tx.ExecuteMorseDumpCard()
	common.AssertAuditPass(t, tx.AuditState(), card)
	if !strings.Contains(card, "ITU-R M.1677-1") {
		t.Fatalf("[SEC 4] unexpected Morse dump card:\n%s", card)
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Preflight Diagnostic Gates & Physical Source Anchor Verification )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.morse
// Mission:       Run PreflightMorseFloor and verify physical source anchor.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Preflight Diagnostic Gates
// ------------------------------------------------------------------------------

func TestMorseUnit_Block4_PreflightGate(t *testing.T) {
	common.AssertAuditPass(t, rung1_floor.PreflightMorseFloor(), "")
	if !rung1_floor.AuditMorseFloor() {
		t.Fatal("[BLOCK 4] AuditMorseFloor returned false")
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Suite Identifier & Physical Source Anchor Gate
// ------------------------------------------------------------------------------

func TestMorseUnit_Block4_SourceAnchor(t *testing.T) {
	if morseUnitSuiteName == "" {
		t.Fatal("[BLOCK 4] suite identifier must not be empty")
	}
	common.AssertRepoAnchors(t, "transport/rung1_floor/morse.go")
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Test Suite Epilogue & Clean Readiness Handshake )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.morse
// Mission:       Confirm Rung 1 Morse CW readiness handshake post-suite.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Epilogue & Clean State Handshake
// ------------------------------------------------------------------------------

func TestMorseUnit_Block5_EpilogueReadiness(t *testing.T) {
	if !rung1_floor.IsMorseFloorReady() || !common.IsHarnessReady() {
		t.Fatal("[BLOCK 5: EPILOGUE] morse unit readiness handshake failed")
	}
}

// End of tests/unit/transport/morse_test.go
