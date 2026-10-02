# Project management evaluation stacks

This directory documents six independent Podman Compose projects from `services-library/project/`. Each application has a dedicated database container, private stack network, persistent named volumes, a unique loopback-only port, and access to the external `microservices-net` network.

Every stack carries the `project-management` tag in its `x-devarch` metadata, so the `devarch` CLI (see [`cli/README.md`](../../cli/README.md)) manages them as a group:

```bash
devarch up --tag project-management        # create the shared network if needed, start all six, print URLs
devarch ls --tag project-management        # state and URL of each stack
devarch ps                                 # running containers, mapped to their stack
devarch down --tag project-management      # stop and remove containers; volumes are retained
```

A start that fails on the first attempt is retried once, as the former `manage.sh` launcher did. To reset one product completely, run `devarch down <name> --volumes`. It names the volumes it will delete and asks first; this permanently deletes the local database and attachments.

The preferred endpoints are `https://redmine.test`, `https://openproject.test`, `https://plane.test`, `https://glpi.test`, `https://leantime.test`, and `https://vikunja.test`. They are routed through the shared Nginx Proxy Manager container. Refresh local name resolution after catalog changes with `devarch hosts sync`.

OpenProject and GLPI read generated credentials from their gitignored, mode-`0600` `.env` files. Tracked `.env.example` files document the required variables. Retrieve a login locally with:

```bash
grep -E '^(OPENPROJECT_ADMIN_LOGIN|OPENPROJECT_ADMIN_PASSWORD)=' services-library/project/openproject/.env
grep '^GLPI_ADMIN_PASSWORD=' services-library/project/glpi/.env
```

Keep these files in a password manager and an encrypted off-host backup. Never commit or copy their values into documentation. GLPI's unused default `tech`, `normal`, and `post-only` accounts should remain disabled.
