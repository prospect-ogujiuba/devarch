# Environment Configuration Cleanup — Initiative Specification r1

- Topic: `env-configuration-cleanup`
- Revision: 1
- Status: draft
- Created: 2026-09-11T16:14:46Z
- Predecessor: none

## Problem

The root `.env.example` describes 21 values, but only WordPress/Laravel bootstrap flows intentionally consume a small subset and only MariaDB Compose interpolates one of them. `scripts/wordpress/bootstrap.sh` shell-sources the entire ignored `.env` with automatic export, allowing shell evaluation and leaking unrelated values to child processes. The template retains removed API, proxy, database, domain, and CORS settings and implies that it centrally configures the Compose catalog.

## Outcome and observable behavior

1. Repository dotenv input is parsed as data, never evaluated as shell code.
2. Each bootstrap accepts only an explicit allowlist and preserves documented file/inherited/default precedence.
3. Unrelated dotenv entries are ignored and are not newly exported to child processes.
4. `.env.example` lists only supported bootstrap inputs, uses non-secret placeholders, and clearly separates host bootstrap configuration from Compose container configuration.
5. Existing `.env` remains the default local path and existing supported aliases continue to work.
6. MariaDB's root password remains the only shared host/Compose credential and continues to be passed explicitly by bootstrap scripts.

## Users

Local DevArch operators running WordPress or Laravel bootstrap scripts with Docker or Podman, plus maintainers updating service definitions and tests.

## Constraints

- Preserve Bash compatibility and the current Docker/Podman behavior.
- Do not read, rewrite, migrate, or commit a developer's ignored `.env`.
- Preserve the current Laravel dotenv quoting/comment/control-byte behavior for supported keys.
- Preserve current documented WordPress keys and precedence.
- Secrets must not be placed in tracked Compose files, command logs, dry-run output, or test output.
- Canonical work is limited to scripts, tests, `.env.example`, and relevant README files.

## Non-goals

- Moving all variables into Compose files.
- Renaming `.env` to `.env.local` in this initiative.
- Introducing a secrets manager, encrypted secret format, interactive prompting, or production deployment scheme.
- Changing service ports, images, volumes, networks, application credentials, or database schema.
- Cleaning unrelated hard-coded development credentials elsewhere in the service catalog.

## Compatibility

- Root `.env` remains optional and ignored by Git.
- Existing `ADMIN_*` WordPress aliases remain accepted alongside `WP_ADMIN_*`.
- `MARIADB_ROOT_PASSWORD`, `GITHUB_USER`, `AIOWM_GIT_URL`, runtime selection, and container-user settings retain their documented behavior.
- Existing Laravel supported keys retain file-over-inherited precedence.
- Removed template keys may remain in an existing local `.env`; they are ignored.

## Migration and rollback

Migration is documentation-only for local files: users compare their existing `.env` against the reduced template and may delete obsolete entries. No automatic mutation occurs. Rollback restores the prior scripts/template/docs; no persistent service or application data changes are involved. Existing containers initialized with a MariaDB password still require that same value after rollback or upgrade.

## Risks

- Parser behavior drift could break quoted or escaped values.
- Export behavior changes could reveal undocumented reliance on unrelated variables.
- A blank/example MariaDB password could conflict with an already initialized volume.
- Tests can be contaminated by a developer's real root `.env` unless they use an explicit fixture seam.

## Acceptance criteria

- **AC1:** A dotenv line containing command substitution or shell syntax is never executed by either bootstrap.
- **AC2:** Only each bootstrap's allowlisted keys can alter its configuration; unrelated keys remain ignored and unexported.
- **AC3:** Supported quoting, comments, empty values, control-byte rejection, and file precedence are covered by host-only tests.
- **AC4:** WordPress and Laravel dry-run behavior and secret redaction remain unchanged for supported inputs.
- **AC5:** `.env.example` contains no unconsumed legacy API/CORS/domain/proxy/database variables and no usable password/token defaults.
- **AC6:** Documentation explains root dotenv scope, Compose interpolation versus container runtime environment, supported aliases, migration, and MariaDB volume/password compatibility.
- **AC7:** Focused and existing regression suites pass without requiring a container runtime or reading the developer's real `.env`.

## Open blockers

None.
