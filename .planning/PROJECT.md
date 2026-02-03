# DevArch — Stacks & Instances Milestone

## What This Is

DevArch is a local microservices development environment where the only prerequisite is Podman. It manages containerized services via a Go API (DB as source of truth, compose YAML generated on-the-fly) with a React dashboard and CLI. This milestone adds **stacks** — composable groups of service instances with automatic wiring, isolation, and declarative export/import — turning DevArch from "manage individual services" into "compose full development environments."

## Core Value

Two stacks using the same service template must never collide. Isolation (naming, networking, ports) is the primitive everything else depends on.

## Requirements

### Validated

- ✓ Service catalog with 166+ compose definitions across 24 categories — existing
- ✓ Go API with service CRUD, container ops, compose generation from DB — existing
- ✓ React dashboard with service management, container views, project management — existing
- ✓ CLI wrapper (service-manager.sh) as thin API client — existing
- ✓ Runtime detection and switching (Docker/Podman) — existing
- ✓ Nginx reverse proxy generation per service — existing
- ✓ DB migrations (up to 012) with service definitions, ports, volumes, env vars, labels, domains, healthchecks, config files — existing
- ✓ WebSocket-based real-time container status — existing

### Active

- [ ] Stack CRUD (create, list, get, update, delete stacks)
- [ ] Service instance creation from template services with full copy-on-write overrides
- [ ] Instance override tables: ports, volumes, env vars, dependencies, labels, domains, healthchecks, config files
- [ ] Effective config resolver (template + instance overrides merged)
- [ ] Deterministic container naming: devarch-{stack}-{instance}
- [ ] Per-stack network isolation (dedicated bridge network per stack)
- [ ] Stack compose generator (one compose YAML per stack with N services)
- [ ] Identity labels on all stack containers (devarch.stack_id, devarch.instance_id, devarch.template_service_id)
- [ ] Plan/Apply workflow: preview changes before applying, advisory locking per stack
- [ ] Service export/import contracts on template catalog (exports + import contracts)
- [ ] Auto-wiring for simple cases (one provider matches one consumer)
- [ ] Explicit contract-based wiring for ambiguous cases
- [ ] Wiring diagnostics (missing/ambiguous required contracts surfaced in plan)
- [ ] Env var injection from wires (DB_HOST, DB_PORT, etc. using internal DNS)
- [ ] devarch.yml export (serialize stack + instances + overrides, secrets redacted)
- [ ] devarch.yml import with reconciliation (create-update mode)
- [ ] Secrets encrypted at rest (AES-256-GCM, key in ~/.devarch/secret.key)
- [ ] Secret redaction in all API responses, plan output, compose previews, exports
- [ ] Resource limits per instance (CPU, memory mapped to compose fields)
- [ ] Runtime abstraction fix (remove hardcoded podman exec.Command, route through container client)
- [ ] Dashboard: full stack management UI (create/manage stacks, add instances, view wiring, plan/apply)
- [ ] Dashboard UI ships per-phase alongside corresponding API work

### Out of Scope

- Multi-machine / remote stack deployment — local dev tool only
- Stack-level CI/CD pipelines — beyond dev environment scope
- Full vault/HSM secret management — simple key file sufficient for local dev
- Stack versioning / rollback — export/import covers this use case simply
- devarch.yml full-reconcile mode (delete removed items) — create-update only for v1
- OAuth/SSO for DevArch itself — single-user/local tool

## Context

- Go API uses chi router, database/sql with PostgreSQL, go-migrate for migrations
- Dashboard: React 18 + Vite + Tailwind + HeadlessUI + Heroicons + TanStack Router
- Current DB schema has services, service_ports, service_volumes, service_env_vars, service_dependencies, service_labels, service_domains, service_healthchecks, service_config_files (migrations 001-012)
- All services currently share one bridge network (microservices-net)
- Container client already abstracts Docker/Podman — but project controller has hardcoded podman exec.Command calls that need fixing
- Compose generator (api/internal/compose/generator.go) generates per-service YAML from DB; stack generator will produce per-stack YAML
- Existing dashboard features: service list/detail, container management, project CRUD, category views, runtime config

## Constraints

- **Runtime**: Must work with both Podman and Docker — container client abstracts this
- **Tech stack**: Go API, React dashboard — no new languages/frameworks
- **DB**: PostgreSQL — all state in DB, compose YAML is derived
- **Backwards compat**: Existing single-service operations must continue working alongside new stack operations
- **Encryption key**: AES-256-GCM with key stored in ~/.devarch/secret.key — auto-generated on first use
- **Naming**: Container names must be deterministic and collision-free across stacks

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Full copy-on-write overrides (including healthchecks, config files) | Tool is for broader adoption — users expect full control | — Pending |
| Isolation as core value over auto-wiring or declarative config | Wiring and config are useless if stacks collide | — Pending |
| Auto-wire + explicit contracts (both layers) | Simple stacks stay simple, complex stacks are possible | — Pending |
| Encryption at rest from v1 | Avoids painful retrofit; builds trust for adoption | — Pending |
| Per-phase dashboard UI (not a dedicated UI phase) | Early feedback loop, testable increments, avoids big-bang UI phase | — Pending |
| devarch.yml for sharing + backup (both equally) | Portable definitions that also serve as state backup | — Pending |
| Redaction only initially was rejected — full encryption chosen | Even for local dev, "secrets are encrypted" is a trust signal for adoption | — Pending |

---
*Last updated: 2026-02-03 after initialization*
