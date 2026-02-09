-- name: projectmember__upsert :exec
INSERT INTO project_members (project_id, user_id, role, created_at, updated_at)
VALUES ($1::uuid, $2::uuid, $3, NOW(), NOW())
ON CONFLICT (project_id, user_id) DO UPDATE
SET role = EXCLUDED.role,
    updated_at = NOW();

