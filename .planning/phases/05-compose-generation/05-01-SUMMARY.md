---
phase: 05-compose-generation
plan: 01
subsystem: compose
tags: [go, compose, yaml, stack, generator]
requires: [phase-04]
provides: [stack-compose-generation, config-materialization, compose-endpoint]
affects: [phase-06-plan-apply, phase-05-02-dashboard]
tech-stack:
  added: []
  patterns: [effective-config-merge, atomic-swap, condition-based-depends-on]
key-files:
  created:
    - api/internal/compose/stack.go
    - api/internal/api/handlers/stack_compose.go
  modified:
    - api/internal/api/routes.go
key-decisions:
  - "Stack compose uses separate types (stackServiceEntry) to avoid touching generator.go"
  - "depends_on: simple list when no healthchecks, condition map when any target has healthcheck"
  - "Identity labels injected via container.BuildLabels, user overrides preserved"
  - "Config files materialized atomically via tmp dir + rename swap"
duration: 2.2min
completed: 2026-02-07
---

# Phase 5 Plan 1: Stack Compose Generator Summary

Stack compose generator transforms DB state (template + instance overrides) into valid multi-service docker-compose YAML with atomic config materialization and dependency-aware depends_on.

## Accomplishments

- GenerateStack method on existing Generator type produces N-service YAML from effective configs
- MaterializeStackConfigs writes config files atomically to compose/stacks/{stack}/{instance}/
- Compose handler at GET /api/v1/stacks/{name}/compose returns JSON with yaml, warnings, instance_count
- depends_on uses simple list for no-healthcheck deps, condition-based map (service_healthy/service_started) when any target has healthcheck
- Identity labels injected via container.BuildLabels (stack_id, instance_id, template_service_id)
- Disabled instances excluded with warnings; dependency refs to disabled/missing instances stripped with warnings
- Port conflicts detected and warned (non-blocking)
- generator.go completely untouched (backward compat COMP-03)

## Task Commits

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Stack compose generator | 9a4c8320 | api/internal/compose/stack.go |
| 2 | Stack compose handler and route wiring | 1bc9edb8 | api/internal/api/handlers/stack_compose.go, api/internal/api/routes.go |

## Files Created

- `api/internal/compose/stack.go` — GenerateStack, MaterializeStackConfigs, effective config loaders
- `api/internal/api/handlers/stack_compose.go` — StackHandler.Compose HTTP handler

## Files Modified

- `api/internal/api/routes.go` — Added GET /stacks/{name}/compose route

## Decisions Made

1. **Separate types for stack compose** — stackServiceEntry with interface{} DependsOn field instead of modifying serviceConfig (which uses []string). Avoids touching generator.go entirely.
2. **Condition-based depends_on logic** — Pure simple deps = string list. Mixed or all-healthcheck = condition map for all entries (simple deps get service_started, healthcheck deps get service_healthy).
3. **Atomic config materialization** — Write to .tmp-{stack} dir, then RemoveAll(final) + Rename(tmp, final). Prevents partial writes.

## Deviations from Plan

None — plan executed exactly as written.

## Issues Encountered

None.

## Next Phase Readiness

- Phase 6 (Plan/Apply) can consume GenerateStack output directly
- 05-02 (Dashboard compose tab) can call GET /stacks/{name}/compose for preview
