-- name: learningfeedback__insert :exec
INSERT INTO learning_feedback (run_id, kind, explanation)
VALUES ($1::uuid, $2, $3);

