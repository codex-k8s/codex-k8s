-- +goose Up
ALTER TABLE runtime_deploy_tasks
    ADD COLUMN IF NOT EXISTS logs_json JSONB NOT NULL DEFAULT '[]'::jsonb;

-- +goose Down
ALTER TABLE runtime_deploy_tasks
    DROP COLUMN IF EXISTS logs_json;
