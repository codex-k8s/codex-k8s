-- name: configentry__upsert_project :one
INSERT INTO config_entries (
    scope,
    kind,
    project_id,
    repository_id,
    key,
    value_plain,
    value_encrypted,
    sync_targets,
    mutability,
    is_dangerous,
    created_by_user_id,
    updated_by_user_id
)
VALUES (
    'project',
    $1::text,
    $2::uuid,
    NULL,
    $3::text,
    $4::text,
    $5::bytea,
    $6::text[],
    $7::text,
    $8::boolean,
    NULLIF($9::text, '')::uuid,
    NULLIF($10::text, '')::uuid
)
ON CONFLICT (scope, project_id, key) WHERE scope = 'project' DO UPDATE
SET kind = EXCLUDED.kind,
    value_plain = EXCLUDED.value_plain,
    value_encrypted = EXCLUDED.value_encrypted,
    sync_targets = EXCLUDED.sync_targets,
    mutability = EXCLUDED.mutability,
    is_dangerous = EXCLUDED.is_dangerous,
    updated_by_user_id = EXCLUDED.updated_by_user_id,
    updated_at = NOW()
RETURNING
    id,
    scope,
    kind,
    COALESCE(project_id::text, '') AS project_id,
    COALESCE(repository_id::text, '') AS repository_id,
    key,
    CASE WHEN kind = 'variable' THEN value_plain ELSE '' END AS value,
    sync_targets,
    mutability,
    is_dangerous,
    updated_at::text AS updated_at;

