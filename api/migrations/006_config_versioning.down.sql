ALTER TABLE services DROP COLUMN IF EXISTS validation_errors;
ALTER TABLE services DROP COLUMN IF EXISTS last_validated_at;
ALTER TABLE services DROP COLUMN IF EXISTS config_status;
DROP TABLE IF EXISTS service_config_versions;
