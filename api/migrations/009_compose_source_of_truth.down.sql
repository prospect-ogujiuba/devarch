ALTER TABLE service_volumes DROP COLUMN IF EXISTS is_external;
ALTER TABLE services DROP COLUMN IF EXISTS env_file;
