# Phase 13: Log aggregation helper

Created: 2026-08-19
Purpose: Make recent or live logs accessible by service, stack, or app while keeping output bounded and attributable.

## Goal

Prefer direct `podman logs`; retain only a container-name resolver if DevArch stack/app selection adds real value.

## Scope

- Resolve one or more service/stack/app selections to container names using Compose labels or `podman ps --filter`.
- Invoke one `podman logs` command with all selected containers. Podman's native `--names`, `--follow`, `--since`, `--until`, `--tail`, and `--timestamps` flags remain unchanged after `--`.
- Filtering is documented as a normal pipeline to `grep`/`rg`; the helper does not absorb text-processing flags.
- When the user already knows container names, documentation directs them to `podman logs` rather than the wrapper.

## Native delegation

Podman owns multi-container interleaving, name prefixes, follow mode, time/tail selection, stream handling, and interruption. DevArch only resolves repository concepts to container names.

## Outputs

- Optional container-name resolver that `exec`s one `podman logs` command, recording tests, and direct native examples.

## Acceptance criteria

- DevArch adds no default tail or follow policy; native defaults/flags remain authoritative.
- Final execution uses `exec podman logs ...`; Podman handles Ctrl-C and streams.
- No custom filter expression is accepted.
- Missing/stopped selections are reported without suppressing available logs from other selections.
- Secret redaction is opt-in and clearly documented as best-effort; support bundles always apply strict redaction separately.

## Verification

- Recording tests assert exact multi-container `podman logs` argv, order, and passthrough.
- Verify no multiplexer, stream parser, prefixer, or child-process pool exists.
- Manual smoke test against two disposable/running services when available.

## Non-goals

No log multiplexer/parser, prefixer, filtering engine, storage, rotation, indexing, alerting, or replacement for Podman/Loki/ELK.
