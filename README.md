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

The restored WordPress workflow starts the PHP and MariaDB services, creates an isolated database, downloads WordPress, installs it, and optionally clones plugin repositories. Sites use the existing Nginx Proxy Manager wildcard routing and PHP-FPM container rather than a separate `wp server` process:

```bash
cp .env.example .env                    # first run only; change the credentials
scripts/wordpress/bootstrap.sh my-site
# ensure 127.0.0.1 my-site.test resolves locally, then open https://my-site.test
```

Select one of the recovered Git-repository profiles with `--profile` (historical `--preset` is also accepted):

```bash
scripts/wordpress/bootstrap.sh my-site --profile clean
scripts/wordpress/bootstrap.sh --list-profiles
```

- `bare` — All-in-One WP Migration.
- `clean` — TypeRocket Pro, MakerMaker, MakerBlocks, MakerStarter, All-in-One WP Migration, and Admin Site Enhancements Pro.
- `custom` — `clean` plus Manual Image Crop.
- `loaded` — `custom` plus the historical WordPress.org development/debugging plugins.

Profile Git repositories are resolved under `GITHUB_USER` and cloned over SSH. Every TypeRocket Pro v6 profile installs MakerMaker and MakerBlocks as regular plugins, and MakerStarter as the active theme; bundled inactive themes are then removed. Every new site deletes WordPress’s default post and configures direct, writable local plugin/theme management without FTP prompts. All-in-One WP Migration remains inactive. Profile files live in `scripts/wordpress/profiles/` and can be reviewed or extended without editing the installer.

Install additional WordPress.org or Git plugins in the same run:

```bash
scripts/wordpress/bootstrap.sh my-site \
  --plugin wp:query-monitor \
  --plugin git:git@github.com:your-user/private-plugin.git

# Or maintain one source per line:
scripts/wordpress/bootstrap.sh my-site \
  --plugins-file scripts/wordpress/plugins.example
```

Private repositories use the host's Git/SSH credentials; tokens are not embedded in clone URLs. Use `--github-plugin NAME` with `GITHUB_USER` for repositories under one GitHub account. Existing sites are preserved unless `--force` is explicit, and forced replacements are moved to `apps/.devarch-backups/`. Run `scripts/wordpress/bootstrap.sh --help` for URL, build, and dry-run options.

The wildcard proxy resolves `<site-name>.test`, serves `apps/<site-name>`, and passes PHP directly to `php:9000` over `microservices-net`. The first run may take longer while the PHP image builds. Warm runs reuse the image and container volumes.

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
