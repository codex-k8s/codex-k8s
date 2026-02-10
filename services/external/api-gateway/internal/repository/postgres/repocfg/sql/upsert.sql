-- name: repocfg__upsert :one
INSERT INTO repositories (
    project_id,
    provider,
    external_id,
    owner,
    name,
    token_encrypted,
    services_yaml_path,
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
    NOW(),
    NOW()
)
ON CONFLICT (provider, external_id) DO UPDATE
SET owner = EXCLUDED.owner,
    name = EXCLUDED.name,
    token_encrypted = EXCLUDED.token_encrypted,
    services_yaml_path = EXCLUDED.services_yaml_path,
    updated_at = NOW()
WHERE repositories.project_id = EXCLUDED.project_id
RETURNING id, project_id, provider, external_id, owner, name, services_yaml_path;

