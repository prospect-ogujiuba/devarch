# EntityCard Component

**Path:** `dashboard/src/components/ui/entity-card.tsx`
**Last Updated:** 2026-02-17

## Overview
Reusable wrapper component that standardizes card styling across entity grids (categories, stacks, projects, networks). Provides consistent border hover effect and optional pointer cursor.

## Props
```typescript
interface EntityCardProps extends React.ComponentProps<'div'> {
  cursor?: 'default' | 'pointer'
}
```

- `cursor` — 'pointer' for clickable cards (projects), 'default' for others
- Inherits div props (className, children, etc.)

## Styling
- Base: `rounded-lg border bg-card text-card-foreground shadow-sm`
- Hover: `border-primary/50` with `transition-colors`
- Optional: `cursor-pointer` when `cursor="pointer"`

## Usage Pattern
Replaces direct Card component in grid views:
```tsx
<EntityCard className="py-4" cursor="pointer">
  <CardHeader>...</CardHeader>
  <CardContent>...</CardContent>
</EntityCard>
```

## Benefits
- Consistent hover feedback across all entity cards
- Single source for card styling (easier to maintain)
- Reduces boilerplate in grid components

## Related Components
- `categories/category-card.tsx` — Uses EntityCard
- `stacks/stack-grid.tsx` — Uses EntityCard
- `projects/project-card.tsx` — Uses EntityCard with cursor="pointer"
- `networks/network-grid.tsx` — Still uses Card (planned migration)
