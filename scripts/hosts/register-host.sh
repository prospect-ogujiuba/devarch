#!/usr/bin/env bash
set -Eeuo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
HOSTNAME_VALUE=""
ADDRESS="127.0.0.1"
DRY_RUN=false

usage() {
  cat <<'EOF'
Usage: scripts/hosts/register-host.sh <hostname> [options]

Idempotently map a local development hostname in the operating system hosts file.

Options:
  --address IP   Address to map (default: 127.0.0.1)
  --dry-run      Print the intended mapping without changing the hosts file
  -h, --help     Show this help
EOF
}

die() {
  printf '[hosts] error: %s\n' "$*" >&2
  exit 1
}

while (($# > 0)); do
  case "$1" in
    --address)
      (($# >= 2)) || die '--address requires a value'
      ADDRESS="$2"
      shift 2
      ;;
    --dry-run) DRY_RUN=true; shift ;;
    -h|--help) usage; exit 0 ;;
    --*) die "unknown option: $1" ;;
    *)
      [[ -z "$HOSTNAME_VALUE" ]] || die "unexpected argument: $1"
      HOSTNAME_VALUE="$1"
      shift
      ;;
  esac
done

[[ "$HOSTNAME_VALUE" =~ ^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?(\.[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?)+$ ]] || \
  die 'hostname must be a lowercase DNS name with at least two labels'
[[ "$ADDRESS" =~ ^([0-9]{1,3}\.){3}[0-9]{1,3}$ ]] || die 'address must be an IPv4 address'
IFS=. read -r -a address_parts <<<"$ADDRESS"
for part in "${address_parts[@]}"; do
  ((10#$part <= 255)) || die 'address must be an IPv4 address'
done

if [[ "$DRY_RUN" == true ]]; then
  printf '[hosts] would register %s %s\n' "$ADDRESS" "$HOSTNAME_VALUE"
  exit 0
fi

platform="${DEVARCH_HOSTS_PLATFORM:-}"
if [[ -z "$platform" ]]; then
  uname_value="$(uname -s)"
  case "$uname_value" in
    Darwin|Linux)
      if [[ "$uname_value" == Linux ]] && grep -qi microsoft /proc/sys/kernel/osrelease 2>/dev/null; then
        platform=windows
      else
        platform=unix
      fi
      ;;
    MINGW*|MSYS*|CYGWIN*) platform=windows ;;
    *) die "unsupported operating system: $uname_value" ;;
  esac
fi

if [[ "$platform" == windows ]]; then
  command -v powershell.exe >/dev/null 2>&1 || die 'powershell.exe is required to update the Windows hosts file'
  if command -v wslpath >/dev/null 2>&1; then
    powershell_script="$(wslpath -w "$SCRIPT_DIR/register-host.ps1")"
  elif command -v cygpath >/dev/null 2>&1; then
    powershell_script="$(cygpath -w "$SCRIPT_DIR/register-host.ps1")"
  else
    powershell_script="$SCRIPT_DIR/register-host.ps1"
  fi
  powershell.exe -NoProfile -ExecutionPolicy Bypass -File "$powershell_script" \
    -HostName "$HOSTNAME_VALUE" -Address "$ADDRESS"
  exit $?
fi

[[ "$platform" == unix ]] || die "unsupported hosts platform override: $platform"
hosts_file="${HOSTS_FILE:-/etc/hosts}"
[[ -f "$hosts_file" ]] || die "hosts file not found: $hosts_file"

if awk -v wanted="$HOSTNAME_VALUE" -v wanted_address="$ADDRESS" '
  /^[[:space:]]*#/ || /^[[:space:]]*$/ { next }
  {
    line = $0
    sub(/#.*/, "", line)
    count = split(line, fields, /[[:space:]]+/)
    address = ""
    aliases = 0
    line_matches = 0
    for (i = 1; i <= count; i++) {
      if (fields[i] == "") continue
      if (address == "") { address = fields[i]; continue }
      aliases++
      if (tolower(fields[i]) == tolower(wanted)) { occurrences++; line_matches++ }
    }
    if (line_matches == 1 && aliases == 1 && address == wanted_address) canonical++
  }
  END { exit !(occurrences == 1 && canonical == 1) }
' "$hosts_file"; then
  printf '[hosts] already registered %s %s in %s\n' "$ADDRESS" "$HOSTNAME_VALUE" "$hosts_file"
  exit 0
fi

temp_file="$(mktemp)"
trap 'rm -f "$temp_file"' EXIT

awk -v wanted="$HOSTNAME_VALUE" '
  /^[[:space:]]*#/ || /^[[:space:]]*$/ { print; next }
  {
    original = $0
    line = $0
    comment = ""
    comment_at = index(line, "#")
    if (comment_at) {
      comment = substr(line, comment_at)
      line = substr(line, 1, comment_at - 1)
    }
    count = split(line, fields, /[[:space:]]+/)
    output = ""
    address = ""
    found = 0
    for (i = 1; i <= count; i++) {
      if (fields[i] == "") continue
      if (address == "") { address = fields[i]; continue }
      if (tolower(fields[i]) == tolower(wanted)) found = 1
      else output = output "\t" fields[i]
    }
    if (!found) { print original; next }
    if (address != "" && output != "") print address output (comment == "" ? "" : " " comment)
    else if (address != "" && comment != "") print comment
  }
' "$hosts_file" > "$temp_file"
printf '%s\t%s\n' "$ADDRESS" "$HOSTNAME_VALUE" >> "$temp_file"

if [[ -w "$hosts_file" ]]; then
  cat "$temp_file" > "$hosts_file"
elif command -v sudo >/dev/null 2>&1; then
  printf '[hosts] administrator access is required to update %s\n' "$hosts_file"
  cat "$temp_file" | sudo tee "$hosts_file" >/dev/null
else
  die "administrator access is required to update $hosts_file (sudo not found)"
fi

printf '[hosts] registered %s %s in %s\n' "$ADDRESS" "$HOSTNAME_VALUE" "$hosts_file"
