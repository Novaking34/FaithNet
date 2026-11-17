// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          cmd/faithnet-ping/main.go
// @id            cmd.faithnet.ping
// @layer         Layer 4: Thin Imperative CLI Entrypoint (First Breath)
// @archetype     operational | cli_entrypoint
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Standalone Offline 5-Cog Sonar Ping CLI
// @role          Thin Imperative Entrypoint: Rung 2 Memory & IPC Loopback Ping
// @maturity      Rev 5 Baseline | Wave 2 First Breath
// @chassis       bounded-heap | Thin CLI orchestrating Layer 0-3 cogs cleanly
// @scale         O(N_pings) deterministic offline heartbeat demonstration
// @seams         in: data, engine, engine/covenant, transport/rung2_host
// @invariants    1. Stratum Baseline: Universal 5-Block Lifecycle Protocol.
//                2. Zero third-party dependencies (100% Go standard library).
//                3. Operates 100% offline with zero network card required.
//                4. `--audit` runs every Layer 0-3 Block 4 pre-flight gate.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Implements Phase 2 ("First Breath") of FaithNet: a thin imperative CLI
//   proving all 5 irreducible cogs working end-to-end completely offline over
//   Rung 2 Host mediums (`--mode=memory` Go RAM channels or `--mode=ipc` local
//   Unix datagram sockets), plus a `--audit` pre-flight verification runner.
//
//   ### For Architects, Operators & Non-Programmers:
//   Lets anyone run `go run ./cmd/faithnet-ping` on an offline computer to see
//   two sovereign nodes exchange real binary FaithNet envelopes, verify tamper
//   seals through the Genesis 3 Airlock, and measure nanosecond ping latency.
//
//   ### For Systems Programmers & Engine Architects:
//   - Ownership & Reentrancy: Spawns one bounded responder goroutine cancelled
//     via `context.WithCancel` + `sync.WaitGroup` before exit (zero leaks).
//   - Explicit Non-Goal: Contains zero codec, routing, or airlock logic;
//     strictly delegates all protocol work to Layers 0..3.
// ==============================================================================

package main

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( CLI Constants, Configuration Struct & Pre-Conditional Flag Validators )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Thin Imperative CLI Entrypoint
// Module:        cmd.faithnet.ping
// Mission:       Define CLI defaults, option parsing chassis, and flag guards.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// --- Tier 2: External Library Dependencies (None — 100% Go Stdlib Invariant) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---
import (
	"github.com/CreativeWorkzStudio/FaithNet/data"
	"github.com/CreativeWorkzStudio/FaithNet/engine"
	"github.com/CreativeWorkzStudio/FaithNet/engine/covenant"
	"github.com/CreativeWorkzStudio/FaithNet/transport"
	"github.com/CreativeWorkzStudio/FaithNet/transport/rung2_host"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Pre-Conditions, Physical Bounds & Default CLI Limits
// ------------------------------------------------------------------------------

const (
	// DefaultPingCount is the default number of offline sonar pulses emitted.
	DefaultPingCount = 4

	// MaxPingCount caps CLI ping loops to a bounded operator ceiling.
	MaxPingCount = 10_000

	// DefaultPingPayload is the default sovereign echo string carried on wire.
	DefaultPingPayload = "FAITHNET-FIRST-BREATH"

	// DefaultNodeAlphaID is the sovereign initiator ID for local loopback CLI.
	DefaultNodeAlphaID data.NodeID = 0xFA17000000000001

	// DefaultNodeBetaID is the sovereign reflector ID for local loopback CLI.
	DefaultNodeBetaID data.NodeID = 0xFA17000000000002
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.2: Operational Logical Anchor
// ------------------------------------------------------------------------------

const (
	// ModuleIDPingCLI is the canonical logical ID for faithnet-ping CLI.
	ModuleIDPingCLI = "cmd.faithnet.ping"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.3: CLI Options Configuration Chassis
// ------------------------------------------------------------------------------

// PingCLIOptions holds validated command-line options for faithnet-ping.
type PingCLIOptions struct {
	Mode      string // Transport mode: "memory" or "ipc"
	Count     int    // Number of ping pulses to exchange (1..MaxPingCount)
	Payload   string // Echo payload text (<= MaxPingEchoBytes)
	AuditOnly bool   // Run all Block 4 pre-flight audits and exit
	ShowCards bool   // Print 60-column diagnostic cards after ping run
}

// ------------------------------------------------------------------------------
// Sub-Assembly 2.4: Pre-Conditional Helper Cogs & Option Validators
// ------------------------------------------------------------------------------

// IsValidMode reports whether mode names a supported Rung 2 transport.
//
// ### For Architects, Operators & Non-Programmers:
// Checks that the operator selected either in-memory (`memory`) or local IPC
// (`ipc`) mode.
//
// ### For Systems Programmers & Engine Architects:
// O(1) whitelist validation preventing unhandled transport dispatch paths.
func IsValidMode(mode string) bool {
	return mode == "memory" || mode == "ipc"
}

// ==============================================================================
// BLOCK 3: BODY BLOCK
// ( Composable CLI Workflows: Reflector Loop & Sovereign Ping Exchange )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Thin Imperative CLI Entrypoint
// Module:        cmd.faithnet.ping
// Mission:       Wire Rung 2 transceivers, Genesis 3 Airlock, Codec, and
//                PingEngine into a live offline 5-Cog loopback exchange.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: Sovereign Echo Reflector Worker
// ------------------------------------------------------------------------------

// runReflector listens on rx, verifies each arriving frame via Codec + Genesis
// 3 Airlock, and reflects OpPing frames back to their sovereign sender.
func runReflector(
	ctx context.Context,
	rx transport.Transceiver,
	selfID data.NodeID,
	airlock *covenant.Filter,
) {
	var in, out data.Frame
	var wire [data.FloorMaxFrameSize]byte

	for {
		env, err := rx.Receive(ctx)
		if err != nil {
			return
		}
		if err := engine.DecodeFrom(env.Slice(), &in); err != nil {
			continue
		}
		if v := airlock.EvaluateFrame(&in); !v.IsPermitted() {
			continue
		}
		if in.Opcode != data.OpPing {
			continue
		}
		if err := engine.ReflectPingInto(&in, &out, selfID); err != nil {
			continue
		}
		n, err := engine.EncodeInto(wire[:], &out)
		if err != nil {
			continue
		}
		_ = rx.Send(out.TargetID, wire[:n])
	}
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: End-to-End Offline Loopback Ping Session
// ------------------------------------------------------------------------------

// executePingSession runs opts.Count sovereign Ping -> Wire -> Reflect -> Pong
// exchanges over the selected Rung 2 transport and prints telemetry.
func executePingSession(opts PingCLIOptions) error {
	var txA, txB transport.Transceiver
	var cleanup func()

	switch opts.Mode {
	case "memory":
		bus := rung2_host.NewMemoryBus()
		a := bus.Attach(DefaultNodeAlphaID)
		b := bus.Attach(DefaultNodeBetaID)
		txA, txB = a, b
		cleanup = func() {
			_ = a.Close()
			_ = b.Close()
		}
	case "ipc":
		dir, err := os.MkdirTemp("", "faithnet-ping-ipc-*")
		if err != nil {
			return err
		}
		a, err := rung2_host.NewIPCTransceiver(DefaultNodeAlphaID, filepath.Join(dir, "a.sock"))
		if err != nil {
			_ = os.RemoveAll(dir)
			return err
		}
		b, err := rung2_host.NewIPCTransceiver(DefaultNodeBetaID, filepath.Join(dir, "b.sock"))
		if err != nil {
			_ = a.Close()
			_ = os.RemoveAll(dir)
			return err
		}
		a.RegisterPeer(DefaultNodeBetaID, b.SocketPath())
		b.RegisterPeer(DefaultNodeAlphaID, a.SocketPath())
		txA, txB = a, b
		cleanup = func() {
			_ = a.Close()
			_ = b.Close()
			_ = os.RemoveAll(dir)
		}
	default:
		return fmt.Errorf("unsupported mode %q (use memory or ipc)", opts.Mode)
	}
	defer cleanup()

	clk := engine.NewMonotonicClock()
	pe := engine.NewPingEngine(DefaultNodeAlphaID, clk)
	airlock := covenant.NewFilter()

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		runReflector(ctx, txB, DefaultNodeBetaID, airlock)
	}()

	fmt.Printf("FAITHNET PING %s (%s | %s) payload=%d bytes\n",
		DefaultNodeBetaID.String(), txA.Name(), txA.Rung().String(), len(opts.Payload))

	var req, resp data.Frame
	var wire [data.FloorMaxFrameSize]byte

	for i := 0; i < opts.Count; i++ {
		if err := pe.BuildPingInto(&req, DefaultNodeBetaID, txA.Rung(), []byte(opts.Payload)); err != nil {
			cancel()
			wg.Wait()
			return err
		}
		n, err := engine.EncodeInto(wire[:], &req)
		if err != nil {
			cancel()
			wg.Wait()
			return err
		}
		if err := txA.Send(DefaultNodeBetaID, wire[:n]); err != nil {
			cancel()
			wg.Wait()
			return err
		}

		recvCtx, recvCancel := context.WithTimeout(ctx, 2*time.Second)
		env, err := txA.Receive(recvCtx)
		recvCancel()
		if err != nil {
			cancel()
			wg.Wait()
			return fmt.Errorf("ping seq=%d receive failed: %w", req.SeqID, err)
		}
		if err := engine.DecodeFrom(env.Slice(), &resp); err != nil {
			cancel()
			wg.Wait()
			return err
		}
		rtt, ok := pe.ProcessPong(&resp)
		if !ok {
			cancel()
			wg.Wait()
			return fmt.Errorf("ping seq=%d failed pong correlation", req.SeqID)
		}
		fmt.Printf("  seq=%d bytes=%d rung=%s truth=%s rtt=%d ns (%.3f us)\n",
			resp.SeqID, env.Length, resp.Rung.String(), resp.Truth.String(),
			rtt, float64(rtt)/1000.0)
	}

	cancel()
	wg.Wait()

	h := pe.HealthSnapshot()
	fmt.Printf("--- %s faithnet-ping summary ---\n", DefaultNodeBetaID.String())
	fmt.Printf("%d sent, %d received, %d%% loss | rtt min/avg/max = %d/%d/%d ns\n",
		h.PingsSent, h.PongsRecv, h.LossPercent(),
		h.MinRTTNano, h.AvgRTTNano, h.MaxRTTNano)

	if opts.ShowCards {
		fmt.Println(pe.ExecutePingDumpCard())
		fmt.Println(engine.ExecuteCodecDumpCard())
		fmt.Println(airlock.ExecuteAirlockDumpCard())
	}
	return nil
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( CLI Option Contract Validation & Aggregated Pre-Flight Gate Runner )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Thin Imperative CLI Entrypoint
// Module:        cmd.faithnet.ping
// Mission:       Validate PingCLIOptions contracts, run a 4-section CLI
//                pre-flight audit, and aggregate Layer 0..3 Block 4 probes.
// @triggers
//   - Block 4 Audits : ValidatePingCLIOptions(), PreflightPingCLI(),
//                      runPreflightAudits()
//   - Layer 4 Tests  : tests/functional/cmd_functional_test.go,
//                      tests/unit/cmd/ping_cli_test.go,
//                      tests/system/cli_binaries_system_test.go
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Contract, Input & CLI Option Verification Cogs
// ------------------------------------------------------------------------------

// ValidatePingCLIOptions audits CLI flags against FaithNet mode, count, and
// 472-byte floor echo payload bounds prior to launching transports.
//
// ### For Architects, Operators & Non-Programmers:
// Checks command-line options up front so invalid modes or oversized payloads
// fail cleanly with a clear explanation instead of crashing mid-run.
//
// ### For Systems Programmers & Engine Architects:
// Pure O(1) contract gate returning structured `data.AuditResult`.
func ValidatePingCLIOptions(opts PingCLIOptions) data.AuditResult {
	const check = "cli_ping_options_contract"

	if !IsValidMode(opts.Mode) {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDPingCLI, check,
			"CLI --mode must be either 'memory' or 'ipc'.")
	}
	if opts.Count < 1 || opts.Count > MaxPingCount {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDPingCLI, check,
			"CLI --count lies outside [1..MaxPingCount].")
	}
	if len(opts.Payload) > engine.MaxPingEchoBytes {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDPingCLI, check,
			"CLI --payload exceeds 472-byte Floppy Sector echo ceiling.")
	}
	return data.AuditPass(ModuleIDPingCLI, check,
		"PingCLIOptions mode, count, and floor payload bounds verified.")
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Machine Self-Audit & Aggregated Pre-Flight Sequence
// ------------------------------------------------------------------------------

// PreflightPingCLI runs a 4-section self-audit verifying (1) default options,
// (2) invalid mode/count rejection, (3) 472-byte echo ceiling, and (4) clean
// Layer 0..3 pre-flight readiness.
//
// ### For Architects, Operators & Non-Programmers:
// Confirms the faithnet-ping CLI validates flags accurately before running.
//
// ### For Systems Programmers & Engine Architects:
// Non-destructive 4-section CLI pre-flight returning structured `data.AuditResult`.
func PreflightPingCLI() data.AuditResult {
	const check = "cli_ping_preflight_4section"

	// Micro Check 1 (Always Pass): Canonical default options pass cleanly.
	def := PingCLIOptions{Mode: "memory", Count: DefaultPingCount, Payload: DefaultPingPayload}
	if r := ValidatePingCLIOptions(def); !r.IsPass() {
		return r
	}
	// Micro Check 2 (Never Fail): Invalid mode & zero count rejected.
	if ValidatePingCLIOptions(PingCLIOptions{Mode: "bogus", Count: 1}).IsPass() ||
		ValidatePingCLIOptions(PingCLIOptions{Mode: "memory", Count: 0}).IsPass() {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDPingCLI, check,
			"Sec 2 invalid mode or zero count bypassed ValidatePingCLIOptions.")
	}
	// Micro Check 3 (Invariants & Limits): 473B payload rejected at 472B ceiling.
	over := make([]byte, engine.MaxPingEchoBytes+1)
	if ValidatePingCLIOptions(PingCLIOptions{Mode: "ipc", Count: 1, Payload: string(over)}).IsPass() {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDPingCLI, check,
			"Sec 3 473B payload bypassed 472B echo ceiling.")
	}
	// Micro Check 4 (Emergence): Underlying Layer 0/1/2 pre-flight readiness.
	if !engine.IsCodecReady() || !engine.IsPingEngineReady() || !rung2_host.IsMemoryBusReady() {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDPingCLI, check,
			"Sec 4 underlying Layer 1/3A readiness probe failed.")
	}
	return data.AuditPass(ModuleIDPingCLI, check,
		"faithnet-ping CLI passed all 4 micro pre-flight gates.")
}

// runPreflightAudits runs every implemented Block 4 pre-flight gate and returns
// false if any subsystem trips a warning, critical, or fatal audit verdict.
func runPreflightAudits() bool {
	results := []data.AuditResult{
		PreflightPingCLI(),
		engine.AuditDataVocabulary(),
		engine.PreflightCodec(),
		engine.PreflightClock(),
		engine.PreflightPingEngine(),
		engine.PreflightRouter(),
		covenant.PreflightCovenantFilter(),
		rung2_host.PreflightMemoryBus(),
		rung2_host.PreflightHostIPC(),
	}

	allPass := true
	for _, res := range results {
		fmt.Println(res.String())
		if !res.IsPass() {
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
// Stratum:       Layer 4: Thin Imperative CLI Entrypoint
// Module:        cmd.faithnet.ping
// Mission:       Parse CLI flags, execute pre-flight or ping workflow, and
//                exit 0 on clean completion or non-zero on error/audit fail.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Main Process Lifecycle Entrypoint
// ------------------------------------------------------------------------------

func main() {
	var opts PingCLIOptions
	flag.StringVar(&opts.Mode, "mode", "memory", "Rung 2 transport mode: memory or ipc")
	flag.IntVar(&opts.Count, "count", DefaultPingCount, "Number of sovereign pings to exchange")
	flag.StringVar(&opts.Payload, "payload", DefaultPingPayload, "Echo payload string (<= 472 bytes)")
	flag.BoolVar(&opts.AuditOnly, "audit", false, "Run all Block 4 pre-flight audits and exit")
	flag.BoolVar(&opts.ShowCards, "cards", true, "Print 60-column diagnostic cards after ping")
	flag.Parse()

	if opts.AuditOnly {
		if !runPreflightAudits() {
			os.Exit(1)
		}
		fmt.Println(engine.ExecuteCodecDumpCard())
		os.Exit(0)
	}

	if !IsValidMode(opts.Mode) {
		fmt.Fprintf(os.Stderr, "error: invalid --mode=%q (must be memory or ipc)\n", opts.Mode)
		os.Exit(2)
	}
	if opts.Count < 1 || opts.Count > MaxPingCount {
		fmt.Fprintf(os.Stderr, "error: --count must be in [1..%d]\n", MaxPingCount)
		os.Exit(2)
	}
	if len(opts.Payload) > engine.MaxPingEchoBytes {
		fmt.Fprintf(os.Stderr, "error: --payload exceeds %d-byte floor echo cap\n", engine.MaxPingEchoBytes)
		os.Exit(2)
	}

	if err := executePingSession(opts); err != nil {
		fmt.Fprintf(os.Stderr, "faithnet-ping failed: %v\n", err)
		os.Exit(1)
	}
}

// End of cmd/faithnet-ping/main.go
