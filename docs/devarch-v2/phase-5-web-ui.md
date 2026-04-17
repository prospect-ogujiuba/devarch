# Phase 5 — Web UI

## Goal
Ship the workspace-first web interface for define -> plan -> apply -> observe without recreating V1 admin sprawl.

## Acceptance

- navigation is limited to Workspaces, Catalog, Activity, and Settings
- workspace detail view exposes overview, resources, graph, plan, logs, and raw config tabs
- users can inspect and edit core manifest content
- plan/apply/status/logs flows work through the shared API
- UI quality is fast, opinionated, and developer-tool oriented

## Recommended sequence

Land shell/navigation first, then workspace pages, then plan/logs/graph, then editing flows.

---

### P5-UI-001 — Build app shell and primary navigation
- **Owner:** surgeon-ui
- **Depends on:** Phase 4 complete
- **Goal:** create the minimal four-area navigation shell
- **Tasks:**
  1. Add top-level routes for Workspaces, Catalog, Activity, and Settings.
  2. Remove or avoid V1-style admin-surface expansion in V2 routes.
  3. Establish shared page layout and loading/error states.
- **Validation:**
  - route-level UI tests
- **Done when:**
  - the V2 shell exists and stays within the planned nav scope

### P5-UI-002 — Implement workspace list and detail frame
- **Owner:** surgeon-ui
- **Depends on:** P5-UI-001
- **Goal:** make workspaces the main entry point of the UI
- **Tasks:**
  1. Add workspace list and selection flow.
  2. Add the workspace detail frame and tab scaffold.
  3. Wire manifest, graph, status, and plan loading states.
- **Validation:**
  - route/component tests
- **Done when:**
  - a user can open a workspace and navigate its core tabs

### P5-UI-003 — Add Overview and Resources tabs
- **Owner:** surgeon-ui
- **Depends on:** P5-UI-002
- **Goal:** present normalized workspace and resource information clearly
- **Tasks:**
  1. Show workspace metadata and high-level runtime state.
  2. Show resource summaries, contracts, ports, and health hints.
  3. Keep visual design compact and developer-oriented.
- **Validation:**
  - component tests
- **Done when:**
  - resource inspection feels smaller and clearer than V1

### P5-UI-004 — Add Graph and Plan tabs
- **Owner:** surgeon-ui
- **Depends on:** P5-UI-002
- **Goal:** visualize dependencies and planned actions
- **Tasks:**
  1. Render import/export graph relationships.
  2. Render plan actions with reasons and action types.
  3. Make ambiguity and unresolved contract states visible.
- **Validation:**
  - component tests
  - Playwright smoke test for plan view
- **Done when:**
  - users can understand what will happen before apply

### P5-UI-005 — Add Logs tab and live activity flow
- **Owner:** surgeon-ui
- **Depends on:** P4-API-002, P5-UI-002
- **Goal:** expose runtime logs and event-backed updates in the UI
- **Tasks:**
  1. Stream or fetch resource logs.
  2. Surface workspace activity and apply progress.
  3. Handle reconnect and empty-state behavior cleanly.
- **Validation:**
  - UI tests
  - Playwright smoke test for logs view
- **Done when:**
  - users can observe running workspaces without leaving the UI

### P5-UI-006 — Add raw manifest editor
- **Owner:** surgeon-ui
- **Depends on:** P5-UI-002
- **Goal:** give advanced users a direct manifest editing path
- **Tasks:**
  1. Add raw manifest view/edit support.
  2. Surface validation errors from schemas clearly.
  3. Keep edits scoped to the canonical manifest model.
- **Validation:**
  - editor tests
- **Done when:**
  - advanced users can edit the manifest directly with validation feedback

### P5-UI-007 — Add structured editor for common resource fields
- **Owner:** surgeon-ui
- **Depends on:** P5-UI-006
- **Goal:** support high-frequency edits without bespoke override pages
- **Tasks:**
  1. Build structured forms for common fields like env, ports, imports, and volumes.
  2. Drive forms from schema/model metadata where practical.
  3. Avoid recreating one editor per override subtype.
- **Validation:**
  - component tests
  - focused Playwright edit flow
- **Done when:**
  - common edits are easier than raw YAML while preserving the manifest-first model

### P5-CAT-001 — Implement catalog browser
- **Owner:** surgeon-ui
- **Depends on:** P5-UI-001, P4-API-001
- **Goal:** browse builtin and loaded templates without expanding into V1 admin CRUD
- **Tasks:**
  1. Add template list and detail views.
  2. Show metadata, contracts, ports, volumes, and runtime defaults.
  3. Keep catalog browsing read-heavy and lightweight.
- **Validation:**
  - route/component tests
- **Done when:**
  - users can inspect reusable templates from the UI

## Parallel-safe packets

After `P5-UI-002`, these can run in parallel if they touch separate route/components:

- `P5-UI-003`
- `P5-UI-004`
- `P5-UI-005`
- `P5-CAT-001`

`P5-UI-006` and `P5-UI-007` should stay sequential unless the architect splits editor ownership clearly.

## Handoff to Phase 6

Phase 6 starts only after:

- users can complete define -> plan -> apply -> observe from the UI
- navigation remains constrained to the four planned areas
- smoke coverage exists for core workspace flows
