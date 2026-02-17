# Networks Page

**Path:** `dashboard/src/routes/networks/index.tsx`
**Last Updated:** 2026-02-17

## Overview
Networks list with search, sort (name/stack/containers/created), status filter (managed/orphaned/external), table/grid view, multi-select, and bulk removal. Manual network creation via dialog.

## Route
- File route: `/networks/`
- Search params: q, sort, dir, view, status, page, size

## Features

### Search
- By network name or stack name (case-insensitive)
- Input field placeholder: "Search networks..."

### Sort Options
1. **Name** (default) — Alphabetical
2. **Stack** — Stack name (or empty)
3. **Containers** — Container count
4. **Created** — Creation date

Sort direction: asc/desc toggle

### Status Filter
- **Managed** — Stack-managed networks
- **Orphaned** — Networks with no stack/containers
- **External** — Unmanaged networks

Shows count for each option (updates dynamically).

### View Modes
- **Table:** Full `NetworkTable` with checkboxes and columns
- **Grid:** Grid of `NetworkGrid` cards (default)

### Selection & Bulk Actions
- Per-network checkbox (toggle individual)
- Table header checkbox (toggle all on page)
- Bulk remove button: visible only when items selected
  - Shows "Remove N" count
  - Calls `handleBulkRemove()` to remove all selected

### Pagination
- Default page size: 12
- Options: [12, 25, 50]
- Resets on search/sort/filter change

### Create Action
- Manual network creation via dialog
- Button shown in action area
- If selected: "Remove N" + "Create" buttons
- Otherwise: "Create Network" button

## State Management

### URL Syncing
- `useUrlSyncedListControls()` — search, sort, filters, view mode
- `useUrlPagination()` — page, pageSize
- State persists in URL

### Local State
- `selected` — Set of selected network names
- `createOpen` — Create dialog visibility
- `newName` — Input value for new network name

### Callbacks
- `handleSearchChange()` — Update search + reset page
- `handleSortByChange()` — Update sort + reset page
- `handleSortDirChange()` — Toggle sort direction + reset page
- `handleViewModeChange()` — Toggle view mode + reset page
- `handleRemove()` — Remove single network
- `handleCreate()` — Create new network
- `handleToggleSelect()` — Toggle single network selection
- `handleToggleAll()` — Toggle all networks on current page
- `handleBulkRemove()` — Remove multiple selected networks

## Special UI States

### Empty State (No Networks & No Search/Filter)
- Simple header with title + create button
- Large empty state card with message and action link
- Create dialog at bottom

### With Networks
- Full `ListPageScaffold` layout
- Controls and stats visible
- Pagination controls if > 1 page

## Stat Cards
- **Total** — Network icon, total count
- **Managed** — Network icon, managed count
- **Orphaned** — AlertTriangle icon, yellow color
- **Containers** — Boxes icon, total container count

Grid: 2 columns (responsive), 4 columns on large screens

## Mutations & Hooks
- `useNetworks()` — Fetch networks
- `useCreateNetwork()` — Create mutation
- `useRemoveNetwork()` — Delete single mutation
- `useBulkRemoveNetworks()` — Delete multiple mutation

## CreateNetworkDialog Component
Inline dialog with:
- Input: network name (autofocus, enter to submit)
- Cancel/Create buttons
- Disabled until name entered
- Loading state on create button

## Dependencies
- `useNetworks()`, `useCreateNetwork()`, `useRemoveNetwork()`, `useBulkRemoveNetworks()` — Mutations
- `NetworkTable`, `NetworkGrid` — View renderers
- `ListPageScaffold` — Page layout
- `Dialog`, `Input`, `Button`, `EmptyState` — UI components
- Icons: Network, AlertTriangle, Boxes, Plus, Trash2

## Recent Changes
- Added checkbox selection to both table and grid views
- Bulk remove button integrated with single create button
- Grid view now supports multi-select via NetworkGrid

## Related Components
- `/components/networks/network-table.tsx` — Table view
- `/components/networks/network-grid.tsx` — Grid view
