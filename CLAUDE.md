# DevArch

A local development environment: a library of plain Compose services (`services-library/`) with one front door, the `devarch` Go CLI and TUI (`cli/`). Users start services by name, switch versions, create WordPress, Laravel and JavaScript projects (`devarch new`), and watch everything in one screen. Read `README.md` for the user view, `cli/README.md` for every command, and `docs/plans/2026-10-02-devarch-cli-and-tui.md` (including its "As built" notes, which win over the original text) for the design.

## Guardrails (hard limits)

DevArch was rebuilt twice before. A database-backed API with a React override editor, then a daemon with plan/apply, were deleted in May 2026 (about 40k lines). The compose files and tested scripts survived. So:

- No daemon, HTTP API, database, background polling, desired-state engine, plan/apply or lock files. Compose files are the desired state.
- No GUI or TUI editing of compose files or overrides.
- Every mutating action runs one visible native command (`podman …`/`podman compose …`, or `docker` when configured), printed when it runs and by `--dry-run`. No per-subcommand wrappers, no reformatting of native output.
- State lives in plain files: `~/.config/devarch/config.yml` and `versions.env` (`DEVARCH_CONFIG_HOME` moves them), plus a project's guard record in `<apps_dir>/.devarch-recovery/` while a bootstrap runs or awaits recovery.

A feature that needs anything on this list is out of scope.

## Where logic lives

- `cli/internal/engine` — the operations shared by CLI and TUI: up/down/wait, versions, hosts registration, `db` (MariaDB/PostgreSQL), JavaScript app start/stop, and the Podman/Docker table in `runtime.go`. Runtime differences go in that table and nowhere else; other code reads its fields and never branches on the runtime name.
- `cli/internal/runner` — the only place processes start. `Run`/`Exec` are for mutations (echoed, skipped by `--dry-run`); `Output` is for read-only queries and always runs.
- `cli/internal/replace` — provisioning safety (guard, backup, recovery) shared by the bootstraps through `devarch app guard|release|recover`.
- `cli/internal/state` — config keys (`runtime`, `apps_dir`, `network`, `hosts.manage`, `hosts.address`, `editor`) and `versions.env`. The `.test` suffix and Nginx Proxy Manager are deliberately not settings.
- `cli/internal/cli` — cobra transport only; `cli/internal/tui` — Bubble Tea screen calling the same engine functions.
- `scripts/{wordpress,laravel,javascript}/bootstrap.sh` — domain logic stays in Bash (WP-CLI, Composer, Artisan, framework scaffolders, profiles). Platform steps call `devarch`: `up … --wait --no-hosts`, `hosts add`, `db create --app`, `app guard|release|recover`, `app start`. Container commands use `"$DEVARCH_RUNTIME" exec --user "$DEVARCH_CONTAINER_USER"` from `devarch_runtime_env`. Rules: `scripts/devarch/README.md`.
- `recipes/*/recipe.yml` — describe a bootstrap's arguments for `devarch new` and the TUI wizard; the script stays the authority on its interface.
- Scripts get catalog and runtime facts from the CLI (`devarch ls --json`, `devarch compose`, `devarch config --env`), not from Bash reimplementations.

Move Bash into Go only when it is platform logic or blocks the CLI/TUI. Domain logic moves only for a recorded trigger (the TUI needs structured data, a recurring quoting or parsing bug class, or native Windows/macOS support).

## Service catalog contract

- Each service is `services-library/<category>/<name>/compose.yml`, optionally with a top-level `x-devarch` block (title, tags, urls, version, ready, requires). Every field is optional.
- A version variable must appear as `${VAR:-default}` so plain compose keeps working.
- The shared network is declared as `microservices-net: {external: true, name: ${DEVARCH_NETWORK:-microservices-net}}`. Apps mounts use `${DEVARCH_APPS_DIR:-../../../apps}`.
- Pin every image to a version (no `:latest`). Before changing a pin for a service with data, check the volume and the local image so the change is not a silent downgrade.
- `devarch lint` enforces these rules and must stay at 0 errors and 0 warnings.

## Testing

- Go: `cd cli && go test ./... && go vet ./... && gofmt -l .`. Tests drive commands through `runner.Fake` and assert the exact native argv; never run real containers in Go tests.
- Bash suites (WordPress, Laravel, JavaScript, foundation; listed in `README.md` under "Development checks") are run with `</dev/null` so none can wait on a terminal. Bootstrap tests fake `devarch` with `DEVARCH_BIN` and pin `DEVARCH_RUNTIME` so a real `devarch` on `PATH` cannot apply the machine's config.
- Port a script's test cases before deleting the script.
- `services-library/backend/node/routing.test.sh` drives real containers. Run it only deliberately.
- Drive the TUI on an isolated tmux server, never the user's default one: wrap `tmux -L devarch-test -f /dev/null` in a shell function (zsh does not word-split a `$VAR`), and set `DEVARCH_CONFIG_HOME` to a scratch directory.
- Don't start real services unless the action is harmless, and clean up afterwards. On the main development machine `podman ps` takes 1–5 seconds and a host Node process holds `127.0.0.1:8080`.

## Conventions

- Never add `Co-Authored-By` trailers naming Claude or any other AI assistant to commits. Commit authorship and attribution must reflect the human contributor.
- Commit messages use conventional prefixes (`feat:`, `fix:`, `docs:`, `chore:`), with an optional scope such as `feat(cli):` or `fix(catalog):`.
- When behavior changes, update the user docs in the same change. Use `README.md` for the overview, `cli/README.md` for commands, the script's README for bootstrap details, and an "As built" note in the plan for design changes. Keep this file consistent with them.
- `apps/` holds the user's projects (separate repositories, ignored here). Never modify them as a side effect.
- `scripts/it-management/` and `.model-artifacts/` hold IT runbooks and initiative records unrelated to the dev environment.
