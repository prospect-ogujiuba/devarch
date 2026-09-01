# Phase 1: Verify Galaxy and runtime

Created: 2026-08-23
Purpose: Establish a working scaffold/runtime baseline before generating application code.

## Goal

Start or repair the Playground dependencies, inventory all three Galaxy surfaces, and prove the app generator can be invoked safely.

## Scope

- Determine the documented Playground start command and database dependency.
- Verify `galaxy`, `galaxy_makermaker`, and `galaxy_playground_app` command inventories.
- Capture exact `make:maker-resource` and `migrate` syntax.
- Confirm `playground-app` is the application-owned generation target and is clean enough to edit.

## Outputs

- Verified command matrix.
- Runtime/database blocker resolution or explicit exception handoff.
- Dry-run or non-mutating generator validation where supported.

## Acceptance criteria

- All available Galaxy surfaces either list commands successfully or have a documented blocker.
- Exact scaffold and migration commands are known before generation.
- No resource files are generated in TypeRocket or MakerMaker core.

## Verification

- Galaxy `list --raw` and relevant `--help` outputs.
- WordPress/database connectivity check.
- Git/worktree scope inspection.

## Non-goals

- No domain resource generation or schema changes.
