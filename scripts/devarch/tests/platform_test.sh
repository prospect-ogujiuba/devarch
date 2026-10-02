#!/usr/bin/env bash
set -Eeuo pipefail

TEST_DIR=$(cd -P -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)
# shellcheck source=../lib/platform.sh
source "$TEST_DIR/../lib/platform.sh"
passed=0
fail() { printf 'FAIL: %s\n' "$*" >&2; exit 1; }
pass() { ((passed += 1)); }
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

[[ $(DEVARCH_BIN=/opt/devarch devarch_bin) == /opt/devarch ]] || fail 'DEVARCH_BIN should win'
pass
mkdir -p "$tmp/bin" && printf '#!/bin/sh\n' >"$tmp/bin/devarch" && chmod +x "$tmp/bin/devarch"
[[ $(env -u DEVARCH_BIN PATH="$tmp/bin:/usr/bin:/bin" bash -c "source '$TEST_DIR/../lib/platform.sh'; devarch_bin") == "$tmp/bin/devarch" ]] ||
    fail 'devarch on PATH should be found'
pass
if env -u DEVARCH_BIN PATH=/nonexistent /bin/bash -c "source '$TEST_DIR/../lib/platform.sh'; devarch_bin" 2>"$tmp/err"; then
    fail 'missing devarch should fail'
fi
grep -q 'go install' "$tmp/err" || fail 'missing devarch should explain how to install it'
pass

# No descriptor: silent no-op even under set -e.
( unset DEVARCH_PROGRESS_FD; devarch_progress step start 'hello' ) || fail 'progress without a reader should succeed'
pass
# Closed descriptor: still a no-op.
( DEVARCH_PROGRESS_FD=9 devarch_progress step start 'hello' ) 2>/dev/null || fail 'progress to a closed descriptor should succeed'
pass

DEVARCH_PROGRESS_FD=3 devarch_progress install done $'quote " slash \\ tab\tnewline\nend' 3>"$tmp/events"
expected='{"step":"install","state":"done","message":"quote \" slash \\ tab\tnewline\nend"}'
[[ $(cat "$tmp/events") == "$expected" ]] || fail "unexpected event: $(cat "$tmp/events")"
pass
if command -v python3 >/dev/null 2>&1; then
    python3 -c 'import json,sys; json.loads(open(sys.argv[1]).read())' "$tmp/events" || fail 'event is not valid JSON'
    pass
fi

printf 'PASS: %d assertions\n' "$passed"
