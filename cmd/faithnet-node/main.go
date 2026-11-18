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
	"bufio"
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// --- Tier 2: External Library Dependencies (None — 100% Go Stdlib Invariant) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---
import (
	"github.com/CreativeWorkzStudio/FaithNet/data"
	"github.com/CreativeWorkzStudio/FaithNet/engine"
	"github.com/CreativeWorkzStudio/FaithNet/engine/covenant"
	"github.com/CreativeWorkzStudio/FaithNet/interpreter"
	"github.com/CreativeWorkzStudio/FaithNet/orchestrator"
	"github.com/CreativeWorkzStudio/FaithNet/transport/rung1_floor"
	"github.com/CreativeWorkzStudio/FaithNet/transport/rung2_host"
	"github.com/CreativeWorkzStudio/FaithNet/transport/rung3_local"
	"github.com/CreativeWorkzStudio/FaithNet/transport/rung4_global"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Pre-Conditions & Default Node Constants
// ------------------------------------------------------------------------------

const (
	// DefaultNodeID is the default sovereign 64-bit node address.
	DefaultNodeID uint64 = 0xFA17000000000001

	// DefaultNodeLabel is the default human-readable sovereign node label.
	DefaultNodeLabel = "FaithNet-Anchor"

	// DefaultLANPingTimeoutMs honors empirical Wi-Fi power-save spikes (>300ms).
	DefaultLANPingTimeoutMs = 2000

	// DefaultBeaconInterval is the periodic LAN discovery OpBeacon cadence.
	DefaultBeaconInterval = 3 * time.Second
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
	BroadcastAddr     string
	JoinAddr          string
	JoinID            uint64
	WellAddr          string
	TLSCertFile       string
	TLSKeyFile        string
	TLSCADir          string
	BeaconSSID        string
	FetchURI          string
	PingTimeoutMs     int
	EnableWell        bool
	EnableBeacon      bool
	VerboseBeacons    bool
	Interactive       bool
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
		// Wire zero-config LAN subnet broadcast discovery when listening on LAN.
		bcastTarget := opts.BroadcastAddr
		if bcastTarget == "" && opts.EnableBeacon {
			host, port, splitErr := net.SplitHostPort(udpTx.Addr())
			if splitErr == nil && host != "127.0.0.1" && host != "::1" && port != "" {
				bcastTarget = net.JoinHostPort(rung3_local.DiscoverLANBroadcastIP(), port)
			}
		}
		if bcastTarget != "" {
			if err := udpTx.SetBroadcastAddr(bcastTarget); err != nil {
				_ = udpTx.Close()
				return fmt.Errorf("invalid --broadcast-addr %s: %w", bcastTarget, err)
			}
		}
		if opts.JoinAddr != "" {
			anchorID := data.NodeID(opts.JoinID)
			if !anchorID.IsValid() || anchorID.IsBroadcast() {
				anchorID = data.NodeID(DefaultNodeID)
			}
			if err := udpTx.RegisterPeer(anchorID, opts.JoinAddr); err != nil {
				_ = udpTx.Close()
				return fmt.Errorf("invalid --join address %s: %w", opts.JoinAddr, err)
			}
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

	// Optional Well captive HTML portal (with 3-Tier Adaptive TLS Trust Ladder
	// and zero-config sovereign peer pairing + self-replicating binary download).
	var well *interpreter.WellServer
	if opts.EnableWell {
		well = interpreter.NewWellServer(disp, opts.WellAddr)
		well.SetTLSOptions(interpreter.WellTLSOptions{
			CertFile: opts.TLSCertFile,
			KeyFile:  opts.TLSKeyFile,
			CADir:    opts.TLSCADir,
		})
		udpPort := "7777"
		if udpTx != nil {
			if _, p, splitErr := net.SplitHostPort(udpTx.Addr()); splitErr == nil && p != "" {
				udpPort = p
			}
		}
		exePath, _ := os.Executable()
		well.SetPairOptions(interpreter.WellPairOptions{
			AnchorUDPPort: udpPort,
			BinaryPath:    exePath,
		})
		if !opts.ProbeOnly {
			if err := well.Start(); err != nil {
				return fmt.Errorf("failed starting The Well portal: %w", err)
			}
			defer func() { _ = well.Close() }()
		}
		fmt.Println(well.ExecuteWellDumpCard())
		if !opts.ProbeOnly {
			fmt.Printf("The Well TLS Trust Tier: %s\n", well.TLSTrustTier())
		}
	}

	if opts.ProbeOnly {
		return nil
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Wire native zero-HTTP FNP/1 resource responder on Port 7777 (OpFetch -> OpResource).
	disp.SetResourceHandler(func(reqPath string, from data.NodeID) ([]byte, bool) {
		switch reqPath {
		case "/status":
			bulletins := 0
			if well != nil {
				var buf [interpreter.MaxWellBulletins]interpreter.BulletinEntry
				bulletins = well.BulletinsSnapshotInto(buf[:])
			}
			return []byte(fmt.Sprintf(
				`{"protocol":"FNP/1","transport":"udp_7777_zero_http","node_id":"%s","label":"%s","active_rung":"%s","peer_count":%d,"bulletins":%d,"airlock_blocked":%d}`,
				manifest.ID.String(),
				manifest.Label,
				disp.Ladder().ActiveRung().String(),
				disp.Router().ActiveCount(),
				bulletins,
				disp.StatsSnapshot().AirlockBlocked,
			)), true
		case "/", "/well":
			bulletins := 0
			if well != nil {
				var buf [interpreter.MaxWellBulletins]interpreter.BulletinEntry
				bulletins = well.BulletinsSnapshotInto(buf[:])
			}
			return []byte(fmt.Sprintf(
				"=== FAITHNET NATIVE FNP/1 ROOM (ZERO HTTP) ===\nNode:      %s (%s)\nProtocol:  FNP/1 on Port 7777/udp | Active Rung: %s\nPeers:     %d | Bulletins: %d\nRoutes:    /status | /well | /protocol | /bulletins",
				manifest.Label,
				manifest.ID.String(),
				disp.Ladder().ActiveRung().String(),
				disp.Router().ActiveCount(),
				bulletins,
			)), true
		case "/bulletins":
			if well == nil {
				return []byte("FNP/1 Bulletins: (The Well bulletin ring disabled on this node)"), true
			}
			var buf [interpreter.MaxWellBulletins]interpreter.BulletinEntry
			n := well.BulletinsSnapshotInto(buf[:])
			if n == 0 {
				return []byte("FNP/1 Bulletins (0): Room is quiet. Use 'say <msg>' or The Well to post."), true
			}
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("FNP/1 Bulletins (%d latest):\n", n))
			start := 0
			if n > 5 {
				start = n - 5
			}
			for i := n - 1; i >= start; i-- {
				e := buf[i]
				line := fmt.Sprintf("[#%d | %s | %s] %s\n", e.Seq, e.Truth, e.ClientOS, e.Message)
				if sb.Len()+len(line) > data.FloorMaxPayload {
					break
				}
				sb.WriteString(line)
			}
			return []byte(strings.TrimRight(sb.String(), "\n")), true
		default:
			return nil, false
		}
	})

	// Forward live incoming peer messages to terminal and The Well (if enabled).
	// Suppress repetitive periodic [BEACON RX] spam so interactive typing stays clean.
	var seenMu sync.Mutex
	seenPeers := make(map[data.NodeID]bool)
	disp.SetDeliverHandler(func(f *data.Frame, v covenant.Verdict) {
		switch f.Opcode {
		case data.OpData, data.OpLegacyWrap:
			msg := string(f.PayloadSlice())
			if opts.Interactive {
				fmt.Printf("\r[RX %s | %s | %s] %s\nfaithnet> ",
					f.Rung.String(), f.SourceID.String(), v.State.String(), msg)
			} else {
				fmt.Printf("[RX %s | %s | %s] %s\n",
					f.Rung.String(), f.SourceID.String(), v.State.String(), msg)
			}
			if well != nil {
				well.RecordBulletin(interpreter.BulletinEntry{
					Seq:      f.SeqID,
					ClientOS: f.SourceID.String(),
					Truth:    v.State.String(),
					Rung:     f.Rung.String(),
					Message:  msg,
				})
			}
		case data.OpBeacon:
			seenMu.Lock()
			firstTime := !seenPeers[f.SourceID]
			seenPeers[f.SourceID] = true
			seenMu.Unlock()
			if firstTime || opts.VerboseBeacons {
				if opts.Interactive {
					fmt.Printf("\r[PEER DISCOVERED] peer=%s rung=%s info=%q\nfaithnet> ",
						f.SourceID.String(), f.Rung.String(), string(f.PayloadSlice()))
				} else {
					fmt.Printf("[PEER DISCOVERED] peer=%s rung=%s info=%q\n",
						f.SourceID.String(), f.Rung.String(), string(f.PayloadSlice()))
				}
			}
		}
	})

	if err := disp.Start(ctx); err != nil {
		return err
	}

	var beaconSeq atomic.Uint32
	if beacon != nil {
		go func() {
			sendOnce := func() {
				bf := beacon.BuildBeaconFrame(manifest.ID, data.Rung3Local, beaconSeq.Add(1))
				_, _ = disp.SendFrame(&bf)
			}
			sendOnce()
			t := time.NewTicker(DefaultBeaconInterval)
			defer t.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-t.C:
					sendOnce()
				}
			}
		}()
	}

	fmt.Printf("FaithNet node %q (%s) online on %s. Press Ctrl-C to stop.\n",
		manifest.Label, manifest.ID.String(), disp.Ladder().ActiveRung().String())

	// If --join was supplied, execute immediate unicast Anchor handshake!
	if opts.JoinAddr != "" {
		anchorID := data.NodeID(opts.JoinID)
		if !anchorID.IsValid() || anchorID.IsBroadcast() {
			anchorID = data.NodeID(DefaultNodeID)
		}
		timeoutMs := opts.PingTimeoutMs
		if timeoutMs <= 0 {
			timeoutMs = DefaultLANPingTimeoutMs
		}
		pCtx, pCancel := context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
		rtt, rung, pErr := disp.PingPeer(pCtx, anchorID, []byte("JOIN-HANDSHAKE:"+manifest.Label))
		pCancel()
		if pErr != nil {
			fmt.Printf("[JOIN WARN] Anchor %s (%s) handshake timed out (%v); periodic beacons remain active.\n",
				anchorID.String(), opts.JoinAddr, pErr)
		} else {
			fmt.Printf("[PAIRED] Connected to Anchor %s (%s) on %s | RTT=%.2fms | +1:ALIGNED!\n",
				anchorID.String(), opts.JoinAddr, rung.String(), float64(rtt)/1e6)
		}
	}

	if opts.Interactive {
		runInteractiveREPL(ctx, stop, disp, beacon, well, manifest, opts.PingTimeoutMs)
	} else {
		<-ctx.Done()
	}

	fmt.Println("FaithNet node shutting down cleanly...")
	return nil
}

func runNativeFNP1Fetch(rawURI string, timeoutMs int) error {
	u, err := engine.ParseFaithNetURI(rawURI)
	if err != nil {
		return fmt.Errorf("invalid --fetch URI %q: %w", rawURI, err)
	}
	if timeoutMs <= 0 {
		timeoutMs = DefaultLANPingTimeoutMs
	}

	host := u.Host
	targetID := data.NodeID(DefaultNodeID)
	if u.HasNodeID() {
		targetID = u.NodeID
		host = engine.FaithNetDefaultHost
	} else if strings.EqualFold(host, "home-anchor") {
		host = engine.FaithNetDefaultHost
	}

	reqPath := "/"
	if u.Action != "" {
		reqPath = "/" + u.Action
		if u.Target != "" {
			reqPath += "/" + u.Target
		}
	}

	clientID := data.NodeID(0xFA17F00000000000 | uint64(time.Now().UnixNano()&0xFFFF))
	manifest := data.DefaultNodeManifest(clientID, "FNP1-Fetch-Client")
	clk := engine.NewMonotonicClock()
	disp, err := orchestrator.NewDispatcher(manifest, clk)
	if err != nil {
		return err
	}
	defer func() { _ = disp.Close() }()

	udpTx, err := rung3_local.NewUDPTransceiver(clientID, "0.0.0.0:0")
	if err != nil {
		return err
	}
	remoteAddr := net.JoinHostPort(host, strconv.Itoa(int(u.Port)))
	if err := udpTx.RegisterPeer(targetID, remoteAddr); err != nil {
		_ = udpTx.Close()
		return fmt.Errorf("failed resolving FNP/1 target %s: %w", remoteAddr, err)
	}
	if err := disp.Ladder().RegisterRung(udpTx); err != nil {
		_ = udpTx.Close()
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()
	if err := disp.Start(ctx); err != nil {
		return err
	}

	payload, rttNs, rung, err := disp.FetchResource(ctx, targetID, reqPath)
	if err != nil {
		return fmt.Errorf("FNP/1 OpFetch %s (%s) failed: %w", rawURI, remoteAddr, err)
	}
	fmt.Printf("[FNP/1 %s | %s | RTT=%.2fms | %dB | ZERO-HTTP]\n%s\n",
		rung.String(), remoteAddr, float64(rttNs)/1e6, len(payload), string(payload))
	return nil
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: Interactive Live Playtest Operator REPL
// ------------------------------------------------------------------------------

func runInteractiveREPL(
	ctx context.Context,
	cancel context.CancelFunc,
	disp *orchestrator.Dispatcher,
	beacon *rung3_local.BeaconSwitch,
	well *interpreter.WellServer,
	manifest data.NodeManifest,
	pingTimeoutMs int,
) {
	if pingTimeoutMs <= 0 {
		pingTimeoutMs = DefaultLANPingTimeoutMs
	}
	fmt.Println("Interactive mode active. Commands: help | status | peers | ping <hex|all> | fetch <path> | say <msg> | beacon | quit")
	sc := bufio.NewScanner(os.Stdin)
	var seq uint32 = 1

	for {
		fmt.Print("faithnet> ")
		if !sc.Scan() {
			cancel()
			return
		}
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		cmd := strings.ToLower(parts[0])
		arg := ""
		if len(parts) > 1 {
			arg = strings.TrimSpace(parts[1])
		}

		switch cmd {
		case "help":
			fmt.Println("  status           - print live Ladder, Dispatcher & Well cards")
			fmt.Println("  peers            - list auto-discovered sovereign peers")
			fmt.Println("  ping <hex|all>   - send live sovereign ping (timeout", pingTimeoutMs, "ms)")
			fmt.Println("  fetch <path>     - native zero-HTTP FNP/1 fetch (/status, /well, /protocol, /bulletins)")
			fmt.Println("  say <message>    - broadcast sovereign +1:ALIGNED bulletin")
			fmt.Println("  beacon           - broadcast immediate LAN discovery OpBeacon")
			fmt.Println("  quit             - shut down cleanly")
		case "status":
			fmt.Println(disp.Ladder().ExecuteLadderDumpCard())
			fmt.Println(disp.ExecuteDispatcherDumpCard())
			if well != nil {
				fmt.Println(well.ExecuteWellDumpCard())
			}
		case "peers":
			var routes [engine.MaxPeers]engine.PeerRoute
			n := disp.Router().SnapshotInto(routes[:])
			if n == 0 {
				fmt.Println("  (no peers discovered yet — run 'beacon' or wait 3s)")
			}
			for i := 0; i < n; i++ {
				r := routes[i]
				fmt.Printf("  [%d] peer=%s nextHop=%s rung=%s hops=%d lastRTT=%dus\n",
					i+1, r.PeerID.String(), r.NextHopID.String(), r.Rung.String(),
					r.HopCount, r.LastRTTNano/1000)
			}
		case "ping":
			var targets []data.NodeID
			if arg == "" || strings.EqualFold(arg, "all") {
				var routes [engine.MaxPeers]engine.PeerRoute
				n := disp.Router().SnapshotInto(routes[:])
				for i := 0; i < n; i++ {
					targets = append(targets, routes[i].PeerID)
				}
			} else {
				cleanHex := strings.TrimPrefix(strings.ToLower(arg), "0x")
				val, err := strconv.ParseUint(cleanHex, 16, 64)
				if err != nil {
					fmt.Printf("  invalid hex NodeID: %q\n", arg)
					continue
				}
				targets = append(targets, data.NodeID(val))
			}
			if len(targets) == 0 {
				fmt.Println("  no peers to ping yet")
				continue
			}
			for _, id := range targets {
				pCtx, pCancel := context.WithTimeout(ctx, time.Duration(pingTimeoutMs)*time.Millisecond)
				rtt, rung, err := disp.PingPeer(pCtx, id, []byte("LIVE-REPL-PING"))
				pCancel()
				if err != nil {
					fmt.Printf("  PING %s -> TIMEOUT/ERR (%s): %v\n", id.String(), rung.String(), err)
				} else {
					fmt.Printf("  PONG %s <- rung=%s rtt=%.2fms (+1:ALIGNED)\n",
						id.String(), rung.String(), float64(rtt)/1e6)
				}
			}
		case "say":
			if arg == "" {
				fmt.Println("  usage: say <message>")
				continue
			}
			var f data.Frame
			f.Version = data.ProtocolVersion
			f.Rung = disp.Ladder().ActiveRung()
			f.Opcode = data.OpData
			f.Flags = data.FlagFloorCompact | data.FlagCovenantVerified
			f.Truth = data.QuadPositive
			f.SourceID = manifest.ID
			f.TargetID = data.NodeIDBroadcast
			f.SeqID = seq
			seq++
			f.SetPayload([]byte(arg))
			rung, err := disp.SendFrame(&f)
			if err != nil {
				fmt.Printf("  send failed: %v\n", err)
			} else {
				fmt.Printf("  broadcast sent on %s (%dB)\n", rung.String(), len(arg))
				if well != nil {
					well.RecordBulletin(interpreter.BulletinEntry{
						Seq:      f.SeqID,
						ClientOS: manifest.Label,
						Truth:    data.QuadPositive.String(),
						Rung:     rung.String(),
						Message:  arg,
					})
				}
			}
		case "beacon":
			if beacon == nil {
				fmt.Println("  beacon disabled (start with --beacon)")
				continue
			}
			bf := beacon.BuildBeaconFrame(manifest.ID, data.Rung3Local, seq)
			seq++
			if rung, err := disp.SendFrame(&bf); err != nil {
				fmt.Printf("  beacon failed: %v\n", err)
			} else {
				fmt.Printf("  beacon broadcast sent on %s\n", rung.String())
			}
		case "fetch":
			reqPath := arg
			if reqPath == "" {
				reqPath = "/status"
			}
			if !strings.HasPrefix(reqPath, "/") {
				reqPath = "/" + reqPath
			}
			var routes [engine.MaxPeers]engine.PeerRoute
			n := disp.Router().SnapshotInto(routes[:])
			if n == 0 {
				fmt.Println("  (no discovered peers yet — wait for beacon or use --join)")
				continue
			}
			fCtx, fCancel := context.WithTimeout(ctx, time.Duration(pingTimeoutMs)*time.Millisecond)
			body, rtt, rung, err := disp.FetchResource(fCtx, routes[0].PeerID, reqPath)
			fCancel()
			if err != nil {
				fmt.Printf("  fetch %s from %s FAILED: %v\n", reqPath, routes[0].PeerID.String(), err)
			} else {
				fmt.Printf("  [FNP/1 %s | %s | RTT=%.2fms | ZERO-HTTP]\n%s\n",
					rung.String(), routes[0].PeerID.String(), float64(rtt)/1e6, string(body))
			}
		case "quit", "exit":
			cancel()
			return
		default:
			fmt.Printf("  unknown command %q (type 'help')\n", cmd)
		}
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Node CLI Contract Validation & Full-Stack 22-Gate Pre-Flight Aggregator )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 4: Thin Imperative Sovereign Node CLI Entrypoint
// Module:        cmd.faithnet.node
// Mission:       Validate NodeCLIOptions, execute 4-section CLI self-audit, and
//                run every Layer 0..4 Block 4 pre-flight probe on --audit.
// @triggers
//   - Block 4 Audits : ValidateNodeCLIOptions(), PreflightNodeCLI(),
//                      runFullNodeAudits()
//   - Layer 4 Tests  : tests/functional/cmd_functional_test.go,
//                      tests/unit/cmd/node_cli_test.go,
//                      tests/system/cli_binaries_system_test.go
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Contract, Input & Node CLI Option Verification Cogs
// ------------------------------------------------------------------------------

// ValidateNodeCLIOptions audits node startup options against sovereign NodeID
// and IEEE 802.11 Beacon SSID bounds prior to starting listeners.
//
// ### For Architects, Operators & Non-Programmers:
// Checks node startup flags up front so bad IDs or oversized Wi-Fi names are
// caught cleanly before opening any ports.
//
// ### For Systems Programmers & Engine Architects:
// Pure O(1) contract gate returning structured `data.AuditResult`.
func ValidateNodeCLIOptions(opts NodeCLIOptions) data.AuditResult {
	const check = "cli_node_options_contract"

	id := data.NodeID(opts.NodeID)
	if !id.IsValid() || id.IsBroadcast() {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDNodeCLI, check,
			"CLI --id must be a valid non-null, non-broadcast 64-bit NodeID.")
	}
	if opts.Label == "" || len(opts.Label) > 64 {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDNodeCLI, check,
			"CLI --label must be 1..64 bytes.")
	}
	if len(opts.BeaconSSID) == 0 || len(opts.BeaconSSID) > rung3_local.MaxBeaconSSIDBytes {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDNodeCLI, check,
			"CLI --beacon-ssid must be 1..32 bytes (IEEE 802.11 bound).")
	}
	if (opts.TLSCertFile != "") != (opts.TLSKeyFile != "") {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDNodeCLI, check,
			"CLI --tls-cert and --tls-key must be supplied together as a pair.")
	}
	return data.AuditPass(ModuleIDNodeCLI, check,
		"NodeCLIOptions sovereign ID and beacon bounds verified.")
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Machine Self-Audit & Full-Stack Pre-Flight Aggregator
// ------------------------------------------------------------------------------

// PreflightNodeCLI executes a 4-section self-audit covering (1) canonical
// defaults, (2) null/broadcast ID rejection, (3) 32-byte SSID ceiling when
// beacon is active, and (4) Layer 2 orchestrator readiness.
//
// ### For Architects, Operators & Non-Programmers:
// Verifies faithnet-node's flag validation rules before launching a node.
//
// ### For Systems Programmers & Engine Architects:
// Non-destructive 4-section CLI pre-flight returning structured `data.AuditResult`.
func PreflightNodeCLI() data.AuditResult {
	const check = "cli_node_preflight_4section"

	// Micro Check 1 (Always Pass): Canonical defaults pass cleanly.
	def := NodeCLIOptions{
		NodeID:     DefaultNodeID,
		Label:      DefaultNodeLabel,
		BeaconSSID: data.DefaultBeaconSSID,
	}
	if r := ValidateNodeCLIOptions(def); !r.IsPass() {
		return r
	}
	// Micro Check 2 (Never Fail): Null and broadcast NodeID rejected.
	if ValidateNodeCLIOptions(NodeCLIOptions{NodeID: 0}).IsPass() ||
		ValidateNodeCLIOptions(NodeCLIOptions{NodeID: uint64(data.NodeIDBroadcast)}).IsPass() {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDNodeCLI, check,
			"Sec 2 null or broadcast --id bypassed ValidateNodeCLIOptions.")
	}
	// Micro Check 3 (Invariants & Limits): 33-byte SSID rejected with --beacon.
	badBeacon := def
	badBeacon.EnableBeacon = true
	badBeacon.BeaconSSID = "123456789012345678901234567890123"
	if ValidateNodeCLIOptions(badBeacon).IsPass() {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDNodeCLI, check,
			"Sec 3 33-byte --beacon-ssid bypassed IEEE 802.11 ceiling.")
	}
	// Micro Check 4 (Emergence): Orchestrator & Well readiness handshakes pass.
	if !orchestrator.IsDispatcherReady() || !interpreter.IsWellServerReady() {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDNodeCLI, check,
			"Sec 4 underlying Dispatcher or WellServer readiness failed.")
	}
	return data.AuditPass(ModuleIDNodeCLI, check,
		"faithnet-node CLI passed all 4 micro pre-flight gates.")
}

func runFullNodeAudits() bool {
	results := []data.AuditResult{
		PreflightNodeCLI(),
		engine.AuditDataVocabulary(),
		engine.PreflightCodec(),
		engine.PreflightTritCodec(),
		engine.PreflightFaithNetURI(),
		engine.PreflightFEC(),
		engine.PreflightChunker(),
		engine.PreflightCrypto(),
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
		rung1_floor.PreflightSerialFloor(),
		rung1_floor.PreflightAudioModem(),
		rung1_floor.PreflightMorseFloor(),
		rung4_global.PreflightGlobalTCP(),
		rung4_global.PreflightGlobalEgress(),
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
	flag.StringVar(&opts.BroadcastAddr, "broadcast-addr", "", "Optional LAN UDP broadcast address (auto-detected when --beacon is on)")
	flag.StringVar(&opts.JoinAddr, "join", "", "Anchor UDP address to pair with automatically (e.g. 10.77.77.1:7777)")
	flag.Uint64Var(&opts.JoinID, "join-id", DefaultNodeID, "Anchor sovereign 64-bit NodeID for --join pairing")
	flag.StringVar(&opts.WellAddr, "well-addr", data.DefaultWellListenAddr, "The Well HTML portal listen address")
	flag.StringVar(&opts.TLSCertFile, "tls-cert", "", "Tier 3 public rollout PEM certificate chain for The Well")
	flag.StringVar(&opts.TLSKeyFile, "tls-key", "", "Tier 3 public rollout PEM private key for The Well")
	flag.StringVar(&opts.TLSCADir, "tls-ca-dir", "", "Tier 2 persistent offline Sovereign Local Root CA directory")
	flag.StringVar(&opts.BeaconSSID, "beacon-ssid", data.DefaultBeaconSSID, "Room Beacon SSID (1..32 bytes)")
	flag.StringVar(&opts.FetchURI, "fetch", "", "Execute native zero-HTTP FNP/1 OpFetch over Port 7777 (e.g. faithnet://faithnet.lan/well)")
	flag.IntVar(&opts.PingTimeoutMs, "ping-timeout", DefaultLANPingTimeoutMs, "Live LAN ping timeout in milliseconds")
	flag.BoolVar(&opts.EnableWell, "well", false, "Enable The Well zero-install HTML captive portal")
	flag.BoolVar(&opts.EnableBeacon, "beacon", false, "Activate Room Beacon switch and periodic LAN OpBeacon discovery")
	flag.BoolVar(&opts.VerboseBeacons, "verbose-beacons", false, "Log every periodic 3s beacon instead of announcing each peer once")
	flag.BoolVar(&opts.Interactive, "interactive", false, "Enable interactive playtest operator console (REPL)")
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

	if opts.FetchURI != "" {
		if err := runNativeFNP1Fetch(opts.FetchURI, opts.PingTimeoutMs); err != nil {
			fmt.Fprintf(os.Stderr, "faithnet-node --fetch failed: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if res := ValidateNodeCLIOptions(opts); !res.IsPass() {
		fmt.Fprintf(os.Stderr, "error: %s\n", res.String())
		os.Exit(2)
	}

	if err := runSovereignNode(opts); err != nil {
		fmt.Fprintf(os.Stderr, "faithnet-node failed: %v\n", err)
		os.Exit(1)
	}
}

// End of cmd/faithnet-node/main.go
