# Phase 1 — Git inventory and safety contract

Created: 2026-08-29
Purpose: Specify exactly which Git facts may cross the API boundary and how cost is bounded.

## Goal

Add an optional `git` object to each app without leaking repository content or making Git availability mandatory.

## Scope

Additive Git metadata schema, safe subprocess boundary, output/time/concurrency limits, error codes, and fixtures. No collector integration or UI.

## API shape

```json
{
  "isRepository": true,
  "branch": "main",
  "detached": false,
  "dirty": true,
  "trackedChanges": 2,
  "untrackedChanges": 1,
  "head": "a1b2c3d",
  "lastCommitAt": "2026-08-29T10:00:00Z",
  "error": null
}
```

Use a discriminated result: return `null` only for a confirmed non-repository; return the object above for success with `error: null`; and return an object with `isRepository: null`, all fact fields `null`, and one bounded error code for a failed/indeterminate check. A timeout, missing Git binary, unsafe path, or unreadable repository must never be reported as a confirmed non-repository. Define how aggregate `capabilities.git` derives `ready`, `partial`, or `unavailable` from these per-app results. Never return raw stderr containing paths or config.

## Safety and cost

- Invoke `git` with argv arrays, `-C <validated app path>`, no shell, `GIT_OPTIONAL_LOCKS=0`, `GIT_TERMINAL_PROMPT=0`, `GIT_PAGER=cat`, and a deterministic locale. Remove inherited Git override variables not explicitly required by the collector.
- Use read-only commands only and disable repository-configured helper execution (including filesystem monitors and hooks) with explicit command configuration. Tests must prove a malicious local helper/config cannot create a marker file or child process.
- Do not follow workspace symlinks outside the validated `apps` root.
- Cap combined output at 256KB/repository, timeout at 750ms/command, concurrent Git processes at 4, repository count at the app inventory cap, and total Git deadline at 5s. Kill and reap timed-out children.
- Parse a documented porcelain format with NUL-safe record handling so unusual filenames cannot corrupt counts; never expose raw output or filenames.
- Normalize branch/detached values and UTC timestamp.

## Outputs

- Git metadata data class/dict contract and error-code enum.
- Command runner seam for deterministic tests.
- Fixture matrix for clean, dirty, untracked, unborn, detached, non-repository, missing Git, timeout, unsafe path, and malicious repository-configured helpers.

## Acceptance criteria

The contract is additive, JSON-safe, secret-safe, bounded, and documented before UI work begins.

## Verification

Schema serialization tests, subprocess/environment inspection, secret-decoy fixtures, timeout/output-limit tests, malicious-helper non-execution, and repository-state fixture review.