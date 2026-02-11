---
phase: 22-identity-service-naming-consolidation
plan: 02
subsystem: identity
tags: [naming, validation, labels, refactoring, consolidation]
dependencies:
  requires: [22-01]
  provides: [identity-migration-complete, container-cleanup]
  affects: [orchestration, compose, export, lock, project, wiring, handlers]
tech-stack:
  added: []
  patterns: [single-source-truth, package-consolidation]
key-files:
  created: []
  modified:
    - internal/orchestration/service.go
    - internal/compose/stack.go
    - internal/export/exporter.go
    - internal/export/importer.go
    - internal/lock/generator.go
    - internal/lock/validator.go
    - internal/project/controller.go
    - internal/wiring/env_injector.go
    - internal/api/handlers/stack.go
    - internal/api/handlers/stack_lifecycle.go
    - internal/api/handlers/stack_compose.go
    - internal/api/handlers/stack_import.go
    - internal/api/handlers/instance.go
    - internal/api/handlers/instance_lifecycle.go
    - internal/api/handlers/instance_effective.go
    - internal/api/handlers/instance_overrides.go
    - internal/api/handlers/project.go
    - internal/api/handlers/network.go
  deleted:
    - internal/container/labels.go
    - internal/container/validation.go
decisions:
  - "Deleted container/labels.go and container/validation.go entirely — identity package is now single source"
  - "Removed extractInstanceName local function from lock/generator.go — use identity.ExtractInstanceName"
  - "container package now contains only runtime concerns (client.go)"
metrics:
  duration: 517
  completed: 2026-02-11T21:19:13Z
---

# Phase 22 Plan 02: Caller Migration and Container Cleanup Summary

**One-liner:** Migrated all naming/validation/label calls to identity package, deleted container naming files, achieved single source of truth

## Completed Tasks

| Task | Name                                                  | Commit  | Files                                                                                                 |
| ---- | ----------------------------------------------------- | ------- | ----------------------------------------------------------------------------------------------------- |
| 1    | Migrate internal packages to identity                | 7cb9e0d | orchestration, compose, export, lock, project, wiring (8 files)                                       |
| 2    | Migrate handlers and remove container naming aliases | d296e49 | 10 handlers + deleted container/labels.go and container/validation.go (12 files)                      |

## Summary

Completed identity package consolidation by migrating all callers from `container` naming/validation/labels to `identity` package equivalents. Eliminated every ad-hoc `fmt.Sprintf("devarch-...")` call outside identity. Deleted `container/labels.go` and `container/validation.go` entirely.

**Task 1: Internal packages** (7 files + validator.go fix):
- `orchestration/service.go` — identity.NetworkName, identity.ContainerName
- `compose/stack.go` — identity.ContainerName, identity.BuildLabels (removed unused container import)
- `export/exporter.go` — identity.NetworkName, identity.BuildLabels (removed container import)
- `export/importer.go` — identity.NetworkName, identity.ContainerName, identity.LabelPrefix
- `lock/generator.go` — identity.NetworkName, identity.LabelStackID, identity.ExtractInstanceName (deleted local extractInstanceName)
- `lock/validator.go` — added identity import for ExtractInstanceName
- `project/controller.go` — identity.NetworkName
- `wiring/env_injector.go` — identity.ContainerName (removed container import)

**Task 2: Handlers and cleanup** (10 handlers + 2 deletions):
- `handlers/stack.go` — identity.NetworkName, identity.ValidateName, identity.LabelStackID, identity.ContainerName
- `handlers/stack_lifecycle.go` — identity.NetworkName (removed fmt import)
- `handlers/stack_compose.go` — identity.NetworkName (removed fmt import)
- `handlers/stack_import.go` — identity.ValidateName (removed container import)
- `handlers/instance.go` — identity.ValidateName, identity.ValidateContainerName, identity.ContainerName
- `handlers/instance_lifecycle.go` — identity.NetworkName
- `handlers/instance_effective.go` — identity.BuildLabels (removed container import)
- `handlers/instance_overrides.go` — identity.LabelPrefix (improved error message)
- `handlers/project.go` — identity.ValidateName, identity.NetworkName, identity.LabelStackID
- `handlers/network.go` — identity.IsDevArchManaged, identity.LabelManagedBy, identity.ManagedByValue
- **Deleted:** `internal/container/labels.go` (67 lines)
- **Deleted:** `internal/container/validation.go` (102 lines)

**Migration pattern:** Add identity import → replace container.* calls with identity.* → remove unused container/fmt imports → verify build.

**Result:** Zero scattered naming calls. `container` package reduced to runtime concerns only (client.go remains). Identity package is single source of truth for all naming, validation, and label operations.

## Deviations from Plan

**[Rule 1 - Bug] Fixed missing identity import in lock/validator.go**
- **Found during:** Task 1 build verification
- **Issue:** validator.go called extractInstanceName which was deleted, needed identity.ExtractInstanceName but missing import
- **Fix:** Added `"github.com/priz/devarch-api/internal/identity"` import to validator.go
- **Files modified:** internal/lock/validator.go
- **Commit:** 7cb9e0d (included in Task 1)

**[Rule 2 - Critical] Removed unused imports during migration**
- **Found during:** Task 1 and Task 2 build verification
- **Issue:** Migrating calls left orphaned imports (container, fmt) causing build errors
- **Fix:** Removed unused imports from compose/stack.go, wiring/env_injector.go, stack_lifecycle.go, stack_compose.go, instance_effective.go, stack_import.go
- **Files modified:** 6 files
- **Commits:** 7cb9e0d (Task 1), d296e49 (Task 2)

## Verification Results

All success criteria met:

- ✓ `go build ./...` passes with zero errors
- ✓ Zero `fmt.Sprintf("devarch-...")` calls outside identity package
- ✓ Zero `container.ValidateName`, `container.ValidateContainerName`, `container.NetworkName`, `container.ContainerName`, `container.BuildLabels` calls
- ✓ Zero `container.LabelStackID`, `container.LabelManagedBy`, `container.ManagedByValue`, `container.LabelPrefix` references
- ✓ container/labels.go and container/validation.go deleted
- ✓ container/client.go unchanged (runtime concerns remain)
- ✓ 15 identity.* references in handlers/stack.go (confirms migration)

**Grep verifications:**
```bash
# No ad-hoc naming outside identity
grep -rn 'fmt.Sprintf("devarch-' internal/ | grep -v identity/ → 0 matches

# No container naming calls
grep -rn 'container\.ValidateName|container\.ContainerName|...' internal/ → 0 matches

# Files deleted
ls internal/container/validation.go internal/container/labels.go → not found

# Container runtime code preserved
ls internal/container/client.go → exists
```

## Next Steps

Phase 22 complete. Identity package consolidation achieved. All naming operations now flow through single source of truth.

## Self-Check: PASSED

**Modified files (Task 1):**
- FOUND: internal/orchestration/service.go
- FOUND: internal/compose/stack.go
- FOUND: internal/export/exporter.go
- FOUND: internal/export/importer.go
- FOUND: internal/lock/generator.go
- FOUND: internal/lock/validator.go
- FOUND: internal/project/controller.go
- FOUND: internal/wiring/env_injector.go

**Modified files (Task 2):**
- FOUND: internal/api/handlers/stack.go
- FOUND: internal/api/handlers/stack_lifecycle.go
- FOUND: internal/api/handlers/stack_compose.go
- FOUND: internal/api/handlers/stack_import.go
- FOUND: internal/api/handlers/instance.go
- FOUND: internal/api/handlers/instance_lifecycle.go
- FOUND: internal/api/handlers/instance_effective.go
- FOUND: internal/api/handlers/instance_overrides.go
- FOUND: internal/api/handlers/project.go
- FOUND: internal/api/handlers/network.go

**Deleted files:**
- CONFIRMED DELETED: internal/container/labels.go
- CONFIRMED DELETED: internal/container/validation.go

**Commits:**
- FOUND: 7cb9e0d
- FOUND: d296e49
