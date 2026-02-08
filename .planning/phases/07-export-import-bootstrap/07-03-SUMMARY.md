---
phase: 07-export-import-bootstrap
plan: 03
subsystem: export-import
tags: [lockfile, digest-resolution, drift-detection, sha256, podman, docker]

# Dependency graph
requires:
  - phase: 07-01
    provides: "Export domain with DevArchFile YAML generation"
  - phase: 07-02
    provides: "Import domain with YAML parsing and DB insertion"
provides:
  - "Lockfile generation with runtime-resolved image digests"
  - "Template version hashing for drift detection"
  - "Config integrity SHA256 hashing"
  - "Lock validation comparing expected vs actual runtime state"
  - "Apply endpoint lock warnings (non-blocking)"
affects: [07-04, EXIM]

# Tech tracking
tech-stack:
  added: [crypto/sha256, podman/docker image inspect]
  patterns: [lockfile-as-JSON, runtime-resolved-digests, warn-only-validation]

key-files:
  created:
    - api/internal/lock/types.go
    - api/internal/lock/generator.go
    - api/internal/lock/validator.go
    - api/internal/lock/integrity.go
    - api/internal/api/handlers/stack_lock.go
  modified:
    - api/internal/api/handlers/stack_apply.go
    - api/internal/api/routes.go

key-decisions:
  - "Lockfile as JSON (not YAML) per phase design decision"
  - "Image digests resolved via container runtime inspect (not registry) for offline reproducibility"
  - "Template hash: SHA256(template_name + created_at) truncated to 16 hex chars"
  - "Lock validation is warn-only in apply response, never blocks"
  - "Empty digest tolerated (non-fatal) if image not yet pulled"

patterns-established:
  - "Lock generation: resolves runtime state (digests, ports, network ID) from running containers"
  - "Lock validation: compares lock against current runtime, returns typed warnings"
  - "Apply integration: optional lock field in request, warnings in response"

# Metrics
duration: 173s
completed: 2026-02-08
---

# Phase 07 Plan 03: Lockfile Domain Summary

**JSON lockfile with runtime-resolved image digests, template version hashing, and warn-only drift detection integrated into apply workflow**

## Performance

- **Duration:** 2m 53s
- **Started:** 2026-02-08T20:15:20Z
- **Completed:** 2026-02-08T20:18:13Z
- **Tasks:** 2
- **Files modified:** 7

## Accomplishments
- Lock domain package with digest resolution via podman/docker image inspect
- Template versioning via SHA256 hash of name + created_at for drift detection
- Lock validation comparing runtime state against lockfile expectations
- Three lock endpoints: generate, validate, refresh
- Apply integration with optional lock_warnings in response (non-blocking)

## Task Commits

Each task was committed atomically:

1. **Task 1: Lock domain package** - `6c943e2f` (feat)
2. **Task 2: Lock HTTP handlers and apply integration** - `19750f0d` (feat)

## Files Created/Modified
- `api/internal/lock/types.go` - LockFile, StackLock, InstLock, LockWarning, ValidationResult JSON structs
- `api/internal/lock/generator.go` - Lock generator resolving digests via container inspect, computing template hashes
- `api/internal/lock/validator.go` - Lock validator comparing lock state against runtime, returning typed warnings
- `api/internal/lock/integrity.go` - SHA256 hash utilities for config content and files
- `api/internal/api/handlers/stack_lock.go` - GenerateLock, ValidateLock, RefreshLock HTTP handlers
- `api/internal/api/handlers/stack_apply.go` - Optional lock validation in apply response
- `api/internal/api/routes.go` - Lock routes: POST /lock, /lock/validate, /lock/refresh

## Decisions Made
- Image digest resolution via container runtime (podman/docker inspect) rather than registry lookups - enables offline reproducibility
- Template hash = SHA256(name + created_at) truncated to 16 chars - detects template version drift
- Empty digest is non-fatal (image may not be pulled yet) - generator returns empty string, validator tolerates it
- Lock validation integrated into apply as optional warn-only feature - never blocks apply operations per phase design

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Lockfile generation and validation complete
- Export/import with lock support ready for Phase 07-04 (CLI commands)
- Foundation for deterministic reproduction via manifest + lockfile pattern established

## Self-Check: PASSED

All created files verified:
- api/internal/lock/types.go
- api/internal/lock/generator.go
- api/internal/lock/validator.go
- api/internal/lock/integrity.go
- api/internal/api/handlers/stack_lock.go

All commits verified:
- 6c943e2f (Task 1)
- 19750f0d (Task 2)
