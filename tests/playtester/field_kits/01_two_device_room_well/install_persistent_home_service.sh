#!/usr/bin/env bash
# ==============================================================================
# BLOCK 1: METADATA BLOCK — FaithNet Persistent Home-Anchor Service Installer
# @project FaithNet
# @file    tests/playtester/field_kits/01_two_device_room_well/install_persistent_home_service.sh
# @role    Build faithnet-node and install/enable persistent systemd --user unit
# ==============================================================================
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../.." && pwd)"
BIN_DIR="${HOME}/.local/bin"
CA_DIR="${HOME}/.local/share/faithnet/ca"
UNIT_DIR="${HOME}/.config/systemd/user"
UNIT_FILE="${UNIT_DIR}/faithnet-home.service"

mkdir -p "${BIN_DIR}" "${CA_DIR}" "${UNIT_DIR}"

echo "[1/3] Compiling standalone native faithnet-node binary..."
(cd "${REPO_ROOT}" && go build -o "${BIN_DIR}/faithnet-node" ./cmd/faithnet-node)

echo "[2/3] Writing systemd --user unit (${UNIT_FILE})..."
cat > "${UNIT_FILE}" <<EOF
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

echo "[3/3] Reloading and starting persistent faithnet-home.service..."
systemctl --user daemon-reload
systemctl --user enable --now faithnet-home.service
systemctl --user --no-pager status faithnet-home.service | head -n 16
