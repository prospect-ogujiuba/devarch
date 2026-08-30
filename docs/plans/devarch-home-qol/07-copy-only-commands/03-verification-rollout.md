# Phase 3 — Security, portability, and verification

Created: 2026-08-29
Purpose: Prove displayed command strings are safe to paste and cannot become an execution surface.

## Goal

Prove every copied command is safely quoted, visibly identical to clipboard content, portable within scope, and never executed by the dashboard.

## Scope

Adversarial quoting, inert-shell round trips, execution-surface audit, clipboard parity, platform documentation, bounded logs, rollout, and future-change policy.

## Work

- Build adversarial fixtures for paths/names containing whitespace, quotes, substitution syntax, separators, control characters, Unicode, and leading dashes.
- Validate rendered commands by parsing/executing only against inert test doubles in temporary directories; never target real containers in automated quoting tests.
- Search server/client code for `shell=True`, `eval`, executable API routes, or dynamic operator concatenation and fail review if introduced.
- Verify absolute paths appear only where already permitted on the localhost dashboard and never enter localStorage or URLs.
- Confirm Linux/macOS POSIX scope and the configured Compose provider; document that WSL users should copy into a POSIX shell and PowerShell is not generated.
- Ensure container logs commands are bounded and service logs use explicit tail counts.

## Outputs

- Adversarial quoting and inert-shell test suites.
- Static execution-surface audit.
- Clipboard/responsive verification evidence and POSIX scope documentation.

## Acceptance criteria

- Adversarial values remain single shell arguments or are rejected.
- Clipboard contents exactly match visible command previews.
- No command is invoked by dashboard tests, server requests, or browser actions.
- The action registry contains no start/stop/restart/delete additions beyond the pre-existing copied service start command.
- Existing search, filters, palette, routes, and manual refresh remain green.

## Verification

Run hostile argv fixtures, inert temporary-directory shell parsing, clipboard/preview comparison, static searches for execution primitives, mobile layout checks, and existing dashboard regressions.

## Rollout

Ship the descriptor/quoting migration first while preserving current UI. Add inspect/navigation commands second. Add palette entries last after card/detail parity. Any future command requires a catalog change, classification, threat review, and tests.

## Exit evidence

Adversarial quoting suite, inert-shell round-trip tests, static execution-surface audit, responsive screenshots, clipboard checks, and README safety documentation.