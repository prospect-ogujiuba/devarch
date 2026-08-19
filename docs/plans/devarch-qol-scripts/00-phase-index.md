# DevArch quality-of-life scripts phase index

Created: 2026-08-19
Purpose: Define the implementation order and shared contracts for eighteen operational scripts without recreating the removed DevArch application or daemon.

## Outcome

Add small Bash entrypoints under `scripts/devarch/` plus shared, testable helpers. Compose files remain the source of truth; scripts discover and operate on them rather than maintaining a second service database.

## Native-first policy

DevArch is repository context around native tools, not a container-management product. Every phase must pass this test before implementation:

1. **Use the native command directly.** Podman operations call `podman`; Compose operations call `podman compose`; database operations call the database's own CLI; certificates use `mkcert`/`openssl`; browser launching uses the operating-system launcher.
2. **Add only DevArch context.** A wrapper may resolve a repository Compose path, choose a declared stack, map an app to its container working directory, or enforce a documented local safety rule. It must not reproduce behavior already exposed by the native tool.
3. **Prefer passthrough over translation.** After resolving DevArch-specific inputs, use `exec` where possible, forward remaining arguments unchanged after `--`, and preserve native stdout, stderr, TTY, signals, prompts, formatting, and exit status.
4. **Do not normalize multiple runtimes.** New QoL commands are Podman-first and call the installed `podman compose` provider. They do not maintain a Docker/Podman compatibility abstraction. Existing bootstrap compatibility is separate legacy scope.
5. **Do not parse when native formatting exists.** Use Podman's `--format`, JSON, labels, filters, `--watch`, and multi-container support. Parse native output only for a narrowly documented DevArch decision that cannot be expressed with native flags.
6. **Do not orchestrate what Compose already orchestrates.** Multi-service lifecycle must be one `podman compose` invocation with declared files/profiles, not loops, schedulers, retry engines, rollback engines, or home-grown dependency graphs.
7. **Do not wrap for naming alone.** If a proposed script would only rename a native command, replace it with documentation, an example, an alias suggestion, or native shell completion.
8. **Keep native help reachable.** Thin wrappers must support `--help-native` or document the exact underlying command, and their own help must show the command that will run.

## Cross-cutting constraints

- Bash 4+, `set -Eeuo pipefail`, `LC_ALL=C`, arrays for commands, and no `eval`.
- Resolve repository paths from `BASH_SOURCE`, not the caller's working directory.
- Accept canonical service IDs as `category/name`; permit a short name only when unique.
- Prefer native dry-run modes such as `podman compose --dry-run`; only add a wrapper dry-run for DevArch-owned filesystem mutations that have no native equivalent.
- Never print secrets or copy native credentials into a DevArch state format.
- Host-only tests use temporary fixture repositories and recording executables to assert exact native argv and passthrough behavior.
- No daemon, API, state database, package manager, runtime abstraction, custom terminal dashboard, or hidden background process.

## Native command ownership map

| Concern | Native owner | DevArch's permitted contribution |
| --- | --- | --- |
| Compose lifecycle/config | `podman compose -f ...` | Resolve Compose file(s), then pass through arguments. |
| Container state/health | `podman ps`, `podman inspect`, `podman healthcheck run` | Apply DevArch labels/filters or provide documented formats. |
| Logs/events | `podman logs`, `podman events` | Resolve container IDs; do not multiplex or recolor. |
| Ports | `podman port`, `podman ps`, `ss`/`lsof` | Compare declared DevArch mappings before startup. |
| Disk/cleanup | `podman system df`, `podman system prune`, object-specific `prune` | Explain scope and pass user-selected native flags. |
| Images/updates | `podman images`, `podman image inspect`, `podman auto-update --dry-run`, `skopeo inspect` | Report configured references and auto-update readiness. |
| Volumes/archive transfer | `podman volume inspect/export/import`, `podman cp` | Map DevArch volume/app names; databases still use native logical dump tools. |
| Database operations | `mariadb`, `pg_dump`/`psql`, `mongodump`/`mongorestore`, `redis-cli`, `sqlite3` via `podman exec` | Resolve the correct container and safe local defaults. |
| Certificates | `mkcert`, `openssl`, OS trust-store tools | Supply DevArch domain/path defaults and show elevation. |
| Browser launch | `xdg-open`, `open`, `powershell.exe Start-Process` | Resolve a declared local URL. |
| Completion | `podman completion` plus each shell's native completion mechanism | Complete only DevArch-specific service/app/stack names. |

## Implementation gate by phase

These are expected shapes, not commitments to create nineteen executables:

- **Thin resolver/passthrough is justified:** 03 service, 05 stack, 06 app, 12 database, and 14 open. Their value is resolving a DevArch name/path before one native `exec`.
- **DevArch-specific policy/script is justified:** 02 doctor, 07 validate, 08 backup/restore selection, 11 support bundle, 15 environment initialization, 16 certificate paths, 17 cross-file ports, and 18 service scaffolding.
- **Documentation/native recipe first; script requires evidence:** 04 status, 09 Podman cleanup, 10 updates, 13 logs, and 19 completion. Podman already owns nearly all behavior in these areas.
- **Shared code is conditional:** phase 01 begins with catalog/path resolution only. A helper is extracted only after two real phases require identical DevArch-specific behavior.

`podman compose` delegates to the installed external provider, so provider-native help/output remains authoritative. If a native capability is unavailable—such as `podman auto-update` through a remote client—DevArch reports the limitation and does not emulate it.

## Plan tracker

Pick any unchecked feature whose dependencies are complete. When a phase satisfies its acceptance criteria and verification, change its checkbox to `[x]` and add `Status: Complete` beneath the `Created:` line in that phase file. Phase 01 is shared prerequisite work, not one of the eighteen feature ideas.

- [ ] `01-shared-foundation.md` — catalog/path resolution, passthrough, safety, and recording-command fixtures. Dependencies: none.
- [ ] `02-doctor.md` — environment and configuration diagnostics. Dependencies: 01.
- [ ] `03-service.md` — single-service lifecycle operations. Dependencies: 01.
- [ ] `04-status.md` — native Podman status recipes and optional label selector. Dependencies: 01, 03.
- [ ] `05-stack.md` — declarative multi-service recipes. Dependencies: 01, 03, 04.
- [ ] `06-app.md` — framework-aware application operations. Dependencies: 01, 03.
- [ ] `07-validate.md` — static validation and regression orchestration. Dependencies: 01.
- [ ] `08-backup-and-restore.md` — durable app/database backups and restores. Dependencies: 01, 06, 12.
- [ ] `09-cleanup.md` — native Podman prune guidance plus separate DevArch-file cleanup. Dependencies: 01, 08.
- [ ] `10-update-report.md` — native Podman/Skopeo image inspection guidance. Dependencies: 01.
- [ ] `11-support-bundle.md` — redacted diagnostic bundles. Dependencies: 02, 04, 07, 17.
- [ ] `12-db.md` — database connection and lifecycle helpers. Dependencies: 01, 03, 06.
- [ ] `13-logs.md` — container-name resolution followed by native `podman logs`. Dependencies: 03, 04, 05, 06.
- [ ] `14-open.md` — endpoint resolution and browser launching. Dependencies: 04, 06.
- [ ] `15-env-init.md` — safe environment initialization and secret generation. Dependencies: 01, 02.
- [ ] `16-certs.md` — local wildcard certificate inspection/generation/trust. Dependencies: 01, 02.
- [ ] `17-ports.md` — static and runtime port inventory/conflict detection. Dependencies: 01.
- [ ] `18-new-service.md` — validated service scaffolding. Dependencies: 07, 17.
- [ ] `19-shell-completions.md` — Podman native completion plus minimal DevArch-name snippets. Dependencies: stable list interfaces from implemented phases.

## Recommended delivery waves

- **Wave 1, substrate:** phases 01–04 and 07. Establish safe discovery, lifecycle, diagnostics, status, and validation.
- **Wave 2, daily workflows:** phases 05–06, 12–14, and 17.
- **Wave 3, data safety:** phases 08–09, then 11.
- **Wave 4, setup and maintenance:** phases 10, 15–16, 18–19.

## Integration map

- A tiny catalog resolver returns Compose paths and declared metadata; it never inventories live Podman state.
- Live state comes directly from `podman ps`/`inspect`/`port` using labels and native formats.
- `service` and `stack` only construct `podman compose -f ...` arguments and then hand over control.
- `app` and `db` only resolve working directory/container identity before `podman exec` invokes the framework/database CLI.
- `support-bundle` captures bounded native diagnostic commands rather than consuming a custom dashboard schema.
- Completions use `podman completion` for Podman and add only DevArch-specific filesystem names.

## Program-level definition of done

- Each idea is implemented as the smallest of: native-command documentation, a thin resolver/passthrough, or a DevArch-specific check; unnecessary scripts are explicitly not created.
- Every implemented wrapper documents and tests its exact native command mapping.
- Recording-command tests prove native argv, stdio, TTY, signals, and exit status are preserved.
- No custom lifecycle engine, status/log parser, prune engine, registry client, database client, certificate implementation, browser launcher, or completion engine exists.
- Existing WordPress, Laravel, and hosts regression suites still pass.
- A full host-only test command performs no container, network, hosts-file, browser, certificate-store, or database mutations.
- README documents discovery rules, safety behavior, supported platforms, and direct native equivalents.
- ShellCheck and syntax checks pass for every added Bash file.

## Review order

Review the index and phase 01 first because every later command depends on those contracts. Then review each delivery wave in order; phases inside a wave may be reprioritized after the foundation exists.
