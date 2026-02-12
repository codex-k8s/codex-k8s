-- +goose Up
CREATE TABLE IF NOT EXISTS agents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_key TEXT NOT NULL,
    role_kind TEXT NOT NULL DEFAULT 'system',
    project_id UUID NULL,
    name TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    settings JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_agents_role_kind CHECK (role_kind IN ('system', 'custom'))
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_agents_system_agent_key
    ON agents (agent_key)
    WHERE project_id IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_agents_project_agent_key
    ON agents (project_id, agent_key)
    WHERE project_id IS NOT NULL;

INSERT INTO agents (agent_key, role_kind, project_id, name)
VALUES
    ('dev', 'system', NULL, 'AI Developer'),
    ('reviewer', 'system', NULL, 'AI Reviewer'),
    ('qa', 'system', NULL, 'AI QA')
ON CONFLICT DO NOTHING;

-- +goose Down
DROP INDEX IF EXISTS uq_agents_project_agent_key;
DROP INDEX IF EXISTS uq_agents_system_agent_key;
DROP TABLE IF EXISTS agents;
