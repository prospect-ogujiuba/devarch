# NetworkGrid Component

**Path:** `dashboard/src/components/networks/network-grid.tsx`
**Last Updated:** 2026-02-17

## Overview
Grid of network cards with checkbox selection, status badges (external, orphaned), stack link, driver/container count, and remove button. Supports multi-select for bulk operations.

## Props
```typescript
interface NetworkGridProps {
  networks: NetworkInfo[]
  selected: Set<string>
  onToggleSelect: (name: string) => void
  onRemove: (name: string) => void
}
```

## Card Layout
- **Header:**
  - Checkbox for selection
  - Network name (monospace font, truncated)
  - Badges: External (if not managed), Orphaned (if orphaned + yellow)

- **Content:**
  - Stack link or name (gray if orphaned)
  - Driver and container count
  - Created date (if available and not zero-date)

- **Footer:**
  - Remove button (disabled if containers > 0 or network is 'podman')
  - Full-width button with title hints

## Selection Behavior
- Checkbox click stops propagation (prevents card click)
- Checkbox state synced with `selected` Set

## Styling
- Hover effect on card border via direct className (not EntityCard yet)
- External/Orphaned badges styled with contextual colors

## Disable States
- Remove disabled if:
  - Network has connected containers (container_count > 0)
  - Network is default 'podman' network
  - Shows appropriate title hint in both cases

## Dependencies
- `Card`, `CardContent`, `CardFooter`, `CardHeader`, `CardTitle` — Radix UI
- `Badge`, `Button`, `Checkbox` — UI components
- `Link` — React Router for stack navigation

## Related Components
- `/routes/networks/index.tsx` — Parent page
- `/components/networks/network-table.tsx` — Table view variant
