# Phase 15: Environment initializer

Created: 2026-08-19
Purpose: Replace insecure example placeholders with a safe, repeatable first-run configuration process.

## Goal

Add `scripts/devarch/env-init.sh` to create or repair supported `.env` values without sourcing or overwriting unrelated configuration.

## Scope

- Initialize from `.env.example`, generate strong local secrets, preserve comments/order, and update only an allowlisted key set.
- Commands/modes: initial creation, `--check`, `--rotate KEY`, `--print-template`, `--dry-run`, and explicit backup location.
- Use `openssl` with `/dev/urandom` fallback; set restrictive file permissions.
- Validate email/domain/user inputs and reject known placeholders.
- Coordinate credential rotation warnings for already-initialized stateful services.

## Outputs

- Environment initializer, non-evaluating assignment parser/writer, tests, and first-run documentation.

## Acceptance criteria

- Existing `.env` is never overwritten wholesale; every mutation creates a mode-0600 backup and uses atomic replacement.
- Generated secrets meet documented entropy/character constraints and never print unless the user explicitly requests a single generated value through a secure mode.
- Shell substitutions, command substitutions, and unrelated keys are preserved as text and never executed.
- `--check` reports missing/placeholders/permissions without mutations or secret disclosure.
- Rotation requires confirmation and explains that persistent services may retain old credentials.

## Verification

- Fixtures cover absent/existing files, quotes/comments, duplicates, controls, malicious shell text, generator fallback, permissions, interrupted writes, and rotation.
- Canary command-substitution fixtures prove no execution occurs.
- Run doctor against a generated temporary environment.

## Non-goals

No production secret manager, automatic live database credential rotation, GitHub token creation, or `.env` synchronization across machines.
