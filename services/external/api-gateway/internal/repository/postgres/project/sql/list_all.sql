-- name: project__list_all :many
SELECT id, slug, name
FROM projects
ORDER BY slug ASC
LIMIT $1;

