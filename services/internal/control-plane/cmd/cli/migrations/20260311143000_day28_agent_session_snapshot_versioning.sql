-- +goose Up

ALTER TABLE agent_sessions
    ADD COLUMN IF NOT EXISTS snapshot_version BIGINT NOT NULL DEFAULT 1;

ALTER TABLE agent_sessions
    ADD COLUMN IF NOT EXISTS snapshot_checksum TEXT NULL;

ALTER TABLE agent_sessions
    ADD COLUMN IF NOT EXISTS snapshot_updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

UPDATE agent_sessions
SET snapshot_version = 1,
    snapshot_checksum = encode(
        digest(
            jsonb_build_object(
                'codex_cli_session_json', COALESCE(codex_cli_session_json, 'null'::jsonb),
                'session_json', COALESCE(session_json, '{}'::jsonb)
            )::text,
            'sha256'
        ),
        'hex'
    ),
    snapshot_updated_at = COALESCE(updated_at, created_at, NOW())
WHERE snapshot_version IS DISTINCT FROM 1
   OR snapshot_checksum IS NULL
   OR snapshot_updated_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_agent_sessions_repo_branch_agent_snapshot_updated
    ON agent_sessions (repository_full_name, branch_name, agent_key, snapshot_updated_at DESC, snapshot_version DESC);

-- +goose Down

DROP INDEX IF EXISTS idx_agent_sessions_repo_branch_agent_snapshot_updated;

ALTER TABLE agent_sessions
    DROP COLUMN IF EXISTS snapshot_updated_at;

ALTER TABLE agent_sessions
    DROP COLUMN IF EXISTS snapshot_checksum;

ALTER TABLE agent_sessions
    DROP COLUMN IF EXISTS snapshot_version;
