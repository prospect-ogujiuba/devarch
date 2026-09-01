# 03-ship-starter-landing-page

Created: 2026-08-22T01:10:12.219Z
Purpose: Define the initial full landing page composition

# Ship the starter landing page

Created: 2026-04-01
Purpose: Ensure the Playground theme opens with a complete, editable conversion-oriented home page.

## Goal

Expand the MakerStarter landing pattern into a full narrative and use it as the Playground child theme front page.

## Scope

- Update `makerstarter/patterns/makerblocks-landing.php`.
- Add `playground-theme/templates/front-page.html` referencing the parent pattern.
- Apply project-owned theme tokens in `playground-theme/theme.json` only when needed for the composition.

## Outputs

A complete page flow: hero, trust/logo proof, benefits, metrics, testimonial, pricing, FAQ, and final CTA.

## Acceptance criteria

- The front-page template starts with the complete pattern.
- Copy is realistic starter copy and remains editable.
- Section order supports a clear conversion narrative.
- Markup references only registered MakerBlocks/core blocks.
- The page remains usable on mobile.

## Verification

MakerStarter tests, PHP syntax check, front-page template reference inspection, and MakerBlocks manifest block-name inspection.

## Open questions

Live browser visual review depends on the local WordPress runtime being available.
