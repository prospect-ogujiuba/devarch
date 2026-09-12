# LAR-005: Laravel scaffold matrix

- Status: planned
- Plan revision: 3
- Spec: r2
- Dependencies: `LAR-004`

## Goal and observable behavior

Add native Laravel matrix commands for lexical list, create, create-all, start, stop, status, and verify over deterministic `apps/showcase-laravel-<profile>` targets.

## Scope

DNS-safe prefix override; `artisan`+`composer.json` completion markers; sequential resumable bulk creation; bounded retry/timeouts only for fresh scaffold acquisition; unique run IDs; disk/network/service prerequisites; preserved per-profile failure logs; independent continuation; concise nonzero failure summary. Never adopt/rename/overwrite/delete unmanaged `apps/showcase-laravel`.

## Non-goals

No top-level adapter, concurrency, default start-all, pruning, recipes, or profile internals.

## Expected files

`scripts/laravel/scaffold-matrix.sh`, matrix tests/docs.

## Acceptance criteria

AC-07 passes through native matrix commands; direct Laravel bootstrap semantics remain unchanged.

## Verification

Red offline fixtures for list/name/state, complete skip, incomplete/unmanaged target, prerequisite, timeout, bounded retry, continuation, quarantine/recovery, unique logs, and final summary; focused/full Laravel suites; opt-in one-profile create/status/verify/stop smoke and cleanup.

## Rollback

Remove matrix; bootstrap and apps remain.
