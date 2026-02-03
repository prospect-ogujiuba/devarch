---
phase: 01-foundation-guardrails
plan: 01
subsystem: runtime-abstraction
completed: 2026-02-03
duration: 3min

requires: []
provides:
  - identity-labels
  - name-validation
  - runtime-interface
  - stacks-schema

affects:
  - 01-02 (uses Runtime interface)
  - 02-* (all Phase 2 plans use labels/validation/schema)

tech-stack:
  added: []
  patterns:
    - label-based-identity
    - dns-safe-naming
    - interface-driven-runtime

key-files:
  created:
    - api/internal/container/labels.go
    - api/internal/container/labels_test.go
    - api/internal/container/validation.go
    - api/internal/container/validation_test.go
    - api/internal/container/types.go
    - api/migrations/013_stacks_instances.up.sql
    - api/migrations/013_stacks_instances.down.sql
  modified: []

decisions:
  - id: use-label-prefix
    what: All DevArch labels use "devarch." prefix
    why: Namespace isolation from user/other tool labels
    impact: Consistent label queries across runtime
  - id: dns-safe-validation
    what: Stack/instance names limited to DNS-safe format (63 chars, lowercase alphanumeric + hyphens)
    why: Container names become DNS names in networks, prevents collision issues
    impact: User-facing naming constraints, prescriptive error messages guide users
  - id: runtime-interface
    what: Define Runtime interface now for future podman/docker abstraction
    why: Plan 02 needs it, establishes contract early
    impact: Clean separation between runtime implementations

tags:
  - validation
  - labels
  - schema
  - runtime
---

# Phase 1 Plan 01: Labels, Validation, Types Summary

**One-liner:** Label constants, DNS-safe name validation with prescriptive errors, Runtime interface, and stack/instance DB schema.

## What Was Built

Created foundational primitives for stack/instance identity:

**Identity Labels** (`labels.go`):
- Constants: `LabelStackID`, `LabelInstanceID`, `LabelTemplateServiceID`, `LabelManagedBy`, `LabelVersion`
- Helpers: `BuildLabels()`, `ContainerName()`, `NetworkName()`, `IsDevArchManaged()`
- All containers tagged with `devarch.managed_by=devarch` for filtering

**Name Validation** (`validation.go`):
- DNS-safe regex: `^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`
- Reserved names blocked: default, devarch, system, none, all
- `ValidateName()` returns prescriptive errors with Slugify suggestions
- Example error: `"My App" is not a valid name: must be lowercase alphanumeric with hyphens — try: my-app`

**Runtime Interface** (`types.go`):
- Unified `Runtime` interface with 14 methods
- `ContainerStatus` and `ContainerMetrics` types for stack-aware code
- Network operations stubbed for Phase 4 implementation

**Database Schema** (migration 013):
- `stacks` table: id, name (unique), description, network_name, enabled
- `service_instances` table: links stack + instance_id to template service
- Foreign keys with CASCADE delete, indexed for performance

## Test Coverage

All code 100% tested:
- 5 label tests (BuildLabels, ContainerName, NetworkName, IsDevArchManaged)
- 3 validation test suites (ValidateName, prescriptive errors, Slugify)
- 14 test cases covering valid/invalid names
- 13 Slugify transformations tested

## Decisions Made

1. **Label prefix "devarch."** — Namespace isolation from other tools
2. **DNS-safe naming (63 char limit)** — Container names become DNS names, prevents network issues
3. **Prescriptive validation errors** — Every error includes Slugify suggestion for user guidance
4. **Runtime interface early** — Plan 02 needs it, establishes clean contract now

## Deviations from Plan

None - plan executed exactly as written.

## Next Phase Readiness

**Blockers:** None

**Enables:**
- Plan 01-02: Runtime adapter implementation (uses Runtime interface)
- Phase 2: All CRUD operations (uses labels, validation, schema)

**Database ready:** Migration 013 creates stacks/service_instances tables with proper constraints

## Artifacts

**Commits:**
- `2c78090`: feat(01-01): add identity labels, validation, and runtime types
- `a1e8185`: feat(01-01): add stacks and instances database schema

**Files:** 7 created (5 Go, 2 SQL)
**Tests:** 100% pass (14 test cases)
**Build:** Clean (`go build`, `go vet`)
