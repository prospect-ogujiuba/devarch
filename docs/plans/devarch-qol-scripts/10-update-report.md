# Phase 10: Container image update report

Created: 2026-08-19
Purpose: Show image freshness and pinning risk without silently changing 174 Compose definitions.

## Goal

Add `scripts/devarch/update-report.sh` for read-only comparison of configured image references with registry metadata.

## Scope

- Inventory image references, digest pinning, floating tags, duplicate images, and locally available digests.
- Optionally query registry metadata using runtime-native manifest inspection, with bounded concurrency, timeout, retry, and local cache.
- Report digest changes for mutable tags; report newer semantic tags only where registry/tag rules are trustworthy.
- Support category/service filters, offline mode, cache TTL, and JSON.

## Outputs

- Update-report script, cache format/location, registry comparison rules, fixtures, and docs.

## Acceptance criteria

- The script never pulls images or edits Compose files.
- Authentication failures and rate limits are isolated per registry and do not expose credentials.
- Digest-change reporting distinguishes “tag moved” from “newer version available.”
- Cache writes are atomic and cache corruption degrades to refetch/offline reporting.
- Results are deterministic once registry fixtures are fixed.

## Verification

- Tests cover Docker Hub and GHCR-style references, digests, floating tags, prereleases, rate limits, timeouts, and corrupt cache.
- Assert bounded worker count and no `pull`/Compose mutation.
- Manually compare a small sample with registry output when network access is available.

## Non-goals

No automated upgrades, compatibility claims, release-note synthesis, or credential provisioning.
