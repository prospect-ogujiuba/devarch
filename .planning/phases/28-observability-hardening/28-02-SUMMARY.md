---
phase: 28-observability-hardening
plan: 02
subsystem: sync
tags: [persistence, database, cleanup, observability]

dependency_graph:
  requires:
    - migration-012-flowstate-build-context
    - sync-manager-daily-cleanup
  provides:
    - sync-job-persistence
    - sync-job-history
  affects:
    - sync-manager
    - daily-cleanup-cycle

tech_stack:
  added: []
  patterns:
    - write-through-cache
    - merged-memory-db-queries
    - 7-day-retention-policy

key_files:
  created:
    - api/migrations/013_sync_jobs.up.sql
    - api/migrations/013_sync_jobs.down.sql
  modified:
    - api/internal/sync/manager.go

decisions:
  - summary: "Write-through persistence removes completed jobs from memory after successful DB insert"
    rationale: "Reduces memory footprint while maintaining observability"
  - summary: "GetJobs merges in-memory (running) + DB (completed) for unified view"
    rationale: "Seamless experience across process restarts"
  - summary: "7-day retention integrated into existing daily cleanup cycle"
    rationale: "Reuses proven batched deletion pattern"

metrics:
  duration_seconds: 157
  tasks_completed: 1
  files_created: 2
  files_modified: 1
  completed_at: "2026-02-12T15:50:06Z"
---

# Phase 28 Plan 02: Sync Job Persistence Summary

**One-liner:** Persist sync job history to PostgreSQL with write-through pattern, 7-day retention, and DB-backed GetJobs for restart survival.

## What Was Built

### Migration 013: sync_jobs Table
- TEXT primary key (timestamp-based job IDs)
- Captures type, status, started_at, ended_at, error
- Indexed by created_at DESC for efficient recent job queries
- Down migration for rollback support

### Sync Manager Persistence
- **Job struct:** Added CreatedAt field for DB scan support
- **TriggerSync:** Write-through persistence after job completion
  - Persists to DB inside jobsMu lock after status update
  - Removes from in-memory map after successful DB insert
  - Graceful degradation: keeps in memory if DB fails
- **GetJobs:** Merged memory + DB approach
  - Queries DB for recent 100 completed jobs
  - Prepends in-memory running jobs to front
  - Falls back to memory-only if DB query fails
- **cleanupSyncJobs:** 7-day retention via batched deletion
  - Reuses deleteInBatches helper with CTE pattern
  - Integrated into daily cleanup ops list (5th operation)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed NewRouter call signature mismatch (pre-existing)**
- **Found during:** Initial compilation check
- **Issue:** main.go was passing logger parameter but code had already been updated in commit f714767 to accept it
- **Fix:** Pre-existing change already resolved - no action needed
- **Files modified:** None (already committed)
- **Commit:** f714767 (pre-existing)

## Verification Results

All verification checks passed:
- ✅ `go build ./...` compiles successfully
- ✅ Migration file exists with CREATE TABLE sync_jobs (TEXT PK)
- ✅ `INSERT INTO sync_jobs` present in TriggerSync (line 725)
- ✅ `SELECT ... FROM sync_jobs` present in GetJobs (lines 742-747)
- ✅ cleanupSyncJobs method and daily ops list integration (lines 438, 566)

## Key Implementation Details

**Write-Through Pattern:**
```go
// After job completion (inside jobsMu.Lock):
_, dbErr := m.db.ExecContext(context.Background(),
    `INSERT INTO sync_jobs ...`, job.ID, job.Type, job.Status, ...)
if dbErr != nil {
    log.Printf("sync: failed to persist job %s: %v", jobID, dbErr)
} else {
    delete(m.jobs, jobID)  // Remove from memory after DB success
}
```

**Merged Query Pattern:**
```go
// 1. Query DB for completed jobs (LIMIT 100)
// 2. Scan into jobs slice
// 3. Prepend in-memory running jobs to front
// 4. Return merged list
```

**Retention Integration:**
- Added to daily cleanup ops after "soft-deleted"
- Uses same CTE + LIMIT + batch pattern as other cleanup operations
- 7-day cutoff: `time.Now().Add(-7 * 24 * time.Hour)`

## Testing Notes

Manual testing steps:
1. Run migration 013: `go run ./cmd/migrate -migrations ./migrations`
2. Trigger sync: `curl -X POST http://localhost:8550/api/v1/sync/trigger -H "X-API-Key: ..."`
3. Check DB: `SELECT * FROM sync_jobs ORDER BY created_at DESC LIMIT 10;`
4. Restart API and verify GetJobs returns historical jobs

## Success Criteria Met

- [x] sync_jobs table migration exists (013_sync_jobs.up.sql + down.sql)
- [x] Completed sync jobs persist to DB via write-through in TriggerSync
- [x] GetJobs returns both in-progress (memory) and completed (DB) jobs
- [x] Job history survives API restart (DB-backed)
- [x] 7-day cleanup integrated into existing daily cleanup cycle

## Self-Check: PASSED

**Files exist:**
- FOUND: api/migrations/013_sync_jobs.up.sql
- FOUND: api/migrations/013_sync_jobs.down.sql
- FOUND: api/internal/sync/manager.go

**Commits exist:**
- FOUND: d6fb6b3 (feat(28-02): persist sync job history to PostgreSQL)

**Code verification:**
- FOUND: CREATE TABLE sync_jobs in migration file
- FOUND: INSERT INTO sync_jobs in manager.go (line 725)
- FOUND: SELECT FROM sync_jobs in manager.go (lines 742-747)
- FOUND: cleanupSyncJobs in manager.go (lines 438, 566)
