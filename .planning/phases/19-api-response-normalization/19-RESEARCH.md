# Phase 19: API Response Normalization - Research

**Researched:** 2026-02-11
**Domain:** Go HTTP API response envelopes and error standardization
**Confidence:** HIGH

## Summary

Phase 19 standardizes API responses across 24+ handler files using consistent JSON envelopes. Current codebase has ~800 instances of `http.Error`, `json.NewEncoder`, and `WriteHeader` across handlers, with no shared response helpers. Each handler implements its own error handling, creating inconsistent client experience.

Standard approach: create shared responder functions that wrap success data in `{"data": ...}` and errors in `{"error": {"code", "message", "details"}}`. The go-chi/render package (already compatible with chi v5) provides battle-tested patterns for this, though a minimal stdlib-only solution is also viable given the project's minimalist dependency philosophy.

**Primary recommendation:** Create internal responder package with helpers for success/error envelopes. Avoid adding go-chi/render as dependency; use stdlib json encoding with custom envelope types. Implement middleware-based panic recovery that enforces envelope consistency.

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| encoding/json | stdlib | JSON marshaling | Project uses stdlib throughout (no external serialization deps) |
| net/http | stdlib | HTTP responses | Foundation of all handler responses |
| github.com/go-chi/chi/v5 | v5.1.0 (existing) | Router | Already in go.mod, provides context and middleware |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| github.com/go-chi/render | v1.0.3+ | Response helpers | **NOT RECOMMENDED** — adds dependency, project favors stdlib |
| github.com/moogar0880/problems | v1.1.1 | RFC 7807 | **NOT RECOMMENDED** — overkill for internal API |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Custom responder | go-chi/render | Render adds convenience but unnecessary dependency; project has no third-party serializers |
| Simple envelope | RFC 7807 (problems pkg) | RFC 7807 is formal standard but verbose; DevArch is single-app API, not public REST |
| Inline errors | Error interface types | Type assertion pattern adds complexity without multiservice benefits |

**Installation:**
```bash
# No new dependencies required — stdlib only
```

## Architecture Patterns

### Recommended Project Structure
```
api/internal/
├── api/
│   ├── handlers/          # HTTP handlers (unchanged)
│   ├── middleware/        # Existing middleware
│   └── respond/           # NEW: response envelope helpers
│       ├── respond.go     # Success/error responders
│       └── types.go       # Envelope structs
```

### Pattern 1: Envelope Response Types

**What:** Standardized wrapper structs for success and error responses

**When to use:** All handler responses except streaming (logs, WebSocket)

**Example:**
```go
// api/internal/api/respond/types.go

// SuccessEnvelope wraps successful responses
type SuccessEnvelope struct {
    Data interface{} `json:"data"`
}

// ErrorEnvelope wraps error responses
type ErrorEnvelope struct {
    Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
    Code    string      `json:"code"`              // machine-readable: "stack_not_found"
    Message string      `json:"message"`           // human-readable: "Stack 'web' not found"
    Details interface{} `json:"details,omitempty"` // optional context
}
```

### Pattern 2: Responder Functions

**What:** Shared functions that set headers, status codes, and encode envelopes

**When to use:** Replace all direct `json.NewEncoder(w).Encode()` and `http.Error()` calls

**Example:**
```go
// api/internal/api/respond/respond.go

func JSON(w http.ResponseWriter, r *http.Request, statusCode int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)

    envelope := SuccessEnvelope{Data: data}
    if err := json.NewEncoder(w).Encode(envelope); err != nil {
        // Logging only — headers already written
        log.Printf("failed to encode response: %v", err)
    }
}

func Error(w http.ResponseWriter, r *http.Request, statusCode int, code, message string, details interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)

    envelope := ErrorEnvelope{
        Error: ErrorDetail{
            Code:    code,
            Message: message,
            Details: details,
        },
    }
    if err := json.NewEncoder(w).Encode(envelope); err != nil {
        log.Printf("failed to encode error response: %v", err)
    }
}
```

### Pattern 3: Convenience Helpers

**What:** Domain-specific error constructors for common HTTP statuses

**When to use:** Reduce boilerplate in handlers

**Example:**
```go
func BadRequest(w http.ResponseWriter, r *http.Request, message string) {
    Error(w, r, http.StatusBadRequest, "bad_request", message, nil)
}

func NotFound(w http.ResponseWriter, r *http.Request, resource, identifier string) {
    Error(w, r, http.StatusNotFound, "not_found",
        fmt.Sprintf("%s '%s' not found", resource, identifier), nil)
}

func InternalError(w http.ResponseWriter, r *http.Request, err error) {
    // Log full error internally, return sanitized message
    log.Printf("internal error: %v", err)
    Error(w, r, http.StatusInternalServerError, "internal_error",
        "An internal error occurred", nil)
}

func Conflict(w http.ResponseWriter, r *http.Request, message string) {
    Error(w, r, http.StatusConflict, "conflict", message, nil)
}
```

### Pattern 4: Middleware for Panic Recovery

**What:** Catch handler panics and return error envelope

**When to use:** Apply globally to all API routes

**Example:**
```go
// api/internal/api/middleware/middleware.go (extend existing)

func RecoverEnvelope(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                log.Printf("panic recovered: %v", err)
                respond.InternalError(w, r, fmt.Errorf("panic: %v", err))
            }
        }()
        next.ServeHTTP(w, r)
    })
}
```

### Pattern 5: Header Preservation

**What:** Some endpoints (e.g., list operations) set custom headers like `X-Total-Count`

**When to use:** Handlers that need pagination headers alongside envelopes

**Example:**
```go
// Service list handler pattern
func (h *ServiceHandler) List(w http.ResponseWriter, r *http.Request) {
    // ... query logic ...

    // Set custom headers BEFORE respond.JSON
    w.Header().Set("X-Total-Count", strconv.Itoa(total))
    w.Header().Set("X-Page", strconv.Itoa(page))
    w.Header().Set("X-Per-Page", strconv.Itoa(limit))
    w.Header().Set("X-Total-Pages", strconv.Itoa(totalPages))

    // respond.JSON will preserve existing headers
    respond.JSON(w, r, http.StatusOK, services)
}
```

### Anti-Patterns to Avoid

- **Mixing envelope and non-envelope responses** — Dashboard clients will expect consistent format. Exempt only non-JSON endpoints (logs streaming, WebSocket upgrade).

- **Exposing internal errors in message field** — Use sanitized messages for clients, log full errors server-side. Example: `fmt.Errorf("failed to query: %w", err)` logs full trace, client sees "An internal error occurred".

- **Using error types/interfaces** — Pattern like `type APIError interface { StatusCode() int }` adds abstraction without value in single-service API.

- **Status code in envelope body** — HTTP status is in headers; don't duplicate it in JSON body (common RFC 7807 approach). Envelope has code (string identifier), not status (int).

- **Wrapping non-errors** — 200/201 responses should not use error envelope shape.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Response encoding logic | Custom per-handler | Shared respond package | 800+ response sites — centralizing prevents drift |
| Error code generation | Ad-hoc strings in handlers | Pre-defined constants | Clients can rely on stable codes for logic |
| Panic handling | Per-handler recovery | Middleware recovery | Ensures all panics return envelope, not plain 500 |
| Content-Type logic | Manual header setting | respond.JSON (sets automatically) | Easy to forget, hard to audit |

**Key insight:** Handlers naturally diverge without shared abstractions. The respond package acts as enforcement point — all responses flow through 2-3 functions instead of 800+ ad-hoc sites.

## Common Pitfalls

### Pitfall 1: Headers Already Sent

**What goes wrong:** Calling `respond.Error()` after `respond.JSON()` or after manual `WriteHeader()` results in no-op or partial writes. HTTP headers cannot be modified after first byte written.

**Why it happens:** Error handling after success path has started response.

**How to avoid:**
- Guard all response calls with early returns
- Never call multiple respond functions in same handler execution
- Middleware recovery must check if headers already sent (noop if true)

**Warning signs:** Logs showing "failed to encode response" but client receives 200 OK with empty body.

### Pitfall 2: Forgetting to Wrap Existing Responses

**What goes wrong:** Incremental migration leaves some endpoints returning bare data, others wrapped in `{"data": ...}`. Dashboard must handle both formats.

**Why it happens:** Large codebase (24 handler files), easy to miss endpoints.

**How to avoid:**
- Audit all `json.NewEncoder(w).Encode` and `http.Error` calls
- Use grep/ripgrep to find all response sites: `rg 'http\.Error|json\.NewEncoder|WriteHeader'`
- Plan endpoint migration in batches (e.g., stacks handlers → instance handlers → service handlers)
- Consider adding interim middleware that logs non-enveloped responses for tracking

**Warning signs:** Dashboard making separate code paths for "old" vs "new" endpoints.

### Pitfall 3: Error Details Field Exposing Internals

**What goes wrong:** Passing raw `err.Error()` string to `details` field leaks SQL errors, file paths, internal package names.

**Why it happens:** Convenience — `respond.Error(w, r, 500, "error", err.Error(), nil)` is easy but unsafe.

**How to avoid:**
- Details field for structured data only (e.g., validation failures: `{"field": "name", "reason": "too long"}`)
- Internal errors → sanitized message, full error logged server-side
- Create helper `InternalError(w, r, err)` that logs err but returns generic message

**Warning signs:** Client logs showing database schema names, file paths, Go stack traces.

### Pitfall 4: HTTP Status Code Confusion

**What goes wrong:** Using 200 OK for error responses with `{"error": ...}` envelope.

**Why it happens:** Forgetting status code parameter, defaulting to 200.

**How to avoid:**
- Status code is first-class in respond functions: `respond.Error(w, r, statusCode, ...)`
- Helper functions encode status: `respond.NotFound()` always uses 404
- Unit tests verify status codes match envelope type

**Warning signs:** Client HTTP clients (axios, fetch) reporting success on errors.

### Pitfall 5: Streaming Endpoints Breaking Pattern

**What goes wrong:** Applying envelope to log streaming endpoint results in `{"data": "log line 1\nlog line 2\n..."}` — breaks real-time UX.

**Why it happens:** Over-application of envelope pattern.

**How to avoid:**
- Explicitly exempt streaming endpoints: logs, metrics, WebSocket upgrade, compose YAML download
- Document exempt endpoints in respond package godoc
- Streaming = response is chunked/SSE/WebSocket, not single JSON object

**Warning signs:** Log viewer receiving entire log as single JSON payload instead of streaming lines.

## Code Examples

Verified patterns from research:

### Handler Migration: Before/After

**Before (current):**
```go
func (h *StackHandler) Get(w http.ResponseWriter, r *http.Request) {
    name := chi.URLParam(r, "name")

    var stack stackResponse
    err := h.db.QueryRow(`...`).Scan(&stack.ID, &stack.Name, ...)

    if err == sql.ErrNoRows {
        http.Error(w, fmt.Sprintf("stack %q not found", name), http.StatusNotFound)
        return
    }
    if err != nil {
        http.Error(w, fmt.Sprintf("failed to get stack: %v", err), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(stack)
}
```

**After (with respond package):**
```go
func (h *StackHandler) Get(w http.ResponseWriter, r *http.Request) {
    name := chi.URLParam(r, "name")

    var stack stackResponse
    err := h.db.QueryRow(`...`).Scan(&stack.ID, &stack.Name, ...)

    if err == sql.ErrNoRows {
        respond.NotFound(w, r, "stack", name)
        return
    }
    if err != nil {
        respond.InternalError(w, r, err)
        return
    }

    respond.JSON(w, r, http.StatusOK, stack)
}
```

### Error Response Structure (Client View)

```json
{
  "error": {
    "code": "stack_not_found",
    "message": "Stack 'web' not found"
  }
}
```

With optional details:
```json
{
  "error": {
    "code": "validation_failed",
    "message": "Invalid stack name",
    "details": {
      "field": "name",
      "reason": "must match pattern ^[a-z0-9-]+$"
    }
  }
}
```

### Success Response Structure (Client View)

Single resource:
```json
{
  "data": {
    "id": 1,
    "name": "web",
    "enabled": true,
    "instance_count": 3
  }
}
```

Collection (preserves existing pagination headers):
```json
{
  "data": [
    {"id": 1, "name": "postgres"},
    {"id": 2, "name": "redis"}
  ]
}
```

Headers: `X-Total-Count: 42`, `X-Page: 1`, `X-Per-Page: 20`, `X-Total-Pages: 3`

### Validation Error Example

```go
// Handler with validation errors
func (h *StackHandler) Create(w http.ResponseWriter, r *http.Request) {
    var req createStackRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respond.BadRequest(w, r, "Invalid JSON body")
        return
    }

    if err := container.ValidateName(req.Name); err != nil {
        respond.Error(w, r, http.StatusBadRequest, "validation_failed",
            err.Error(), map[string]string{"field": "name"})
        return
    }

    // ... success path ...
    respond.JSON(w, r, http.StatusCreated, stack)
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Plain-text http.Error | JSON error envelopes | 2020-2024 | Structured errors enable client-side logic |
| Bare JSON responses | `{"data": ...}` wrapper | 2022-2025 | Consistent parsing, easier to add metadata |
| Handler panics → 500 text | Panic recovery middleware | 2021+ | Graceful degradation, no broken JSON |
| RFC 7807 full spec | Simplified error envelope | 2023+ | Reduced complexity for single-app APIs |

**Deprecated/outdated:**
- **JSend format** (`status: "success"/"fail"/"error"`): Redundant with HTTP status codes, verbose
- **go-chi/render Renderer interface**: Adds boilerplate (Render method on every type) without clear benefit for simple CRUD APIs
- **Global error type hierarchies**: Over-engineering for non-microservice architectures

## Open Questions

1. **Should pagination metadata move inside envelope?**
   - What we know: Current pattern uses HTTP headers (`X-Total-Count`, etc.)
   - What's unclear: Should envelope become `{"data": [...], "meta": {"total": 42}}`?
   - Recommendation: Keep headers initially — less client disruption. Consider meta field in future phase if needed.

2. **How to handle WebSocket status messages?**
   - What we know: WebSocket sends real-time JSON messages (not HTTP responses)
   - What's unclear: Should WS messages use same envelope format?
   - Recommendation: Out of scope for Phase 19. WS messages are push notifications, not request/response. Revisit in Phase 25 if dashboard needs consistency.

3. **Should error codes be constants or strings?**
   - What we know: Examples show string codes (`"stack_not_found"`)
   - What's unclear: Should we define const ErrorCodeStackNotFound = "stack_not_found"?
   - Recommendation: Start with inline strings, extract constants if pattern emerges (e.g., reused across multiple handlers).

## Sources

### Primary (HIGH confidence)
- [go-chi/render package documentation](https://pkg.go.dev/github.com/go-chi/render) - Official chi render package API
- [chi REST example](https://github.com/go-chi/chi/blob/master/_examples/rest/main.go) - Canonical error envelope pattern
- DevArch codebase - 24 handler files, 800+ response sites analyzed

### Secondary (MEDIUM confidence)
- [API responses in Go (Mat Ryer)](https://medium.com/@matryer/api-responses-in-go-1ef8f7b74997) - Industry patterns from experienced Go author
- [Crafting Custom Errors in Go APIs](https://leapcell.io/blog/crafting-custom-errors-and-http-status-codes-in-go-apis) - Error response structure patterns
- [go-chi/chi GitHub](https://github.com/go-chi/chi) - Router documentation and middleware patterns

### Tertiary (LOW confidence)
- [RFC 7807 implementations](https://pkg.go.dev/github.com/moogar0880/problems) - Formal error standard (deemed too heavy for DevArch)
- [REST API Best Practices](https://medium.com/@sukhadamorgaonkar28/rest-api-best-practices-239f4d0bd6f5) - General guidance (not Go-specific)

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - stdlib-only approach verified against existing codebase patterns
- Architecture: HIGH - Patterns sourced from chi maintainers and production codebases
- Pitfalls: HIGH - Derived from codebase analysis (800+ response sites) and common handler patterns

**Research date:** 2026-02-11
**Valid until:** 2026-03-15 (30 days for stable stdlib patterns; Go 1.22 dependencies unchanged)
