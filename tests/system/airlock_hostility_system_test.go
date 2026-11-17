// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/system/airlock_hostility_system_test.go
// @id            test.system.airlock_hostility
// @layer         Layer 4: Tier-4 System End-to-End & Stress Verification Suite
// @archetype     verification_suite | system_security_hostility
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Genesis 3 Airlock & Full-Stack Hostility Storm Suite
// @role          Tier-4 System Security Gate: Zero Breaches Under Attack Storm
// @maturity      Rev 5 Baseline | Wave 6 4x4 Symmetry Expansion
// @chassis       bounded-heap | Concurrent adversarial storm across ALL borders
// @scale         Multi-vector concurrent attack storm (Codec, Airlock, Well, Egress)
// @seams         in: data, engine, engine/covenant, orchestrator, interpreter, rung4_global
// @invariants    1. Macro Tier 4 (System Security): zero hostile payloads cross
//                   ANY border—wire Codec, Genesis 3 Airlock, Dispatcher,
//                   Translator, WellServer HTTP, or Rung 4 EgressGateway.
//                2. Zero panics, zero OOM allocations, zero data races under
//                   concurrent multi-goroutine hostile bombardment.
//                3. Clean sovereign & 0:AWAIT traffic continues unimpeded during
//                   active hostile bombardment (zero denial of service).
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Dedicated Tier-4 System security proving ground subjecting EVERY FaithNet
//   ingress/egress boundary simultaneously to XSS, SQL injection, PHP tags,
//   null-byte smuggling, forged `FlagCovenantVerified` spoofing, malformed wire
//   headers, and >480B OOM body attacks—verifying 0 breaches and 0 crashes.
//
//   ### For Architects, Operators & Non-Programmers:
//   Attacks every door of FaithNet at the same time (wire frames, browser forms,
//   and outside web fetches) while clean community messages keep flowing,
//   proving the Genesis 3 Airlock holds firm with zero leaks or crashes.
//
//   ### For Systems Programmers & Engine Architects:
//   Concurrently drives `engine.DecodeFrom`, `covenant.Filter`, `Dispatcher`,
//   `Translator`, `WellServer`, and `EgressGateway` (`httptest`) under `-race`.
// ==============================================================================

package system_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Identifier, Verification Imports & Adversarial Corpus )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-4 System Verification Suite
// Module:        test.system.airlock_hostility
// Mission:       Define adversarial corpus covering all Genesis 3 threat classes.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// --- Tier 2: External Library Dependencies (None — 100% Go Stdlib Invariant) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---
import (
	"github.com/CreativeWorkzStudio/FaithNet/data"
	"github.com/CreativeWorkzStudio/FaithNet/engine"
	"github.com/CreativeWorkzStudio/FaithNet/engine/covenant"
	"github.com/CreativeWorkzStudio/FaithNet/interpreter"
	"github.com/CreativeWorkzStudio/FaithNet/orchestrator"
	"github.com/CreativeWorkzStudio/FaithNet/tests/common"
	"github.com/CreativeWorkzStudio/FaithNet/transport"
	"github.com/CreativeWorkzStudio/FaithNet/transport/rung4_global"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Operational Suite Constants & Hostile Attack Corpus
// ------------------------------------------------------------------------------

const (
	hostilitySuiteName = "Tier 4 System: Genesis 3 Airlock Hostility Storm"
)

var hostileAttackCorpus = [][]byte{
	[]byte("<script>alert(1)</script>"),
	[]byte("<SCRIPT SRC=//evil.test/x.js></SCRIPT>"),
	[]byte("javascript:fetch('http://evil')"),
	[]byte("SELECT * FROM users; DROP TABLE sessions;"),
	[]byte("admin' UNION SELECT password FROM vault --"),
	[]byte("<?php system($_GET['cmd']); ?>"),
	[]byte("clean-prefix\x00hidden-smuggled-bytes"),
	[]byte(strings.Repeat("A", data.FloorMaxPayload+32)),
}

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-4 System Verification Suite
// Module:        test.system.airlock_hostility
// Mission:       Verify zero breaches across Airlock, Well, Egress, and Codec
//                while clean traffic survives uninterrupted under bombardment.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: [ SECTION 1: ALWAYS PASS (CLEAN TRAFFIC DURING STORM) ]
// ------------------------------------------------------------------------------

func TestAirlockHostilitySystem_Sec1_CleanTrafficFlowsUnimpededDuringStorm(t *testing.T) {
	disp, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(common.NodeAlpha, "fortress-node"), engine.NewMonotonicClock(),
	)
	defer func() { _ = disp.Close() }()
	well := interpreter.NewWellServer(disp, "127.0.0.1:0")

	// Hammer Well with all hostile payloads first.
	for _, atk := range hostileAttackCorpus {
		rec := httptest.NewRecorder()
		well.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/message",
			strings.NewReader(string(atk))))
		if rec.Code == http.StatusAccepted || rec.Code == http.StatusOK {
			t.Fatalf("[SEC 1] hostile payload accepted by Well: %q (HTTP %d)", atk, rec.Code)
		}
	}

	// Clean neighborhood message must still be accepted cleanly after bombardment.
	okRec := httptest.NewRecorder()
	well.Handler().ServeHTTP(okRec, httptest.NewRequest(http.MethodPost, "/api/message",
		strings.NewReader("Water pump 2 is online and verified.")))
	if okRec.Code != http.StatusAccepted {
		t.Fatalf("[SEC 1] clean message blocked after storm: HTTP %d", okRec.Code)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: [ SECTION 2: NEVER FAIL (EVERY BORDER BLOCKS EVERY VECTOR) ]
// ------------------------------------------------------------------------------

func TestAirlockHostilitySystem_Sec2_AllFourSecurityBordersBlockEveryThreatVector(t *testing.T) {
	airlock := covenant.NewFilter()
	tr := interpreter.NewTranslator(airlock)
	disp, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(common.NodeAlpha, "border-guard"), engine.NewMonotonicClock(),
	)
	defer func() { _ = disp.Close() }()

	// 1. Airlock direct + Translator boundary.
	for i, atk := range hostileAttackCorpus {
		if v := airlock.EvaluateExternalPayload(atk); !v.IsBlocked() {
			t.Fatalf("[SEC 2] vector #%d bypassed Airlock: %q -> %+v", i, atk, v)
		}
		if _, v, err := tr.WrapForeignMessage(
			common.NodeExternal, common.NodeAlpha, data.Rung3Local, uint32(i+1), atk,
		); err == nil || !v.IsBlocked() {
			t.Fatalf("[SEC 2] vector #%d bypassed Translator: v=%+v err=%v", i, v, err)
		}
	}

	// 2. Wire Codec boundary (truncated, bad magic, bad version, tampered CRC).
	malformedWires := [][]byte{
		nil,
		{},
		{0xDE, 0xAD, 0xBE, 0xEF},
		make([]byte, data.FrameHeaderSize),
		make([]byte, data.FloorMaxFrameSize+16),
	}
	var dst data.Frame
	for i, raw := range malformedWires {
		if err := engine.DecodeFrom(raw, &dst); err == nil {
			t.Fatalf("[SEC 2] malformed wire #%d bypassed Codec", i)
		}
	}

	// 3. Dispatcher inbound boundary with forged FlagCovenantVerified on external origin.
	forged := data.Frame{
		Version:  data.ProtocolVersion,
		Rung:     data.Rung4Global,
		Opcode:   data.OpLegacyWrap,
		Flags:    data.FlagExternalOrigin | data.FlagCovenantVerified,
		Truth:    data.QuadPositive,
		SourceID: common.NodeExternal,
		TargetID: common.NodeAlpha,
		SeqID:    99,
	}
	forged.SetPayload([]byte("SPOOFED-TRUST"))
	wire := common.EncodeFixture(t, &forged)
	var env transport.WireEnvelope
	env.SenderID = common.NodeExternal
	env.TargetID = common.NodeAlpha
	env.Length = len(wire)
	copy(env.Bytes[:], wire)
	outcome := disp.HandleInboundEnvelope(env)
	if !outcome.Decoded || !outcome.Verdict.IsBlocked() {
		t.Fatalf("[SEC 2] forged FlagCovenantVerified bypassed Dispatcher: %+v", outcome)
	}

	// 4. Rung 4 EgressGateway against hostile httptest server.
	m := data.DefaultNodeManifest(common.NodeAlpha, "egress-guard")
	m.AllowGlobalEgress = true
	gw := rung4_global.NewEgressGateway(m, airlock)

	for i, atk := range hostileAttackCorpus {
		payload := atk
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write(payload)
		}))
		_, v, err := gw.FetchEmpiricalFrame(context.Background(), srv.URL, uint32(i+1))
		srv.Close()
		if err == nil || !v.IsBlocked() {
			t.Fatalf("[SEC 2] vector #%d bypassed Rung 4 EgressGateway: v=%+v err=%v", i, v, err)
		}
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: [ SECTION 3: INVARIANTS (CONCURRENT -RACE ATTACK STORM) ]
// ------------------------------------------------------------------------------

func TestAirlockHostilitySystem_Sec3_ConcurrentMultiGoroutineAttackStormZeroRaces(t *testing.T) {
	disp, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(common.NodeAlpha, "concurrent-shield"), engine.NewMonotonicClock(),
	)
	defer func() { _ = disp.Close() }()
	well := interpreter.NewWellServer(disp, "127.0.0.1:0")

	const goroutines = 12
	const iters = 25
	var wg sync.WaitGroup
	breachCh := make(chan string, goroutines*iters)

	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < iters; i++ {
				atk := hostileAttackCorpus[(id+i)%len(hostileAttackCorpus)]
				rec := httptest.NewRecorder()
				well.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/message",
					strings.NewReader(string(atk))))
				if rec.Code == http.StatusAccepted || rec.Code == http.StatusOK {
					breachCh <- string(atk)
				}
			}
		}(g)
	}
	wg.Wait()
	close(breachCh)

	for leaked := range breachCh {
		t.Fatalf("[SEC 3] concurrent storm breach: %q", leaked)
	}
	common.AssertAuditPass(t, disp.Airlock().AuditPolicy(), disp.Airlock().ExecuteAirlockDumpCard())
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.4: [ SECTION 4: EMERGENCE (AUDIT TELEMETRY COUNTS EVERY BLOCK) ]
// ------------------------------------------------------------------------------

func TestAirlockHostilitySystem_Sec4_DispatcherTelemetryTracksExactBlockedCount(t *testing.T) {
	disp, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(common.NodeAlpha, "telemetry-guard"), engine.NewMonotonicClock(),
	)
	defer func() { _ = disp.Close() }()

	// Inject hostile external frames through Dispatcher inbound envelope path.
	blockedExpected := uint64(0)
	for i, atk := range hostileAttackCorpus {
		if len(atk) > data.FloorMaxPayload {
			continue // Wire frames cap at 480B; oversize tested via Egress/Well.
		}
		f := data.Frame{
			Version:  data.ProtocolVersion,
			Rung:     data.Rung4Global,
			Opcode:   data.OpLegacyWrap,
			Flags:    data.FlagExternalOrigin,
			Truth:    data.QuadAwait,
			SourceID: common.NodeExternal,
			TargetID: common.NodeAlpha,
			SeqID:    uint32(i + 1),
		}
		f.SetPayload(atk)
		wire := common.EncodeFixture(t, &f)
		var env transport.WireEnvelope
		env.SenderID = common.NodeExternal
		env.TargetID = common.NodeAlpha
		env.Length = len(wire)
		copy(env.Bytes[:], wire)
		out := disp.HandleInboundEnvelope(env)
		if !out.Verdict.IsBlocked() {
			t.Fatalf("[SEC 4] expected Dispatcher to block hostile vector #%d", i)
		}
		blockedExpected++
	}

	// Plus one clean sovereign OpData frame (+1) that must be delivered cleanly.
	clean := common.MakeSovereignFrame(
		common.NodeBeta, common.NodeAlpha, data.Rung2Host, data.OpData, 900, []byte("CLEAN"),
	)
	wire := common.EncodeFixture(t, &clean)
	var env transport.WireEnvelope
	env.SenderID = common.NodeBeta
	env.TargetID = common.NodeAlpha
	env.Length = len(wire)
	copy(env.Bytes[:], wire)
	if out := disp.HandleInboundEnvelope(env); !out.Decoded || !out.Verdict.IsPermitted() {
		t.Fatalf("[SEC 4] clean sovereign frame was not delivered: %+v", out)
	}

	stats := disp.StatsSnapshot()
	if stats.AirlockBlocked != blockedExpected || stats.DataDelivered != 1 {
		t.Fatalf("[SEC 4] Dispatcher telemetry drift: want blocked=%d delivered=1, got %+v",
			blockedExpected, stats)
	}
	_ = time.Now()
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Preflight Diagnostic Gates & Physical Source Anchor Verification )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-4 System Verification Suite
// Module:        test.system.airlock_hostility
// Mission:       Assert Airlock, Interpreter & Egress preflights and anchors.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Block 4 Pre-Flight Contract Gates
// ------------------------------------------------------------------------------

func TestAirlockHostilitySystem_Block4_PreflightGates(t *testing.T) {
	if hostilitySuiteName == "" {
		t.Fatal("[BLOCK 4] suite identifier must not be empty")
	}
	common.AssertAuditPass(t, covenant.PreflightCovenantFilter(), "")
	common.AssertAuditPass(t, interpreter.PreflightTranslator(), "")
	common.AssertAuditPass(t, interpreter.PreflightWellServer(), "")
	common.AssertAuditPass(t, rung4_global.PreflightGlobalEgress(), "")
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Physical Source Anchor Verification
// ------------------------------------------------------------------------------

func TestAirlockHostilitySystem_Block4_SourceAnchors(t *testing.T) {
	common.AssertRepoAnchors(t,
		"engine/covenant/filter.go",
		"interpreter/translator.go",
		"interpreter/captive_web.go",
		"transport/rung4_global/egress.go",
	)
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Test Suite Epilogue & Clean Readiness Handshake )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-4 System Verification Suite
// Module:        test.system.airlock_hostility
// Mission:       Confirm Airlock & border subsystem readiness post-storm.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Epilogue Readiness Verification
// ------------------------------------------------------------------------------

func TestAirlockHostilitySystem_Block5_EpilogueReadiness(t *testing.T) {
	if !covenant.IsAirlockReady() || !interpreter.IsTranslatorReady() ||
		!interpreter.IsWellServerReady() || !rung4_global.IsGlobalEgressReady() ||
		!common.IsHarnessReady() {
		t.Fatal("[BLOCK 5: EPILOGUE] Airlock hostility system readiness failed")
	}
}

// End of tests/system/airlock_hostility_system_test.go
