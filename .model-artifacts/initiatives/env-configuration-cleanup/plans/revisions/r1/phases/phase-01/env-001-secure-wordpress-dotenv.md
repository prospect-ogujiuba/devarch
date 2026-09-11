# ENV-001 — Secure WordPress dotenv loading

- Contract ID: `ENV-001`
- Parent: `PHASE-01`
- Kind: implementation
- Status: planned
- Plan revision: 1
- Spec: `.model-artifacts/initiatives/env-configuration-cleanup/specs/2026-09-11_1614-initiative-spec-r1.md`
- Plan: `.model-artifacts/initiatives/env-configuration-cleanup/plans/2026-09-11_1614-plan-index-r1.md`

## Goal and observable behavior

Introduce a reusable, non-evaluating dotenv loader and use it in WordPress so only declared keys affect provisioning. Shell payloads and unrelated variables are inert, supported values retain precedence, and tests never read the developer's real `.env`.

## Scope

- Add a focused shared Bash dotenv loader under `scripts/devarch/lib/`.
- Define an API accepting an explicit file and allowlist; validate NUL/control bytes for accepted assignments; preserve Laravel-compatible unquoted, single-quoted, double-quoted, escaped, empty, and trailing-comment semantics.
- Replace WordPress `source`/`set -a` loading with the shared loader.
- Add an explicit test-only/operator env-file override while keeping `<repo>/.env` as default.
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
- TDD: first prove a command-substitution fixture is inert and unrelated variables remain unset, then add supported parsing/precedence cases.
- Compatibility: retain default file path, supported aliases, and file-over-inherited behavior.
- Operations: tests select a fixture/disabled env file explicitly.

## Acceptance criteria

1. AC1 and AC2 pass for WordPress with a sentinel command-substitution fixture.
2. Recognized quoted/commented/empty assignments match specified semantics; accepted control bytes fail safely.
3. Existing documented WordPress precedence and dry-run secret redaction remain intact.
4. WordPress tests are isolated from the repository's ignored `.env`.

## Verification

- Red during implementation: add focused tests showing current shell evaluation/export behavior violates AC1/AC2; record actual failing output without secrets.
- Green/refactor: `bash scripts/devarch/tests/dotenv_test.sh`.
- Regression: `bash scripts/wordpress/bootstrap.test.sh`.
- Syntax: `bash -n scripts/devarch/lib/dotenv.sh scripts/wordpress/bootstrap.sh scripts/wordpress/bootstrap.test.sh`.

## Rollback/migration constraints

Rollback removes the shared loader and restores prior WordPress loading. No persistent data migration. Never modify the operator's `.env`.
