---
phase: 23-performance-optimization
plan: 01
subsystem: api-handlers
tags: [performance, optimization, database, batch-loading]
dependency_graph:
  requires: [phase-22]
  provides: [optimized-list-endpoints]
  affects: [service-handler, instance-handler]
tech_stack:
  added: []
  patterns: [batch-loading, filtered-count, union-all-aggregation]
key_files:
  created: []
  modified:
    - api/internal/api/handlers/service.go
    - api/internal/api/handlers/instance.go
decisions:
  - Batch loading via GetBatchServiceData eliminates N+1 queries for service includes
  - Shared filter clause ensures X-Total-Count matches filtered results
  - UNION ALL + GROUP BY pattern replaces 11 scalar subqueries with single aggregated query
metrics:
  duration_seconds: 116
  tasks_completed: 2
  files_modified: 2
  commits: 2
  completed_at: 2026-02-11T21:48:06Z
---

# Phase 23 Plan 01: Performance Optimization Summary

**One-liner:** Optimized service and instance list endpoints by eliminating N+1 runtime calls, fixing unfiltered count headers, and replacing scalar subquery chains with aggregated queries.

## What Was Built

Three targeted performance optimizations to reduce API latency:

1. **PERF-01: Batch service includes** - Service List handler now calls `GetBatchServiceData` once for all services instead of N individual `GetServiceState`/`GetServiceMetrics` calls when `include=status,metrics` is requested.

2. **PERF-02: Filtered count header** - Service List count query now applies the same category/search/enabled filters as the data query, ensuring `X-Total-Count` reflects the actual filtered result set instead of total unfiltered count.

3. **PERF-03: Aggregated override counts** - Instance List handler replaced 11 chained scalar subqueries with a single `UNION ALL` + `GROUP BY` + `LEFT JOIN` pattern for calculating override counts across all instance override tables.

## Implementation Details

### Service Handler Changes (service.go)

**Filter clause extraction:**
- Moved WHERE condition building to top of handler
- Created `filterClause` and `filterArgs` shared by both data and count queries
- Data query extends `filterArgs` with LIMIT/OFFSET parameters
- Count query uses only filter args

**Batch include loading:**
- Replaced per-service loop calling `loadServiceIncludes`
- Collect all service names into array
- Single `GetBatchServiceData` call returns map of states and metrics
- Map results back to services array
- `loadServiceIncludes` function preserved for single-service Get endpoint

### Instance Handler Changes (instance.go)

**Override count aggregation:**
- Replaced `(SELECT COUNT(*) ...) + (SELECT COUNT(*) ...) + ...` pattern (11 subqueries)
- New pattern: `LEFT JOIN (SELECT instance_id, COUNT(*) FROM (UNION ALL...) GROUP BY)`
- All 11 override tables: ports, volumes, env_vars, labels, domains, healthchecks, dependencies, config_files, env_files, networks, config_mounts
- Same semantics, single aggregated query execution

## Deviations from Plan

None - plan executed exactly as written.

## Verification Results

All success criteria met:

- [x] PERF-01: Service List calls GetBatchServiceData for include=status,metrics (one call, not N)
- [x] PERF-02: X-Total-Count reflects active filters (category, search, enabled)
- [x] PERF-03: Instance List uses UNION ALL + GROUP BY for override counts (no scalar subqueries)
- [x] Full build passes: `go build ./...`

Verification commands:
```bash
go build ./...  # clean compilation
grep GetBatchServiceData service.go  # batch call confirmed
grep "countQuery.*filterClause" service.go  # filtered count confirmed
grep "UNION ALL" instance.go  # aggregated pattern confirmed (10 occurrences)
```

## Self-Check: PASSED

### Created Files
None - only modifications.

### Modified Files
- [x] `/home/fhcadmin/projects/devarch/api/internal/api/handlers/service.go` exists
- [x] `/home/fhcadmin/projects/devarch/api/internal/api/handlers/instance.go` exists

### Commits
- [x] `28ccd6a` exists: refactor(23-01): batch service includes and filtered count
- [x] `b18aa74` exists: refactor(23-01): aggregated instance override count query

## Impact

**Performance improvements:**
- Service list with includes: O(N) runtime calls reduced to O(1)
- Service list pagination: Count now reflects actual filtered results (correct UX)
- Instance list: 11 scalar subqueries consolidated to 1 aggregated query (DB efficiency)

**No breaking changes:**
- Response schemas unchanged
- Single-service Get endpoints unchanged (still use per-call pattern where appropriate)
- Override count semantics identical (same 11 tables, same total)

## Next Steps

Phase 23-01 complete. Ready for next plan in performance optimization phase.
