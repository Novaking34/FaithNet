// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/unit/cmd/node_cli_test.go
// @id            test.unit.cmd.node_cli
// @layer         Layer 4: Tier-2 Unit Verification Suite (CLI Sub-Stratum)
// @archetype     verification_suite | unit_isolation
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet `cmd/faithnet-node` Isolated CLI Flags & Audit Suite
// @role          Tier-2 Unit Isolation Gate for cmd.faithnet.node
// @maturity      Rev 5 Baseline | 4x4 Verification Matrix
// @chassis       bounded-heap | Exercises compiled faithnet-node flag matrix
// @scale         O(1) deterministic node CLI option & exit-code boundary tests
// @seams         in: cmd/faithnet-node, orchestrator, tests/common
// @invariants    1. `--audit` runs all 22 structured Block 4 gates & exits 0.
//                2. `--probe-only` renders all 6 diagnostic cards & exits 0.
//                3. Null/broadcast --id or >32B --beacon-ssid exits 2 cleanly.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Dedicated 1:1 Tier-2 Unit isolation suite for `cmd/faithnet-node/main.go`,
//   mirroring `tools/tests/unit/cmd/dispatch.rs`.
//
//   ### For Architects, Operators & Non-Programmers:
//   Verifies every startup flag of the sovereign node daemon cleanly validates
//   node identity, Wi-Fi beacon bounds, and full-stack pre-flight gates.
//
//   ### For Systems Programmers & Engine Architects:
//   Compiles `cmd/faithnet-node` into `t.TempDir()` and verifies exit codes
//   and stdout diagnostic cards across Sections 1-4.
// ==============================================================================

package cmd_unit_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Identifier, Verification Imports & Tier-2 Unit Sensor Anchor )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.cmd.node_cli
// Mission:       Bind node CLI test helper and orchestrator readiness probes.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"os/exec"
	"strings"
	"testing"
)

// --- Tier 2: External Library Dependencies (None) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---
import (
	"github.com/CreativeWorkzStudio/FaithNet/orchestrator"
	"github.com/CreativeWorkzStudio/FaithNet/tests/common"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Operational Suite Constants & Identifiers
// ------------------------------------------------------------------------------

const (
	nodeCLIUnitSuiteName = "Tier 2 Unit: cmd.faithnet.node (Node CLI & Audit)"
)

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.cmd.node_cli
// Mission:       Verify faithnet-node flags, bounds, probe cards & 22-gate audit.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: [ SECTION 1: ALWAYS PASS (NOMINAL GOLDEN SURFACE COGS) ]
// ------------------------------------------------------------------------------

func TestNodeCLIUnit_Sec1_ProbeOnlyWithBeaconAndWellCards(t *testing.T) {
	bin := buildUnitBin(t, "./cmd/faithnet-node", "faithnet-node")

	out, err := exec.Command(bin,
		"--probe-only", "--label=UnitAnchor", "--beacon", "--well",
	).CombinedOutput()
	s := string(out)
	if err != nil ||
		!strings.Contains(s, "FAITHNET HOST BODY HARDWARE PROBE CARD") ||
		!strings.Contains(s, "FAITHNET 4-RUNG ADAPTIVE LADDER CARD") ||
		!strings.Contains(s, "FAITHNET RUNG 3 ROOM BEACON SWITCH CARD") ||
		!strings.Contains(s, "FAITHNET THE WELL CAPTIVE PORTAL CARD") {
		t.Fatalf("[SEC 1] --probe-only output incomplete: err=%v\n%s", err, s)
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: [ SECTION 2: NEVER FAIL (FAULT INVERSION & DEGRADATION) ]
// ------------------------------------------------------------------------------

func TestNodeCLIUnit_Sec2_RejectsNullAndBroadcastNodeIDs(t *testing.T) {
	bin := buildUnitBin(t, "./cmd/faithnet-node", "faithnet-node")

	_, errNull := exec.Command(bin, "--id=0", "--probe-only").CombinedOutput()
	assertExitCode(t, errNull, 2)

	_, errBcast := exec.Command(bin, "--id=18446744073709551615", "--probe-only").CombinedOutput()
	assertExitCode(t, errBcast, 2)
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.3: [ SECTION 3: INVARIANTS, SATURATION & LIMIT GUARDS ]
// ------------------------------------------------------------------------------

func TestNodeCLIUnit_Sec3_Enforces32ByteBeaconSSIDCeiling(t *testing.T) {
	bin := buildUnitBin(t, "./cmd/faithnet-node", "faithnet-node")

	exact32 := strings.Repeat("S", 32)
	if out, err := exec.Command(bin, "--beacon", "--beacon-ssid="+exact32, "--probe-only").CombinedOutput(); err != nil {
		t.Fatalf("[SEC 3] exact 32B SSID failed: %v\n%s", err, out)
	}
	over33 := strings.Repeat("S", 33)
	_, errOver := exec.Command(bin, "--beacon", "--beacon-ssid="+over33, "--probe-only").CombinedOutput()
	assertExitCode(t, errOver, 2)
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.4: [ SECTION 4: EMERGENCE & COMPOSITE WORKFLOWS ]
// ------------------------------------------------------------------------------

func TestNodeCLIUnit_Sec4_FullStack22GateAuditPassesCleanly(t *testing.T) {
	bin := buildUnitBin(t, "./cmd/faithnet-node", "faithnet-node")
	out, err := exec.Command(bin, "--audit").CombinedOutput()
	if err != nil || strings.Count(string(out), "[PASS]") < 22 ||
		!strings.Contains(string(out), "cli_node_preflight_4section") {
		t.Fatalf("[SEC 4] faithnet-node --audit expected >=22 [PASS], err=%v\n%s", err, out)
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Preflight Diagnostic Gates & Physical Source Anchor Verification )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.cmd.node_cli
// Mission:       Run harness self-audit and verify cmd/faithnet-node anchor.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Preflight Diagnostic Gates
// ------------------------------------------------------------------------------

func TestNodeCLIUnit_Block4_PreflightGate(t *testing.T) {
	common.AssertAuditPass(t, common.SelfAudit(), common.ExecuteHarnessDumpCard())
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Suite Identifier & Physical Source Anchor Gate
// ------------------------------------------------------------------------------

func TestNodeCLIUnit_Block4_SourceAnchor(t *testing.T) {
	if nodeCLIUnitSuiteName == "" {
		t.Fatal("[BLOCK 4] suite identifier must not be empty")
	}
	common.AssertRepoAnchors(t, "cmd/faithnet-node/main.go")
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Test Suite Epilogue & Clean Readiness Handshake )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Tier-2 Unit Verification Suite
// Module:        test.unit.cmd.node_cli
// Mission:       Confirm orchestrator & harness readiness post-suite.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Epilogue & Clean State Handshake
// ------------------------------------------------------------------------------

func TestNodeCLIUnit_Block5_EpilogueReadiness(t *testing.T) {
	if !common.IsHarnessReady() || !orchestrator.IsDispatcherReady() {
		t.Fatal("[BLOCK 5: EPILOGUE] node CLI unit readiness handshake failed")
	}
}

// End of tests/unit/cmd/node_cli_test.go
