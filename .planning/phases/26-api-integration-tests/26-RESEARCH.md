# Phase 26: API Integration Tests - Research

**Researched:** 2026-02-12
**Domain:** Go integration testing, testcontainers-go, testify, httptest
**Confidence:** HIGH

## Summary

Integration testing for Go APIs uses testcontainers-go to spin up real Postgres instances, httptest for HTTP-level testing, and testify for assertions. The standard pattern: testcontainers provides isolated DB per test suite via TestMain, httptest.NewRecorder + httptest.NewRequest test HTTP handlers with chi router, testify/require provides fail-fast assertions, and build tags (`//go:build integration`) separate slow integration tests from fast unit tests.

The codebase already has unit tests using stdlib testing (container package), uses lib/pq for Postgres, and has chi router + custom respond package for envelope responses. Integration tests will verify CRUD endpoints return correct HTTP status codes + envelope format + DB state, without exercising container runtime operations.

**Primary recommendation:** Use testcontainers-go with Postgres module + httptest + testify/require + build tags. Table truncation for isolation (not transaction rollback) because advisory lock tests need separate DB connections.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- **Test Framework:** testify/assert + testify/require — new go.mod dependencies, better assertion messages vs plain stdlib
- **Database Provisioning:** testcontainers-go with Postgres module — spins up real Postgres per test suite, migrations from api/migrations/ run at suite startup, self-contained (no dependency on running compose stack), same approach in CI (no GitHub Actions services: needed)
- **Test Level:** HTTP-level via httptest.NewServer + chi router for CRUD and soft-delete tests (proves full request path: middleware → handler → DB → envelope). Direct DB-level function calls for staleness token validation and advisory lock tests (pure DB concerns, avoids container runtime coupling)
- **Container Runtime:** Stub/no-op container client for HTTP-level tests — paths that call container ops won't be exercised. container.Client is a struct: create with nil/disconnected socket or introduce minimal test boundary. No real Docker/Podman needed during test execution
- **Test Isolation:** Table truncation between tests via truncateAll(t, db) helper. Not transaction rollback — advisory lock tests need separate DB connections (locks are session-scoped). Testcontainers provides isolated DB instance, so truncation is fast
- **Build Tags:** //go:build integration on all integration test files. Default go test ./... skips them. Explicit go test -tags=integration ./tests/integration/... to run
- **File Organization:** api/tests/integration/ directory, package integration_test. Files: stack_test.go, instance_test.go, staleness_test.go, lock_test.go, helpers_test.go. Separate from internal packages — integration tests cross package boundaries
- **CI Integration:** New GitHub Actions workflow .github/workflows/integration-tests.yml. Triggered on PR changes to api/. Testcontainers handles Postgres provisioning in CI (self-contained)
- **Test Data Setup:** Go helper functions in helpers_test.go: createStack(t, db, name), createInstance(t, db, stackID, serviceName). Return created entity IDs for assertions. Each test sets up exactly what it needs — no fixtures or seed files
- **Assertion Scope:** Assert HTTP status codes + response body (envelope format) + DB state via direct queries. Example: POST stack → assert 201 + {"data": {...}} → SELECT confirms row exists
- **Staleness Tests:** Test plan.ValidateToken() and plan.GenerateToken() directly against test DB. Create stack + instances → generate token → mutate state → assert validation fails
- **Advisory Lock Tests:** Two separate *sql.DB connections to test DB. Connection A acquires pg_try_advisory_lock(stackID), connection B attempts same → assert false. A releases, B acquires → assert true. Simulates concurrent apply attempts from orchestration.Service.ApplyPlan()

### Claude's Discretion
- Exact testcontainers setup/teardown lifecycle
- Whether to wrap router setup in TestMain vs per-test
- Helper function signatures and return types
- Migration runner implementation (reuse cmd/migrate logic or simplified version)

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within phase scope
</user_constraints>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| testcontainers-go | v0.36+ | Container lifecycle management | Industry standard for integration tests requiring real dependencies, auto-cleanup, works in CI |
| testcontainers-go/modules/postgres | v0.36+ | Postgres-specific module | Simplifies Postgres setup, exposes WithDatabase/WithUsername/WithPassword, returns connection string |
| testify/require | v1.11.1+ | Fail-fast assertions | 17,733+ importers, stops test on first failure (critical for setup), clearer error messages than stdlib |
| testify/assert | v1.11.1+ | Non-fatal assertions | Same package as require, use for optional checks where you want to see all failures |
| net/http/httptest | stdlib | HTTP handler testing | Standard library tool for mocking HTTP requests/responses, zero dependencies |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| database/sql | stdlib | DB connection management | Create multiple *sql.DB for advisory lock tests, already in use via lib/pq |
| encoding/json | stdlib | Response body parsing | Verify envelope structure after HTTP requests |
| context | stdlib | Request context + testcontainers lifecycle | Pass to postgres.Run(), attach to httptest.NewRequest() |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| testcontainers-go | dockertest | dockertest is older, less maintained, harder API. testcontainers has official modules and cross-language consistency |
| testify | stdlib testing only | Stdlib works but verbose (manual error messages, no fail-fast). testify reduces boilerplate |
| httptest.NewRecorder | httptest.NewServer | NewServer starts real TCP listener (slower), NewRecorder captures response in-memory (faster, sufficient for unit-style integration tests) |
| Table truncation | go-txdb (transaction rollback) | txdb wraps all tests in transactions — breaks advisory lock tests (locks are session-scoped, require separate connections) |

**Installation:**
```bash
cd api
go get github.com/testcontainers/testcontainers-go
go get github.com/testcontainers/testcontainers-go/modules/postgres
go get github.com/stretchr/testify
go mod tidy
```

## Architecture Patterns

### Recommended Project Structure
```
api/
├── tests/
│   └── integration/              # Integration test package
│       ├── helpers_test.go       # DB setup, fixtures, assertions
│       ├── stack_test.go         # Stack CRUD + soft-delete
│       ├── instance_test.go      # Instance CRUD
│       ├── staleness_test.go     # plan.ValidateToken() tests
│       └── lock_test.go          # pg_try_advisory_lock tests
├── internal/
│   ├── api/
│   │   ├── routes.go            # Router factory (testable)
│   │   └── handlers/            # HTTP handlers
│   ├── plan/
│   │   └── staleness.go         # ValidateToken, GenerateToken (testable)
│   └── orchestration/
│       └── service.go           # ApplyPlan uses advisory locks
└── cmd/migrate/main.go          # Migration logic (reusable)
```

### Pattern 1: TestMain with Testcontainers Lifecycle
**What:** TestMain runs once per package, sets up shared Postgres container, runs all tests, tears down
**When to use:** When multiple tests share same DB schema (fast startup amortized across many tests)
**Example:**
```go
// Source: https://golang.testcontainers.org/modules/postgres/
var testDB *sql.DB

func TestMain(m *testing.M) {
    ctx := context.Background()

    // Start Postgres container
    postgresContainer, err := postgres.Run(ctx,
        "postgres:16-alpine",
        postgres.WithDatabase("devarch_test"),
        postgres.WithUsername("test"),
        postgres.WithPassword("test"),
    )
    if err != nil {
        log.Fatalf("failed to start container: %s", err)
    }
    defer testcontainers.TerminateContainer(postgresContainer)

    // Get connection string
    connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
    if err != nil {
        log.Fatalf("failed to get connection string: %s", err)
    }

    // Open DB connection
    testDB, err = sql.Open("postgres", connStr)
    if err != nil {
        log.Fatalf("failed to connect: %s", err)
    }
    defer testDB.Close()

    // Run migrations (reuse cmd/migrate logic or simplified version)
    if err := runMigrations(testDB); err != nil {
        log.Fatalf("failed to run migrations: %s", err)
    }

    // Run tests
    code := m.Run()
    os.Exit(code)
}
```

### Pattern 2: HTTP-Level Testing with httptest + chi
**What:** Test full HTTP request → chi router → handler → DB → envelope response
**When to use:** For stack/instance CRUD endpoints, soft-delete semantics
**Example:**
```go
// Source: https://www.newline.co/@kchan/testing-a-go-and-chi-restful-api-route-handlers-part-1--6b105194
func TestStackCreate(t *testing.T) {
    // Arrange: truncate tables + create stub container client
    truncateAll(t, testDB)
    stubClient := &container.Client{} // Minimal no-op client
    router := api.NewRouter(testDB, stubClient, nil, nil, nil, nil, nil, nil, nil, security.Disabled)

    reqBody := `{"name":"test-stack","description":"Test description"}`
    req := httptest.NewRequest(http.MethodPost, "/api/v1/stacks", strings.NewReader(reqBody))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()

    // Act
    router.ServeHTTP(w, req)

    // Assert: HTTP status + envelope format + DB state
    require.Equal(t, http.StatusCreated, w.Code)

    var envelope respond.SuccessEnvelope
    err := json.NewDecoder(w.Body).Decode(&envelope)
    require.NoError(t, err)
    require.NotNil(t, envelope.Data)

    // Verify DB insertion
    var name string
    err = testDB.QueryRow("SELECT name FROM stacks WHERE name = $1 AND deleted_at IS NULL", "test-stack").Scan(&name)
    require.NoError(t, err)
    require.Equal(t, "test-stack", name)
}
```

### Pattern 3: Direct DB Function Testing (Staleness)
**What:** Test plan.ValidateToken() and plan.GenerateToken() without HTTP layer
**When to use:** Pure DB logic that doesn't need HTTP context
**Example:**
```go
func TestStalenessDetection(t *testing.T) {
    truncateAll(t, testDB)

    // Arrange: create stack + instance
    stackID := createStack(t, testDB, "test-stack")
    instID := createInstance(t, testDB, stackID, "postgres")

    // Generate token for current state
    var stackUpdatedAt time.Time
    err := testDB.QueryRow("SELECT updated_at FROM stacks WHERE id = $1", stackID).Scan(&stackUpdatedAt)
    require.NoError(t, err)

    instances := []plan.InstanceTimestamp{
        {InstanceID: instID, UpdatedAt: time.Now()},
    }
    token := plan.GenerateToken(stackUpdatedAt, instances)

    // Act: mutate state (update instance)
    _, err = testDB.Exec("UPDATE service_instances SET updated_at = NOW() WHERE instance_id = $1", instID)
    require.NoError(t, err)

    // Assert: token validation fails
    err = plan.ValidateToken(testDB, stackID, token)
    require.ErrorIs(t, err, plan.ErrStalePlan)
}
```

### Pattern 4: Advisory Lock Testing with Multiple Connections
**What:** Test pg_try_advisory_lock behavior with concurrent attempts
**When to use:** Verify orchestration.Service.ApplyPlan() lock conflict detection
**Example:**
```go
// Source: https://engineering.qubecinema.com/2019/08/26/unlocking-advisory-locks.html
func TestAdvisoryLockConflict(t *testing.T) {
    truncateAll(t, testDB)
    stackID := createStack(t, testDB, "test-stack")

    // Open second connection (locks are per-session)
    connStr, _ := postgresContainer.ConnectionString(context.Background(), "sslmode=disable")
    db2, err := sql.Open("postgres", connStr)
    require.NoError(t, err)
    defer db2.Close()

    // Connection A acquires lock
    var acquiredA bool
    err = testDB.QueryRow("SELECT pg_try_advisory_lock($1)", stackID).Scan(&acquiredA)
    require.NoError(t, err)
    require.True(t, acquiredA)

    // Connection B attempts same lock
    var acquiredB bool
    err = db2.QueryRow("SELECT pg_try_advisory_lock($1)", stackID).Scan(&acquiredB)
    require.NoError(t, err)
    require.False(t, acquiredB) // Lock already held by A

    // A releases lock
    _, err = testDB.Exec("SELECT pg_advisory_unlock($1)", stackID)
    require.NoError(t, err)

    // B can now acquire
    err = db2.QueryRow("SELECT pg_try_advisory_lock($1)", stackID).Scan(&acquiredB)
    require.NoError(t, err)
    require.True(t, acquiredB)
}
```

### Pattern 5: Table Truncation for Isolation
**What:** Reset DB state between tests by truncating tables in dependency order
**When to use:** Every test to ensure clean state, faster than transaction rollback for testcontainers
**Example:**
```go
func truncateAll(t *testing.T, db *sql.DB) {
    t.Helper()

    // Order matters: delete child tables before parents (FK constraints)
    tables := []string{
        "service_instances",
        "stacks",
        "projects",
        "categories",
        "services",
    }

    for _, table := range tables {
        _, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
        require.NoError(t, err, "failed to truncate %s", table)
    }
}
```

### Pattern 6: Soft-Delete Testing
**What:** Verify deleted_at column filters stacks from List, includes in ListTrash, allows Restore
**When to use:** Testing trash/restore endpoints
**Example:**
```go
func TestSoftDeleteSemantics(t *testing.T) {
    truncateAll(t, testDB)
    router := setupRouter(t)

    // Create stack
    createStackViaHTTP(t, router, "test-stack")

    // Delete stack (soft-delete)
    req := httptest.NewRequest(http.MethodDelete, "/api/v1/stacks/test-stack", nil)
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    require.Equal(t, http.StatusNoContent, w.Code)

    // Assert: not in List
    req = httptest.NewRequest(http.MethodGet, "/api/v1/stacks", nil)
    w = httptest.NewRecorder()
    router.ServeHTTP(w, req)
    var list respond.SuccessEnvelope
    json.NewDecoder(w.Body).Decode(&list)
    stacks := list.Data.([]interface{})
    require.Empty(t, stacks)

    // Assert: present in ListTrash
    req = httptest.NewRequest(http.MethodGet, "/api/v1/stacks/trash", nil)
    w = httptest.NewRecorder()
    router.ServeHTTP(w, req)
    json.NewDecoder(w.Body).Decode(&list)
    trash := list.Data.([]interface{})
    require.Len(t, trash, 1)

    // Restore
    req = httptest.NewRequest(http.MethodPost, "/api/v1/stacks/trash/test-stack/restore", nil)
    w = httptest.NewRecorder()
    router.ServeHTTP(w, req)
    require.Equal(t, http.StatusOK, w.Code)

    // Assert: back in List
    req = httptest.NewRequest(http.MethodGet, "/api/v1/stacks", nil)
    w = httptest.NewRecorder()
    router.ServeHTTP(w, req)
    json.NewDecoder(w.Body).Decode(&list)
    stacks = list.Data.([]interface{})
    require.Len(t, stacks, 1)
}
```

### Anti-Patterns to Avoid
- **Testing with mocked DB:** Integration tests should use real Postgres (testcontainers) not mocks — goal is to verify DB queries work
- **Shared mutable state between tests:** Each test must truncate tables or risk flaky failures from test order dependencies
- **Testing container operations in integration tests:** User decision: stub container client, focus on HTTP/DB path only
- **Transaction rollback for advisory lock tests:** Advisory locks are session-scoped — rollback doesn't release them, need separate connections
- **Hardcoding connection strings:** Use postgresContainer.ConnectionString() to get dynamic port assigned by testcontainers
- **Not using t.Helper() in test utilities:** Helper functions should call t.Helper() so failures report correct line number in test code
- **Parallel tests without isolation:** Don't use t.Parallel() with shared testDB unless using per-test transactions or separate DBs

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Postgres container lifecycle | Custom Docker CLI wrapper, manual port allocation | testcontainers-go/modules/postgres | Auto port allocation, wait strategies, cross-platform, auto-cleanup on test failure |
| Migration runner | Custom SQL file reader + executor | Reuse cmd/migrate logic or simplified version | Already tested migration runner exists, knows about schema_migrations table |
| HTTP request mocking | Manual http.Request construction with Body readers | httptest.NewRequest | Handles body encoding, sets correct headers, integrates with context |
| Response capture | Custom ResponseWriter implementation | httptest.NewRecorder | Captures status, headers, body correctly, supports Result() method for inspection |
| Assertion messages | fmt.Sprintf for every failure | testify/require | Auto-generated diff messages, better formatting for complex types |
| Container stub | Complex interface mocking | Minimal struct with no-op methods or nil fields | container.Client is a struct, can create with NewClient then override or pass disconnected runtime |

**Key insight:** Integration testing is about composition — testcontainers handles containers, httptest handles HTTP, testify handles assertions. Don't reimplement any of these.

## Common Pitfalls

### Pitfall 1: Advisory Lock Not Released
**What goes wrong:** Tests hang or fail because pg_advisory_unlock not called after pg_try_advisory_lock
**Why it happens:** Advisory locks are session-scoped — must unlock on same connection, not released on transaction rollback
**How to avoid:** Always defer unlock immediately after successful lock acquisition
**Warning signs:** Tests timeout, pg_locks table shows locks held after test completion
**Source:** [The Pitfall of Using PostgreSQL Advisory Locks with Go's DB Connection Pool](https://engineering.qubecinema.com/2019/08/26/unlocking-advisory-locks.html)

### Pitfall 2: Shared Loop Variable in Parallel Table Tests
**What goes wrong:** Parallel table-driven tests all see same test case data
**Why it happens:** Loop variable shared across goroutines, overwritten as loop continues
**How to avoid:** Capture loop variable: `tt := tt` before t.Run()
**Warning signs:** Random test failures, all subtests show same assertion error
**Source:** [Parallel Table-Driven Tests in Go](https://medium.com/@rosgluk/parallel-table-driven-tests-in-go-d06d53a02b1a)

### Pitfall 3: Container Not Cleaned Up After Test Failure
**What goes wrong:** Failed test exits without stopping container, orphaned containers accumulate
**Why it happens:** Not using defer for testcontainers.TerminateContainer()
**How to avoid:** Always defer termination immediately after successful container start
**Warning signs:** docker ps shows many postgres containers, disk space fills up
**Source:** [Getting started with Testcontainers for Go](https://testcontainers.com/guides/getting-started-with-testcontainers-for-go/)

### Pitfall 4: Wrong Envelope Assertion Strategy
**What goes wrong:** Test asserts response.Data directly instead of checking envelope structure first
**Why it happens:** json.Decoder returns generic interface{}, requires type assertion
**How to avoid:** First decode into respond.SuccessEnvelope, check Data != nil, then type assert
**Warning signs:** Test panics with "interface conversion" or "nil pointer dereference"

### Pitfall 5: Testing Against Wrong Database
**What goes wrong:** Tests run against production/development DB instead of testcontainers DB
**Why it happens:** DATABASE_URL env var set globally, overrides testcontainers connection string
**How to avoid:** Unset DATABASE_URL in test environment, pass explicit connection string to sql.Open()
**Warning signs:** Tests mutate real data, CI fails with connection refused

### Pitfall 6: Truncation Order Violates Foreign Keys
**What goes wrong:** TRUNCATE fails with "cannot truncate a table referenced in a foreign key constraint"
**Why it happens:** Truncating parent table before child tables
**How to avoid:** Truncate in dependency order (service_instances before stacks) or use CASCADE
**Warning signs:** truncateAll() returns FK constraint errors

### Pitfall 7: Testcontainers in CI Without Docker Socket
**What goes wrong:** CI job fails with "Cannot connect to the Docker daemon"
**Why it happens:** GitHub Actions runner needs Docker daemon running
**How to avoid:** Use ubuntu-latest image (has Docker preinstalled), don't use custom containers without docker-in-docker
**Warning signs:** Tests pass locally but fail in CI with socket errors
**Source:** [Running Testcontainers Tests Using GitHub Actions](https://www.docker.com/blog/running-testcontainers-tests-using-github-actions/)

## Code Examples

Verified patterns from official sources:

### Testcontainers Setup with Migration Runner
```go
// Source: https://golang.testcontainers.org/modules/postgres/
func setupTestDB(t *testing.T) *sql.DB {
    ctx := context.Background()

    container, err := postgres.Run(ctx,
        "postgres:16-alpine",
        postgres.WithDatabase("devarch_test"),
        postgres.WithUsername("test"),
        postgres.WithPassword("test"),
    )
    require.NoError(t, err)
    t.Cleanup(func() {
        if err := testcontainers.TerminateContainer(container); err != nil {
            t.Logf("failed to terminate container: %s", err)
        }
    })

    connStr, err := container.ConnectionString(ctx, "sslmode=disable")
    require.NoError(t, err)

    db, err := sql.Open("postgres", connStr)
    require.NoError(t, err)
    require.NoError(t, db.Ping())

    // Run migrations (simplified from cmd/migrate logic)
    migrateUp(t, db, "../../migrations")

    return db
}

func migrateUp(t *testing.T, db *sql.DB, dir string) {
    t.Helper()

    // Ensure migrations table
    _, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS schema_migrations (
            version VARCHAR(255) PRIMARY KEY,
            applied_at TIMESTAMPTZ DEFAULT NOW()
        )
    `)
    require.NoError(t, err)

    // Read and apply migrations
    files, err := filepath.Glob(filepath.Join(dir, "*.up.sql"))
    require.NoError(t, err)

    for _, file := range files {
        version := filepath.Base(strings.TrimSuffix(file, ".up.sql"))

        var exists bool
        err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", version).Scan(&exists)
        require.NoError(t, err)
        if exists {
            continue
        }

        content, err := os.ReadFile(file)
        require.NoError(t, err)

        _, err = db.Exec(string(content))
        require.NoError(t, err)

        _, err = db.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", version)
        require.NoError(t, err)
    }
}
```

### HTTP Handler Test with Envelope Verification
```go
// Source: https://pkg.go.dev/net/http/httptest + project respond package
func TestStackList(t *testing.T) {
    db := setupTestDB(t)
    truncateAll(t, db)

    // Seed data
    createStack(t, db, "stack-a")
    createStack(t, db, "stack-b")

    // Setup router
    stubClient := &container.Client{}
    router := api.NewRouter(db, stubClient, nil, nil, nil, nil, nil, nil, nil, security.Disabled)

    // Make request
    req := httptest.NewRequest(http.MethodGet, "/api/v1/stacks", nil)
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    // Assert response
    require.Equal(t, http.StatusOK, w.Code)
    require.Equal(t, "application/json", w.Header().Get("Content-Type"))

    var envelope respond.SuccessEnvelope
    err := json.NewDecoder(w.Body).Decode(&envelope)
    require.NoError(t, err)
    require.NotNil(t, envelope.Data)

    // Type assert to slice
    stacks, ok := envelope.Data.([]interface{})
    require.True(t, ok, "expected data to be array")
    require.Len(t, stacks, 2)
}
```

### Test Helper Functions
```go
// Source: User decision from CONTEXT.md
func createStack(t *testing.T, db *sql.DB, name string) int {
    t.Helper()

    var id int
    err := db.QueryRow(`
        INSERT INTO stacks (name, description, network_name, enabled)
        VALUES ($1, $2, $3, true)
        RETURNING id
    `, name, "Test stack", name+"-net").Scan(&id)
    require.NoError(t, err)

    return id
}

func createInstance(t *testing.T, db *sql.DB, stackID int, serviceName string) string {
    t.Helper()

    var instanceID string
    err := db.QueryRow(`
        INSERT INTO service_instances (stack_id, service_name, instance_id)
        VALUES ($1, $2, $3)
        RETURNING instance_id
    `, stackID, serviceName, serviceName+"-1").Scan(&instanceID)
    require.NoError(t, err)

    return instanceID
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `// +build integration` | `//go:build integration` | Go 1.17 (2021) | New syntax is parsed as expression, old syntax deprecated but still works. Use new syntax for consistency |
| dockertest | testcontainers-go | 2020+ shift | testcontainers has official modules (postgres, mysql, etc), better wait strategies, cross-language consistency |
| Manual docker compose in CI | testcontainers in CI | 2023+ | No GitHub Actions services: needed, testcontainers self-contained, same code locally and CI |
| assert.Equal(t.testify/assert is non-fatal) | require.Equal(t.testify/require is fail-fast) | Always coexisted | Use require for critical assertions (setup, DB queries), assert for optional checks |
| Transaction rollback per test | Table truncation per test | Depends on test needs | Rollback faster but breaks advisory lock tests (session-scoped). Truncation works for all cases |

**Deprecated/outdated:**
- `// +build` syntax: Still works but `//go:build` preferred
- dockertest library: Maintained but testcontainers has better ecosystem support
- GitHub Actions services: for DB: Still valid but testcontainers more flexible (can run locally with same code)

## Open Questions

1. **Container Client Stubbing Strategy**
   - What we know: container.Client is a struct with runtime field, NewClient() tries to detect docker/podman
   - What's unclear: Cleanest way to create stub — pass nil runtime, create disconnected client, or introduce minimal test interface
   - Recommendation: Create container.Client with explicit RuntimeType but don't call NewClient() — initialize struct directly with stub fields. Avoids runtime detection logic in tests

2. **Migration Runner Approach**
   - What we know: cmd/migrate has full-featured runner with up/down/status, uses schema_migrations table
   - What's unclear: Whether to extract reusable function from cmd/migrate or inline simplified version in test helpers
   - Recommendation: Inline simplified version in helpers_test.go — tests need up only, not down/status. Less dependency coupling

3. **TestMain vs Per-Test Container**
   - What we know: TestMain amortizes startup cost across all tests, per-test gives true isolation
   - What's unclear: Whether truncation is sufficient isolation or if per-test container needed
   - Recommendation: TestMain with truncation — user chose table truncation explicitly, testcontainers startup is 2-3 seconds (too slow per-test)

4. **Parallel Test Execution**
   - What we know: Tests share testDB from TestMain, advisory lock tests need separate connections
   - What's unclear: Can tests run with t.Parallel() if they all truncate tables
   - Recommendation: Don't use t.Parallel() — table truncation is not transaction-isolated, concurrent truncations could conflict. Sequential execution is simpler

## Sources

### Primary (HIGH confidence)
- [testcontainers-go Postgres module](https://golang.testcontainers.org/modules/postgres/) - Container setup, connection string retrieval
- [testify/require package](https://pkg.go.dev/github.com/stretchr/testify/require) - Assertion functions, require vs assert differences
- [net/http/httptest package](https://pkg.go.dev/net/http/httptest) - NewRecorder, NewRequest usage
- [Getting started with Testcontainers for Go](https://testcontainers.com/guides/getting-started-with-testcontainers-for-go/) - Lifecycle management, cleanup patterns
- [Testing a Go and chi RESTful API - Route Handlers](https://www.newline.co/@kchan/testing-a-go-and-chi-restful-api-route-handlers-part-1--6b105194) - Chi router testing with httptest

### Secondary (MEDIUM confidence)
- [Separate Your Go Tests with Build Tags](https://mickey.dev/posts/go-build-tags-testing/) - Build tag syntax and usage
- [4 practical principles of high-quality database integration tests in Go](https://threedots.tech/post/database-integration-testing/) - Database isolation strategies
- [The Pitfall of Using PostgreSQL Advisory Locks with Go's DB Connection Pool](https://engineering.qubecinema.com/2019/08/26/unlocking-advisory-locks.html) - Advisory lock session-scoping, unlock requirements
- [Isolating Integration Tests in Go](https://mtekmir.com/blog/golang-sql-integration-test-isolation/) - Transaction rollback vs table truncation tradeoffs
- [Running Testcontainers Tests Using GitHub Actions](https://www.docker.com/blog/running-testcontainers-tests-using-github-actions/) - CI setup requirements

### Tertiary (LOW confidence)
- [Parallel Table-Driven Tests in Go](https://medium.com/@rosgluk/parallel-table-driven-tests-in-go-d06d53a02b1a) - Loop variable capture pattern (verified in testing docs)

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - testcontainers-go, testify, httptest are industry standard, verified via official docs and pkg.go.dev
- Architecture: HIGH - Patterns verified via official examples, chi router testing confirmed via community guides
- Pitfalls: HIGH - Advisory lock session-scoping verified via PostgreSQL docs + engineering blog, loop variable issue documented in Go testing guides
- CI setup: MEDIUM - GitHub Actions + testcontainers verified via official blog, but project-specific workflow needs testing

**Research date:** 2026-02-12
**Valid until:** 2026-04-12 (60 days — Go testing ecosystem stable, testcontainers-go mature project with infrequent breaking changes)
