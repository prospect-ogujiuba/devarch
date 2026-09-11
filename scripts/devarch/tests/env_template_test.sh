#!/usr/bin/env bash
set -Eeuo pipefail

SCRIPT_DIR=$(cd -P -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)
# shellcheck source=test-helper.sh
source "$SCRIPT_DIR/test-helper.sh"
PROJECT_ROOT=$(cd -P -- "$DEVARCH_DIR/../.." && pwd -P)
TEMPLATE="$PROJECT_ROOT/.env.example"

extract_allowlist() {
    local file=$1 declaration=$2
    awk -v declaration="$declaration" '
        $0 ~ "^" declaration "=\\(" { in_list = 1; next }
        in_list && /^[[:space:]]*\)/ { exit }
        in_list {
            sub(/#.*/, "")
            for (i = 1; i <= NF; i++) print $i
        }
    ' "$file"
}

allowed_keys=$TEST_TMP/allowed-keys
{
    extract_allowlist "$PROJECT_ROOT/scripts/wordpress/bootstrap.sh" WORDPRESS_DOTENV_KEYS
    extract_allowlist "$PROJECT_ROOT/scripts/laravel/bootstrap.sh" LARAVEL_DOTENV_KEYS
} | sort -u >"$allowed_keys"
[[ -s $allowed_keys ]] || fail "bootstrap dotenv allowlists were not readable"

template_keys=$TEST_TMP/template-keys
awk '
    /^[[:space:]]*(#|$)/ { next }
    /^[A-Z][A-Z0-9_]*=/ {
        key = $0
        sub(/=.*/, "", key)
        print key
        next
    }
    { exit 1 }
' "$TEMPLATE" | sort -u >"$template_keys" || fail "template contains an invalid active line"

while IFS= read -r key; do
    grep -Fxq -- "$key" "$allowed_keys" || fail "template key is not accepted by a bootstrap: $key"
done <"$template_keys"

if grep -Fxq -- DEVARCH_ENV_FILE "$template_keys"; then
    fail "host control remains active in template: DEVARCH_ENV_FILE"
fi

unexpected_allowed_key=$(comm -23 "$allowed_keys" "$template_keys" | head -n 1)
if [[ -n $unexpected_allowed_key ]]; then
    fail "bootstrap accepts a key missing from the canonical template: $unexpected_allowed_key"
fi

while IFS='=' read -r key value; do
    [[ $key =~ (PASSWORD|TOKEN|API_KEY)$ ]] || continue
    if [[ -n $value && $value != *INVALID* && $value != *REPLACE_ME* ]]; then
        fail "template credential has a usable-looking value: $key"
    fi
done < <(grep -E '^[A-Z][A-Z0-9_]*=' "$TEMPLATE")

printf 'ok - dotenv template matches bootstrap allowlists\n'
