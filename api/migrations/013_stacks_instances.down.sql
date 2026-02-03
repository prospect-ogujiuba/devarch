DROP INDEX IF EXISTS idx_service_instances_template_service_id;
DROP INDEX IF EXISTS idx_service_instances_stack_id;
DROP INDEX IF EXISTS idx_stacks_deleted_at;
DROP INDEX IF EXISTS uq_stacks_name_active;
DROP TABLE IF EXISTS service_instances;
DROP TABLE IF EXISTS stacks;
