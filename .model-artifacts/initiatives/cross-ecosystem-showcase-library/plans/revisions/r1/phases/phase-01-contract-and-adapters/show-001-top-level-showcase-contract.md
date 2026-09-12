# SHOW-001: top-level showcase contract and orchestrator

- Status: planned
- Plan revision: 1
- Spec: r1
- Dependencies: none

## Goal and observable behavior

Create a versioned, ecosystem-neutral CLI at `scripts/showcase/showcase.sh` that deterministically lists tracked showcase definitions and dispatches validated `create`, `start`, `stop`, `verify`, and `status` actions to ecosystem adapters. Listing reports ecosystem, definition, deterministic app names, support tier, and local state. Unsupported actions fail clearly.

## Scope

- Define adapter protocol v1 and status vocabulary.
- Implement data-only adapter registry/discovery, lexical ordering, validation, argument boundary, dispatch, and exit propagation.
- Add fixture tests and concise README.

## Non-goals

No Podman abstraction, generated apps, ecosystem profile changes, recipes, concurrency, or arbitrary adapter paths from user data.

## Entry inputs

Approved plan r1; existing script conventions; existing ignored `apps/` policy.

## Expected files

`scripts/showcase/showcase.sh`, `scripts/showcase/adapters/`, `scripts/showcase/showcase.test.sh`, `scripts/showcase/README.md`.

## Acceptance criteria

- AC-01/02 listing and dispatch behavior pass fixtures.
- Registry/adapters are repository-contained regular files, validated, uniquely named, and protocol-version checked.
- State derives only from adapter-declared completion markers, not from definition authority in generated apps.
- `--` cleanly separates top-level arguments from adapter-owned arguments.
- Invalid IDs/actions and duplicate adapters fail before mutation.

## Verification

Plan-time TDD: first Red tests cover sorted list, unknown adapter/action, duplicate registry ID, traversal/symlink rejection, exact argv/exit propagation, and state reporting. Then `bash -n`, focused fixtures, ShellCheck if installed.

## Rollback

Delete the additive `scripts/showcase/` surface; native ecosystem commands remain untouched.
