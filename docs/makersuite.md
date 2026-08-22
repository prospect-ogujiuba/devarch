# MakerSuite architecture and delivery guide

This is the canonical guide to the MakerSuite architecture, ownership model, and normal development workflow in DevArch. Detailed command options and package internals remain in the linked component documentation.

## Purpose

MakerSuite is designed for building and maintaining multiple substantial WordPress client sites without mixing reusable platform code with client-owned behavior.

Its governing rule is:

> Framework core is replaceable. Client project workspaces are independently owned and must survive core updates unchanged.

This separation makes upgrades, rollback, testing, handoff, and reuse predictable. It is most valuable for an agency or productized delivery system maintaining several sites. A disposable one-page site may not justify the additional repository and release discipline.

## System model

```text
DevArch
├── shared local infrastructure and safe provisioning
└── Maker stack
    ├── MakerStarter  — design system and WordPress site shell
    ├── MakerBlocks   — reusable Gutenberg block foundation
    └── MakerMaker    — application and domain-code generator

Each client site
├── <site>-theme      — branding, templates, patterns, presentation
├── <site>-blocks     — client-specific Gutenberg blocks
└── <site>-app        — models, controllers, policies, routes, migrations
```

DevArch supplies shared PHP-FPM, MariaDB, Nginx Proxy Manager, and optional services over `microservices-net`. Sites live in `apps/<site>` and route through `https://<site>.test` without a separate development server per site.

Playground is the integration and release site for the three Maker core packages:

```text
apps/playground/wp-content/themes/makerstarter/
apps/playground/wp-content/plugins/makerblocks/
apps/playground/wp-content/plugins/makermaker/
```

## Ownership boundaries

Every change must be classified before implementation.

| Change | Owner and destination |
| --- | --- |
| Branding, `theme.json`, templates, parts, patterns, assets, project PHP | `<site>-theme` |
| Client-specific blocks, editor code, frontend code, and block tests | `<site>-blocks` |
| Models, controllers, fields, policies, routes, views, migrations, and domain tests | `<site>-app` |
| Reusable design-system behavior or stable tokens | MakerStarter core |
| Reusable blocks or block-generation improvements | MakerBlocks core |
| Reusable application generation or runtime integration | MakerMaker core |
| Site Editor state stored in posts/options | Database-owned unless deliberately exported to `<site>-theme` |

### Non-negotiable invariants

1. Never place client-specific code in MakerStarter, MakerBlocks, or MakerMaker.
2. Never import MakerBlocks private source, templates, globals, or build internals from a project plugin.
3. Never use a Maker core feature worktree for site-owned code.
4. Never overwrite or merge an existing generated project workspace automatically.
5. Never run project database migrations during core synchronization.
6. Never synchronize dirty, untrusted, or origin-mismatched core repositories.
7. Never treat database-owned Site Editor state as a core Git change.

Core directories carry `CORE-BOUNDARY.md`. Generated workspaces carry `PROJECT OWNED — EDIT HERE` markers and are initialized as independent Git repositories on `main`, without an initial commit or remote.

## Why this benefits developers

### Clear placement decisions

Developers can decide where code belongs before implementation. Presentation, content components, domain behavior, and reusable platform improvements have separate homes and histories.

### Smaller regression scope

Each project workspace is independently buildable and disableable. A block change does not require editing the parent theme, and a domain change does not require modifying the block framework.

### Native WordPress contracts

MakerStarter uses normal child-theme resolution. MakerBlocks uses native block metadata and the official `@wordpress/create-block` tooling. MakerMaker generates sibling plugins rather than hiding project behavior inside its own repository.

### Efficient reuse without copy-paste drift

Reusable improvements return to the matching core package through review and release. Client code remains isolated. Generator templates flow one way from reviewed MakerBlocks core into project workspaces and affect only future blocks.

### Parallel core development

Playground core checkouts stay clean on `main`. Linked Git worktrees allow isolated feature branches without duplicating repositories or destabilizing the integration site.

## Why this benefits clients

- Exact core versions are recorded in `.devarch-maker.lock`.
- Core updates do not modify project workspaces, uploads, or database content.
- Updates are staged and atomically published.
- Failed publication restores all prior core directories and lock state.
- Retained rollback packages permit explicit restoration.
- Client behavior remains in identifiable, transferable repositories.
- Native WordPress conventions reduce dependence on undocumented framework behavior.
- Shared accessibility, security, and compatibility fixes can be delivered without merging over client customizations.

MakerMaker also validates namespaces and paths, refuses existing destinations, rejects symbolic-link template sources, requires authorization for admin generation, and stages generated resources before collision-checked publication.

## Create a client site

Use the `clean` profile for normal Maker client development. Use `custom` when Manual Image Crop is required and `loaded` when the additional debugging/development plugins are deliberately wanted.

Always preview provisioning first:

```bash
scripts/wordpress/bootstrap.sh acme --profile clean --dry-run
scripts/wordpress/bootstrap.sh acme --profile clean
```

The result is:

```text
apps/acme/wp-content/
├── themes/
│   ├── makerstarter/       # framework core — do not edit
│   └── acme-theme/         # project-owned presentation
└── plugins/
    ├── makerblocks/        # framework core — do not edit
    ├── acme-blocks/        # project-owned blocks
    ├── makermaker/         # framework core — do not edit
    └── acme-app/           # project-owned domain behavior
```

To add only the project workspaces to an existing Maker-enabled site:

```bash
scripts/wordpress/bootstrap.sh acme \
  --profile clean \
  --scaffolds-only
```

Existing workspace destinations are refused and left byte-for-byte unchanged.

After generation, create the first commit and configure the intended remote independently in each workspace. Do not assume that the top-level DevArch repository owns these histories.

## Efficient feature workflow

### 1. Define one vertical outcome

Describe the smallest user-visible result and its acceptance checks. Avoid starting with a list of files or framework layers.

### 2. Select the minimum necessary layer

Use this order:

1. Theme only when the change is presentation or composition.
2. A static block when saved post markup is sufficient.
3. A server block when output is dynamic but browser interaction is unnecessary.
4. A full block only when frontend interaction or hydration is required.
5. The app plugin when the feature owns structured data, authorization, routes, or migrations.

A feature may cross layers, but each responsibility remains with its owner. For example, a testimonial system can use:

```text
acme-app       → Testimonial model, policy, fields, and admin resource
acme-blocks    → Testimonial display block
acme-theme     → Brand-specific visual treatment
```

### 3. Implement and verify the complete slice

Test the relevant editor, frontend, admin, accessibility, fallback, and migration behavior before expanding the feature.

### 4. Commit only the affected workspace

Keep theme, block, and application histories independently reviewable. Promote a change to core only when it is generic enough to benefit unrelated clients.

## Theme development

Place branding, composition, templates, parts, patterns, assets, and project PHP in:

```text
apps/<site>/wp-content/themes/<site>-theme/
```

The child theme declares MakerStarter through the standard parent relationship:

```css
/*
Theme Name: Acme
Template: makerstarter
*/
```

Prefer MakerStarter's stable token slugs over duplicated hard-coded values:

```css
.component {
  color: var(--wp--preset--color--ink);
  padding: var(--wp--preset--spacing--40);
  border-radius: var(--wp--custom--radius--medium, 1rem);
}
```

Token slugs are the public contract; token values may evolve.

## Project block development

Install dependencies once in the project block workspace:

```bash
cd apps/acme/wp-content/plugins/acme-blocks
npm install
```

Generate the least complex suitable block:

```bash
npm run create:block -- pricing-calculator
npm run create:block -- account-summary --profile=server
npm run create:block -- callout --profile=static
```

- `full` is the default and includes frontend interaction with semantic PHP fallback.
- `server` omits frontend React.
- `static` saves markup into post content and omits server/frontend runtime.

Use the normal loop:

```bash
npm start
npm run build
npm test
```

Use the project namespace, never `makerblocks/*`. Existing block destinations are always refused.

### Updating creation templates

A project workspace owns a creation-time snapshot of MakerBlocks templates. Inspect and apply reviewed template changes explicitly:

```bash
npm run templates:check
npm run templates:dry-run
npm run templates:sync
```

Synchronization backs up and replaces only `templates/create-block/` and updates `.makerblocks-template.json`. It never changes existing generated blocks, `src/blocks/`, build output, or package configuration.

## Project application development

Generate application resources into `<site>-app` instead of modifying MakerMaker:

```bash
wp makermaker resource Product \
  --plugin=acme-app \
  --namespace='Maker\\Acme' \
  --plural=products \
  --migration \
  --views \
  --factory \
  --tests
```

The plugin-specific Galaxy launcher provides the corresponding workflow:

```bash
php galaxy_acme_app make:maker-resource Product \
  --migration \
  --tests
```

Generated resources include a model, controller, fields, deny-by-default policy, registry entry, and optional migrations, views, factories, and tests.

Run project migrations explicitly:

```bash
php galaxy_acme_app migrate up
```

## Reusable core development

The three package directories in Playground are independent repositories and primary Git worktrees. Keep each clean on `main`.

Create linked feature worktrees outside the WordPress tree:

```bash
cd apps/playground/wp-content/plugins/makerblocks
git worktree add \
  "$HOME/projects/worktrees/makerblocks/new-card" \
  -b feat/new-card
```

Develop, test, and commit there. Merge through the primary Playground checkout, verify again, and remove the linked worktree:

```bash
git switch main
git pull --ff-only
git merge feat/new-card
npm ci
npm run build
npm test
git worktree remove "$HOME/projects/worktrees/makerblocks/new-card"
git branch -d feat/new-card
```

Core verification commands are:

```bash
# MakerStarter
npm test

# MakerBlocks
npm ci
npm run build
npm test

# MakerMaker
composer lint
composer test
```

## Release the Maker stack

One semantic Maker stack version maps to exact tags and commits for all three core packages in `scripts/wordpress/maker-stack.json`. The schema is `scripts/wordpress/maker-stack.schema.json`.

After all package changes are committed, changelogs are updated, and Playground is using the exact clean commits, preview the release gates:

```bash
scripts/wordpress/release-maker.sh 1.0.0 --dry-run
```

Create local annotated tags and update the stack manifest only after the preview passes:

```bash
scripts/wordpress/release-maker.sh 1.0.0
```

The release command tests all three packages and their integration in WordPress. It deliberately never pushes. Review and push package tags and the DevArch manifest commit separately.

Use semantic-version intent consistently:

- major: breaking token, hook, schema, generator, or public API change;
- minor: additive compatible behavior;
- patch: compatible fix.

`stable` must point to a reviewed tag-only stack. `main` is allowed only for explicit local development.

## Synchronize a client

Use the profile appropriate to the site and preview every update:

```bash
scripts/wordpress/sync-maker.sh acme \
  --profile clean \
  --to stable \
  --dry-run
```

Prefer an exact reviewed version when applying an update:

```bash
scripts/wordpress/sync-maker.sh acme \
  --profile clean \
  --to 1.0.0
```

Synchronization validates the current lock, clean core checkouts, origins, tag/commit pairs, core markers, health files, and dependency installation. It replaces only the three declared core directories and atomically rewrites `.devarch-maker.lock`.

Restore a retained version when required:

```bash
scripts/wordpress/sync-maker.sh acme \
  --rollback <rollback-id>
```

## Existing-site migration

Audit before changing an existing Maker site:

```bash
scripts/wordpress/audit-maker.sh --all
scripts/wordpress/audit-maker.sh acme --sync-ready --runtime-check
```

Preview migration against a published stack:

```bash
scripts/wordpress/migrate-maker.sh acme \
  --profile clean \
  --to stable \
  --dry-run
```

Migration evidence, Git state, binary patches, untracked files, and audit reports are retained under `apps/.devarch-maker-migrations/`. Dirty, missing, or untrusted core stops synchronization until every difference is classified into project-owned or reusable core work.

Pilot migrations on a disposable copy. Verify routes, screenshots, active packages, admin resources, Galaxy commands, workspace hashes, a same-version no-op synchronization, a next-version synchronization, and one rollback before production rollout.

## Delivery checklist

Before presenting or deploying a client feature:

- [ ] Every changed file has the correct owner.
- [ ] Maker core repositories remain clean and client-neutral.
- [ ] Project workspaces have reviewed commits and configured remotes or retained backup receipts.
- [ ] Production assets are built.
- [ ] Relevant package tests pass.
- [ ] Editor, frontend, admin, and no-JavaScript/fallback behavior are checked where applicable.
- [ ] Authorization and migration behavior are verified.
- [ ] The exact Maker stack lock is retained.
- [ ] Client acceptance evidence is captured.
- [ ] Update and rollback procedures are known before production mutation.

## Detailed references

- [WordPress bootstrap, profiles, synchronization, and tests](../scripts/wordpress/README.md)
- [Maker release and compatibility policy](../scripts/wordpress/MAKER-RELEASES.md)
- [Existing Maker site migration](../scripts/wordpress/MIGRATE-MAKER-SITES.md)
- [MakerStarter package guide](../apps/playground/wp-content/themes/makerstarter/README.md)
- [MakerBlocks package guide](../apps/playground/wp-content/plugins/makerblocks/README.md)
- [MakerMaker package guide](../apps/playground/wp-content/plugins/makermaker/README.md)

When this guide and a component reference differ, treat ownership and architecture decisions here as canonical, then update the component reference in the same reviewed change so commands and implementation details remain synchronized.
