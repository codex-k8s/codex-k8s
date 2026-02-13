---
doc_id: EPC-CK8S-S3-D5
type: epic
title: "Epic S3 Day 5: Owner feedback handle and HTTP approver/executor interfaces"
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

# Epic S3 Day 5: Owner feedback handle and HTTP approver/executor interfaces

## TL;DR
- Цель: стандартизовать канал оперативных решений владельца для run-процессов.
- MVP-результат: feedback handle с вариантами ответов и Telegram adapter как референс реализации.

## Priority
- `P0`.

## Scope
### In scope
- MCP tool `owner.feedback.request` (question + options + optional custom input).
- HTTP contracts для approver/executor (request, callback, retry, idempotency).
- Telegram adapter baseline с поддержкой approve/deny/option/custom.
- Интеграция с wait queue и timeout pause/resume.

### Out of scope
- Нативные UI-адаптеры под Slack/Jira/Mattermost (только контрактная совместимость).

## Критерии приемки
- Агент может получить структурированный ответ Owner и продолжить run без ручного вмешательства в БД.
- Callback-события корректно обновляют state run и audit.
