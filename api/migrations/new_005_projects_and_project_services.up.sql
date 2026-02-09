CREATE TABLE projects (
    id SERIAL PRIMARY KEY,
    name VARCHAR(128) UNIQUE NOT NULL,
    path VARCHAR(512) NOT NULL,
    project_type VARCHAR(32) NOT NULL,
    framework VARCHAR(64),
    language VARCHAR(32),
    package_manager VARCHAR(16),
    description TEXT,
    version VARCHAR(32),
    license VARCHAR(64),
    entry_point VARCHAR(256),
    has_frontend BOOLEAN DEFAULT false,
    frontend_framework VARCHAR(64),
    domain VARCHAR(256),
    proxy_port INTEGER,
    dependencies JSONB DEFAULT '{}',
    scripts JSONB DEFAULT '{}',
    git_remote VARCHAR(512),
    git_branch VARCHAR(128),
    compose_path VARCHAR(512),
    service_count INTEGER DEFAULT 0,
    stack_id INTEGER,  -- FK to stacks added in 006_stacks_core.up.sql
    last_scanned_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_projects_type ON projects(project_type);
CREATE INDEX idx_projects_language ON projects(language);
CREATE INDEX idx_projects_stack_id ON projects(stack_id);

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
