-- name: runqueue__ensure_project_exists :exec
INSERT INTO projects (id, slug, name, settings, created_at, updated_at)
VALUES ($1::uuid, $2, $3, COALESCE($4::jsonb, '{}'::jsonb), NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

