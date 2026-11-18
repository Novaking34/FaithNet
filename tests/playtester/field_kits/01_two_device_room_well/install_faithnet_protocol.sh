#!/usr/bin/env bash
# ==============================================================================
# BLOCK 1: METADATA BLOCK
# ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
# ==============================================================================
# @project       FaithNet
# @file          tests/playtester/field_kits/01_two_device_room_well/install_faithnet_protocol.sh
# @id            playtest.field_kit.01.linux.install_protocol
# @layer         Layer 4: Proving Ground (Field Kit 01 — FNP/1 OS Protocol Setup)
# @archetype     field_script | operational
# @version       0.1.0-dev
# @interface_rev Rev 5
# @date          2026-11-17
# @updated       2026-11-17
# @author        CreativeWorkzStudio LLC / ProfessorSeanEX
# @repository    https://github.com/CreativeWorkzStudio/FaithNet
# @system        FaithNet Official OS Protocol (`faithnet://`) & Port `7777` Setup
# @role          Register x-scheme-handler/faithnet, faithnet-open & /etc/services 7777
# @maturity      Rev 5 Baseline | Wave 15 Official FaithNet Protocol (FNP/1)
# @chassis       user-desktop + optional /etc/services | 100% reversible
# @scale         O(1) idempotent protocol registration
# @seams         in: xdg-mime, curl | exports: ~/.local/bin/faithnet-open, faithnet://
# @invariants    1. Universal 5-Block Lifecycle Protocol (Rev 5).
#                2. Strictly additive to /etc/services (never edits existing ports).
#                3. Validates scheme prefix before dispatching any URI action.
#
# PURPOSE & ARCHITECTURAL ROLE:
#   Installs native OS support for the official FaithNet Protocol (`FNP/1`) on
#   Linux:
#     1. Installs `~/.local/bin/faithnet-open` CLI & desktop handler for
#        `faithnet://faithnet.lan/well`, `faithnet://home-anchor/status`,
#        `faithnet://10.77.77.1/ca`, etc.
#     2. Registers `x-scheme-handler/faithnet` via `xdg-mime` so clicking any
#        `faithnet://` link launches FaithNet natively.
#     3. Registers canonical `faithnet 7777/udp` and `faithnet 7777/tcp` in
#        `/etc/services` (when sudo is available) so OS tools (`ss`, `netstat`,
#        `getent services faithnet`) recognize Port 7777 as `faithnet`!
# ==============================================================================
set -euo pipefail

# ==============================================================================
# BLOCK 2: OPENING BLOCK
# ( Paths & Canonical FNP/1 Constants )
# ==============================================================================

readonly BIN_DIR="${HOME}/.local/bin"
readonly HANDLER_BIN="${BIN_DIR}/faithnet-open"
readonly APP_DIR="${HOME}/.local/share/applications"
readonly DESKTOP_FILE="${APP_DIR}/faithnet-uri.desktop"
readonly FNP_PORT="7777"
readonly WELL_PORT="8077"

# ==============================================================================
# BLOCK 3: BODY BLOCK
# ( Install faithnet-open Handler, Desktop Entry & /etc/services Port 7777 )
# ==============================================================================

install_uri_handler() {
  mkdir -p "${BIN_DIR}" "${APP_DIR}"
  echo "[1/3] Installing ${HANDLER_BIN}..."
  cat > "${HANDLER_BIN}" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

URI="${1:-faithnet://faithnet.lan/well}"
if [[ "${URI}" != faithnet://* ]]; then
  echo "faithnet-open: URI must start with faithnet:// (got: ${URI})" >&2
  exit 1
fi

REST="${URI#faithnet://}"
AUTHORITY="${REST%%/*}"
PATH_PART=""
if [[ "${REST}" == */* ]]; then
  PATH_PART="${REST#*/}"
fi
HOST="${AUTHORITY%%:*}"
ACTION="${PATH_PART%%/*}"
ACTION="${ACTION,,}"

if [[ -z "${HOST}" || "${HOST}" == "home-anchor" || "${HOST}" == 0x* || "${HOST}" == 0X* ]]; then
  HOST="faithnet.lan"
fi

case "${ACTION}" in
  ""|"well")
    URL="https://${HOST}:8077/"
    if [[ -n "${DISPLAY:-}${WAYLAND_DISPLAY:-}" ]] && command -v xdg-open >/dev/null 2>&1; then
      exec xdg-open "${URL}"
    else
      echo "${URL}"
    fi
    ;;
  "status")
    curl -fsS "https://${HOST}:8077/api/status" && echo ""
    ;;
  "pair")
    echo "https://${HOST}:8077/pair.sh"
    ;;
  "ca")
    echo "https://${HOST}:8077/faithnet-root-ca.crt"
    ;;
  "ping")
    echo "FNP/1 native ping target: ${HOST}:7777 (UDP Rung 3)"
    if [[ -x "${HOME}/.local/bin/faithnet-node" ]]; then
      echo "Use: ${HOME}/.local/bin/faithnet-node --join=${HOST}:7777 --interactive"
    fi
    ;;
  *)
    echo "faithnet-open: unsupported action '${ACTION}' in ${URI}" >&2
    exit 1
    ;;
esac
EOF
  chmod 0755 "${HANDLER_BIN}"

  echo "[2/3] Registering x-scheme-handler/faithnet desktop entry..."
  cat > "${DESKTOP_FILE}" <<EOF
[Desktop Entry]
Type=Application
Name=FaithNet Protocol Handler (FNP/1)
Comment=Open faithnet:// sovereign local network links
Exec=${HANDLER_BIN} %u
Terminal=false
NoDisplay=true
MimeType=x-scheme-handler/faithnet;
EOF
  if command -v xdg-mime >/dev/null 2>&1; then
    xdg-mime default faithnet-uri.desktop x-scheme-handler/faithnet || true
  fi
}

register_etc_services_port() {
  echo "[3/3] Checking canonical FNP/1 port ${FNP_PORT} in /etc/services..."
  if grep -qE '^faithnet[[:space:]]+7777/' /etc/services 2>/dev/null; then
    echo "      /etc/services already lists faithnet (${FNP_PORT}/udp & ${FNP_PORT}/tcp)."
    return 0
  fi
  if sudo -n true 2>/dev/null; then
    sudo tee -a /etc/services >/dev/null <<EOF

# FaithNet Sovereign Adaptive Networking Protocol (FNP/1)
faithnet        ${FNP_PORT}/tcp                         # FaithNet Protocol (Rung 4 TCP)
faithnet        ${FNP_PORT}/udp                         # FaithNet Protocol (Rung 3 UDP)
faithnet-well   ${WELL_PORT}/tcp                        # FaithNet The Well Legacy Bridge
EOF
    echo "      Registered faithnet (${FNP_PORT}/tcp, ${FNP_PORT}/udp) & faithnet-well (${WELL_PORT}/tcp) in /etc/services!"
  else
    echo "      (Run with sudo to add canonical port 7777/udp & 7777/tcp to /etc/services)"
  fi
}

# ==============================================================================
# BLOCK 4: AUDIT PROCEDURES BLOCK
# ( Verify Registered Handler & Port Lookup )
# ==============================================================================

postflight_verify_protocol() {
  echo ""
  echo "=== Post-Flight Verification ==="
  if command -v xdg-mime >/dev/null 2>&1; then
    echo "  x-scheme-handler/faithnet -> $(xdg-mime query default x-scheme-handler/faithnet 2>/dev/null || echo 'unregistered')"
  fi
  if command -v getent >/dev/null 2>&1; then
    getent services faithnet/udp || true
    getent services faithnet/tcp || true
  fi
}

# ==============================================================================
# BLOCK 5: CLOSING BLOCK
# ==============================================================================

main() {
  install_uri_handler
  register_etc_services_port
  postflight_verify_protocol
}

main "$@"

# End of tests/playtester/field_kits/01_two_device_room_well/install_faithnet_protocol.sh
