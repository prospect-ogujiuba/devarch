---
phase: 22-identity-service-naming-consolidation
plan: 01
subsystem: identity
tags: [naming, validation, labels, refactoring]
dependencies:
  requires: []
  provides: [identity-package, naming-service, validation-functions, label-constants]
  affects: []
tech-stack:
  added: [api/internal/identity]
  patterns: [pure-functions, transport-agnostic]
key-files:
  created:
    - api/internal/identity/service.go
    - api/internal/identity/validation.go
    - api/internal/identity/labels.go
  modified: []
decisions:
  - "Package-level functions only (no struct) — no DB dependency needed per research"
  - "Accept custom names as parameters — transport-agnostic per Phase 21 decision"
  - "New ValidateLabelKey function enforces devarch.* prefix reservation"
metrics:
  duration: 43
  completed: 2026-02-11T21:08:33Z
---

# Phase 22 Plan 01: Identity Package Creation Summary

**One-liner:** Pure function naming service with DNS validation and label constants replicated from container/ package

## Completed Tasks

| Task | Name                                                    | Commit  | Files                                                                                            |
| ---- | ------------------------------------------------------- | ------- | ------------------------------------------------------------------------------------------------ |
| 1    | Create identity package with naming service and validation | c0e3916 | api/internal/identity/service.go, api/internal/identity/validation.go, api/internal/identity/labels.go |

## Summary

Created `api/internal/identity/` package consolidating all naming logic, validation rules, and label constants currently scattered across `container/labels.go` and `container/validation.go`.

**service.go** (7 functions):
- `NetworkName(stackName)` — returns `devarch-{stack}-net`
- `ContainerName(stackName, instanceName)` — returns `devarch-{stack}-{instance}`
- `ResolveNetworkName(stackName, customNetworkName)` — custom override or computed default
- `ResolveContainerName(stackName, instanceName, customContainerName)` — custom override or computed default
- `ExportFileName(stackName)` — returns `{stack}-devarch.yml`
- `LockFileName(stackName)` — returns `{stack}-devarch.lock`
- `ExtractInstanceName(stackName, containerName)` — reverses container name to instance name

**validation.go** (5 functions):
- `ValidateName(name)` — DNS-safe, length, reserved names, slugify suggestion
- `ValidateContainerName(stackName, instanceName)` — combined length 127 limit
- `ValidateNetworkName(stackName)` — generated name length 63 limit
- `Slugify(input)` — converts to DNS-safe name
- `ValidateLabelKey(key)` — NEW: enforces devarch.* prefix reservation

**labels.go** (2 functions + 7 constants):
- `BuildLabels(stackID, instanceID, templateServiceID)` — produces label map
- `IsDevArchManaged(labels)` — checks managed_by label
- Constants: LabelPrefix, LabelStackID, LabelInstanceID, LabelTemplateServiceID, LabelManagedBy, LabelVersion, ManagedByValue

Package uses stdlib only (fmt, strings, regexp) — no external dependencies. Transport-agnostic per Phase 21 decision (no net/http imports).

## Deviations from Plan

None — plan executed exactly as written.

## Verification Results

All success criteria met:
- ✓ identity package compiles independently (`go build ./internal/identity/`)
- ✓ 7 naming functions in service.go
- ✓ 5 validation functions in validation.go
- ✓ 2 functions + 7 constants in labels.go
- ✓ All naming patterns (`devarch-*`) consolidated
- ✓ No external dependencies (stdlib only)
- ✓ No net/http imports

## Next Steps

Plan 02 will migrate all callers from container/ to identity/ package and remove duplicated functions from container/.

## Self-Check: PASSED

**Created files:**
- FOUND: api/internal/identity/service.go
- FOUND: api/internal/identity/validation.go
- FOUND: api/internal/identity/labels.go

**Commits:**
- FOUND: c0e3916
