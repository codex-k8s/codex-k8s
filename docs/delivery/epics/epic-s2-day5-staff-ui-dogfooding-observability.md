---
doc_id: EPC-CK8S-S2-D5
type: epic
title: "Epic S2 Day 5: Staff UI for dogfooding visibility (runs/issues/PRs)"
status: planned
owner_role: EM
created_at: 2026-02-10
updated_at: 2026-02-11
related_issues: []
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: ""
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
- Видимость paused/waiting статусов (`waiting_owner_review`, `waiting_mcp`) и resumable признака сессии.

### Out of scope
- Полный UI для управления документами/шаблонами (отдельный этап).

## Критерии приемки эпика
- По одному экрану можно понять: что запущено, где работает (namespace/job) и что получилось (PR).
