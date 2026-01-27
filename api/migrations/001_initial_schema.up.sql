CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(64) UNIQUE NOT NULL,
    display_name VARCHAR(128),
    color VARCHAR(7),
    startup_order INTEGER,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE services (
    id SERIAL PRIMARY KEY,
    name VARCHAR(128) UNIQUE NOT NULL,
    category_id INTEGER REFERENCES categories(id),
    image_name VARCHAR(256) NOT NULL,
    image_tag VARCHAR(64) DEFAULT 'latest',
    restart_policy VARCHAR(32) DEFAULT 'unless-stopped',
    command TEXT,
    user_spec VARCHAR(64),
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_services_category ON services(category_id);
CREATE INDEX idx_services_enabled ON services(enabled);

CREATE TABLE service_ports (
    id SERIAL PRIMARY KEY,
    service_id INTEGER REFERENCES services(id) ON DELETE CASCADE,
    host_ip VARCHAR(45) DEFAULT '127.0.0.1',
    host_port INTEGER NOT NULL,
    container_port INTEGER NOT NULL,
    protocol VARCHAR(8) DEFAULT 'tcp',
    UNIQUE(host_ip, host_port)
);

CREATE INDEX idx_service_ports_service ON service_ports(service_id);

CREATE TABLE service_volumes (
    id SERIAL PRIMARY KEY,
    service_id INTEGER REFERENCES services(id) ON DELETE CASCADE,
    volume_type VARCHAR(16) NOT NULL,
    source VARCHAR(512) NOT NULL,
    target VARCHAR(512) NOT NULL,
    read_only BOOLEAN DEFAULT false
);

CREATE INDEX idx_service_volumes_service ON service_volumes(service_id);

CREATE TABLE service_env_vars (
    id SERIAL PRIMARY KEY,
    service_id INTEGER REFERENCES services(id) ON DELETE CASCADE,
    key VARCHAR(256) NOT NULL,
    value TEXT,
    is_secret BOOLEAN DEFAULT false,
    UNIQUE(service_id, key)
);

CREATE INDEX idx_service_env_vars_service ON service_env_vars(service_id);

CREATE TABLE service_dependencies (
    id SERIAL PRIMARY KEY,
    service_id INTEGER REFERENCES services(id) ON DELETE CASCADE,
    depends_on_service_id INTEGER REFERENCES services(id),
    condition VARCHAR(32) DEFAULT 'service_started',
    UNIQUE(service_id, depends_on_service_id)
);

CREATE INDEX idx_service_dependencies_service ON service_dependencies(service_id);
CREATE INDEX idx_service_dependencies_depends_on ON service_dependencies(depends_on_service_id);

CREATE TABLE service_healthchecks (
    id SERIAL PRIMARY KEY,
    service_id INTEGER REFERENCES services(id) ON DELETE CASCADE UNIQUE,
    test TEXT NOT NULL,
    interval_seconds INTEGER DEFAULT 30,
    timeout_seconds INTEGER DEFAULT 10,
    retries INTEGER DEFAULT 3,
    start_period_seconds INTEGER DEFAULT 0
);

CREATE TABLE service_domains (
    id SERIAL PRIMARY KEY,
    service_id INTEGER REFERENCES services(id) ON DELETE CASCADE,
    domain VARCHAR(256) UNIQUE NOT NULL,
    proxy_port INTEGER
);

CREATE INDEX idx_service_domains_service ON service_domains(service_id);

CREATE TABLE service_labels (
    id SERIAL PRIMARY KEY,
    service_id INTEGER REFERENCES services(id) ON DELETE CASCADE,
    key VARCHAR(256) NOT NULL,
    value TEXT,
    UNIQUE(service_id, key)
);

CREATE INDEX idx_service_labels_service ON service_labels(service_id);
