# DevArch scripts

These Bash helpers add only DevArch repository discovery and safety checks that native commands cannot provide. They are not a container-runtime framework.

## Shared foundation contract

Source `lib/catalog.sh` for catalog operations; it sources `lib/common.sh`. Public functions are:

- `devarch_repo_root` — print the physical repository root. Set `DEVARCH_REPO_ROOT` only when intentionally targeting another checkout (including tests).
- `devarch_require_podman` — require Podman without inventing a Docker fallback.
- `devarch_require_compose` — run `podman compose --help`, allowing Podman to select and diagnose its configured external Compose provider.
- `devarch_run [--exec] -- command [args...]` — log a shell-escaped argv vector and invoke it unchanged. Use `--exec` for the final native process when no post-processing is required.
- `devarch_catalog_root` — print the validated physical `services-library` path.
- `devarch_catalog_list` — print validated canonical `category/name` IDs in deterministic order.
- `devarch_catalog_resolve ID_OR_UNIQUE_NAME` — resolve an exact canonical ID or an unambiguous short name.
- `devarch_catalog_compose_file ID_OR_UNIQUE_NAME` — print the validated absolute Compose file path.

Keep native behavior native: do not wrap individual Podman subcommands, reformat native output, maintain a Podman/Docker feature matrix, or add runtime adapters. Wrapper-owned options end at `--`; forward every later argument unchanged. Prefer `exec` for the final process so stdin, stdout, stderr, TTY state, signals, and exit status remain native.

DevArch dry-run behavior is limited to filesystem changes owned by a wrapper. Compose simulation must be forwarded to `podman compose --dry-run` when the installed provider supports it. Reuse native `--format`, `--filter`, JSON, `--watch`, and completion support.

## Foundation tests

```bash
scripts/devarch/tests/run-tests.sh
bash -n scripts/devarch/lib/*.sh scripts/devarch/tests/*.sh
shellcheck scripts/devarch/lib/*.sh
```

The fixture helper creates temporary repositories and recording `podman`, database CLI (`psql`), launcher (`devarch-launcher`), and certificate (`mkcert`) executables for later script phases.
