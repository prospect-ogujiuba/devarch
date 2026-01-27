-- Config version history for audit/diffing
CREATE TABLE service_config_versions (
    id SERIAL PRIMARY KEY,
    service_id INTEGER NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    version INTEGER NOT NULL DEFAULT 1,
    config_snapshot JSONB NOT NULL,
    change_summary TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_config_versions_service ON service_config_versions(service_id, version DESC);

-- Validation status tracking on services
ALTER TABLE services ADD COLUMN config_status VARCHAR(20) DEFAULT 'imported'
    CHECK (config_status IN ('imported', 'validated', 'modified', 'broken'));
ALTER TABLE services ADD COLUMN last_validated_at TIMESTAMPTZ;
ALTER TABLE services ADD COLUMN validation_errors JSONB;
