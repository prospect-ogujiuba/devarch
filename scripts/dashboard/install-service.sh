#!/usr/bin/env bash
set -euo pipefail

script_dir="$(CDPATH='' cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)"
repo_root="$(CDPATH='' cd -- "$script_dir/../.." && pwd -P)"
template="$script_dir/devarch-dashboard.service.in"
unit_dir="${XDG_CONFIG_HOME:-$HOME/.config}/systemd/user"
unit_file="$unit_dir/devarch-dashboard.service"
python_bin="$(command -v python3 || true)"

if [[ -z "$python_bin" ]]; then
  printf 'python3 is required to install DevArch Home.\n' >&2
  exit 1
fi
if ! command -v systemctl >/dev/null 2>&1; then
  printf 'systemctl is required to install the user service.\n' >&2
  exit 1
fi

mkdir -p "$unit_dir"
temporary="$(mktemp)"
trap 'rm -f "$temporary"' EXIT

"$python_bin" - "$template" "$temporary" "$repo_root" "$python_bin" <<'PY'
from pathlib import Path
import sys

template, output, repo_root, python = sys.argv[1:]
for value in (repo_root, python):
    if any(character in value for character in ('\n', '\r', '"')):
        raise SystemExit(f"Unsupported character in service path: {value!r}")
text = Path(template).read_text(encoding="utf-8")
text = text.replace("@REPO_ROOT@", repo_root).replace("@PYTHON@", python)
Path(output).write_text(text, encoding="utf-8")
PY

install -m 0644 "$temporary" "$unit_file"
systemctl --user daemon-reload
if command -v podman >/dev/null 2>&1; then
  systemctl --user enable --now podman.socket >/dev/null 2>&1 || \
    printf 'Warning: Podman API socket could not be enabled; container inventory may be unavailable.\n' >&2
fi
systemctl --user enable devarch-dashboard.service
systemctl --user restart devarch-dashboard.service

if ! systemctl --user is-active --quiet devarch-dashboard.service; then
  systemctl --user --no-pager --full status devarch-dashboard.service >&2 || true
  exit 1
fi

"$python_bin" - <<'PY'
import socket
import time

for _ in range(30):
    try:
        with socket.create_connection(("127.0.0.1", 7411), timeout=0.2):
            break
    except OSError:
        time.sleep(0.1)
else:
    raise SystemExit("DevArch Home did not begin listening on port 7411.")
PY

if command -v podman >/dev/null 2>&1 && podman container exists nginx-proxy-manager; then
  podman exec nginx-proxy-manager nginx -t
  podman exec nginx-proxy-manager nginx -s reload
else
  printf 'Warning: Nginx Proxy Manager is not running; start it before opening devarch.test.\n' >&2
fi

printf 'DevArch Home is running at https://devarch.test\n'
printf 'Service: systemctl --user status devarch-dashboard.service\n'

if command -v loginctl >/dev/null 2>&1; then
  linger="$(loginctl show-user "$USER" -p Linger --value 2>/dev/null || true)"
  if [[ "$linger" != "yes" ]]; then
    printf 'Note: enable lingering if the dashboard must run while you are logged out: loginctl enable-linger %q\n' "$USER"
  fi
fi
