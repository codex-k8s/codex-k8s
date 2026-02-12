-- Deprecated: use get_latest_by_repository_branch_and_agent.sql in runtime code.
-- Kept temporarily for compatibility with older references.
SELECT
  id,
  run_id,
  correlation_id,
  project_id,
  repository_full_name,
  agent_key,
  issue_number,
  branch_name,
  pr_number,
  pr_url,
  trigger_kind,
  template_kind,
  template_source,
  template_locale,
  model,
  reasoning_effort,
  status,
  session_id,
  session_json,
  codex_cli_session_path,
  codex_cli_session_json,
  started_at,
  finished_at,
  created_at,
  updated_at
FROM agent_sessions
WHERE repository_full_name = $1
  AND branch_name = $2
  AND agent_key = $3
ORDER BY created_at DESC
LIMIT 1;
