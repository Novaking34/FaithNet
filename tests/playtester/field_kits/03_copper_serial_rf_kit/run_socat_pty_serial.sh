#!/usr/bin/env bash
# ==============================================================================
# BLOCK 1: METADATA BLOCK
# ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
# ==============================================================================
# @project       FaithNet
# @file          tests/playtester/field_kits/03_copper_serial_rf_kit/run_socat_pty_serial.sh
# @id            playtest.field_kit.03.linux.socat_pty_serial
# @layer         Layer 4: Proving Ground (Field Kit 03 Phase 1 — Kernel PTY Serial)
# @archetype     field_script | operational
# @version       0.2.0-dev
# @interface_rev Rev 5
# @date          2026-11-17
# @updated       2026-11-17
# @author        CreativeWorkzStudio LLC / ProfessorSeanEX
# @repository    https://github.com/CreativeWorkzStudio/FaithNet
# @system        FaithNet Rung 1 Copper Serial Kernel PTY (`socat`) Pre-Test
# @role          Verify Rung 1 SLIP Over Real Linux `/dev/pts/*` Raw Kernel TTYs
# @maturity      Rev 5 Baseline | Wave 17 Field Kit 03 Phase 1
# @chassis       single-laptop-pty | Zero external cables required for Phase 1
# @scale         Real kernel TTY discipline (`stty raw -echo`) + 512B SLIP + glitch resync
# @seams         in: socat, stty, go, pty_serial_bench | exports: PTY SLIP telemetry
# @invariants    1. Universal 5-Block Lifecycle Protocol (Rev 5).
#                2. Creates ephemeral `/dev/pts/*` pair via socat and cleans up on EXIT.
#                3. Tests literal 0xC0/0xDB SLIP escaping + mid-stream noise resync.
#
# PURPOSE & ARCHITECTURAL ROLE:
#   Executes Field Kit 03 Phase 1 on a single Linux laptop without requiring
#   USB-TTL copper cables yet: creates two real Linux kernel pseudoterminal
#   serial devices (`/dev/pts/N` <-> `/dev/pts/M`) with `socat`, configures raw
#   8N1 serial line discipline (`stty raw -echo`), and runs FaithNet's Rung 1
#   `SerialTransceiver` (RFC 1055 SLIP) benchmark & cable-glitch resync suite
#   directly across the live kernel TTY layer!
# ==============================================================================
set -euo pipefail

# ==============================================================================
# BLOCK 2: OPENING BLOCK
# ( Operational Paths & Socat Process Cleanup Trap )
# ==============================================================================

readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly REPO_ROOT="$(cd "${SCRIPT_DIR}/../../../.." && pwd)"
readonly SOCAT_LOG="/tmp/faithnet-kit03-socat.log"
SOCAT_PID=""

cleanup() {
  if [[ -n "${SOCAT_PID}" ]]; then
    kill "${SOCAT_PID}" 2>/dev/null || true
    SOCAT_PID=""
  fi
  rm -f "${SOCAT_LOG}"
}
trap cleanup EXIT INT TERM

# ==============================================================================
# BLOCK 3: BODY BLOCK
# ( Kernel PTY Creation, Raw TTY Discipline & Live Rung 1 SLIP Run )
# ==============================================================================

run_socat_pty_test() {
  echo "============================================================"
  echo "FAITHNET FIELD KIT 03 (PHASE 1): KERNEL PTY SERIAL (SOCAT)"
  echo "============================================================"

  : > "${SOCAT_LOG}"
  socat -d -d pty,raw,echo=0 pty,raw,echo=0 2>"${SOCAT_LOG}" &
  SOCAT_PID=$!

  local pty_a="" pty_b=""
  for _ in $(seq 1 20); do
    mapfile -t ptys < <(grep -oE '/dev/pts/[0-9]+' "${SOCAT_LOG}" 2>/dev/null || true)
    if (( ${#ptys[@]} >= 2 )); then
      pty_a="${ptys[0]}"
      pty_b="${ptys[1]}"
      break
    fi
    sleep 0.1
  done

  if [[ -z "${pty_a}" || -z "${pty_b}" ]]; then
    echo "[FAIL] socat did not allocate two /dev/pts/* devices:" >&2
    cat "${SOCAT_LOG}" >&2
    return 1
  fi

  echo "[1/3] Real Linux Kernel PTY Pair Allocated: ${pty_a} <-> ${pty_b}"
  stty -F "${pty_a}" raw -echo -echoe -echok
  stty -F "${pty_b}" raw -echo -echoe -echok
  echo "[2/3] Configured raw serial TTY line discipline (stty raw -echo) on both ends."

  echo "[3/3] Running FaithNet Rung 1 SLIP SerialTransceiver across ${pty_a} <-> ${pty_b}:"
  (cd "${REPO_ROOT}" && go run ./tests/playtester/harness/pty_serial_bench --pty-a="${pty_a}" --pty-b="${pty_b}")
}

# ==============================================================================
# BLOCK 4: AUDIT PROCEDURES BLOCK
# ( Pre-Flight Dependency Check )
# ==============================================================================

preflight_kit03_phase1() {
  for bin in socat stty go; do
    if ! command -v "${bin}" >/dev/null 2>&1; then
      echo "[PREFLIGHT FAIL] Missing required tool: ${bin}" >&2
      return 1
    fi
  done
  return 0
}

# ==============================================================================
# BLOCK 5: CLOSING BLOCK
# ==============================================================================

main() {
  preflight_kit03_phase1
  run_socat_pty_test
}

main "$@"

# End of tests/playtester/field_kits/03_copper_serial_rf_kit/run_socat_pty_serial.sh
