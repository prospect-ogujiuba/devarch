# ServiceTable Component

**Path:** `dashboard/src/components/services/service-table.tsx`
**Last Updated:** 2026-02-17

## Overview
Table display of services with checkbox selection, filtering, and action buttons. Shows name, category, image, ports, CPU/memory metrics, status, and controls.

## Props
```typescript
interface ServiceTableProps {
  services: Service[]
  selected: Set<string>
  onToggleSelect: (name: string) => void
}
```

- `services` — Array of service objects
- `selected` — Set of selected service names
- `onToggleSelect` — Callback to toggle selection state

## Columns (Left to Right)

1. **Checkbox** — Select individual or all services
   - Header checkbox toggles all visible services
   - Individual checkboxes toggle per-service state

2. **Name** — Service name (linked to detail page /services/$name)

3. **Category** — Category badge with label

4. **Image** — `{image_name}:{image_tag}` format
   - Handles null/empty image_name or image_tag
   - Shows "—" if both are empty
   - Max width 200px with truncation

5. **Ports** — Comma-separated host:container pairs
   - Shows "—" if no ports

6. **CPU** — Resource bar (only if running and metrics > 0)
   - Shows "—" otherwise
   - Width: 80px

7. **Memory** — Resource bar (only if running and metrics > 0)
   - Shows "—" otherwise
   - Width: 80px

8. **Status** — Status badge (running/stopped/created/exited)
   - Uses `StatusBadge` component
   - Color-coded based on service state

9. **Actions** — Action button (play/stop/logs)
   - Right-aligned
   - Uses `ActionButton` component with service name and status

## Selection Behavior
- Header checkbox: toggle all visible services at once
- If all selected: clicking header deselects all
- If some/none selected: clicking header selects all
- Individual checkboxes: toggle single service
- Selection state is a Set for O(1) membership testing

## Empty State
- Shows "No services found" message when services array is empty
- Spans all columns

## Dependencies
- `Link` — React Router navigation
- `Badge` — Category badge
- `ResourceBar` — Visual metric display
- `StatusBadge`, `ActionButton` — Service-specific UI
- `getServiceStatus()` — Derive status from service object
- `titleCase()`, `categoryLabel()` — Formatting utilities

## Styling
- Row hover: cursor-pointer style
- Truncated image name
- Muted foreground for secondary columns
- Checkbox styling with border-muted-foreground

## Recent Changes
- Fixed image display: now correctly handles cases where image_name or image_tag is null/empty
- Removed empty state footer (no "Showing N services" counter)

## Related Components
- `/routes/services/index.tsx` — Parent page
- `/routes/index.tsx` (overview) — May use similar table patterns
