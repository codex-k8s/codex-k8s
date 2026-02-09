-- name: learningfeedback__list_for_run :many
SELECT
    id,
    run_id,
    COALESCE(repository_id::text, '') AS repository_id,
    COALESCE(pr_number, 0) AS pr_number,
    COALESCE(file_path, '') AS file_path,
    COALESCE(line, 0) AS line,
    kind,
    explanation,
    created_at
FROM learning_feedback
WHERE run_id = $1::uuid
ORDER BY created_at ASC
LIMIT $2;

