---
doc_id: EPC-CK8S-S2-D5
type: epic
title: "Epic S2 Day 5: Staff UI for dogfooding visibility (runs/issues/PRs)"
status: completed
owner_role: EM
created_at: 2026-02-10
updated_at: 2026-02-13
related_issues: []
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-13-s2-day5"
---

# Epic S2 Day 5: Staff UI for dogfooding visibility (runs/issues/PRs)

## TL;DR
- Цель эпика: дать оператору платформы видимость по issue-driven run pipeline.
- Ключевая ценность: меньше “слепых зон” при dogfooding.
- MVP-результат: UI показывает Issue -> Run -> Job/Namespace -> PR и даёт drilldown по событиям/логам.

## Priority
- `P1`.

## Scope
### In scope
- UI разделы/таблицы для run requests и их статусов.
- Отображение связанного PR и ссылок.
- Базовый drilldown по `flow_events`, `agent_sessions`, `token_usage` и traceability `links`.
- Базовое отображение snapshot логов из `agent_runs.agent_logs_json` в run details.
- Видимость paused/waiting статусов (`waiting_owner_review`, `waiting_mcp`) и resumable признака сессии.
- Видимость Day4 execution-артефактов:
  - branch name, PR link/number;
  - `template_source/template_locale/template_version`;
  - session/thread identity для resume диагностики.

### Out of scope
- Полный UI для управления документами/шаблонами (отдельный этап).
- Live-stream логов агента (SSE/WebSocket) — фиксируется как follow-up после базового snapshot drilldown.

## Критерии приемки эпика
- По одному экрану можно понять: что запущено, где работает (namespace/job) и что получилось (PR).

## Реализация (2026-02-13)
- Run details API/UI доведены до операторского сценария удаления namespace:
  - runtime state (`job_name`, `job_namespace`, `namespace`, `job_exists`, `namespace_exists`) отдается в `GET /staff/runs/{id}`;
  - кнопка удаления namespace отображается только если есть активная job.
- Исправлен сценарий `DELETE /staff/runs/{id}/namespace` для кейсов без status-comment:
  - добавлен fallback поиска namespace по `run_id` label;
  - создается/обновляется статус-комментарий с фазой удаления.
- Доработан run details UX:
  - явные блоки Проект/Issue/PR/Trigger;
  - события в обратной сортировке (свежее сверху);
  - learning feedback убран из run-details сценария;
  - даты/время в UI форматируются без timezone suffix.
- Доработан runs list UX:
  - убран столбец `Создано`;
  - добавлены столбцы Issue/PR, run type, trigger label;
  - добавлена пагинация по 20 элементов на страницу.
- Ошибки в UI автоскрываются через 5 секунд.
- Контракт синхронизирован:
  - обновлены `proto`/gRPC модели run;
  - обновлен OpenAPI и codegen backend/frontend артефактов.
