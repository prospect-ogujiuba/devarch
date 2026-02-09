CREATE TABLE instance_env_files (
    id SERIAL PRIMARY KEY,
    instance_id INTEGER NOT NULL REFERENCES service_instances(id) ON DELETE CASCADE,
    path TEXT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    UNIQUE(instance_id, path)
);

CREATE INDEX idx_instance_env_files_instance_id ON instance_env_files(instance_id);

CREATE TABLE instance_networks (
    id SERIAL PRIMARY KEY,
    instance_id INTEGER NOT NULL REFERENCES service_instances(id) ON DELETE CASCADE,
    network_name TEXT NOT NULL,
    UNIQUE(instance_id, network_name)
);

CREATE INDEX idx_instance_networks_instance_id ON instance_networks(instance_id);

CREATE TABLE instance_config_mounts (
    id SERIAL PRIMARY KEY,
    instance_id INTEGER NOT NULL REFERENCES service_instances(id) ON DELETE CASCADE,
    config_file_id INTEGER REFERENCES service_config_files(id),
    source_path TEXT NOT NULL,
    target_path TEXT NOT NULL,
    readonly BOOLEAN NOT NULL DEFAULT false,
    UNIQUE(instance_id, target_path)
);

CREATE INDEX idx_instance_config_mounts_instance_id ON instance_config_mounts(instance_id);
