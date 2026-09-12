# REC-001: recoverable multi-app recipe engine

- Status: planned
- Plan revision: 1
- Spec: r1
- Dependencies: `SHOW-002`, `LAR-003`

## Goal and observable behavior

Allow tracked data-only recipes to coordinate multiple adapter-owned applications with deterministic names, validated dependency order, explicit configuration links, and durable recovery state.

## Scope

Define recipe protocol v1 with allowlisted directives for app nodes, dependency edges, and named configuration outputs/inputs; validate unique IDs, acyclic graph, adapter/profile existence, target collisions, and all inputs before mutation. Execute stable topological order. Persist a recovery manifest recording pre-existing, created, failed, and pending nodes; never delete pre-existing targets.

## Non-goals

No arbitrary commands/templates, secret values in recipe files/manifests, concurrent creation, distributed transactions, or automatic rollback deletion.

## Expected files

`scripts/showcase/recipes/`, recipe parser/runner and tests, recovery documentation.

## Acceptance criteria

- AC-09 and recipe-engine portion of AC-10 pass.
- Validation rejects cycles, duplicate nodes/outputs, missing adapters/profiles, traversal, unsupported links, and target collisions before mutation.
- Stable Kahn topological ordering uses declaration order only as deterministic tie-breaker.
- Failure stops dependent nodes, preserves independent/pre-existing nodes, quarantines partial adapter targets, and writes a secret-free actionable recovery manifest.
- Resume validates manifest/definition identity before continuing.

## Verification

Red parser/graph fixtures, failure injection at every node transition, resume/stale-definition/pre-existing-target tests, secret scan, then one two-app fixture integration.

## Rollback

Remove recipe runner/definitions; preserve all generated and quarantined trees for manual recovery.
