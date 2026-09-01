# 03-assets-printers-provisions

Created: 2026-08-23T22:30:12.201Z
Purpose: Expand inventory and provisioning with full assignment history

# Phase 3 — Assets, printers, and provisions

Created: 2026-08-23
Purpose: Cover workbook printer and provision records with a stronger unified inventory and lifecycle model.

## Goal

Make equipment ownership, location, printer supplies, issue/return state, and custody history operationally reliable.

## Scope

- Extend Asset with department, primary employee, current assignment, condition, IP/hostname, purchase/warranty/service dates, supplier links, notes, and lifecycle status while retaining tag, serial, manufacturer, model, location, and flexible metadata.
- Printer subtype fields: toner model, toner stock quantity/reorder threshold, supplier URLs, IP address, last service date; keep them first-class and conditionally displayed.
- AssetAssignment resource maps workbook Provisions: human-readable assignment number, employee, asset or ad-hoc item description, item type, serial/tag snapshot, issued/expected-return/returned dates, status, issue/return condition, notes, issued/returned/created/updated actor IDs.
- Transactional issue, transfer, return, loss, repair, retire, and stock-adjustment services.
- Preserve historical snapshots if employee or asset names/details change.
- Derived current owner/status on Asset must remain consistent with open assignments.

## Outputs

- Additive asset migration and AssetAssignment resource.
- Printer-focused form tab and filtered inventory view without creating a duplicate printer identity table.
- Issue/return/transfer commands or controller actions with immutable events.
- Low-toner, overdue-return, warranty-expiry, and service-due filters/dashboard metrics.

## Acceptance criteria

- All Printers/Printer Form/Provisions/Provision Form fields are supported.
- Only one open custody assignment exists per serialized asset.
- Returns require return date, condition, and actor; issued assets identify recipient and issuer.
- Toner stock cannot silently become negative; adjustments are audited.
- Deleting/deactivating an employee or asset does not erase custody history.
- Common operational fields are not hidden inside raw JSON.

## Verification

- Database uniqueness/index/FK contracts.
- Transaction rollback and concurrency tests for issue/transfer/return.
- Printer conditional-form and URL/IP/date validation contracts.
- Policy coverage for view, inventory administration, issue/return, and sensitive history.

## Non-goals

No automatic network discovery or supplier purchasing integration in this phase.
