# ENV-001 — Secure WordPress dotenv loading

- Contract ID: `ENV-001`
- Parent: `PHASE-01`
- Kind: implementation
- Status: planned
- Plan revision: 2
- Spec: `.model-artifacts/initiatives/env-configuration-cleanup/specs/2026-09-11_1625-initiative-spec-r2.md`
- Plan: `.model-artifacts/initiatives/env-configuration-cleanup/plans/2026-09-11_1625-plan-index-r2.md`

## Goal and observable behavior

Introduce a reusable, non-evaluating dotenv loader and use it in WordPress so only declared keys affect provisioning. Shell payloads and unrelated variables are inert, supported values retain precedence, and tests never read the developer's real `.env`.

## Scope

- Add a focused shared Bash dotenv loader under `scripts/devarch/lib/`.
- Define an API accepting an explicit file and allowlist; validate NUL/control bytes for accepted assignments; preserve Laravel-compatible unquoted, single-quoted, double-quoted, escaped, empty, and trailing-comment semantics.
- Replace WordPress `source`/`set -a` loading with the shared loader.
- Add `DEVARCH_ENV_FILE` selection: unset uses optional `<repo>/.env`; a nonempty explicit path must be a regular file or fail before mutation. The control is not loadable from dotenv.
- Allowlist documented WordPress inputs: `WP_ADMIN_USER`, `ADMIN_USER`, `WP_ADMIN_PASSWORD`, `ADMIN_PASSWORD`, `WP_ADMIN_EMAIL`, `ADMIN_EMAIL`, `MARIADB_ROOT_PASSWORD`, `GITHUB_USER`, `AIOWM_GIT_URL`, `CONTAINER_RUNTIME`, and `WORDPRESS_CONTAINER_USER`.
- Add unit/security characterization and WordPress regression coverage.

## Non-goals

Do not change WordPress provisioning, profiles, CLI options, Compose definitions, secret values, or local `.env`.

## Dependencies and entry inputs

- Depends on: none.
- Entry inputs: clean worktree awareness; active approved spec/plan; current WordPress regression tests; current Laravel parser semantics as characterization reference.
- Required capabilities: Bash editing/testing and security-aware input parsing.

## Expected files/outputs

- `scripts/devarch/lib/dotenv.sh` (new)
- `scripts/devarch/tests/dotenv_test.sh` (new)
- `scripts/devarch/tests/run-tests.sh`
- `scripts/wordpress/bootstrap.sh`
- `scripts/wordpress/bootstrap.test.sh`

## Specialist decisions incorporated

- Security: never evaluate dotenv text; allowlist before assignment; do not auto-export unrelated values; suppress secret material in failures.
- TDD: first exercise current WordPress loading with an isolated sentinel command-substitution/unrelated-export fixture; then add table-driven parsing cases including duplicate recognized keys (last wins).
- Compatibility: retain default file path, supported aliases, and file-over-inherited behavior.
- Operations: tests select a fixture/disabled env file explicitly.

## Acceptance criteria

1. AC1 and AC2 pass for WordPress with an isolated sentinel command-substitution and unrelated-child-environment fixture.
2. Recognized quoted/commented/empty/duplicate assignments match specified semantics; accepted controls fail without reflecting values, while malformed unrecognized lines remain inert.
3. Existing documented WordPress precedence and dry-run secret redaction remain intact.
4. WordPress tests are isolated from the repository's ignored `.env`.

## Verification

- Red during implementation: add focused tests showing current shell evaluation/export behavior violates AC1/AC2; record actual failing output without secrets.
- Green/refactor: `bash scripts/devarch/tests/dotenv_test.sh`.
- Regression: `bash scripts/wordpress/bootstrap.test.sh`.
- Syntax: `bash -n scripts/devarch/lib/dotenv.sh scripts/wordpress/bootstrap.sh scripts/wordpress/bootstrap.test.sh`.

## Rollback/migration constraints

Rollback removes the shared loader and restores prior WordPress loading. No persistent data migration. Never modify the operator's `.env`.
