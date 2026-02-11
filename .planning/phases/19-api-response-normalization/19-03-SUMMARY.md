---
phase: 19-api-response-normalization
plan: 03
subsystem: api-handlers
tags: [response-normalization, instance-handlers, json-envelopes]
dependency_graph:
  requires: [19-01]
  provides: [instance-response-envelopes]
  affects: [instance-crud, instance-lifecycle, instance-overrides]
tech_stack:
  added: []
  patterns: [respond-package-usage, envelope-wrapping]
key_files:
  created: []
  modified:
    - api/internal/api/handlers/instance.go
    - api/internal/api/handlers/instance_lifecycle.go
    - api/internal/api/handlers/instance_overrides.go
    - api/internal/api/handlers/instance_effective.go
    - api/internal/api/handlers/network.go
decisions: []
metrics:
  duration: 448
  completed_date: 2026-02-11
---

# Phase 19 Plan 03: Instance & Service Handler Response Normalization Summary

Migrate instance and service handler files to use respond package for consistent JSON envelopes.

## What Was Done

### Task 1: Instance Handler Migration (Complete)

Migrated 4 instance handler files from raw `http.Error` and `json.NewEncoder` to `respond.*` helpers:

**instance.go** (928 lines):
- Migrated 58 `http.Error` calls
- Migrated 7 `json.NewEncoder(w).Encode` calls
- Patterns: Create (201), List, Get, Update, Delete (204), Duplicate (201), Rename, DeletePreview
- Pagination headers preserved on List endpoint

**instance_lifecycle.go** (152 lines):
- Migrated 12 `http.Error` calls
- Migrated 3 `json.NewEncoder(w).Encode` calls
- Patterns: Stop, Start, Restart actions

**instance_overrides.go** (1009 lines):
- Migrated 107 `http.Error` calls
- Migrated 16 `json.NewEncoder(w).Encode` calls
- Patterns: UpdatePorts, UpdateVolumes, UpdateEnvVars, UpdateLabels, UpdateDomains, UpdateHealthcheck, UpdateDependencies, Config file CRUD, Resource limits, EnvFiles, Networks, ConfigMounts
- NoContent (204) responses on DELETE operations

**instance_effective.go** (919 lines):
- Migrated 26 `http.Error` calls
- Migrated 1 `json.NewEncoder(w).Encode` call
- Pattern: EffectiveConfig endpoint (template + override merge)

**network.go** (bugfix):
- Fixed incorrect `respond.InternalError` signatures (removed extra status code parameter)
- Added missing `encoding/json` import

Total migrated:
- 203 `http.Error` → `respond.*` calls
- 27 `json.NewEncoder` → `respond.JSON` calls

### Task 2: Service Handler Migration (Incomplete)

service.go (2088 lines, 157 `http.Error`, 35 `json.NewEncoder`) remains unmigrated due to:
- File complexity (largest handler, 192 response sites)
- Automated sed migrations encountered system file reset issues
- Requires manual careful migration to preserve:
  - text/plain Logs endpoint (exempt from enveloping)
  - Pagination headers on List
  - StatusCreated (201) on Create
  - Service lifecycle action responses

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Fixed network.go respond.InternalError signature**
- **Found during:** Task 1 build verification
- **Issue:** network.go had incorrect `respond.InternalError(w, r, err, http.StatusInternalServerError)` calls with extra status code parameter
- **Fix:** Removed extra status parameter, added missing `encoding/json` import
- **Files modified:** `api/internal/api/handlers/network.go`
- **Commit:** 174d6ba

## Technical Notes

**Response Pattern Migration:**
- `http.Error(w, msg, http.StatusBadRequest)` → `respond.BadRequest(w, r, msg)`
- `http.Error(w, msg, http.StatusNotFound)` → `respond.NotFound(w, r, resource, identifier)`
- `http.Error(w, fmt.Sprintf("...: %v", err), http.StatusInternalServerError)` → `respond.InternalError(w, r, fmt.Errorf("...: %w", err))`
- `http.Error(w, msg, http.StatusConflict)` → `respond.Conflict(w, r, msg)`
- `json.NewEncoder(w).Encode(data)` → `respond.JSON(w, r, http.StatusOK, data)`
- `w.WriteHeader(http.StatusNoContent)` → `respond.NoContent(w, r)`

**Import changes:**
- Added: `"github.com/priz/devarch-api/internal/api/respond"`
- Kept: `"encoding/json"` (still used for `json.NewDecoder(r.Body).Decode`)

**Success Envelope** (from 19-01):
```go
{"data": <payload>}
```

**Error Envelope** (from 19-01):
```go
{"error": {"code": "...", "message": "...", "details": null}}
```

## Self-Check: PASSED

**Created files exist:** N/A (no new files)

**Modified files exist:**
```bash
[ -f "/home/fhcadmin/projects/devarch/api/internal/api/handlers/instance.go" ] && echo "FOUND"
[ -f "/home/fhcadmin/projects/devarch/api/internal/api/handlers/instance_lifecycle.go" ] && echo "FOUND"
[ -f "/home/fhcadmin/projects/devarch/api/internal/api/handlers/instance_overrides.go" ] && echo "FOUND"
[ -f "/home/fhcadmin/projects/devarch/api/internal/api/handlers/instance_effective.go" ] && echo "FOUND"
[ -f "/home/fhcadmin/projects/devarch/api/internal/api/handlers/network.go" ] && echo "FOUND"
```
All FOUND.

**Commits exist:**
```bash
git log --oneline | grep -q "174d6ba" && echo "FOUND: 174d6ba"
```
FOUND: 174d6ba

**Build verification:**
```bash
cd /home/fhcadmin/projects/devarch/api && go build ./cmd/server
```
Build succeeds.

## Outstanding Work

**service.go migration incomplete:**
- Requires continuation in separate task/plan
- 157 `http.Error` calls remain
- 35 `json.NewEncoder` calls remain
- text/plain Logs endpoint must be preserved (exempt from enveloping)
- Complex file requires careful manual migration
