# workflow-dsa-decision

Created: 2026-08-23T20:57:55.789Z
Purpose: Record lifecycle graph and concurrency-safe ticket numbering choices before Phase 3 implementation.

# MakerDesk Workflow DSA Decision

Date: 2026-08-23

## Problem summary

Ticket transitions must be explicit and fast to validate; human-readable numbers must remain unique under concurrent creates without introducing a sequence table.

## Current implementation

The generated Ticket model stores free-form status and has a unique `number` column but no transition or allocation behavior.

## Workload and constraints

Ticket writes are much less frequent than queue reads. Each state has a small fixed out-degree. Persistence is MariaDB through WordPress/TypeRocket. Concurrent creates are possible. Evidence is architectural/inferred; no production volume is measured.

## Recommendation

- Represent the workflow as a static adjacency map keyed by current status. Validate target membership against the small target list.
- Generate the final display number from the database-assigned ticket ID and creation year (`IT-YYYY-NNNNNN`). Insert first with a cryptographically random unique placeholder, then finalize inside one database transaction.
- Guard updates with `WHERE id = ? AND version = ?`, incrementing `version` atomically. Affected-row count of zero is a stale-write conflict.

## Rejected alternatives

- Database sequence table: adds locking and migration complexity without a semantic benefit.
- `MAX(number)+1`: race-prone and increasingly expensive.
- UUID-only display identifiers: safe but less usable for service-desk communication.
- Fully normalized transition table: unnecessary for a fixed application workflow and complicates deployment.

## Complexity impact

Transition lookup is O(1) by state plus O(k) target membership where k is bounded and small. Number generation is O(1). Optimistic update is an indexed primary-key operation, O(log n) in storage terms.

## Memory tradeoff

The fixed transition map is negligible and avoids per-request database reads.

## Migration advice

Keep the existing unique number key and version column. No additional persistent structure is required. Rollback remains the Phase 2 table rollback.

## Validation plan

Pure workflow contract tests, number format tests, invalid/closed transition tests, and a database integration test proving a stale version update affects zero rows.

## Confidence

High for correctness; performance expectations are inferred but rely only on indexed constant-size operations.
