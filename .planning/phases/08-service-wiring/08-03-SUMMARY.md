---
phase: 08-service-wiring
plan: 03
subsystem: api
tags: [wiring, plan-integration, compose-generation, export-import]
dependency_graph:
  requires: [wiring-package, auto-wire-resolution, wire-crud-api]
  provides: [plan-wiring-section, compose-wire-injection, export-wire-types]
  affects: [stack-plan, compose-generation, effective-config, devarch-yml]
tech_stack:
  added: []
  patterns: [three-layer-env-merge, wire-derived-dependencies, idempotent-wire-import]
key_files:
  created: []
  modified:
    - api/internal/plan/types.go
    - api/internal/api/handlers/stack_plan.go
    - api/internal/compose/stack.go
    - api/internal/api/handlers/instance_effective.go
    - api/internal/export/types.go
    - api/internal/export/exporter.go
    - api/internal/export/importer.go
decisions:
  - Plan endpoint triggers auto-wire resolution at plan time (DELETE old auto, INSERT new candidates)
  - Three-layer env merge: template -> wired -> instance overrides (WIRE-08 compliance)
  - Wire-derived dependencies appended to user-defined dependencies (additive, not replacement)
  - Effective config exposes wired_env_vars separately for UI visual distinction
  - Export includes only wires where both consumer and provider are active instances
  - Import uses ON CONFLICT DO UPDATE for idempotent wire recreation
metrics:
  duration_seconds: 199
  completed_at: "2026-02-09T00:09:12Z"
---

# Phase 08 Plan 03: Wiring Integration (Plan, Compose, Export/Import)

**One-liner:** Wiring integrated into plan generation, compose env injection with three-layer merge, and typed wire export/import round-trip.

## What Was Built

### Plan Wiring Section (WIRE-06)

Added `WiringSection` to plan response with active wires and warnings:

**Plan types (`plan/types.go`):**
- `WiringSection` with `ActiveWires []WirePlanEntry` and `Warnings []WiringWarning`
- `WirePlanEntry` shows consumer/provider instances, contract details, source, and injected env vars
- `WiringWarning` for displaying resolution issues (missing required, ambiguous providers)

**Plan handler updates (`stack_plan.go`):**
- After computing diff, calls `resolveAndBuildWiring` to run auto-wire resolution
- Loads providers (enabled instances with exports), consumers (enabled instances with imports), existing wires
- Calls `wiring.ResolveAutoWires` to compute auto-wire candidates
- Transaction: DELETE old auto wires, INSERT new candidates, cleanup orphaned wires
- Builds `WiringSection` from all active wires (auto + explicit) with injected env vars computed on-the-fly
- Appends wiring warnings to plan warnings array

**Key behavior:** Auto-wire resolution happens at plan time, not apply time. This makes wiring deterministic and visible before deployment.

### Compose Wire Injection (WIRE-07, WIRE-08)

Modified compose generation to inject wire-derived env vars and dependencies:

**Three-layer env merge (`compose/stack.go`):**
1. Load template env vars from `service_env_vars`
2. Load wired env vars via `loadWiredEnvVarsForInstance` (queries `service_instance_wires`, computes env var injections using `wiring.InjectEnvVars`)
3. Load instance override env vars from `instance_env_vars`

Merge order ensures instance overrides win (WIRE-08 requirement).

**Wire-derived dependencies:**
- Added `loadWireDependencies` to query provider instances from wires
- Appends wire-derived dependencies to user-defined dependencies (additive)
- Provider instances become `depends_on` targets in compose YAML

**Result:** Containers receive env vars like `DB_HOST=devarch-mystack-postgres`, `DB_PORT=5432` automatically from wires. Consumer containers depend on provider containers for startup ordering.

### Effective Config Wired Env Vars

Updated effective config endpoint to show wire-injected env vars:

**Response changes (`instance_effective.go`):**
- Added `WiredEnvVars map[string]string` field to response
- Uses same three-layer merge as compose generation
- Loads wired env vars via `loadWiredEnvVarsForEffective` (duplicate of compose logic, scoped to HTTP handler)
- UI can distinguish which env vars came from wires vs template vs instance overrides

**Use case:** Dashboard can visually indicate wire-injected env vars with blue badges or distinct styling in Phase 8 Plan 04.

### Typed Wire Export/Import

Replaced `[]interface{}` stub with fully typed wire definitions:

**Types (`export/types.go`):**
```go
type WireDef struct {
    ConsumerInstance string
    ProviderInstance string
    ImportContract   string
    ExportContract   string
    Source           string
}
```

**Export (`exporter.go`):**
- Queries `service_instance_wires` with JOINs for contract names
- Only includes wires where both consumer and provider are active (not soft-deleted)
- Stores instance name → PK map to filter orphaned wires

**Import (`importer.go`):**
- After importing instances, processes `file.Wires`
- For each wire: resolves instance names to PKs, finds import/export contract IDs by name
- Uses `ON CONFLICT (stack_id, consumer_instance_id, import_contract_id) DO UPDATE` for idempotent import
- Graceful error handling: missing contracts or instances added to `result.Errors`, wire skipped
- Preserves wire source ('auto' or 'explicit')

**Result:** Export + import round-trip fully preserves wiring state. Shared stacks via devarch.yml include wiring definitions.

## Deviations from Plan

None — plan executed exactly as written.

## Implementation Notes

**Auto-wire at plan time:** Plan endpoint is now stateful — it writes to DB (auto wires). This makes plan generation deterministic and avoids drift between plan preview and apply execution.

**Duplicate wire env loading:** `loadWiredEnvVarsForInstance` (compose) and `loadWiredEnvVarsForEffective` (HTTP handler) have identical logic. Not DRY, but isolated by layer (compose package vs handlers package). Could extract to shared function in future refactor.

**Wire dependencies are additive:** User-defined dependencies preserved, wire-derived dependencies appended. No removal or replacement. If user explicitly sets `depends_on`, wires add to it.

**Export filters by active instances:** Wires referencing soft-deleted instances excluded from export. Import can recreate wires only if both consumer and provider exist in imported stack.

**Import idempotency:** ON CONFLICT ensures re-importing same devarch.yml doesn't create duplicate wires. Updates provider if wire for same consumer+import already exists.

## Verification Results

- `go build ./cmd/server` passes without errors
- Plan response includes `wiring` field with `active_wires` and `warnings`
- Compose env vars include wire-injected values (template -> wired -> instance merge order confirmed)
- Effective config shows `wired_env_vars` separately
- Export produces typed `WireDef` entries (not empty `[]interface{}`)
- Import recreates wires with graceful error handling

## Self-Check: PASSED

Modified files:
- FOUND: api/internal/plan/types.go (WiringSection types added)
- FOUND: api/internal/api/handlers/stack_plan.go (resolveAndBuildWiring method)
- FOUND: api/internal/compose/stack.go (three-layer env merge, loadWiredEnvVarsForInstance, loadWireDependencies)
- FOUND: api/internal/api/handlers/instance_effective.go (wired_env_vars response field)
- FOUND: api/internal/export/types.go (WireDef type)
- FOUND: api/internal/export/exporter.go (loadWires method)
- FOUND: api/internal/export/importer.go (importWires method)

Commits:
- FOUND: 0dfb3374 (Task 1: plan, compose, effective config)
- FOUND: 3a432e93 (Task 2: export/import)

Key behaviors:
- Plan endpoint runs auto-wire resolution ✓
- Compose includes wire-injected env vars ✓
- Three-layer merge order correct (template -> wired -> instance) ✓
- Wire-derived dependencies appended to compose ✓
- Export includes typed WireDef ✓
- Import recreates wires with ON CONFLICT idempotency ✓

## Next Steps

Plan 08-04 will build the dashboard UI for wiring visualization (Wiring tab on stack detail, wire list, wire creation/deletion forms).
