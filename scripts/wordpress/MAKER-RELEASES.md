# Maker stack releases and compatibility

Playground is the integration and release site for MakerStarter, MakerBlocks, and MakerMaker. Consumer projects synchronize only these core checkouts; sibling project workspaces, uploads, and database content are outside the release boundary.

## Version policy

- Stack and package tags use semantic versions: `vMAJOR.MINOR.PATCH`; manifest stack keys omit `v`.
- `stable` points to the latest reviewed tag-only stack. `main` is permitted only for local development with `sync-maker.sh --allow-main`.
- Breaking token, hook, schema, generator, or public API changes require a major version and migration notes.
- Additive compatible behavior uses a minor version; compatible fixes use a patch version.
- Database migrations belong to generated project app plugins and are never run by core synchronization.
- Every package keeps a `CHANGELOG.md` using `## [X.Y.Z] - YYYY-MM-DD`, grouped under `Added`, `Changed`, `Fixed`, `Deprecated`, `Removed`, and `Security` as applicable. Breaking releases link migration notes.

## Compatibility matrix

`scripts/wordpress/maker-stack.json` is the machine-readable compatibility matrix. One stack version maps exactly one repository URL, semantic tag, commit, package type, health file, and install contract for each core package. `maker-stack.schema.json` is its schema.

The committed manifest intentionally starts empty. A release is added only by the gated playground release command after the three package repositories are clean, independently tested, and healthy together in WordPress.

## Playground release checklist

1. Finish and commit scoped changes independently in all three nested repositories.
2. Add each package's changelog entry and migration notes for breaking changes.
3. Ensure the playground site uses those exact clean commits.
4. Run the release gates without changing tags or the manifest:

   ```bash
   scripts/wordpress/release-maker.sh 1.0.0 --dry-run
   ```

5. Create matching local annotated tags and update the manifest atomically:

   ```bash
   scripts/wordpress/release-maker.sh 1.0.0
   ```

6. Review the manifest diff and commit it in DevArch. Do not include unrelated changes.
7. Push the three package tags and the DevArch manifest commit only after review. If any push fails, stop; do not move `stable` again until the remote state is reconciled explicitly.
8. Synchronize two disposable consumers and compare exact core commits and workspace hashes before promoting the stack beyond local use.

The release command runs MakerStarter `npm test`, MakerBlocks `npm test`, MakerMaker `composer test`, and a playground WP-CLI matrix covering WordPress installation, both active plugins, MakerStarter installation, and MakerMaker Galaxy registration. It never pushes automatically.

## Consumer synchronization

Preview a stable update:

```bash
scripts/wordpress/sync-maker.sh my-site \
  --profile loaded \
  --to stable \
  --dry-run
```

Apply an exact version:

```bash
scripts/wordpress/sync-maker.sh my-site --profile loaded --to 1.0.0
```

The command validates the current lock, clean core checkouts, exact origin URLs, trusted tag/commit pairs, core markers, release health files, and declared dependency installation. It stages complete replacements, retains `.devarch-maker-rollbacks/<id>/`, publishes only the three declared core directories, checks installed commits, and atomically rewrites `.devarch-maker.lock`. Any failure after publication automatically restores all prior core directories and the prior lock.

Restore retained versions explicitly:

```bash
scripts/wordpress/sync-maker.sh my-site --rollback 20260820T153000Z-12345
```

Synchronization never stashes, resets, merges, or pulls a dirty checkout. It never reads or writes `<site>-theme`, `<site>-blocks`, `<site>-app`, uploads, or database-owned content.
