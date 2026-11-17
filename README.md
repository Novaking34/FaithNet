<!--
================================================================================
BLOCK 1: METADATA BLOCK
( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
================================================================================
@project       FaithNet
@file          README.md
@id            docs.readme.faithnet
@layer         Layer 5: Root Repository Orientation & Covenant Anchor
@archetype     manifest | operational
@version       0.1.0-dev
@interface_rev Rev 5
@date          2026-11-17
@updated       2026-11-17
@author        CreativeWorkzStudio LLC / ProfessorSeanEX
@repository    https://github.com/CreativeWorkzStudio/FaithNet
@system        FaithNet Sovereign Adaptive Networking Substrate
@role          Root Orientation, Architectural Specification & Verification Index
@maturity      Rev 5 Baseline | Wave 8 Complete (Session 002 All Gaps Closed)
@chassis       zero-external-deps | Pure Go 1.26.8 Standard Library Substrate
@scale         O(1) orientation across 5 layers, 4 adaptive rungs & 4 test tiers
@seams         in: docs/ROADMAP.adoc | exports: operator & engineer entrypoint
@invariants    1. Universal 5-Block Lifecycle Protocol (Rev 5) across all files.
               2. Zero third-party dependencies (100% Go standard library).
               3. 512-byte Floppy Sector Invariant (32B header + 480B Floor MTU).
               4. Downward Dependency Law: Layer 0/1 never import net or os.
               5. 4x4 Macro-Micro Verification Matrix with structured Block 4
                  `data.AuditResult` preflight gates on every operational cog.

PURPOSE & ARCHITECTURAL ROLE:
  Primary root orientation and architectural manual for FaithNet, the sovereign
  offline-first adaptive networking substrate for OmniCode and MillenniumOS.

  ### For Architects, Operators & Non-Programmers:
  Explains what FaithNet is, why it can communicate over everything from analog
  radio/Morse tones to local Wi-Fi without relying on commercial ISPs, how the
  Genesis 3 Airlock protects local nodes, and how to run live offline ping
  diagnostics right now.

  ### For Systems Programmers & Engine Architects:
  Documents the 5 Irreducible Cogs, the 32-byte wire header layout, the 4-Rung
  Adaptive Ladder, layered package boundaries, Go Block 4 `data.AuditResult`
  pre-flight standards, 4x4 Macro-Micro testing layout, and wave milestones.
================================================================================
-->

# FaithNet

> **Sovereign, First-Principles, Data-Driven Networking Substrate for OmniCode & MillenniumOS.**  
> *From a scrap of copper wire and analog radio tones to a federated local mesh — compatibility with the modern world without dependency on it.*

| Attribute | Specification |
| :--- | :--- |
| **Organization** | CreativeWorkzStudio LLC |
| **Lead Architect** | ProfessorSeanEX (Seanje Lenox-Wise) |
| **Runtime Language** | Go 1.26.8 (**100% Standard Library**, Zero Third-Party Dependencies) |
| **Truth Space** | Balanced Ternary Quad-State ($\mathbb{T}_4 = \{-1, 0, +1, \text{NULL}\}$) |
| **Frame Invariant** | 32-Byte Fixed Header + 480-Byte Floor MTU = **512-Byte Floppy Sector Invariant** |
| **Memory Discipline** | `ALLOC_ZERO_HEAP` (0.00 allocs/op) on Layer 0/1 wire & heartbeat hot paths |
| **Adaptive Resilience** | 4-Rung Bidirectional Ladder (`Rung 1 Floor` $\leftrightarrow$ `Rung 4 Global`) |
| **Verification Matrix** | **4x4 Lab Matrix** (46 test files, 329 tests) + **`tests/playtester/`** (5 live harnesses, 4 field kits) |
| **Current Milestone** | **Wave 8 Complete** (All 3 Session 001 Gaps Closed & Verified in Session 002; Field-Test Ready) |
| **Documentation Series** | [`docs/ROADMAP.adoc`](docs/ROADMAP.adoc) & [`docs/01..05`](docs/01_WHAT_IS_FAITHNET.adoc) |

---

<!-- =========================================================================
BLOCK 2: OPENING BLOCK / FIRST PRINCIPLES: 5 COGS & 4-RUNG ADAPTIVE LADDER
Stratum: Root Orientation | Machine: docs.readme.faithnet
========================================================================== -->

## First-Principles Architecture

### The 5 Irreducible Cogs of Sovereign Communication

Built on OmniCode's first-principles discipline, FaithNet decomposes networking into five self-contained, portable cogs:

1. **Cog 1 — Sovereign Identity (`data/types.go`):** 64-bit `NodeID` independent of DHCP, DNS, IPv4/IPv6, or ISP registries (`0x0` null sentinel, `0xFFFFFFFFFFFFFFFF` broadcast).
2. **Cog 2 — Canonical Envelope & Codec (`data/frame.go`, `engine/codec.go`):** Bit-exact 32-byte fixed header + up to 480-byte inline payload = **512-Byte Floppy Sector Invariant**, sealed with inline zero-heap FNV-1a 32-bit checksums.
3. **Cog 3 — Polymorphic Transceiver (`transport/transceiver.go`):** Medium-agnostic `Send` / `Receive(ctx)` / `Healthy` / `Close` wall-plug decoupling upper layers from RAM, IPC, UDP, TCP, serial wire, AFSK audio, or Morse CW.
4. **Cog 4 — Offline Monotonic Clock (`engine/clock.go`):** NTP-independent hardware monotonic stopwatch (`MonotonicClock`) paired with an atomic lock-free `MockClock` for deterministic zero-flake laboratory testing.
5. **Cog 5 — Heartbeat & Echo Engine (`engine/ping.go`):** Fixed 64-slot power-of-two in-flight ring (`PingEngine`) + stateless zero-alloc `ReflectPingInto` mirror computing exact min/avg/max RTT and packet loss.

### 32-Byte Canonical Wire Frame Layout (`0xFA17`)

```text
Offset   Size   Field         Description
------------------------------------------------------------------------------
[0..1]   2 B    Magic         0xFA, 0x17 ("FAIT[H]")
[2]      1 B    Ver | Rung    High nibble = ProtocolVersion (1); Low = Rung (1..4)
[3]      1 B    Opcode        PING(1), PONG(2), BEACON(3), CPI_SYNC(4), ...
[4]      1 B    Flags         CovenantVerified, ExternalOrigin, LegacyTranslated, ...
[5]      1 B    QuadState     T4 Truth Space: NULL(0), -1(1), 0:AWAIT(2), +1(3)
[6..7]   2 B    PayloadLen    Big-endian uint16 (0..480 bytes)
[8..15]  8 B    SourceID      Big-endian uint64 sovereign originator NodeID
[16..23] 8 B    TargetID      Big-endian uint64 sovereign destination NodeID
[24..27] 4 B    SeqID         Big-endian uint32 monotonic per-origin sequence
[28..31] 4 B    Checksum      Big-endian uint32 FNV-1a seal (over header[0..27]+payload)
[32..N] <=480B  Payload       Inline zero-heap payload (total wire frame <= 512 bytes)
```

### The 4-Rung Adaptive Networking Ladder

Mirroring Track 0 of **MillenniumOS** (Legacy BIOS $\leftrightarrow$ Stage 2 $\leftrightarrow$ Multiboot $\leftrightarrow$ UEFI), FaithNet dynamically probes whatever host machine it boots onto, **Rises Up** when richer hardware is available, and executes a **Defensive Cascade (`EVT_CASCADE_DOWN`)** when links fail:

| Rung | Medium | Purpose | Wave Status |
| :--- | :--- | :--- | :--- |
| **Rung 1 — Civilizational Floor** | `transport/rung1_floor/` (`serial` SLIP, `audio` 1200/2400 Hz AFSK + Goertzel, `morse` ITU CW) | Copper wire, dial-up tones, analog walkie-talkie/radio, 1844 telegraph floor | **Implemented & Verified (Wave 5)** |
| **Rung 2 — Host Substrate** | `transport/rung2_host/` (`memory`, `ipc` unixgram) | Zero-NIC in-RAM channels & local cross-process IPC | **Implemented & Verified (Wave 2)** |
| **Rung 3 — The Room & The Well** | `transport/rung3_local/` (`udp`, `beacon`) + `interpreter/` | Peer-to-peer LAN UDP, Wi-Fi AP switch, zero-install HTML portal | **Implemented & Verified (Wave 4)** |
| **Rung 4 — Global Compatibility** | `transport/rung4_global/` (`tcp`, `egress`) + `engine/covenant/` | Optional TCP federation & Airlock-gated outbound egress | **Implemented & Verified (Wave 5)** |

---

<!-- =========================================================================
BLOCK 3: BODY BLOCK / LAYERED REPOSITORY MAP & PROTOCOL GOVERNANCE
Stratum: Root Orientation | Machine: docs.readme.faithnet
========================================================================== -->

## Layered Repository Map (Downward Dependency Law)

Imports flow strictly downward (`Layer 4 -> 3 -> 2 -> 1 -> 0`). **Layer 0 (`data/`) and Layer 1 (`engine/`) NEVER import `net`, `os`, or external packages.**

```text
FaithNet/
├── data/                       # Layer 0: Pure Invariant Vocabulary (pure_types)
│   ├── types.go                #   NodeID, RungLevel (StepUp/StepDown), QuadState (T4)
│   ├── opcodes.go              #   Opcode verbs (0x01..0x08) & orthogonal FrameFlags
│   ├── frame.go                #   32B Header + 480B inline Frame (512B Floppy Sector)
│   ├── profile.go              #   HardwareProfile, NodeManifest, LinkHealth, CascadeEvent
│   └── audit.go                #   Structured Block 4 AuditSeverity & AuditResult verdicts
├── engine/                     # Layer 1: Pure Functional State Machines (ALLOC_ZERO_HEAP)
│   ├── codec.go                #   Zero-alloc EncodeInto / DecodeFrom / FNV-1a seal
│   ├── clock.go                #   MonotonicClock (NTP-free) & atomic MockClock
│   ├── ping.go                 #   64-slot PingEngine + stateless ReflectPingInto
│   ├── router.go               #   Fixed 128-peer Router (direct + multi-hop relays)
│   └── covenant/filter.go      #   Genesis 3 Airlock (T4 screening: +1 / 0:AWAIT / -1 / NULL)
├── orchestrator/               # Layer 2: Dynamic BodyProbe, AdaptiveLadder, Dispatcher
│   ├── probe.go                #   Live OS/NIC/serial/audio hardware probe & policy ceiling
│   ├── ladder.go               #   4-Rung Rising-Up & EVT_CASCADE_DOWN (64-slot ring)
│   └── dispatcher.go           #   Central node loop: Codec -> Airlock -> Router -> Ping
├── transport/                  # Layer 3A: Cog 3 Transceiver & 4-Rung Physical/Virtual Wires
│   ├── transceiver.go          #   Universal Transceiver interface & WireEnvelope
│   ├── rung1_floor/            #   SLIP Serial, 1200/2400Hz AFSK Goertzel Modem, ITU Morse CW
│   ├── rung2_host/             #   MemoryBus + Unixgram IPCTransceiver
│   ├── rung3_local/            #   Auto-learning LAN UDP & Wi-Fi AP Beacon Switch
│   └── rung4_global/           #   Length-prefixed TCP Stream & Airlock EgressGateway
├── interpreter/                # Layer 3B: "The Well" Universal OS/Web Portal Bridge
│   ├── translator.go           #   Airlock-gated OpLegacyWrap translator (0:AWAIT only)
│   ├── captive_web.go          #   Zero-install offline HTML portal & 32-bulletin ring
│   └── legacy_bridge.go        #   Data-driven OS classifier (Win/Mac/Linux/iOS/Android/CLI)
├── cmd/                        # Layer 4: Thin Imperative CLI Entrypoints
│   ├── faithnet-ping/          #   Standalone offline 5-Cog ping & --audit (9 gates) CLI
│   └── faithnet-node/          #   Full orchestrated node, --audit (22 gates) & The Well CLI
├── docs/                       # Layer 5: Canonical Roadmap & 5-Volume AsciiDoc Series
│   ├── ROADMAP.adoc            #   Master 5-Block Architecture & Wave Plan
│   ├── 01_WHAT_IS_FAITHNET.adoc                # Vol 01: Philosophy, CPI-SI & The 5 Cogs
│   ├── 02_ARCHITECTURE_AND_CODE_WALKTHROUGH.adoc # Vol 02: Layer-by-Layer Code & 5-Block Guide
│   ├── 03_FAITHNET_VS_MODERN_INTERNET.adoc     # Vol 03: Comparison vs Commercial Internet Stack
│   ├── 04_PLAYTESTING_BENCHMARKS_AND_GAPS.adoc # Vol 04: Empirical Benchmarks & Honest Gaps
│   └── 05_FIELD_PLAYTEST_RUNBOOKS.adoc         # Vol 05: Physical Field Playtest Runbooks
└── tests/                      # Layer 4: 4x4 Lab Matrix + Playtester Proving Ground
    ├── common/fixtures.go      #   Shared identities, forensic AuditResult asserts & anchors
    ├── functional/             #   Tier 1: 7 per-subsystem public API suites + benchmarks
    ├── unit/                   #   Tier 2: 28 strict 1:1 per-cog unit suites (all 28 prod files)
    ├── integrated/             #   Tier 3: 6 cross-subsystem multi-combination suites
    ├── system/                 #   Tier 4: 5 full-stack E2E, blackout, hostility, CLI & stress suites
    └── playtester/             #   Proving Ground: 5 live OS/chaos/audio harnesses, 4 field kits & reports
```

### Universal 5-Block Lifecycle Protocol (Rev 5) & Go Block 4 Standard

Every source, test, build, and documentation file in FaithNet adheres to OmniCode Rev 5:

* **Block 1 (`METADATA BLOCK`):** Movement-resilient `@id`, `@layer`, `@archetype`, `@chassis`, `@invariants`, and Dual-Audience `PURPOSE & ARCHITECTURAL ROLE`.
* **Block 2 (`OPENING BLOCK`):** Tiered imports, constants/limits (`2.1`), logical anchor (`2.2`), public ontology (`2.3`), pure O(1) predicates (`2.4`), and monitored support infrastructure (`2.5`).
* **Block 3 (`BODY BLOCK`):** Composable non-lifecycle business logic (or Statutory Pure-Types Notice in `data/`), with Dual-Audience docstrings on every exported symbol.
* **Block 4 (`AUDIT PROCEDURES BLOCK`):**
  * **Sub-Assembly 4.1:** Non-destructive live-instance & contract validators returning structured `data.AuditResult` (`Passed`, `Severity`, `Subsystem`, `CheckName`, `Details`).
  * **Sub-Assembly 4.2:** `Preflight<Cog>() data.AuditResult` running all **4 Micro Checks** (1: Always Pass, 2: Never Fail, 3: Invariants/Limits, 4: Emergence/Reset), with `@triggers` linking to Layer 4 tests.
* **Block 5 (`CLOSING BLOCK`):** `Is<Cog>Ready() bool` epilogue readiness handshake, 60-column `+---+` diagnostic dump cards (`5.1`), and lifecycle `New*` / `Reset` / `Close` (`5.2`).

---

<!-- =========================================================================
BLOCK 4: AUDIT PROCEDURES BLOCK / VERIFICATION SUITE & LIVE CLI USAGE
Stratum: Verification | Machine: docs.readme.faithnet
========================================================================== -->

## Verification Suite & Live CLI Usage

```bash
# 1. Confirm zero third-party dependencies
go list -m all

# 2. Confirm Downward Dependency Law (zero net/os imports in Layer 0/1)
go list -f '{{.ImportPath}} -> {{.Imports}}' ./data ./engine ./engine/covenant

# 3. Run static analysis across the entire repository
go vet ./...

# 4. Execute ALL 4 verification tiers (Functional, Unit, Integrated, System) with -race
go test -v -count=1 -race ./tests/...

# 5. Verify ALLOC_ZERO_HEAP hot paths (0 B/op, 0 allocs/op)
go test -bench=. -benchmem ./tests/functional/...

# 6. Run ALL 22 live Block 4 Pre-Flight Audit gates across every subsystem & CLI
go run ./cmd/faithnet-node --audit

# 7. Run live offline 5-Cog loopback ping (Rung 2 RAM or local Unix IPC)
go run ./cmd/faithnet-ping --mode=memory --count=4
go run ./cmd/faithnet-ping --mode=ipc --count=4

# 8. Inspect live host body, adaptive ladder, Room Beacon & The Well cards
go run ./cmd/faithnet-node --probe-only --beacon --well
```

### Verified Baseline Telemetry & Reality Matrix (Wave 8 / Session 002)

* **External Dependencies:** `0` (`go list -m all` reports only `github.com/CreativeWorkzStudio/FaithNet`).
* **Wire Codec Throughput:** `BenchmarkEncodeInto` `50.89 ns/op` (`0 B/op, 0 allocs/op`), `BenchmarkDecodeFrom` `86.72 ns/op` (`0 B/op, 0 allocs/op`).
* **4x4 Lab Matrix (`tests/{functional,unit,integrated,system}`):** **46 dedicated 5-Block test files** (`7` functional, `28` strict 1:1 unit, `6` integrated, `5` system) + `1` shared harness = **330 verified `Test*` functions** (`go test -count=1 -race ./tests/...` PASS) + **22/22 live Block 4 pre-flight gates**.
* **Playtester Proving Ground ([`SESSION_002_WAVE8_HARDENED.adoc`](tests/playtester/reports/SESSION_002_WAVE8_HARDENED.adoc)):**

| Capability | Lab Verified (4x4) | Live Playtested (Session 002 — Real OS/Audio/HTTP) | Physical Field Status |
| :--- | :--- | :--- | :--- |
| **Cross-Process UDP / TCP** | PASS (`-race`) | **PASS:** UDP `p50=39.1 us` (`0.85 MB`), TCP `p50=338.4 us` (`1.28 MB`), `0%` loss | LAN broadcast `OpBeacon` + `--interactive` REPL ready for Laptop #2 (`192.168.1.99`) |
| **Chaos Weather & Cascade** | PASS (`-race`) | **PASS:** `8/8` bit-flips caught; `UDP->RAM` hard-cut `21 us`; **Silent Blackhole auto-cascade rescued `2/2` on `RUNG_2_HOST`!** | Field Kit 04 (`ip netns` + `tc netem`) staged |
| **Rung 1 SLIP Serial** | PASS (`-race`) | **PASS:** Real kernel FIFO `1,126.6 KB/s` (`9.23 Mbps`), glitch resync `126.9 us` | Field Kit 03 (`socat` installed + `/dev/ttyUSB0` copper checklist) |
| **Rung 1 AFSK & Morse** | PASS (`-race`) | **PASS:** Survives `+3.93 dB` SNR, `1%` vol, `ffmpeg` 32k MP3 & phone bandpass; **mic leading silence `0..4,000` samples `6/6` PASS!** | Field Kit 02 (`aplay`/`arecord`) ready for live room-air test |
| **The Well & Airlock** | PASS (`-race`) | **PASS:** `3,375–4,887 req/s`, `p50=2.1–2.9 ms`, `120/120` attacks blocked (`0` breaches) | Field Kit 01 ready for live phone / Laptop #2 browser test |

### Documentation Series (`docs/`)

1. [`docs/01_WHAT_IS_FAITHNET.adoc`](docs/01_WHAT_IS_FAITHNET.adoc) — Philosophy, first principles, CPI-SI communication & the 5 Irreducible Cogs.
2. [`docs/02_ARCHITECTURE_AND_CODE_WALKTHROUGH.adoc`](docs/02_ARCHITECTURE_AND_CODE_WALKTHROUGH.adoc) — Layer-by-layer code walkthrough, wire format, `[5][4]` ladder & Rev 5 5-Block standard.
3. [`docs/03_FAITHNET_VS_MODERN_INTERNET.adoc`](docs/03_FAITHNET_VS_MODERN_INTERNET.adoc) — Direct architectural comparison vs. the commercial internet & why the compiled C/Rust/Go trinity matters.
4. [`docs/04_PLAYTESTING_BENCHMARKS_AND_GAPS.adoc`](docs/04_PLAYTESTING_BENCHMARKS_AND_GAPS.adoc) — Consolidated empirical benchmarks & honest gaps exposed by Playtest Session 001.
5. [`docs/05_FIELD_PLAYTEST_RUNBOOKS.adoc`](docs/05_FIELD_PLAYTEST_RUNBOOKS.adoc) — Ready-to-execute operator runbooks for all 4 staged physical Field Kits.

---

<!-- =========================================================================
BLOCK 5: CLOSING BLOCK / COVENANT AFFIRMATIONS
Stratum: Governance | Machine: docs.readme.faithnet
========================================================================== -->

## Covenant Invariants

* **Constraint is Not Limitation:** Strict physical boundaries (512-byte floor frames, zero-alloc hot paths, zero third-party packages) yield eternal portability from bare metal to modern clusters.
* **Compatibility Without Dependency:** FaithNet can speak with Windows, macOS, Linux, iOS, Android, and the legacy web without ever being enslaved to them.
* **Genesis 3 Airlock:** Untrusted external data never enters sovereign core state as `+1:ALIGNED`; clean empirical facts enter strictly in `0:AWAIT` quarantine while hostile payloads are halted as `-1:REJECT` or `NULL:HALT`.
* **Canonical Architecture Roadmap:** [`docs/ROADMAP.adoc`](docs/ROADMAP.adoc)

<!-- End of README.md -->
