# ADR 0003 — V2 local API remains thin and engine-backed

- Status: Accepted
- Date: 2026-04-17

## Context

V1 accumulated a large CRUD-heavy API surface that mirrored database decomposition more than the primary user workflow.

V2 needs an API for the web UI and local operator tooling, but repeating V1's transport sprawl would undermine the rewrite.

## Decision

DevArch V2 will expose a thin local API whose handlers delegate to shared engine services.

The API should:

- expose only the minimal catalog, workspace, runtime, import, and scan flows needed by the product
- return computed views such as manifest, graph, plan, status, logs, and events
- avoid broad child-table mutation surfaces and entity-first CRUD expansion
- share the same engine behavior as the CLI

## Consequences

### Positive

- CLI and API stay aligned through shared services
- behavior differences are easier to detect and test
- the web UI consumes stable task-oriented endpoints instead of transport-only CRUD shapes

### Negative

- some V1 endpoint patterns will intentionally disappear or be rejected during migration
- transport-layer convenience must not short-circuit the engine

## Guardrails

- if a feature can only be implemented by adding transport-specific logic, revisit the engine boundary first
- parity checks should compare CLI and API outcomes at the engine result level where practical
- daemon bootstrap, routing, and event transport remain thin wrappers around shared services
