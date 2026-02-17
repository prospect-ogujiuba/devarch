# DevArch Dashboard Documentation Index

> Auto-generated 2026-02-17 | 12 entries

## Components

### Layout
- [Header](components/layout/header.md) — Sticky top nav with logo, route links, theme toggle, mobile menu. Active link styling with accent foreground.

### UI
- [EntityCard](components/ui/entity-card.md) — Reusable card wrapper with consistent hover border effect. Base primitive for all entity cards.

### Categories
- [CategoryCard](components/categories/category-card.md) — Category display card with service count progress bar. Uses EntityCard.

### Networks
- [NetworkGrid](components/networks/network-grid.md) — Grid of network cards with checkbox selection, status badges, stack links, and remove actions.

### Projects
- [ProjectCard](components/projects/project-card.md) — Project display card with type badge, service count, description. Uses EntityCard.

### Services
- [ServiceTable](components/services/service-table.md) — Sortable service table with image name:tag display, status badges, and action buttons.

### Stacks
- [StackGrid](components/stacks/stack-grid.md) — Stack card grid with start/stop/delete/network actions, service count, and description. Uses EntityCard.

## Routes

- [Dashboard Home](routes/index.md) — Overview page with stat cards, runtime info bar, top CPU services, and category grid.
- [Categories](routes/categories.md) — Categories list page with search, sort, grid/table view. Default sort by name.
- [Images](routes/images.md) — Container images page with search, sort, grid/table toggle, pull and prune dialogs.
- [Networks](routes/networks.md) — Networks list with multi-select, bulk remove, status filter, create dialog.
- [Settings](routes/settings.md) — Configuration page with runtime switcher, socket management, API key settings.
