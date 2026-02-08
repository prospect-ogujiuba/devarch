---
phase: 07-export-import-bootstrap
plan: 04
subsystem: export-import-ui-cli
tags: [ui, cli, export, import, bootstrap, diagnostics]
dependency_graph:
  requires: [07-01, 07-02, 07-03]
  provides: [dashboard-export-ui, dashboard-import-ui, cli-init, cli-doctor]
  affects: [dashboard-stacks, cli-devarch]
tech_stack:
  added: []
  patterns: [blob-download, multipart-upload, bash-scripting, curl-api-client]
key_files:
  created:
    - scripts/devarch-init.sh
    - scripts/devarch-doctor.sh
  modified:
    - dashboard/src/features/stacks/queries.ts
    - dashboard/src/routes/stacks/$name.tsx
    - dashboard/src/types/api.ts
    - scripts/devarch
decisions: []
metrics:
  duration: 275s
  tasks: 2
  files: 6
  completed: 2026-02-08
---

# Phase 7 Plan 4: Dashboard Export/Import UI + CLI Bootstrap Summary

Dashboard export/import buttons with toast feedback and devarch init/doctor bash scripts for teammate onboarding.

## What Was Built

### Dashboard Export/Import UI
- **ImportResult type** in api.ts for import response structure (stack_name, stack_created, created/updated arrays, errors)
- **useExportStack mutation** triggers blob download as `{name}-devarch.yml` via responseType: 'blob'
- **useImportStack mutation** accepts File, sends multipart/form-data, shows toast with "Created: N, Updated: N" summary
- **Export/Import buttons** added to stack detail dropdown menu (alongside Edit, Clone, Rename, Delete)
- **Hidden file input** with ref handler for import file picker (.yml/.yaml filter)
- Query invalidation on import success (stacks list + specific stack)

### CLI Bootstrap (devarch init)
- **Import step** POSTs yml to /stacks/import, parses ImportResult with jq, shows created/updated counts
- **Pull step** extracts unique images from yml, pulls with detected runtime (podman/docker)
- **Apply step** POSTs to /stacks/{name}/apply with empty token (triggers plan + apply)
- Environment vars: DEVARCH_API_URL (default: http://localhost:8550), DEVARCH_API_KEY (default: test)
- Error handling: exits 1 on import failure, continues with warnings on pull/apply failures

### CLI Diagnostics (devarch doctor)
- **Check 1: Runtime** - detects podman/docker, shows version, tests responsiveness (podman ps / docker ps). FAIL if neither found.
- **Check 2: API** - curls health endpoint. WARN if unreachable (not FAIL - API might start later).
- **Check 3: Disk** - df checks available space. WARN if <1GB.
- **Check 4: Tools** - checks curl, jq, grep presence. WARN if missing.
- **Check 5: Ports** - nc -z checks 5432, 3306, 6379, 8080. WARN if in use.
- Exit code 1 on any FAIL, 0 otherwise

### CLI Integration
- Added `init` and `doctor` subcommands to devarch dispatcher
- Added help text for both commands (usage, examples, environment vars)
- Namespace aliases: `init` (no alias), `doctor|doc`
- Scripts are executable with Unix line endings

## Deviations from Plan

None - plan executed exactly as written.

## Key Decisions

None - followed existing patterns.

## Verification Results

All verification checks passed:
- TypeScript compiles cleanly (npx tsc --noEmit)
- useExportStack and useImportStack hooks exist in queries.ts
- Export/Import buttons visible in stack detail dropdown
- devarch-init.sh syntax valid (bash -n)
- devarch-doctor.sh syntax valid (bash -n)
- init and doctor subcommands registered in devarch CLI
- Both bash scripts are executable

## Files Changed

**Created (2):**
- scripts/devarch-init.sh - Bootstrap workflow (import + pull + apply)
- scripts/devarch-doctor.sh - Diagnostics (Pass/Warn/Fail checks)

**Modified (4):**
- dashboard/src/types/api.ts - Added ImportResult type
- dashboard/src/features/stacks/queries.ts - Added useExportStack and useImportStack mutations
- dashboard/src/routes/stacks/$name.tsx - Added export/import UI (dropdown buttons + file input)
- scripts/devarch - Registered init and doctor subcommands with help text

## Commits

- `4de4f7de` - feat(07-04): dashboard export/import mutations and UI
- `fa81184a` - feat(07-04): CLI bootstrap and diagnostics scripts

## Testing Notes

Manual testing required:
- Dashboard: Export button → downloads {name}-devarch.yml blob
- Dashboard: Import button → file picker → shows toast with counts → refreshes stack/instances
- CLI: `devarch init devarch.yml` → import + pull + apply workflow
- CLI: `devarch doctor` → Pass/Warn/Fail output with correct exit codes

API endpoints `/api/v1/stacks/{name}/export` and `/api/v1/stacks/import` must exist (implemented in 07-02).

## Next Steps

Phase 7 complete (all 4 plans done). Export/import surface is now complete:
- API endpoints (07-02)
- Lockfile generation/validation (07-03)
- Dashboard UI + CLI scripts (07-04)

Ready for Phase 8 (Secrets Management) or Phase 9 (Resource Limits & Plan Validation).

## Self-Check: PASSED

All claimed files exist and all commits are present in git history.
