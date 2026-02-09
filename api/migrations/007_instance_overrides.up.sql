CREATE TABLE instance_ports (
    id SERIAL PRIMARY KEY,
    instance_id INTEGER NOT NULL REFERENCES service_instances(id) ON DELETE CASCADE,
    host_ip VARCHAR(45) DEFAULT '0.0.0.0',
    host_port INTEGER NOT NULL,
    container_port INTEGER NOT NULL,
    protocol VARCHAR(10) DEFAULT 'tcp'
);

CREATE INDEX idx_instance_ports_instance_id ON instance_ports(instance_id);

CREATE TABLE instance_volumes (
    id SERIAL PRIMARY KEY,
    instance_id INTEGER NOT NULL REFERENCES service_instances(id) ON DELETE CASCADE,
    volume_type VARCHAR(20) DEFAULT 'bind',
    source TEXT NOT NULL,
    target TEXT NOT NULL,
    read_only BOOLEAN DEFAULT false,
    is_external BOOLEAN DEFAULT false
);

CREATE INDEX idx_instance_volumes_instance_id ON instance_volumes(instance_id);

CREATE TABLE instance_env_vars (
    id SERIAL PRIMARY KEY,
    instance_id INTEGER NOT NULL REFERENCES service_instances(id) ON DELETE CASCADE,
    key VARCHAR(255) NOT NULL,
    value TEXT,
    is_secret BOOLEAN DEFAULT false,
    encrypted_value TEXT,
    encryption_version INTEGER NOT NULL DEFAULT 0,
    UNIQUE(instance_id, key)
);

CREATE INDEX idx_instance_env_vars_instance_id ON instance_env_vars(instance_id);

CREATE TABLE instance_labels (
    id SERIAL PRIMARY KEY,
    instance_id INTEGER NOT NULL REFERENCES service_instances(id) ON DELETE CASCADE,
    key VARCHAR(255) NOT NULL,
    value TEXT NOT NULL,
    UNIQUE(instance_id, key)
);

CREATE INDEX idx_instance_labels_instance_id ON instance_labels(instance_id);

CREATE TABLE instance_domains (
    id SERIAL PRIMARY KEY,
    instance_id INTEGER NOT NULL REFERENCES service_instances(id) ON DELETE CASCADE,
    domain VARCHAR(255) NOT NULL,
    proxy_port INTEGER
);

CREATE INDEX idx_instance_domains_instance_id ON instance_domains(instance_id);

CREATE TABLE instance_healthchecks (
    id SERIAL PRIMARY KEY,
    instance_id INTEGER NOT NULL REFERENCES service_instances(id) ON DELETE CASCADE UNIQUE,
    test TEXT NOT NULL,
    interval_seconds INTEGER DEFAULT 30,
    timeout_seconds INTEGER DEFAULT 10,
    retries INTEGER DEFAULT 3,
    start_period_seconds INTEGER DEFAULT 0
);

CREATE TABLE instance_config_files (
    id SERIAL PRIMARY KEY,
    instance_id INTEGER NOT NULL REFERENCES service_instances(id) ON DELETE CASCADE,
    file_path TEXT NOT NULL,
    content TEXT NOT NULL,
    file_mode VARCHAR(10) DEFAULT '0644',
    is_template BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(instance_id, file_path)
);

CREATE INDEX idx_instance_config_files_instance_id ON instance_config_files(instance_id);

CREATE TABLE instance_dependencies (
    id SERIAL PRIMARY KEY,
    instance_id INTEGER NOT NULL REFERENCES service_instances(id) ON DELETE CASCADE,
    depends_on TEXT NOT NULL,
    condition VARCHAR(32) DEFAULT 'service_started',
    UNIQUE(instance_id, depends_on)
);

CREATE INDEX idx_instance_dependencies_instance_id ON instance_dependencies(instance_id);

CREATE TABLE instance_resource_limits (
    id SERIAL PRIMARY KEY,
    instance_id INTEGER NOT NULL REFERENCES service_instances(id) ON DELETE CASCADE,
    cpu_limit TEXT,
    cpu_reservation TEXT,
    memory_limit TEXT,
    memory_reservation TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(instance_id)
);
