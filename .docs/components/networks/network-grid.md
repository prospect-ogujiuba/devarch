# NetworkGrid Component

**Path:** `dashboard/src/components/networks/network-grid.tsx`
**Last Updated:** 2026-02-17

## Overview
Grid of network cards with checkbox selection, status badges (external, orphaned), stack link, driver/container info, and remove button. Supports multi-select for bulk operations.

## Props
```typescript
interface NetworkGridProps {
  networks: NetworkInfo[]
  selected: Set<string>
  onToggleSelect: (name: string) => void
  onRemove: (name: string) => void
}
```

- `networks` — Array of NetworkInfo objects
- `selected` — Set of selected network names
- `onToggleSelect` — Callback to toggle selection
- `onRemove` — Callback to remove network

## Card Layout

### Header
- **Left:** Checkbox (selection toggle, stops propagation)
- **Center:** Network name (monospace font, truncated, flex-1 to take space)
- **Right:** Status badges
  - External badge: shown if !managed (Shield icon + "External")
  - Orphaned badge: shown if orphaned (AlertTriangle icon + "Orphaned", yellow color)

### Content
- **Stack Link:**
  - If orphaned: plain text (muted)
  - If not orphaned and stack_name exists: blue link to `/stacks/$name`
  - If no stack: "No stack" (muted)

- **Driver & Container Count:**
  - Driver name (e.g., "bridge", "overlay")
  - Container count with singular/plural suffix

- **Created Date:**
  - Formatted date (if exists and not zero-date)
  - Text size: xs, muted foreground

### Footer
- **Remove Button:**
  - Full-width outline button
  - Disabled if:
    - network has connected containers (container_count > 0)
    - network is default "podman" network
  - Title hints:
    - "Default network" — if podman
    - "Has connected containers" — if containers > 0
    - "Remove network" — otherwise

## Selection Behavior
- Checkbox click: stops propagation (prevents card triggering parent events)
- Checkbox state reflects `selected` Set membership
- Clicking checkbox calls `onToggleSelect(net.name)`

## Empty State
- Shows `EmptyState` with Network icon and message: "No networks match your filters"

## Styling
- Hover effect: `border-primary/50 transition-colors` (direct className, not EntityCard)
- Card base: `py-4` padding
- Badges: Contextual colors (yellow for orphaned)
- Stack link: blue-500 with underline on hover

## Grid Layout
- Responsive: 1-4 columns (sm, lg, xl breakpoints)
- Gap: 4 units
- Grid: `sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4`

## Dependencies
- `Card`, `CardContent`, `CardFooter`, `CardHeader`, `CardTitle` — Radix UI card primitives
- `Badge`, `Button`, `Checkbox` — UI components
- `Link` — React Router navigation
- `EmptyState` — Empty state display
- Icons: Network, Trash2, AlertTriangle, Shield

## Recent Changes
- Added checkbox selection support (new props: selected, onToggleSelect)
- Now supports bulk removal workflows (via parent page)

## Planned Migrations
- Migrate from Card to EntityCard for consistency with other grids

## Related Components
- `/routes/networks/index.tsx` — Parent page
- `/components/networks/network-table.tsx` — Table view variant
