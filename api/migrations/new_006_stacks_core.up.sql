CREATE TABLE stacks (
    id SERIAL PRIMARY KEY,
    name VARCHAR(63) NOT NULL,
    description TEXT DEFAULT '',
    network_name VARCHAR(63),
    source VARCHAR(32) DEFAULT 'manual',
    enabled BOOLEAN DEFAULT true,
    deleted_at TIMESTAMPTZ DEFAULT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE UNIQUE INDEX uq_stacks_name_active ON stacks(name) WHERE deleted_at IS NULL;
CREATE INDEX idx_stacks_deleted_at ON stacks(deleted_at);

CREATE TABLE service_instances (
    id SERIAL PRIMARY KEY,
    stack_id INTEGER NOT NULL REFERENCES stacks(id) ON DELETE CASCADE,
    instance_id VARCHAR(63) NOT NULL,
    template_service_id INTEGER REFERENCES services(id),
    container_name VARCHAR(127),
    description TEXT DEFAULT '',
    enabled BOOLEAN DEFAULT true,
    deleted_at TIMESTAMPTZ DEFAULT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE UNIQUE INDEX uq_instances_stack_instance_active ON service_instances(stack_id, instance_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_service_instances_stack_id ON service_instances(stack_id);
CREATE INDEX idx_service_instances_template_service_id ON service_instances(template_service_id);
CREATE INDEX idx_service_instances_deleted_at ON service_instances(deleted_at);

ALTER TABLE projects ADD CONSTRAINT fk_projects_stack_id FOREIGN KEY (stack_id) REFERENCES stacks(id);
