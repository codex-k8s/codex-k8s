-- +goose Up
ALTER TABLE agent_runs
    ADD COLUMN IF NOT EXISTS agent_logs_json JSONB NULL;

-- +goose Down
ALTER TABLE agent_runs
    DROP COLUMN IF EXISTS agent_logs_json;
