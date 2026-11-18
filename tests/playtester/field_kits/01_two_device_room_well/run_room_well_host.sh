#!/usr/bin/env bash
# ==============================================================================
# BLOCK 1: METADATA BLOCK
# ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
# ==============================================================================
# @project       FaithNet
# @file          tests/playtester/field_kits/01_two_device_room_well/run_room_well_host.sh
# @id            playtest.field_kit.01.linux.room_well_host
# @layer         Layer 4: Proving Ground (Field Kit 01 — Linux Room & Well Host)
# @archetype     field_script | operational
# @version       0.1.0-dev
# @interface_rev Rev 5
# @date          2026-11-17
# @updated       2026-11-17
# @author        CreativeWorkzStudio LLC / ProfessorSeanEX
# @repository    https://github.com/CreativeWorkzStudio/FaithNet
# @system        FaithNet Field Kit 01 — Two-Device LAN Room & The Well Launcher
# @role          Foreground Room & Well Node Launcher (Linux / Tier 2 CA)
# @maturity      Rev 5 Baseline | Wave 14 5-Block Cross-Platform Parity
# @chassis       stateless-wrapper | Zero permanent host mutations
# @scale         O(1) host LAN IP detection + foreground faithnet-node launch
# @seams         in: cmd/faithnet-node | exports: live UDP :7777 & Well :8077
# @invariants    1. Universal 5-Block Lifecycle Protocol (Rev 5).
#                2. Zero permanent edits to host network or firewall config.
#                3. Enables Tier 2 Sovereign Local Root CA for offline padlock.
#
# PURPOSE & ARCHITECTURAL ROLE:
#   Foreground Linux launcher for Field Kit 01 that discovers the active LAN
#   IPv4 address, prints visitor onboarding URLs and curl test commands, and
#   runs `faithnet-node` with Rung 3 LAN UDP, Room Beacon, and The Well portal.
#
#   ### For Architects, Operators & Non-Programmers:
#   One-command script to turn a Linux laptop into a live local FaithNet room
#   anchor so phones and nearby laptops can connect and test The Well.
#
#   ### For Systems Programmers & Engine Architects:
#   Executes `go run ./cmd/faithnet-node` (forwarding extra CLI flags `"$@"`)
#   with `--tls-ca-dir` persisted under `/tmp/faithnet-sovereign-ca`.
# ==============================================================================
set -euo pipefail

# ==============================================================================
# BLOCK 2: OPENING BLOCK
# ( Operational Defaults, Network Constants & Helper Predicates )
# ==============================================================================
# Engine:        FaithNet Adaptive Networking Substrate
# Stratum:       Layer 4: Proving Ground (Field Kit 01 — Linux)
# Module:        playtest.field_kit.01.linux.room_well_host
# Mission:       Pin default ports, SSID, CA directory, and LAN IP resolver.
# ==============================================================================

# ------------------------------------------------------------------------------
# Sub-Assembly 2.1: Operational Constants
# ------------------------------------------------------------------------------
readonly NODE_LABEL="${NODE_LABEL:-Home-Anchor}"
readonly UDP_ADDR="${UDP_ADDR:-0.0.0.0:7777}"
readonly WELL_ADDR="${WELL_ADDR:-0.0.0.0:8077}"
readonly BEACON_SSID="${BEACON_SSID:-FaithNet-HomeRoom}"
readonly TLS_CA_DIR="${TLS_CA_DIR:-/tmp/faithnet-sovereign-ca}"
readonly REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../.." && pwd)"

# ------------------------------------------------------------------------------
# Sub-Assembly 2.2: LAN IPv4 Discovery Helper
# ------------------------------------------------------------------------------
detect_host_ip() {
  local ip_val=""
  ip_val="$(ip -4 -o addr show dev wlp3s0 2>/dev/null | awk '{print $4}' | cut -d/ -f1 | head -n1 || true)"
  if [[ -z "${ip_val}" ]]; then
    ip_val="$(hostname -I 2>/dev/null | awk '{print $1}' || echo '127.0.0.1')"
  fi
  echo "${ip_val}"
}

# ==============================================================================
# BLOCK 3: BODY BLOCK
# ( Banner Telemetry & Sovereign Node Execution )
# ==============================================================================
# Engine:        FaithNet Adaptive Networking Substrate
# Stratum:       Layer 4: Proving Ground (Field Kit 01 — Linux)
# Module:        playtest.field_kit.01.linux.room_well_host
# Mission:       Print visitor endpoints and launch foreground faithnet-node.
# ==============================================================================

print_visitor_banner() {
  local host_ip="$1"
  echo "=== FAITHNET FIELD KIT 01: TWO-DEVICE ROOM & THE WELL (LINUX) ==="
  echo "Host LAN IP : ${host_ip}"
  echo "The Well URL: http://${host_ip}:8077/  (or https://${host_ip}:8077/)"
  echo "LAN UDP Port: ${host_ip}:7777"
  echo "Root CA Cert: http://${host_ip}:8077/faithnet-root-ca.crt"
  echo "Auto-Pair   : curl -fsSL http://${host_ip}:8077/pair.sh | sh"
  echo ""
  echo "Quick curl checks from Laptop #2:"
  echo "  curl -s http://${host_ip}:8077/api/status"
  echo "  curl -s -X POST -d 'Hello from Laptop 2' http://${host_ip}:8077/api/message"
  echo "  curl -i -X POST -d '<script>alert(1)</script>' http://${host_ip}:8077/api/message"
  echo ""
}

launch_room_well_node() {
  cd "${REPO_ROOT}"
  exec go run ./cmd/faithnet-node \
    --label="${NODE_LABEL}" \
    --udp="${UDP_ADDR}" \
    --well \
    --well-addr="${WELL_ADDR}" \
    --tls-ca-dir="${TLS_CA_DIR}" \
    --beacon \
    --beacon-ssid="${BEACON_SSID}" \
    "$@"
}

# ==============================================================================
# BLOCK 4: AUDIT PROCEDURES BLOCK
# ( Pre-Flight Toolchain & Repository Anchor Verification )
# ==============================================================================
# Engine:        FaithNet Adaptive Networking Substrate
# Stratum:       Layer 4: Proving Ground (Field Kit 01 — Linux)
# Module:        playtest.field_kit.01.linux.room_well_host
# Mission:       Verify `go` toolchain and FaithNet `go.mod` anchor before boot.
# ==============================================================================

preflight_room_well_host() {
  if ! command -v go >/dev/null 2>&1; then
    echo "[PREFLIGHT FAIL] Go toolchain not found in PATH." >&2
    return 1
  fi
  if [[ ! -f "${REPO_ROOT}/go.mod" ]]; then
    echo "[PREFLIGHT FAIL] FaithNet go.mod anchor missing at ${REPO_ROOT}." >&2
    return 1
  fi
  return 0
}

# ==============================================================================
# BLOCK 5: CLOSING BLOCK
# ( Lifecycle Entrypoint & Clean Exec Handoff )
# ==============================================================================
# Engine:        FaithNet Adaptive Networking Substrate
# Stratum:       Layer 4: Proving Ground (Field Kit 01 — Linux)
# Module:        playtest.field_kit.01.linux.room_well_host
# Mission:       Run pre-flight audit, print banner, and hand off to node.
# ==============================================================================

main() {
  preflight_room_well_host
  local host_ip
  host_ip="$(detect_host_ip)"
  print_visitor_banner "${host_ip}"
  launch_room_well_node "$@"
}

main "$@"

# End of tests/playtester/field_kits/01_two_device_room_well/run_room_well_host.sh
