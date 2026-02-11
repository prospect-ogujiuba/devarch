CREATE TABLE IF NOT EXISTS project_services (
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

DROP INDEX IF EXISTS idx_stacks_project_id;
ALTER TABLE stacks DROP COLUMN IF EXISTS project_id;

ALTER TABLE projects ALTER COLUMN stack_id DROP NOT NULL;
