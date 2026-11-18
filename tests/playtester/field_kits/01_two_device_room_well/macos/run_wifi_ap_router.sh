#!/usr/bin/env bash
# ==============================================================================
# BLOCK 1: METADATA BLOCK
# ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
# ==============================================================================
# @project       FaithNet
# @file          tests/playtester/field_kits/01_two_device_room_well/macos/run_wifi_ap_router.sh
# @id            playtest.field_kit.01.macos.wifi_ap_router
# @layer         Layer 4: Proving Ground (Field Kit 01 — macOS Internet Sharing)
# @archetype     field_script | operational
# @version       0.1.0-dev
# @interface_rev Rev 5
# @date          2026-11-17
# @updated       2026-11-17
# @author        CreativeWorkzStudio LLC / ProfessorSeanEX
# @repository    https://github.com/CreativeWorkzStudio/FaithNet
# @system        FaithNet Micro-ISP AP Helper & Hardware Check (macOS)
# @role          Inspect macOS Interfaces & Guide Safe Internet Sharing AP
# @maturity      Rev 5 Baseline | Wave 14 5-Block Cross-Platform Parity
# @chassis       read-only-safe | Zero unapproved pf/sysctl mutations
# @scale         Documents Apple single-radio constraint vs. Ethernet/USB + Wi-Fi
# @seams         in: networksetup, ifconfig | exports: bridge100 Well endpoint
# @invariants    1. Universal 5-Block Lifecycle Protocol (Rev 5).
#                2. Honest hardware truth: macOS forbids simultaneous STA+AP on
#                   a single built-in Wi-Fi card (`en0`), requiring Ethernet or
#                   a second USB Wi-Fi adapter for upstream + Wi-Fi AP sharing.
#
# PURPOSE & ARCHITECTURAL ROLE:
#   macOS counterpart to Linux `run_wifi_ap_router.sh`. Probes active macOS
#   network hardware (`networksetup -listallhardwareports`), detects whether
#   macOS Internet Sharing (`bridge100`) is live, and prints exact onboarding
#   instructions for broadcasting `FaithNet-HomeRoom` cleanly on macOS.
#
#   ### For Architects, Operators & Non-Programmers:
#   Explains honestly how to turn a Mac into a FaithNet Wi-Fi router using
#   Apple's built-in Internet Sharing (without hacking kernel drivers).
#
#   ### For Systems Programmers & Engine Architects:
#   Pure read-only inspection script compatible with macOS `/bin/bash` 3.2+.
# ==============================================================================
set -euo pipefail

# ==============================================================================
# BLOCK 2: OPENING BLOCK
# ( Operational Constants )
# ==============================================================================
# Engine:        FaithNet Adaptive Networking Substrate
# Stratum:       Layer 4: Proving Ground (Field Kit 01 — macOS)
# Module:        playtest.field_kit.01.macos.wifi_ap_router
# Mission:       Pin canonical FaithNet SSID and macOS sharing interface name.
# ==============================================================================

SSID="${SSID:-FaithNet-HomeRoom}"
BRIDGE_IF="${BRIDGE_IF:-bridge100}"

# ==============================================================================
# BLOCK 3: BODY BLOCK
# ( macOS Hardware Probe & Internet Sharing Guide )
# ==============================================================================
# Engine:        FaithNet Adaptive Networking Substrate
# Stratum:       Layer 4: Proving Ground (Field Kit 01 — macOS)
# Module:        playtest.field_kit.01.macos.wifi_ap_router
# Mission:       Inspect bridge100 and report exact Apple AP setup steps.
# ==============================================================================

show_macos_ap_status() {
  echo "=== FAITHNET WI-FI AP ROUTER STATUS (MACOS) ==="
  if command -v networksetup >/dev/null 2>&1; then
    networksetup -listallhardwareports | head -n 20
  fi
  echo ""
  if ifconfig "${BRIDGE_IF}" >/dev/null 2>&1; then
    local bip
    bip="$(ipconfig getifaddr "${BRIDGE_IF}" 2>/dev/null || echo '192.168.2.1')"
    echo "[LIVE] macOS Internet Sharing (${BRIDGE_IF}) active at ${bip}!"
    echo "       The Well: https://${bip}:8077/ | Root CA: http://${bip}:8077/faithnet-root-ca.crt"
  else
    echo "[INFO] macOS Hardware Note: Apple restricts simultaneous STA+AP on a single Wi-Fi chip."
    echo "       To broadcast '${SSID}' from macOS while staying online:"
    echo "       1. Connect upstream via Ethernet, USB-C tether, or a second Wi-Fi adapter."
    echo "       2. System Settings -> General -> Sharing -> Internet Sharing:"
    echo "          Share from: <Upstream> | To computers using: Wi-Fi -> Wi-Fi Options:"
    echo "          Network Name: ${SSID} | Security: WPA2/WPA3 Personal."
    echo "       3. Once enabled, ${BRIDGE_IF} comes up (typically 192.168.2.1) and"
    echo "          persistent faithnet-home serves The Well on https://192.168.2.1:8077/!"
  fi
}

# ==============================================================================
# BLOCK 4: AUDIT PROCEDURES BLOCK
# ( Pre-Flight macOS Platform Check )
# ==============================================================================
# Engine:        FaithNet Adaptive Networking Substrate
# Stratum:       Layer 4: Proving Ground (Field Kit 01 — macOS)
# Module:        playtest.field_kit.01.macos.wifi_ap_router
# Mission:       Confirm Darwin environment or print clean dry-run notice.
# ==============================================================================

preflight_macos_ap() {
  if [[ "$(uname -s)" != "Darwin" ]]; then
    echo "[PREFLIGHT INFO] Running on $(uname -s); displaying macOS AP guidance only."
  fi
  return 0
}

# ==============================================================================
# BLOCK 5: CLOSING BLOCK
# ( Lifecycle Entrypoint )
# ==============================================================================
# Engine:        FaithNet Adaptive Networking Substrate
# Stratum:       Layer 4: Proving Ground (Field Kit 01 — macOS)
# Module:        playtest.field_kit.01.macos.wifi_ap_router
# Mission:       Execute pre-flight and display macOS AP router status/guide.
# ==============================================================================

main() {
  preflight_macos_ap
  show_macos_ap_status
}

main "$@"

# End of tests/playtester/field_kits/01_two_device_room_well/macos/run_wifi_ap_router.sh
