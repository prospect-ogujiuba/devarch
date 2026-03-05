#!/usr/bin/env bash
set -euo pipefail

IN_DIR="${DAW_IN_DIR:-/work/in}"
OUT_DIR="${DAW_OUT_DIR:-/work/out}"

mkdir -p "$OUT_DIR"

shopt -s nullglob
for f in "$IN_DIR"/*; do
  base="$(basename "$f")"
  stem="${base%.*}"

  essentia_streaming_extractor_music "$f" "$OUT_DIR/${stem}.sig" || true

  if command -v essentia_streaming_extractor_freesound >/dev/null 2>&1; then
    essentia_streaming_extractor_freesound "$f" "$OUT_DIR/${stem}.freesound.sig" || true
  fi

  echo "done: $base"
done
