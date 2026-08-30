# Phase 2 — Accessible palette interaction

Created: 2026-08-29
Purpose: Deliver fast keyboard and touch navigation without a heavy UI framework.

## Goal

Implement an accessible modal palette using native/standards-based semantics and the action registry.

## Scope

Dialog markup, global/mobile entry points, focus lifecycle, keyboard navigation, grouped results, execution dispatch, and responsive states.

## UX

- Open with `Ctrl+K` and `Cmd+K`; add **Command palette** to mobile navigation.
- Use modal `<dialog>.showModal()` for the defined evergreen-browser baseline. Do not add a fallback preemptively; if the verified target matrix lacks required behavior, return to this phase contract before implementing an alternative.
- Focus the search input on open; Arrow Up/Down moves active result; Enter executes; Escape closes.
- Trap focus inside the dialog and restore focus to the opener on close.
- Group results visually and announce total/result changes with a debounced/polite live region; do not announce on every keystroke when the meaningful result summary is unchanged.
- Show action type and keyboard hint; do not hide safety-relevant labels behind icons.
- On mobile, use an inset full-width sheet with safe-area padding and large touch targets. Honor reduced-motion preferences and keep functionality independent of transition completion.

## State behavior

Palette query is ephemeral and never enters route URLs or localStorage. Inventory refresh while closed rebuilds actions. Manual refresh while open closes the palette before applying new inventory to avoid stale result execution.

## Outputs

- Dialog markup, Tailwind styles, focus/keyboard controller, and result renderer.
- One execution dispatcher for navigate/open/copy.
- Mobile menu integration and discoverable shortcut hint.

## Acceptance criteria

- Keyboard-only users can open, search, inspect, execute, and close.
- Focus never escapes an open palette.
- Empty/no-match/loading states are distinct.
- Opening the palette does not change browser history.
- Executed navigation closes the mobile menu and palette consistently.

## Verification

Keyboard matrix, screen-reader role/name/state inspection, 320px layout, reduced-motion mode, and external popup-block behavior.