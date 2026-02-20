-- +goose Up
ALTER TABLE runtime_deploy_tasks
    DROP CONSTRAINT IF EXISTS chk_runtime_deploy_tasks_status;

ALTER TABLE runtime_deploy_tasks
    ADD CONSTRAINT chk_runtime_deploy_tasks_status
        CHECK (status IN ('pending', 'running', 'succeeded', 'failed', 'canceled'));

-- +goose Down
ALTER TABLE runtime_deploy_tasks
    DROP CONSTRAINT IF EXISTS chk_runtime_deploy_tasks_status;

ALTER TABLE runtime_deploy_tasks
    ADD CONSTRAINT chk_runtime_deploy_tasks_status
        CHECK (status IN ('pending', 'running', 'succeeded', 'failed'));
