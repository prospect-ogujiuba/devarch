# Categories Page

**Path:** `dashboard/src/routes/categories/index.tsx`
**Last Updated:** 2026-02-17

## Overview
Full-featured categories list with search, sort (name/services/running/order), status filter (all/running/partial/stopped), table/grid view toggle, pagination, and create dialog.

## Route
- File route: `/categories/`
- Search params: q, status, sort, dir, view, page, size

## Features

### Search
- Case-insensitive category name matching
- Input field placeholder: "Search categories..."

### Sort Options
1. **Name** (default) — Alphabetical
2. **Services** — Total service count
3. **Running** — Running service count
4. **Startup Order** — startup_order field

Sort direction: asc/desc toggle

### Status Filter
- **All** — All categories
- **All Running** — All services in category running
- **Partial** — Some services running
- **Stopped** — No services running

Each filter option shows count of matching categories.

### View Modes
- **Table:** Full `CategoryTable` with all columns
- **Grid:** 2-col (mobile), 3-col (tablet/desktop) grid of full `CategoryCard`

### Pagination
- Configurable page size (default: 12)
- Page size options: [12, 25, 50]
- Resets page on search/sort/filter change

### Create Action
- Button opens `CreateCategoryDialog`
- Dialog state: `createOpen`

## State Management

### URL Syncing
- `useUrlSyncedListControls()` — search, sort, filters, view mode
- `useUrlPagination()` — page, pageSize
- All state persists in URL query params

### Filters
- `status` — Filter function checks running vs total services

### Sort Functions
- Implemented for all sort options with proper null handling

### Callbacks
- `handleSearchChange()` — Resets pagination
- `handleSortByChange()` — Resets pagination
- `handleSortDirChange()` — Resets pagination
- `handleViewModeChange()` — Resets pagination
- `handleStatusFilterChange()` — Resets pagination

## Stat Cards
- **Categories** — Server icon, total count
- **Services Running** — Play icon, green text
- **Services Stopped** — Square icon, muted text

## Grid Layout
- `md:grid-cols-2 lg:grid-cols-3` responsive

## Dependencies
- `useCategories()` — Fetch categories
- `useUrlSyncedListControls()` — URL syncing
- `useUrlPagination()` — Pagination
- `CategoryCard`, `CategoryTable` — View renderers
- `CreateCategoryDialog` — Create flow
- `FilterBar` — Status filter UI
- `ListPageScaffold` — Page layout
- Icons: Server, Play, Square, Plus

## Recent Changes
- Uses `ListPageScaffold` for consistent page structure
- Create button is action button in header
- Stat cards show running/stopped counts
- Status filter options show counts

## Related Pages
- `/` (overview) — Shows categories with compact=true cards
- `/services/` — Filter by category link
