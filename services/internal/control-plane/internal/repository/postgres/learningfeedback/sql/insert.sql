-- name: learningfeedback__insert :one
INSERT INTO learning_feedback (
    run_id,
    repository_id,
    pr_number,
    file_path,
    line,
    kind,
    explanation
)
VALUES (
    $1::uuid,
    NULLIF($2, '')::uuid,
    $3,
    $4,
    $5,
    $6,
    $7
)
RETURNING id;

