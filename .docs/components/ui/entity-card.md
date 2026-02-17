# EntityCard Component

**Path:** `dashboard/src/components/ui/entity-card.tsx`
**Last Updated:** 2026-02-17

## Overview
Reusable wrapper component that standardizes card styling across entity grids (categories, stacks, projects, networks). Provides consistent border, hover effects, and optional pointer cursor.

## Props
```typescript
interface EntityCardProps extends React.ComponentProps<'div'> {
  cursor?: 'default' | 'pointer'
}
```

- `cursor` — 'pointer' for clickable cards (e.g., projects), 'default' for others (default)
- Inherits all div props: className, children, onClick, etc.

## Styling
- **Base:** `rounded-lg border bg-card text-card-foreground shadow-sm`
- **Hover:** `border-primary/50` with `transition-colors`
- **Cursor:** `cursor-pointer` when `cursor="pointer"` (toggleable)

## Usage Pattern
Replaces direct Card component in grid displays:
```tsx
<EntityCard className="py-4">
  <CardHeader>...</CardHeader>
  <CardContent>...</CardContent>
</EntityCard>
```

For clickable cards:
```tsx
<EntityCard className="py-4" cursor="pointer">
  ...
</EntityCard>
```

## Benefits
- Consistent hover feedback across all entity grids
- Single source for card styling (easier maintenance)
- Reduces boilerplate in grid components
- Supports custom className and all div props

## Current Users
- `categories/category-card.tsx` — Uses EntityCard
- `stacks/stack-grid.tsx` — Uses EntityCard
- `projects/project-card.tsx` — Uses EntityCard with cursor="pointer"

## Planned Migrations
- `networks/network-grid.tsx` — Still uses Card component directly (pending)
