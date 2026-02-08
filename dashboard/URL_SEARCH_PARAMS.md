# URL Search Params Convention

Use URL search params for list state and tab state so pages are shareable and back/forward works.

## Core keys

- `q`: text search
- `sort`: active sort field
- `dir`: sort direction (`asc` or `desc`)
- `view`: list view mode (`grid` or `table`)
- `page`: current list page (>= 1)
- `size`: items per page

Only keep non-default values in URL.

- default `page` is `1` (omit from URL)
- default `size` is `24` (omit from URL)
- allowed `size` values: `12`, `24`, `50`, `100`, `200`

## Filter keys

Use explicit, domain-named keys per page.

- Overview/Categories: `status`
- Services: `status`, `category`
- Projects: `type`, `language`
- Stacks: no extra filter key today

Avoid generic keys like `filter1`.

## Tab keys

Use `tab` when the route is the only tab owner.

- `services/$name`: `tab`
- `projects/$name`: `tab`
- `stacks/$name`: `tab`

If a child route can inherit parent search params and has its own tabs, use a namespaced key.

- `stacks/$name/instances/$instance`: `instanceTab`

Rule: one tab key per route level, and child key must not collide with parent key.

## Validation

Each route with search params should define `validateSearch` with Zod.

- use `z.enum([...]).optional()` for bounded values (`sort`, `dir`, `view`, tab keys)
- use `z.string().optional()` for free text (`q`, dynamic filters)

## Sync pattern

For list pages, keep bidirectional sync:

1. URL -> state effect (hydrates controls)
2. state -> URL effect (writes normalized params)

Use a ref guard to skip one write cycle immediately after URL hydration.

## New page checklist

1. Add `validateSearch`
2. Read `Route.useSearch()`
3. Hydrate local controls from search
4. Write controls back to search with defaults omitted
5. Add collision-safe tab key if nested tab routes exist
