# Phase 4: Catalog expansion and hardening

Created: 2026-08-30
Purpose: Expand from the proven vertical slice according to protocol families and explicit support quality.

## Goal

Offer a documented, tested subset of the existing service catalog without implying that every service has identical ORM, migration, or framework semantics.

## Scope

Add adapters incrementally by family:

- SQL: MariaDB/MySQL via `mysql2` and Drizzle; CockroachDB through its PostgreSQL protocol after compatibility tests.
- Document/key-value: MongoDB via `mongodb`, CouchDB via `nano`, SurrealDB via its maintained SDK.
- Cache: Memcached via a maintained Node client.
- Specialized: ClickHouse, Neo4j, Cassandra, MSSQL, and EdgeDB/Gel only after client/version and smoke-test decisions.

Each manifest declares support level (`stable`, `experimental`, or `unsupported`), capabilities such as migrations/transactions, and framework/runtime requirements. Listing output exposes these labels.

## Compatibility policy

A catalog Compose service is not automatically a JavaScript scaffold capability. Inclusion requires a maintained Node client, deterministic configuration, container connectivity, a smoke test, and documentation. SQL services need not share one ORM when feature support differs.

## Outputs

- Adapter manifests/configurators by protocol family.
- Generated module naming and environment-variable conventions.
- Pairwise compatibility test strategy instead of exhaustive product scaffolding.
- Documentation table mapping service, role, client, ORM/migrations, support level, and limitations.

## Acceptance criteria

- Adding one adapter requires one manifest/configurator and focused tests, not edits to every framework profile.
- Stable adapters pass real connectivity smoke tests.
- Unsupported combinations fail with a reason rather than silently generating partial code.
- Listing accurately distinguishes database, cache, and specialist capabilities.
- Test runtime scales approximately with profiles + adapters + selected integration pairs, not their full Cartesian product.

## Verification

- Manifest schema test across the whole registry.
- Unit tests for every adapter generator.
- Contract tests for supported server profiles.
- Pairwise integration coverage plus one canonical full product (`next/fullstack + postgres + redis`).

## Open questions

- Promote analytics/graph/search to distinct CLI axes or retain a future repeatable `--service` escape hatch.
- ORM policy for MSSQL and specialist databases.
- Version pinning and deprecation policy when services or Node clients rename, such as EdgeDB/Gel.

## Non-goals

Exhaustively testing every mathematical combination, normalizing all databases behind one lowest-common-denominator repository API, or hiding vendor-specific capabilities.
