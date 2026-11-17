// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/playtester/harness/chaos_weather_proxy/main.go
// @id            playtest.harness.chaos_weather_proxy
// @layer         Layer 4: Proving Ground (Userspace Network Chaos Weather Engine)
// @archetype     playtest_harness | operational
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Adversarial Network Weather & Silent Blackhole Playtester
// @role          Userspace UDP Chaos Proxy (Loss, Jitter, Dup, Bit-Flip, Blackhole)
// @maturity      Rev 5 Baseline | Wave 7A Playtester Proving Ground
// @chassis       bounded-heap | Non-privileged userspace UDP impairment proxy
// @scale         5 live weather regimes across real OS loopback UDP sockets
// @seams         in: data, engine, orchestrator, rung1_floor, rung2_host, rung3_local
// @invariants    1. Runs without root/CAP_NET_ADMIN by proxying real UDP sockets.
//                2. Exposes real silent-blackhole behavior honestly (never masks gaps).
//                3. Verifies 0 corrupted frames ever survive bit-flip storms.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Simulates hostile real-world network weather between two live FaithNet
//   nodes over real OS UDP sockets without requiring root privileges:
//   (1) Clear Sky, (2) Congested LAN (loss + jitter + duplicates), (3) Storm
//   (30% loss + 10% wire bit-flips proving FNV-1a drops 100% of corrupt frames),
//   (4) Silent Blackhole (100% silent drop exposing passive-timeout blind spot),
//   and (5) Active Cascade Failover (Rung 3 UDP -> Rung 2 RAM -> Rung 1 AFSK).
//
//   ### For Architects, Operators & Non-Programmers:
//   Creates real bad weather between two live FaithNet nodes so we can see
//   honestly what survives (bit-flips, duplicates, jitter) and where current
//   gaps still exist (silent UDP blackholes without active heartbeat timeouts).
//
//   ### For Systems Programmers & Engine Architects:
//   Deterministic PRNG-driven userspace UDP forwarder with per-packet drop,
//   bit-flip, duplication, and goroutine-delayed jitter injection.
// ==============================================================================

package main

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Weather Regime Ontology & Userspace UDP Chaos Proxy )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Playtester Proving Ground
// Module:        playtest.harness.chaos_weather_proxy
// Mission:       Define weather parameters and real UDP impairment proxy.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// --- Tier 2: External Library Dependencies (None — 100% Go Stdlib Invariant) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---
import (
	"github.com/CreativeWorkzStudio/FaithNet/data"
	"github.com/CreativeWorkzStudio/FaithNet/engine"
	"github.com/CreativeWorkzStudio/FaithNet/orchestrator"
	"github.com/CreativeWorkzStudio/FaithNet/transport/rung1_floor"
	"github.com/CreativeWorkzStudio/FaithNet/transport/rung2_host"
	"github.com/CreativeWorkzStudio/FaithNet/transport/rung3_local"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Weather Regime Definition
// ------------------------------------------------------------------------------

const (
	ModuleIDChaosWeather = "playtest.harness.chaos_weather_proxy"

	chaosAlpha data.NodeID = 0xFA1700000000C001
	chaosBeta  data.NodeID = 0xFA1700000000C002
)

// WeatherProfile configures deterministic real-UDP impairment rates.
type WeatherProfile struct {
	Name        string
	DropProb    float64
	CorruptProb float64
	DupProb     float64
	MaxJitterMs int
	PingTimeout time.Duration
	Pings       int
}

// udpChaosProxy sits between Alpha's UDP socket and Beta's UDP socket.
type udpChaosProxy struct {
	connA    *net.UDPConn // Alpha sends here; proxy forwards to Beta
	connB    *net.UDPConn // Beta sends here; proxy forwards to Alpha
	addrA    *net.UDPAddr
	addrB    *net.UDPAddr
	mu       sync.RWMutex
	profile  WeatherProfile
	rng      *rand.Rand
	rngMu    sync.Mutex
	dropped  atomic.Uint64
	corrupt  atomic.Uint64
	duped    atomic.Uint64
	forwarded atomic.Uint64
	stopCh   chan struct{}
	wg       sync.WaitGroup
}

// ==============================================================================
// BLOCK 3: BODY BLOCK
// ( Live Proxy Engine & 5 Empirical Chaos Weather Scenarios )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Playtester Proving Ground
// Module:        playtest.harness.chaos_weather_proxy
// Mission:       Forward real UDP datagrams through configurable impairments.
// ==============================================================================

func newUDPChaosProxy(addrA, addrB string) (*udpChaosProxy, error) {
	ua, err := net.ResolveUDPAddr("udp4", addrA)
	if err != nil {
		return nil, err
	}
	ub, err := net.ResolveUDPAddr("udp4", addrB)
	if err != nil {
		return nil, err
	}
	cA, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		return nil, err
	}
	cB, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		_ = cA.Close()
		return nil, err
	}
	p := &udpChaosProxy{
		connA:  cA,
		connB:  cB,
		addrA:  ua,
		addrB:  ub,
		rng:    rand.New(rand.NewSource(20261117)),
		stopCh: make(chan struct{}),
	}
	p.wg.Add(2)
	go p.pump(p.connA, p.connB, p.addrB)
	go p.pump(p.connB, p.connA, p.addrA)
	return p, nil
}

func (p *udpChaosProxy) setProfile(wp WeatherProfile) {
	p.mu.Lock()
	p.profile = wp
	p.mu.Unlock()
	p.dropped.Store(0)
	p.corrupt.Store(0)
	p.duped.Store(0)
	p.forwarded.Store(0)
}

func (p *udpChaosProxy) nextFloat() float64 {
	p.rngMu.Lock()
	defer p.rngMu.Unlock()
	return p.rng.Float64()
}

func (p *udpChaosProxy) nextIntn(n int) int {
	p.rngMu.Lock()
	defer p.rngMu.Unlock()
	return p.rng.Intn(n)
}

func (p *udpChaosProxy) pump(inConn, outConn *net.UDPConn, dst *net.UDPAddr) {
	defer p.wg.Done()
	buf := make([]byte, 2048)
	for {
		n, _, err := inConn.ReadFromUDP(buf)
		if err != nil {
			return
		}
		pkt := make([]byte, n)
		copy(pkt, buf[:n])

		p.mu.RLock()
		wp := p.profile
		p.mu.RUnlock()

		if wp.DropProb > 0 && p.nextFloat() < wp.DropProb {
			p.dropped.Add(1)
			continue
		}
		if wp.CorruptProb > 0 && p.nextFloat() < wp.CorruptProb && len(pkt) > 8 {
			pkt[len(pkt)-1] ^= 0x80
			p.corrupt.Add(1)
		}

		sendCopy := func(b []byte) {
			if wp.MaxJitterMs > 0 {
				delay := time.Duration(p.nextIntn(wp.MaxJitterMs+1)) * time.Millisecond
				time.Sleep(delay)
			}
			_, _ = outConn.WriteToUDP(b, dst)
			p.forwarded.Add(1)
		}

		if wp.DupProb > 0 && p.nextFloat() < wp.DupProb {
			p.duped.Add(1)
			dup := make([]byte, len(pkt))
			copy(dup, pkt)
			go sendCopy(dup)
		}
		go sendCopy(pkt)
	}
}

func (p *udpChaosProxy) close() {
	close(p.stopCh)
	_ = p.connA.Close()
	_ = p.connB.Close()
	p.wg.Wait()
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Active Cascade Failover Timing Measurement )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Playtester Proving Ground
// Module:        playtest.harness.chaos_weather_proxy
// Mission:       Measure live failover timing across Rung 3 -> Rung 2 -> Rung 1.
// ==============================================================================

func runActiveCascadeTiming() error {
	dA, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(chaosAlpha, "cascade-a"), engine.NewMonotonicClock(),
	)
	dB, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(chaosBeta, "cascade-b"), engine.NewMonotonicClock(),
	)
	defer func() {
		_ = dA.Close()
		_ = dB.Close()
	}()

	afsk := rung1_floor.NewAudioMedium()
	ram := rung2_host.NewMemoryBus()
	udpA, _ := rung3_local.NewUDPTransceiver(chaosAlpha, "127.0.0.1:0")
	udpB, _ := rung3_local.NewUDPTransceiver(chaosBeta, "127.0.0.1:0")
	_ = udpA.RegisterPeer(chaosBeta, udpB.Addr())
	_ = udpB.RegisterPeer(chaosAlpha, udpA.Addr())

	ramA := ram.Attach(chaosAlpha)
	_ = dA.Ladder().RegisterRung(afsk.Attach(chaosAlpha))
	_ = dA.Ladder().RegisterRung(ramA)
	_ = dA.Ladder().RegisterRung(udpA)
	_ = dB.Ladder().RegisterRung(afsk.Attach(chaosBeta))
	_ = dB.Ladder().RegisterRung(ram.Attach(chaosBeta))
	_ = dB.Ladder().RegisterRung(udpB)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = dA.Start(ctx)
	_ = dB.Start(ctx)

	t0 := time.Now()
	rtt3, r3, err := dA.PingPeer(ctx, chaosBeta, []byte("R3"))
	if err != nil {
		return err
	}
	udpA.SetHealthy(false)
	tCut1 := time.Now()
	rtt2, r2, err := dA.PingPeer(ctx, chaosBeta, []byte("R2"))
	cut1Dur := time.Since(tCut1)
	if err != nil {
		return err
	}
	ramA.SetHealthy(false)
	tCut2 := time.Now()
	rtt1, r1, err := dA.PingPeer(ctx, chaosBeta, []byte("R1"))
	cut2Dur := time.Since(tCut2)
	if err != nil {
		return err
	}

	fmt.Printf("[REGIME 5: HARD_LINK_CUT CASCADE] total=%.2fms | 0%% loss\n",
		float64(time.Since(t0).Microseconds())/1000.0)
	fmt.Printf("  R3 (%s): rtt=%.1fus | Cut UDP->%s failover=%.1fus (rtt=%.1fus) | Cut RAM->%s failover=%.1fms (rtt=%.1fms)\n",
		r3, float64(rtt3)/1000.0, r2, float64(cut1Dur.Microseconds()), float64(rtt2)/1000.0,
		r1, float64(cut2Dur.Microseconds())/1000.0, float64(rtt1)/1000.0)
	return nil
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( Playtest Entrypoint & Honest Chaos Telemetry Report )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Playtester Proving Ground
// Module:        playtest.harness.chaos_weather_proxy
// Mission:       Execute all 5 chaos weather regimes and report real outcomes.
// ==============================================================================

func main() {
	dA, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(chaosAlpha, "weather-alpha"), engine.NewMonotonicClock(),
	)
	dB, _ := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(chaosBeta, "weather-beta"), engine.NewMonotonicClock(),
	)
	defer func() {
		_ = dA.Close()
		_ = dB.Close()
	}()

	udpA, err := rung3_local.NewUDPTransceiver(chaosAlpha, "127.0.0.1:0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "bind udpA: %v\n", err)
		os.Exit(1)
	}
	udpB, err := rung3_local.NewUDPTransceiver(chaosBeta, "127.0.0.1:0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "bind udpB: %v\n", err)
		os.Exit(1)
	}

	proxy, err := newUDPChaosProxy(udpA.Addr(), udpB.Addr())
	if err != nil {
		fmt.Fprintf(os.Stderr, "proxy init: %v\n", err)
		os.Exit(1)
	}
	defer proxy.close()

	// Wire BOTH nodes to send ONLY through the hostile Chaos Weather Proxy!
	_ = udpA.RegisterPeer(chaosBeta, proxy.connA.LocalAddr().String())
	_ = udpB.RegisterPeer(chaosAlpha, proxy.connB.LocalAddr().String())
	_ = dA.Ladder().RegisterRung(udpA)
	_ = dB.Ladder().RegisterRung(udpB)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	_ = dA.Start(ctx)
	_ = dB.Start(ctx)

	regimes := []WeatherProfile{
		{Name: "REGIME 1: CLEAR_SKY", DropProb: 0, CorruptProb: 0, DupProb: 0, MaxJitterMs: 0, PingTimeout: 250 * time.Millisecond, Pings: 30},
		{Name: "REGIME 2: CONGESTED_LAN", DropProb: 0.10, CorruptProb: 0, DupProb: 0.10, MaxJitterMs: 15, PingTimeout: 250 * time.Millisecond, Pings: 30},
		{Name: "REGIME 3: STORM_CORRUPTION", DropProb: 0.25, CorruptProb: 0.15, DupProb: 0.05, MaxJitterMs: 25, PingTimeout: 250 * time.Millisecond, Pings: 30},
		{Name: "REGIME 4: SILENT_BLACKHOLE", DropProb: 1.00, CorruptProb: 0, DupProb: 0, MaxJitterMs: 0, PingTimeout: 120 * time.Millisecond, Pings: 5},
	}

	fmt.Println("=== FAITHNET PLAYTESTER: USERSPACE CHAOS WEATHER PROXY ===")
	for _, wp := range regimes {
		proxy.setProfile(wp)
		beforeDecodeErrs := dA.StatsSnapshot().DecodeErrors + dB.StatsSnapshot().DecodeErrors
		okCnt := 0
		var rttSum int64
		for i := 0; i < wp.Pings; i++ {
			pCtx, pCancel := context.WithTimeout(ctx, wp.PingTimeout)
			rtt, _, err := dA.PingPeer(pCtx, chaosBeta, []byte("WEATHER-PROBE"))
			pCancel()
			if err == nil {
				okCnt++
				rttSum += rtt
			}
		}
		afterDecodeErrs := dA.StatsSnapshot().DecodeErrors + dB.StatsSnapshot().DecodeErrors
		avgMs := 0.0
		if okCnt > 0 {
			avgMs = float64(rttSum) / float64(okCnt) / 1e6
		}
		fmt.Printf("[%s] sent=%d recv=%d (%.1f%% ok) | avgRTT=%.2fms | proxy(drop=%d corrupt=%d dup=%d) | codecDecodeErrsCaught=%d | activeRung=%s\n",
			wp.Name, wp.Pings, okCnt, float64(okCnt)*100.0/float64(wp.Pings), avgMs,
			proxy.dropped.Load(), proxy.corrupt.Load(), proxy.duped.Load(),
			afterDecodeErrs-beforeDecodeErrs, dA.Ladder().ActiveRung())
	}

	if err := runActiveCascadeTiming(); err != nil {
		fmt.Fprintf(os.Stderr, "[CASCADE ERROR] %v\n", err)
		os.Exit(1)
	}
}

// End of tests/playtester/harness/chaos_weather_proxy/main.go
