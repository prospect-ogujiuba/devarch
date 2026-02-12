# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-11)

**Core value:** Two stacks using the same service template must never collide — isolation is the primitive everything else depends on.
**Current focus:** Phase 28 complete — Observability hardening (v1.1.1 milestone complete)

## Current Position

Phase: 28 of 28
Plan: 02 of 02 complete
Status: Phase 28 complete — v1.1.1 milestone complete
Last activity: 2026-02-12 — Observability hardening (structured logging + sync job persistence)

Progress: Phases 16-28 complete (v1.1.1 milestone shipped)

## Performance Metrics

**Velocity:**
- Total plans completed: 30 (v1.0) + 14 (v1.1) + 1 (v1.1.1) + 1 (v1.1.2) + 2 (v1.1.3) + 4 (v1.1.4) + 2 (v1.1.5) + 2 (v1.1.6) + 1 (v1.1.7) + 1 (v1.1.8) + 2 (v1.1.9) + 1 (v1.1.10) + 2 (v1.1.11) + 2 (v1.1.12) + 2 (v1.1.13) = 67
- Average duration: ~5.0 minutes per plan
- Total execution time: ~5.6 hours

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| v1.0 (1-9) | 30 | ~1.6h | ~3.2min |
| v1.1 (10-15) | 14 | ~1.8h | ~7.7min |
| v1.1.1 (16) | 1 | 42s | 42s |
| v1.1.2 (17) | 1 | 98s | 98s |
| v1.1.3 (18) | 2 | 302s | 151s |
| v1.1.4 (19) | 4 | 1392s | 348s |
| v1.1.5 (20) | 2 | 760s | 380s |
| v1.1.6 (21) | 2 | 323s | 161s |
| v1.1.7 (22) | 1 | 43s | 43s |
| v1.1.8 (23) | 1 | 116s | 116s |
| v1.1.9 (24) | 2 | 680s | 340s |
| v1.1.10 (25) | 1 | 32s | 32s |
| v1.1.11 (26) | 2 | 292s | 146s |
| v1.1.12 (27) | 2 | 308s | 154s |
| v1.1.13 (28) | 2 | 607s | 304s |

**Recent Trend:**
- v1.0 shipped successfully on 2026-02-09
- v1.1 shipped successfully on 2026-02-10
- v1.1.1 shipped successfully on 2026-02-12

*Updated after each plan completion*
| Phase 16 P01 | 42 | 1 tasks | 2 files |
| Phase 17 P01 | 98 | 2 tasks | 4 files |
| Phase 18 P01 | 165 | 2 tasks | 7 files |
| Phase 18 P02 | 137 | 2 tasks | 6 files |
| Phase 19 P01 | 91 | 2 tasks | 4 files |
| Phase 19 P02 | 652 | 2 tasks | 9 files |
| Phase 19 P03 | 448 | 1 task (partial) | 5 files |
| Phase 19 P04 | 649 | 2 tasks | 11 files |
| Phase 20 P01 | 240 | 2 tasks | 11 files |
| Phase 20 P02 | 520 | 2 tasks | 20 files |
| Phase 21 P01 | 147 | 1 tasks | 2 files |
| Phase 21 P02 | 176 | 2 tasks | 5 files |
| Phase 22 P01 | 43 | 1 tasks | 3 files |
| Phase 22 P02 | 517 | 2 tasks | 20 files |
| Phase 23 P01 | 116 | 2 tasks | 2 files |
| Phase 24 P01 | 393 | 2 tasks | 4 files |
| Phase 24 P02 | 287 | 2 tasks | 3 files |
| Phase 24 P03 | 557 | 2 tasks | 10 files |
| Phase 25 P01 | 32 | 1 tasks | 1 files |
| Phase 26 P01 | 172 | 2 tasks | 4 files |
| Phase 26 P02 | 120 | 2 tasks | 4 files |
| Phase 27 P01 | 183 | 2 tasks | 4 files |
| Phase 27 P02 | 125 | 2 tasks | 2 files |
| Phase 28 P02 | 157 | 1 tasks | 3 files |
| Phase 28 P01 | 450 | 2 tasks | 8 files |

## Accumulated Context

### Decisions

Decisions are logged in PROJECT.md Key Decisions table.
Recent decisions affecting current work:

- Fresh baseline migrations over incremental (v1.1): Clean slate for domain-separated DDL
- Streaming multipart for large imports (v1.1): Avoids OOM on 256MB payloads
- MaxBodySize scope isolation (Phase 15-02): Register large-limit routes outside restrictive groups
- [Phase 17]: Use comma-separated ALLOWED_ORIGINS env var for both CORS and WebSocket (single source of truth)
- [Phase 17]: Default to wildcard when ALLOWED_ORIGINS unset (dev-friendly, no breaking changes)
- [Phase 18-01]: Empty SECURITY_MODE defaults to dev-open for backward compatibility
- [Phase 18-01]: Mode validation at startup prevents invalid runtime configuration
- [Phase 18]: HMAC-SHA256 signing using DEVARCH_API_KEY as secret (no additional secret needed)
- [Phase 18]: 60-second WS token TTL (sufficient for upgrade handshake)
- [Phase 19-01]: Envelope structure separates success (data) from errors (error.code/message/details)
- [Phase 19-01]: InternalError logs full error server-side but returns generic message to client
- [Phase 19-01]: RecoverEnvelope replaces chi Recoverer for JSON panic responses
- [Phase 19-02]: ExportStack YAML response remains exempt from envelope (retains application/x-yaml content-type)
- [Phase 19-04]: Auth Validate endpoint returns JSON envelope {"valid": true} instead of empty 204 for parseable client responses
- [Phase 20-02]: Full annotation coverage — 143 routes annotated across 24 handler files (120 unique spec paths)
- [Phase 20-02]: Envelope response syntax uses swaggo nested `respond.SuccessEnvelope{data=TYPE}` for data and `{data=respond.ActionResponse}` for actions
- [Phase 21-01]: Service layer accepts Go types only — no net/http imports for transport independence
- [Phase 21-01]: Sentinel errors enable handlers to map service errors to HTTP status codes
- [Phase 21-02]: Orchestration service created in NewRouter and injected into StackHandler (consistent with existing handler creation pattern)
- [Phase 21-02]: Orchestration service created in NewRouter and injected into StackHandler (consistent with existing handler creation pattern)
- [Phase 22]: Package-level functions only (no struct) — no DB dependency needed per research
- [Phase 22]: Accept custom names as parameters — transport-agnostic per Phase 21 decision
- [Phase 22]: New ValidateLabelKey function enforces devarch.* prefix reservation
- [Phase 23-01]: Batch loading via GetBatchServiceData eliminates N+1 queries for service includes
- [Phase 23-01]: Shared filter clause ensures X-Total-Count matches filtered results
- [Phase 23-01]: UNION ALL + GROUP BY pattern replaces 11 scalar subqueries with single aggregated query
- [Phase 24-01]: useMutationHelper factory pattern for DRY toast+invalidation boilerplate
- [Phase 24-01]: Controller hooks consolidate query orchestration - business logic separate from presentation
- [Phase 24-03]: ProxyConfigPanel error type changed to unknown for useMutationHelper compatibility
- [Phase 24-03]: All 9 feature query files migrated to useMutationHelper pattern (FE-04 complete)
- [Phase 24-02]: Instance mutations use useMutationHelper - 21 mutations refactored with consistent error handling
- [Phase 24-02]: Override mutations preserve effective-config invalidation for real-time UI sync
- [Phase 25]: Single predicate covers all stack and instance queries via ['stacks', ...] prefix
- [Phase 27-01]: test-utils.tsx extension required for JSX support in wrapper component
- [Phase 27-01]: localStorage mock in setup.ts enables testing components importing API module
- [Phase 27-01]: Mock factory helpers (mockQueryResult, mockMutation) reduce test boilerplate
- [Phase 27-02]: All 3 controller hooks tested following consistent patterns (TEST-03 complete)
- [Phase 27-02]: CI workflow triggers only on dashboard/** changes for efficiency
- [Phase 28]: Write-through persistence removes completed jobs from memory after successful DB insert
- [Phase 28]: GetJobs merges in-memory (running) + DB (completed) for unified view
- [Phase 28]: 7-day retention integrated into existing daily cleanup cycle
- [Phase 28]: SlogMiddleware replaces chi middleware.Logger for structured JSON logging
- [Phase 28]: Request-scoped logger with request_id propagated via context

### Pending Todos

None.

### Blockers/Concerns

None.

## Session Continuity

Last session: 2026-02-12
Stopped at: Completed Phase 28 — v1.1.1 milestone complete
Resume file: None
Next: /gsd:complete-milestone

---
*Last updated: 2026-02-12*
