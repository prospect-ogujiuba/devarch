ALTER TABLE projects ADD COLUMN stack_id INTEGER REFERENCES stacks(id);
CREATE INDEX idx_projects_stack_id ON projects(stack_id);
