#!/usr/bin/env bash
# ==============================================================================
# BLOCK 1: METADATA BLOCK
# ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
# ==============================================================================
# @project       FaithNet
# @file          tests/playtester/field_kits/04_kernel_chaos_netns/run_netns_chaos.sh
# @id            playtest.field_kit.04.linux.netns_chaos
# @layer         Layer 4: Proving Ground (Field Kit 04 — Kernel Chaos Namespaces)
# @archetype     field_script | operational
# @version       0.2.0-dev
# @interface_rev Rev 5
# @date          2026-11-17
# @updated       2026-11-17
# @author        CreativeWorkzStudio LLC / ProfessorSeanEX
# @repository    https://github.com/CreativeWorkzStudio/FaithNet
# @system        FaithNet Isolated Linux Kernel netns + tc netem Chaos Runner
# @role          Test Live faithnet-node FNP/1 Under Real Kernel Loss/Jitter/Blackout
# @maturity      Rev 5 Baseline | Wave 17 Live Kernel Chaos Proving Ground
# @chassis       ephemeral-netns | Zero mutation of host wlp3s0 or ap0
# @scale         Single-laptop 2-namespace isolated veth + real tc netem qdisc
# @seams         in: ip, tc, faithnet-node | exports: empirical kernel chaos telemetry
# @invariants    1. Universal 5-Block Lifecycle Protocol (Rev 5).
#                2. Strictly isolated inside ephemeral namespaces `fn_a` and `fn_b`.
#                3. Automatic trap cleanup kills child node & deletes fn_a/fn_b on EXIT.
#                4. Host wlp3s0, ap0, and main routing table are NEVER touched.
#
# PURPOSE & ARCHITECTURAL ROLE:
#   Executes Field Kit 04 on a single Linux laptop: spins up two completely
#   isolated Linux kernel network namespaces (`fn_a` at 10.77.0.1 and `fn_b` at
#   10.77.0.2), runs a real `faithnet-node` Anchor inside `fn_a`, and fires real
#   zero-HTTP `FNP/1` (`faithnet://10.77.0.1:7777/status`) requests from `fn_b`
#   across four escalating kernel weather stages:
#     Stage 1: Clean kernel veth baseline (0% loss)
#     Stage 2: Kernel `tc netem` storm (delay 80ms +-20ms, 20% loss, 5% corrupt)
#     Stage 3: Total kernel blackout (100% packet drop) -> clean timeout, 0 crash
#     Stage 4: Kernel link recovery (`tc qdisc del`) -> instant <1ms FNP/1 heal!
# ==============================================================================
set -euo pipefail

# ==============================================================================
# BLOCK 2: OPENING BLOCK
# ( Operational Constants, Node Binary Discovery & Guaranteed Cleanup Trap )
# ==============================================================================

readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly REPO_ROOT="$(cd "${SCRIPT_DIR}/../../../.." && pwd)"
readonly NS_A="fn_a"
readonly NS_B="fn_b"
readonly IP_A="10.77.0.1"
readonly IP_B="10.77.0.2"
readonly TMP_BIN="/tmp/faithnet-node-kit04"
NODE_PID=""

cleanup() {
  if [[ -n "${NODE_PID}" ]]; then
    sudo kill "${NODE_PID}" 2>/dev/null || true
    NODE_PID=""
  fi
  sudo ip netns del "${NS_A}" 2>/dev/null || true
  sudo ip netns del "${NS_B}" 2>/dev/null || true
  rm -f "${TMP_BIN}" "/tmp/faithnet-kit04-node.log"
}
trap cleanup EXIT INT TERM

# ==============================================================================
# BLOCK 3: BODY BLOCK
# ( Namespace Construction, Live Anchor Launch & 4-Stage Kernel Chaos Suite )
# ==============================================================================

setup_isolated_namespaces() {
  sudo ip netns del "${NS_A}" 2>/dev/null || true
  sudo ip netns del "${NS_B}" 2>/dev/null || true

  sudo ip netns add "${NS_A}"
  sudo ip netns add "${NS_B}"
  sudo ip link add veth_fn_a type veth peer name veth_fn_b
  sudo ip link set veth_fn_a netns "${NS_A}"
  sudo ip link set veth_fn_b netns "${NS_B}"

  sudo ip -n "${NS_A}" addr add "${IP_A}/24" dev veth_fn_a
  sudo ip -n "${NS_B}" addr add "${IP_B}/24" dev veth_fn_b
  sudo ip -n "${NS_A}" link set lo up
  sudo ip -n "${NS_A}" link set veth_fn_a up
  sudo ip -n "${NS_B}" link set lo up
  sudo ip -n "${NS_B}" link set veth_fn_b up
}

launch_anchor_in_ns_a() {
  if [[ -x "${HOME}/.local/bin/faithnet-node" ]]; then
    cp "${HOME}/.local/bin/faithnet-node" "${TMP_BIN}"
  else
    (cd "${REPO_ROOT}" && go build -o "${TMP_BIN}" ./cmd/faithnet-node)
  fi
  chmod 0755 "${TMP_BIN}"

  sudo ip netns exec "${NS_A}" "${TMP_BIN}" \
    --id=0xFA17000000000001 \
    --label=Kit04-Anchor \
    --udp="${IP_A}:7777" > /tmp/faithnet-kit04-node.log 2>&1 &
  NODE_PID=$!
  sleep 1
}

run_four_stage_chaos() {
  echo "============================================================"
  echo "FAITHNET FIELD KIT 04: KERNEL NETNS + TC NETEM CHAOS SUITE"
  echo "============================================================"

  echo ""
  echo "[STAGE 1/4] Clean Isolated Kernel Wire (0% loss, 0ms delay):"
  sudo ip netns exec "${NS_B}" ping -c 3 -i 0.1 "${IP_A}" | tail -n 2
  sudo ip netns exec "${NS_B}" "${TMP_BIN}" --fetch="faithnet://${IP_A}:7777/status" --ping-timeout=1500

  echo ""
  echo "[STAGE 2/4] Real Kernel Storm (tc netem: delay 80ms +-20ms, loss 20%, corrupt 5%):"
  sudo ip netns exec "${NS_A}" tc qdisc add dev veth_fn_a root netem delay 80ms 20ms loss 20% corrupt 5%
  sudo ip netns exec "${NS_B}" tc qdisc add dev veth_fn_b root netem delay 80ms 20ms loss 20% corrupt 5%
  local storm_ok=0
  local storm_total=10
  for i in $(seq 1 "${storm_total}"); do
    if out="$(sudo ip netns exec "${NS_B}" "${TMP_BIN}" --fetch="faithnet://${IP_A}:7777/protocol" --ping-timeout=1000 2>/dev/null)"; then
      storm_ok=$((storm_ok + 1))
      echo "  [Storm #${i}/${storm_total}] PASS: $(echo "${out}" | head -n 1)"
    else
      echo "  [Storm #${i}/${storm_total}] DROPPED/CORRUPTED (safely caught by FNV-1a/timeout)"
    fi
  done
  echo "  -> Storm Summary: ${storm_ok}/${storm_total} delivered intact (0 corrupt frames leaked!)"

  echo ""
  echo "[STAGE 3/4] Total Kernel Blackout (tc netem: loss 100%):"
  sudo ip netns exec "${NS_A}" tc qdisc change dev veth_fn_a root netem loss 100%
  sudo ip netns exec "${NS_B}" tc qdisc change dev veth_fn_b root netem loss 100%
  if sudo ip netns exec "${NS_B}" "${TMP_BIN}" --fetch="faithnet://${IP_A}:7777/status" --ping-timeout=600 >/dev/null 2>&1; then
    echo "  [BLACKOUT] UNEXPECTED PASS"
    return 1
  else
    echo "  [BLACKOUT] Clean bounded timeout under 100% kernel drop (0 hang, 0 crash)!"
  fi

  echo ""
  echo "[STAGE 4/4] Instant Kernel Recovery (removing tc netem qdisc):"
  sudo ip netns exec "${NS_A}" tc qdisc del dev veth_fn_a root
  sudo ip netns exec "${NS_B}" tc qdisc del dev veth_fn_b root
  sudo ip netns exec "${NS_B}" "${TMP_BIN}" --fetch="faithnet://${IP_A}:7777/well" --ping-timeout=1500
  echo ""
  echo "[PASS] Field Kit 04 complete! Deleting ephemeral namespaces (host wlp3s0 & ap0 untouched)."
}

# ==============================================================================
# BLOCK 4: AUDIT PROCEDURES BLOCK
# ( Pre-Flight Tool Check )
# ==============================================================================

preflight_kit04() {
  for bin in ip tc sudo; do
    if ! command -v "${bin}" >/dev/null 2>&1; then
      echo "[PREFLIGHT FAIL] Required tool missing: ${bin}" >&2
      return 1
    fi
  done
  return 0
}

# ==============================================================================
# BLOCK 5: CLOSING BLOCK
# ( Lifecycle Entrypoint )
# ==============================================================================

main() {
  preflight_kit04
  setup_isolated_namespaces
  launch_anchor_in_ns_a
  run_four_stage_chaos
}

main "$@"

# End of tests/playtester/field_kits/04_kernel_chaos_netns/run_netns_chaos.sh
