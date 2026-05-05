# 06 — Retire Scripts into Compatibility Shims

## Goal

Stop scripts from being the source of truth after Go workflow parity exists.

### Task 13: Replace script implementations with `devarch` shims after parity

**Objective:** Convert selected shell scripts into compatibility wrappers around the new CLI.

**Files:**
- Modify: `scripts/devarch-doctor.sh`
- Modify: `scripts/socket-manager.sh`
- Modify: `scripts/runtime-switcher.sh`
- Modify: `scripts/service-manager.sh`
- Create: `docs/devarch-v2/script-migration.md`

**Example shim pattern:**

```bash
#!/usr/bin/env bash
set -euo pipefail
exec devarch doctor "$@"
```

**Validation:**

```bash
go test ./...
bash -n scripts/devarch-doctor.sh scripts/socket-manager.sh scripts/runtime-switcher.sh scripts/service-manager.sh
```
