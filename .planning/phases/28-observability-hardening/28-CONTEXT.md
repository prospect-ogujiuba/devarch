# Phase 28: Observability Hardening - Context

**Gathered:** 2026-02-12
**Status:** Ready for planning

<domain>
## Phase Boundary

Structured logging with request correlation across all API handlers; sync job history persists to DB and survives process restarts. No new endpoints, no dashboard changes, no alerting.

</domain>

<decisions>
## Implementation Decisions

### Logging library
- `log/slog` (stdlib) — Go 1.22 project, no new dependencies, built-in JSON handler
- Replace stdlib `log` usage in handler/service code with slog

### Log format
- JSON output in all environments (no text mode toggle)
- Locked by success criteria: "Logs parseable by structured log tools (JSON format)"

### Log level control
- `LOG_LEVEL` env var, default `info`
- Follows existing env-var config pattern (ALLOWED_ORIGINS, SECURITY_MODE)

### Core handler scope
- All handler methods serving `/api/v1` routes
- Excludes CLI tools (cmd/migrate, cmd/import) and test code

### Logger injection
- Request-scoped slog.Logger created in logging middleware
- Middleware extracts `request_id` from chi's existing RequestID middleware, attaches logger to context
- Replaces chi's built-in `middleware.Logger`
- Handlers and services extract logger from context

### Structured log fields
- `request_id` — from chi RequestID middleware (propagated via context)
- `op` — derived from `chi.RouteContext().RoutePattern()` (e.g., `GET /api/v1/stacks/{stack}`)
- `stack` / `instance` — extracted from chi URL params when present on route
- `duration_ms` — middleware-level timing wrapping handler execution

### Service layer logging
- Services extract logger from `context.Context` (already accept context)
- No net/http imports — maintains Phase 21 transport independence

### Sync job persistence — table schema
- New `sync_jobs` table: id (text PK), type, status, started_at, ended_at, error, created_at
- New migration file, mirrors existing in-memory Job struct fields

### Sync job persistence — storage pattern
- Write-through: active jobs tracked in-memory during execution, persisted to DB on completion
- `GetJobs` reads from DB (recent history) instead of in-memory map
- Satisfies "survives restart" requirement

### Sync job retention
- 7-day TTL, purged by existing daily cleanup cycle in sync manager
- Prevents unbounded growth

### Claude's Discretion
- Exact slog handler configuration (output destination, source location inclusion)
- Whether to add `method` and `path` as separate fields beyond `op`
- Sync job summary fields beyond the core struct (e.g., items_processed count)
- Error log verbosity levels for different failure modes

</decisions>

<specifics>
## Specific Ideas

No specific requirements — open to standard approaches

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 28-observability-hardening*
*Context gathered: 2026-02-12*
