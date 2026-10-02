# Recipes

`devarch new <recipe> <name> [flags]` runs a recipe: a `recipe.yml` that points at an executable bootstrap script and describes its arguments.

```bash
devarch new                                   # list recipes
devarch new wordpress shop --profile clean
devarch new laravel api --database sqlite
devarch new javascript store --framework next --profile fullstack --start
devarch --dry-run new wordpress shop          # the script prints its plan
```

The script is the authority on its interface: flags the manifest does not declare are passed through unchanged, and `devarch new <recipe> --help` shows the script's own help. The manifest lets `devarch` validate names and choices before anything runs and gives the TUI enough to build a form.

| Field | Meaning |
|---|---|
| `entry` | Script path relative to the checkout |
| `requires` | Services the script starts (shown before running) |
| `args[].positional` / `flag` | How the value reaches the script |
| `args[].type` | `string` (default), `bool`, or `path` |
| `args[].pattern`, `choices`, `choices_from` | Validation; `choices_from` globs files and may reference other args as `{name}` |
| `args[].confirm` | Question the TUI asks before passing a destructive flag |
| `args[].preview` | The `choices_from` files are line-oriented profiles (`directive value [option]`, `#` comments). `directives` lists `{kind, label}` in display order and `include` names the directive that pulls in another file; the TUI wizard then shows what the chosen profile installs. Go only groups the lines; the script still interprets them. |
| `no_hosts_flag` | Arguments that skip hostname registration; the TUI adds them and offers `H` afterwards, because `sudo` cannot prompt inside it |

Scripts run with `DEVARCH_BIN` (the running `devarch`) and `DEVARCH_ROOT` set. They report progress through `devarch_progress STEP STATE MESSAGE` from `scripts/devarch/lib/platform.sh`, which writes JSON lines to `DEVARCH_PROGRESS_FD` when a reader such as the TUI provides one and does nothing otherwise.

A personal recipe in `~/.config/devarch/recipes/<name>/recipe.yml` overrides a built-in one with the same name.
