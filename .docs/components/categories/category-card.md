# CategoryCard Component

**Path:** `dashboard/src/components/categories/category-card.tsx`
**Last Updated:** 2026-02-17

## Overview
Displays a single category with running service count, resource bar, and action buttons (start/stop all, edit, delete). Supports compact and full modes; used in overview page and categories list page.

## Props
```typescript
interface CategoryCardProps {
  category: Category
  compact?: boolean
}
```

- `category` — Category object with name, service_count, runningCount, startup_order
- `compact` — If true, minimal horizontal layout; if false, full card layout (default)

## Features

### Compact Mode (Overview Page)
- Single horizontal row layout
- Small resource bar with running count badge (right side)
- Play/stop buttons in small icon-only style
- Minimal height for dense layout

### Full Mode (Categories Page)
- Large running count display (e.g., "3/5 services running")
- Full-width resource bar showing percentage
- Edit/delete buttons in header (small icons)
- Full-width Start All / Stop All buttons
- Edit and Delete dialogs

### Shared Logic
- Calculates `allRunning` — all services running (show stop button)
- Calculates `allStopped` — no services running (show start button)
- Only shows Start button if not all running
- Only shows Stop button if not all stopped
- Disabled state during mutation

## State
- `editOpen`, `deleteOpen` — Dialog visibility
- Uses mutations: `useStartCategory()`, `useStopCategory()`
- Mutation loading state disables all buttons

## Styling
- Uses `EntityCard` for consistent hover effects
- Resource bar: `bg-green-500` (full width, transitions smoothly)
- Link styling: `hover:underline` for category name

## Mutations
- `useStartCategory()` — Start all services in category
- `useStopCategory()` — Stop all services in category

## Dialogs
- `EditCategoryDialog` — Edit category name/order
- `DeleteCategoryDialog` — Confirm deletion

## Dependencies
- `EntityCard` — Card wrapper with hover styling
- `EditCategoryDialog`, `DeleteCategoryDialog` — Edit/delete flows
- `ResourceBar` — Visual percentage display
- `useStartCategory()`, `useStopCategory()` — Mutations
- `categoryLabel()` — Format category name display
- Icons: Play, Square, Loader2, Pencil, Trash2

## Related Components
- `/routes/index.tsx` (overview) — Uses compact=true
- `/routes/categories/index.tsx` — Uses compact=false (or omitted)
