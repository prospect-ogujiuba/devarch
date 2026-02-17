# Images Page

**Path:** `dashboard/src/routes/images/index.tsx`
**Last Updated:** 2026-02-17

## Overview
Container image management page with table/grid view toggle, search, sort, and image pull/prune dialogs. Displays image stats (total size, dangling count) and streaming pull progress.

## Key Functions
- `ImagesPage()` — Main route component (file route `/images/`)
- `imageTableView()` — Table renderer showing image details (tag, ID, size, age)
- `handlePullStart()` — Pull image by reference with streaming progress
- `handleRemove()` — Remove single image by tag
- `handlePrune()` — Remove all dangling images
- `formatAge()` — Helper to format Unix timestamps as human-readable age

## Props/State
- `search`, `sortBy`, `sortDir`, `viewMode` — List controls synced to URL
- `pullOpen`, `pruneOpen`, `removeTarget` — Dialog states
- `pullProgress` — Array of streaming pull events for progress display
- `filteredImages` — Memoized search/sort result

## Dependencies
- `useImages()`, `useRemoveImage()`, `usePruneImages()`, `pullImageWithProgress()` — Image queries
- `ListPageScaffold` — Standard page layout with controls, stats, filters
- `Dialog`, `Table` — Radix UI components
- `formatBytes()` — Utility for byte display

## Recent Changes
- Added `viewMode` state (table/grid toggle) wired to controls
- Refactored to use `ListPageScaffold` for consistent page structure
- Grid view renders manual card layout (not yet EntityCard-based)
- Pull progress auto-scrolls and streams incrementally

## Related Components
- `/routes/services`, `/routes/categories` — Similar page pattern with ListPageScaffold
