# Phase 01: Shared scripting foundation

Created: 2026-08-19
Purpose: Prevent eighteen scripts from duplicating unsafe runtime detection, Compose discovery, output formatting, and confirmation logic.

## Goal

Create a small sourced library and fixture-based test harness that later scripts can reuse without becoming a monolithic CLI.

## Scope

- Add `scripts/devarch/lib/common.sh`, `catalog.sh`, `runtime.sh`, `output.sh`, and `safety.sh` with narrow public functions.
- Discover `services-library/*/*/compose.yml` and emit canonical `category/name` IDs in deterministic order.
- Detect Podman/Docker and Compose provider once; represent commands as Bash arrays.
- Define exit codes, log/error prefixes, `--json` conventions, dry-run rendering, confirmation, and secret redaction.
- Add temporary fixture repositories and fake/rejecting runtime executables.

## Outputs

- Shared libraries with documented public functions.
- `scripts/devarch/tests/test-helper.sh` and foundation regression tests.
- A short contributor contract in `scripts/devarch/README.md`.

## Acceptance criteria

- Canonical IDs resolve exactly; ambiguous short IDs fail with candidates.
- Discovery rejects paths escaping the repository and malformed/missing Compose files.
- Runtime selection honors `CONTAINER_RUNTIME`, prefers Podman, and supports `podman-compose` fallback.
- JSON output is valid and diagnostics go to stderr.
- Dry-run and confirmation helpers cannot execute the supplied mutation callback accidentally.
- Redaction covers allowlisted credential names and URL userinfo.

## Verification

- `bash -n scripts/devarch/lib/*.sh scripts/devarch/tests/*.sh`
- Run foundation tests against fixture catalogs and fake Podman/Docker commands.
- Run ShellCheck on the shared libraries.
- Verify tests pass from both repository root and an unrelated working directory.

## Non-goals

No lifecycle command, persistent index, daemon, dependency solver, or general-purpose framework is introduced here.
