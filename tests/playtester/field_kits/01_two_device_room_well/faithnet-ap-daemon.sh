#!/usr/bin/env bash
# ==============================================================================
# BLOCK 1: METADATA BLOCK — FaithNet Persistent Event-Driven Micro-ISP AP Daemon
# @project FaithNet
# @file    tests/playtester/field_kits/01_two_device_room_well/faithnet-ap-daemon.sh
# @role    Persistent boot-time systemd daemon managing simultaneous STA+AP
#          (FaithNet-HomeRoom on ap0) with immediate radio release whenever
#          wlp3s0 switches Wi-Fi/hotspot so Realtek rtw89 NEVER locks up!
# @safety  Zero-brick: never modifies wlp3s0; releases ap0 immediately on
#          connecting/disconnected events; clean trap EXIT teardown.
# ==============================================================================
set -euo pipefail

STA_IF="${STA_IF:-wlp3s0}"
AP_IF="${AP_IF:-ap0}"
SSID="${SSID:-FaithNet-HomeRoom}"
PASS="${PASS:-FaithNet2026}"
AP_IP="${AP_IP:-10.77.77.1}"
AP_SUBNET="${AP_SUBNET:-10.77.77.0/24}"
RUN_DIR="/run/faithnet-ap"
HOSTAPD_PID=""
DNSMASQ_PID=""
CURRENT_CHAN=""
AP_ACTIVE=0

mkdir -p "${RUN_DIR}"

log() {
  echo "[faithnet-ap $(date '+%H:%M:%S')] $*"
}

stop_ap() {
  if [[ "${AP_ACTIVE}" -eq 1 ]] || ip link show "${AP_IF}" >/dev/null 2>&1; then
    log "Releasing ${AP_IF} cleanly (leaving ${STA_IF} free & untouched)..."
  fi
  if [[ -n "${HOSTAPD_PID}" ]]; then
    kill "${HOSTAPD_PID}" 2>/dev/null || true
    HOSTAPD_PID=""
  fi
  if [[ -n "${DNSMASQ_PID}" ]]; then
    kill "${DNSMASQ_PID}" 2>/dev/null || true
    DNSMASQ_PID=""
  fi
  pkill -f "${RUN_DIR}/hostapd.conf" 2>/dev/null || true
  pkill -f "${RUN_DIR}/dnsmasq.conf" 2>/dev/null || true
  iptables -t nat -D POSTROUTING -s "${AP_SUBNET}" -o "${STA_IF}" -j MASQUERADE 2>/dev/null || true
  iptables -D FORWARD -i "${AP_IF}" -o "${STA_IF}" -j ACCEPT 2>/dev/null || true
  iptables -D FORWARD -i "${STA_IF}" -o "${AP_IF}" -m state --state RELATED,ESTABLISHED -j ACCEPT 2>/dev/null || true
  ip link set "${AP_IF}" down 2>/dev/null || true
  iw dev "${AP_IF}" del 2>/dev/null || true
  AP_ACTIVE=0
  CURRENT_CHAN=""
}

cleanup() {
  stop_ap
  rm -rf "${RUN_DIR}"
}
trap cleanup EXIT INT TERM

derive_ap_mac() {
  local base_mac first new_first
  base_mac="$(cat "/sys/class/net/${STA_IF}/address" 2>/dev/null || echo "44:fa:66:89:f6:f5")"
  first="${base_mac%%:*}"
  new_first="$(printf '%02x' $(( (0x${first} | 0x02) ^ 0x04 )))"
  echo "${new_first}:${base_mac#*:}"
}

start_ap() {
  local chan="$1"
  local mode="g"
  if (( chan > 14 )); then
    mode="a"
  fi

  stop_ap
  sleep 0.3

  local ap_mac
  ap_mac="$(derive_ap_mac)"

  log "Bringing up '${SSID}' on ${AP_IF} (${ap_mac}) matching Channel ${chan} (mode ${mode})..."
  if ! iw dev "${STA_IF}" interface add "${AP_IF}" type __ap 2>/dev/null; then
    log "WARN: Could not add virtual ${AP_IF} right now; keeping ${STA_IF} untouched."
    return 1
  fi
  nmcli dev set "${AP_IF}" managed no 2>/dev/null || true
  ip link set dev "${AP_IF}" address "${ap_mac}" 2>/dev/null || true

  cat > "${RUN_DIR}/hostapd.conf" <<EOF
interface=${AP_IF}
driver=nl80211
ssid=${SSID}
hw_mode=${mode}
channel=${chan}
ieee80211n=1
wmm_enabled=1
auth_algs=1
wpa=2
wpa_passphrase=${PASS}
wpa_key_mgmt=WPA-PSK
rsn_pairwise=CCMP
EOF

  if ! hostapd -B -P "${RUN_DIR}/hostapd.pid" "${RUN_DIR}/hostapd.conf" >/dev/null 2>&1; then
    log "WARN: hostapd refused Channel ${chan} (e.g. 5GHz DFS/LAR restriction on ${STA_IF}); tearing down ${AP_IF} cleanly so upstream Wi-Fi stays 100% healthy."
    stop_ap
    return 1
  fi
  sleep 0.8
  HOSTAPD_PID="$(cat "${RUN_DIR}/hostapd.pid" 2>/dev/null || true)"
  if [[ -z "${HOSTAPD_PID}" ]] || ! kill -0 "${HOSTAPD_PID}" 2>/dev/null; then
    log "WARN: hostapd did not stay active on Channel ${chan}; releasing ${AP_IF} safely."
    stop_ap
    return 1
  fi

  ip addr flush dev "${AP_IF}"
  ip addr add "${AP_IP}/24" dev "${AP_IF}"

  cat > "${RUN_DIR}/dnsmasq.conf" <<EOF
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
EOF
  dnsmasq -C "${RUN_DIR}/dnsmasq.conf" --pid-file="${RUN_DIR}/dnsmasq.pid"
  DNSMASQ_PID="$(cat "${RUN_DIR}/dnsmasq.pid" 2>/dev/null || true)"

  sysctl -w net.ipv4.ip_forward=1 >/dev/null
  iptables -t nat -C POSTROUTING -s "${AP_SUBNET}" -o "${STA_IF}" -j MASQUERADE 2>/dev/null || \
    iptables -t nat -A POSTROUTING -s "${AP_SUBNET}" -o "${STA_IF}" -j MASQUERADE
  iptables -C FORWARD -i "${AP_IF}" -o "${STA_IF}" -j ACCEPT 2>/dev/null || \
    iptables -A FORWARD -i "${AP_IF}" -o "${STA_IF}" -j ACCEPT
  iptables -C FORWARD -i "${STA_IF}" -o "${AP_IF}" -m state --state RELATED,ESTABLISHED -j ACCEPT 2>/dev/null || \
    iptables -A FORWARD -i "${STA_IF}" -o "${AP_IF}" -m state --state RELATED,ESTABLISHED -j ACCEPT

  CURRENT_CHAN="${chan}"
  AP_ACTIVE=1
  log "LIVE: '${SSID}' broadcasting on ${AP_IF} (${AP_IP}/24, Channel ${chan})!"
  return 0
}

get_sta_state() {
  nmcli -t -f DEVICE,STATE dev status 2>/dev/null | awk -F: -v dev="${STA_IF}" '$1==dev {print $2; exit}'
}

get_sta_channel() {
  iw dev "${STA_IF}" info 2>/dev/null | awk '/channel / {print $2; exit}'
}

# 1. Boot settle: give NetworkManager up to 15s to associate wlp3s0 cleanly
#    before creating ap0 so boot-time Wi-Fi never collides on rtw89.
log "Starting FaithNet persistent Micro-ISP AP daemon on ${STA_IF}..."
for _ in $(seq 1 15); do
  st="$(get_sta_state)"
  ch="$(get_sta_channel)"
  if [[ "${st}" == "connected" && -n "${ch}" ]]; then
    break
  fi
  sleep 1
done

init_ch="$(get_sta_channel)"
if [[ -n "${init_ch}" ]]; then
  start_ap "${init_ch}" || true
else
  log "No upstream Wi-Fi after boot settle; launching offline island on Channel 6..."
  start_ap 6 || true
fi

# 2. Event-driven monitor (process substitution keeps main shell PID/traps!).
#    Releases ap0 IMMEDIATELY on any connecting/disconnected event so Realtek
#    rtw89 never locks up during Wi-Fi or mobile-hotspot channel switches!
exec 3< <(nmcli device monitor "${STA_IF}" 2>/dev/null)
OFFLINE_WAIT=0

while true; do
  if [[ "${OFFLINE_WAIT}" -eq 1 ]]; then
    if ! IFS= read -r -t 10 -u 3 line; then
      # 10s elapsed with no new NM event: check if still disconnected (blackout).
      OFFLINE_WAIT=0
      if [[ "$(get_sta_state)" == "disconnected" && "${AP_ACTIVE}" -eq 0 ]]; then
        log "Upstream still offline after 10s -> booting standalone offline island on Channel 6!"
        start_ap 6 || true
      fi
      continue
    fi
  else
    if ! IFS= read -r -u 3 line; then
      break
    fi
  fi

  case "${line}" in
    *"connecting"*|*"configuring"*|*"deactivating"*|*"unavailable"*)
      OFFLINE_WAIT=0
      if [[ "${AP_ACTIVE}" -eq 1 ]]; then
        log "Upstream event (${line}) -> immediately releasing ${AP_IF} for clean radio handoff!"
        stop_ap
      fi
      ;;
    *"disconnected"*)
      if [[ "${AP_ACTIVE}" -eq 1 ]]; then
        log "Upstream disconnected -> releasing ${AP_IF} so NetworkManager can scan freely..."
        stop_ap
      fi
      OFFLINE_WAIT=1
      ;;
    *": connected"*)
      OFFLINE_WAIT=0
      sleep 2
      new_ch="$(get_sta_channel)"
      if [[ -n "${new_ch}" && ( "${AP_ACTIVE}" -eq 0 || "${new_ch}" != "${CURRENT_CHAN}" ) ]]; then
        log "Upstream connected on Channel ${new_ch} -> bringing '${SSID}' online!"
        start_ap "${new_ch}" || true
      fi
      ;;
  esac
done
