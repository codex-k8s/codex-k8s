---
doc_id: DM-CK8S-0001
type: data-model
title: "codex-k8s — Data Model"
status: active
owner_role: SA
created_at: 2026-02-06
updated_at: 2026-02-14
related_issues: [1, 19]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-19-full-docset"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-19
---

# Data Model: codex-k8s

## TL;DR
- Ключевые сущности: users, projects, system_settings, repositories, project_databases, agent_policies, agents, agent_runs, agent_sessions, token_usage, slots, flow_events, links, prompt_templates, docs_meta, doc_chunks.
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
| settings | jsonb | no | '{}'::jsonb |  | user-configurable (`learning_mode_default`, `locale`, etc.) |

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
- Назначение: глобальные настройки платформы (включая default locale).
- Важные инварианты: ключ уникален.
- Поля:

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| key | text | no |  | pk | |
| value | jsonb | no | '{}'::jsonb |  | e.g. `{"default_locale":"ru"}` |
| updated_at | timestamptz | no | now() |  | |

### Entity: repositories
- Назначение: подключённые репозитории проектов.
- Важные инварианты: provider + external_id уникальны.
- Поля:

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| id | uuid | no | gen_random_uuid() | pk | |
| project_id | uuid | no |  | fk -> projects | |
| provider | text | no |  | check(github/gitlab) | |
| owner | text | no |  |  | |
| name | text | no |  |  | |
| token_encrypted | bytea | no |  |  | app-level encrypted |
| services_yaml_path | text | no | "services.yaml" |  | per-repo override |

Примечание по token scope (S2 Day4+):
- `repositories.token_encrypted` используется только для операций управления проектом/репозиторием
  (validate repository, ensure/delete webhook и т.п. staff management path).
- Runtime сообщения и label-операции в run/mcp контуре используют bot-token из singleton сущности `platform_github_tokens`.

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
- Назначение: системные и project-scoped custom-агенты.
- Важные инварианты: уникальный agent_key; для custom-агента обязательна привязка к проекту.
- Поля:

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| id | uuid | no | gen_random_uuid() | pk | |
| agent_key | text | no |  | unique | pm/sa/em/dev/reviewer/qa/sre/km или custom key |
| role_kind | text | no | "system" | check(system/custom) | |
| policy_id | bigint | yes |  | fk -> agent_policies | null allowed only until policy bootstrap |
| project_id | uuid | yes |  | fk -> projects | not null for role_kind=custom |
| name | text | no |  |  | |
| github_nick | text | yes |  |  | |
| email | text | yes |  |  | |
| token_encrypted | bytea | yes |  |  | rotated by worker |
| instruction_template | text | no |  |  | fallback markdown |
| review_template | text | yes |  |  | optional review fallback |
| settings | jsonb | no | '{}'::jsonb |  | runtime/options |
| policies | jsonb | no | '{}'::jsonb |  | execution/timeout/approval policy snapshot |

### Entity: agent_policies
- Назначение: централизованные policy-профили работы агентов.
- Важные инварианты: `agents` всегда ссылается на policy (явно или через default).
- Поля:

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| id | bigserial | no |  | pk | |
| policy_key | text | no |  | unique | e.g. `default-dev`, `default-review` |
| runtime_mode | text | no |  | check(full-env/code-only) | |
| max_parallel_runs | int | no | 1 |  | per project baseline |
| run_timeout_sec | int | no | 0 |  | 0 = no hard timeout |
| owner_review_timeout_sec | int | yes |  |  | pause/resume aware |
| kill_on_mcp_wait_timeout | bool | no | false |  | must stay false |
| approval_required_for_run_labels | bool | no | true |  | |
| config | jsonb | no | '{}'::jsonb |  | extensible policy fields, включая MCP tool/resource matrix и label-based overrides |
| created_at | timestamptz | no | now() |  | |

Planned extension (Day6+):
- policy snapshot хранит отдельные блоки:
  - `mcp.default_capabilities` (базовый набор ручек/ресурсов по агенту);
  - `mcp.label_overrides` (дополнительные ограничения/разрешения по типу задачи, например `run:dev`, `run:dev:revise`);
  - `mcp.composite_tools` (профили для комбинированных ручек вида GitHub+Kubernetes).

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
| lease_owner | text | yes |  |  | worker instance currently owning running-run reconciliation |
| lease_until | timestamptz | yes |  |  | reconciliation lease expiration |
| started_at | timestamptz | yes |  |  | |
| finished_at | timestamptz | yes |  |  | |

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

### Entity: prompt_templates
- Назначение: хранение global/project override шаблонов промптов (`work`/`revise`).
- Важные инварианты: уникальность active версии по `(scope_type, scope_id, role_key, template_kind, locale)`.
- Поля:

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| id | bigserial | no |  | pk | |
| scope_type | text | no |  | check(global/project) | |
| scope_id | uuid | yes |  | fk -> projects | null for global |
| role_key | text | no |  |  | pm/sa/em/dev/reviewer/qa/sre/km/custom |
| template_kind | text | no |  | check(work/revise) | |
| locale | text | no | "en" |  | i18n locale key (`ru`, `en`, ...) |
| body_markdown | text | no |  |  | |
| source | text | no | "db_override" |  | db_override/repo_seed_ref |
| render_context_version | text | no | "v1" |  | contract version for template context |
| version | int | no | 1 |  | |
| is_active | bool | no | true |  | |
| created_at | timestamptz | no | now() |  | |

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
| tool_name | text | no |  |  | e.g. `secret.sync.github_k8s` |
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
- `system_settings` задаёт глобальные fallback policy (включая default locale)
- `projects` 1:N `repositories`
- `projects` M:N `users` через `project_members`
- `agent_policies` 1:N `agents`
- `agents` 1:N `agent_runs`
- `projects` 1:N `agents` (для custom-агентов)
- `agent_runs` 1:1 `agent_sessions` (текущий baseline Day4 по `run_id unique`; может эволюционировать до 1:N при multi-session run)
- `agent_sessions` 1:N `token_usage`
- `projects` 1:N `slots`
- `docs_meta` 1:N `doc_chunks`
- `agent_runs` 1:N `flow_events` (по `correlation_id`)
- `agent_runs` 1:N `learning_feedback`
- `agent_runs` 1:N `mcp_action_requests`
- `projects` 1:N `prompt_templates` (scope=project)
- `links` хранит M:N трассировки между `issue/pr/run/doc/adr`

## Логическое размещение по БД-контурам (MVP)
- PostgreSQL cluster единый.
- Core contour: `users`, `projects`, `project_members`, `system_settings`, `repositories`, `agent_policies`, `agents`, `agent_runs`, `slots`, `docs_meta`, `learning_feedback`, `prompt_templates`.
- Audit/chunks contour: `agent_sessions`, `token_usage`, `flow_events`, `links`, `doc_chunks`, `mcp_action_requests`.
- Связи между контурами — через устойчивые ключи (`correlation_id`, `doc_id`), без требования к cross-contour FK.

## Индексы и запросы (критичные)
- Запрос: выбрать ожидающие webhook jobs по статусу/времени.
- Индексы: `agent_runs(status, started_at)`, `agent_runs(status, lease_until, started_at)`, `flow_events(correlation_id, created_at)`.
- Запрос: аудит сессий и стоимости по run/agent/model.
- Индексы: `agent_sessions(run_id, started_at)`, `token_usage(session_id, created_at)`.
- Запрос: найти pending/failed MCP action requests.
- Индексы: `mcp_action_requests(approval_state, created_at)`, `mcp_action_requests(correlation_id)`.
- Запрос: возобновление прерванной/ожидающей сессии по run.
- Индексы: `agent_sessions(run_id, wait_state, last_heartbeat_at)`.
- Запрос: выбор effective prompt template по role/kind/locale.
- Индексы: `prompt_templates(scope_type, scope_id, role_key, template_kind, locale, is_active)`.
- Запрос: traceability issue/pr/run/doc.
- Индексы: `links(source_type, source_id, created_at)`, `links(target_type, target_id, created_at)`.
- Запрос: поиск релевантных doc chunks.
- Индексы: `doc_chunks using ivfflat/hnsw (embedding)`, плюс `metadata` GIN.

## Политика хранения данных
- Retention: flow_events, agent_sessions.session_json, agent_sessions.codex_cli_session_json и token_usage с ротацией/архивом по сроку.
- `agent_runs.agent_logs_json` очищается периодическим cleanup loop в `control-plane` для завершённых run старше `CODEXK8S_RUN_AGENT_LOGS_RETENTION_DAYS` (default: `14`).
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
- Шаблоны промптов поддерживают модель `repo seed + DB override` c фиксацией effective version/hash.
- Для paused-состояний сохраняется `codex-cli` session snapshot, чтобы run можно было продолжить с того же места.
- При ожидании ответа MCP (`wait_state=mcp`) timeout-kill для pod/run не применяется до завершения ожидания.

## Апрув
- request_id: owner-2026-02-06-mvp
- Решение: approved
- Комментарий: Модель данных MVP зафиксирована.
