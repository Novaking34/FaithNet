#!/usr/bin/env bash
# ==============================================================================
# BLOCK 1: METADATA BLOCK — FaithNet Persistent Boot-Time AP Router Installer
# @project FaithNet
# @file    tests/playtester/field_kits/01_two_device_room_well/install_persistent_ap_router.sh
# @role    Install & enable persistent faithnet-ap.service so the laptop boots
#          automatically as a FaithNet-HomeRoom micro-ISP router every startup.
# ==============================================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "[1/3] Installing /usr/local/sbin/faithnet-ap-daemon.sh..."
sudo install -m 0755 "${SCRIPT_DIR}/faithnet-ap-daemon.sh" /usr/local/sbin/faithnet-ap-daemon.sh

echo "[2/3] Writing /etc/systemd/system/faithnet-ap.service..."
sudo tee /etc/systemd/system/faithnet-ap.service >/dev/null <<'EOF'
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

echo "[3/3] Enabling and starting persistent faithnet-ap.service..."
sudo systemctl daemon-reload
sudo systemctl enable faithnet-ap.service
sudo systemctl restart faithnet-ap.service
sleep 3
sudo systemctl --no-pager status faithnet-ap.service | head -n 20
