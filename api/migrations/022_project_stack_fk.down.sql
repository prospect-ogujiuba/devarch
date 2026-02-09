DROP INDEX IF EXISTS idx_projects_stack_id;
ALTER TABLE projects DROP COLUMN IF EXISTS stack_id;
