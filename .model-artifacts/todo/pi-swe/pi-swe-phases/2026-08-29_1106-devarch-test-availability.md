# devarch-test-availability

Created: 2026-08-29T11:06:41.799Z
Purpose: Define the persistent dashboard availability implementation slice.

# DevArch test availability

Created: 2026-08-29
Purpose: Make the completed dashboard persistently reachable at the canonical local domain.

## Goal
Serve DevArch Home at `https://devarch.test` without manual terminal startup.

## Scope
- Repository-owned systemd user unit template and idempotent installer.
- Dashboard host binding suitable for Nginx Proxy Manager container access.
- Replace the stale `devarch-app:80` upstream in the existing `devarch.test` Nginx block.
- Documentation and focused installer tests where practical.

## Outputs
- `scripts/dashboard/devarch-dashboard.service.in`
- `scripts/dashboard/install-service.sh`
- Updated Nginx custom configuration and dashboard documentation.

## Acceptance criteria
- User service starts and restarts automatically.
- Nginx container reaches the host dashboard through `host.containers.internal:7411`.
- `https://devarch.test` and `/api/inventory` return HTTP 200.
- No automatic browser polling is introduced.

## Verification
Shell syntax, Python and dashboard tests, systemd unit state, `nginx -t`, container-to-host request, and HTTPS smoke requests.

## Non-goals
Containerizing the dashboard, exposing lifecycle controls, or replacing Nginx Proxy Manager.
