---
doc_id: EPC-CK8S-S3-D2
type: epic
title: "Epic S3 Day 2: Staff runtime debug console"
status: planned
owner_role: EM
created_at: 2026-02-13
updated_at: 2026-02-13
related_issues: [19]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: ""
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
