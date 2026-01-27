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
    last_scanned_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_projects_type ON projects(project_type);
CREATE INDEX idx_projects_language ON projects(language);
