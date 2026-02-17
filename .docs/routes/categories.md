# Categories Page

**Path:** `dashboard/src/routes/categories/index.tsx`
**Last Updated:** 2026-02-17

## Overview
Full-featured categories list with search, sort (name/services/running/order), status filter (all/running/partial/stopped), table/grid view toggle, and pagination.

## Features
- **Search:** By category name
- **Sort Options:** Name (default), Services count, Running count, Startup order
- **Status Filter:** All, All Running (all services in category running), Partial, Stopped
- **View Modes:** Table or grid of cards
- **Pagination:** With configurable page size
- **Create Action:** Dialog to create new category

## Default Sort Changed
- **Previous:** `defaultSort: 'order'` (startup order)
- **Current:** `defaultSort: 'name'` (alphabetical)

## State
- Uses `useUrlSyncedListControls()` for search/sort/filter sync to URL
- Uses `useUrlPagination()` for page/size sync to URL
- `createOpen` — Create dialog visibility

## Stat Cards
- Categories count
- Services running (green)
- Services stopped (muted gray)

## Grid Layout
- 2 col on mobile, 3 on tablet, wide screens
- Full `CategoryCard` (non-compact)
- Each card has edit/delete buttons

## Table Layout
- Full `CategoryTable` view

## Dependencies
- `useCategories()` — Fetch categories
- `CategoryCard`, `CategoryTable` — View renderers
- `CreateCategoryDialog` — Create flow
- `FilterBar` — Status filter UI
- `ListPageScaffold` — Page layout with controls

## Route & Navigation
- Path: `/categories/`
- Uses TanStack Router search params: q, status, sort, dir, view, page, size
