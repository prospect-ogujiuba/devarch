# 02-add-conversion-blocks

Created: 2026-08-22T01:10:04.343Z
Purpose: Define reusable landing-page section blocks

# Add conversion blocks

Created: 2026-04-01
Purpose: Cover the sections most often needed in a high-converting landing page.

## Goal

Add reusable, accessible section blocks that editors can customize with native controls.

## Scope

Create six static or InnerBlocks-based MakerBlocks: Logo Cloud, Stats, Testimonial, Pricing, FAQ, and Call to Action.

## Outputs

Each block has `block.json`, editor/save implementation, editor Sass, frontend Sass, and compiled assets.

## Acceptance criteria

- Blocks are discoverable in the Design category.
- Default inserter states provide useful editable content structures.
- FAQ uses semantic native details/summary markup.
- Pricing and CTA expose clear action links.
- Sections are responsive and inherit theme tokens.

## Verification

MakerBlocks build, lint, tests, and generated manifest inspection.

## Non-goals

Payment processing, form submission, carousel behavior, external icon/font packages, or third-party testimonials.
