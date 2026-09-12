# SHOW-003: WordPress existing-profile adapter

- Status: planned
- Plan revision: 1
- Spec: r1
- Dependencies: `SHOW-001`

## Goal and observable behavior

List and individually provision WordPress's existing `bare`, `clean`, `custom`, and `loaded` profiles through adapter protocol v1 without changing profile contents or native bootstrap behavior.

## Scope

Add WordPress adapter lifecycle mapping, deterministic showcase names, state/verify behavior, and fixture/dry-run tests.

## Non-goals

No new WordPress profiles, bulk provision-all default, profile parser changes, or restore automation through the showcase adapter.

## Expected files

`scripts/showcase/adapters/wordpress.sh`, showcase fixtures/tests, documentation.

## Acceptance criteria

- AC-11 passes.
- `list` is mutation-free; `create` delegates one selected profile.
- Restore and credential-bearing options are rejected at the showcase boundary and remain available only through native bootstrap.
- Existing WordPress tests remain green.

## Verification

Red fixtures for list/name/unsupported restore and exact dry-run dispatch; WordPress bootstrap suite and one adapted profile dry-run.

## Rollback

Remove adapter only.
