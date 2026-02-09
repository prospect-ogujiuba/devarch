---
phase: 08-service-wiring
plan: 02
subsystem: api
tags: [wiring, auto-resolver, contracts, endpoints]
dependency_graph:
  requires: [wiring-schema, service-exports, service-import-contracts, service-instance-wires]
  provides: [wiring-package, auto-wire-resolution, env-injection, wire-crud-api]
  affects: [stack-plan, compose-generation]
tech_stack:
  added: [internal/wiring]
  patterns: [stateless-domain-logic, auto-wire-resolution, env-template-injection]
key_files:
  created:
    - api/internal/wiring/resolver.go
    - api/internal/wiring/env_injector.go
    - api/internal/wiring/validator.go
    - api/internal/api/handlers/stack_wiring.go
  modified:
    - api/internal/api/routes.go
decisions:
  - Auto-wire skips explicit wires (explicit overrides auto per locked decision)
  - Exact type matching for contract resolution (no fuzzy matching)
  - Ambiguous providers (>1 match) left unwired with warning (user must create explicit wire)
  - Circular dependency detection prevents invalid wire graphs
  - Env var injection uses container DNS names (devarch-{stack}-{instance} pattern)
metrics:
  duration_seconds: 153
  completed_at: "2026-02-09T00:03:41Z"
---

# Phase 08 Plan 02: Wiring Domain Package & Wire Management API

**One-liner:** Auto-wire resolution algorithm with 0/1/N provider matching, env var injection using container DNS, and wire CRUD endpoints.

## What Was Built

### Wiring Domain Package (`api/internal/wiring/`)

Created stateless wiring logic package with three modules:

**resolver.go — Auto-wire resolution:**
- `ResolveAutoWires`: Core algorithm that matches import contracts to export contracts by type
- Skips consumer imports that already have explicit wires (explicit overrides auto)
- Handles three provider cases:
  - 0 matches + required: adds warning "Missing required contract"
  - 0 matches + optional: skips silently
  - 1 match: creates auto-wire candidate with injected env vars
  - >1 matches: adds warning "Ambiguous: N providers — create explicit wire"
- Deterministic output sorted by (consumer instance ID, contract name)

**env_injector.go — Env var injection:**
- `InjectEnvVars`: Template replacement using container DNS pattern
- Replaces `{{hostname}}` with `devarch-{stack}-{instance}` (internal DNS)
- Replaces `{{port}}` with container port (not host port)
- Replaces `{{protocol}}` and `{{name}}` with provider values
- Returns resolved env var map for each wire

**validator.go — Wiring validation:**
- `ValidateWiring`: Circular dependency detection using DFS graph traversal
- Warns on self-dependencies (instance depends on itself)
- Returns error if cycle detected, preventing invalid wire graphs
- `FindOrphanedWires`: DB query to find wires referencing deleted instances

### Wire Management HTTP Handlers (`stack_wiring.go`)

Implemented five endpoints under `/api/v1/stacks/{name}/wires`:

**GET /wires (ListWires):**
- Returns all wires for stack with JOIN to get instance names, contract details
- Computes injected env var values using `wiring.InjectEnvVars`
- Includes unresolved imports (contracts with no wire) with reason
- Response: `{wires: [...], unresolved: [...]}`

**POST /wires/resolve (ResolveWires):**
- Loads all exports and imports for stack instances
- Calls `wiring.ResolveAutoWires` to compute candidates
- Validates with `wiring.ValidateWiring` (rejects if circular)
- Deletes existing auto-wires, inserts new candidates in transaction
- Returns `{resolved: N, warnings: [...]}`

**POST /wires (CreateWire):**
- Accepts `{consumer_instance_id, provider_instance_id, import_contract_name}`
- Resolves instance names to PKs, finds matching contracts
- Validates: same stack, type match, no self-dependency
- Deletes existing wire for same (consumer, import contract) if any
- Inserts with source='explicit'
- Returns created wire with 201 status

**DELETE /wires/{wireId} (DeleteWire):**
- Deletes wire by ID, verifies it belongs to correct stack
- Returns 204 on success, 404 if not found

**POST /wires/cleanup (CleanupOrphanedWires):**
- Calls `wiring.FindOrphanedWires` to find wires to deleted instances
- Deletes orphaned wires
- Returns `{deleted: N}`

### Route Registration

Added wire routes in `routes.go` inside `/stacks/{name}` block:
- Registered `/wires/resolve` and `/wires/cleanup` BEFORE `/wires` to avoid chi parameter conflicts
- Pattern matches existing trash route registration (specific before parameterized)

## Deviations from Plan

None — plan executed exactly as written.

## Implementation Notes

- Wiring package is stateless — receives data from handlers, returns results
- DB queries happen in handler layer, not domain layer (except `FindOrphanedWires` which inherently needs DB)
- Auto-wire resolution uses exact type matching (postgres == postgres, mysql != postgres)
- Explicit wires tracked via wireKey map to skip during auto-resolution
- CreateWire uses DELETE+INSERT transaction to replace existing wire for same import contract
- ListWires computes injected env vars on-the-fly for each wire (not stored in DB)

## Verification Results

- `go build ./cmd/server` passes without errors
- All 5 wire endpoints registered in routes.go
- Wiring package has resolver.go, env_injector.go, validator.go
- Auto-wire handles 0/1/N provider cases correctly
- Env injection uses container DNS pattern (devarch-{stack}-{instance})

## Self-Check: PASSED

Created files:
- FOUND: api/internal/wiring/resolver.go
- FOUND: api/internal/wiring/env_injector.go
- FOUND: api/internal/wiring/validator.go
- FOUND: api/internal/api/handlers/stack_wiring.go

Modified files:
- FOUND: api/internal/api/routes.go (wire routes registered)

Commits:
- FOUND: 8d00a836 (wiring domain package)
- FOUND: 84395672 (wire management HTTP handlers)

Route registration:
- FOUND: GET /wires → stackHandler.ListWires
- FOUND: POST /wires/resolve → stackHandler.ResolveWires
- FOUND: POST /wires/cleanup → stackHandler.CleanupOrphanedWires
- FOUND: POST /wires → stackHandler.CreateWire
- FOUND: DELETE /wires/{wireId} → stackHandler.DeleteWire

## Next Steps

Plan 08-03 will integrate wiring into the plan/apply workflow and add wire visualization in the dashboard.
