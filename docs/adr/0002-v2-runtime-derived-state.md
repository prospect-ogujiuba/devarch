# ADR 0002 — V2 runtime state is derived and non-canonical

- Status: Accepted
- Date: 2026-04-17

## Context

V2 still needs runtime inspection, apply history, recent events, logs metadata, and optional caching. However, if that data becomes canonical desired state, the rewrite recreates the same ambiguity V2 is meant to remove.

## Decision

All runtime-oriented state in V2 is derived, cached, or historical — never canonical desired state.

This includes:

- runtime snapshots
- planner inputs sourced from runtime inspection
- apply history
- event streams
- optional SQLite caches for snapshots and history

The only canonical desired-state inputs are workspace manifests and catalog templates.

## Consequences

### Positive

- the product keeps a clean separation between desired and observed state
- caches can be optional, replaceable, and rebuildable
- planner and apply behavior stay reproducible from manifests plus runtime inspection

### Negative

- the engine must be explicit about refresh, invalidation, and cache miss behavior
- operators cannot rely on caches as hidden config stores

## Guardrails

- no cache write path may be required to persist desired-state edits
- apply history must not be used as configuration input
- UI and API responses must label runtime-derived data clearly when needed
