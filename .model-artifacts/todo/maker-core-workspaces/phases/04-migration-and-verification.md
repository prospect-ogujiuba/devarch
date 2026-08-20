# Phase 4: Existing-site migration and isolation verification

Created: 2026-08-20
Purpose: Move existing project edits out of core checkouts before enabling synchronized updates.

## Goal

Classify and migrate current changes without data loss, then prove that future core replacement cannot affect project-owned work.

## Scope

- Inventory every site using a Maker-enabled DevArch profile.
- For each core repository, compare its working tree and commit to the expected upstream release.
- Classify differences as:
  - reusable core improvement to upstream through playground;
  - site theme customization to `<site>-theme`;
  - site block customization to `<site>-blocks`;
  - site domain/application behavior to `<site>-app`;
  - database/Site Editor content that should remain database-owned.
- Create sibling workspaces without overwriting destinations, move classified behavior, and verify parity.
- Reset/replace core only after the migrated workspaces are committed or backed up and acceptance checks pass.
- Activate the child theme and required project plugins; record exact core refs in the lock manifest.

## Outputs

- Per-site migration inventory and ownership decisions.
- A report-only audit command that finds modified/untracked core files and missing workspace markers.
- Migration guidance for Site Editor template overrides, custom blocks, generated resources, and project hooks.
- A rollout checklist beginning with a disposable copy, then one pilot site, then remaining sites.

## Acceptance criteria

- No known site behavior depends on uncommitted or site-specific files inside a core directory.
- All three core repositories are clean and match lock-file commits after migration.
- Project workspaces are independently versioned/backed up.
- Core synchronization passes while representative site branding, templates, blocks, and domain behavior remain intact.
- Audit failures are actionable and block synchronization.

## Verification

For every migrated site:

1. Capture screenshots/routes, active package state, critical WP-CLI checks, and workspace hashes.
2. Test theme templates/parts, representative blocks in editor/frontend, MakerMaker-generated admin resources, migrations, and Galaxy commands.
3. Synchronize core to the same version as a no-op, then to the next test release.
4. Re-run checks and confirm unchanged workspace hashes.
5. Roll back once and confirm the previous stack remains functional.

## Open questions

- Which existing projects currently contain local changes in the three nested core repositories.
- Whether database Site Editor customizations should be exported into child-theme files per project or intentionally remain database-owned.
- Whether all consumer projects are development-only or whether staged release promotion is required.
