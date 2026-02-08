---
phase: 07-export-import-bootstrap
plan: 02
subsystem: export
tags: [import, reconciliation, advisory-lock, validation]
dependency_graph:
  requires:
    - "07-01 (export types)"
    - "Phase 3 (instance overrides)"
    - "Phase 4 (container validation)"
  provides:
    - "POST /stacks/import endpoint"
    - "Importer with DB reconciliation"
  affects:
    - "stacks table"
    - "service_instances table"
    - "instance override tables"
tech_stack:
  added:
    - "gopkg.in/yaml.v3 (decoder)"
  patterns:
    - "Multipart form upload"
    - "DELETE+INSERT override reconciliation"
    - "Advisory transaction lock (pg_try_advisory_xact_lock)"
    - "Fail-fast template validation"
key_files:
  created:
    - "api/internal/export/importer.go"
    - "api/internal/api/handlers/stack_import.go"
  modified:
    - "api/internal/api/routes.go"
decisions:
  - "Import is additive/update only — instances in DB but not in YAML are left untouched"
  - "Secret placeholders (\${SECRET:X}) preserved as literal values during import"
  - "Auto-compute network_name with devarch-{stack}-net if not provided"
  - "Skip devarch.* system labels during import — auto-injected in effective config"
  - "Volume type inferred from source path (bind vs named volume)"
metrics:
  duration: 116
  completed_at: "2026-02-08T20:13:26Z"
---

# Phase 7 Plan 2: Import Domain Logic & HTTP Endpoint Summary

Stack import with create-update reconciliation, template validation, and advisory locking.

## Completion Summary

**Tasks completed:** 2/2
**Files created:** 2 (importer.go, stack_import.go)
**Files modified:** 1 (routes.go)

### Task Commits

| Task | Name                                 | Commit   | Files                               |
| ---- | ------------------------------------ | -------- | ----------------------------------- |
| 1    | Importer with create-update reconciliation | 9932b16e | api/internal/export/importer.go     |
| 2    | Import HTTP handler and route wiring | 808b3e7c | stack_import.go, routes.go          |

## Implementation Details

### Importer Logic (importer.go)

**ImportResult struct:**
- StackName, StackCreated (bool), Created/Updated (instance name lists), Errors

**Import flow (single transaction):**
1. Validate all referenced templates exist FIRST — collect missing names, fail-fast if any missing
2. Check if stack exists by name (WHERE deleted_at IS NULL)
3. If exists: acquire `pg_try_advisory_xact_lock(stackID)`, return 409 if locked, update stack
4. If doesn't exist: INSERT new stack with auto-computed network_name
5. For each instance (keyed by name):
   - Look up template service ID
   - Check if instance exists (by instance_id + stack_id)
   - If exists: UPDATE template + enabled, DELETE all overrides, INSERT fresh from YAML
   - If doesn't exist: INSERT instance, INSERT overrides
   - Container name: `devarch-{stackName}-{instanceName}`
6. Commit transaction

**Override insertion patterns:**
- Ports, volumes, env_vars, labels, domains, healthcheck, dependencies, config_files
- Secret detection: `is_secret = strings.Contains(value, "${SECRET:")`
- Labels: skip devarch.* prefix (system labels)
- Volume type: inferred from source path (bind vs volume)
- Healthcheck: duration parser supports "30s", "5m" formats

### HTTP Handler (stack_import.go)

**ImportStack method:**
- Parse multipart form (10MB max)
- Get file from form field "file"
- Decode YAML into DevArchFile
- Validate version == 1 (return 400 if unsupported)
- Validate stack name is present and DNS-safe (container.ValidateName)
- Call importer.Import()
- Error handling:
  - "not found in catalog" → 400 with result JSON showing missing templates
  - "locked" → 409 conflict
  - Other errors → 500
- Success → 200 JSON with ImportResult

**Route registration:**
- `r.Post("/import", stackHandler.ImportStack)` at `/stacks/import`
- Positioned BEFORE `/{name}` routes to avoid chi parameter conflicts
- Same pattern as trash routes

## Verification Results

1. ✅ Full binary compiles (`go build ./cmd/server`)
2. ✅ ImportResult type exists in importer.go
3. ✅ Fail-fast error message: "Template X not found in catalog. Import the template first."
4. ✅ Advisory lock pattern: `pg_try_advisory_xact_lock`
5. ✅ Route wired at POST /stacks/import (before /{name})

## Success Criteria Met

- ✅ Importer validates templates exist before any mutations
- ✅ Advisory lock acquired for existing stacks during import
- ✅ Override insertion follows DELETE+INSERT pattern from existing codebase
- ✅ Import route registered at POST /stacks/import (before /{name} routes)
- ✅ Stack auto-created when name doesn't exist

## Deviations from Plan

None — plan executed exactly as written.

## Self-Check: PASSED

All created files exist:
- ✅ api/internal/export/importer.go
- ✅ api/internal/api/handlers/stack_import.go

All commits exist:
- ✅ 9932b16e (Task 1: Importer)
- ✅ 808b3e7c (Task 2: HTTP handler)
