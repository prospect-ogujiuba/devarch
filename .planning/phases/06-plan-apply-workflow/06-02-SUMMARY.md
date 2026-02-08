---
phase: 06-plan-apply-workflow
plan: 02
subsystem: api-handlers
tags: [http, advisory-lock, postgres, compose, docker]
requires: [06-01, 05-01, 04-01, 03-01]
provides:
  - GET /stacks/{name}/plan endpoint
  - POST /stacks/{name}/apply endpoint
  - advisory lock coordination
  - staleness validation
affects: [06-03]
tech-stack:
  added: []
  patterns: [pg_try_advisory_lock, sequential-apply-flow]
key-files:
  created:
    - api/internal/api/handlers/stack_plan.go
    - api/internal/api/handlers/stack_apply.go
  modified:
    - api/internal/api/routes.go
decisions:
  - title: "pg_try_advisory_lock per stack"
    rationale: "Prevents concurrent applies on same stack, uses stack.id as lock key"
    alternatives: ["distributed lock", "optimistic locking"]
    chosen: "postgres advisory lock (already available, simple)"
  - title: "Sequential apply flow"
    rationale: "Network -> configs -> compose up, with cleanup on error"
    alternatives: ["atomic rollback", "partial apply retry"]
    chosen: "sequential with cleanup (configs left for debugging on compose failure)"
metrics:
  duration: ~2.0min
  completed: 2026-02-07
---

# Phase 6 Plan 02: Plan/Apply HTTP Handlers Summary

**One-liner:** GET /plan returns diff preview with staleness token, POST /apply acquires advisory lock and runs sequential flow (network, configs, compose up).

## What Was Built

Created two new handler methods on existing `StackHandler`:

**Plan handler** (`stack_plan.go`):
- Query stack + all instances (including disabled) with timestamps
- Query running containers via `ListContainersWithLabels("devarch.stack_id": stackName)`
- Compute diff via `plan.ComputeDiff(desired, running)`
- Generate staleness token via `plan.GenerateToken(stackUpdatedAt, instanceTimestamps)`
- Return structured JSON: stack name, ID, changes array, token, generated timestamp

**Apply handler** (`stack_apply.go`):
- Decode JSON body for token (required)
- Query stack by name
- **Acquire advisory lock** via `pg_try_advisory_lock(stackID)` — return 409 if already locked
- **Validate token** via `plan.ValidateToken(db, stackID, token)` — return 409 if stale
- **Ensure network** via `CreateNetwork(netName, labels)` (idempotent)
- **Materialize configs** via `gen.MaterializeStackConfigs(stackName, projectRoot)`
- **Generate compose YAML** via `gen.GenerateStack(stackName)`
- **Run compose up** via `RunCompose(tmpFile, "--project-name", "devarch-"+stackName, "up", "-d")`
- Return 200 JSON with status "applied" and compose output
- **Defer unlock** via `pg_advisory_unlock(stackID)` with 5s timeout context

**Route wiring** (`routes.go`):
- Added `r.Get("/plan", stackHandler.Plan)` under `/{name}` stack group
- Added `r.Post("/apply", stackHandler.Apply)` under `/{name}` stack group

## Technical Decisions Made

1. **Advisory lock per stack (not global)**: Each stack has independent lock using `stack.id` as key. Multiple stacks can apply concurrently, same stack cannot.

2. **Lock acquisition uses `pg_try_advisory_lock`**: Non-blocking. Returns 409 immediately if already locked (user sees "Stack is being applied by another session").

3. **Unlock deferred with background context**: If request context is cancelled, unlock still completes (5s timeout). Prevents orphaned locks.

4. **Staleness validation after lock acquisition**: Token validated after lock to ensure no TOCTOU race. If stale, lock released immediately, user gets 409.

5. **Sequential apply flow**: Network -> configs -> compose up. No rollback on failure (configs left for debugging). Network creation is idempotent.

6. **Compose output returned on success**: User sees full `docker compose up -d` output for troubleshooting.

7. **Empty running containers gracefully handled in Plan**: If `ListContainersWithLabels` fails (runtime down), plan uses empty slice. All instances show as "add" — user can still generate plan.

## Implementation Notes

**Deviation tracking:**
- [Rule 1 - Bug] Removed unused `container` import from `stack_apply.go` after initial compilation error

**Advisory lock pattern:**
- Same pattern as `sync/manager.go` cleanup loop
- `pg_try_advisory_lock` returns `bool` — scan result to check acquisition
- Unlock uses background context (not request context) to guarantee cleanup

**Error handling:**
- Network failure: return 500 (nothing to roll back)
- Config materialization failure: clean up config dir, return 500
- Compose generation failure: clean up config dir, return 500
- Compose up failure: leave configs (for debugging), return 500 with output

**Container label for plan query:**
- Uses `devarch.stack_id` label (set by identity labels in Phase 4)
- Query pattern: `ListContainersWithLabels(map[string]string{"devarch.stack_id": stackName})`

## Testing Suggestions

**Plan endpoint:**
1. GET /stacks/test-stack/plan with no instances → empty changes array
2. GET /stacks/test-stack/plan with disabled instance + running container → modify action
3. GET /stacks/test-stack/plan with enabled instance + no container → add action
4. GET /stacks/test-stack/plan with running container not in DB → remove action
5. Verify token changes when stack.updated_at or instance.updated_at changes

**Apply endpoint:**
1. POST /stacks/test-stack/apply with valid token → success
2. POST /stacks/test-stack/apply with stale token → 409 "Plan is stale"
3. Concurrent POST /stacks/test-stack/apply → second request gets 409 "being applied by another session"
4. POST /stacks/test-stack/apply with network creation error → 500
5. POST /stacks/test-stack/apply with compose up failure → 500 with output

## Files Changed

**Created:**
- `api/internal/api/handlers/stack_plan.go` (108 lines) — Plan handler
- `api/internal/api/handlers/stack_apply.go` (151 lines) — Apply handler with advisory lock

**Modified:**
- `api/internal/api/routes.go` — Added `/plan` (GET) and `/apply` (POST) routes

## Deviations from Plan

**Auto-fixed Issues:**

**1. [Rule 1 - Bug] Unused import in stack_apply.go**
- **Found during:** Task 2 compilation
- **Issue:** Imported `internal/container` but didn't use it
- **Fix:** Removed import (no container name computation needed in apply handler)
- **Files modified:** api/internal/api/handlers/stack_apply.go
- **Commit:** ecf64a33

No other deviations — plan executed as written.

## Next Phase Readiness

**For Phase 6 Plan 03 (Dashboard UI):**
- Plan endpoint returns structured JSON ready for diff display
- Apply endpoint returns 409 on staleness (UI can prompt "regenerate plan")
- Apply endpoint returns 409 on concurrent apply (UI can show "in progress" state)
- Compose output returned for log display in UI

**Blockers:** None

**Dependencies:** Phase 6 Plan 03 depends on these endpoints being available.
