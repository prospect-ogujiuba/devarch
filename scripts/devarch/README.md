# DevArch Bash libraries

The `devarch` CLI owns platform work: starting services, readiness, hosts, databases, provisioning guards and app runtimes. These libraries are what the bootstraps still need in Bash. They are not a container-runtime framework.

## `lib/platform.sh` (sourced by every bootstrap)

- `devarch_bin` — print the `devarch` executable: `$DEVARCH_BIN` (set by `devarch new`), else the one on `PATH`.
- `devarch_runtime_env [DEVARCH]` — set `DEVARCH_RUNTIME` (podman or docker) and `DEVARCH_CONTAINER_USER` (the uid:gid for container exec calls). `devarch new` passes both; a bootstrap run directly reads them from `devarch config --env`. The Podman/Docker differences themselves live in one table in `cli/internal/engine/runtime.go`.
- `devarch_env_value KEY` — read one value from `NAME=value` output on stdin, such as `devarch db create --env` or `devarch app guard`.
- `devarch_progress STEP STATE [MESSAGE]` — report progress as a JSON line on `DEVARCH_PROGRESS_FD` when a reader (the TUI) provided one; otherwise do nothing.

## `lib/dotenv.sh`

`devarch_load_dotenv FILE KEY...` parses only the named keys from a dotenv file as data: no command substitution, no export, and controls rejected. The WordPress and Laravel bootstraps use it for the repository `.env`.

Catalog lookups and runtime checks belong to the CLI: use `devarch ls --json`, `devarch compose <service> -- <args>` and `devarch doctor` from scripts. (The earlier `catalog.sh` and `common.sh` libraries were removed once nothing used them.)

## Rules for bootstrap code

- Call `devarch` for platform steps: `devarch up … --wait --no-hosts`, `devarch hosts add`, `devarch db create --app`, `devarch app guard|release|recover`, `devarch app start`. Do not reimplement them in Bash.
- Run container commands with `"$DEVARCH_RUNTIME" exec --user "$DEVARCH_CONTAINER_USER" …`, never a hard-coded `podman`.
- Keep native behavior native: do not wrap individual runtime subcommands or reformat native output. Wrapper-owned options end at `--`; forward later arguments unchanged.
- A bootstrap's `--dry-run` covers its own steps and prints the `devarch` commands it would run.

## Tests

```bash
scripts/devarch/tests/run-tests.sh </dev/null
bash -n scripts/devarch/lib/*.sh scripts/devarch/tests/*.sh
shellcheck scripts/devarch/lib/*.sh
```

Running Bash suites with stdin from `/dev/null` keeps them from waiting on a terminal.
