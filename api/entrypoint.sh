#!/bin/sh
set -e

MIGRATIONS_DIR="${MIGRATIONS_DIR:-/app/migrations}"

echo "waiting for postgres..."
until devarch-migrate -cmd create-db 2>/dev/null; do
  sleep 2
done

echo "running migrations..."
devarch-migrate -cmd up -migrations "$MIGRATIONS_DIR"

echo "starting devarch-api..."
exec devarch-api
