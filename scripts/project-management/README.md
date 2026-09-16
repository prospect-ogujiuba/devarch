# Project management evaluation stacks

This launcher manages six independent Podman Compose projects from `services-library/project/`. Each application has a dedicated database container, private stack network, persistent named volumes, a unique loopback-only port, and access to the external `microservices-net` network.

```bash
scripts/project-management/manage.sh up
scripts/project-management/manage.sh status
scripts/project-management/manage.sh urls
scripts/project-management/manage.sh down
```

`down` removes the containers and private networks but retains database and application volumes. To reset one product completely, run `podman compose down -v` from that product's service-library directory.

The preferred endpoints are `https://redmine.test`, `https://openproject.test`, `https://plane.test`, `https://glpi.test`, `https://leantime.test`, and `https://vikunja.test`. They are routed through the shared Nginx Proxy Manager container. Refresh local name resolution after catalog changes with `scripts/hosts/sync-hosts.sh`.

These credentials and secrets are for local evaluation only. Change them before exposing any service beyond loopback.
