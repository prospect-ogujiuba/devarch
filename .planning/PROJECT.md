# DevArch

## What This Is

DevArch is a local microservices development environment where the only prerequisite is Podman (or Docker). It manages containerized services via a Go API (DB as source of truth, compose YAML generated on-the-fly) with a React dashboard and CLI. Developers compose full development environments using **stacks** — isolated groups of service instances with automatic wiring, export/import portability, and encrypted secrets.

Portability promise: `devarch.yml` (intent) + `devarch.lock` (resolved ports/digests/template versions) + `devarch init`/`devarch doctor` = same environment on any machine.

## Core Value

Two stacks using the same service template must never collide. Isolation (naming, networking, ports) is the primitive everything else depends on.

## Current Milestone: v1.5 Architecture Hardening

**Goal:** Harden security model, normalize API contracts, decompose monolithic handlers, optimize query paths, extract frontend controllers, and establish test/observability baselines.

**Target features:**
- Security modes (dev-open/dev-keyed/strict), CORS allowlist, WS auth, secret hygiene
- Unified API response envelopes and error normalization
- Deploy orchestration and identity domain services extracted from handlers
- Batch status/metrics retrieval, filtered count fixes, query optimization
- Frontend controller hooks for stacks/services/instances, shared mutation helpers
- API integration tests, frontend controller tests, OpenAPI spec generation
- Structured logging, persisted sync job history

## Current State (v1.1 shipped 2026-02-10)

- 172+ service templates across 24 categories, 1:1 legacy parity verified
- 9 domain-separated migrations (001-009), 39 tables, zero seed data
- Parser preserves env_files, container_name, networks, config mounts with cross-service FK provenance
- Compose generator emits all fields faithfully from DB — verified by verify-parity tool
- Stack CRUD with soft-delete, clone-as-rename
- Service instances with full copy-on-write overrides (ports, volumes, env, labels, deps, domains, healthchecks, config files, env_files, networks, config_mounts)
- Per-stack network isolation, deterministic container naming (devarch-{stack}-{instance})
- Stack compose generation with config materialization
- Terraform-style plan/apply with advisory locking + staleness detection
- Contract-based auto-wiring + explicit wiring for ambiguous cases
- devarch.yml export/import + devarch.lock + devarch init/doctor
- Streaming multipart import (256MB cap), prepared statement batching
- AES-256-GCM secret encryption at rest, redaction in all outputs
- Resource limits (CPU/memory) mapped to compose deploy.resources
- Dashboard: full stack management UI with env_files, networks, config_mounts editing
- Living verification tools: verify-parity (parity proof), verify-boundary (size limit proof)
- Codebase: ~25K LOC Go, ~18K LOC TypeScript

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
- ✓ Fresh baseline migrations (9 domain-separated, 39 tables, zero seeds) — v1.1
- ✓ Parser preserves env_file, container_name, networks, config mounts — v1.1
- ✓ Config mount provenance with cross-service FK linking — v1.1
- ✓ Compose generator 1:1 legacy parity (172/173 pass, 1 whitelisted) — v1.1
- ✓ Streaming multipart import (256MB cap, prepared statements, idempotent upserts) — v1.1
- ✓ Dashboard env_files, networks, config_mounts editing (service + instance) — v1.1
- ✓ Golden service parity verified (7/7 pass with zero exceptions) — v1.1
- ✓ Import boundary tests (200MB accepted, 300MB rejected with 413) — v1.1

### Active

- [ ] Security hardening: CORS allowlist, WS origin checks, WS auth tokens, security modes, secret hygiene
- [ ] API contract normalization: shared responder envelopes, standardized error payloads, action conventions
- [ ] Backend decomposition: deploy orchestration service, identity domain service, naming deduplication
- [ ] Performance optimization: batch status/metrics, filtered count fix, override-count query optimization
- [ ] Frontend extraction: stack/instance/service detail controllers, shared mutation helpers, WS invalidation expansion
- [ ] Test coverage: API integration tests (CRUD, plan/apply), frontend controller tests
- [ ] Observability: structured log fields, persisted sync job history, OpenAPI spec generation

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
- **DB**: PostgreSQL, 9 domain-separated migrations (v1.1 fresh baseline)
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

| Fresh baseline over incremental migration | 23 patch migrations create fragility; clean slate is simpler to maintain | ✓ Good |
| Domain-separated DDL files | Clean boundaries, each table created once in final form | ✓ Good |
| Dedicated service_config_mounts table | Config provenance needs its own model, not overloaded config_files | ✓ Good |
| Streaming multipart for large imports | ParseMultipartForm buffers entire body; streaming avoids OOM | ✓ Good |
| MaxBodySize scope isolation | Import route registered outside /api/v1 group to avoid middleware stacking | ✓ Good |
| Whitelist governance for parity exceptions | Expected differences documented with reason, golden services never whitelisted | ✓ Good |

---
*Last updated: 2026-02-11 after v1.5 milestone start*
