---
phase: 06-plan-apply-workflow
plan: 01
subsystem: api
tags: [go, plan, diff, staleness, sha256]

requires:
  - phase: 05-compose-generation
    provides: compose generation foundation
provides:
  - Plan domain package with types, diff computation, staleness detection
  - Stateless differ for add/modify/remove change detection
  - SHA256-based staleness token for optimistic concurrency
affects: [06-02-plan-endpoint, 06-03-apply-endpoint]

tech-stack:
  added: [crypto/sha256]
  patterns: [stateless differ pattern, staleness token via SHA256]

key-files:
  created:
    - api/internal/plan/types.go
    - api/internal/plan/differ.go
    - api/internal/plan/staleness.go
  modified: []

key-decisions:
  - "Stateless ComputeDiff - caller provides desired + running inputs"
  - "Modifications scoped to enabled/disabled toggle only (config drift deferred)"
  - "Terraform-style change ordering: removes, modifies, adds"
  - "Deterministic token via sorted instances + RFC3339Nano timestamps"

patterns-established:
  - "Differ pattern: stateless function, caller provides all inputs"
  - "Staleness detection: SHA256 hash of sorted timestamps"
  - "Sentinel error pattern: ErrStalePlan for caller error handling"

duration: 1min
completed: 2026-02-07
---

# Phase 6 Plan 01: Plan Domain Package Summary

**Stateless plan domain with diff computation (add/modify/remove) and SHA256 staleness tokens**

## Performance

- **Duration:** 1 min
- **Started:** 2026-02-07T19:46:16Z
- **Completed:** 2026-02-07T19:47:13Z
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments
- Plan types with JSON tags for full diff schema (add/modify/remove with per-field detail)
- Stateless ComputeDiff handles add/modify/remove detection from desired vs running inputs
- Deterministic staleness tokens via SHA256 hash of sorted timestamps
- ErrStalePlan sentinel enables graceful optimistic concurrency control

## Task Commits

Each task was committed atomically:

1. **Task 1: Plan types and diff computation** - `cbb278ee` (feat)
2. **Task 2: Staleness token generation and validation** - `1e9ccb80` (feat)

## Files Created/Modified
- `api/internal/plan/types.go` - Plan, Change, FieldChange, Action types with JSON tags
- `api/internal/plan/differ.go` - ComputeDiff stateless differ, DesiredInstance input type
- `api/internal/plan/staleness.go` - GenerateToken, ValidateToken, ErrStalePlan sentinel

## Decisions Made

**Stateless differ pattern**
- ComputeDiff receives desired instances + running containers as inputs
- No DB access in differ - keeps it testable and focused
- Caller (plan endpoint handler) queries DB and passes inputs

**Modification scope**
- Only enabled/disabled toggle tracked as modifications
- Config drift detection deferred to future phase
- Source field prepared for Phase 8 wire tracking

**Change ordering**
- Terraform convention: removes → modifies → adds
- Ensures dependencies handled correctly during apply

**Token determinism**
- Sort instances by instance_id before hashing
- RFC3339Nano format for timestamp precision
- SHA256 ensures collision resistance

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

Plan domain foundation complete. Ready for 06-02 (plan endpoint) and 06-03 (apply endpoint).

Exports available:
- Types: Plan, Change, FieldChange, Action
- Differ: ComputeDiff, DesiredInstance
- Staleness: GenerateToken, ValidateToken, InstanceTimestamp, ErrStalePlan

---
*Phase: 06-plan-apply-workflow*
*Completed: 2026-02-07*
