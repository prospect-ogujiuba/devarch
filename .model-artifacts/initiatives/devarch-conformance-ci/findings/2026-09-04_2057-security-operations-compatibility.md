# Specialist Finding: Conformance Safety, Migration, Performance, Operations, and Compatibility

- Topic: `devarch-conformance-ci`
- Specialists: security, migration, performance, operations, compatibility
- Decision: complete; incorporate before approval
- Assessed spec: `.model-artifacts/initiatives/devarch-conformance-ci/specs/2026-09-04_2057-initiative-spec-r1.md`
- Assessed plan: `.model-artifacts/initiatives/devarch-conformance-ci/plans/2026-09-04_2057-plan-index-r1.md`
- Created: 2026-09-04T20:57:54Z

## Security

Use sanitized child environments and placeholder fixtures; reports expose variable names, never values. Reject secret-bearing metadata. Run rootless without privileged containers, host mutation, or application access to the container socket. Use least GitHub token permissions, fork-safe triggers, immutable action SHA pins, bounded redacted artifacts, and no production/private credentials.

## Migration

Version `conformance.yml` from v1, fail unknown versions/fields, and migrate category-by-category. Mechanical generation may populate IDs and Compose facts, but readiness, smoke tier, and exceptions require manual review. Exceptions identify rule, rationale, owner, and review date. Avoid unrelated Compose normalization. CI enforcement follows full local migration success and can be rolled back independently.

## Performance

Record per-check and scenario durations. Keep static/unit checks near the characterized baseline, bound subprocess/log/artifact/scenario/global time and bytes, and cache only immutable dependencies. Do not parallelize legacy shared resources until an isolation canary passes. Compute reliability over 30 completed runs and emit `insufficient-sample` before that threshold.

## Operations

Use collision-resistant run IDs and labels on all owned projects, containers, volumes, networks, apps, and temp paths. Teardown must be unconditional, signal-safe, timeout-safe, idempotent, and ownership-scoped; a leak query gates success. Capability absence is not pass. CI required-check activation and rollback are explicit administrator steps after an actual stable run.

## Compatibility

The verifier and metadata are additive. Retain existing script paths/flags/direct tests, rootless user model, canonical service IDs, and configured `podman compose` provider selection. Do not add a Docker fallback to shared helpers. Treat one-shot jobs and aggregate project stacks semantically rather than imposing service-only health rules.

## Finding

Approve these specialist areas if r2 incorporates the boundaries above into affected contracts and verification, and keeps branch-protection mutation outside repository automation.
