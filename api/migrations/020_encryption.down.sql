ALTER TABLE instance_env_vars DROP COLUMN encryption_version;
ALTER TABLE instance_env_vars DROP COLUMN encrypted_value;
ALTER TABLE service_env_vars DROP COLUMN encryption_version;
ALTER TABLE service_env_vars DROP COLUMN encrypted_value;
