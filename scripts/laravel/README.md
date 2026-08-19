# Laravel bootstrap

`scripts/laravel/bootstrap.sh` creates a fresh Laravel application at `apps/<app-name>` and configures it for DevArch's shared PHP container and wildcard `.test` proxy.

## Requirements

- Bash 4+
- Podman with `podman compose`/`podman-compose`, or Docker with Compose
- `awk`, `tr`, `od`, and `sha256sum` or `shasum`
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
scripts/laravel/bootstrap.sh demo --profile standard
scripts/laravel/bootstrap.sh demo --profile loaded --seed
scripts/laravel/bootstrap.sh demo --package 'vendor/package:^1.0'
scripts/laravel/bootstrap.sh demo --dev-package 'vendor/tool:dev-main'
cp scripts/laravel/packages.example /tmp/laravel-packages
# Edit /tmp/laravel-packages, then run:
scripts/laravel/bootstrap.sh demo --packages-file /tmp/laravel-packages
scripts/laravel/bootstrap.sh demo --version '^12.0'
scripts/laravel/bootstrap.sh demo --force
scripts/laravel/bootstrap.sh demo --force --dry-run
```

Provisioning defaults to an isolated MariaDB database/user, migrations enabled, Laravel's `log` mailer, file cache, and synchronous queues. `--database sqlite` avoids MariaDB. `--with-mailpit` selects `mailpit:1025`. `--with-redis` uses `phpredis`, `redis:6379`, shared Compose authentication, Redis DB `0`, and cache DB `1`.

Application names must be 1–63 lowercase ASCII letters, digits, or interior hyphens; they cannot begin or end with a hyphen. Version constraints accept numeric versions/wildcards, comparison/caret/tilde operators, stability suffixes, `dev-*` branches, hyphen ranges, comma/space intersections, and `|`/`||` unions; malformed constraints fail during preflight.

## Profiles and packages

The default `bare` profile adds no optional services. `standard` enables Mailpit; `loaded` enables Mailpit and Redis. List profiles and their first description comments with:

```bash
scripts/laravel/bootstrap.sh --list-profiles
```

Profiles accept only `feature mailpit`, `feature redis`, `composer-package vendor/package [constraint]`, and `composer-dev-package vendor/package [constraint]`. Canonical v1 profiles contain no Composer packages.

`--package` and `--dev-package` accept `vendor/package` or `vendor/package:constraint` and are repeatable. A packages file accepts only `composer-package` and `composer-dev-package` directives. Blank lines, full-line comments, and trailing comments are ignored. Copy `scripts/laravel/packages.example`, uncomment and replace its placeholders, then pass the edited copy with `--packages-file`; the repository example selects no packages by default. Profile entries are applied first, packages-file entries second, and CLI additions last, preserving order. Production and development requirements are installed separately inside the PHP container. Package names, constraints, directive fields, and features are validated before provisioning; profile/package text is never evaluated as shell. Composer authentication remains in the user's existing Composer configuration.

Existing targets are refused unless `--force` is supplied. Before any provisioning mutation, the script creates a durable recovery guard under `apps/.devarch-recovery/`. Forced replacement preserves the prior tree beneath `apps/.devarch-backups/`; post-backup failures quarantine a partial target, remove only database resources created by that run, and restore the old tree. An incomplete recovery keeps the guard so later retries are refused, even if the detailed marker update fails.

`--dry-run` validates the request and runtime/Compose provider, then prints a redacted ordered plan without filesystem, container, network, database, Composer, or Artisan mutation.

`--help` and `--list-profiles` must each be used alone and do not require provisioning dependencies.

Run persistent focused checks with `scripts/laravel/tests/bootstrap_test.sh`.
