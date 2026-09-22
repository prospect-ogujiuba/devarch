# GLPI and OpenProject production migration plan

## Pilot baseline

- GLPI image: `glpi/glpi:11.0.9`
- MariaDB image: `mariadb:11.4`
- OpenProject image: `openproject/openproject:17.8.0`
- PostgreSQL image: `postgres:17-alpine`
- Both web ports are bound to `127.0.0.1` and published through the local HTTPS proxy.
- Runtime secrets are stored only in each stack's gitignored, mode-`0600` `.env` file.

Do not change application or database major versions during the transfer. Restore onto the same versions first, validate, and upgrade separately.

## Pilot operations

1. Put both `.env` files in a password manager and encrypted off-host backup.
2. Back up nightly to encrypted storage outside the workstation.
3. Retain at least 7 daily and 4 weekly restore points.
4. Every backup set must include:
   - GLPI MariaDB dump and the entire `glpi-data` volume, including `files/`, configuration, `glpicrypt.key`, plugins, and marketplace content.
   - OpenProject PostgreSQL dump and the entire `openproject-assets` volume.
   - Both `.env` files, especially `OPENPROJECT_SECRET_KEY_BASE`.
   - A manifest containing image tags/digests, creation time, and checksums.
5. Perform a restore drill into isolated temporary volumes before relying on the backup.

## Production preparation

1. Provision a patched Linux host with supported Podman or Docker Compose, persistent storage, monitoring, and encrypted off-host backups.
2. Use production DNS names and a trusted TLS certificate. Expose only ports 80/443 through a reverse proxy; keep databases private.
3. Generate new database root credentials where feasible. Preserve `OPENPROJECT_SECRET_KEY_BASE` and GLPI encryption keys because existing encrypted values depend on them.
4. Configure SMTP, time zone, authentication/SSO, retention, and monitoring.
5. Restrict host and backup access, configure a firewall, and enable automatic security updates.
6. Record the production image digests and test the target with disposable data.

## Final cutover

1. Announce a maintenance window and prevent new writes in both applications.
2. Stop application containers while leaving the database containers available for final dumps.
3. Create final database dumps and archive application data volumes.
4. Generate SHA-256 checksums; copy the encrypted backup set to production and verify checksums.
5. Start the same application/database versions on production with empty volumes.
6. Restore databases into empty databases and restore application data with ownership and permissions preserved.
7. Restore the OpenProject `SECRET_KEY_BASE` and GLPI encryption key/configuration. Do not regenerate them.
8. Start the stacks. Run required migrations only through the vendor-provided startup/seeder or CLI for the deployed version.
9. Validate login, projects/tickets, attachments, email, scheduled jobs, API integrations, permissions, and audit history.
10. Switch DNS/reverse-proxy routing, monitor logs and health checks, and keep the workstation copy stopped and unchanged for rollback.

## Acceptance and rollback

Production acceptance requires successful admin/user login, matching record counts, sampled attachment downloads, working outbound email, healthy scheduled jobs, and a fresh production backup that has passed a restore test.

If validation fails before acceptance, return DNS/routing to the stopped pilot, restart it, and preserve the failed production state for diagnosis. Never merge writes from both environments; repeat the cutover from a fresh final backup.

## Post-cutover

- Rotate administrative passwords after handoff.
- Confirm GLPI default accounts remain disabled.
- Remove pilot data only after production has operated successfully through the agreed rollback window.
- Apply upgrades one product at a time after taking and testing a new backup.
