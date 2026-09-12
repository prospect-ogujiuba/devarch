# VER-001: integrated golden paths and documentation

- Status: planned
- Plan revision: 3
- Spec: r2
- Dependencies: `SHOW-003`, `REC-002`

## Goal and observable behavior

Deliver one documented, verified showcase workflow covering top-level discovery, JavaScript compatibility, Laravel profiles/matrix, WordPress adapter, and the Laravel–TanStack recipe.

## Scope

Run default offline focused/adjacent tests, characterization, static shell checks, deterministic list snapshots, WordPress adapted dry-run, secret-canary scans, and failure/recovery exercises. Run explicitly opt-in, uniquely named, prerequisite-checked, bounded golden paths for five representative Laravel profiles, existing JavaScript runtime, and TanStack recipe HTTPS/API/service behavior; preserve failures and document cleanup. Update root/ecosystem/showcase documentation and examples.

## Non-goals

No new features, profile expansion, performance optimization, all-profile concurrent startup, or external SaaS validation.

## Expected files/outputs

Root README updates, `scripts/showcase/README.md`, ecosystem READMEs, test runner integration, durable verification report under this initiative.

## Acceptance criteria

- AC-12/13 and the complete acceptance-to-evidence matrix pass or explicitly record a blocking gap.
- Documentation distinguishes definitions, unmanaged existing apps, and matrix-generated apps; names support tiers, supported actions, prerequisites, trusted-upstream execution, recovery, and cleanup.
- Commands shown in documentation are exercised by tests or the runtime smoke report.
- Existing unrelated worktree changes are preserved.

## Verification

Default: all focused offline suites, characterization, `bash -n`, ShellCheck when installed, existing DevArch/bootstrap/matrix suites, Windows/Unix hosts helpers, and synthetic canary scan of argv/logs/definitions/recovery. Opt-in: bounded uniquely named runtime golden paths with disk/network/service probes, timeouts, preserved logs, and cleanup. Produce the complete acceptance-to-evidence report.

## Rollback

Documentation/test-runner changes revert independently; prior verified implementation contracts remain intact.
