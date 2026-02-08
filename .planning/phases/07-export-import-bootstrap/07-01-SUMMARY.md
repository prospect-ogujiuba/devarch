---
phase: 07-export-import-bootstrap
plan: 01
subsystem: export
tags: [export, yaml, secrets, http-api]
completed: 2026-02-08
duration: 2min

dependency_graph:
  requires: [stacks, instances, effective-config]
  provides: [export-types, exporter, secret-redaction, export-endpoint]
  affects: [import-plan-02, lockfile-plan-03]

tech_stack:
  added: [gopkg.in/yaml.v3]
  patterns: [effective-config-merge, keyword-based-secret-detection]

key_files:
  created:
    - api/internal/export/types.go
    - api/internal/export/exporter.go
    - api/internal/export/secrets.go
    - api/internal/api/handlers/stack_export.go
  modified:
    - api/internal/api/routes.go

decisions:
  - Keyword-based secret detection (password, secret, key, token, api_key, apikey, auth, private, credential, passwd)
  - ${SECRET:VAR_NAME} placeholder syntax for redacted secrets
  - Version 1 format with reserved wires field (empty array)
  - Export includes all instances (enabled and disabled) with Enabled boolean
  - Identity labels included in export (devarch.stack_id, devarch.instance_id, devarch.template_service_id)
  - Effective config merge follows existing instance_effective.go patterns

metrics:
  tasks_completed: 2
  commits: 2
  files_created: 4
  files_modified: 1
  loc_added: ~730
---

# Phase 7 Plan 01: Export Domain Package Summary

Export foundation: types, exporter with effective config merge, and keyword-based secret redaction with ${SECRET:VAR_NAME} placeholders.

## Tasks Completed

| Task | Name                                           | Commit   | Files                                                                           |
| ---- | ---------------------------------------------- | -------- | ------------------------------------------------------------------------------- |
| 1    | Export types, exporter, and secret redaction   | 44892052 | types.go, exporter.go, secrets.go                                               |
| 2    | Export HTTP handler and route wiring           | ae6fd9e5 | stack_export.go, routes.go                                                      |

## Implementation Details

**Export Package Structure:**

- `types.go`: DevArchFile YAML schema with version 1, stack config, instance map, reserved wires field
- `exporter.go`: Exporter struct with db, Export method loads stack + instances with merged effective config
- `secrets.go`: IsSecretKey (case-insensitive keyword match), RedactSecrets (returns new map with placeholders)

**Effective Config Merge:**

Follows instance_effective.go patterns:
- Full replacement: ports, volumes, domains, healthcheck, dependencies, config files
- Key-based merge: environment variables, labels
- Identity labels auto-injected if not present (preserves user overrides)
- Image: template image unless instance overrides (format: name:tag)

**Secret Redaction:**

Keyword list: password, secret, key, token, api_key, apikey, auth, private, credential, passwd
- Case-insensitive substring match on env var keys
- Redacted values use ${SECRET:VAR_NAME} placeholder syntax
- Non-secrets exported with actual values

**HTTP Endpoint:**

GET /stacks/{name}/export returns YAML:
- Content-Type: application/x-yaml
- Content-Disposition: attachment; filename="{stackName}-devarch.yml"
- 404 on stack not found
- 500 on export failure

## Deviations from Plan

None - plan executed exactly as written.

## Verification Results

1. Export package compiles without errors
2. Server binary compiles with export handler
3. DevArchFile type exists in types.go
4. SECRET: placeholder syntax present in secrets.go
5. Export route registered at GET /stacks/{name}/export

## Self-Check

Verifying created files and commits.

**Files:**
- FOUND: api/internal/export/types.go
- FOUND: api/internal/export/exporter.go
- FOUND: api/internal/export/secrets.go
- FOUND: api/internal/api/handlers/stack_export.go

**Commits:**
- FOUND: 44892052 (Task 1)
- FOUND: ae6fd9e5 (Task 2)

## Self-Check: PASSED
