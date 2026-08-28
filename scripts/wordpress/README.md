# WordPress development scripts

`scripts/wordpress/bootstrap.sh` provisions local WordPress sites in `apps/<site-name>` using DevArch's shared PHP, MariaDB, and Nginx Proxy Manager services.

## Quick start

```bash
cp .env.example .env
scripts/wordpress/bootstrap.sh my-site
```

The site is available at `https://my-site.test`; its document root is `apps/my-site` and its database is `wp_my_site`.

Requirements:

- Podman or Docker with Compose
- Git
- the repository's local wildcard certificate files
- `sudo` or Windows UAC for automatic hosts-file registration

## Configuration

The script sources the repository `.env`.

| Variable | Purpose | Default / requirement |
| --- | --- | --- |
| `WP_ADMIN_USER` / `ADMIN_USER` | Administrator login | `admin` |
| `WP_ADMIN_PASSWORD` / `ADMIN_PASSWORD` | Administrator password | Required outside dry runs |
| `WP_ADMIN_EMAIL` / `ADMIN_EMAIL` | Administrator email | `admin@devarch.test` |
| `MARIADB_ROOT_PASSWORD` | MariaDB root password | `devarch` |
| `GITHUB_USER` | Owner used by profiles and `--github-plugin` | Required for private GitHub plugins |
| `AIOWM_GIT_URL` | All-in-One WP Migration repository used by `--restore` | `git@github.com:$GITHUB_USER/all-in-one-wp-migration.git` |
| `CONTAINER_RUNTIME` | Force `podman` or `docker` | Auto-detected |
| `WORDPRESS_CONTAINER_USER` | User for WP-CLI and Composer | `0:0` for Podman; host UID/GID for Docker |

Passwords are supplied to WP-CLI through standard input and are redacted in dry runs.

## Options

```text
-t, --title TITLE
-u, --url URL
    --profile NAME
    --list-profiles
-p, --plugin SOURCE
    --github-plugin NAME
    --plugins-file FILE
-r, --restore FILE
    --build
-f, --force
    --no-hosts
    --dry-run
```

Run `scripts/wordpress/bootstrap.sh --help` for details.

A plugin source may be a WordPress.org slug (`query-monitor` or `wp:query-monitor`) or a Git URL prefixed with `git:`. Credential-bearing HTTPS URLs are rejected. Plugin files accept one source per line with blank lines and `#` comments.

## Profiles

```bash
scripts/wordpress/bootstrap.sh --list-profiles
scripts/wordpress/bootstrap.sh my-site --profile clean
```

| Profile | Contents |
| --- | --- |
| `bare` | All-in-One WP Migration, inactive. |
| `clean` | Bare plus Admin Site Enhancements Pro. |
| `custom` | Clean plus Manual Image Crop. |
| `loaded` | Custom plus common WordPress development and debugging plugins. |

Profile files live in `scripts/wordpress/profiles/`. Supported directives are:

```text
include profile-fragment
github-plugin repository [active|inactive]
github-theme repository
github-mu-plugin repository
wp-plugin wordpress-org-slug
```

GitHub repository names resolve under `GITHUB_USER`. Git components are cloned on the host; Composer runs inside the PHP container when `composer.json` exists.

## WordPress defaults

New sites use direct filesystem access, post-name permalinks, flat uploads, and writable local `wp-content`. The default post, Akismet, and Hello Dolly are removed.

## Replacing or restoring a site

`--force` moves an existing site to `apps/.devarch-backups/` before recreating its database and files.

```bash
scripts/wordpress/bootstrap.sh my-site --force
scripts/wordpress/bootstrap.sh my-site --restore /backups/site.wpress
```

Restore uses the native All-in-One WP Migration CLI. Existing sites receive a safety backup first. Archives inside a replaced site are staged outside it before replacement. Restored `home` and `siteurl` values are normalized to the requested URL.

When invoked below an existing `apps/<site-name>` WordPress tree, the script can infer the site name from `wp-config.php`.

## Runtime notes

- The script creates `microservices-net` and starts the shared PHP, MariaDB, and proxy services.
- Podman defaults WP-CLI to `0:0`; Docker uses the host UID/GID.
- Use `--no-hosts` when host registration is managed separately.
- Use `--dry-run` to validate and print a secret-safe plan.

## Tests

```bash
bash scripts/wordpress/bootstrap.test.sh
```
