# DevArch

## What This Is

DevArch is a local microservices development environment where the only prerequisite is Podman (or Docker). It manages containerized services via a Go API (DB as source of truth, compose YAML generated on-the-fly) with a React dashboard and CLI. Developers compose full development environments using **stacks** — isolated groups of service instances with automatic wiring, export/import portability, and encrypted secrets.

Portability promise: `devarch.yml` (intent) + `devarch.lock` (resolved ports/digests/template versions) + `devarch init`/`devarch doctor` = same environment on any machine.

## Core Value

Two stacks using the same service template must never collide. Isolation (naming, networking, ports) is the primitive everything else depends on.

## Current State (v1.0 shipped 2026-02-09)

- 166+ service templates across 24 categories
- Stack CRUD with soft-delete, clone-as-rename
- Service instances with full copy-on-write overrides (ports, volumes, env, labels, deps, domains, healthchecks, config files)
- Per-stack network isolation, deterministic container naming (devarch-{stack}-{instance})
- Stack compose generation with config materialization
- Terraform-style plan/apply with advisory locking + staleness detection
- Contract-based auto-wiring + explicit wiring for ambiguous cases
- devarch.yml export/import + devarch.lock + devarch init/doctor
- AES-256-GCM secret encryption at rest, redaction in all outputs
- Resource limits (CPU/memory) mapped to compose deploy.resources
- Dashboard: full stack management UI shipped per-phase
- DB migrations: 001-019+

## Requirements

### Validated

- ✓ Service catalog with 166+ compose definitions across 24 categories — pre-existing
- ✓ Go API with service CRUD, container ops, compose generation from DB — pre-existing
- ✓ React dashboard with service management, container views, project management — pre-existing
- ✓ CLI wrapper (service-manager.sh) as thin API client — pre-existing
- ✓ Runtime detection and switching (Docker/Podman) — pre-existing
- ✓ Nginx reverse proxy generation per service — pre-existing
- ✓ WebSocket-based real-time container status — pre-existing
- ✓ Stack CRUD (create, list, get, update, delete, clone, enable/disable) — v1.0
- ✓ Service instances with full copy-on-write overrides — v1.0
- ✓ Effective config resolver (template + overrides merged) — v1.0
- ✓ Per-stack network isolation with deterministic naming — v1.0
- ✓ Stack compose generator (one YAML per stack with N services) — v1.0
- ✓ Plan/Apply workflow with advisory locking — v1.0
- ✓ Contract-based auto-wiring + explicit wiring — v1.0
- ✓ devarch.yml export/import with reconciliation — v1.0
- ✓ devarch.lock generation/validation — v1.0
- ✓ devarch init + devarch doctor — v1.0
- ✓ AES-256-GCM secret encryption at rest — v1.0
- ✓ Secret redaction in all outputs — v1.0
- ✓ Resource limits per instance — v1.0
- ✓ Runtime abstraction fix (no hardcoded podman calls) — v1.0
- ✓ Dashboard: full stack management UI — v1.0

### Active

(Next milestone TBD — run `/gsd:new-milestone`)

### Out of Scope

- Multi-machine / remote stack deployment — local dev tool only
- Stack-level CI/CD pipelines — beyond dev environment scope
- Full vault/HSM secret management — simple key file sufficient for local dev
- Stack versioning / rollback — export/import covers this simply
- devarch.yml full-reconcile mode (delete removed items) — create-update only for now
- Production deployment support — anti-feature, creates false confidence

## Context

- **API**: Go 1.22, chi router, lib/pq, gorilla/websocket, yaml.v3
- **Dashboard**: React 19, Vite, TanStack Router + Query, Tailwind 4, Radix UI, Zod, CodeMirror
- **DB**: PostgreSQL, 19+ migrations
- **Runtime**: Podman or Docker via container.Client abstraction
- **Scripts**: bash CLI wrapper (devarch), service manager, runtime switcher

## Constraints

- **Runtime**: Must work with both Podman and Docker
- **Tech stack**: Go API, React dashboard — no new languages/frameworks
- **DB**: PostgreSQL — all state in DB, compose YAML is derived
- **Backwards compat**: Single-service operations continue working alongside stacks
- **Naming**: Container names deterministic and collision-free across stacks

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Full copy-on-write overrides (including healthchecks, config files) | Users expect full control | ✓ Good |
| Isolation as core value over auto-wiring or declarative config | Wiring is useless if stacks collide | ✓ Good |
| Auto-wire + explicit contracts (both layers) | Simple stacks stay simple, complex stacks possible | ✓ Good |
| Encryption at rest from v1 | Avoids painful retrofit, builds trust | ✓ Good |
| Per-phase dashboard UI (not a dedicated UI phase) | Early feedback loop, testable increments | ✓ Good |
| Stack name is immutable ID; renames via clone | Deterministic naming, stable resources | ✓ Good |
| devarch.yml for sharing + backup (both equally) | Portable definitions that also serve as state backup | ✓ Good |
| Full encryption over redaction-only | "Secrets are encrypted" is a trust signal | ✓ Good |
| Soft-delete with partial unique index | Safe deletes, restore with conflict check | ✓ Good |
| Advisory lock per stack (pg_try_advisory_lock) | Prevents concurrent applies, no external deps | ✓ Good |
| Lockfile as JSON (not YAML) | Consistency with export/import workflow | ✓ Good |
| Three-layer env merge: template → wired → overrides | WIRE-08 compliance, user always wins | ✓ Good |

---
*Last updated: 2026-02-09 after v1.0 milestone*
