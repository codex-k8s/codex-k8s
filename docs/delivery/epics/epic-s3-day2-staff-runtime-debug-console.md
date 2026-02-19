---
doc_id: EPC-CK8S-S3-D2
type: epic
title: "Epic S3 Day 2: Staff runtime debug console"
status: completed
owner_role: EM
created_at: 2026-02-13
updated_at: 2026-02-13
related_issues: [19]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-19-full-docset"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-19
---

# Epic S3 Day 2: Staff runtime debug console

## TL;DR
- Цель: дать оператору минимально достаточную runtime-диагностику в staff UI.
- MVP-результат: видны running jobs, live/history logs и queue ожидающих run.

## Priority
- `P0`.

## Scope
### In scope
- Экран running jobs с фильтрацией по stage/agent/status.
- Live log tail + исторический архив логов/flow events.
- Экран wait queue: `waiting_mcp`, `waiting_owner_review`, причина ожидания и SLA таймер.
- Ссылки на issue/pr/namespace/job и переходы в traceability.

### Out of scope
- Полная observability-платформа и кастомные dashboard-конструкторы.

## Критерии приемки
- По одному run оператор видит runtime-состояние, историю и причину блокировки без доступа к raw pod.

## Фактический результат (выполнено)
- Добавлены backend-методы control-plane/staff:
  - `ListRunJobs` (фильтры: `trigger_kind`, `status`, `agent_key`);
  - `ListRunWaits` (фильтры: `trigger_kind`, `status`, `agent_key`, `wait_state`);
  - `GetRunLogs` (snapshot + tail lines).
- Расширена модель `Run` runtime-полями:
  - `agent_key`, `wait_since`, `last_heartbeat_at`.
- В `staffrun` PostgreSQL repository добавлены:
  - выборки jobs/waits c фильтрацией и RBAC-scope (admin/user projects);
  - выборка run logs из `agent_runs.agent_logs_json`;
  - нормализация wait-reason (`waiting_mcp`, `waiting_owner_review`).
- В api-gateway добавлены staff endpoints:
  - `GET /api/v1/staff/runs/jobs`;
  - `GET /api/v1/staff/runs/waits`;
  - `GET /api/v1/staff/runs/{run_id}/logs`.
- В staff web-console реализовано:
  - отдельный блок `Running jobs` с фильтрами и таблицей runtime-полей;
  - отдельный блок `Wait queue` с `wait_state`, `wait_since`, SLA elapsed, `last_heartbeat_at`;
  - в деталях run добавлен блок логов: `tail_lines` + raw snapshot JSON.
- Сохранена traceability-навигация:
  - ссылки на project/issue/PR из runtime-таблиц;
  - drilldown в детали run.

## Проверки
- `go test ./services/internal/control-plane/... ./services/external/api-gateway/...` — passed.
- `make lint-go` — passed.
- `make dupl-go` — passed.
- `npm run build` (`services/staff/web-console`) — passed.
