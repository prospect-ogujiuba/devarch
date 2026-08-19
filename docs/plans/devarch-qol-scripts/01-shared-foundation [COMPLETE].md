# Phase 01: Shared scripting foundation

Created: 2026-08-19
Purpose: Prevent eighteen scripts from duplicating unsafe runtime detection, Compose discovery, output formatting, and confirmation logic.

## Goal

Create only the repository-resolution and safety helpers that native commands cannot provide; deliberately avoid a container-runtime framework.

## Scope

- Start with `scripts/devarch/lib/catalog.sh` and `common.sh`; add another library only after two implemented phases demonstrate the same DevArch-specific need.
- Discover `services-library/*/*/compose.yml` and emit canonical `category/name` IDs in deterministic order.
- Require `podman`; invoke `podman compose` directly and let Podman select its configured external Compose provider.
- Provide one array-safe `run`/`exec` helper only for logging exact argv and preserving native behavior; do not wrap individual Podman subcommands.
- Reuse native `--format`, `--filter`, JSON, `--watch`, `--dry-run`, and completion support instead of output/runtime abstractions.
- Add temporary fixture repositories and recording `podman`, database CLI, launcher, and certificate executables.

## Native delegation

- `podman compose --help` is the capability check; no Podman/Docker feature matrix is maintained.
- Wrapper arguments end at `--`; everything after it is forwarded unchanged to the documented native command.
- Commands use `exec` for the final native process whenever no post-processing is required.

## Outputs

- Minimal catalog/common helpers with documented public functions and an explicit no-runtime-abstraction boundary.
- `scripts/devarch/tests/test-helper.sh` and foundation regression tests.
- A short contributor contract in `scripts/devarch/README.md`.

## Acceptance criteria

- Canonical IDs resolve exactly; ambiguous short IDs fail with candidates.
- Discovery rejects paths escaping the repository and malformed/missing Compose files.
- Missing Podman or Compose provider produces the native installation/configuration guidance; no Docker fallback is invented.
- Native output is not reformatted unless a phase documents a DevArch-only data join.
- Wrapper dry-run is limited to DevArch-owned filesystem changes; Compose simulation delegates to `podman compose --dry-run` when supported by the installed provider.
- Passthrough preserves stdout, stderr, stdin, TTY, signals, and exit status.

## Verification

- `bash -n scripts/devarch/lib/*.sh scripts/devarch/tests/*.sh`
- Run foundation tests against fixture catalogs and a recording Podman command.
- Run ShellCheck on the shared libraries.
- Verify tests pass from both repository root and an unrelated working directory.

## Non-goals

No runtime adapter, JSON schema framework, lifecycle command, persistent index, daemon, dependency solver, or general-purpose shell framework is introduced here.
