# Maker core/workspace boundary plan

Created: 2026-08-20
Purpose: Define an update-safe ownership boundary for MakerStarter, MakerBlocks, and MakerMaker before implementation.

## Outcome

Treat the three maintained repositories as replaceable framework packages and put project-owned work in sibling WordPress packages. Playground remains the integration and release source; DevArch profiles install versioned releases into consumer sites.

## Design rule

Do not put a `custom/` directory inside a cloned core repository. That is only a naming convention, leaves a dirty Git checkout, and makes replacement/rollback ambiguous. Use WordPress package boundaries—the equivalent of Laravel's `vendor/` versus `app/` and WordPress parent versus child themes.

## Target site layout

```text
wp-content/
├── themes/
│   ├── makerstarter/          # maintained core; replaceable
│   └── <site>-theme/          # project workspace; child theme
├── plugins/
│   ├── makerblocks/           # maintained core; replaceable
│   ├── <site>-blocks/         # project block workspace
│   ├── makermaker/            # maintained generator; replaceable
│   └── <site>-app/            # project domain/MVC workspace
└── devarch-maker.lock         # installed core refs; no secrets
```

Only the three core directories are managed by the Maker repositories. The three site packages are owned by the consumer project and must never be modified by profile synchronization.

## Implementation order

1. [01-package-contracts \[COMPLETE\].md](01-package-contracts%20%5BCOMPLETE%5D.md) — establish ownership and extension contracts.
2. [02-scaffolds-and-profiles.md](02-scaffolds-and-profiles.md) — generate project workspaces during provisioning.
3. [03-release-and-sync.md](03-release-and-sync.md) — add tagged, lockable multi-project updates.
4. [04-migration-and-verification.md](04-migration-and-verification.md) — migrate existing sites and prove isolation.

## Non-goals

- Moving uploads, WordPress content, or database-managed Site Editor changes into Git.
- Allowing profile sync to overwrite project-owned files.
- Automatically tracking an unpinned `main` branch in production-like projects.
- Combining all six packages into one repository.
