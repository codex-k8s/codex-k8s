-- name: projectmember__list :many
SELECT pm.project_id, pm.user_id, u.email, pm.role
FROM project_members pm
JOIN users u ON u.id = pm.user_id
WHERE pm.project_id = $1::uuid
ORDER BY u.email ASC
LIMIT $2;

