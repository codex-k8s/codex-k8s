-- +goose Up
CREATE TABLE IF NOT EXISTS realtime_events (
    id BIGSERIAL PRIMARY KEY,
    topic TEXT NOT NULL,
    scope JSONB NOT NULL DEFAULT '{}'::jsonb,
    payload_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    correlation_id TEXT NOT NULL DEFAULT '',
    project_id UUID NULL REFERENCES projects(id) ON DELETE SET NULL,
    run_id UUID NULL REFERENCES agent_runs(id) ON DELETE SET NULL,
    task_id UUID NULL REFERENCES runtime_deploy_tasks(run_id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_realtime_events_created_at
    ON realtime_events (created_at DESC);

CREATE INDEX IF NOT EXISTS idx_realtime_events_topic_id
    ON realtime_events (topic, id DESC);

CREATE INDEX IF NOT EXISTS idx_realtime_events_project_id
    ON realtime_events (project_id, id DESC);

CREATE INDEX IF NOT EXISTS idx_realtime_events_run_id
    ON realtime_events (run_id, id DESC);

CREATE INDEX IF NOT EXISTS idx_realtime_events_task_id
    ON realtime_events (task_id, id DESC);

CREATE OR REPLACE FUNCTION notify_realtime_event_insert() RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    PERFORM pg_notify('codex_realtime', NEW.id::text);
    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_realtime_events_notify ON realtime_events;
CREATE TRIGGER trg_realtime_events_notify
AFTER INSERT ON realtime_events
FOR EACH ROW
EXECUTE FUNCTION notify_realtime_event_insert();

CREATE OR REPLACE FUNCTION publish_realtime_from_flow_events() RETURNS trigger
LANGUAGE plpgsql
AS $$
DECLARE
    v_run_id UUID;
    v_project_id UUID;
BEGIN
    SELECT ar.id, ar.project_id
      INTO v_run_id, v_project_id
      FROM agent_runs ar
     WHERE ar.correlation_id = NEW.correlation_id
     LIMIT 1;

    INSERT INTO realtime_events (
        topic,
        scope,
        payload_json,
        correlation_id,
        project_id,
        run_id,
        task_id
    )
    VALUES (
        'run.events',
        jsonb_strip_nulls(
            jsonb_build_object(
                'project_id', v_project_id,
                'run_id', v_run_id,
                'correlation_id', NEW.correlation_id
            )
        ),
        jsonb_build_object(
            'flow_event_id', NEW.id,
            'event_type', NEW.event_type,
            'actor_type', NEW.actor_type,
            'actor_id', NEW.actor_id,
            'created_at', NEW.created_at,
            'payload', NEW.payload
        ),
        COALESCE(NEW.correlation_id, ''),
        v_project_id,
        v_run_id,
        NULL
    );

    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_flow_events_realtime_publish ON flow_events;
CREATE TRIGGER trg_flow_events_realtime_publish
AFTER INSERT ON flow_events
FOR EACH ROW
EXECUTE FUNCTION publish_realtime_from_flow_events();

CREATE OR REPLACE FUNCTION publish_realtime_from_agent_runs_status() RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    INSERT INTO realtime_events (
        topic,
        scope,
        payload_json,
        correlation_id,
        project_id,
        run_id,
        task_id
    )
    VALUES (
        'run.status',
        jsonb_strip_nulls(
            jsonb_build_object(
                'project_id', NEW.project_id,
                'run_id', NEW.id,
                'correlation_id', NEW.correlation_id
            )
        ),
        jsonb_build_object(
            'run_id', NEW.id,
            'status', NEW.status,
            'started_at', NEW.started_at,
            'finished_at', NEW.finished_at,
            'updated_at', NEW.updated_at
        ),
        COALESCE(NEW.correlation_id, ''),
        NEW.project_id,
        NEW.id,
        NULL
    );

    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_agent_runs_status_realtime_publish ON agent_runs;
CREATE TRIGGER trg_agent_runs_status_realtime_publish
AFTER INSERT OR UPDATE OF status, started_at, finished_at, updated_at ON agent_runs
FOR EACH ROW
WHEN (
    TG_OP = 'INSERT'
    OR OLD.status IS DISTINCT FROM NEW.status
    OR OLD.started_at IS DISTINCT FROM NEW.started_at
    OR OLD.finished_at IS DISTINCT FROM NEW.finished_at
)
EXECUTE FUNCTION publish_realtime_from_agent_runs_status();

CREATE OR REPLACE FUNCTION publish_realtime_from_agent_runs_logs() RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    INSERT INTO realtime_events (
        topic,
        scope,
        payload_json,
        correlation_id,
        project_id,
        run_id,
        task_id
    )
    VALUES (
        'run.logs',
        jsonb_strip_nulls(
            jsonb_build_object(
                'project_id', NEW.project_id,
                'run_id', NEW.id,
                'correlation_id', NEW.correlation_id
            )
        ),
        jsonb_build_object(
            'run_id', NEW.id,
            'status', NEW.status,
            'updated_at', NEW.updated_at
        ),
        COALESCE(NEW.correlation_id, ''),
        NEW.project_id,
        NEW.id,
        NULL
    );

    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_agent_runs_logs_realtime_publish ON agent_runs;
CREATE TRIGGER trg_agent_runs_logs_realtime_publish
AFTER UPDATE OF agent_logs_json ON agent_runs
FOR EACH ROW
WHEN (OLD.agent_logs_json IS DISTINCT FROM NEW.agent_logs_json)
EXECUTE FUNCTION publish_realtime_from_agent_runs_logs();

CREATE OR REPLACE FUNCTION publish_realtime_from_runtime_deploy_tasks_status() RETURNS trigger
LANGUAGE plpgsql
AS $$
DECLARE
    v_project_id UUID;
    v_correlation_id TEXT;
BEGIN
    SELECT ar.project_id, ar.correlation_id
      INTO v_project_id, v_correlation_id
      FROM agent_runs ar
     WHERE ar.id = NEW.run_id
     LIMIT 1;

    INSERT INTO realtime_events (
        topic,
        scope,
        payload_json,
        correlation_id,
        project_id,
        run_id,
        task_id
    )
    VALUES (
        'deploy.events',
        jsonb_strip_nulls(
            jsonb_build_object(
                'project_id', v_project_id,
                'run_id', NEW.run_id,
                'task_id', NEW.run_id
            )
        ),
        jsonb_build_object(
            'run_id', NEW.run_id,
            'status', NEW.status,
            'target_env', NEW.target_env,
            'namespace', NEW.namespace,
            'updated_at', NEW.updated_at
        ),
        COALESCE(v_correlation_id, ''),
        v_project_id,
        NEW.run_id,
        NEW.run_id
    );

    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_runtime_deploy_tasks_status_realtime_publish ON runtime_deploy_tasks;
CREATE TRIGGER trg_runtime_deploy_tasks_status_realtime_publish
AFTER INSERT OR UPDATE OF status, updated_at, started_at, finished_at, last_error, result_namespace, result_target_env ON runtime_deploy_tasks
FOR EACH ROW
WHEN (
    TG_OP = 'INSERT'
    OR OLD.status IS DISTINCT FROM NEW.status
    OR OLD.last_error IS DISTINCT FROM NEW.last_error
    OR OLD.result_namespace IS DISTINCT FROM NEW.result_namespace
    OR OLD.result_target_env IS DISTINCT FROM NEW.result_target_env
    OR OLD.started_at IS DISTINCT FROM NEW.started_at
    OR OLD.finished_at IS DISTINCT FROM NEW.finished_at
)
EXECUTE FUNCTION publish_realtime_from_runtime_deploy_tasks_status();

CREATE OR REPLACE FUNCTION publish_realtime_from_runtime_deploy_tasks_logs() RETURNS trigger
LANGUAGE plpgsql
AS $$
DECLARE
    v_project_id UUID;
    v_correlation_id TEXT;
BEGIN
    SELECT ar.project_id, ar.correlation_id
      INTO v_project_id, v_correlation_id
      FROM agent_runs ar
     WHERE ar.id = NEW.run_id
     LIMIT 1;

    INSERT INTO realtime_events (
        topic,
        scope,
        payload_json,
        correlation_id,
        project_id,
        run_id,
        task_id
    )
    VALUES (
        'deploy.logs',
        jsonb_strip_nulls(
            jsonb_build_object(
                'project_id', v_project_id,
                'run_id', NEW.run_id,
                'task_id', NEW.run_id
            )
        ),
        jsonb_build_object(
            'run_id', NEW.run_id,
            'status', NEW.status,
            'updated_at', NEW.updated_at
        ),
        COALESCE(v_correlation_id, ''),
        v_project_id,
        NEW.run_id,
        NEW.run_id
    );

    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_runtime_deploy_tasks_logs_realtime_publish ON runtime_deploy_tasks;
CREATE TRIGGER trg_runtime_deploy_tasks_logs_realtime_publish
AFTER UPDATE OF logs_json ON runtime_deploy_tasks
FOR EACH ROW
WHEN (OLD.logs_json IS DISTINCT FROM NEW.logs_json)
EXECUTE FUNCTION publish_realtime_from_runtime_deploy_tasks_logs();

CREATE OR REPLACE FUNCTION publish_realtime_from_runtime_errors() RETURNS trigger
LANGUAGE plpgsql
AS $$
DECLARE
    v_topic TEXT;
BEGIN
    v_topic := 'system.errors';

    INSERT INTO realtime_events (
        topic,
        scope,
        payload_json,
        correlation_id,
        project_id,
        run_id,
        task_id
    )
    VALUES (
        v_topic,
        jsonb_strip_nulls(
            jsonb_build_object(
                'project_id', NEW.project_id,
                'run_id', NEW.run_id
            )
        ),
        jsonb_build_object(
            'runtime_error_id', NEW.id,
            'level', NEW.level,
            'source', NEW.source,
            'message', NEW.message,
            'viewed_at', NEW.viewed_at,
            'created_at', NEW.created_at
        ),
        COALESCE(NEW.correlation_id, ''),
        NEW.project_id,
        NEW.run_id,
        NULL
    );

    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_runtime_errors_realtime_publish ON runtime_errors;
CREATE TRIGGER trg_runtime_errors_realtime_publish
AFTER INSERT OR UPDATE OF viewed_at ON runtime_errors
FOR EACH ROW
WHEN (
    TG_OP = 'INSERT'
    OR OLD.viewed_at IS DISTINCT FROM NEW.viewed_at
)
EXECUTE FUNCTION publish_realtime_from_runtime_errors();

-- +goose Down
DROP TRIGGER IF EXISTS trg_runtime_errors_realtime_publish ON runtime_errors;
DROP FUNCTION IF EXISTS publish_realtime_from_runtime_errors();

DROP TRIGGER IF EXISTS trg_runtime_deploy_tasks_logs_realtime_publish ON runtime_deploy_tasks;
DROP FUNCTION IF EXISTS publish_realtime_from_runtime_deploy_tasks_logs();

DROP TRIGGER IF EXISTS trg_runtime_deploy_tasks_status_realtime_publish ON runtime_deploy_tasks;
DROP FUNCTION IF EXISTS publish_realtime_from_runtime_deploy_tasks_status();

DROP TRIGGER IF EXISTS trg_agent_runs_logs_realtime_publish ON agent_runs;
DROP FUNCTION IF EXISTS publish_realtime_from_agent_runs_logs();

DROP TRIGGER IF EXISTS trg_agent_runs_status_realtime_publish ON agent_runs;
DROP FUNCTION IF EXISTS publish_realtime_from_agent_runs_status();

DROP TRIGGER IF EXISTS trg_flow_events_realtime_publish ON flow_events;
DROP FUNCTION IF EXISTS publish_realtime_from_flow_events();

DROP TRIGGER IF EXISTS trg_realtime_events_notify ON realtime_events;
DROP FUNCTION IF EXISTS notify_realtime_event_insert();

DROP INDEX IF EXISTS idx_realtime_events_task_id;
DROP INDEX IF EXISTS idx_realtime_events_run_id;
DROP INDEX IF EXISTS idx_realtime_events_project_id;
DROP INDEX IF EXISTS idx_realtime_events_topic_id;
DROP INDEX IF EXISTS idx_realtime_events_created_at;

DROP TABLE IF EXISTS realtime_events;
