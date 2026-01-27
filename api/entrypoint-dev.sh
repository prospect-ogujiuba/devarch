#!/bin/sh
set -e

MIGRATIONS_DIR="${MIGRATIONS_DIR:-/app/migrations}"
COMPOSE_DIR="${COMPOSE_DIR:-/workspace/compose}"

echo "waiting for postgres..."
until devarch-migrate -cmd create-db 2>/dev/null; do
  sleep 2
done

echo "running migrations..."
devarch-migrate -cmd up -migrations "$MIGRATIONS_DIR"

if [ -d "$COMPOSE_DIR" ]; then
  SVC_COUNT=$(devarch-import --count-only 2>/dev/null || echo "0")
  if [ "$SVC_COUNT" = "0" ]; then
    echo "importing compose services (first boot)..."
    COMPOSE_DIR="$COMPOSE_DIR" devarch-import
  else
    echo "skipping import ($SVC_COUNT services already in DB)"
  fi
fi

echo "starting air (hot reload)..."
exec air -c .air.toml
