# Phase 4 — CLI + Thin Local API

## Goal
Expose the engine through a small operator surface while keeping CLI, API, and engine behavior aligned.

## Acceptance

- CLI supports core workspace, catalog, import, and scan commands
- local API daemon exposes the minimal V2 endpoint set
- JSON output exists where automation needs it
- API handlers are thin wrappers over engine calls
- CLI and API produce equivalent outcomes for the same operations

## Recommended sequence

Start with shared engine-facing service boundaries, then land CLI commands and API endpoints in parallel where safe.

---

### P4-SVC-001 — Define shared application service layer
- **Owner:** architect
- **Depends on:** Phase 3 complete
- **Goal:** create one engine-facing service boundary reused by CLI and API
- **Tasks:**
  1. Define service methods for workspace, catalog, runtime, import, and scan flows.
  2. Keep transport concerns out of service interfaces.
  3. Document parity expectations for CLI and API callers.
- **Validation:**
  - design review
- **Done when:**
  - transport code can stay thin and consistent

### P4-CLI-001 — Implement workspace CLI commands
- **Owner:** surgeon-engine
- **Depends on:** P4-SVC-001
- **Goal:** expose list, open, plan, apply, status, logs, and exec commands
- **Tasks:**
  1. Add `workspace` subcommands.
  2. Support human-readable output first.
  3. Add command-level tests where practical.
- **Validation:**
  - `go test ./cmd/devarch/...`
  - manual CLI smoke checks
- **Done when:**
  - core workspace operations work without the UI

### P4-CLI-002 — Implement catalog/import/scan CLI commands
- **Owner:** surgeon-import
- **Depends on:** P4-SVC-001
- **Goal:** expose catalog, import, and project scan commands
- **Tasks:**
  1. Add `catalog list/show` commands.
  2. Add `import v1-stack`, `import v1-library`, and `scan project` commands.
  3. Keep command output structured and script-friendly where needed.
- **Validation:**
  - `go test ./cmd/devarch/...`
- **Done when:**
  - non-UI operator workflows are reachable from the CLI

### P4-CLI-003 — Add JSON output mode
- **Owner:** surgeon-engine
- **Depends on:** P4-CLI-001, P4-CLI-002
- **Goal:** support machine-readable output for automation
- **Tasks:**
  1. Add a shared JSON output flag approach.
  2. Ensure plan, status, and catalog outputs serialize cleanly.
  3. Document stable output expectations.
- **Validation:**
  - command tests or snapshot tests for JSON output
- **Done when:**
  - automation can consume the CLI without scraping text

### P4-API-001 — Implement catalog and workspace read endpoints
- **Owner:** surgeon-api
- **Depends on:** P4-SVC-001
- **Goal:** expose minimal read endpoints for catalog and workspace data
- **Tasks:**
  1. Add endpoints for templates, workspaces, manifest, graph, status, and plan.
  2. Wrap shared service calls only.
  3. Add handler tests for success and failure cases.
- **Validation:**
  - `go test ./internal/api/...`
- **Done when:**
  - read-only V2 API surface is available and thin

### P4-API-002 — Implement apply, logs, events, and exec endpoints
- **Owner:** surgeon-api
- **Depends on:** P4-API-001
- **Goal:** expose operational endpoints over the runtime-backed engine
- **Tasks:**
  1. Add apply trigger endpoint.
  2. Add logs and event subscription endpoints.
  3. Add exec transport wiring.
- **Validation:**
  - `go test ./internal/api/...`
  - integration smoke tests for apply/status/logs
- **Done when:**
  - API can drive core workspace operations end to end

### P4-DAEMON-001 — Implement local API daemon bootstrap
- **Owner:** surgeon-api
- **Depends on:** P4-API-001
- **Goal:** make the daemon usable as the UI backend
- **Tasks:**
  1. Add config/bootstrap for the local daemon process.
  2. Wire routing, lifecycle, and graceful shutdown.
  3. Document default local usage.
- **Validation:**
  - daemon smoke start test
- **Done when:**
  - UI and external tools have a stable local API process to call

### P4-PAR-001 — Add CLI/API parity checks
- **Owner:** verifier
- **Depends on:** P4-CLI-003, P4-API-002
- **Goal:** prove that CLI and API hit the same underlying behaviors
- **Tasks:**
  1. Define parity scenarios for plan, status, and catalog reads.
  2. Compare outputs at the service/result level.
  3. Record any intentional transport differences.
- **Validation:**
  - parity test suite
- **Done when:**
  - Phase 4 acceptance has evidence, not just intent

## Parallel-safe packets

After `P4-SVC-001`, these can run in parallel with clear file boundaries:

- `P4-CLI-001`
- `P4-CLI-002`
- `P4-API-001`

Do not parallelize `P4-CLI-003` before the underlying command outputs stabilize.

## Handoff to Phase 5

Phase 5 starts only after:

- core flows work from CLI without the UI
- the local API daemon is usable
- parity checks demonstrate shared-engine behavior
