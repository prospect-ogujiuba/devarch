---
phase: 05-compose-generation
plan: 02
subsystem: dashboard
tags: [react, codemirror, yaml, compose, preview, download]
requires: [phase-05-01]
provides: [compose-preview-tab, compose-download, compose-warnings-ui]
affects: [phase-06-plan-apply]
tech-stack:
  added: []
  patterns: [tabbed-detail-layout, blob-download, read-only-codemirror]
key-files:
  created: []
  modified:
    - dashboard/src/types/api.ts
    - dashboard/src/features/stacks/queries.ts
    - dashboard/src/routes/stacks/$name.tsx
key-decisions:
  - "Tabbed layout (Instances + Compose) replaces flat card layout on stack detail"
  - "CodeMirror with yaml language and readOnly for syntax-highlighted preview"
  - "Blob download pattern for client-side YAML file save"
duration: 1.8min
completed: 2026-02-07
---

# Phase 5 Plan 2: Dashboard Compose Preview Summary

Compose tab on stack detail page with CodeMirror YAML preview, warning display, and download capability — users see exactly what YAML will be generated for their stack.

## Accomplishments

- StackCompose type added (yaml, warnings, instance_count)
- useStackCompose query hook fetches from /stacks/{name}/compose (on-demand, no polling)
- Stack detail page restructured: stats cards at top, then tabbed layout (Instances tab + Compose tab)
- Compose tab: CodeMirror read-only YAML preview with syntax highlighting
- Warnings section with AlertTriangle icon and yellow styling (dark mode aware)
- Download button creates docker-compose-{stack}.yml via Blob URL
- Loading spinner while compose fetches, empty state when no YAML available
- Instances tab preserves all existing functionality (instance grid, network card, add button)

## Task Commits

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Stack compose type and query hook | dcdd8457 | api.ts, queries.ts |
| 2 | Compose tab on stack detail page | a50285c1 | stacks/$name.tsx |

## Files Modified

- `dashboard/src/types/api.ts` -- Added StackCompose interface
- `dashboard/src/features/stacks/queries.ts` -- Added useStackCompose hook, imported StackCompose type
- `dashboard/src/routes/stacks/$name.tsx` -- Added Tabs layout, Compose tab with CodeMirror, warnings, download

## Decisions Made

1. **Tabbed layout on stack detail** -- Instances and Compose as peer tabs keeps the page organized as features grow. Stats cards remain above tabs for always-visible context.
2. **CodeMirror over pre tag** -- Syntax highlighting via existing CodeEditor component (yaml language mode). Read-only with no-op onChange.
3. **Blob download pattern** -- Client-side file creation avoids needing a separate download endpoint. Creates and revokes object URL immediately.

## Deviations from Plan

None -- plan executed exactly as written.

## Issues Encountered

None.

## Next Phase Readiness

- Phase 5 complete: stack compose generator (05-01) + dashboard preview (05-02)
- Phase 6 (Plan/Apply) can build on compose generation + preview infrastructure
- Compose tab provides visual validation before plan/apply workflow adds execution
