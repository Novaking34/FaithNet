#!/usr/bin/env bash
# ==============================================================================
# BLOCK 1: METADATA BLOCK
# ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
# ==============================================================================
# @project       FaithNet
# @file          tests/playtester/field_kits/01_two_device_room_well/faithnet-netctl.sh
# @id            playtest.field_kit.01.linux.netctl
# @layer         Layer 4: Proving Ground (Field Kit 01 — Sovereign Host Control)
# @archetype     field_cli | operational
# @version       0.2.0-dev
# @interface_rev Rev 5
# @date          2026-11-17
# @updated       2026-11-17
# @author        CreativeWorkzStudio LLC / ProfessorSeanEX
# @repository    https://github.com/Novaking34/FaithNet
# @system        FaithNet Sovereign Host Profile Operator Control CLI
# @role          Inspect Radio Mode, Switch Upstream Wi-Fi Cleanly, Toggle Egress & Trust CA
# @maturity      Rev 5 Baseline | Wave 21 Seamless Upstream Switch & Multi-Channel Support
# @chassis       zero-brick-cli | USR1/USR2 AP handoff + scoped iptables + resolvectl + CA trust
# @scale         O(1) instant operator toggle & zero-queue-flush Wi-Fi/hotspot switch
# @seams         in: /run/faithnet-ap, nmcli, iptables, resolvectl, iw | exports: faithnet-netctl
# @invariants    1. Universal 5-Block Lifecycle Protocol (Rev 5).
#                2. `mode bridge` always restores outbound commercial WAN immediately.
#                3. `switch` yields ap0 BEFORE nmcli touches wlp3s0 and auto-rolls back
#                   to the prior working connection on failure (Zero-Brick Guarantee).
#                4. `adapters` and `wifi` are 100% read-only; `trust-host-ca` is idempotent.
#
# PURPOSE & ARCHITECTURAL ROLE:
#   Operator control plane for Laptop #1's Wave 15 Sovereign Host Profile.
#   Lets Seanje inspect live radio mode (`SHARED_RADIO` vs `DEDICATED_RADIO`),
#   toggle commercial Rung 4 WAN egress (`bridge` vs `island`), scan for a
#   dedicated 2nd USB Wi-Fi adapter, and install FaithNet's Sovereign Root CA
#   into the host OS trust store so `https://faithnet.lan:8077` opens natively
#   with a green padlock on this PC itself.
# ==============================================================================
set -euo pipefail

# ==============================================================================
# BLOCK 2: OPENING BLOCK
# ( Operational Anchors & Runtime Paths )
# ==============================================================================
# Engine:        FaithNet Adaptive Networking Substrate
# Stratum:       Layer 4: Proving Ground (Field Kit 01 — Linux)
# Module:        playtest.field_kit.01.linux.netctl
# Mission:       Define runtime state files and default interface anchors.
# ==============================================================================

readonly RUN_DIR="/run/faithnet-ap"
readonly STATE_FILE="${RUN_DIR}/state.env"
readonly MODE_FILE="${RUN_DIR}/egress_mode"
readonly SWITCH_HOLD_FILE="${RUN_DIR}/switch.hold"
readonly DAEMON_PID_FILE="${RUN_DIR}/daemon.pid"
readonly CONF_FILE="/etc/faithnet-ap.conf"
readonly WAN_GUARD_CHAIN="FAITHNET_WAN_GUARD"
readonly WAN_EGRESS_CHAIN="FAITHNET_WAN_EGRESS"

STA_IF="wlp3s0"
ACTIVE_AP_IF="ap0"
AP_SUBNET="10.77.77.0/24"
RADIO_MODE="UNKNOWN"
AP_IP="10.77.77.1"
SSID="FaithNet-HomeRoom"
CURRENT_CHAN="?"

load_state() {
  if [[ -f "${STATE_FILE}" ]]; then
    # shellcheck disable=SC1090
    source "${STATE_FILE}"
  fi
}

# ==============================================================================
# BLOCK 3: BODY BLOCK
# ( Status Dashboard, Bridge/Island Toggle, USB Adapter Scanner & Host CA Trust )
# ==============================================================================
# Engine:        FaithNet Adaptive Networking Substrate
# Stratum:       Layer 4: Proving Ground (Field Kit 01 — Linux)
# Module:        playtest.field_kit.01.linux.netctl
# Mission:       Implement status, mode, adapters, and trust-host-ca subcommands.
# ==============================================================================

cmd_status() {
  load_state
  local egress="bridge"
  [[ -f "${MODE_FILE}" ]] && egress="$(cat "${MODE_FILE}")"

  echo "+------------------------------------------------------------+"
  echo "| FAITHNET SOVEREIGN HOST PROFILE STATUS (WAVE 15)           |"
  echo "+------------------------------------------------------------+"
  printf "| Radio Mode:  %-45s |\n" "${RADIO_MODE}"
  printf "| HomeRoom AP: %-45s |\n" "${SSID} on ${ACTIVE_AP_IF} (${AP_IP}/24, Ch ${CURRENT_CHAN})"
  printf "| WAN Uplink:  %-45s |\n" "${STA_IF} (Egress Mode: ${egress})"
  echo "+------------------------------------------------------------+"

  echo ""
  echo "[1] Split-DNS Check (faithnet.lan on THIS PC):"
  if command -v resolvectl >/dev/null 2>&1; then
    resolvectl query faithnet.lan 2>&1 | head -n 3 || true
  fi

  echo ""
  echo "[2] Dual-Stack Rung 4 WAN Firewall (${WAN_GUARD_CHAIN} on ${STA_IF}):"
  if sudo -n iptables -L "${WAN_GUARD_CHAIN}" -n --line-numbers >/dev/null 2>&1; then
    echo "  -- IPv4 --"
    sudo iptables -L "${WAN_GUARD_CHAIN}" -n --line-numbers
    if command -v ip6tables >/dev/null 2>&1 && sudo -n ip6tables -L "${WAN_GUARD_CHAIN}" -n --line-numbers >/dev/null 2>&1; then
      echo "  -- IPv6 --"
      sudo ip6tables -L "${WAN_GUARD_CHAIN}" -n --line-numbers
    fi
  else
    echo "    (Run with sudo to inspect live iptables/ip6tables chains)"
  fi

  echo ""
  echo "[3] Connected FaithNet-HomeRoom Clients (${ACTIVE_AP_IF}):"
  iw dev "${ACTIVE_AP_IF}" station dump 2>/dev/null | awk '/Station / {print "    Peer MAC:", $2} /signal:/ {sub(/^[ \t]+/, ""); print "     ", $0}' || echo "    (No clients attached right now)"
}

cmd_mode() {
  local target="${1:-}"
  if [[ "${target}" != "bridge" && "${target}" != "island" ]]; then
    echo "Usage: sudo faithnet-netctl mode <bridge|island>" >&2
    return 1
  fi
  if [[ "$(id -u)" -ne 0 ]]; then
    echo "Please run with sudo: sudo faithnet-netctl mode ${target}" >&2
    return 1
  fi
  load_state
  mkdir -p "${RUN_DIR}"
  echo "${target}" > "${MODE_FILE}"

  # Persist preference in /etc/faithnet-ap.conf cleanly
  if [[ -f "${CONF_FILE}" ]]; then
    grep -v '^EGRESS_MODE=' "${CONF_FILE}" > "${CONF_FILE}.tmp" || true
    mv "${CONF_FILE}.tmp" "${CONF_FILE}"
  fi
  echo "EGRESS_MODE=${target}" >> "${CONF_FILE}"

  # Apply immediately without restarting hostapd (zero drop for HomeRoom devices!)
  iptables -D OUTPUT -o "${STA_IF}" -j "${WAN_EGRESS_CHAIN}" 2>/dev/null || true
  iptables -F "${WAN_EGRESS_CHAIN}" 2>/dev/null || true
  iptables -X "${WAN_EGRESS_CHAIN}" 2>/dev/null || true

  if command -v ip6tables >/dev/null 2>&1; then
    ip6tables -D OUTPUT -o "${STA_IF}" -j "${WAN_EGRESS_CHAIN}" 2>/dev/null || true
    ip6tables -F "${WAN_EGRESS_CHAIN}" 2>/dev/null || true
    ip6tables -X "${WAN_EGRESS_CHAIN}" 2>/dev/null || true
  fi

  if [[ "${target}" == "island" ]]; then
    iptables -t nat -D POSTROUTING -s "${AP_SUBNET}" -o "${STA_IF}" -j MASQUERADE 2>/dev/null || true
    iptables -D FORWARD -i "${ACTIVE_AP_IF}" -o "${STA_IF}" -j ACCEPT 2>/dev/null || true
    iptables -D FORWARD -i "${STA_IF}" -o "${ACTIVE_AP_IF}" -m state --state RELATED,ESTABLISHED -j ACCEPT 2>/dev/null || true

    iptables -N "${WAN_EGRESS_CHAIN}"
    iptables -A "${WAN_EGRESS_CHAIN}" -p udp --sport 68 --dport 67 -j ACCEPT
    iptables -A "${WAN_EGRESS_CHAIN}" -j REJECT --reject-with icmp-net-prohibited
    iptables -I OUTPUT 1 -o "${STA_IF}" -j "${WAN_EGRESS_CHAIN}"

    if command -v ip6tables >/dev/null 2>&1; then
      ip6tables -N "${WAN_EGRESS_CHAIN}"
      ip6tables -A "${WAN_EGRESS_CHAIN}" -p ipv6-icmp -j ACCEPT
      ip6tables -A "${WAN_EGRESS_CHAIN}" -p udp --sport 546 --dport 547 -j ACCEPT
      ip6tables -A "${WAN_EGRESS_CHAIN}" -j REJECT --reject-with icmp6-adm-prohibited
      ip6tables -I OUTPUT 1 -o "${STA_IF}" -j "${WAN_EGRESS_CHAIN}"
    fi
    echo "[ISLAND MODE ACTIVE] Dual-stack IPv4+IPv6 commercial WAN (${STA_IF}) severed. FaithNet-HomeRoom (${ACTIVE_AP_IF}) stays 100% live!"
    echo "Restore internet anytime with: sudo faithnet-netctl mode bridge"
  else
    sysctl -w net.ipv4.ip_forward=1 >/dev/null
    iptables -t nat -C POSTROUTING -s "${AP_SUBNET}" -o "${STA_IF}" -j MASQUERADE 2>/dev/null || \
      iptables -t nat -A POSTROUTING -s "${AP_SUBNET}" -o "${STA_IF}" -j MASQUERADE
    iptables -C FORWARD -i "${ACTIVE_AP_IF}" -o "${STA_IF}" -j ACCEPT 2>/dev/null || \
      iptables -A FORWARD -i "${ACTIVE_AP_IF}" -o "${STA_IF}" -j ACCEPT
    iptables -C FORWARD -i "${STA_IF}" -o "${ACTIVE_AP_IF}" -m state --state RELATED,ESTABLISHED -j ACCEPT 2>/dev/null || \
      iptables -A FORWARD -i "${STA_IF}" -o "${ACTIVE_AP_IF}" -m state --state RELATED,ESTABLISHED -j ACCEPT
    echo "[BRIDGE MODE ACTIVE] Dual-stack inbound ${STA_IF} firewalled; outbound internet & HomeRoom NAT restored!"
  fi
}

cmd_adapters() {
  echo "=== FaithNet Physical Wi-Fi Radio Scanner (Dedicated USB Adapter Ready) ==="
  iw dev 2>/dev/null | awk '
    /^phy#/ { phy=$1; print "\n[" phy "]" }
    /Interface / { print "  Interface:", $2 }
    /type / { print "  Type:     ", $2 }
    /channel / { print "  Channel:  ", $2, $3, $4 }
  '
  local phy_count
  phy_count="$(iw dev 2>/dev/null | grep -c '^phy#' || true)"
  echo ""
  if (( phy_count >= 2 )); then
    echo "[READY] ${phy_count} physical Wi-Fi radios detected!"
    echo "        Restart faithnet-ap (`sudo systemctl restart faithnet-ap`) to auto-bind"
    echo "        FaithNet-HomeRoom onto the dedicated 2nd radio (DEDICATED_RADIO mode)!"
  else
    echo "[INFO] 1 physical Wi-Fi radio detected (running SHARED_RADIO STA+AP mode)."
    echo "       When you plug in any Linux-compatible USB Wi-Fi adapter supporting AP mode"
    echo "       (e.g., MediaTek MT7612U / MT7921AU or Atheros AR9271), just run:"
    echo "         sudo systemctl restart faithnet-ap"
    echo "       and FaithNet will automatically claim the dedicated radio with zero config!"
  fi
}

cmd_trust_host_ca() {
  local user_home="${HOME}"
  if [[ -n "${SUDO_USER:-}" ]]; then
    user_home="$(getent passwd "${SUDO_USER}" | cut -d: -f6)"
  fi
  local ca_src="${user_home}/.local/share/faithnet/ca/ca.crt"
  local ca_dst="/usr/local/share/ca-certificates/faithnet-root-ca.crt"
  if [[ ! -f "${ca_src}" ]]; then
    echo "[ERROR] Sovereign Root CA not found at ${ca_src}. Start faithnet-home.service first." >&2
    return 1
  fi
  echo "[1/2] Installing FaithNet Sovereign Root CA into host Ubuntu trust store..."
  sudo install -m 0644 "${ca_src}" "${ca_dst}"
  echo "[2/2] Refreshing system CA certificates (update-ca-certificates)..."
  sudo update-ca-certificates
  echo "[PASS] Host PC now natively trusts https://faithnet.lan:8077/ with zero --cacert flag!"
}

resolve_wifi_profile() {
  local want="$1"
  local exact=""
  local trimmed_match=""
  local want_trim
  want_trim="$(printf '%s' "${want}" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')"

  while IFS=: read -r name ctype; do
    [[ "${ctype}" == "802-11-wireless" ]] || continue
    if [[ "${name}" == "${want}" ]]; then
      exact="${name}"
      break
    fi
    local name_trim
    name_trim="$(printf '%s' "${name}" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')"
    if [[ "${name_trim,,}" == "${want_trim,,}" ]]; then
      trimmed_match="${name}"
    fi
  done < <(nmcli -t -f NAME,TYPE connection show 2>/dev/null)

  if [[ -n "${exact}" ]]; then
    printf '%s' "${exact}"
    return 0
  fi
  if [[ -n "${trimmed_match}" ]]; then
    printf '%s' "${trimmed_match}"
    return 0
  fi
  return 1
}

cmd_wifi() {
  load_state
  local cur_conn cur_chan cur_ip
  cur_conn="$(nmcli -t -f DEVICE,CONNECTION dev status 2>/dev/null | awk -F: -v d="${STA_IF}" '$1==d {print $2; exit}')"
  cur_chan="$(iw dev "${STA_IF}" info 2>/dev/null | awk '/channel / {print $2, $3; exit}')"
  cur_ip="$(ip -4 -o addr show dev "${STA_IF}" 2>/dev/null | awk '{print $4; exit}')"

  echo "=== FaithNet Upstream Wi-Fi & Switchable Profiles (${STA_IF}) ==="
  echo "  Active Uplink: ${cur_conn:-disconnected}  (Channel: ${cur_chan:-none}, IP: ${cur_ip:-none})"
  echo "  HomeRoom AP:   ${SSID} on ${ACTIVE_AP_IF} (Radio: ${RADIO_MODE}, Ch: ${CURRENT_CHAN:-idle})"
  echo ""
  echo "[Saved Wi-Fi Profiles]"
  nmcli -t -f NAME,TYPE,DEVICE connection show 2>/dev/null | awk -F: '$2=="802-11-wireless" {
    active = ($3 != "" ? " [ACTIVE on " $3 "]" : "")
    printf "  - \"%s\"%s\n", $1, active
  }'
  echo ""
  echo "Switch cleanly while FaithNet stays live:"
  echo "  sudo faithnet-netctl switch \"TheTribe\""
  echo "  sudo faithnet-netctl switch \"Shein iPhone\""
}

cmd_switch() {
  local target_arg="${1:-}"
  local force_switch=0
  if [[ "${2:-}" == "--force" ]]; then
    force_switch=1
  fi
  if [[ -z "${target_arg}" ]]; then
    echo "Usage: sudo faithnet-netctl switch \"<Wi-Fi Profile Name>\" [--force]" >&2
    echo "Run 'faithnet-netctl wifi' to see saved profiles." >&2
    return 1
  fi
  if [[ "$(id -u)" -ne 0 ]]; then
    echo "Please run with sudo: sudo faithnet-netctl switch \"${target_arg}\"" >&2
    return 1
  fi

  load_state
  local target_profile
  if ! target_profile="$(resolve_wifi_profile "${target_arg}")"; then
    echo "[ERROR] No saved Wi-Fi profile matching '${target_arg}'." >&2
    echo "        Run 'faithnet-netctl wifi' to list available saved profiles." >&2
    return 1
  fi

  local target_ssid
  target_ssid="$(nmcli -g 802-11-wireless.ssid connection show id "${target_profile}" 2>/dev/null || echo "${target_profile}")"
  if [[ "${force_switch}" -eq 0 ]]; then
    echo "[0/4] Checking if '${target_ssid}' is actually in range before dropping current Wi-Fi..."
    nmcli device wifi rescan ifname "${STA_IF}" >/dev/null 2>&1 || true
    sleep 1
    if ! nmcli -t -f SSID device wifi list ifname "${STA_IF}" 2>/dev/null | grep -Fxq "${target_ssid}"; then
      echo "[HOLD] '${target_ssid}' is NOT showing up on Wi-Fi right now." >&2
      echo "       Keeping your current connection untouched so you don't lose internet." >&2
      echo "       Tip: Check if the router is powered on/in range, or pass --force if it is hidden." >&2
      return 1
    fi
  fi

  local prev_conn
  prev_conn="$(nmcli -t -f DEVICE,CONNECTION dev status 2>/dev/null | awk -F: -v d="${STA_IF}" '$1==d {print $2; exit}')"
  local daemon_pid=""
  [[ -f "${DAEMON_PID_FILE}" ]] && daemon_pid="$(cat "${DAEMON_PID_FILE}" 2>/dev/null || true)"

  echo "[1/4] Yielding ${ACTIVE_AP_IF} cleanly so ${STA_IF} switches with zero radio conflict..."
  mkdir -p "${RUN_DIR}"
  echo "switching:${target_profile}" > "${SWITCH_HOLD_FILE}"
  if [[ -n "${daemon_pid}" ]] && kill -0 "${daemon_pid}" 2>/dev/null; then
    kill -USR1 "${daemon_pid}" 2>/dev/null || true
    for _ in $(seq 1 20); do
      if ! ip link show "${ACTIVE_AP_IF}" >/dev/null 2>&1; then
        break
      fi
      sleep 0.2
    done
  fi
  sleep 0.5

  echo "[2/4] Switching ${STA_IF} cleanly -> '${target_profile}' (FaithNet services stay live)..."
  nmcli device disconnect "${STA_IF}" >/dev/null 2>&1 || true
  sleep 0.5

  local switched=0
  if nmcli --wait 30 connection up id "${target_profile}"; then
    switched=1
  else
    echo "[WARN] Failed to activate '${target_profile}'!" >&2
    echo "       Run 'faithnet-netctl logs 5' for a plain-English readout of why it failed." >&2
    if [[ -n "${prev_conn}" && "${prev_conn}" != "--" && "${prev_conn}" != "${target_profile}" ]]; then
      echo "[ROLLBACK] Restoring previous working uplink '${prev_conn}' (Zero-Brick Guarantee)..." >&2
      nmcli --wait 25 connection up id "${prev_conn}" >/dev/null 2>&1 || true
    fi
  fi

  echo "[3/4] Releasing switch hold -> re-syncing '${SSID}' onto active channel..."
  rm -f "${SWITCH_HOLD_FILE}"
  if [[ -n "${daemon_pid}" ]] && kill -0 "${daemon_pid}" 2>/dev/null; then
    kill -USR2 "${daemon_pid}" 2>/dev/null || true
    for _ in $(seq 1 32); do
      load_state
      if [[ "${AP_ACTIVE:-0}" -eq 1 ]] && ip link show "${ACTIVE_AP_IF}" >/dev/null 2>&1; then
        break
      fi
      sleep 0.25
    done
  fi

  echo "[4/4] Verifying final uplink & HomeRoom state..."
  local final_conn final_chan final_ip
  final_conn="$(nmcli -t -f DEVICE,CONNECTION dev status 2>/dev/null | awk -F: -v d="${STA_IF}" '$1==d {print $2; exit}')"
  final_chan="$(iw dev "${STA_IF}" info 2>/dev/null | awk '/channel / {print $2, $3; exit}')"
  final_ip="$(ip -4 -o addr show dev "${STA_IF}" 2>/dev/null | awk '{print $4; exit}')"
  load_state
  echo "  Active Uplink: ${final_conn:-disconnected} (Channel: ${final_chan:-none}, IP: ${final_ip:-none})"
  echo "  HomeRoom AP:   ${SSID} on ${ACTIVE_AP_IF} (Ch: ${CURRENT_CHAN:-pending}, Active: ${AP_ACTIVE:-0})"

  if [[ "${switched}" -eq 1 ]]; then
    echo "[PASS] Switched cleanly to '${target_profile}' while FaithNet stayed live!"
    return 0
  fi
  return 1
}

describe_signal_dbm() {
  local dbm="${1:-}"
  if [[ -z "${dbm}" || ! "${dbm}" =~ ^-?[0-9]+$ ]]; then
    echo "Unknown"
    return
  fi
  if (( dbm >= -50 )); then
    echo "Excellent (${dbm} dBm — right next to you)"
  elif (( dbm >= -65 )); then
    echo "Good (${dbm} dBm — solid connection)"
  elif (( dbm >= -75 )); then
    echo "Fair (${dbm} dBm — usable, some walls/distance)"
  else
    echo "Weak (${dbm} dBm — far away; may drop or time out)"
  fi
}

cmd_health() {
  load_state
  local cur_state cur_conn cur_ip cur_gw cur_chan cur_freq band_label sig_dbm sig_text
  cur_state="$(nmcli -t -f DEVICE,STATE dev status 2>/dev/null | awk -F: -v d="${STA_IF}" '$1==d {print $2; exit}')"
  cur_conn="$(nmcli -t -f DEVICE,CONNECTION dev status 2>/dev/null | awk -F: -v d="${STA_IF}" '$1==d {print $2; exit}')"
  cur_ip="$(ip -4 -o addr show dev "${STA_IF}" 2>/dev/null | awk '{print $4; exit}')"
  cur_gw="$(ip -4 route show default dev "${STA_IF}" 2>/dev/null | awk '{print $3; exit}')"
  cur_chan="$(iw dev "${STA_IF}" info 2>/dev/null | awk '/channel / {print $2; exit}')"
  cur_freq="$(iw dev "${STA_IF}" info 2>/dev/null | awk '/channel / {gsub(/[()]/, "", $3); print $3; exit}')"
  sig_dbm="$(iw dev "${STA_IF}" link 2>/dev/null | awk '/signal:/ {print $2; exit}')"
  sig_text="$(describe_signal_dbm "${sig_dbm}")"

  band_label="Unknown"
  if [[ "${cur_freq}" =~ ^[0-9]+$ ]]; then
    if (( cur_freq < 3000 )); then band_label="2.4 GHz (longer range)";
    elif (( cur_freq < 5900 )); then band_label="5 GHz (faster)";
    else band_label="6 GHz"; fi
  fi

  local inet_tag="[OFFLINE]"
  local inet_msg="No internet response right now"
  if ping -c 1 -W 2 1.1.1.1 >/dev/null 2>&1; then
    inet_tag="[  OK   ]"
    inet_msg="Online (internet responding normally)"
  elif [[ -n "${cur_ip}" ]]; then
    inet_tag="[ WARN  ]"
    inet_msg="Connected to router (${cur_ip}), but no internet beyond it"
  fi

  local ap_svc="stopped"
  systemctl is-active --quiet faithnet-ap.service 2>/dev/null && ap_svc="running"
  local home_svc="stopped"
  systemctl --user is-active --quiet faithnet-home.service 2>/dev/null && home_svc="running"
  local peer_count="0"
  if ip link show "${ACTIVE_AP_IF}" >/dev/null 2>&1; then
    peer_count="$(iw dev "${ACTIVE_AP_IF}" station dump 2>/dev/null | grep -c '^Station ' || true)"
  fi

  local rtw_state="Default (sleep enabled)"
  local l1_val ps_val
  l1_val="$(cat /sys/module/rtw89_pci/parameters/disable_aspm_l1 2>/dev/null || echo "?")"
  ps_val="$(cat /sys/module/rtw89_core/parameters/disable_ps_mode 2>/dev/null || echo "?")"
  if [[ "${l1_val}" == "Y" && "${ps_val}" == "Y" ]]; then
    rtw_state="Tuned & Active (ASPM L1 + deep sleep disabled)"
  elif [[ -f /etc/modprobe.d/99-faithnet-rtw89-stability.conf ]]; then
    rtw_state="Installed (will activate cleanly on next reboot)"
  fi

  local wifi_tag="[  OK   ]"
  [[ "${cur_state}" != "connected" ]] && wifi_tag="[ WARN  ]"
  local ap_tag="[ INFO  ]"
  [[ "${ap_svc}" == "running" ]] && ap_tag="[  OK   ]"

  echo "========================================================================"
  echo "  FAITHNET PLAIN-ENGLISH NETWORK HEALTH CHECK"
  echo "========================================================================"
  echo "  ${wifi_tag} Wi-Fi Network:   ${cur_conn:-Not connected} (${cur_state:-unknown})"
  if [[ "${cur_state}" == "connected" ]]; then
    echo "            Band & Channel:  Channel ${cur_chan:-?} — ${band_label}"
    echo "            Signal Quality:  ${sig_text}"
    echo "            IP & Router:     ${cur_ip:-none} (Router: ${cur_gw:-none})"
  fi
  echo "  ${inet_tag} Internet:        ${inet_msg}"
  echo "  ${ap_tag} FaithNet AP:     ${ap_svc} (${SSID}, ${peer_count} device(s) joined)"
  echo "  [ INFO  ] FaithNet Node:   ${home_svc} (Port 7777 / The Well :8077)"
  echo "  [ INFO  ] Wi-Fi Chip Tune: ${rtw_state}"
  echo "------------------------------------------------------------------------"

  if [[ "${cur_state}" != "connected" ]]; then
    echo "  Next Step: Run 'faithnet-netctl logs' to see why Wi-Fi disconnected,"
    echo "             or 'faithnet-netctl wifi' to see available networks."
  elif [[ "${inet_tag}" == "[ WARN  ]" ]]; then
    echo "  Next Step: Your laptop reached the router, but the router has no internet"
    echo "             (or FaithNet is in 'island' mode — check 'faithnet-netctl status')."
  elif [[ "${ap_svc}" == "stopped" ]]; then
    echo "  Next Step: Wi-Fi is running stock Ubuntu (FaithNet AP is paused)."
    echo "             Turn it back on anytime with: sudo systemctl start faithnet-ap"
  else
    echo "  Everything looks healthy. Run 'faithnet-netctl logs' for recent history."
  fi
  echo "========================================================================"
}

translate_network_logs() {
  awk '
    function flush_dup() {
      if (last_msg != "") {
        if (dup_cnt > 1) {
          printf "  %-8s %s  %s (x%d)\n", last_tag, last_ts, last_msg, dup_cnt
        } else {
          printf "  %-8s %s  %s\n", last_tag, last_ts, last_msg
        }
        fflush()
        last_msg = ""
        last_tag = ""
        dup_cnt = 0
      }
    }
    function emit(tag, ts, msg) {
      if (msg == last_msg && tag == last_tag) {
        dup_cnt++
        last_ts = ts
        return
      }
      flush_dup()
      last_tag = tag
      last_ts = ts
      last_msg = msg
      dup_cnt = 1
    }
    {
      ts = $3
      line = $0

      if (line ~ /timed out to flush queues/) {
        emit("[RADIO]", ts, "Wi-Fi chip had a queue hiccup while changing state/channels (rtw89 flush timeout)")
        next
      }
      if (line ~ /Activation: starting connection /) {
        name = line
        sub(/.*Activation: starting connection \x27/, "", name)
        sub(/\x27.*/, "", name)
        emit("[TRY]", ts, "Trying to connect to Wi-Fi \x27" name "\x27...")
        next
      }
      if (line ~ /Connected to wireless network "/) {
        name = line
        sub(/.*Connected to wireless network "/, "", name)
        sub(/".*/, "", name)
        emit("[WAIT]", ts, "Joined \x27" name "\x27 radio! Waiting for router to give us an IP address (DHCP)...")
        next
      }
      if (line ~ /dhcp4 \(wlp3s0\): state changed new lease, address=/) {
        ip = line
        sub(/.*address=/, "", ip)
        sub(/,.*/, "", ip)
        if (line !~ /acd pending/) {
          emit("[OK]", ts, "Router gave us IP address " ip " (connection ready!)")
        }
        next
      }
      if (line ~ /Activation: failed for connection /) {
        name = line
        sub(/.*Activation: failed for connection \x27/, "", name)
        sub(/\x27.*/, "", name)
        emit("[FAIL]", ts, "Could not finish connecting to \x27" name "\x27 (router did not respond or timed out)")
        next
      }
      if (line ~ /reason \x27ip-config-unavailable\x27/) {
        emit("[FAIL]", ts, "Router never handed out an IP address (DHCP timed out — check router or signal)")
        next
      }
      if (line ~ /reason \x27no-secrets\x27/) {
        emit("[FAIL]", ts, "Wi-Fi password was missing or rejected")
        next
      }
      if (line ~ /authentication with .* timed out/) {
        emit("[WARN]", ts, "Router did not answer our login knock (weak signal or router busy)")
        next
      }
      if (line ~ /association with .* timed out/) {
        emit("[WARN]", ts, "Router heard us but dropped the join handshake (signal dropped or router busy)")
        next
      }
      if (line ~ /Limiting TX power to 0 /) {
        emit("[INFO]", ts, "Hotspot asked laptop to use low transmit power (0 dBm — normal when phone is close)")
        next
      }
      if (line ~ /SIGUSR1 received|immediately releasing ap0|Yielding ap0 cleanly/) {
        emit("[AP]", ts, "FaithNet-HomeRoom paused cleanly to let Wi-Fi switch networks")
        next
      }
      if (line ~ /LIVE \[/) {
        ch = line
        sub(/.*Channel /, "", ch)
        sub(/\).*/, "", ch)
        emit("[AP]", ts, "FaithNet-HomeRoom is LIVE and broadcasting (Channel " ch ", 10.77.77.1)")
        next
      }
      if (line ~ /DFS\/6GHz\/unsupported|hostapd refused Channel/) {
        emit("[AP]", ts, "Upstream channel does not allow shared AP -> keeping your Wi-Fi untouched")
        next
      }
      if (line ~ /Stopped faithnet-ap\.service/) {
        emit("[AP]", ts, "FaithNet AP service stopped cleanly (Wi-Fi is back to stock Ubuntu)")
        next
      }
      if (line ~ /DHCPACK\(ap0\)/) {
        split(line, parts, " ")
        ip = ""
        for (i = 1; i <= NF; i++) {
          if ($i ~ /^10\.77\.77\./) { ip = $i; break }
        }
        emit("[PEER]", ts, "A device joined FaithNet-HomeRoom (assigned local IP " ip ")")
        next
      }
    }
    END { flush_dup() }
  '
}

cmd_logs() {
  local arg="${1:-30}"
  local j_match=(
    _SYSTEMD_UNIT=NetworkManager.service +
    _SYSTEMD_UNIT=wpa_supplicant.service +
    _SYSTEMD_UNIT=faithnet-ap.service +
    _TRANSPORT=kernel
  )
  if [[ "${arg}" == "--live" || "${arg}" == "-f" ]]; then
    echo "========================================================================"
    echo "  FAITHNET LIVE PLAIN-ENGLISH NETWORK LOGS (Press Ctrl+C to stop)"
    echo "========================================================================"
    journalctl -f -n 20 --no-pager "${j_match[@]}" 2>/dev/null \
      | translate_network_logs
    return 0
  fi

  local mins="30"
  local raw=0
  if [[ "${arg}" == "--raw" ]]; then
    raw=1
    mins="${2:-30}"
  elif [[ "${arg}" =~ ^[0-9]+$ ]]; then
    mins="${arg}"
  fi

  if [[ "${raw}" -eq 1 ]]; then
    journalctl --since "${mins} min ago" --no-pager "${j_match[@]}" 2>/dev/null \
      | grep -iE "wlp3s0|ap0|rtw89|faithnet|dhcp|TheTribe|Shein" | tail -n 60 || true
    return 0
  fi

  echo "========================================================================"
  echo "  FAITHNET PLAIN-ENGLISH NETWORK STORY (Last ${mins} minutes)"
  echo "  Tags: [OK]=Connected  [TRY]=Joining  [WAIT]=Waiting IP  [WARN]=Weak/Retry"
  echo "        [FAIL]=Failed   [RADIO]=Chip   [AP]=HomeRoom      [PEER]=Device"
  echo "========================================================================"
  local out
  out="$(journalctl --since "${mins} min ago" --no-pager "${j_match[@]}" 2>/dev/null \
    | translate_network_logs)"
  if [[ -z "${out}" ]]; then
    echo "  (No major Wi-Fi or FaithNet events in the last ${mins} minutes.)"
    echo "  Tip: Try 'faithnet-netctl logs 120' for the last 2 hours, or 'faithnet-netctl health'."
  else
    echo "${out}"
  fi
  echo "========================================================================"
  echo "  Tip: 'faithnet-netctl logs --live' watches in real time | '--raw' shows full system logs"
}

# ==============================================================================
# BLOCK 4: AUDIT PROCEDURES BLOCK
# ( Usage & Subcommand Validation )
# ==============================================================================

usage() {
  cat <<'EOF'
Usage: faithnet-netctl <command>

Commands:
  health            Plain-English 5-second traffic-light check (Wi-Fi, Signal, IP, Internet, FaithNet)
  logs [mins|--live|--raw]
                    Plain-English timeline of what Wi-Fi & FaithNet have been doing (default: 30m)
  status            Technical Sovereign Host Profile status, Split-DNS, WAN guard & clients
  wifi              Show active Wi-Fi uplink, channel, and switchable saved profiles
  switch <SSID> [--force]
                    Cleanly switch Wi-Fi/hotspot while FaithNet stays live (pre-scan + auto-rollback)
  mode bridge       Daily driver: inbound WAN firewalled, outbound internet + NAT active
  mode island       Air-gapped sovereign test: cut commercial WAN, keep HomeRoom 100% live
  adapters          Scan physical Wi-Fi PHYs for a dedicated 2nd USB Wi-Fi adapter
  trust-host-ca     Install FaithNet Sovereign Root CA into this laptop's system trust store
EOF
}

# ==============================================================================
# BLOCK 5: CLOSING BLOCK
# ( Subcommand Dispatch )
# ==============================================================================

main() {
  local cmd="${1:-health}"
  shift || true
  case "${cmd}" in
    health)        cmd_health "$@" ;;
    logs)          cmd_logs "$@" ;;
    status)        cmd_status "$@" ;;
    wifi)          cmd_wifi "$@" ;;
    switch)        cmd_switch "$@" ;;
    mode)          cmd_mode "$@" ;;
    adapters)      cmd_adapters "$@" ;;
    trust-host-ca) cmd_trust_host_ca "$@" ;;
    -h|--help)     usage ;;
    *)             usage; exit 1 ;;
  esac
}

main "$@"

# End of tests/playtester/field_kits/01_two_device_room_well/faithnet-netctl.sh
