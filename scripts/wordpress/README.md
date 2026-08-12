# WordPress development scripts

This directory contains the DevArch WordPress bootstrap workflow. It creates sites under `apps/` and serves them through the existing infrastructure:

- Nginx Proxy Manager handles wildcard `*.test` HTTPS routing.
- PHP-FPM runs in the `php` container and is reached as `php:9000` over `microservices-net`.
- MariaDB runs in the `mariadb` container.
- Nginx maps `<site-name>.test` to `apps/<site-name>`.

The bootstrap does **not** start a separate `wp server` process.

## Initial setup

Run commands from the DevArch repository root.

```bash
cp .env.example .env
```

Set at least the following values in `.env`:

```dotenv
ADMIN_USER=admin
ADMIN_PASSWORD=choose-a-secure-password
ADMIN_EMAIL=you@example.com
MARIADB_ROOT_PASSWORD=devarch
GITHUB_USER=prospect-ogujiuba
```

Add each site hostname to `/etc/hosts` or your local DNS:

```text
127.0.0.1 my-site.test
```

The local wildcard certificate must also be trusted by your browser if you want to avoid a certificate warning.

## Create a site

The default WordPress URL is `https://<site-name>.test`.

```bash
scripts/wordpress/bootstrap.sh my-site
```

This creates `apps/my-site`, creates an isolated database and database user, installs WordPress, and configures its URL as:

```text
https://my-site.test
```

Open:

```text
https://my-site.test
https://my-site.test/wp-admin
```

Set a custom title:

```bash
scripts/wordpress/bootstrap.sh my-site --title "My Company"
```

Set a non-default URL when required:

```bash
scripts/wordpress/bootstrap.sh my-site --url https://custom.test
```

## Preview the plan

A dry run validates arguments and prints the planned operations without changing files, containers, or databases:

```bash
scripts/wordpress/bootstrap.sh my-site --dry-run
```

## Profiles

List available profiles:

```bash
scripts/wordpress/bootstrap.sh --list-profiles
```

Create a site with a profile:

```bash
scripts/wordpress/bootstrap.sh my-site --profile bare
scripts/wordpress/bootstrap.sh my-site --profile clean
scripts/wordpress/bootstrap.sh my-site --profile custom
scripts/wordpress/bootstrap.sh my-site --profile loaded
```

Profiles are defined under `scripts/wordpress/profiles/`:

- `bare` — minimal WordPress with the historical migration plugin.
- `clean` — TypeRocket Pro, the three custom packages, migration, and core development repositories.
- `custom` — `clean` plus accessible image tooling.
- `loaded` — `custom` plus WordPress.org development and debugging plugins.

Every profile containing TypeRocket Pro v6 (`clean`, `custom`, and `loaded`) also installs MakerMaker and MakerBlocks as regular plugins, and MakerStarter as the active theme. `bare` does not install these packages.

`--preset` remains an alias for `--profile`.

## Install plugins

Install a WordPress.org plugin:

```bash
scripts/wordpress/bootstrap.sh my-site --plugin wp:query-monitor
```

A bare WordPress.org slug is also accepted:

```bash
scripts/wordpress/bootstrap.sh my-site --plugin query-monitor
```

Install multiple plugins:

```bash
scripts/wordpress/bootstrap.sh my-site \
  --plugin wp:query-monitor \
  --plugin wp:debug-bar
```

Clone a Git plugin:

```bash
scripts/wordpress/bootstrap.sh my-site \
  --plugin git:git@github.com:owner/plugin.git
```

Clone a repository under `GITHUB_USER`:

```bash
scripts/wordpress/bootstrap.sh my-site --github-plugin makerblocks
```

Load sources from a file:

```bash
scripts/wordpress/bootstrap.sh my-site \
  --plugins-file scripts/wordpress/plugins.example
```

Private repositories use the host's Git and SSH credentials. Credential-bearing HTTPS URLs are rejected.

## Replace an existing site

The bootstrap refuses to overwrite an existing site unless `--force` is explicit:

```bash
scripts/wordpress/bootstrap.sh my-site --force
```

The previous directory is moved under `apps/.devarch-backups/`, and the isolated site database is reset.

## Rebuild PHP

Rebuild the PHP image before starting the services:

```bash
scripts/wordpress/bootstrap.sh my-site --build
```

## CLI reference

```bash
scripts/wordpress/bootstrap.sh --help
```

## Tests

Run the bootstrap regression suite:

```bash
bash scripts/wordpress/bootstrap.test.sh
```

The suite covers help and profile output, validation, secret-safe dry runs, runtime user mapping, plugin sources, and recovered profile contents.

## Current Maker site

The local architecture test site is at:

```text
apps/maker-site
https://maker-site.test
```

Its local administrator credentials are stored with mode `0600` at:

```text
apps/maker-site/.devarch-admin-credentials
```

### MakerBlocks

```bash
cd apps/maker-site/wp-content/plugins/makerblocks
npm install
npm start
npm run build
npm test
npm run lint
npm run plugin-zip
```

Create another block:

```bash
npm run create:block -- testimonial
npm run create:block -- listing --dynamic
```

### MakerStarter

```bash
cd apps/maker-site/wp-content/themes/makerstarter
npm install
npm test
```
