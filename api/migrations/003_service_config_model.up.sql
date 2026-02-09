CREATE TABLE service_config_files (
    id SERIAL PRIMARY KEY,
    service_id INTEGER NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    file_path VARCHAR(512) NOT NULL,
    content TEXT NOT NULL DEFAULT '',
    file_mode VARCHAR(10) NOT NULL DEFAULT '0644',
    is_template BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(service_id, file_path)
);

CREATE INDEX idx_service_config_files_service_id ON service_config_files(service_id);

CREATE TABLE service_config_mounts (
    id SERIAL PRIMARY KEY,
    service_id INTEGER NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    config_file_id INTEGER REFERENCES service_config_files(id),
    source_path TEXT NOT NULL,
    target_path TEXT NOT NULL,
    readonly BOOLEAN NOT NULL DEFAULT false,
    UNIQUE(service_id, target_path)
);

CREATE TABLE service_config_versions (
    id SERIAL PRIMARY KEY,
    service_id INTEGER NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    version INTEGER NOT NULL DEFAULT 1,
    config_snapshot JSONB NOT NULL,
    change_summary TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_config_versions_service ON service_config_versions(service_id, version DESC);
