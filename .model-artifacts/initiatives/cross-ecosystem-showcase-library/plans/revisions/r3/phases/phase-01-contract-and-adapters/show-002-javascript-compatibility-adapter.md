# SHOW-002: JavaScript compatibility adapter

- Status: planned
- Plan revision: 3
- Spec: r2
- Dependencies: `SHOW-001`

## Goal and observable behavior

Expose every existing JavaScript framework/profile through adapter protocol v1 while preserving direct bootstrap/matrix commands, deterministic names, skip/retry behavior, logs, and exit codes.

## Scope

Implement the JavaScript adapter mapping list/create/start/stop/status/verify to existing scripts; add additive compatibility assertions.

## Non-goals

No JavaScript profile changes, bootstrap rewrite, bulk concurrency, or generated app commits.

## Expected files

`scripts/showcase/adapters/javascript.sh`, focused showcase fixtures, additive `scripts/javascript/scaffold-matrix.test.sh` assertions.

## Acceptance criteria

- AC-02/03 pass.
- Direct and adapted plans/argv are behavior-equivalent for representative profiles.
- Existing app prefix and adapter passthrough remain supported.
- Adapter does not parse or source JavaScript profile contents itself.

## Verification

Red characterization for direct-vs-adapter list/create/start/stop, then focused and full JavaScript bootstrap/matrix suites.

## Rollback

Remove adapter only.
