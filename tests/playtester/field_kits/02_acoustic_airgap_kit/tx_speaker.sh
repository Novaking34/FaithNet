#!/usr/bin/env bash
# ==============================================================================
# BLOCK 1: METADATA BLOCK — FaithNet Field Kit 02 (Acoustic Air-Gap TX)
# @project FaithNet | @file tests/playtester/field_kits/02_acoustic_airgap_kit/tx_speaker.sh
# @role    Modulate FaithNet Rung 1 AFSK 1200/2400 Hz WAV and play on speaker
# ==============================================================================
set -euo pipefail

MSG="${1:-BLOCK-7-LOCAL-NEWS:AIRGAP-LIVE}"
OUT_WAV="/tmp/faithnet_airgap_tx.wav"

go run ./tests/playtester/harness/audio_torture_bench \
  --export-wav="${OUT_WAV}" \
  --msg="${MSG}"

echo "[TX] Playing FaithNet Rung 1 AFSK waveform via aplay..."
aplay "${OUT_WAV}"
