#!/usr/bin/env bash

set -o pipefail

TEST_DIR=$(cd -P -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)
DEVARCH_DIR=$(cd -P -- "$TEST_DIR/.." && pwd -P)
export DEVARCH_DIR
TEST_TMP_BASE=${TEST_TMP:-${TMPDIR:-/tmp}}
[[ -d $TEST_TMP_BASE ]] || {
    printf 'test temp base does not exist: %s\n' "$TEST_TMP_BASE" >&2
    return 1
}
TEST_TMP=$(mktemp -d "$TEST_TMP_BASE/devarch-foundation.XXXXXX") || return
TEST_TMP_OWNED=$TEST_TMP

cleanup_test_tmp() {
    rm -rf -- "$TEST_TMP_OWNED"
}
trap cleanup_test_tmp EXIT

fail() {
    printf 'not ok - %s\n' "$*" >&2
    exit 1
}

assert_eq() {
    local expected=$1 actual=$2 message=${3:-values differ}
    [[ $actual == "$expected" ]] || {
        printf 'expected: %q\nactual:   %q\n' "$expected" "$actual" >&2
        fail "$message"
    }
}

assert_contains() {
    local haystack=$1 needle=$2 message=${3:-missing expected text}
    [[ $haystack == *"$needle"* ]] || fail "$message: $needle"
}

new_fixture_repo() {
    local name=${1:-repository}
    FIXTURE_REPO=$TEST_TMP/$name
    mkdir -p "$FIXTURE_REPO/services-library"
    export DEVARCH_REPO_ROOT=$FIXTURE_REPO
}

add_fixture_service() {
    local category=$1 name=$2 content
    if (($# >= 3)); then
        content=$3
    else
        content=$'services:\n  app:\n    image: example.invalid/test:latest\n'
    fi
    mkdir -p "$FIXTURE_REPO/services-library/$category/$name"
    printf '%s' "$content" >"$FIXTURE_REPO/services-library/$category/$name/compose.yml"
}

install_recording_executables() {
    RECORD_BIN=$TEST_TMP/bin
    DEVARCH_RECORD_FILE=$TEST_TMP/recorded-argv
    DEVARCH_RECORD_STDIN=$TEST_TMP/recorded-stdin
    mkdir -p "$RECORD_BIN"
    : >"$DEVARCH_RECORD_FILE"
    export DEVARCH_RECORD_FILE DEVARCH_RECORD_STDIN

    cat >"$RECORD_BIN/recorder" <<'RECORDER'
#!/bin/bash
{
    printf '%q' "${0##*/}"
    for argument in "$@"; do
        printf ' %q' "$argument"
    done
    printf '\n'
} >>"$DEVARCH_RECORD_FILE"

if [[ ! -t 0 ]]; then
    cat >"$DEVARCH_RECORD_STDIN"
fi
[[ -n ${RECORD_STDOUT:-} ]] && printf '%s' "$RECORD_STDOUT"
[[ -n ${RECORD_STDERR:-} ]] && printf '%s' "$RECORD_STDERR" >&2
if [[ ${0##*/} == podman && ${1:-} == compose && ${2:-} == --help ]]; then
    exit "${RECORD_COMPOSE_HELP_STATUS:-0}"
fi
exit "${RECORD_EXIT_STATUS:-0}"
RECORDER
    chmod +x "$RECORD_BIN/recorder"

    local executable
    for executable in podman psql devarch-launcher mkcert; do
        cp "$RECORD_BIN/recorder" "$RECORD_BIN/$executable"
    done
}
