#!/usr/bin/env bash
set -euo pipefail

SOURCE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEST_TMP="$(mktemp -d)"
PROJECT_ROOT="$TEST_TMP/project"
MATRIX="$PROJECT_ROOT/scripts/javascript/scaffold-matrix.sh"
CALL_LOG="$TEST_TMP/calls.log"
passed=0

cleanup() { rm -rf "$TEST_TMP"; }
trap cleanup EXIT
pass() { ((passed += 1)); }
fail() { printf 'FAIL: %s\n' "$*" >&2; exit 1; }
assert_contains() { grep -Fq -- "$2" <<<"$1" || fail "$3 (missing '$2')"; pass; }

mkdir -p \
  "$PROJECT_ROOT/scripts/javascript/profiles/alpha" \
  "$PROJECT_ROOT/scripts/javascript/profiles/beta" \
  "$PROJECT_ROOT/scripts/node" \
  "$PROJECT_ROOT/services-library/backend/node" \
  "$PROJECT_ROOT/apps" \
  "$TEST_TMP/bin"
cp "$SOURCE_DIR/scaffold-matrix.sh" "$MATRIX"
printf "FRAMEWORK_DESCRIPTION='Alpha'\nDEFAULT_PROFILE='one'\n" > "$PROJECT_ROOT/scripts/javascript/profiles/alpha/framework.conf"
printf "PROFILE_DESCRIPTION='One'\nSCAFFOLD=(true)\nDEVARCH_SCRIPT='true'\n" > "$PROJECT_ROOT/scripts/javascript/profiles/alpha/one.profile"
printf "PROFILE_DESCRIPTION='Two'\nSCAFFOLD=(true)\nDEVARCH_SCRIPT='true'\n" > "$PROJECT_ROOT/scripts/javascript/profiles/alpha/two.profile"
printf "FRAMEWORK_DESCRIPTION='Beta'\nDEFAULT_PROFILE='basic'\n" > "$PROJECT_ROOT/scripts/javascript/profiles/beta/framework.conf"
printf "PROFILE_DESCRIPTION='Basic'\nSCAFFOLD=(true)\nDEVARCH_SCRIPT='true'\n" > "$PROJECT_ROOT/scripts/javascript/profiles/beta/basic.profile"
printf 'services: {}\n' > "$PROJECT_ROOT/services-library/backend/node/app.compose.yml"

cat > "$PROJECT_ROOT/scripts/javascript/bootstrap.sh" <<'FAKE'
#!/usr/bin/env bash
set -euo pipefail
printf 'scaffold %s\n' "$*" >> "${MATRIX_CALL_LOG:?}"
app=$1
mkdir -p "${MATRIX_PROJECT_ROOT:?}/apps/$app"
printf '{"name":"%s","scripts":{"devarch":"true"}}\n' "$app" > "${MATRIX_PROJECT_ROOT:?}/apps/$app/package.json"
FAKE
cat > "$PROJECT_ROOT/scripts/node/bootstrap.sh" <<'FAKE'
#!/usr/bin/env bash
printf 'start %s\n' "$*" >> "${MATRIX_CALL_LOG:?}"
FAKE
cat > "$TEST_TMP/bin/podman" <<'FAKE'
#!/usr/bin/env bash
printf 'stop app=%s %s\n' "${DEVARCH_NODE_APP_NAME:-}" "$*" >> "${MATRIX_CALL_LOG:?}"
FAKE
chmod +x "$MATRIX" "$PROJECT_ROOT/scripts/javascript/bootstrap.sh" "$PROJECT_ROOT/scripts/node/bootstrap.sh" "$TEST_TMP/bin/podman"
: > "$CALL_LOG"

run_matrix() {
  PATH="$TEST_TMP/bin:$PATH" \
  MATRIX_CALL_LOG="$CALL_LOG" \
  MATRIX_PROJECT_ROOT="$PROJECT_ROOT" \
  "$MATRIX" "$@"
}

list_output=$(run_matrix list)
assert_contains "$list_output" 'alpha' 'list should include discovered frameworks'
assert_contains "$list_output" 'showcase-alpha-one' 'list should include deterministic app names'
assert_contains "$list_output" 'Total: 3 profiles' 'list should count discovered profiles'

run_matrix scaffold alpha one >/dev/null
[[ -f "$PROJECT_ROOT/apps/showcase-alpha-one/package.json" ]] || fail 'scaffold should create the deterministic application'
pass
grep -Fqx 'scaffold showcase-alpha-one --framework alpha --profile one' "$CALL_LOG" || fail 'scaffold should delegate exact framework and profile arguments'
pass

before=$(grep -c '^scaffold ' "$CALL_LOG")
skip_output=$(run_matrix scaffold alpha one)
after=$(grep -c '^scaffold ' "$CALL_LOG")
[[ "$before" == "$after" ]] || fail 'resuming should not scaffold an existing application again'
pass
assert_contains "$skip_output" 'skip alpha/one' 'resuming should report skipped applications'

all_output=$(run_matrix scaffold-all)
[[ -f "$PROJECT_ROOT/apps/showcase-alpha-two/package.json" ]] || fail 'scaffold-all should create alpha/two'
pass
[[ -f "$PROJECT_ROOT/apps/showcase-beta-basic/package.json" ]] || fail 'scaffold-all should create beta/basic'
pass
assert_contains "$all_output" 'total=3 created=2 skipped=1' 'scaffold-all should summarize sequential results'

run_matrix start beta basic --no-hosts >/dev/null
grep -Fqx 'start showcase-beta-basic --no-hosts' "$CALL_LOG" || fail 'start should delegate to the Node bootstrap and preserve options'
pass

run_matrix stop beta basic >/dev/null
grep -Fq 'stop app=showcase-beta-basic compose -p devarch-node-showcase-beta-basic' "$CALL_LOG" || fail 'stop should target the isolated Compose project'
pass

if run_matrix scaffold missing basic >"$TEST_TMP/error" 2>&1; then
  fail 'unknown framework should fail'
fi
assert_contains "$(<"$TEST_TMP/error")" 'unknown framework: missing' 'invalid combinations should explain the failure'

mkdir -p "$PROJECT_ROOT/apps/showcase-alpha-broken"
if run_matrix scaffold alpha broken >"$TEST_TMP/error" 2>&1; then
  fail 'unknown profile should fail'
fi
assert_contains "$(<"$TEST_TMP/error")" 'unknown profile: alpha/broken' 'unknown profiles should explain the failure'

printf 'scaffold matrix tests passed (%d assertions)\n' "$passed"
