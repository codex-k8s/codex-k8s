-- name: runqueue__upsert_project :exec
-- Preserve existing explicit learning_mode_default in project settings.
-- Merge incoming defaults only when the key is still absent.
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
