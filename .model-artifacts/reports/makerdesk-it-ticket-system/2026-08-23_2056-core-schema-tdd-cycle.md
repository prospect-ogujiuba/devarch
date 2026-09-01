# core-schema-tdd-cycle

Created: 2026-08-23T20:56:52.848Z
Purpose: Record the scaffold-derived core domain Red/Green/Refactor and migration verification.

# MakerDesk Core Schema TDD Cycle

Date: 2026-08-23

## Behavior

The generated MakerDesk resources expose the approved ticket-domain columns, queue indexes, and model fillables, and their migrations can be applied and rolled back.

## Test level

Static schema/model contract plus disposable-database integration.

## Red

- Added `playground-app/tests/schema-contract.php` after Galaxy generated all six resources.
- Command: `php tests/schema-contract.php`
- Result: failed with 72 missing domain-contract tokens because the generated resources still contained generic `name/description/is_active/sort_order` schemas.

## Green

- Overwrote only scaffold-generated migrations, models, HTTP fields, and factories.
- Added ticket identifiers, assignment/reference IDs, lifecycle/SLA timestamps, immutable activity data, catalog fields, and queue-focused indexes.
- Command: `php tests/schema-contract.php`
- Result: passed for six migrations and six models.

## Refactor

- Kept common generated audit conventions.
- Avoided cross-resource foreign keys where migration ordering and WordPress portability would make rollback fragile; retained user foreign keys and the activity-to-ticket ownership key.
- No MVC foundations were added by hand.

## Verification

- PHP lint: 57/57 files passed.
- All six generated models load inside WordPress.
- `migrate up`: six MakerDesk tables created.
- Confirmed destructive `migrate down 6`, then `migrate up`: 0 tables after rollback, 6 after reapply.
- TypeRocket and MakerMaker core worktrees were not modified.

## Follow-up

Phase 3 owns ticket number allocation, lifecycle transitions, capability installation, policy expansion, and activity creation. Phase 4 owns replacement of generic generated admin views with requester and agent experiences.
