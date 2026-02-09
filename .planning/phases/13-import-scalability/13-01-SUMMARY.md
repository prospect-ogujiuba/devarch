---
phase: 13-import-scalability
plan: 01
subsystem: stack-import
tags: [streaming, multipart, performance, scalability]

dependency_graph:
  requires: [api-middleware, export-types]
  provides: [streaming-import, configurable-size-limits]
  affects: [stack-import-handler, route-middleware]

tech_stack:
  added: [mime/multipart, io.LimitReader]
  patterns: [streaming-multipart, route-level-middleware]

key_files:
  created: []
  modified:
    - api/internal/api/handlers/stack_import.go
    - api/internal/api/routes.go
    - api/internal/export/importer.go

decisions:
  - Streaming multipart via multipart.NewReader eliminates full-body buffering
  - io.LimitReader caps YAML part size, not multipart overhead
  - Route-level MaxBodySize(256MB) overrides global 10MB via r.With()
  - HTTP 413 response for payloads exceeding STACK_IMPORT_MAX_BYTES

metrics:
  duration: 2m22s
  completed: 2026-02-09
  tasks: 2
  commits: 2
---

# Phase 13 Plan 01: Streaming Stack Import Summary

**Streaming multipart import handler with configurable 256MB limit via STACK_IMPORT_MAX_BYTES env var**

## Tasks Completed

### Task 1: Replace ParseMultipartForm with streaming multipart reader
**Commit:** `a8db74b`
**Files:** `api/internal/api/handlers/stack_import.go`, `api/internal/export/importer.go`

Replaced full-buffering `ParseMultipartForm(10 << 20)` with streaming `multipart.NewReader`:
- Extract boundary from Content-Type header via `mime.ParseMediaType`
- Create streaming reader with `multipart.NewReader(r.Body, boundary)`
- Loop through parts to find form field "file"
- Apply `io.LimitReader(filePart, importMaxBytes)` to cap YAML content
- Decode directly from limited reader via `yaml.NewDecoder(limitedReader)`
- Return HTTP 413 when limit exceeded (detected via EOF errors)

Removed `r.ParseMultipartForm()` and `r.FormFile()` entirely — no intermediate buffering.

### Task 2: Add route-level size override for import endpoint
**Commit:** `4cc779a`
**Files:** `api/internal/api/routes.go`

Added dedicated 256MB size limit for stack import route:
- Read `STACK_IMPORT_MAX_BYTES` env var in `NewRouter` (default: `256 << 20`)
- Apply route-specific limit via `r.With(mw.MaxBodySize(importMaxBytes)).Post("/import", ...)`
- Global `r.Use(mw.MaxBodySize(10 << 20))` unchanged — all non-import routes keep 10MB cap
- Chi's innermost `MaxBytesReader` takes precedence, so route-level overrides global

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Fixed unused imports in export/importer.go**
- **Found during:** Task 1 verification (go build failure)
- **Issue:** `log` and `time` imports declared but never used, plus unused `start := time.Now()` variable
- **Fix:** Removed `log` and `time` imports, removed `start` variable declaration
- **Files modified:** `api/internal/export/importer.go`
- **Commit:** Included in `a8db74b` (Task 1 commit)

This was a blocking issue (Rule 3) — build wouldn't compile without fix.

## Verification Results

**Build status:** ✓ All packages compile cleanly
**Streaming verification:** ✓ `multipart.NewReader` used, zero `ParseMultipartForm` in codebase
**Size limit verification:** ✓ `STACK_IMPORT_MAX_BYTES` read in routes.go, applied to import route
**Global limit verification:** ✓ `r.Use(mw.MaxBodySize(10 << 20))` still present
**Part limiting:** ✓ `io.LimitReader` applied to file part

## Technical Details

**Before:** `ParseMultipartForm(10 << 20)` buffered entire multipart body in memory (temp files for large parts). Hard-coded 10MB limit applied to all endpoints including import.

**After:** Streaming multipart reader processes parts incrementally. Only the YAML content is memory-resident during decode. Import route has dedicated 256MB cap (configurable); other routes retain 10MB.

**Key improvement:** Large stack imports (e.g., 100+ services with config files) no longer risk OOM. The `io.LimitReader` cap prevents unbounded memory growth from malicious/malformed payloads.

## Impact

- **Stack import scalability:** Can now handle stack exports up to 256MB (configurable)
- **Memory safety:** Streaming read + size cap prevents OOM on large imports
- **API flexibility:** Route-specific size limits without affecting global defaults
- **Error clarity:** HTTP 413 response clearly indicates payload size issue

## Self-Check: PASSED

**Created files:** None (modifications only)

**Modified files verified:**
```
FOUND: api/internal/api/handlers/stack_import.go
FOUND: api/internal/api/routes.go
FOUND: api/internal/export/importer.go
```

**Commits verified:**
```
FOUND: a8db74b
FOUND: 4cc779a
```

**Build status:** Clean (no errors, no warnings)
