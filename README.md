# DevArch

DevArch is now a simple local service library: a collection of Podman-compatible Compose service definitions plus application workspaces.

The old Go CLI, planning workflow, daemon/API code, and generated workspace surfaces have been removed. Use the compose files directly.

## Layout

- `services-library/<category>/<service>/compose.yml` — service compose definitions.
- `services-library/<category>/<service>/config/` — optional service configuration.
- `apps/<app>/` — application workspaces, typically separate repositories ignored by the top-level DevArch repository.
- `.env.example` — example environment values.

## Usage

Pick a service and run it with Podman and the configured Compose provider:

```bash
cd services-library/database/postgres
podman compose up -d
```

Systems configured with the standalone provider may use `podman-compose up -d` instead. Do not mix rootless users: containers and networks created by one user are not visible to another.

Most compose files attach to the external network `microservices-net`. Create it once as the same service user that runs the stack:

```bash
podman network create microservices-net
```

For persistent production services, manage the containers with systemd/Quadlet (or reviewed generated units) and enable lingering for the rootless service user.

## Development checks

```bash
# validate compose YAML structure
python - <<'PY'
from pathlib import Path
import yaml
for path in sorted(Path('services-library').glob('**/compose.yml')):
    data = yaml.safe_load(path.read_text())
    assert isinstance(data, dict) and 'services' in data, path
print('compose yaml ok')
PY

# check PHP syntax
php -l apps/serverinfo/index.php
```
