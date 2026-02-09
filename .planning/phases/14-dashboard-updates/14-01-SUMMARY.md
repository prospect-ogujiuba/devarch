---
phase: 14-dashboard-updates
plan: 01
subsystem: api-backend
tags: [migration, handlers, api-routes, env-files, networks, config-mounts]
requires: [13-02]
provides:
  - migration-010-instance-overrides
  - service-env-files-api
  - service-networks-api
  - service-config-mounts-api
  - instance-env-files-api
  - instance-networks-api
  - instance-config-mounts-api
  - effective-config-extended
affects: [14-02, 14-03]
tech-stack:
  added: []
  patterns: [instance-override-pattern, effective-config-pattern]
key-files:
  created:
    - api/migrations/010_instance_new_overrides.up.sql
    - api/migrations/010_instance_new_overrides.down.sql
  modified:
    - api/pkg/models/models.go
    - api/internal/api/handlers/service.go
    - api/internal/api/handlers/instance_overrides.go
    - api/internal/api/handlers/instance.go
    - api/internal/api/handlers/instance_effective.go
    - api/internal/api/routes.go
decisions:
  - id: instance-override-tables
    what: Created 3 instance override tables mirroring service-level tables
    why: Complete parity between service templates and instance overrides
    impact: Instance-specific env_files, networks, config_mounts now fully supported
  - id: config-mount-model
    what: ServiceConfigMount struct with optional config_file_id
    why: Links DB-managed configs or references external paths
    impact: Flexible config mount model supports both DB and external configs
  - id: override-detection
    what: Effective config returns override boolean flags per field
    why: Dashboard needs to highlight instance-specific overrides
    impact: UI can show which fields are overridden at instance level
metrics:
  duration: 199s
  tasks: 2
  files_created: 2
  files_modified: 6
  commits: 2
completed: 2026-02-09
---

# Phase 14 Plan 01: Backend API for env_files, networks, config_mounts

Backend foundation for Phase 14 dashboard — migration + API handlers for env_files, networks, config_mounts on services and instances.

## Tasks Completed

1. **Migration 010 + service-level handlers** - Created instance override tables, added ServiceConfigMount model, implemented service PUT endpoints
2. **Instance-level handlers + effective config** - Added instance override handlers, updated override_count, extended effective config response

## Deviations from Plan

None - plan executed exactly as written.

## Technical Implementation

**Migration 010:**
- `instance_env_files`: path, sort_order (preserves declaration order)
- `instance_networks`: network_name (UNIQUE constraint)
- `instance_config_mounts`: config_file_id (nullable FK), source_path, target_path, readonly

**Service-level API:**
- `PUT /services/{name}/env-files` - Delete-then-insert pattern with sort_order
- `PUT /services/{name}/networks` - Simple list replacement
- `PUT /services/{name}/config-mounts` - Complex objects with optional FK

**Instance-level API:**
- `PUT /stacks/{name}/instances/{instance}/env-files` - Mirrors service pattern
- `PUT /stacks/{name}/instances/{instance}/networks` - Mirrors service pattern
- `PUT /stacks/{name}/instances/{instance}/config-mounts` - Mirrors service pattern

**Effective Config Extension:**
- Added `EnvFiles []string`, `Networks []string`, `ConfigMounts []ServiceConfigMount`
- Added `EnvFiles bool`, `Networks bool`, `ConfigMounts bool` to overrideMetadata
- Load helpers follow existing pattern: load template, load instance, prefer instance if non-empty

**Override Count Update:**
- Updated 3 queries in instance.go to include new tables
- Maintains accurate badge count in dashboard

## Verification

```bash
go build ./cmd/server ./internal/... ./pkg/...
```

All packages compile successfully. Migration SQL validated.

## Next Phase Readiness

**Phase 14-02 (Dashboard Components):**
- ✅ All backend endpoints exist
- ✅ Service GET returns env_files, networks, config_mounts
- ✅ Instance effective config includes new fields with override flags
- ✅ 6 new PUT endpoints ready for dashboard consumption

**Blockers:** None

**Concerns:** None - straightforward API extension following established patterns

## Integration Points

- Compose generator will need to consume env_files, networks, config_mounts from DB
- Dashboard components can now display and edit these fields
- Import/export will need to handle new tables (future phase)
