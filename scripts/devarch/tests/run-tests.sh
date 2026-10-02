#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR=$(cd -P -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)

"$SCRIPT_DIR/dotenv_test.sh"
"$SCRIPT_DIR/env_template_test.sh"
"$SCRIPT_DIR/platform_test.sh"
