DROP TABLE IF EXISTS instance_config_files;
DROP TABLE IF EXISTS instance_healthchecks;
DROP TABLE IF EXISTS instance_domains;
DROP TABLE IF EXISTS instance_labels;
DROP TABLE IF EXISTS instance_env_vars;
DROP TABLE IF EXISTS instance_volumes;
DROP TABLE IF EXISTS instance_ports;

DROP INDEX IF EXISTS idx_service_instances_deleted_at;
DROP INDEX IF EXISTS uq_instances_stack_instance_active;

ALTER TABLE service_instances
    DROP COLUMN IF EXISTS deleted_at,
    DROP COLUMN IF EXISTS description;

ALTER TABLE service_instances
    ADD CONSTRAINT service_instances_stack_id_instance_id_key UNIQUE(stack_id, instance_id);
