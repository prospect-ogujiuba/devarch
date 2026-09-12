# LAR-003: reviewed Laravel capability handlers

- Status: planned
- Plan revision: 3
- Spec: r2
- Dependencies: `LAR-001`

## Goal and observable behavior

Implement code-owned plan/install/configure/verify handlers for `sanctum`, `horizon`, `reverb`, `telescope`, `pulse`, `pennant`, `scout-meilisearch`, `filament`, and `pest`, including exact conflicts, required catalog services/health probes, compatibility checks, deterministic ordering, and recovery integration.

## Scope

One registry maps fixed IDs to shell functions and metadata; profile text never supplies commands. Resolve all requested handlers/conflicts/services before mutation. Handler execution uses argv arrays, validated package constraints, stdin/files for secrets, bounded timeouts, local resolved-version evidence, and explicit post-install verification.

## Non-goals

No starter engine changes, profile catalog, external SaaS, billing, social auth, tenancy, Octane, or arbitrary third-party capability extension.

## Expected files

`scripts/laravel/lib/capabilities.sh` or equivalent; Laravel bootstrap integration; handler fixture tests/docs.

## Specialist decisions incorporated

Stable ordered de-duplication; fixed trust roots/handlers; service preflight and health; offline default tests; explicit upstream package trust.

## Acceptance criteria

- AC-08/09 pass for all handlers.
- Exact duplicates coalesce; conflicts/missing services/failed probes/incompatible packages fail before app capability mutation.
- Only selected services start; every handler verifies its installed/configured state.
- Synthetic secrets never enter argv/output/log/recovery/definitions.

## Verification

Red table-driven handler plan/order/conflict/service/health/compatibility tests; fake Composer/Artisan execution and postcondition failures; canary scan; full Laravel suite. No real multi-capability profile smoke until `LAR-004`.

## Rollback

Remove handlers/integration; parser and starter engine remain, legacy profiles remain usable.
