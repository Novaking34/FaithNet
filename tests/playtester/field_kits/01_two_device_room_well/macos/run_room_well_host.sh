#!/usr/bin/env bash
# ==============================================================================
# BLOCK 1: METADATA BLOCK
# ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
# ==============================================================================
# @project       FaithNet
# @file          tests/playtester/field_kits/01_two_device_room_well/macos/run_room_well_host.sh
# @id            playtest.field_kit.01.macos.room_well_host
# @layer         Layer 4: Proving Ground (Field Kit 01 — macOS Room & Well)
# @archetype     field_script | operational
# @version       0.1.0-dev
# @interface_rev Rev 5
# @date          2026-11-17
# @updated       2026-11-17
# @author        CreativeWorkzStudio LLC / ProfessorSeanEX
# @repository    https://github.com/CreativeWorkzStudio/FaithNet
# @system        FaithNet Field Kit 01 — Two-Device Room & Well Launcher (macOS)
# @role          Foreground Room & Well Node Launcher (macOS Apple Silicon/Intel)
# @maturity      Rev 5 Baseline | Wave 14 5-Block Cross-Platform Parity
# @chassis       stateless-wrapper | Zero permanent macOS system modifications
# @scale         O(1) en0/en1 IP discovery + foreground faithnet-node launch
# @seams         in: cmd/faithnet-node | exports: live UDP :7777 & Well :8077
# @invariants    1. Universal 5-Block Lifecycle Protocol (Rev 5).
#                2. Uses macOS native `ipconfig getifaddr` (en0/en1) cleanly.
#                3. Zero external dependencies beyond Go toolchain.
#
# PURPOSE & ARCHITECTURAL ROLE:
#   macOS counterpart to Linux `run_room_well_host.sh`. Discovers the active
#   Wi-Fi/Ethernet IPv4 via `ipconfig getifaddr` and starts `faithnet-node`
#   with Tier 2 Sovereign Local Root CA and zero-config `/pair.sh`.
#
#   ### For Architects, Operators & Non-Programmers:
#   One-command script turning any MacBook or Mac into a live FaithNet local
#   room anchor so nearby phones and laptops can open The Well immediately.
#
#   ### For Systems Programmers & Engine Architects:
#   Pure POSIX/bash 3.2+ compatible (macOS default `/bin/bash` friendly).
# ==============================================================================
set -euo pipefail

# ==============================================================================
# BLOCK 2: OPENING BLOCK
# ( Operational Defaults & macOS LAN IPv4 Discovery )
# ==============================================================================
# Engine:        FaithNet Adaptive Networking Substrate
# Stratum:       Layer 4: Proving Ground (Field Kit 01 — macOS)
# Module:        playtest.field_kit.01.macos.room_well_host
# Mission:       Define macOS defaults and resolve active en0/en1 IPv4.
# ==============================================================================

NODE_LABEL="${NODE_LABEL:-Home-Anchor-Mac}"
UDP_ADDR="${UDP_ADDR:-0.0.0.0:7777}"
WELL_ADDR="${WELL_ADDR:-0.0.0.0:8077}"
BEACON_SSID="${BEACON_SSID:-FaithNet-HomeRoom}"
TLS_CA_DIR="${TLS_CA_DIR:-${TMPDIR:-/tmp}/faithnet-sovereign-ca}"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.." && pwd)"

detect_macos_ip() {
  local ip_val=""
  if command -v ipconfig >/dev/null 2>&1; then
    ip_val="$(ipconfig getifaddr en0 2>/dev/null || ipconfig getifaddr en1 2>/dev/null || true)"
  fi
  if [[ -z "${ip_val}" ]]; then
    ip_val="127.0.0.1"
  fi
  echo "${ip_val}"
}

# ==============================================================================
# BLOCK 3: BODY BLOCK
# ( Visitor Onboarding Banner & Node Launch )
# ==============================================================================
# Engine:        FaithNet Adaptive Networking Substrate
# Stratum:       Layer 4: Proving Ground (Field Kit 01 — macOS)
# Module:        playtest.field_kit.01.macos.room_well_host
# Mission:       Print onboarding URLs and run foreground faithnet-node.
# ==============================================================================

print_banner() {
  local host_ip="$1"
  echo "=== FAITHNET FIELD KIT 01: TWO-DEVICE ROOM & THE WELL (MACOS) ==="
  echo "Host LAN IP : ${host_ip}"
  echo "The Well URL: http://${host_ip}:8077/  (or https://${host_ip}:8077/)"
  echo "LAN UDP Port: ${host_ip}:7777"
  echo "Root CA Cert: http://${host_ip}:8077/faithnet-root-ca.crt"
  echo "Auto-Pair   : curl -fsSL http://${host_ip}:8077/pair.sh | sh"
  echo ""
}

launch_node() {
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
# ( Pre-Flight Toolchain & Anchor Check )
# ==============================================================================
# Engine:        FaithNet Adaptive Networking Substrate
# Stratum:       Layer 4: Proving Ground (Field Kit 01 — macOS)
# Module:        playtest.field_kit.01.macos.room_well_host
# Mission:       Verify `go` and FaithNet `go.mod` anchor before launch.
# ==============================================================================

preflight_macos_host() {
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
# ( Lifecycle Entrypoint )
# ==============================================================================
# Engine:        FaithNet Adaptive Networking Substrate
# Stratum:       Layer 4: Proving Ground (Field Kit 01 — macOS)
# Module:        playtest.field_kit.01.macos.room_well_host
# Mission:       Run pre-flight, display banner, and launch faithnet-node.
# ==============================================================================

main() {
  preflight_macos_host
  local host_ip
  host_ip="$(detect_macos_ip)"
  print_banner "${host_ip}"
  launch_node "$@"
}

main "$@"

# End of tests/playtester/field_kits/01_two_device_room_well/macos/run_room_well_host.sh
