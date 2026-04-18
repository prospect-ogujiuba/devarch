# ADR 0001 — V2 manifest-first desired state

- Status: Accepted
- Date: 2026-04-17

## Context

DevArch V1 stores desired state in relational tables and generates compose output on demand. That made desired state fragmented, hard to reason about, and expensive to evolve.

V2 needs one canonical desired-state model that is portable, diffable, and understandable without a running database.

## Decision

DevArch V2 will use files on disk as the canonical desired-state source.

Specifically:

- each workspace has one canonical manifest (`devarch.workspace.yaml`)
- catalog templates live as plain files under catalog roots
- manifests and templates are versioned, validated, and loaded into normalized in-memory models
- runtime views, plans, and apply history are derived from manifests plus runtime inspection

## Consequences

### Positive

- desired state is portable and VCS-friendly
- onboarding does not require a canonical database before the product is useful
- schema validation and example-driven development become straightforward
- engine behavior can be golden-tested against files

### Negative

- file validation and normalization must be robust
- some V1 CRUD flows must be redesigned into manifest editing flows
- importer work must map V1 relational exports into V2 manifests intentionally

## Follow-on implications

- schema files under `schemas/` become first-class contracts
- the resolver and planner must treat manifests as the source of truth
- runtime caches must never backflow into desired-state storage
