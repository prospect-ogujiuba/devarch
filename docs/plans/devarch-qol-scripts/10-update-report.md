# Phase 10: Container image update report

Created: 2026-08-19
Purpose: Show image freshness and pinning risk without silently changing 174 Compose definitions.

## Goal

Use Podman/Skopeo's native image inspection and auto-update dry-run; do not build a registry client or version resolver.

## Scope

- Use `podman images`, `podman image inspect`, and Compose config output for configured/local references.
- For locally managed containers using supported systemd/Quadlet auto-update workflows, use `podman auto-update --dry-run` and its native output. Detect and report that `podman auto-update` is unavailable with the remote Podman client; do not emulate it.
- For remote read-only inspection outside auto-update, invoke `skopeo inspect docker://IMAGE` directly; if Skopeo is absent, document that limitation.
- Report only native digest/tag facts and whether DevArch services are configured for Podman auto-update. Do not infer semantic “latest version.”
- Add a script only if joining configured Compose image references to those native commands saves meaningful repetition; otherwise ship a command recipe.

## Native delegation

Podman owns local image metadata and supported local auto-update decisions. Skopeo owns registry transport/authentication/manifest inspection. Remote-client or unsupported-auto-update environments receive direct native guidance, not a DevArch fallback implementation. DevArch does not implement registry HTTP, retries, concurrency, caching, semver, or release comparison.

## Outputs

- Native Podman/Skopeo command recipe or thin configured-image iterator, recording fixtures, and docs; no cache or registry rules engine.

## Acceptance criteria

- Reporting never pulls images or edits Compose files; only native read-only/dry-run operations are invoked.
- Authentication failures and rate limits are isolated per registry and do not expose credentials.
- Output does not claim “newer version available” unless the native tool does; digest/tag facts are labeled precisely.
- Authentication, rate-limit, and transport diagnostics remain Skopeo/Podman output.

## Verification

- Recording tests cover exact `podman images`/`image inspect`/supported-local `auto-update --dry-run` and `skopeo inspect` argv, plus the remote-client unsupported path.
- Assert no `pull`, registry HTTP client, worker pool, cache, or Compose mutation exists.
- Manually compare a small sample with registry output when network access is available.

## Non-goals

No registry client, cache, retry/concurrency engine, semantic-version resolver, automated upgrades, compatibility claims, release-note synthesis, or credential provisioning.
