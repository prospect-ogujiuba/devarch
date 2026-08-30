# Phase 3 — Performance, failure, and verification

Created: 2026-08-29
Purpose: Ensure useful Git context does not make manual inventory refresh slow or fragile.

## Goal

Prove Git context is bounded, secret-safe, process-safe, and non-blocking across repository and failure conditions.

## Scope

Benchmarks, timeout/process cleanup, error sanitization, path boundaries, secret-field audit, capability rollback, and documentation.

## Work

- Benchmark no-repository, clean, dirty, and moderately large repositories across representative app counts.
- Test timeout cancellation and executor cleanup; no child process may survive the request. Include a repository-configured helper/monitor decoy and prove collection never executes it.
- Cap and sanitize every stderr/error outcome into stable codes such as `git-missing`, `timeout`, `unreadable`, or `unknown`.
- Verify symlink and nested-worktree boundaries against temporary directories outside `apps`.
- Confirm response contains no commit subject, author, email, remote, URL, config, diff, or path list.
- Document Git status semantics and refresh behavior.

## Outputs

- Git collector benchmark and failure-injection suite.
- Serialized-response secret audit.
- Process cleanup and path-boundary evidence.
- Capability rollback and semantics documentation.

## Acceptance criteria

- With the shared 50-app fixture, inventory including Git targets ≤2s and must stop by the 5s hard deadline; record median and p95 across 20 runs.
- Timeouts return the rest of inventory normally.
- Process count never exceeds the configured worker limit.
- Unit and integration fixtures cover all contract states.
- Existing Podman socket inventory and route tests remain green.
- Page load causes one inventory request; no timers update commit age.

## Verification

Run temporary repository fixtures, synthetic app-count benchmarks, timeout/cleanup and malicious-helper checks, symlink/nested-worktree tests, response audits, and single-fetch assertions.

## Rollout

Ship metadata behind graceful capability detection. If Git discovery causes operational problems, disable the collector while leaving UI tolerant of absent `git` fields; no migration is required.

## Exit evidence

Benchmark table, failure-injection tests, secret-field audit, process cleanup check, screenshots, and README/API documentation.