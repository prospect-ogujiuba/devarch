CREATE TABLE service_config_files (
    id SERIAL PRIMARY KEY,
    service_id INTEGER NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    file_path VARCHAR(512) NOT NULL,
    content TEXT NOT NULL DEFAULT '',
    file_mode VARCHAR(10) NOT NULL DEFAULT '0644',
    is_template BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE (service_id, file_path)
);

CREATE INDEX idx_service_config_files_service_id ON service_config_files(service_id);
