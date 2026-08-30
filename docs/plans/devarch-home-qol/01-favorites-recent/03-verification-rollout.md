# Phase 3 — Hardening, migration, and verification

Created: 2026-08-29
Purpose: Prove preference behavior remains local, bounded, accessible, and recoverable.

## Goal

Complete test coverage, documentation, and safe rollout of the first client-persisted state.

## Scope

Storage failure modes, bounds, migration behavior, clear-history UX, privacy documentation, regression checks, and rollback.

## Work

- Add regression fixtures for malformed JSON, duplicate IDs, stale items, unsupported versions, quota errors, and storage-disabled browsers.
- Verify recency timestamps use UTC ISO strings and ordering handles equal timestamps deterministically.
- Verify no inventory refresh occurs when pinning, unpinning, or opening the dashboard sections.
- Document the storage key, retained fields, clearing procedure, and privacy boundary.
- Add a small **Clear dashboard history** preference action only if recent data cannot otherwise be removed; it must not clear favorites without separate confirmation.
- Confirm future schema migration has an explicit version switch rather than implicit shape guessing.

## Outputs

- Complete preference-store and UI regression suite.
- Storage/privacy documentation and clear-history behavior.
- Rollout and rollback checklist.

## Acceptance criteria

- Focused tests pass for every storage failure mode.
- Existing routes, search, refresh, and mobile navigation remain unchanged.
- Lighthouse/accessibility inspection reports no new control-name or contrast failures.
- Stored state remains under 10KB with maximum supported entries.
- Clearing recents is immediate and does not affect inventory or favorites.

## Verification

Run storage fixtures, reload/persistence scenarios, keyboard and mobile walkthroughs, storage-size inspection, and no-network assertions.

## Rollout

Ship as a browser-local feature with no server migration. If defects occur, the feature can be disabled by ignoring/removing the single versioned key without affecting dashboard inventory.

## Exit evidence

Test results, storage-size check, keyboard walkthrough, mobile screenshots, and updated `scripts/dashboard/README.md`.