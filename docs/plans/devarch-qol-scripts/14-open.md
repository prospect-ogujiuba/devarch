# Phase 14: Endpoint opener

Created: 2026-08-19
Purpose: Open applications and service UIs without remembering ports or hand-assembling URLs.

## Goal

Add `scripts/devarch/open.sh` to resolve a service or app to a documented local endpoint and optionally launch the platform browser.

## Scope

- Resolve app URLs from validated app configuration/default `<name>.test` conventions.
- Resolve service URLs from an explicit declarative endpoint registry with scheme, host, published port, path, and optional description.
- Commands/modes: default open, `--print`, `--list`, `--all-endpoints`, and endpoint selection for multi-UI services.
- Support Linux (`xdg-open`), macOS (`open`), WSL/Windows (`powershell.exe Start-Process`), and `BROWSER` override without shell evaluation.
- Optionally check runtime state/reachability and offer an actionable message; never auto-start by default.

## Outputs

- Open script, endpoint registry/schema, platform launcher adapter, tests, and docs.

## Acceptance criteria

- Endpoint metadata is explicit and reviewable; ports are not guessed from arbitrary container ports.
- `--print` and `--list` work headlessly and never launch a browser.
- URLs and launcher arguments are passed as single array elements and reject controls/unsafe schemes.
- Unknown, ambiguous, stopped, and unreachable targets produce distinct guidance.
- Every registry endpoint references an existing canonical service and published port.

## Verification

- Fake launchers assert exact argv on Linux, macOS, WSL, and override paths.
- Validate registry references and test multi-endpoint selection, app URLs, IPv6/host formatting if supported, and malicious values.
- Manual smoke test uses `--print` first, then one app and one service UI.

## Non-goals

No proxy configuration, automatic hosts edits, automatic service startup, or endpoint scraping from container logs.
