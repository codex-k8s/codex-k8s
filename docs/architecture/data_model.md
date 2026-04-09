---
doc_id: DM-CK8S-0001
type: data-model
title: "kodex — Data Model"
status: active
owner_role: SA
created_at: 2026-02-06
updated_at: 2026-03-15
related_issues: [1, 19, 100, 247, 248, 249, 500]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-19-full-docset"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-19
---

# Data Model: kodex

## TL;DR
- Ключевые сущности: users, projects, system_settings, system_setting_changes, repositories, project_databases, agents, agent_runs, worker_instances, agent_sessions, token_usage, slots, runtime_deploy_tasks, flow_events, links, docs_meta, doc_chunks.
- Основные связи: user<->project (RBAC), project->repositories, project->project_databases, agent->agent_runs, issue/pr->doc links.
- Риски миграций: ранний выбор индексов для webhook/event throughput и vector search.

## Сущности
### Entity: users
- Назначение: пользователи платформы (без self-signup).
- Важные инварианты: уникальный email; OAuth login должен матчиться с разрешённым email.
- Поля:

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| id | uuid | no | gen_random_uuid() | pk | |
| email | text | no |  | unique | |
| github_login | text | yes |  |  | |
| role_global | text | no | "user" | check enum | |
| created_at | timestamptz | no | now() |  | |

### Entity: projects
- Назначение: проекты в платформе.
- Важные инварианты: уникальное имя проекта.
- Поля:

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| id | uuid | no | gen_random_uuid() | pk | |
| key | text | no |  | unique | short id |
| name | text | no |  | unique | |
| settings | jsonb | no | '{}'::jsonb |  | project-level flags (`learning_mode_default`, repository/onboarding hints, etc.) |

### Entity: project_members
- Назначение: доступы пользователей к проектам.
- Важные инварианты: один пользователь имеет одну роль в проекте.
- Поля:

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| project_id | uuid | no |  | fk -> projects | |
| user_id | uuid | no |  | fk -> users | |
| role | text | no |  | check(read/read_write/admin) | |
| learning_mode_enabled | bool | no | false |  | user-level override |

### Entity: system_settings
- Назначение: глобальные настройки платформы.
- Важные инварианты:
  - ключ уникален и берётся из typed platform settings catalog;
  - `version` монотонно растёт при каждом изменении effective value;
  - `value_json` хранит typed persisted snapshot, а не произвольный settings blob.
- Поля:

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| key | text | no |  | pk | typed setting key |
| value_kind | text | no |  | check(boolean) | typed value contract |
| value_json | jsonb | no |  |  | persisted effective value |
| source | text | no |  | check(default/staff) | current owner of value |
| version | bigint | no |  | check(>0) | durable monotonic version |
| updated_by_user_id | uuid | yes |  | fk -> users | nullable for seeded/default values |
| updated_by_email | text | yes |  |  | actor snapshot for audit/UI |
| updated_at | timestamptz | no | now() |  | last durable update time |

### Entity: system_setting_changes
- Назначение: durable audit/versioning ledger для platform settings.
- Важные инварианты:
  - `(setting_key, version)` уникальны;
  - `previous_value_json` nullable только для seed baseline;
  - change log остаётся durable source для reconnect/catch-up, `LISTEN/NOTIFY` используется только как wake-up signal.
- Поля:

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| id | bigserial | no |  | pk | |
| setting_key | text | no |  | fk -> system_settings.key | |
| value_kind | text | no |  | check(boolean) | |
| value_json | jsonb | no |  |  | value after change |
| previous_value_json | jsonb | yes |  |  | nullable for initial seed |
| source | text | no |  | check(default/staff) | |
| version | bigint | no |  | check(>0), unique(setting_key, version) | |
| change_kind | text | no |  | check(seeded/updated/reset) | |
| actor_user_id | uuid | yes |  | fk -> users | |
| actor_email | text | yes |  |  | |
| created_at | timestamptz | no | now() |  | |

### Entity: repositories
- Назначение: подключённые репозитории проектов.
- Важные инварианты:
  - `project_id + alias` уникальны (стабильный repo-key внутри проекта);
  - `project_id + provider + owner + name` уникальны (одна реальная интеграция на проект).
- Поля:

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| id | uuid | no | gen_random_uuid() | pk | |
| project_id | uuid | no |  | fk -> projects | |
| alias | text | no |  | unique(project_id, alias) | stable key for multi-repo imports/docs refs |
| provider | text | no |  | check(github/gitlab) | |
| owner | text | no |  |  | |
| name | text | no |  |  | |
| role | text | no | "service" | check(orchestrator/service/docs/mixed) | topology role in project composition |
| default_ref | text | no | "main" |  | default branch/tag for resolve |
| token_encrypted | bytea | no |  |  | app-level encrypted |
| services_yaml_path | text | no | "services.yaml" |  | per-repo override |
| docs_root_path | text | yes |  |  | optional default docs root for repo-aware docs context |

Примечание по token scope (S2 Day4+):
- `repositories.token_encrypted` используется только для операций управления проектом/репозиторием
  (validate repository, ensure/delete webhook и т.п. staff management path).
- Runtime сообщения и label-операции в run/mcp контуре используют bot-token из singleton сущности `platform_github_tokens`.

### Entity: repository_compositions
- Назначение: фиксировать результат runtime-компоновки multi-repo manifest до reconcile/deploy.
- Важные инварианты:
  - одна активная компоновка на `(project_id, environment, composition_key)`;
  - `status=ready` допускается только при валидном `effective_manifest_json`.
- Поля:

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| id | bigserial | no |  | pk | |
| project_id | uuid | no |  | fk -> projects | |
| environment | text | no |  |  | e.g. `dev`, `production`, `ai-slot` |
| composition_key | text | no |  | unique(project_id, environment, composition_key) | deterministic resolver key |
| root_repository_alias | text | yes |  | fk -> repositories.alias scoped by project | null for virtual-root mode |
| resolved_repositories_json | jsonb | no | '[]'::jsonb |  | aliases + pinned refs/commits |
| effective_manifest_json | jsonb | no | '{}'::jsonb |  | final typed compose payload |
| status | text | no | "pending" | check(pending/ready/failed) | resolver state |
| error_message | text | yes |  |  | last resolver error |
| created_at | timestamptz | no | now() |  | |
| updated_at | timestamptz | no | now() |  | |

### Entity: repository_doc_sources
- Назначение: role-aware источники docs из разных repo для prompt context и doc governance.
- Важные инварианты:
  - `project_id + repository_alias + path` уникальны;
  - path всегда относительный к repo root.
- Поля:

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| id | bigserial | no |  | pk | |
| project_id | uuid | no |  | fk -> projects | |
| repository_alias | text | no |  | fk -> repositories.alias scoped by project | repo binding key |
| path | text | no |  |  | docs root path in repo |
| description | text | yes |  |  | human-readable docs source purpose |
| roles_json | jsonb | no | '[]'::jsonb |  | allowed roles for docs context |
| optional | bool | no | false |  | resolver may skip if source missing |
| priority | int | no | 100 |  | lower value = higher priority |
| created_at | timestamptz | no | now() |  | |
| updated_at | timestamptz | no | now() |  | |

### Entity: project_databases
- Назначение: ownership registry для MCP tool `database.lifecycle`.
- Важные инварианты: одна БД принадлежит только одному проекту; delete/describe разрешены только владельцу.
- Поля:

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| project_id | uuid | no |  | fk -> projects | ownership project |
| environment | text | no |  | check not empty | env scope (`dev/production/prod/...`) |
| database_name | text | no |  | pk | global DB identifier |
| created_at | timestamptz | no | now() |  | |
| updated_at | timestamptz | no | now() |  | |

### Entity: platform_github_tokens
- Назначение: singleton-хранилище платформенных GitHub токенов.
- Важные инварианты: в таблице всегда максимум одна запись (`id=1`).
- Поля:

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| id | smallint | no |  | pk, check(id=1) | singleton row |
| platform_token_encrypted | bytea | yes |  |  | platform token (wide scope, management paths) |
| bot_token_encrypted | bytea | yes |  |  | bot token (run/messaging/labels paths) |
| created_at | timestamptz | no | now() |  | |
| updated_at | timestamptz | no | now() |  | |

### Entity: agents
- Назначение: реестр системных агентных профилей и возможных project-scoped overrides.
- Важные инварианты: уникальный `agent_key` для system-profile; project-scoped запись с тем же `agent_key` имеет приоритет при резолве effective agent.
- Поля:

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| id | uuid | no | gen_random_uuid() | pk | |
| agent_key | text | no |  | unique | pm/sa/em/dev/reviewer/qa/sre/km или custom key |
| role_kind | text | no | "system" | check(system/custom) | |
| project_id | uuid | yes |  | fk -> projects | not null for role_kind=custom |
| name | text | no |  |  | |
| is_active | bool | no | true |  | inactive записи исключаются из effective resolve |
| created_at | timestamptz | no | now() |  | |
| updated_at | timestamptz | no | now() |  | |

Примечание:
- В S7 cleanup из `agents` удалены non-MVP поля `settings` и `settings_version`.
- Runtime mode, locale и prompt policy больше не хранятся как user-editable agent settings в БД.
- Эти параметры определяются platform defaults, label policy, `services.yaml` и repo seeds.

### Entity: agent_runs
- Назначение: запуски и сессии агентов.
- Важные инварианты: уникальный correlation_id.
- Поля:

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| id | uuid | no | gen_random_uuid() | pk | |
| correlation_id | text | no |  | unique | webhook/job correlation |
| project_id | uuid | no |  | fk -> projects | |
| agent_id | uuid | no |  | fk -> agents | |
| status | text | no | "pending" | check enum | pending/running/waiting_owner_review/waiting_mcp/succeeded/failed/timed_out/cancelled |
| run_payload | jsonb | no | '{}'::jsonb |  | session metadata/log refs |
| agent_logs_json | jsonb | yes |  |  | persisted agent execution logs snapshot for staff observability |
| learning_mode | bool | no | false |  | run-level effective mode |
| timeout_at | timestamptz | yes |  |  | hard timeout deadline |
| timeout_paused | bool | no | false |  | true while paused on allowed waits |
| wait_reason | text | yes |  |  | owner_review/mcp/none |
| lease_owner | text | yes |  |  | worker instance currently owning running-run reconciliation; during mixed-version rollout missing `worker_instances` row is cross-checked against live worker pods |
| lease_until | timestamptz | yes |  |  | reconciliation lease expiration |
| stale_reclaim_pending | bool | no | false |  | set when stale worker lease was released and next claim must emit recovery event |
| started_at | timestamptz | yes |  |  | |
| finished_at | timestamptz | yes |  |  | |

### Entity: worker_instances
- Назначение: liveness-модель worker pod/instance для быстрого recovery stale running leases.
- Важные инварианты: один `worker_id` описывает одну активную worker-instance запись; heartbeat обновляет `expires_at`, graceful shutdown переводит `status=stopped`, а mixed-version rollout использует Kubernetes live worker pod set как fallback для owner без строки в `worker_instances`.
- Поля:

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| worker_id | text | no |  | pk | logical worker instance id (обычно pod hostname/name) |
| namespace | text | no | '' |  | worker pod namespace |
| pod_name | text | no | '' |  | worker pod name for runtime diagnostics |
| status | text | no | 'active' | check enum | active/stopped |
| started_at | timestamptz | no | now() |  | worker process start time |
| heartbeat_at | timestamptz | no | now() |  | latest successful heartbeat |
| expires_at | timestamptz | no | now() | index | stale threshold for lease recovery |
| created_at | timestamptz | no | now() |  | |
| updated_at | timestamptz | no | now() |  | |

### Entity: agent_sessions
- Назначение: детальная телеметрия и аудит выполнения агентной сессии.
- Важные инварианты (текущий baseline Day4): одна запись на `run_id` (unique), сессия связана с run.
- Поля:

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| id | bigserial | no |  | pk | |
| run_id | uuid | no |  | fk -> agent_runs | |
| correlation_id | text | no |  |  | |
| project_id | uuid | yes |  | fk -> projects | |
| repository_full_name | text | no |  |  | |
| issue_number | int | yes |  |  | |
| branch_name | text | yes |  |  | |
| pr_number | int | yes |  |  | |
| pr_url | text | yes |  |  | |
| trigger_kind | text | yes |  |  | |
| template_kind | text | yes |  |  | |
| template_source | text | yes |  |  | |
| template_locale | text | yes |  |  | |
| model | text | yes |  |  | |
| reasoning_effort | text | yes |  |  | |
| status | text | no | "running" | check enum | running/succeeded/failed/cancelled/failed_precondition |
| session_id | text | yes |  |  | external/model session id |
| session_json | jsonb | no | '{}'::jsonb |  | run execution snapshot (report + condensed runtime logs) |
| codex_cli_session_path | text | yes |  |  | path to saved session file in workspace/storage |
| codex_cli_session_json | jsonb | yes |  |  | persisted codex-cli session snapshot for resume |
| snapshot_version | bigint | no | 1 |  | CAS-like snapshot rewrite version |
| snapshot_checksum | text | yes |  |  | sha256 canonical checksum of persisted snapshot payload |
| snapshot_updated_at | timestamptz | no | now() |  | latest successful snapshot rewrite timestamp |
| wait_state | text | yes |  | check(owner_review/mcp) | current wait-state for timeout governance |
| timeout_guard_disabled | bool | no | false |  | `true` while timeout-kill must stay paused |
| last_heartbeat_at | timestamptz | yes |  |  | heartbeat for wait-state/recovery |
| started_at | timestamptz | no | now() |  | |
| finished_at | timestamptz | yes |  |  | |
| created_at | timestamptz | no | now() |  | |
| updated_at | timestamptz | no | now() |  | |

Реализовано в S2 Day6:
- wait-state/time-guard поля добавлены и используются в approval lifecycle (`wait_state`, `timeout_guard_disabled`, `last_heartbeat_at`);
- pause/resume ожидания MCP синхронизируется через `agent_sessions` + `flow_events`.

### Entity: token_usage
- Назначение: учёт токенов/стоимости по сессиям и моделям.
- Важные инварианты: запись append-only.
- Поля:

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| id | bigserial | no |  | pk | |
| session_id | text | no |  | fk/logical -> agent_sessions.session_id | |
| model | text | no |  |  | |
| prompt_tokens | int | no | 0 |  | |
| completion_tokens | int | no | 0 |  | |
| total_tokens | int | no | 0 |  | |
| cost_usd | numeric(18,6) | yes |  |  | optional |
| created_at | timestamptz | no | now() | index | |

### Entity: slots
- Назначение: слоты и их lease-состояние для конкурентных pod.
- Важные инварианты: один активный lease на слот.
- Поля:

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| id | uuid | no | gen_random_uuid() | pk | |
| project_id | uuid | no |  | fk -> projects | |
| slot_no | int | no |  | unique(project_id, slot_no) | |
| state | text | no | "free" | check enum | free/leased/releasing |
| lease_owner | text | yes |  |  | pod/run id |
| lease_until | timestamptz | yes |  |  | |

### Entity: runtime_deploy_tasks
- Назначение: persisted desired/actual state для декларативного deploy-контура (`services.yaml`) с идемпотентным reconciler execution.
- Важные инварианты: один deploy task на один `run_id`; lease-механизм предотвращает двойную параллельную обработку.
- Поля:

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| run_id | uuid | no |  | pk, fk -> agent_runs(id) | one task per run |
| runtime_mode | text | no | '' |  | requested runtime mode |
| namespace | text | no | '' |  | desired namespace override |
| target_env | text | no | '' |  | requested target env |
| slot_no | int | no | 0 |  | slot index from run payload |
| repository_full_name | text | no | '' |  | owner/repo |
| services_yaml_path | text | no | '' |  | path hint from payload |
| build_ref | text | no | '' |  | commit/branch ref for build |
| deploy_only | bool | no | false |  | deploy-only run flag |
| status | text | no | 'pending' | check enum | pending/running/succeeded/failed/canceled |
| lease_owner | text | yes |  |  | reconciler instance id |
| lease_until | timestamptz | yes |  |  | lease expiration |
| attempts | int | no | 0 |  | reconcile attempts |
| last_error | text | yes |  |  | last terminal failure details |
| result_namespace | text | yes |  |  | effective namespace after render |
| result_target_env | text | yes |  |  | effective env after render |
| created_at | timestamptz | no | now() |  | |
| updated_at | timestamptz | no | now() |  | |
| started_at | timestamptz | yes |  |  | first claim timestamp |
| finished_at | timestamptz | yes |  |  | terminal timestamp |

### Entity: flow_events
- Назначение: аудит системных/агентных/человеческих действий.
- Важные инварианты: append-only.
- Поля:

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| id | bigserial | no |  | pk | |
| correlation_id | text | no |  | index | |
| actor_type | text | no |  | check enum | human/agent/system |
| actor_id | text | yes |  |  | |
| event_type | text | no |  | index | |
| payload | jsonb | no | '{}'::jsonb |  | includes approval/executor callbacks and label/runtime action metadata |
| created_at | timestamptz | no | now() | index | |

### Entity: links
- Назначение: трассировка связей между Issue/PR/run/doc/ADR.
- Важные инварианты: уникальность пары source-target по типу связи.
- Поля:

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| id | bigserial | no |  | pk | |
| source_type | text | no |  |  | issue/pr/run/doc/adr |
| source_id | text | no |  |  | |
| target_type | text | no |  |  | issue/pr/run/doc/adr |
| target_id | text | no |  |  | |
| link_type | text | no |  |  | references/implements/supersedes |
| metadata | jsonb | no | '{}'::jsonb |  | |
| created_at | timestamptz | no | now() | index | |

### Entity: docs_meta
- Назначение: шаблоны и документы платформы.
- Важные инварианты: уникальный doc_id.
- Поля:

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| id | uuid | no | gen_random_uuid() | pk | |
| doc_id | text | no |  | unique | |
| title | text | no |  |  | |
| type | text | no |  |  | |
| status | text | no | "draft" |  | |
| body_markdown | text | no |  |  | |
| meta | jsonb | no | '{}'::jsonb |  | frontmatter mirror |

### Entity: learning_feedback
- Назначение: хранение образовательных объяснений для выполненных задач (inline и post-PR).
- Важные инварианты: каждая запись связана с конкретным run/file/опционально line.
- Поля:

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| id | bigserial | no |  | pk | |
| run_id | uuid | no |  | fk -> agent_runs | |
| repository_id | uuid | yes |  | fk -> repositories | |
| pr_number | int | yes |  |  | |
| file_path | text | yes |  |  | |
| line | int | yes |  |  | optional line-level note |
| kind | text | no |  | check(inline,post_pr) | |
| explanation | text | no |  |  | why/tradeoffs/better patterns |
| created_at | timestamptz | no | now() |  | |

### Entity: mcp_action_requests
- Назначение: журнал запросов к MCP control tools и их approval lifecycle.
- Важные инварианты: один action request имеет единственный current status, переходы append-only в `flow_events`.
- Поля:

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| id | bigserial | no |  | pk | |
| correlation_id | text | no |  | index | |
| run_id | uuid | yes |  | fk -> agent_runs | |
| tool_name | text | no |  |  | e.g. `secret.sync.k8s` |
| action | text | no |  |  | create/update/delete/request |
| target_ref | jsonb | no | '{}'::jsonb |  | project/repo/env refs + policy/idempotency_key |
| approval_mode | text | no | "owner" | check enum | none/owner/delegated |
| approval_state | text | no | "requested" | check enum | requested/approved/denied/expired/failed/applied |
| requested_by | text | no |  |  | actor id |
| applied_by | text | yes |  |  | actor id |
| payload | jsonb | no | '{}'::jsonb |  | masked request/result metadata (для secret sync хранится encrypted value) |
| created_at | timestamptz | no | now() | index | |
| updated_at | timestamptz | no | now() |  | |

### Entity: doc_chunks
- Назначение: чанки документов для поиска.
- Важные инварианты: уникальный (doc_id, chunk_no).
- Поля:

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| id | bigserial | no |  | pk | |
| doc_id | text | no |  | fk/logical to docs_meta.doc_id | |
| chunk_no | int | no |  | unique(doc_id, chunk_no) | |
| chunk_text | text | no |  |  | |
| embedding | vector(3072) | yes |  | ivfflat/hnsw index | pgvector |
| metadata | jsonb | no | '{}'::jsonb |  | headings, links |

## Связи
- `system_settings` хранит глобальные platform-wide настройки
- `system_settings` 1:N `system_setting_changes`
- `projects` 1:N `repositories`
- `projects` M:N `users` через `project_members`
- `agents` 1:N `agent_runs`
- `worker_instances` 1:N `agent_runs` по `agent_runs.lease_owner -> worker_instances.worker_id` (soft relation для liveness/recovery)
- `projects` 1:N `agents` (только для project-scoped overrides/custom profiles, если они присутствуют)
- `agent_runs` 1:1 `agent_sessions` (текущий baseline Day4 по `run_id unique`; может эволюционировать до 1:N при multi-session run)
- `agent_sessions` 1:N `token_usage`
- `projects` 1:N `slots`
- `docs_meta` 1:N `doc_chunks`
- `agent_runs` 1:N `flow_events` (по `correlation_id`)
- `agent_runs` 1:N `learning_feedback`
- `agent_runs` 1:N `mcp_action_requests`
- `links` хранит M:N трассировки между `issue/pr/run/doc/adr`

## Логическое размещение по БД-контурам (MVP)
- PostgreSQL cluster единый.
- Core contour: `users`, `projects`, `project_members`, `system_settings`, `system_setting_changes`, `repositories`, `agents`, `agent_runs`, `worker_instances`, `slots`, `runtime_deploy_tasks`, `docs_meta`, `learning_feedback`.
- Audit/chunks contour: `agent_sessions`, `token_usage`, `flow_events`, `links`, `doc_chunks`, `mcp_action_requests`.
- Связи между контурами — через устойчивые ключи (`correlation_id`, `doc_id`), без требования к cross-contour FK.

## Индексы и запросы (критичные)
- Запрос: выбрать ожидающие webhook jobs по статусу/времени.
- Индексы: `agent_runs(status, started_at)`, `agent_runs(status, lease_until, started_at)`, `agent_runs(status, lease_owner, lease_until, started_at)`, `worker_instances(status, expires_at)`, `flow_events(correlation_id, created_at)`.
- Запрос: аудит сессий и стоимости по run/agent/model.
- Индексы: `agent_sessions(run_id, started_at)`, `token_usage(session_id, created_at)`.
- Запрос: найти pending/failed MCP action requests.
- Индексы: `mcp_action_requests(approval_state, created_at)`, `mcp_action_requests(correlation_id)`.
- Запрос: возобновление прерванной/ожидающей сессии по run.
- Индексы: `agent_sessions(run_id, wait_state, last_heartbeat_at)`.
- Запрос: traceability issue/pr/run/doc.
- Индексы: `links(source_type, source_id, created_at)`, `links(target_type, target_id, created_at)`.
- Запрос: поиск релевантных doc chunks.
- Индексы: `doc_chunks using ivfflat/hnsw (embedding)`, плюс `metadata` GIN.

## Политика хранения данных
- Retention: flow_events, agent_sessions.session_json, agent_sessions.codex_cli_session_json и token_usage с ротацией/архивом по сроку.
- `agent_runs.agent_logs_json` очищается периодическим cleanup loop в `control-plane` для завершённых run старше `KODEX_RUN_AGENT_LOGS_RETENTION_DAYS` (default: `14`).
- Архивирование: ежедневный backup БД в production.
- PII/комплаенс: email хранится, токены только в шифрованном виде.

Roadmap (Day5+):
- добавить live-stream канал логов (SSE/WebSocket) и отдельную staff UI вьюшку run-деталей с обновлением действий агента в реальном времени;
- после включения стриминга оставить `agent_runs.agent_logs_json` как fallback snapshot для пост-фактум аудита.

## Миграции (ссылка)
См. `migrations_policy.md` (будет добавлен на этапе design) + миграции в держателе схемы:
`services/<zone>/<db-owner-service>/cmd/cli/migrations`.

## Решения Owner
- Размер вектора `3072` подтверждён как базовый для MVP.
- Отдельный `event_outbox` на MVP не вводится; используем статусы `agent_runs` + `flow_events`.
- Контур аудита и учета обязателен: `agent_sessions`, `token_usage`, `links`.
- Шаблоны промптов в текущем MVP поддерживают repo-only seed model; effective source фиксируется в `agent_sessions.template_source`.
- Для paused-состояний сохраняется `codex-cli` session snapshot, чтобы run можно было продолжить с того же места.
- Для rewrite-safe snapshot persistence используются `snapshot_version`, `snapshot_checksum`, `snapshot_updated_at`;
  repeated replay с тем же payload остаётся идемпотентным, stale rewrite не должен затирать более новую версию.
- При ожидании ответа MCP (`wait_state=mcp`) timeout-kill для pod/run не применяется до завершения ожидания.

## Апрув
- request_id: owner-2026-02-06-mvp
- Решение: approved
- Комментарий: Модель данных MVP зафиксирована.
