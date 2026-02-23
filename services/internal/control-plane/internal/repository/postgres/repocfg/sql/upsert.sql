-- name: repocfg__upsert :one
INSERT INTO repositories (
    project_id,
    alias,
    role,
    default_ref,
    provider,
    external_id,
    owner,
    name,
    token_encrypted,
    bot_token_encrypted,
    services_yaml_path,
    docs_root_path,
    bot_username,
    bot_email,
    preflight_report_json,
    preflight_updated_at,
    created_at,
    updated_at
)
VALUES (
    $1::uuid,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    NULL,
    $10,
    NULLIF($11, ''),
    '',
    '',
    '{}'::jsonb,
    NULL,
    NOW(),
    NOW()
)
ON CONFLICT (provider, external_id) DO UPDATE
SET alias = EXCLUDED.alias,
    role = EXCLUDED.role,
    default_ref = EXCLUDED.default_ref,
    owner = EXCLUDED.owner,
    name = EXCLUDED.name,
    token_encrypted = EXCLUDED.token_encrypted,
    -- Preserve bot params and preflight report on normal repo upsert.
    services_yaml_path = EXCLUDED.services_yaml_path,
    docs_root_path = EXCLUDED.docs_root_path,
    updated_at = NOW()
WHERE repositories.project_id = EXCLUDED.project_id
RETURNING
    id,
    project_id,
    alias,
    role,
    default_ref,
    provider,
    external_id,
    owner,
    name,
    services_yaml_path,
    COALESCE(docs_root_path, '') AS docs_root_path,
    bot_username,
    bot_email,
    COALESCE(preflight_updated_at::text, '') AS preflight_updated_at;
