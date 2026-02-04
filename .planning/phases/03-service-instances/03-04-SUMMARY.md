---
phase: 03-service-instances
plan: 04
subsystem: dashboard-overrides
tags: [react, tanstack-router, instance-detail, override-editors, codemirror]
requires: [03-02, 03-03]
provides:
  - instance-detail-page
  - override-editing-ui
  - port-volume-env-label-domain-healthcheck-config-editors
affects: [03-05]
tech-stack:
  added: []
  patterns: [override-editor-pattern, template-value-as-placeholder, explicit-save-with-dirty-tracking]
key-files:
  created:
    - dashboard/src/routes/stacks/$name.instances.$instance.tsx
    - dashboard/src/components/instances/override-ports.tsx
    - dashboard/src/components/instances/override-volumes.tsx
    - dashboard/src/components/instances/override-env-vars.tsx
    - dashboard/src/components/instances/override-labels.tsx
    - dashboard/src/components/instances/override-domains.tsx
    - dashboard/src/components/instances/override-healthcheck.tsx
    - dashboard/src/components/instances/override-config-files.tsx
  modified:
    - dashboard/src/routes/stacks/$name.tsx
    - dashboard/src/routeTree.gen.ts
decisions:
  - slug: override-editor-ux-pattern
    summary: Template values shown muted, overrides shown with blue left border
    rationale: Clear visual distinction between template defaults and instance-specific config
  - slug: explicit-save-not-autosave
    summary: All override editors require explicit save button click
    rationale: Prevents accidental changes, gives user control over when mutations happen
  - slug: per-field-reset-plus-reset-all
    summary: Each override field has X icon to remove, plus Reset All button in section header
    rationale: Granular control for single overrides, convenience for clearing all at once
  - slug: config-files-use-codemirror
    summary: Config file editing uses CodeMirror with language detection
    rationale: Reuse existing code-editor component, syntax highlighting for YAML/JSON/XML
  - slug: template-config-files-as-reference
    summary: Template file content shown read-only above editable override content
    rationale: User can reference original template while editing override
metrics:
  duration: 5min
  completed: 2026-02-04
---

# Phase 3 Plan 4: Instance Override Editors Summary

Instance detail page with tabbed override editors for all resource types

## One-liner

Instance detail page with 7 override editor tabs (ports, volumes, env vars, labels, domains, healthcheck, config files), explicit save, template values as context

## What Was Built

### Instance Detail Page
- File-based route: `/stacks/{name}/instances/{instance}`
- Breadcrumb: Stacks > {stack-name} > {instance-name}
- Header: instance name, template name, container name, enabled status dot
- Actions: Enable/Disable, Duplicate, Rename, Delete (dialogs)
- 9 tabs: Info, Ports, Volumes, Environment, Labels, Domains, Healthcheck, Config Files, Effective Config (placeholder)
- Info tab: metadata (name, template, container, created/updated), editable description
- Instance cards in stack detail page now link to instance detail

### Override Editor Pattern
All 7 override editors follow consistent UX:
1. **Template section (read-only)**: Show template values in muted text
2. **Override section (editable)**: Show instance-specific config with blue left border
3. **Add Override button**: Add new override row
4. **Per-field reset**: X icon on each override row to remove that override
5. **Reset All button**: Clear all overrides in section (confirmation)
6. **Save button**: Disabled when no changes, triggers PUT mutation
7. **Dirty tracking**: Compare local state to initial data
8. **Loading state**: Button shows spinner during mutation

### Override Ports Editor
- Template ports shown as: `{host_ip}:{host_port}:{container_port}/{protocol}`
- Override form: host_ip input (default 0.0.0.0), host_port, container_port, protocol dropdown (tcp/udp)
- Each override row has colored left border

### Override Volumes Editor
- Fields: volume_type dropdown (bind/volume/tmpfs), source, target, read_only checkbox, is_external checkbox
- Template volumes shown with type badge + source → target
- Override volumes grouped per volume with checkboxes

### Override Env Vars Editor
- Template env vars shown as: `KEY = value` (or `••••••••` if secret)
- Override form: key input, value input (password-style if secret), secret toggle
- For keys in template: show template value as placeholder in value field
- Show/hide secrets toggle per row

### Override Labels Editor
- System labels (devarch.*) shown with lock icon, read-only
- Template labels shown muted
- Override labels editable
- Client-side validation: reject "devarch." prefix with error message

### Override Domains Editor
- Simple list: domain input + optional proxy_port input
- Template domains shown with badge
- Override domains shown with blue border

### Override Healthcheck Editor
- Individual field editing: test (textarea), interval, timeout, retries, start_period (all number inputs)
- Template healthcheck shown as read-only reference above override form
- Single save for all fields
- If test field empty on save, sends null (removes healthcheck override)

### Override Config Files Editor
- List view: template files + custom override files
- Template files show "View" or "Edit Override" button
- Custom files (not in template) show "Edit" + "Delete" buttons
- Click file opens dialog with CodeMirror editor
- Template content shown read-only above editable override content
- Language detection from file extension (JSON, YAML, XML)
- File mode input (default 0644)
- Add Config File button for new files not in template
- Reset to Template button (DELETE override, reverts to template)

## Deviations from Plan

None — plan executed as written

## Technical Notes

### Override Editor Component Pattern
```tsx
interface Props {
  instance: InstanceDetail
  templateData: Service
  stackName: string
  instanceId: string
}
```

All editors:
- Fetch template via `useService(instance.template_name)`
- Override data from `instance.{ports|volumes|env_vars|...}`
- Local state for editing (`useState` with drafts array)
- Dirty tracking: `JSON.stringify(drafts) !== JSON.stringify(initial)`
- Mutation hooks: `useUpdateInstance{Ports|Volumes|...}(stackName, instanceId)`
- Toast notifications on success/error
- Cache invalidation: instance detail + effective config + list

### CodeMirror Integration
Reused existing `dashboard/src/components/services/code-editor.tsx`:
- Language extensions: JSON, YAML, XML
- OneDark theme
- Line numbers, active line highlight
- History + undo/redo
- Read-only mode for template content
- 400px height

### File Paths
Instance detail route: `dashboard/src/routes/stacks/$name.instances.$instance.tsx`
Override editors: `dashboard/src/components/instances/override-{type}.tsx`

## Next Phase Readiness

**Blocks:** None

**Enables:**
- 03-05: Effective Config tab (merge logic to show final composed config)
- Phase 4 networking can wire instances to networks
- Phase 6 compose generation can use instance overrides

**Context for Next Plans:**
- Override editing UX established (explicit save, dirty tracking, reset options)
- Template data shown as read-only reference alongside editable overrides
- All 7 override types have working editors with consistent patterns

## State Changes

### Files Created
8 new files:
- 1 route (instance detail page)
- 7 override editor components

### Files Modified
- `dashboard/src/routes/stacks/$name.tsx` — instance cards now link to detail page
- `dashboard/src/routeTree.gen.ts` — auto-generated route tree

### Dependencies
No new dependencies — reused existing CodeMirror setup from service editors

## Verification

1. Navigate to stack detail page
2. Click instance card → routes to `/stacks/{name}/instances/{instance}`
3. Instance detail page renders with header, actions, 9 tabs
4. Info tab shows metadata
5. Ports tab shows template ports + override editor
6. Volumes tab shows template volumes + override editor
7. Environment tab shows template env vars + override editor
8. Labels tab shows system labels (locked) + template labels + override editor
9. Domains tab shows template domains + override editor
10. Healthcheck tab shows template healthcheck + override editor
11. Config Files tab shows template files + custom files
12. Effective Config tab shows placeholder
13. Edit override → Save button disabled until change → Save succeeds → Toast confirmation

## Task Breakdown

| Task | Description | Commit | Files |
|------|-------------|--------|-------|
| 1 | Instance detail page with tabbed layout | b4c7ae1d | $name.instances.$instance.tsx, $name.tsx, routeTree.gen.ts |
| 2 | Override editor components for all 7 types | 3faabb99 | override-ports/volumes/env-vars/labels/domains/healthcheck/config-files.tsx |

## Time Breakdown

- Task 1 (detail page): 2min
- Task 2 (7 editors): 3min
- Total: 5min
