CREATE TABLE project_services (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    service_name VARCHAR(128) NOT NULL,
    container_name VARCHAR(128),
    image VARCHAR(256),
    service_type VARCHAR(32),
    ports JSONB DEFAULT '[]',
    depends_on JSONB DEFAULT '[]',
    UNIQUE(project_id, service_name)
);

ALTER TABLE projects ADD COLUMN compose_path VARCHAR(512);
ALTER TABLE projects ADD COLUMN service_count INTEGER DEFAULT 0;
