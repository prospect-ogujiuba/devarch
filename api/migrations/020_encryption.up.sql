ALTER TABLE service_env_vars ADD COLUMN encrypted_value TEXT;
ALTER TABLE service_env_vars ADD COLUMN encryption_version INTEGER NOT NULL DEFAULT 0;
ALTER TABLE instance_env_vars ADD COLUMN encrypted_value TEXT;
ALTER TABLE instance_env_vars ADD COLUMN encryption_version INTEGER NOT NULL DEFAULT 0;
