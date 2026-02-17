# Dashboard Home (Overview Page)

**Path:** `dashboard/src/routes/index.tsx`
**Last Updated:** 2026-02-17

## Overview
Landing page showing system stats, runtime info, top services by CPU, and categories list with status filters. Serves as quick health check and navigation hub.

## Main Sections

### Stat Cards (Grid)
- Total Services, Running, Stopped, Categories, Avg CPU, Total Memory
- Uses icon badges with values

### Runtime Info Bar
- Container runtime (podman/docker)
- Socket path (if available)
- Count of enabled services
- Yellow highlight for missing network indicator

### Top Services by CPU
- Table of top 5 services sorted by CPU usage
- Shows name, category, CPU%, memory, status
- Empty state if no metrics available

### Categories Section
- Search, sort, view toggle (table/grid)
- Status filter: All, All Running, Partial, Stopped
- Paginated grid of category cards (compact mode)
- Shows running count per category

## State & Logic
- `useStatusOverview()` — Overview data (categories, status counts)
- `useServices()` — All services (for top CPU calculation)
- `topServices` — Memoized top 5 by CPU percentage
- `serviceStats` — Average CPU and total memory across running services
- `statusCounts` — Running/partial/stopped category counts

## Recent Changes
- Added `topServices` memoization with CPU-based sorting
- Runtime info bar now displays container runtime, socket path, enabled count
- Category compact cards used (not full layout)

## Dependencies
- `useStatusOverview()`, `useServices()` — Query hooks
- `ListPageScaffold` — Not used; custom layout
- `CategoryCard` — Compact card renderer
- `CategoryTable` — Table view
- `StatCard`, `FilterBar`, `ListToolbar` — UI components

## Navigation Entry Point
- Route: `/` (home, exact match)
