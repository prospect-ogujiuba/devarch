CREATE TABLE service_exports (
    id SERIAL PRIMARY KEY,
    service_id INTEGER NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    name VARCHAR(128) NOT NULL,
    type VARCHAR(64) NOT NULL,
    port INTEGER NOT NULL,
    protocol VARCHAR(10) DEFAULT 'tcp',
    UNIQUE(service_id, name)
);

CREATE INDEX idx_service_exports_service_id ON service_exports(service_id);

CREATE TABLE service_import_contracts (
    id SERIAL PRIMARY KEY,
    service_id INTEGER NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    name VARCHAR(128) NOT NULL,
    type VARCHAR(64) NOT NULL,
    required BOOLEAN DEFAULT true,
    env_vars JSONB DEFAULT '{}',
    UNIQUE(service_id, name)
);

CREATE INDEX idx_service_import_contracts_service_id ON service_import_contracts(service_id);

CREATE TABLE service_instance_wires (
    id SERIAL PRIMARY KEY,
    stack_id INTEGER NOT NULL REFERENCES stacks(id) ON DELETE CASCADE,
    consumer_instance_id INTEGER NOT NULL REFERENCES service_instances(id) ON DELETE CASCADE,
    provider_instance_id INTEGER NOT NULL REFERENCES service_instances(id) ON DELETE CASCADE,
    import_contract_id INTEGER NOT NULL REFERENCES service_import_contracts(id),
    export_contract_id INTEGER NOT NULL REFERENCES service_exports(id),
    source VARCHAR(10) NOT NULL DEFAULT 'auto' CHECK (source IN ('auto', 'explicit')),
    provider_contract_type VARCHAR(64),
    consumer_contract_type VARCHAR(64),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(stack_id, consumer_instance_id, import_contract_id)
);

CREATE INDEX idx_service_instance_wires_stack_id ON service_instance_wires(stack_id);
CREATE INDEX idx_service_instance_wires_consumer ON service_instance_wires(consumer_instance_id);
CREATE INDEX idx_service_instance_wires_provider ON service_instance_wires(provider_instance_id);

CREATE TABLE sync_state (
    key TEXT PRIMARY KEY,
    value TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE container_states (
    id SERIAL PRIMARY KEY,
    service_id INTEGER REFERENCES services(id) ON DELETE CASCADE UNIQUE,
    container_id VARCHAR(64),
    status VARCHAR(32) NOT NULL,
    health_status VARCHAR(32),
    restart_count INTEGER DEFAULT 0,
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    exit_code INTEGER,
    error TEXT,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_container_states_service ON container_states(service_id);
CREATE INDEX idx_container_states_status ON container_states(status);

CREATE TABLE container_metrics (
    id SERIAL PRIMARY KEY,
    service_id INTEGER REFERENCES services(id) ON DELETE CASCADE,
    cpu_percentage DECIMAL(5,2),
    memory_used_mb DECIMAL(10,2),
    memory_limit_mb DECIMAL(10,2),
    memory_percentage DECIMAL(5,2),
    network_rx_bytes BIGINT,
    network_tx_bytes BIGINT,
    recorded_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_container_metrics_service_time ON container_metrics(service_id, recorded_at DESC);
