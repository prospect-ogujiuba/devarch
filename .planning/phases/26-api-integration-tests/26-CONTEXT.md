# Phase 26: API Integration Tests - Context

**Gathered:** 2026-02-12
**Status:** Ready for planning

<domain>
## Phase Boundary

Integration tests covering stack/instance CRUD, soft-delete semantics, plan token staleness detection, and advisory lock conflict prevention. DB-centric validation — container runtime operations are not under test. Requirements: TEST-01, TEST-02.

</domain>

<decisions>
## Implementation Decisions

### Test Framework
- `testify/assert` + `testify/require` — new go.mod dependencies
- Better assertion messages and fail-fast on setup errors vs plain stdlib

### Database Provisioning
- `testcontainers-go` with Postgres module — spins up real Postgres per test suite
- Migrations from `api/migrations/` run at suite startup against test DB
- Self-contained: no dependency on running compose stack
- Same approach in CI — no GitHub Actions `services:` needed

### Test Level
- HTTP-level via `httptest.NewServer` + chi router for CRUD and soft-delete tests (proves full request path: middleware → handler → DB → envelope)
- Direct DB-level function calls for staleness token validation and advisory lock tests (pure DB concerns, avoids container runtime coupling)

### Container Runtime
- Stub/no-op container client for HTTP-level tests — paths that call container ops won't be exercised
- `container.Client` is a struct: create with nil/disconnected socket or introduce minimal test boundary
- No real Docker/Podman needed during test execution

### Test Isolation
- Table truncation between tests via `truncateAll(t, db)` helper
- Not transaction rollback — advisory lock tests need separate DB connections (locks are session-scoped)
- Testcontainers provides isolated DB instance, so truncation is fast

### Build Tags
- `//go:build integration` on all integration test files
- Default `go test ./...` skips them
- Explicit `go test -tags=integration ./tests/integration/...` to run

### File Organization
- `api/tests/integration/` directory, package `integration_test`
- Files: `stack_test.go`, `instance_test.go`, `staleness_test.go`, `lock_test.go`, `helpers_test.go`
- Separate from internal packages — integration tests cross package boundaries

### CI Integration
- New GitHub Actions workflow `.github/workflows/integration-tests.yml`
- Triggered on PR changes to `api/`
- Testcontainers handles Postgres provisioning in CI (self-contained)

### Test Data Setup
- Go helper functions in `helpers_test.go`: `createStack(t, db, name)`, `createInstance(t, db, stackID, serviceName)`
- Return created entity IDs for assertions
- Each test sets up exactly what it needs — no fixtures or seed files

### Assertion Scope
- Assert HTTP status codes + response body (envelope format) + DB state via direct queries
- Example: POST stack → assert 201 + `{"data": {...}}` → SELECT confirms row exists

### Staleness Tests
- Test `plan.ValidateToken()` and `plan.GenerateToken()` directly against test DB
- Create stack + instances → generate token → mutate state → assert validation fails

### Advisory Lock Tests
- Two separate `*sql.DB` connections to test DB
- Connection A acquires `pg_try_advisory_lock(stackID)`, connection B attempts same → assert false
- A releases, B acquires → assert true
- Simulates concurrent apply attempts from `orchestration.Service.ApplyPlan()`

### Claude's Discretion
- Exact testcontainers setup/teardown lifecycle
- Whether to wrap router setup in TestMain vs per-test
- Helper function signatures and return types
- Migration runner implementation (reuse cmd/migrate logic or simplified version)

</decisions>

<specifics>
## Specific Ideas

- HTTP-level tests should verify envelope format from Phase 19 (`{"data": ...}` / `{"error": {"code", "message", "details"}}`)
- Soft-delete tests: deleted stacks excluded from List, present in ListTrash, restorable via Restore
- Lock tests model the exact `pg_try_advisory_lock` pattern used in `orchestration/service.go:181` and `export/importer.go:96`

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 26-api-integration-tests*
*Context gathered: 2026-02-12*
