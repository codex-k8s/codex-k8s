---
doc_id: EPC-CK8S-S3-D4
type: epic
title: "Epic S3 Day 4: MCP database lifecycle (create/delete per env)"
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

# Epic S3 Day 4: MCP database lifecycle (create/delete per env)

## TL;DR
- Цель: дать управляемый инструмент для создания и удаления БД в выбранном окружении.
- MVP-результат: DB lifecycle операции стандартизованы, аудируются и защищены апрувом.

## Priority
- `P0`.

## Scope
### In scope
- MCP tool `database.lifecycle` (`create`, `delete`, `describe`) с environment scoping.
- Policy checks (allowed envs, naming rules, ownership checks).
- Safeguards для destructive операций (`delete` только с явным подтверждением).
- Аудит и traceability в `flow_events`/`links`.

### Out of scope
- Полноценный DBaaS и автоматический backup/restore orchestration.

## Критерии приемки
- Создание/удаление БД воспроизводимы и проходят через approval flow.
- Операции отражаются в UI и audit с `correlation_id`.
