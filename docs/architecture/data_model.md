---
doc_id: DM-CK8S-0001
type: data-model
title: "codex-k8s — Data Model"
status: draft
owner_role: SA
created_at: 2026-02-06
updated_at: 2026-02-06
related_issues: [1]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-06-mvp"
---

# Data Model: codex-k8s

## TL;DR
- Ключевые сущности: users, projects, repositories, agents, agent_runs, slots, flow_events, docs_meta, doc_chunks.
- Основные связи: user<->project (RBAC), project->repositories, agent->agent_runs, issue/pr->doc links.
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
| settings | jsonb | no | '{}'::jsonb |  | user-configurable (`learning_mode_default`, etc.) |

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

### Entity: agents
- Назначение: штатные агенты (фиксированный набор ролей).
- Важные инварианты: уникальный agent_key.
- Поля:

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| id | uuid | no | gen_random_uuid() | pk | |
| agent_key | text | no |  | unique | pm/sa/em/qa/sre/km/auditor |
| name | text | no |  |  | |
| github_nick | text | yes |  |  | |
| email | text | yes |  |  | |
| token_encrypted | bytea | yes |  |  | rotated by worker |
| instruction_template | text | no |  |  | markdown |
| settings | jsonb | no | '{}'::jsonb |  | |

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
| status | text | no | "pending" | check enum | |
| run_payload | jsonb | no | '{}'::jsonb |  | session metadata/log refs |
| learning_mode | bool | no | false |  | run-level effective mode |
| started_at | timestamptz | yes |  |  | |
| finished_at | timestamptz | yes |  |  | |

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
| payload | jsonb | no | '{}'::jsonb |  | |
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
- `projects` 1:N `repositories`
- `projects` M:N `users` через `project_members`
- `agents` 1:N `agent_runs`
- `projects` 1:N `slots`
- `docs_meta` 1:N `doc_chunks`
- `agent_runs` 1:N `flow_events` (по `correlation_id`)
- `agent_runs` 1:N `learning_feedback`

## Логическое размещение по БД-контурам (MVP)
- PostgreSQL cluster единый.
- Core contour: `users`, `projects`, `project_members`, `repositories`, `agents`, `agent_runs`, `slots`, `docs_meta`, `learning_feedback`.
- Audit/chunks contour: `flow_events`, `doc_chunks`.
- Связи между контурами — через устойчивые ключи (`correlation_id`, `doc_id`), без требования к cross-contour FK.

## Индексы и запросы (критичные)
- Запрос: выбрать ожидающие webhook jobs по статусу/времени.
- Индексы: `agent_runs(status, started_at)`, `flow_events(correlation_id, created_at)`.
- Запрос: поиск релевантных doc chunks.
- Индексы: `doc_chunks using ivfflat/hnsw (embedding)`, плюс `metadata` GIN.

## Политика хранения данных
- Retention: flow_events и session JSON с ротацией/архивом по сроку.
- Архивирование: ежедневный backup БД в staging.
- PII/комплаенс: email хранится, токены только в шифрованном виде.

## Миграции (ссылка)
См. `migrations_policy.md` (будет добавлен на этапе design) + список миграций в `cmd/cli/migrations`.

## Решения Owner
- Размер вектора `3072` подтверждён как базовый для MVP.
- Отдельный `event_outbox` на MVP не вводится; используем статусы `agent_runs` + `flow_events`.

## Апрув
- request_id: owner-2026-02-06-mvp
- Решение: approved
- Комментарий: Модель данных MVP зафиксирована.
