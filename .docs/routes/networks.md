# Networks Page

**Path:** `dashboard/src/routes/networks/index.tsx`
**Last Updated:** 2026-02-17

## Overview
Networks list with search, sort (name/stack/containers/created), status filter (managed/orphaned/external), table/grid view, bulk selection, and single/bulk removal. Create dialog for manual network creation.

## Features
- **Search:** By network name or stack name
- **Sort Options:** Name, Stack, Container count, Created date
- **Status Filter:** Managed, Orphaned, External
- **View Modes:** Table or grid
- **Selection:** Multi-select with per-item and bulk toggle
- **Bulk Actions:** Remove multiple networks (if selected)
- **Create Action:** Manual network creation dialog

## State Management
- `selected` — Set<string> of selected network names
- `newName` — Create dialog input
- Synced to URL via `useUrlSyncedListControls()` and `useUrlPagination()`

## Stat Cards
- Total networks
- Managed count
- Orphaned count (yellow)
- Total containers

## Default View
- Default sort: 'name'
- Default view: 'table'

## Conditional UI
- If no networks and no search/filter:
  - Show empty state with quick create action
  - Simpler header layout
- Otherwise:
  - Full ListPageScaffold with controls
  - Action button shows "Remove N" if items selected, else "Create Network"

## Selection Behavior
- `handleToggleSelect()` — Toggle single network
- `handleToggleAll()` — Toggle all networks on current page
- `handleBulkRemove()` — Remove all selected networks and clear selection

## NetworkGrid Changes
- **New Props:** `selected` (Set), `onToggleSelect` (function)
- NetworkGrid now renders checkboxes and wires selection state

## Dependencies
- `useNetworks()` — Fetch networks
- `useCreateNetwork()`, `useRemoveNetwork()`, `useBulkRemoveNetworks()` — Mutations
- `NetworkTable`, `NetworkGrid` — View renderers
- `ListPageScaffold` — Page layout

## Route & Navigation
- Path: `/networks/`
- URL params: q, sort, dir, view, status, page, size
