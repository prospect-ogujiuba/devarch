# Phase 3 — Fixture coverage, migration, and verification

Created: 2026-08-29
Purpose: Prove relationships are accurate, secret-safe, and compatible with current running stacks.

## Goal

Demonstrate relationship accuracy, ambiguity safety, secret exclusion, and compatibility with current live stacks before replacing heuristics.

## Scope

Sanitized fixtures, secret-decoy audit, live parity review, convention documentation, bounds, staged backend/frontend migration, and rollback.

## Work

- Capture sanitized fixture shapes from representative Podman API versions, structured CLI fallback, single-service, multi-service, app runtime, unlabeled, conflicting-label, and duplicate-name containers.
- Add negative fixtures containing secrets in arbitrary labels and assert they never appear in serialized inventory.
- Compare new resolver output with current name heuristic for the live inventory; manually review differences before removing frontend fallback code.
- Document conventions and how future bootstraps may add explicit `devarch.app`/`devarch.service` labels.
- Verify service runtime filters and app/service detail counts use the same resolved data.
- Bound label value length and relationship count per container.

## Outputs

- Representative Podman/Compose fixture corpus.
- Serialized-response secret audit and live parity report.
- Relationship convention documentation and frontend migration checklist.

## Acceptance criteria

- All exact links are explainable by an allowlisted source.
- No ambiguous target is selected.
- Secret-decoy tests inspect the complete serialized response, not only intermediate objects.
- Live inventory retains expected app/service relationships or records reviewed intentional differences.
- Missing Podman and API fallback behavior remains graceful.
- Existing container fields and URLs remain backward compatible.

## Verification

Run source/precedence fixtures, complete-response secret checks, live inventory parity comparison, service runtime filters, detail counts, fallback paths, and responsive chips.

## Rollout

Ship additive backend fields first while retaining the frontend heuristic. Compare/debug in development. Switch the frontend to normalized relationships only after fixture and live parity review; retain convention confidence from the backend as the deliberate fallback.

## Exit evidence

Fixture suite, serialized-response secret audit, live parity report, service-filter tests, detail-page screenshots, and updated API/convention documentation.