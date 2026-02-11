INSERT INTO stacks (name, network_name, source, created_at, updated_at)
SELECT p.name, 'devarch-' || p.name || '-net', 'project', NOW(), NOW()
FROM projects p WHERE p.stack_id IS NULL
ON CONFLICT (name) WHERE deleted_at IS NULL DO NOTHING;

UPDATE projects p SET stack_id = s.id
FROM stacks s WHERE s.name = p.name AND p.stack_id IS NULL AND s.deleted_at IS NULL;

ALTER TABLE projects ALTER COLUMN stack_id SET NOT NULL;

ALTER TABLE stacks ADD COLUMN project_id INTEGER REFERENCES projects(id) ON DELETE SET NULL;
CREATE INDEX idx_stacks_project_id ON stacks(project_id);
UPDATE stacks s SET project_id = p.id FROM projects p WHERE p.stack_id = s.id;

DROP TABLE IF EXISTS project_services;
