// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          transport/rung3_local/beacon.go
// @id            cog.transport.rung3_beacon
// @layer         Layer 3A: Physical & Virtual Transport Substrate (Rung 3 Local)
// @archetype     operational | stateful
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Room Wi-Fi AP Beacon "On/Off Switch" Controller
// @role          Safe Cross-Platform Local Beacon State Machine & Frame Builder
// @maturity      Rev 5 Baseline | Wave 4 The Room & The Well
// @chassis       zero-heap | Bounded value config + atomic toggle telemetry
// @scale         O(1) toggle, state query, and OpBeacon frame construction
// @seams         in: data, engine.Clock | uses: os/exec (capability probe only)
// @invariants    1. Stratum Baseline: Downward Dependency Law, Rev 5 5-Block.
//                2. Default mode is BeaconModeSimulated: NEVER mutates host OS
//                   Wi-Fi settings during tests or without explicit operator flag.
//                3. SSID strictly bounded to IEEE 802.11 32-byte limit.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Implements FaithNet's Rung 3 Room Beacon ("On/Off Switch"). Manages the
//   sovereign local Access Point broadcast state (SSID, IEEE channel, active
//   switch) and constructs canonical `data.OpBeacon` wire frames announcing
//   this node's presence and active survival rung to nearby neighbors.
//
//   ### For Architects, Operators & Non-Programmers:
//   Gives a FaithNet device a clean On/Off switch to broadcast a local room
//   presence beacon (like turning a laptop into a local campfire hotspot) while
//   staying 100% safe and non-destructive when running in simulation mode.
//
//   ### For Systems Programmers & Engine Architects:
//   - Ownership & Reentrancy: Mutex-guarded state machine + immutable clock.
//   - Extension Seam: `BeaconModeSystem` probes host wireless tooling (`nmcli`,
//     `iw`) cleanly without embedding fragile OS shell assumptions in tests.
// ==============================================================================

package rung3_local

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( IEEE 802.11 Bounds, Modes, BeaconSwitch Chassis & O(1) Accessors )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3A: Physical & Virtual Transport Substrate (Rung 3 Local)
// Module:        cog.transport.rung3_beacon
// Mission:       Define IEEE 802.11 SSID/channel limits, beacon operating modes,
//                and thread-safe BeaconSwitch state queries.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"errors"
	"fmt"
	"os/exec"
	"sync"
)

// --- Tier 2: External Library Dependencies (None — 100% Go Stdlib Invariant) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---
import (
	"github.com/CreativeWorkzStudio/FaithNet/data"
	"github.com/CreativeWorkzStudio/FaithNet/engine"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Pre-Conditions, IEEE 802.11 Bounds & Sentinel Errors
// ------------------------------------------------------------------------------

const (
	// MaxBeaconSSIDBytes is the IEEE 802.11 maximum SSID length (32 bytes).
	MaxBeaconSSIDBytes = 32

	// MinBeaconChannel is the lowest 2.4 GHz IEEE 802.11 channel.
	MinBeaconChannel = 1

	// MaxBeaconChannel is the highest common IEEE 802.11 channel.
	MaxBeaconChannel = 14

	// DefaultBeaconChannel is non-overlapping 2.4 GHz channel 6.
	DefaultBeaconChannel = 6
)

var (
	// ErrInvalidBeaconSSID indicates an empty or >32-byte SSID.
	ErrInvalidBeaconSSID = errors.New("faithnet/beacon: SSID must be 1..32 bytes")

	// ErrInvalidBeaconChannel indicates an out-of-range Wi-Fi channel.
	ErrInvalidBeaconChannel = errors.New("faithnet/beacon: channel must be in [1..14]")

	// ErrBeaconBackendUnavailable indicates system Wi-Fi AP tooling is absent.
	ErrBeaconBackendUnavailable = errors.New("faithnet/beacon: host AP backend unavailable")
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.2: Operational Logical Anchor
// ------------------------------------------------------------------------------

const (
	// ModuleIDBeacon is the canonical logical ID for the Rung 3 Beacon switch.
	ModuleIDBeacon = "cog.transport.rung3_beacon"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.3: Public Tagged Ontology (BeaconMode, Status & BeaconSwitch)
// ------------------------------------------------------------------------------

// BeaconMode controls whether BeaconSwitch runs safely in simulation or probes
// native OS wireless AP tools.
type BeaconMode uint8

const (
	// BeaconModeSimulated manages beacon state safely in memory (default).
	BeaconModeSimulated BeaconMode = iota

	// BeaconModeSystem probes native OS wireless tooling (`nmcli`/`iw`).
	BeaconModeSystem
)

// BeaconStatus captures a read-only value snapshot of BeaconSwitch state.
type BeaconStatus struct {
	Active        bool       // True when room beacon switch is ON
	SSID          string     // Active broadcast SSID (1..32 bytes)
	Channel       int        // IEEE 802.11 channel (1..14)
	Mode          BeaconMode // Simulated or System backend mode
	ToggleCount   uint64     // Cumulative ON/OFF switch transitions
	StartedAtNano int64      // Monotonic timestamp of most recent TurnOn()
}

// BeaconSwitch is the Rung 3 Room Wi-Fi AP "On/Off Switch" state machine.
//
// ### For Architects, Operators & Non-Programmers:
// Turns local room visibility ON or OFF with one call and builds clean
// FaithNet presence announcements (`OpBeacon`) for neighbors to hear.
//
// ### For Systems Programmers & Engine Architects:
// Mutex-guarded deterministic controller with zero host side effects in
// `BeaconModeSimulated`.
type BeaconSwitch struct {
	mu            sync.RWMutex
	ssid          string
	channel       int
	mode          BeaconMode
	clock         engine.Clock
	active        bool
	toggleCount   uint64
	startedAtNano int64
}

// ------------------------------------------------------------------------------
// Sub-Assembly 2.4: Pre-Conditional Helper Cogs & O(1) Mode Stringifiers
// ------------------------------------------------------------------------------

// String returns the canonical symbolic name of BeaconMode.
//
// ### For Architects, Operators & Non-Programmers:
// Labels whether the room beacon is running in safe simulation or system mode.
//
// ### For Systems Programmers & Engine Architects:
// Pure O(1) static string switch (0 heap allocations).
func (m BeaconMode) String() string {
	switch m {
	case BeaconModeSimulated:
		return "SIMULATED_SAFE"
	case BeaconModeSystem:
		return "SYSTEM_HOST_AP"
	default:
		return "UNKNOWN_MODE"
	}
}

// IsActive reports whether the Room Beacon switch is currently ON.
//
// ### For Architects, Operators & Non-Programmers:
// Checks whether this device is currently broadcasting its room beacon.
//
// ### For Systems Programmers & Engine Architects:
// Read-locked O(1) boolean query.
func (b *BeaconSwitch) IsActive() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.active
}

// ==============================================================================
// BLOCK 3: BODY BLOCK
// ( Composable Business Logic: Switch On/Off & OpBeacon Frame Construction )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3A: Physical & Virtual Transport Substrate (Rung 3 Local)
// Module:        cog.transport.rung3_beacon
// Mission:       Toggle local room beacon state safely and construct canonical
//                OpBeacon presence frames for LAN/Wi-Fi announcement.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: Beacon On/Off Switch Control
// ------------------------------------------------------------------------------

// TurnOn activates the local room beacon switch cleanly.
//
// ### For Architects, Operators & Non-Programmers:
// Flips the FaithNet room campfire switch ON so nearby phones and laptops can
// discover this sovereign node.
//
// ### For Systems Programmers & Engine Architects:
// In `BeaconModeSystem`, verifies `nmcli` or `iw` exists before flipping active;
// idempotent when already ON.
func (b *BeaconSwitch) TurnOn() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.active {
		return nil
	}
	if b.mode == BeaconModeSystem {
		_, errNM := exec.LookPath("nmcli")
		_, errIW := exec.LookPath("iw")
		if errNM != nil && errIW != nil {
			return ErrBeaconBackendUnavailable
		}
	}
	b.active = true
	b.toggleCount++
	b.startedAtNano = b.clock.NowNano()
	return nil
}

// TurnOff deactivates the local room beacon switch cleanly.
//
// ### For Architects, Operators & Non-Programmers:
// Flips the FaithNet room beacon OFF immediately (idempotent if already off).
//
// ### For Systems Programmers & Engine Architects:
// Write-locked transition updating `active` and `toggleCount` atomically.
func (b *BeaconSwitch) TurnOff() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.active {
		return
	}
	b.active = false
	b.toggleCount++
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: Sovereign OpBeacon Announcement Builder
// ------------------------------------------------------------------------------

// BuildBeaconFrame populates a canonical `data.OpBeacon` broadcast Frame
// advertising selfID, activeRung, and this beacon's SSID.
//
// ### For Architects, Operators & Non-Programmers:
// Creates a compact announcement envelope saying "I am here in the room on
// this SSID and survival rung."
//
// ### For Systems Programmers & Engine Architects:
// Value-typed `data.Frame` builder targeting `data.NodeIDBroadcast` cleanly.
func (b *BeaconSwitch) BuildBeaconFrame(
	selfID data.NodeID,
	activeRung data.RungLevel,
	seq uint32,
) data.Frame {
	b.mu.RLock()
	ssid := b.ssid
	ch := b.channel
	b.mu.RUnlock()

	var f data.Frame
	f.Version = data.ProtocolVersion
	f.Rung = activeRung
	if !f.Rung.IsValid() {
		f.Rung = data.Rung3Local
	}
	f.Opcode = data.OpBeacon
	f.Flags = data.FlagFloorCompact | data.FlagCovenantVerified
	f.Truth = data.QuadPositive
	f.SourceID = selfID
	f.TargetID = data.NodeIDBroadcast
	f.SeqID = seq
	f.SetPayload([]byte(fmt.Sprintf("BEACON:%s:CH%d", ssid, ch)))
	return f
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Contract Verification Cogs & Pre-Flight Validation Tripwires )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3A: Physical & Virtual Transport Substrate (Rung 3 Local)
// Module:        cog.transport.rung3_beacon
// Mission:       Audit live BeaconSwitch IEEE bounds and execute a 4-section
//                pre-flight check in safe simulation mode.
// @triggers
//   - Block 4 Audits : (*BeaconSwitch).AuditState(), PreflightBeaconSwitch(),
//                      AuditBeaconSwitch()
//   - Layer 4 Tests  : tests/unit/transport/rung3_local_test.go,
//                      tests/integrated/room_and_well_test.go
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Contract, Input & Live-Instance Verification Cogs
// ------------------------------------------------------------------------------

// AuditState inspects a live BeaconSwitch non-destructively to confirm IEEE
// 802.11 SSID (1..32B), valid channel (1..14), and attached Clock.
//
// ### For Architects, Operators & Non-Programmers:
// Checks the room beacon switch settings to verify its Wi-Fi name and radio
// channel are legal before broadcasting.
//
// ### For Systems Programmers & Engine Architects:
// Read-locked O(1) live probe returning structured `data.AuditResult`.
func (b *BeaconSwitch) AuditState() data.AuditResult {
	const check = "rung3_beacon_live_state_contract"

	if b == nil {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDBeacon, check,
			"BeaconSwitch receiver is nil.")
	}
	b.mu.RLock()
	ssid := b.ssid
	ch := b.channel
	clk := b.clock
	b.mu.RUnlock()

	if len(ssid) == 0 || len(ssid) > MaxBeaconSSIDBytes {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDBeacon, check,
			"BeaconSwitch SSID breaches IEEE 802.11 [1..32] byte bounds.")
	}
	if ch < MinBeaconChannel || ch > MaxBeaconChannel {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDBeacon, check,
			"BeaconSwitch channel lies outside [1..14].")
	}
	if clkRes := engine.ValidateClockContract(clk); !clkRes.IsPass() {
		return clkRes
	}
	return data.AuditPass(ModuleIDBeacon, check,
		"Live Rung 3 BeaconSwitch SSID, channel, and clock invariants verified.")
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Machine Self-Audit & Pre-Flight Validation Tripwires
// ------------------------------------------------------------------------------

// PreflightBeaconSwitch executes a 4-section non-destructive pre-flight audit
// in safe simulation mode covering (1) On/Off toggle & OpBeacon build,
// (2) invalid SSID/channel rejection, (3) 32-byte SSID ceiling, and (4) Close.
//
// ### For Architects, Operators & Non-Programmers:
// Bench-tests the room beacon On/Off switch safely in memory without touching
// your computer's live Wi-Fi adapter.
//
// ### For Systems Programmers & Engine Architects:
// Isolated 4-section pre-flight gate returning structured `data.AuditResult`.
func PreflightBeaconSwitch() data.AuditResult {
	const check = "rung3_beacon_preflight_4section"

	// Micro Check 1 (Always Pass): Nominal TurnOn -> BuildBeaconFrame -> TurnOff.
	sw, err := NewBeaconSwitch(data.DefaultBeaconSSID, DefaultBeaconChannel,
		BeaconModeSimulated, engine.NewMockClock(100))
	if err != nil {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDBeacon, check,
			"Sec 1 NewBeaconSwitch nominal construction failed.")
	}
	if err := sw.TurnOn(); err != nil || !sw.IsActive() {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDBeacon, check,
			"Sec 1 BeaconSwitch TurnOn failed.")
	}
	bf := sw.BuildBeaconFrame(0x01, data.Rung3Local, 1)
	if bf.Opcode != data.OpBeacon || !bf.TargetID.IsBroadcast() {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDBeacon, check,
			"Sec 1 BuildBeaconFrame produced invalid OpBeacon frame.")
	}

	// Micro Check 2 (Never Fail): Reject empty SSID, >32B SSID, and bad channel.
	if _, err := NewBeaconSwitch("", 6, BeaconModeSimulated, nil); err != ErrInvalidBeaconSSID {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDBeacon, check,
			"Sec 2 empty SSID was not rejected.")
	}
	if _, err := NewBeaconSwitch("123456789012345678901234567890123", 6, BeaconModeSimulated, nil); err != ErrInvalidBeaconSSID {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDBeacon, check,
			"Sec 2 33-byte SSID bypassed IEEE 802.11 limit.")
	}
	if _, err := NewBeaconSwitch("OK", 99, BeaconModeSimulated, nil); err != ErrInvalidBeaconChannel {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDBeacon, check,
			"Sec 2 OOB Wi-Fi channel 99 was not rejected.")
	}

	// Micro Check 3 (Invariants & Limits): Exact 32-byte SSID accepted.
	if _, err := NewBeaconSwitch("12345678901234567890123456789012", 14, BeaconModeSimulated, nil); err != nil {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDBeacon, check,
			"Sec 3 exact 32-byte SSID on channel 14 was rejected.")
	}

	// Micro Check 4 (Emergence): Live AuditState & graceful Close() -> OFF.
	if live := sw.AuditState(); !live.IsPass() {
		return live
	}
	_ = sw.Close()
	if sw.IsActive() {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDBeacon, check,
			"Sec 4 Close() failed to turn BeaconSwitch OFF.")
	}
	return data.AuditPass(ModuleIDBeacon, check,
		"Rung 3 BeaconSwitch passed all 4 micro pre-flight gates.")
}

// AuditBeaconSwitch returns true iff PreflightBeaconSwitch() passes cleanly.
//
// ### For Architects, Operators & Non-Programmers:
// Quick pass/fail check confirming the room beacon switch passed safety.
//
// ### For Systems Programmers & Engine Architects:
// Convenience boolean bridge delegating to `PreflightBeaconSwitch().IsPass()`.
func AuditBeaconSwitch() bool {
	return PreflightBeaconSwitch().IsPass()
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK / EXECUTABLE PROCEDURES & LIFECYCLE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3A: Physical & Virtual Transport Substrate (Rung 3 Local)
// Module:        cog.transport.rung3_beacon
// Mission:       Expose status snapshots, 60-col cards, epilogue readiness, and
//                BeaconSwitch initialization/Close teardown.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Executable Procedures, Inspection, Readiness & Cards
// ------------------------------------------------------------------------------

// IsBeaconSwitchReady reports whether the Rung 3 BeaconSwitch passed preflight.
//
// ### For Architects, Operators & Non-Programmers:
// Epilogue confirmation that the room beacon controller is healthy and ready.
//
// ### For Systems Programmers & Engine Architects:
// Matches OmniCode's Block 5 epilogue readiness predicate (`is_*_ready`).
func IsBeaconSwitchReady() bool {
	return PreflightBeaconSwitch().IsPass()
}

// StatusSnapshot returns a thread-safe value copy of current BeaconStatus.
//
// ### For Architects, Operators & Non-Programmers:
// Reads the beacon's current ON/OFF state, SSID, channel, and toggle count.
//
// ### For Systems Programmers & Engine Architects:
// Read-locked value copy with zero shared references.
func (b *BeaconSwitch) StatusSnapshot() BeaconStatus {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return BeaconStatus{
		Active:        b.active,
		SSID:          b.ssid,
		Channel:       b.channel,
		Mode:          b.mode,
		ToggleCount:   b.toggleCount,
		StartedAtNano: b.startedAtNano,
	}
}

// ExecuteBeaconDumpCard renders a 60-column inner-width Room Beacon card.
//
// ### For Architects, Operators & Non-Programmers:
// Prints a neat framed switchboard card showing SSID, channel, mode, and ON/OFF.
//
// ### For Systems Programmers & Engine Architects:
// Non-mutating diagnostic card builder invoking `StatusSnapshot()`.
func (b *BeaconSwitch) ExecuteBeaconDumpCard() string {
	s := b.StatusSnapshot()
	state := "OFF"
	if s.Active {
		state = "ON (BROADCASTING)"
	}
	return fmt.Sprintf(
		"+------------------------------------------------------------+\n"+
			"| %-58s |\n"+
			"+------------------------------------------------------------+\n"+
			"| Module:   %-48s |\n"+
			"| Switch:   %-48s |\n"+
			"| SSID:     %-48s |\n"+
			"| Channel:  %-48d |\n"+
			"| Mode:     %-48s |\n"+
			"+------------------------------------------------------------+",
		"FAITHNET RUNG 3 ROOM BEACON SWITCH CARD",
		ModuleIDBeacon,
		state,
		s.SSID,
		s.Channel,
		s.Mode.String(),
	)
}

// ------------------------------------------------------------------------------
// Sub-Assembly 5.2: Subsystem Lifecycle Initialization & Teardown
// ------------------------------------------------------------------------------

// NewBeaconSwitch constructs a validated Room Beacon Switch in OFF state.
//
// ### For Architects, Operators & Non-Programmers:
// Prepares a local room beacon switch with a validated Wi-Fi name and channel.
//
// ### For Systems Programmers & Engine Architects:
// Enforces IEEE 802.11 SSID (1..32B) and channel (1..14) bounds upfront.
func NewBeaconSwitch(
	ssid string,
	channel int,
	mode BeaconMode,
	clock engine.Clock,
) (*BeaconSwitch, error) {
	if len(ssid) == 0 || len(ssid) > MaxBeaconSSIDBytes {
		return nil, ErrInvalidBeaconSSID
	}
	if channel < MinBeaconChannel || channel > MaxBeaconChannel {
		return nil, ErrInvalidBeaconChannel
	}
	if clock == nil {
		clock = engine.NewMonotonicClock()
	}
	return &BeaconSwitch{
		ssid:    ssid,
		channel: channel,
		mode:    mode,
		clock:   clock,
	}, nil
}

// Close ensures the Room Beacon switch is cleanly turned OFF.
//
// ### For Architects, Operators & Non-Programmers:
// Safely switches off room broadcasting during node shutdown.
//
// ### For Systems Programmers & Engine Architects:
// Delegates idempotently to `TurnOff()`.
func (b *BeaconSwitch) Close() error {
	b.TurnOff()
	return nil
}

// End of transport/rung3_local/beacon.go
