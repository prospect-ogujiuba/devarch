# Images Page

**Path:** `dashboard/src/routes/images/index.tsx`
**Last Updated:** 2026-02-17

## Overview
Container image management page with table/grid view toggle, search, sort, and image pull/prune dialogs. Displays image stats (total size, dangling count) and streaming pull progress.

## Route
- File route: `/images/`
- No search param validation

## Key Functions
- `ImagesPage()` — Main route component
- `imageTableView()` — Table renderer (image, tag, ID, size, created)
- `handlePullStart()` — Pull image by reference with streaming progress
- `handleRemove()` — Remove single image with confirmation
- `handlePrune()` — Remove all dangling images
- `formatAge()` — Convert Unix timestamp to human-readable age (e.g., "5d ago")

## State Management
- `search`, `sortBy`, `sortDir`, `viewMode` — List controls
- `pullOpen`, `pruneOpen`, `removeTarget` — Dialog visibility states
- `pulling` — Pull operation in progress
- `pullProgress` — Array of streaming pull events
- `filteredImages` — Memoized search/sort result

## Stat Cards
- Total Images — Count of all images
- Total Size — Formatted bytes
- Dangling Count — Yellow badge if > 0

## View Modes

### Table View
Columns: Image | Tag | ID | Size | Created | Actions
- Shows full metadata in rows
- Delete button per image

### Grid View
Cards with: repo name, tag, short SHA256, size, age
- Delete button per card
- Responsive: 2-4 columns

## Dialogs

### Pull Image
- Input: image reference (e.g., nginx:latest)
- Progress: auto-scrolling output log (max 50 lines)
- States: input → pulling → done
- Success: toast + query invalidation

### Remove Image
- Confirmation: "Remove {tag}?"
- Mutation-driven removal

### Prune Dialog
- Removes all dangling/untagged images
- Confirmation prompt

## Search & Sort
- Search: case-insensitive tag matching
- Sort by: Name, Size, Age
- Sort dir: Asc/Desc

## Dependencies
- `useImages()`, `useRemoveImage()`, `usePruneImages()` — Image queries
- `pullImageWithProgress()` — Streaming pull
- `ListPageScaffold` — Page layout component
- `Dialog`, `Table` — Radix UI
- `formatBytes()` — Format file sizes
- Icons: HardDrive, Database, AlertTriangle, Download, Trash2, Eraser, Loader2

## Recent Changes
- Added grid/table view toggle wired to controls
- Grid view renders manual card layout with stats
- Pull progress auto-scrolls to bottom
- Dangling count has conditional yellow styling

## Related Pages
- `/routes/categories/`, `/routes/services/` — Similar ListPageScaffold pattern
