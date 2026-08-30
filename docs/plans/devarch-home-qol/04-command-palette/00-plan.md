# Command palette

Created: 2026-08-29
Status: Planned
Depends on: favorites/recent items; collection routes; shared action registry

## Outcome

A keyboard-first `Ctrl/Cmd+K` palette searches the current inventory and executes safe navigation, open, and copy actions without introducing management controls.

## Scope

- Shared action registry consumed by cards, detail pages, and palette.
- Palette dialog with grouped results for apps, services, containers, and navigation.
- Deterministic fuzzy-lite ranking, favorites boost, and recent boost.
- Safe actions: navigate, open URL/port, copy path/ID/native command.

## Subphases

1. [Action registry and search contract](01-action-search-contract.md)
2. [Accessible palette interaction](02-dialog-interaction.md)
3. [Ranking, integration, and verification](03-verification-rollout.md)

## Definition of done

- Palette opens globally with keyboard and mobile menu entry.
- Results and actions are predictable, accessible, and shared with existing UI.
- No action mutates containers, files, services, or inventory.
- Opening/closing/searching the palette causes no API request.

## Non-goals

Shell command execution, arbitrary user commands, plugins, nested command workflows, natural-language interpretation, or remote search.