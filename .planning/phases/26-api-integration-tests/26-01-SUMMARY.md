---
phase: 26-api-integration-tests
plan: 01
subsystem: api-testing
tags: [integration-tests, testcontainers, stack-crud, soft-delete]

dependency_graph:
  requires: []
  provides:
    - integration-test-infrastructure
    - testcontainers-lifecycle
    - stack-crud-tests
    - soft-delete-tests
  affects:
    - api/tests/integration

tech_stack:
  added:
    - testcontainers-go@0.40.0
    - testcontainers-postgres@0.40.0
    - testify@1.11.1
  patterns:
    - TestMain lifecycle management
    - httptest.NewRecorder for HTTP simulation
    - Envelope assertion pattern

key_files:
  created:
    - api/tests/integration/helpers_test.go
    - api/tests/integration/stack_test.go
  modified:
    - api/go.mod
    - api/go.sum

decisions: []

metrics:
  duration: 172
  tasks_completed: 2
  files_created: 2
  files_modified: 2
  test_coverage:
    - TestStackCreate
    - TestStackCreateDuplicate
    - TestStackList
    - TestStackGet
    - TestStackGetNotFound
    - TestStackUpdate
    - TestStackDelete
    - TestStackSoftDeleteExcludedFromList
    - TestStackSoftDeleteInTrash
    - TestStackRestore
  completed_at: 2026-02-12
---

# Phase 26 Plan 01: Integration Test Infrastructure Summary

**One-liner:** Self-contained integration test scaffold with testcontainers, migration runner, and 10 stack CRUD/soft-delete tests verifying envelope format.

## What Was Built

### Test Infrastructure (helpers_test.go)

**TestMain lifecycle:**
- Postgres 16-alpine via testcontainers with database "devarch_test"
- Automatic migration runner (applies *.up.sql files from api/migrations/)
- Connection string exposed via `testConnStr` for lock tests
- Cleanup on exit (DB close, container terminate)

**Helper functions:**
- `migrateUp(db, dir)` — inline migration runner with schema_migrations tracking
- `truncateAll(t, db)` — truncates all 30+ tables with CASCADE (safe test isolation)
- `setupRouter(t)` — creates chi router with stub container client and ModeDevOpen
- `createStackViaDB(t, db, name)` — direct DB stack insertion for test setup
- `seedServiceTemplate(t, db, serviceName)` — ensures category + service exist for FK constraints
- `createInstanceViaDB(t, db, stackID, serviceName)` — creates instance with template reference

### Stack Integration Tests (stack_test.go)

**10 test functions covering:**

1. **TestStackCreate** — POST creates stack, returns 201, envelope has data, DB row exists
2. **TestStackCreateDuplicate** — duplicate name returns error (409/400/500), envelope has error object
3. **TestStackList** — GET returns 200, envelope data is array with 2 stacks
4. **TestStackGet** — GET /{name} returns 200, envelope data has correct name
5. **TestStackGetNotFound** — nonexistent stack returns 404, envelope has error.code/message
6. **TestStackUpdate** — PUT updates description, DB reflects change
7. **TestStackDelete** — DELETE soft-deletes stack (deleted_at non-null)
8. **TestStackSoftDeleteExcludedFromList** — deleted stack absent from GET /stacks
9. **TestStackSoftDeleteInTrash** — deleted stack present in GET /stacks/trash
10. **TestStackRestore** — POST /trash/{name}/restore clears deleted_at, stack back in list

**Test patterns:**
- `doRequest(t, router, method, path, body)` helper for HTTP simulation
- Envelope assertions via `map[string]interface{}` decoding
- DB verification after HTTP operations
- No t.Parallel() — truncation is not transaction-isolated

## Deviations from Plan

None — plan executed exactly as written.

## Verification

**Compile check:**
```bash
cd api && go build -tags=integration ./tests/integration/...
```
✅ Compiles successfully

**Runtime tests (requires Docker):**
Tests designed to run via:
```bash
cd api && go test -tags=integration -v -count=1 ./tests/integration/... -run TestStack
```

All 10 TestStack* functions verify:
- HTTP status codes
- Envelope structure (data vs error)
- DB state consistency

## Dependencies Installed

- `testcontainers-go@0.40.0` — container lifecycle management
- `testcontainers-go/modules/postgres@0.40.0` — Postgres-specific testcontainer
- `testify@1.11.1` — require assertions

## Files Modified

**api/go.mod, api/go.sum:**
- Added testcontainers + testify dependencies
- ~40 transitive dependencies (Docker client, OpenTelemetry, etc.)

**api/tests/integration/helpers_test.go (272 lines):**
- Build tag: `//go:build integration`
- Package: `integration_test` (external test package)
- TestMain with Postgres lifecycle
- 6 helper functions

**api/tests/integration/stack_test.go (255 lines):**
- Build tag: `//go:build integration`
- 10 test functions
- 1 HTTP helper function

## Test Coverage

**Stack operations covered:**
- CRUD: Create, Read (list/get), Update, Delete
- Soft-delete lifecycle: delete → trash list → restore
- Error cases: duplicate name, not found

**Not covered (future plans):**
- Stack enable/disable
- Stack clone/rename
- Stack lifecycle operations (start/stop/restart)
- Network management
- Instance operations
- Export/import

## Next Steps

Phase 26 Plan 02 will add instance CRUD tests, override tests, and effective-config tests using this infrastructure.

## Self-Check

**Created files exist:**
```bash
[ -f "api/tests/integration/helpers_test.go" ] && echo "FOUND: helpers_test.go" || echo "MISSING"
[ -f "api/tests/integration/stack_test.go" ] && echo "FOUND: stack_test.go" || echo "MISSING"
```
✅ FOUND: helpers_test.go
✅ FOUND: stack_test.go

**Commits exist:**
```bash
git log --oneline --all | grep -q "68e32f6" && echo "FOUND: 68e32f6" || echo "MISSING"
git log --oneline --all | grep -q "379badd" && echo "FOUND: 379badd" || echo "MISSING"
```
✅ FOUND: 68e32f6
✅ FOUND: 379badd

## Self-Check: PASSED
