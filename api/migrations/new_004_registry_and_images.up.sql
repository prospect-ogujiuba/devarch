CREATE TABLE registries (
    id SERIAL PRIMARY KEY,
    name VARCHAR(64) UNIQUE NOT NULL,
    base_url VARCHAR(256) NOT NULL,
    api_version VARCHAR(16),
    enabled BOOLEAN DEFAULT true,
    rate_limit_remaining INTEGER,
    rate_limit_reset_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE images (
    id SERIAL PRIMARY KEY,
    registry_id INTEGER REFERENCES registries(id),
    repository VARCHAR(256) NOT NULL,
    description TEXT,
    star_count INTEGER,
    pull_count BIGINT,
    is_official BOOLEAN DEFAULT false,
    last_synced_at TIMESTAMPTZ,
    UNIQUE(registry_id, repository)
);

CREATE INDEX idx_images_registry ON images(registry_id);
CREATE INDEX idx_images_repository ON images(repository);

CREATE TABLE image_tags (
    id SERIAL PRIMARY KEY,
    image_id INTEGER REFERENCES images(id) ON DELETE CASCADE,
    tag VARCHAR(128) NOT NULL,
    digest VARCHAR(128),
    size_bytes BIGINT,
    pushed_at TIMESTAMPTZ,
    last_synced_at TIMESTAMPTZ,
    UNIQUE(image_id, tag)
);

CREATE INDEX idx_image_tags_image ON image_tags(image_id);
CREATE INDEX idx_image_tags_tag ON image_tags(tag);

CREATE TABLE image_architectures (
    id SERIAL PRIMARY KEY,
    tag_id INTEGER REFERENCES image_tags(id) ON DELETE CASCADE,
    os VARCHAR(32) NOT NULL,
    architecture VARCHAR(32) NOT NULL,
    variant VARCHAR(32),
    digest VARCHAR(128),
    size_bytes BIGINT
);

CREATE INDEX idx_image_architectures_tag ON image_architectures(tag_id);

CREATE TABLE vulnerabilities (
    id SERIAL PRIMARY KEY,
    cve_id VARCHAR(32) UNIQUE NOT NULL,
    severity VARCHAR(16) NOT NULL,
    title VARCHAR(512),
    description TEXT,
    cvss_score DECIMAL(3,1),
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_vulnerabilities_severity ON vulnerabilities(severity);
CREATE INDEX idx_vulnerabilities_cve ON vulnerabilities(cve_id);

CREATE TABLE image_tag_vulnerabilities (
    id SERIAL PRIMARY KEY,
    tag_id INTEGER REFERENCES image_tags(id) ON DELETE CASCADE,
    vulnerability_id INTEGER REFERENCES vulnerabilities(id),
    package_name VARCHAR(256),
    installed_version VARCHAR(128),
    fixed_version VARCHAR(128),
    scanned_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(tag_id, vulnerability_id, package_name)
);

CREATE INDEX idx_image_tag_vulns_tag ON image_tag_vulnerabilities(tag_id);
CREATE INDEX idx_image_tag_vulns_vuln ON image_tag_vulnerabilities(vulnerability_id);
