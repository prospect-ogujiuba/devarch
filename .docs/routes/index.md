# Dashboard Home (Overview Page)

**Path:** `dashboard/src/routes/index.tsx`
**Last Updated:** 2026-02-17

## Overview
Landing page showing system stats, runtime info, top services by CPU, and categories list with status filters. Serves as quick health check and navigation hub.

## Route
- File route: `/` (exact match)
- Search params: q (search), status (filter), sort, dir, view, page, size

## Main Sections

### 1. Stat Cards (Top Grid)
6-column responsive grid showing:
- **Total Services** — Server icon
- **Running** — Play icon, green text
- **Stopped** — Square icon, muted text
- **Categories** — Activity icon
- **Avg CPU** — CPU icon, formatted as percent
- **Total Memory** — Memory icon, formatted as MB

Uses `StatCard` component with icon, label, and value.

### 2. Runtime Info Bar
Single-line info bar showing:
- Container runtime (podman/docker) with Server icon
- Socket path (monospace, if available)
- Enabled services count
- Network status: name + exists indicator (yellow if missing)

### 3. Top Services by CPU
Table showing top 5 services by CPU usage:
- Columns: Name (linked), Category, CPU%, Memory MB, Status
- Memoized: `topServices` array sorted by CPU percentage
- Empty state: "No metrics available" if no running services with metrics

### 4. Categories Section
Search, sort, view toggle + status filter + paginated grid:
- **Search:** By category name
- **Sort:** Name (default), Services count, Running count
- **View:** Table or grid
- **Filter:** Status (all/running/partial/stopped) with counts

Uses `CategoryCard` in compact mode, or `CategoryTable` for table view.

## State & Data Management

### Queries
- `useStatusOverview()` — Overview data: categories, service counts, container runtime
- `useServices()` — All services (for metrics calculation)

### Computed Values
- `topServices` — Memoized top 5 services sorted by CPU % (descending)
- `serviceStats` — Average CPU and total memory across all services with metrics
- `statusCounts` — Category counts by status (running/partial/stopped)

### URL Syncing
- Uses `useUrlSyncedListControls()` — Search/sort/filter → URL
- Uses `useUrlPagination()` — Page number/size → URL
- Pagination resets on search/sort/filter change

## Statistics Calculation
```
avgCpu = totalCpu / runningWithMetrics
totalMem = sum of all service memory_used_mb
```

Categories status breakdown:
- Running: all services in category running
- Partial: some services running
- Stopped: no services running

## Recent Changes
- Added `topServices` memoization with CPU-based filtering and sorting
- Runtime info bar displays container runtime, socket path, enabled count, network status
- Categories use compact `CategoryCard` (not full layout)
- Status filter shows counts for each option

## Dependencies
- `useStatusOverview()`, `useServices()` — Query hooks
- `CategoryCard` — Compact card renderer
- `CategoryTable` — Table view alternative
- `StatCard`, `FilterBar`, `ListToolbar`, `EmptyState`, `PaginationControls` — UI components
- `getServiceStatus()` — Service status helper
- Icons: Server, Play, Square, Activity, Loader2, Cpu, MemoryStick, FolderOpen

## Related Pages
- `/categories/` — Full categories page
- `/services/` — Services list
