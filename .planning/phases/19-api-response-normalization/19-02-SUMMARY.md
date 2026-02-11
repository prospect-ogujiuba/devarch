---
phase: 19-api-response-normalization
plan: 02
subsystem: api
tags: [response-normalization, stack-handlers, error-handling]
dependency_graph:
  requires: [respond-package]
  provides: [normalized-stack-responses]
  affects: [stack-handlers]
tech_stack:
  added: []
  patterns: [envelope-pattern]
key_files:
  created: []
  modified:
    - api/internal/api/handlers/stack.go
    - api/internal/api/handlers/stack_lifecycle.go
    - api/internal/api/handlers/stack_apply.go
    - api/internal/api/handlers/stack_compose.go
    - api/internal/api/handlers/stack_plan.go
    - api/internal/api/handlers/stack_lock.go
    - api/internal/api/handlers/stack_wiring.go
    - api/internal/api/handlers/stack_export.go
    - api/internal/api/handlers/stack_import.go
decisions:
  - ExportStack YAML response remains exempt from envelope (retains application/x-yaml content-type)
  - Lock file generation returns JSON with file download headers (not wrapped in envelope)
  - Import error responses use respond.Error for payload_too_large with structured details
  - Removed custom writeImportError function in favor of standard respond helpers
metrics:
  duration: 652
  completed_at: 2026-02-11
  tasks_completed: 2
  files_modified: 9
---

# Phase 19 Plan 02: Stack Handler Response Migration Summary

**One-liner:** Migrated all 9 stack handler files (~170 response sites) to use respond package for consistent JSON envelopes

## What Was Built

### Migrated Stack Handler Files

**Task 1: Core Stack Operations (5 files)**
- `stack.go` — CRUD endpoints (Create, List, Get, Update, Delete, Enable, Disable, NetworkStatus, Clone, Rename, Restore, etc.)
- `stack_lifecycle.go` — Start, Stop, Restart endpoints
- `stack_apply.go` — Apply endpoint with plan token validation and advisory locking
- `stack_compose.go` — Compose generation endpoint
- `stack_plan.go` — Plan generation with wiring resolution and resource limits

**Task 2: Advanced Stack Operations (4 files)**
- `stack_lock.go` — Lock file generation, validation, and refresh
- `stack_wiring.go` — Service wiring CRUD (ListWires, CreateWire, DeleteWire, ResolveWires, CleanupWires)
- `stack_export.go` — Stack export to YAML (errors only — YAML response exempt from envelope)
- `stack_import.go` — Stack import from multipart YAML with streaming validation

### Migration Patterns Applied

**Error Response Migration:**
- `http.Error(w, fmt.Sprintf("stack %q not found", name), http.StatusNotFound)` → `respond.NotFound(w, r, "stack", name)`
- `http.Error(w, fmt.Sprintf("failed to X: %v", err), http.StatusInternalServerError)` → `respond.InternalError(w, r, err)`
- `http.Error(w, "...", http.StatusBadRequest)` → `respond.BadRequest(w, r, "...")`
- `http.Error(w, "...", http.StatusConflict)` → `respond.Conflict(w, r, "...")`

**Success Response Migration:**
- `w.Header().Set("Content-Type", "application/json")` + `json.NewEncoder(w).Encode(data)` → `respond.JSON(w, r, http.StatusOK, data)`
- `w.WriteHeader(http.StatusCreated)` + `json.NewEncoder(w).Encode(data)` → `respond.JSON(w, r, http.StatusCreated, data)`

**Special Cases:**
- `stack_import.go`: Replaced custom `writeImportError` function with respond helpers
- `stack_export.go`: YAML response body unchanged (only error responses migrated)
- `stack_lock.go`: Lock file download headers preserved (JSON content-type with attachment disposition)

## Deviations from Plan

None — plan executed exactly as written.

## Verification Results

```bash
cd /home/fhcadmin/projects/devarch/api
go build ./cmd/server                    # ✓ compiles cleanly
grep -c 'http\.Error' internal/api/handlers/stack*.go
# All 9 files: 0 (all migrated)
```

All verification passed.

## Task Breakdown

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Migrate stack.go, stack_lifecycle.go, stack_apply.go, stack_compose.go, stack_plan.go | 9084d1c | 5 files |
| 2 | Migrate stack_lock.go, stack_wiring.go, stack_export.go, stack_import.go | 4abfef7 | 4 files |

## Impact

**Migration Statistics:**
- 9 files migrated
- ~170 response sites converted
- 169 `http.Error` calls replaced
- ~40 `json.NewEncoder` calls replaced
- 1 custom error function removed

**Breaking Changes:**
All stack endpoints now return JSON envelopes:
- Success: `{"data": {...}}`
- Errors: `{"error": {"code": "...", "message": "...", "details": ...}}`

**Exemptions (as specified):**
- `POST /api/v1/stacks/{name}/export` — Returns `application/x-yaml` (YAML file download)
- Lock file endpoints preserve `Content-Disposition: attachment` headers

**Next Steps:**
Phase 19 Plan 03 will migrate service handler responses.

## Self-Check: PASSED

**Modified files verified:**
```bash
[ -f "api/internal/api/handlers/stack.go" ] && echo "FOUND: stack.go"
```
FOUND: stack.go

```bash
[ -f "api/internal/api/handlers/stack_wiring.go" ] && echo "FOUND: stack_wiring.go"
```
FOUND: stack_wiring.go

**Commits verified:**
```bash
git log --oneline --all | grep -q "9084d1c" && echo "FOUND: 9084d1c"
```
FOUND: 9084d1c

```bash
git log --oneline --all | grep -q "4abfef7" && echo "FOUND: 4abfef7"
```
FOUND: 4abfef7
