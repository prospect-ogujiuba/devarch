# Phase 28: Observability Hardening - Research

**Researched:** 2026-02-12
**Domain:** Structured logging with Go log/slog, sync job persistence to PostgreSQL
**Confidence:** HIGH

## Summary

Phase 28 implements structured JSON logging across 148 HTTP handlers using Go's stdlib `log/slog` (Go 1.21+), replacing unstructured `log` package calls. Request-scoped loggers propagate through context with correlation IDs from chi's RequestID middleware. Sync job history moves from in-memory-only to DB-persisted storage, surviving process restarts. The phase is API-only, no dashboard changes, no new endpoints, no alerting infrastructure.

User decisions (CONTEXT.md) lock down: `log/slog` as the logging library, JSON format everywhere, `LOG_LEVEL` env var for level control, request-scoped logger injection via middleware, sync jobs table schema (`id`, `type`, `status`, `started_at`, `ended_at`, `error`, `created_at`), and write-through persistence (active jobs in-memory during execution, persisted on completion). Claude has discretion over exact handler configuration, additional log fields beyond required set, and sync job summary fields.

**Primary recommendation:** Use `slog.JSONHandler` with `HandlerOptions.Level` set to `LevelVar` for runtime control, inject request-scoped loggers via custom middleware replacing chi's built-in `middleware.Logger`, extract logger from context in handlers/services with type-safe context key pattern, persist sync jobs with `TEXT` primary key matching timestamp-based ID format, index on `created_at` DESC for recent history queries.

## User Constraints

<user_constraints>
### Locked Decisions (from CONTEXT.md)

**Logging library:** `log/slog` (stdlib) — Go 1.22 project (actually Go 1.24 per go.mod), no new dependencies, built-in JSON handler. Replace stdlib `log` usage in handler/service code with slog.

**Log format:** JSON output in all environments (no text mode toggle). Locked by success criteria: "Logs parseable by structured log tools (JSON format)".

**Log level control:** `LOG_LEVEL` env var, default `info`. Follows existing env-var config pattern (ALLOWED_ORIGINS, SECURITY_MODE).

**Core handler scope:** All handler methods serving `/api/v1` routes. Excludes CLI tools (cmd/migrate, cmd/import) and test code.

**Logger injection:** Request-scoped slog.Logger created in logging middleware. Middleware extracts `request_id` from chi's existing RequestID middleware, attaches logger to context. Replaces chi's built-in `middleware.Logger`. Handlers and services extract logger from context.

**Structured log fields:** `request_id` (from chi RequestID middleware via context), `op` (derived from `chi.RouteContext().RoutePattern()` e.g. `GET /api/v1/stacks/{stack}`), `stack`/`instance` (extracted from chi URL params when present on route), `duration_ms` (middleware-level timing wrapping handler execution).

**Service layer logging:** Services extract logger from `context.Context` (already accept context). No net/http imports — maintains Phase 21 transport independence.

**Sync job persistence — table schema:** New `sync_jobs` table: `id` (text PK), `type`, `status`, `started_at`, `ended_at`, `error`, `created_at`. New migration file, mirrors existing in-memory Job struct fields.

**Sync job persistence — storage pattern:** Write-through: active jobs tracked in-memory during execution, persisted to DB on completion. `GetJobs` reads from DB (recent history) instead of in-memory map. Satisfies "survives restart" requirement.

**Sync job retention:** 7-day TTL, purged by existing daily cleanup cycle in sync manager. Prevents unbounded growth.

### Claude's Discretion

- Exact slog handler configuration (output destination, source location inclusion)
- Whether to add `method` and `path` as separate fields beyond `op`
- Sync job summary fields beyond the core struct (e.g., items_processed count)
- Error log verbosity levels for different failure modes

### Deferred Ideas (OUT OF SCOPE)

None — discussion stayed within phase scope.
</user_constraints>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| log/slog | stdlib (Go 1.21+) | Structured logging | Official Go stdlib, zero dependencies, designed for production use, built-in JSON handler |
| github.com/go-chi/chi/v5 | v5.1.0 | Router with RequestID middleware | Already in use, provides `middleware.RequestID` and `middleware.GetReqID()` for correlation |
| database/sql | stdlib | Sync job persistence | Already used for all DB operations, no ORM needed |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| testing/slogtest | stdlib | Handler testing | Unit tests verifying log output (optional for this phase) |
| log (stdlib bridge) | stdlib | Bridge old log calls | slog.SetDefault() redirects stdlib log to slog handler |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| log/slog | zerolog, zap | Faster (zap: 420 ns/op vs slog: 650 ns/op) but require dependencies; slog is stdlib, sufficient for non-extreme throughput |
| JSON handler | Text handler | Human-readable but not parseable by log aggregation tools; JSON locked by success criteria |
| Context-based logger | Global logger | Simpler but loses request correlation; context-based required for request_id propagation |

**Installation:**
No new dependencies — `log/slog` is stdlib in Go 1.21+, project uses Go 1.24.

## Architecture Patterns

### Recommended Project Structure
```
api/internal/api/
├── middleware/
│   ├── middleware.go         # existing middleware
│   └── logging.go            # NEW: slog middleware replacing chi's middleware.Logger
├── handlers/                  # 24 handler files, 148 handler methods
│   └── *.go                   # extract logger from context, log with structured fields
api/internal/sync/
├── manager.go                 # add DB persistence to TriggerSync/GetJobs
api/migrations/
└── 013_sync_jobs.up.sql       # NEW: sync_jobs table
```

### Pattern 1: Request-Scoped Logger Middleware
**What:** Middleware creates `slog.Logger` with request-scoped attributes, stores in context, wraps response writer for timing
**When to use:** Replace chi's `middleware.Logger` on all `/api/v1` routes
**Example:**
```go
// Source: https://betterstack.com/community/guides/logging/golang-contextual-logging/
// and https://github.com/samber/slog-chi/blob/main/middleware.go

type contextKey string
const LoggerCtxKey contextKey = "logger"

func SlogMiddleware(base *slog.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()

            // Extract request_id from chi's RequestID middleware
            requestID := middleware.GetReqID(r.Context())

            // Create request-scoped logger with base attributes
            logger := base.With(
                "request_id", requestID,
            )

            // Attach logger to context
            ctx := context.WithValue(r.Context(), LoggerCtxKey, logger)

            // Wrap response writer to capture status code
            ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

            defer func() {
                // Extract route pattern AFTER handler runs (RouteContext populated)
                op := ""
                if rctx := chi.RouteContext(ctx); rctx != nil {
                    op = rctx.RoutePattern()
                }

                // Extract stack/instance from URL params if present
                var stack, instance string
                if rctx := chi.RouteContext(ctx); rctx != nil {
                    stack = rctx.URLParam("name")      // stack name
                    instance = rctx.URLParam("instance") // instance name
                }

                duration := time.Since(start).Milliseconds()

                // Log completed request with all structured fields
                logger.Info("request completed",
                    "op", op,
                    "stack", stack,
                    "instance", instance,
                    "duration_ms", duration,
                    "status", ww.Status(),
                    "bytes", ww.BytesWritten(),
                )
            }()

            next.ServeHTTP(ww, r.WithContext(ctx))
        })
    }
}
```

### Pattern 2: Extract Logger from Context
**What:** Type-safe context key retrieval of request-scoped logger
**When to use:** All handlers and services needing to log
**Example:**
```go
// Source: https://betterstack.com/community/guides/logging/golang-contextual-logging/

// In handler or service
func (h *Handler) SomeMethod(w http.ResponseWriter, r *http.Request) {
    logger := r.Context().Value(LoggerCtxKey).(*slog.Logger)

    logger.Info("processing request", "service", "example")

    // Pass context to services — logger propagates automatically
    err := h.service.DoWork(r.Context(), data)
    if err != nil {
        logger.Error("operation failed", "error", err)
        // ... respond with error
    }
}

// In service layer (maintains transport independence)
func (s *Service) DoWork(ctx context.Context, data string) error {
    logger := ctx.Value(LoggerCtxKey).(*slog.Logger)

    logger.Debug("starting work", "data_length", len(data))
    // ... do work
    return nil
}
```

### Pattern 3: Sync Job DB Persistence
**What:** Write-through pattern — active jobs in-memory, persist on completion, read from DB
**When to use:** Sync manager's TriggerSync and GetJobs methods
**Example:**
```go
// Source: Existing sync.Manager pattern adapted for DB

func (m *Manager) TriggerSync(syncType string) string {
    jobID := time.Now().Format("20060102150405") // TEXT primary key

    job := &Job{
        ID:        jobID,
        Type:      syncType,
        Status:    "running",
        StartedAt: time.Now(),
    }

    // Track in-memory during execution
    m.jobsMu.Lock()
    m.jobs[jobID] = job
    m.jobsMu.Unlock()

    go func() {
        ctx := context.Background()
        var err error

        // ... execute sync work ...

        m.jobsMu.Lock()
        now := time.Now()
        job.EndedAt = &now
        if err != nil {
            job.Status = "failed"
            job.Error = err.Error()
        } else {
            job.Status = "completed"
        }

        // PERSIST to DB on completion (write-through)
        _, dbErr := m.db.Exec(`
            INSERT INTO sync_jobs (id, type, status, started_at, ended_at, error, created_at)
            VALUES ($1, $2, $3, $4, $5, $6, NOW())
        `, job.ID, job.Type, job.Status, job.StartedAt, job.EndedAt, job.Error)
        if dbErr != nil {
            log.Printf("sync: failed to persist job %s: %v", jobID, dbErr)
        }

        // Remove from in-memory map after persistence
        delete(m.jobs, jobID)
        m.jobsMu.Unlock()
    }()

    return jobID
}

func (m *Manager) GetJobs() []*Job {
    // Read from DB (recent history survives restart)
    rows, err := m.db.Query(`
        SELECT id, type, status, started_at, ended_at, error, created_at
        FROM sync_jobs
        ORDER BY created_at DESC
        LIMIT 100
    `)
    if err != nil {
        return nil
    }
    defer rows.Close()

    var jobs []*Job
    for rows.Next() {
        var job Job
        rows.Scan(&job.ID, &job.Type, &job.Status, &job.StartedAt, &job.EndedAt, &job.Error, &job.CreatedAt)
        jobs = append(jobs, &job)
    }
    return jobs
}
```

### Pattern 4: Dynamic Log Level Control
**What:** `slog.LevelVar` enables runtime level changes via env var
**When to use:** Server initialization, read `LOG_LEVEL` env var
**Example:**
```go
// Source: https://pkg.go.dev/log/slog

var programLevel = new(slog.LevelVar) // Defaults to LevelInfo

func initLogger() *slog.Logger {
    // Parse LOG_LEVEL env var
    levelStr := os.Getenv("LOG_LEVEL")
    var level slog.Level
    switch strings.ToLower(levelStr) {
    case "debug":
        level = slog.LevelDebug
    case "info", "":
        level = slog.LevelInfo
    case "warn", "warning":
        level = slog.LevelWarn
    case "error":
        level = slog.LevelError
    default:
        level = slog.LevelInfo
    }
    programLevel.Set(level)

    handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: programLevel,
        AddSource: false, // Set true for file:line (slower, useful for debugging)
    })

    logger := slog.New(handler)
    slog.SetDefault(logger) // Bridge stdlib log calls to slog

    return logger
}
```

### Anti-Patterns to Avoid
- **Global logger without context:** Loses request correlation, violates success criteria requirement for `request_id` in all core handlers
- **Log in middleware with handler details:** RouteContext is nil until handler runs; must log in deferred function after handler completes
- **Separate migration for cleanup logic:** Reuse existing daily cleanup cycle in sync manager (already has advisory lock, batching, timeout handling)
- **UUID primary key for sync jobs:** Existing Job.ID uses timestamp format `20060102150405`; TEXT primary key matches, no schema change needed

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Request ID generation | Custom ID format | chi `middleware.RequestID` | Already deployed, generates `hostname/random-XXXXXX` format, increments atomically, preserves `X-Request-Id` header if present |
| Response writer wrapping | Custom status capture | chi `middleware.NewWrapResponseWriter` | Already used in chi's logger, captures status code and bytes written, works across HTTP versions |
| Log level parsing | String switch | slog `Level.UnmarshalText()` | Built-in parser handles "DEBUG", "INFO", "WARN", "ERROR" case-insensitively, returns error for invalid values |
| JSON serialization | Manual JSON building | `slog.JSONHandler` | Optimized single-write serialization, handles special characters, timestamp formatting, structured attribute nesting |
| Context value extraction | `interface{}` casting everywhere | Helper function with panic recovery | Centralizes nil check and type assertion, prevents panics if middleware skipped |

**Key insight:** Structured logging middleware is deceptively complex — request ID correlation, route pattern extraction timing (RouteContext populated after routing), response status capture, and panic recovery all have edge cases. Chi's existing middleware primitives solve these; slog-chi reference implementation (https://github.com/samber/slog-chi) demonstrates integration patterns.

## Common Pitfalls

### Pitfall 1: Extracting RoutePattern Too Early
**What goes wrong:** `chi.RouteContext(ctx).RoutePattern()` returns empty string, logs show `op: ""` instead of `GET /api/v1/stacks/{name}`
**Why it happens:** Chi populates RouteContext during routing, AFTER middleware runs but BEFORE handler executes. Extracting in middleware function body (before `next.ServeHTTP`) sees nil RouteContext.
**How to avoid:** Extract route pattern in deferred function that runs AFTER handler completes.
**Warning signs:** All logs have empty `op` field, route patterns never appear in logs.

### Pitfall 2: Nil Logger Panics in Services
**What goes wrong:** Service calls `ctx.Value(LoggerCtxKey).(*slog.Logger)`, gets nil, panics with "invalid memory address or nil pointer dereference"
**Why it happens:** Logger middleware skipped (e.g., unit test bypasses middleware, CLI tool calls service directly, health check endpoint exempt from middleware)
**How to avoid:** Provide fallback: helper function checks nil and returns `slog.Default()` if logger not in context. Alternative: always attach logger in tests/CLI tools.
**Warning signs:** Panics in service layer, stack trace shows type assertion on context value retrieval.

### Pitfall 3: Over-Logging in Loops
**What goes wrong:** Handler logs inside loop over 1000 instances, generates 1000 log lines per request, floods logs, slows request processing
**Why it happens:** Structured logging makes it easy to log everywhere; performance impact not obvious until production scale
**How to avoid:** Log summary after loop completes, not individual iterations. Use `slog.Debug` for detailed iteration logging (disabled in production with `LOG_LEVEL=info`).
**Warning signs:** Log volume spikes correlate with request count × collection size, handler latency increases with collection size.

### Pitfall 4: Logging Sensitive Data
**What goes wrong:** Logs contain passwords, API keys, secrets from env vars or request bodies
**Why it happens:** Structured logging's ease of adding fields makes it tempting to log entire request/response objects
**How to avoid:** Never log: passwords, tokens, API keys, secrets, full request bodies (especially POST/PUT), connection strings. Use `ReplaceAttr` in HandlerOptions to redact sensitive keys.
**Warning signs:** grep logs for "password", "token", "secret" — if matches found, data leaking.

### Pitfall 5: Incorrect Log Levels
**What goes wrong:** INFO used for debug details (loop iterations, variable values), ERROR used for expected failures (validation errors)
**Why it happens:** Unclear understanding of level semantics in production
**How to avoid:** Follow level guidelines — DEBUG: internal state, INFO: significant business events, WARN: degraded but functional, ERROR: operation failure requiring attention, per https://betterstack.com/community/guides/logging/log-levels-explained/
**Warning signs:** Production logs at INFO level are noisy with internal details, genuine errors buried in noise.

## Code Examples

Verified patterns from official sources:

### Handler Options Configuration
```go
// Source: https://pkg.go.dev/log/slog

opts := &slog.HandlerOptions{
    // Level: minimum log level (nil = INFO)
    // Use LevelVar for dynamic runtime control
    Level: programLevel,

    // AddSource: include file:line in log output
    // Adds ~40% overhead, useful for debugging, disable in production
    AddSource: false,

    // ReplaceAttr: redact sensitive fields
    ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
        // Redact password/token/secret keys
        if a.Key == "password" || a.Key == "token" || a.Key == "secret" {
            return slog.String(a.Key, "REDACTED")
        }
        return a
    },
}

handler := slog.NewJSONHandler(os.Stdout, opts)
logger := slog.New(handler)
```

### Chi Request ID Extraction
```go
// Source: https://github.com/go-chi/chi/blob/master/middleware/request_id.go

import "github.com/go-chi/chi/v5/middleware"

// In middleware
requestID := middleware.GetReqID(r.Context())

// Chi's RequestID middleware must be installed before logging middleware:
// r.Use(middleware.RequestID)   // MUST come first
// r.Use(SlogMiddleware(logger))  // Can extract request_id
```

### Chi Route Pattern Extraction
```go
// Source: https://github.com/samber/slog-chi/blob/main/middleware.go

// Extract route pattern from RouteContext
var op string
if rctx := chi.RouteContext(ctx); rctx != nil {
    op = rctx.RoutePattern()  // e.g., "GET /api/v1/stacks/{name}"
}

// Extract URL params
var stack, instance string
if rctx := chi.RouteContext(ctx); rctx != nil {
    stack = rctx.URLParam("name")       // {name} param -> stack name
    instance = rctx.URLParam("instance") // {instance} param -> instance name
}
```

### Sync Jobs Table Migration
```sql
-- Source: Adapted from existing migration 008_wiring_contracts_sync_security.up.sql

CREATE TABLE sync_jobs (
    id TEXT PRIMARY KEY,                    -- timestamp format: 20060102150405
    type VARCHAR(32) NOT NULL,              -- 'containers', 'metrics', 'registry', 'trivy', 'all'
    status VARCHAR(32) NOT NULL,            -- 'running', 'completed', 'failed'
    started_at TIMESTAMPTZ NOT NULL,
    ended_at TIMESTAMPTZ,                   -- NULL while running
    error TEXT,                             -- NULL if status='completed'
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for GetJobs() query (ORDER BY created_at DESC LIMIT 100)
CREATE INDEX idx_sync_jobs_created_at ON sync_jobs(created_at DESC);

-- Cleanup handled by existing daily cleanup cycle in sync.Manager
```

### Daily Cleanup for Sync Jobs
```go
// Source: Adapted from sync/manager.go cleanupLoop pattern

func (m *Manager) cleanupSyncJobs(ctx context.Context) error {
    // 7-day retention from CONTEXT.md
    cutoff := time.Now().Add(-7 * 24 * time.Hour)

    _, err := m.deleteInBatches(ctx,
        `WITH doomed AS (
            SELECT id FROM sync_jobs
            WHERE created_at < $1
            LIMIT $2
        )
        DELETE FROM sync_jobs
        USING doomed WHERE sync_jobs.id = doomed.id`,
        cutoff,
        m.cleanupBatch,
        "cleanup old sync jobs",
    )
    return err
}

// Add to runDailyCleanupIfDue() ops list:
// {"sync jobs", m.cleanupSyncJobs}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| stdlib `log` package | `log/slog` structured logging | Go 1.21 (Aug 2023) | JSON output, key-value pairs, handler abstraction, level control, context integration |
| Manual context key types | `slog.Logger` in context | slog release | Standardized pattern for request-scoped loggers |
| Third-party logging (zap, zerolog, logrus) | stdlib `log/slog` | Go 1.21+ | Zero dependencies, good-enough performance (650 ns/op vs zap 420 ns/op), official support |
| Global logger with printf | Request-scoped structured logger | 2020+ | Request correlation via request_id, automated attribute propagation |

**Deprecated/outdated:**
- `log.Printf` unstructured logging: Replaced by `slog.Info/Warn/Error` with structured fields, though stdlib bridge exists via `slog.SetDefault()`
- Global logger pattern: Replaced by context-propagated request-scoped loggers for correlation
- Text handler in production: JSON handler standard for log aggregation/parsing tools

## Open Questions

1. **Should AddSource be enabled for ERROR level only?**
   - What we know: `AddSource: true` adds 40% overhead but provides file:line for debugging
   - What's unclear: Whether production error logs need source location for troubleshooting
   - Recommendation: Start with `AddSource: false`, enable temporarily via env var if debugging production errors. Avoid always-on overhead.

2. **How to handle startup logs before middleware attached?**
   - What we know: Server initialization (DB connect, migrations, container client) happens before routes registered
   - What's unclear: Whether startup logs need request_id (they don't have requests)
   - Recommendation: Use base logger (no request_id) for startup, request-scoped logger only attaches after middleware. Acceptable — startup logs are one-time per process.

3. **Should method and path be separate fields beyond op?**
   - What we know: `op` contains full route pattern (e.g., `GET /api/v1/stacks/{name}`)
   - What's unclear: Whether log aggregation queries benefit from separate `method` and `path` fields
   - Recommendation: Start with `op` only (one field, matches route definition), add `method`/`path` if log analysis shows query patterns that need them. Marked as Claude's discretion.

4. **Should GetJobs return in-progress jobs from in-memory map?**
   - What we know: Write-through pattern persists on completion, GetJobs reads DB
   - What's unclear: Whether API consumers need to see currently-running jobs (not yet in DB)
   - Recommendation: Merge in-memory (running) + DB (completed) for GetJobs response. Provides complete picture, survives restart (only completed jobs, but that's expected).

## Sources

### Primary (HIGH confidence)
- [log/slog official docs](https://pkg.go.dev/log/slog) — API reference, HandlerOptions, Level types, LevelVar
- [Structured Logging with slog - Go Blog](https://go.dev/blog/slog) — Design rationale, performance considerations, handler patterns
- [slog-chi middleware reference implementation](https://github.com/samber/slog-chi/blob/main/middleware.go) — Chi integration, RoutePattern extraction, request ID correlation
- [chi middleware.RequestID source](https://github.com/go-chi/chi/blob/master/middleware/request_id.go) — RequestIDKey constant, GetReqID function, context storage mechanism
- [chi middleware package docs](https://pkg.go.dev/github.com/go-chi/chi/v5/middleware) — NewWrapResponseWriter, RequestID middleware

### Secondary (MEDIUM confidence)
- [Contextual Logging in Go with Slog | Better Stack](https://betterstack.com/community/guides/logging/golang-contextual-logging/) — Context key pattern, middleware example, logger extraction
- [Log Levels Explained | Better Stack](https://betterstack.com/community/guides/logging/log-levels-explained/) — DEBUG/INFO/WARN/ERROR usage guidelines, when to use each
- [Logging in Go with Slog: The Ultimate Guide | Better Stack](https://betterstack.com/community/guides/logging/logging-in-go/) — JSON handler configuration, best practices, performance optimization
- [High-Performance Structured Logging in Go | Leapcell](https://leapcell.io/blog/high-performance-structured-logging-in-go-with-slog-and-zerolog) — Performance benchmarks (slog: 650 ns/op, 48 B/op, 1 alloc/op), sync.Pool for buffer reuse
- [Database Design for Audit Logging | Redgate](https://www.red-gate.com/blog/database-design-for-audit-logging) — Audit log table best practices, indexing strategies
- [pg_cron job_run_details schema | AWS](https://aws.amazon.com/blogs/database/schedule-jobs-with-pg_cron-on-your-amazon-rds-for-postgresql-or-amazon-aurora-for-postgresql-databases/) — Job history table reference (jobid, status, start_time, end_time, return_message)

### Tertiary (LOW confidence)
- [testing/slogtest package](https://pkg.go.dev/testing/slogtest) — Handler testing utilities (mentioned for completeness, testing out of scope for this phase)

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — `log/slog` is stdlib Go 1.21+, chi v5 already in use, patterns verified in official docs
- Architecture: HIGH — Request-scoped logger pattern standard (Better Stack, slog-chi reference), sync job persistence mirrors existing cleanup patterns
- Pitfalls: MEDIUM-HIGH — RouteContext timing verified in slog-chi source, nil logger issue common in context patterns, over-logging/sensitive data/log levels from general best practices (not DevArch-specific)

**Research date:** 2026-02-12
**Valid until:** 2026-03-15 (30 days, stable domain — slog API stable since Go 1.21, chi v5 stable)
