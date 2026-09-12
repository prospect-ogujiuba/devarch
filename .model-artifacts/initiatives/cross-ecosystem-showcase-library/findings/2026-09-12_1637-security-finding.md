# Security finding: cross-ecosystem showcase library

- Reviewed spec: r1 (`sha256:8d212d0bd6e0f2770f79080659bfd8b3fc7a81a7bc57b38172487185f3b4ec92`)
- Reviewed plan: r1 (`sha256:a2904fe4dd7cef344fa39f67fc701ba48f2d4d08dac0bc1fd77a3a38bb9c470d`)
- Decision: accept with required incorporation in r2

## Trust boundaries

Tracked definitions are repository-controlled data, but may be changed by contributors. Adapters and capability handlers execute package managers and framework installers with network access and write access to `apps/`. Generated package lifecycle scripts are upstream executable code. Local dotenv, Composer/npm credentials, SSH agents, databases, and hosts elevation are sensitive boundaries.

## Required decisions

- Parse new registry/profile/recipe files as data; never `source`, `eval`, shell-expand, or execute definition fields.
- Resolve adapter and capability IDs only through repository-contained, non-symlink regular files under fixed roots; canonicalize paths and reject traversal, duplicates, controls, whitespace ambiguity, and unsupported protocol versions before mutation.
- Use argv arrays and `--` boundaries. Do not construct shell command strings.
- Capability handlers own exact package/Artisan/npm operations and configuration keys. Definitions may select IDs and validated constraints only.
- Treat Composer/npm install scripts as explicit trusted-upstream execution; display package/starter sources in plans, provide a no-scripts diagnostic mode where feasible, and never claim profiles are sandboxed.
- Pass secrets through stdin or permission-restricted files where supported; redact plans/logs; recovery records contain IDs/status/paths/hashes only. Add secret-pattern scans using synthetic canaries.
- Reject credential-bearing Git/HTTPS URLs and external-account profiles from unattended golden paths.
- Prevent symlink writes/target escape beneath `apps/`; preserve pre-existing targets and refuse dangerous cleanup paths.
- Recipe cross-app outputs must be allowlisted non-secret values (origin/URL/name). Secret sharing requires an adapter-owned secure channel and is deferred.

## Verification

Malicious fixture corpus for command substitutions, separators, newlines/controls, traversal, symlinks, duplicate IDs, hostile constraints/URLs, stale recovery hashes, and forged adapter output. Assert no canary appears in argv, stdout/stderr, run logs, recovery state, or tracked definitions.

## Residual risk

Upstream installer/package code executes with the shared PHP/Node container's access. Document this and pin/validate sources; eliminating it is outside scope.
