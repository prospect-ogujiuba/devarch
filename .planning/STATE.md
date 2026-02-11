# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-11)

**Core value:** Two stacks using the same service template must never collide — isolation is the primitive everything else depends on.
**Current focus:** Phase 19: API Response Normalization

## Current Position

Phase: 19 of 28 (in progress)
Plan: 19-03 (partial)
Status: Plan partially complete — service.go migration requires continuation
Last activity: 2026-02-11 — Completed instance handler migration (19-03 Task 1)

Progress: Phases 16-18 complete, Phase 19 in progress (1.5 plans complete)

## Performance Metrics

**Velocity:**
- Total plans completed: 30 (v1.0) + 14 (v1.1) + 1 (v1.1.1) + 1 (v1.1.2) + 2 (v1.1.3) + 1.5 (v1.1.4) = 49.5
- Average duration: ~4.3 minutes per plan
- Total execution time: ~3.7 hours

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| v1.0 (1-9) | 30 | ~1.6h | ~3.2min |
| v1.1 (10-15) | 14 | ~1.8h | ~7.7min |
| v1.1.1 (16) | 1 | 42s | 42s |
| v1.1.2 (17) | 1 | 98s | 98s |
| v1.1.3 (18) | 2 | 302s | 151s |
| v1.1.4 (19) | 1.5 | 539s | 269s |

**Recent Trend:**
- v1.0 shipped successfully on 2026-02-09
- v1.1 shipped successfully on 2026-02-10
- v1.1.1-v1.1.3 in progress

*Updated after each plan completion*
| Phase 16 P01 | 42 | 1 tasks | 2 files |
| Phase 17 P01 | 98 | 2 tasks | 4 files |
| Phase 18 P01 | 165 | 2 tasks | 7 files |
| Phase 18 P02 | 137 | 2 tasks | 6 files |
| Phase 19 P01 | 91 | 2 tasks | 4 files |
| Phase 19 P03 | 448 | 1 task (partial) | 5 files |

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

### Pending Todos

None.

### Blockers/Concerns

None.

## Session Continuity

Last session: 2026-02-11
Stopped at: Completed Phase 19 Plan 03 (partial — Task 1 complete, Task 2 requires continuation)
Resume file: None
Next: Complete service.go migration (19-03 Task 2) or continue with next plan

---
*Last updated: 2026-02-11*
