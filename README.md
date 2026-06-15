# DevArch

DevArch is now a simple local service library: a collection of ready-to-run Docker Compose service definitions plus a small PHP server-info app.

The old Go CLI, planning workflow, daemon/API code, and generated workspace surfaces have been removed. Use the compose files directly.

## Layout

- `services-library/<category>/<service>/compose.yml` — service compose definitions.
- `services-library/<category>/<service>/config/` — optional service configuration.
- `apps/serverinfo/` — small PHP server info app.
- `.env.example` — example environment values.

## Usage

Pick a service and run it with Docker Compose or a compatible Compose implementation:

```bash
cd services-library/database/postgres
docker compose up -d
```

Most compose files attach to the external network `microservices-net`. Create it once if your Compose runtime reports it missing:

```bash
docker network create microservices-net
```

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
