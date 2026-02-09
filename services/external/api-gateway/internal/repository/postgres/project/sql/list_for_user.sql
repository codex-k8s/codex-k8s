-- name: project__list_for_user :many
SELECT p.id, p.slug, p.name, pm.role
FROM projects p
JOIN project_members pm ON pm.project_id = p.id
WHERE pm.user_id = $1::uuid
ORDER BY p.slug ASC
LIMIT $2;

