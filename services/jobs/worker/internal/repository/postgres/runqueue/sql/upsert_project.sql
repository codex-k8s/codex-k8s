-- name: runqueue__upsert_project :exec
INSERT INTO projects (id, slug, name, settings, created_at, updated_at)
VALUES ($1::uuid, $2, $3, COALESCE($4::jsonb, '{}'::jsonb), NOW(), NOW())
ON CONFLICT (id) DO UPDATE
SET slug = EXCLUDED.slug,
    name = EXCLUDED.name,
    settings = CASE
        WHEN (COALESCE(projects.settings, '{}'::jsonb) ? 'learning_mode_default') THEN projects.settings
        ELSE COALESCE(projects.settings, '{}'::jsonb) || COALESCE(EXCLUDED.settings, '{}'::jsonb)
    END,
    updated_at = NOW();
