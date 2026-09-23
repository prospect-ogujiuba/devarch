# IT management workbook import

These scripts migrate `IT_Management_System.xlsm` into the local GLPI, OpenProject, and Taiga deployments without modifying the workbook.

## Mapping

- Employees become GLPI users and department groups, OpenProject users and department groups, plus Taiga users and department project roles.
- Current and archived tickets become GLPI tickets, OpenProject work packages, and Taiga issues.
- Printers become GLPI printer assets and dedicated OpenProject/Taiga work-package or issue types.
- Equipment provisions become GLPI peripheral assets and dedicated OpenProject/Taiga work-package or issue types.
- Workbook-only values are retained in GLPI comments/source identifiers and OpenProject/Taiga custom fields or attributes.
- GLPI and OpenProject map workbook workflow labels to their closest native status while retaining the original values.
- Taiga creates the workbook ticket statuses and normalized priorities within the imported project.
- Taiga has no global group equivalent, so department roles are used in the imported project.
- Values in `Parent Task` are retained as references. The workbook values are categories rather than ticket IDs, so no parent-child relationships are created.

Accounts with real, unique email addresses keep those addresses. Workbook rows containing the repeated placeholder `fluidhose.com` receive unique addresses in the reserved `employees.invalid` domain. New accounts receive random/unusable passwords and no invitation email is sent.

## Files

- `normalize_workbook.py` reads XLSM package XML and writes protected normalized JSON.
- `import_glpi.php` performs an authenticated, transactional, idempotent GLPI import through GLPI application models.
- `import_openproject.rb` performs an atomic, idempotent OpenProject import through Rails models.
- `import_taiga.py` performs an atomic, idempotent Taiga import through Django models.

All importers use deterministic source keys and update existing imported records on subsequent runs rather than duplicating them. Back up the databases and application-data volumes before every production run.

## GLPI and OpenProject runbook

Normalize the workbook without changing it:

```bash
python3 scripts/it-management/normalize_workbook.py \
  '/path/to/IT_Management_System.xlsm' \
  /tmp/it-management-normalized.json
```

Copy the protected JSON and importer into each application container. GLPI requires the administrator password only for the duration of the import; do not place it in command history or documentation.

```bash
podman cp scripts/it-management/import_glpi.php glpi:/tmp/import_glpi.php
podman cp /tmp/it-management-normalized.json glpi:/tmp/it-management-normalized.json
podman exec -e GLPI_IMPORT_ADMIN_LOGIN=glpi \
  -e GLPI_IMPORT_ADMIN_PASSWORD glpi \
  php /tmp/import_glpi.php /tmp/it-management-normalized.json

podman cp scripts/it-management/import_openproject.rb openproject:/tmp/import_openproject.rb
podman cp /tmp/it-management-normalized.json openproject:/tmp/it-management-normalized.json
podman exec openproject bundle exec rails runner \
  /tmp/import_openproject.rb /tmp/it-management-normalized.json
```

Run each importer a second time and verify it reports only updates, with no new users, assets, tickets, or work packages. Remove copied credentials and normalized data from the containers after verification.
