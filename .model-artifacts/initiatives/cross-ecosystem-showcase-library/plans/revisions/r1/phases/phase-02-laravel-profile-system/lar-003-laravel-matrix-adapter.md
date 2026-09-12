# LAR-003: Laravel matrix and adapter

- Status: planned
- Plan revision: 1
- Spec: r1
- Dependencies: `LAR-002`

## Goal and observable behavior

Add a deterministic Laravel matrix and protocol-v1 adapter supporting list/create/create-all/start/stop/status/verify for `apps/showcase-laravel-<profile>`.

## Scope

Implement lexical profile discovery, DNS-safe naming/prefix override, completion-marker checks (`artisan` and `composer.json`), sequential resumable bulk creation, bounded retries only around fresh scaffold acquisition, per-profile logs, independent continuation, concise summary, and adapter mapping.

## Non-goals

No concurrency, automatic start-all, profile internals, recipes, or deletion/pruning command.

## Expected files

`scripts/laravel/scaffold-matrix.sh`, matrix tests, `scripts/showcase/adapters/laravel.sh`, docs.

## Acceptance criteria

- AC-01/02/07 pass.
- Existing complete targets skip; incomplete targets fail without mutation; existing unrelated apps are untouched.
- Failure quarantines through bootstrap recovery, later profiles continue, final exit is nonzero with exact failures/log path.
- Start/stop/verify act only on one explicit profile; bulk creation does not run all apps.

## Verification

Red matrix fixtures mirror JavaScript lifecycle cases plus recovery-marker/quarantine assertions; adapter parity fixtures; representative create/verify and safe cleanup smoke.

## Rollback

Remove matrix/adapter; native Laravel bootstrap and generated apps remain.
