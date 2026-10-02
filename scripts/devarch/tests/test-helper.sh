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
