#!/usr/bin/env bash
# ==============================================================================
# BLOCK 1: METADATA BLOCK — FaithNet Field Kit 02 (Acoustic Air-Gap RX)
# @project FaithNet | @file tests/playtester/field_kits/02_acoustic_airgap_kit/rx_mic.sh
# @role    Record 9600 Hz mono 16-bit PCM WAV from microphone and decode FaithNet
# ==============================================================================
set -euo pipefail

SECONDS_REC="${1:-4}"
IN_WAV="/tmp/faithnet_airgap_rx.wav"

echo "[RX] Recording ${SECONDS_REC}s at 9600 Hz S16_LE mono via arecord..."
arecord -f S16_LE -r 9600 -c 1 -d "${SECONDS_REC}" "${IN_WAV}"
echo "[RX] Running FaithNet Rung 1 Goertzel demodulator + Codec CRC verification..."
go run ./tests/playtester/harness/audio_torture_bench --decode-wav="${IN_WAV}"
