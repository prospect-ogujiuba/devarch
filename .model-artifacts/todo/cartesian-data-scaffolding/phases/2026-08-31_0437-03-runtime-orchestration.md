# Phase 3: Runtime orchestration

Created: 2026-08-30
Purpose: Make declared data services part of the existing isolated Node application lifecycle.

## Goal

Have `scripts/node/bootstrap.sh` safely discover an application's generated manifest, start required shared services, wait for health, and then start the app.

## Scope

- Parse `.devarch/data-services.json` with Python/Node already required by the runtime.
- Validate manifest version and service ids against the repository registry.
- Ensure `microservices-net`, start selected Compose files, and wait for health before the Node app.
- Extend dry-run output with required services and exact allowlisted Compose targets.
- Define environment loading consistently for Next.js and future server frameworks.
- Preserve existing apps with no manifest: they start exactly as today.
- Ensure a failed dependency does not leave a falsely reported ready application.

## Safety invariants

- Never accept an arbitrary Compose path from generated application JSON.
- Runtime maps service ids to repository-owned manifests.
- `--force` replaces application files but does not drop persistent service volumes or databases.
- Shared services are not stopped when one application stops.
- Health timeout/error output identifies the failed service and inspection command.

## Outputs

- Manifest-aware Node bootstrap.
- Service startup and readiness helpers.
- Optional app-container command helper for migrations, or a documented Compose exec invocation.
- Runtime tests covering absent, valid, corrupt, unknown, and unavailable dependencies.

## Acceptance criteria

- Starting a PostgreSQL+Redis app starts both dependencies before the app.
- A second app reuses healthy shared services idempotently.
- Existing non-data apps have no new required files or services.
- Unknown ids and schema versions fail before Compose mutation.
- Dry-run remains mutation-free.

## Verification

- `bash scripts/node/bootstrap.test.sh`
- Existing routing tests.
- Podman/Docker integration test with two apps sharing PostgreSQL and Redis.
- Failure test with one deliberately unhealthy dependency.

## Open questions

- Per-app logical databases/users versus shared development credentials.
- Whether migrations run automatically on startup or only through an explicit command; default recommendation is explicit.

## Non-goals

Stopping shared databases automatically, deleting volumes, production orchestration, or migration rollback automation.
