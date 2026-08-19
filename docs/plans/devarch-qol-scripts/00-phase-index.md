# DevArch quality-of-life scripts phase index

Created: 2026-08-19
Purpose: Define the implementation order and shared contracts for eighteen operational scripts without recreating the removed DevArch application or daemon.

## Outcome

Add small Bash entrypoints under `scripts/devarch/` plus shared, testable helpers. Compose files remain the source of truth; scripts discover and operate on them rather than maintaining a second service database.

## Cross-cutting constraints

- Bash 4+, `set -Eeuo pipefail`, `LC_ALL=C`, arrays for commands, and no `eval`.
- Support Podman (`podman compose` or `podman-compose`) and Docker Compose through one runtime adapter.
- Resolve repository paths from `BASH_SOURCE`, not the caller's working directory.
- Accept canonical service IDs as `category/name`; permit a short name only when unique.
- Mutating/destructive commands provide `--dry-run`; destructive cleanup/reset/restore additionally requires explicit confirmation or `--yes`.
- Never print secrets. Parse only allowlisted `.env` keys without evaluating shell syntax.
- Human-readable output is the baseline; machine-readable `--json` is required where another script consumes results.
- Host-only regression tests use temporary fixture repositories and rejecting/fake runtimes by default.
- No daemon, API, state database, package manager, or hidden background process.

## Plan tracker

Pick any unchecked feature whose dependencies are complete. When a phase satisfies its acceptance criteria and verification, change its checkbox to `[x]` and add `Status: Complete` beneath the `Created:` line in that phase file. Phase 01 is shared prerequisite work, not one of the eighteen feature ideas.

- [ ] `01-shared-foundation.md` — discovery, runtime, output, safety, and fixture contracts. Dependencies: none.
- [ ] `02-doctor.md` — environment and configuration diagnostics. Dependencies: 01.
- [ ] `03-service.md` — single-service lifecycle operations. Dependencies: 01.
- [ ] `04-status.md` — consolidated runtime state and endpoint inventory. Dependencies: 01, 03.
- [ ] `05-stack.md` — declarative multi-service recipes. Dependencies: 01, 03, 04.
- [ ] `06-app.md` — framework-aware application operations. Dependencies: 01, 03.
- [ ] `07-validate.md` — static validation and regression orchestration. Dependencies: 01.
- [ ] `08-backup-and-restore.md` — durable app/database backups and restores. Dependencies: 01, 06, 12.
- [ ] `09-cleanup.md` — conservative stale-resource cleanup. Dependencies: 01, 08.
- [ ] `10-update-report.md` — read-only image update reporting. Dependencies: 01.
- [ ] `11-support-bundle.md` — redacted diagnostic bundles. Dependencies: 02, 04, 07, 17.
- [ ] `12-db.md` — database connection and lifecycle helpers. Dependencies: 01, 03, 06.
- [ ] `13-logs.md` — service/stack log aggregation. Dependencies: 03, 04, 05, 06.
- [ ] `14-open.md` — endpoint resolution and browser launching. Dependencies: 04, 06.
- [ ] `15-env-init.md` — safe environment initialization and secret generation. Dependencies: 01, 02.
- [ ] `16-certs.md` — local wildcard certificate inspection/generation/trust. Dependencies: 01, 02.
- [ ] `17-ports.md` — static and runtime port inventory/conflict detection. Dependencies: 01.
- [ ] `18-new-service.md` — validated service scaffolding. Dependencies: 07, 17.
- [ ] `19-shell-completions.md` — generated completions for dynamic names. Dependencies: stable list interfaces from implemented phases.

## Recommended delivery waves

- **Wave 1, substrate:** phases 01–04 and 07. Establish safe discovery, lifecycle, diagnostics, status, and validation.
- **Wave 2, daily workflows:** phases 05–06, 12–14, and 17.
- **Wave 3, data safety:** phases 08–09, then 11.
- **Wave 4, setup and maintenance:** phases 10, 15–16, 18–19.

## Integration map

- `doctor`, `validate`, `ports`, and `new-service` share the static Compose catalog.
- `service`, `status`, `stack`, `logs`, and `open` share runtime discovery.
- `app`, `db`, `backup/restore`, and `cleanup` share app/database identity rules.
- `support-bundle` consumes bounded/redacted output from `doctor`, `status`, `ports`, and runtime commands.
- Completions consume only stable list modes; they never start the runtime.

## Program-level definition of done

- All eighteen ideas have documented and tested commands.
- Existing WordPress, Laravel, and hosts regression suites still pass.
- A full host-only test command performs no container, network, hosts-file, browser, certificate-store, or database mutations.
- README documents discovery rules, safety behavior, supported platforms, and examples.
- ShellCheck and syntax checks pass for every added Bash file.

## Review order

Review the index and phase 01 first because every later command depends on those contracts. Then review each delivery wave in order; phases inside a wave may be reprioritized after the foundation exists.
