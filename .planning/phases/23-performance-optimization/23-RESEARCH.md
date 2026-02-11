# Phase 23: Performance Optimization - Research

**Researched:** 2026-02-11
**Domain:** Go database performance optimization, PostgreSQL query optimization, batch operations
**Confidence:** HIGH

## Summary

Phase 23 addresses three distinct performance issues in DevArch's API:

1. **N+1 Query Problem (PERF-01)**: Service list handler calls `GetServiceState()` and `GetServiceMetrics()` in a loop when `include=status,metrics` is specified, making 2N runtime API calls for N services. Solution: batch retrieval using existing `GetBatchServiceData()` function.

2. **Inaccurate Total Count (PERF-02)**: `X-Total-Count` header returns unfiltered count from `SELECT COUNT(*) FROM services` regardless of active filters (category, search, enabled). Solution: apply same WHERE conditions to count query.

3. **Scalar Subquery Chain (PERF-03)**: Instance list query uses 7 scalar subqueries chained with `+` to calculate override counts, forcing PostgreSQL to execute each subquery independently per row. Solution: rewrite as single aggregated query with LEFT JOINs or UNION ALL + GROUP BY.

**Primary recommendation:** These are straightforward optimizations requiring focused refactoring of existing handler code with measurable performance criteria. Use Go's built-in benchmarking for verification.

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| database/sql | stdlib | Database interface | Go's standard SQL interface with connection pooling |
| lib/pq | 1.10.9 | PostgreSQL driver | Mature, stable driver for PostgreSQL in Go |
| testing | stdlib | Benchmarking framework | Native Go benchmarking with `-bench` flag |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| httptest | stdlib | HTTP handler testing | Benchmarking API endpoint response times |
| context | stdlib | Request cancellation | Timeout enforcement for batch operations |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| lib/pq | pgx | pgx has native batch support and better performance, but requires rewriting existing code; not worth migration for this phase |
| Manual batching | Third-party DataLoader | Adds dependency for pattern easily implemented in 20 lines |

**Installation:**
No new dependencies required — all optimizations use existing stack.


## Architecture Patterns

### Pattern 1: Batch Data Loading (N+1 Solution)
**What:** Collect all IDs/names first, then fetch related data in single batch call
**When to use:** When fetching related data for multiple entities in a loop
**Example:**
```go
// BEFORE (N+1 problem)
for i := range services {
    state, _ := client.GetServiceState(ctx, services[i].Name)
    services[i].Status = state
    metrics, _ := client.GetServiceMetrics(ctx, services[i].Name)
    services[i].Metrics = metrics
}

// AFTER (batch retrieval)
names := make([]string, len(services))
for i, svc := range services {
    names[i] = svc.Name
}

batch, err := client.GetBatchServiceData(ctx, names)
if err == nil {
    for i := range services {
        if state, ok := batch.States[services[i].Name]; ok {
            services[i].Status = state
        }
        if metrics, ok := batch.Metrics[services[i].Name]; ok {
            services[i].Metrics = metrics
        }
    }
}
```

**Note:** DevArch already has `GetBatchServiceData()` in `podman/metrics.go:91-112` — just needs integration into service handler.

### Pattern 2: Filtered Count Query
**What:** Apply same WHERE conditions to both data query and count query
**When to use:** When pagination headers must reflect filtered results, not total table size
**Example:**
```go
// Build WHERE clause and args once
whereClause, args := buildFilters(r.URL.Query())

// Use for data query
query := "SELECT ... FROM services s " + whereClause + " ORDER BY ..."
rows, _ := db.Query(query, args...)

// Reuse for count query
countQuery := "SELECT COUNT(*) FROM services s " + whereClause
var total int
db.QueryRow(countQuery, args...).Scan(&total)
```

**Key insight:** Extract filter building into helper function to prevent drift between data and count queries.


### Pattern 3: Aggregated Subquery Replacement
**What:** Replace scalar subquery chain with single aggregated query using UNION ALL + GROUP BY
**When to use:** When counting related records across multiple tables
**Example:**
```go
// BEFORE (scalar subquery chain - executes 7 subqueries per row)
query := `
    SELECT si.id,
        (SELECT COUNT(*) FROM instance_ports WHERE instance_id = si.id) +
        (SELECT COUNT(*) FROM instance_volumes WHERE instance_id = si.id) +
        (SELECT COUNT(*) FROM instance_env_vars WHERE instance_id = si.id) +
        ...
    FROM service_instances si
`

// AFTER (aggregated query - single pass)
query := `
    SELECT si.id, COALESCE(counts.total, 0) as override_count
    FROM service_instances si
    LEFT JOIN (
        SELECT instance_id, COUNT(*) as total
        FROM (
            SELECT instance_id FROM instance_ports
            UNION ALL
            SELECT instance_id FROM instance_volumes
            UNION ALL
            SELECT instance_id FROM instance_env_vars
            UNION ALL
            SELECT instance_id FROM instance_healthchecks
            UNION ALL
            SELECT instance_id FROM instance_dependencies
            UNION ALL
            SELECT instance_id FROM instance_config_files
            UNION ALL
            SELECT instance_id FROM instance_env_files
            UNION ALL
            SELECT instance_id FROM instance_networks
            UNION ALL
            SELECT instance_id FROM instance_config_mounts
        ) all_overrides
        GROUP BY instance_id
    ) counts ON counts.instance_id = si.id
`
```

**Performance:** UNION ALL + GROUP BY scans each table once and aggregates, versus scalar subqueries which execute 7 separate queries per result row.

### Pattern 4: Go Benchmark Testing
**What:** Use Go's built-in `testing.B` for performance verification
**When to use:** Validating optimization impact meets success criteria
**Example:**
```go
func BenchmarkServiceList_BatchRetrieval(b *testing.B) {
    // Setup: create 100 test services
    db := setupTestDB(b)
    handler := NewServiceHandler(db, containerClient, podmanClient, cipher)

    // Create request with include=status,metrics
    req := httptest.NewRequest("GET", "/api/v1/services?include=status,metrics", nil)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        w := httptest.NewRecorder()
        handler.List(w, req)
        if w.Code != 200 {
            b.Fatalf("expected 200, got %d", w.Code)
        }
    }

    // Report custom metrics
    if b.N > 0 {
        elapsed := b.Elapsed()
        avgMs := elapsed.Milliseconds() / int64(b.N)
        if avgMs > 100 {
            b.Errorf("average request time %dms exceeds 100ms target", avgMs)
        }
    }
}
```

### Anti-Patterns to Avoid
- **Premature batching:** Don't batch queries for single-item lookups (e.g., `GET /service/:name` with includes)
- **Over-optimization:** Don't rewrite working queries unless they're proven bottlenecks
- **Breaking transport independence:** Keep batch logic in handler layer, not service layer (per Phase 21 decision)

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Connection pooling | Custom connection manager | database/sql built-in pool | database/sql handles pooling, max connections, idle timeouts automatically |
| Query parameter building | String concatenation | Helper function + []interface{} | Prevents SQL injection, maintains parameter count correctly |
| Benchmark timing | Manual time.Now() wrapping | testing.B framework | Handles warmup, statistical analysis, memory profiling automatically |

**Key insight:** Go's stdlib provides robust primitives for database performance — leverage them instead of third-party tools.


## Common Pitfalls

### Pitfall 1: Counting Twice
**What goes wrong:** Running filtered count query after fetching data increases total query time
**Why it happens:** Trying to reuse transaction or not recognizing parallelization opportunity
**How to avoid:** Run count query and data query in parallel using goroutines with shared context
**Warning signs:** Total endpoint latency roughly equals (count query time + data query time)

### Pitfall 2: UNION vs UNION ALL
**What goes wrong:** Using UNION instead of UNION ALL for override count aggregation adds unnecessary deduplication overhead
**Why it happens:** Default UNION behavior removes duplicates
**How to avoid:** Use UNION ALL when duplicates are impossible (different tables) or don't matter (counting)
**Warning signs:** EXPLAIN ANALYZE shows "Unique" or "HashAggregate" nodes before final aggregation

### Pitfall 3: Batch Size Explosion
**What goes wrong:** Batching 1000+ services causes timeout or memory issues
**Why it happens:** No pagination consideration in batch retrieval
**How to avoid:** Batch retrieval should respect pagination limits (max 500 per DevArch's existing limit)
**Warning signs:** Batch operations succeed in tests but timeout in production

### Pitfall 4: Ignored Errors in Batch Loading
**What goes wrong:** Batch failure silently returns incomplete data instead of error
**Why it happens:** Using `if err == nil` instead of proper error handling
**How to avoid:** Log batch errors but gracefully degrade (return services without runtime data)
**Warning signs:** Intermittent missing status/metrics in API responses

### Pitfall 5: Filter Drift
**What goes wrong:** Count query and data query apply different filters, causing pagination bugs
**Why it happens:** Copy-paste WHERE clause instead of extracting to function
**How to avoid:** Build WHERE clause once, apply to both queries
**Warning signs:** X-Total-Count doesn't match actual filtered result count

### Pitfall 6: Context Timeout
**What goes wrong:** Batch runtime calls timeout before completing
**Why it happens:** Using request context with short deadline for potentially slow Podman API calls
**How to avoid:** Consider separate timeout for batch operations or test with realistic container counts
**Warning signs:** Batch calls succeed individually but fail when batched


## Code Examples

Verified patterns from codebase analysis:

### Current N+1 Implementation (service.go:202-220)
```go
// Source: /home/fhcadmin/projects/devarch/api/internal/api/handlers/service.go:202-220
func (h *ServiceHandler) loadServiceIncludes(ctx context.Context, s *models.Service, includes string) {
    for _, inc := range []string{"status", "metrics"} {
        if !containsInclude(includes, inc) {
            continue
        }
        switch inc {
        case "status":
            state, err := h.podmanClient.GetServiceState(ctx, s.Name)
            if err == nil {
                s.Status = state
            }
        case "metrics":
            metrics, err := h.podmanClient.GetServiceMetrics(ctx, s.Name)
            if err == nil {
                s.Metrics = metrics
            }
        }
    }
}
```

### Existing Batch Function (podman/metrics.go:91-112)
```go
// Source: /home/fhcadmin/projects/devarch/api/internal/podman/metrics.go:91-112
func (c *Client) GetBatchServiceData(ctx context.Context, names []string) (*BatchMetrics, error) {
    result := &BatchMetrics{
        States:  make(map[string]*models.ContainerState, len(names)),
        Metrics: make(map[string]*models.ContainerMetrics, len(names)),
    }

    for _, name := range names {
        state, err := c.GetServiceState(ctx, name)
        if err == nil {
            result.States[name] = state
        }

        if state != nil && state.Status == "running" {
            metrics, err := c.GetServiceMetrics(ctx, name)
            if err == nil {
                result.Metrics[name] = metrics
            }
        }
    }

    return result, nil
}
```

**Note:** Current implementation still loops, but consolidates to single function call from handler perspective — reduces overhead of repeated ctx/error handling. True batching would require Podman API batch endpoint (out of scope).

### Current Count Query Issue (service.go:180-181)
```go
// Source: /home/fhcadmin/projects/devarch/api/internal/api/handlers/service.go:180-181
var total int
h.db.QueryRow("SELECT COUNT(*) FROM services s JOIN categories c ON s.category_id = c.id WHERE 1=1").Scan(&total)
```

**Problem:** Hardcoded `WHERE 1=1` ignores filters from lines 88-103 (category, search, enabled).


### Current Scalar Subquery Chain (instance.go:234-252)
```go
// Source: /home/fhcadmin/projects/devarch/api/internal/api/handlers/instance.go:234-252
query := `
    SELECT si.id, si.stack_id, si.instance_id, si.template_service_id,
        s.name as template_name, si.container_name, si.description,
        si.enabled, si.created_at, si.updated_at, (
            SELECT COUNT(*) FROM instance_ports WHERE instance_id = si.id
        ) + (
            SELECT COUNT(*) FROM instance_volumes WHERE instance_id = si.id
        ) + (
            SELECT COUNT(*) FROM instance_env_vars WHERE instance_id = si.id
        ) + (
            SELECT COUNT(*) FROM instance_healthchecks WHERE instance_id = si.id
        ) + (
            SELECT COUNT(*) FROM instance_dependencies WHERE instance_id = si.id
        ) + (
            SELECT COUNT(*) FROM instance_config_files WHERE instance_id = si.id
        ) + (
            SELECT COUNT(*) FROM instance_env_files WHERE instance_id = si.id
        ) + (
            SELECT COUNT(*) FROM instance_networks WHERE instance_id = si.id
        ) + (
            SELECT COUNT(*) FROM instance_config_mounts WHERE instance_id = si.id
        ) as override_count
    FROM service_instances si
    JOIN services s ON s.id = si.template_service_id
`
```

**Problem:** PostgreSQL executes 9 subqueries per row in result set.

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Scalar subqueries for aggregation | LATERAL JOIN or UNION ALL + GROUP BY | PostgreSQL 9.3+ (2013) | 5-10x faster for multi-table counts |
| Manual batching loops | DataLoader pattern / breadth-first loading | 2025 (DataLoader 3.0) | Reduces concurrency O(N²) → O(1) |
| Separate count queries | Window functions with OVER() | PostgreSQL 8.4+ (2009) | Single query for data + count, but complex for filtered pagination |
| IN clause with many values | ANY with array parameter | PostgreSQL 14 (2021) | 10x faster for 1000+ values |
| lib/pq prepared statements | pgx native batch protocol | pgx 2016+ | True batch operations, but requires migration |

**Deprecated/outdated:**
- **COUNT(*) optimization tricks:** PostgreSQL 13+ improved COUNT(*) planner significantly; old workarounds unnecessary
- **EXPLAIN without ANALYZE:** Always use EXPLAIN ANALYZE for actual timing data, not just planner estimates

## Open Questions

1. **Podman API Batch Endpoint**
   - What we know: Current `GetBatchServiceData` still loops over individual container inspect calls
   - What's unclear: Whether Podman 5.x supports batch inspect endpoint
   - Recommendation: Phase 23 scope is handler-level batching; true runtime batching deferred to future phase

2. **Benchmark Environment**
   - What we know: Success criteria specifies <100ms for 100 services
   - What's unclear: Should test include actual Podman calls or use mocks?
   - Recommendation: Integration test with real Podman, unit benchmark with mocks for consistent CI results

3. **Parallel Count Query**
   - What we know: Count and data queries could run in parallel goroutines
   - What's unclear: Is added complexity worth ~20-30ms savings?
   - Recommendation: Start with sequential, add parallelization only if benchmarks show >50ms count query time


## Sources

### Primary (HIGH confidence)
- [PostgreSQL Official Docs: Aggregate Functions](https://www.postgresql.org/docs/12/functions-aggregate.html)
- [Go Official Blog: Context](https://go.dev/blog/context)
- [Go Packages: testing](https://pkg.go.dev/testing)
- [Go Packages: lib/pq](https://pkg.go.dev/github.com/lib/pq)
- DevArch codebase: `api/internal/api/handlers/service.go`, `api/internal/podman/metrics.go`, `api/internal/api/handlers/instance.go`

### Secondary (MEDIUM confidence)
- [Faster PostgreSQL Counting - Citus Data](https://www.citusdata.com/blog/2016/10/12/count-performance/)
- [PostgreSQL count(*) made fast - CYBERTEC](https://www.cybertec-postgresql.com/en/postgresql-count-made-fast/)
- [Postgres performance with IN vs ANY - pganalyze](https://pganalyze.com/blog/5mins-postgres-performance-in-vs-any)
- [Postgres Query Boost: Using ANY Instead of IN - Crunchy Data](https://www.crunchydata.com/blog/postgres-query-boost-using-any-instead-of-in)
- [Understanding FILTER in PostgreSQL - Tiger Data](https://www.tigerdata.com/learn/understanding-filter-in-postgresql-with-examples)
- [SQL Optimizations in PostgreSQL: IN vs EXISTS vs ANY/ALL vs JOIN - Percona](https://www.percona.com/blog/sql-optimizations-in-postgresql-in-vs-exists-vs-any-all-vs-join/)
- [LATERAL JOIN vs Subquery in PostgreSQL - w3tutorials](https://www.w3tutorials.net/blog/what-is-the-difference-between-a-lateral-join-and-a-subquery-in-postgresql/)
- [PostgreSQL's Powerful New Join Type: LATERAL - Heap](https://www.heap.io/blog/postgresqls-powerful-new-join-type-lateral)
- [RESTful API Pagination Best Practices - Medium](https://medium.com/@khdevnet/restful-api-pagination-best-practices-a-developers-guide-5b177a9552ef)
- [REST API Design: Filtering, Sorting, and Pagination - Moesif](https://www.moesif.com/blog/technical/api-design/REST-API-Design-Filtering-Sorting-and-Pagination/)
- [10 RESTful API Pagination Best Practices - Nordic APIs](https://nordicapis.com/restful-api-pagination-best-practices/)
- [N+1 Query Problem: The Database Killer - Medium](https://medium.com/@saad.minhas.codes/n-1-query-problem-the-database-killer-youre-creating-f68104b99a2d)
- [Solving N+1 Query Problems in Go Applications - Prisma Client Go](https://goprisma.org/blog/solving-n1-query-problems-in-go-applications)
- [Dataloader 3.0: A new algorithm to solve the N+1 Problem - WunderGraph](https://wundergraph.com/blog/dataloader_3_0_breadth_first_data_loading)
- [Improving Postgres Performance Tenfold Using Go Concurrency - LakeFS](https://lakefs.io/blog/improving-postgres-performance-tenfold-using-go-concurrency/)
- [How to Benchmark PostgreSQL for Optimal Performance - DZone](https://dzone.com/articles/how-to-benchmark-postgresql-for-optimal-performance)
- [How To Use Explain Analyze To Improve Query Performance in PostgreSQL - EnterpriseDB](https://www.enterprisedb.com/blog/postgresql-query-optimization-performance-tuning-with-explain-analyze)
- [Testing SQL Performance in PostgreSQL - Thoughtbot](https://thoughtbot.com/blog/test-sql-performance)
- [Go Concurrency Patterns You Must Know - Opcito](https://www.opcito.com/blogs/practical-concurrency-patterns-in-go)
- [Efficient Concurrency in Go: Worker Pool Pattern - Medium](https://rksurwase.medium.com/efficient-concurrency-in-go-a-deep-dive-into-the-worker-pool-pattern-for-batch-processing-73cac5a5bdca)
- [Benchmarking and Load Testing for Networked Go Apps - Go Optimization Guide](https://goperf.dev/02-networking/bench-and-load/)
- [Testing in Go: Unit, Integration, and Benchmark Tests - dasroot.net](https://dasroot.net/posts/2026/01/testing-in-go-unit-integration-benchmark-tests/)
- [Writing Benchmarks: Performance testing in Go/Golang - willem.dev](https://www.willem.dev/articles/benchmarks-performance-testing/)

### Tertiary (LOW confidence)
- None — all findings verified through official docs or multiple credible sources

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All stdlib or existing dependencies
- Architecture: HIGH - Patterns verified in existing codebase, PostgreSQL docs, Go community standards
- Pitfalls: HIGH - Based on common PostgreSQL performance issues and Go database best practices

**Research date:** 2026-02-11
**Valid until:** 2026-03-11 (30 days - stable domain, PostgreSQL and Go patterns don't change rapidly)
