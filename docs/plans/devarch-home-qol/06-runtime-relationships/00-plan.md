# Runtime relationships

Created: 2026-08-29
Status: Planned
Depends on: current Podman socket inventory; app and service identities

## Outcome

App and service pages show related running containers through allowlisted Compose/DevArch metadata with explicit confidence, replacing silent name-only guesses.

## Scope

- Read only allowlisted identity labels from Podman responses.
- Normalize Compose project/service identity without returning the raw label map.
- Resolve relationships against stable app and catalog IDs.
- Preserve a clearly marked, deterministic naming fallback where metadata is absent.
- Show relationship source/confidence on details when useful.
- Add the deferred Services `hasRuntimeMatch` facet from plan 02 only after normalized relationships are authoritative.

## Subphases

1. [Metadata allowlist and relationship contract](01-identity-contract.md)
2. [Resolver and UI integration](02-resolver-and-ui.md)
3. [Fixture coverage, migration, and verification](03-verification-rollout.md)

## Definition of done

- Known Compose/app metadata produces exact relationships.
- Arbitrary container labels remain secret and absent from API output.
- Ambiguous matches produce no exact relationship.
- Existing app/service/container inventory survives missing labels.
- No management action is attached to a relationship.

## Non-goals

Topology visualization, dependency graphs, logs, lifecycle controls, arbitrary label browsing, Compose mutation, or multi-host relationships.