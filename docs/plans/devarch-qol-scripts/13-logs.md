# Phase 13: Log aggregation helper

Created: 2026-08-19
Purpose: Make recent or live logs accessible by service, stack, or app while keeping output bounded and attributable.

## Goal

Add `scripts/devarch/logs.sh` as a focused wrapper over runtime logs.

## Scope

- Select one or more canonical services, a category, a stack, an app's infrastructure, or all running DevArch containers.
- Support follow, since, until, tail, timestamps, grep/fixed filtering, and no-color.
- Prefix multiplexed lines with canonical service IDs and preserve single-service raw mode.
- Reuse runtime inventory and stack/app resolution.

## Outputs

- Logs script, stream multiplexer/parser, tests, and examples.

## Acceptance criteria

- Non-follow mode has a safe default tail limit.
- Follow mode handles Ctrl-C and terminates all child processes.
- User filters are passed as data, not evaluated shell expressions.
- Missing/stopped selections are reported without suppressing available logs from other selections.
- Secret redaction is opt-in and clearly documented as best-effort; support bundles always apply strict redaction separately.

## Verification

- Simulated interleaved streams test attribution, partial lines, UTF-8, child failure, filters, bounds, and cleanup.
- Assert no orphaned fake log processes after interruption.
- Manual smoke test against two disposable/running services when available.

## Non-goals

No log storage, rotation, indexing, alerting, or replacement for Loki/ELK.
