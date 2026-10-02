# DevArch CLI, Bash Migration, and TUI Plan

- Date: 2026-10-02
- Status: Proposed

## Goal

Give DevArch one front door, a `devarch` binary, that makes the existing service library feel like Laragon: start services by name, switch versions, create projects, and see everything from one interactive screen. The binary is a thin layer over native Podman Compose. It does not replace the service library, the proxy, or the bootstraps.

## Lessons this plan is built on

DevArch has already been rebuilt twice:

1. **V1:** a database-backed Go API, plus a React dashboard that edited per-instance overrides.
2. **V2:** a manifest-first engine with a daemon, a local API, and plan/apply.

Both were removed on 2026-05-16 (`cfeaa91`, about 40k lines). Each time, the plain compose files and the tested scripts were what survived. This plan grows from those surviving pieces and treats the guardrails below as hard limits.

## Guardrails (non-goals)

- No daemon, no HTTP API, no database, no background polling.
- No desired-state engine, plan/apply, or lock files. Compose files are the desired state.
- No GUI or TUI editing of compose overrides. Changing a service means editing its file.
- No per-subcommand Podman wrappers and no reformatting of native output. Every mutating action runs one visible `podman compose …` or `podman …` command, which `--dry-run` prints.
- Podman by default; Docker is a setting, not an adapter layer. The differences live in one table (`cli/internal/engine/runtime.go`), and everything else runs the configured binary with identical arguments.
- State lives in plain files under `~/.config/devarch/`.

If a feature needs something on this list, it is out of scope.

## Architecture

```text
cli/                          Go module (separate go.mod so Go tooling never walks apps/)
  cmd/devarch/main.go         entry point
  internal/root/              locate the DevArch checkout
  internal/catalog/           discover services, read x-devarch metadata, resolve names
  internal/runner/            the only place that executes processes (fakeable in tests)
  internal/compose/           build podman compose argv + env for a service
  internal/state/             ~/.config/devarch files: config, versions.env
  internal/hosts/             managed hosts block: Linux, macOS, WSL→Windows
  internal/health/            wait for healthchecks / readiness probes
  internal/recipe/            run recipes, parse progress events
  internal/doctor/            environment checks
  internal/lint/              catalog contract validation
  internal/tui/               Bubble Tea application
recipes/                      recipe manifests (entry points remain under scripts/)
```

**Libraries:**
- `spf13/cobra` for commands and shell completion.
- `charmbracelet/bubbletea`, `bubbles` and `lipgloss` for the TUI.
- `gopkg.in/yaml.v3` for reading compose metadata.

compose-go is deliberately not used. DevArch only reads a few fields and leaves interpolation to Podman Compose, so the two never disagree. When fully resolved values are needed (doctor, lint), call `podman compose config`.

**Finding the root**, in order:
1. The `DEVARCH_ROOT` environment variable.
2. `root` in `~/.config/devarch/config.yml`.
3. Walk up from the current directory looking for `services-library/`.

The first successful walk-up offers to save `root` to the config file.

**Process execution:** everything goes through `runner.Runner`, an interface with `Run(ctx, Cmd) error` and `Exec(Cmd)`. The real implementation logs the shell-escaped argv, the same way `devarch_run` does, and execs the final native process where possible, so TTY, signals and exit codes stay native. Tests use a recording fake.

## Service contract (`x-devarch`)

Compose ignores top-level `x-` keys, so metadata lives in the existing `compose.yml`:

```yaml
x-devarch:
  title: PostgreSQL
  description: Relational database
  tags: [database, sql]
  version:
    var: POSTGRES_VERSION
    default: "18.2"
    choices: ["18.2", "17", "16"]
    rebuild: false            # true for services built from a Dockerfile (php)
  urls:
    - https://postgres.test    # optional; shown by `open` and the TUI
  ready:                       # optional; defaults to the compose healthcheck
    exec: ["pg_isready", "-U", "postgres"]
  requires: []                 # other catalog IDs started first by `up`
services:
  postgres:
    image: postgres:${POSTGRES_VERSION:-18.2}
```

**Rules:**
- Every field is optional. A service without `x-devarch` still works with `up`, `down`, `logs` and `ls`; it just isn't parametric yet.
- The variable named in `version.var` must appear in the file as `${VAR:-default}`, so `podman compose` from the service directory keeps working without DevArch.
- For built services, the Dockerfile takes the version as an `ARG`, and compose passes it through `build.args`. For example, `ARG PHP_VERSION=8.4` with `FROM php:${PHP_VERSION}-fpm`.

**Selected versions** live in `~/.config/devarch/versions.env`, written by `devarch use`. When running compose, the binary sets them as process environment variables, not with `--env-file`. Shell environment takes precedence over a service's own `.env` for interpolation and leaves that `.env` loading untouched; GLPI and OpenProject keep credentials in their `.env` files.

**User library:** catalog search paths are `[~/.config/devarch/services, <root>/services-library]`, and the first match for an ID wins. The search path is built in Phase 1 but only documented and supported from Phase 6.

> **As built (configuration, 2026-10-02):** `config.yml` gained six optional keys whose defaults are the behavior above: `runtime`, `apps_dir`, `network`, `hosts.manage`, `hosts.address` and `editor`. `devarch config` shows effective values with their source, and `get`/`set` change only these keys. Compose files read the network and apps directory as `${DEVARCH_NETWORK:-microservices-net}` and `${DEVARCH_APPS_DIR:-../../../apps}`, which DevArch sets only when they differ from the defaults. The `.test` suffix and Nginx Proxy Manager stay fixed (see `cli/README.md`).

## Command surface

| Command | Behavior |
|---|---|
| `devarch` | Open the TUI when stdout is a TTY; otherwise print help. |
| `devarch ls [query] [--tag t] [--json]` | Catalog with title, tags, selected version and running state. |
| `devarch up <svc>... [--wait] [--no-hosts]` | Ensure `microservices-net`, start `requires` first, `compose up -d`, optionally wait for readiness, and register hosts. |
| `devarch down <svc>... [--volumes]` | `compose down`. `--volumes` asks for confirmation and names the volumes it will delete. |
| `devarch restart <svc>...` | `compose restart`. |
| `devarch ps [--json]` | Running DevArch containers, matched to catalog IDs. |
| `devarch logs <svc> [-f]` | `podman logs`, exec'd natively. |
| `devarch use <svc> [version]` | No version: list the choices. With a version: write `versions.env`, rebuild if `rebuild: true`, recreate if running. |
| `devarch new [recipe] [name] [flags]` | Run a recipe. No arguments in a TTY opens the TUI wizard. |
| `devarch open <svc\|app> [--editor\|--shell]` | Browser URL, `code <dir>`, or `podman exec -it <c> sh`. |
| `devarch hosts sync\|add\|remove\|list` | Managed hosts block (replaces `scripts/hosts/*`). |
| `devarch doctor` | Podman, compose provider, rootless user, network, port conflicts, certs, hosts, root path. |
| `devarch lint` | Validate every compose file, `x-devarch` fields, host-port collisions and unpinned images. Replaces the Python snippet in the README. |
| `devarch compose <svc> -- <args>` | Raw passthrough with the version environment applied. |
| `devarch <name>` (unknown) | Exec `devarch-<name>` from PATH, git-style. |

**Global flags:** `--dry-run` prints the native commands and changes nothing. `--json` is accepted only where listed above.

**Names:** a service can be referred to by its full ID (`database/postgres`) or a unique short name (`postgres`), matching the existing `devarch_catalog_resolve` semantics.

## Recipes and the progress protocol

This is the bridge that lets the bootstraps stay in Bash while the binary and TUI drive them.

**Manifest** (`recipes/wordpress/recipe.yml`):

```yaml
name: wordpress
title: WordPress site
entry: ../../scripts/wordpress/bootstrap.sh
requires: [backend/php, database/mariadb, proxy/nginx-proxy-manager]
ready:
  backend/php: { exec: ["wp", "--info"] }
args:
  name: { positional: true, required: true, pattern: "^[a-z0-9][a-z0-9-]*$" }
  profile: { flag: --profile, choices_from: ../../scripts/wordpress/profiles/*.profile, default: bare }
  title: { flag: --title }
  restore: { flag: --restore, type: path }
  force: { flag: --force, type: bool, confirm: "Back up and replace the existing site?" }
```

**Run sequence:**
1. Resolve and validate args, prompting in the TUI and failing fast in the CLI.
2. Start `requires` and wait for readiness. This replaces each script's `ensure_network`, `start_services` and `wait_for_services`.
3. Exec `entry`, passing the user's flags through unchanged plus these environment variables:
   - `DEVARCH_ROOT`
   - `DEVARCH_MANAGED=1`: tells the script to skip its own platform steps.
   - `DEVARCH_PROGRESS_FD=3`
4. Read JSON lines from fd 3, for example `{"step":"install_plugins","state":"start","message":"Installing 4 plugins"}`. Valid `state` values are `start`, `done`, `warn` and `fail`.
5. Register the host and print the URL.

> **As built (Phase 4):** the two-mode design above was simplified. Bootstraps always perform their platform steps by calling `devarch up <services> --wait --no-hosts` and `devarch hosts add`, whether they are run directly or through `devarch new`, so there is one code path and no `DEVARCH_MANAGED` flag. `devarch new` validates arguments, sets `DEVARCH_BIN` and `DEVARCH_ROOT`, and execs the entry script; environment such as `MARIADB_ROOT_PASSWORD` reaches compose because the script passes it to `devarch`. `requires` in `recipe.yml` is descriptive (shown by `devarch new` and the TUI). The progress helper lives in `scripts/devarch/lib/platform.sh`, not `common.sh`, so bootstraps can source it without the catalog library.

**Bash side:** add `devarch_progress STEP STATE MESSAGE` to `scripts/devarch/lib/common.sh`. It writes to fd 3 only when `DEVARCH_PROGRESS_FD` is set and is a no-op otherwise. **Every bootstrap keeps working when run directly**, and its existing `*.test.sh` suites keep passing throughout.

## Moving logic out of Bash

Bash code moves only when it is platform logic duplicated across scripts, or when it blocks the CLI or TUI. Domain logic stays in Bash until there is a concrete trigger.

### Layer 1: platform (moves first, Phases 1–4)

| Bash today | Go owner | Phase | Then |
|---|---|---|---|
| `common.sh`: `devarch_require_podman`, `devarch_require_compose`, `devarch_run` | `runner`, `doctor` | 1 | Library stays for scripts until Phase 4. |
| `catalog.sh`: list, resolve, compose_file, validation | `catalog`, `lint` | 1 | `catalog.sh` becomes `devarch ls --json`/`devarch compose` calls, then is deleted. |
| `detect_runtime` (wordpress, laravel, node) | `runner` | 1 | Deleted from scripts in Phase 4. |
| `ensure_network` (wordpress, laravel, node) | `compose` (`up`) | 1 | Same. |
| `start_services` / `compose_up` | `compose` (`up`) | 1 | Same. |
| `wait_for_services` (90×1s exec loops) | `health` + recipe `ready` | 1 | Same. |
| `project-management/manage.sh` | `devarch up --tag project-management`, `devarch ls --tag … ` (urls) | 1 | Script deleted. |
| `sync-hosts.sh`, `register-host.sh`, `register-host.ps1`, `register_*_host` | `hosts` (embeds the PowerShell helper and uses the existing `/init powershell.exe` WSL route) | 3 | `sync-hosts.*` deleted in Phase 3 with its cases ported to Go. `register-host.*` stays until Phase 4 because the bootstraps call it directly. |
| `dotenv.sh` | `state` (reads the root `.env` with the same data-only rules) | 4 | Kept while any Bash consumer remains. |
| `print_plan` / dry-run plumbing | `--dry-run` on every command | 1–4 | Bootstrap dry-run keeps covering only its own filesystem steps. |

**Phase 4 cut-over:** once the binary is the supported entry point, the bootstraps drop their copies of the Layer 1 functions and call `devarch up … --wait` and `devarch hosts add` instead. The duplication is removed for real, not just hidden behind the binary. (Done: WordPress, Laravel and the Node runtime now require the `devarch` CLI. Their ad hoc Docker detection was removed; Docker returned later as the `runtime` setting, which they receive as `DEVARCH_RUNTIME` and `DEVARCH_CONTAINER_USER`.)

### Layer 2: shared provisioning (Phase 6, only once Layer 1 has landed)

| Bash today | Go owner | Why move |
|---|---|---|
| `generate_db_password`, `reset_database`, `create_database`, `db_exec`, `db_query`, `assert_database_available` | `devarch db create\|drop <name> --engine mariadb\|postgres` | Duplicated across WordPress and Laravel, useful on its own, and gives recipes one tested path. |
| `choose_unique_path`, `backup_existing_site`, `move_existing_target`, `create_recovery_guard`, `rollback` | `internal/replace` with `devarch app backup` | Data-safety logic that deserves typed tests. Laravel's recovery guard becomes the shared model. |
| WordPress/Laravel `.profile` directive parsing | `recipe` parses into typed lists, shown in the TUI before running | Line-oriented directives, easy to parse. Lets the wizard preview exactly what a profile installs. |
| `node/bootstrap.sh` (app runtime container + router) | `devarch app start\|stop <app>` | Mostly platform: network, container, host. |

### Layer 3: domain logic (stays in Bash)

`wp_exec`, plugin/theme/MU-plugin installation, AIOWM backup and restore, Composer component installs, `artisan`, Laravel package handling, and the JavaScript framework profiles. The JavaScript `.profile` files define `configure_app()` functions with inline Node; they are code, not config, and should stay that way.

**Triggers for moving a Layer 3 piece:**
1. The TUI needs structured data the progress protocol can't carry.
2. There is a recurring bug class rooted in Bash quoting or parsing.
3. Native Windows or macOS support needs behavior the script can't provide.

Write down which trigger applied in the PR that moves it.

## TUI

**Launch:** `devarch` with no arguments, using Bubble Tea and full-screen alt-screen.

```text
┌ DevArch ─ Services  Apps  Running  Doctor ──────────────── /filter ┐
│ ● backend/php            8.4   running  healthy    php.test          │
│ ● database/mariadb       11.8  running  healthy                      │
│ ○ database/postgres      18.2  stopped                               │
│ ○ database/redis         8     stopped                               │
│ ● proxy/nginx-proxy-mgr  2     running  healthy    :81               │
│ ○ mail/mailpit           1.27  stopped             mailpit.test      │
│   … 169 more  (tab: group by category)                               │
├──────────────────────────────────────────────────────────────────────┤
│ PostgreSQL · database, sql · versions 18.2 17 16                     │
│ services-library/database/postgres/compose.yml                       │
├──────────────────────────────────────────────────────────────────────┤
│ $ podman compose -f …/postgres/compose.yml up -d                     │
│ u up  d down  r restart  v version  l logs  o open  n new  ? help  q │
└──────────────────────────────────────────────────────────────────────┘
```

**Views:**
- **Services:** the full catalog. Typing `/` filters by name, tag or category, and Tab toggles a grouped view. The detail pane shows metadata and the compose file path.
- **Apps:** directories under `apps/` with detected framework, URL, and app runtime state (from Phase 6). Keys: `o` opens, `e` opens the editor, `s` opens a shell.
- **Running:** `podman ps` for containers on `microservices-net`, including non-catalog ones.
- **Doctor:** results of `devarch doctor`, each with a fix hint and the command to run.
- **Logs:** a full-screen follow of `podman logs -f` with search. Esc returns.
- **Version picker (`v`):** lists `version.choices`, marks the current one, and confirms "recreate now?" if the service is running.
- **New wizard (`n`):** the steps are:
  1. Choose a recipe.
  2. Name it, validated live.
  3. Choose a profile, with a preview of what it installs (from Phase 6 profile parsing).
  4. Set options.
  5. Confirm on a screen listing the exact commands.
  6. Watch a progress view fed by fd 3 events, with raw output collapsible underneath.

**Behavior rules:**
- **Live state without polling:** while the TUI is open, it streams `podman events --format json --filter type=container` and updates rows on events. Nothing runs after you quit.
- **One code path:** every action calls the same internal functions as the CLI. The footer always shows the native command being run.
- **No editing:** the TUI never edits compose files or overrides.
- **Graceful degradation:** if Podman is unavailable, the catalog and apps still render and the Doctor tab is selected.

> **As built (Phase 5):**
> - Logs are followed inside the screen (streamed, filterable, Esc stops the process) rather than by suspending it, so a Ctrl-C meant for the log never reaches the TUI.
> - Hostname registration is never run in the background, because `sudo` cannot prompt inside a full-screen app: `up` and the wizard pass `--no-hosts`/`no_hosts_flag`, report missing hostnames, and `H` runs `devarch hosts sync` interactively.
> - The grouped-by-category toggle and the `s` shell key were left out; the shell remains `devarch compose <svc> -- exec …`.
> - Views are tested by driving the model directly with synthetic messages (including a real recipe run with progress events) instead of `teatest`; the screen was also exercised in an isolated tmux server.
> - Container refreshes are coalesced (one `podman ps` in flight, one queued) because `podman ps` takes one to five seconds on WSL.

## Phases

Each phase can ship on its own and has a clear exit check.

### Phase 0: contract groundwork (no Go)
- Fix the host-port collisions on `127.0.0.1:8091` and `127.0.0.1:9216`.
- Add `x-devarch` and a version variable to the core set: php, mariadb, postgres, redis, memcached, nginx-proxy-manager, mailpit, node, adminer, phpmyadmin.
- Turn the PHP base image into `ARG PHP_VERSION` and pass it from compose `build.args`.
- **Exit check:** `podman compose up -d` still works unchanged in each touched directory, and `PHP_VERSION=8.3 podman compose build` produces PHP 8.3.

### Phase 1: binary core
- Root discovery, catalog with the search path, `runner`, `compose`, `health`.
- `ls`, `up`, `down`, `restart`, `ps`, `logs`, `compose`, external commands, completion, `--dry-run`, `--json`.
- Replace `manage.sh` with tags.
- **Exit check:** on a fresh rootless Podman user, `devarch up php mariadb nginx-proxy-manager --wait` brings the shared stack up, creates the network automatically, and `--dry-run` prints exactly the commands that ran.

### Phase 2: versions
- `use`, `versions.env`, rebuild/recreate logic, versions shown in `ls`.
- **Exit check:** `devarch use php 8.3` followed by `podman exec php php -v` reports 8.3, and running `podman compose up -d` directly still gives the default.

### Phase 3: hosts, doctor, lint
- Go `hosts` module with Linux/macOS sudo and the WSL→Windows PowerShell route.
- `up` registers hosts.
- Port the `sync-hosts.test.sh` and `register-host.test.sh` cases to Go.
- Add `doctor` and `lint`, and replace the README's Python validation snippet with `devarch lint`.
- **Exit check:** hosts tests pass in Go, `sync-hosts.*` is deleted (`register-host.*` follows in Phase 4, once the bootstraps stop calling it), and `devarch lint` exits 0 on a clean checkout.

### Phase 4: recipes
- Add `recipe.yml` for wordpress, laravel and javascript.
- Add the progress protocol and `devarch_progress` in `common.sh`, plus `DEVARCH_MANAGED` handling.
- Remove the duplicated Layer 1 functions from the bootstraps.
- Make `devarch new` the documented entry point.
- **Exit check:** all existing `bootstrap.test.sh` suites pass, `devarch new wordpress shop --profile clean` produces a working `https://shop.test`, and running `bootstrap.sh` directly still works.

### Phase 5: TUI
- Services, Running, Logs, Doctor, the version picker, and the new-app wizard with progress.
- **Exit check:** every Phase 1–4 workflow can be done from the TUI without typing a command, and `teatest` golden tests cover each view.

### Phase 6: provisioning and extensibility
- Layer 2 moves: `db`, `replace`, profile parsing with wizard previews, and `app start|stop` replacing `node/bootstrap.sh`.
- The Apps view.
- Document and support the user service library and external commands.
- Optionally, the dashboard reads `devarch ls --json` and `devarch ps --json` instead of doing its own discovery.
- **Exit check:** `scripts/node/` is deleted, WordPress and Laravel create their databases through `devarch db create`, and a user service in `~/.config/devarch/services/` shows up in `ls` and the TUI.
- **As built (`db`):** `devarch db create|drop` covers MariaDB and PostgreSQL with `--existing fail|reuse|replace`; replace dumps the old database and saves the user's definition (password hash and grants) so it can be restored. Identifier derivation stays in each bootstrap, which is domain logic.
- **As built (replace):** `internal/replace` owns the Laravel recovery model for both bootstraps. `devarch app guard|release|recover` operate on a JSON record at the old marker path, `db create --app` adds databases to it, and recovery saves each finished step so it can resume. WordPress gained rollback; its `--force` no longer loses the replaced database. `devarch app backup` copies a project with `cp -a`.
- **Status (2026-10-02):** the extensibility half is done: user services and recipes overlay the built-in ones, `devarch-<name>` commands run, and all three are documented and tested. The Layer 2 moves (`db`, `replace`, profile previews, `app start|stop`) have not started; they rewrite data-safety code in the WordPress and Laravel bootstraps and should follow real use of Phases 1–5.

## Testing

- **Go unit tests:** every command builds its argv through the recording fake runner, with golden tests for compose argv and environment.
- **Fixtures:** catalog fixture trees, including invalid, ambiguous, overlay and missing-metadata cases.
- **Hosts:** table tests for rewriting the managed block, idempotency, and preserving content outside the markers.
- **TUI:** `teatest` golden output per view and key path.
- **Existing suites:** `scripts/devarch/tests/run-tests.sh` and every `*.test.sh` stay green until the script they cover is deleted. Delete a script only after its cases are ported.
- **Integration:** an opt-in `make integration` target that runs Phase 1–4 exit checks against real rootless Podman.

## Distribution

- **Now:** `go install ./cli/cmd/devarch` from the checkout, with `devarch completion zsh|bash|fish`.
- **Later:** GoReleaser builds for linux/darwin × amd64/arm64 and a Homebrew tap. Version info is embedded at build time.

## Decisions taken in this plan

- **Go module in `cli/`, not the repo root.** Keeps Go tooling out of `apps/` and keeps the repo's identity as a service library.
- **YAML for `x-devarch` and `recipe.yml`.** One format, the same as compose.
- **Versions passed as process environment, not `--env-file`.** This preserves per-service `.env` credentials and avoids differences in multi-env-file support between Compose providers.
- **Podman by default, Docker as a setting.** Originally Podman only. Reversed on 2026-10-02: `runtime: docker` in `config.yml` switches the compose command, existence checks, readiness, the bootstrap container user, `ps`/events parsing and the doctor checks, all from one table. No per-runtime code paths exist outside it and the bootstraps’ `DEVARCH_RUNTIME`/`DEVARCH_CONTAINER_USER`.

## Open questions

1. **Readiness defaults:** use the compose healthcheck everywhere and add `ready.exec` only where it isn't enough (as with WP-CLI on php)? Recommended: yes.
2. **Where `requires` lives:** in `x-devarch` rather than compose `depends_on`, because the services are separate compose projects. Confirm.
3. **IT-management runbooks and GLPI production material:** move them to a separate repository before Phase 1 so this plan's scope stays the dev environment.
