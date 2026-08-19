#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR=$(cd -P -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)

"$SCRIPT_DIR/catalog_test.sh"
"$SCRIPT_DIR/common_test.sh"
