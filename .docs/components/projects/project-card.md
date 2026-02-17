# ProjectCard Component

**Path:** `dashboard/src/components/projects/project-card.tsx`
**Last Updated:** 2026-02-17

## Overview
Clickable card showing project metadata: name, type, version, description, framework/language, domain, git branch, dependencies, plugins, themes, and proxy port. Used in projects grid.

## Props
```typescript
interface { project: Project }
```

- `project` — Project object with metadata, dependencies, framework info

## Card Layout

### Header
- **Left:** Project logo (from type/framework/language) + project name
- **Right:** Badges (running count, version, project type)
  - Running count badge: green bg, shown if running_count > 0
  - Version badge: secondary variant, shown if version exists
  - Project type badge: color-coded by typeColors map

### Content
- **Description:** Text with `text-muted-foreground/50`, line clamped to 2 lines
  - Fallback italic "No description" if missing
- **Framework/Language Badges:** Secondary variant, small text
  - Shows: framework, language, frontend_framework (if has_frontend=true)
- **Metadata Row:** Flex wrap, small text, muted foreground
  - Domain (if exists) — Globe icon
  - Git branch (if exists) — GitBranch icon
  - Dependencies count (if > 0) — Package icon
  - Plugins count (if WordPress and > 0) — Puzzle icon
  - Themes count (if WordPress and > 0) — Palette icon
  - Proxy port (if exists) — ExternalLink icon

## Type Colors Map
- laravel → red
- wordpress → blue
- node → green
- go → cyan
- rust → orange
- python → yellow
- php → purple

## Dependencies Counter
- Counts items in:
  - `dependencies.plugins` (array)
  - `dependencies.themes` (array)
  - Other fields in dependencies object
- Helper function `countDeps()` iterates and counts recursively

## Styling
- Uses `EntityCard` with `cursor="pointer"` for clickable styling
- Full height (`h-full`) for grid alignment
- Grid gap and responsive columns handled by parent

## Link Behavior
- Wrapped in Link to `/projects/$name`
- Cursor styling via EntityCard cursor="pointer"
- Entire card is clickable

## Recent Changes
- Migrated from Card to EntityCard with cursor="pointer"
- Description text color: `text-muted-foreground/50` (lighter)
- Always renders description with fallback placeholder

## Dependencies
- `EntityCard` — Card wrapper with pointer cursor
- `CardContent`, `CardHeader`, `CardTitle` — Card structure
- `Badge` — Framework/language/type badges
- `ProjectLogo` — Logo component
- `Link` — React Router navigation
- Icons: Globe, GitBranch, Package, ExternalLink, Puzzle, Palette

## Related Components
- `/routes/projects/index.tsx` — Parent grid page
- `projects/project-logo.tsx` — Logo rendering
