DROP TABLE IF EXISTS project_services;
ALTER TABLE projects DROP COLUMN IF EXISTS compose_path;
ALTER TABLE projects DROP COLUMN IF EXISTS service_count;
