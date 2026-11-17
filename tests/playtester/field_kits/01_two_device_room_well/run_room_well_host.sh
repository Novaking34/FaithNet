#!/usr/bin/env bash
# ==============================================================================
# BLOCK 1: METADATA BLOCK — FaithNet Field Kit 01 (Two-Device Room & The Well)
# @project FaithNet | @file tests/playtester/field_kits/01_two_device_room_well/run_room_well_host.sh
# @role    Launch FaithNet Node on LAN 0.0.0.0:7777 (UDP) + 0.0.0.0:8077 (Well)
# ==============================================================================
set -euo pipefail

HOST_IP="$(ip -4 -o addr show dev wlp3s0 2>/dev/null | awk '{print $4}' | cut -d/ -f1 | head -n1 || true)"
if [[ -z "${HOST_IP}" ]]; then
  HOST_IP="$(hostname -I 2>/dev/null | awk '{print $1}' || echo '127.0.0.1')"
fi

echo "=== FAITHNET FIELD KIT 01: TWO-DEVICE ROOM & THE WELL ==="
echo "Host LAN IP : ${HOST_IP}"
echo "The Well URL: http://${HOST_IP}:8077/"
echo "LAN UDP Port: ${HOST_IP}:7777"
echo ""
echo "From Laptop #2 or Phone on the same Wi-Fi, open:"
echo "  http://${HOST_IP}:8077/"
echo "Or test via curl:"
echo "  curl -s http://${HOST_IP}:8077/api/status"
echo "  curl -s -X POST -d 'Hello from Laptop 2' http://${HOST_IP}:8077/api/message"
echo "  curl -i -X POST -d '<script>alert(1)</script>' http://${HOST_IP}:8077/api/message"
echo ""
exec go run ./cmd/faithnet-node \
  --label="Home-Anchor" \
  --udp="0.0.0.0:7777" \
  --well \
  --well-addr="0.0.0.0:8077" \
  --beacon \
  --beacon-ssid="FaithNet-HomeRoom" \
  "$@"
