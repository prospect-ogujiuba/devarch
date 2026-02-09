---
phase: 08-service-wiring
plan: 01
subsystem: api, dashboard
tags: [schema, contracts, wiring, foundation]
dependency_graph:
  requires: [migration-018, service-handler, routes]
  provides: [service-exports, service-import-contracts, service-instance-wires, contract-crud-api]
  affects: [service-templates, wiring-engine]
tech_stack:
  added: [migration-019]
  patterns: [template-level-contracts, delete-insert-transaction]
key_files:
  created:
    - api/migrations/019_service_wiring.up.sql
    - api/migrations/019_service_wiring.down.sql
  modified:
    - api/internal/api/handlers/service.go
    - api/internal/api/routes.go
    - dashboard/src/types/api.ts
decisions:
  - Contract CRUD follows existing service handler pattern (DELETE+INSERT transaction)
  - Contracts are template-level only (no instance overrides)
  - service_instance_wires enforces one wire per import contract per consumer per stack via UNIQUE constraint
metrics:
  duration_seconds: 140
  completed_at: "2026-02-08T23:59:24Z"
---

# Phase 08 Plan 01: Service Wiring Schema & Contract CRUD

**One-liner:** Foundation schema for service wiring with template-level export/import contracts and resolved wire persistence, plus contract CRUD API.

## What Was Built

### Migration 019: Service Wiring Tables

Created three tables for the wiring foundation:

**service_exports** — Template-level declarations of what a service provides:
- name, type, port, protocol
- Unique per service_id + name
- Example: postgres template exports `{name: "database", type: "postgres", port: 5432}`

**service_import_contracts** — Template-level declarations of what a service needs:
- name, type, required, env_vars (JSONB)
- Unique per service_id + name
- env_vars maps env var names to template strings (e.g., `{"DB_HOST": "{{hostname}}", "DB_PORT": "{{port}}"}`)
- Example: Laravel template imports `{name: "database", type: "postgres", required: true}`

**service_instance_wires** — Resolved wires between instances:
- Tracks consumer_instance_id → provider_instance_id relationships
- References both import_contract_id and export_contract_id
- source: 'auto' | 'explicit' (informational)
- UNIQUE constraint per stack_id + consumer_instance_id + import_contract_id (one wire per import per consumer per stack)
- Cascade deletes when stack or instances are deleted

### Contract CRUD Endpoints

Added four handlers to service.go following existing patterns:

**GET/PUT /api/v1/services/{name}/exports** — List/update export contracts on template
**GET/PUT /api/v1/services/{name}/imports** — List/update import contracts on template

All use DELETE+INSERT transaction pattern consistent with UpdatePorts, UpdateVolumes, etc.

### TypeScript Types

Added complete wiring type definitions:
- ServiceExport, ServiceImportContract, Wire
- WiringSection, WirePlanEntry, WiringWarning (for plan diagnostics)

## Deviations from Plan

None — plan executed exactly as written.

## Implementation Notes

- JSONB env_vars column uses `COALESCE(env_vars, '{}')` in queries for safe empty object handling
- Protocol defaults to 'tcp' if not provided
- Contracts are immutable per service template — no per-instance overrides (aligns with Phase 8 context decisions)
- Wire table's UNIQUE constraint enforces "one wire per import contract per consumer per stack" invariant

## Verification Results

- Migration 019 applied successfully to running database
- All three tables created with correct schema and indexes
- Go server compiles without errors
- Dashboard TypeScript compiles without errors
- Routes registered at /api/v1/services/{name}/exports and /api/v1/services/{name}/imports

## Self-Check: PASSED

Created files:
- FOUND: api/migrations/019_service_wiring.up.sql
- FOUND: api/migrations/019_service_wiring.down.sql

Modified files:
- FOUND: api/internal/api/handlers/service.go (contract handlers added)
- FOUND: api/internal/api/routes.go (contract routes registered)
- FOUND: dashboard/src/types/api.ts (wiring types added)

Commits:
- FOUND: ac2a87a3 (migration 019)
- FOUND: e4ba1f08 (contract CRUD + types)

Database tables:
- FOUND: service_exports
- FOUND: service_import_contracts
- FOUND: service_instance_wires

## Next Steps

Plan 08-02 will implement the auto-wiring resolver algorithm and integrate wires into the plan/apply workflow.
