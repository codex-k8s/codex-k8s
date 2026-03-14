-- +goose Up
ALTER TABLE mission_control_commands
    ADD COLUMN IF NOT EXISTS lease_owner TEXT,
    ADD COLUMN IF NOT EXISTS lease_until TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_mission_control_commands_claimable
    ON mission_control_commands (status, lease_until, updated_at DESC, requested_at DESC)
    WHERE status IN ('accepted', 'queued');

-- +goose Down
DROP INDEX IF EXISTS idx_mission_control_commands_claimable;

ALTER TABLE mission_control_commands
    DROP COLUMN IF EXISTS lease_until,
    DROP COLUMN IF EXISTS lease_owner;
