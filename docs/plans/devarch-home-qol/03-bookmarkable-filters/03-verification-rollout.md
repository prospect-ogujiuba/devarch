# Phase 3 — Compatibility and verification

Created: 2026-08-29
Purpose: Protect existing routes and make URL behavior predictable across releases.

## Goal

Guarantee shareable collection URLs remain canonical, backward-compatible, safe, and faithful through reload and browser history.

## Scope

Route fallback, parse/serialize edge cases, history sequences, aliases, privacy review, compatibility policy, documentation, and rollout.

## Work

- Add route tests proving supported collection paths with query strings still return the HTML shell.
- Add pure parse/serialize tests and browser-level history scenarios.
- Verify malformed percent encoding, unsupported keys, oversized queries, zero-count selected values, and temporarily missing inventory degrade deterministically without exceptions or destructive bookmark rewrites.
- Confirm detail pages ignore collection-only query parameters rather than accidentally showing filters.
- Document URL compatibility policy: existing parameter meanings remain stable; new facets are additive; removed facets are ignored.
- Verify copied URLs never include localStorage data, absolute paths, container IDs unless explicitly represented by a future route, or secrets.

## Outputs

- URL contract and browser-history regression suite.
- Compatibility policy and README parameter examples.
- Per-collection rollout checklist.

## Acceptance criteria

- URL state survives reload and manual inventory refresh.
- Browser Back/Forward has no duplicate or skipped semantic states.
- Default state URL remains clean after clear/reset.
- Existing deep links and `/projects` aliases continue working.
- Query operations trigger zero inventory requests.

## Verification

Run query fixtures, direct HTTPS loads, Back/Forward sequences, refresh reapplication, clipboard inspection, malformed-query cases, and no-fetch assertions.

## Rollout

Enable per collection in the same order used for filters. Do not ship partial parse-only behavior: each route must parse, render, update, navigate, and canonicalize as one slice.

## Exit evidence

URL fixture tests, recorded history walkthrough, direct HTTPS route smoke tests, clipboard inspection, and README examples.