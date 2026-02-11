---
phase: 21-deploy-orchestration-service
plan: 01
subsystem: api
tags: [service-layer, orchestration, refactoring]
dependency_graph:
  requires: [plan, wiring, compose, container, lock]
  provides: [orchestration-service]
  affects: []
tech_stack:
  added: []
  patterns: [service-layer-extraction, sentinel-errors]
key_files:
  created:
    - api/internal/orchestration/service.go
    - api/internal/orchestration/errors.go
  modified: []
decisions:
  - "Service accepts Go types only — no net/http imports for transport independence"
  - "Sentinel errors enable handlers to map service errors to HTTP status codes"
  - "loadProviders vs loadAllProviders distinction: plan generation uses enabled filter, wiring resolution includes disabled instances"
metrics:
  duration_seconds: 147
  completed_at: 2026-02-11T19:43:59Z
---

# Phase 21 Plan 01: Deploy Orchestration Service Summary

Transport-agnostic service layer encapsulating deploy orchestration logic: plan generation, plan application, and wiring resolution.

## Objective

Extract deploy orchestration business logic from HTTP handlers into reusable service layer with three public methods (GeneratePlan, ApplyPlan, ResolveWiring).

## Implementation Summary

Created `api/internal/orchestration/` package with two files:

**errors.go** — Sentinel error definitions:
- ErrStackNotFound, ErrStackDisabled, ErrStalePlan, ErrLockConflict, ErrProjectRoot, ErrValidation
- Enable clean error-to-HTTP-status mapping in handlers (Plan 02)

**service.go** — Service with dependencies (db, containerClient):

**GeneratePlan(stackName string)** extracts from stack_plan.go:
1. Query stack (id, network_name, updated_at, enabled)
2. Query instances (instance_id, template_name, container_name, enabled, updated_at)
3. Build desired instances slice and timestamps
4. List running containers (graceful failure: empty on error)
5. Compute diff, generate token
6. Resolve and build wiring (auto-wires persisted)
7. Load resource limits
8. Return plan.Plan with changes, token, wiring, resource limits, warnings

**ApplyPlan(ctx, stackName, token, lockFile)** extracts from stack_apply.go:
1. Query stack (id, network_name, enabled)
2. Acquire advisory lock (returns ErrLockConflict if not acquired)
3. Defer advisory unlock with 5s timeout
4. Validate token (wraps plan.ErrStalePlan as ErrStalePlan)
5. Resolve network name (custom or devarch-{name}-net)
6. Create network
7. Read PROJECT_ROOT, HOST_PROJECT_ROOT, WORKSPACE_ROOT env vars
8. Generate compose YAML, materialize configs
9. Write YAML to temp file, run compose up
10. Validate lock file if provided
11. Return ApplyResult with output and optional lock warnings

**ResolveWiring(stackName)** extracts from stack_wiring.go:
1. Query stack ID
2. Load providers (no enabled filter — includes disabled instances)
3. Load consumers (no enabled filter)
4. Load existing wires
5. Resolve auto-wires
6. Validate wiring (wraps validation errors with ErrValidation)
7. Transaction: delete auto wires, insert candidates, commit
8. Return ResolveResult with resolved count and warnings

**Private helpers** (8 methods):
- loadResourceLimitsForStack
- resolveAndBuildWiring (used by GeneratePlan)
- loadProviders (enabled filter — for plan generation)
- loadAllProviders (no enabled filter — for wiring resolution)
- loadConsumers (enabled filter — for plan generation)
- loadAllConsumers (no enabled filter — for wiring resolution)
- loadExistingWires
- loadAllWires (with env var injection and secret redaction)

## Key Links Verified

✅ `orchestration.Service` → `plan.ComputeDiff`, `plan.GenerateToken`, `plan.ValidateToken`
✅ `orchestration.Service` → `wiring.ResolveAutoWires`, `wiring.ValidateWiring`
✅ `orchestration.Service` → `compose.NewGenerator`
✅ No `net/http` imports — transport-agnostic

## Deviations from Plan

None — plan executed exactly as written.

## Tasks Completed

| Task | Description | Commit | Files |
|------|-------------|--------|-------|
| 1 | Create orchestration service with GeneratePlan, ApplyPlan, ResolveWiring | 5f58e23 | api/internal/orchestration/service.go, api/internal/orchestration/errors.go |

## Verification Results

1. ✅ `go build ./internal/orchestration/` compiles successfully
2. ✅ Service struct has db and containerClient fields
3. ✅ Three public methods exist: GeneratePlan, ApplyPlan, ResolveWiring
4. ✅ `grep -r "net/http" api/internal/orchestration/` returns nothing
5. ✅ errors.go defines all sentinel errors

## Success Criteria Met

- [x] Orchestration package compiles independently
- [x] Three public methods contain all business logic currently in handlers
- [x] Service is transport-agnostic (no HTTP imports)
- [x] Sentinel errors enable clean error-to-status-code mapping in handlers

## Next Steps

Plan 02: Refactor handlers to use orchestration service (thin HTTP wrappers calling service methods).

## Self-Check: PASSED

Created files exist:
```bash
FOUND: api/internal/orchestration/service.go
FOUND: api/internal/orchestration/errors.go
```

Commit exists:
```bash
FOUND: 5f58e23
```
