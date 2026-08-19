#!/usr/bin/env bash
set -Eeuo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REGISTER="$SCRIPT_DIR/register-host.sh"
TEST_TMP="$(mktemp -d)"
HOSTS_FIXTURE="$TEST_TMP/hosts"
passed=0

cleanup() { rm -rf "$TEST_TMP"; }
trap cleanup EXIT
fail() { printf 'FAIL: %s\n' "$*" >&2; exit 1; }
pass() { ((passed += 1)); }
assert_count() {
  local expected="$1" pattern="$2"
  [[ "$(grep -Ec "$pattern" "$HOSTS_FIXTURE" || true)" == "$expected" ]] || fail "expected $expected matches for $pattern"
  pass
}

cat > "$HOSTS_FIXTURE" <<'EOF'
127.0.0.1    localhost
::1  localhost ip6-localhost
192.0.2.10    demo.test    retained.test    # local aliases
# keep this comment
EOF

expected_new="$TEST_TMP/expected-new"
cp "$HOSTS_FIXTURE" "$expected_new"
printf '127.0.0.1\tmoneytrees.test\n' >> "$expected_new"
HOSTS_FILE="$HOSTS_FIXTURE" DEVARCH_HOSTS_PLATFORM=unix "$REGISTER" moneytrees.test >/dev/null
cmp -s "$HOSTS_FIXTURE" "$expected_new" || fail 'adding a new hostname reformatted existing Unix hosts lines'
pass

before="$(cksum "$HOSTS_FIXTURE")"
HOSTS_FILE="$HOSTS_FIXTURE" DEVARCH_HOSTS_PLATFORM=unix "$REGISTER" moneytrees.test >/dev/null
after="$(cksum "$HOSTS_FIXTURE")"
[[ "$before" == "$after" ]] || fail 're-registering a canonical Unix mapping changed the hosts file'
pass

HOSTS_FILE="$HOSTS_FIXTURE" DEVARCH_HOSTS_PLATFORM=unix "$REGISTER" demo.test >/dev/null
assert_count 1 '^127\.0\.0\.1[[:space:]]+demo\.test$'
assert_count 1 '^192\.0\.2\.10[[:space:]]+retained\.test[[:space:]]+# local aliases$'
grep -Fxq '127.0.0.1    localhost' "$HOSTS_FIXTURE" || fail 'updating another hostname reformatted an unrelated Unix line'
pass

before="$(cksum "$HOSTS_FIXTURE")"
HOSTS_FILE="$HOSTS_FIXTURE" DEVARCH_HOSTS_PLATFORM=unix "$REGISTER" demo.test >/dev/null
assert_count 1 'demo\.test'
after="$(cksum "$HOSTS_FIXTURE")"
[[ "$before" == "$after" ]] || fail 're-registering an updated Unix mapping changed the hosts file'
pass

before="$(cksum "$HOSTS_FIXTURE")"
HOSTS_FILE="$HOSTS_FIXTURE" DEVARCH_HOSTS_PLATFORM=unix "$REGISTER" dry-run.test --dry-run >/dev/null
after="$(cksum "$HOSTS_FIXTURE")"
[[ "$before" == "$after" ]] || fail 'dry-run changed the hosts file'
pass

if HOSTS_FILE="$HOSTS_FIXTURE" DEVARCH_HOSTS_PLATFORM=unix "$REGISTER" 'unsafe host.test' >/dev/null 2>&1; then
  fail 'invalid hostname should fail'
fi
pass
if HOSTS_FILE="$HOSTS_FIXTURE" DEVARCH_HOSTS_PLATFORM=unix "$REGISTER" demo.test --address 999.1.1.1 >/dev/null 2>&1; then
  fail 'invalid IPv4 address should fail'
fi
pass

if command -v powershell.exe >/dev/null 2>&1 && command -v wslpath >/dev/null 2>&1; then
  POWERSHELL_REGISTER="$SCRIPT_DIR/register-host.ps1"
  WINDOWS_FIXTURE="$TEST_TMP/windows-hosts"
  WINDOWS_EXPECTED="$TEST_TMP/windows-expected"
  printf '# comment\r\n127.0.0.1    first.test    second.test\r\n' > "$WINDOWS_FIXTURE"
  cp "$WINDOWS_FIXTURE" "$WINDOWS_EXPECTED"
  printf '127.0.0.1\tmoneytrees.test\r\n' >> "$WINDOWS_EXPECTED"
  powershell.exe -NoProfile -ExecutionPolicy Bypass -File "$(wslpath -w "$POWERSHELL_REGISTER")" \
    -HostName moneytrees.test -HostsPath "$(wslpath -w "$WINDOWS_FIXTURE")" >/dev/null
  cmp -s "$WINDOWS_FIXTURE" "$WINDOWS_EXPECTED" || fail 'adding a new hostname reformatted existing Windows hosts lines or line endings'
  pass

  before="$(cksum "$WINDOWS_FIXTURE")"
  powershell.exe -NoProfile -ExecutionPolicy Bypass -File "$(wslpath -w "$POWERSHELL_REGISTER")" \
    -HostName moneytrees.test -HostsPath "$(wslpath -w "$WINDOWS_FIXTURE")" >/dev/null
  after="$(cksum "$WINDOWS_FIXTURE")"
  [[ "$before" == "$after" ]] || fail 're-registering a canonical Windows mapping changed the hosts file'
  pass
fi

printf 'PASS: %d assertions\n' "$passed"
