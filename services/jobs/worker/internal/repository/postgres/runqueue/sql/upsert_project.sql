-- name: runqueue__upsert_project :exec
INSERT INTO projects (id, slug, name, created_at, updated_at)
VALUES ($1::uuid, $2, $3, NOW(), NOW())
ON CONFLICT (id) DO UPDATE
SET slug = EXCLUDED.slug,
    name = EXCLUDED.name,
    updated_at = NOW();

