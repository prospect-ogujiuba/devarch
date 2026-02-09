CREATE TABLE instance_resource_limits (
    id SERIAL PRIMARY KEY,
    instance_id INTEGER NOT NULL REFERENCES service_instances(id) ON DELETE CASCADE,
    cpu_limit TEXT,
    cpu_reservation TEXT,
    memory_limit TEXT,
    memory_reservation TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(instance_id)
);
