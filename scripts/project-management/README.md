# Project management evaluation stacks

This launcher manages six independent Podman Compose projects from `services-library/project/`. Each application has a dedicated database container, private stack network, persistent named volumes, a unique loopback-only port, and access to the external `microservices-net` network.

```bash
scripts/project-management/manage.sh up
scripts/project-management/manage.sh status
scripts/project-management/manage.sh urls
scripts/project-management/manage.sh down
```

`down` removes the containers and private networks but retains database and application volumes. To reset one product completely, run `podman compose down -v` from that product's service-library directory. This permanently deletes its local database and attachments.

The preferred endpoints are `https://redmine.test`, `https://openproject.test`, `https://plane.test`, `https://glpi.test`, `https://leantime.test`, and `https://vikunja.test`. They are routed through the shared Nginx Proxy Manager container. Refresh local name resolution after catalog changes with `scripts/hosts/sync-hosts.sh`.

OpenProject and GLPI read generated credentials from their gitignored, mode-`0600` `.env` files. Tracked `.env.example` files document the required variables. Retrieve a login locally with:

```bash
grep -E '^(OPENPROJECT_ADMIN_LOGIN|OPENPROJECT_ADMIN_PASSWORD)=' services-library/project/openproject/.env
grep '^GLPI_ADMIN_PASSWORD=' services-library/project/glpi/.env
```

Keep these files in a password manager and an encrypted off-host backup. Never commit or copy their values into documentation. GLPI's unused default `tech`, `normal`, and `post-only` accounts should remain disabled.
