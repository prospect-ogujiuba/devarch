# Phase 2: Workspace scaffolds and DevArch profiles

Created: 2026-08-20
Purpose: Make the safe customization path the default for sites created from Maker-enabled profiles.

## Goal

Have a Maker-enabled profile install the three core packages and create three clearly project-owned workspaces without overwriting existing work.

## Scope

- Extend the WordPress profile model to distinguish `core` packages from `workspace` scaffolds.
- Derive a normalized project slug and PHP/JS namespace from the DevArch site name, with explicit override flags for exceptional names.
- Provision:
  - `themes/<site>-theme` from the MakerStarter child scaffold and activate it;
  - `plugins/<site>-blocks` from the MakerBlocks project scaffold and activate it;
  - `plugins/<site>-app` through MakerMaker's generator and activate it.
- Add root README markers in workspaces: `PROJECT OWNED — EDIT HERE`.
- Add root markers in core packages: `FRAMEWORK CORE — DO NOT EDIT; update from playground releases`.
- Refuse scaffold generation when a destination exists. Never merge, delete, or overwrite a workspace.
- Keep current clean/custom/loaded profile composition, but have their shared Maker stack declaration reference one reusable profile fragment to prevent drift.

## Proposed profile semantics

```text
maker-core theme makerstarter ref=v1.x
maker-core plugin makerblocks ref=v1.x
maker-core plugin makermaker ref=v1.x
maker-workspace child-theme makerstarter
maker-workspace blocks-plugin makerblocks
maker-workspace app-plugin makermaker
```

The exact syntax can change during implementation; the required distinction and behavior cannot.

## Outputs

- Reusable Maker stack profile fragment.
- Idempotent scaffold commands usable independently of full site bootstrap.
- Dry-run output that labels core installs and workspace creation separately.
- Per-site lock manifest recording repository URL, resolved tag/commit, package type, and install time.
- Updated bootstrap regression tests and WordPress script documentation.

## Acceptance criteria

- A new Maker profile site starts with all six packages in the target layout.
- The child theme—not MakerStarter—is active.
- Re-running provisioning leaves all workspace bytes unchanged.
- Existing workspaces cause a clear skip/refusal rather than an attempted merge.
- `clean`, `custom`, and `loaded` cannot drift in their Maker core declarations.
- Dry-run performs no clone, scaffold, activation, or lock-file write.

## Verification

- Bootstrap a disposable site for each Maker-enabled profile.
- Hash all workspace files, re-run provisioning, and assert identical hashes.
- Confirm WordPress reports the expected active theme/plugins.
- Test invalid slugs/namespaces, pre-existing destinations, interrupted scaffold publication, and rollback.

## Non-goals

- Installing Node dependencies in deployed consumer sites automatically.
- Generating branded content or database records beyond required activation/setup.
