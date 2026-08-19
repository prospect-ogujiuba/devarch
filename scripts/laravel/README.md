# Laravel bootstrap

`scripts/laravel/bootstrap.sh` creates a fresh Laravel application at `apps/<app-name>` and configures it for DevArch's shared PHP container and wildcard `.test` proxy.

## Requirements

- Bash 4+
- Podman with `podman compose`/`podman-compose`, or Docker with Compose
- `awk`, `tr`, and `sha256sum` or `shasum`
- `openssl` or `/dev/urandom` for real MariaDB provisioning
- The shared Compose definitions under `services-library/`
- An available external port/runtime setup for the selected services

The script reads only its approved Laravel/runtime values from the repository `.env`. Supported overrides are:

- `LARAVEL_APP_NAME`
- `LARAVEL_APP_URL`
- `LARAVEL_DB_ROOT_PASSWORD` (then `MARIADB_ROOT_PASSWORD`)
- `LARAVEL_CONTAINER_USER`
- `CONTAINER_RUNTIME`

Supported values may be unquoted, single-quoted, or double-quoted. Quoted values may be followed by a comment; escaped matching quotes and backslashes are decoded, and double-quoted dollars may be escaped. Supported assignments reject ASCII controls and DEL rather than evaluating shell syntax.

## Core usage

```bash
scripts/laravel/bootstrap.sh demo
scripts/laravel/bootstrap.sh demo --database sqlite
scripts/laravel/bootstrap.sh demo --with-mailpit --with-redis --seed
scripts/laravel/bootstrap.sh demo --version '^12.0'
scripts/laravel/bootstrap.sh demo --force
scripts/laravel/bootstrap.sh demo --force --dry-run
```

Provisioning defaults to an isolated MariaDB database/user, migrations enabled, Laravel's `log` mailer, file cache, and synchronous queues. `--database sqlite` avoids MariaDB. `--with-mailpit` selects `mailpit:1025`. `--with-redis` uses `phpredis`, `redis:6379`, shared Compose authentication, Redis DB `0`, and cache DB `1`.

Application names must be 1–63 lowercase ASCII letters, digits, or interior hyphens; they cannot begin or end with a hyphen. Version constraints accept numeric versions/wildcards, comparison/caret/tilde operators, stability suffixes, `dev-*` branches, hyphen ranges, comma/space intersections, and `|`/`||` unions; malformed constraints fail during preflight.

Existing targets are refused unless `--force` is supplied. Before any provisioning mutation, the script creates a durable recovery guard under `apps/.devarch-recovery/`. Forced replacement preserves the prior tree beneath `apps/.devarch-backups/`; post-backup failures quarantine a partial target, remove only database resources created by that run, and restore the old tree. An incomplete recovery keeps the guard so later retries are refused, even if the detailed marker update fails.

`--dry-run` validates the request and runtime/Compose provider, then prints a redacted ordered plan without filesystem, container, network, database, Composer, or Artisan mutation.

`--help` must be used alone and does not require provisioning dependencies. `--list-profiles`, profile selection, and package options are reserved for phase 03 and fail clearly in this phase.

Run persistent focused checks with `scripts/laravel/tests/bootstrap_test.sh`.
