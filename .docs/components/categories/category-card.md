# CategoryCard Component

**Path:** `dashboard/src/components/categories/category-card.tsx`
**Last Updated:** 2026-02-17

## Overview
Displays a single category with running service count, resource bar, and action buttons (start/stop all, edit, delete). Supports compact and full modes; used in overview page and categories list.

## Props
```typescript
interface CategoryCardProps {
  category: Category
  compact?: boolean
}
```

- `category` — Category object with name, service_count, runningCount, startup_order
- `compact` — If true, minimal horizontal layout; false = full card layout

## Features

### Compact Mode (Overview Page)
- Horizontal resource bar with running count
- Single row play/stop buttons
- Minimal height

### Full Mode (Categories Page)
- Large running count display (e.g., "3/5")
- Resource bar showing percentage
- Edit/delete buttons in header
- Full Start All / Stop All buttons

### Shared Logic
- Calculates `allRunning` (all services running) and `allStopped` (none running)
- Only shows Start button if not all running; Stop button if not all stopped
- Disabled state during mutation

## State
- `editOpen`, `deleteOpen` — Dialog visibility
- Uses `useStartCategory()`, `useStopCategory()` mutations

## Styling Changes
- Migrated from Card to EntityCard for hover effects
- Resource bar uses `bg-green-500` (was `bg-success`)

## Dependencies
- `EntityCard` — Wrapper component
- `EditCategoryDialog`, `DeleteCategoryDialog` — Edit/delete flows
- `useStartCategory()`, `useStopCategory()` — Mutations
- `categoryLabel()` — Format category name

## Related Pages
- `/routes/index.tsx` (overview) — Compact cards
- `/routes/categories/index.tsx` — Full cards
