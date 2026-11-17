// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/playtester/harness/pty_serial_bench/main.go
// @id            playtest.harness.pty_serial_bench
// @layer         Layer 4: Proving Ground (Rung 1 OS FIFO/Serial SLIP Workbench)
// @archetype     playtest_harness | operational
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Rung 1 Civilizational Floor OS Serial Stream Workbench
// @role          Real Kernel FIFO/FD SLIP Throughput, Latency & Cable-Pull Test
// @maturity      Rev 5 Baseline | Wave 7A Playtester Proving Ground
// @chassis       bounded-heap | Real Linux kernel FIFO file descriptors
// @scale         Sustained 512B Floppy Sector SLIP stream + mid-line noise resync
// @seams         in: data, engine, orchestrator, transport/rung1_floor
// @invariants    1. Every byte crosses real Linux kernel file descriptors (Mkfifo).
//                2. Tests full 512-byte Floppy Sector frames containing literal
//                   SLIPEND (0xC0) and SLIPESC (0xDB) bytes.
//                3. Injects deliberate mid-line noise/cable-pull garbage and
//                   proves RFC 1055 SLIP resyncs cleanly on the next frame.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Playtests Rung 1 Civilizational Floor `SerialTransceiver` over real Linux
//   kernel named-FIFO file descriptors (`syscall.Mkfifo` + `os.OpenFile`),
//   measuring real kernel stream throughput (KB/s), round-trip ping latency,
//   512-byte worst-case escape-byte stuffing overhead, and cable-pull resync.
//
//   ### For Architects, Operators & Non-Programmers:
//   Simulates two machines wired together with a copper serial cable using real
//   operating-system byte streams, including someone yanking/bumping the wire
//   mid-sentence to prove the next message still comes through intact.
//
//   ### For Systems Programmers & Engine Architects:
//   Uses `syscall.Mkfifo` inside `os.MkdirTemp` so zero external utilities or
//   root privileges are required; cleanly closes and unlinks temporary FIFOs.
// ==============================================================================

package main

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Operational Constants & Full-Duplex Kernel FIFO Pair Builder )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Playtester Proving Ground
// Module:        playtest.harness.pty_serial_bench
// Mission:       Create real OS kernel FIFO byte streams for SLIP playtesting.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

// --- Tier 2: External Library Dependencies (None — 100% Go Stdlib Invariant) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---
import (
	"github.com/CreativeWorkzStudio/FaithNet/data"
	"github.com/CreativeWorkzStudio/FaithNet/engine"
	"github.com/CreativeWorkzStudio/FaithNet/engine/covenant"
	"github.com/CreativeWorkzStudio/FaithNet/orchestrator"
	"github.com/CreativeWorkzStudio/FaithNet/transport/rung1_floor"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Operational Constants
// ------------------------------------------------------------------------------

const (
	ModuleIDSerialBench = "playtest.harness.pty_serial_bench"

	serialAlpha data.NodeID = 0xFA17000000005001
	serialBeta  data.NodeID = 0xFA17000000005002
)

// ==============================================================================
// BLOCK 3: BODY BLOCK
// ( Real Kernel FIFO Full-Duplex SLIP Ping, 512B Throughput & Noise Resync )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Playtester Proving Ground
// Module:        playtest.harness.pty_serial_bench
// Mission:       Measure real OS serial stream throughput and cable-pull resync.
// ==============================================================================

func openFIFOPair(path string) (rd *os.File, wr *os.File, err error) {
	if err := syscall.Mkfifo(path, 0600); err != nil {
		return nil, nil, err
	}
	// Open reader with O_RDWR on Linux so opening does not deadlock before writer.
	rd, err = os.OpenFile(path, os.O_RDWR, 0600)
	if err != nil {
		return nil, nil, err
	}
	wr, err = os.OpenFile(path, os.O_WRONLY, 0600)
	if err != nil {
		_ = rd.Close()
		return nil, nil, err
	}
	return rd, wr, nil
}

func main() {
	tmpDir, err := os.MkdirTemp("", "faithnet-slip-playtest-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "MkdirTemp: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Full-duplex pair of real Linux kernel FIFOs (A->B and B->A).
	rB, wA, err := openFIFOPair(filepath.Join(tmpDir, "a_to_b.fifo"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "openFIFO a_to_b: %v\n", err)
		os.Exit(1)
	}
	rA, wB, err := openFIFOPair(filepath.Join(tmpDir, "b_to_a.fifo"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "openFIFO b_to_a: %v\n", err)
		os.Exit(1)
	}

	sA := rung1_floor.NewSerialTransceiver(serialAlpha, rA, wA, wA)
	sB := rung1_floor.NewSerialTransceiver(serialBeta, rB, wB, wB)

	dA, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(serialAlpha, "serial-alpha"), engine.NewMonotonicClock(),
	)
	dB, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(serialBeta, "serial-beta"), engine.NewMonotonicClock(),
	)
	defer func() {
		_ = dA.Close()
		_ = dB.Close()
		_ = rA.Close()
		_ = rB.Close()
	}()

	_ = dA.Ladder().RegisterRung(sA)
	_ = dB.Ladder().RegisterRung(sB)

	rxData := make(chan int, 256)
	dB.SetDeliverHandler(func(f *data.Frame, _ covenant.Verdict) {
		if f.Opcode == data.OpData {
			rxData <- int(f.PayloadLen)
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_ = dA.Start(ctx)
	_ = dB.Start(ctx)

	fmt.Println("=== FAITHNET PLAYTESTER: RUNG 1 OS KERNEL FIFO SLIP BENCHMARK ===")

	// 1. Live RTT pings over real kernel FIFO streams.
	const pingCount = 50
	var rttTotal int64
	for i := 0; i < pingCount; i++ {
		rtt, rung, err := dA.PingPeer(ctx, serialBeta, []byte("SLIP-OS-FIFO"))
		if err != nil || rung != data.Rung1Floor {
			fmt.Fprintf(os.Stderr, "[PING FAIL] i=%d rung=%s err=%v\n", i, rung, err)
			os.Exit(1)
		}
		rttTotal += rtt
	}
	fmt.Printf("[1. OS FIFO PING] %d/%d OK (0%% loss) | avgRTT=%.1fus\n",
		pingCount, pingCount, float64(rttTotal)/float64(pingCount)/1000.0)

	// 2. Worst-case 512-byte Floppy Sector burst full of literal SLIPEND/SLIPESC bytes!
	const burstFrames = 200
	var trickyPayload [data.FloorMaxPayload]byte
	for i := range trickyPayload {
		switch i % 3 {
		case 0:
			trickyPayload[i] = rung1_floor.SLIPEND
		case 1:
			trickyPayload[i] = rung1_floor.SLIPESC
		default:
			trickyPayload[i] = byte(i & 0xFF)
		}
	}

	t0 := time.Now()
	for i := 1; i <= burstFrames; i++ {
		var f data.Frame
		f.Version = data.ProtocolVersion
		f.Rung = data.Rung1Floor
		f.Opcode = data.OpData
		f.Flags = data.FlagFloorCompact | data.FlagCovenantVerified
		f.Truth = data.QuadPositive
		f.SourceID = serialAlpha
		f.TargetID = serialBeta
		f.SeqID = uint32(i)
		f.SetPayload(trickyPayload[:])
		if _, err := dA.SendFrame(&f); err != nil {
			fmt.Fprintf(os.Stderr, "[BURST SEND FAIL] i=%d err=%v\n", i, err)
			os.Exit(1)
		}
	}
	totalPayloadBytes := 0
	for i := 0; i < burstFrames; i++ {
		select {
		case n := <-rxData:
			totalPayloadBytes += n
		case <-ctx.Done():
			fmt.Fprintf(os.Stderr, "[BURST TIMEOUT] got %d/%d frames\n", i, burstFrames)
			os.Exit(1)
		}
	}
	elapsed := time.Since(t0)
	kbps := float64(totalPayloadBytes) / 1024.0 / elapsed.Seconds()
	fmt.Printf("[2. 512B WORST-CASE ESCAPE BURST] %d/%d frames (%d payload bytes) in %.2fms | %.1f KB/s (equiv %.0f bps)\n",
		burstFrames, burstFrames, totalPayloadBytes, float64(elapsed.Microseconds())/1000.0,
		kbps, kbps*1024*8)

	// 3. Deliberate mid-line cable-pull / RF burst noise injection + SLIP resync!
	garbage := []byte{0xDE, 0xAD, 0xBE, 0xEF, rung1_floor.SLIPESC, 0x99, 0xFF, 0x00, 0x13, 0x37}
	_, _ = wA.Write(garbage)

	// Next sovereign frame MUST automatically resync via double-END framing!
	rtt, rung, err := dA.PingPeer(ctx, serialBeta, []byte("POST-GLITCH-RESYNC"))
	if err != nil || rung != data.Rung1Floor {
		fmt.Fprintf(os.Stderr, "[RESYNC FAIL] err=%v\n", err)
		os.Exit(1)
	}
	fmt.Printf("[3. CABLE-GLITCH NOISE RESYNC] injected %dB corrupt burst -> next ping RECOVERED on %s in %.1fus!\n",
		len(garbage), rung, float64(rtt)/1000.0)
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Inline Verification — Non-Zero Exit on Any Stream Loss or Resync Failure )
// ==============================================================================

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( End of Rung 1 OS Kernel FIFO SLIP Benchmark )
// ==============================================================================

// End of tests/playtester/harness/pty_serial_bench/main.go
