-- +goose Up

CREATE TABLE IF NOT EXISTS repositories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    provider TEXT NOT NULL,
    external_id BIGINT NOT NULL,
    owner TEXT NOT NULL,
    name TEXT NOT NULL,
    token_encrypted BYTEA NOT NULL,
    services_yaml_path TEXT NOT NULL DEFAULT 'services.yaml',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_repositories_provider CHECK (provider IN ('github', 'gitlab')),
    CONSTRAINT uq_repositories_provider_external_id UNIQUE (provider, external_id),
    CONSTRAINT uq_repositories_project_provider_owner_name UNIQUE (project_id, provider, owner, name)
);

CREATE INDEX IF NOT EXISTS idx_repositories_project_id
    ON repositories (project_id);

CREATE INDEX IF NOT EXISTS idx_repositories_provider_owner_name
    ON repositories (provider, owner, name);

-- +goose Down
DROP INDEX IF EXISTS idx_repositories_provider_owner_name;
DROP INDEX IF EXISTS idx_repositories_project_id;
DROP TABLE IF EXISTS repositories;
