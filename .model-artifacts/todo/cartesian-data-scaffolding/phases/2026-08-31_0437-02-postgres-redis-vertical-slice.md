# Phase 2: PostgreSQL and Redis vertical slice

Created: 2026-08-30
Purpose: Prove that two independent capability packs can generate one coherent server data layer.

## Goal

Generate a Next.js server application with PostgreSQL via Drizzle and Redis without creating any combined framework/database/cache profile.

## Scope

- Implement `postgres` database and `redis` cache manifests/configurators.
- Install all resolved runtime/dev packages in one npm operation.
- Generate `.devarch/data-services.json`, `.env.example`, ignored `.env`, and server-only TypeScript modules.
- PostgreSQL packages: `drizzle-orm`, `postgres`; dev packages: `drizzle-kit`, `dotenv`.
- Redis package: choose and standardize `redis` unless tests expose a concrete need for `ioredis`.
- Add Drizzle configuration, schema directory, migration scripts, and singleton connection lifecycle suitable for development hot reload.
- Add a framework-specific health/example route only for the validated Next.js integration; adapter modules remain framework-neutral.

## Generated contract

```text
.devarch/data-services.json
.env.example
.env
src/lib/server/data/postgres.ts
src/lib/server/data/redis.ts
src/lib/server/data/schema.ts
drizzle.config.ts
```

Inside `microservices-net`, URLs use `postgres:5432` and `redis:6379`. Document that host commands use `127.0.0.1:8502` and `127.0.0.1:8504`; generated database scripts should run in the app container or accept explicit host URL overrides.

## Outputs

- Two capability packs and configurators.
- Machine-readable generated manifest version 1.
- Environment validation with actionable missing-variable errors.
- Database migration and connection smoke test.
- Redis set/get smoke test.

## Acceptance criteria

- `--database postgres`, `--cache redis`, either alone, and neither all scaffold successfully on a server profile.
- Selecting both installs each dependency once and produces no duplicated environment entries.
- Connection modules cannot be imported into a browser bundle in the supported Next.js profile.
- Secrets are absent from committed `.env.example`; local development values live only in ignored files.
- Re-running generation in staging is deterministic.

## Verification

- Bootstrap fixture tests inspect package.json, manifest, env example, and generated modules.
- Typecheck generated Next.js application.
- Real-service smoke test: migrate/query PostgreSQL and set/get/delete Redis.

## Open questions

- Whether to generate a per-application database/user now or initially use the shared `devarch` database.
- Whether Redis should expose one URL or separate host/port/password variables.

## Non-goals

Other frameworks, production secrets, deployment configuration, or cross-service transactions.
