#!/usr/bin/env bash
# ==============================================================================
# BLOCK 1: METADATA BLOCK — FaithNet Field Kit 04 (Privileged Kernel Chaos Kit)
# @project FaithNet | @file tests/playtester/field_kits/04_kernel_chaos_netns/run_netns_chaos.sh
# @role    Non-destructive isolated Linux network namespace + tc netem chaos demo
# ==============================================================================
set -euo pipefail

# Strict safety: creates ONLY temporary isolated namespaces `fn_a` and `fn_b`
# and tears them down automatically on EXIT. Never touches host `wlp3s0`!
cleanup() {
  sudo ip netns del fn_a 2>/dev/null || true
  sudo ip netns del fn_b 2>/dev/null || true
}
trap cleanup EXIT

echo "=== FAITHNET FIELD KIT 04: ISOLATED KERNEL NETNS + TC NETEM CHAOS ==="
cleanup

sudo ip netns add fn_a
sudo ip netns add fn_b
sudo ip link add veth_fn_a type veth peer name veth_fn_b
sudo ip link set veth_fn_a netns fn_a
sudo ip link set veth_fn_b netns fn_b

sudo ip -n fn_a addr add 10.77.0.1/24 dev veth_fn_a
sudo ip -n fn_b addr add 10.77.0.2/24 dev veth_fn_b
sudo ip -n fn_a link set lo up && sudo ip -n fn_a link set veth_fn_a up
sudo ip -n fn_b link set lo up && sudo ip -n fn_b link set veth_fn_b up

echo "[1/3] Baseline clean veth link (10.77.0.1 -> 10.77.0.2):"
sudo ip netns exec fn_a ping -c 3 -i 0.2 10.77.0.2

echo "[2/3] Applying tc netem storm (delay 80ms +-20ms, loss 20%, corrupt 5%):"
sudo ip netns exec fn_a tc qdisc add dev veth_fn_a root netem delay 80ms 20ms loss 20% corrupt 5%
sudo ip netns exec fn_a ping -c 5 -i 0.2 10.77.0.2 || true

echo "[3/3] Cleaning up temporary namespaces (host wlp3s0 untouched)."
