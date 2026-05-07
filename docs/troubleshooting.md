# Troubleshooting

## `zsh: no such file or directory: /home/priz/projects/devarch/scripts/devarch`

Cause: your shell still has an old alias from the retired `scripts/setup-aliases.sh` flow.

Fix current shell:

```bash
unalias devarch dvrc dv da 2>/dev/null || true
hash -r
```

Then remove lines like these from shell config or dotfiles:

```bash
alias devarch='/home/priz/projects/devarch/scripts/devarch'
alias dvrc='/home/priz/projects/devarch/scripts/devarch'
alias dv='/home/priz/projects/devarch/scripts/devarch'
alias da='/home/priz/projects/devarch/scripts/devarch'
```

Install the Go CLI instead:

```bash
go install ./cmd/devarch
export PATH="$HOME/go/bin:$PATH"
devarch --help
```

## `devarch: command not found`

Cause: the Go binary is not installed or not on `PATH`.

Fix:

```bash
go install ./cmd/devarch
export PATH="$HOME/go/bin:$PATH"
command -v devarch
devarch --help
```

Optional symlink:

```bash
mkdir -p "$HOME/.local/bin"
ln -sf "$HOME/go/bin/devarch" "$HOME/.local/bin/devarch"
export PATH="$HOME/.local/bin:$PATH"
```

## Global flags seem ignored

Global flags must come before the command.

```bash
# Good
devarch --workspace-root ./examples/v2/workspaces workspace list

# Wrong
devarch workspace list --workspace-root ./examples/v2/workspaces
```

## `doctor` warns about workspace roots or catalog roots

This is expected when no defaults are configured.

Pass roots explicitly:

```bash
devarch \
  --workspace-root ./examples/v2/workspaces \
  --catalog-root ./catalog/builtin \
  workspace list
```

## Podman socket unavailable

Check:

```bash
devarch socket status
podman ps
```

Start the user socket:

```bash
devarch socket start
# or
systemctl --user start podman.socket
```

Then re-run:

```bash
devarch --json doctor
```

## I ran `pcleanall`; is DevArch broken?

Usually no. `pcleanall` removes containers, images, volumes, networks, and pods. DevArch can recreate declared workspace resources with `workspace apply`.

After `pcleanall`:

```bash
devarch --json doctor
devarch --workspace-root <root> --catalog-root <catalog> workspace plan <name>
devarch --workspace-root <root> --catalog-root <catalog> workspace apply <name>
```

Expect images to be pulled again.

## Does the CLI need a DevArch API container?

No.

The current supported CLI is a local Go binary. It calls local tools such as Podman/systemctl through the runtime workflow. It does not need the old DevArch API/app container.

Container requirements are resource-specific:

- `devarch doctor`: needs local Go binary and local runtime checks.
- `devarch catalog list/show`: no containers required.
- `devarch workspace plan`: needs manifests/catalogs; runtime snapshot may query Podman.
- `devarch workspace apply/status/logs/exec/restart`: needs Podman and relevant managed containers.

## Does the web UI need an API?

The web UI is separate from the CLI workflow. `web/README.md` currently says the dev server proxies `/api` to `http://127.0.0.1:7777` for a local `devarchd` instance.

That means:

- CLI workflows do not need the web UI or API.
- If you work on the web UI, you need whatever local `devarchd` API implementation is current for that UI slice.
- Do not bring back the old custom PHP mega container just to run the CLI.

## Plan after apply still shows `modify`

This can happen because runtime inspect reports normalized container state, including:

- fully qualified image names
- default command/entrypoint
- injected image environment
- normalized port bindings
- runtime labels
- volume details

If the containers are healthy and behavior works, treat this as a runtime adapter normalization gap. Capture the plan output and file a focused bug.

## Cleanup a tutorial workspace

For the small DB tutorial:

```bash
podman rm -f devarch-tiny-db-adminer devarch-tiny-db-mariadb
podman network rm devarch-tiny-db-net
podman volume rm workspace.mariadb-data
```

For any workspace, list managed resources first:

```bash
podman ps -a --filter label=devarch.managed-by=devarch-v2 --format '{{.Names}} {{.Labels}}'
podman network ls --filter label=devarch.managed-by=devarch-v2
```
