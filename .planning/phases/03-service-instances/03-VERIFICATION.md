---
phase: 03-service-instances
verified: 2026-02-06T13:45:00Z
status: passed
score: 6/6 must-haves verified
---

# Phase 3: Service Instances Verification Report

**Phase Goal:** Users can create service instances from templates with full copy-on-write overrides
**Verified:** 2026-02-06T13:45:00Z
**Status:** passed
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | User can add service instance to stack from template catalog via dashboard | ✓ VERIFIED | `add-instance-dialog.tsx` renders template catalog with search, calls `useCreateInstance` mutation, validates instance names |
| 2 | User can override instance ports, volumes, env vars, dependencies, labels, domains, healthchecks, config files | ✓ VERIFIED | 8 override editor components exist (`override-*.tsx`), each with edit UI calling respective `useUpdateInstance*` mutations |
| 3 | Effective config API returns merged template + overrides (overrides win) | ✓ VERIFIED | `instance_effective.go` EffectiveConfig handler merges template+instance with explicit override precedence, returns `overrides_applied` metadata |
| 4 | User can view effective config before applying (no surprises) | ✓ VERIFIED | `effective-config-tab.tsx` displays merged config with override indicators (blue borders), YAML/JSON export via copy button |
| 5 | User can edit instance overrides after creation | ✓ VERIFIED | Instance detail page has tabbed editors for each resource type, all wired to PUT endpoints, save buttons trigger mutations |
| 6 | User can remove instance from stack | ✓ VERIFIED | `instance-actions.tsx` DeleteInstanceDialog fetches preview, calls `useDeleteInstance` mutation, soft-deletes via deleted_at column |

**Score:** 6/6 truths verified

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `api/migrations/014_instance_overrides.up.sql` | 7 override tables (ports, volumes, env_vars, labels, domains, healthchecks, config_files) | ✓ VERIFIED | 87 lines, all 7 tables with FK to service_instances, ON DELETE CASCADE, indexes on instance_id, soft-delete support |
| `api/migrations/017_instance_dependencies.up.sql` | instance_dependencies table | ✓ VERIFIED | 10 lines, FK to service_instances, UNIQUE(instance_id, depends_on) |
| `api/internal/api/handlers/instance.go` | Instance CRUD handler | ✓ VERIFIED | 787 lines, 8 endpoints: Create, List, Get, Update, Delete, Duplicate, Rename, DeletePreview. Override count aggregated from 8 tables. |
| `api/internal/api/handlers/instance_overrides.go` | Override PUT handlers | ✓ VERIFIED | 619 lines (200 shown), handlers for UpdatePorts, UpdateVolumes, UpdateEnvVars, UpdateLabels, UpdateDomains, UpdateHealthcheck, UpdateDependencies, config file CRUD |
| `api/internal/api/handlers/instance_effective.go` | Effective config resolver | ✓ VERIFIED | 609 lines, EffectiveConfig endpoint, loads template+instance for all resource types, merges with override precedence, returns overrides_applied metadata |
| `api/internal/api/routes.go` | Route wiring | ✓ VERIFIED | Lines 153-179, routes under /stacks/{name}/instances/{instance} for all CRUD + override + effective-config endpoints |
| `dashboard/src/components/stacks/add-instance-dialog.tsx` | Add instance dialog | ✓ VERIFIED | 100+ lines, template catalog with search, name validation, instance count display, calls useCreateInstance |
| `dashboard/src/components/instances/override-*.tsx` | Override editors (8 files) | ✓ VERIFIED | 8 files (ports, volumes, env-vars, labels, domains, healthcheck, dependencies, config-files), each with edit/save/reset UI |
| `dashboard/src/components/instances/effective-config-tab.tsx` | Effective config viewer | ✓ VERIFIED | 100+ lines, structured display matching override editor layout, override indicators (blue border + badge), YAML/JSON copy |
| `dashboard/src/components/instances/instance-actions.tsx` | Instance lifecycle dialogs | ✓ VERIFIED | Delete, Duplicate, Rename dialogs with validation, blast radius preview for delete |
| `dashboard/src/features/instances/queries.ts` | Query/mutation hooks | ✓ VERIFIED | 304 lines, hooks for all instance operations: useInstances, useInstance, useEffectiveConfig, useCreate/Update/Delete/Duplicate/Rename, 8 useUpdateInstance* hooks for overrides |
| `dashboard/src/routes/stacks/$name.instances.$instance.tsx` | Instance detail page | ✓ VERIFIED | 150+ lines, header with enable/disable toggle, dropdown menu for actions, tabbed interface for overrides + effective config |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|----|--------|---------|
| instance.go handlers | routes.go | chi route registration | ✓ WIRED | Line 45: `instanceHandler := handlers.NewInstanceHandler(db, containerClient)`, lines 153-179: all routes registered |
| override-ports.tsx | useUpdateInstancePorts | import + call | ✓ WIRED | Line 8: import, line 37: hook usage, line 50: mutate call with ports array |
| effective-config-tab.tsx | useEffectiveConfig | import + call | ✓ WIRED | Line 6: import, line 16: hook usage fetches merged config |
| instance-actions.tsx | useDeleteInstance | import + call | ✓ WIRED | DeleteInstanceDialog calls mutation with preview data |
| EffectiveConfig handler | merge functions | function calls | ✓ WIRED | Lines 137, 152, 227: mergeEnvVars, mergeLabels, mergeConfigFiles called with template+instance data |
| Migration 014 | service_instances table | ALTER TABLE + FK | ✓ WIRED | Lines 1-3: adds description + deleted_at columns, lines 11-86: all override tables FK to service_instances(id) ON DELETE CASCADE |

### Requirements Coverage

| Requirement | Status | Blocking Issue |
|-------------|--------|----------------|
| INST-01: Create instance from template within stack | ✓ SATISFIED | All truths verified |
| INST-02: Override instance ports | ✓ SATISFIED | override-ports.tsx + UpdatePorts handler |
| INST-03: Override instance volumes | ✓ SATISFIED | override-volumes.tsx + UpdateVolumes handler |
| INST-04: Override instance env vars | ✓ SATISFIED | override-env-vars.tsx + UpdateEnvVars handler |
| INST-05: Override instance dependencies | ✓ SATISFIED | override-dependencies.tsx + UpdateDependencies handler |
| INST-06: Override instance labels | ✓ SATISFIED | override-labels.tsx + UpdateLabels handler |
| INST-07: Override instance domains | ✓ SATISFIED | override-domains.tsx + UpdateDomains handler |
| INST-08: Override instance healthchecks | ✓ SATISFIED | override-healthcheck.tsx + UpdateHealthcheck handler |
| INST-09: Override instance config files | ✓ SATISFIED | override-config-files.tsx + config file CRUD handlers |
| INST-10: Effective config resolver | ✓ SATISFIED | instance_effective.go with merge logic |
| INST-11: List/get/update/delete instances | ✓ SATISFIED | All CRUD operations implemented |
| INST-12: Dashboard UI for instance management | ✓ SATISFIED | Full UI: add dialog, detail page, override editors, actions |
| MIGR-02: Migration 014 (override tables) | ✓ SATISFIED | Migration 014 + 017 (dependencies) exist and applied |

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| None | - | - | - | No anti-patterns detected |

**Notes:**
- No TODO/FIXME comments in handler code
- No stub patterns (empty returns, console.log-only handlers)
- No orphaned files (all components imported and used)
- Override editors all follow consistent pattern: edit mode, save/cancel, reset to template
- Effective config properly tracks `overrides_applied` metadata for each resource type
- Migration 014 properly defines 7 override tables, migration 017 adds instance_dependencies
- All dashboard components use TanStack Query hooks for API integration
- Soft-delete pattern consistent across stacks and instances

### Human Verification Required

#### 1. Template Catalog Display

**Test:** Navigate to stack detail page, click "Add Instance"
**Expected:** 
- Template catalog appears with search
- Services show image name, category, instance count
- Selecting service auto-generates unique instance name
- Name validation rejects invalid characters, duplicates
**Why human:** UI rendering, visual feedback, interaction flow

#### 2. Override Editors with Template Placeholders

**Test:** Create instance, open instance detail, navigate to each override tab (Ports, Volumes, Environment, Labels, Domains, Healthcheck, Dependencies, Config Files)
**Expected:**
- Template values shown as muted placeholders or "from template" badges
- Edit mode allows adding/modifying/removing overrides
- Save persists changes, shows success toast
- Reset button clears overrides back to template defaults
**Why human:** Visual indicators, interaction states, toast notifications

#### 3. Effective Config Preview

**Test:** After adding overrides, navigate to "Effective Config" tab
**Expected:**
- Merged config displayed in structured cards (Image, Ports, Volumes, Environment, etc.)
- Overridden sections have blue left border + "Overridden" badge
- YAML/JSON toggle works, copy button copies to clipboard
- Config shows merged result: template values + overrides (overrides win)
**Why human:** Visual indicators, clipboard interaction, merge correctness

#### 4. Instance Lifecycle Actions

**Test:** From instance detail page, test Duplicate, Rename, Enable/Disable, Delete
**Expected:**
- Duplicate: dialog with auto-generated name, creates new instance with all overrides copied
- Rename: validates new name, updates instance_id and container_name
- Enable/Disable: toggles enabled state, shows in instance list
- Delete: shows blast radius preview (instance name, template, override count, container name), soft-deletes
**Why human:** Dialog flows, navigation after actions, data consistency

#### 5. Instance List Display

**Test:** Navigate to stack detail page, view instance list
**Expected:**
- Instances show name, template, override count, enabled status
- Override count reflects total across all resource types
- Instance cards clickable to detail page
- Actions accessible from list (delete, duplicate, enable/disable)
**Why human:** UI rendering, computed values, interaction flow

---

_Verified: 2026-02-06T13:45:00Z_
_Verifier: Claude (gsd-verifier)_
