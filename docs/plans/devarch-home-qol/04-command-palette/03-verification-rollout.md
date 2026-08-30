# Phase 3 — Ranking, integration, and verification

Created: 2026-08-29
Purpose: Validate result quality, action safety, and parity with existing controls.

## Goal

Validate registry parity, search quality, keyboard accessibility, payload safety, and performance before enabling the palette globally.

## Scope

Visible-action refactor, ranking fixtures, safety denylist, action parity, performance, mobile/keyboard checks, documentation, and staged rollout.

## Work

- Replace duplicate card/detail action construction with registry lookups before enabling the palette.
- Test exact/prefix/token/substring ranking, diacritics, ties, favorite boosts, recent boosts, and empty query.
- Verify every palette action has an equivalent visible action or route elsewhere; the palette must not become a hidden capability surface.
- Confirm copy feedback, external opens, recency updates, and same-origin navigation behave identically from cards and palette.
- Add an explicit denylist test proving lifecycle terms/action kinds cannot become palette-eligible. In plan 04 none exist in the registry; when plan 07 later registers the sole pre-existing copied service-start action, this test must continue to exclude it from palette results.
- Document shortcut, result groups, and safety boundary.

## Outputs

- Registry parity and ranking test suites.
- Safety denylist and invalid-payload fixtures.
- Palette accessibility/performance evidence and shortcut documentation.

## Acceptance criteria

- Registry parity tests cover apps, services, containers, and global navigation.
- No unsafe scheme (`javascript:`, `file:`, arbitrary custom scheme) reaches open execution; editor links remain visible detail-page actions unless explicitly and safely modeled.
- Ranking completes within 50ms for 5,000 generated actions on target hardware.
- Palette adds no recurring timers or network calls.
- Existing mobile menu, route, filter, and refresh tests remain green.

## Verification

Run action snapshots, ranking benchmark, keyboard/focus walkthrough, mobile layout checks, popup/clipboard cases, no-network scan, and HTTPS navigation smoke tests.

## Rollout

Ship registry refactor first with unchanged visible behavior. Ship palette UI second. Enable favorites/recent ranking only after those plans are complete; otherwise use text-only ranking without provisional storage.

## Exit evidence

Action registry tests, ranking benchmark, keyboard recording/checklist, mobile screenshots, HTTPS smoke navigation, and README shortcut documentation.