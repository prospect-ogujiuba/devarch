# Phase 1 — Metadata allowlist and relationship contract

Created: 2026-08-29
Purpose: Define a secret-safe identity boundary before consuming container labels.

## Goal

Extract only identity metadata required for discovery and represent resolved links with source and confidence.

## Scope

Explicit label allowlist, normalized Compose identity, relationship source/confidence, precedence, ambiguity behavior, limits, and fixture contracts. No UI integration.

## Allowed raw labels

- `io.podman.compose.project`
- `io.podman.compose.service`
- `com.docker.compose.project`
- `com.docker.compose.service`
- Future explicit `devarch.app` and `devarch.service` only after their producer contract is documented and tested. They are not accepted in the first release merely because the resolver understands the concept.

Never return the raw `Labels` object. Do not allow prefix wildcards. Explicitly exclude working-directory, config-file, environment, orchestration, and third-party labels.

## Normalized container fields

```json
{
  "compose": {"project": "postgres", "service": "postgres"},
  "relationships": [
    {"kind": "service", "id": "database/postgres", "source": "compose-project", "confidence": "exact"}
  ]
}
```

Schema-reserved sources are `devarch-label`, `compose-project`, `compose-service`, and `name-convention`; `devarch-label` is not emitted in the first release because those labels are not yet allowlisted. Allowed confidence is `exact` or `convention`. Relationship targets use existing stable app/service IDs.

## Resolution precedence

For the first release, resolution starts at Compose project below. The explicit-label step becomes active only in a later schema/version after the `devarch.app`/`devarch.service` producer contract and allowlist are documented and tested:

1. When enabled in that later version, explicit DevArch ID label with exact existing target.
2. Compose project matching one unique catalog short name or app runtime convention.
3. Compose service matching one unique target when project is insufficient.
4. Documented container naming convention (`node-<app>`, exact catalog name) as convention confidence.
5. Otherwise no relationship.

Ambiguity never falls through to a guessed exact result.

## Outputs

- Allowlist extraction and normalization helpers.
- Relationship schema and precedence table.
- Positive, ambiguous, malformed, and sensitive-label fixtures.

## Acceptance criteria

- Allowlist extraction is tested against sensitive decoy labels.
- Identity normalization rejects empty, oversized, control-character, and non-string values.
- Resolved IDs always exist in current inventory.
- Relationship ordering is deterministic and duplicate-free, with at most 8 resolved targets per container and bounded normalized label values.
- Conflicting Podman/Docker Compose keys follow a documented Podman-first precedence and produce a test-visible capability warning rather than two contradictory exact links.

## Verification

Unit fixtures for every allowed source and precedence branch; serialized-response assertions proving arbitrary labels and decoy secrets are absent.