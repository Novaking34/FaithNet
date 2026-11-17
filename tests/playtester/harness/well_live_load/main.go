// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          tests/playtester/harness/well_live_load/main.go
// @id            playtest.harness.well_live_load
// @layer         Layer 4: Proving Ground (Live Well HTTP Portal Load & Security)
// @archetype     playtest_harness | operational
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Layer 3B Live Well Captive Portal Load Playtester
// @role          Real TCP/HTTP Mixed Client + Hostile Attack Load Benchmark
// @maturity      Rev 5 Baseline | Wave 7A Playtester Proving Ground
// @chassis       bounded-heap | Real net/http server + concurrent OS client pool
// @scale         Hundreds of real TCP/HTTP requests across 6 OS User-Agents
// @seams         in: data, engine, orchestrator, interpreter, rung2_host
// @invariants    1. Binds a real TCP HTTP listener on 127.0.0.1:0 (not httptest).
//                2. Mixes real Android/iOS/Win/Mac/Linux/CLI clients + attacks.
//                3. Verifies 0 hostile attacks accepted and reports p50/p95/p99.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Playtests Layer 3B ("The Well" captive portal) over a REAL OS TCP listener
//   (`well.Start()`) with real `net/http.Client` connections simulating a room
//   full of visitor phones, laptops, captive-portal probes, and hostile web
//   attackers simultaneously—measuring real HTTP latency percentiles and throughput.
//
//   ### For Architects, Operators & Non-Programmers:
//   Simulates dozens of neighbors hitting The Well portal at once from Android,
//   iPhone, Windows, and Mac while attackers throw XSS/SQLi at the same door,
//   measuring real response speeds and proving zero attacks sneak through.
//
//   ### For Systems Programmers & Engine Architects:
//   Drives `WellServer.Start()` over real loopback TCP sockets with concurrent
//   worker pools and collects exact request/sec, p50/p95/p99 latencies, and
//   `/api/status` JSON audit counters.
// ==============================================================================

package main

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Playtest Constants, Multi-OS Client Profiles & Attack Corpus )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Playtester Proving Ground
// Module:        playtest.harness.well_live_load
// Mission:       Define realistic visitor User-Agents and hostile payloads.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// --- Tier 2: External Library Dependencies (None — 100% Go Stdlib Invariant) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---
import (
	"github.com/CreativeWorkzStudio/FaithNet/data"
	"github.com/CreativeWorkzStudio/FaithNet/engine"
	"github.com/CreativeWorkzStudio/FaithNet/interpreter"
	"github.com/CreativeWorkzStudio/FaithNet/orchestrator"
	"github.com/CreativeWorkzStudio/FaithNet/transport/rung2_host"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Operational Constants & Realistic Multi-OS Visitors
// ------------------------------------------------------------------------------

const (
	ModuleIDWellLiveLoad = "playtest.harness.well_live_load"
	wellNodeID           = data.NodeID(0xFA1700000000D001)
)

var visitorAgents = []string{
	"Mozilla/5.0 (Linux; Android 14; Pixel 8)",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 17_4 like Mac OS X)",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 14_4)",
	"Mozilla/5.0 (X11; Ubuntu; Linux x86_64)",
	"curl/8.5.0",
}

var hostileBodies = []string{
	"<script>document.cookie</script>",
	"1 UNION SELECT password FROM users",
	"SELECT * FROM secrets; DROP TABLE bulletins;",
	"<?php system('id'); ?>",
	"javascript:alert(1)",
	strings.Repeat("OOM", 200),
}

// ==============================================================================
// BLOCK 3: BODY BLOCK
// ( Live HTTP Load & Mixed Hostility Driver )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Playtester Proving Ground
// Module:        playtest.harness.well_live_load
// Mission:       Hammer live TCP WellServer with mixed real HTTP traffic.
// ==============================================================================

func main() {
	disp, err := orchestrator.NewDispatcher(
		data.DefaultNodeManifest(wellNodeID, "Room-Well-Playtest"), engine.NewMonotonicClock(),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "NewDispatcher: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = disp.Close() }()

	bus := rung2_host.NewMemoryBus()
	_ = disp.Ladder().RegisterRung(bus.Attach(wellNodeID))

	well := interpreter.NewWellServer(disp, "127.0.0.1:0")
	if err := well.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "WellServer.Start: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = well.Close() }()

	baseURL := "http://" + well.ListenAddr()
	client := &http.Client{
		Timeout: 3 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	const workers = 12
	const reqsPerWorker = 40
	var cleanOK, captiveOK, attackBlocked, attackBreach, netErrs atomic.Uint64
	latCh := make(chan int64, workers*reqsPerWorker)

	start := time.Now()
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < reqsPerWorker; i++ {
				ua := visitorAgents[(id+i)%len(visitorAgents)]
				t0 := time.Now()
				switch i % 4 {
				case 0: // GET /
					req, _ := http.NewRequest(http.MethodGet, baseURL+"/", nil)
					req.Header.Set("User-Agent", ua)
					resp, err := client.Do(req)
					if err != nil {
						netErrs.Add(1)
						continue
					}
					_, _ = io.Copy(io.Discard, resp.Body)
					_ = resp.Body.Close()
					if resp.StatusCode == http.StatusOK {
						cleanOK.Add(1)
					}
				case 1: // POST clean local news
					msg := fmt.Sprintf("Neighbor %d update #%d: block power nominal", id, i)
					req, _ := http.NewRequest(http.MethodPost, baseURL+"/api/message", strings.NewReader(msg))
					req.Header.Set("User-Agent", ua)
					resp, err := client.Do(req)
					if err != nil {
						netErrs.Add(1)
						continue
					}
					_, _ = io.Copy(io.Discard, resp.Body)
					_ = resp.Body.Close()
					if resp.StatusCode == http.StatusAccepted {
						cleanOK.Add(1)
					}
				case 2: // Captive portal OS probe
					req, _ := http.NewRequest(http.MethodGet, baseURL+"/generate_204", nil)
					req.Header.Set("User-Agent", ua)
					resp, err := client.Do(req)
					if err != nil {
						netErrs.Add(1)
						continue
					}
					_, _ = io.Copy(io.Discard, resp.Body)
					_ = resp.Body.Close()
					if resp.StatusCode == http.StatusFound {
						captiveOK.Add(1)
					}
				case 3: // Hostile attack attempt
					atk := hostileBodies[(id+i)%len(hostileBodies)]
					req, _ := http.NewRequest(http.MethodPost, baseURL+"/api/message", strings.NewReader(atk))
					req.Header.Set("User-Agent", ua)
					resp, err := client.Do(req)
					if err != nil {
						netErrs.Add(1)
						continue
					}
					_, _ = io.Copy(io.Discard, resp.Body)
					_ = resp.Body.Close()
					if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusRequestEntityTooLarge {
						attackBlocked.Add(1)
					} else {
						attackBreach.Add(1)
					}
				}
				latCh <- time.Since(t0).Microseconds()
			}
		}(w)
	}
	wg.Wait()
	elapsed := time.Since(start)
	close(latCh)

	lats := make([]int64, 0, workers*reqsPerWorker)
	for v := range latCh {
		lats = append(lats, v)
	}
	sort.Slice(lats, func(i, j int) bool { return lats[i] < lats[j] })
	pct := func(p float64) int64 {
		if len(lats) == 0 {
			return 0
		}
		return lats[int(float64(len(lats)-1)*p)]
	}

	// Fetch final /api/status telemetry from live server.
	stResp, err := client.Get(baseURL + "/api/status")
	var st interpreter.WellStatusPayload
	if err == nil {
		_ = json.NewDecoder(stResp.Body).Decode(&st)
		_ = stResp.Body.Close()
	}

	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	totalReqs := workers * reqsPerWorker
	rps := float64(totalReqs) / elapsed.Seconds()

	fmt.Println("=== FAITHNET PLAYTESTER: LIVE WELL HTTP LOAD & AIRLOCK STRESS ===")
	fmt.Printf("Endpoint: %s | Workers: %d | Total Reqs: %d | Elapsed: %.1fms | Throughput: %.0f req/s\n",
		baseURL, workers, totalReqs, float64(elapsed.Microseconds())/1000.0, rps)
	fmt.Printf("Results: cleanOK=%d captiveRedirects=%d attacksBlocked=%d breaches=%d netErrs=%d\n",
		cleanOK.Load(), captiveOK.Load(), attackBlocked.Load(), attackBreach.Load(), netErrs.Load())
	fmt.Printf("Latency (us): min=%d  p50=%d  p95=%d  p99=%d  max=%d | HeapAlloc=%.2fMB\n",
		pct(0), pct(0.50), pct(0.95), pct(0.99), pct(1.0), float64(ms.Alloc)/(1024*1024))
	fmt.Printf("Live /api/status: node=%s rung=%s bulletins=%d/32 airlockBlocked=%d\n",
		st.Label, st.ActiveRung, st.Bulletins, st.AirlockBlocked)

	if attackBreach.Load() > 0 || netErrs.Load() > 0 {
		os.Exit(1)
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Playtest Contract Verification — Inline Status & Breach Gate )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Playtester Proving Ground
// Module:        playtest.harness.well_live_load
// Mission:       Non-zero exit on any Airlock breach or live TCP error.
// ==============================================================================

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ( End of Live Well Playtest Runner )
// ==============================================================================

// End of tests/playtester/harness/well_live_load/main.go
