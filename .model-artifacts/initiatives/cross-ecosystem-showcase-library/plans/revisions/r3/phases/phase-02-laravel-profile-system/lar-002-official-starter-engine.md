# LAR-002: official Laravel starter engine

- Status: planned
- Plan revision: 3
- Spec: r2
- Dependencies: `LAR-001`

## Goal and observable behavior

Implement non-interactive current-major scaffolding for starter IDs `none`, `livewire`, `react`, `vue`, and `svelte`, including frontend dependency/build steps only when selected, compatibility probes/resolved-version evidence, timeouts, and existing recovery/quarantine behavior.

## Scope

Research and pin/validate current Laravel installer flags during implementation; provide code-owned starter handlers; expose trusted upstream execution in plans; provide diagnostic no-scripts behavior where feasible; record resolved Laravel/starter/frontend versions in local logs/evidence.

## Non-goals

No capability packages, new profile catalog, matrix, recipes, or external auth variants.

## Expected files

Laravel bootstrap/starter handler, PHP image only if installer availability requires it, tests/docs.

## Specialist decisions incorporated

TDD starter Red order; security argv arrays/fixed handlers/upstream disclosure; compatibility probes before later mutation; operations bounded timeout/log/quarantine.

## Acceptance criteria

- Each starter produces an exact redacted dry-run plan and valid app completion markers.
- Unknown starter or incompatible current-major resolution fails before database/later capability mutation where feasible.
- Failure leaves no ambiguous target and produces preserved actionable logs/recovery state.
- Legacy `none` path remains equivalent.

## Verification

Offline fake-installer fixtures for every starter, incompatibility, timeout, scripts mode, frontend build selection, and failure stages; full Laravel suite; one opt-in uniquely named `none`/minimal real smoke with prerequisite and cleanup checks.

## Rollback

Revert starter handlers; parser remains additive and legacy `none` behavior remains usable.
