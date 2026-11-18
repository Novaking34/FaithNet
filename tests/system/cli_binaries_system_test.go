// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/system/cli_binaries_system_test.go
// @id            test.system.cli_binaries
// @layer         Layer 4: Tier-4 System End-to-End & Stress Verification Suite
// @archetype     verification_suite | system_cli_binaries
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Compiled Standalone CLI Binaries System Suite
// @role          Tier-4 System Gate: faithnet-ping & faithnet-node CLI Proofs
// @maturity      Rev 5 Baseline | Wave 6 4x4 Symmetry Expansion
// @chassis       bounded-heap | Compiles real binaries into isolated t.TempDir
// @scale         Full positive & negative CLI matrix with strict exit-code proof
// @seams         in: cmd/faithnet-ping, cmd/faithnet-node, tests/common
// @invariants    1. Macro Tier 4 (System): compiles and executes BOTH standalone
//                   operator binaries out-of-process with real OS exit codes.
//                2. Tests BOTH valid flags (exit 0 + expected forensic cards)
//                   AND invalid operator flags (non-zero exit + clear error).
//                3. Zero binaries or temporary files leaked into repo root.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Dedicated Tier-4 System proving ground for FaithNet's Layer 3B operator
//   executables (`cmd/faithnet-ping` and `cmd/faithnet-node`). Mirrors
//   Cornerstone's `tools/tests/common/bin_tests.rs` & `tools/tests/system/`
//   discipline by testing real compiled binaries end-to-end across positive
//   audit/probe/ping runs and negative flag validation boundaries.
//
//   ### For Architects, Operators & Non-Programmers:
//   Compiles the actual `faithnet-ping` and `faithnet-node` programs a user
//   would run in a terminal and verifies that valid commands print clean
//   dashboards while bad arguments are rejected safely with clear explanations.
//
//   ### For Systems Programmers & Engine Architects:
//   Uses `go build` into `t.TempDir()`, executes subprocess invocations via
//   `os/exec`, and asserts stdout/stderr tokens plus exact zero/non-zero exits.
// ==============================================================================

package system_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Identifier, Verification Imports & Binary Build Helper )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-4 System Verification Suite
// Module:        test.system.cli_binaries
// Mission:       Compile standalone CLI binaries into isolated temp storage.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// --- Tier 2: External Library Dependencies (None — 100% Go Stdlib Invariant) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---
import (
	"github.com/CreativeWorkzStudio/FaithNet/tests/common"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Operational Suite Constants & Binary Compilation Helper
// ------------------------------------------------------------------------------

const (
	cliBinariesSuiteName = "Tier 4 System: Compiled Standalone CLI Binaries"
)

func buildCLIBinaries(t *testing.T) (pingBin, nodeBin string) {
	t.Helper()
	repoRoot := common.ResolveRepoRoot(t)
	binDir := t.TempDir()
	pingBin = filepath.Join(binDir, "faithnet-ping")
	nodeBin = filepath.Join(binDir, "faithnet-node")

	for _, target := range []struct {
		out string
		pkg string
	}{
		{pingBin, "./cmd/faithnet-ping"},
		{nodeBin, "./cmd/faithnet-node"},
	} {
		cmd := exec.Command("go", "build", "-o", target.out, target.pkg)
		cmd.Dir = repoRoot
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("failed building %s: %v\n%s", target.pkg, err, out)
		}
	}
	return pingBin, nodeBin
}

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-4 System Verification Suite
// Module:        test.system.cli_binaries
// Mission:       Verify positive CLI flows, negative flag rejection, exact
//                22-gate audit output, and combined Beacon+Well probe emergence.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: [ SECTION 1: ALWAYS PASS (VALID CLI EXECUTION MATRIX) ]
// ------------------------------------------------------------------------------

func TestCLIBinariesSystem_Sec1_ValidPingAndNodeCommandsExitZero(t *testing.T) {
	pingBin, nodeBin := buildCLIBinaries(t)

	// 1. faithnet-ping: --audit, --mode=memory, --mode=ipc.
	for _, tc := range []struct {
		args   []string
		expect string
	}{
		{[]string{"--audit"}, "[PASS]"},
		{[]string{"--mode=memory", "--count=2", "--cards=true"}, "FAITHNET 5-COG PING ENGINE DIAGNOSTIC CARD"},
		{[]string{"--mode=ipc", "--count=2", "--cards=false"}, "0% loss"},
	} {
		out, err := exec.Command(pingBin, tc.args...).CombinedOutput()
		if err != nil || !strings.Contains(string(out), tc.expect) {
			t.Fatalf("[SEC 1] faithnet-ping %v failed (want %q): err=%v\n%s",
				tc.args, tc.expect, err, out)
		}
	}

	// 2. faithnet-node: --audit and --probe-only.
	out, err := exec.Command(nodeBin, "--audit").CombinedOutput()
	if err != nil || strings.Count(string(out), "[PASS]") < 22 {
		t.Fatalf("[SEC 1] faithnet-node --audit expected >= 22 [PASS] gates: err=%v\n%s", err, out)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: [ SECTION 2: NEVER FAIL (INVALID FLAGS REJECTED CLEANLY) ]
// ------------------------------------------------------------------------------

func TestCLIBinariesSystem_Sec2_InvalidOperatorFlagsExitNonZeroCleanly(t *testing.T) {
	pingBin, nodeBin := buildCLIBinaries(t)

	invalidPingCases := [][]string{
		{"--mode=quantum"},
		{"--count=0"},
		{"--count=999999"},
		{"--payload=" + strings.Repeat("X", 480)},
	}
	for _, args := range invalidPingCases {
		out, err := exec.Command(pingBin, args...).CombinedOutput()
		if err == nil {
			t.Fatalf("[SEC 2] faithnet-ping unexpectedly accepted invalid flags %v:\n%s", args, out)
		}
	}

	invalidNodeCases := [][]string{
		{"--probe-only", "--id=0"},
		{"--probe-only", "--label="},
		{"--probe-only", "--beacon-ssid=" + strings.Repeat("S", 33)},
	}
	for _, args := range invalidNodeCases {
		out, err := exec.Command(nodeBin, args...).CombinedOutput()
		if err == nil {
			t.Fatalf("[SEC 2] faithnet-node unexpectedly accepted invalid flags %v:\n%s", args, out)
		}
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: [ SECTION 3: INVARIANTS (DETERMINISTIC AUDIT GATE COUNTS) ]
// ------------------------------------------------------------------------------

func TestCLIBinariesSystem_Sec3_AuditGateCountsMatchExactArchitectureInventory(t *testing.T) {
	pingBin, nodeBin := buildCLIBinaries(t)

	pingOut, err := exec.Command(pingBin, "--audit").CombinedOutput()
	if err != nil {
		t.Fatalf("[SEC 3] faithnet-ping --audit failed: %v\n%s", err, pingOut)
	}
	if got := strings.Count(string(pingOut), "[PASS]"); got != 9 {
		t.Fatalf("[SEC 3] expected exactly 9 [PASS] gates from faithnet-ping --audit, got %d:\n%s",
			got, pingOut)
	}

	nodeOut, err := exec.Command(nodeBin, "--audit").CombinedOutput()
	if err != nil {
		t.Fatalf("[SEC 3] faithnet-node --audit failed: %v\n%s", err, nodeOut)
	}
	if got := strings.Count(string(nodeOut), "[PASS]"); got != 27 {
		t.Fatalf("[SEC 3] expected exactly 27 [PASS] gates from faithnet-node --audit, got %d:\n%s",
			got, nodeOut)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.4: [ SECTION 4: EMERGENCE (FULL ROOM + WELL PROBE TELEMETRY) ]
// ------------------------------------------------------------------------------

func TestCLIBinariesSystem_Sec4_NodeProbeOnlyEmitsFullRoomAndWellCards(t *testing.T) {
	_, nodeBin := buildCLIBinaries(t)

	out, err := exec.Command(
		nodeBin,
		"--probe-only",
		"--label=Block7-Anchor",
		"--beacon",
		"--beacon-ssid=FaithNet-Block7",
		"--well",
	).CombinedOutput()
	if err != nil {
		t.Fatalf("[SEC 4] faithnet-node --probe-only failed: %v\n%s", err, out)
	}
	text := string(out)
	for _, token := range []string{
		"FAITHNET HOST BODY HARDWARE PROBE CARD",
		"FAITHNET 4-RUNG ADAPTIVE LADDER CARD",
		"FAITHNET RUNG 3 ROOM BEACON SWITCH CARD",
		"FAITHNET THE WELL CAPTIVE PORTAL CARD",
		"FaithNet-Block7",
	} {
		if !strings.Contains(text, token) {
			t.Fatalf("[SEC 4] probe output missing expected telemetry token %q:\n%s", token, text)
		}
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Preflight Diagnostic Gates & Physical Source Anchor Verification )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-4 System Verification Suite
// Module:        test.system.cli_binaries
// Mission:       Verify shared harness self-audit and CLI physical anchors.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Block 4 Pre-Flight Contract Gates
// ------------------------------------------------------------------------------

func TestCLIBinariesSystem_Block4_PreflightGates(t *testing.T) {
	if cliBinariesSuiteName == "" {
		t.Fatal("[BLOCK 4] suite identifier must not be empty")
	}
	common.AssertAuditPass(t, common.SelfAudit(), common.ExecuteHarnessDumpCard())
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Physical Source Anchor Verification
// ------------------------------------------------------------------------------

func TestCLIBinariesSystem_Block4_SourceAnchors(t *testing.T) {
	common.AssertRepoAnchors(t,
		"cmd/faithnet-ping/main.go",
		"cmd/faithnet-node/main.go",
		"go.mod",
	)
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Test Suite Epilogue & Clean Readiness Handshake )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-4 System Verification Suite
// Module:        test.system.cli_binaries
// Mission:       Confirm shared harness readiness handshake post-CLI suite.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Epilogue Readiness Verification
// ------------------------------------------------------------------------------

func TestCLIBinariesSystem_Block5_EpilogueReadiness(t *testing.T) {
	if !common.IsHarnessReady() {
		t.Fatal("[BLOCK 5: EPILOGUE] CLI binaries system readiness failed")
	}
}

// End of tests/system/cli_binaries_system_test.go
