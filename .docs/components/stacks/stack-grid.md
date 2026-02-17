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

## Card Layout
- **Header:** Stack name (link), enabled/disabled badge
- **Content:**
  - Optional description (conditional render)
  - Instance count with status color (green if all running, yellow if partial, muted if none)
  - Network name display if network_name exists (blue text)
- **Footer:** Action buttons
  - Enable/Disable toggle
  - Divider separator
  - Create/Remove network buttons (toggle based on network_active)
  - Delete button

## UX Improvements
- Conditional description render (no "No description" text)
- Network buttons show hover titles
- Delete button title on hover
- Divider separates primary action (enable/disable) from secondary (network, delete)

## Styling
- Migrated from Card to EntityCard for consistent hover effect
- Removed inline `hover:border-primary/50` (now in EntityCard)

## Dependencies
- `EntityCard` — Card wrapper
- `Badge`, `Button` — UI primitives
- `cn()` — Classname utility

## Related Components
- `/routes/stacks/index.tsx` — Parent page using StackGrid
