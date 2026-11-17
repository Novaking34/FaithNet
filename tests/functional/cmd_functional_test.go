// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/functional/cmd_functional_test.go
// @id            test.functional.cmd
// @layer         Layer 4: Tier-1 Functional Verification Suite
// @archetype     verification_suite | functional_surface
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet CLI Entrypoints Functional Surface Suite
// @role          Tier-1 Functional Diagnostic Sensor for cmd/faithnet-{ping,node}
// @maturity      Rev 5 Baseline | 4x4 Verification Matrix
// @chassis       bounded-heap | Exercises both CLI surfaces via `go run` / bins
// @scale         O(1) deterministic CLI surface, flag & exit-code probes
// @seams         in: cmd/faithnet-ping, cmd/faithnet-node, tests/common
// @invariants    1. Macro Tier 1 (Functional): mirrors `tools/tests/functional/cmd.rs`.
//                2. Both CLIs pass `--audit` cleanly with exit code 0.
//                3. Invalid CLI flags fail cleanly with non-zero exit codes.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Tier-1 Functional surface suite exercising FaithNet's two Layer 4 CLI
//   entrypoints (`cmd/faithnet-ping` and `cmd/faithnet-node`) across nominal
//   modes, invalid flag fault inversion, payload/SSID ceilings, and Block 4/5
//   preflight & anchor verification.
//
//   ### For Architects, Operators & Non-Programmers:
//   Verifies both command-line tools respond predictably to normal flags,
//   print clean status cards, and reject bad operator input safely.
//
//   ### For Systems Programmers & Engine Architects:
//   Paired with `tests/unit/cmd/{ping,node}_cli_test.go` and Tier 4
//   `tests/system/cli_binaries_system_test.go`.
// ==============================================================================

package functional_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Identifier, Verification Imports & Ephemeral CLI Builder )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-1 Functional Verification Suite
// Module:        test.functional.cmd
// Mission:       Compile ephemeral CLI binaries into TempDir for fast probes.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
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
	cmdFunctionalSuiteName = "Tier 1: Layer 4 CLI Entrypoints Surface"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.2: Ephemeral CLI Compilation Fixture
// ------------------------------------------------------------------------------

func buildCLI(t *testing.T, pkgRel string, binName string) string {
	t.Helper()
	root := common.ResolveRepoRoot(t)
	outPath := filepath.Join(t.TempDir(), binName)
	cmd := exec.Command("go", "build", "-o", outPath, pkgRel)
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed building %s: %v\n%s", pkgRel, err, out)
	}
	return outPath
}

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-1 Functional Verification Suite
// Module:        test.functional.cmd
// Mission:       Verify both CLIs across Sections 1-4 (nominal/fault/limit/e2e).
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: [ SECTION 1: ALWAYS PASS (NOMINAL CLI SURFACE INVOCATIONS) ]
// ------------------------------------------------------------------------------

func TestCmdFunctional_Sec1_PingAndNodeNominalSurface(t *testing.T) {
	pingBin := buildCLI(t, "./cmd/faithnet-ping", "faithnet-ping")
	nodeBin := buildCLI(t, "./cmd/faithnet-node", "faithnet-node")

	outMem, err := exec.Command(pingBin, "--mode=memory", "--count=2", "--cards=false").CombinedOutput()
	if err != nil || !strings.Contains(string(outMem), "0% loss") {
		t.Fatalf("[SEC 1] faithnet-ping memory failed: %v\n%s", err, outMem)
	}

	outNode, err := exec.Command(nodeBin, "--probe-only").CombinedOutput()
	if err != nil || !strings.Contains(string(outNode), "FAITHNET HOST BODY HARDWARE PROBE CARD") {
		t.Fatalf("[SEC 1] faithnet-node --probe-only failed: %v\n%s", err, outNode)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: [ SECTION 2: NEVER FAIL (INVALID CLI FLAG INVERSION) ]
// ------------------------------------------------------------------------------

func TestCmdFunctional_Sec2_RejectsInvalidModeAndNullNodeID(t *testing.T) {
	pingBin := buildCLI(t, "./cmd/faithnet-ping", "faithnet-ping")
	nodeBin := buildCLI(t, "./cmd/faithnet-node", "faithnet-node")

	if _, err := exec.Command(pingBin, "--mode=carrier-pigeon").CombinedOutput(); err == nil {
		t.Fatal("[SEC 2] invalid --mode must exit non-zero")
	}
	if _, err := exec.Command(nodeBin, "--id=0", "--probe-only").CombinedOutput(); err == nil {
		t.Fatal("[SEC 2] null --id=0 must exit non-zero")
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: [ SECTION 3: INVARIANTS, SATURATION & LIMIT GUARDS ]
// ------------------------------------------------------------------------------

func TestCmdFunctional_Sec3_EnforcesPayloadAndSSIDCeilings(t *testing.T) {
	pingBin := buildCLI(t, "./cmd/faithnet-ping", "faithnet-ping")
	nodeBin := buildCLI(t, "./cmd/faithnet-node", "faithnet-node")

	overEcho := strings.Repeat("X", engine.MaxPingEchoBytes+1)
	if _, err := exec.Command(pingBin, "--payload="+overEcho).CombinedOutput(); err == nil {
		t.Fatal("[SEC 3] 473B --payload bypassed 472B echo ceiling")
	}
	overSSID := strings.Repeat("A", 33)
	if _, err := exec.Command(nodeBin, "--beacon", "--beacon-ssid="+overSSID, "--probe-only").CombinedOutput(); err == nil {
		t.Fatal("[SEC 3] 33B --beacon-ssid bypassed IEEE 802.11 ceiling")
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.4: [ SECTION 4: EMERGENCE & COMPOSITE WORKFLOWS ]
// ------------------------------------------------------------------------------

func TestCmdFunctional_Sec4_BothCLIAuditModesEmitFullPassCards(t *testing.T) {
	pingBin := buildCLI(t, "./cmd/faithnet-ping", "faithnet-ping")
	nodeBin := buildCLI(t, "./cmd/faithnet-node", "faithnet-node")

	pingOut, err := exec.Command(pingBin, "--audit").CombinedOutput()
	if err != nil || !strings.Contains(string(pingOut), "cli_ping_preflight_4section") {
		t.Fatalf("[SEC 4] faithnet-ping --audit failed: %v\n%s", err, pingOut)
	}
	nodeOut, err := exec.Command(nodeBin, "--audit").CombinedOutput()
	if err != nil || strings.Count(string(nodeOut), "[PASS]") < 22 {
		t.Fatalf("[SEC 4] faithnet-node --audit expected >=22 [PASS], err=%v\n%s", err, nodeOut)
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Preflight Diagnostic Gates & Physical Source Anchor Verification )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-1 Functional Verification Suite
// Module:        test.functional.cmd
// Mission:       Verify shared harness self-audit and physical CLI anchors.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Preflight Diagnostic Gates
// ------------------------------------------------------------------------------

func TestCmdFunctional_Block4_PreflightHarness(t *testing.T) {
	common.AssertAuditPass(t, common.SelfAudit(), common.ExecuteHarnessDumpCard())
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Suite Identifier & Physical Source Anchor Gate
// ------------------------------------------------------------------------------

func TestCmdFunctional_Block4_SourceAnchors(t *testing.T) {
	if cmdFunctionalSuiteName == "" {
		t.Fatal("[BLOCK 4] suite identifier must not be empty")
	}
	common.AssertRepoAnchors(t,
		"cmd/faithnet-ping/main.go",
		"cmd/faithnet-node/main.go",
	)
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Test Suite Epilogue & Clean Readiness Handshake )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-1 Functional Verification Suite
// Module:        test.functional.cmd
// Mission:       Confirm harness readiness after CLI functional surface sweep.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Epilogue & Clean State Handshake
// ------------------------------------------------------------------------------

func TestCmdFunctional_Block5_EpilogueReadiness(t *testing.T) {
	if !common.IsHarnessReady() || !engine.IsCodecReady() {
		t.Fatal("[BLOCK 5: EPILOGUE] CLI functional readiness handshake failed")
	}
}

// End of tests/functional/cmd_functional_test.go
