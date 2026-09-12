# REC-001: recoverable multi-app recipe engine

- Status: planned
- Plan revision: 3
- Spec: r2
- Dependencies: `SHOW-002`, `LAR-006`

## Goal and observable behavior

Allow tracked data-only recipes to coordinate multiple adapter-owned applications with deterministic names, validated dependency order, explicit configuration links, and durable recovery state.

## Scope

Define recipe protocol v1 with allowlisted directives for app nodes, dependency edges, and named non-secret configuration outputs/inputs; validate unique IDs, acyclic graph, adapter/profile/action support, target collisions, and all inputs before mutation. Execute Kahn topological order with lexical node-ID tie-breaking. Persist a versioned restricted line-record recovery file recording pre-existing, creating, created, failed, and pending nodes plus recipe hash/prefix; replace it atomically through sibling temporary write and rename; never delete pre-existing targets.

## Non-goals

No arbitrary commands/templates, secret values in recipe files/manifests, concurrent creation, distributed transactions, or automatic rollback deletion.

## Expected files

`scripts/showcase/recipes/`, recipe parser/runner and tests, recovery documentation.

## Acceptance criteria

- AC-09 and recipe-engine portion of AC-10 pass.
- Validation rejects cycles, duplicate nodes/outputs, missing adapters/profiles, traversal, unsupported links, and target collisions before mutation.
- Stable Kahn topological ordering uses declaration order only as deterministic tie-breaker.
- Failure stops dependent nodes, preserves independent/pre-existing nodes, quarantines partial adapter targets, and writes a secret-free actionable recovery record.
- `status` and `resume` validate recovery schema, recipe hash, prefix, paths, and definition identity before reporting or continuing; stale/corrupt records fail closed.

## Verification

Red parser/graph fixtures including shuffled declarations and lexical tie order; failure injection at every atomically persisted node transition; interrupted-write, resume/stale-definition/prefix/pre-existing-target tests; synthetic canary secret scan; then one offline two-app fixture integration.

## Rollback

Remove recipe runner/definitions; preserve all generated and quarantined trees for manual recovery.
