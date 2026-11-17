// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          cmd/faithnet-node/main.go
// @id            cmd.faithnet.node
// @layer         Layer 4: Thin Imperative Sovereign Node CLI Entrypoint
// @archetype     operational | cli_entrypoint
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Orchestrated Sovereign Node & "The Well" Daemon CLI
// @role          Thin Entrypoint: BodyProbe -> AdaptiveLadder -> Room & Well
// @maturity      Rev 5 Baseline | Wave 4 The Room & The Well
// @chassis       bounded-heap | Thin CLI orchestrating Layers 0..3B cleanly
// @scale         O(1) boot sequence + bounded background receive/HTTP goroutines
// @seams         in: data, engine, orchestrator, transport/rung2_host,
//                transport/rung3_local, interpreter
// @invariants    1. Stratum Baseline: Universal 5-Block Lifecycle Protocol.
//                2. Offline-First Default: AllowGlobalEgress defaults to false.
//                3. Graceful SIGINT/SIGTERM shutdown tears down UDP, Beacon,
//                   Dispatcher, and WellServer with zero leaked goroutines.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Implements the primary sovereign FaithNet node daemon (`cmd/faithnet-node`).
//   Probes the current host body, wires Rung 2 Host + Rung 3 Local UDP onto the
//   AdaptiveLadder, optionally switches ON the Room Beacon and The Well zero-
//   install HTML portal, and provides `--audit` and `--probe-only` inspection.
//
//   ### For Architects, Operators & Non-Programmers:
//   Lets anyone launch a live offline FaithNet node on Linux, macOS, or Windows,
//   print full hardware and ladder health cards, run every built-in pre-flight
//   safety check, or serve The Well offline webpage to nearby browsers.
//
//   ### For Systems Programmers & Engine Architects:
//   - Ownership & Reentrancy: Uses `signal.NotifyContext` so Ctrl-C or SIGTERM
//     drains HTTP, UDP, Beacon, and Dispatcher cleanly before process exit.
// ==============================================================================

package main

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( CLI Constants, Options Chassis & Logical Anchor )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Thin Imperative Sovereign Node CLI Entrypoint
// Module:        cmd.faithnet.node
// Mission:       Define node startup flags, defaults, and CLI option chassis.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// --- Tier 2: External Library Dependencies (None — 100% Go Stdlib Invariant) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---
import (
	"github.com/CreativeWorkzStudio/FaithNet/data"
	"github.com/CreativeWorkzStudio/FaithNet/engine"
	"github.com/CreativeWorkzStudio/FaithNet/engine/covenant"
	"github.com/CreativeWorkzStudio/FaithNet/interpreter"
	"github.com/CreativeWorkzStudio/FaithNet/orchestrator"
	"github.com/CreativeWorkzStudio/FaithNet/transport/rung2_host"
	"github.com/CreativeWorkzStudio/FaithNet/transport/rung3_local"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Pre-Conditions & Default Node Constants
// ------------------------------------------------------------------------------

const (
	// DefaultNodeID is the default sovereign 64-bit node address.
	DefaultNodeID uint64 = 0xFA17000000000001

	// DefaultNodeLabel is the default human-readable sovereign node label.
	DefaultNodeLabel = "FaithNet-Anchor"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.2: Operational Logical Anchor
// ------------------------------------------------------------------------------

const (
	// ModuleIDNodeCLI is the canonical logical ID for faithnet-node CLI.
	ModuleIDNodeCLI = "cmd.faithnet.node"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.3: CLI Options Configuration Chassis
// ------------------------------------------------------------------------------

// NodeCLIOptions captures operator startup flags for faithnet-node.
type NodeCLIOptions struct {
	NodeID            uint64
	Label             string
	UDPAddr           string
	WellAddr          string
	BeaconSSID        string
	EnableWell        bool
	EnableBeacon      bool
	AllowGlobalEgress bool
	AuditOnly         bool
	ProbeOnly         bool
}

// ==============================================================================
// BLOCK 3: BODY BLOCK
// ( Composable Node Boot, Diagnostic Inspection & Signal-Bounded Run Loop )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Thin Imperative Sovereign Node CLI Entrypoint
// Module:        cmd.faithnet.node
// Mission:       Boot a sovereign FaithNet node from BodyProbe -> Ladder ->
//                Dispatcher -> Room UDP / Beacon / Well Portal cleanly.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: Sovereign Node Boot & Runtime Execution
// ------------------------------------------------------------------------------

func runSovereignNode(opts NodeCLIOptions) error {
	body := orchestrator.ProbeHostBody()

	manifest := data.DefaultNodeManifest(data.NodeID(opts.NodeID), opts.Label)
	manifest.AllowGlobalEgress = opts.AllowGlobalEgress
	if opts.AllowGlobalEgress {
		manifest.PreferredMaxRung = data.Rung4Global
	}
	manifest.EnableWellPortal = opts.EnableWell
	manifest.WellListenAddr = opts.WellAddr
	manifest.BeaconSSID = opts.BeaconSSID

	clk := engine.NewMonotonicClock()
	disp, err := orchestrator.NewDispatcher(manifest, clk)
	if err != nil {
		return err
	}
	defer func() { _ = disp.Close() }()

	// Always attach Rung 2 Host RAM bus as baseline local substrate.
	memBus := rung2_host.NewMemoryBus()
	if err := disp.Ladder().RegisterRung(memBus.Attach(manifest.ID)); err != nil {
		return err
	}

	// Attach Rung 3 Local UDP transceiver when requested or available.
	var udpTx *rung3_local.UDPTransceiver
	if opts.UDPAddr != "" {
		udpTx, err = rung3_local.NewUDPTransceiver(manifest.ID, opts.UDPAddr)
		if err != nil {
			return fmt.Errorf("failed binding Rung 3 UDP at %s: %w", opts.UDPAddr, err)
		}
		if err := disp.Ladder().RegisterRung(udpTx); err != nil {
			_ = udpTx.Close()
			return err
		}
	}

	// Optional Room Beacon switch (safe simulated mode by default).
	var beacon *rung3_local.BeaconSwitch
	if opts.EnableBeacon {
		beacon, err = rung3_local.NewBeaconSwitch(
			manifest.BeaconSSID,
			rung3_local.DefaultBeaconChannel,
			rung3_local.BeaconModeSimulated,
			clk,
		)
		if err != nil {
			return err
		}
		if err := beacon.TurnOn(); err != nil {
			return err
		}
		defer func() { _ = beacon.Close() }()
	}

	// Print complete diagnostic dashboard cards on startup.
	fmt.Println(orchestrator.ExecuteProbeDumpCard(body))
	fmt.Println(disp.Ladder().ExecuteLadderDumpCard())
	fmt.Println(disp.ExecuteDispatcherDumpCard())
	if udpTx != nil {
		fmt.Println(udpTx.ExecuteUDPDumpCard())
	}
	if beacon != nil {
		fmt.Println(beacon.ExecuteBeaconDumpCard())
	}

	// Optional Well captive HTML portal.
	var well *interpreter.WellServer
	if opts.EnableWell {
		well = interpreter.NewWellServer(disp, opts.WellAddr)
		if !opts.ProbeOnly {
			if err := well.Start(); err != nil {
				return fmt.Errorf("failed starting The Well portal: %w", err)
			}
			defer func() { _ = well.Close() }()
		}
		fmt.Println(well.ExecuteWellDumpCard())
	}

	if opts.ProbeOnly {
		return nil
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := disp.Start(ctx); err != nil {
		return err
	}
	fmt.Printf("FaithNet node %q (%s) online on %s. Press Ctrl-C to stop.\n",
		manifest.Label, manifest.ID.String(), disp.Ladder().ActiveRung().String())

	<-ctx.Done()
	fmt.Println("FaithNet node shutting down cleanly...")
	return nil
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Full-Stack Layer 0..3B Pre-Flight Audit Aggregator )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Thin Imperative Sovereign Node CLI Entrypoint
// Module:        cmd.faithnet.node
// Mission:       Execute every Layer 0..3B Block 4 pre-flight probe on --audit.
// ==============================================================================

func runFullNodeAudits() bool {
	results := []data.AuditResult{
		engine.AuditDataVocabulary(),
		engine.PreflightCodec(),
		engine.PreflightClock(),
		engine.PreflightPingEngine(),
		engine.PreflightRouter(),
		covenant.PreflightCovenantFilter(),
		rung2_host.PreflightMemoryBus(),
		rung2_host.PreflightHostIPC(),
		orchestrator.PreflightProbe(),
		orchestrator.PreflightLadder(),
		orchestrator.PreflightDispatcher(),
		rung3_local.PreflightLocalUDP(),
		rung3_local.PreflightBeaconSwitch(),
		interpreter.PreflightTranslator(),
		interpreter.PreflightWellServer(),
		interpreter.PreflightLegacyBridge(),
	}
	allPass := true
	for _, r := range results {
		fmt.Println(r.String())
		if !r.IsPass() {
			allPass = false
		}
	}
	return allPass
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Imperative CLI Lifecycle Entrypoint & Exit Code Handshake )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Thin Imperative Sovereign Node CLI Entrypoint
// Module:        cmd.faithnet.node
// Mission:       Parse flags, execute pre-flight or sovereign node boot, and
//                exit with deterministic status code.
// ==============================================================================

func main() {
	var opts NodeCLIOptions
	flag.Uint64Var(&opts.NodeID, "id", DefaultNodeID, "Sovereign 64-bit NodeID")
	flag.StringVar(&opts.Label, "label", DefaultNodeLabel, "Human-readable node label")
	flag.StringVar(&opts.UDPAddr, "udp", rung3_local.DefaultLocalUDPAddr, "Rung 3 Local UDP listen address")
	flag.StringVar(&opts.WellAddr, "well-addr", data.DefaultWellListenAddr, "The Well HTML portal listen address")
	flag.StringVar(&opts.BeaconSSID, "beacon-ssid", data.DefaultBeaconSSID, "Room Beacon SSID (1..32 bytes)")
	flag.BoolVar(&opts.EnableWell, "well", false, "Enable The Well zero-install HTML captive portal")
	flag.BoolVar(&opts.EnableBeacon, "beacon", false, "Activate Room Beacon switch (safe simulated mode)")
	flag.BoolVar(&opts.AllowGlobalEgress, "allow-global", false, "Explicitly opt in to Rung 4 Global egress")
	flag.BoolVar(&opts.AuditOnly, "audit", false, "Run all Layer 0..3B Block 4 pre-flight audits and exit")
	flag.BoolVar(&opts.ProbeOnly, "probe-only", false, "Print host body, ladder & node cards and exit")
	flag.Parse()

	if opts.AuditOnly {
		if !runFullNodeAudits() {
			os.Exit(1)
		}
		os.Exit(0)
	}

	if err := runSovereignNode(opts); err != nil {
		fmt.Fprintf(os.Stderr, "faithnet-node failed: %v\n", err)
		os.Exit(1)
	}
}

// End of cmd/faithnet-node/main.go
