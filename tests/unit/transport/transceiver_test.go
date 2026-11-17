// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/unit/transport/transceiver_test.go
// @id            test.unit.transport.transceiver
// @layer         Layer 4: Tier-2 Unit Verification Suite (Transport Sub-Stratum)
// @archetype     verification_suite | unit_isolation
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Cog 3 Universal Transceiver Contract & Envelope Suite
// @role          Tier-2 Unit Isolation Gate for cog.transport.transceiver
// @maturity      Rev 5 Baseline | 4x4 Verification Matrix
// @chassis       zero-heap | Verifies WireEnvelope bounds, Slice() & sentinels
// @scale         O(1) deterministic contract & envelope saturation proofs
// @seams         in: data, transport, tests/common | target: transceiver.go
// @invariants    1. WireEnvelope.Slice() saturates [0..512] without panicking.
//                2. All 5 transport sentinel errors are non-nil and distinct.
//                3. WireEnvelope.Slice() performs 0.0 allocs/op.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Dedicated 1:1 Tier-2 Unit isolation suite for `transport/transceiver.go`.
//
//   ### For Architects, Operators & Non-Programmers:
//   Verifies FaithNet's universal wire plug and 512-byte carrier envelope
//   behave identically and safely across all four survival rungs.
//
//   ### For Systems Programmers & Engine Architects:
//   100% branch coverage of `WireEnvelope.Slice()`, sentinel error uniqueness,
//   and interface satisfaction verification across all 4 survival rungs.
// ==============================================================================

package transport_unit_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Identifier, Verification Imports & Tier-2 Unit Sensor Anchor )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.transceiver
// Mission:       Bind Cog 3 Transceiver contract and concrete implementations.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"bytes"
	"testing"
)

// --- Tier 2: External Library Dependencies (None) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---
import (
	"github.com/CreativeWorkzStudio/FaithNet/data"
	"github.com/CreativeWorkzStudio/FaithNet/tests/common"
	"github.com/CreativeWorkzStudio/FaithNet/transport"
	"github.com/CreativeWorkzStudio/FaithNet/transport/rung1_floor"
	"github.com/CreativeWorkzStudio/FaithNet/transport/rung2_host"
	"github.com/CreativeWorkzStudio/FaithNet/transport/rung3_local"
	"github.com/CreativeWorkzStudio/FaithNet/transport/rung4_global"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Operational Suite Constants & Identifiers
// ------------------------------------------------------------------------------

const (
	transceiverUnitSuiteName = "Tier 2 Unit: cog.transport.transceiver (Cog 3)"
)

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.transceiver
// Mission:       Verify WireEnvelope, sentinels, and Cog 3 interface parity.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: [ SECTION 1: ALWAYS PASS (NOMINAL GOLDEN SURFACE COGS) ]
// ------------------------------------------------------------------------------

func TestTransceiverUnit_Sec1_WireEnvelopeNominalSliceAndSentinels(t *testing.T) {
	var env transport.WireEnvelope
	payload := []byte("FAITHNET-COG3")
	env.SenderID = common.NodeAlpha
	env.TargetID = common.NodeBeta
	env.Length = len(payload)
	copy(env.Bytes[:], payload)

	if !bytes.Equal(env.Slice(), payload) {
		t.Fatalf("[SEC 1] WireEnvelope.Slice() mismatch: got %q", env.Slice())
	}
	sentinels := []error{
		transport.ErrTransceiverClosed,
		transport.ErrTransceiverOffline,
		transport.ErrPeerUnreachable,
		transport.ErrWireFrameOversize,
		transport.ErrReceiveTimeout,
	}
	for i := 0; i < len(sentinels); i++ {
		if sentinels[i] == nil {
			t.Fatalf("[SEC 1] sentinel %d is nil", i)
		}
		for j := i + 1; j < len(sentinels); j++ {
			if sentinels[i] == sentinels[j] {
				t.Fatalf("[SEC 1] duplicate sentinel errors at %d and %d", i, j)
			}
		}
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: [ SECTION 2: NEVER FAIL (FAULT INVERSION & DEGRADATION) ]
// ------------------------------------------------------------------------------

func TestTransceiverUnit_Sec2_WireEnvelopeNegativeAndOversizeLengthClamp(t *testing.T) {
	neg := transport.WireEnvelope{Length: -42}
	if len(neg.Slice()) != 0 {
		t.Fatalf("[SEC 2] negative Length did not clamp to 0: got %d", len(neg.Slice()))
	}
	over := transport.WireEnvelope{Length: 99_999}
	if len(over.Slice()) != data.FloorMaxFrameSize {
		t.Fatalf("[SEC 2] oversize Length did not clamp to 512: got %d", len(over.Slice()))
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: [ SECTION 3: INVARIANTS, SATURATION & LIMIT GUARDS ]
// ------------------------------------------------------------------------------

func TestTransceiverUnit_Sec3_ZeroAllocSliceAndCompileTimeImplementations(t *testing.T) {
	var env transport.WireEnvelope
	env.Length = 64
	allocs := testing.AllocsPerRun(500, func() {
		_ = env.Slice()
	})
	if allocs != 0 {
		t.Fatalf("[SEC 3] WireEnvelope.Slice() violated ALLOC_ZERO_HEAP: %.2f", allocs)
	}

	// Compile-time proof that all 7 concrete mediums satisfy Cog 3 Transceiver.
	var (
		_ transport.Transceiver = (*rung1_floor.SerialTransceiver)(nil)
		_ transport.Transceiver = (*rung1_floor.AudioTransceiver)(nil)
		_ transport.Transceiver = (*rung1_floor.MorseTransceiver)(nil)
		_ transport.Transceiver = (*rung2_host.MemoryTransceiver)(nil)
		_ transport.Transceiver = (*rung2_host.IPCTransceiver)(nil)
		_ transport.Transceiver = (*rung3_local.UDPTransceiver)(nil)
		_ transport.Transceiver = (*rung4_global.TCPTransceiver)(nil)
	)
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.4: [ SECTION 4: EMERGENCE & COMPOSITE WORKFLOWS ]
// ------------------------------------------------------------------------------

func TestTransceiverUnit_Sec4_Exact512ByteEnvelopeRoundTrip(t *testing.T) {
	var env transport.WireEnvelope
	full := bytes.Repeat([]byte{0xFA}, data.FloorMaxFrameSize)
	env.Length = len(full)
	copy(env.Bytes[:], full)
	if !bytes.Equal(env.Slice(), full) || transport.ModuleIDTransceiver == "" {
		t.Fatal("[SEC 4] exact 512B WireEnvelope view or module ID failed")
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Preflight Diagnostic Gates & Physical Source Anchor Verification )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.transceiver
// Mission:       Run harness self-audit and verify transport/transceiver.go.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Preflight Diagnostic Gates
// ------------------------------------------------------------------------------

func TestTransceiverUnit_Block4_PreflightGate(t *testing.T) {
	common.AssertAuditPass(t, common.SelfAudit(), common.ExecuteHarnessDumpCard())
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Suite Identifier & Physical Source Anchor Gate
// ------------------------------------------------------------------------------

func TestTransceiverUnit_Block4_SourceAnchor(t *testing.T) {
	if transceiverUnitSuiteName == "" {
		t.Fatal("[BLOCK 4] suite identifier must not be empty")
	}
	common.AssertRepoAnchors(t, "transport/transceiver.go")
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Test Suite Epilogue & Clean Readiness Handshake )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.transceiver
// Mission:       Confirm harness readiness after Cog 3 contract unit suite.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Epilogue & Clean State Handshake
// ------------------------------------------------------------------------------

func TestTransceiverUnit_Block5_EpilogueReadiness(t *testing.T) {
	if !common.IsHarnessReady() {
		t.Fatal("[BLOCK 5: EPILOGUE] transceiver unit readiness handshake failed")
	}
}

// End of tests/unit/transport/transceiver_test.go
