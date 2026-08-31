# Phase 1: Registry and resolver

Created: 2026-08-30
Purpose: Introduce orthogonal, validated data-service selection without changing generated application behavior yet.

## Goal

Represent database and cache options as a small registry and resolve a valid product with a framework profile before mutation.

## Scope

- Add `--database NAME`, `--cache NAME`, and `--list-data-services`.
- Default both axes to `none`; permit at most one selection per axis initially.
- Add profile metadata such as `RUNTIME_CAPABILITIES=(browser server)`; direct data access requires `server`.
- Add allowlisted manifests under `scripts/javascript/data-services/<name>/service.conf`.
- Fields cover kind, description, Compose file/service, runtime endpoint, environment key/URL, dependencies, configurator, requirements, and conflicts.
- Resolve into ordered arrays and associative maps, deduplicate npm packages, validate conflicts, and print the plan.
- Do not start Compose or generate connection code yet.

## Representation

Use one manifest per service, never one per combination. Store selections in ordered Bash arrays plus associative maps keyed by service id. Lookup is O(1); iteration is deterministic. Slug-validate and allowlist every user-selected id.

```bash
SERVICE_KIND=database
SERVICE_DESCRIPTION='PostgreSQL relational database'
COMPOSE_FILE='services-library/database/postgres/compose.yml'
COMPOSE_SERVICE=postgres
RUNTIME_HOST=postgres
RUNTIME_PORT=5432
ENV_KEY=DATABASE_URL
DEPENDENCIES=(drizzle-orm postgres)
DEV_DEPENDENCIES=(drizzle-kit dotenv)
CONFIGURATOR='configure.mjs'
REQUIRES_CAPABILITIES=(server)
```

## Outputs

- Registry schema validation.
- CLI/listing/help/dry-run support.
- Framework/profile capability metadata.
- Resolver tests and extension documentation.

## Acceptance criteria

- Existing invocations remain compatible.
- `next/fullstack + postgres + redis` resolves without creating files.
- `vite-react/typescript + postgres` fails before staging.
- Manifest paths cannot escape approved project directories.
- Package lists are deterministic and deduplicated.

## Verification

- `bash scripts/javascript/bootstrap.test.sh`
- Resolver fixture tests for valid products, `none`, invalid ids, duplicate roles, conflicts, and browser-only rejection.

## Open questions

- `RUNTIME_CAPABILITIES` versus `PROFILE_CAPABILITIES` naming.
- Whether specialist stores later become axes such as analytics/graph.

## Non-goals

Credentials, startup, migrations, generated modules, multiple primary databases, or blanket support for every catalog service.
