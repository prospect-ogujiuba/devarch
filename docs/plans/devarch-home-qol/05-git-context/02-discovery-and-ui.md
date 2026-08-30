# Phase 2 — Git discovery and UI

Created: 2026-08-29
Purpose: Implement bounded metadata collection and present it as context, not source control tooling.

## Goal

Collect bounded Git facts for each app and show concise, non-operational context on cards and details.

## Scope

Read-only Git subprocess collection, bounded concurrency, deterministic integration, app card/detail presentation, graceful failures, and fixtures.

## Backend

- Detect worktrees without traversing arbitrary parent repositories; an app should represent its own worktree boundary.
- Use porcelain status to derive clean/dirty and counts, handling renamed/conflicted records without exposing names.
- Resolve branch or detached HEAD and retrieve only short HEAD plus committer ISO timestamp.
- Run repositories through a small bounded executor; preserve deterministic app ordering regardless of completion order.
- Cache nothing beyond the current inventory request; manual refresh recomputes current state.

## UI

- App cards: compact branch label and clean/modified dot/badge; omit the row for non-repositories.
- App details: branch/detached label, clean/modified summary, short HEAD, and humanized last commit age with exact timestamp in accessible text/title.
- Git unavailable/timed out: subtle “Git status unavailable” text on details only; no alarming global runtime error.
- Filters/sorting plan may later consume Git state only through an additive follow-up, not in this phase.

## Outputs

- Git collector integrated into project discovery.
- Additive app inventory fields and failure codes.
- Responsive card badges and app-detail Git context.
- Repository fixture tests.

## Acceptance criteria

- Cards remain compact at mobile widths.
- Dirty status clearly means local working-tree changes, not remote divergence.
- Humanized age is computed client-side from a stable timestamp without recurring timers.
- One repository failure cannot suppress app, container, or service inventory.
- No additional endpoint or background refresh is introduced.

## Verification

Temporary Git repository fixtures, detached/unborn cases, missing executable, timeouts, JSON contract tests, and responsive app-card/detail checks.