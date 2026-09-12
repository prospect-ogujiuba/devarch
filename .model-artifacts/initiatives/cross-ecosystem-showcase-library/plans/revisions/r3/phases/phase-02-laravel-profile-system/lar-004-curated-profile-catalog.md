# LAR-004: curated Laravel profile catalog

- Status: planned
- Plan revision: 3
- Spec: r2
- Dependencies: `LAR-002`, `LAR-003`

## Goal and observable behavior

Add 13 explicit, non-inheriting profiles: `minimal`, `web`, `api`, `livewire`, `react`, `vue`, `svelte`, `queues`, `realtime`, `admin`, `search`, `observability`, and `testing`, while retaining canonical `bare`, `standard`, and `loaded` files/names.

## Profile definitions

- `minimal`: SQLite, starter none.
- `web`: MariaDB, Mailpit, starter none.
- `api`: MariaDB, Sanctum.
- `livewire`/`react`/`vue`/`svelte`: matching official starter, MariaDB, Mailpit.
- `queues`: MariaDB, Redis, Mailpit, Horizon.
- `realtime`: MariaDB, Redis, Reverb.
- `admin`: MariaDB, Mailpit, Filament.
- `search`: MariaDB, Meilisearch, Scout.
- `observability`: MariaDB, Redis, Telescope, Pulse.
- `testing`: SQLite, Pest and deterministic test execution.

Each declares support tier, compatibility probe, required catalog services/health probes, and exact description. Existing unmanaged `apps/showcase-laravel` is untouched.

## Non-goals

No external-account, billing/social, tenancy, Octane, inheritance, matrix, or recipes.

## Expected files

Laravel profile files, table-driven catalog tests, README.

## Acceptance criteria

AC-05/06/08/09 pass; all profiles list lexically and resolve exact secret-safe plans with no shipped conflict.

## Verification

Red offline exact-plan/support/compatibility table for every legacy/new profile. Opt-in uniquely named bounded real smokes for `minimal`, `api`, `react`, `queues`, `observability`, each with prerequisite checks, preserved failure log, HTTPS/Artisan/service postconditions, and cleanup instructions.

## Rollback

Remove new files only; legacy profiles and generated apps remain.
