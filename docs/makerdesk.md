# MakerDesk integration and operations guide

This is the canonical DevArch guide for integrating and operating MakerDesk in Playground. MakerDesk is a standalone TypeRocket/MakerMaker plugin at [github.com/prospect-ogujiuba/makerdesk](https://github.com/prospect-ogujiuba/makerdesk); its repository remains the authority for plugin internals and development. Playground clones it into `apps/playground/wp-content/plugins/makerdesk` as an integration site.

Operational entry points:

- Portal: `https://playground.test/makerdesk/`
- App Galaxy: `podman exec php php /var/www/html/playground/galaxy_makerdesk`
- Apply migrations: append `migrate up --no-interaction`
- Run SLA sweep: append `makerdesk:sla-sweep`
- Seed repeatable local demo data: append `makerdesk:seed --no-interaction`
- Rotate demo-user credentials: append `makerdesk:seed --reset-passwords --no-interaction`
- WordPress admin: one **MakerDesk** top-level menu with native TypeRocket resource submenus and resource-specific Add New actions
- [Plugin operations and verification](https://github.com/prospect-ogujiuba/makerdesk/blob/main/README.md)
- [Standalone roadmap](https://github.com/prospect-ogujiuba/makerdesk/blob/main/docs/plans/2026-08-23_2229-00-phase-index.md)

## Fresh Playground integration

MakerDesk is deliberately pulled after the generic Playground profile is installed:

```bash
scripts/wordpress/bootstrap.sh playground --profile clean --force --no-hosts
podman exec php wp plugin deactivate playground-app --path=/var/www/html/playground --allow-root
podman exec php wp plugin delete playground-app --path=/var/www/html/playground --allow-root
git clone git@github.com:prospect-ogujiuba/makerdesk.git apps/playground/wp-content/plugins/makerdesk
podman exec php wp plugin activate makerdesk --path=/var/www/html/playground --allow-root
podman exec php wp makermaker register-plugin-galaxy --plugin=makerdesk \
  --namespace=Maker/MakerDesk --path=/var/www/html/playground --allow-root
podman exec php php /var/www/html/playground/galaxy_makerdesk migrate up --no-interaction
```

`bootstrap.sh --force` moves the prior site into `apps/.devarch-backups/` before resetting its database. The generic profile currently generates a temporary `playground-app`; remove it before cloning MakerDesk so there is no competing application plugin.

Application resources must continue to start from the MakerDesk Galaxy scaffold. Do not place project behavior in MakerMaker or TypeRocket core.
