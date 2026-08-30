# Phase 2 — Entity actions and responsive UI

Created: 2026-08-29
Purpose: Present useful commands without turning detail pages into operational consoles.

## Goal

Add a compact **Native commands** section to app/service details and contextual copy actions to containers.

## Scope

Native-command detail sections, contextual container copies, command-palette parity, clipboard fallback, responsive previews, and accessible labels.

## UX

- Detail pages show command label, one-line explanation, code preview, and **Copy** button.
- Group navigation/inspection commands first. Keep the existing service start command in a separate “Starts a service” area with state-changing wording; do not style it as the default action.
- Container rows expose at most one **Copy logs command** action; the command palette may expose that same bounded `inspect`-classified logs action, but no other container inspection action and no lifecycle action.
- Copied feedback identifies the command, not just “Copied.”
- On mobile, code previews wrap or scroll inside their card without forcing page overflow; actions remain 44px targets.
- If clipboard permission fails, select/show the command for manual copying. Never fall back to deprecated implicit copy APIs that obscure what text is copied.

## Integration

- When an action appears on more than one surface, cards, details, and command palette consume the same registry action ID; the palette still receives only the `paletteEligible` subset.
- Runtime relationships may surface service-level Compose commands from related containers, but never duplicate the same command.
- Existing service `command` field migrates to the descriptor contract with backward compatibility during one release slice.

## Outputs

- App/service native-command sections.
- Contextual container log-command action.
- Registry-backed copy actions and manual-copy fallback.
- Responsive command preview component.

## Acceptance criteria

- No button says “Run,” “Execute,” or implies dashboard execution.
- Lifecycle classification is visually and textually distinct.
- Keyboard users can reach previews and copy controls in logical order.
- Long commands and unusual names do not break layouts.
- Copying updates recency only if the favorites/recent plan explicitly treats command copies as recent actions; default is no recency update.

## Verification

Entity-by-entity action matrix, mobile widths, clipboard success/failure, command-palette parity, and screen-reader labels.