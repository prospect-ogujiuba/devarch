CREATE TABLE instance_dependencies (
    id SERIAL PRIMARY KEY,
    instance_id INTEGER NOT NULL REFERENCES service_instances(id) ON DELETE CASCADE,
    depends_on TEXT NOT NULL,
    condition VARCHAR(32) DEFAULT 'service_started',
    UNIQUE(instance_id, depends_on)
);

CREATE INDEX idx_instance_dependencies_instance_id ON instance_dependencies(instance_id);
