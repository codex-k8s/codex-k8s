---
doc_id: DMD-CK8S-0119
type: design-data-model
title: "Issue #119 — Design Data Model and Evidence Schema"
status: draft
owner_role: SA
created_at: 2026-02-24
updated_at: 2026-02-24
related_issues: [119, 118]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-24-issue-119-design"
---

# Data Model: Issue #119

## TL;DR
- Новые таблицы/поля не требуются.
- Evidence для A+B строится на существующих сущностях: `agent_runs`, `flow_events`, `agent_sessions`, `links`.
- Основная задача design-этапа: зафиксировать обязательный набор атрибутов для проверки pass/fail.

## Используемые сущности

| Entity | Роль в issue #119 | Ключевые поля |
|---|---|---|
| `agent_runs` | факт запуска и итоговый статус | `id`, `status`, `project_id`, `run_payload`, `started_at`, `finished_at` |
| `flow_events` | канонический аудит transitions | `correlation_id`, `event_type`, `payload`, `created_at` |
| `agent_sessions` | effective profile и session telemetry | `run_id`, `model`, `reasoning_effort`, `status`, `session_json` |
| `links` | трассировка issue/pr/doc/run связей | source/target refs + link type |

## Evidence schema (логическая модель)
- ScenarioRecord:
  - `scenario_key` (`A1|A2|A3|B1|B2|B3`)
  - `result` (`pass|fail`)
  - `issue_number`
  - `pr_number` (nullable для doc-only шагов)
  - `run_id`
  - `correlation_id`
  - `expected_transitions[]`
  - `actual_transitions[]`
  - `notes`

## Инварианты данных
- Для каждого ScenarioRecord должен существовать как минимум один `run_id`.
- Для B2 не должно существовать run с trigger `run:<stage>:revise` в момент ambiguity.
- Для B3 `agent_sessions.model/reasoning_effort` соответствуют policy resolver chain.
- Все ссылки на evidence в Issue #118 должны иметь трассировку на `run_id` и `flow_events`.

## SQL-шаблоны проверки (read-only)

```sql
-- Срез запусков по issue #119
SELECT id, status, started_at, finished_at
FROM agent_runs
WHERE run_payload ->> 'issue_number' = '119'
ORDER BY started_at DESC;
```

```sql
-- Проверка transitions и ambiguity-path
SELECT event_type, payload, created_at
FROM flow_events
WHERE payload ->> 'issue_number' = '119'
ORDER BY created_at DESC;
```

## Migration impact
- DB migration: none.
- Backfill: none.
- Data retention policy: без изменений.

## Runtime impact
- Только дополнительный анализ существующих данных для evidence bundle.
- Путь записи данных не меняется.

## Апрув
- request_id: owner-2026-02-24-issue-119-design
- Решение: pending
- Комментарий:
