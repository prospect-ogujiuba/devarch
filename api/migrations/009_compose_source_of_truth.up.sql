ALTER TABLE services ADD COLUMN env_file VARCHAR(512);
ALTER TABLE service_volumes ADD COLUMN is_external BOOLEAN NOT NULL DEFAULT false;
