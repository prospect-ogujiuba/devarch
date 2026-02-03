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
