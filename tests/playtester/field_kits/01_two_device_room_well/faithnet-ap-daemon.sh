#!/usr/bin/env bash
# ==============================================================================
# BLOCK 1: METADATA BLOCK
# ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
# ==============================================================================
# @project       FaithNet
# @file          tests/playtester/field_kits/01_two_device_room_well/faithnet-ap-daemon.sh
# @id            playtest.field_kit.01.linux.ap_daemon
# @layer         Layer 4: Proving Ground (Field Kit 01 — Sovereign Micro-ISP Daemon)
# @archetype     field_daemon | operational
# @version       0.3.0-dev
# @interface_rev Rev 5
# @date          2026-11-17
# @updated       2026-11-17
# @author        CreativeWorkzStudio LLC / ProfessorSeanEX
# @repository    https://github.com/Novaking34/FaithNet
# @system        FaithNet Persistent Sovereign Micro-ISP & Firewalled WAN Daemon
# @role          Boot-Persistent AP + Multi-Channel Switch Handoff + Split-DNS + WAN Guard
# @maturity      Rev 5 Baseline | Wave 21 Seamless Upstream Switch & Multi-Channel Support
# @chassis       systemd-daemon | Non-blocking FIFO nmcli monitor + USR1/USR2 switch handoff
# @scale         Auto-selects DEDICATED_RADIO (2nd USB Wi-Fi) or SHARED_RADIO (STA+AP)
# @seams         in: nmcli, iw, hostapd, dnsmasq, iptables, resolvectl | exports: ap0
# @invariants    1. Universal 5-Block Lifecycle Protocol (Rev 5).
#                2. Zero-brick guarantee: releases shared ap0 IMMEDIATELY on USR1 or
#                   upstream switching events so rtw89 never hits queue-flush timeouts.
#                3. Valid US non-DFS channel support across 2.4GHz (1-11) & 5GHz
#                   (36-48, 149-165) with explicit country_code=US & roam tracking.
#                4. Scoped Rung 4 WAN firewall (`FAITHNET_WAN_GUARD`): blocks
#                   unsolicited inbound traffic on wlp3s0 while keeping ap0 & lo open.
#                5. Clean <1s SIGTERM shutdown killing nmcli monitor (zero SIGKILL).
#
# PURPOSE & ARCHITECTURAL ROLE:
#   System-level persistent daemon backing `/etc/systemd/system/faithnet-ap.service`.
#   Makes `FaithNet-HomeRoom` this PC's primary sovereign room identity (with
#   native local `.lan`/`.home.arpa` split-DNS and firewalled commercial WAN on
#   `wlp3s0`), supports instant `bridge` vs. `island` egress switching, and
#   auto-detects any dedicated 2nd USB Wi-Fi adapter for zero-blink AP radio.
#
#   ### For Architects, Operators & Non-Programmers:
#   Keeps FaithNet-HomeRoom automatically live on boot, shields this laptop from
#   unsolicited commercial Wi-Fi probes, lets this laptop open `faithnet.lan`
#   natively, and automatically upgrades to a dedicated USB Wi-Fi adapter the
#   moment one is plugged in.
#
#   ### For Systems Programmers & Engine Architects:
#   Loads optional `/etc/faithnet-ap.conf`, detects secondary `wiphy` radios via
#   `iw dev`, applies isolated `FAITHNET_WAN_GUARD` / `FAITHNET_WAN_EGRESS`
#   iptables chains, configures `resolvectl` routing-only domains, and tears down
#   every rule idempotently on EXIT/INT/TERM.
# ==============================================================================
set -euo pipefail

# ==============================================================================
# BLOCK 2: OPENING BLOCK
# ( Operational Config, Mutable Daemon State & Zero-Brick Teardown Trap )
# ==============================================================================
# Engine:        FaithNet Adaptive Networking Substrate
# Stratum:       Layer 4: Proving Ground (Field Kit 01 — Linux Daemon)
# Module:        playtest.field_kit.01.linux.ap_daemon
# Mission:       Load config, define scoped firewall/DNS state & teardown trap.
# ==============================================================================

readonly CONF_FILE="/etc/faithnet-ap.conf"
if [[ -f "${CONF_FILE}" ]]; then
  # shellcheck disable=SC1090
  source "${CONF_FILE}"
fi

STA_IF="${STA_IF:-wlp3s0}"
AP_IF="${AP_IF:-ap0}"
DEDICATED_AP_IF="${DEDICATED_AP_IF:-}"
DEDICATED_CHANNEL="${DEDICATED_CHANNEL:-6}"
SSID="${SSID:-FaithNet-HomeRoom}"
PASS="${PASS:-FaithNet2026}"
AP_IP="${AP_IP:-10.77.77.1}"
AP_SUBNET="${AP_SUBNET:-10.77.77.0/24}"
EGRESS_MODE="${EGRESS_MODE:-bridge}"
COUNTRY_CODE="${COUNTRY_CODE:-US}"
readonly RUN_DIR="/run/faithnet-ap"
readonly SWITCH_HOLD_FILE="${RUN_DIR}/switch.hold"
readonly DAEMON_PID_FILE="${RUN_DIR}/daemon.pid"
readonly NM_FIFO="${RUN_DIR}/nm_events.fifo"
readonly WAN_GUARD_CHAIN="FAITHNET_WAN_GUARD"
readonly WAN_EGRESS_CHAIN="FAITHNET_WAN_EGRESS"

HOSTAPD_PID=""
DNSMASQ_PID=""
NM_MON_PID=""
CURRENT_CHAN=""
AP_ACTIVE=0
RADIO_MODE="SHARED_RADIO"
ACTIVE_AP_IF="${AP_IF}"
CREATED_VIRTUAL_AP=0
SWITCH_HOLD=0
NEED_REBIND=0
SHUTTING_DOWN=0
ORIG_IP_FORWARD="0"

log() {
  echo "[faithnet-ap $(date '+%H:%M:%S')] $*"
}

clear_host_split_dns() {
  local iface="${1:-${ACTIVE_AP_IF}}"
  if command -v resolvectl >/dev/null 2>&1; then
    resolvectl revert "${iface}" 2>/dev/null || true
  fi
}

clear_wan_firewall() {
  iptables -D INPUT -i "${STA_IF}" -j "${WAN_GUARD_CHAIN}" 2>/dev/null || true
  iptables -F "${WAN_GUARD_CHAIN}" 2>/dev/null || true
  iptables -X "${WAN_GUARD_CHAIN}" 2>/dev/null || true

  iptables -D OUTPUT -o "${STA_IF}" -j "${WAN_EGRESS_CHAIN}" 2>/dev/null || true
  iptables -F "${WAN_EGRESS_CHAIN}" 2>/dev/null || true
  iptables -X "${WAN_EGRESS_CHAIN}" 2>/dev/null || true

  if command -v ip6tables >/dev/null 2>&1; then
    ip6tables -D INPUT -i "${STA_IF}" -j "${WAN_GUARD_CHAIN}" 2>/dev/null || true
    ip6tables -F "${WAN_GUARD_CHAIN}" 2>/dev/null || true
    ip6tables -X "${WAN_GUARD_CHAIN}" 2>/dev/null || true

    ip6tables -D OUTPUT -o "${STA_IF}" -j "${WAN_EGRESS_CHAIN}" 2>/dev/null || true
    ip6tables -F "${WAN_EGRESS_CHAIN}" 2>/dev/null || true
    ip6tables -X "${WAN_EGRESS_CHAIN}" 2>/dev/null || true
  fi
}

stop_ap() {
  if [[ "${AP_ACTIVE}" -eq 1 ]] || ip link show "${ACTIVE_AP_IF}" >/dev/null 2>&1; then
    log "Releasing ${ACTIVE_AP_IF} cleanly (leaving ${STA_IF} free & untouched)..."
  fi
  clear_host_split_dns "${ACTIVE_AP_IF}"
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
  iptables -D FORWARD -i "${ACTIVE_AP_IF}" -o "${STA_IF}" -j ACCEPT 2>/dev/null || true
  iptables -D FORWARD -i "${STA_IF}" -o "${ACTIVE_AP_IF}" -m state --state RELATED,ESTABLISHED -j ACCEPT 2>/dev/null || true
  if [[ "${CREATED_VIRTUAL_AP}" -eq 1 ]] || [[ "${RADIO_MODE}" == "SHARED_RADIO" && "${ACTIVE_AP_IF}" == "${AP_IF}" ]]; then
    ip link set "${ACTIVE_AP_IF}" down 2>/dev/null || true
    iw dev "${ACTIVE_AP_IF}" del 2>/dev/null || true
    CREATED_VIRTUAL_AP=0
  else
    ip addr flush dev "${ACTIVE_AP_IF}" 2>/dev/null || true
  fi
  AP_ACTIVE=0
  CURRENT_CHAN=""
}

handle_switch_yield() {
  SWITCH_HOLD=1
  NEED_REBIND=0
  log "SIGUSR1 received -> yielding ${ACTIVE_AP_IF} cleanly for upstream Wi-Fi switch (FaithNet stays live)..."
  stop_ap
  write_runtime_state 2>/dev/null || true
}

handle_switch_resume() {
  SWITCH_HOLD=0
  NEED_REBIND=1
  log "SIGUSR2 received -> resuming FaithNet-HomeRoom on active upstream channel..."
}

handle_shutdown() {
  SHUTTING_DOWN=1
  exit 0
}

cleanup() {
  if [[ -n "${NM_MON_PID}" ]]; then
    kill "${NM_MON_PID}" 2>/dev/null || true
    NM_MON_PID=""
  fi
  stop_ap
  clear_wan_firewall
  sysctl -w "net.ipv4.ip_forward=${ORIG_IP_FORWARD}" >/dev/null 2>&1 || true
  rm -rf "${RUN_DIR}"
}
trap cleanup EXIT
trap handle_shutdown INT TERM
trap handle_switch_yield USR1
trap handle_switch_resume USR2

# ==============================================================================
# BLOCK 3: BODY BLOCK
# ( Adapter Auto-Detect, Scoped WAN Firewall, Split-DNS & AP Lifecycle )
# ==============================================================================
# Engine:        FaithNet Adaptive Networking Substrate
# Stratum:       Layer 4: Proving Ground (Field Kit 01 — Linux Daemon)
# Module:        playtest.field_kit.01.linux.ap_daemon
# Mission:       Manage DEDICATED_RADIO / SHARED_RADIO AP, WAN guard & DNS.
# ==============================================================================

detect_dedicated_adapter() {
  if [[ -n "${DEDICATED_AP_IF}" ]] && ip link show "${DEDICATED_AP_IF}" >/dev/null 2>&1; then
    RADIO_MODE="DEDICATED_RADIO"
    ACTIVE_AP_IF="${DEDICATED_AP_IF}"
    return 0
  fi

  local wan_phy
  wan_phy="$(iw dev "${STA_IF}" info 2>/dev/null | awk '/wiphy/ {print $2; exit}')"
  if [[ -z "${wan_phy}" ]]; then
    RADIO_MODE="SHARED_RADIO"
    ACTIVE_AP_IF="${AP_IF}"
    return 1
  fi

  local cand
  cand="$(iw dev 2>/dev/null | awk -v wphy="phy#${wan_phy}" '
    /^phy#/ { cur=$1; next }
    /Interface / {
      if (cur != wphy && $2 != "") { print $2; exit }
    }
  ')"

  if [[ -n "${cand}" ]] && ip link show "${cand}" >/dev/null 2>&1; then
    RADIO_MODE="DEDICATED_RADIO"
    ACTIVE_AP_IF="${cand}"
    return 0
  fi

  RADIO_MODE="SHARED_RADIO"
  ACTIVE_AP_IF="${AP_IF}"
  return 1
}

apply_wan_firewall() {
  # Scoped dual-stack (IPv4 + IPv6) inbound firewall on commercial WAN (wlp3s0):
  # Permits replies to outbound traffic, ICMP/ICMPv6, and DHCP/DHCPv6 renewals;
  # drops unsolicited inbound probes so only FaithNet-HomeRoom (ap0 + lo)
  # reaches local services (:8077 The Well, :7777 FNP/1).
  iptables -N "${WAN_GUARD_CHAIN}" 2>/dev/null || iptables -F "${WAN_GUARD_CHAIN}"
  iptables -A "${WAN_GUARD_CHAIN}" -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT
  iptables -A "${WAN_GUARD_CHAIN}" -p icmp -j ACCEPT
  iptables -A "${WAN_GUARD_CHAIN}" -p udp --sport 67 --dport 68 -j ACCEPT
  iptables -A "${WAN_GUARD_CHAIN}" -p udp --dport 68 -j ACCEPT
  iptables -A "${WAN_GUARD_CHAIN}" -j DROP
  iptables -C INPUT -i "${STA_IF}" -j "${WAN_GUARD_CHAIN}" 2>/dev/null || \
    iptables -I INPUT 1 -i "${STA_IF}" -j "${WAN_GUARD_CHAIN}"

  if command -v ip6tables >/dev/null 2>&1; then
    ip6tables -N "${WAN_GUARD_CHAIN}" 2>/dev/null || ip6tables -F "${WAN_GUARD_CHAIN}"
    ip6tables -A "${WAN_GUARD_CHAIN}" -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT
    ip6tables -A "${WAN_GUARD_CHAIN}" -p ipv6-icmp -j ACCEPT
    ip6tables -A "${WAN_GUARD_CHAIN}" -p udp --sport 547 --dport 546 -j ACCEPT
    ip6tables -A "${WAN_GUARD_CHAIN}" -j DROP
    ip6tables -C INPUT -i "${STA_IF}" -j "${WAN_GUARD_CHAIN}" 2>/dev/null || \
      ip6tables -I INPUT 1 -i "${STA_IF}" -j "${WAN_GUARD_CHAIN}"
  fi

  apply_egress_policy
}

apply_egress_policy() {
  local mode="${EGRESS_MODE}"
  if [[ -f "${RUN_DIR}/egress_mode" ]]; then
    mode="$(cat "${RUN_DIR}/egress_mode" 2>/dev/null || echo "${EGRESS_MODE}")"
  fi

  iptables -D OUTPUT -o "${STA_IF}" -j "${WAN_EGRESS_CHAIN}" 2>/dev/null || true
  iptables -F "${WAN_EGRESS_CHAIN}" 2>/dev/null || true
  iptables -X "${WAN_EGRESS_CHAIN}" 2>/dev/null || true

  if command -v ip6tables >/dev/null 2>&1; then
    ip6tables -D OUTPUT -o "${STA_IF}" -j "${WAN_EGRESS_CHAIN}" 2>/dev/null || true
    ip6tables -F "${WAN_EGRESS_CHAIN}" 2>/dev/null || true
    ip6tables -X "${WAN_EGRESS_CHAIN}" 2>/dev/null || true
  fi

  if [[ "${mode}" == "island" ]]; then
    log "Egress mode = ISLAND (dual-stack IPv4+IPv6 WAN ${STA_IF} firewalled off; FaithNet-HomeRoom stays 100% live)."
    iptables -t nat -D POSTROUTING -s "${AP_SUBNET}" -o "${STA_IF}" -j MASQUERADE 2>/dev/null || true
    iptables -D FORWARD -i "${ACTIVE_AP_IF}" -o "${STA_IF}" -j ACCEPT 2>/dev/null || true
    iptables -D FORWARD -i "${STA_IF}" -o "${ACTIVE_AP_IF}" -m state --state RELATED,ESTABLISHED -j ACCEPT 2>/dev/null || true

    iptables -N "${WAN_EGRESS_CHAIN}" 2>/dev/null || iptables -F "${WAN_EGRESS_CHAIN}"
    iptables -A "${WAN_EGRESS_CHAIN}" -p udp --sport 68 --dport 67 -j ACCEPT
    iptables -A "${WAN_EGRESS_CHAIN}" -j REJECT --reject-with icmp-net-prohibited
    iptables -I OUTPUT 1 -o "${STA_IF}" -j "${WAN_EGRESS_CHAIN}"

    if command -v ip6tables >/dev/null 2>&1; then
      ip6tables -N "${WAN_EGRESS_CHAIN}" 2>/dev/null || ip6tables -F "${WAN_EGRESS_CHAIN}"
      ip6tables -A "${WAN_EGRESS_CHAIN}" -p ipv6-icmp -j ACCEPT
      ip6tables -A "${WAN_EGRESS_CHAIN}" -p udp --sport 546 --dport 547 -j ACCEPT
      ip6tables -A "${WAN_EGRESS_CHAIN}" -j REJECT --reject-with icmp6-adm-prohibited
      ip6tables -I OUTPUT 1 -o "${STA_IF}" -j "${WAN_EGRESS_CHAIN}"
    fi
  else
    log "Egress mode = BRIDGE (dual-stack inbound ${STA_IF} firewalled; outbound WAN & HomeRoom NAT active)."
    sysctl -w net.ipv4.ip_forward=1 >/dev/null
    iptables -t nat -C POSTROUTING -s "${AP_SUBNET}" -o "${STA_IF}" -j MASQUERADE 2>/dev/null || \
      iptables -t nat -A POSTROUTING -s "${AP_SUBNET}" -o "${STA_IF}" -j MASQUERADE
    iptables -C FORWARD -i "${ACTIVE_AP_IF}" -o "${STA_IF}" -j ACCEPT 2>/dev/null || \
      iptables -A FORWARD -i "${ACTIVE_AP_IF}" -o "${STA_IF}" -j ACCEPT
    iptables -C FORWARD -i "${STA_IF}" -o "${ACTIVE_AP_IF}" -m state --state RELATED,ESTABLISHED -j ACCEPT 2>/dev/null || \
      iptables -A FORWARD -i "${STA_IF}" -o "${ACTIVE_AP_IF}" -m state --state RELATED,ESTABLISHED -j ACCEPT
  fi
  echo "${mode}" > "${RUN_DIR}/egress_mode"
}

apply_host_split_dns() {
  if command -v resolvectl >/dev/null 2>&1; then
    resolvectl dns "${ACTIVE_AP_IF}" "${AP_IP}" 2>/dev/null || true
    resolvectl domain "${ACTIVE_AP_IF}" "~faithnet.lan" "~faithnet.home.arpa" 2>/dev/null || true
    resolvectl default-route "${ACTIVE_AP_IF}" no 2>/dev/null || true
    log "Host split-DNS active on ${ACTIVE_AP_IF}: ~faithnet.lan & ~faithnet.home.arpa -> ${AP_IP}"
  fi
}

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

get_sta_freq() {
  iw dev "${STA_IF}" info 2>/dev/null | awk '/channel / {gsub(/[()]/, "", $3); print $3; exit}'
}

is_valid_ap_channel() {
  local chan="${1:-}"
  local freq="${2:-0}"
  [[ "${chan}" =~ ^[0-9]+$ ]] || return 1
  if [[ "${freq}" =~ ^[0-9]+$ ]] && (( freq >= 5900 )); then
    return 1
  fi
  if (( chan >= 1 && chan <= 11 )); then
    return 0
  fi
  case "${chan}" in
    36|40|44|48|149|153|157|161|165) return 0 ;;
    *) return 1 ;;
  esac
}

is_switch_held() {
  [[ "${SWITCH_HOLD}" -eq 1 || -f "${SWITCH_HOLD_FILE}" ]]
}

write_runtime_state() {
  cat > "${RUN_DIR}/state.env" <<EOF
RADIO_MODE=${RADIO_MODE}
STA_IF=${STA_IF}
ACTIVE_AP_IF=${ACTIVE_AP_IF}
SSID=${SSID}
AP_IP=${AP_IP}
CURRENT_CHAN=${CURRENT_CHAN}
AP_ACTIVE=${AP_ACTIVE}
SWITCH_HOLD=${SWITCH_HOLD}
EOF
}

start_ap() {
  local chan="$1"
  local freq="${2:-0}"
  if ! is_valid_ap_channel "${chan}" "${freq}"; then
    log "INFO: Upstream Channel ${chan} (${freq} MHz) is DFS/6GHz/unsupported for shared AP; keeping ${STA_IF} free & undisturbed."
    return 1
  fi

  local mode="g"
  if (( chan > 14 )); then
    mode="a"
  fi

  stop_ap
  sleep 0.3

  if [[ "${RADIO_MODE}" == "DEDICATED_RADIO" ]]; then
    log "Bringing up '${SSID}' on DEDICATED radio ${ACTIVE_AP_IF} (Channel ${chan}, mode ${mode})..."
    nmcli dev set "${ACTIVE_AP_IF}" managed no 2>/dev/null || true
    CREATED_VIRTUAL_AP=0
  else
    local ap_mac
    ap_mac="$(derive_ap_mac)"
    log "Bringing up '${SSID}' on SHARED radio ${ACTIVE_AP_IF} (${ap_mac}) matching Channel ${chan} (mode ${mode}, country ${COUNTRY_CODE})..."
    if ! iw dev "${STA_IF}" interface add "${ACTIVE_AP_IF}" type __ap 2>/dev/null; then
      log "WARN: Could not add virtual ${ACTIVE_AP_IF} right now; keeping ${STA_IF} untouched."
      return 1
    fi
    CREATED_VIRTUAL_AP=1
    nmcli dev set "${ACTIVE_AP_IF}" managed no 2>/dev/null || true
    ip link set dev "${ACTIVE_AP_IF}" address "${ap_mac}" 2>/dev/null || true
  fi

  cat > "${RUN_DIR}/hostapd.conf" <<EOF
interface=${ACTIVE_AP_IF}
driver=nl80211
country_code=${COUNTRY_CODE}
ieee80211d=1
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
    log "WARN: hostapd refused Channel ${chan} (e.g. 5GHz DFS/LAR); releasing ${ACTIVE_AP_IF} cleanly."
    stop_ap
    return 1
  fi
  sleep 0.8
  HOSTAPD_PID="$(cat "${RUN_DIR}/hostapd.pid" 2>/dev/null || true)"
  if [[ -z "${HOSTAPD_PID}" ]] || ! kill -0 "${HOSTAPD_PID}" 2>/dev/null; then
    log "WARN: hostapd did not stay active on Channel ${chan}; releasing ${ACTIVE_AP_IF} safely."
    stop_ap
    return 1
  fi

  ip addr flush dev "${ACTIVE_AP_IF}"
  ip addr add "${AP_IP}/24" dev "${ACTIVE_AP_IF}"

  cat > "${RUN_DIR}/dnsmasq.conf" <<EOF
interface=${ACTIVE_AP_IF}
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

  apply_host_split_dns
  apply_wan_firewall

  CURRENT_CHAN="${chan}"
  AP_ACTIVE=1
  NEED_REBIND=0
  write_runtime_state
  log "LIVE [${RADIO_MODE}]: '${SSID}' broadcasting on ${ACTIVE_AP_IF} (${AP_IP}/24, Channel ${chan})!"
  return 0
}

run_dedicated_radio_loop() {
  log "Dedicated 2nd Wi-Fi adapter (${ACTIVE_AP_IF}) active -> FaithNet-HomeRoom stays permanently up even while ${STA_IF} roams!"
  start_ap "${DEDICATED_CHANNEL}" || true
  while [[ "${SHUTTING_DOWN}" -eq 0 ]]; do
    sleep 5 &
    wait $! 2>/dev/null || true
    if [[ -n "${HOSTAPD_PID}" ]] && ! kill -0 "${HOSTAPD_PID}" 2>/dev/null; then
      log "Dedicated hostapd exited unexpectedly; recovering on Channel ${DEDICATED_CHANNEL}..."
      start_ap "${DEDICATED_CHANNEL}" || true
    fi
  done
}

sync_shared_ap_to_upstream() {
  if is_switch_held; then
    if [[ "${AP_ACTIVE}" -eq 1 ]]; then
      stop_ap
      write_runtime_state
    fi
    return 0
  fi

  local state new_ch new_freq
  state="$(get_sta_state)"
  if [[ "${state}" != "connected" ]]; then
    return 0
  fi

  new_ch="$(get_sta_channel)"
  new_freq="$(get_sta_freq)"
  if [[ -n "${new_ch}" && ( "${AP_ACTIVE}" -eq 0 || "${NEED_REBIND}" -eq 1 || "${new_ch}" != "${CURRENT_CHAN}" ) ]]; then
    if is_valid_ap_channel "${new_ch}" "${new_freq}"; then
      log "Upstream connected on Channel ${new_ch} (${new_freq} MHz) -> syncing '${SSID}'!"
      start_ap "${new_ch}" "${new_freq}" || true
    else
      if [[ "${AP_ACTIVE}" -eq 1 ]]; then
        log "Upstream roamed to unsupported/DFS Channel ${new_ch} (${new_freq} MHz) -> yielding ${ACTIVE_AP_IF} cleanly."
        stop_ap
        write_runtime_state
      fi
      NEED_REBIND=0
    fi
  fi
}

run_shared_event_monitor_loop() {
  rm -f "${NM_FIFO}"
  mkfifo "${NM_FIFO}"
  nmcli device monitor "${STA_IF}" > "${NM_FIFO}" 2>/dev/null &
  NM_MON_PID=$!
  exec 3<>"${NM_FIFO}"

  local offline_ticks=0
  local line=""
  local rc=0

  while [[ "${SHUTTING_DOWN}" -eq 0 ]]; do
    rc=0
    line=""
    IFS= read -r -t 2 -u 3 line || rc=$?

    if is_switch_held; then
      offline_ticks=0
      if [[ "${AP_ACTIVE}" -eq 1 ]]; then
        stop_ap
        write_runtime_state
      fi
      continue
    fi

    if [[ "${rc}" -eq 0 && -n "${line}" ]]; then
      case "${line}" in
        *"connecting"*|*"configuring"*|*"deactivating"*|*"unavailable"*)
          offline_ticks=0
          if [[ "${AP_ACTIVE}" -eq 1 ]]; then
            log "Upstream event (${line}) -> immediately releasing ${ACTIVE_AP_IF} for clean handoff!"
            stop_ap
            write_runtime_state
          fi
          ;;
        *"disconnected"*)
          if [[ "${AP_ACTIVE}" -eq 1 ]]; then
            log "Upstream disconnected -> releasing ${ACTIVE_AP_IF} so NetworkManager scans freely..."
            stop_ap
            write_runtime_state
          fi
          offline_ticks=1
          ;;
        *": connected"*)
          offline_ticks=0
          sleep 1
          if ! is_switch_held && [[ "$(get_sta_state)" == "connected" ]]; then
            sync_shared_ap_to_upstream
          fi
          ;;
      esac
    else
      # Periodic 2s health / roam / post-switch rebind check
      local cur_state
      cur_state="$(get_sta_state)"
      if [[ "${cur_state}" == "connected" ]]; then
        offline_ticks=0
        sync_shared_ap_to_upstream
      elif [[ "${cur_state}" == "disconnected" ]]; then
        if [[ "${AP_ACTIVE}" -eq 0 ]]; then
          offline_ticks=$(( offline_ticks + 1 ))
          if (( offline_ticks >= 5 )); then
            offline_ticks=0
            log "Upstream offline >10s -> booting standalone offline island on Channel 6!"
            start_ap 6 2437 || true
          fi
        fi
      else
        offline_ticks=0
      fi
    fi
  done
}

# ==============================================================================
# BLOCK 4: AUDIT PROCEDURES BLOCK
# ( Root Privilege, Binary Presence & Upstream Interface Pre-Flight Checks )
# ==============================================================================
# Engine:        FaithNet Adaptive Networking Substrate
# Stratum:       Layer 4: Proving Ground (Field Kit 01 — Linux Daemon)
# Module:        playtest.field_kit.01.linux.ap_daemon
# Mission:       Verify dependencies before mutating any kernel netlink state.
# ==============================================================================

preflight_daemon() {
  if [[ "$(id -u)" -ne 0 ]]; then
    log "PREFLIGHT FAIL: faithnet-ap-daemon.sh must run as root (systemd unit)." >&2
    return 1
  fi
  for bin in iw ip hostapd dnsmasq iptables nmcli; do
    if ! command -v "${bin}" >/dev/null 2>&1; then
      log "PREFLIGHT FAIL: required tool missing: ${bin}" >&2
      return 1
    fi
  done
  return 0
}

# ==============================================================================
# BLOCK 5: CLOSING BLOCK
# ( Daemon Bootstrap, Radio Mode Selection & Continuous Sovereign Guard )
# ==============================================================================
# Engine:        FaithNet Adaptive Networking Substrate
# Stratum:       Layer 4: Proving Ground (Field Kit 01 — Linux Daemon)
# Module:        playtest.field_kit.01.linux.ap_daemon
# Mission:       Boot WAN guard, select dedicated vs. shared radio, and supervise.
# ==============================================================================

main() {
  preflight_daemon
  mkdir -p "${RUN_DIR}"
  rm -f "${SWITCH_HOLD_FILE}"
  ORIG_IP_FORWARD="$(cat /proc/sys/net/ipv4/ip_forward 2>/dev/null || echo 0)"
  echo "$$" > "${DAEMON_PID_FILE}"
  echo "${EGRESS_MODE}" > "${RUN_DIR}/egress_mode"
  log "FaithNet Persistent Sovereign Micro-ISP Daemon starting (PID=$$, WAN=${STA_IF}, Egress=${EGRESS_MODE}, Country=${COUNTRY_CODE})..."

  apply_wan_firewall

  if detect_dedicated_adapter; then
    run_dedicated_radio_loop
    return 0
  fi

  local init_state="" init_ch="" init_freq=""
  for _ in $(seq 1 15); do
    init_state="$(get_sta_state)"
    init_ch="$(get_sta_channel)"
    init_freq="$(get_sta_freq)"
    if [[ "${init_state}" == "connected" && -n "${init_ch}" ]]; then
      break
    fi
    sleep 1
  done

  if [[ "${init_state}" == "connected" && -n "${init_ch}" ]]; then
    start_ap "${init_ch}" "${init_freq}" || true
  else
    log "No upstream Wi-Fi connected on boot -> starting standalone offline island on Channel 6!"
    start_ap 6 2437 || true
  fi

  run_shared_event_monitor_loop
}

main "$@"

# End of tests/playtester/field_kits/01_two_device_room_well/faithnet-ap-daemon.sh
