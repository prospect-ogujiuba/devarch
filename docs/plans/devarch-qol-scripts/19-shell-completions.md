# Phase 19: Shell completions

Created: 2026-08-19
Purpose: Make dynamic service, stack, app, profile, section, and command names discoverable at the prompt.

## Goal

Add generated completion support for Bash, Zsh, and Fish using stable, side-effect-free list interfaces.

## Scope

- Complete subcommands, flags, canonical service IDs, unambiguous service names, categories, stack names, app names, profiles, validation/doctor sections, database adapters, and endpoint names.
- Add `scripts/devarch/completions.sh bash|zsh|fish` to emit scripts and an optional `install --user` mode with dry-run.
- Keep completion-time discovery local and bounded; cache only when necessary with safe invalidation by directory metadata.
- Never invoke runtime mutations, network access, browser launch, privilege elevation, or interactive prompts during completion.

## Outputs

- Completion generator, shell-specific templates, tests, installation/uninstallation instructions, and examples.

## Acceptance criteria

- Completion candidates come from existing stable `list`/metadata functions rather than duplicated hardcoded service lists.
- Paths and names containing shell metacharacters cannot inject code into generated scripts.
- Missing optional scripts or catalogs produce no candidates rather than terminal errors.
- User installation does not modify global shell configuration and refuses to overwrite unmanaged files.
- Generation is deterministic for a fixed repository state.

## Verification

- Golden-output tests for Bash, Zsh, and Fish; syntax-parse generated files when each shell is installed.
- Test spaces, metacharacters, ambiguous short names, changed catalogs, missing repository, and safe install/uninstall.
- Verify completion execution records no fake-runtime or network calls.

## Non-goals

No shell framework/plugin packaging, global system install, runtime health completion, or background completion daemon.
