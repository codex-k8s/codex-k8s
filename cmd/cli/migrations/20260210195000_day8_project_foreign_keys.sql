-- +goose Up

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

