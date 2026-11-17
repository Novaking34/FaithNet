#!/usr/bin/env bash
# ==============================================================================
# BLOCK 1: METADATA BLOCK — FaithNet Simultaneous STA+AP Wi-Fi Router Launcher
# @project FaithNet
# @file    tests/playtester/field_kits/01_two_device_room_well/run_wifi_ap_router.sh
# @role    Create real simultaneous AP (ap0: FaithNet-HomeRoom) while keeping
#          upstream Wi-Fi (wlp3s0) connected; NATs traffic + serves The Well.
# @safety  Zero-brick: never modifies wlp3s0; trap EXIT tears down ap0 cleanly.
# ==============================================================================
set -euo pipefail

STA_IF="${STA_IF:-wlp3s0}"
AP_IF="${AP_IF:-ap0}"
SSID="${SSID:-FaithNet-HomeRoom}"
PASS="${PASS:-FaithNet2026}"
AP_IP="${AP_IP:-10.77.77.1}"
AP_SUBNET="${AP_SUBNET:-10.77.77.0/24}"
TMP_DIR="$(mktemp -d /tmp/faithnet-ap-XXXXXX)"
DNSMASQ_PID=""

cleanup() {
  echo ""
  echo "[CLEANUP] Tearing down FaithNet AP (${AP_IF}) safely (leaving ${STA_IF} untouched)..."
  [[ -n "${DNSMASQ_PID}" ]] && sudo kill "${DNSMASQ_PID}" 2>/dev/null || true
  sudo iptables -t nat -D POSTROUTING -s "${AP_SUBNET}" -o "${STA_IF}" -j MASQUERADE 2>/dev/null || true
  sudo iptables -D FORWARD -i "${AP_IF}" -o "${STA_IF}" -j ACCEPT 2>/dev/null || true
  sudo iptables -D FORWARD -i "${STA_IF}" -o "${AP_IF}" -m state --state RELATED,ESTABLISHED -j ACCEPT 2>/dev/null || true
  sudo ip link set "${AP_IF}" down 2>/dev/null || true
  sudo iw dev "${AP_IF}" del 2>/dev/null || true
  rm -rf "${TMP_DIR}"
  echo "[CLEANUP] Done."
}
trap cleanup EXIT

# 1. Read active STA channel (RTL8852CE requires STA + AP on same channel).
CHAN="$(iw dev "${STA_IF}" info | awk '/channel / {print $2; exit}')"
if [[ -z "${CHAN}" ]]; then
  echo "ERROR: Could not detect active channel on ${STA_IF}." >&2
  exit 1
fi
HW_MODE="g"
if (( CHAN > 14 )); then
  HW_MODE="a"
fi

# 2. Derive distinct locally-administered MAC for virtual AP interface.
BASE_MAC="$(cat "/sys/class/net/${STA_IF}/address")"
FIRST_OCTET_HEX="${BASE_MAC%%:*}"
NEW_FIRST="$(printf '%02x' $(( (0x${FIRST_OCTET_HEX} | 0x02) ^ 0x04 )))"
AP_MAC="${NEW_FIRST}:${BASE_MAC#*:}"

echo "=== FAITHNET REAL WI-FI AP ROUTER (SIMULTANEOUS STA + AP) ==="
echo "Upstream STA : ${STA_IF} (Channel ${CHAN}, mode ${HW_MODE}) — UNTOUCHED"
echo "Virtual AP   : ${AP_IF} (${AP_MAC}) -> SSID '${SSID}'"
echo "Router IP    : ${AP_IP}/24 | The Well: http://${AP_IP}:8077/ | http://faithnet.local:8077/"
echo "WPA2 Pass    : ${PASS}"
echo ""

# 3. Create virtual AP interface and keep NetworkManager hands off ap0.
sudo iw dev "${AP_IF}" del 2>/dev/null || true
sudo iw dev "${STA_IF}" interface add "${AP_IF}" type __ap
sudo ip link set dev "${AP_IF}" address "${AP_MAC}"
sudo nmcli dev set "${AP_IF}" managed no 2>/dev/null || true
sudo ip addr flush dev "${AP_IF}"
sudo ip addr add "${AP_IP}/24" dev "${AP_IF}"
sudo ip link set "${AP_IF}" up

# 4. Restart persistent faithnet-home.service so Tier 2 CA signs 10.77.77.1 SAN!
if systemctl --user is-active --quiet faithnet-home.service 2>/dev/null; then
  echo "[SYNC] Restarting faithnet-home.service to include ${AP_IP} in Tier 2 TLS cert SANs..."
  systemctl --user restart faithnet-home.service
fi

# 5. Start isolated dnsmasq DHCP + local faithnet.local DNS bound ONLY to ap0.
cat > "${TMP_DIR}/dnsmasq.conf" <<EOF
interface=${AP_IF}
bind-interfaces
port=53
dhcp-range=10.77.77.10,10.77.77.100,255.255.255.0,12h
dhcp-option=option:router,${AP_IP}
dhcp-option=option:dns-server,${AP_IP}
address=/faithnet.local/${AP_IP}
server=1.1.1.1
server=8.8.8.8
no-resolv
log-dhcp
EOF
sudo dnsmasq -C "${TMP_DIR}/dnsmasq.conf" --pid-file="${TMP_DIR}/dnsmasq.pid"
DNSMASQ_PID="$(cat "${TMP_DIR}/dnsmasq.pid")"

# 6. Enable NAT router forwarding so connected clients stay online via STA_IF.
sudo sysctl -w net.ipv4.ip_forward=1 >/dev/null
sudo iptables -t nat -C POSTROUTING -s "${AP_SUBNET}" -o "${STA_IF}" -j MASQUERADE 2>/dev/null || \
  sudo iptables -t nat -A POSTROUTING -s "${AP_SUBNET}" -o "${STA_IF}" -j MASQUERADE
sudo iptables -C FORWARD -i "${AP_IF}" -o "${STA_IF}" -j ACCEPT 2>/dev/null || \
  sudo iptables -A FORWARD -i "${AP_IF}" -o "${STA_IF}" -j ACCEPT
sudo iptables -C FORWARD -i "${STA_IF}" -o "${AP_IF}" -m state --state RELATED,ESTABLISHED -j ACCEPT 2>/dev/null || \
  sudo iptables -A FORWARD -i "${STA_IF}" -o "${AP_IF}" -m state --state RELATED,ESTABLISHED -j ACCEPT

# 7. Run hostapd in foreground (Ctrl-C tears everything down cleanly).
cat > "${TMP_DIR}/hostapd.conf" <<EOF
interface=${AP_IF}
driver=nl80211
ssid=${SSID}
hw_mode=${HW_MODE}
channel=${CHAN}
ieee80211n=1
wmm_enabled=1
auth_algs=1
wpa=2
wpa_passphrase=${PASS}
wpa_key_mgmt=WPA-PSK
rsn_pairwise=CCMP
EOF

echo "[LIVE] Broadcasting Wi-Fi SSID '${SSID}'! Connect phone or Laptop #2 now (Ctrl-C to stop)."
sudo hostapd "${TMP_DIR}/hostapd.conf"
