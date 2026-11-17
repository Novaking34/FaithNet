// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/common/fixtures.go
// @id            test.common.fixtures
// @layer         Layer 4: Shared Test Fixtures & Diagnostic Sensor Helpers
// @archetype     operational | stateless
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet 4x4 Verification Matrix Shared Fixtures
// @role          Shared Node IDs, Sample Frame Builders & Assertion Helpers
// @maturity      Rev 5 Baseline | Wave 1 Verification
// @chassis       bounded-heap | Reusable test harness builders (Layer 4 only)
// @scale         O(1) deterministic test fixture construction
// @seams         in: data, engine | exports: canonical test IDs & builders
// @invariants    1. Stratum Baseline: Universal 5-Block Lifecycle Protocol.
//                2. Deterministic fixtures: zero wall-clock or random drift.
//                3. Mirrors engine/tests/common/cornerstone_test.h pattern.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Provides shared sovereign NodeIDs, deterministic sample Frame builders,
//   and reusable assertion utilities consumed across all 4 Macro Tiers
//   (`functional/`, `unit/`, `integrated/`, `system/`).
//
//   ### For Architects, Operators & Non-Programmers:
//   Supplies consistent test characters (NodeAlpha, NodeBeta, NodeRelay,
//   NodeExternal) so every test tier speaks the same clear vocabulary.
//
//   ### For Systems Programmers & Engine Architects:
//   Eliminates boilerplate duplication across the 4x4 verification matrix while
//   keeping every fixture 100% deterministic and race-detector clean.
// ==============================================================================

package common

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Canonical Sovereign Test Identities & Hardware Body Profiles )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Shared Test Fixtures
// Module:        test.common.fixtures
// Mission:       Pin canonical sovereign NodeIDs and synthetic body profiles.
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
	"github.com/CreativeWorkzStudio/FaithNet/engine"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Canonical Sovereign Test Identities
// ------------------------------------------------------------------------------

const (
	NodeAlpha    data.NodeID = 0xFA17000000000001
	NodeBeta     data.NodeID = 0xFA17000000000002
	NodeRelay    data.NodeID = 0xFA17000000000003
	NodeExternal data.NodeID = 0xEEEE000000000099
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.2: Operational Logical Anchor
// ------------------------------------------------------------------------------

const (
	ModuleIDTestFixtures = "test.common.fixtures"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.3: Synthetic Hardware Body Profiles
// ------------------------------------------------------------------------------
// Deterministic host profiles simulating rich, offline, and crippled bodies.
// ------------------------------------------------------------------------------

// RichHostProfile simulates a modern laptop with all 4 rungs available.
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
// ( Reusable Diagnostic Sensor Assertions )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Shared Test Fixtures
// Module:        test.common.fixtures
// Mission:       Provide clear, non-flaky frame equality and round-trip helpers.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: Frame Equality & Round-Trip Assertions
// ------------------------------------------------------------------------------

// AssertFramesEqual verifies every header field and active payload byte matches.
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

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Fixture Integrity Self-Check )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Shared Test Fixtures
// Module:        test.common.fixtures
// Mission:       Guarantee test node IDs and synthetic profiles remain valid.
// ==============================================================================

// AuditFixtures verifies all canonical test identities and body profiles.
func AuditFixtures() bool {
	if !NodeAlpha.IsValid() || !NodeBeta.IsValid() || !NodeRelay.IsValid() {
		return false
	}
	return RichHostProfile().HighestViableRung() == data.Rung4Global &&
		OfflineRoomProfile().HighestViableRung() == data.Rung3Local &&
		EmergencyFloorProfile().HighestViableRung() == data.Rung1Floor
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK / STATUTORY EXEMPTION & LIFECYCLE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Shared Test Fixtures
// Module:        test.common.fixtures
// Mission:       Shared Layer 4 verification fixture library.
// Stateless Exemption Notice: None. (Stateless fixture module).
// ==============================================================================

// End of tests/common/fixtures.go
