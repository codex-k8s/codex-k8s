-- +goose Up
INSERT INTO agents (agent_key, role_kind, project_id, name)
VALUES
    ('pm', 'system', NULL, 'AI Product Manager'),
    ('sa', 'system', NULL, 'AI Solution Architect'),
    ('em', 'system', NULL, 'AI Engineering Manager'),
    ('sre', 'system', NULL, 'AI SRE'),
    ('km', 'system', NULL, 'AI Knowledge Manager')
ON CONFLICT DO NOTHING;

-- +goose Down
DELETE FROM agents
WHERE project_id IS NULL
  AND agent_key IN ('pm', 'sa', 'em', 'sre', 'km');
