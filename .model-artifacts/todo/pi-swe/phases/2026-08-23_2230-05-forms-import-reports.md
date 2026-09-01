# 05-forms-import-reports

Created: 2026-08-23T22:30:43.390Z
Purpose: Deliver operational UI, workbook import, and actionable dashboards

# Phase 5 — Forms, workbook import, and reports

Created: 2026-08-23
Purpose: Replace spreadsheet record forms with efficient application workflows and safely migrate/reference workbook data.

## Goal

Make the expanded data model practical for daily IT work and provide script-driven ingestion instead of manual re-entry.

## Scope

- Apply established TypeRocket form conventions: left tabs, fieldsets, rows, searchable model selects, help text, required markers, conditional sections, read-only system/audit panels.
- Build focused list/detail/edit views for tickets, employees, assets/printers, and assignments/provisions.
- Add a `makerdesk:import-workbook` Galaxy command using an approved XLSX/XLSM reader, with `--dry-run`, sheet selection, mapping report, collision policy, resumable batches, and import-run ID.
- Do not execute workbook macros; read cell values only.
- Normalize Excel serial dates, formula-cache values, blank/zero ambiguity, human names, inconsistent enumerations, and broken form references.
- Resolve people by explicit mapping (employee number/email/confirmed WordPress link), never names alone when ambiguous.
- Store row-level import success/error/warning evidence and actor/source attribution; reruns must be idempotent.
- Add CSV exports and dashboards for open/overdue/SLA work, workload, effort, assets by lifecycle, printer toner/service needs, provisions due/overdue, employees without/with WordPress links, and recent audit changes.
- Seed configurable workbook lookup values from DataTables without locking future options to the spreadsheet.

## Outputs

- Completed operational forms/lists/detail panels.
- Workbook import command plus mapping/config and dry-run report.
- Dashboard widgets, saved filters, and bounded formula-safe exports.
- User/operator documentation for backup, dry run, import, reconciliation, and rollback.

## Acceptance criteria

- The four spreadsheet record forms have equal-or-better application equivalents.
- Import dry run changes no data and reports every row decision.
- Import rerun does not duplicate tickets, employees, printers/assets, or provisions.
- Ambiguous identity matches are quarantined for review.
- All exports are capability-gated, bounded, escaped, and spreadsheet-formula-safe.
- Daily tasks can be completed without editing JSON or entering raw IDs.

## Verification

- Fixture workbook import contract covering valid rows, broken formulas, duplicates, ambiguous people, dates, and rollback.
- Form and list-screen contracts.
- Report/export bounds and formula-prefix tests.
- Seed/import integration smoke test against WordPress/MariaDB.
