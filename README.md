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
@role          Root Orientation, Architectural Overview & Verification Index
@maturity      Rev 5 Baseline | Wave 0 Genesis
@chassis       zero-external-deps | Pure Go Standard Library Substrate
@scale         O(1) orientation across 6 layers, 4 adaptive rungs & 4 test tiers
@seams         in: docs/ROADMAP.adoc | exports: operator & engineer entrypoint
@invariants    1. Universal 5-Block Lifecycle Protocol enforced across files.
               2. Zero third-party dependencies (100% Go standard library).
               3. 512-byte Floppy Sector Invariant (32B header + 480B Floor MTU).
               4. Downward dependency flow: Layer 0/1 never import net or OS.

PURPOSE & ARCHITECTURAL ROLE:
  Primary root orientation guide for FaithNet, the sovereign offline-first
  networking substrate for OmniCode and MillenniumOS.

  ### For Architects, Operators & Non-Programmers:
  Explains what FaithNet is, how it enables sovereign local communication
  without commercial ISP dependency, and how to run standalone ping and node
  demonstrations offline.

  ### For Systems Programmers & Engine Architects:
  Summarizes the 4-Rung Adaptive Networking Ladder, the 5 Irreducible Cogs,
  layered package boundaries, and exact verification commands across all
  4 test tiers (`functional/`, `unit/`, `integrated/`, `system/`).
================================================================================
-->

# FaithNet

> **Sovereign, First-Principles, Data-Driven Networking Substrate for OmniCode & MillenniumOS.**  
> *From a scrap of copper wire and analog radio tones to a federated local mesh — zero dependency on commercial telecom monopolies.*

| Attribute | Specification |
| :--- | :--- |
| **Organization** | CreativeWorkzStudio LLC |
| **Lead Architect** | ProfessorSeanEX (Seanje Lenox-Wise) |
| **Runtime Language** | Go 1.26.8 (100% Standard Library, Zero External Dependencies) |
| **Truth Space** | Balanced Ternary Quad-State ($\mathbb{T}_4 = \{-1, 0, +1, \text{NULL}\}$) |
| **Frame Invariant** | 32-Byte Fixed Header + 480-Byte Floor MTU = **512-Byte Floppy Sector Invariant** |
| **Adaptive Resilience** | 4-Rung Bidirectional Ladder (`Rung 1 Floor` $\leftrightarrow$ `Rung 4 Global`) |
| **Verification Suite** | 4-Tier Macro Suite (`tests/functional`, `tests/unit`, `tests/integrated`, `tests/system`) |
| **Canonical Roadmap** | [`docs/ROADMAP.adoc`](docs/ROADMAP.adoc) |

---

<!-- =========================================================================
BLOCK 2: OPENING BLOCK / THE 4-RUNG ADAPTIVE NETWORKING LADDER
Stratum: Root Orientation | Machine: docs.readme.faithnet
========================================================================== -->

## The 4-Rung Adaptive Networking Ladder

Mirroring Track 0 of **MillenniumOS** (Legacy BIOS $\leftrightarrow$ Stage 2 $\leftrightarrow$ Multiboot $\leftrightarrow$ UEFI), FaithNet probes the host machine dynamically, **Rises Up** when rich hardware is present, and executes a **Defensive Cascade (`EVT_CASCADE_DOWN`)** when links fail or hardware degrades:

1. **Rung 1 — The Civilizational Floor (`transport/rung1_floor/`):** SLIP-framed RS-232 serial / copper wire, AFSK audio-frequency modulation (dial-up / analog walkie-talkie / CB radio), and International Morse Code (CW).
2. **Rung 2 — The Host Substrate (`transport/rung2_host/`):** In-memory Go channels and local Unix IPC sockets (zero network cards required).
3. **Rung 3 — The Room & The Well (`transport/rung3_local/` + `interpreter/`):** Peer-to-peer LAN UDP, in-device Wi-Fi AP Beacon controller, and zero-install HTML Captive Portal so Windows, macOS, Linux, iOS, and Android devices can connect without custom apps.
4. **Rung 4 — Federated Global Backbone (`transport/rung4_global/` + `engine/covenant/`):** Optional TCP encapsulation and legal outbound egress strictly gated behind the **Genesis 3 Covenant Airlock**.

---

<!-- =========================================================================
BLOCK 3: BODY BLOCK / LAYERED DATA-ORIENTED ARCHITECTURE
Stratum: Root Orientation | Machine: docs.readme.faithnet
========================================================================== -->

## Layered Architecture (Downward Dependency Flow)

* **Layer 0 (`data/`):** Pure invariant data structures (`NodeID`, `Frame`, `Opcode`, `QuadState`, `HardwareProfile`). Zero OS or network imports.
* **Layer 1 (`engine/`):** Pure functional state machines (`Codec`, `Clock`, `PingEngine`, `Router`, `CovenantFilter`). Zero-allocation hot paths (`ALLOC_ZERO_HEAP`).
* **Layer 2 (`orchestrator/`):** Dynamic `BodyProbe`, `AdaptiveLadder`, and central `Dispatcher`.
* **Layer 3 (`transport/` & `interpreter/`):** Pluggable physical wires and universal legacy translators ("The Well").
* **Layer 4 (`cmd/` & `tests/`):** Thin imperative CLI entrypoints (`faithnet-ping`, `faithnet-node`) and 4-Tier verification suites.

---

<!-- =========================================================================
BLOCK 4: AUDIT PROCEDURES BLOCK / VERIFICATION SUITE
Stratum: Verification | Machine: docs.readme.faithnet
========================================================================== -->

## Verification Suite

```bash
# 1. Confirm zero third-party dependencies
go list -m all

# 2. Static analysis across all packages
go vet ./...

# 3. Execute all 4 verification tiers with race detector
go test -v -race ./tests/...

# 4. Verify zero-allocation hot paths (ALLOC_ZERO_HEAP audit)
go test -bench=. -benchmem ./tests/functional/... ./tests/unit/...

# 5. Run standalone offline loopback ping
go run ./cmd/faithnet-ping

# 6. Launch orchestrated sovereign node & optional captive portal ("The Well")
go run ./cmd/faithnet-node --help
```

---

<!-- =========================================================================
BLOCK 5: CLOSING BLOCK / COVENANT AFFIRMATIONS
Stratum: Governance | Machine: docs.readme.faithnet
========================================================================== -->

## Covenant Invariants

* **Constraint is Not Limitation:** Strict physical boundaries (512-byte floor frames, zero-alloc hot paths, zero third-party packages) yield eternal portability.
* **Compatibility Without Dependency:** FaithNet speaks to the modern world without ever being enslaved to it.
* **See Full Specification:** [`docs/ROADMAP.adoc`](docs/ROADMAP.adoc)

<!-- End of README.md -->
