# ENV-003 — Clean the dotenv template and configuration guidance

- Contract ID: `ENV-003`
- Parent: `PHASE-02`
- Kind: implementation
- Status: planned
- Plan revision: 1
- Spec: `.model-artifacts/initiatives/env-configuration-cleanup/specs/2026-09-11_1614-initiative-spec-r1.md`
- Plan: `.model-artifacts/initiatives/env-configuration-cleanup/plans/2026-09-11_1614-plan-index-r1.md`

## Goal and observable behavior

Operators can copy a minimal template, understand which values each host script consumes, avoid committing usable secrets, and migrate an existing `.env` without assuming it configures every Compose service.

## Scope

- Reduce `.env.example` to supported WordPress/Laravel/shared inputs.
- Use blank or unmistakably non-usable secret placeholders and comments describing required values.
- Keep `ADMIN_*` compatibility documented while preferring `WP_ADMIN_*` in the template.
- Update root, WordPress, and Laravel README configuration guidance.
- Explain Compose substitution versus `environment:` and explicit MariaDB credential forwarding.
- Document that obsolete entries are harmless/ignored, local `.env` is never automatically rewritten, and existing MariaDB volumes require their initialization password.
- Add a deterministic assertion that template keys are a subset of declared bootstrap allowlists.

## Non-goals

No `.env` rename, automatic migration, secret manager, Compose redesign, or cleanup of unrelated service-catalog defaults.

## Dependencies and entry inputs

- Depends on: `ENV-002`.
- Entry inputs: final allowlists exposed by the green shared loader consumers.
- Required capabilities: documentation and configuration-contract testing.

## Expected files/outputs

- `.env.example`
- `README.md`
- `scripts/wordpress/README.md`
- `scripts/laravel/README.md`
- Appropriate host-only contract test under `scripts/devarch/tests/`

## Specialist decisions incorporated

- Security: example secrets are not usable defaults; no real `.env` content enters tests or documentation.
- Migration/compatibility: preserve old aliases and ignored old entries; give manual, reversible migration steps.
- Operations: distinguish service-directory Compose invocation from root bootstrap configuration and warn about initialized MariaDB volumes.

## Acceptance criteria

1. AC5 and AC6 pass through template-contract assertions and documentation review.
2. Every active template key is accepted by at least one bootstrap; every omitted legacy key is identified in migration guidance by category.
3. No tracked Compose file receives a secret or host-only bootstrap input.
4. All focused and existing regressions satisfy AC7.

## Verification

- Focused template/allowlist contract test.
- `bash scripts/devarch/tests/run-tests.sh`
- `bash scripts/wordpress/bootstrap.test.sh`
- `bash scripts/laravel/tests/bootstrap_test.sh`
- `bash scripts/laravel/bootstrap.test.sh`
- `git diff --check` and targeted search confirming removed legacy keys occur only in migration documentation/tests where intended.

## Rollback/migration constraints

Revert template/docs/tests. Do not alter `.env`, volumes, databases, or generated application configuration.
