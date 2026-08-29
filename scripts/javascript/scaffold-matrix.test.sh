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
attempt_dir="${MATRIX_ATTEMPT_DIR:?}"
mkdir -p "$attempt_dir"
attempt_file="$attempt_dir/$app"
attempt=0
[[ ! -f "$attempt_file" ]] || attempt=$(<"$attempt_file")
((attempt += 1))
printf '%d\n' "$attempt" > "$attempt_file"
printf 'upstream output for %s attempt %d\n' "$app" "$attempt"
if [[ ${MATRIX_ALWAYS_FAIL_APP:-} == "$app" ]]; then
  printf 'persistent fake failure for %s\n' "$app" >&2
  exit 9
fi
if [[ ${MATRIX_FAIL_ONCE_APP:-} == "$app" && $attempt -eq 1 ]]; then
  printf 'transient fake failure for %s\n' "$app" >&2
  exit 8
fi
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
  MATRIX_ATTEMPT_DIR="$TEST_TMP/attempts" \
  DEVARCH_MATRIX_LOG_DIR="$TEST_TMP/logs" \
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

all_output=$(MATRIX_FAIL_ONCE_APP=showcase-alpha-two run_matrix scaffold-all 2>&1)
[[ -f "$PROJECT_ROOT/apps/showcase-alpha-two/package.json" ]] || fail 'scaffold-all should create alpha/two after a transient failure'
pass
[[ -f "$PROJECT_ROOT/apps/showcase-beta-basic/package.json" ]] || fail 'scaffold-all should continue to beta/basic'
pass
assert_contains "$all_output" '[2/3] alpha/two — attempt 1/2' 'scaffold-all should show numbered progress'
assert_contains "$all_output" 'attempt 1 failed, retrying' 'scaffold-all should report automatic retries'
assert_contains "$all_output" 'total=3 created=2 skipped=1 failed=0' 'scaffold-all should summarize sequential results'
[[ "$(<"$TEST_TMP/attempts/showcase-alpha-two")" == 2 ]] || fail 'transient failure should be attempted twice'
pass
run_log=$(find "$TEST_TMP/logs" -type f -name '*-scaffold-all*.log' | head -n1)
[[ -n "$run_log" ]] || fail 'scaffold-all should create a run log'
pass
assert_contains "$(<"$run_log")" 'upstream output for showcase-alpha-two attempt 1' 'run log should capture upstream output'

rm -rf "$PROJECT_ROOT/apps/showcase-alpha-two" "$PROJECT_ROOT/apps/showcase-beta-basic" "$TEST_TMP/attempts"
: > "$CALL_LOG"
if MATRIX_ALWAYS_FAIL_APP=showcase-alpha-two run_matrix scaffold-all >"$TEST_TMP/failure-output" 2>&1; then
  fail 'persistent profile failure should produce a nonzero final status'
fi
failure_output=$(<"$TEST_TMP/failure-output")
[[ -f "$PROJECT_ROOT/apps/showcase-beta-basic/package.json" ]] || fail 'persistent failure should not prevent later profiles from being attempted'
pass
assert_contains "$failure_output" '[2/3] alpha/two — failed; log tail follows' 'persistent failure should be explicit'
assert_contains "$failure_output" 'total=3 created=1 skipped=1 failed=1' 'failure summary should include every outcome'
assert_contains "$failure_output" 'failed profiles: alpha/two' 'failure summary should identify retry targets'

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
