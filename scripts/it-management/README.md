# IT management workbook import

These scripts migrate `IT_Management_System.xlsm` into the local OpenProject and Taiga deployments without modifying the workbook.

## Mapping

- Employees become OpenProject users and department groups, plus Taiga users and department project roles.
- Current and archived tickets become OpenProject work packages and Taiga issues.
- Printers and equipment provisions become dedicated work-package/issue types.
- Workbook-only values are retained in custom fields/custom attributes.
- OpenProject maps workbook workflow labels to its closest native status and keeps the original label in `Original Status`.
- Taiga creates the workbook ticket statuses and normalized priorities within the imported project.
- Taiga has no global group equivalent, so department roles are used in the imported project.
- Values in `Parent Task` are retained as references. The workbook values are categories rather than ticket IDs, so no parent-child relationships are created.

Accounts with real, unique email addresses keep those addresses. Workbook rows containing the repeated placeholder `fluidhose.com` receive unique addresses in the reserved `employees.invalid` domain. New accounts receive random/unusable passwords and no invitation email is sent.

## Files

- `normalize_workbook.py` reads XLSM package XML and writes protected normalized JSON.
- `import_openproject.rb` performs an atomic, idempotent OpenProject import through Rails models.
- `import_taiga.py` performs an atomic, idempotent Taiga import through Django models.

Both importers use deterministic source keys and update their isolated `IT Management System` projects on subsequent runs rather than duplicating records. Back up both databases before every production run.
