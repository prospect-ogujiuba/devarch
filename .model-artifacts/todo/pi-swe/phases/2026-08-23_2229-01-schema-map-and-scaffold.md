# 01-schema-map-and-scaffold

Created: 2026-08-23T22:29:42.176Z
Purpose: Map spreadsheet requirements and generate resource skeletons before custom code

# Phase 1 — Schema map and scaffold

Created: 2026-08-23
Purpose: Turn the workbook into a testable minimum contract and generate the application skeletons before manual implementation.

## Goal

Create an explicit workbook-to-domain mapping and scaffold the smallest coherent new resource set.

## Scope

- Treat sheets as minimum contracts: Tickets, Archive, Employees, Printers, Provisions, four record forms, and DataTables.
- Map Ticket columns: number, title/subject, description, type, priority, status, requester, assignee, parent task, start/due/completed dates, estimated/actual hours, notes.
- Map Employee columns: employee ID, full name, department, title, email, employee type, access-control ID, time-clock ID, card ID, status, notes; preserve hire date because the Employee Form requires it even though the source-table formula is broken.
- Map Printer columns: ID, location, department, model, primary user, status, toner model/stock, supplier links, notes; include form-intended serial number, IP address, and last service fields despite broken workbook references.
- Map Provision columns: ID, employee, item type/description, serial, asset tag, issue/return dates, status, condition, notes.
- Map DataTables values to validated configuration/enums without hard-coding organization-specific values in views.
- Decide additive resources: Employee, Department, AssetAssignment (provision history), AuditEvent, and optionally LookupOption; extend existing Asset for printers rather than duplicating inventory identity.
- Run MakerMaker resource scaffolds from the plugin Galaxy launcher for each approved new resource, including migration, views, factory, and tests.

## Outputs

- Durable field-mapping/spec artifact referenced by tests.
- Generated models, controllers, fields, policies, migrations, forms, indexes, factories, and test skeletons.
- Command transcript identifying which files were generated versus manually customized.

## Acceptance criteria

- Every workbook field has a destination, type, validation rule, relationship, and audit classification.
- Archive maps to soft-delete/closed lifecycle plus filtered views, not a disconnected copy table.
- Employee and WordPress user identities are separate but linkable.
- Printer is an asset subtype with first-class operational fields.
- Provision is modeled as assignment history, not only current asset owner state.
- Existing generated/custom files are never overwritten without review.

## Verification

- Run each generator with `--help`/dry-safe checks first where supported.
- Confirm expected scaffold files exist and core MakerMaker/TypeRocket repositories remain clean.
- Add a schema-map contract that fails for unmapped workbook fields.

## Non-goals

No manual business logic or UI customization before scaffold generation.
