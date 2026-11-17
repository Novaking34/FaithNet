// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          transport/rung4_global/egress.go
// @id            cog.transport.rung4_egress
// @layer         Layer 3A: Physical & Virtual Transport Substrate (Rung 4 Global)
// @archetype     operational | stateful
// @version       0.1.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Rung 4 Airlock-Gated Legal Outbound Egress Gateway
// @role          Opt-In Commercial Web Fetcher Strictly Screened by Genesis 3 T4
// @maturity      Rev 5 Baseline | Wave 5 Global Compatibility & Stress
// @chassis       bounded-heap | Bounded 480B Floor MTU fetch + Genesis 3 Airlock
// @scale         O(N_body <= 481B) bounded outbound HTTP read & airlock screen
// @seams         in: data, engine/covenant | uses: net/http, io, sync/atomic
// @invariants    1. Stratum Baseline: Downward Dependency Law, Rev 5 5-Block.
//                2. Hard Opt-In Lock: Fetch() ALWAYS fails with
//                   `ErrGlobalEgressDisabled` when `AllowGlobalEgress == false`.
//                3. Every single fetched byte passes `covenant.Filter`:
//                   clean facts enter strictly as `0:AWAIT`; hostile/oversize
//                   responses are blocked (`-1:REJECT`) and NEVER returned.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Implements FaithNet's Rung 4 legal outbound egress gateway ("Compatibility
//   Without Dependency"). When an operator explicitly opts in to fetch factual
//   public data (e.g., NOAA weather, emergency alerts) from the commercial web,
//   EgressGateway bounds the HTTP read, screens every byte through the Genesis 3
//   Covenant Airlock, and wraps clean empirical data strictly as a `0:AWAIT`
//   `data.OpLegacyWrap` frame.
//
//   ### For Architects, Operators & Non-Programmers:
//   Acts like an airtight bio-containment airlock when reaching out to the
//   commercial internet: locked shut by default, and even when opened, strips
//   dangerous web scripts and stamps clean facts "0:AWAIT" so outside chaos
//   can never pollute sovereign local truth.
//
//   ### For Systems Programmers & Engine Architects:
//   - Ownership & Reentrancy: Owns bounded `*http.Client` (timeout-enforced)
//     and immutable `*covenant.Filter`; 100% goroutine-safe.
// ==============================================================================

package rung4_global

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Constants, Sentinel Errors, Logical Anchor & EgressGateway Chassis )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3A: Physical & Virtual Transport Substrate (Rung 4 Global)
// Module:        cog.transport.rung4_egress
// Mission:       Define outbound egress bounds, sentinel errors, and the
//                Airlock-gated EgressGateway chassis.
// ==============================================================================

// --- Tier 1: Standard Go Library Packages ---
import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"time"
)

// --- Tier 2: External Library Dependencies (None — 100% Go Stdlib Invariant) ---

// --- Tier 3: Internal FaithNet Foundation Packages ---
import (
	"github.com/CreativeWorkzStudio/FaithNet/data"
	"github.com/CreativeWorkzStudio/FaithNet/engine/covenant"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.1: Pre-Conditions, Fetch Bounds & Sentinel Errors
// ------------------------------------------------------------------------------

const (
	// DefaultEgressTimeout bounds outbound commercial HTTP requests strictly.
	DefaultEgressTimeout = 3 * time.Second

	// DefaultEgressOriginID stamps commercial web origin on wrapped frames.
	DefaultEgressOriginID data.NodeID = 0xEEEE000000000004
)

var (
	// ErrGlobalEgressDisabled indicates NodeManifest.AllowGlobalEgress == false.
	ErrGlobalEgressDisabled = errors.New("faithnet/egress: global egress disabled by offline-first policy")

	// ErrEgressAirlockBlocked indicates fetched external bytes tripped Airlock.
	ErrEgressAirlockBlocked = errors.New("faithnet/egress: response blocked by Genesis 3 Airlock")
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.2: Operational Logical Anchor
// ------------------------------------------------------------------------------

const (
	// ModuleIDGlobalEgress is the canonical logical ID for Rung 4 egress.
	ModuleIDGlobalEgress = "cog.transport.rung4_egress"
)

// ------------------------------------------------------------------------------
// Sub-Assembly 2.3: Public Tagged Ontology (EgressGateway Chassis)
// ------------------------------------------------------------------------------

// EgressGateway performs opt-in, Airlock-screened HTTP GET fetches into
// bounded 512-byte `0:AWAIT` FaithNet `data.OpLegacyWrap` frames.
type EgressGateway struct {
	manifest   data.NodeManifest
	airlock    *covenant.Filter
	client     *http.Client
	fetchedCnt atomic.Uint64
	blockedCnt atomic.Uint64
}

// ------------------------------------------------------------------------------
// Sub-Assembly 2.4: Pre-Conditional Helper Cogs & O(1) Accessors
// ------------------------------------------------------------------------------

// IsEnabled reports whether this gateway is explicitly permitted to reach out.
//
// ### For Architects, Operators & Non-Programmers:
// Checks whether the operator explicitly unlocked outside internet egress.
//
// ### For Systems Programmers & Engine Architects:
// O(1) immutable check against `manifest.AllowGlobalEgress`.
func (g *EgressGateway) IsEnabled() bool {
	return g.manifest.AllowGlobalEgress
}

// ==============================================================================
// BLOCK 3: BODY BLOCK
// ( Composable Business Logic: Opt-In Bounded Fetch & Genesis 3 Screening )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3A: Physical & Virtual Transport Substrate (Rung 4 Global)
// Module:        cog.transport.rung4_egress
// Mission:       Fetch empirical external bytes strictly under AllowGlobalEgress
//                and screen every byte through the Genesis 3 Covenant Airlock.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 3.1: Airlock-Gated Outbound Fetch
// ------------------------------------------------------------------------------

// FetchEmpiricalFrame performs a bounded HTTP GET against rawURL (ONLY when
// `manifest.AllowGlobalEgress == true`), reads at most `FloorMaxPayload+1`
// bytes, audits the response through `covenant.Filter`, and returns a sealed
// `0:AWAIT` `data.OpLegacyWrap` frame—or blocks hostile/oversize responses.
//
// ### For Architects, Operators & Non-Programmers:
// Safely fetches a short factual bulletin from an outside web URL when enabled,
// refusing oversized pages or dangerous scripts at the Genesis 3 Airlock door.
//
// ### For Systems Programmers & Engine Architects:
// Uses `io.LimitReader(resp.Body, data.FloorMaxPayload+1)` so hostile remote
// servers can NEVER trigger unbounded memory allocation (`OOM` DoS defense).
func (g *EgressGateway) FetchEmpiricalFrame(
	ctx context.Context,
	rawURL string,
	seq uint32,
) (data.Frame, covenant.Verdict, error) {
	if !g.manifest.AllowGlobalEgress {
		g.blockedCnt.Add(1)
		return data.Frame{}, covenant.Verdict{
			State:  data.QuadNegative,
			Reason: covenant.ReasonForgedCovenantClaim,
		}, ErrGlobalEgressDisabled
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return data.Frame{}, covenant.Verdict{
			State:  data.QuadNull,
			Reason: covenant.ReasonMalformedStructure,
		}, err
	}
	req.Header.Set("User-Agent", "FaithNet-Egress-Airlock/0.1")

	resp, err := g.client.Do(req)
	if err != nil {
		return data.Frame{}, covenant.Verdict{
			State:  data.QuadNull,
			Reason: covenant.ReasonMalformedStructure,
		}, err
	}
	defer func() { _ = resp.Body.Close() }()

	// Bounded read: at most 481 bytes so oversize trips Airlock cleanly with 0 OOM risk.
	raw, err := io.ReadAll(io.LimitReader(resp.Body, int64(data.FloorMaxPayload+1)))
	if err != nil {
		return data.Frame{}, covenant.Verdict{
			State:  data.QuadNull,
			Reason: covenant.ReasonMalformedStructure,
		}, err
	}

	verdict := g.airlock.EvaluateExternalPayload(raw)
	if verdict.IsBlocked() {
		g.blockedCnt.Add(1)
		return data.Frame{}, verdict, ErrEgressAirlockBlocked
	}

	var f data.Frame
	f.Version = data.ProtocolVersion
	f.Rung = data.Rung4Global
	f.Opcode = data.OpLegacyWrap
	f.Flags = data.FlagExternalOrigin | data.FlagFloorCompact
	f.Truth = verdict.State // Strictly 0:AWAIT.
	f.SourceID = DefaultEgressOriginID
	f.TargetID = g.manifest.ID
	f.SeqID = seq
	f.SetPayload(raw)

	g.fetchedCnt.Add(1)
	return f, verdict, nil
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ( Contract Verification Cogs & Pre-Flight Validation Tripwires )
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3A: Physical & Virtual Transport Substrate (Rung 4 Global)
// Module:        cog.transport.rung4_egress
// Mission:       Audit live EgressGateway policy and run a 4-section local
//                httptest pre-flight verifying opt-in lock, 0:AWAIT, and -1 XSS.
// @triggers
//   - Block 4 Audits : (*EgressGateway).AuditState(), PreflightGlobalEgress(),
//                      AuditGlobalEgress()
//   - Layer 4 Tests  : tests/functional/transport_functional_test.go,
//                      tests/unit/transport/egress_test.go,
//                      tests/integrated/global_airlock_bridge_test.go,
//                      tests/system/airlock_hostility_system_test.go
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 4.1: Contract, Input & Live-Instance Verification Cogs
// ------------------------------------------------------------------------------

// AuditState inspects a live EgressGateway non-destructively to confirm its
// manifest, HTTP client, and armed Genesis 3 Airlock policy.
//
// ### For Architects, Operators & Non-Programmers:
// Verifies the outside-web airlock gateway has its security filter armed and
// timeout guards configured cleanly.
//
// ### For Systems Programmers & Engine Architects:
// Non-mutating live probe returning structured `data.AuditResult`.
func (g *EgressGateway) AuditState() data.AuditResult {
	const check = "rung4_egress_live_state_contract"

	if g == nil || g.airlock == nil || g.client == nil {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDGlobalEgress, check,
			"EgressGateway is nil or missing airlock/http.Client.")
	}
	if !g.manifest.IsValid() {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDGlobalEgress, check,
			"EgressGateway holds invalid NodeManifest.")
	}
	return g.airlock.AuditPolicy()
}

// ------------------------------------------------------------------------------
// Sub-Assembly 4.2: Machine Self-Audit & Pre-Flight Validation Tripwires
// ------------------------------------------------------------------------------

// PreflightGlobalEgress runs a 4-section 100% local `httptest` pre-flight check
// covering (1) clean empirical fetch -> 0:AWAIT, (2) offline-first default lock
// & hostile XSS response block, (3) 481B oversize OOM cap, and (4) AuditState.
//
// ### For Architects, Operators & Non-Programmers:
// Rehearses outside web fetches against a local simulated server to prove the
// airlock stays sealed by default and stops malicious web payloads every time.
//
// ### For Systems Programmers & Engine Architects:
// Zero external network calls: uses local `httptest.NewServer` cleanly.
func PreflightGlobalEgress() data.AuditResult {
	const check = "rung4_egress_preflight_4section"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/clean":
			_, _ = w.Write([]byte("NOAA Local Advisory: Clear 70F"))
		case "/xss":
			_, _ = w.Write([]byte("<html><script>evil()</script></html>"))
		case "/huge":
			buf := make([]byte, 1024)
			for i := range buf {
				buf[i] = 'X'
			}
			_, _ = w.Write(buf)
		}
	}))
	defer srv.Close()

	// Micro Check 2 (Never Fail) — Offline default MUST refuse egress!
	offlineGw := NewEgressGateway(data.DefaultNodeManifest(0x01, "offline"), nil)
	if _, _, err := offlineGw.FetchEmpiricalFrame(context.Background(), srv.URL+"/clean", 1); err != ErrGlobalEgressDisabled {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDGlobalEgress, check,
			"Sec 2 offline-first EgressGateway allowed unauthorized outbound fetch.")
	}

	// Micro Check 1 (Always Pass): Opt-in gateway fetches clean fact as 0:AWAIT.
	optIn := data.DefaultNodeManifest(0x01, "egress")
	optIn.AllowGlobalEgress = true
	optIn.PreferredMaxRung = data.Rung4Global
	gw := NewEgressGateway(optIn, nil)
	if live := gw.AuditState(); !live.IsPass() {
		return live
	}
	f, v, err := gw.FetchEmpiricalFrame(context.Background(), srv.URL+"/clean", 1)
	if err != nil || v.State != data.QuadAwait || f.Truth != data.QuadAwait {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDGlobalEgress, check,
			"Sec 1 clean empirical fetch did not yield 0:AWAIT OpLegacyWrap.")
	}

	// Micro Check 2b: Hostile XSS web response blocked cleanly.
	if _, hv, err := gw.FetchEmpiricalFrame(context.Background(), srv.URL+"/xss", 2); err != ErrEgressAirlockBlocked || hv.State != data.QuadNegative {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDGlobalEgress, check,
			"Sec 2 hostile XSS web response bypassed Genesis 3 Airlock.")
	}

	// Micro Check 3 (Invariants & Limits): 1024B huge page capped & rejected.
	if _, ov, err := gw.FetchEmpiricalFrame(context.Background(), srv.URL+"/huge", 3); err != ErrEgressAirlockBlocked || ov.Reason != covenant.ReasonPayloadOversize {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDGlobalEgress, check,
			"Sec 3 oversized remote page bypassed 480B Floor MTU airlock cap.")
	}

	// Micro Check 4 (Emergence): Counters accurately reflect 1 fetched, 2 blocked.
	if gw.fetchedCnt.Load() != 1 || gw.blockedCnt.Load() != 2 {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDGlobalEgress, check,
			"Sec 4 EgressGateway telemetry counts diverged.")
	}
	return data.AuditPass(ModuleIDGlobalEgress, check,
		"Rung 4 Global Egress passed all 4 micro pre-flight gates.")
}

// AuditGlobalEgress returns true iff PreflightGlobalEgress() passes cleanly.
//
// ### For Architects, Operators & Non-Programmers:
// Quick pass/fail shortcut confirming the outbound web airlock works.
//
// ### For Systems Programmers & Engine Architects:
// Convenience boolean bridge delegating to `PreflightGlobalEgress().IsPass()`.
func AuditGlobalEgress() bool {
	return PreflightGlobalEgress().IsPass()
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK / EXECUTABLE PROCEDURES & LIFECYCLE
// ==============================================================================
// Engine:        FaithNet Adaptive Networking Substrate
// Stratum:       Layer 3A: Physical & Virtual Transport Substrate (Rung 4 Global)
// Module:        cog.transport.rung4_egress
// Mission:       Expose epilogue readiness, 60-col egress cards, and lifecycle
//                construction for EgressGateway.
// ==============================================================================

// ------------------------------------------------------------------------------
// Sub-Assembly 5.1: Executable Procedures, Inspection, Readiness & Cards
// ------------------------------------------------------------------------------

// IsGlobalEgressReady reports whether the Rung 4 EgressGateway passed preflight.
//
// ### For Architects, Operators & Non-Programmers:
// Epilogue confirmation that FaithNet's commercial-web airlock is ready.
//
// ### For Systems Programmers & Engine Architects:
// Matches OmniCode's Block 5 epilogue readiness predicate (`is_*_ready`).
func IsGlobalEgressReady() bool {
	return PreflightGlobalEgress().IsPass()
}

// ExecuteEgressDumpCard renders a 60-column inner-width Rung 4 egress card.
//
// ### For Architects, Operators & Non-Programmers:
// Shows whether global egress is locked or unlocked, plus fetched/blocked counts.
//
// ### For Systems Programmers & Engine Architects:
// Non-mutating diagnostic card builder using lock-free atomic telemetry.
func (g *EgressGateway) ExecuteEgressDumpCard() string {
	lockState := "SEALED_OFFLINE (AllowGlobalEgress=false)"
	if g.IsEnabled() {
		lockState = "OPT_IN_AIRLOCKED (0:AWAIT ONLY)"
	}
	return fmt.Sprintf(
		"+------------------------------------------------------------+\n"+
			"| %-58s |\n"+
			"+------------------------------------------------------------+\n"+
			"| Module:  %-49s |\n"+
			"| Policy:  %-49s |\n"+
			"| MaxBody: %-49s |\n"+
			"| Traffic: %-49s |\n"+
			"+------------------------------------------------------------+",
		"FAITHNET RUNG 4 GLOBAL EGRESS AIRLOCK CARD",
		ModuleIDGlobalEgress,
		lockState,
		fmt.Sprintf("%d bytes (OOM-bounded LimitReader)", data.FloorMaxPayload),
		fmt.Sprintf("%d clean AWAIT / %d blocked", g.fetchedCnt.Load(), g.blockedCnt.Load()),
	)
}

// ------------------------------------------------------------------------------
// Sub-Assembly 5.2: Subsystem Lifecycle Initialization & Teardown
// ------------------------------------------------------------------------------

// NewEgressGateway initializes an Airlock-gated EgressGateway from manifest.
//
// ### For Architects, Operators & Non-Programmers:
// Builds an outside-internet airlock gateway obeying the node's offline policy.
//
// ### For Systems Programmers & Engine Architects:
// Defaults nil `airlock` to `covenant.NewFilter()` and pins `DefaultEgressTimeout`.
func NewEgressGateway(
	manifest data.NodeManifest,
	airlock *covenant.Filter,
) *EgressGateway {
	if airlock == nil {
		airlock = covenant.NewFilter()
	}
	return &EgressGateway{
		manifest: manifest,
		airlock:  airlock,
		client:   &http.Client{Timeout: DefaultEgressTimeout},
	}
}

// End of transport/rung4_global/egress.go
