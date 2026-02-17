# StackGrid Component

**Path:** `dashboard/src/components/stacks/stack-grid.tsx`
**Last Updated:** 2026-02-17

## Overview
Grid of stack cards showing instance count, running status, description, and network state. Provides enable/disable, network create/remove, and delete actions per stack.

## Props
```typescript
interface StackGridProps {
  stacks: Stack[]
  onEnable: (name: string) => void
  onDisable: (name: string) => void
  onDelete: (name: string) => void
  onCreateNetwork: (name: string) => void
  onRemoveNetwork: (name: string) => void
}
```

All callbacks receive stack name as string parameter.

## Card Layout

### Header
- Stack name (blue link to stack detail page)
- Enabled/Disabled badge (right side)

### Content
- Optional description (conditional render, line clamped to 2 lines)
- Instance count and running status (e.g., "5 instances" + "3/5 running")
- Running status color:
  - Green if all running
  - Yellow if partial (some running)
  - Muted gray if none running
- Network name display (if exists) with blue text and globe icon

### Footer Actions
- **Enable/Disable:** Toggles based on stack.enabled state
- **Divider:** Separates primary from secondary actions
- **Create/Remove Network:** Toggle buttons based on network_active state
- **Delete:** Remove stack

## Button States
- Create network: disabled if network_active=true
- Remove network: disabled if network_active=false
- Delete: always enabled
- Buttons show hover titles for clarity

## Empty State
- Shows empty state icon/message when stacks.length === 0

## Styling
- Uses `EntityCard` for consistent hover effect
- Responsive grid: 1-4 columns (sm, lg, xl breakpoints)
- Divider separator uses border styling

## Dependencies
- `EntityCard` — Card wrapper
- `Badge`, `Button` — UI primitives
- `Link` — React Router navigation to /stacks/$name
- `cn()` — Classname utility
- Icons: Layers, Trash2, Power, PowerOff, Globe, Network, Unplug

## Recent Changes
- Migrated from Card to EntityCard for consistent hover effect
- Removed inline hover styling (now in EntityCard)

## Related Components
- `/routes/stacks/index.tsx` — Parent page using StackGrid
