# VER-001: integrated golden paths and documentation

- Status: planned
- Plan revision: 1
- Spec: r1
- Dependencies: `SHOW-003`, `REC-002`

## Goal and observable behavior

Deliver one documented, verified showcase workflow covering top-level discovery, JavaScript compatibility, Laravel profiles/matrix, WordPress adapter, and the Laravel–TanStack recipe.

## Scope

Run focused and adjacent tests, static shell checks, deterministic list snapshots, five representative Laravel runtime smokes, WordPress adapted dry-run, existing JavaScript runtime smoke, TanStack recipe smoke, HTTPS/API/service assertions, secret scan, failure/recovery exercises, and cleanup. Update root/ecosystem/showcase documentation and examples.

## Non-goals

No new features, profile expansion, performance optimization, all-profile concurrent startup, or external SaaS validation.

## Expected files/outputs

Root README updates, `scripts/showcase/README.md`, ecosystem READMEs, test runner integration, durable verification report under this initiative.

## Acceptance criteria

- AC-12/13 and the complete acceptance-to-evidence matrix pass or explicitly record a blocking gap.
- Documentation distinguishes definitions from generated apps and names support tiers/prerequisites.
- Commands shown in documentation are exercised by tests or the runtime smoke report.
- Existing unrelated worktree changes are preserved.

## Verification

All focused suites; `bash -n`; ShellCheck when installed; existing DevArch/bootstrap/matrix suites; bounded runtime golden paths; Windows/Unix hosts helpers; secret scan of logs/definitions; acceptance-to-evidence report.

## Rollback

Documentation/test-runner changes revert independently; prior verified implementation contracts remain intact.
