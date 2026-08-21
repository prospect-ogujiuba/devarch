# WordPress development scripts

`scripts/wordpress/bootstrap.sh` provisions local WordPress sites in `apps/<site-name>` using DevArch's shared containers and wildcard proxy. It does not run `wp server` or create a container per site.

## What the bootstrap does

A successful run performs this sequence:

1. Loads the repository `.env`, validates arguments, and resolves a site name.
2. Selects Podman or Docker and a Compose provider.
3. Creates `microservices-net` when it does not exist.
4. Starts the shared `php`, `mariadb`, and Nginx Proxy Manager services, optionally rebuilding PHP.
5. Waits up to 90 seconds for WP-CLI, MariaDB, and Nginx Proxy Manager.
6. Protects and moves an existing site when replacement is requested.
7. Creates a dedicated database and database user.
8. Downloads WordPress and writes `wp-config.php` without logging passwords.
9. Applies the selected profile and any extra plugins.
10. For Maker-enabled profiles, records core revisions and atomically publishes project-owned theme, blocks, and app workspaces.
11. Makes `wp-content` writable for local PHP-FPM use.
12. Optionally restores an All-in-One WP Migration `.wpress` archive.
13. Idempotently registers `127.0.0.1 <site-name>.test` in the system hosts file.

The resulting site is served by Nginx Proxy Manager at `https://<site-name>.test`. PHP requests are sent to `php:9000` over `microservices-net`, and MariaDB is reached as `mariadb`.

## Requirements

Run the script from this repository or invoke it by absolute path from an existing site.

- Podman or Docker
- `podman compose`, `podman-compose`, or `docker compose`
- Git; host SSH credentials or a Git credential helper for private repositories
- PHP on the host when using a MakerMaker profile (used to generate the portable Galaxy launcher)
- The trusted local wildcard certificate files used by the repository's Nginx Proxy Manager Compose service
- `sudo` on Linux/macOS, or Windows PowerShell/UAC under WSL or Git Bash, for automatic hosts-file registration

Copy the environment template before the first real run:

```bash
cp .env.example .env
```

## Configuration

The script sources `<repository>/.env` before resolving its settings. Values exported by the caller are used when the file does not replace them; WordPress-specific variables take precedence over shared admin values.

| Variable | Purpose | Default / requirement |
| --- | --- | --- |
| `WP_ADMIN_USER` / `ADMIN_USER` | WordPress administrator login | `admin` |
| `WP_ADMIN_PASSWORD` / `ADMIN_PASSWORD` | WordPress administrator password | Required for a real run |
| `WP_ADMIN_EMAIL` / `ADMIN_EMAIL` | WordPress administrator email | `admin@devarch.test` |
| `MARIADB_ROOT_PASSWORD` | Root password passed to the MariaDB service and client | `devarch` |
| `GITHUB_USER` | Owner used by profiles and `--github-plugin` | Required for those features; the placeholder `github-user` is rejected |
| `AIOWM_GIT_URL` | Native-CLI All-in-One WP Migration repository used by `--restore` | `git@github.com:$GITHUB_USER/all-in-one-wp-migration.git` |
| `CONTAINER_RUNTIME` | Force `podman` or `docker` instead of auto-detection | Podman first, then Docker |
| `WORDPRESS_CONTAINER_USER` | User passed to WP-CLI and Composer inside PHP | `0:0` for Podman; host UID/GID for Docker |

Passwords are supplied to WP-CLI through standard input. Dry runs redact both WordPress and MariaDB secrets.

## Quick start

```bash
scripts/wordpress/bootstrap.sh my-site
```

This creates:

- document root: `apps/my-site`
- database: `wp_my_site`
- database user: `wp_my_site` (truncated to 32 characters when necessary)
- site URL: `https://my-site.test`
- admin URL: `https://my-site.test/wp-admin`
- title: `My Site`

The database password is generated per run and written only to `wp-config.php`.

Customize the title or URL:

```bash
scripts/wordpress/bootstrap.sh my-site \
  --title "My Company" \
  --url https://wordpress.custom.test
```

Site names must match `[a-z0-9][a-z0-9-]{0,59}`. URLs must start with `http://` or `https://` and contain no spaces.

## CLI reference

| Option | Behavior |
| --- | --- |
| `-t, --title TITLE` | Set the site title; otherwise title-case the site name. |
| `-u, --url URL` | Set `home`/`siteurl`; otherwise use `https://<site-name>.test`. |
| `--profile NAME` | Load `bare`, `clean`, `custom`, `loaded`, or another profile file. |
| `--preset NAME` | Backward-compatible alias for `--profile`. |
| `--list-profiles` | Print discovered profile names/descriptions and exit. |
| `-p, --plugin SOURCE` | Add and activate a plugin; repeatable. |
| `--github-plugin NAME` | Clone and activate `git@github.com:$GITHUB_USER/NAME.git`; repeatable. |
| `--plugins-file FILE` | Read additional plugin sources, one per line. |
| `-r, --restore FILE` | Replace/install the site, then restore a `.wpress` archive. |
| `--build` | Pass `--build` while starting the PHP Compose service. |
| `-f, --force` | Move an existing site aside and reset its database. |
| `--no-hosts` | Skip automatic registration of `127.0.0.1 <site-name>.test`. |
| `--scaffolds-only` | Provision/refuse Maker workspaces in an existing site without running full bootstrap. Requires a Maker-enabled profile. |
| `--project-slug SLUG` | Override the lowercase kebab-case workspace prefix derived from the site name. |
| `--php-namespace NS` | Override the generated PHP namespace (default `Maker\\<SiteName>`). |
| `--js-namespace NS` | Override the lowercase block namespace (default project slug). |
| `--dry-run` | Validate and print core installs and workspace creation separately without changing hosts, files, containers, databases, activation, or lock state. |
| `-h, --help` | Print built-in help. |
| `--no-server` | Deprecated no-op retained for compatibility; shared proxy routing is always used. |

Preview a complete plan safely:

```bash
scripts/wordpress/bootstrap.sh my-site --profile clean --dry-run
```

## Base WordPress configuration

Every new installation:

- sets `FS_METHOD` to `direct`;
- disables year/month upload folders and removes empty upload subdirectories;
- sets permalinks to the **Post name** structure (`/%postname%/`);
- deletes the default post;
- deletes Akismet and Hello Dolly when present;
- grants local read/write access to `wp-content` so PHP can manage plugins, themes, uploads, and migration state without FTP prompts.

Git components are shallow-cloned on the host. If a cloned plugin, must-use plugin, or theme contains `composer.json`, `composer install --no-interaction` runs inside the PHP container as the configured container user.

## Profiles

List profiles or select one during installation:

```bash
scripts/wordpress/bootstrap.sh --list-profiles
scripts/wordpress/bootstrap.sh my-site --profile clean
```

| Profile | Installed components |
| --- | --- |
| `bare` | All-in-One WP Migration, installed but inactive. |
| `clean` | TypeRocket Pro v6 as an MU plugin; MakerMaker, MakerBlocks, and MakerStarter as framework core; three project workspaces; All-in-One WP Migration inactive; Admin Site Enhancements Pro active. |
| `custom` | Everything in `clean`, plus Manual Image Crop. |
| `loaded` | Everything in `custom`, plus Debug Bar, Debug Bar Actions and Filters Addon, Classic Editor, Default Featured Image, Plugin Inspector, Log Deprecated Notices, Query Monitor, Theme Check, WordPress Beta Tester, Show Current Template, Theme Inspector, and View Admin As. |

Repositories named by a profile are cloned over SSH from `GITHUB_USER`. All components are active unless a profile marks a Git plugin `inactive`. `clean`, `custom`, and `loaded` include the same `maker-stack.fragment`, so their Maker core/workspace declarations cannot drift. MakerStarter is installed as core, but the generated `<site>-theme` child theme is the final active theme.

### Maker core and project workspaces

A new Maker-enabled site contains:

```text
wp-content/themes/makerstarter/          # framework core
wp-content/themes/<site>-theme/          # PROJECT OWNED — EDIT HERE
wp-content/plugins/makerblocks/          # framework core
wp-content/plugins/<site>-blocks/        # PROJECT OWNED — EDIT HERE
wp-content/plugins/makermaker/           # framework core
wp-content/plugins/<site>-app/           # PROJECT OWNED — EDIT HERE
```

Core roots carry `CORE-BOUNDARY.md` with `FRAMEWORK CORE — DO NOT EDIT; update from playground releases`; workspace root READMEs contain `PROJECT OWNED — EDIT HERE`. Each generated project workspace is initialized as an independent Git repository on `main`; no initial commit or remote is created. The app plugin is generated through `wp makermaker create`. DevArch renders MakerStarter's maintained `scaffolds/child-theme/` and MakerBlocks' `scaffolds/project-plugin/` into sibling staging directories, then atomically publishes them; interrupted staging is rolled back before the destination becomes visible.

Existing workspace destinations are never merged, deleted, or overwritten. Provisioning logs a refusal for each existing workspace, then may activate it; all existing bytes remain unchanged. To provision independently in an existing site:

```bash
scripts/wordpress/bootstrap.sh my-site --profile clean --scaffolds-only
```

The default project slug is the site name, the PHP namespace is `Maker\\<SiteName>`, and the JS block namespace is the project slug. Use `--project-slug`, `--php-namespace`, and `--js-namespace` for exceptional names. Invalid values fail validation before provisioning.

`.devarch-maker.lock` records each core repository URL, exact tag when available (otherwise commit), commit, package type, and UTC install time. An existing manifest is preserved. Dry-run labels core installs and workspaces but performs no clone, scaffold publication, activation, or manifest write.

### Maker releases and core synchronization

Playground remains the integration site and owns the three nested core repositories. `maker-stack.json` maps a reviewed stack version or channel to exact package tags and commits; `maker-stack.schema.json` defines the contract. See [MAKER-RELEASES.md](MAKER-RELEASES.md) for semantic-version policy, compatibility/changelog conventions, and the gated release checklist.

Preview or apply a core-only consumer update:

```bash
scripts/wordpress/sync-maker.sh my-site --profile loaded --dry-run  # profile's stable channel
scripts/wordpress/sync-maker.sh my-site --profile loaded --to 1.0.0
```

The sync command refuses dirty repositories, unknown origins, non-semantic refs unless `--allow-main` is explicit, ref/commit mismatches, missing health files, and failed declared dependency installation. It stages full replacements, retains rollback metadata, atomically updates `.devarch-maker.lock`, and automatically restores prior core directories and lock state after a publication failure. Project workspaces, uploads, and databases are outside its target set.

Release maintainers run `release-maker.sh VERSION --dry-run`, then the same command without `--dry-run` only after all gates pass. It creates local tags and updates the manifest but deliberately never pushes.

The three package directories inside `apps/playground/wp-content/` are independent repositories and their **primary Git worktrees**. Keep each on clean `main`; `release-maker.sh` enforces this. They are not consumer deployments, so `sync-maker.sh` refuses `playground` rather than replacing those worktrees with detached release clones.

Use linked worktrees only for parallel core feature branches, outside the WordPress tree:

```bash
# Replace makerblocks with makerstarter or makermaker as needed.
cd apps/playground/wp-content/plugins/makerblocks
git worktree add "$HOME/projects/worktrees/makerblocks/my-feature" \
  -b feat/my-feature
```

Commit and test in the linked checkout, merge the branch from the playground primary `main`, then remove it with `git worktree remove <path>` and `git branch -d <branch>`. `git worktree list` shows every checkout sharing that package repository. A branch can be active in only one worktree. Site-owned code never belongs in these core feature worktrees; use `<site>-theme`, `<site>-blocks`, or `<site>-app`. Each package README lists its exact merge/test commands.

### Existing-site ownership audit and migration

Inventory all detected Maker sites without changing them:

```bash
scripts/wordpress/audit-maker.sh --all
scripts/wordpress/audit-maker.sh my-site --sync-ready --runtime-check
```

`audit-maker.sh` reports dirty/untracked core paths with ownership destinations, unknown remotes, missing or mismatched locks, missing workspace markers, absent independent backup/versioning receipts, and activation failures. `sync-maker.sh` runs the sync-readiness audit and refuses updates when it fails.

Prepare a site with a durable core-diff/untracked backup, refusal-safe workspaces, exact lock, synchronization, and runtime audit:

```bash
scripts/wordpress/migrate-maker.sh my-site --profile loaded --to stable --dry-run
scripts/wordpress/migrate-maker.sh my-site --profile loaded --to 0.1.0
```

Dirty or untrusted core stops after preservation and workspace preparation; it is never reset automatically. See [MIGRATE-MAKER-SITES.md](MIGRATE-MAKER-SITES.md) for ownership decisions, Site Editor handling, evidence requirements, and the disposable-copy → pilot → remaining-sites rollout.

### TypeRocket and MakerMaker integration

The `clean`, `custom`, and `loaded` profiles additionally:

- install `typerocket-pro-v6` under `wp-content/mu-plugins/` and copy its root entry file into the MU-plugin directory;
- install MakerMaker as a normal plugin;
- generate executable `galaxy` and `galaxy-config.php` files in the site root from MakerMaker's `GalaxyContext`;
- load MakerMaker's idempotent `typerocket_galaxy_commands` runtime filter;
- backfill MakerMaker's plugin-specific Galaxy context with `register-plugin-galaxy`.

This makes MakerMaker's Galaxy commands, including `make:maker-resource`, available without editing TypeRocket configuration. Bootstrap and release integration checks require the TypeRocket repository to remain clean.

### Extending profiles

Profiles are line-oriented files in `scripts/wordpress/profiles/<name>.profile`. The first `# ` comment becomes the description shown by `--list-profiles`.

Supported directives are:

```text
include fragment-name
maker-stack-channel channel-name
maker-core plugin|theme repository-name
maker-workspace child-theme|blocks-plugin|app-plugin core-repository-name
github-plugin repository-name [active|inactive]
github-theme repository-name
github-mu-plugin repository-name
wp-plugin wordpress-org-slug
```

Repository names are resolved as `git@github.com:$GITHUB_USER/<repository-name>.git`. Unknown directives, unsafe names, extra fields, and unsupported activation values stop the run before provisioning.

## Additional plugins

`--plugin` accepts:

```bash
# WordPress.org; a bare slug is equivalent to wp:<slug>
scripts/wordpress/bootstrap.sh my-site --plugin wp:query-monitor
scripts/wordpress/bootstrap.sh my-site --plugin query-monitor

# Git over SSH or HTTPS
scripts/wordpress/bootstrap.sh my-site \
  --plugin git:git@github.com:owner/private-plugin.git \
  --plugin git:https://github.com/owner/public-plugin.git

# Repository beneath GITHUB_USER
scripts/wordpress/bootstrap.sh my-site --github-plugin makerblocks
```

The target directory is derived from the repository filename, minus `.git`. Existing plugin directories are not recloned, but activation and applicable Composer setup still run. Command-line plugins are activated; profiles may explicitly leave Git plugins inactive.

A plugin file supports blank lines, full-line comments, and trailing `#` comments:

```bash
scripts/wordpress/bootstrap.sh my-site \
  --plugins-file scripts/wordpress/plugins.example
```

Credential-bearing HTTPS URLs such as `https://token@github.com/...` are rejected. Use SSH or a configured Git credential helper instead.

## Replacing an existing site

A normal run refuses to overwrite `apps/<site-name>`:

```bash
scripts/wordpress/bootstrap.sh my-site --force
```

With `--force`, the site directory is moved to:

```text
apps/.devarch-backups/<site-name>-YYYYMMDD-HHMMSS
```

The matching database is dropped and recreated with a newly generated password. This is a local safety move, not a database export; use the restore workflow when a native WordPress backup is required.

## Restoring a `.wpress` archive

```bash
scripts/wordpress/bootstrap.sh my-site --restore /absolute/or/relative/site.wpress
```

Restore requires `AIOWM_GIT_URL` or a usable `GITHUB_USER`. DevArch deliberately installs that Git repository rather than the WordPress.org package because the established repository exposes native `wp ai1wm backup` and `wp ai1wm restore`; the WordPress.org build gates CLI restore behind the Unlimited Extension.

For an existing WordPress target, the script:

1. Installs/activates the native-CLI AIOWM repository in the old site if necessary.
2. Creates writable `wp-content/ai1wm-backups` and plugin `storage` directories.
3. Runs `wp ai1wm backup` before replacement.
4. Moves the old site to `apps/.devarch-backups/`.
5. Installs a fresh site and its requested profile/plugins.
6. Copies the archive to the new `wp-content/ai1wm-backups/` directory.
7. Runs `wp ai1wm restore <archive-name>`.
8. Normalizes restored `home` and `siteurl` values and flushes the cache.

`--restore` implies permission to replace the target, so `--force` is not required. If the input archive is inside the site being replaced, it is first copied to `apps/.devarch-backups/imports/`.

The archive must exist, use `.wpress`, and have a filename containing only letters, numbers, dots, underscores, and hyphens. AIOWM's native CLI does not support multisite restores.

When called from anywhere beneath an existing `apps/<site-name>` WordPress tree, the site name can be inferred:

```bash
cd apps/my-site/wp-content/plugins
/path/to/devarch/scripts/wordpress/bootstrap.sh \
  --restore /backups/my-site.wpress
```

Discovery requires `wp-config.php`, and the detected WordPress root must be directly below this repository's `apps/` directory.

## Runtime and troubleshooting

- Set `CONTAINER_RUNTIME=docker` or `podman` to bypass auto-detection.
- Podman defaults WP-CLI to `--user 0:0`, which maps bind-mounted ownership through rootless Podman. Docker defaults to the host UID/GID.
- Set `WORDPRESS_CONTAINER_USER` only when the service image or bind-mount ownership requires another mapping.
- The script creates `microservices-net`, but Nginx Proxy Manager must already be configured to route `*.test` document roots to the shared PHP service.
- A first run may take longer while service images and WordPress are downloaded. Subsequent runs reuse images and persistent container volumes.
- A readiness timeout usually means the `php` container lacks WP-CLI, MariaDB rejected `MARIADB_ROOT_PASSWORD`, or the expected container names were changed.

## Tests

Run the shell regression suites without provisioning a real site:

```bash
bash scripts/wordpress/bootstrap.test.sh
bash scripts/wordpress/audit-maker.test.sh
bash scripts/wordpress/migrate-maker.test.sh
bash scripts/wordpress/sync-maker.test.sh
bash scripts/wordpress/release-maker.test.sh
```

They cover help/profile output, validation, site discovery, secret-safe dry runs, Podman/Docker user mapping, plugin sources, shared Maker profile composition, workspace markers and hash-stable refusal, interrupted publication rollback, Galaxy integration, native AIOWM backup/restore planning, ownership audit/classification, migration backup and dirty-core stop behavior, exact core synchronization, dirty/unknown/ref mismatch refusal, workspace hash preservation, automatic failure rollback, explicit retained rollback, and gated local tag/manifest publication.

## Current Maker site

The local architecture test site is `apps/maker-site` at `https://maker-site.test`. Its administrator credentials are stored with mode `0600` in `apps/maker-site/.devarch-admin-credentials`.

```bash
# MakerBlocks
cd apps/maker-site/wp-content/plugins/makerblocks
npm install
npm start
npm run build
npm test
npm run lint
npm run plugin-zip
npm run create:block -- testimonial
npm run create:block -- listing --dynamic

# MakerStarter
cd apps/maker-site/wp-content/themes/makerstarter
npm install
npm test
```
