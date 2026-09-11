# ENV-002 — Migrate Laravel to the shared dotenv loader

- Contract ID: `ENV-002`
- Parent: `PHASE-01`
- Kind: implementation
- Status: planned
- Plan revision: 2
- Spec: `.model-artifacts/initiatives/env-configuration-cleanup/specs/2026-09-11_1625-initiative-spec-r2.md`
- Plan: `.model-artifacts/initiatives/env-configuration-cleanup/plans/2026-09-11_1625-plan-index-r2.md`

## Goal and observable behavior

Remove Laravel's private dotenv parser in favor of the shared tested loader without changing its six-key allowlist, parsing, precedence, validation, dry-run, or fixture behavior.

## Scope

- Source the shared loader from Laravel bootstrap.
- Preserve the exact allowlist: `LARAVEL_APP_NAME`, `LARAVEL_APP_URL`, `LARAVEL_DB_ROOT_PASSWORD`, `MARIADB_ROOT_PASSWORD`, `LARAVEL_CONTAINER_USER`, `CONTAINER_RUNTIME`.
- Preserve file-over-inherited precedence and ignored unrelated entries.
- Adopt the same `DEVARCH_ENV_FILE` semantics as WordPress; the control is not loadable from dotenv.
- Update isolated Laravel test fixtures to copy/select the shared loader.
- Remove superseded private parser code only after parity is green.

## Non-goals

No Laravel provisioning, generated application `.env`, package/profile, Compose, database, or runtime behavior changes.

## Dependencies and entry inputs

- Depends on: `ENV-001`.
- Entry inputs: green shared-loader tests and existing Laravel parser characterization tests.
- Required capabilities: Bash refactoring and regression testing.

## Expected files/outputs

- `scripts/laravel/bootstrap.sh`
- `scripts/laravel/tests/bootstrap_test.sh`
- `scripts/laravel/bootstrap.test.sh`
- Shared loader/tests only if parity exposes an in-contract defect.

## Specialist decisions incorporated

- TDD: existing parser characterization is the baseline; add missing path/fixture parity test before removing code.
- Security: maintain allowlist-first assignment and control-byte handling.
- Compatibility/migration: preserve all supported values and do not touch generated application dotenv files.

## Acceptance criteria

1. AC1–AC4 pass for Laravel.
2. The private parser is removed only after shared-loader parity passes.
3. Both Laravel suites run in temporary fixtures without the developer's `.env`.
4. Generated Laravel application dotenv encoding remains unchanged.

## Verification

- Red: add a Laravel fixture expecting the shared-loader path and explicit env-file selection before wiring it; do not claim existing green tests as Red evidence.
- `bash scripts/devarch/tests/dotenv_test.sh`
- `bash scripts/laravel/tests/bootstrap_test.sh`
- `bash scripts/laravel/bootstrap.test.sh`
- `bash -n scripts/laravel/bootstrap.sh scripts/laravel/tests/bootstrap_test.sh scripts/laravel/bootstrap.test.sh`.

## Rollback/migration constraints

Restore the private parser and fixture layout. No persistent migration and no local `.env` mutation.
