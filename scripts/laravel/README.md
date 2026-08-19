# Laravel bootstrap

`scripts/laravel/bootstrap.sh` creates a fresh Laravel application in `apps/<app-name>` with DevArch's shared PHP-FPM, database, mail, Redis, and wildcard `.test` infrastructure. It does not create a per-application container or start queue workers, the scheduler, Reverb, or a Vite development server.

## Architecture

The host script validates every input before provisioning, selects existing Compose services, and runs Composer and Artisan inside the shared `php` container. The repository `apps/` directory is mounted there at `/var/www/html`, so `apps/demo` is `/var/www/html/demo` in the container.

The wildcard proxy resolves `<app-name>.test` and sends PHP to `php:9000` over `microservices-net`. It automatically selects `apps/<app-name>/public` as the document root when `public/index.php` (or `public/index.html`) exists. A normal Laravel scaffold therefore routes through `public/index.php` without per-app proxy configuration.

MariaDB applications receive an isolated database and user derived from the app name. SQLite applications use `database/database.sqlite`. Mailpit and Redis are optional shared services. The script makes Laravel's runtime directories writable by the shared PHP container, writes the selected settings to Laravel's `.env`, generates `APP_KEY`, and optionally runs migrations and the seeder.

## Prerequisites

- Bash 4+
- Podman with `podman compose` or `podman-compose`, or Docker with Compose
- `awk`, `tr`, `od`, and either `sha256sum` or `shasum`
- `openssl` or `/dev/urandom` for real MariaDB provisioning
- the Compose definitions under `services-library/`
- a wildcard proxy, trusted local certificate, and local DNS/hosts routing for `*.test`
- access to Composer package sources from the shared PHP container

The runtime user must be able to create `microservices-net`, start the selected services, and write beneath `apps/`. Do not mix rootless container users: their containers and networks are isolated.

## Quick start

```bash
cp .env.example .env                    # first run; replace example credentials
scripts/laravel/bootstrap.sh demo --dry-run
scripts/laravel/bootstrap.sh demo
# Resolve demo.test to 127.0.0.1, then open https://demo.test
```

Defaults are the `bare` profile, MariaDB, migrations enabled, no seeding, Laravel's `log` mailer, file cache, and synchronous queues.

Use SQLite for a scaffold without MariaDB:

```bash
scripts/laravel/bootstrap.sh demo-sqlite --database sqlite
```

## CLI

```text
scripts/laravel/bootstrap.sh <app-name> [options]
```

| Option | Behavior |
| --- | --- |
| `--version CONSTRAINT` | Pass a validated Composer constraint to `laravel/laravel`. Current stable is used when omitted. |
| `--profile NAME` | Load `bare`, `standard`, or `loaded`; default `bare`. |
| `--list-profiles` | List profiles and exit. Must be the sole argument. |
| `--package PACKAGE` | Add a production Composer requirement; repeatable. |
| `--dev-package PACKAGE` | Add a development Composer requirement; repeatable. |
| `--packages-file FILE` | Read package directives from a declarative file. |
| `--database mariadb\|sqlite` | Select the database; default `mariadb`. |
| `--with-mailpit` | Enable Mailpit and configure SMTP at `mailpit:1025`. |
| `--with-redis` | Enable Redis for cache and queues. |
| `--migrate` | Run migrations; this is the default. |
| `--no-migrate` | Skip migrations. It cannot be combined with `--migrate`. |
| `--seed` | Run `db:seed` once, independently of the migration choice. |
| `--force` | Preserve and replace an existing target. |
| `--dry-run` | Validate runtime and configuration, then print a redacted mutation-free plan. |
| `--help` | Show help and exit. Must be the sole argument. |

Application names are 1–63 lowercase ASCII letters, digits, or interior hyphens. They cannot begin or end with a hyphen. Options from later versions (`--repository`, `--branch`, `--database-import`, `--npm-install`, `--build`, and `--npm-build`) are intentionally rejected.

## Configuration and precedence

The script reads only these keys from the repository `.env`:

| Key | Purpose | Default |
| --- | --- | --- |
| `LARAVEL_APP_NAME` | Laravel `APP_NAME` display value, not the app directory name. | Title-cased app name |
| `LARAVEL_APP_URL` | Laravel `APP_URL`. | `https://<app-name>.test` |
| `LARAVEL_DB_ROOT_PASSWORD` | MariaDB administrative password used during provisioning. | `MARIADB_ROOT_PASSWORD`, then `devarch` |
| `MARIADB_ROOT_PASSWORD` | Fallback MariaDB administrative password. | `devarch` |
| `LARAVEL_CONTAINER_USER` | Explicit numeric `uid:gid` for container commands. | Runtime-specific mapping |
| `CONTAINER_RUNTIME` | Force `podman` or `docker`. | Auto-detect Podman, then Docker |

Repository `.env` assignments are loaded after CLI parsing and replace inherited values for the same supported key. Command-line options control only their documented settings; there are no CLI forms for `APP_NAME`, `APP_URL`, the root password, runtime, or container user. For the root password, `LARAVEL_DB_ROOT_PASSWORD` takes precedence over `MARIADB_ROOT_PASSWORD` after `.env` loading.

Values may be unquoted, single-quoted, or double-quoted. Quoted values may have trailing comments. Escaped matching quotes and backslashes are decoded, and a dollar sign may be escaped in a double-quoted value. The parser does not evaluate shell syntax. Supported assignments containing ASCII control bytes or DEL are rejected; unrelated `.env` keys are ignored.

## Profiles

```bash
scripts/laravel/bootstrap.sh --list-profiles
scripts/laravel/bootstrap.sh demo --profile standard
scripts/laravel/bootstrap.sh demo --profile loaded --seed
```

- `bare` — selected database, no optional service.
- `standard` — `bare` plus Mailpit.
- `loaded` — `bare` plus Mailpit and Redis.

Profiles are declarative files in `scripts/laravel/profiles/`. Allowed directives are:

```text
feature mailpit
feature redis
composer-package vendor/package [constraint]
composer-dev-package vendor/package [constraint]
```

Only `mailpit` and `redis` are valid features. Canonical v1 profiles install no Composer packages.

## Composer package syntax

`--package` and `--dev-package` accept `vendor/package` or `vendor/package:constraint` and may be repeated:

```bash
scripts/laravel/bootstrap.sh demo \
  --package 'vendor/runtime:^1.0' \
  --dev-package 'vendor/tool:dev-main'
```

A packages file accepts only `composer-package` and `composer-dev-package` directives. The constraint is separated from the package by whitespace:

```text
composer-package vendor/runtime ^1.0
composer-dev-package vendor/tool dev-main
```

Blank lines, full-line comments, and trailing comments are ignored. Start from `scripts/laravel/packages.example`; it selects no packages until placeholders are uncommented and replaced. Profile entries are applied first, packages-file entries second, and CLI entries last, preserving source order. Production and development packages are installed in separate Composer commands.

Package names must use lowercase Composer `vendor/package` syntax. Constraints support numeric versions and wildcards, comparison/caret/tilde operators, stability suffixes, `dev-*` branches, hyphen ranges, comma/space intersections, and `|`/`||` unions. Malformed names, constraints, directive fields, controls, and NUL bytes fail during preflight. Declarative text is never evaluated as shell; Composer authentication remains in the user's Composer configuration.

## Services and runtime mapping

The script prefers Podman when no runtime is selected, then Docker. It uses the runtime's integrated `compose` command; Podman may fall back to `podman-compose`.

- Podman defaults container commands to `--user 0:0`, which maps correctly for the rootless bind mount.
- Docker defaults to the host's numeric `uid:gid`.
- `LARAVEL_CONTAINER_USER=<uid>:<gid>` overrides either default.

PHP is always selected. MariaDB is selected by the default database, Mailpit by `standard`, `loaded`, or `--with-mailpit`, and Redis by `loaded` or `--with-redis`. Redis uses `phpredis`, host `redis`, port `6379`, the shared Compose password, queue/default DB `0`, and cache DB `1`. Dry-run output redacts database and Redis credentials.

## Replacement and recovery safety

An existing `apps/<app-name>` is refused unless `--force` is supplied. Before provisioning mutation, the script creates a durable guard in `apps/.devarch-recovery/`. Forced replacement moves the old tree beneath `apps/.devarch-backups/` rather than deleting it.

If provisioning fails after backup, the partial target is quarantined beneath `apps/.devarch-failed/`, database resources created by that run are removed, and the prior tree is restored. If recovery cannot complete, the guard remains and blocks retries for manual inspection. `--dry-run --force` only reports the chosen backup path; it does not move the target or create recovery state.

## Regression tests

The default regression suite is host-only. It copies the bootstrap assets into a trap-owned temporary project root, so repository `.env` values and existing `apps/` workspaces cannot affect it. It replaces Podman/Docker with rejecting fakes and permits only `compose version`, so it cannot create containers, networks, databases, or applications:

```bash
bash -n scripts/laravel/bootstrap.sh scripts/laravel/bootstrap.test.sh
bash scripts/laravel/bootstrap.test.sh
```

Focused parser and rollback-contract checks are also available at `scripts/laravel/tests/bootstrap_test.sh`. They use temporary fixtures and fake commands; they do not perform real provisioning.

## Optional real-provisioning smoke test

This is intentionally separate from the default regression suite because it starts services, downloads Laravel, and creates a real database and application. Use a disposable name and the default `bare`/MariaDB profile:

```bash
scripts/laravel/bootstrap.sh laravel-smoke --dry-run
scripts/laravel/bootstrap.sh laravel-smoke
test -f apps/laravel-smoke/public/index.php
curl -fk https://laravel-smoke.test/
```

After inspection, remove the disposable database/user with the active runtime and then remove the application. For Podman:

```bash
podman exec mariadb sh -c 'mariadb -uroot -p"$MARIADB_ROOT_PASSWORD" -e "DROP DATABASE IF EXISTS laravel_laravel_smoke; DROP USER IF EXISTS '\''lv_laravel_smoke'\''@'\''%'\'';"'
rm -rf apps/laravel-smoke
```

Review the commands and confirm the name is disposable before running them. Use `docker exec` instead when Docker owns the services.

## Troubleshooting

- **`Podman or Docker is required` / no Compose provider:** install the runtime and Compose integration, or set `CONTAINER_RUNTIME` correctly.
- **Runtime cannot see an existing network/container:** run all DevArch services as the same rootless user.
- **Proxy returns 503:** start the shared PHP service and verify it is attached to `microservices-net`.
- **Wrong document root:** confirm `apps/<name>/public/index.php` exists; the wildcard proxy then selects `public` automatically.
- **MariaDB readiness/authentication fails:** make `LARAVEL_DB_ROOT_PASSWORD` (or its fallback) match the shared MariaDB service.
- **Target already exists:** use a different name or review `--force` replacement safety first.
- **Unresolved recovery marker:** inspect the matching recovery, backup, and failed paths; do not delete the guard until the application/database state is understood.
- **Composer authentication fails:** configure credentials in the Composer environment used by the shared PHP container.
- **Permission errors:** use the documented runtime mapping or a numeric `LARAVEL_CONTAINER_USER` that can write the bind mount.
