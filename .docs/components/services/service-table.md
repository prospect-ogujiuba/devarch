# ServiceTable Component

**Path:** `dashboard/src/components/services/service-table.tsx`
**Last Updated:** 2026-02-17

## Overview
Table display of services with filtering, selection, and action buttons. Shows name, category, image, ports, CPU/memory metrics, status, and controls. Used in services list and may appear on dashboard.

## Props
```typescript
interface ServiceTableProps {
  services: Service[]
  selected: Set<string>
  onToggleSelect: (name: string) => void
}
```

## Columns
- **Checkbox** — Select individual or all services
- **Name** — Linked to detail page
- **Category** — Badge with category label
- **Image** — `{image_name}:{image_tag}` format (fixed: handles partial names)
- **Ports** — Comma-separated host:container pairs
- **CPU** — Resource bar if running with metrics, else dash
- **Memory** — Resource bar if running with metrics, else dash
- **Status** — Status badge (running, stopped, etc.)
- **Actions** — Action button (play/stop/logs)

## Recent Changes
- Fixed image name display: now handles `image_name` or `image_tag` being null/empty
  - Format: `image_name:tag`, `image_name` (if no tag), or `—` (if both empty)
- Removed "Showing N services" counter at bottom

## Selection Behavior
- Header checkbox toggles all visible services
- Individual checkboxes toggle per-service
- `selected` is a Set for O(1) membership

## Dependencies
- `Link` — React Router navigation
- `Badge`, `ResourceBar` — UI components
- `StatusBadge`, `ActionButton` — Service-specific controls
- `getServiceStatus()` — Derive status from service state

## Related Components
- `/routes/services/index.tsx` — Parent page
- `/routes/index.tsx` (overview) — May use similar table layout
