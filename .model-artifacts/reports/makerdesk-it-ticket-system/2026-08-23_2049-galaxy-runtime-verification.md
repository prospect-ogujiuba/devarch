# galaxy-runtime-verification

Created: 2026-08-23T20:49:24.957Z
Purpose: Record Phase 1 Galaxy and runtime verification before scaffold generation.

# MakerDesk Galaxy and Runtime Verification

Date: 2026-08-23

## Runtime

- Container runtime: Podman 5.8.2.
- `php`, `mariadb`, and `nginx-proxy-manager` are running and healthy.
- Host-side Galaxy execution cannot resolve the container-only `mariadb` hostname; Galaxy commands must run in the PHP container, for example:
  `podman exec php php /var/www/html/playground/galaxy_playground_app ...`

## Galaxy surfaces

The site, MakerMaker, and playground-app launchers all boot successfully in the PHP container and currently expose the same TypeRocket command inventory, including:

- `make:maker-resource`
- `make:model`, `make:controller`, `make:fields`, `make:policy`
- `make:migration`, `make:service`, `make:job`, `make:command`
- `migrate`

The three contexts are operational; application writes will use the most specific `galaxy_playground_app` launcher.

## Resource generator contract

`make:maker-resource [options] <PascalCaseName>` accepts `--plugin`, `--namespace`, `--plural`, `--migration`, `--views`, `--factory`, and `--tests`. It generates required MVC files and an explicit registry entry and refuses overwrites.

Approved application command shape:

```bash
podman exec php php /var/www/html/playground/galaxy_playground_app \
  make:maker-resource Ticket --migration --views --factory --tests --no-interaction
```

Because the launcher is plugin-specific, plugin and namespace are resolved from its context.

## Migration contract

`migrate <type> [steps]` supports `up`, `down`, `reload`, and `flush`, plus explicit `--path` and `--wp_option`. There is no `status` migration type.

## Worktree scope

- TypeRocket core: independent `master` worktree, clean.
- MakerMaker core: independent `main` worktree, clean.
- `playground-app`: independent `main` worktree containing the existing uncommitted application scaffold. MakerDesk changes belong only here.
- DevArch root had only a pre-existing untracked `.pi/pi-gate/pi-gate.json` before plan creation.

## Gate result

Phase 1 passes. The database and all three Galaxy surfaces are operational through Podman. Core repositories are clean and excluded from application writes.
