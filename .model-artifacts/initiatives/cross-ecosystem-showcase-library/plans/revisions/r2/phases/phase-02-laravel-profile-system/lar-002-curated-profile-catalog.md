# LAR-002: curated Laravel profile catalog

- Status: planned
- Plan revision: 2
- Spec: r2
- Dependencies: `LAR-001`

## Goal and observable behavior

Provide 13 deterministic, documented, locally useful Laravel profiles: `minimal`, `web`, `api`, `livewire`, `react`, `vue`, `svelte`, `queues`, `realtime`, `admin`, `search`, `observability`, and `testing`.

## Scope

Define each new profile additively and explicitly without inheritance. Retain `bare`, `standard`, and `loaded` as canonical public files/names with characterized equivalent behavior:

- `minimal`: SQLite, no optional service/capability.
- `web`: MariaDB and Mailpit.
- `api`: MariaDB plus Sanctum.
- `livewire`/`react`/`vue`/`svelte`: corresponding official starter, MariaDB, Mailpit.
- `queues`: MariaDB, Redis, Mailpit, Horizon.
- `realtime`: MariaDB, Redis, Reverb.
- `admin`: MariaDB, Mailpit, Filament.
- `search`: MariaDB, Meilisearch, Scout integration.
- `observability`: MariaDB, Redis, Telescope and Pulse.
- `testing`: SQLite, Pest, deterministic test execution.

Each profile declares support tier, compatibility probe, required catalog service IDs and health probes. Research and record current compatible constraints/installer flags during implementation; record resolved versions in local evidence, prefer current-major ranges, and fail closed over silent drift.

## Non-goals

No external-account profiles, multi-tenancy, billing, Socialite providers, Octane, or profile inheritance.

## Expected files

` scripts/laravel/profiles/*.profile` additions/legacy adjustments, catalog tests, Laravel README.

## Acceptance criteria

- AC-06/08/09 pass for all 13 profiles.
- Listing is lexical and descriptions/support tiers/prerequisites/supported actions are explicit; existing unmanaged `apps/showcase-laravel` is never adopted or changed.
- Profile plans identify starter, services, capabilities, packages, frontend build, and verification without secrets.
- Conflicting capabilities/services are absent from shipped profiles.

## Verification

Default table-driven offline fixture test for exact resolved plans of all profiles. Opt-in, uniquely named, bounded smokes for `minimal`, `api`, `react`, `queues`, `observability` perform disk/network/service prerequisites, preserve failure logs, document cleanup, and make HTTPS/Artisan/service assertions appropriate to each.

## Rollback

Remove new profile files; retain legacy profiles and generated apps.
