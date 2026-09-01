# 01-modernize-core-blocks

Created: 2026-08-22T01:09:57.478Z
Purpose: Define the existing-block modernization slice

# Modernize core MakerBlocks

Created: 2026-04-01
Purpose: Give the five existing blocks a cohesive modern baseline and stronger conversion capabilities.

## Goal

Upgrade Hero, Card, Card Grid, Icon, and Section visuals and editor behavior while preserving their role as reusable primitives.

## Scope

- `apps/playground/wp-content/plugins/makerblocks/src/blocks/{hero,card,card-grid,icon,section}`
- Add optional Hero primary/secondary actions.
- Add optional Card icon/eyebrow treatment.
- Improve responsive spacing, typography, depth, focus states, and editor parity.

## Outputs

Updated metadata, edit/save implementations, Sass, and compiled `build/` assets.

## Acceptance criteria

- Existing blocks render as a cohesive modern system.
- Hero and Card support conversion-oriented optional content.
- Interactive links have visible keyboard focus.
- Layouts respond cleanly on narrow screens.

## Verification

`npm run build`, `npm run lint`, and `npm test` in MakerBlocks.

## Non-goals

No dynamic frontend runtime, analytics, forms, or project-specific branding.
