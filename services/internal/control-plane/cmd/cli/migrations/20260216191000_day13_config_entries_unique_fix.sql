-- +goose Up

DROP INDEX IF EXISTS uq_config_entries_scope_key;

CREATE UNIQUE INDEX IF NOT EXISTS uq_config_entries_platform_key
    ON config_entries (scope, key)
    WHERE scope = 'platform';

CREATE UNIQUE INDEX IF NOT EXISTS uq_config_entries_project_key
    ON config_entries (scope, project_id, key)
    WHERE scope = 'project';

CREATE UNIQUE INDEX IF NOT EXISTS uq_config_entries_repository_key
    ON config_entries (scope, repository_id, key)
    WHERE scope = 'repository';

-- +goose Down
DROP INDEX IF EXISTS uq_config_entries_repository_key;
DROP INDEX IF EXISTS uq_config_entries_project_key;
DROP INDEX IF EXISTS uq_config_entries_platform_key;

CREATE UNIQUE INDEX IF NOT EXISTS uq_config_entries_scope_key
    ON config_entries (scope, project_id, repository_id, key);

