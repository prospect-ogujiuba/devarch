# connection-refused-diagnosis

Created: 2026-08-29T11:06:32.630Z
Purpose: Record the confirmed cause and smallest persistent devarch.test fix.

# Diagnosis: DevArch dashboard connection refusal

## Problem
`127.0.0.1:7411` refused connections and the desired canonical URL is `https://devarch.test`.

## Reproduction
- `curl http://127.0.0.1:7411/` returned connection refused.
- `devarch.test` resolves to `127.0.0.1`.

## Minimized case
The dashboard implementation was present but no dashboard process was running. The existing Nginx `devarch.test` block still targeted removed legacy upstream `devarch-app:80`.

## Hypotheses and evidence
- Confirmed: absent process caused port 7411 refusal.
- Confirmed: a temporary dashboard bound to `0.0.0.0:7411` returned HTTP 200 from the Nginx container through `host.containers.internal:7411`.
- Rejected: DNS or Nginx container availability; `devarch.test` resolves locally and Nginx Proxy Manager is running on ports 80/443.

## Candidate fix-slice
Add a repository-owned user-systemd unit installer for persistent startup, bind the service for container-to-host access, and update the existing Nginx block to proxy `host.containers.internal:7411`. Retain localhost/default safety for direct manual starts and document installation.

## Verification gaps
Verify the installed unit, Nginx configuration reload, HTTPS page, and HTTPS inventory endpoint.
