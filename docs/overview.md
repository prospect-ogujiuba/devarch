# DevArch overview

DevArch is a local development workspace orchestrator.

It is currently shipped as one Go CLI, `devarch`, that plans and applies local development resources from declarative workspace manifests.

## The problem

Local development environments often drift into a mix of:

- project-specific shell scripts
- copied Compose snippets
- implicit aliases
- undocumented container names
- manual Podman/Docker commands
- framework-specific bootstrap scripts
- stale API/dashboard assumptions

That makes environments hard to reproduce, hard to inspect, and risky to clean with commands such as `pcleanall`.

## The DevArch approach

DevArch makes the local environment explicit:

1. A **workspace** declares the resources a project needs.
2. A **catalog** provides reusable templates for common services.
3. DevArch resolves the workspace and catalog into desired runtime state.
4. `workspace plan` shows the changes before touching the runtime.
5. `workspace apply` creates or updates containers/networks.
6. `workspace status`, `logs`, `exec`, and `restart` observe and operate on the result.

The default workflow is:

```txt
define -> plan -> apply -> observe -> iterate
```

## What DevArch is today

DevArch v2 is intentionally narrow:

- a local Go CLI
- workspace discovery
- catalog inspection
- deterministic planning
- Podman-oriented apply/status/logs/exec/restart
- project scanning
- v1 import helpers for migration

## What DevArch is not today

DevArch CLI is not currently:

- a required API container
- a PHP/Laravel/WordPress scaffolding tool
- a dashboard-first product
- a shell alias collection
- a replacement for every Compose feature

Those can be added later as explicit, tested CLI/API contracts, but they are not required for the core local workflow.

## Runtime requirement

The CLI itself is just a local binary. It does not need a DevArch app container.

Commands that inspect or run containers need a working local container runtime. Current workflow support is Podman-oriented, so `devarch doctor` and `devarch socket status` are the first checks to run on a new or cleaned machine.
