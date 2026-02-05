-- Persists cleanup timestamps across restarts; advisory locks prevent duplicate work.

CREATE TABLE sync_state (
    key        TEXT PRIMARY KEY,
    value      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO sync_state (key, value, updated_at)
VALUES ('last_daily_cleanup', '1970-01-01T00:00:00Z', '1970-01-01T00:00:00Z');
