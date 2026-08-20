# Phase 3: Release and multi-project synchronization

Created: 2026-08-20
Purpose: Update core packages across projects predictably without touching project workspaces.

## Goal

Make playground the integration/release site and give consumer projects an explicit, lockable core synchronization command.

## Scope

### Playground release flow

- Continue maintaining the three nested repositories in `apps/playground/wp-content/...`.
- Test each repository independently, then run an integration matrix in playground.
- Tag each repository using semantic versions. A Maker stack release manifest maps one stack version to compatible MakerStarter, MakerBlocks, and MakerMaker refs.
- Push repository tags and publish/update the stack manifest only after integration passes.

### Consumer synchronization

Add a command conceptually like:

```bash
scripts/wordpress/sync-maker.sh <site> --profile loaded --to <stack-version> --dry-run
```

It must:

1. Read the requested profile/stack manifest and the site's lock file.
2. Validate that each target is one of the three declared core directories.
3. Refuse dirty core repositories by default; never stash or discard changes silently.
4. Fetch and check out the exact resolved tag/commit, preferably via atomic staged replacement rather than an in-place pull.
5. Run package installation/build steps only when required by the release contract; committed MakerBlocks `build/` keeps Node optional for deployment.
6. Run health checks, update the lock file atomically, and retain rollback metadata.
7. Ignore sibling `<site>-theme`, `<site>-blocks`, `<site>-app`, uploads, and database-owned content.

### Version policy

- Profiles select a channel or compatible range; each site lock records exact commits.
- Local development may opt into `main`; shared/staging/production sites default to tags.
- Breaking token, hook, schema, or public API changes require a major release and migration notes.
- Database migrations belong to project app plugins and run only through explicit commands, never merely because files synchronized.

## Outputs

- Maker stack release manifest and schema.
- Release checklist/command from playground.
- Consumer sync command with `--dry-run`, explicit target, backup, verification, and rollback.
- Compatibility matrix and changelog conventions.

## Acceptance criteria

- One command can update only the three core packages in a consumer project to exact known refs.
- Workspace hashes before and after synchronization are identical.
- Dirty core, unknown remote, untrusted ref, failed dependency install, and failed health check stop safely.
- A failed update restores the prior core versions and lock file.
- The installed state is reproducible from profile plus lock manifest.

## Verification

- Create two disposable consumer sites with different workspace changes and synchronize both.
- Compare core commits with the release manifest and workspace hashes with pre-sync values.
- Exercise dirty-core refusal, interrupted replacement, incompatible versions, and rollback.
- Run the existing WordPress shell regression suite plus package-specific tests and playground smoke tests.

## Non-goals

- Blind `git pull` across all application directories.
- Automatic production updates without review or a rollback point.
