# Phase 06: Framework-aware application helper

Created: 2026-08-19
Purpose: Provide consistent daily commands across Laravel and WordPress workspaces without hiding their native tools.

## Goal

Add `scripts/apps/app.sh` to detect application type and safely dispatch common framework commands through shared infrastructure.

## Scope

- DevArch-owned discovery modes: `list`, `type`, and `url`; execution modes only resolve the app/container path before calling a native tool.
- Detect Laravel using `artisan` plus `composer.json`; detect WordPress using `wp-config.php` plus core structure; report ambiguous/unknown trees.
- Map host app paths to `/var/www/html/<app>` and use the shared PHP container with existing user-mapping conventions.
- Dispatch with direct commands such as `podman exec -it --workdir /var/www/html/<app> php ...`, then native `composer`, `php artisan`, `wp`, or package-manager arguments after `--`.
- Use `podman logs` for logs and the operating-system launcher for open; do not proxy those implementations inside app logic.

## Native delegation

Podman owns exec/TTY/user behavior; Composer, Artisan, WP-CLI, npm/pnpm/bun own their command parsing and output. DevArch only validates the app path/type and supplies container/working-directory defaults. Where a direct native command is clear, documentation is preferred over another app subcommand.

## Outputs

- App helper, framework detector, fake-runtime tests, and docs linked from both bootstrap READMEs.

## Acceptance criteria

- Only direct children of `apps/` are addressable; hidden DevArch backup/recovery directories are excluded.
- App names and resolved paths cannot escape `apps/` through symlinks or traversal.
- Framework-specific commands fail before execution on the wrong app type.
- `list` and `type` work with no runtime.
- Final execution uses `exec podman exec ...` where possible and preserves exit codes, stdin/stdout/stderr, terminal interactivity, signals, and argument boundaries.
- No application files are modified unless the requested native command does so.

## Verification

- Fixture apps cover Laravel, WordPress, unknown, ambiguous, symlink escape, spaces, and missing runtime.
- Recording runtime asserts working directory, user mapping, container, and forwarded arguments.
- Existing Laravel and WordPress bootstrap suites pass.

## Non-goals

No per-app containers, command reformatting, log proxy, browser implementation, automatic dependency installation, queue/scheduler management, or framework abstraction beyond path-aware native dispatch.
