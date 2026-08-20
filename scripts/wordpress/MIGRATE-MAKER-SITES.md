# Existing Maker site migration

Use this workflow to move project behavior out of replaceable Maker core checkouts without data loss. Never reset, clean, stash, merge, or replace a dirty core directory merely to make synchronization pass.

## Ownership decisions

Classify every modified or untracked core path before replacement:

| Current location/change | Destination or decision |
| --- | --- |
| MakerStarter branding, `theme.json`, styles, templates, parts, patterns, assets, or project hooks | `<site>-theme` using normal `Template: makerstarter` child-theme resolution |
| Site-specific blocks, block builds, editor/frontend code, or tests under MakerBlocks | Independently buildable `<site>-blocks`; use the project namespace, never `makerblocks/*` |
| Models, controllers, fields, policies, routes, resources, views, migrations, assets, or project hooks under MakerMaker | `<site>-app`, generated/registered through MakerMaker where possible |
| Generic fixes, stable tokens/hooks, shared generators, or reusable blocks | Commit upstream in the matching playground core repository and release through the Maker stack |
| Site Editor template/style changes stored in WordPress posts/options | Keep database-owned deliberately, or export reviewed overrides into `<site>-theme`; do not count database state as a core Git change |

Generated app migrations remain explicit project operations. Core synchronization never runs database migrations.

## Commands

Report every detected Maker site without changing it:

```bash
scripts/wordpress/audit-maker.sh --all
scripts/wordpress/audit-maker.sh my-site --sync-ready --runtime-check
```

Prepare or finish one site against a published stack:

```bash
scripts/wordpress/migrate-maker.sh my-site \
  --profile loaded \
  --to stable \
  --dry-run

scripts/wordpress/migrate-maker.sh my-site \
  --profile loaded \
  --to 0.1.0
```

The migration command first stores Git status, exact HEAD, remotes, binary patches, untracked-file archives, and the audit report under `apps/.devarch-maker-migrations/`. It then provisions refusal-safe sibling workspaces. Dirty, missing, or untrusted core stops with exit 3 after backup and workspace preparation; it does not write a lock or replace core. A clean trusted site receives an exact current lock, synchronizes through `sync-maker.sh`, and must pass ownership plus runtime activation audit.

After manually classifying a dirty site:

1. Copy/move project behavior into the named workspace; never delete the backup.
2. Test parity before removing the original core edit.
3. Commit the workspace in its own repository or retain a reviewed external backup and `.devarch-workspace-backup` receipt.
4. Promote reusable core work through playground and a semantic stack release.
5. Restore the core checkout to a clean trusted commit only after parity and backup are confirmed.
6. Rerun `migrate-maker.sh`; audit failures continue to block synchronization.

## Rollout checklist

### 1. Disposable copy

- Clone/copy the site and database to a disposable name.
- Capture routes/screenshots, WP-CLI active theme/plugins, Site Editor state, critical admin resources, Galaxy commands, and workspace hashes.
- Run migration to the same stack as a no-op, then the next test release.
- Recheck behavior and hashes; perform one retained rollback.

### 2. Pilot site

- Choose one low-risk site with clean or easily classified core.
- Preserve the migration evidence directory outside the site.
- Migrate, review frontend/editor/admin behavior, and observe before promotion.

### 3. Remaining sites

- Process one site at a time from `audit-maker.sh --all` output.
- Stop on unknown remotes, missing direct Git checkouts, dirty core, lock mismatches, missing ownership markers, failed activation, or behavior parity failures.
- Record per-site ownership decisions and evidence; never bulk-reset application directories.

## Per-site evidence

For each migrated site retain:

- migration backup directory and audit report;
- ownership decision for every core difference;
- workspace repository/backup reference and before/after hashes;
- exact lock and active package state;
- screenshots/routes and representative editor/frontend/admin checks;
- same-version no-op sync, next-version sync, and rollback results.
