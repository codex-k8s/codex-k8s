-- +goose Up
CREATE TABLE IF NOT EXISTS project_databases (
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    environment TEXT NOT NULL,
    database_name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT pk_project_databases PRIMARY KEY (database_name),
    CONSTRAINT chk_project_databases_environment_not_empty CHECK (BTRIM(environment) <> '')
);

CREATE INDEX IF NOT EXISTS idx_project_databases_project_environment
    ON project_databases (project_id, environment);

-- +goose Down
DROP INDEX IF EXISTS idx_project_databases_project_environment;
DROP TABLE IF EXISTS project_databases;
