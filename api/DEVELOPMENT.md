# DevArch API Development Setup

## Prerequisites

- Go 1.21+
- Node.js 18+
- PostgreSQL (via devarch service)
- Podman or Docker

## Quick Start

```bash
# 1. Start PostgreSQL
./scripts/service-manager.sh up postgres

# 2. Run migrations
cd api
DATABASE_URL="postgres://postgres:admin1234567@localhost:8502/postgres?sslmode=disable" \
  go run ./cmd/migrate -migrations ./migrations

# 3. Import compose services
DATABASE_URL="postgres://postgres:admin1234567@localhost:8502/postgres?sslmode=disable" \
  go run ./cmd/import -compose-dir ../compose

# 4. Start API server
PORT=8081 DATABASE_URL="postgres://postgres:admin1234567@localhost:8502/postgres?sslmode=disable" \
  go run ./cmd/server

# 5. Start dashboard (separate terminal)
cd dashboard
npm run dev
```

Dashboard: http://localhost:5174
API: http://localhost:8081

## API Endpoints

| Endpoint | Description |
|----------|-------------|
| `GET /health` | Health check |
| `GET /api/v1/status` | Overview stats |
| `GET /api/v1/services` | List services |
| `GET /api/v1/services/{name}` | Get service |
| `POST /api/v1/services/{name}/start` | Start service |
| `POST /api/v1/services/{name}/stop` | Stop service |
| `GET /api/v1/categories` | List categories |
| `GET /api/v1/ws/status` | WebSocket status stream |

## Configuration

| Env Var | Default | Description |
|---------|---------|-------------|
| `DATABASE_URL` | `postgres://devarch:devarch@localhost:5432/devarch?sslmode=disable` | PostgreSQL connection |
| `PORT` | `8080` | API server port |
| `COMPOSE_DIR` | (required for import) | Path to compose files |

## Vite Proxy

Dashboard proxies `/api` requests to the Go API. Configure target in `dashboard/vite.config.ts`:

```ts
proxy: {
  '/api': {
    target: 'http://localhost:8081',
    changeOrigin: true,
  },
},
```
