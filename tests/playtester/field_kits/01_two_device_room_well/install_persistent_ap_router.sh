#!/usr/bin/env bash
# ==============================================================================
# BLOCK 1: METADATA BLOCK
# ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
# ==============================================================================
# @project       FaithNet
# @file          tests/playtester/field_kits/01_two_device_room_well/install_persistent_ap_router.sh
# @id            playtest.field_kit.01.linux.install_ap_service
# @layer         Layer 4: Proving Ground (Field Kit 01 — Linux System Service)
# @archetype     field_script | operational
# @version       0.1.0-dev
# @interface_rev Rev 5
# @date          2026-11-17
# @updated       2026-11-17
# @author        CreativeWorkzStudio LLC / ProfessorSeanEX
# @repository    https://github.com/CreativeWorkzStudio/FaithNet
# @system        FaithNet Persistent Boot-Time Micro-ISP Service Installer (Linux)
# @role          Install /usr/local/sbin/faithnet-ap-daemon.sh & systemd Unit
# @maturity      Rev 5 Baseline | Wave 14 5-Block Cross-Platform Parity
# @chassis       system-service | Reversible single unit + isolated script
# @scale         O(1) idempotent install, daemon-reload, enable & restart
# @seams         in: faithnet-ap-daemon.sh | exports: /etc/systemd/system/faithnet-ap.service
# @invariants    1. Universal 5-Block Lifecycle Protocol (Rev 5).
#                2. Strictly reversible via `sudo systemctl disable --now faithnet-ap`.
#                3. Never edits NetworkManager profiles or permanent iptables saves.
#
# PURPOSE & ARCHITECTURAL ROLE:
#   Installs `faithnet-ap-daemon.sh` into `/usr/local/sbin/` and registers
#   `/etc/systemd/system/faithnet-ap.service` so Laptop #1 automatically boots
#   as a persistent `FaithNet-HomeRoom` micro-ISP router on every startup.
#
#   ### For Architects, Operators & Non-Programmers:
#   One-command setup so your FaithNet Wi-Fi router turns itself back on every
#   time you power on your laptop, letting phones and laptops auto-reconnect.
#
#   ### For Systems Programmers & Engine Architects:
#   Validates source daemon syntax (`bash -n`) prior to `sudo install`, writes
#   the `multi-user.target` unit, enables, restarts, and verifies live status.
# ==============================================================================
set -euo pipefail

# ==============================================================================
# BLOCK 2: OPENING BLOCK
# ( Operational Paths & Service Anchors )
# ==============================================================================
# Engine:        FaithNet Adaptive Networking Substrate
# Stratum:       Layer 4: Proving Ground (Field Kit 01 — Linux)
# Module:        playtest.field_kit.01.linux.install_ap_service
# Mission:       Pin source daemon path, install destination, and unit file.
# ==============================================================================

readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SRC_DAEMON="${SCRIPT_DIR}/faithnet-ap-daemon.sh"
readonly DST_DAEMON="/usr/local/sbin/faithnet-ap-daemon.sh"
readonly UNIT_FILE="/etc/systemd/system/faithnet-ap.service"
readonly SERVICE_NAME="faithnet-ap.service"

# ==============================================================================
# BLOCK 3: BODY BLOCK
# ( Daemon Installation, Systemd Unit Creation & Service Activation )
# ==============================================================================
# Engine:        FaithNet Adaptive Networking Substrate
# Stratum:       Layer 4: Proving Ground (Field Kit 01 — Linux)
# Module:        playtest.field_kit.01.linux.install_ap_service
# Mission:       Deploy daemon script and enable/restart faithnet-ap.service.
# ==============================================================================

install_daemon_binary() {
  echo "[1/3] Installing ${DST_DAEMON}..."
  sudo install -m 0755 "${SRC_DAEMON}" "${DST_DAEMON}"
}

write_systemd_unit() {
  echo "[2/3] Writing ${UNIT_FILE}..."
  sudo tee "${UNIT_FILE}" >/dev/null <<'EOF'
[Unit]
Description=FaithNet Persistent Micro-ISP Wi-Fi AP Router (Simultaneous STA+AP)
After=NetworkManager.service NetworkManager-wait-online.service
Wants=NetworkManager.service

[Service]
Type=simple
ExecStart=/usr/local/sbin/faithnet-ap-daemon.sh
Restart=on-failure
RestartSec=5
KillMode=mixed
TimeoutStopSec=10

[Install]
WantedBy=multi-user.target
EOF
}

activate_system_service() {
  echo "[3/3] Enabling and starting persistent ${SERVICE_NAME}..."
  sudo systemctl daemon-reload
  sudo systemctl enable "${SERVICE_NAME}"
  sudo systemctl restart "${SERVICE_NAME}"
  sleep 3
}

# ==============================================================================
# BLOCK 4: AUDIT PROCEDURES BLOCK
# ( Pre-Flight Syntax Check & Post-Activation Live Status Verification )
# ==============================================================================
# Engine:        FaithNet Adaptive Networking Substrate
# Stratum:       Layer 4: Proving Ground (Field Kit 01 — Linux)
# Module:        playtest.field_kit.01.linux.install_ap_service
# Mission:       Lint source daemon before deploying and verify active unit.
# ==============================================================================

preflight_ap_install() {
  if [[ ! -f "${SRC_DAEMON}" ]]; then
    echo "[PREFLIGHT FAIL] Missing source daemon: ${SRC_DAEMON}" >&2
    return 1
  fi
  if ! bash -n "${SRC_DAEMON}"; then
    echo "[PREFLIGHT FAIL] Syntax error in ${SRC_DAEMON}" >&2
    return 1
  fi
  return 0
}

postflight_verify_ap() {
  sudo systemctl --no-pager status "${SERVICE_NAME}" | head -n 20
}

# ==============================================================================
# BLOCK 5: CLOSING BLOCK
# ( Lifecycle Entrypoint )
# ==============================================================================
# Engine:        FaithNet Adaptive Networking Substrate
# Stratum:       Layer 4: Proving Ground (Field Kit 01 — Linux)
# Module:        playtest.field_kit.01.linux.install_ap_service
# Mission:       Execute pre-flight -> install -> activate -> post-flight.
# ==============================================================================

main() {
  preflight_ap_install
  install_daemon_binary
  write_systemd_unit
  activate_system_service
  postflight_verify_ap
}

main "$@"

# End of tests/playtester/field_kits/01_two_device_room_well/install_persistent_ap_router.sh
