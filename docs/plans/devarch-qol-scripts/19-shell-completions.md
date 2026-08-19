# Phase 19: Shell completions

Created: 2026-08-19
Purpose: Make dynamic service, stack, app, profile, section, and command names discoverable at the prompt.

## Goal

Install/use Podman's native completion and add only minimal completions for DevArch-specific resolver arguments after those commands exist.

## Scope

- Document/install Podman completion directly with `podman completion bash|zsh|fish|powershell` and its native `--file` option.
- Do not copy Podman subcommands or flags into DevArch completion definitions.
- Only if `service.sh`, `stack.sh`, or `app.sh` survives its implementation gate, add shell-native completion snippets for canonical filesystem-discovered service/stack/app names.
- Passthrough after `--` is left to Podman/provider completion where the shell can support it; otherwise do not approximate native flags.
- No DevArch completion generator or installer is created unless multiple static snippets prove unmaintainable.

## Native delegation

Podman owns its command/flag completion generation. Bash/Zsh/Fish/PowerShell own installation and dispatch. DevArch snippets expose only repository names unavailable to Podman.

## Outputs

- Podman native-completion instructions plus only the minimal shell-specific DevArch name snippets justified by implemented wrappers.

## Acceptance criteria

- Podman candidates come exclusively from `podman completion`; DevArch candidates come from filesystem/list functions rather than duplicated lists.
- Paths and names containing shell metacharacters cannot inject code into generated scripts.
- Missing optional scripts or catalogs produce no candidates rather than terminal errors.
- Documentation recommends the shell's normal user completion directory; any optional copy helper refuses to overwrite unmanaged files.
- DevArch snippets are deterministic for a fixed repository state.

## Verification

- Syntax-test only the small DevArch snippets for installed shells; smoke-test the documented `podman completion` commands without snapshotting Podman's generated output.
- Test spaces, metacharacters, ambiguous short names, changed catalogs, missing repository, and safe install/uninstall.
- Verify completion execution records no fake-runtime or network calls.

## Non-goals

No reimplementation/snapshot of Podman completion, DevArch completion generator unless justified, shell framework packaging, global install, runtime health completion, or background daemon.
