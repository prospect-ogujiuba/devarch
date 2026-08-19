#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR=$(cd -P -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)
# shellcheck source-path=SCRIPTDIR
# shellcheck source=test-helper.sh
source "$SCRIPT_DIR/test-helper.sh"
# shellcheck disable=SC1091 # Resolved from the test helper's physical path.
source "$DEVARCH_DIR/lib/common.sh"

caller_tmp_base=$TEST_TMP/caller-owned-base
caller_tmp_sentinel=$caller_tmp_base/sentinel
nested_tmp_path=$TEST_TMP/nested-test-tmp
mkdir -p "$caller_tmp_base"
printf 'keep' >"$caller_tmp_sentinel"
TEST_TMP=$caller_tmp_base TEST_HELPER=$SCRIPT_DIR/test-helper.sh NESTED_TMP_PATH=$nested_tmp_path \
    bash -c 'source "$TEST_HELPER"; printf "%s" "$TEST_TMP" >"$NESTED_TMP_PATH"'
[[ -f $caller_tmp_sentinel ]] || fail "test helper removed caller-provided TEST_TMP base"
nested_tmp=$(<"$nested_tmp_path")
[[ $nested_tmp == "$caller_tmp_base"/* ]] || fail "test helper did not create a child of caller-provided TEST_TMP"
[[ ! -e $nested_tmp ]] || fail "test helper did not clean its owned temporary directory"

new_fixture_repo repository-root
original_pwd=$PWD
cd "$TEST_TMP"
assert_eq "$FIXTURE_REPO" "$(devarch_repo_root)" "repository resolution must not depend on cwd"
cd "$original_pwd"

error_file=$TEST_TMP/common-error
original_path=$PATH
# shellcheck disable=SC2123 # Intentionally test command discovery with an empty PATH.
PATH=$TEST_TMP/no-executables
if devarch_require_podman 2>"$error_file"; then
    fail "missing Podman unexpectedly passed"
fi
PATH=$original_path
missing_error=$(<"$error_file")
assert_contains "$missing_error" "podman is required" "missing Podman guidance"
if [[ $missing_error == *Docker* || $missing_error == *docker* ]]; then
    fail "missing Podman guidance invented a Docker fallback"
fi

install_recording_executables
PATH=$RECORD_BIN:$original_path
export PATH
compose_stdout=$TEST_TMP/compose-stdout
export RECORD_COMPOSE_HELP_STATUS=47 RECORD_STDOUT='provider stdout guidance' \
    RECORD_STDERR='provider stderr guidance'
if devarch_require_compose >"$compose_stdout" 2>"$error_file"; then
    fail "failing Compose provider unexpectedly passed"
fi
assert_contains "$(<"$compose_stdout")" "provider stdout guidance" \
    "native provider stdout guidance must remain visible"
assert_contains "$(<"$error_file")" "provider stderr guidance" \
    "native provider stderr guidance must remain visible"
assert_contains "$(<"$DEVARCH_RECORD_FILE")" "podman compose --help" "Compose capability check must be native"
unset RECORD_COMPOSE_HELP_STATUS RECORD_STDOUT RECORD_STDERR

export RECORD_STDOUT='successful help stdout' RECORD_STDERR='successful help stderr'
devarch_require_compose >"$compose_stdout" 2>"$error_file"
assert_eq "" "$(<"$compose_stdout")" "successful Compose capability stdout must stay quiet"
assert_eq "" "$(<"$error_file")" "successful Compose capability stderr must stay quiet"
unset RECORD_STDOUT RECORD_STDERR

: >"$DEVARCH_RECORD_FILE"
export RECORD_STDOUT='native stdout' RECORD_STDERR='native stderr' RECORD_EXIT_STATUS=23
set +e
stdout=$(devarch_run -- psql 'space argument' --format=json 2>"$error_file")
status=$?
set -e
assert_eq "23" "$status" "devarch_run must preserve exit status"
assert_eq "native stdout" "$stdout" "devarch_run must preserve stdout"
run_stderr=$(<"$error_file")
assert_contains "$run_stderr" "+ psql space\\ argument --format=json" "devarch_run must log exact argv"
assert_contains "$run_stderr" "native stderr" "devarch_run must preserve stderr"
assert_contains "$(<"$DEVARCH_RECORD_FILE")" "psql space\\ argument --format=json" "recorder must receive exact argv"
unset RECORD_STDOUT RECORD_STDERR RECORD_EXIT_STATUS

printf 'stdin payload' | devarch_run -- psql >/dev/null 2>"$error_file"
assert_eq "stdin payload" "$(<"$DEVARCH_RECORD_STDIN")" "devarch_run must preserve stdin"

before_pid=$TEST_TMP/before-pid
after_pid=$TEST_TMP/after-pid
cat >"$TEST_TMP/exec-check.sh" <<'EXEC_CHECK'
#!/bin/bash
source "$COMMON_LIBRARY"
printf '%s' "$BASHPID" >"$BEFORE_PID"
devarch_run --exec -- bash -c 'printf "%s" "$BASHPID" >"$AFTER_PID"'
EXEC_CHECK
chmod +x "$TEST_TMP/exec-check.sh"
export COMMON_LIBRARY=$DEVARCH_DIR/lib/common.sh BEFORE_PID=$before_pid AFTER_PID=$after_pid
"$TEST_TMP/exec-check.sh" 2>"$error_file"
assert_eq "$(<"$before_pid")" "$(<"$after_pid")" "--exec must replace the wrapper process"

child_pid_file=$TEST_TMP/run-child-pid
cat >"$TEST_TMP/signal-check.sh" <<'SIGNAL_CHECK'
#!/bin/bash
source "$COMMON_LIBRARY"
devarch_run -- bash -c 'printf "%s" "$BASHPID" >"$CHILD_PID_FILE"; exec sleep 30'
SIGNAL_CHECK
chmod +x "$TEST_TMP/signal-check.sh"
export CHILD_PID_FILE=$child_pid_file
"$TEST_TMP/signal-check.sh" 2>"$error_file" &
wrapper_pid=$!
for ((attempt = 0; attempt < 100; attempt++)); do
    [[ -s $child_pid_file ]] && break
    sleep 0.01
done
[[ -s $child_pid_file ]] || {
    kill -TERM "$wrapper_pid" 2>/dev/null || true
    fail "signal test child did not start"
}
child_pid=$(<"$child_pid_file")
kill -TERM "$wrapper_pid"
set +e
wait "$wrapper_pid"
wrapper_status=$?
set -e
assert_eq "143" "$wrapper_status" "signaled devarch_run wrapper must preserve SIGTERM status"
for ((attempt = 0; attempt < 100; attempt++)); do
    kill -0 "$child_pid" 2>/dev/null || break
    sleep 0.01
done
if kill -0 "$child_pid" 2>/dev/null; then
    kill -TERM "$child_pid" 2>/dev/null || true
    fail "terminating devarch_run left its child running"
fi

printf 'ok - repository, capability, and passthrough helpers\n'
