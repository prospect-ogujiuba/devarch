# devarch-test-verification

Created: 2026-08-29T11:14:54.751Z
Purpose: Record persistent devarch.test availability verification.

# Verification evidence: devarch.test availability

Timestamp: 2026-08-29 07:16 EDT
Scope: persistent DevArch dashboard user service, Nginx route, Podman socket inventory, and canonical HTTPS endpoint.

## Checks

- Dashboard unit installation: `scripts/dashboard/install-service.sh`
  - Result: pass; installed, enabled, and restarted `devarch-dashboard.service`.
- User units: `systemctl --user is-active devarch-dashboard.service podman.socket`
  - Result: pass; both active.
- Unit tests: `python -m unittest discover -s scripts/dashboard/tests -v`
  - Result: pass; 7 tests.
- Syntax: Bash installer/start script, Python compile, JavaScript check
  - Result: pass.
- Nginx: `podman exec nginx-proxy-manager nginx -t`
  - Result: pass.
- Page: `curl -k https://devarch.test/`
  - Result: pass; HTTP 200.
- Inventory: `curl -k https://devarch.test/api/inventory`
  - Result: pass; HTTP 200 with 31 projects, 30 containers, 174 services, and no runtime error.
- Host allowlist
  - Result: direct canonical host returned 200; unrelated host returned 421.

## Gaps

- The repository's local TLS certificate is not trusted by command-line curl on this host, so verification used `-k`. Browser trust depends on the existing DevArch local-certificate setup.

## Outcome

Pass. DevArch Home is persistently running and available at `https://devarch.test` with complete inventory.
