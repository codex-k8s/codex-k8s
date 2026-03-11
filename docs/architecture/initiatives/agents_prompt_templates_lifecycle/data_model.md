---
doc_id: DM-APT-CK8S-0001
type: data-model
title: "codex-k8s — Data model: agents settings and prompt templates lifecycle"
status: in-review
owner_role: SA
created_at: 2026-02-25
updated_at: 2026-02-25
related_issues: [184, 185, 187, 189, 195, 197]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-25-issue-195-data-model"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-25
---

# Data Model: agents settings and prompt templates lifecycle

## TL;DR
- Основной owner схемы остаётся `services/internal/control-plane`.
- Базовые сущности: `agents`, `agent_policies`, `prompt_templates`, `flow_events`.
- Вариант ADR-0009 сохраняется: история версий внутри `prompt_templates` + audit в `flow_events`.
- Миграционный риск сосредоточен в инварианте «одна active версия на template key».

## Сущности
### Entity: `agents` (reuse + minor settings hardening)
- Назначение: хранение runtime/policy ссылок агента.
- Важные инварианты:
  - агент связан с policy;
  - edit правки разрешены только project admin.
- Модель поля (design scope): без обязательного расширения схемы на этапе S6 Day5; изменения касаются transport-level typed settings DTO.

### Entity: `prompt_templates` (target extension)
- Назначение: хранение lifecycle версий prompt templates (`work`/`revise`, locale-aware).
- Важные инварианты:
  - уникальный `version` внутри template key;
  - только одна `active` версия на template key;
  - любые write изменения фиксируются audit-событием.

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| id | bigserial | no |  | pk | existing |
| scope_type | text | no |  | check(global/project) | existing |
| scope_id | uuid | yes |  | fk -> projects | existing |
| role_key | text | no |  |  | existing |
| template_kind | text | no |  | check(work/revise) | existing |
| locale | text | no | `en` |  | existing |
| body_markdown | text | no |  |  | existing |
| source | text | no | `db_override` | check(project_override/global_override/repo_seed) | existing; `repo_seed` используется для baseline bootstrap и fallback-трассировки |
| render_context_version | text | no | `v1` |  | existing |
| version | int | no | 1 |  | existing |
| is_active | bool | no | true |  | existing, remains for compatibility |
| status | text | no | `draft` | check(draft/active/archived) | new |
| checksum | text | no |  |  | new, sha256(body_markdown) |
| change_reason | text | yes |  |  | new, required for activate/archive |
| supersedes_version | int | yes |  |  | new, logical link to previous version |
| updated_by | text | no |  |  | new, actor id |
| updated_at | timestamptz | no | now() |  | new |
| activated_at | timestamptz | yes |  |  | new |
| metadata | jsonb | no | '{}'::jsonb |  | new, optional UI/runtime hints |
| created_at | timestamptz | no | now() |  | existing |

### Entity: `flow_events` (reuse, contract hardening)
- Назначение: append-only audit событий lifecycle.
- Важные инварианты:
  - событие для template write сохраняется в той же DB-транзакции,
  - `payload.template_key`, `payload.version`, `payload.status` обязательны для template events.

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| id | bigserial | no |  | pk | existing |
| correlation_id | text | no |  | index | existing |
| actor_type | text | no |  | check enum | existing |
| actor_id | text | yes |  |  | existing |
| event_type | text | no |  | index | existing |
| payload | jsonb | no | '{}'::jsonb |  | existing, typed payload schema by event_type |
| created_at | timestamptz | no | now() | index | existing |

## Связи
- `projects` 1:N `prompt_templates` (project scope).
- `agent_runs` 1:N `flow_events` по `correlation_id`.
- `prompt_templates` -> `flow_events` логическая связь через `payload.template_key`/`version`.

## Индексы и запросы (критичные)
- Query: active version by template key.
  - Index: partial unique on `(scope_type, scope_id, role_key, template_kind, locale)` where `status='active'`.
- Query: list versions by key sorted desc.
  - Index: `(scope_type, scope_id, role_key, template_kind, locale, version desc)`.
- Query: conflict check for writes.
  - Index: `(scope_type, scope_id, role_key, template_kind, locale, version)` unique.
- Query: audit list for template actions.
  - Index: `flow_events(event_type, created_at desc)` + GIN index on `payload` for template filters.

## Политика хранения данных
- `prompt_templates`: retain full history; archive via `status='archived'`.
- `flow_events`: retention/archival policy из базовой модели данных, без silent delete.
- PII/secret policy: в `body_markdown` запрещены secret-like значения; violation -> validation error.

## Инварианты домена
- `version` монотонно увеличивается в рамках template key.
- `activate` разрешён только для существующей `draft|archived` версии.
- `is_active` и `status` должны быть консистентны (`status='active' => is_active=true`).
- `flow_events` write обязателен для каждой mutating операции.
- При отсутствии DB-записей по ключу effective template берётся из embed seed; это не нарушает инварианты БД и не требует synthetic rows.

## Runtime impact / Migration impact
- Runtime impact (`run:design`): отсутствует, документы-only.
- Migration impact (`run:dev`): DDL расширения `prompt_templates`, backfill existing rows, индексы и post-check инвариантов.

## Context7 dependency check
- Подтверждено, что текущий стек (`kin-openapi` + `monaco-editor`) покрывает контракты и UI diff use-case.
- Новые внешние зависимости для data-model части не требуются.

## Апрув
- request_id: owner-2026-02-25-issue-195-data-model
- Решение: approved
- Комментарий: Data model изменения согласованы для handover в `run:dev`.
