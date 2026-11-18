#!/usr/bin/env bash
# ==============================================================================
# BLOCK 1: METADATA BLOCK
# ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
# ==============================================================================
# @project       FaithNet
# @file          tests/playtester/field_kits/01_two_device_room_well/install_persistent_home_service.sh
# @id            playtest.field_kit.01.linux.install_home_service
# @layer         Layer 4: Proving Ground (Field Kit 01 — Linux User Service)
# @archetype     field_script | operational
# @version       0.1.0-dev
# @interface_rev Rev 5
# @date          2026-11-17
# @updated       2026-11-17
# @author        CreativeWorkzStudio LLC / ProfessorSeanEX
# @repository    https://github.com/CreativeWorkzStudio/FaithNet
# @system        FaithNet Persistent Home-Anchor User Service Installer (Linux)
# @role          Compile Native faithnet-node & Enable systemd --user Unit
# @maturity      Rev 5 Baseline | Wave 14 5-Block Cross-Platform Parity
# @chassis       user-scoped | Zero root required; clean systemd --user unit
# @scale         O(1) deterministic build + idempotent service enable & restart
# @seams         in: cmd/faithnet-node | exports: ~/.config/systemd/user/faithnet-home.service
# @invariants    1. Universal 5-Block Lifecycle Protocol (Rev 5).
#                2. Zero sudo required (installs strictly into $HOME/.local).
#                3. Preserves persistent Tier 2 Sovereign Root CA across updates.
#
# PURPOSE & ARCHITECTURAL ROLE:
#   Compiles the standalone zero-dependency native `faithnet-node` executable
#   and installs/restarts `faithnet-home.service` under `systemd --user` so the
#   Home-Anchor node, The Well, Tier 2 Sovereign Root CA, and `/pair.sh` stay
#   online automatically across logins and reboots.
#
#   ### For Architects, Operators & Non-Programmers:
#   Turns your Linux computer into a persistent always-on FaithNet Home-Anchor
#   with zero hanging terminal windows.
#
#   ### For Systems Programmers & Engine Architects:
#   Idempotent build + unit write + `daemon-reload` + `enable` + `restart`.
# ==============================================================================
set -euo pipefail

# ==============================================================================
# BLOCK 2: OPENING BLOCK
# ( Operational Paths, Service Constants & Anchors )
# ==============================================================================
# Engine:        FaithNet Adaptive Networking Substrate
# Stratum:       Layer 4: Proving Ground (Field Kit 01 — Linux)
# Module:        playtest.field_kit.01.linux.install_home_service
# Mission:       Define user-local binary, CA storage, and systemd unit paths.
# ==============================================================================

readonly REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../.." && pwd)"
readonly BIN_DIR="${HOME}/.local/bin"
readonly CA_DIR="${HOME}/.local/share/faithnet/ca"
readonly UNIT_DIR="${HOME}/.config/systemd/user"
readonly UNIT_FILE="${UNIT_DIR}/faithnet-home.service"
readonly SERVICE_NAME="faithnet-home.service"

# ==============================================================================
# BLOCK 3: BODY BLOCK
# ( Native Build, systemd --user Unit Generation & Idempotent Activation )
# ==============================================================================
# Engine:        FaithNet Adaptive Networking Substrate
# Stratum:       Layer 4: Proving Ground (Field Kit 01 — Linux)
# Module:        playtest.field_kit.01.linux.install_home_service
# Mission:       Build faithnet-node, write unit file, and restart cleanly.
# ==============================================================================

build_native_binary() {
  mkdir -p "${BIN_DIR}" "${CA_DIR}" "${UNIT_DIR}"
  echo "[1/3] Compiling standalone native faithnet-node binary..."
  (cd "${REPO_ROOT}" && go build -o "${BIN_DIR}/faithnet-node" ./cmd/faithnet-node)
}

write_systemd_user_unit() {
  echo "[2/3] Writing systemd --user unit (${UNIT_FILE})..."
  cat > "${UNIT_FILE}" <<'EOF'
[Unit]
Description=FaithNet Sovereign Home-Anchor Node (Rung 3 LAN + The Well Tier 2 CA)
After=network-online.target

[Service]
Type=simple
ExecStart=%h/.local/bin/faithnet-node --label=Home-Anchor --udp=0.0.0.0:7777 --well --well-addr=0.0.0.0:8077 --tls-ca-dir=%h/.local/share/faithnet/ca --beacon --beacon-ssid=FaithNet-HomeRoom
Restart=on-failure
RestartSec=3

[Install]
WantedBy=default.target
EOF
}

activate_user_service() {
  echo "[3/3] Reloading and restarting persistent ${SERVICE_NAME}..."
  systemctl --user daemon-reload
  systemctl --user enable "${SERVICE_NAME}"
  systemctl --user restart "${SERVICE_NAME}"
}

# ==============================================================================
# BLOCK 4: AUDIT PROCEDURES BLOCK
# ( Pre-Flight Toolchain Check & Post-Activation Live Health Audit )
# ==============================================================================
# Engine:        FaithNet Adaptive Networking Substrate
# Stratum:       Layer 4: Proving Ground (Field Kit 01 — Linux)
# Module:        playtest.field_kit.01.linux.install_home_service
# Mission:       Verify prerequisites before build and confirm active status.
# ==============================================================================

preflight_install() {
  for tool in go systemctl; do
    if ! command -v "${tool}" >/dev/null 2>&1; then
      echo "[PREFLIGHT FAIL] Missing required tool: ${tool}" >&2
      return 1
    fi
  done
  if [[ ! -f "${REPO_ROOT}/go.mod" ]]; then
    echo "[PREFLIGHT FAIL] FaithNet go.mod anchor not found at ${REPO_ROOT}" >&2
    return 1
  fi
  return 0
}

postflight_verify() {
  systemctl --user --no-pager status "${SERVICE_NAME}" | head -n 16
}

# ==============================================================================
# BLOCK 5: CLOSING BLOCK
# ( Lifecycle Orchestration Entrypoint )
# ==============================================================================
# Engine:        FaithNet Adaptive Networking Substrate
# Stratum:       Layer 4: Proving Ground (Field Kit 01 — Linux)
# Module:        playtest.field_kit.01.linux.install_home_service
# Mission:       Execute pre-flight -> build -> install -> post-flight cleanly.
# ==============================================================================

main() {
  preflight_install
  build_native_binary
  write_systemd_user_unit
  activate_user_service
  postflight_verify
}

main "$@"

# End of tests/playtester/field_kits/01_two_device_room_well/install_persistent_home_service.sh
