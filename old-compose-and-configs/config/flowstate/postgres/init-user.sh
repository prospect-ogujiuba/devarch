#!/bin/bash
set -e

# Only create production user if not in dev mode
if [ "${APP_ENV:-local}" = "production" ]; then
    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
        CREATE SCHEMA IF NOT EXISTS flowstate;
        CREATE USER IF NOT EXISTS flowstate_user WITH PASSWORD '${DB_APP_PASSWORD:-changeme}';
        GRANT CONNECT ON DATABASE $POSTGRES_DB TO flowstate_user;
        GRANT USAGE ON SCHEMA flowstate TO flowstate_user;
        GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA flowstate TO flowstate_user;
        GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA flowstate TO flowstate_user;
        ALTER DEFAULT PRIVILEGES IN SCHEMA flowstate GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO flowstate_user;
        ALTER DEFAULT PRIVILEGES IN SCHEMA flowstate GRANT USAGE, SELECT ON SEQUENCES TO flowstate_user;
        ALTER USER flowstate_user SET search_path TO flowstate;
EOSQL
fi
