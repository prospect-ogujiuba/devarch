# DevArch concepts

## Workspace

A workspace is a declarative manifest for a local development environment.

Typical file:

```txt
devarch.workspace.yaml
```

A workspace names the environment, selects runtime behavior, points at catalog sources, and declares resources.

## Resource

A resource is one thing DevArch manages inside a workspace.

Examples:

- `postgres`
- `redis`
- `mariadb`
- `adminer`
- `api`
- `web`

A resource can reference a catalog template and override environment, ports, volumes, dependencies, imports, and exports.

## Template

A template is a reusable service definition stored in a catalog.

Templates describe defaults such as:

- container image
- command or entrypoint
- environment
- ports
- volumes
- health check
- imports and exports

Builtin templates live under:

```txt
catalog/builtin/<category>/<template-name>/template.yaml
```

## Catalog

A catalog is a set of templates that workspaces can reference.

A workspace can declare catalog sources:

```yaml
catalog:
  sources:
    - ../../../../catalog/builtin
```

The CLI can inspect catalogs:

```bash
devarch --catalog-root ./catalog/builtin catalog list
devarch --catalog-root ./catalog/builtin catalog show postgres
```

## Runtime provider

The runtime provider is the local execution backend. Current workflows are Podman-oriented.

A workspace can request a provider:

```yaml
runtime:
  provider: podman
  isolatedNetwork: true
  namingStrategy: workspace-resource
```

`isolatedNetwork: true` tells DevArch to create a workspace network. `namingStrategy: workspace-resource` gives deterministic runtime names such as `devarch-shop-local-api`.

## Plan

`workspace plan` compares desired state with runtime state and reports actions:

- `add`
- `modify`
- `remove`
- `restart`
- `noop`

Plan is the safe preview step. Run it before apply when changing manifests.

## Apply

`workspace apply` executes the planned changes through the runtime adapter.

For Podman, this means creating/replacing containers and networks with DevArch labels so later status/logs/exec operations can find them.

## Status

`workspace status` shows both:

- desired state resolved from manifests and templates
- observed runtime snapshot from Podman

Use it to answer: “What should exist?” and “What is actually running?”

## Logs, exec, restart

Once a resource exists:

```bash
devarch --workspace-root <root> workspace logs <workspace> <resource>
devarch --workspace-root <root> workspace exec <workspace> <resource> -- <command...>
devarch --workspace-root <root> workspace restart <workspace> <resource>
```

## Imports and exports

Templates/resources can expose contracts and consume contracts.

Example: a database exports connection variables, and an app imports them.

```yaml
exports:
  - contract: postgres
    env:
      DB_HOST: "${resource.host}"
      DB_PORT: "${resource.port.5432}"
```

```yaml
imports:
  - contract: postgres
    from: postgres
```

This is how DevArch can wire resources without hard-coding every environment variable in every workspace.

## Current limits

DevArch does not promise full Compose parity. Some image defaults reported by runtime inspect can differ from template intent, so a follow-up `plan` can sometimes report `modify` for normalized image, command, entrypoint, env, port, or volume differences. Treat plan output as the source of truth and report noisy diffs as bugs or adapter gaps.
