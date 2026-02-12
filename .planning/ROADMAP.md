# Roadmap: DevArch

## Milestones

- ✅ **v1.0 Stacks & Instances** — Phases 1-9 (shipped 2026-02-09)
- ✅ **v1.1 Schema Reconciliation** — Phases 10-15 (shipped 2026-02-10)
- 🚧 **v1.1.1 Architecture Hardening** — Phases 16-28 (in progress)

## Phases

<details>
<summary>✅ v1.0 Stacks & Instances (Phases 1-9) — SHIPPED 2026-02-09</summary>

- [x] Phase 1: Foundation & Guardrails (2/2 plans) — 2026-02-03
- [x] Phase 2: Stack CRUD (5/5 plans) — 2026-02-03
- [x] Phase 3: Service Instances (5/5 plans) — 2026-02-04
- [x] Phase 4: Network Isolation (2/2 plans) — 2026-02-06
- [x] Phase 5: Compose Generation (2/2 plans) — 2026-02-07
- [x] Phase 6: Plan/Apply Workflow (3/3 plans) — 2026-02-07
- [x] Phase 7: Export/Import & Bootstrap (4/4 plans) — 2026-02-08
- [x] Phase 8: Service Wiring (4/4 plans) — 2026-02-08
- [x] Phase 9: Secrets & Resources (3/3 plans) — 2026-02-08

Full details: [milestones/v1.0-ROADMAP.md](milestones/v1.0-ROADMAP.md)

</details>

<details>
<summary>✅ v1.1 Schema Reconciliation (Phases 10-15) — SHIPPED 2026-02-10</summary>

- [x] Phase 10: Fresh Baseline Migrations (3/3 plans) — 2026-02-09
- [x] Phase 11: Parser & Importer Updates (2/2 plans) — 2026-02-09
- [x] Phase 12: Compose Generator Parity (2/2 plans) — 2026-02-09
- [x] Phase 13: Import Scalability (2/2 plans) — 2026-02-09
- [x] Phase 14: Dashboard Updates (3/3 plans) — 2026-02-09
- [x] Phase 15: Validation & Parity (2/2 plans) — 2026-02-10

Full details: [milestones/v1.1-ROADMAP.md](milestones/v1.1-ROADMAP.md)

</details>

### 🚧 v1.1.1 Architecture Hardening (In Progress)

**Milestone Goal:** Harden security model, normalize API contracts, decompose monolithic handlers, optimize query paths, extract frontend controllers, and establish test/observability baselines.

#### Phase 16: Security Configuration
**Goal**: API key loads from environment, not hardcoded in repo
**Depends on**: Nothing (first phase of v1.1.1)
**Requirements**: SEC-01
**Success Criteria** (what must be TRUE):
  1. docker-compose.yml loads TEST_API_KEY from .env file
  2. Repository contains .env.example with placeholder values
  3. .env is gitignored and never committed
  4. API container receives TEST_API_KEY via environment variable
**Plans**: 1 plan

Plans:
- [x] 16-01-PLAN.md — Externalize DEVARCH_API_KEY from compose.yml into .env

#### Phase 17: CORS & Origin Hardening
**Goal**: API enforces origin restrictions for HTTP and WebSocket connections
**Depends on**: Phase 16
**Requirements**: SEC-02, SEC-03
**Success Criteria** (what must be TRUE):
  1. API reads ALLOWED_ORIGINS from config and applies CORS middleware
  2. HTTP requests from disallowed origins receive 403 responses
  3. WebSocket upgrade rejects connections from disallowed origins
  4. Dashboard can connect when running on allowed origin
**Plans**: 1 plan

Plans:
- [x] 17-01-PLAN.md — Wire ALLOWED_ORIGINS into CORS middleware and WebSocket upgrader

#### Phase 18: WebSocket Authentication & Security Modes
**Goal**: WebSocket connections authenticate when API auth enabled; security profiles control auth behavior
**Depends on**: Phase 17
**Requirements**: SEC-04, SEC-05
**Success Criteria** (what must be TRUE):
  1. Browser WS clients include signed token in query parameter when auth enabled
  2. API rejects WS connections with missing or invalid tokens in strict mode
  3. Security mode (dev-open/dev-keyed/strict) configurable via environment
  4. API startup validation fails fast if security mode config invalid
  5. Dev-open mode skips auth checks; dev-keyed validates API key; strict enforces all checks
**Plans**: 2 plans

Plans:
- [x] 18-01-PLAN.md — Security mode profiles (dev-open/dev-keyed/strict) with startup validation
- [x] 18-02-PLAN.md — HMAC-signed WS token auth with dashboard integration

#### Phase 19: API Response Normalization
**Goal**: All endpoints return consistent JSON envelopes for success and errors
**Depends on**: Phase 18
**Requirements**: API-01, API-02
**Success Criteria** (what must be TRUE):
  1. Success responses wrap data in `{"data": ...}` envelope
  2. Error responses use `{"error": {"code", "message", "details"}}` structure
  3. No plain-text http.Error responses remain on core endpoints
  4. Shared responder functions enforce envelope consistency
**Plans**: 4 plans

Plans:
- [x] 19-01-PLAN.md — Create respond package + panic recovery middleware
- [x] 19-02-PLAN.md — Migrate stack handlers to response envelopes
- [x] 19-03-PLAN.md — Migrate instance + service handlers to response envelopes
- [x] 19-04-PLAN.md — Migrate remaining handlers + final audit

#### Phase 20: Action Endpoint Consistency & OpenAPI
**Goal**: Action endpoints share consistent response fields; OpenAPI spec documents all routes
**Depends on**: Phase 19
**Requirements**: API-03, API-04
**Success Criteria** (what must be TRUE):
  1. Start/stop/restart/apply/resolve actions return consistent fields (container_id, status, etc.)
  2. OpenAPI spec generated from router definitions
  3. CI pipeline fails when undocumented endpoint changes detected
  4. Spec includes all route paths, methods, request/response schemas
**Plans**: 2 plans

Plans:
- [x] 20-01-PLAN.md — Standardize action endpoints with ActionResponse struct
- [x] 20-02-PLAN.md — OpenAPI annotations, spec generation, CI validation

#### Phase 21: Deploy Orchestration Service
**Goal**: Deploy orchestration logic (plan/apply/wiring) extracted from handlers into application service
**Depends on**: Phase 20
**Requirements**: BE-01
**Success Criteria** (what must be TRUE):
  1. GeneratePlan function lives in orchestration service, not handler
  2. ApplyPlan function lives in orchestration service, not handler
  3. Wiring logic lives in orchestration service, not handler
  4. Handlers delegate to service layer for all orchestration operations
  5. Plan/apply tests pass with new service layer
**Plans**: 2 plans

Plans:
- [x] 21-01-PLAN.md — Create orchestration service package (GeneratePlan, ApplyPlan, ResolveWiring)
- [x] 21-02-PLAN.md — Refactor handlers to delegate to orchestration service + wire into server

#### Phase 22: Identity Service & Naming Consolidation
**Goal**: All naming logic (stack/instance/network/container) consolidated in identity service
**Depends on**: Phase 21
**Requirements**: BE-02, BE-03
**Success Criteria** (what must be TRUE):
  1. Identity service owns stack naming validation rules
  2. Identity service owns instance naming validation rules
  3. Identity service owns network naming conventions
  4. Identity service owns container naming conventions
  5. No ad-hoc `fmt.Sprintf("devarch-...")` calls remain outside identity service
**Plans**: 2 plans

Plans:
- [x] 22-01-PLAN.md — Create identity package (naming service, validation, labels)
- [x] 22-02-PLAN.md — Migrate all callers to identity package, delete container naming/validation

#### Phase 23: Performance Optimization
**Goal**: Status/metrics batch retrieval, accurate filtered counts, optimized override queries
**Depends on**: Phase 21
**Requirements**: PERF-01, PERF-02, PERF-03
**Success Criteria** (what must be TRUE):
  1. Service list with `include=status,metrics` makes batch runtime calls instead of N+1 per-service calls
  2. X-Total-Count header reflects active filters (category, search term)
  3. Instance list uses aggregated query for override counts (no scalar subquery chain)
  4. Performance tests verify batch operations complete in <100ms for 100 services
**Plans**: 1 plan

Plans:
- [ ] 23-01-PLAN.md — Batch service includes, filtered count query, aggregated override counts

#### Phase 24: Frontend Controller Extraction
**Goal**: Stack/instance/service detail pages delegate orchestration to controller hooks
**Depends on**: Phase 20
**Requirements**: FE-01, FE-02, FE-03, FE-04
**Success Criteria** (what must be TRUE):
  1. Stack detail page uses `useStackDetailController` hook for orchestration
  2. Instance detail page uses `useInstanceDetailController` hook for orchestration
  3. Service detail page uses feature-layer controller hook for orchestration
  4. Mutation boilerplate replaced with shared helper for toast + invalidation
  5. Controller hooks encapsulate query orchestration, state derivation, and action handlers
**Plans**: 3 plans

Plans:
- [x] 24-01-PLAN.md — Create shared mutation helper + stack detail controller extraction
- [x] 24-02-PLAN.md — Instance mutation refactor + instance detail controller extraction
- [x] 24-03-PLAN.md — Service + remaining feature mutation refactor + service detail controller extraction

#### Phase 25: WebSocket Expansion & Frontend Auth
**Goal**: WebSocket invalidates stack/instance queries; browser clients authenticate WS connections
**Depends on**: Phase 18, Phase 24
**Requirements**: FE-05
**Success Criteria** (what must be TRUE):
  1. WebSocket status updates invalidate stack detail queries
  2. WebSocket status updates invalidate instance detail queries
  3. Dashboard WS client includes signed token when API auth enabled
  4. Live container status updates trigger UI refresh via query invalidation
**Plans**: 1 plan

Plans:
- [x] 25-01-PLAN.md — Add stacks/instances predicate invalidation to WS status handler

#### Phase 26: API Integration Tests
**Goal**: Integration tests cover stack/instance CRUD, soft-delete, plan staleness, advisory locks
**Depends on**: Phase 22
**Requirements**: TEST-01, TEST-02
**Success Criteria** (what must be TRUE):
  1. Tests verify stack CRUD operations with DB assertions
  2. Tests verify instance CRUD operations with DB assertions
  3. Tests verify soft-delete semantics (deleted stacks not listed but restorable)
  4. Tests verify plan token staleness detection
  5. Tests verify advisory lock conflicts prevent concurrent applies
  6. CI pipeline runs integration tests and fails build on failure
**Plans**: 2 plans

Plans:
- [x] 26-01-PLAN.md — Test infrastructure (testcontainers, helpers, TestMain) + stack CRUD/soft-delete tests
- [x] 26-02-PLAN.md — Instance CRUD + staleness token + advisory lock tests + CI workflow

#### Phase 27: Frontend Controller Tests
**Goal**: Controller hooks have test coverage for orchestration flows
**Depends on**: Phase 24
**Requirements**: TEST-03
**Success Criteria** (what must be TRUE):
  1. `useStackDetailController` tests cover query orchestration
  2. `useInstanceDetailController` tests cover query orchestration
  3. Service detail controller tests cover query orchestration
  4. Tests verify state derivation logic (loading, error, success states)
  5. Tests verify action handler delegation
**Plans**: 2 plans

Plans:
- [ ] 27-01-PLAN.md — Test infrastructure + instance & stack controller tests
- [ ] 27-02-PLAN.md — Service controller tests + CI workflow

#### Phase 28: Observability Hardening
**Goal**: Structured logging with request correlation; sync job history persists across restarts
**Depends on**: Phase 19
**Requirements**: OPS-01, OPS-02
**Success Criteria** (what must be TRUE):
  1. Core handlers emit structured log fields: request_id, stack, instance, op, duration_ms
  2. Request IDs propagate through call stack for correlation
  3. Sync job summaries persist to DB (table: sync_jobs)
  4. Sync job history survives API process restarts
  5. Logs parseable by structured log tools (JSON format)
**Plans**: TBD

Plans:
- [ ] TBD

## Progress

**Execution Order:**
Phases execute in numeric order: 16 → 17 → 18 → 19 → 20 → 21 → 22 → 23 → 24 → 25 → 26 → 27 → 28

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 16. Security Configuration | 1/1 | ✓ Complete | 2026-02-11 |
| 17. CORS & Origin Hardening | 1/1 | ✓ Complete | 2026-02-11 |
| 18. WebSocket Authentication & Security Modes | 2/2 | ✓ Complete | 2026-02-11 |
| 19. API Response Normalization | 4/4 | ✓ Complete | 2026-02-11 |
| 20. Action Endpoint Consistency & OpenAPI | 2/2 | ✓ Complete | 2026-02-11 |
| 21. Deploy Orchestration Service | 2/2 | ✓ Complete | 2026-02-11 |
| 22. Identity Service & Naming Consolidation | 2/2 | ✓ Complete | 2026-02-11 |
| 23. Performance Optimization | 1/1 | ✓ Complete | 2026-02-11 |
| 24. Frontend Controller Extraction | 3/3 | ✓ Complete | 2026-02-11 |
| 25. WebSocket Expansion & Frontend Auth | 1/1 | ✓ Complete | 2026-02-12 |
| 26. API Integration Tests | 2/2 | ✓ Complete | 2026-02-12 |
| 27. Frontend Controller Tests | 0/TBD | Not started | - |
| 28. Observability Hardening | 0/TBD | Not started | - |

---
*Created: 2026-02-03*
*Last updated: 2026-02-12 — Phase 26 complete*
