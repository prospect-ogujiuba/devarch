#!/usr/bin/env bash
set -euo pipefail

script_dir="$(CDPATH='' cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)"

if [[ ! -f "$script_dir/static/dashboard.css" ]]; then
  printf 'Dashboard CSS is missing. Run: cd %q && npm install && npm run build\n' "$script_dir" >&2
  exit 1
fi

exec python3 "$script_dir/server.py" "$@"
