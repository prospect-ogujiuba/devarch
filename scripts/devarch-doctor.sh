#!/usr/bin/env bash
set -euo pipefail

# Compatibility shim: replacement command is `devarch doctor`.
exec devarch doctor "$@"
