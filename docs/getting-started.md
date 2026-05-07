# Getting started

## Prerequisites

- Go installed.
- Podman installed for runtime operations.
- A working user Podman socket for socket/status workflows.

Check Podman:

```bash
podman --version
podman ps
```

## Install the CLI

From the DevArch repository root:

```bash
go install ./cmd/devarch
export PATH="$HOME/go/bin:$PATH"
devarch --help
```

If you prefer a stable user-local path:

```bash
mkdir -p "$HOME/.local/bin"
ln -sf "$HOME/go/bin/devarch" "$HOME/.local/bin/devarch"
export PATH="$HOME/.local/bin:$PATH"
```

No aliases are required.

## Remove obsolete aliases

Old setups may alias `devarch` to the retired shell script path:

```txt
/home/priz/projects/devarch/scripts/devarch
```

If your shell reports that path as missing, remove or unalias it:

```bash
unalias devarch dvrc dv da 2>/dev/null || true
hash -r
```

Then remove the aliases from your shell config or dotfiles.

## First checks

Run diagnostics:

```bash
devarch --json doctor
```

Expected on a healthy local runtime:

- Podman available: pass
- Podman socket: pass

Warnings about missing default workspace/catalog roots are okay if you pass roots explicitly on commands.

## List catalog templates

```bash
devarch --catalog-root ./catalog/builtin catalog list
```

Show a template:

```bash
devarch --catalog-root ./catalog/builtin catalog show postgres
```

## Run against example workspaces

List workspaces:

```bash
devarch --workspace-root ./examples/v2/workspaces workspace list
```

Plan an example:

```bash
devarch \
  --workspace-root ./examples/v2/workspaces \
  --catalog-root ./catalog/builtin \
  workspace plan shop-local
```

Apply an example only when you are ready for DevArch to create/update local containers:

```bash
devarch \
  --workspace-root ./examples/v2/workspaces \
  --catalog-root ./catalog/builtin \
  workspace apply shop-local
```

Observe:

```bash
devarch --workspace-root ./examples/v2/workspaces --catalog-root ./catalog/builtin workspace status shop-local
devarch --workspace-root ./examples/v2/workspaces --catalog-root ./catalog/builtin workspace logs shop-local api
devarch --workspace-root ./examples/v2/workspaces --catalog-root ./catalog/builtin workspace exec shop-local api -- echo ok
```

## Important flag rule

Global flags go before the command:

```bash
# Good
devarch --workspace-root ./examples/v2/workspaces workspace list

# Wrong
devarch workspace list --workspace-root ./examples/v2/workspaces
```

## Next step

For a full small-container walkthrough, see [Small DB tutorial](tutorial-small-db.md).
