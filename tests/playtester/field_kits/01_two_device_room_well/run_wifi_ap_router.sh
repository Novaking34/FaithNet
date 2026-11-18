#!/usr/bin/env bash
# ==============================================================================
# BLOCK 1: METADATA BLOCK
# ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
# ==============================================================================
# @project       FaithNet
# @file          tests/playtester/field_kits/01_two_device_room_well/run_wifi_ap_router.sh
# @id            playtest.field_kit.01.linux.wifi_ap_router
# @layer         Layer 4: Proving Ground (Field Kit 01 — Interactive Linux AP)
# @archetype     field_script | operational
# @version       0.1.0-dev
# @interface_rev Rev 5
# @date          2026-11-17
# @updated       2026-11-17
# @author        CreativeWorkzStudio LLC / ProfessorSeanEX
# @repository    https://github.com/CreativeWorkzStudio/FaithNet
# @system        FaithNet Interactive Simultaneous STA+AP Wi-Fi Router (Linux)
# @role          Foreground Virtual AP Launcher With Auto-Channel & Clean Trap
# @maturity      Rev 5 Baseline | Wave 14 5-Block Cross-Platform Parity
# @chassis       ephemeral-ap | Zero permanent host network config mutations
# @scale         Single-radio simultaneous STA+AP (or Ch 6 offline island)
# @seams         in: wlp3s0, hostapd, dnsmasq, iptables | exports: ap0 10.77.77.1
# @invariants    1. Universal 5-Block Lifecycle Protocol (Rev 5).
#                2. Zero-brick safety: never modifies upstream wlp3s0 config.
#                3. `trap cleanup EXIT` guarantees full teardown on Ctrl-C.
#                4. Uses RFC 8375 `faithnet.home.arpa` + `faithnet.lan` unicast DNS.
#
# PURPOSE & ARCHITECTURAL ROLE:
#   Foreground interactive launcher for testing simultaneous STA+AP Wi-Fi
#   routing on Linux (`ap0` broadcasting `FaithNet-HomeRoom` at `10.77.77.1/24`
#   while upstream `wlp3s0` stays connected—or Channel 6 offline island mode).
#   For persistent reboot-surviving Micro-ISP service, see
#   `install_persistent_ap_router.sh` and `faithnet-ap-daemon.sh`.
#
#   ### For Architects, Operators & Non-Programmers:
#   Lets an operator test turning a Linux laptop into a live FaithNet Wi-Fi
#   router directly in a terminal window with automatic cleanup on Ctrl-C.
#
#   ### For Systems Programmers & Engine Architects:
#   Follows the empirically verified Realtek `rtw89` bring-up sequence: create
#   `__ap` interface DOWN -> `nmcli managed no` -> `hostapd` nl80211 enable ->
#   assign `10.77.77.1/24` -> isolated `dnsmasq` -> NAT forwarding.
# ==============================================================================
set -euo pipefail

# ==============================================================================
# BLOCK 2: OPENING BLOCK
# ( Operational Constants, Mutable PID State & Signal Cleanup Guard )
# ==============================================================================
# Engine:        FaithNet Adaptive Networking Substrate
# Stratum:       Layer 4: Proving Ground (Field Kit 01 — Linux)
# Module:        playtest.field_kit.01.linux.wifi_ap_router
# Mission:       Declare interface bounds, SSID/subnet config, and teardown.
# ==============================================================================

readonly STA_IF="${STA_IF:-wlp3s0}"
readonly AP_IF="${AP_IF:-ap0}"
readonly SSID="${SSID:-FaithNet-HomeRoom}"
readonly PASS="${PASS:-FaithNet2026}"
readonly AP_IP="${AP_IP:-10.77.77.1}"
readonly AP_SUBNET="${AP_SUBNET:-10.77.77.0/24}"
readonly TMP_DIR="$(mktemp -d /tmp/faithnet-ap-XXXXXX)"

DNSMASQ_PID=""
HOSTAPD_PID=""
CHAN=""
UPSTREAM_MODE="CONNECTED"

cleanup() {
  echo ""
  echo "[CLEANUP] Tearing down FaithNet AP (${AP_IF}) safely (leaving ${STA_IF} untouched)..."
  [[ -n "${HOSTAPD_PID}" ]] && sudo kill "${HOSTAPD_PID}" 2>/dev/null || true
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

# ==============================================================================
# BLOCK 3: BODY BLOCK
# ( Radio Channel Sync, hostapd Bring-Up, Isolated dnsmasq & NAT Forwarding )
# ==============================================================================
# Engine:        FaithNet Adaptive Networking Substrate
# Stratum:       Layer 4: Proving Ground (Field Kit 01 — Linux)
# Module:        playtest.field_kit.01.linux.wifi_ap_router
# Mission:       Project FaithNet-HomeRoom on matching live channel + NAT.
# ==============================================================================

detect_live_channel() {
  local ch
  ch="$(iw dev "${STA_IF}" info 2>/dev/null | awk '/channel / {print $2; exit}')"
  if [[ -z "${ch}" ]]; then
    UPSTREAM_MODE="OFFLINE_ISLAND (no upstream Wi-Fi)"
    echo "6"
  else
    UPSTREAM_MODE="CONNECTED"
    echo "${ch}"
  fi
}

derive_ap_mac() {
  local base_mac first new_first
  base_mac="$(cat "/sys/class/net/${STA_IF}/address")"
  first="${base_mac%%:*}"
  new_first="$(printf '%02x' $(( (0x${first} | 0x02) ^ 0x04 )))"
  echo "${new_first}:${base_mac#*:}"
}

start_hostapd_on_channel() {
  local ch="$1"
  local mode="g"
  if (( ch > 14 )); then
    mode="a"
  fi
  cat > "${TMP_DIR}/hostapd.conf" <<EOF
interface=${AP_IF}
driver=nl80211
ssid=${SSID}
hw_mode=${mode}
channel=${ch}
ieee80211n=1
wmm_enabled=1
auth_algs=1
wpa=2
wpa_passphrase=${PASS}
wpa_key_mgmt=WPA-PSK
rsn_pairwise=CCMP
EOF
  sudo hostapd -B -P "${TMP_DIR}/hostapd.pid" "${TMP_DIR}/hostapd.conf"
  sleep 1
  HOSTAPD_PID="$(cat "${TMP_DIR}/hostapd.pid" 2>/dev/null || true)"
  sudo ip addr flush dev "${AP_IF}"
  sudo ip addr add "${AP_IP}/24" dev "${AP_IF}"
}

start_isolated_dnsmasq_and_nat() {
  cat > "${TMP_DIR}/dnsmasq.conf" <<EOF
interface=${AP_IF}
bind-interfaces
port=53
dhcp-range=10.77.77.10,10.77.77.100,255.255.255.0,12h
dhcp-option=option:router,${AP_IP}
dhcp-option=option:dns-server,${AP_IP}
address=/faithnet.lan/${AP_IP}
address=/faithnet.home.arpa/${AP_IP}
address=/faithnet.local/${AP_IP}
server=1.1.1.1
server=8.8.8.8
no-resolv
log-dhcp
EOF
  sudo dnsmasq -C "${TMP_DIR}/dnsmasq.conf" --pid-file="${TMP_DIR}/dnsmasq.pid"
  DNSMASQ_PID="$(cat "${TMP_DIR}/dnsmasq.pid")"

  sudo sysctl -w net.ipv4.ip_forward=1 >/dev/null
  sudo iptables -t nat -C POSTROUTING -s "${AP_SUBNET}" -o "${STA_IF}" -j MASQUERADE 2>/dev/null || \
    sudo iptables -t nat -A POSTROUTING -s "${AP_SUBNET}" -o "${STA_IF}" -j MASQUERADE
  sudo iptables -C FORWARD -i "${AP_IF}" -o "${STA_IF}" -j ACCEPT 2>/dev/null || \
    sudo iptables -A FORWARD -i "${AP_IF}" -o "${STA_IF}" -j ACCEPT
  sudo iptables -C FORWARD -i "${STA_IF}" -o "${AP_IF}" -m state --state RELATED,ESTABLISHED -j ACCEPT 2>/dev/null || \
    sudo iptables -A FORWARD -i "${STA_IF}" -o "${AP_IF}" -m state --state RELATED,ESTABLISHED -j ACCEPT
}

# ==============================================================================
# BLOCK 4: AUDIT PROCEDURES BLOCK
# ( Pre-Flight Dependency & Radio Interface Verification )
# ==============================================================================
# Engine:        FaithNet Adaptive Networking Substrate
# Stratum:       Layer 4: Proving Ground (Field Kit 01 — Linux)
# Module:        playtest.field_kit.01.linux.wifi_ap_router
# Mission:       Confirm iw, hostapd, dnsmasq, iptables, and STA_IF exist.
# ==============================================================================

preflight_ap_router() {
  for tool in iw ip hostapd dnsmasq iptables nmcli; do
    if ! command -v "${tool}" >/dev/null 2>&1; then
      echo "[PREFLIGHT FAIL] Missing required wireless tool: ${tool}" >&2
      return 1
    fi
  done
  if ! ip link show "${STA_IF}" >/dev/null 2>&1; then
    echo "[PREFLIGHT FAIL] Upstream Wi-Fi interface ${STA_IF} not found." >&2
    return 1
  fi
  return 0
}

# ==============================================================================
# BLOCK 5: CLOSING BLOCK
# ( Lifecycle Entrypoint & Auto-Channel Monitor Loop )
# ==============================================================================
# Engine:        FaithNet Adaptive Networking Substrate
# Stratum:       Layer 4: Proving Ground (Field Kit 01 — Linux)
# Module:        playtest.field_kit.01.linux.wifi_ap_router
# Mission:       Bring up FaithNet-HomeRoom and monitor channel shifts cleanly.
# ==============================================================================

main() {
  preflight_ap_router
  CHAN="$(detect_live_channel)"
  local ap_mac
  ap_mac="$(derive_ap_mac)"

  echo "=== FAITHNET REAL WI-FI AP ROUTER (SIMULTANEOUS STA + AP) ==="
  echo "Upstream STA : ${STA_IF} (${UPSTREAM_MODE} | Channel ${CHAN}) — UNTOUCHED"
  echo "Virtual AP   : ${AP_IF} (${ap_mac}) -> SSID '${SSID}'"
  echo "Router IP    : ${AP_IP}/24 | The Well: https://faithnet.lan:8077/ | https://${AP_IP}:8077/"
  echo "WPA2 Pass    : ${PASS}"
  echo ""

  sudo pkill -f "/tmp/faithnet-ap-" 2>/dev/null || true
  sleep 0.3
  sudo iw dev "${AP_IF}" del 2>/dev/null || true
  sudo iw dev "${STA_IF}" interface add "${AP_IF}" type __ap
  sudo nmcli dev set "${AP_IF}" managed no 2>/dev/null || true
  sudo ip link set dev "${AP_IF}" address "${ap_mac}"

  start_hostapd_on_channel "${CHAN}"
  if systemctl --user is-active --quiet faithnet-home.service 2>/dev/null; then
    echo "[SYNC] Restarting faithnet-home.service to confirm ${AP_IP} in Tier 2 SANs..."
    systemctl --user restart faithnet-home.service
  fi
  start_isolated_dnsmasq_and_nat

  echo "[LIVE] Wi-Fi SSID '${SSID}' is LIVE on ${AP_IF} (${AP_IP}/24, Channel ${CHAN})!"
  echo "[LIVE] Auto-channel follower active (Ctrl-C tears down cleanly)..."
  while sudo kill -0 "${HOSTAPD_PID}" 2>/dev/null; do
    sleep 5
    local new_ch
    new_ch="$(detect_live_channel)"
    if [[ "${new_ch}" != "${CHAN}" ]]; then
      echo "[AUTO-CHANNEL] Upstream shifted Channel ${CHAN} -> ${new_ch}; retuning ${AP_IF}..."
      sudo kill "${HOSTAPD_PID}" 2>/dev/null || true
      sleep 0.5
      CHAN="${new_ch}"
      start_hostapd_on_channel "${CHAN}"
      echo "[AUTO-CHANNEL] '${SSID}' live on Channel ${CHAN} (${AP_IP}/24)!"
    fi
  done
}

main "$@"

# End of tests/playtester/field_kits/01_two_device_room_well/run_wifi_ap_router.sh
