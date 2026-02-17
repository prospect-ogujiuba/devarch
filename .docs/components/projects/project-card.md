# ProjectCard Component

**Path:** `dashboard/src/components/projects/project-card.tsx`
**Last Updated:** 2026-02-17

## Overview
Clickable card showing project metadata: name, type, version, description, framework/language, domain, git branch, dependencies (plugins, themes), and proxy port. Used in projects grid.

## Props
```typescript
export function ProjectCard({ project }: { project: Project })
```

## Card Layout
- **Header:**
  - Project logo (from project type/framework/language)
  - Project name
  - Running count badge (if > 0)
  - Version badge
  - Project type badge (colored by `typeColors` map)

- **Content:**
  - Description (italic "No description" placeholder if missing)
  - Framework/language/frontend framework badges
  - Metadata row: domain, git branch, dependencies, plugins, themes, proxy port

## Styling
- Uses `EntityCard` with `cursor="pointer"` for clickable styling
- Full height (`h-full`) for grid alignment
- Description text uses `text-muted-foreground/50` with line clamping

## Type Colors
- laravel → red, wordpress → blue, node → green, go → cyan, rust → orange, python → yellow, php → purple

## Dependencies Counter
- Counts items in `dependencies.plugins`, `dependencies.themes`, and other dep fields
- Only shown if depCount > 0

## Recent Changes
- Migrated from Card to EntityCard with cursor="pointer"
- Description now always renders with fallback placeholder (not conditional)
- Description text color changed to `text-muted-foreground/50`
- Removed conditional rendering of project type badge (now always renders if colorClass exists)

## Related Components
- `/routes/projects/index.tsx` — Parent grid page
- `projects/project-logo.tsx` — Logo component
