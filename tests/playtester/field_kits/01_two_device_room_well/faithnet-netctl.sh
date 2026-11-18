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
  if [[ -z "${target_arg}" ]]; then
    echo "Usage: sudo faithnet-netctl switch \"<Wi-Fi Profile Name>\"" >&2
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

# ==============================================================================
# BLOCK 4: AUDIT PROCEDURES BLOCK
# ( Usage & Subcommand Validation )
# ==============================================================================

usage() {
  cat <<'EOF'
Usage: faithnet-netctl <command>

Commands:
  status            Show Sovereign Host Profile status, Split-DNS, WAN guard & clients
  wifi              Show active Wi-Fi uplink, channel, and switchable saved profiles
  switch <SSID>     Cleanly switch Wi-Fi/hotspot while FaithNet stays live (auto-rollback)
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
  local cmd="${1:-status}"
  shift || true
  case "${cmd}" in
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
