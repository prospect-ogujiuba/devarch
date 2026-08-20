# Phase 1: Package ownership and extension contracts

Created: 2026-08-20
Purpose: Give each maintained Maker package a stable public surface while keeping all site-specific work outside its checkout.

## Goal

Document and enforce which files DevArch may replace and where users are expected to work.

## Scope

### MakerStarter

- Keep `themes/makerstarter/` entirely core-owned. Its current `theme.json`, `templates/`, `parts/`, `patterns/`, `styles/`, `functions.php`, and tests remain framework assets.
- Add a maintained child-theme scaffold, but generate it as `themes/<site>-theme/`.
- The child theme owns branding and site composition: child `theme.json`, style variations, templates, template parts, patterns, assets, and `functions.php`.
- Use `Template: makerstarter`; child overrides rely on standard WordPress resolution rather than custom file discovery.
- Define stable token slugs and documented hooks as MakerStarter's compatibility contract. Removing/renaming tokens or hooks requires a major version.

### MakerBlocks

- Keep `plugins/makerblocks/` entirely core-owned. Generic blocks, shared frontend behavior, generators, committed builds, and tests remain here.
- Generate `plugins/<site>-blocks/` as a standalone WordPress block plugin for project-specific blocks.
- The site plugin owns its own `src/`, `build/`, package files, and tests. Its block namespace must be project-specific, not `makerblocks/*`.
- Site blocks may use documented MakerBlocks PHP/JS APIs, but must remain independently buildable. No imports from private MakerBlocks source paths.
- Expose versioned public APIs only where reuse is real; otherwise copy a scaffold at creation time so future core changes cannot silently alter generated code.

### MakerMaker

- Keep `plugins/makermaker/` entirely core-owned. Its current generator, CLI, Galaxy integration, templates, and tests are framework code.
- Treat `plugins/<site>-app/` (and any additional generated domain plugins) as the user workspace. This aligns with MakerMaker's existing behavior of generating sibling plugins and refusing overwrite.
- Generated plugins own models, controllers, fields, policies, resources, views, routes, migrations, assets, and tests.
- MakerMaker updates may improve future scaffolds and runtime integration but must never rewrite an existing generated plugin unless a separate, explicit, diff-producing migration command is introduced.

## Outputs

- `CORE-BOUNDARY.md` in each maintained repository with core-owned paths, public APIs, compatibility policy, and prohibited edits.
- README sections that point users to the corresponding sibling workspace.
- Runtime version/dependency checks with actionable admin/CLI messages.
- Stable hooks/contracts where site packages genuinely need them.

## Acceptance criteria

- Every file beneath the three core package directories is classified as replaceable core.
- Every normal site customization has a named location in one of the sibling workspaces.
- No documented workflow asks users to edit a core checkout.
- MakerStarter child overrides use WordPress conventions.
- MakerBlocks and MakerMaker extension packages can be disabled independently without corrupting core packages.

## Verification

- Review ownership tables against current repository trees.
- Add contract tests for declared hooks, token slugs, dependency/version checks, and package boot without site workspaces.
- Confirm each core repository has a clean `git status` after making representative site changes.

## Open questions

- Whether a site needs one `<site>-app` plugin or several bounded-domain plugins; default to one, allow more.
- Which MakerBlocks facilities are stable enough to expose versus scaffold-copy only.
