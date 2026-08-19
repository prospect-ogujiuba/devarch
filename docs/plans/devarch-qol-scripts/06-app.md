# Phase 06: Framework-aware application helper

Created: 2026-08-19
Purpose: Provide consistent daily commands across Laravel and WordPress workspaces without hiding their native tools.

## Goal

Add `scripts/apps/app.sh` to detect application type and safely dispatch common framework commands through shared infrastructure.

## Scope

- Commands: `list`, `type`, `url`, `open`, `shell`, `composer`, `npm`, `artisan`, `wp`, and `logs`.
- Detect Laravel using `artisan` plus `composer.json`; detect WordPress using `wp-config.php` plus core structure; report ambiguous/unknown trees.
- Map host app paths to `/var/www/html/<app>` and use the shared PHP container with existing user-mapping conventions.
- Forward native arguments only after `--` where ambiguity exists.

## Outputs

- App helper, framework detector, fake-runtime tests, and docs linked from both bootstrap READMEs.

## Acceptance criteria

- Only direct children of `apps/` are addressable; hidden DevArch backup/recovery directories are excluded.
- App names and resolved paths cannot escape `apps/` through symlinks or traversal.
- Framework-specific commands fail before execution on the wrong app type.
- `list` and `type` work with no runtime.
- Commands preserve exit codes, terminal interactivity, and argument boundaries.
- No application files are modified unless the requested native command does so.

## Verification

- Fixture apps cover Laravel, WordPress, unknown, ambiguous, symlink escape, spaces, and missing runtime.
- Recording runtime asserts working directory, user mapping, container, and forwarded arguments.
- Existing Laravel and WordPress bootstrap suites pass.

## Non-goals

No per-app containers, automatic dependency installation, queue/scheduler management, or framework abstraction beyond dispatch.
