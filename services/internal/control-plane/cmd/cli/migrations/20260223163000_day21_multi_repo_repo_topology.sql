-- +goose Up

ALTER TABLE repositories
    ADD COLUMN IF NOT EXISTS alias TEXT,
    ADD COLUMN IF NOT EXISTS role TEXT NOT NULL DEFAULT 'service',
    ADD COLUMN IF NOT EXISTS default_ref TEXT NOT NULL DEFAULT 'main',
    ADD COLUMN IF NOT EXISTS docs_root_path TEXT NULL;

UPDATE repositories
SET alias = TRIM(BOTH '-' FROM LOWER(regexp_replace(COALESCE(owner, '') || '-' || COALESCE(name, ''), '[^a-zA-Z0-9._-]+', '-', 'g')))
WHERE alias IS NULL OR BTRIM(alias) = '';

UPDATE repositories
SET alias = SUBSTRING(id::text FROM 1 FOR 8)
WHERE alias IS NULL OR BTRIM(alias) = '';

WITH ranked AS (
    SELECT
        id,
        project_id,
        alias,
        ROW_NUMBER() OVER (PARTITION BY project_id, alias ORDER BY created_at, id) AS rn
    FROM repositories
)
UPDATE repositories AS r
SET alias = r.alias || '-' || ranked.rn
FROM ranked
WHERE r.id = ranked.id
  AND ranked.rn > 1;

ALTER TABLE repositories
    ALTER COLUMN alias SET NOT NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'chk_repositories_role'
    ) THEN
        ALTER TABLE repositories
            ADD CONSTRAINT chk_repositories_role
                CHECK (role IN ('orchestrator', 'service', 'docs', 'mixed'));
    END IF;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS uq_repositories_project_alias
    ON repositories (project_id, alias);

-- +goose Down

DROP INDEX IF EXISTS uq_repositories_project_alias;

ALTER TABLE repositories
    DROP CONSTRAINT IF EXISTS chk_repositories_role,
    DROP COLUMN IF EXISTS docs_root_path,
    DROP COLUMN IF EXISTS default_ref,
    DROP COLUMN IF EXISTS role,
    DROP COLUMN IF EXISTS alias;
