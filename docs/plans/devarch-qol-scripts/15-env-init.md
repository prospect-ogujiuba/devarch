# Phase 15: Environment initializer

Created: 2026-08-19
Purpose: Replace insecure example placeholders with a safe, repeatable first-run configuration process.

## Goal

Keep initialization close to native file/secret tools: create a fresh `.env` safely and check DevArch placeholders; avoid a general dotenv editor.

## Scope

- Initial creation uses native `install -m 600 .env.example .env` (or equivalent `umask` + copy) with no overwrite.
- Generate only documented fresh-install secrets with `openssl rand`; if OpenSSL is missing, stop with installation guidance rather than implement a random generator.
- `--check` uses literal allowlisted assignment matching to report missing/placeholders/permissions without loading the file.
- Defer credential rotation: established stateful services require service-specific native procedures and must not be “fixed” by editing `.env` alone.
- Validate only DevArch-owned email/domain/user inputs needed during fresh initialization.

## Native delegation

`install`/`chmod` own file mode, `openssl rand` owns secret generation, and the user/editor owns arbitrary dotenv editing. DevArch supplies the template and checks repository-specific placeholders; it does not source files or implement dotenv semantics.

## Outputs

- Fresh-file initializer, literal placeholder checker, recording OpenSSL tests, and first-run documentation.

## Acceptance criteria

- Existing `.env` is never modified; the initializer exits with guidance. Future migration/rotation must be a separate reviewed phase.
- Generated secrets come directly from `openssl rand`, meet documented byte requirements, and are written to the new file without console output.
- Shell substitutions, command substitutions, and unrelated keys are preserved as text and never executed.
- `--check` reports missing/placeholders/permissions without mutations or secret disclosure.
- No generic rotation mode exists; documentation points to database/service-native credential rotation.

## Verification

- Fixtures cover absent/existing files, literal malicious shell text, placeholders, OpenSSL failure, permissions, and interrupted fresh-file creation.
- Canary command-substitution fixtures prove no execution occurs.
- Run doctor against a generated temporary environment.

## Non-goals

No dotenv parser/writer, random generator, generic repair/rotation engine, production secret manager, live database credential rotation, GitHub token creation, or `.env` synchronization across machines.
