// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/common/fixtures.go
// @id            test.common.harness
// @layer         Layer 4: Shared Verification Harness & Diagnostic Fixtures
// @archetype     shared_harness | stateless
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet 4x4 Verification Matrix Shared Harness
// @role          Shared Fixtures, Forensic Assertions, Repo Root & Self-Audit
// @maturity      Rev 5 Baseline | Wave 1.5 Audit & Coverage Elevation
// @chassis       bounded-heap | Reusable Layer 4 test scaffolding only
// @scale         O(1) deterministic fixture construction & assertion helpers
// @seams         in: data, engine | exports: canonical IDs, assertions, probes
// @invariants    1. Stratum Baseline: Universal 5-Block Lifecycle Protocol.
//                2. Mirrors Cornerstone `tests/common/cornerstone_test.h` and
//                   `tools/tests/common/mod.rs` forensic failure reporting.
//                3. Deterministic fixtures: zero wall-clock or random drift.
//                4. Zero transient disk leaks across all 4 macro test tiers.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Provides shared sovereign NodeIDs, synthetic HardwareProfiles, sample
//   Frame builders, forensic `AuditResult` assertions (printing 60-column
//   diagnostic dump cards on failure), repository root discovery, and Block 4
//   self-audit probes consumed across `functional/`, `unit/`, `integrated/`,
//   and `system/` verification tiers.
//
//   ### For Architects, Operators & Non-Programmers:
//   Supplies consistent test characters (NodeAlpha, NodeBeta, NodeRelay,
//   NodeExternal) and ensures that whenever any automated test fails, it prints
//   a clear diagnostic card explaining exactly what tripped and why.
//
//   ### For Systems Programmers & Engine Architects:
//   - Ownership & Reentrancy: Stateless helper functions safe for `t.Parallel()`
//     and `-race` execution across all Layer 4 packages.
//   - Extension Seam: Add cross-tier assertion helpers or synthetic profiles in
//     Sub-Assembly 2.3/3.1 without duplicating boilerplate across test packages.
// ==============================================================================

package common

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Suite Constants, Sovereign Test Identities & Hardware Body Profiles )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Shared Verification Harness
// Module:        test.common.harness
// Mission:       Pin canonical sovereign NodeIDs, synthetic host body profiles,
//                and deterministic sample Frame builders for all 4 test tiers.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// --- Tier 2: External Library Dependencies (None — 100% Go Stdlib Invariant) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---
import (
	"github.com/CreativeWorkzStudio/FaithNet/data"
	"github.com/CreativeWorkzStudio/FaithNet/engine"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Operational Harness Constants & Canonical Sovereign IDs
// ------------------------------------------------------------------------------

const (
	// CommonHarnessVersion pins the shared Layer 4 harness revision.
	CommonHarnessVersion = "0.1.0-dev"

	// CommonHarnessSuiteName is the canonical shared harness identifier.
	CommonHarnessSuiteName = "FaithNet Shared 4x4 Verification Harness"

	// NodeAlpha is the primary sovereign originator identity in test suites.
	NodeAlpha data.NodeID = 0xFA17000000000001

	// NodeBeta is the primary sovereign responder identity in test suites.
	NodeBeta data.NodeID = 0xFA17000000000002

	// NodeRelay is the multi-hop ("Across the Street") relay identity.
	NodeRelay data.NodeID = 0xFA17000000000003

	// NodeExternal represents an untrusted commercial web origin identity.
	NodeExternal data.NodeID = 0xEEEE000000000099
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.2: Operational Logical Anchor
// ------------------------------------------------------------------------------

const (
	// ModuleIDTestFixtures is the canonical logical ID for this shared harness.
	ModuleIDTestFixtures = "test.common.harness"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.3: Synthetic Hardware Body Profiles
// ------------------------------------------------------------------------------
// Deterministic host profiles simulating rich, offline, and crippled bodies.
// ------------------------------------------------------------------------------

// RichHostProfile simulates a modern laptop with all 4 rungs available.
//
// ### For Architects, Operators & Non-Programmers:
// Represents a full-featured computer with Wi-Fi, Ethernet, sound, and serial.
//
// ### For Systems Programmers & Engine Architects:
// Deterministic `data.HardwareProfile` whose `HighestViableRung() == Rung4Global`.
func RichHostProfile() data.HardwareProfile {
	return data.HardwareProfile{
		OSName:          "linux",
		Arch:            "amd64",
		CPUCores:        8,
		HasMemoryBus:    true,
		HasLocalIPC:     true,
		HasLocalUDP:     true,
		HasWiFiHardware: true,
		HasSerialPorts:  true,
		HasAudioDevice:  true,
		HasGlobalTCP:    true,
		InterfaceCount:  3,
	}
}

// OfflineRoomProfile simulates an air-gapped laptop in a room (Rung 1..3).
//
// ### For Architects, Operators & Non-Programmers:
// Represents a laptop with local Wi-Fi/LAN but zero outside internet access.
//
// ### For Systems Programmers & Engine Architects:
// Deterministic `data.HardwareProfile` whose `HighestViableRung() == Rung3Local`.
func OfflineRoomProfile() data.HardwareProfile {
	return data.HardwareProfile{
		OSName:          "linux",
		Arch:            "amd64",
		CPUCores:        4,
		HasMemoryBus:    true,
		HasLocalIPC:     true,
		HasLocalUDP:     true,
		HasWiFiHardware: true,
		HasSerialPorts:  true,
		HasAudioDevice:  true,
		HasGlobalTCP:    false,
		InterfaceCount:  1,
	}
}

// EmergencyFloorProfile simulates a degraded host with only Rung 1 floor.
//
// ### For Architects, Operators & Non-Programmers:
// Represents bare survival hardware with only copper serial or audio tones.
//
// ### For Systems Programmers & Engine Architects:
// Deterministic `data.HardwareProfile` whose `HighestViableRung() == Rung1Floor`.
func EmergencyFloorProfile() data.HardwareProfile {
	return data.HardwareProfile{
		OSName:         "bare-metal",
		Arch:           "x86",
		CPUCores:       1,
		HasMemoryBus:   false,
		HasLocalIPC:    false,
		HasLocalUDP:    false,
		HasSerialPorts: true,
		HasAudioDevice: true,
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 2.4: O(1) Deterministic Sample Frame Builders
// ------------------------------------------------------------------------------

// MakeSovereignFrame populates a valid sovereign Frame on rung with payload.
//
// ### For Architects, Operators & Non-Programmers:
// Builds a clean, properly sealed FaithNet message ready for any test rung.
//
// ### For Systems Programmers & Engine Architects:
// Returns value-typed `data.Frame` with `FlagFloorCompact | FlagCovenantVerified`
// and `QuadPositive` stamped cleanly.
func MakeSovereignFrame(
	src, dst data.NodeID,
	rung data.RungLevel,
	op data.Opcode,
	seq uint32,
	payload []byte,
) data.Frame {
	var f data.Frame
	f.Version = data.ProtocolVersion
	f.Rung = rung
	f.Opcode = op
	f.Flags = data.FlagFloorCompact | data.FlagCovenantVerified
	f.Truth = data.QuadPositive
	f.SourceID = src
	f.TargetID = dst
	f.SeqID = seq
	f.SetPayload(payload)
	return f
}

// ==============================================================================
// BLOCK 3: BODY BLOCK
// ( Reusable Forensic Assertions, Encoding Helpers & Repo Anchor Discovery )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Shared Verification Harness
// Module:        test.common.harness
// Mission:       Provide clear, forensic frame/audit assertions and repository
//                anchor resolution shared across all 4 verification tiers.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: Forensic AuditResult & Frame Equality Assertions
// ------------------------------------------------------------------------------

// AssertAuditPass asserts that a Block 4 `data.AuditResult` passed cleanly;
// on failure it emits the structured verdict plus an optional 60-col dump card.
//
// ### For Architects, Operators & Non-Programmers:
// Verifies a safety check passed and prints its full forensic dashboard card
// immediately if anything tripped.
//
// ### For Systems Programmers & Engine Architects:
// Mirrors Cornerstone's `TEST_ASSERT` + debugger forensic card dump pattern.
func AssertAuditPass(t *testing.T, res data.AuditResult, dumpCard string) {
	t.Helper()
	if !res.IsPass() {
		if dumpCard != "" {
			t.Fatalf("Block 4 audit gate failed: %s\n%s", res.String(), dumpCard)
		}
		t.Fatalf("Block 4 audit gate failed: %s", res.String())
	}
}

// AssertAuditFail asserts that a Block 4 `data.AuditResult` failed cleanly
// with at least the expected severity (fault-inversion verification).
//
// ### For Architects, Operators & Non-Programmers:
// Confirms a safety check caught a deliberate fault rather than ignoring it.
//
// ### For Systems Programmers & Engine Architects:
// Enforces `!res.IsPass()` and `res.Severity >= minSev` for Section 2 tests.
func AssertAuditFail(t *testing.T, res data.AuditResult, minSev data.AuditSeverity) {
	t.Helper()
	if res.IsPass() {
		t.Fatalf("expected Block 4 audit failure (>= %s), got PASS: %s",
			minSev.String(), res.String())
	}
	if res.Severity < minSev {
		t.Fatalf("expected severity >= %s, got %s: %s",
			minSev.String(), res.Severity.String(), res.String())
	}
}

// AssertFramesEqual verifies every header field and active payload byte matches.
//
// ### For Architects, Operators & Non-Programmers:
// Confirms an unpacked message matches the original sent envelope bit-for-bit.
//
// ### For Systems Programmers & Engine Architects:
// Compares all 9 scalar header fields plus active `PayloadSlice()` equality.
func AssertFramesEqual(t *testing.T, want, got *data.Frame) {
	t.Helper()
	if got.Version != want.Version ||
		got.Rung != want.Rung ||
		got.Opcode != want.Opcode ||
		got.Flags != want.Flags ||
		got.Truth != want.Truth ||
		got.SourceID != want.SourceID ||
		got.TargetID != want.TargetID ||
		got.SeqID != want.SeqID ||
		got.PayloadLen != want.PayloadLen ||
		!bytes.Equal(got.PayloadSlice(), want.PayloadSlice()) {
		t.Fatalf("frame mismatch:\n  want=%+v (%q)\n  got =%+v (%q)",
			want, want.PayloadSlice(), got, got.PayloadSlice())
	}
}

// EncodeFixture encodes f into a fresh 512-byte floor slice or fails the test.
//
// ### For Architects, Operators & Non-Programmers:
// Packs a test envelope into raw wire bytes so transport tests can transmit it.
//
// ### For Systems Programmers & Engine Architects:
// Invokes `engine.EncodeInto` into a stack buffer and returns an isolated copy.
func EncodeFixture(t *testing.T, f *data.Frame) []byte {
	t.Helper()
	var wire [data.FloorMaxFrameSize]byte
	n, err := engine.EncodeInto(wire[:], f)
	if err != nil {
		t.Fatalf("EncodeFixture failed: %v", err)
	}
	out := make([]byte, n)
	copy(out, wire[:n])
	return out
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: Repository Root & Physical Source Anchor Discovery
// ------------------------------------------------------------------------------

// ResolveRepoRoot walks upward from the current working directory until it
// finds FaithNet's `go.mod` anchor, returning its absolute directory path.
//
// ### For Architects, Operators & Non-Programmers:
// Locates the FaithNet project root reliably no matter which test folder runs.
//
// ### For Systems Programmers & Engine Architects:
// Mirrors `tools/tests/common/mod.rs::resolve_repo_root()` for Block 4.2
// physical source anchor verification gates.
func ResolveRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("ResolveRepoRoot failed os.Getwd: %v", err)
	}
	for i := 0; i < 6; i++ {
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatalf("ResolveRepoRoot could not locate FaithNet go.mod anchor")
	return ""
}

// AssertRepoAnchors verifies that every relative file path exists under root.
//
// ### For Architects, Operators & Non-Programmers:
// Confirms expected source files are physically present before/after tests.
//
// ### For Systems Programmers & Engine Architects:
// Implements Block 4.2 repository anchor verification gates across test tiers.
func AssertRepoAnchors(t *testing.T, relPaths ...string) {
	t.Helper()
	root := ResolveRepoRoot(t)
	for _, rel := range relPaths {
		full := filepath.Join(root, rel)
		info, err := os.Stat(full)
		if err != nil || info.IsDir() {
			t.Fatalf("[BLOCK 4: PREFLIGHT] Required source anchor missing: %s", rel)
		}
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Shared Harness Preflight & Fixture Self-Audit Gate )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Shared Verification Harness
// Module:        test.common.harness
// Mission:       Guarantee canonical test identities, synthetic profiles, and
//                sample frame builders pass self-audit prior to suite runs.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Harness Self-Audit Probe
// ------------------------------------------------------------------------------

// SelfAudit verifies all canonical test identities, body profiles, and sample
// frame builders, returning a structured `data.AuditResult`.
//
// ### For Architects, Operators & Non-Programmers:
// Checks the shared test toolbox itself so test failures always point to real
// code issues rather than broken test fixtures.
//
// ### For Systems Programmers & Engine Architects:
// Mirrors `tools/tests/common/mod.rs::self_audit()` with structured verdicts.
func SelfAudit() data.AuditResult {
	const check = "common_harness_self_audit"

	if !NodeAlpha.IsValid() || !NodeBeta.IsValid() ||
		!NodeRelay.IsValid() || !NodeExternal.IsValid() {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDTestFixtures, check,
			"One or more canonical sovereign test NodeIDs is invalid.")
	}
	if NodeAlpha == NodeBeta || NodeBeta == NodeRelay || NodeAlpha == NodeExternal {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDTestFixtures, check,
			"Canonical sovereign test NodeIDs must be pairwise distinct.")
	}
	if RichHostProfile().HighestViableRung() != data.Rung4Global ||
		OfflineRoomProfile().HighestViableRung() != data.Rung3Local ||
		EmergencyFloorProfile().HighestViableRung() != data.Rung1Floor {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDTestFixtures, check,
			"Synthetic HardwareProfile ladder ceiling invariants violated.")
	}
	sample := MakeSovereignFrame(
		NodeAlpha, NodeBeta, data.Rung1Floor, data.OpPing, 1, []byte("AUDIT"),
	)
	if v := engine.ValidateFrameContract(&sample); !v.IsPass() {
		return v
	}
	return data.AuditPass(ModuleIDTestFixtures, check,
		"Shared 4x4 verification harness identities, profiles, and frames verified.")
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Boolean Convenience Audit Wrapper
// ------------------------------------------------------------------------------

// AuditFixtures returns true iff SelfAudit() completes with AuditSeverityPass.
//
// ### For Architects, Operators & Non-Programmers:
// Quick pass/fail shortcut asking whether the shared test fixtures are valid.
//
// ### For Systems Programmers & Engine Architects:
// Convenience boolean bridge delegating to `SelfAudit().IsPass()`.
func AuditFixtures() bool {
	return SelfAudit().IsPass()
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK / EXECUTABLE PROCEDURES & READINESS HANDSHAKE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Shared Verification Harness
// Module:        test.common.harness
// Mission:       Expose epilogue readiness handshake and diagnostic card for
//                the shared Layer 4 verification harness.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Epilogue Readiness Probe & Diagnostic Card
// ------------------------------------------------------------------------------

// IsHarnessReady reports whether the shared verification harness is ready.
//
// ### For Architects, Operators & Non-Programmers:
// Epilogue confirmation that the shared test harness remained clean and healthy.
//
// ### For Systems Programmers & Engine Architects:
// Matches `tools/tests/common/mod.rs::is_common_harness_ready()`.
func IsHarnessReady() bool {
	return SelfAudit().IsPass()
}

// ExecuteHarnessDumpCard renders a 60-column inner-width harness summary card.
//
// ### For Architects, Operators & Non-Programmers:
// Displays the shared test harness version, canonical identities, and status.
//
// ### For Systems Programmers & Engine Architects:
// Non-mutating diagnostic card builder invoking `SelfAudit()`.
func ExecuteHarnessDumpCard() string {
	sa := SelfAudit()
	return fmt.Sprintf(
		"+------------------------------------------------------------+\n"+
			"| %-58s |\n"+
			"+------------------------------------------------------------+\n"+
			"| Module:    %-47s |\n"+
			"| Version:   %-47s |\n"+
			"| Alpha ID:  %-47s |\n"+
			"| Beta ID:   %-47s |\n"+
			"| SelfAudit: %-47s |\n"+
			"+------------------------------------------------------------+",
		"FAITHNET SHARED 4x4 TEST HARNESS CARD",
		ModuleIDTestFixtures,
		CommonHarnessVersion,
		NodeAlpha.String(),
		NodeBeta.String(),
		sa.Severity.String(),
	)
}

// ------------------------------------------------------------------------------
// Sub-Assembly 5.2: Subsystem Lifecycle Initialization & Teardown
// ------------------------------------------------------------------------------
// Stateless Exemption Notice: None. (Stateless fixture module).
// ------------------------------------------------------------------------------

// End of tests/common/fixtures.go
