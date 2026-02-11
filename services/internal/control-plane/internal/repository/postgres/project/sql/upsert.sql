-- name: project__upsert :one
INSERT INTO projects (id, slug, name, settings, created_at, updated_at)
VALUES (
    $1::uuid,
    $2,
    $3,
    COALESCE($4::jsonb, '{}'::jsonb),
    NOW(),
    NOW()
)
ON CONFLICT (slug) DO UPDATE
SET name = EXCLUDED.name,
    settings = COALESCE(projects.settings, '{}'::jsonb) || COALESCE(EXCLUDED.settings, '{}'::jsonb),
    updated_at = NOW()
RETURNING id, slug, name;

