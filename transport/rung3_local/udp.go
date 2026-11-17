// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          transport/rung3_local/udp.go
// @id            cog.transport.rung3_udp
// @layer         Layer 3A: Physical & Virtual Transport Substrate (Rung 3 Local)
// @archetype     operational | stateful
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Rung 3 Peer-to-Peer LAN UDP Transceiver ("The Room")
// @role          Local Room & Neighborhood Datagram Transport (Cog 3 impl)
// @maturity      Rev 5 Baseline | Wave 4 The Room & The Well
// @chassis       bounded-heap | Fixed 512B datagram framing + peer UDP map
// @scale         O(1) unicast UDP send/recv; O(N_peers) room broadcast
// @seams         in: data, transport | uses: net, encoding/binary, sync/atomic
// @invariants    1. Stratum Baseline: Downward Dependency Law, Rev 5 5-Block.
//                2. Enforces 512-byte Floppy Sector ceiling on every datagram.
//                3. Zero external internet dependency: works strictly on local
//                   LAN, ad-hoc Wi-Fi, or 127.0.0.1 loopback.
//                4. Receive honors ctx cancellation cleanly (zero leaks).
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Implements Cog 3 (`transport.Transceiver`) for Rung 3 Local ("The Room &
//   Across the Street") using standard library UDP datagrams. Enables multiple
//   sovereign machines in a room or local neighborhood Wi-Fi to discover each
//   other, exchange 512-byte FaithNet frames, and echo pings without an ISP.
//
//   ### For Architects, Operators & Non-Programmers:
//   Lets laptops and devices in the same room or neighborhood talk directly
//   over local Wi-Fi or Ethernet with zero internet connection or cloud server.
//
//   ### For Systems Programmers & Engine Architects:
//   - Ownership & Reentrancy: Owns one `*net.UDPConn` closed idempotently via
//     `atomic.Bool.CompareAndSwap`; auto-learns peer `*net.UDPAddr` when valid
//     `0xFA17` FaithNet headers arrive on the socket.
//   - Extension Seam: Registers directly on Layer 2 `AdaptiveLadder` at Rung 3.
// ==============================================================================

package rung3_local

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Constants, Logical Anchor, UDPTransceiver Chassis & O(1) Accessors )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3A: Physical & Virtual Transport Substrate (Rung 3 Local)
// Module:        cog.transport.rung3_udp
// Mission:       Define Rung 3 UDP transceiver state, peer address registry,
//                and O(1) Cog 3 Transceiver metadata predicates.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"context"
	"encoding/binary"
	"fmt"
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
// Sub-Assembly 2.1: Pre-Conditions & Poll Bounds
// ------------------------------------------------------------------------------

const (
	// DefaultLocalUDPAddr binds an ephemeral loopback UDP port safely by default.
	DefaultLocalUDPAddr = "127.0.0.1:0"

	udpReadPollInterval = 40 * time.Millisecond
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.2: Operational Logical Anchor
// ------------------------------------------------------------------------------

const (
	// ModuleIDLocalUDP is the canonical logical ID for Rung 3 LAN UDP.
	ModuleIDLocalUDP = "cog.transport.rung3_udp"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.3: Public Tagged Ontology (UDPTransceiver Chassis)
// ------------------------------------------------------------------------------

// UDPTransceiver implements `transport.Transceiver` over local LAN/loopback UDP.
//
// ### For Architects, Operators & Non-Programmers:
// Acts as a local room intercom sending and receiving 512-byte FaithNet
// envelopes over local Wi-Fi or Ethernet without touching the global web.
//
// ### For Systems Programmers & Engine Architects:
// Thread-safe wrapper around `*net.UDPConn` with mutex-guarded `NodeID -> UDPAddr`
// directory and atomic lock-free telemetry counters.
type UDPTransceiver struct {
	selfID  data.NodeID
	conn    *net.UDPConn
	mu      sync.RWMutex
	peers   map[data.NodeID]*net.UDPAddr
	healthy atomic.Bool
	closed  atomic.Bool
	sentCnt atomic.Uint64
	recvCnt atomic.Uint64
}

// ------------------------------------------------------------------------------
// Sub-Assembly 2.4: Pre-Conditional Helper Cogs & O(1) Accessors
// ------------------------------------------------------------------------------

// Name returns the canonical Rung 3 UDP transport identifier.
//
// ### For Architects, Operators & Non-Programmers:
// Names this local room UDP channel clearly in logs and diagnostic cards.
//
// ### For Systems Programmers & Engine Architects:
// O(1) static string getter implementing `transport.Transceiver`.
func (u *UDPTransceiver) Name() string {
	return "rung3:udp"
}

// Rung returns data.Rung3Local.
//
// ### For Architects, Operators & Non-Programmers:
// Identifies this transport as belonging to Rung 3 (The Room & Local Airspace).
//
// ### For Systems Programmers & Engine Architects:
// O(1) constant accessor implementing `transport.Transceiver`.
func (u *UDPTransceiver) Rung() data.RungLevel {
	return data.Rung3Local
}

// Healthy reports whether the local UDP socket is open and healthy.
//
// ### For Architects, Operators & Non-Programmers:
// Checks whether the local LAN radio/socket is up and ready for packets.
//
// ### For Systems Programmers & Engine Architects:
// Lock-free atomic check verifying `!closed && healthy`.
func (u *UDPTransceiver) Healthy() bool {
	return !u.closed.Load() && u.healthy.Load()
}

// SetHealthy toggles live health state for deterministic cascade verification.
//
// ### For Architects, Operators & Non-Programmers:
// Allows operators or tests to simulate Wi-Fi dropping or coming back online.
//
// ### For Systems Programmers & Engine Architects:
// Lock-free atomic boolean store gating `Healthy()` and outbound `Send()`.
func (u *UDPTransceiver) SetHealthy(ok bool) {
	u.healthy.Store(ok)
}

// Addr returns the bound local UDP address string (e.g. "127.0.0.1:41902").
//
// ### For Architects, Operators & Non-Programmers:
// Reports the exact local IP and port this node is listening on in the room.
//
// ### For Systems Programmers & Engine Architects:
// Delegates safely to `u.conn.LocalAddr().String()`.
func (u *UDPTransceiver) Addr() string {
	if u.conn == nil {
		return ""
	}
	return u.conn.LocalAddr().String()
}

// RegisterPeer binds a sovereign peer NodeID to its local UDP endpoint string.
//
// ### For Architects, Operators & Non-Programmers:
// Saves a neighbor's local room IP/port so FaithNet can send it frames directly.
//
// ### For Systems Programmers & Engine Architects:
// Resolves `udpAddr` via `net.ResolveUDPAddr("udp", addrStr)` under write lock.
func (u *UDPTransceiver) RegisterPeer(peerID data.NodeID, addrStr string) error {
	resolved, err := net.ResolveUDPAddr("udp", addrStr)
	if err != nil {
		return err
	}
	u.mu.Lock()
	u.peers[peerID] = resolved
	u.mu.Unlock()
	return nil
}

// ==============================================================================
// BLOCK 3: BODY BLOCK
// ( Composable Business Logic: Unicast/Broadcast UDP Send & Context Receive )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3A: Physical & Virtual Transport Substrate (Rung 3 Local)
// Module:        cog.transport.rung3_udp
// Mission:       Transmit and receive 512-byte FaithNet datagrams over local
//                UDP sockets with automatic peer endpoint learning.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: Unicast & Room Broadcast UDP Transmission
// ------------------------------------------------------------------------------

// Send transmits raw wire bytes to target (or all registered peers when target
// is NodeIDBroadcast) over the local UDP socket.
//
// ### For Architects, Operators & Non-Programmers:
// Sends a FaithNet envelope across the local room Wi-Fi or Ethernet to one
// neighbor or announces a beacon to every known neighbor in the room.
//
// ### For Systems Programmers & Engine Architects:
// Enforces 512B Floppy Sector wire ceiling and closed/offline guards upfront.
func (u *UDPTransceiver) Send(target data.NodeID, raw []byte) error {
	if u.closed.Load() {
		return transport.ErrTransceiverClosed
	}
	if !u.healthy.Load() {
		return transport.ErrTransceiverOffline
	}
	if len(raw) > data.FloorMaxFrameSize {
		return transport.ErrWireFrameOversize
	}

	if target.IsBroadcast() {
		u.mu.RLock()
		targets := make([]*net.UDPAddr, 0, len(u.peers))
		for _, addr := range u.peers {
			targets = append(targets, addr)
		}
		u.mu.RUnlock()

		for _, addr := range targets {
			if _, err := u.conn.WriteToUDP(raw, addr); err == nil {
				u.sentCnt.Add(1)
			}
		}
		return nil
	}

	u.mu.RLock()
	peerAddr, ok := u.peers[target]
	u.mu.RUnlock()
	if !ok {
		return transport.ErrPeerUnreachable
	}

	if _, err := u.conn.WriteToUDP(raw, peerAddr); err != nil {
		return transport.ErrPeerUnreachable
	}
	u.sentCnt.Add(1)
	return nil
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: Context-Aware UDP Reception & Peer Auto-Learning
// ------------------------------------------------------------------------------

// Receive waits for one incoming UDP datagram or exits cleanly when ctx cancels.
//
// ### For Architects, Operators & Non-Programmers:
// Listens for envelopes arriving from neighbors in the room and automatically
// remembers the sender's local return address so replies flow effortlessly.
//
// ### For Systems Programmers & Engine Architects:
// Uses bounded `udpReadPollInterval` deadlines so `ctx.Done()` exits promptly.
// Extracts `SourceID` from valid `0xFA17` headers to auto-populate `u.peers`.
func (u *UDPTransceiver) Receive(ctx context.Context) (transport.WireEnvelope, error) {
	if u.closed.Load() {
		return transport.WireEnvelope{}, transport.ErrTransceiverClosed
	}

	var env transport.WireEnvelope
	for {
		if ctx.Err() != nil {
			return transport.WireEnvelope{}, ctx.Err()
		}
		if u.closed.Load() {
			return transport.WireEnvelope{}, transport.ErrTransceiverClosed
		}

		_ = u.conn.SetReadDeadline(time.Now().Add(udpReadPollInterval))
		n, remoteAddr, err := u.conn.ReadFromUDP(env.Bytes[:])
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				continue
			}
			if u.closed.Load() {
				return transport.WireEnvelope{}, transport.ErrTransceiverClosed
			}
			return transport.WireEnvelope{}, err
		}

		env.TargetID = u.selfID
		env.Length = n
		// Auto-learn peer return UDP address when FaithNet 0xFA17 header is present.
		if n >= data.FrameHeaderSize &&
			env.Bytes[0] == data.FrameMagicHi &&
			env.Bytes[1] == data.FrameMagicLo &&
			remoteAddr != nil {
			srcID := data.NodeID(binary.BigEndian.Uint64(env.Bytes[data.OffsetSourceID:]))
			if srcID.IsValid() && srcID != u.selfID {
				env.SenderID = srcID
				u.mu.Lock()
				u.peers[srcID] = remoteAddr
				u.mu.Unlock()
			}
		}

		u.recvCnt.Add(1)
		return env, nil
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Contract Verification Cogs & Pre-Flight Validation Tripwires )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3A: Physical & Virtual Transport Substrate (Rung 3 Local)
// Module:        cog.transport.rung3_udp
// Mission:       Audit live UDPTransceiver state and run a 4-section loopback
//                pre-flight check verifying datagram parity and 512B ceiling.
// @triggers
//   - Block 4 Audits : (*UDPTransceiver).AuditState(), PreflightLocalUDP(),
//                      AuditLocalUDP()
//   - Layer 4 Tests  : tests/functional/transport_functional_test.go,
//                      tests/unit/transport/udp_test.go,
//                      tests/integrated/room_and_well_test.go,
//                      tests/system/sovereign_mesh_e2e_test.go
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Contract, Input & Live-Instance Verification Cogs
// ------------------------------------------------------------------------------

// AuditState inspects a live UDPTransceiver non-destructively to confirm bound
// socket, valid NodeID, and closed/healthy consistency.
//
// ### For Architects, Operators & Non-Programmers:
// Checks a live room UDP endpoint to ensure its local socket is valid and open.
//
// ### For Systems Programmers & Engine Architects:
// Non-mutating O(1) live probe returning structured `data.AuditResult`.
func (u *UDPTransceiver) AuditState() data.AuditResult {
	const check = "rung3_udp_live_state_contract"

	if u == nil {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDLocalUDP, check,
			"UDPTransceiver receiver is nil.")
	}
	if u.conn == nil {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDLocalUDP, check,
			"UDPTransceiver missing active net.UDPConn.")
	}
	if !u.selfID.IsValid() || u.selfID.IsBroadcast() {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDLocalUDP, check,
			"UDPTransceiver selfID is null or broadcast.")
	}
	if u.closed.Load() && u.healthy.Load() {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDLocalUDP, check,
			"Closed UDPTransceiver erroneously reports Healthy()==true.")
	}
	return data.AuditPass(ModuleIDLocalUDP, check,
		"Live Rung 3 UDPTransceiver socket and identity invariants verified.")
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Machine Self-Audit & Pre-Flight Validation Tripwires
// ------------------------------------------------------------------------------

// PreflightLocalUDP executes a 4-section loopback pre-flight self-audit
// covering (1) ephemeral UDP bind & send/recv, (2) unknown peer rejection,
// (3) 512B ceiling enforcement, and (4) clean Close() teardown.
//
// ### For Architects, Operators & Non-Programmers:
// Runs a quick local loopback test to verify room-level UDP communication
// works cleanly on this machine before opening to neighbors.
//
// ### For Systems Programmers & Engine Architects:
// Ephemeral `127.0.0.1:0` 4-section pre-flight returning `data.AuditResult`.
func PreflightLocalUDP() data.AuditResult {
	const check = "rung3_udp_preflight_4section"

	a, err := NewUDPTransceiver(0x01, DefaultLocalUDPAddr)
	if err != nil {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDLocalUDP, check,
			"Sec 1 failed binding ephemeral loopback UDP socket A.")
	}
	defer func() { _ = a.Close() }()

	b, err := NewUDPTransceiver(0x02, DefaultLocalUDPAddr)
	if err != nil {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDLocalUDP, check,
			"Sec 1 failed binding ephemeral loopback UDP socket B.")
	}
	defer func() { _ = b.Close() }()

	// Micro Check 1 (Always Pass): Live audit + loopback datagram exchange.
	if live := a.AuditState(); !live.IsPass() {
		return live
	}
	if err := a.RegisterPeer(0x02, b.Addr()); err != nil {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDLocalUDP, check,
			"Sec 1 RegisterPeer failed for local loopback peer.")
	}
	if err := a.Send(0x02, []byte("UDP-ROOM-PING")); err != nil {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDLocalUDP, check,
			"Sec 1 nominal UDP Send failed.")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	env, err := b.Receive(ctx)
	if err != nil || string(env.Slice()) != "UDP-ROOM-PING" {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDLocalUDP, check,
			"Sec 1 nominal UDP Receive mismatch.")
	}

	// Micro Check 2 (Never Fail): Unknown peer rejected cleanly.
	if err := a.Send(0x9999, []byte("GHOST")); err != transport.ErrPeerUnreachable {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDLocalUDP, check,
			"Sec 2 unknown peer did not return ErrPeerUnreachable.")
	}

	// Micro Check 3 (Invariants & Limits): 513B payload rejected at 512B ceiling.
	var over [data.FloorMaxFrameSize + 1]byte
	if err := a.Send(0x02, over[:]); err != transport.ErrWireFrameOversize {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDLocalUDP, check,
			"Sec 3 513B datagram bypassed 512B Floppy Sector ceiling.")
	}

	// Micro Check 4 (Emergence): Close() marks endpoint unhealthy cleanly.
	if err := b.Close(); err != nil || b.Healthy() {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDLocalUDP, check,
			"Sec 4 Close() failed to mark UDPTransceiver unhealthy.")
	}
	return data.AuditPass(ModuleIDLocalUDP, check,
		"Rung 3 Local UDP passed all 4 micro pre-flight gates.")
}

// AuditLocalUDP returns true iff PreflightLocalUDP() passes cleanly.
//
// ### For Architects, Operators & Non-Programmers:
// Quick pass/fail check confirming local room UDP networking works.
//
// ### For Systems Programmers & Engine Architects:
// Convenience boolean bridge delegating to `PreflightLocalUDP().IsPass()`.
func AuditLocalUDP() bool {
	return PreflightLocalUDP().IsPass()
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK / EXECUTABLE PROCEDURES & LIFECYCLE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3A: Physical & Virtual Transport Substrate (Rung 3 Local)
// Module:        cog.transport.rung3_udp
// Mission:       Expose epilogue readiness, 60-col UDP cards, and lifecycle
//                binding/teardown for local room UDP sockets.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Executable Procedures, Inspection, Readiness & Cards
// ------------------------------------------------------------------------------

// IsLocalUDPReady reports whether Rung 3 LAN UDP passed pre-flight clearance.
//
// ### For Architects, Operators & Non-Programmers:
// Epilogue confirmation that local room UDP networking is healthy and ready.
//
// ### For Systems Programmers & Engine Architects:
// Matches OmniCode's Block 5 epilogue readiness predicate (`is_*_ready`).
func IsLocalUDPReady() bool {
	return PreflightLocalUDP().IsPass()
}

// ExecuteUDPDumpCard renders a 60-column inner-width Rung 3 UDP card.
//
// ### For Architects, Operators & Non-Programmers:
// Displays this node's bound room IP/port, peer count, health, and packet tally.
//
// ### For Systems Programmers & Engine Architects:
// Non-mutating read-locked diagnostic formatter.
func (u *UDPTransceiver) ExecuteUDPDumpCard() string {
	u.mu.RLock()
	peerCount := len(u.peers)
	u.mu.RUnlock()

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
		"FAITHNET RUNG 3 LOCAL UDP TRANSCEIVER CARD",
		ModuleIDLocalUDP,
		u.selfID.String(),
		u.Addr(),
		peerCount,
		fmt.Sprintf("%d sent / %d recv (healthy=%t)",
			u.sentCnt.Load(), u.recvCnt.Load(), u.Healthy()),
	)
}

// ------------------------------------------------------------------------------
// Sub-Assembly 5.2: Subsystem Lifecycle Initialization & Teardown
// ------------------------------------------------------------------------------

// NewUDPTransceiver binds a local UDP socket at listenAddr for selfID.
//
// ### For Architects, Operators & Non-Programmers:
// Opens a local room UDP ear so nearby FaithNet devices can send/echo packets.
//
// ### For Systems Programmers & Engine Architects:
// Resolves listenAddr (defaulting to `127.0.0.1:0` if empty) via `net.ListenUDP`.
func NewUDPTransceiver(selfID data.NodeID, listenAddr string) (*UDPTransceiver, error) {
	if listenAddr == "" {
		listenAddr = DefaultLocalUDPAddr
	}
	addr, err := net.ResolveUDPAddr("udp", listenAddr)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}
	u := &UDPTransceiver{
		selfID: selfID,
		conn:   conn,
		peers:  make(map[data.NodeID]*net.UDPAddr),
	}
	u.healthy.Store(true)
	return u, nil
}

// Close closes the underlying UDP socket idempotently.
//
// ### For Architects, Operators & Non-Programmers:
// Closes the local room intercom cleanly with zero leaked sockets.
//
// ### For Systems Programmers & Engine Architects:
// Atomic CompareAndSwap guard ensures repeated Close() calls are safe.
func (u *UDPTransceiver) Close() error {
	if !u.closed.CompareAndSwap(false, true) {
		return nil
	}
	u.healthy.Store(false)
	return u.conn.Close()
}

// End of transport/rung3_local/udp.go
