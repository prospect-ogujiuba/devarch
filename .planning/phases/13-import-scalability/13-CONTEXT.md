# Phase 13: Import Scalability - Context

**Gathered:** 2026-02-09
**Status:** Ready for planning

<domain>
## Phase Boundary

Stack import handles large payloads without memory exhaustion. Streaming multipart read replaces full buffering, route-specific size cap overrides global limit, bulk writes use prepared statements in single transaction, and re-importing the same data succeeds idempotently. No new import features — strictly scalability and robustness of existing import flow.

</domain>

<decisions>
## Implementation Decisions

### Streaming strategy
- Replace `ParseMultipartForm` with `multipart.NewReader(r.Body, boundary)` — reads part-by-part without buffering entire form
- Apply `io.LimitReader` on the multipart part (not the body) so the 256MB cap targets YAML content size, not multipart overhead
- Feed part directly to `yaml.NewDecoder(part)` — already accepts `io.Reader`, no intermediate buffer needed
- Forward consideration: Phase 15 tests 200MB success / 300MB rejection — part-level limit makes boundary behavior predictable for test assertions

### Size limit architecture
- Route-level override via `r.With(mw.MaxBodySize(importMax)).Post("/import", ...)` — chi processes innermost `MaxBytesReader` last, so route-level replaces global
- Global `r.Use(mw.MaxBodySize(10 << 20))` stays untouched — all non-import routes keep 10MB default
- Import cap read from `STACK_IMPORT_MAX_BYTES` env var, default 256MB (`256 << 20`)
- Forward consideration: Phase 15 can assert rejection threshold against env var value; no hardcoded magic numbers in tests

### Conflict/idempotency approach
- `ON CONFLICT DO UPDATE` (upsert) for all insert paths — stacks by name, instances by stack_id + service_name, wires by stack_id + source + target
- Delete-then-reinsert rejected: breaks FK cascades, creates ID gaps, unnecessary WAL churn
- Forward consideration: Phase 14 dashboard CRUD and import operations don't interfere — re-import after dashboard edits updates cleanly without orphaning records. Phase 15 re-import of golden services must be idempotent.

### Batching strategy
- `tx.Prepare()` once per entity type, `stmt.Exec()` in loop, all within single `tx.Begin()/tx.Commit()`
- No external batching library — Go's `sql.Tx` + `sql.Stmt` handles this natively
- `COPY` rejected: doesn't support `ON CONFLICT`. Multi-row `INSERT ... VALUES` rejected: dynamic SQL construction + 65K parameter limit
- Forward consideration: Prepared statements amortize parse cost across thousands of services (Phase 15's 200MB test). Single transaction = all-or-nothing semantics.

### Claude's Discretion
- Override delete pattern for instance overrides (delete-then-reinsert vs upsert per override type)
- Error message format for size limit rejection
- Whether to log import duration/stats
- Chunk size for streaming reads if buffering is needed for YAML parsing

</decisions>

<specifics>
## Specific Ideas

- Chi `r.With()` pattern chosen over restructuring route groups — additive, not disruptive to existing route registration
- Env var `STACK_IMPORT_MAX_BYTES` for runtime configurability without code change
- All choices validated against Phase 14 (dashboard stability) and Phase 15 (boundary testing) requirements

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 13-import-scalability*
*Context gathered: 2026-02-09*
