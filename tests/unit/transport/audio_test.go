// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/unit/transport/audio_test.go
// @id            test.unit.transport.audio
// @layer         Layer 4: Tier-2 Unit Verification Suite (Transport Sub-Stratum)
// @archetype     verification_suite | unit_isolation
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Rung 1 AFSK 1200/2400 Hz Audio Modem Unit Suite
// @role          Tier-2 Unit Isolation Gate for cog.transport.rung1_audio
// @maturity      Rev 5 Baseline | 4x4 Verification Matrix
// @chassis       bounded-heap | Real PCM sine synthesis & Goertzel DSP tests
// @scale         O(N_bits * 16) acoustic modulation, noise & medium tests
// @seams         in: data, engine, transport, rung1_floor, common
// @invariants    1. Orthogonal 1200/2400 Hz AFSK recovers 0xFA17 frames bit-exact.
//                2. Survives additive analog channel static cleanly.
//                3. Enforces 512-byte Floppy Sector ceiling strictly.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Dedicated 1:1 Tier-2 Unit isolation suite for `transport/rung1_floor/audio.go`.
//
//   ### For Architects, Operators & Non-Programmers:
//   Verifies FaithNet's audible radio/dial-up modem turns frames into tones and
//   decodes them back cleanly even through analog static.
//
//   ### For Systems Programmers & Engine Architects:
//   100% branch coverage of `ModulateAFSK`, `DemodulateAFSK`, `AudioMedium`,
//   and `AudioTransceiver`.
// ==============================================================================

package transport_unit_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Identifier, Verification Imports & Tier-2 Unit Sensor Anchor )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.audio
// Mission:       Bind Rung 1 AFSK audio modem and shared test harness.
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
	audioUnitSuiteName = "Tier 2 Unit: cog.transport.rung1_audio (AFSK Modem)"
)

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.audio
// Mission:       Verify PCM modulation, Goertzel static immunity & AudioMedium.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: [ SECTION 1: ALWAYS PASS (NOMINAL GOLDEN SURFACE COGS) ]
// ------------------------------------------------------------------------------

func TestAudioUnit_Sec1_AFSKUnicastBroadcastAndFaithNetFrame(t *testing.T) {
	f := common.MakeSovereignFrame(
		common.NodeAlpha, common.NodeBeta, data.Rung1Floor, data.OpCPISync, 3,
		[]byte("AFSK-WALKIE-TALKIE"),
	)
	wire := common.EncodeFixture(t, &f)

	med := rung1_floor.NewAudioMedium()
	aA := med.Attach(common.NodeAlpha)
	aB := med.Attach(common.NodeBeta)
	defer func() {
		_ = aA.Close()
		_ = aB.Close()
	}()
	if aA.Name() != "rung1:audio-afsk" || aA.Rung() != data.Rung1Floor || !aA.Healthy() {
		t.Fatal("[SEC 1] unexpected AudioTransceiver metadata")
	}
	if err := aA.Send(data.NodeIDBroadcast, wire); err != nil {
		t.Fatalf("[SEC 1] AFSK broadcast failed: %v", err)
	}
	env, err := aB.Receive(context.Background())
	if err != nil {
		t.Fatalf("[SEC 1] AFSK receive failed: %v", err)
	}
	var out data.Frame
	if err := engine.DecodeFrom(env.Slice(), &out); err != nil {
		t.Fatalf("[SEC 1] AFSK decoded bytes failed Codec: %v", err)
	}
	common.AssertFramesEqual(t, &f, &out)
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: [ SECTION 2: NEVER FAIL (FAULT INVERSION & DEGRADATION) ]
// ------------------------------------------------------------------------------

func TestAudioUnit_Sec2_SilenceTruncatedOfflineClosedAndNilAudit(t *testing.T) {
	if _, err := rung1_floor.ModulateAFSK(nil); err != rung1_floor.ErrAFSKTruncated {
		t.Fatalf("[SEC 2] empty ModulateAFSK got %v", err)
	}
	if _, err := rung1_floor.DemodulateAFSK(nil); err != rung1_floor.ErrAFSKTruncated {
		t.Fatalf("[SEC 2] nil DemodulateAFSK got %v", err)
	}
	if _, err := rung1_floor.DemodulateAFSK(make([]int16, 1024)); err != rung1_floor.ErrAFSKBadPreamble {
		t.Fatalf("[SEC 2] dead silence DemodulateAFSK got %v", err)
	}

	var nilA *rung1_floor.AudioTransceiver
	common.AssertAuditFail(t, nilA.AuditState(), data.AuditSeverityCritical)

	med := rung1_floor.NewAudioMedium()
	tx := med.Attach(common.NodeAlpha)
	if err := tx.Send(common.NodeBeta, []byte("MISSING")); err != transport.ErrPeerUnreachable {
		t.Fatalf("[SEC 2] missing peer got %v", err)
	}
	tx.SetHealthy(false)
	if err := tx.Send(common.NodeBeta, []byte("OFF")); err != transport.ErrTransceiverOffline {
		t.Fatalf("[SEC 2] offline Audio Send got %v", err)
	}
	tx.SetHealthy(true)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := tx.Receive(ctx); err != context.Canceled {
		t.Fatalf("[SEC 2] canceled Audio Receive got %v", err)
	}
	_ = tx.Close()
	_ = tx.Close()
	if err := tx.Send(common.NodeBeta, []byte("CLOSED")); err != transport.ErrTransceiverClosed {
		t.Fatalf("[SEC 2] closed Audio Send got %v", err)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: [ SECTION 3: INVARIANTS, SATURATION & LIMIT GUARDS ]
// ------------------------------------------------------------------------------

func TestAudioUnit_Sec3_512ByteCeilingAndAdditiveAnalogStaticImmunity(t *testing.T) {
	exact := bytes.Repeat([]byte{0xFA, 0x17, 0x55, 0xAA}, 128)
	pcm, err := rung1_floor.ModulateAFSK(exact)
	if err != nil {
		t.Fatalf("[SEC 3] exact 512B AFSK modulate failed: %v", err)
	}
	// Inject deterministic +/- 5000 PCM static (~20% peak amplitude) across all samples.
	for i := range pcm {
		if i%3 == 0 {
			pcm[i] += 5000
		} else {
			pcm[i] -= 5000
		}
	}
	dec, err := rung1_floor.DemodulateAFSK(pcm)
	if err != nil || !bytes.Equal(dec, exact) {
		t.Fatalf("[SEC 3] Goertzel demodulator failed 512B recovery under static: %v", err)
	}

	over := make([]byte, data.FloorMaxFrameSize+1)
	if _, err := rung1_floor.ModulateAFSK(over); err != transport.ErrWireFrameOversize {
		t.Fatalf("[SEC 3] 513B bypassed AFSK ceiling: %v", err)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.4: [ SECTION 4: EMERGENCE & COMPOSITE WORKFLOWS ]
// ------------------------------------------------------------------------------

func TestAudioUnit_Sec4_LiveAuditAndDumpCard(t *testing.T) {
	tx := rung1_floor.NewAudioMedium().Attach(common.NodeAlpha)
	defer func() { _ = tx.Close() }()
	card := tx.ExecuteAudioDumpCard()
	common.AssertAuditPass(t, tx.AuditState(), card)
	if !strings.Contains(card, "1200Hz Mark / 2400Hz Space") {
		t.Fatalf("[SEC 4] unexpected audio dump card:\n%s", card)
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Preflight Diagnostic Gates & Physical Source Anchor Verification )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.audio
// Mission:       Run PreflightAudioModem and verify physical source anchor.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Preflight Diagnostic Gates
// ------------------------------------------------------------------------------

func TestAudioUnit_Block4_PreflightGate(t *testing.T) {
	common.AssertAuditPass(t, rung1_floor.PreflightAudioModem(), "")
	if !rung1_floor.AuditAudioModem() {
		t.Fatal("[BLOCK 4] AuditAudioModem returned false")
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Suite Identifier & Physical Source Anchor Gate
// ------------------------------------------------------------------------------

func TestAudioUnit_Block4_SourceAnchor(t *testing.T) {
	if audioUnitSuiteName == "" {
		t.Fatal("[BLOCK 4] suite identifier must not be empty")
	}
	common.AssertRepoAnchors(t, "transport/rung1_floor/audio.go")
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Test Suite Epilogue & Clean Readiness Handshake )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.transport.audio
// Mission:       Confirm Rung 1 Audio readiness handshake post-suite.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Epilogue & Clean State Handshake
// ------------------------------------------------------------------------------

func TestAudioUnit_Block5_EpilogueReadiness(t *testing.T) {
	if !rung1_floor.IsAudioModemReady() || !common.IsHarnessReady() {
		t.Fatal("[BLOCK 5: EPILOGUE] audio unit readiness handshake failed")
	}
}

// End of tests/unit/transport/audio_test.go
