// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          orchestrator/probe.go
// @id            cog.orchestrator.probe
// @layer         Layer 2: Adaptive Orchestration & Ladder Control
// @archetype     operational | stateless
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Dynamic Host Body Hardware Prober
// @role          Data-Driven Hardware Discovery & Policy Ceiling Evaluation
// @maturity      Rev 5 Baseline | Wave 3 Orchestrator
// @chassis       bounded-heap | Inspects host interfaces into value struct
// @scale         O(N_ifaces) startup capability scan; O(1) policy ceiling math
// @seams         in: data.HardwareProfile, data.NodeManifest | uses: net, os
// @invariants    1. Stratum Baseline: Downward Dependency Law, Rev 5 5-Block.
//                2. Zero OS-name branching inside core networking logic: host
//                   reality is measured once into data.HardwareProfile.
//                3. Offline-First Guard: EvaluatePolicyCeiling NEVER returns
//                   Rung4Global unless manifest.AllowGlobalEgress == true.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Implements FaithNet's dynamic host body prober ("What body did I wake up
//   in?"). Inspects OS runtime, network interfaces, local IPC writability,
//   serial ports, and audio devices into a pure `data.HardwareProfile` value
//   struct so `AdaptiveLadder` can rise up or fall down purely from data.
//
//   ### For Architects, Operators & Non-Programmers:
//   Acts like an automatic morning checkup when FaithNet starts on any machine
//   (old laptop, bare-metal PC, or modern server), noting which wires, radios,
//   or sound cards exist so the network adapts with zero manual config.
//
//   ### For Systems Programmers & Engine Architects:
//   - Ownership & Reentrancy: Pure stateless functions returning caller-owned
//     `data.HardwareProfile` values; 100% goroutine-safe.
//   - Extension Seam: Add platform-specific device probes cleanly in
//     Sub-Assembly 2.5 without altering `data.HardwareProfile` consumers.
// ==============================================================================

package orchestrator

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Probe Constants, Logical Anchor & Host Capability Support Helpers )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 2: Adaptive Orchestration & Ladder Control
// Module:        cog.orchestrator.probe
// Mission:       Inspect host machine capabilities into data.HardwareProfile
//                and compute safe data-driven rung ceilings.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// --- Tier 2: External Library Dependencies (None — 100% Go Stdlib Invariant) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---
import "github.com/CreativeWorkzStudio/FaithNet/data"

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Pre-Conditions & Device Discovery Globs
// ------------------------------------------------------------------------------

var wirelessPrefixes = [...]string{
	"wl", "wlan", "wifi", "wi-fi", "en0", "ath", "iwl",
}

var linuxSerialGlobs = [...]string{
	"/dev/ttyS*",
	"/dev/ttyUSB*",
	"/dev/ttyACM*",
}

// ------------------------------------------------------------------------------
// Sub-Assembly 2.2: Operational Logical Anchor
// ------------------------------------------------------------------------------

const (
	// ModuleIDProbe is the canonical logical ID for the host body prober.
	ModuleIDProbe = "cog.orchestrator.probe"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.3: Public Tagged Ontology (Reuses Layer 0 data.HardwareProfile)
// ------------------------------------------------------------------------------

// ------------------------------------------------------------------------------
// Sub-Assembly 2.4: Pre-Conditional Helper Cogs & O(1) Policy Ceiling Math
// ------------------------------------------------------------------------------

// EvaluatePolicyCeiling resolves the highest survival rung permitted by BOTH
// the probed physical host body and the sovereign NodeManifest policy.
//
// ### For Architects, Operators & Non-Programmers:
// Combines what hardware the computer actually has with what the operator
// explicitly allowed—ensuring a node never climbs onto the commercial internet
// unless the operator explicitly opted in.
//
// ### For Systems Programmers & Engine Architects:
// Pure O(1) intersection returning `data.RungNone` when no rung satisfies
// `[manifest.MinimumFloorRung .. min(body.HighestViableRung(), policyMax)]`.
func EvaluatePolicyCeiling(
	body data.HardwareProfile,
	manifest data.NodeManifest,
) data.RungLevel {
	if !manifest.IsValid() {
		return data.RungNone
	}
	hwMax := body.HighestViableRung()
	if !hwMax.IsValid() {
		return data.RungNone
	}

	policyMax := manifest.PreferredMaxRung
	if !manifest.AllowGlobalEgress && policyMax > data.Rung3Local {
		policyMax = data.Rung3Local
	}

	best := hwMax
	if policyMax < best {
		best = policyMax
	}
	if best < manifest.MinimumFloorRung {
		return data.RungNone
	}
	return best
}

// ------------------------------------------------------------------------------
// Sub-Assembly 2.5: Monitored Support Infrastructure
// ------------------------------------------------------------------------------
// Support Role:  Best-effort OS interface and device-node inspection helpers.
// Pattern Watch: If Windows/macOS specific hardware probes grow substantially,
//                split into build-tagged `probe_linux.go` / `probe_other.go`.
// ------------------------------------------------------------------------------

func looksWirelessName(name string) bool {
	lower := strings.ToLower(name)
	for i := 0; i < len(wirelessPrefixes); i++ {
		if strings.HasPrefix(lower, wirelessPrefixes[i]) {
			return true
		}
	}
	return false
}

func probeSerialDevicePresence() bool {
	for i := 0; i < len(linuxSerialGlobs); i++ {
		matches, err := filepath.Glob(linuxSerialGlobs[i])
		if err == nil && len(matches) > 0 {
			return true
		}
	}
	return false
}

func probeAudioDevicePresence() bool {
	if _, err := os.Stat("/dev/snd"); err == nil {
		return true
	}
	if _, err := os.Stat("/dev/dsp"); err == nil {
		return true
	}
	return false
}

func probeLocalIPCAvailable() bool {
	tmp := os.TempDir()
	if tmp == "" {
		return false
	}
	info, err := os.Stat(tmp)
	return err == nil && info.IsDir()
}

// ==============================================================================
// BLOCK 3: BODY BLOCK
// ( Composable Business Logic: Dynamic Host Body Hardware Inspection )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 2: Adaptive Orchestration & Ladder Control
// Module:        cog.orchestrator.probe
// Mission:       Measure live host capabilities non-destructively into a
//                portable data.HardwareProfile value record.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: Live Host Body Probing
// ------------------------------------------------------------------------------

// ProbeHostBody inspects the current host machine's runtime, network
// interfaces, IPC directory, serial ports, and audio devices into a
// deterministic `data.HardwareProfile` value struct.
//
// ### For Architects, Operators & Non-Programmers:
// Scans what communication hardware this machine currently has—never guessing
// and never crashing if a laptop has no Wi-Fi, no Ethernet, or no sound card.
//
// ### For Systems Programmers & Engine Architects:
// Non-blocking, read-only OS inspection; never opens outbound connections or
// modifies host configuration. `HasMemoryBus` is always true on any live CPU.
func ProbeHostBody() data.HardwareProfile {
	prof := data.HardwareProfile{
		OSName:         runtime.GOOS,
		Arch:           runtime.GOARCH,
		CPUCores:       runtime.NumCPU(),
		HasMemoryBus:   true,
		HasLocalIPC:    probeLocalIPCAvailable(),
		HasSerialPorts: probeSerialDevicePresence(),
		HasAudioDevice: probeAudioDevicePresence(),
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		return prof
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		prof.InterfaceCount++
		if looksWirelessName(iface.Name) {
			prof.HasWiFiHardware = true
		}

		addrs, addrErr := iface.Addrs()
		if addrErr != nil || len(addrs) == 0 {
			continue
		}
		prof.HasLocalUDP = true
		if iface.Flags&net.FlagLoopback == 0 {
			prof.HasGlobalTCP = true
		}
	}
	return prof
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Contract Verification Cogs & Pre-Flight Validation Tripwires )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 2: Adaptive Orchestration & Ladder Control
// Module:        cog.orchestrator.probe
// Mission:       Audit probed HardwareProfile invariants and run a 4-section
//                pre-flight check verifying offline-first policy capping.
// @triggers
//   - Block 4 Audits : AuditHardwareProfile(), PreflightProbe(), AuditProbe()
//   - Layer 4 Tests  : tests/functional/orchestrator_functional_test.go,
//                      tests/unit/orchestrator/probe_test.go,
//                      tests/integrated/orchestrator_ladder_test.go
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Contract, Input & Live-Instance Verification Cogs
// ------------------------------------------------------------------------------

// AuditHardwareProfile validates a probed `data.HardwareProfile` against
// structural consistency bounds (non-empty OS/Arch, positive cores, valid rung).
//
// ### For Architects, Operators & Non-Programmers:
// Double-checks that a machine's capability chart is complete and sensible
// before handing it to the adaptive ladder.
//
// ### For Systems Programmers & Engine Architects:
// Pure O(1) contract gate returning structured `data.AuditResult`.
func AuditHardwareProfile(p data.HardwareProfile) data.AuditResult {
	const check = "hardware_profile_contract"

	if p.OSName == "" || p.Arch == "" {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDProbe, check,
			"HardwareProfile missing OSName or Arch identity.")
	}
	if p.CPUCores < 1 {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDProbe, check,
			"HardwareProfile CPUCores must be >= 1.")
	}
	if p.InterfaceCount < 0 {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDProbe, check,
			"HardwareProfile InterfaceCount cannot be negative.")
	}
	if !p.HighestViableRung().IsValid() {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDProbe, check,
			"HardwareProfile exposes zero viable survival rungs.")
	}
	return data.AuditPass(ModuleIDProbe, check,
		"HardwareProfile OS, CPU, and viable survival rung invariants verified.")
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Machine Self-Audit & Pre-Flight Validation Tripwires
// ------------------------------------------------------------------------------

// PreflightProbe executes a 4-section non-destructive pre-flight self-audit
// covering (1) live host probe, (2) invalid manifest rejection, (3) offline
// policy capping below Rung 4, and (4) emergency floor body resolution.
//
// ### For Architects, Operators & Non-Programmers:
// Verifies FaithNet's hardware scanner works cleanly on this computer and
// obeys offline-first rules before any transport rung starts.
//
// ### For Systems Programmers & Engine Architects:
// Stack-pure 4-section pre-flight gate returning structured `data.AuditResult`.
func PreflightProbe() data.AuditResult {
	const check = "probe_preflight_4section"

	// Micro Check 1 (Always Pass): Live host body scan produces valid profile.
	live := ProbeHostBody()
	if res := AuditHardwareProfile(live); !res.IsPass() {
		return res
	}
	if !live.HasMemoryBus {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDProbe, check,
			"Sec 1 live host body failed to expose Rung 2 MemoryBus.")
	}

	// Micro Check 2 (Never Fail): Invalid manifest yields RungNone safely.
	badManifest := data.DefaultNodeManifest(data.NodeIDNull, "bad")
	if EvaluatePolicyCeiling(live, badManifest) != data.RungNone {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDProbe, check,
			"Sec 2 invalid NodeManifest was not rejected with RungNone.")
	}
	if AuditHardwareProfile(data.HardwareProfile{}).IsPass() {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDProbe, check,
			"Sec 2 empty HardwareProfile bypassed AuditHardwareProfile.")
	}

	// Micro Check 3 (Invariants & Limits): Offline manifest caps Rich body at Rung 3.
	rich := data.HardwareProfile{
		OSName: "linux", Arch: "amd64", CPUCores: 4,
		HasMemoryBus: true, HasLocalUDP: true, HasGlobalTCP: true,
	}
	offlineManifest := data.DefaultNodeManifest(0x01, "offline")
	if EvaluatePolicyCeiling(rich, offlineManifest) != data.Rung3Local {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDProbe, check,
			"Sec 3 offline-first manifest failed to cap Rung4Global at Rung3Local.")
	}

	// Micro Check 4 (Emergence): Crippled floor body resolves cleanly to Rung 1.
	floorBody := data.HardwareProfile{
		OSName: "bare-metal", Arch: "x86", CPUCores: 1, HasSerialPorts: true,
	}
	if EvaluatePolicyCeiling(floorBody, offlineManifest) != data.Rung1Floor {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDProbe, check,
			"Sec 4 emergency serial-only body failed to resolve to Rung1Floor.")
	}

	return data.AuditPass(ModuleIDProbe, check,
		"BodyProbe passed all 4 micro pre-flight gates (live/fault/cap/floor).")
}

// AuditProbe returns true iff PreflightProbe() completes with AuditSeverityPass.
//
// ### For Architects, Operators & Non-Programmers:
// Quick pass/fail shortcut confirming the hardware scanner is ready.
//
// ### For Systems Programmers & Engine Architects:
// Convenience boolean bridge delegating to `PreflightProbe().IsPass()`.
func AuditProbe() bool {
	return PreflightProbe().IsPass()
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK / EXECUTABLE PROCEDURES & LIFECYCLE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 2: Adaptive Orchestration & Ladder Control
// Module:        cog.orchestrator.probe
// Mission:       Expose epilogue readiness handshake, 60-column host capability
//                cards, and declare stateless lifecycle compliance.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Executable Procedures, Inspection, Readiness & Cards
// ------------------------------------------------------------------------------

// IsProbeReady reports whether the dynamic host body prober passed pre-flight.
//
// ### For Architects, Operators & Non-Programmers:
// Epilogue confirmation that FaithNet's hardware body scanner is healthy.
//
// ### For Systems Programmers & Engine Architects:
// Matches OmniCode's Block 5 epilogue readiness predicate (`is_*_ready`).
func IsProbeReady() bool {
	return PreflightProbe().IsPass()
}

// ExecuteProbeDumpCard renders a 60-column inner-width host capability card.
//
// ### For Architects, Operators & Non-Programmers:
// Prints a neat framed medical chart summarizing this computer's OS, CPU count,
// detected hardware organs, and highest viable FaithNet survival rung.
//
// ### For Systems Programmers & Engine Architects:
// Non-mutating diagnostic formatter consuming a value `data.HardwareProfile`.
func ExecuteProbeDumpCard(p data.HardwareProfile) string {
	return fmt.Sprintf(
		"+------------------------------------------------------------+\n"+
			"| %-58s |\n"+
			"+------------------------------------------------------------+\n"+
			"| Module:   %-48s |\n"+
			"| Host:     %-48s |\n"+
			"| TopRung:  %-48s |\n"+
			"| R1 Floor: %-48s |\n"+
			"| R2 Host:  %-48s |\n"+
			"| R3/R4:    %-48s |\n"+
			"+------------------------------------------------------------+",
		"FAITHNET HOST BODY HARDWARE PROBE CARD",
		ModuleIDProbe,
		fmt.Sprintf("%s/%s (%d cores, %d ifaces)", p.OSName, p.Arch, p.CPUCores, p.InterfaceCount),
		p.HighestViableRung().String(),
		fmt.Sprintf("serial=%t audio=%t", p.HasSerialPorts, p.HasAudioDevice),
		fmt.Sprintf("memory=%t ipc=%t", p.HasMemoryBus, p.HasLocalIPC),
		fmt.Sprintf("udp=%t wifi=%t tcp=%t", p.HasLocalUDP, p.HasWiFiHardware, p.HasGlobalTCP),
	)
}

// ------------------------------------------------------------------------------
// Sub-Assembly 5.2: Subsystem Lifecycle Initialization & Teardown
// ------------------------------------------------------------------------------
// Stateless Exemption Notice: None. (Stateless module / pure stack-allocated execution).
// ------------------------------------------------------------------------------

// End of orchestrator/probe.go
