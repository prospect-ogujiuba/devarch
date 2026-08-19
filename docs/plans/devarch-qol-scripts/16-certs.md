# Phase 16: Local certificate manager

Created: 2026-08-19
Purpose: Make wildcard `.test` certificate state, renewal, and trust understandable and reproducible.

## Goal

Add `scripts/devarch/certs.sh` for certificate inspection, generation, and explicitly approved platform trust operations.

## Scope

- Commands: `status`, `generate`, `trust`, `untrust`, and `paths`.
- Inspect subject, SANs, issuer, validity, key/certificate match, permissions, and renewal threshold without exposing private material.
- Prefer `mkcert` when available; define an OpenSSL fallback with SANs for the configured domain suffix and wildcard.
- Atomically replace certificate/key only after validation and preserve a timestamped backup.
- Implement platform-specific trust adapters with dry-run and visible elevation boundaries.

## Outputs

- Certificate script, config templates if needed, trust adapters, tests, and platform documentation.

## Acceptance criteria

- `status` is read-only and works without a container runtime.
- Generated private keys use restrictive permissions and are never printed.
- Generation verifies SAN coverage and key match before replacing active files.
- Trust/untrust require explicit confirmation, show exact certificate fingerprint, and never hide sudo/UAC interaction.
- Existing certificates remain intact on generation or trust failure.

## Verification

- Generate temporary certificates and test SAN, expiry, mismatch, permissions, atomic failure, and backup behavior.
- Fake platform trust stores assert exact commands without host mutation.
- Manual trust testing is opt-in and documented separately.

## Open questions

Confirm whether `.test` wildcard trust should use a local CA (`mkcert`, recommended) or a directly trusted self-signed leaf on each supported platform.

## Non-goals

No public ACME issuance, production certificates, unattended privilege escalation, or automatic proxy restart.
