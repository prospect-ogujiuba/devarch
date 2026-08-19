# Phase 05: Declarative stack recipes

Created: 2026-08-19
Purpose: Start and stop repeatable groups such as WordPress infrastructure, Laravel-loaded, observability, and local AI without copying command sequences.

## Goal

Add `scripts/devarch/stack.sh` backed by auditable declarative recipe files.

## Scope

- Commands: `list`, `show`, `up`, `down`, `restart`, `status`, and `logs`.
- Add recipe files for `wordpress`, `laravel-bare`, `laravel-standard`, `laravel-loaded`, `observability`, and `local-ai` after validating their exact members.
- Define ordered members and optional members using a non-executable line format.
- Start in declared order, stop in reverse order, and report partial failures clearly.

## Outputs

- Stack script, recipe directory/schema, canonical recipes, tests, and README examples.

## Acceptance criteria

- Recipe parsing rejects unknown directives, duplicate IDs, cycles/includes, and unsafe text before mutation.
- `show` and `--dry-run` reveal exact ordered operations.
- A failed member stops subsequent startup and identifies already-started members; no implicit destructive rollback.
- Stack status/logs reuse shared commands rather than reimplementing runtime parsing.
- Users can add a recipe without editing Bash.

## Verification

- Fixture tests cover valid recipes, missing services, duplicates, partial startup, reverse shutdown, and dry-run.
- Validate canonical recipes against the real catalog.
- Syntax, ShellCheck, and fake-runtime regression tests pass.

## Open questions

Decide during implementation whether recipe inclusion is needed in v1; default recommendation is no inclusion until a real duplication problem appears.

## Non-goals

No general dependency solver, parallel scheduler, or Kubernetes-style reconciliation.
