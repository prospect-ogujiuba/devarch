# Phase 2: Scaffold core ticket domain

Created: 2026-08-23
Purpose: Generate the complete core domain foundation before applying ticket-specific edits.

## Goal

Create scaffold-derived resources for Ticket, TicketActivity, SupportTeam, TicketCategory, Asset, and SlaPolicy.

## Scope

- Generate each resource through `galaxy_playground_app make:maker-resource` with migration, views, factory, and tests where supported.
- Review generated paths before editing.
- Define ticket/reference schemas, foreign keys or indexed identifiers, timestamps, and safe rollback behavior.
- Use WordPress users and Media attachments rather than duplicating identity or blob storage.

## Outputs

- Generated models, controllers, fields, policies, registry entries, migrations, views, factories, and tests.
- Edited migrations and models for the core relationships.

## Acceptance criteria

- No resource MVC foundation is handwritten.
- Ticket supports requester, assignee/team, status, priority, impact, urgency, category, asset, SLA, due/resolution/closure timestamps, and searchable content.
- Activity records support public replies, internal notes, transitions, and immutable actor/timestamp data.
- Migrations are reversible and indexed for primary queue access patterns.

## Verification

- Generator output inspection.
- Focused generated tests plus schema assertions.
- Migration up/down on the disposable Playground database.

## Open questions

- Use native DB foreign keys only where WordPress table conventions make rollback and portability safe.
