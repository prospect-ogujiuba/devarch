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
    config_status VARCHAR(20) DEFAULT 'imported' CHECK (config_status IN ('imported','validated','modified','broken')),
    last_validated_at TIMESTAMPTZ,
    validation_errors JSONB,
    compose_overrides JSONB DEFAULT '{}',
    container_name_template TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_services_category ON services(category_id);
CREATE INDEX idx_services_enabled ON services(enabled);
