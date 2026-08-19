# Phase 17: Port inventory and conflict detector

Created: 2026-08-19
Purpose: Prevent service startup failures caused by duplicate Compose mappings or ports already occupied on the host.

## Goal

Add `scripts/devarch/ports.sh` to provide deterministic static allocation and live availability reports.

## Scope

- Ask each provider to normalize its file with `podman compose -f FILE config`; consume structured config output when supported. If structured output is unavailable, use an established YAML library in the validation tool—not a shell YAML parser.
- Commands/modes: table, `--conflicts`, `--available`, category/service filters, and `--json`.
- Detect cross-file static collisions (the DevArch-specific gap) and use native `podman port`/`podman ps` plus `ss` or `lsof` for live listeners.
- Suggest free ports within documented category ranges without editing files.
- Distinguish TCP/UDP and loopback/wildcard binding overlap semantics.

## Native delegation

Compose owns interpolation/normalization; Podman owns published-container port reporting; `ss`/`lsof` own host listener inventory. DevArch only compares its many independent Compose definitions and optionally documents suggested ranges.

## Outputs

- Ports script, minimal normalized-config extractor, platform command selection, optional allocation-policy document, tests, and docs.

## Acceptance criteria

- Reports all owners of a collision, including the current port `8091` assignments.
- Correctly treats `0.0.0.0:PORT` as overlapping `127.0.0.1:PORT`, while TCP and UDP remain distinct.
- Compose/provider interpolation errors are shown natively; unresolved values are reported as unknown, never guessed.
- `--available` avoids both static reservations and live listeners and is race-qualified as advisory.
- Static inventory works without runtime or elevated access.

## Verification

- Provider/YAML fixtures cover IPv4/IPv6, wildcard/loopback, ranges, long syntax, protocols, variables, duplicates, and malformed mappings without re-testing a home-grown YAML grammar.
- Fake `ss`/`lsof` outputs cover Linux/macOS and permission limitations.
- Validate normalized-config extraction against all real Compose files and compare live results with native `podman port`/`ss` output.

## Non-goals

No Compose/YAML parser in Bash, socket scanner, automatic port reassignment, process termination, firewall configuration, or permanent reservation daemon.
