# Requirements: DevArch

**Defined:** 2026-02-11
**Core Value:** Two stacks using the same service template must never collide — isolation is the primitive everything else depends on.

## v1.1.1 Requirements

Requirements for v1.1.1 Architecture Hardening. Each maps to roadmap phases.

### Security

- [ ] **SEC-01**: Compose file loads API key from env file, not hardcoded in repo
- [ ] **SEC-02**: API enforces CORS origin allowlist via ALLOWED_ORIGINS config
- [ ] **SEC-03**: WebSocket upgrader rejects connections from disallowed origins
- [ ] **SEC-04**: Browser WS clients authenticate via signed query token when API auth is enabled
- [ ] **SEC-05**: API supports security mode profiles (dev-open, dev-keyed, strict) with startup validation

### API Contract

- [ ] **API-01**: All endpoints use shared responder for `{"data":...}` and `{"error":{"code","message","details"}}` envelopes
- [ ] **API-02**: No `http.Error` plain-text responses remain on core endpoints
- [ ] **API-03**: Action endpoints (start/stop/restart/apply/resolve) share consistent response fields
- [ ] **API-04**: OpenAPI spec is generated from router and CI fails on undocumented endpoint changes

### Backend Decomposition

- [ ] **BE-01**: Deploy orchestration (GeneratePlan, ApplyPlan, wiring) lives in application service, not handlers
- [ ] **BE-02**: Identity service owns stack/instance/network/container naming and validation rules
- [ ] **BE-03**: No ad-hoc `fmt.Sprintf("devarch-...")` naming paths remain outside identity service

### Performance

- [ ] **PERF-01**: Service list with `include=status,metrics` uses batch retrieval instead of per-service runtime calls
- [ ] **PERF-02**: `X-Total-Count` header reflects active filters on service listing
- [ ] **PERF-03**: Instance list uses aggregated query for override counts instead of scalar subquery chain

### Frontend

- [ ] **FE-01**: Stack detail page delegates orchestration to `useStackDetailController` hook
- [ ] **FE-02**: Instance detail page delegates orchestration to `useInstanceDetailController` hook
- [ ] **FE-03**: Service detail page delegates orchestration to feature-layer controller hook
- [ ] **FE-04**: Mutation boilerplate replaced with shared helper for toast + invalidation maps
- [ ] **FE-05**: WebSocket invalidation covers stack/instance query keys for live updates

### Testing

- [ ] **TEST-01**: API integration tests cover stack/instance CRUD + soft-delete semantics
- [ ] **TEST-02**: API integration tests verify plan token staleness and advisory lock conflicts
- [ ] **TEST-03**: Frontend controller tests cover stacks/services/instances detail flows

### Observability

- [ ] **OPS-01**: Core handlers emit structured log fields (request_id, stack, instance, op, duration_ms)
- [ ] **OPS-02**: Sync job summaries persist to DB and survive process restarts

## Future Requirements

### RBAC / Multi-User

- **RBAC-01**: Role-based access control with explicit permission matrix
- **RBAC-02**: Per-stack permission scoping

### Advanced Observability

- **OBS-01**: Distributed tracing pipeline with request correlation
- **OBS-02**: Prometheus-compatible metrics endpoint
- **OBS-03**: Event-driven WS updates (replace polling fallback)

## Out of Scope

| Feature | Reason |
|---------|--------|
| Full RBAC/role matrix | API-key auth sufficient for local dev tool; v2+ consideration |
| Distributed tracing pipeline | Structured logs cover v1.1.1 needs; full tracing is over-engineering |
| Event-driven WS (replacing polling) | Polling works; event-first is optimization for future |
| OpenAPI client SDK generation | Spec generation is sufficient; SDK gen adds maintenance burden |
| Bootstrap package restructure | Boot sequence works; phased startup is over-engineering for single-process |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| SEC-01 | Phase 16 | Pending |
| SEC-02 | Phase 17 | Pending |
| SEC-03 | Phase 17 | Pending |
| SEC-04 | Phase 18 | Pending |
| SEC-05 | Phase 18 | Pending |
| API-01 | Phase 19 | Pending |
| API-02 | Phase 19 | Pending |
| API-03 | Phase 20 | Pending |
| API-04 | Phase 20 | Pending |
| BE-01 | Phase 21 | Pending |
| BE-02 | Phase 22 | Pending |
| BE-03 | Phase 22 | Pending |
| PERF-01 | Phase 23 | Pending |
| PERF-02 | Phase 23 | Pending |
| PERF-03 | Phase 23 | Pending |
| FE-01 | Phase 24 | Pending |
| FE-02 | Phase 24 | Pending |
| FE-03 | Phase 24 | Pending |
| FE-04 | Phase 24 | Pending |
| FE-05 | Phase 25 | Pending |
| TEST-01 | Phase 26 | Pending |
| TEST-02 | Phase 26 | Pending |
| TEST-03 | Phase 27 | Pending |
| OPS-01 | Phase 28 | Pending |
| OPS-02 | Phase 28 | Pending |

**Coverage:**
- v1.1.1 requirements: 25 total
- Mapped to phases: 25
- Unmapped: 0

**Distribution:**
- Phase 16: 1 requirement (SEC-01)
- Phase 17: 2 requirements (SEC-02, SEC-03)
- Phase 18: 2 requirements (SEC-04, SEC-05)
- Phase 19: 2 requirements (API-01, API-02)
- Phase 20: 2 requirements (API-03, API-04)
- Phase 21: 1 requirement (BE-01)
- Phase 22: 2 requirements (BE-02, BE-03)
- Phase 23: 3 requirements (PERF-01, PERF-02, PERF-03)
- Phase 24: 4 requirements (FE-01, FE-02, FE-03, FE-04)
- Phase 25: 1 requirement (FE-05)
- Phase 26: 2 requirements (TEST-01, TEST-02)
- Phase 27: 1 requirement (TEST-03)
- Phase 28: 2 requirements (OPS-01, OPS-02)

---
*Requirements defined: 2026-02-11*
*Last updated: 2026-02-11 after v1.1.1 roadmap creation*
