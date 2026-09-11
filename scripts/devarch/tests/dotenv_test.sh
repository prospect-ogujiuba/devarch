#!/usr/bin/env bash
set -Eeuo pipefail

SCRIPT_DIR=$(cd -P -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)
# shellcheck source=test-helper.sh
source "$SCRIPT_DIR/test-helper.sh"
# shellcheck source=../lib/dotenv.sh
source "$DEVARCH_DIR/lib/dotenv.sh"

fixture=$TEST_TMP/values.env
sentinel=$TEST_TMP/payload-ran
export SENTINEL_PATH=$sentinel
cat >"$fixture" <<'EOF'
# supported dotenv forms
TEST_UNQUOTED=plain # trailing comment
TEST_HASH=plain#literal
TEST_SINGLE='single \'quote\' and \\ path' # comment
TEST_DOUBLE="quoted \"value\" \\ path \$cash" # comment
TEST_EMPTY=
TEST_DUP=first
UNRELATED=$(touch "$SENTINEL_PATH")
DEVARCH_ENV_FILE=/must/not/be/loaded
export TEST_DUP=last
TEST_EXPORTED=file-value
EOF

export TEST_EXPORTED=inherited
export DEVARCH_DOTENV_VALUE=ambient-scratch
unset TEST_UNQUOTED TEST_HASH TEST_SINGLE TEST_DOUBLE TEST_EMPTY TEST_DUP UNRELATED
original_env_file=${DEVARCH_ENV_FILE-}
devarch_load_dotenv "$fixture" \
    TEST_UNQUOTED TEST_HASH TEST_SINGLE TEST_DOUBLE TEST_EMPTY TEST_DUP TEST_EXPORTED

assert_eq plain "$TEST_UNQUOTED" "unquoted value and trailing comment"
assert_eq 'plain#literal' "$TEST_HASH" "literal hash in unquoted value"
assert_eq "single 'quote' and \\ path" "$TEST_SINGLE" "single-quoted escapes"
assert_eq 'quoted "value" \ path $cash' "$TEST_DOUBLE" "double-quoted escapes"
assert_eq '' "$TEST_EMPTY" "empty assignment"
assert_eq last "$TEST_DUP" "last recognized assignment wins"
assert_eq file-value "$TEST_EXPORTED" "file assignment overrides inherited value"
assert_eq ambient-scratch "$DEVARCH_DOTENV_VALUE" "loader scratch state does not alter or expose caller state"
[[ ! -e $sentinel ]] || fail "dotenv command substitution executed"
[[ ! -v UNRELATED ]] || fail "unrecognized key was assigned"
assert_eq "$original_env_file" "${DEVARCH_ENV_FILE-}" "dotenv cannot select its own path"
if env | grep -q '^TEST_UNQUOTED='; then
    fail "new dotenv value was automatically exported"
fi
assert_contains "$(env)" 'TEST_EXPORTED=file-value' "inherited export attribute is preserved"

malformed_unrecognized=$TEST_TMP/malformed-unrecognized.env
printf 'UNRELATED="unterminated\nUNRELATED_CONTROL=bad\tvalue\nUNRELATED_NUL=bad\0value\nTEST_UNQUOTED=accepted\n' >"$malformed_unrecognized"
unset TEST_UNQUOTED
devarch_load_dotenv "$malformed_unrecognized" TEST_UNQUOTED
assert_eq accepted "$TEST_UNQUOTED" "malformed unrecognized lines are inert"

assert_secret_safe_failure() {
    local file=$1 key=$2 forbidden=$3 error=$TEST_TMP/error
    if (devarch_load_dotenv "$file" "$key") 2>"$error"; then
        fail "malformed recognized assignment unexpectedly succeeded: $key"
    fi
    if grep -Fq -- "$forbidden" "$error"; then
        fail "dotenv diagnostic reflected value for $key"
    fi
    assert_contains "$(<"$error")" "$key" "dotenv diagnostic identifies recognized key"
}

printf 'TEST_BAD="do-not-print\n' >"$TEST_TMP/bad-quote.env"
assert_secret_safe_failure "$TEST_TMP/bad-quote.env" TEST_BAD do-not-print
printf 'TEST_CONTROL=do-not-print\tvalue\n' >"$TEST_TMP/control.env"
assert_secret_safe_failure "$TEST_TMP/control.env" TEST_CONTROL do-not-print
printf 'TEST_NUL=do-not-print\0value\n' >"$TEST_TMP/nul.env"
assert_secret_safe_failure "$TEST_TMP/nul.env" TEST_NUL do-not-print

if (devarch_load_dotenv "$fixture" 'BAD-KEY') >/dev/null 2>&1; then
    fail "invalid allowlist key unexpectedly succeeded"
fi

printf 'ok - secure allowlisted dotenv parsing\n'
