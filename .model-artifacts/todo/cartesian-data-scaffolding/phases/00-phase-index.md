# Cartesian data-service scaffolding phase index

Created: 2026-08-30
Purpose: Define an editable implementation sequence for composing JavaScript framework profiles with database and cache capabilities.

## Problem summary

The current scaffold selects exactly one framework and one project profile. Putting database support inside profiles would create framework/profile × database × cache duplication. DevArch already has fourteen database/cache Compose services on `microservices-net`, so data services should be orthogonal capability packs resolved at scaffold time and started at runtime.

## Target model

A generated application is the constrained product:

`framework profile × primary database (including none) × cache (including none)`

Do not materialize that product as profile files. Resolve declarative capability manifests dynamically. Initially allow at most one primary database and one cache; arbitrary repeated stores are a later composition problem.

Server capability is a compatibility constraint: browser-only profiles cannot select a direct database/cache connection. They must use a server-capable profile or separate API.

## Phases

1. [Registry and resolver](01-registry-and-resolver.md)
2. [PostgreSQL and Redis vertical slice](02-postgres-redis-vertical-slice.md)
3. [Runtime orchestration](03-runtime-orchestration.md)
4. [Catalog expansion and hardening](04-catalog-expansion-and-hardening.md)

## Architectural decisions

- Keep project profiles product-oriented (`fullstack`, `api`, `tested`); do not add `postgres-*` copies.
- Add orthogonal `--database` and `--cache` dimensions, each defaulting to `none`.
- Use an allowlisted registry under `scripts/javascript/data-services/`; never derive Compose paths directly from user input.
- Record selections in `.devarch/data-services.json` inside the generated app.
- Generate server-only modules using plain Node environment access; framework-specific route examples are optional hooks.
- Batch dependency installation once after resolving all selected packs.
- Keep generated development environment files ignored; commit only `.env.example`.

## Global definition of done

- Existing commands remain compatible without data options.
- `next/fullstack + postgres + redis` scaffolds and connects over `microservices-net`.
- Invalid roles and browser-only profiles fail before mutation.
- Dry runs show framework, profile, data services, packages, files, and runtime services.
- Maintained manifests grow additively rather than as framework/profile × service files.

## Review order

Review this target model first, then phases in numeric order. Resolve Phase 1 open questions before implementation.
