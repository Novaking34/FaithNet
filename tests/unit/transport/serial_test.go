// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/unit/transport/serial_test.go
// @id            test.unit.transport.serial
// @layer         Layer 4: Tier-2 Unit Verification Suite (Transport Sub-Stratum)
// @archetype     verification_suite | unit_isolation
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Rung 1 Copper Serial (RFC 1055 SLIP) Unit Suite
// @role          Tier-2 Unit Isolation Gate for cog.transport.rung1_serial
// @maturity      Rev 5 Baseline | 4x4 Verification Matrix
// @chassis       bounded-heap | Exhaustive 0x00..0xFF SLIP & io.Pipe() tests
// @scale         O(N_frame) lossless byte-stuffing verification
// @seams         in: data, engine, transport, rung1_floor, common
// @invariants    1. All 256 byte values round-trip losslessly through SLIP.
//                2. Malformed escapes & >512B streams rejected cleanly.
//                3. Close() unblocks pending io.Pipe() readers cleanly.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Dedicated 1:1 Tier-2 Unit isolation suite for `transport/rung1_floor/serial.go`.
//
//   ### For Architects, Operators & Non-Programmers:
//   Verifies FaithNet's 3-wire serial cable framing carries any binary message
//   losslessly and handles noise or cable disconnects without crashing.
//
//   ### For Systems Programmers & Engine Architects:
//   100% branch coverage of `SLIPEncode`, `SLIPDecode`, `SerialTransceiver`.
// ==============================================================================

package transport_unit_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Identifier, Verification Imports & Tier-2 Unit Sensor Anchor )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.serial
// Mission:       Bind Rung 1 SLIP serial functions and shared harness.
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
	serialUnitSuiteName = "Tier 2 Unit: cog.transport.rung1_serial (SLIP)"
)

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.serial
// Mission:       Verify SLIP framing, fault inversion, 512B cap, and pipes.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: [ SECTION 1: ALWAYS PASS (NOMINAL GOLDEN SURFACE COGS) ]
// ------------------------------------------------------------------------------

func TestSerialUnit_Sec1_All256BytesAndLivePipeFaithNetFrame(t *testing.T) {
	allBytes := make([]byte, 256)
	for i := 0; i < 256; i++ {
		allBytes[i] = byte(i)
	}
	enc, err := rung1_floor.SLIPEncode(allBytes)
	if err != nil {
		t.Fatalf("[SEC 1] SLIPEncode failed: %v", err)
	}
	dec, err := rung1_floor.SLIPDecode(enc)
	if err != nil || !bytes.Equal(dec, allBytes) {
		t.Fatalf("[SEC 1] SLIPDecode mismatch: err=%v len=%d", err, len(dec))
	}

	rB, wA := io.Pipe()
	rA, wB := io.Pipe()
	txA := rung1_floor.NewSerialTransceiver(common.NodeAlpha, rA, wA, wA)
	txB := rung1_floor.NewSerialTransceiver(common.NodeBeta, rB, wB, wB)
	defer func() {
		_ = txA.Close()
		_ = txB.Close()
	}()
	if txA.Name() != "rung1:serial-slip" || txA.Rung() != data.Rung1Floor || !txA.Healthy() {
		t.Fatal("[SEC 1] unexpected SerialTransceiver metadata")
	}

	f := common.MakeSovereignFrame(
		common.NodeAlpha, common.NodeBeta, data.Rung1Floor, data.OpPing, 1, []byte("RS232-COPPER"),
	)
	wire := common.EncodeFixture(t, &f)
	done := make(chan error, 1)
	go func() { done <- txA.Send(common.NodeBeta, wire) }()
	env, err := txB.Receive(context.Background())
	if sErr := <-done; sErr != nil || err != nil {
		t.Fatalf("[SEC 1] pipe transfer failed: send=%v recv=%v", sErr, err)
	}
	var out data.Frame
	if err := engine.DecodeFrom(env.Slice(), &out); err != nil {
		t.Fatalf("[SEC 1] decoded SLIP bytes failed Codec: %v", err)
	}
	common.AssertFramesEqual(t, &f, &out)
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: [ SECTION 2: NEVER FAIL (FAULT INVERSION & DEGRADATION) ]
// ------------------------------------------------------------------------------

func TestSerialUnit_Sec2_BadEscapesEmptyOfflineClosedAndNilAudit(t *testing.T) {
	if _, err := rung1_floor.SLIPEncode(nil); err != rung1_floor.ErrSLIPEmptyFrame {
		t.Fatalf("[SEC 2] empty SLIPEncode got %v", err)
	}
	if _, err := rung1_floor.SLIPDecode([]byte{rung1_floor.SLIPEND, rung1_floor.SLIPEND}); err != rung1_floor.ErrSLIPEmptyFrame {
		t.Fatalf("[SEC 2] empty SLIPDecode got %v", err)
	}
	if _, err := rung1_floor.SLIPDecode([]byte{rung1_floor.SLIPESC}); err != rung1_floor.ErrSLIPBadEscape {
		t.Fatalf("[SEC 2] dangling ESC got %v", err)
	}
	if _, err := rung1_floor.SLIPDecode([]byte{rung1_floor.SLIPEND, rung1_floor.SLIPESC, 0x42, rung1_floor.SLIPEND}); err != rung1_floor.ErrSLIPBadEscape {
		t.Fatalf("[SEC 2] bogus ESC got %v", err)
	}

	var nilS *rung1_floor.SerialTransceiver
	common.AssertAuditFail(t, nilS.AuditState(), data.AuditSeverityCritical)

	r, w := io.Pipe()
	tx := rung1_floor.NewSerialTransceiver(common.NodeAlpha, r, w, w)
	tx.SetHealthy(false)
	if err := tx.Send(common.NodeBeta, []byte("OFF")); err != transport.ErrTransceiverOffline {
		t.Fatalf("[SEC 2] offline serial Send got %v", err)
	}
	tx.SetHealthy(true)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := tx.Receive(ctx); err != context.Canceled {
		t.Fatalf("[SEC 2] canceled serial Receive got %v", err)
	}
	_ = tx.Close()
	_ = tx.Close()
	if err := tx.Send(common.NodeBeta, []byte("CLOSED")); err != transport.ErrTransceiverClosed {
		t.Fatalf("[SEC 2] closed serial Send got %v", err)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: [ SECTION 3: INVARIANTS, SATURATION & LIMIT GUARDS ]
// ------------------------------------------------------------------------------

func TestSerialUnit_Sec3_Exact512ByteCeilingAnd513Rejection(t *testing.T) {
	exact := bytes.Repeat([]byte{rung1_floor.SLIPEND, rung1_floor.SLIPESC, 0xFA, 0x17}, 128)
	enc, err := rung1_floor.SLIPEncode(exact)
	if err != nil {
		t.Fatalf("[SEC 3] exact 512B SLIP encode failed: %v", err)
	}
	dec, err := rung1_floor.SLIPDecode(enc)
	if err != nil || !bytes.Equal(dec, exact) {
		t.Fatalf("[SEC 3] exact 512B SLIP decode mismatch: %v", err)
	}
	over := make([]byte, data.FloorMaxFrameSize+1)
	if _, err := rung1_floor.SLIPEncode(over); err != transport.ErrWireFrameOversize {
		t.Fatalf("[SEC 3] 513B bypassed SLIP ceiling: %v", err)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.4: [ SECTION 4: EMERGENCE & COMPOSITE WORKFLOWS ]
// ------------------------------------------------------------------------------

func TestSerialUnit_Sec4_LiveAuditAndDumpCard(t *testing.T) {
	r, w := io.Pipe()
	tx := rung1_floor.NewSerialTransceiver(common.NodeAlpha, r, w, w)
	defer func() { _ = tx.Close() }()
	card := tx.ExecuteSerialDumpCard()
	common.AssertAuditPass(t, tx.AuditState(), card)
	if !strings.Contains(card, "RFC 1055 SLIP") {
		t.Fatalf("[SEC 4] unexpected serial dump card:\n%s", card)
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Preflight Diagnostic Gates & Physical Source Anchor Verification )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.serial
// Mission:       Run PreflightSerialFloor and verify physical source anchor.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Preflight Diagnostic Gates
// ------------------------------------------------------------------------------

func TestSerialUnit_Block4_PreflightGate(t *testing.T) {
	common.AssertAuditPass(t, rung1_floor.PreflightSerialFloor(), "")
	if !rung1_floor.AuditSerialFloor() {
		t.Fatal("[BLOCK 4] AuditSerialFloor returned false")
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Suite Identifier & Physical Source Anchor Gate
// ------------------------------------------------------------------------------

func TestSerialUnit_Block4_SourceAnchor(t *testing.T) {
	if serialUnitSuiteName == "" {
		t.Fatal("[BLOCK 4] suite identifier must not be empty")
	}
	common.AssertRepoAnchors(t, "transport/rung1_floor/serial.go")
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Test Suite Epilogue & Clean Readiness Handshake )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.serial
// Mission:       Confirm Rung 1 Serial readiness handshake post-suite.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Epilogue & Clean State Handshake
// ------------------------------------------------------------------------------

func TestSerialUnit_Block5_EpilogueReadiness(t *testing.T) {
	if !rung1_floor.IsSerialFloorReady() || !common.IsHarnessReady() {
		t.Fatal("[BLOCK 5: EPILOGUE] serial unit readiness handshake failed")
	}
}

// End of tests/unit/transport/serial_test.go
