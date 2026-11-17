// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          transport/rung2_host/ipc.go
// @id            cog.transport.rung2_ipc
// @layer         Layer 3A: Physical & Virtual Transport Substrate (Rung 2 Host)
// @archetype     operational | stateful
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Rung 2 Local Unix Domain Datagram IPC Transceiver
// @role          Runtime Engine: Inter-Process Socket Bridge (Zero NIC Required)
// @maturity      Rev 5 Baseline | Wave 2 First Breath
// @chassis       bounded-heap | Single UnixConn + 512-byte stack read buffer
// @scale         O(1) kernel unixgram syscall per 512-byte floor frame
// @seams         in: data, transport | uses: context, net, os, sync, time
// @invariants    1. Stratum Baseline: Universal 5-Block Lifecycle Protocol.
//                2. Operates across separate OS processes with zero network
//                   cards, zero IP addresses, and zero TCP/IP stack involvement.
//                3. Cleans up bound socket path automatically upon Close().
//                4. Strict 512-byte Floor MTU boundary enforced on Send/Recv.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Implements the inter-process Rung 2 (Host Substrate) transceiver using
//   POSIX Unix domain datagram sockets (`unixgram`). Enables two independent
//   local processes (e.g., MillenniumOS kernel daemon and a CLI tool) to
//   exchange FaithNet frames locally without touching Ethernet or Wi-Fi.
//
//   ### For Architects, Operators & Non-Programmers:
//   Lets two separate programs running on the same laptop talk directly through
//   a local file-system mailbox—even if the computer has no Wi-Fi card at all.
//
//   ### For Systems Programmers & Engine Architects:
//   - Controlled Memory Chassis: Reuses a fixed `[FloorMaxFrameSize]byte` read
//     buffer per Receive call and polls read deadlines so context cancellation
//     exits cleanly without hanging blocked kernel syscalls.
//   - Extension Seam: Peer socket paths are bound explicitly via `RegisterPeer`
//     so sovereign NodeIDs remain decoupled from filesystem paths.
// ==============================================================================

package rung2_host

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Bounds, Logical Anchor, IPCTransceiver Struct & O(1) Accessors )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3A: Physical & Virtual Transport Substrate (Rung 2 Host)
// Module:        cog.transport.rung2_ipc
// Mission:       Declare IPC poll bounds, IPCTransceiver chassis, and O(1)
//                health/identity/peer-mapping helpers.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
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
// Sub-Assembly 2.1: Pre-Conditions & Poll Deadlines
// ------------------------------------------------------------------------------

const (
	// ipcReadPollInterval bounds Receive responsiveness to ctx cancellation.
	ipcReadPollInterval = 25 * time.Millisecond
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.2: Operational Logical Anchor
// ------------------------------------------------------------------------------

const (
	// ModuleIDHostIPC is the canonical logical ID for Rung 2 Unix IPC.
	ModuleIDHostIPC = "cog.transport.rung2_ipc"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.3: Public Tagged Ontology (IPCTransceiver Chassis)
// ------------------------------------------------------------------------------

// IPCTransceiver exchanges 512-byte FaithNet frames over local unixgram sockets.
type IPCTransceiver struct {
	mu       sync.RWMutex
	selfID   data.NodeID
	sockPath string
	conn     *net.UnixConn
	peers    map[data.NodeID]string
	healthy  atomic.Bool
	closed   atomic.Bool
	sentCnt  atomic.Uint64
	recvCnt  atomic.Uint64
}

// ------------------------------------------------------------------------------
// Sub-Assembly 2.4: Pre-Conditional Helper Cogs & O(1) Accessors
// ------------------------------------------------------------------------------

// Name returns the canonical Rung 2 IPC transport identifier.
//
// ### For Architects, Operators & Non-Programmers:
// Names this local cross-process mailbox clearly in status cards and logs.
//
// ### For Systems Programmers & Engine Architects:
// O(1) static string getter implementing `transport.Transceiver`.
func (u *IPCTransceiver) Name() string {
	return "rung2:ipc"
}

// Rung returns data.Rung2Host.
//
// ### For Architects, Operators & Non-Programmers:
// Identifies this local IPC socket as belonging to Rung 2 (Host Substrate).
//
// ### For Systems Programmers & Engine Architects:
// O(1) constant accessor implementing `transport.Transceiver`.
func (u *IPCTransceiver) Rung() data.RungLevel {
	return data.Rung2Host
}

// Healthy reports whether the local Unix socket is open and healthy.
//
// ### For Architects, Operators & Non-Programmers:
// Checks whether this local mailbox is open and ready to exchange packets.
//
// ### For Systems Programmers & Engine Architects:
// Lock-free atomic read verifying `!closed && healthy`.
func (u *IPCTransceiver) Healthy() bool {
	return !u.closed.Load() && u.healthy.Load()
}

// SetHealthy toggles live health state for deterministic cascade verification.
//
// ### For Architects, Operators & Non-Programmers:
// Lets operators or test harnesses simulate cutting or restoring the IPC link.
//
// ### For Systems Programmers & Engine Architects:
// Lock-free atomic boolean store gating `Healthy()` and outbound `Send()`.
func (u *IPCTransceiver) SetHealthy(ok bool) {
	u.healthy.Store(ok)
}

// SocketPath returns the bound local filesystem socket path.
//
// ### For Architects, Operators & Non-Programmers:
// Reports the file path where this local mailbox listens for neighbors.
//
// ### For Systems Programmers & Engine Architects:
// O(1) immutable field getter.
func (u *IPCTransceiver) SocketPath() string {
	return u.sockPath
}

// RegisterPeer binds a sovereign peer NodeID to its local unixgram socket path.
//
// ### For Architects, Operators & Non-Programmers:
// Saves another local FaithNet process's mailbox path so we can send it frames.
//
// ### For Systems Programmers & Engine Architects:
// Write-locked map update associating `peerID -> peerSockPath`.
func (u *IPCTransceiver) RegisterPeer(peerID data.NodeID, peerSockPath string) {
	u.mu.Lock()
	u.peers[peerID] = peerSockPath
	u.mu.Unlock()
}

// ==============================================================================
// BLOCK 3: BODY BLOCK
// ( Composable Business Logic: Unixgram Send & Context-Aware Receive )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3A: Physical & Virtual Transport Substrate (Rung 2 Host)
// Module:        cog.transport.rung2_ipc
// Mission:       Transmit and receive bounded 512-byte FaithNet envelopes across
//                independent host processes over local unixgram sockets.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: Point-to-Point & Broadcast Unixgram Transmission
// ------------------------------------------------------------------------------

// Send writes raw FaithNet wire bytes to target's registered local socket path.
//
// ### For Architects, Operators & Non-Programmers:
// Drops an encoded FaithNet envelope directly into another local program's
// socket mailbox on the same computer.
//
// ### For Systems Programmers & Engine Architects:
// Uses `WriteToUnix` datagrams preserving exact frame boundaries (zero stream
// deframing needed) and enforces the 512-byte Floor MTU ceiling.
func (u *IPCTransceiver) Send(target data.NodeID, raw []byte) error {
	if u.closed.Load() {
		return transport.ErrTransceiverClosed
	}
	if !u.healthy.Load() {
		return transport.ErrTransceiverOffline
	}
	if len(raw) > data.FloorMaxFrameSize {
		return transport.ErrWireFrameOversize
	}

	u.mu.RLock()
	if target.IsBroadcast() {
		paths := make([]string, 0, len(u.peers))
		for _, p := range u.peers {
			paths = append(paths, p)
		}
		u.mu.RUnlock()
		for _, p := range paths {
			_, _ = u.conn.WriteToUnix(raw, &net.UnixAddr{Name: p, Net: "unixgram"})
		}
		u.sentCnt.Add(1)
		return nil
	}

	peerPath, ok := u.peers[target]
	u.mu.RUnlock()
	if !ok {
		return transport.ErrPeerUnreachable
	}

	if _, err := u.conn.WriteToUnix(raw, &net.UnixAddr{
		Name: peerPath,
		Net:  "unixgram",
	}); err != nil {
		return transport.ErrPeerUnreachable
	}
	u.sentCnt.Add(1)
	return nil
}

// ------------------------------------------------------------------------------
// Sub-Assembly 3.2: Context-Aware Unixgram Reception
// ------------------------------------------------------------------------------

// Receive waits for one incoming datagram from the local Unix socket or stops
// cleanly when ctx is cancelled.
//
// ### For Architects, Operators & Non-Programmers:
// Listens on this process's local mailbox and hands each arriving envelope up
// to the dispatcher cleanly.
//
// ### For Systems Programmers & Engine Architects:
// Loops with short read deadlines (`ipcReadPollInterval`) so context
// cancellation exits promptly without leaking blocked syscalls.
func (u *IPCTransceiver) Receive(ctx context.Context) (transport.WireEnvelope, error) {
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

		_ = u.conn.SetReadDeadline(time.Now().Add(ipcReadPollInterval))
		n, _, err := u.conn.ReadFromUnix(env.Bytes[:])
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
		u.recvCnt.Add(1)
		return env, nil
	}
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Contract Verification Cogs & Pre-Flight Validation Tripwires )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3A: Physical & Virtual Transport Substrate (Rung 2 Host)
// Module:        cog.transport.rung2_ipc
// Mission:       Audit live unixgram IPC endpoints and execute a 4-section
//                pre-flight check inside an isolated temporary directory.
// @triggers
//   - Block 4 Audits : (*IPCTransceiver).AuditState(), PreflightHostIPC(),
//                      AuditHostIPC()
//   - Layer 4 Tests  : tests/functional/transport_functional_test.go,
//                      tests/unit/transport/ipc_test.go,
//                      tests/integrated/host_ipc_mesh_test.go,
//                      tests/system/sovereign_mesh_e2e_test.go
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Contract, Input & Live-Instance Verification Cogs
// ------------------------------------------------------------------------------

// AuditState inspects a live IPCTransceiver non-destructively to verify its
// bound socket, valid NodeID, socket path, and closed/healthy consistency.
//
// ### For Architects, Operators & Non-Programmers:
// Checks a running local IPC mailbox to confirm its OS socket is valid and
// properly tracked before sending or receiving cross-process packets.
//
// ### For Systems Programmers & Engine Architects:
// Non-mutating O(1) live probe returning structured `data.AuditResult`.
func (u *IPCTransceiver) AuditState() data.AuditResult {
	const check = "rung2_ipc_live_state_contract"

	if u == nil {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDHostIPC, check,
			"IPCTransceiver receiver is nil.")
	}
	if u.conn == nil || u.sockPath == "" {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDHostIPC, check,
			"IPCTransceiver missing active unixgram socket or path.")
	}
	if !u.selfID.IsValid() || u.selfID.IsBroadcast() {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDHostIPC, check,
			"IPCTransceiver selfID is null or broadcast.")
	}
	if u.closed.Load() && u.healthy.Load() {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDHostIPC, check,
			"Closed IPCTransceiver erroneously reports Healthy()==true.")
	}
	return data.AuditPass(ModuleIDHostIPC, check,
		"Live Rung 2 IPCTransceiver socket and state invariants verified.")
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Machine Self-Audit & Pre-Flight Validation Tripwires
// ------------------------------------------------------------------------------

// PreflightHostIPC executes a 4-section non-destructive pre-flight self-audit
// covering (1) unixgram bind/send/recv, (2) unknown peer & oversize rejection,
// (3) 512B ceiling enforcement, and (4) Close() socket-file removal.
//
// ### For Architects, Operators & Non-Programmers:
// Bench-tests cross-process communication in a temporary scratch folder and
// verifies every temporary socket file disappears cleanly afterwards.
//
// ### For Systems Programmers & Engine Architects:
// Isolated 4-section pre-flight gate returning structured `data.AuditResult`.
func PreflightHostIPC() data.AuditResult {
	const check = "rung2_ipc_preflight_4section"

	dir, err := os.MkdirTemp("", "faithnet-ipc-audit-*")
	if err != nil {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDHostIPC, check,
			"Unable to allocate isolated temporary directory for IPC audit.")
	}
	defer func() { _ = os.RemoveAll(dir) }()

	sockA := filepath.Join(dir, "a.sock")
	sockB := filepath.Join(dir, "b.sock")

	a, err := NewIPCTransceiver(0x01, sockA)
	if err != nil {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDHostIPC, check,
			"Sec 1 failed binding temporary unixgram socket A.")
	}
	defer func() { _ = a.Close() }()

	b, err := NewIPCTransceiver(0x02, sockB)
	if err != nil {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDHostIPC, check,
			"Sec 1 failed binding temporary unixgram socket B.")
	}
	defer func() { _ = b.Close() }()

	// Micro Check 1 (Always Pass): Live audit + nominal datagram exchange.
	if live := a.AuditState(); !live.IsPass() {
		return live
	}
	a.RegisterPeer(0x02, sockB)
	if err := a.Send(0x02, []byte("IPC-PING")); err != nil {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDHostIPC, check,
			"Sec 1 nominal unixgram Send failed.")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	env, err := b.Receive(ctx)
	if err != nil || string(env.Slice()) != "IPC-PING" {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDHostIPC, check,
			"Sec 1 nominal unixgram Receive payload mismatch.")
	}

	// Micro Check 2 (Never Fail): Unregistered peer rejected cleanly.
	if err := a.Send(0x9999, []byte("GHOST")); err != transport.ErrPeerUnreachable {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDHostIPC, check,
			"Sec 2 unregistered peer did not return ErrPeerUnreachable.")
	}

	// Micro Check 3 (Invariants & Limits): 513B payload rejected at 512B cap.
	var oversize [data.FloorMaxFrameSize + 1]byte
	if err := a.Send(0x02, oversize[:]); err != transport.ErrWireFrameOversize {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDHostIPC, check,
			"Sec 3 513B payload bypassed 512B Floppy Sector ceiling.")
	}

	// Micro Check 4 (Emergence): Close() removes socket file from disk cleanly.
	if err := b.Close(); err != nil || b.Healthy() {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDHostIPC, check,
			"Sec 4 Close() failed to mark IPCTransceiver unhealthy.")
	}
	if _, statErr := os.Stat(sockB); !os.IsNotExist(statErr) {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDHostIPC, check,
			"Sec 4 Close() leaked unixgram socket file on disk.")
	}

	return data.AuditPass(ModuleIDHostIPC, check,
		"Rung 2 Host IPC passed all 4 micro pre-flight gates.")
}

// AuditHostIPC returns true iff PreflightHostIPC() passes cleanly.
//
// ### For Architects, Operators & Non-Programmers:
// Quick pass/fail check confirming local cross-process IPC is working.
//
// ### For Systems Programmers & Engine Architects:
// Convenience boolean bridge delegating to `PreflightHostIPC().IsPass()`.
func AuditHostIPC() bool {
	return PreflightHostIPC().IsPass()
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK / EXECUTABLE PROCEDURES & LIFECYCLE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3A: Physical & Virtual Transport Substrate (Rung 2 Host)
// Module:        cog.transport.rung2_ipc
// Mission:       Render 60-column IPC diagnostic cards, expose epilogue
//                readiness, and own unixgram socket lifecycle binding/teardown.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Executable Procedures, Inspection, Readiness & Cards
// ------------------------------------------------------------------------------

// IsHostIPCReady reports whether Rung 2 unixgram IPC passed pre-flight audit.
//
// ### For Architects, Operators & Non-Programmers:
// Epilogue confirmation that local cross-process IPC mailboxes are ready.
//
// ### For Systems Programmers & Engine Architects:
// Matches OmniCode's Block 5 epilogue readiness predicate (`is_*_ready`).
func IsHostIPCReady() bool {
	return PreflightHostIPC().IsPass()
}

// ExecuteIPCDumpCard renders a 60-column inner-width diagnostic card.
//
// ### For Architects, Operators & Non-Programmers:
// Prints a framed status summary of this IPC mailbox's ID, health, and traffic.
//
// ### For Systems Programmers & Engine Architects:
// Non-mutating diagnostic formatter using lock-free atomic telemetry loads.
func (u *IPCTransceiver) ExecuteIPCDumpCard() string {
	return fmt.Sprintf(
		"+------------------------------------------------------------+\n"+
			"| %-58s |\n"+
			"+------------------------------------------------------------+\n"+
			"| Module:   %-48s |\n"+
			"| NodeID:   %-48s |\n"+
			"| Rung:     %-48s |\n"+
			"| Healthy:  %-48t |\n"+
			"| Traffic:  %-48s |\n"+
			"+------------------------------------------------------------+",
		"FAITHNET RUNG 2 UNIX IPC TRANSCEIVER CARD",
		ModuleIDHostIPC,
		u.selfID.String(),
		u.Rung().String(),
		u.Healthy(),
		fmt.Sprintf("%d sent / %d recv", u.sentCnt.Load(), u.recvCnt.Load()),
	)
}

// ------------------------------------------------------------------------------
// Sub-Assembly 5.2: Subsystem Lifecycle Initialization & Teardown
// ------------------------------------------------------------------------------

// NewIPCTransceiver binds a fresh unixgram socket at sockPath for selfID.
//
// ### For Architects, Operators & Non-Programmers:
// Opens a private local mailbox file on this machine so separate FaithNet
// processes can talk offline with zero network card required.
//
// ### For Systems Programmers & Engine Architects:
// Removes any stale socket at sockPath prior to `net.ListenUnixgram`.
func NewIPCTransceiver(selfID data.NodeID, sockPath string) (*IPCTransceiver, error) {
	_ = os.Remove(sockPath)
	conn, err := net.ListenUnixgram("unixgram", &net.UnixAddr{
		Name: sockPath,
		Net:  "unixgram",
	})
	if err != nil {
		return nil, err
	}
	u := &IPCTransceiver{
		selfID:   selfID,
		sockPath: sockPath,
		conn:     conn,
		peers:    make(map[data.NodeID]string),
	}
	u.healthy.Store(true)
	return u, nil
}

// Close closes the underlying Unix socket and removes the socket file cleanly.
//
// ### For Architects, Operators & Non-Programmers:
// Shuts down this local IPC mailbox and cleans up its socket file from disk.
//
// ### For Systems Programmers & Engine Architects:
// Idempotent via atomic CompareAndSwap; safe for concurrent teardown calls.
func (u *IPCTransceiver) Close() error {
	if !u.closed.CompareAndSwap(false, true) {
		return nil
	}
	u.healthy.Store(false)
	err := u.conn.Close()
	_ = os.Remove(u.sockPath)
	return err
}

// End of transport/rung2_host/ipc.go
