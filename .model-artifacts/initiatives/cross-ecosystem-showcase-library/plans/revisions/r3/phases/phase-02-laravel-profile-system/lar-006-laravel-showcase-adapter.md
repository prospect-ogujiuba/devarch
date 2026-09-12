# LAR-006: Laravel showcase adapter

- Status: planned
- Plan revision: 3
- Spec: r2
- Dependencies: `LAR-005`

## Goal and observable behavior

Expose the native Laravel matrix through adapter protocol v1 with exact supported-action negotiation, validated dispatch, state, prerequisites, support tier, and exit propagation.

## Scope

Map top-level list/create/start/stop/status/verify to native matrix actions. `resume` is declared unsupported for single-app matrix entries; recipe resume remains recipe-owned. Add direct-versus-adapter parity fixtures.

## Non-goals

No matrix/profile/bootstrap changes, action emulation, or generated apps.

## Expected files

`scripts/showcase/adapters/laravel.sh`, focused adapter tests/docs.

## Acceptance criteria

AC-01/02 pass for Laravel; unsupported actions are reported, not emulated; exact argv/output/exit parity holds.

## Verification

Red protocol/version/action/argv/exit/list-state parity fixtures, then native Laravel matrix/bootstrap and top-level showcase suites.

## Rollback

Remove adapter only.
