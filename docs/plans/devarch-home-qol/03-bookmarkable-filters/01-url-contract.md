# Phase 1 — Canonical URL schema

Created: 2026-08-29
Purpose: Define a human-readable, stable query contract before router integration.

## Goal

Serialize only meaningful collection view state with deterministic ordering and safe validation.

## Scope

Query parameter names, validation, canonical ordering, parse/serialize functions, limits, compatibility rules, and contract tests. No router or control integration.

## Schema

Common parameters:

- `q=<text>` — trimmed search text.
- `sort=<allowed-key>` — omitted when default.
- `dir=desc` — omitted for ascending/default.

Collection facets:

- Apps: repeated `kind=<value>`.
- Containers: repeated `state=<value>` and `ports=published`.
- Services: repeated `category=<value>` and `runtime=matched`.

Rules:

- Use `URLSearchParams`; never manually concatenate or decode.
- Sort repeated values and parameter keys for one canonical URL.
- Omit default, empty, malformed, unknown-key, and duplicate values. Validate enum facets (`ports`, `runtime`, sort, direction) against their route allowlists. Validate dynamic inventory facets (`kind`, `state`, `category`) by bounded printable-string syntax only—not membership in the current inventory—so a syntactically valid selected value survives with a zero count.
- Reject control characters and unpaired surrogates. Cap search length at 200 Unicode code points, each repeated facet at 50 values, each facet value at 100 code points, and the canonical query string at 2,048 UTF-8 bytes. Sort first, then discard the final values deterministically until the byte cap is met, and expose a non-blocking reset notice.
- Preserve no unrelated parameters because these routes are owned entirely by the dashboard.

## Outputs

- Pure `parseCollectionQuery(route, searchParams, facetDefinitions)` and `serializeCollectionQuery(viewState)` functions; `facetDefinitions` contains route keys, enum allowlists, and syntax validators, never current inventory values.
- Canonicalization decision: replace malformed/noncanonical URLs after parse without creating history entries.
- Table in dashboard README documenting parameters and examples.

## Acceptance criteria

- Parse/serialize round trips are stable.
- Reserved characters and Unicode round trip through native URL encoding.
- Invalid values cannot become HTML or command payloads.
- Default apps/services/containers URLs remain `/apps`, `/services`, and `/containers`.

## Verification

Fixture matrix for defaults, repeated values, Unicode, duplicates, invalid sort/facets, excessive lengths, and canonical ordering.