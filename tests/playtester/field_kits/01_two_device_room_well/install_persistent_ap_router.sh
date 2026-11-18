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
# @version       0.2.0-dev
# @interface_rev Rev 5
# @date          2026-11-17
# @updated       2026-11-17
# @author        CreativeWorkzStudio LLC / ProfessorSeanEX
# @repository    https://github.com/CreativeWorkzStudio/FaithNet
# @system        FaithNet Persistent Sovereign Micro-ISP & Netctl Installer (Linux)
# @role          Install faithnet-ap-daemon.sh, faithnet-netctl & systemd Unit
# @maturity      Rev 5 Baseline | Wave 15 Sovereign Host Profile
# @chassis       system-service | Reversible single unit + isolated CLI
# @scale         O(1) idempotent install, daemon-reload, enable & restart
# @seams         in: faithnet-ap-daemon.sh, faithnet-netctl.sh | exports: systemd unit
# @invariants    1. Universal 5-Block Lifecycle Protocol (Rev 5).
#                2. Strictly reversible via `sudo systemctl disable --now faithnet-ap`.
#                3. Never edits NetworkManager profiles or permanent iptables saves.
#
# PURPOSE & ARCHITECTURAL ROLE:
#   Installs `faithnet-ap-daemon.sh` into `/usr/local/sbin/`, installs
#   `faithnet-netctl` into `/usr/local/bin/`, and registers
#   `/etc/systemd/system/faithnet-ap.service` so Laptop #1 automatically boots
#   into the Wave 15 Sovereign Host Profile (Split-DNS + Firewalled WAN +
#   Dedicated USB Adapter Auto-Detect).
# ==============================================================================
set -euo pipefail

# ==============================================================================
# BLOCK 2: OPENING BLOCK
# ( Operational Paths & Service Anchors )
# ==============================================================================

readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SRC_DAEMON="${SCRIPT_DIR}/faithnet-ap-daemon.sh"
readonly DST_DAEMON="/usr/local/sbin/faithnet-ap-daemon.sh"
readonly SRC_NETCTL="${SCRIPT_DIR}/faithnet-netctl.sh"
readonly DST_NETCTL="/usr/local/bin/faithnet-netctl"
readonly UNIT_FILE="/etc/systemd/system/faithnet-ap.service"
readonly SERVICE_NAME="faithnet-ap.service"

# ==============================================================================
# BLOCK 3: BODY BLOCK
# ( Daemon & CLI Installation, Systemd Unit Creation & Service Activation )
# ==============================================================================

install_binaries() {
  echo "[1/3] Installing ${DST_DAEMON} and ${DST_NETCTL}..."
  sudo install -m 0755 "${SRC_DAEMON}" "${DST_DAEMON}"
  sudo install -m 0755 "${SRC_NETCTL}" "${DST_NETCTL}"
}

write_systemd_unit() {
  echo "[2/3] Writing ${UNIT_FILE}..."
  sudo tee "${UNIT_FILE}" >/dev/null <<'EOF'
[Unit]
Description=FaithNet Sovereign Micro-ISP Wi-Fi Router & Firewalled WAN Guard
After=NetworkManager.service NetworkManager-wait-online.service systemd-resolved.service
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

preflight_ap_install() {
  for src in "${SRC_DAEMON}" "${SRC_NETCTL}"; do
    if [[ ! -f "${src}" ]]; then
      echo "[PREFLIGHT FAIL] Missing source script: ${src}" >&2
      return 1
    fi
    if ! bash -n "${src}"; then
      echo "[PREFLIGHT FAIL] Syntax error in ${src}" >&2
      return 1
    fi
  done
  return 0
}

postflight_verify_ap() {
  sudo systemctl --no-pager status "${SERVICE_NAME}" | head -n 22
}

# ==============================================================================
# BLOCK 5: CLOSING BLOCK
# ( Lifecycle Entrypoint )
# ==============================================================================

main() {
  preflight_ap_install
  install_binaries
  write_systemd_unit
  activate_system_service
  postflight_verify_ap
}

main "$@"

# End of tests/playtester/field_kits/01_two_device_room_well/install_persistent_ap_router.sh
