-- +goose Up

-- Legacy cleanup: slots/agent_runs predate project FK constraints.
-- If a project was deleted or runs were created without a proper catalog entry,
-- keep the DB consistent by:
-- - dropping orphan slots (they will be re-created by the worker as needed)
-- - nulling orphan run.project_id (preserve run history)
DELETE FROM slots
WHERE project_id NOT IN (SELECT id FROM projects);

UPDATE agent_runs
SET project_id = NULL
WHERE project_id IS NOT NULL
  AND project_id NOT IN (SELECT id FROM projects);

ALTER TABLE agent_runs
    DROP CONSTRAINT IF EXISTS fk_agent_runs_project_id;

ALTER TABLE agent_runs
    ADD CONSTRAINT fk_agent_runs_project_id
        FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE;

ALTER TABLE slots
    DROP CONSTRAINT IF EXISTS fk_slots_project_id;

ALTER TABLE slots
    ADD CONSTRAINT fk_slots_project_id
        FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE;

-- +goose Down

ALTER TABLE slots
    DROP CONSTRAINT IF EXISTS fk_slots_project_id;

ALTER TABLE agent_runs
    DROP CONSTRAINT IF EXISTS fk_agent_runs_project_id;
