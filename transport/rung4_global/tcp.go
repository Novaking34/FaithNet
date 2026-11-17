// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          transport/rung4_global/tcp.go
// @id            cog.transport.rung4_tcp
// @layer         Layer 3A: Physical & Virtual Transport Substrate (Rung 4 Global)
// @archetype     operational | stateful
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Rung 4 Global Federated TCP Stream Transceiver
// @role          Length-Prefixed TCP Frame Tunnel (Optional Compatibility Cog 3)
// @maturity      Rev 5 Baseline | Wave 5 Global Compatibility & Stress
// @chassis       bounded-heap | 2-byte length prefix + 512B Floppy Sector cap
// @scale         O(1) per length-prefixed TCP frame send/receive
// @seams         in: data, transport | uses: encoding/binary, io, net, sync
// @invariants    1. Stratum Baseline: Downward Dependency Law, Rev 5 5-Block.
//                2. Enforces 512-byte Floppy Sector ceiling on BOTH outbound
//                   writes and inbound uint16 length headers.
//                3. Strictly optional compatibility layer: FaithNet never
//                   depends on Rung 4 to function offline across Rungs 1..3.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Implements Cog 3 (`transport.Transceiver`) for Rung 4 Global using length-
//   prefixed TCP streams (`uint16` big-endian length + raw FaithNet wire bytes).
//   Allows two sovereign FaithNet nodes to federate across standard modern
//   IP/TCP routers when explicitly opted in via `NodeManifest.AllowGlobalEgress`.
//
//   ### For Architects, Operators & Non-Programmers:
//   Provides optional high-speed global TCP compatibility so two distant
//   FaithNet nodes can link over the modern internet—while remaining completely
//   detachable so cutting TCP drops cleanly back to local room/floor rungs.
//
//   ### For Systems Programmers & Engine Architects:
//   - Ownership & Reentrancy: Owns one `net.Listener` accepting bounded inbound
//     connections into a fixed channel (`DefaultTCPQueueDepth`); `Close()`
//     terminates the accept loop cleanly with zero leaked goroutines.
// ==============================================================================

package rung4_global

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Constants, Logical Anchor, TCPTransceiver Chassis & O(1) Accessors )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3A: Physical & Virtual Transport Substrate (Rung 4 Global)
// Module:        cog.transport.rung4_tcp
// Mission:       Define length-prefixed TCP framing bounds, peer registry, and
//                Rung 4 TCPTransceiver state & accessors.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// --- Tier 2: External Library Dependencies (None — 100% Go Stdlib Invariant) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---
import (
	"github.com/CreativeWorkzStudio/FaithNet/data"
	"github.com/CreativeWorkzStudio/FaithNet/transport"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Pre-Conditions & TCP Framing Limits
// ------------------------------------------------------------------------------

const (
	// DefaultGlobalTCPAddr binds an ephemeral loopback TCP listener by default.
	DefaultGlobalTCPAddr = "127.0.0.1:0"

	// DefaultTCPQueueDepth bounds queued inbound frames per TCPTransceiver.
	DefaultTCPQueueDepth = 64

	tcpDialTimeout = 2 * time.Second
	tcpIOTimeout   = 2 * time.Second
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.2: Operational Logical Anchor
// ------------------------------------------------------------------------------

const (
	// ModuleIDGlobalTCP is the canonical logical ID for Rung 4 Global TCP.
	ModuleIDGlobalTCP = "cog.transport.rung4_tcp"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.3: Public Tagged Ontology (TCPTransceiver Chassis)
// ------------------------------------------------------------------------------

// TCPTransceiver implements `transport.Transceiver` on `data.Rung4Global`
// using 2-byte big-endian length-prefixed TCP connections.
type TCPTransceiver struct {
	selfID   data.NodeID
	listener net.Listener
	mu       sync.RWMutex
	peers    map[data.NodeID]string
	inbox    chan transport.WireEnvelope
	healthy  atomic.Bool
	closed   atomic.Bool
	sentCnt  atomic.Uint64
	recvCnt  atomic.Uint64
	wg       sync.WaitGroup
}

// ------------------------------------------------------------------------------
// Sub-Assembly 2.4: Pre-Conditional Helper Cogs & O(1) Accessors
// ------------------------------------------------------------------------------

// Name returns the canonical Rung 4 Global TCP identifier.
//
// ### For Architects, Operators & Non-Programmers:
// Names this global TCP tunnel clearly in status cards and logs.
//
// ### For Systems Programmers & Engine Architects:
// O(1) static getter implementing `transport.Transceiver`.
func (t *TCPTransceiver) Name() string {
	return "rung4:tcp"
}

// Rung returns data.Rung4Global.
//
// ### For Architects, Operators & Non-Programmers:
// Identifies this TCP tunnel as belonging to Rung 4 (Global Compatibility).
//
// ### For Systems Programmers & Engine Architects:
// O(1) constant accessor implementing `transport.Transceiver`.
func (t *TCPTransceiver) Rung() data.RungLevel {
	return data.Rung4Global
}

// Healthy reports whether the TCP listener is open and operational.
//
// ### For Architects, Operators & Non-Programmers:
// Checks whether the global TCP link is currently up and healthy.
//
// ### For Systems Programmers & Engine Architects:
// Lock-free atomic check verifying `!closed && healthy`.
func (t *TCPTransceiver) Healthy() bool {
	return !t.closed.Load() && t.healthy.Load()
}

// SetHealthy toggles live health state for deterministic cascade testing.
//
// ### For Architects, Operators & Non-Programmers:
// Simulates losing or regaining global TCP backbone connectivity cleanly.
//
// ### For Systems Programmers & Engine Architects:
// Lock-free atomic store gating `Healthy()` and outbound `Send()`.
func (t *TCPTransceiver) SetHealthy(ok bool) {
	t.healthy.Store(ok)
}

// Addr returns the bound TCP listener address string (e.g. "127.0.0.1:41999").
//
// ### For Architects, Operators & Non-Programmers:
// Reports the TCP IP and port where this node listens for federated peers.
//
// ### For Systems Programmers & Engine Architects:
// Delegates safely to `t.listener.Addr().String()`.
func (t *TCPTransceiver) Addr() string {
	if t.listener == nil {
		return ""
	}
	return t.listener.Addr().String()
}

// RegisterPeer binds a sovereign peer NodeID to its TCP address string.
//
// ### For Architects, Operators & Non-Programmers:
// Records a remote federated FaithNet peer's TCP host and port.
//
// ### For Systems Programmers & Engine Architects:
// Write-locked map update associating `peerID -> addrStr`.
func (t *TCPTransceiver) RegisterPeer(peerID data.NodeID, addrStr string) {
	t.mu.Lock()
	t.peers[peerID] = addrStr
	t.mu.Unlock()
}

// ==============================================================================
// BLOCK 3: BODY BLOCK
// ( Composable Business Logic: Length-Prefixed TCP Send & Accept/Receive Loop )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3A: Physical & Virtual Transport Substrate (Rung 4 Global)
// Module:        cog.transport.rung4_tcp
// Mission:       Transmit and receive length-prefixed 512B FaithNet envelopes
//                over standard TCP connections cleanly.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: Length-Prefixed TCP Transmission
// ------------------------------------------------------------------------------

// Send dials target's registered TCP address and writes a 2-byte big-endian
// length prefix followed by raw FaithNet wire bytes (<= 512B).
//
// ### For Architects, Operators & Non-Programmers:
// Sends a sealed FaithNet envelope across a standard TCP connection to a
// remote peer.
//
// ### For Systems Programmers & Engine Architects:
// Enforces `tcpDialTimeout`, `tcpIOTimeout`, and 512B Floppy Sector ceiling.
func (t *TCPTransceiver) Send(target data.NodeID, raw []byte) error {
	if t.closed.Load() {
		return transport.ErrTransceiverClosed
	}
	if !t.healthy.Load() {
		return transport.ErrTransceiverOffline
	}
	if len(raw) == 0 || len(raw) > data.FloorMaxFrameSize {
		return transport.ErrWireFrameOversize
	}

	t.mu.RLock()
	peerAddr, ok := t.peers[target]
	t.mu.RUnlock()
	if !ok {
		return transport.ErrPeerUnreachable
	}

	conn, err := net.DialTimeout("tcp", peerAddr, tcpDialTimeout)
	if err != nil {
		return transport.ErrPeerUnreachable
	}
	defer func() { _ = conn.Close() }()

	_ = conn.SetWriteDeadline(time.Now().Add(tcpIOTimeout))
	var hdr [2]byte
	binary.BigEndian.PutUint16(hdr[:], uint16(len(raw)))
	if _, err := conn.Write(hdr[:]); err != nil {
		return transport.ErrPeerUnreachable
	}
	if _, err := conn.Write(raw); err != nil {
		return transport.ErrPeerUnreachable
	}
	t.sentCnt.Add(1)
	return nil
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: Background TCP Accept Worker & Context-Aware Receive
// ------------------------------------------------------------------------------

func (t *TCPTransceiver) acceptLoop() {
	defer t.wg.Done()
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			return
		}
		t.handleConn(conn)
	}
}

func (t *TCPTransceiver) handleConn(conn net.Conn) {
	defer func() { _ = conn.Close() }()
	_ = conn.SetReadDeadline(time.Now().Add(tcpIOTimeout))

	var hdr [2]byte
	if _, err := io.ReadFull(conn, hdr[:]); err != nil {
		return
	}
	frameLen := int(binary.BigEndian.Uint16(hdr[:]))
	if frameLen == 0 || frameLen > data.FloorMaxFrameSize {
		return
	}

	var env transport.WireEnvelope
	if _, err := io.ReadFull(conn, env.Bytes[:frameLen]); err != nil {
		return
	}
	env.TargetID = t.selfID
	env.Length = frameLen
	if frameLen >= data.FrameHeaderSize &&
		env.Bytes[0] == data.FrameMagicHi &&
		env.Bytes[1] == data.FrameMagicLo {
		env.SenderID = data.NodeID(binary.BigEndian.Uint64(env.Bytes[data.OffsetSourceID:]))
	}

	if t.closed.Load() {
		return
	}
	select {
	case t.inbox <- env:
		t.recvCnt.Add(1)
	default:
	}
}

// Receive waits for one inbound length-prefixed TCP frame or ctx cancellation.
//
// ### For Architects, Operators & Non-Programmers:
// Waits for arriving envelopes from federated TCP peers and stops cleanly on
// shutdown.
//
// ### For Systems Programmers & Engine Architects:
// Reads from bounded `t.inbox` channel honoring `ctx.Done()`.
func (t *TCPTransceiver) Receive(ctx context.Context) (transport.WireEnvelope, error) {
	if t.closed.Load() {
		return transport.WireEnvelope{}, transport.ErrTransceiverClosed
	}
	select {
	case <-ctx.Done():
		return transport.WireEnvelope{}, ctx.Err()
	case env, ok := <-t.inbox:
		if !ok {
			return transport.WireEnvelope{}, transport.ErrTransceiverClosed
		}
		return env, nil
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Contract Verification Cogs & Pre-Flight Validation Tripwires )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3A: Physical & Virtual Transport Substrate (Rung 4 Global)
// Module:        cog.transport.rung4_tcp
// Mission:       Audit live TCPTransceiver state and execute a 4-section
//                loopback TCP pre-flight check with clean Close() teardown.
// @triggers
//   - Block 4 Audits : (*TCPTransceiver).AuditState(), PreflightGlobalTCP(),
//                      AuditGlobalTCP()
//   - Layer 4 Tests  : tests/functional/transport_functional_test.go,
//                      tests/unit/transport/tcp_test.go,
//                      tests/integrated/global_airlock_bridge_test.go,
//                      tests/system/sovereign_mesh_e2e_test.go
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Contract, Input & Live-Instance Verification Cogs
// ------------------------------------------------------------------------------

// AuditState inspects a live TCPTransceiver non-destructively to confirm bound
// listener, valid NodeID, and closed/healthy consistency.
//
// ### For Architects, Operators & Non-Programmers:
// Checks a live global TCP endpoint to verify its socket is valid and active.
//
// ### For Systems Programmers & Engine Architects:
// Non-mutating O(1) live probe returning structured `data.AuditResult`.
func (t *TCPTransceiver) AuditState() data.AuditResult {
	const check = "rung4_tcp_live_state_contract"

	if t == nil || t.listener == nil {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDGlobalTCP, check,
			"TCPTransceiver is nil or missing active net.Listener.")
	}
	if !t.selfID.IsValid() || t.selfID.IsBroadcast() {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDGlobalTCP, check,
			"TCPTransceiver selfID is null or broadcast.")
	}
	if t.closed.Load() && t.healthy.Load() {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDGlobalTCP, check,
			"Closed TCPTransceiver erroneously reports Healthy()==true.")
	}
	return data.AuditPass(ModuleIDGlobalTCP, check,
		"Live Rung 4 TCPTransceiver listener and identity invariants verified.")
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Machine Self-Audit & Pre-Flight Validation Tripwires
// ------------------------------------------------------------------------------

// PreflightGlobalTCP executes a 4-section loopback TCP pre-flight audit
// covering (1) length-prefixed TCP send/recv, (2) unknown peer rejection,
// (3) 512B ceiling enforcement, and (4) clean Close() goroutine exit.
//
// ### For Architects, Operators & Non-Programmers:
// Runs a local loopback test confirming FaithNet's optional TCP federation
// channel exchanges sealed frames cleanly and shuts down with zero leaks.
//
// ### For Systems Programmers & Engine Architects:
// Ephemeral `127.0.0.1:0` 4-section pre-flight gate returning `AuditResult`.
func PreflightGlobalTCP() data.AuditResult {
	const check = "rung4_tcp_preflight_4section"

	a, err := NewTCPTransceiver(0x01, DefaultGlobalTCPAddr)
	if err != nil {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDGlobalTCP, check,
			"Sec 1 failed binding ephemeral loopback TCP listener A.")
	}
	defer func() { _ = a.Close() }()

	b, err := NewTCPTransceiver(0x02, DefaultGlobalTCPAddr)
	if err != nil {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDGlobalTCP, check,
			"Sec 1 failed binding ephemeral loopback TCP listener B.")
	}
	defer func() { _ = b.Close() }()

	// Micro Check 1 (Always Pass): Live audit + loopback TCP frame exchange.
	if live := a.AuditState(); !live.IsPass() {
		return live
	}
	a.RegisterPeer(0x02, b.Addr())
	if err := a.Send(0x02, []byte("TCP-GLOBAL-PING")); err != nil {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDGlobalTCP, check,
			"Sec 1 nominal TCP Send failed.")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	env, err := b.Receive(ctx)
	if err != nil || string(env.Slice()) != "TCP-GLOBAL-PING" {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDGlobalTCP, check,
			"Sec 1 nominal TCP Receive mismatch.")
	}

	// Micro Check 2 (Never Fail): Unregistered peer rejected cleanly.
	if err := a.Send(0x9999, []byte("GHOST")); err != transport.ErrPeerUnreachable {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDGlobalTCP, check,
			"Sec 2 unregistered peer did not return ErrPeerUnreachable.")
	}

	// Micro Check 3 (Invariants & Limits): 513B payload rejected at 512B ceiling.
	var over [data.FloorMaxFrameSize + 1]byte
	if err := a.Send(0x02, over[:]); err != transport.ErrWireFrameOversize {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDGlobalTCP, check,
			"Sec 3 513B TCP payload bypassed 512B Floppy Sector ceiling.")
	}

	// Micro Check 4 (Emergence): Close() drains acceptLoop & marks unhealthy.
	if err := b.Close(); err != nil || b.Healthy() {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDGlobalTCP, check,
			"Sec 4 Close() failed to mark TCPTransceiver unhealthy.")
	}
	return data.AuditPass(ModuleIDGlobalTCP, check,
		"Rung 4 Global TCP passed all 4 micro pre-flight gates.")
}

// AuditGlobalTCP returns true iff PreflightGlobalTCP() passes cleanly.
//
// ### For Architects, Operators & Non-Programmers:
// Quick pass/fail shortcut confirming optional TCP federation works.
//
// ### For Systems Programmers & Engine Architects:
// Convenience boolean bridge delegating to `PreflightGlobalTCP().IsPass()`.
func AuditGlobalTCP() bool {
	return PreflightGlobalTCP().IsPass()
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK / EXECUTABLE PROCEDURES & LIFECYCLE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3A: Physical & Virtual Transport Substrate (Rung 4 Global)
// Module:        cog.transport.rung4_tcp
// Mission:       Expose epilogue readiness, 60-col TCP cards, and TCP listener
//                lifecycle binding/teardown.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Executable Procedures, Inspection, Readiness & Cards
// ------------------------------------------------------------------------------

// IsGlobalTCPReady reports whether Rung 4 Global TCP passed pre-flight checks.
//
// ### For Architects, Operators & Non-Programmers:
// Epilogue confirmation that optional TCP federation is healthy and ready.
//
// ### For Systems Programmers & Engine Architects:
// Matches OmniCode's Block 5 epilogue readiness predicate (`is_*_ready`).
func IsGlobalTCPReady() bool {
	return PreflightGlobalTCP().IsPass()
}

// ExecuteTCPDumpCard renders a 60-column inner-width Rung 4 Global TCP card.
//
// ### For Architects, Operators & Non-Programmers:
// Displays the TCP listen port, peer count, health, and packet tally.
//
// ### For Systems Programmers & Engine Architects:
// Non-mutating read-locked diagnostic card builder.
func (t *TCPTransceiver) ExecuteTCPDumpCard() string {
	t.mu.RLock()
	peers := len(t.peers)
	t.mu.RUnlock()
	return fmt.Sprintf(
		"+------------------------------------------------------------+\n"+
			"| %-58s |\n"+
			"+------------------------------------------------------------+\n"+
			"| Module:   %-48s |\n"+
			"| NodeID:   %-48s |\n"+
			"| Listen:   %-48s |\n"+
			"| Peers:    %-48d |\n"+
			"| Traffic:  %-48s |\n"+
			"+------------------------------------------------------------+",
		"FAITHNET RUNG 4 GLOBAL TCP TRANSCEIVER CARD",
		ModuleIDGlobalTCP,
		t.selfID.String(),
		t.Addr(),
		peers,
		fmt.Sprintf("%d sent / %d recv (healthy=%t)",
			t.sentCnt.Load(), t.recvCnt.Load(), t.Healthy()),
	)
}

// ------------------------------------------------------------------------------
// Sub-Assembly 5.2: Subsystem Lifecycle Initialization & Teardown
// ------------------------------------------------------------------------------

// NewTCPTransceiver binds a TCP listener at listenAddr and starts accepting
// length-prefixed FaithNet frames in a bounded background goroutine.
//
// ### For Architects, Operators & Non-Programmers:
// Opens an optional global TCP port so distant FaithNet peers can connect.
//
// ### For Systems Programmers & Engine Architects:
// Defaults empty `listenAddr` to `127.0.0.1:0` and tracks acceptLoop via `wg`.
func NewTCPTransceiver(selfID data.NodeID, listenAddr string) (*TCPTransceiver, error) {
	if listenAddr == "" {
		listenAddr = DefaultGlobalTCPAddr
	}
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return nil, err
	}
	t := &TCPTransceiver{
		selfID:   selfID,
		listener: ln,
		peers:    make(map[data.NodeID]string),
		inbox:    make(chan transport.WireEnvelope, DefaultTCPQueueDepth),
	}
	t.healthy.Store(true)
	t.wg.Add(1)
	go t.acceptLoop()
	return t, nil
}

// Close closes the TCP listener, waits for acceptLoop exit, and closes inbox.
//
// ### For Architects, Operators & Non-Programmers:
// Shuts down global TCP federation cleanly with zero leaked sockets.
//
// ### For Systems Programmers & Engine Architects:
// Idempotent via atomic CompareAndSwap; waits on `t.wg` prior to closing inbox.
func (t *TCPTransceiver) Close() error {
	if !t.closed.CompareAndSwap(false, true) {
		return nil
	}
	t.healthy.Store(false)
	err := t.listener.Close()
	t.wg.Wait()
	close(t.inbox)
	return err
}

// End of transport/rung4_global/tcp.go
