#!/usr/bin/env bash
set -euo pipefail

LIBRARY_DIR="${1:-../services-library}"

if [[ ! -d "$LIBRARY_DIR/compose" ]]; then
  echo "ERROR: $LIBRARY_DIR/compose not found"
  exit 1
fi

yaml_moved=0
config_moved=0

for category_dir in "$LIBRARY_DIR"/compose/*/; do
  category=$(basename "$category_dir")
  for yml in "$category_dir"*.yml; do
    [[ -f "$yml" ]] || continue
    svc=$(basename "$yml" .yml)
    target_dir="$LIBRARY_DIR/$category/$svc"
    mkdir -p "$target_dir"
    mv "$yml" "$target_dir/"
    yaml_moved=$((yaml_moved + 1))
  done
done

for config_dir in "$LIBRARY_DIR"/config/*/; do
  [[ -d "$config_dir" ]] || continue
  svc=$(basename "$config_dir")

  found_category=""
  for category_dir in "$LIBRARY_DIR"/*/; do
    category=$(basename "$category_dir")
    [[ "$category" == "compose" || "$category" == "config" ]] && continue
    if [[ -d "$category_dir/$svc" ]]; then
      found_category="$category"
      break
    fi
  done

  if [[ -z "$found_category" ]]; then
    echo "WARNING: no category found for config/$svc, skipping"
    continue
  fi

  mv "$config_dir" "$LIBRARY_DIR/$found_category/$svc/config"
  config_moved=$((config_moved + 1))
done

rmdir "$LIBRARY_DIR"/compose/*/ 2>/dev/null || true
rmdir "$LIBRARY_DIR/compose" 2>/dev/null || true
rmdir "$LIBRARY_DIR/config" 2>/dev/null || true

echo "Done: $yaml_moved YAMLs moved, $config_moved config dirs moved"
