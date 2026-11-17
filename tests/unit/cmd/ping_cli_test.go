// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/unit/cmd/ping_cli_test.go
// @id            test.unit.cmd.ping_cli
// @layer         Layer 4: Tier-2 Unit Verification Suite (CLI Sub-Stratum)
// @archetype     verification_suite | unit_isolation
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet `cmd/faithnet-ping` Isolated CLI Argument & Mode Suite
// @role          Tier-2 Unit Isolation Gate for cmd.faithnet.ping
// @maturity      Rev 5 Baseline | 4x4 Verification Matrix
// @chassis       bounded-heap | Exercises compiled faithnet-ping flag matrix
// @scale         O(1) deterministic CLI option & exit-code boundary tests
// @seams         in: cmd/faithnet-ping, engine, tests/common
// @invariants    1. Valid --mode=memory, --mode=ipc, --audit exit 0 cleanly.
//                2. Invalid mode, count<=0, count>10000, payload>472B exit 2.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Dedicated 1:1 Tier-2 Unit isolation suite for `cmd/faithnet-ping/main.go`,
//   mirroring `tools/tests/unit/cmd/args.rs`.
//
//   ### For Architects, Operators & Non-Programmers:
//   Verifies every command-line option of `faithnet-ping` accepts good inputs
//   and exits cleanly with a helpful error on invalid arguments.
//
//   ### For Systems Programmers & Engine Architects:
//   Compiles `cmd/faithnet-ping` once per test into `t.TempDir()` and probes
//   every nominal, fault-inversion, boundary, and diagnostic card path.
// ==============================================================================

package cmd_unit_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Identifier, Verification Imports & Ephemeral Binary Builder )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.cmd.ping_cli
// Mission:       Compile faithnet-ping into TempDir for isolated flag testing.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// --- Tier 2: External Library Dependencies (None) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---
import (
	"github.com/CreativeWorkzStudio/FaithNet/engine"
	"github.com/CreativeWorkzStudio/FaithNet/tests/common"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Operational Suite Constants & Identifiers
// ------------------------------------------------------------------------------

const (
	pingCLIUnitSuiteName = "Tier 2 Unit: cmd.faithnet.ping (CLI Flags & Modes)"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.2: Ephemeral Binary Helper
// ------------------------------------------------------------------------------

func buildUnitBin(t *testing.T, pkgRel, name string) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), name)
	cmd := exec.Command("go", "build", "-o", bin, pkgRel)
	cmd.Dir = common.ResolveRepoRoot(t)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build %s failed: %v\n%s", pkgRel, err, out)
	}
	return bin
}

func assertExitCode(t *testing.T, err error, wantCode int) {
	t.Helper()
	var ee *exec.ExitError
	if !errors.As(err, &ee) || ee.ExitCode() != wantCode {
		t.Fatalf("expected exit code %d, got err=%v", wantCode, err)
	}
}

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.cmd.ping_cli
// Mission:       Verify faithnet-ping modes, flag bounds, cards, and --audit.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: [ SECTION 1: ALWAYS PASS (NOMINAL GOLDEN SURFACE COGS) ]
// ------------------------------------------------------------------------------

func TestPingCLIUnit_Sec1_MemoryAndIPCNominalWithCards(t *testing.T) {
	bin := buildUnitBin(t, "./cmd/faithnet-ping", "faithnet-ping")

	outMem, err := exec.Command(bin, "--mode=memory", "--count=2", "--cards=true").CombinedOutput()
	if err != nil || !strings.Contains(string(outMem), "FAITHNET 5-COG PING ENGINE DIAGNOSTIC CARD") {
		t.Fatalf("[SEC 1] memory+cards failed: %v\n%s", err, outMem)
	}
	outIPC, err := exec.Command(bin, "--mode=ipc", "--count=2", "--cards=false").CombinedOutput()
	if err != nil || !strings.Contains(string(outIPC), "rung2:ipc") || !strings.Contains(string(outIPC), "0% loss") {
		t.Fatalf("[SEC 1] ipc run failed: %v\n%s", err, outIPC)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: [ SECTION 2: NEVER FAIL (FAULT INVERSION & DEGRADATION) ]
// ------------------------------------------------------------------------------

func TestPingCLIUnit_Sec2_RejectsBadModeAndOutOfRangeCounts(t *testing.T) {
	bin := buildUnitBin(t, "./cmd/faithnet-ping", "faithnet-ping")

	_, err1 := exec.Command(bin, "--mode=invalid").CombinedOutput()
	assertExitCode(t, err1, 2)

	_, err2 := exec.Command(bin, "--count=0").CombinedOutput()
	assertExitCode(t, err2, 2)

	_, err3 := exec.Command(bin, "--count=99999").CombinedOutput()
	assertExitCode(t, err3, 2)
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: [ SECTION 3: INVARIANTS, SATURATION & LIMIT GUARDS ]
// ------------------------------------------------------------------------------

func TestPingCLIUnit_Sec3_Exact472BytePayloadAcceptedAnd473Rejected(t *testing.T) {
	bin := buildUnitBin(t, "./cmd/faithnet-ping", "faithnet-ping")

	exact := strings.Repeat("P", engine.MaxPingEchoBytes)
	outOK, err := exec.Command(bin, "--mode=memory", "--count=1", "--cards=false", "--payload="+exact).CombinedOutput()
	if err != nil || !strings.Contains(string(outOK), "bytes=512") {
		t.Fatalf("[SEC 3] exact 472B echo (512B sector) failed: %v\n%s", err, outOK)
	}

	over := strings.Repeat("P", engine.MaxPingEchoBytes+1)
	_, errOver := exec.Command(bin, "--mode=memory", "--count=1", "--payload="+over).CombinedOutput()
	assertExitCode(t, errOver, 2)
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.4: [ SECTION 4: EMERGENCE & COMPOSITE WORKFLOWS ]
// ------------------------------------------------------------------------------

func TestPingCLIUnit_Sec4_AuditModeRunsAllNinePreflightGates(t *testing.T) {
	bin := buildUnitBin(t, "./cmd/faithnet-ping", "faithnet-ping")
	out, err := exec.Command(bin, "--audit").CombinedOutput()
	if err != nil || strings.Count(string(out), "[PASS]") < 9 {
		t.Fatalf("[SEC 4] --audit expected >=9 [PASS] gates, err=%v\n%s", err, out)
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Preflight Diagnostic Gates & Physical Source Anchor Verification )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.cmd.ping_cli
// Mission:       Run harness self-audit and verify cmd/faithnet-ping anchor.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Preflight Diagnostic Gates
// ------------------------------------------------------------------------------

func TestPingCLIUnit_Block4_PreflightGate(t *testing.T) {
	common.AssertAuditPass(t, common.SelfAudit(), common.ExecuteHarnessDumpCard())
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Suite Identifier & Physical Source Anchor Gate
// ------------------------------------------------------------------------------

func TestPingCLIUnit_Block4_SourceAnchor(t *testing.T) {
	if pingCLIUnitSuiteName == "" {
		t.Fatal("[BLOCK 4] suite identifier must not be empty")
	}
	common.AssertRepoAnchors(t, "cmd/faithnet-ping/main.go")
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Test Suite Epilogue & Clean Readiness Handshake )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.cmd.ping_cli
// Mission:       Confirm harness readiness after faithnet-ping CLI unit suite.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Epilogue & Clean State Handshake
// ------------------------------------------------------------------------------

func TestPingCLIUnit_Block5_EpilogueReadiness(t *testing.T) {
	if !common.IsHarnessReady() || !engine.IsPingEngineReady() {
		t.Fatal("[BLOCK 5: EPILOGUE] ping CLI unit readiness handshake failed")
	}
}

// End of tests/unit/cmd/ping_cli_test.go
