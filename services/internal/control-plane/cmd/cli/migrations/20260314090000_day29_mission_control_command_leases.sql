ALTER TABLE mission_control_commands
    ADD COLUMN lease_owner TEXT,
    ADD COLUMN lease_until TIMESTAMPTZ;

CREATE INDEX idx_mission_control_commands_claimable
    ON mission_control_commands (status, lease_until, updated_at DESC, requested_at DESC)
    WHERE status IN ('accepted', 'queued');
