# Phase 17: Port inventory and conflict detector

Created: 2026-08-19
Purpose: Prevent service startup failures caused by duplicate Compose mappings or ports already occupied on the host.

## Goal

Add `scripts/devarch/ports.sh` to provide deterministic static allocation and live availability reports.

## Scope

- Parse short and long Compose port syntax, protocol, interface, ranges, and variable interpolation that can be resolved safely.
- Commands/modes: table, `--conflicts`, `--available`, category/service filters, and `--json`.
- Detect static collisions across the catalog and live listeners using `ss`, `lsof`, or platform fallbacks.
- Suggest free ports within documented category ranges without editing files.
- Distinguish TCP/UDP and loopback/wildcard binding overlap semantics.

## Outputs

- Ports script, parser, platform listener adapters, optional allocation-policy document, tests, and docs.

## Acceptance criteria

- Reports all owners of a collision, including the current port `8091` assignments.
- Correctly treats `0.0.0.0:PORT` as overlapping `127.0.0.1:PORT`, while TCP and UDP remain distinct.
- Unresolved environment expressions are reported as unknown, never guessed.
- `--available` avoids both static reservations and live listeners and is race-qualified as advisory.
- Static inventory works without runtime or elevated access.

## Verification

- Fixture matrix covers IPv4/IPv6, wildcard/loopback, ranges, long syntax, protocols, variables, duplicates, and malformed mappings.
- Fake `ss`/`lsof` outputs cover Linux/macOS and permission limitations.
- Validate parser results against all real Compose files.

## Non-goals

No automatic port reassignment, process termination, firewall configuration, or permanent reservation daemon.
