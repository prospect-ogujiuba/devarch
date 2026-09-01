# 02-people-directory

Created: 2026-08-23T22:29:56.361Z
Purpose: Build an auditable employee directory with optional WordPress user linkage

# Phase 2 — People directory and WordPress identity

Created: 2026-08-23
Purpose: Represent employees as operational people records while preserving WordPress as the authentication and audit-actor system.

## Goal

Deliver employee/department management that supports users with and without WordPress accounts.

## Scope

- Employee: unique employee number, display/full name, department, title, email, employee type, access-control ID, time-clock ID, card ID, hire date, employment status, notes, optional unique `wp_user_id`, timestamps, soft delete, created/updated actor IDs.
- Department: name, optional parent department, status, manager employee, timestamps and actor IDs.
- Explicit link/unlink flow to an existing WordPress user; never silently create or merge accounts by matching names.
- Optional controlled account-provisioning action with capability check and explicit confirmation.
- Searchable employee and WordPress selectors showing meaningful identity context.
- Employee detail tabs: Employment, Identity/access, IT relationships, System/audit.
- Related views for tickets requested/assigned, assets held, and provisioning history.

## Outputs

- Employee and Department migrations/resources/policies/forms/list screens.
- Identity-link service that records old/new linkage and actor.
- Employee status transitions for active, leave, terminated, contractor/inactive as configurable values.
- Factories, seed examples, and relationship queries.

## Acceptance criteria

- Duplicate employee numbers and duplicate non-null WordPress links are rejected.
- Deactivating an employee does not delete WordPress users, tickets, assets, or history.
- WordPress user deletion nulls the login link while employee history remains.
- Sensitive physical-access/time-clock identifiers require manager capability and are omitted from requester-facing surfaces.
- Created/updated/link/unlink events identify the acting WordPress user.

## Verification

- Migration and FK/index contract.
- Policy tests for requester/agent/manager/admin roles.
- Identity link/unlink collision and user-deletion tests.
- Form contract for grouped fields, searchable relationships, help text, and read-only audit fields.

## Non-goals

No HR payroll or full identity-provider synchronization.
