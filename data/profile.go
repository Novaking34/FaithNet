// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          data/profile.go
// @id            cog.data.profile
// @layer         Layer 0: Pure Invariant Data Substrate
// @archetype     pure_types | stateless
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Data-Driven Host Profile & Telemetry Vocabulary
// @role          Pure Contract & Vocabulary: HardwareProfile, Manifest & Health
// @maturity      Rev 5 Baseline | Wave 1 Seed
// @chassis       zero-heap | Zero stateful globals; caller-owned value records
// @scale         O(1) pure profile queries & bounded telemetry records
// @seams         in: data/types.go, data/opcodes.go | exports: profile structs
// @invariants    1. Stratum Baseline: ALLOC_ZERO_HEAP, Downward Dependency Law,
//                   Universal 5-Block Lifecycle Protocol (Rev 5).
//                2. Zero OS or network imports (Downward Dependency Law).
//                3. DefaultNodeManifest enforces offline-first sovereignty
//                   (`AllowGlobalEgress == false`, floor = Rung1Floor).
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Defines the pure data-driven schemas consumed by the Orchestrator's
//   BodyProbe, AdaptiveLadder, and Dispatcher. Rather than branching on OS
//   names throughout the engine, FaithNet measures concrete body capabilities
//   into HardwareProfile and drives ladder decisions purely from data.
//
//   ### For Architects, Operators & Non-Programmers:
//   Acts as a medical chart for whatever computer FaithNet wakes up on—noting
//   whether it has Wi-Fi, Ethernet, serial ports, audio hardware, or only RAM,
//   and keeping a clear flight log whenever the network climbs up or falls down.
//
//   ### For Systems Programmers & Engine Architects:
//   - Ownership & Reentrancy: Pure caller-owned POD value records; 100%
//     stateless and safe across goroutines.
//   - Extension Seam: Add new hardware probes or cascade telemetry fields here
//     so Layer 2 (`orchestrator/`) and Layer 4 (`cmd/`) stay strictly aligned.
//   - Explicit Non-Goal: Performs zero OS hardware probing (owned by Layer 2
//     `orchestrator/probe.go`).
// ==============================================================================

package data

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Shared Pure Structural Vocabulary, Value Bounds & Stateless Predicates )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 0: Pure Invariant Data Substrate
// Module:        cog.data.profile
// Mission:       Provide hardware capability, node manifest, link health, and
//                cascade event value records with O(1) pure queries.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages (None — Pure Primitive Vocabulary) ---

// --- Tier 2: External Library Dependencies (None — 100% Go Stdlib Invariant) ---

// --- Tier 3: Internal FaithNet Foundation Packages (Same Package: data) ---

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Pre-Conditions, Default Constants & Profile Limits
// ------------------------------------------------------------------------------
// Sovereign offline-first baseline defaults for zero-config startup.
// ------------------------------------------------------------------------------

const (
	// DefaultWellListenAddr is the loopback portal address for The Well.
	DefaultWellListenAddr = "127.0.0.1:8077"

	// DefaultBeaconSSID is the default local Wi-Fi AP beacon identifier.
	DefaultBeaconSSID = "FaithNet-Beacon"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.2: Pure-Types Logical Anchor
// ------------------------------------------------------------------------------
// Movement-resilient identifier shared by diagnostic and audit subsystems.
// ------------------------------------------------------------------------------

const (
	// ModuleIDProfile is the canonical logical ID for this pure_types leaf.
	ModuleIDProfile = "cog.data.profile"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.3: Public Tagged Ontology (Profile, Manifest & Health Records)
// ------------------------------------------------------------------------------
// Canonical caller-owned data records driving Layer 2 orchestration.
// ------------------------------------------------------------------------------

// HardwareProfile describes the probed transport capabilities of the host PC.
//
// ### For Architects, Operators & Non-Programmers:
// Summarizes what communication organs the current machine possesses so the
// node can automatically adapt to an ancient laptop, a desktop tower, or a
// modern server without manual configuration.
//
// ### For Systems Programmers & Engine Architects:
// Populated by `orchestrator.ProbeHostBody()`. Injectable verbatim in unit
// tests to simulate crippled or offline hardware bodies deterministically.
type HardwareProfile struct {
	OSName          string // Probed host operating system (linux/windows/darwin)
	Arch            string // Probed host CPU architecture (amd64/arm64/etc.)
	CPUCores        int    // Number of logical CPU cores available
	HasMemoryBus    bool   // Rung 2: Always true on any running computer
	HasLocalIPC     bool   // Rung 2: Unix sockets / named pipes available
	HasLocalUDP     bool   // Rung 3: LAN / loopback UDP interfaces available
	HasWiFiHardware bool   // Rung 3: Wireless interface detected on host
	HasSerialPorts  bool   // Rung 1: RS-232 / ttyS / ttyUSB / COM detected
	HasAudioDevice  bool   // Rung 1: Sound hardware for AFSK / dial-up tones
	HasGlobalTCP    bool   // Rung 4: Outbound TCP stack available
	InterfaceCount  int    // Total non-zero network interfaces detected
}

// NodeManifest is the data-driven startup specification for a FaithNet node.
//
// ### For Architects, Operators & Non-Programmers:
// Tells a FaithNet node its sovereign name, whether to stay strictly offline
// or open the optional internet airlock, and whether to open The Well portal.
//
// ### For Systems Programmers & Engine Architects:
// Immutable startup descriptor passed into Layer 2 `orchestrator.NewDispatcher`.
type NodeManifest struct {
	ID                NodeID    // Sovereign 64-bit identity of this node
	Label             string    // Human-readable operator label for diagnostics
	PreferredMaxRung  RungLevel // Highest rung allowed by operator policy
	MinimumFloorRung  RungLevel // Lowest survival floor (default: Rung1Floor)
	AllowGlobalEgress bool      // Explicit opt-in for Rung 4 commercial egress
	EnableWellPortal  bool      // Enable zero-install HTML portal ("The Well")
	WellListenAddr    string    // Bind address for The Well HTTP listener
	BeaconSSID        string    // Wireless SSID announced when AP switch is ON
}

// LinkHealth captures deterministic heartbeat statistics for a peer or link.
//
// ### For Architects, Operators & Non-Programmers:
// A scorecard showing how many sonar pings were sent, how many echoes came
// back, and the fastest, slowest, and average response speed.
//
// ### For Systems Programmers & Engine Architects:
// Value-copied out of `engine.PingEngine` via Block 5 `InspectHealth()` with
// zero lock retention by callers.
type LinkHealth struct {
	PeerID      NodeID    // Sovereign ID of the last peer communicating on link
	ActiveRung  RungLevel // Survival rung carrying the active heartbeat
	PingsSent   uint64    // Cumulative OpPing pulses emitted
	PongsRecv   uint64    // Cumulative valid OpPong echoes correlated
	LastRTTNano int64     // Most recent monotonic round-trip latency (ns)
	MinRTTNano  int64     // Minimum observed round-trip latency (-1 if none)
	MaxRTTNano  int64     // Maximum observed round-trip latency (ns)
	AvgRTTNano  int64     // Running integer mean round-trip latency (ns)
	LastSeenNs  int64     // Monotonic timestamp of most recent verified echo
}

// CascadeEvent records a single Rising-Up or Falling-Down ladder transition.
//
// ### For Architects, Operators & Non-Programmers:
// A flight-recorder entry noting the exact moment FaithNet stepped down to a
// survival floor or climbed back up when a faster link healed.
//
// ### For Systems Programmers & Engine Architects:
// Stored inside `orchestrator.AdaptiveLadder`'s fixed-capacity history ring.
type CascadeEvent struct {
	TimestampNano int64     // Monotonic timestamp of ladder transition
	FromRung      RungLevel // Previous active survival rung
	ToRung        RungLevel // Newly activated survival rung
	Direction     Opcode    // OpCascadeDown or OpCascadeUp
	Reason        string    // Human-readable cause of transition
}

// ------------------------------------------------------------------------------
// Sub-Assembly 2.4: Pre-Conditional Helper Cogs & Stateless Profile Predicates
// ------------------------------------------------------------------------------
// Pure O(1) capability queries and sovereign default manifest constructor.
// ------------------------------------------------------------------------------

// HighestViableRung returns the top adaptive rung supported by the host body.
//
// ### For Architects, Operators & Non-Programmers:
// Reads the machine's capability chart and picks the fastest survival tier
// the physical hardware can actually sustain.
//
// ### For Systems Programmers & Engine Architects:
// Pure O(1) descending priority cascade from Rung4Global down to RungNone.
func (p HardwareProfile) HighestViableRung() RungLevel {
	if p.HasGlobalTCP {
		return Rung4Global
	}
	if p.HasLocalUDP || p.HasWiFiHardware {
		return Rung3Local
	}
	if p.HasMemoryBus || p.HasLocalIPC {
		return Rung2Host
	}
	if p.HasSerialPorts || p.HasAudioDevice {
		return Rung1Floor
	}
	return RungNone
}

// IsValid reports whether a NodeManifest obeys sovereign identity and rung
// ordering invariants (valid ID, valid rungs, Minimum <= PreferredMax, and
// Rung4Global prohibited unless AllowGlobalEgress is explicitly true).
//
// ### For Architects, Operators & Non-Programmers:
// Verifies a node's startup configuration is safe and never accidentally opens
// outside internet access without explicit operator consent.
//
// ### For Systems Programmers & Engine Architects:
// Pure O(1) policy gate evaluated by Layer 2 Dispatcher initialization and
// Block 4 manifest contract audits.
func (m NodeManifest) IsValid() bool {
	if !m.ID.IsValid() || m.ID.IsBroadcast() {
		return false
	}
	if !m.MinimumFloorRung.IsValid() || !m.PreferredMaxRung.IsValid() {
		return false
	}
	if m.MinimumFloorRung > m.PreferredMaxRung {
		return false
	}
	if m.PreferredMaxRung == Rung4Global && !m.AllowGlobalEgress {
		return false
	}
	return true
}

// LossPercent returns integer packet loss percentage clamped to [0, 100].
//
// ### For Architects, Operators & Non-Programmers:
// Computes what percent of outgoing sonar pings went unanswered so the
// adaptive ladder knows when a link is deteriorating.
//
// ### For Systems Programmers & Engine Architects:
// Pure O(1) integer arithmetic; returns 0 when PingsSent == 0 or PongsRecv >=
// PingsSent, eliminating divide-by-zero and unsigned underflow traps.
func (h LinkHealth) LossPercent() int {
	if h.PingsSent == 0 || h.PongsRecv >= h.PingsSent {
		return 0
	}
	lost := h.PingsSent - h.PongsRecv
	pct := int((lost * 100) / h.PingsSent)
	if pct > 100 {
		return 100
	}
	return pct
}

// DefaultNodeManifest returns a sovereign offline-first manifest (Rung 1..3).
//
// ### For Architects, Operators & Non-Programmers:
// Builds a safe default node configuration that runs completely offline in the
// room or neighborhood and keeps global internet egress sealed off by default.
//
// ### For Systems Programmers & Engine Architects:
// Pure value constructor guaranteeing `AllowGlobalEgress == false` and
// `MinimumFloorRung == Rung1Floor`.
func DefaultNodeManifest(id NodeID, label string) NodeManifest {
	return NodeManifest{
		ID:                id,
		Label:             label,
		PreferredMaxRung:  Rung3Local,
		MinimumFloorRung:  Rung1Floor,
		AllowGlobalEgress: false,
		EnableWellPortal:  false,
		WellListenAddr:    DefaultWellListenAddr,
		BeaconSSID:        DefaultBeaconSSID,
	}
}

// ==============================================================================
// BLOCK 3: BODY BLOCK
// ( Statutory Pure-Types Notice & Structural Boundary )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 0: Pure Invariant Data Substrate
// Module:        cog.data.profile
// Mission:       Pure-types profile vocabulary module: define purely static
//                capability manifests, telemetry records, and O(1) queries;
//                live hardware probing and ladder cascade engines reside in
//                Layer 2 (`orchestrator/probe.go` and `orchestrator/ladder.go`).
// ==============================================================================

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Statutory Pure-Types Audit Cross-Reference )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 0: Pure Invariant Data Substrate
// Module:        cog.data.profile
// Mission:       Functional audits, offline-first manifest tripwires, and
//                pre-flight routines verifying host profiles are declared in
//                Layer 2 (`orchestrator/probe.go`: `AuditProbe`, `PreflightProbe`).
// ==============================================================================

// ==============================================================================
// BLOCK 5: CLOSING BLOCK / STATUTORY EXEMPTION & LIFECYCLE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 0: Pure Invariant Data Substrate
// Module:        cog.data.profile
// Mission:       Operational inspection cards, subsystem lifecycle init, and
//                teardown routines are declared in Layer 2 (`orchestrator/`).
// Stateless Exemption Notice: None. (Stateless module / pure stack-allocated execution).
// ==============================================================================

// End of data/profile.go
