#!/usr/bin/env bash
# ==============================================================================
# BLOCK 1: METADATA BLOCK
# ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
# ==============================================================================
# @project       FaithNet
# @file          tests/playtester/field_kits/01_two_device_room_well/faithnet-ap-daemon.sh
# @id            playtest.field_kit.01.linux.ap_daemon
# @layer         Layer 4: Proving Ground (Field Kit 01 — Persistent AP Daemon)
# @archetype     field_daemon | operational
# @version       0.1.0-dev
# @interface_rev Rev 5
# @date          2026-11-17
# @updated       2026-11-17
# @author        CreativeWorkzStudio LLC / ProfessorSeanEX
# @repository    https://github.com/CreativeWorkzStudio/FaithNet
# @system        FaithNet Persistent Event-Driven Micro-ISP AP Daemon (Linux)
# @role          Boot-Persistent Simultaneous STA+AP With Real-Time Radio Guard
# @maturity      Rev 5 Baseline | Wave 14 5-Block Cross-Platform Parity
# @chassis       systemd-daemon | Non-blocking nmcli monitor + zero-brick safety
# @scale         Single-radio Realtek rtw89/mac80211 safe STA+AP + Ch 6 island
# @seams         in: nmcli, iw, hostapd, dnsmasq, iptables | exports: ap0
# @invariants    1. Universal 5-Block Lifecycle Protocol (Rev 5).
#                2. Zero-brick guarantee: releases ap0 IMMEDIATELY on upstream
#                   connecting/disconnected events so rtw89 never locks PCI.
#                3. Graceful degradation: logs warning and preserves upstream
#                   Wi-Fi if a 5 GHz DFS channel forbids simultaneous AP mode.
#                4. Automatic offline island fallback (Ch 6) after 10s blackout.
#
# PURPOSE & ARCHITECTURAL ROLE:
#   System-level persistent daemon backing `/etc/systemd/system/faithnet-ap.service`.
#   Ensures Laptop #1 auto-boots into a FaithNet-HomeRoom micro-ISP router so
#   previously paired phones, laptops, and consoles auto-reconnect on startup,
#   while safely releasing `ap0` during upstream Wi-Fi/hotspot switches.
#
#   ### For Architects, Operators & Non-Programmers:
#   Keeps FaithNet Wi-Fi automatically running whenever the laptop turns on and
#   safely steps aside for 2 seconds whenever you switch to a phone hotspot.
#
#   ### For Systems Programmers & Engine Architects:
#   Uses process substitution (`exec 3< <(nmcli device monitor ...)`) so traps
#   and PID variables remain in the main shell, with non-blocking `read -t 10`
#   blackout detection and idempotent iptables/iw teardown.
# ==============================================================================
set -euo pipefail

# ==============================================================================
# BLOCK 2: OPENING BLOCK
# ( Operational Constants, Mutable Daemon State & Safe Teardown Trap )
# ==============================================================================
# Engine:        FaithNet Adaptive Networking Substrate
# Stratum:       Layer 4: Proving Ground (Field Kit 01 — Linux Daemon)
# Module:        playtest.field_kit.01.linux.ap_daemon
# Mission:       Define runtime paths, state flags, logging, and trap cleanup.
# ==============================================================================

readonly STA_IF="${STA_IF:-wlp3s0}"
readonly AP_IF="${AP_IF:-ap0}"
readonly SSID="${SSID:-FaithNet-HomeRoom}"
readonly PASS="${PASS:-FaithNet2026}"
readonly AP_IP="${AP_IP:-10.77.77.1}"
readonly AP_SUBNET="${AP_SUBNET:-10.77.77.0/24}"
readonly RUN_DIR="/run/faithnet-ap"

HOSTAPD_PID=""
DNSMASQ_PID=""
CURRENT_CHAN=""
AP_ACTIVE=0

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

# ==============================================================================
# BLOCK 3: BODY BLOCK
# ( Radio Query Helpers, Safe AP Bring-Up & Event-Driven Switch Loop )
# ==============================================================================
# Engine:        FaithNet Adaptive Networking Substrate
# Stratum:       Layer 4: Proving Ground (Field Kit 01 — Linux Daemon)
# Module:        playtest.field_kit.01.linux.ap_daemon
# Mission:       Manage ap0 lifecycle across boot, hotspot switch & blackout.
# ==============================================================================

derive_ap_mac() {
  local base_mac first new_first
  base_mac="$(cat "/sys/class/net/${STA_IF}/address" 2>/dev/null || echo "44:fa:66:89:f6:f5")"
  first="${base_mac%%:*}"
  new_first="$(printf '%02x' $(( (0x${first} | 0x02) ^ 0x04 )))"
  echo "${new_first}:${base_mac#*:}"
}

get_sta_state() {
  nmcli -t -f DEVICE,STATE dev status 2>/dev/null | awk -F: -v dev="${STA_IF}" '$1==dev {print $2; exit}'
}

get_sta_channel() {
  iw dev "${STA_IF}" info 2>/dev/null | awk '/channel / {print $2; exit}'
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
    log "WARN: hostapd refused Channel ${chan} (e.g. 5GHz DFS/LAR); releasing ${AP_IF} cleanly."
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

run_event_monitor_loop() {
  exec 3< <(nmcli device monitor "${STA_IF}" 2>/dev/null)
  local offline_wait=0
  local line=""

  while true; do
    if [[ "${offline_wait}" -eq 1 ]]; then
      if ! IFS= read -r -t 10 -u 3 line; then
        offline_wait=0
        if [[ "$(get_sta_state)" == "disconnected" && "${AP_ACTIVE}" -eq 0 ]]; then
          log "Upstream offline >10s -> booting standalone offline island on Channel 6!"
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
        offline_wait=0
        if [[ "${AP_ACTIVE}" -eq 1 ]]; then
          log "Upstream event (${line}) -> immediately releasing ${AP_IF} for clean handoff!"
          stop_ap
        fi
        ;;
      *"disconnected"*)
        if [[ "${AP_ACTIVE}" -eq 1 ]]; then
          log "Upstream disconnected -> releasing ${AP_IF} so NetworkManager scans freely..."
          stop_ap
        fi
        offline_wait=1
        ;;
      *": connected"*)
        offline_wait=0
        sleep 2
        local new_ch
        new_ch="$(get_sta_channel)"
        if [[ -n "${new_ch}" && ( "${AP_ACTIVE}" -eq 0 || "${new_ch}" != "${CURRENT_CHAN}" ) ]]; then
          log "Upstream connected on Channel ${new_ch} -> bringing '${SSID}' online!"
          start_ap "${new_ch}" || true
        fi
        ;;
    esac
  done
}

# ==============================================================================
# BLOCK 4: AUDIT PROCEDURES BLOCK
# ( Pre-Flight Privilege, Runtime Directory & Wireless Tool Audit )
# ==============================================================================
# Engine:        FaithNet Adaptive Networking Substrate
# Stratum:       Layer 4: Proving Ground (Field Kit 01 — Linux Daemon)
# Module:        playtest.field_kit.01.linux.ap_daemon
# Mission:       Verify root capability and required Linux wireless toolchain.
# ==============================================================================

preflight_ap_daemon() {
  if [[ "${EUID}" -ne 0 ]]; then
    log "PREFLIGHT FAIL: faithnet-ap-daemon.sh requires root (run via systemd)."
    return 1
  fi
  for tool in iw ip hostapd dnsmasq iptables nmcli; do
    if ! command -v "${tool}" >/dev/null 2>&1; then
      log "PREFLIGHT FAIL: missing required wireless dependency '${tool}'."
      return 1
    fi
  done
  mkdir -p "${RUN_DIR}"
  return 0
}

# ==============================================================================
# BLOCK 5: CLOSING BLOCK
# ( Boot Settle Sequence & Event-Loop Lifecycle Entrypoint )
# ==============================================================================
# Engine:        FaithNet Adaptive Networking Substrate
# Stratum:       Layer 4: Proving Ground (Field Kit 01 — Linux Daemon)
# Module:        playtest.field_kit.01.linux.ap_daemon
# Mission:       Settle upstream Wi-Fi cleanly on boot, then run monitor loop.
# ==============================================================================

main() {
  preflight_ap_daemon
  log "Starting FaithNet persistent Micro-ISP AP daemon on ${STA_IF}..."
  for _ in $(seq 1 15); do
    if [[ "$(get_sta_state)" == "connected" && -n "$(get_sta_channel)" ]]; then
      break
    fi
    sleep 1
  done

  local init_ch
  init_ch="$(get_sta_channel)"
  if [[ -n "${init_ch}" ]]; then
    start_ap "${init_ch}" || true
  else
    log "No upstream Wi-Fi after boot settle; launching offline island on Channel 6..."
    start_ap 6 || true
  fi

  run_event_monitor_loop
}

main "$@"

# End of tests/playtester/field_kits/01_two_device_room_well/faithnet-ap-daemon.sh
