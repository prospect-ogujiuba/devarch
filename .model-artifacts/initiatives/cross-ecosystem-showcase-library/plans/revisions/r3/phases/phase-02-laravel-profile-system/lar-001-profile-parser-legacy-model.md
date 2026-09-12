# LAR-001: Laravel profile parser and legacy model

- Status: planned
- Plan revision: 3
- Spec: r2
- Dependencies: `SHOW-001`

## Goal and observable behavior

Characterize existing `bare`, `standard`, and `loaded` behavior, then extend Laravel's data-only profile parser/model for one allowlisted starter ID, repeatable capability IDs, support tier, compatibility probe ID, and required catalog service/health IDs without installing new starters or capabilities yet.

## Scope

Use indexed arrays for declaration order and associative maps for membership/conflicts; coalesce exact repeatables at first occurrence; reject duplicate singletons, controls, arbitrary commands, traversal, unknown metadata IDs, and declared conflicts before mutation. Preserve all current profile/package/database/migrate/seed/force/no-hosts plans and CLI output.

## Non-goals

No new starter execution, capability handlers, profiles, matrix, services, or generated apps.

## Expected files

`scripts/laravel/bootstrap.sh`, parser/model tests, legacy characterization fixtures, Laravel directive documentation.

## Specialist decisions incorporated

DSA arrays/maps/no inheritance; TDD characterization-first; security data-only parsing; migration preserves canonical profile files/names.

## Acceptance criteria

- AC-04 and parser portion of AC-05/08/09 pass.
- Legacy resolved plans are byte-stable except explicitly additive metadata lines approved by fixtures.
- New metadata resolves into a deterministic secret-safe plan but unknown handler IDs fail preflight.

## Verification

Red legacy characterization first; then valid model, controls/command text, duplicate singleton/repeatable, conflict, missing service/health/compatibility ID, stable order, and synthetic secret-canary fixtures. Run focused and full Laravel bootstrap tests, `bash -n`, ShellCheck if installed.

## Rollback

Revert parser/model while retaining untouched legacy profile files and generated apps.
