# The server (5.161.85.187)

This is the `server` branch: the devarch checkout that the Hetzner server runs, at
`~/projects/devarch`. Develop on `master` (on the laptop), then merge `master` into `server`
and run `git pull` on the server. Changes that only apply to the server go on this branch.

- Ubuntu 24.04, rootless Podman 4.9.3 (user `priz`, lingering on), 2 vCPU, 2 GB memory + 2 GB swap.
- The `devarch` binary on the server is built on the laptop
  (`cd cli && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o devarch ./cmd/devarch`) and copied
  to `~/.local/bin/devarch`, because the server has no Go.
- `devarch doctor` passes. The `.test` hosts block is not used on the server.

## What runs here (2026-11-04)

| Name | What | Started by |
| --- | --- | --- |
| `nginx-proxy-manager` | The proxy on ports 80 and 443, on `microservices-net` | `~/.config/systemd/user/container-nginx-proxy-manager.service` |
| `prospect-app`, `prospect-db` | prospectogujiuba.com (repo `prospect-ogujiuba`) | Quadlets in `~/.config/containers/systemd/` |
| `mxcro-web` | mxcro.com, the static build of `growth-partner` in `~/mxcro/dist` | Quadlet `mxcro-web.container` |
| `flowstate-*` | FlowState: **stopped** on 2026-11-04 (data kept; dump in `~/backups/`) | `podman start` by hand |
| `devarch-db`, `devarch-api`, `devarch-app` | The old devarch dashboard: stopped since March 2026 | none |

## The proxy

Cloudflare is in front of every domain and ends HTTPS; nginx gets plain HTTP on port 80.
Server blocks are written by hand, not in the nginx-proxy-manager UI:

- `/data/nginx/custom/http.conf`: the FlowState block, and one line that includes
  `/data/nginx/custom/sites/*.conf`.
- `/data/nginx/custom/sites/sites.conf`: prospectogujiuba.com, mxcro.com and the redirects
  (www, mediagrowthpartner.com). The source is `deploy/proxy/sites.conf` in the
  `prospect-ogujiuba` repo.
- `/data/nginx/proxy_host/1.conf` (FlowState, from the UI) is moved aside as
  `1.conf.off-flowstate-stopped`: it named a stopped container, which stops nginx from starting.

After a change: `podman exec nginx-proxy-manager nginx -t && podman exec nginx-proxy-manager nginx -s reload`.

## Still to do (needs sudo)

- SSH: `PasswordAuthentication no`, `PermitRootLogin prohibit-password`.
- Firewall: allow 22, 80 and 443 only.
- Reboot for the pending kernel and libc updates (every service above starts again on boot).
- Backups: a nightly database dump to a second place.
