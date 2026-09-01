# 06-audit-security-verification

Created: 2026-08-23T22:31:00.713Z
Purpose: Harden cross-domain auditability, permissions, and release evidence

# Phase 6 — Audit, security, and verification

Created: 2026-08-23
Purpose: Ensure the spreadsheet expansion meets MakerDesk's existing auditability and WordPress authorization standards.

## Goal

Provide durable actor-attributed change history across people, inventory, provisions, tickets, imports, and configuration.

## Scope

- Scaffold/implement append-only AuditEvent with entity type/ID, action, actor WordPress user ID, source, request/import correlation ID, before/after or field-level diff, reason, timestamp, and safe metadata.
- Keep TicketActivity as the human/operational conversation and workflow timeline; use AuditEvent for administrative data mutations. Link events where useful rather than replacing one with the other.
- Centralize mutation services so create/update/delete/link/issue/return/import actions write record and audit evidence transactionally.
- Add least-privilege capabilities for people, sensitive employee IDs, inventory, provisioning, imports, configuration, audit viewing, and exports.
- Redact secrets and sensitive identifiers from logs, notifications, exports, and requester surfaces.
- Add retention and soft-delete rules; audit records are not generically editable/deletable.
- Verify upgrade/rollback, cron, routes, capabilities, and existing MakerDesk functionality.

## Outputs

- Audit schema/service/view and correlation support.
- Policy/capability matrix and privacy classification.
- Migration/rollback notes and release verification report.
- Updated README/IT operations guidance.

## Acceptance criteria

- Every privileged mutation records who, what, when, source, and changed fields.
- Subject user/employee and acting WordPress user cannot be confused.
- Audit failures prevent the protected mutation rather than silently losing evidence.
- Requesters cannot see internal notes, sensitive employee identifiers, or unrestricted audit diffs.
- Existing ticket activity remains immutable.
- Migration can be applied to existing data without destructive table replacement.

## Verification

- PHP lint for all non-vendor files.
- Existing pure contracts: schema, workflow, policy, SLA/report.
- New pure contracts: spreadsheet mapping, employee identity, asset assignment, ticket hierarchy/effort/recurrence, audit, import.
- WordPress/MariaDB integration contracts: database, portal, SLA notifications, admin sample, import fixture.
- Manual smoke: each role, each main form, issue/return, ticket completion, archive filters, audit trail, dry-run import, reports/exports.

## Non-goals

Do not claim production readiness without backup/restore rehearsal and real-data dry-run reconciliation.
