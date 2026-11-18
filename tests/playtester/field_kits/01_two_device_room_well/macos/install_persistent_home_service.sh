#!/usr/bin/env bash
# ==============================================================================
# BLOCK 1: METADATA BLOCK
# ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
# ==============================================================================
# @project       FaithNet
# @file          tests/playtester/field_kits/01_two_device_room_well/macos/install_persistent_home_service.sh
# @id            playtest.field_kit.01.macos.install_home_service
# @layer         Layer 4: Proving Ground (Field Kit 01 — macOS LaunchAgent)
# @archetype     field_script | operational
# @version       0.1.0-dev
# @interface_rev Rev 5
# @date          2026-11-17
# @updated       2026-11-17
# @author        CreativeWorkzStudio LLC / ProfessorSeanEX
# @repository    https://github.com/CreativeWorkzStudio/FaithNet
# @system        FaithNet Persistent Home-Anchor Installer (macOS launchd)
# @role          Compile Native faithnet-node & Install ~/Library/LaunchAgents
# @maturity      Rev 5 Baseline | Wave 14 5-Block Cross-Platform Parity
# @chassis       user-scoped | Zero sudo required; clean macOS LaunchAgent plist
# @scale         O(1) deterministic build + idempotent launchctl reload
# @seams         in: cmd/faithnet-node | exports: com.creativeworkzstudio.faithnet.home.plist
# @invariants    1. Universal 5-Block Lifecycle Protocol (Rev 5).
#                2. Zero sudo required (strictly under $HOME/Library).
#                3. Preserves persistent Tier 2 Sovereign Root CA across updates.
#
# PURPOSE & ARCHITECTURAL ROLE:
#   macOS counterpart to Linux `install_persistent_home_service.sh`. Builds the
#   native `faithnet-node` binary and registers a user-level `launchd` agent
#   (`com.creativeworkzstudio.faithnet.home`) so Home-Anchor boots automatically.
#
#   ### For Architects, Operators & Non-Programmers:
#   Turns a Mac into a persistent always-on FaithNet Home-Anchor across logins.
#
#   ### For Systems Programmers & Engine Architects:
#   Writes standard XML LaunchAgent plist (`KeepAlive=true`, `RunAtLoad=true`)
#   and reloads cleanly via `launchctl unload` / `launchctl load`.
# ==============================================================================
set -euo pipefail

# ==============================================================================
# BLOCK 2: OPENING BLOCK
# ( Operational macOS Paths & LaunchAgent Identifiers )
# ==============================================================================
# Engine:        FaithNet Adaptive Networking Substrate
# Stratum:       Layer 4: Proving Ground (Field Kit 01 — macOS)
# Module:        playtest.field_kit.01.macos.install_home_service
# Mission:       Define macOS binary, Application Support CA, and plist paths.
# ==============================================================================

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.." && pwd)"
BIN_DIR="${HOME}/.local/bin"
CA_DIR="${HOME}/Library/Application Support/FaithNet/ca"
AGENT_DIR="${HOME}/Library/LaunchAgents"
PLIST_LABEL="com.creativeworkzstudio.faithnet.home"
PLIST_PATH="${AGENT_DIR}/${PLIST_LABEL}.plist"
LOG_PATH="${HOME}/Library/Logs/faithnet-home.log"

# ==============================================================================
# BLOCK 3: BODY BLOCK
# ( Native Build, LaunchAgent Plist Generation & Activation )
# ==============================================================================
# Engine:        FaithNet Adaptive Networking Substrate
# Stratum:       Layer 4: Proving Ground (Field Kit 01 — macOS)
# Module:        playtest.field_kit.01.macos.install_home_service
# Mission:       Compile native macOS binary and load LaunchAgent cleanly.
# ==============================================================================

build_binary() {
  mkdir -p "${BIN_DIR}" "${CA_DIR}" "${AGENT_DIR}" "$(dirname "${LOG_PATH}")"
  echo "[1/3] Compiling standalone native faithnet-node binary..."
  (cd "${REPO_ROOT}" && go build -o "${BIN_DIR}/faithnet-node" ./cmd/faithnet-node)
}

write_launchagent_plist() {
  echo "[2/3] Writing LaunchAgent plist (${PLIST_PATH})..."
  cat > "${PLIST_PATH}" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>${PLIST_LABEL}</string>
  <key>ProgramArguments</key>
  <array>
    <string>${BIN_DIR}/faithnet-node</string>
    <string>--label=Home-Anchor-Mac</string>
    <string>--udp=0.0.0.0:7777</string>
    <string>--well</string>
    <string>--well-addr=0.0.0.0:8077</string>
    <string>--tls-ca-dir=${CA_DIR}</string>
    <string>--beacon</string>
    <string>--beacon-ssid=FaithNet-HomeRoom</string>
  </array>
  <key>RunAtLoad</key>
  <true/>
  <key>KeepAlive</key>
  <true/>
  <key>StandardOutPath</key>
  <string>${LOG_PATH}</string>
  <key>StandardErrorPath</key>
  <string>${LOG_PATH}</string>
</dict>
</plist>
EOF
}

activate_launchagent() {
  echo "[3/3] Reloading LaunchAgent ${PLIST_LABEL}..."
  launchctl unload "${PLIST_PATH}" 2>/dev/null || true
  launchctl load -w "${PLIST_PATH}"
  echo "[LIVE] LaunchAgent ${PLIST_LABEL} active. Logs: ${LOG_PATH}"
}

# ==============================================================================
# BLOCK 4: AUDIT PROCEDURES BLOCK
# ( Pre-Flight OS & Toolchain Verification )
# ==============================================================================
# Engine:        FaithNet Adaptive Networking Substrate
# Stratum:       Layer 4: Proving Ground (Field Kit 01 — macOS)
# Module:        playtest.field_kit.01.macos.install_home_service
# Mission:       Confirm macOS launchctl, Go toolchain, and go.mod anchor.
# ==============================================================================

preflight_macos_install() {
  if [[ "$(uname -s)" != "Darwin" ]]; then
    echo "[PREFLIGHT FAIL] This installer targets macOS (Darwin); current OS is $(uname -s)." >&2
    return 1
  fi
  for tool in go launchctl; do
    if ! command -v "${tool}" >/dev/null 2>&1; then
      echo "[PREFLIGHT FAIL] Missing required tool: ${tool}" >&2
      return 1
    fi
  done
  if [[ ! -f "${REPO_ROOT}/go.mod" ]]; then
    echo "[PREFLIGHT FAIL] FaithNet go.mod anchor missing at ${REPO_ROOT}." >&2
    return 1
  fi
  return 0
}

# ==============================================================================
# BLOCK 5: CLOSING BLOCK
# ( Lifecycle Entrypoint )
# ==============================================================================
# Engine:        FaithNet Adaptive Networking Substrate
# Stratum:       Layer 4: Proving Ground (Field Kit 01 — macOS)
# Module:        playtest.field_kit.01.macos.install_home_service
# Mission:       Execute pre-flight -> build -> write plist -> activate.
# ==============================================================================

main() {
  preflight_macos_install
  build_binary
  write_launchagent_plist
  activate_launchagent
}

main "$@"

# End of tests/playtester/field_kits/01_two_device_room_well/macos/install_persistent_home_service.sh
