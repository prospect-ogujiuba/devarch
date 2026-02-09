---
phase: 09-secrets-resources
plan: 02
subsystem: security,resources
tags: [secret-redaction, resource-limits, compose-deploy, plan-output]
dependency_graph:
  requires: [09-01 encryption foundation]
  provides: [secret redaction in outputs, resource limits CRUD, deploy.resources compose section]
  affects: [compose generator, plan output, export handler, instance overrides API]
tech_stack:
  added: [deploy.resources YAML structure, memory string parser, resource validation]
  patterns: [belt-and-suspenders redaction, UPSERT with cleanup, validation warnings]
key_files:
  created: []
  modified:
    - api/internal/compose/stack.go
    - api/internal/api/handlers/stack_plan.go
    - api/internal/api/handlers/stack_compose.go
    - api/internal/export/exporter.go
    - api/internal/api/handlers/instance_overrides.go
    - api/internal/api/routes.go
    - api/internal/plan/types.go
decisions:
  - Compose preview uses redaction mode, apply uses real values for container runtime
  - Secret redaction: belt-and-suspenders (is_secret flag + keyword heuristic)
  - Resource limits: DELETE row when all fields empty (clean up nulls)
  - Validation warnings never block operations (always warn-only)
  - Memory parser supports k/m/g suffixes (case-insensitive)
metrics:
  duration: 219
  completed: 2026-02-09T04:01:23Z
---

# Phase 09 Plan 02: Secret Redaction & Resource Limits Summary

**One-liner:** Secret redaction in compose preview/plan/export outputs + resource limits CRUD with deploy.resources compose integration and validation warnings.

## What Was Built

### Task 1: Secret Redaction in Outputs

**Compose Generator (`stack.go`):**
- `GenerateStackWithRedaction(stackName, redactSecrets bool)` replaces plaintext `GenerateStack`
- `loadEffectiveEnvVarsWithSecrets` tracks `is_secret` flag from `service_env_vars` and `instance_env_vars`
- Three-layer merge preserves secret flags: template → wired → instance overrides
- Redaction mode replaces secret values with `***` in environment map before YAML generation
- Apply flow uses `GenerateStack()` (defaults to `redactSecrets=false`) for real values

**Compose Handler (`stack_compose.go`):**
- Calls `GenerateStackWithRedaction(stackName, true)` for preview/download endpoint
- Apply handler unchanged — uses real values for `podman compose up`

**Plan Handler (`stack_plan.go`):**
- Redacts `InjectedEnvVars` in `WirePlanEntry` using `export.IsSecretKey()` heuristic
- Checks injected env var keys against keyword list (password, secret, key, token, etc.)
- Redacted keys replaced with `***` in plan response

**Export Handler (`exporter.go`):**
- Queries `is_secret` column from both `service_env_vars` and `instance_env_vars`
- Belt-and-suspenders approach: `is_secret=true` → `${SECRET:KEY_NAME}` placeholder
- Also applies keyword heuristic via `RedactSecrets()` (catches non-flagged secrets)
- Both layers ensure comprehensive coverage

### Task 2: Resource Limits CRUD + Compose Integration

**Instance Overrides Handler (`instance_overrides.go`):**
- `GetResourceLimits` — returns CPU/memory limits + reservations or 404 if none set
- `UpdateResourceLimits` — UPSERT pattern with cleanup:
  - All fields empty → DELETE row (clean up)
  - At least one field → INSERT ... ON CONFLICT DO UPDATE
  - `NULLIF($n, '')` converts empty strings to NULL in DB
- `parseMemoryString` — parses "4m", "512m", "1g", "2G" (case-insensitive)
- Validation warnings (never blocking):
  - Memory < 4MB: "memory limit very low, container may fail to start"
  - CPU < 0.01: "CPU limit extremely low"

**Routes (`routes.go`):**
```
GET  /api/v1/stacks/{name}/instances/{instance}/resources
PUT  /api/v1/stacks/{name}/instances/{instance}/resources
```

**Compose Generator (`stack.go`):**
- `deployConfig` / `resourcesConfig` / `resourceLimits` YAML types
- `loadResourceLimits(instancePK)` queries `instance_resource_limits` table
- Returns nil if no row or all fields null (omits `deploy:` section)
- Populates `deploy.resources.limits` (CPUs, Memory) and/or `deploy.resources.reservations`
- Integrated into `GenerateStackWithRedaction` loop — loads per instance

**Plan Output (`stack_plan.go`, `types.go`):**
- `ResourceLimitEntry` type with CPU/memory fields
- `loadResourceLimitsForStack` queries all instances in stack
- Plan response includes `resource_limits` map: `instance_id -> ResourceLimitEntry`
- Enables dashboard validation display before apply

## Request/Response Examples

**PUT resource limits:**
```json
{
  "cpu_limit": "2.0",
  "cpu_reservation": "0.5",
  "memory_limit": "1g",
  "memory_reservation": "512m"
}
```

**Response:**
```json
{
  "status": "updated",
  "cpu_limit": "2.0",
  "cpu_reservation": "0.5",
  "memory_limit": "1g",
  "memory_reservation": "512m",
  "warnings": []
}
```

**Compose YAML output (with limits):**
```yaml
services:
  postgres:
    image: postgres:16
    container_name: devarch-mystack-postgres
    deploy:
      resources:
        limits:
          cpus: "2.0"
          memory: "1g"
        reservations:
          cpus: "0.5"
          memory: "512m"
```

## Deviations from Plan

None — plan executed exactly as written.

## Verification Results

**Compilation:** ✅ `go build ./...` passes

**Secret redaction coverage:**
- Compose preview: env vars with `is_secret=true` → `***`
- Plan output: wired env vars matching keyword heuristic → `***`
- Export YAML: DB-flagged secrets + keyword matches → `${SECRET:KEY}`

**Resource limits:**
- CRUD endpoints registered at `/instances/{instance}/resources`
- Compose YAML includes `deploy.resources` section when limits exist
- No `deploy` section when no limits configured (clean YAML)
- Plan output includes `resource_limits` map for validation

## Technical Notes

**Belt-and-suspenders redaction:** Two-layer approach ensures comprehensive coverage:
1. DB `is_secret` flag (explicit user intention)
2. Keyword heuristic (catches non-flagged secrets like "DB_PASSWORD")

**Compose redaction split:** Preview endpoint uses `redactSecrets=true`, apply uses `redactSecrets=false`. This allows:
- Users see redacted YAML in UI (security)
- Runtime receives real values for containers (functionality)

**Resource limit cleanup:** DELETE row when all fields empty avoids storing rows with all NULLs. Compose generator returns nil deploy config when no limits, preventing empty `deploy:` sections in YAML.

**Memory parsing:** Case-insensitive suffix handling (k/K, m/M, g/G) matches Docker/Podman conventions. Validation warns but never blocks (user may know better).

**Plan validation:** Resource limits in plan output enable dashboard to show warnings before apply (e.g., "postgres has memory_limit < 4MB — may fail to start").

## Next Steps (Plan 03)

- Dashboard UI for secret input fields (password type, show/hide toggle)
- Dashboard UI for resource limits configuration
- Testing with actual secret values and resource limits
- End-to-end verification of compose apply with deploy.resources

## Self-Check: PASSED

All modified files exist:
- ✅ api/internal/compose/stack.go
- ✅ api/internal/api/handlers/stack_plan.go
- ✅ api/internal/api/handlers/stack_compose.go
- ✅ api/internal/export/exporter.go
- ✅ api/internal/api/handlers/instance_overrides.go
- ✅ api/internal/api/routes.go
- ✅ api/internal/plan/types.go

All commits exist:
- ✅ 136c5342: feat(09-02): add secret redaction to compose, plan, and export outputs
- ✅ 15f1d72a: feat(09-02): add resource limits CRUD and compose deploy.resources
