-- name: realtime__check_project_membership :one
SELECT EXISTS (
    SELECT 1
    FROM project_members pm
    WHERE pm.project_id = $1::uuid
      AND pm.user_id = $2::uuid
);

