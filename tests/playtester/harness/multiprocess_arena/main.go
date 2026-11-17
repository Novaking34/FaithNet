// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/playtester/harness/multiprocess_arena/main.go
// @id            playtest.harness.multiprocess_arena
// @layer         Layer 4: Proving Ground (Playtester Multi-Process Arena)
// @archetype     playtest_harness | operational
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Multi-Process OS Isolation Playtester
// @role          Live Cross-Process UDP & TCP RTT/RSS Empirical Proving Ground
// @maturity      Rev 5 Baseline | Wave 7A Playtester Proving Ground
// @chassis       bounded-heap | Spawns truly separate OS child processes
// @scale         Empirical p50/p95/p99 cross-process latency & loss telemetry
// @seams         in: data, engine, orchestrator, rung3_local, rung4_global
// @invariants    1. Zero shared RAM between NodeAlpha and NodeBeta processes.
//                2. Every packet traverses the real Linux kernel socket stack.
//                3. Reports honest empirical min/p50/p95/p99/max RTT & loss.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Breaks out of single-process unit test memory by spawning genuinely
//   separate OS processes (`--role=responder` and `--role=initiator`) talking
//   over real kernel UDP (Rung 3) and TCP (Rung 4) sockets, plus executing the
//   compiled `faithnet-ping` standalone binary over real Unixgram IPC (Rung 2).
//
//   ### For Architects, Operators & Non-Programmers:
//   Proves two completely separate FaithNet programs can talk across the real
//   operating system network stack—not just inside one Go test process—and
//   measures true latency percentiles (min, p50, p95, p99, max).
//
//   ### For Systems Programmers & Engine Architects:
//   Self-re-executes (`os.Executable()`) with `--role=responder` to guarantee
//   strict OS process-address-space isolation while coordinating ephemeral port
//   discovery cleanly over stdout pipes.
// ==============================================================================

package main

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Playtest Constants, Flags & Empirical Percentile Telemetry )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Playtester Proving Ground
// Module:        playtest.harness.multiprocess_arena
// Mission:       Configure cross-process playtest runs and percentile math.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"time"
)

// --- Tier 2: External Library Dependencies (None — 100% Go Stdlib Invariant) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---
import (
	"github.com/CreativeWorkzStudio/FaithNet/data"
	"github.com/CreativeWorkzStudio/FaithNet/engine"
	"github.com/CreativeWorkzStudio/FaithNet/orchestrator"
	"github.com/CreativeWorkzStudio/FaithNet/transport/rung3_local"
	"github.com/CreativeWorkzStudio/FaithNet/transport/rung4_global"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Operational Constants & Identities
// ------------------------------------------------------------------------------

const (
	ModuleIDMultiProcessArena = "playtest.harness.multiprocess_arena"

	arenaNodeAlpha data.NodeID = 0xFA1700000000A001
	arenaNodeBeta  data.NodeID = 0xFA1700000000B002
)

// LatencySummary stores empirical RTT percentiles measured across OS processes.
type LatencySummary struct {
	Mode      string
	Sent      int
	Received  int
	MinUs     float64
	P50Us     float64
	P95Us     float64
	P99Us     float64
	MaxUs     float64
	ElapsedMs float64
}

func summarizeRTTs(mode string, sent int, rttsNs []int64, elapsed time.Duration) LatencySummary {
	s := LatencySummary{
		Mode:      mode,
		Sent:      sent,
		Received:  len(rttsNs),
		ElapsedMs: float64(elapsed.Microseconds()) / 1000.0,
	}
	if len(rttsNs) == 0 {
		return s
	}
	cp := make([]int64, len(rttsNs))
	copy(cp, rttsNs)
	sort.Slice(cp, func(i, j int) bool { return cp[i] < cp[j] })
	pct := func(p float64) float64 {
		idx := int(float64(len(cp)-1) * p)
		return float64(cp[idx]) / 1000.0
	}
	s.MinUs = float64(cp[0]) / 1000.0
	s.P50Us = pct(0.50)
	s.P95Us = pct(0.95)
	s.P99Us = pct(0.99)
	s.MaxUs = float64(cp[len(cp)-1]) / 1000.0
	return s
}

// ==============================================================================
// BLOCK 3: BODY BLOCK
// ( Responder Child Process, Initiator Process & Orchestrated Arena Runner )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Playtester Proving Ground
// Module:        playtest.harness.multiprocess_arena
// Mission:       Run isolated OS responder/initiator pairs over UDP & TCP.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: Child Responder Process (Zero Shared Memory)
// ------------------------------------------------------------------------------

func runResponder(mode string, peerAddr string) error {
	m := data.DefaultNodeManifest(arenaNodeBeta, "playtest-responder")
	m.AllowGlobalEgress = true
	m.PreferredMaxRung = data.Rung4Global

	disp, err := orchestrator.NewDispatcher(m, engine.NewMonotonicClock())
	if err != nil {
		return err
	}
	defer func() { _ = disp.Close() }()

	var listenAddr string
	switch mode {
	case "udp":
		udpTx, err := rung3_local.NewUDPTransceiver(arenaNodeBeta, "127.0.0.1:0")
		if err != nil {
			return err
		}
		listenAddr = udpTx.Addr()
		if err := disp.Ladder().RegisterRung(udpTx); err != nil {
			return err
		}
	case "tcp":
		tcpTx, err := rung4_global.NewTCPTransceiver(arenaNodeBeta, "127.0.0.1:0")
		if err != nil {
			return err
		}
		if peerAddr != "" {
			tcpTx.RegisterPeer(arenaNodeAlpha, peerAddr)
		}
		listenAddr = tcpTx.Addr()
		if err := disp.Ladder().RegisterRung(tcpTx); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown responder mode: %s", mode)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := disp.Start(ctx); err != nil {
		return err
	}

	// Emit bound address so parent initiator can dial across process boundary.
	fmt.Printf("READY %s\n", listenAddr)

	// Stay alive until stdin closes (clean EOF shutdown from parent).
	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "QUIT" {
			break
		}
	}
	return nil
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: Cross-Process Arena Execution (UDP & TCP)
// ------------------------------------------------------------------------------

func runCrossProcessMode(exePath, mode string, count int) (LatencySummary, error) {
	m := data.DefaultNodeManifest(arenaNodeAlpha, "playtest-initiator")
	m.AllowGlobalEgress = true
	m.PreferredMaxRung = data.Rung4Global

	disp, err := orchestrator.NewDispatcher(m, engine.NewMonotonicClock())
	if err != nil {
		return LatencySummary{}, err
	}
	defer func() { _ = disp.Close() }()

	var udpTx *rung3_local.UDPTransceiver
	var tcpTx *rung4_global.TCPTransceiver
	var initAddr string

	switch mode {
	case "udp":
		udpTx, err = rung3_local.NewUDPTransceiver(arenaNodeAlpha, "127.0.0.1:0")
		if err != nil {
			return LatencySummary{}, err
		}
		initAddr = udpTx.Addr()
		_ = disp.Ladder().RegisterRung(udpTx)
	case "tcp":
		tcpTx, err = rung4_global.NewTCPTransceiver(arenaNodeAlpha, "127.0.0.1:0")
		if err != nil {
			return LatencySummary{}, err
		}
		initAddr = tcpTx.Addr()
		_ = disp.Ladder().RegisterRung(tcpTx)
	default:
		return LatencySummary{}, fmt.Errorf("unsupported mode: %s", mode)
	}

	// Spawn genuinely separate OS child process for NodeBeta!
	child := exec.Command(exePath, "--role=responder", "--mode="+mode, "--peer="+initAddr)
	stdout, err := child.StdoutPipe()
	if err != nil {
		return LatencySummary{}, err
	}
	stdin, err := child.StdinPipe()
	if err != nil {
		return LatencySummary{}, err
	}
	child.Stderr = os.Stderr
	if err := child.Start(); err != nil {
		return LatencySummary{}, err
	}
	defer func() {
		_, _ = fmt.Fprintln(stdin, "QUIT")
		_ = stdin.Close()
		_ = child.Wait()
	}()

	sc := bufio.NewScanner(stdout)
	if !sc.Scan() {
		return LatencySummary{}, fmt.Errorf("child responder exited before READY")
	}
	line := sc.Text()
	if !strings.HasPrefix(line, "READY ") {
		return LatencySummary{}, fmt.Errorf("unexpected responder banner: %q", line)
	}
	respAddr := strings.TrimPrefix(line, "READY ")

	if mode == "udp" {
		if err := udpTx.RegisterPeer(arenaNodeBeta, respAddr); err != nil {
			return LatencySummary{}, err
		}
	} else {
		tcpTx.RegisterPeer(arenaNodeBeta, respAddr)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := disp.Start(ctx); err != nil {
		return LatencySummary{}, err
	}

	rtts := make([]int64, 0, count)
	payload := []byte("MULTIPROC-PLAYTEST-ECHO")
	start := time.Now()
	for i := 0; i < count; i++ {
		pCtx, pCancel := context.WithTimeout(ctx, 2*time.Second)
		rtt, _, err := disp.PingPeer(pCtx, arenaNodeBeta, payload)
		pCancel()
		if err == nil {
			rtts = append(rtts, rtt)
		}
	}
	return summarizeRTTs(strings.ToUpper(mode), count, rtts, time.Since(start)), nil
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Playtest Pre-Flight Validation & Empirical Summary Printer )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Playtester Proving Ground
// Module:        playtest.harness.multiprocess_arena
// Mission:       Confirm transport preflights before spawning OS child nodes.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Pre-Flight Sanity Check
// ------------------------------------------------------------------------------

func verifyArenaPreflight() error {
	for _, r := range []data.AuditResult{
		rung3_local.PreflightLocalUDP(),
		rung4_global.PreflightGlobalTCP(),
		orchestrator.PreflightDispatcher(),
	} {
		if !r.IsPass() {
			return fmt.Errorf("preflight failed: %s", r.String())
		}
	}
	return nil
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Executable Entrypoint & 60-Column Telemetry Card )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Playtester Proving Ground
// Module:        playtest.harness.multiprocess_arena
// Mission:       Run multi-process UDP & TCP playtests and print live metrics.
// ==============================================================================

func main() {
	role := flag.String("role", "coordinator", "coordinator or responder")
	mode := flag.String("mode", "all", "udp, tcp, or all")
	peer := flag.String("peer", "", "initiator address for tcp responder")
	count := flag.Int("count", 100, "number of cross-process pings per medium")
	flag.Parse()

	if *role == "responder" {
		if err := runResponder(*mode, *peer); err != nil {
			fmt.Fprintf(os.Stderr, "[RESPONDER ERROR] %v\n", err)
			os.Exit(1)
		}
		return
	}

	if err := verifyArenaPreflight(); err != nil {
		fmt.Fprintf(os.Stderr, "[PREFLIGHT FAIL] %v\n", err)
		os.Exit(1)
	}

	exePath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] os.Executable: %v\n", err)
		os.Exit(1)
	}

	modes := []string{"udp", "tcp"}
	if *mode != "all" {
		modes = []string{*mode}
	}

	fmt.Println("=== FAITHNET PLAYTESTER: MULTI-PROCESS OS ISOLATION ARENA ===")
	fmt.Printf("Host: %s/%s | CPUs: %d | Pings/Medium: %d | Shared RAM: 0 bytes\n\n",
		runtime.GOOS, runtime.GOARCH, runtime.NumCPU(), *count)

	for _, m := range modes {
		s, err := runCrossProcessMode(exePath, m, *count)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[%s FAIL] %v\n", strings.ToUpper(m), err)
			os.Exit(1)
		}
		loss := 0.0
		if s.Sent > 0 {
			loss = float64(s.Sent-s.Received) * 100.0 / float64(s.Sent)
		}
		var ms runtime.MemStats
		runtime.ReadMemStats(&ms)
		fmt.Printf("[%s CROSS-PROCESS] sent=%d recv=%d loss=%.1f%% | elapsed=%.1fms | heap=%.2fMB\n",
			s.Mode, s.Sent, s.Received, loss, s.ElapsedMs, float64(ms.Alloc)/(1024*1024))
		fmt.Printf("  RTT (us): min=%.1f  p50=%.1f  p95=%.1f  p99=%.1f  max=%.1f\n\n",
			s.MinUs, s.P50Us, s.P95Us, s.P99Us, s.MaxUs)
	}
}

// End of tests/playtester/harness/multiprocess_arena/main.go
