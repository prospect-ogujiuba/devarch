# LAR-003: Laravel matrix and adapter

- Status: planned
- Plan revision: 2
- Spec: r2
- Dependencies: `LAR-002`

## Goal and observable behavior

Add a deterministic Laravel matrix and protocol-v1 adapter supporting list/create/create-all/start/stop/status/verify for `apps/showcase-laravel-<profile>`.

## Scope

Implement lexical profile discovery, DNS-safe naming/prefix override, completion-marker checks (`artisan` and `composer.json`), sequential resumable bulk creation, bounded retries/timeouts only around fresh scaffold acquisition, unique run IDs, prerequisite checks, per-profile preserved failure logs, independent continuation, concise summary, and adapter mapping.

## Non-goals

No concurrency, automatic start-all, profile internals, recipes, or deletion/pruning command.

## Expected files

`scripts/laravel/scaffold-matrix.sh`, matrix tests, `scripts/showcase/adapters/laravel.sh`, docs.

## Acceptance criteria

- AC-01/02/07 pass.
- Existing complete matrix targets skip; incomplete targets fail without mutation; unrelated/unmanaged apps including `apps/showcase-laravel` are never adopted, renamed, overwritten, or deleted.
- Failure quarantines through bootstrap recovery, later profiles continue, final exit is nonzero with exact failures/log path.
- Start/stop/verify act only on one explicit profile; bulk creation does not run all apps.

## Verification

Red offline matrix fixtures mirror JavaScript lifecycle cases plus timeout, prerequisite, recovery-marker/quarantine, unique-run-log, and unmanaged-target assertions; adapter parity fixtures; opt-in representative create/verify and explicit safe-cleanup smoke.

## Rollback

Remove matrix/adapter; native Laravel bootstrap and generated apps remain.
