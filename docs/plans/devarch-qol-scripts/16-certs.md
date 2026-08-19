# Phase 16: Local certificate manager

Created: 2026-08-19
Purpose: Make wildcard `.test` certificate state, renewal, and trust understandable and reproducible.

## Goal

Use `mkcert` as the native local-CA/trust workflow and OpenSSL only for inspection; DevArch supplies domain/path defaults.

## Scope

- `status` invokes `openssl x509`/`openssl pkey` for subject, SANs, issuer, validity, key match, and permissions.
- `install-ca` delegates to `mkcert -install`; `uninstall-ca` delegates to `mkcert -uninstall`, preserving mkcert's native prompts/platform behavior.
- `generate` invokes `mkcert -cert-file PATH -key-file PATH localhost <suffix> '*.<suffix>'` with DevArch paths/domains.
- Atomically move newly generated files into the proxy config only after native OpenSSL verification; preserve a timestamped backup.
- If mkcert is absent, stop with installation/manual OpenSSL guidance. Do not maintain custom OS trust-store adapters.

## Native delegation

mkcert owns local CA creation, OS trust-store integration, and certificate generation. OpenSSL owns certificate/key inspection. DevArch only resolves domain names and repository target paths.

## Outputs

- Domain/path resolver around mkcert and OpenSSL, recording tests, and platform documentation.

## Acceptance criteria

- `status` is read-only and works without a container runtime.
- Generated private keys use restrictive permissions and are never printed.
- Generation verifies SAN coverage and key match before replacing active files.
- Trust/untrust show the mkcert command and fingerprint/context, then preserve mkcert's own sudo/UAC interaction and exit status.
- Existing certificates remain intact on generation or trust failure.

## Verification

- Generate temporary certificates and test SAN, expiry, mismatch, permissions, atomic failure, and backup behavior.
- Recording mkcert/OpenSSL executables assert exact argv without host trust mutation.
- Manual trust testing is opt-in and documented separately.

## Open questions

Confirm the exact SAN set. The trust model is decided: use mkcert's local CA, not a directly trusted self-signed leaf.

## Non-goals

No custom CA, OpenSSL certificate generator, OS trust-store adapter, public ACME issuance, production certificates, unattended privilege escalation, or automatic proxy restart.
