<!--
================================================================================
BLOCK 1: METADATA BLOCK
( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
================================================================================
@project       FaithNet
@file          README.md
@id            docs.readme.faithnet
@layer         Layer 5: Root Repository Orientation & Covenant Anchor
@archetype     manifest | dual_audience_protocol_manual
@version       0.2.0-dev
@interface_rev Rev 5
@date          2026-11-17
@updated       2026-11-17
@author        CreativeWorkzStudio LLC / ProfessorSeanEX (Seanje Lenox-Wise)
@repository    https://github.com/CreativeWorkzStudio/FaithNet
@system        FaithNet Sovereign Adaptive Networking Substrate (FNP/1)
@role          Plain-English Orientation, Official FNP/1 Spec & Verification Index
@maturity      Rev 5 Baseline | Wave 15 Complete (FNP/1 Port 7777, Bit<->Trit, Sovereign Host)
@chassis       zero-external-deps | Pure Go 1.26.8 Standard Library Substrate
@scale         O(1) orientation across 5 layers, 4 adaptive rungs & 4 test tiers
@seams         in: docs/ROADMAP.adoc, docs/discovery/ | exports: human & engineer entrypoint
@invariants    1. Universal 5-Block Lifecycle Protocol (Rev 5) across all files.
               2. Zero third-party dependencies (100% Go standard library).
               3. Official FNP/1 Wire Protocol on Port 7777 (UDP/TCP) & faithnet:// URI.
               4. 512-byte Floppy Sector Invariant (32B header + 480B Floor MTU).
               5. Bit <-> Balanced-Trit <-> Bit lossless OmniCode submodule bridge.
               6. Downward Dependency Law: Layer 0/1 never import net or os.

PURPOSE & ARCHITECTURAL ROLE:
  Primary root orientation and dual-audience manual for FaithNet — written so
  both everyday non-technical readers (neighbors, creatives, community hosts)
  and deep systems programmers immediately understand WHY FaithNet exists, HOW
  it works in plain English, and EVERY exact wire/protocol invariant underneath.
================================================================================
-->

# FaithNet (`FNP/1`)

> **A sovereign, offline-first networking protocol and micro-ISP substrate built for real people, creative studios, and local communities.**  
> *Compatibility with the modern internet—without ever being held hostage by it.*

---

<!-- =========================================================================
BLOCK 2: OPENING BLOCK / FOR EVERYONE (PLAIN-ENGLISH GUIDE — ZERO JARGON)
Stratum: Root Orientation | Machine: docs.readme.faithnet
========================================================================== -->

## Part 1: What Is FaithNet? *(Plain-English Guide for Everyone)*

Look around at the news right now: wave after wave of cloud and internet outages—AWS, Microsoft, Apple, airlines, banks, local ISPs—knocking whole regions offline from one bad update or router glitch. That happens because the modern internet puts everybody's communication on a handful of distant, centralized corporate towers. When one tower hiccups, two people sitting in the **same house or the same neighborhood** suddenly can't even send each other a message.

**FaithNet flips that upside down—one piece at a time.**

### Who Built This & Why?
**CreativeWorkzStudio LLC** isn't a typical software vendor chasing subscriptions. We are building a **sovereign publishing house**—stories, games (*Tenebris and Lux*), art, and creative works meant to serve people directly. To protect a studio's creative lifeblood and serve a local community during shaky times, you need an in-house foundation you actually own from the ground up:

* **MillenniumOS** — the sovereign operating system ground.
* **OmniCode** — our language, design discipline, and full vertical stack (built around balanced three-way truth: `-1` Contradict, `0` Await/Verify, `+1` Aligned).
* **FaithNet** — the nervous system and networking protocol (`FNP/1`) that lets machines, studios, and local rooms talk **whether the commercial internet is working or completely dead.**

### How FaithNet Actually Works (No Tech Degree Needed)

Think of networking the way normal human communication works:

1. **The 4-Rung Ladder (Never Stranded in Silence):**
   Imagine you need to get a message to your neighbor across the street:
   * **Rung 4 — Global Highway (Internet / TCP):** If the commercial internet is up and you *choose* to bridge through it, FaithNet speaks smoothly across it—while keeping a firewall wall up so strangers can't walk into your house.
   * **Rung 3 — The Room & Neighborhood Street (Local Wi-Fi / `FaithNet-HomeRoom`):** Your laptop acts as its **own built-in local Wi-Fi micro-ISP**. Your phone, a second laptop, or even a PlayStation connects straight to `FaithNet-HomeRoom`—even if the commercial ISP goes dark!
   * **Rung 2 — Next Door (Direct Host Memory / Local IPC):** Programs on the same computer talk directly inside RAM with zero network card required.
   * **Rung 1 — Civilizational Floor (Copper Wire, Radio Tones & 1844 Morse Code):** If Wi-Fi and the power grid fail, FaithNet steps all the way down to a plain copper serial wire, audible walkie-talkie/speaker-to-mic radio tones, or classic Morse code—carrying the **exact same message format**.

2. **The Well (Zero-Install Local Gathering Place):**
   You shouldn't have to install special software just to walk into a local room. When any everyday device (iPhone, Android, Windows, Mac, Linux, or console) connects to `FaithNet-HomeRoom` and opens `https://faithnet.lan:8077` (or clicks `faithnet://faithnet.lan/well`), it lands in **The Well**—a fast, offline local bulletin board, trust-certificate desk, and one-command node pairing station.

3. **The Genesis 3 Airlock (Safety Without Paranoia):**
   Every message coming from an outside browser or untrusted wire passes through our **Genesis 3 Airlock** before touching core data:
   * Trusted, paired FaithNet nodes speak at **`+1:ALIGNED`**.
   * Clean messages from visiting browsers enter safely in **`0:AWAIT`** quarantine so everyone can see where they came from.
   * Hostile web attacks, corrupted wires, or bad scripts are stopped cold (**`-1:CONTRADICT`** / `403 Forbidden`) without ever crashing the room.

4. **Graceful Degradation Over Bricking:**
   Real hardware has physical limits. FaithNet is engineered from day one to **yield safely, contain faults, and step down cleanly** rather than locking up your computer or destroying your existing Wi-Fi.

---

<!-- =========================================================================
BLOCK 3: BODY BLOCK / FOR ENGINEERS & BUILDERS (OFFICIAL FNP/1 PROTOCOL SPEC)
Stratum: Technical Specification | Machine: docs.readme.faithnet
========================================================================== -->

## Part 2: Under the Hood *(Official `FNP/1` Protocol & Architecture Spec)*

| Attribute | Specification |
| :--- | :--- |
| **Organization / Studio** | CreativeWorkzStudio LLC (Sovereign Publishing House & In-House Stack) |
| **Lead Architect** | ProfessorSeanEX (Seanje Lenox-Wise) |
| **Wire Protocol** | **`FNP/1`** (FaithNet Protocol v1 — `0xFA17` canonical 512-byte sector frames) |
| **Canonical Port & URI** | **Port `7777`** (`7777/udp` Rung 3, `7777/tcp` Rung 4), **`8077/tcp`** (The Well), **`faithnet://`** URI scheme |
| **Runtime Language** | Go 1.26.8 (**100% Standard Library**, Zero Third-Party Dependencies) |
| **Truth Space & OmniCode Seam** | Balanced Ternary $\mathbb{T}_4 = \{-1, 0, +1, \text{NULL}\}$ + **Zero-Heap Bit $\leftrightarrow$ Trit $\leftrightarrow$ Bit Codec** (`engine/trit.go`) |
| **Frame Invariant** | 32-Byte Fixed Header + 480-Byte Floor MTU = **512-Byte Floppy Sector Invariant** |
| **Memory Discipline** | `ALLOC_ZERO_HEAP` (`0.00 allocs/op`) on Layer 0/1 wire, heartbeat & trit hot paths |
| **Verification Matrix** | **4x4 Lab Matrix** (`48` test files, `30:30` strict 1:1 unit pairing) + **24/24 live Block 4 gates** + **6 live playtest reports** |
| **Current Milestone** | **Wave 15 Complete** (`FNP/1` Port `7777` & `faithnet://` URI, Bit$\leftrightarrow$Trit$\leftrightarrow$Bit bridge, Sovereign Host Split-DNS, Firewalled WAN Guard, Dedicated USB Radio Auto-Detect) |

### The Official FaithNet Protocol (`FNP/1`), Port `7777` & `faithnet://` URI

FaithNet introduces a self-contained, medium-independent networking protocol—**`FNP/1`**—decoupled from commercial IP ownership, DHCP authorities, and centralized DNS:

* **Canonical Wire Port:** **`7777/udp`** (Rung 3 LAN mesh & beacons) and **`7777/tcp`** (Rung 4 global stream bridge).
* **Legacy Web Bridge Port:** **`8077/tcp`** (dual-mode `0x16` HTTP/HTTPS peek listener for The Well + 3-Tier Adaptive TLS Trust Ladder).
* **Canonical URI Scheme (`engine/uri.go`):**
  ```text
  faithnet://<host-or-64bit-NodeID>[:7777][/action[/target]]

  Examples:
    faithnet://faithnet.lan/well             -> Opens local The Well portal
    faithnet://home-anchor/status            -> Queries live Sovereign Anchor telemetry
    faithnet://0xFA17000000000001:7777/ping  -> Targets sovereign 64-bit NodeID directly
    faithnet://10.77.77.1/ca                 -> Downloads Tier 2 Sovereign Root CA (.crt)
  ```

### 32-Byte Canonical `FNP/1` Wire Frame Layout (`0xFA17`)

```text
Offset   Size   Field         Description
------------------------------------------------------------------------------
[0..1]   2 B    Magic         0xFA, 0x17 ("FAIT[H]")
[2]      1 B    Ver | Rung    High nibble = ProtocolVersion (1); Low = Rung (1..4)
[3]      1 B    Opcode        PING(1), PONG(2), BEACON(3), CPI_SYNC(4), DATA(5)...
[4]      1 B    Flags         CovenantVerified, ExternalOrigin, LegacyTranslated...
[5]      1 B    QuadState     T4 Truth Space: NULL(0), -1(1), 0:AWAIT(2), +1(3)
[6..7]   2 B    PayloadLen    Big-endian uint16 (0..480 bytes)
[8..15]  8 B    SourceID      Big-endian uint64 sovereign originator NodeID
[16..23] 8 B    TargetID      Big-endian uint64 sovereign destination NodeID
[24..27] 4 B    SeqID         Big-endian uint32 monotonic per-origin sequence
[28..31] 4 B    Checksum      Big-endian uint32 FNV-1a seal (header[0..27]+payload)
[32..N] <=480B  Payload       Inline zero-heap payload (total wire frame <= 512 bytes)
```

### Bit $\leftrightarrow$ Balanced-Trit $\leftrightarrow$ Bit OmniCode Submodule Bridge (`engine/trit.go`)

When FaithNet mounts inside **OmniCode** as a native submodule, binary physical wires (Wi-Fi, Ethernet, UART, PCM audio) must translate cleanly to and from OmniCode's balanced ternary world ($\{-1, 0, +1\}$) without heap churn or data loss:

* **Compact Ternary Packing (`Tryte5`):** `5` balanced trits = $3^5 = 243$ signed states (`-121..+121`) packed cleanly into 1 binary byte (`0x00..0xF2`), leaving `13` reserved hardware/control sentinels (`0xF3..0xFF`, with `0xFF = NULL:HALT`).
* **Lossless Binary Bridge (`BytesToBalancedTrits` / `BalancedTritsToBytes`):** Any arbitrary 8-bit binary byte (`0..255`) maps bijectively onto `6` balanced trits ($3^6 = 729 > 256$) and back bit-for-bit with **`0.00 allocs/op`**.
* **$\mathbb{T}_4$ Isolation:** `QuadStateToTrit` explicitly preserves `StateNull` isolation so missing/halted state is never confused with neutral `0:AWAIT`.

### The 5 Irreducible Cogs & 4-Rung Adaptive Ladder

1. **Cog 1 — Sovereign Identity (`data/types.go`):** 64-bit `NodeID` independent of DHCP, DNS, IPv4/IPv6, or ISP registries.
2. **Cog 2 — Canonical Envelope & Codecs (`data/frame.go`, `engine/codec.go`, `engine/trit.go`, `engine/uri.go`):** 512-byte sector wire frames + FNV-1a seal + Bit$\leftrightarrow$Trit$\leftrightarrow$Bit bridge + `faithnet://` URI parser.
3. **Cog 3 — Polymorphic Transceiver (`transport/transceiver.go`):** Medium-agnostic `Send` / `Receive(ctx)` / `Healthy` / `Close` wall-plug across all 4 rungs (`[5][4]` multi-sibling bounded chassis).
4. **Cog 4 — Offline Monotonic Clock (`engine/clock.go`):** NTP-independent hardware stopwatch + deterministic lock-free `MockClock`.
5. **Cog 5 — Heartbeat & Echo Engine (`engine/ping.go`):** 64-slot power-of-two in-flight ring + zero-alloc `ReflectPingInto` mirror.

### Sovereign Host Profile (Wave 15 — Laptop as Primary Micro-ISP, Commercial Internet Firewalled)

* **Inside-Out Host Design:** `FaithNet-HomeRoom` (`10.77.77.1/24`) is this PC's primary sovereign room identity with local `systemd-resolved` split-DNS (`~faithnet.lan`, `~faithnet.home.arpa`) and host-trusted Tier 2 Sovereign Root CA.
* **Firewalled Rung 4 WAN (`FAITHNET_WAN_GUARD`):** Unsolicited inbound probes arriving on commercial Wi-Fi (`wlp3s0`) are dropped at the kernel boundary while `ap0` and `lo` stay open for authenticated HomeRoom peers.
* **Instant `bridge` / `island` Egress Switch (`faithnet-netctl`):** Daily-driver `bridge` mode allows outbound internet/NAT while keeping inbound firewalled; `island` mode severs commercial WAN egress cleanly on command to verify 100% offline local operation.
* **Dedicated 2nd USB Wi-Fi Adapter Auto-Detect:** Plugging in any Linux AP-capable USB Wi-Fi adapter (`MT7612U`, `MT7921AU`, `AR9271`) automatically switches `faithnet-ap.service` from `SHARED_RADIO` (`STA+AP`) to **`DEDICATED_RADIO`** mode—zero blink even while `wlp3s0` switches hotspots!

### Layered Repository Map (Downward Dependency Law)

Imports flow strictly downward (`Layer 4 -> 3 -> 2 -> 1 -> 0`). **Layer 0 (`data/`) and Layer 1 (`engine/`) NEVER import `net`, `os`, or external packages.**

```text
FaithNet/
├── data/                       # Layer 0: Pure Invariant Vocabulary (5 files)
├── engine/                     # Layer 1: Pure Engine Core — ALLOC_ZERO_HEAP (7 files)
│   ├── codec.go                #   Zero-alloc FNP/1 32B Wire Codec & FNV-1a seal
│   ├── trit.go                 #   Zero-alloc Bit <-> Balanced-Trit <-> Bit OmniCode bridge
│   ├── uri.go                  #   Canonical FNP/1 Port 7777 & faithnet:// URI parser
│   ├── clock.go, ping.go, router.go & covenant/filter.go (Genesis 3 Airlock)
├── orchestrator/               # Layer 2: BodyProbe, [5][4] AdaptiveLadder, Dispatcher
├── transport/                  # Layer 3A: Rung 1 (SLIP/AFSK/Morse), Rung 2 (RAM/IPC), Rung 3 (UDP/Beacon), Rung 4 (TCP/Egress)
├── interpreter/                # Layer 3B: "The Well" Dual-Mode TLS Portal, /pair.sh, /pair.ps1, OS Classifier
├── cmd/                        # Layer 4: faithnet-ping & faithnet-node CLIs (24 live Block 4 gates)
├── docs/                       # Layer 5: Roadmap, 5-Volume Core Series & Discovery Series
│   ├── ROADMAP.adoc & 01..05_*.adoc
│   └── discovery/              #   D1: Field Kit 01 Lessons | D2: Safety & Degradation | D3: Live vs. Lab Benchmarks
└── tests/                      # Layer 4: 4x4 Lab Matrix (48 files, 30:30 1:1 units) + Playtester Proving Ground
    └── playtester/             #   5 live harnesses, cross-platform Field Kits (Linux/Win/macOS) & 6 live reports
```

---

<!-- =========================================================================
BLOCK 4: AUDIT PROCEDURES BLOCK / QUICK START, CONTROLS & LIVE BENCHMARKS
Stratum: Verification | Machine: docs.readme.faithnet
========================================================================== -->

## Quick Start, Sovereign Controls & Verification Suite

### Everyday Operator Controls (Linux Sovereign Host Profile)

```bash
# Inspect Sovereign Host status (Radio mode, Split-DNS, WAN Firewall, connected devices)
faithnet-netctl status

# Scan physical Wi-Fi radios (checks if a dedicated 2nd USB Wi-Fi adapter is plugged in)
faithnet-netctl adapters

# Toggle commercial internet egress cleanly (zero reboot, zero HomeRoom drop)
sudo faithnet-netctl mode island   # True air-gapped sovereign room mode
sudo faithnet-netctl mode bridge   # Daily-driver mode (inbound WAN still firewalled!)

# Open native faithnet:// protocol links on this machine
faithnet-open faithnet://faithnet.lan/status
faithnet-open faithnet://faithnet.lan/well
```

### Full Developer Verification Battery

```bash
# 1. Confirm zero external dependencies & strict Layer 0/1 Downward Dependency Law
go list -m all
go list -f '{{.ImportPath}} -> {{.Imports}}' ./data ./engine ./engine/covenant

# 2. Run static analysis & ALL 4 verification tiers with -race
go vet ./...
go test -v -count=1 -race ./tests/...

# 3. Run ALL 24 live Block 4 Pre-Flight Audit gates across every subsystem & CLI
go run ./cmd/faithnet-node --audit

# 4. Run live offline 5-Cog loopback ping (Rung 2 RAM or Unix IPC)
go run ./cmd/faithnet-ping --mode=memory --count=4
go run ./cmd/faithnet-ping --mode=ipc --count=4
```

### Verified Reality Matrix (Lab 4x4 vs. Harnesses vs. Live Physical Field Kit 01)

| Capability | Lab 4x4 Verified | Local-OS Harnesses (S002) | Physical Field Status (S003–S006 & Wave 15) |
| :--- | :--- | :--- | :--- |
| **`FNP/1` Codec & Bit$\leftrightarrow$Trit** | `50.89 ns` (`0 B/op`), `243` trytes + `256` bytes (`0` allocs) | `8/8` storm bit-flips caught (`100%`) | Live `+1:ALIGNED` mesh frames across Wi-Fi; `faithnet://` & Port `7777` registered |
| **Rung 3 Wi-Fi Mesh & Pairing** | PASS (`-race`) | UDP `p50=39.1 us`, TCP `p50=338.4 us` | **LIVE VERIFIED:** 2-laptop mesh (`peer_count=1`, `11.1 MB` `/pair.sh`, `~3s` rejoin) + `/pair.ps1` (`pwsh 7.7.0` tested) |
| **Micro-ISP Router & WAN Guard** | PASS (`-race`) | Hard-cut `1.28 ms`; silent blackhole auto-rescued `2/2` | **LIVE VERIFIED:** `~10–20 MB` RSS, `>36 MB` routed across Laptop #2, iPhone & PlayStation YouTube; zero-lockup hotspot roaming; Split-DNS + `FAITHNET_WAN_GUARD` + USB radio auto-detect |
| **The Well & 3-Tier TLS** | PASS (`-race`) | `6,126 req/s`, `120/120` attacks blocked (`0` breaches) | **LIVE VERIFIED:** Tier 2 Sovereign Root CA green padlock on Laptop #1, Laptop #2 & iPhone (`https://faithnet.lan:8077`) |
| **Rung 1 Floor (AFSK/Morse/SLIP)** | PASS (`-race`) | AFSK `+3.93 dB` SNR & `0..4,000` mic silence `6/6` PASS; FIFO SLIP `9.23 Mbps` | Staged in Field Kits 02 (Acoustic Air-Gap) & 03 (Copper Serial) |

### Documentation & Discovery Series (`docs/` & `docs/discovery/`)

1. [Vol 01: What Is FaithNet](docs/01_WHAT_IS_FAITHNET.adoc) — Philosophy, CPI-SI & the 5 Irreducible Cogs.
2. [Vol 02: Architecture & Code Walkthrough](docs/02_ARCHITECTURE_AND_CODE_WALKTHROUGH.adoc) — Layer-by-layer code guide & 5-Block standard.
3. [Vol 03: FaithNet vs. Modern Internet](docs/03_FAITHNET_VS_MODERN_INTERNET.adoc) — Comparison vs. commercial internet stack.
4. [Vol 04: Playtesting Benchmarks & Gaps](docs/04_PLAYTESTING_BENCHMARKS_AND_GAPS.adoc) — Empirical Session 001/002 harness baselines.
5. [Vol 05: Field Playtest Runbooks](docs/05_FIELD_PLAYTEST_RUNBOOKS.adoc) — Operator runbooks for Field Kits 01–04.
6. [Discovery 01: Field Kit 01 Experience & Lessons](docs/discovery/01_FIELD_KIT_01_EXPERIENCE_AND_LESSONS.adoc) — 9 real moving-part edge cases & cross-platform OS parity.
7. [Discovery 02: Safety Locks & Graceful Degradation](docs/discovery/02_SAFETY_AND_GRACEFUL_DEGRADATION.adoc) — Why containment & graceful degradation beat hard aborts.
8. [Discovery 03: Live vs. Lab Benchmark Comparison](docs/discovery/03_LIVE_VS_LAB_BENCHMARK_COMPARISON.adoc) — Side-by-side empirical ledger across Lab, Harnesses & Live RF.

---

<!-- =========================================================================
BLOCK 5: CLOSING BLOCK / COVENANT AFFIRMATIONS
Stratum: Governance | Machine: docs.readme.faithnet
========================================================================== -->

## Covenant Invariants

* **Constraint Is Not Limitation:** Strict physical boundaries (512-byte floor frames, zero-alloc hot paths, zero third-party packages) yield eternal portability from bare metal to modern machines.
* **Compatibility Without Dependency:** FaithNet speaks fluently with Windows, macOS, Linux, iOS, Android, consoles, and legacy browsers—without ever being enslaved to commercial clouds or ISPs.
* **Genesis 3 Airlock:** Untrusted external data never enters sovereign core state as `+1:ALIGNED`; clean facts enter strictly in `0:AWAIT` quarantine while hostile payloads halt as `-1:CONTRADICT` or `NULL:HALT`.
* **Canonical Roadmap:** [docs/ROADMAP.adoc](docs/ROADMAP.adoc)

<!-- End of README.md -->
