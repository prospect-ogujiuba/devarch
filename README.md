# DevArch

DevArch is now a simple local service library: a collection of Podman-compatible Compose service definitions plus application workspaces.

The old Go CLI, planning workflow, daemon/API code, and generated workspace surfaces have been removed. Use the compose files directly.

## Layout

- `services-library/<category>/<service>/compose.yml` — service compose definitions.
- `services-library/<category>/<service>/config/` — optional service configuration.
- `apps/<app>/` — application workspaces, typically separate repositories ignored by the top-level DevArch repository.
- `.env.example` — example environment values.

## Usage

Pick a service and run it with Podman and the configured Compose provider:

```bash
cd services-library/database/postgres
podman compose up -d
```

Systems configured with the standalone provider may use `podman-compose up -d` instead. Do not mix rootless users: containers and networks created by one user are not visible to another.

Most compose files attach to the external network `microservices-net`. Create it once as the same service user that runs the stack:

```bash
podman network create microservices-net
```

For persistent production services, manage the containers with systemd/Quadlet (or reviewed generated units) and enable lingering for the rootless service user.

## Rapid WordPress bootstrap

`scripts/wordpress/bootstrap.sh` creates local sites in `apps/<site-name>` using the shared PHP-FPM, MariaDB, and Nginx Proxy Manager infrastructure. It creates `microservices-net` when needed, starts the PHP and MariaDB Compose services, waits for WP-CLI/database readiness, creates an isolated database and user, installs WordPress, and applies profiles or additional plugins. It never starts a separate `wp server` process.

```bash
cp .env.example .env                    # first run; replace example credentials
scripts/wordpress/bootstrap.sh my-site
# map 127.0.0.1 my-site.test locally, then open https://my-site.test
```

Each base install sets `FS_METHOD=direct`, disables date-based upload folders, deletes the default post and bundled sample plugins, and makes `wp-content` writable by the shared PHP container. The default database is `wp_<site_name>`, its password is generated per run and written only to `wp-config.php`, and the default URL/title are derived from the site name.

Useful operations:

```bash
scripts/wordpress/bootstrap.sh my-site --dry-run
scripts/wordpress/bootstrap.sh my-site --title "My Site" --url https://custom.test
scripts/wordpress/bootstrap.sh my-site --build
scripts/wordpress/bootstrap.sh my-site --force
scripts/wordpress/bootstrap.sh --list-profiles
```

`--dry-run` validates and prints a secret-safe plan. `--build` rebuilds PHP before service startup. `--force` moves an existing directory to `apps/.devarch-backups/<site>-<timestamp>` and recreates its database; without it, existing sites are preserved.

### Profiles and plugins

```bash
scripts/wordpress/bootstrap.sh my-site --profile clean
scripts/wordpress/bootstrap.sh my-site \
  --plugin wp:query-monitor \
  --plugin git:git@github.com:your-user/private-plugin.git \
  --plugins-file scripts/wordpress/plugins.example
```

- `bare` — All-in-One WP Migration, inactive.
- `clean` — TypeRocket Pro v6 as an MU plugin; MakerMaker and MakerBlocks as plugins; MakerStarter as the active theme; All-in-One WP Migration inactive; Admin Site Enhancements Pro active.
- `custom` — `clean` plus Manual Image Crop.
- `loaded` — `custom` plus 12 WordPress.org development and debugging plugins.

Profile repositories are shallow-cloned over SSH from `GITHUB_USER`; `--github-plugin NAME` provides the same shorthand for an extra active plugin. Git components with `composer.json` run Composer inside the PHP container. TypeRocket profiles also generate the site `galaxy` launcher/config, register MakerMaker's Galaxy commands idempotently, activate MakerStarter, and remove inactive bundled themes. `--preset` remains an alias for `--profile`.

### Restore workflow

```bash
scripts/wordpress/bootstrap.sh my-site --restore /path/to/site.wpress
```

For an existing WordPress target, restore first installs/activates the established native-CLI `GITHUB_USER/all-in-one-wp-migration` repository, prepares writable backup/storage directories, and runs `wp ai1wm backup`. It then moves the old site aside, performs a fresh installation, copies the archive into `wp-content/ai1wm-backups`, runs `wp ai1wm restore`, and normalizes `home` and `siteurl` to the requested local URL. `--restore` implies replacement, so `--force` is unnecessary. Archives located inside the replaced site are staged under `apps/.devarch-backups/imports/` first.

The WordPress.org AIOWM build is intentionally not used because it gates CLI restore behind its Unlimited Extension. Site name can be omitted when the command runs beneath an existing `apps/<site-name>` WordPress tree.

See [`scripts/wordpress/README.md`](scripts/wordpress/README.md) for prerequisites, environment-variable precedence, every option, exact profile contents/directives, plugin-source validation, runtime user mapping, restore safety details, troubleshooting, and tests. The wildcard proxy resolves `<site-name>.test`, serves `apps/<site-name>`, and sends PHP to `php:9000` over `microservices-net`.

## Development checks

```bash
# validate compose YAML structure
python - <<'PY'
from pathlib import Path
import yaml
for path in sorted(Path('services-library').glob('**/compose.yml')):
    data = yaml.safe_load(path.read_text())
    assert isinstance(data, dict) and 'services' in data, path
print('compose yaml ok')
PY

# check PHP syntax
php -l apps/serverinfo/index.php
```
