#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
config="$repo_root/.pi/mcp.json"
runtime_check=false

case "${1:-}" in
  "") ;;
  --runtime) runtime_check=true ;;
  *)
    printf 'usage: %s [--runtime]\n' "$0" >&2
    exit 2
    ;;
esac

python3 - "$config" <<'PY'
import json
import sys
from pathlib import Path

path = Path(sys.argv[1])
data = json.loads(path.read_text())
server = data.get("mcpServers", {}).get("playwright")
assert isinstance(server, dict), "missing mcpServers.playwright"
assert server.get("command") == "npx", "playwright command must be npx"
args = server.get("args")
assert isinstance(args, list), "playwright args must be a list"
assert "-y" in args, "npx must run non-interactively"
packages = [arg for arg in args if arg.startswith("@playwright/mcp@")]
assert len(packages) == 1, "pin exactly one @playwright/mcp version"
assert packages[0] != "@playwright/mcp@latest", "pin Playwright MCP for reproducibility"
print(f"config ok: {packages[0]}")
PY

if "$runtime_check"; then
  package="$(python3 - "$config" <<'PY'
import json
import sys
from pathlib import Path

args = json.loads(Path(sys.argv[1]).read_text())["mcpServers"]["playwright"]["args"]
print(next(arg for arg in args if arg.startswith("@playwright/mcp@")))
PY
)"
  npx -y "$package" --help >/dev/null
  printf 'runtime ok: %s\n' "$package"
fi
