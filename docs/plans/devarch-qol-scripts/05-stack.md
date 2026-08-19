# Phase 05: Declarative stack recipes

Created: 2026-08-19
Purpose: Start and stop repeatable groups such as WordPress infrastructure, Laravel-loaded, observability, and local AI without copying command sequences.

## Goal

Add a Compose-file-set selector that resolves named DevArch recipes and performs one native `podman compose` invocation.

## Scope

- DevArch-owned modes are `list`, `show`, and `run STACK -- COMPOSE_ARGS...`; lifecycle words remain native Compose arguments.
- Prefer canonical Compose override files/profiles as recipes. If a text recipe is unavoidable, it contains only ordered canonical Compose file paths.
- Resolve a stack to repeated `-f FILE` arguments and call `podman compose ...` exactly once.
- Add `wordpress`, `laravel-bare`, `laravel-standard`, `laravel-loaded`, `observability`, and `local-ai` only after `podman compose ... config` proves each merged project valid.

## Native delegation

Compose owns dependency order, parallelism, health/dependency conditions, lifecycle, partial-failure reporting, logs, and shutdown. The wrapper never loops through members. Example: `stack.sh observability -- up -d` becomes one `podman compose -f ... -f ... up -d` invocation.

## Outputs

- Stack script, recipe directory/schema, canonical recipes, tests, and README examples.

## Acceptance criteria

- Recipe parsing rejects unknown directives, duplicate IDs, cycles/includes, and unsafe text before mutation.
- `show` and `--dry-run` reveal exact ordered operations.
- Provider output and exit status report startup failures; DevArch adds no rollback or partial-start accounting.
- Status/logs are native arguments passed to the same Compose project.
- Users can add a recipe without editing Bash.

## Verification

- Fixture tests cover valid recipes, missing services, duplicates, exact repeated `-f` argv, passthrough, and provider dry-run.
- Validate canonical recipes against the real catalog.
- Syntax, ShellCheck, and fake-runtime regression tests pass.

## Open questions

Confirm that merging today's one-service Compose files preserves volumes, networks, container names, and project behavior. If not, use a canonical generated/handwritten stack Compose file rather than scripting sequential lifecycle calls.

## Non-goals

No per-service lifecycle loop, reverse-shutdown loop, dependency solver, parallel scheduler, retry/rollback engine, log aggregator, or Kubernetes-style reconciliation.
