---
doc_id: EPC-CK8S-S3-D5
type: epic
title: "Epic S3 Day 5: Owner feedback handle and HTTP approver/executor interfaces"
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
- Контрактная политика:
  - контракт публикуется и поддерживается платформой `codex-k8s`;
  - внешние команды самостоятельно реализуют совместимые адаптеры без изменений core.

### Out of scope
- Нативные UI-адаптеры под Slack/Jira/Mattermost (только контрактная совместимость).

## Критерии приемки
- Агент может получить структурированный ответ Owner и продолжить run без ручного вмешательства в БД.
- Callback-события корректно обновляют state run и audit.

## Фактический результат (выполнено)
- MCP control tool приведён к каноническому имени:
  - `owner.feedback.request`.
- Добавлены и задокументированы HTTP callback-контракты для внешних approver/executor адаптеров:
  - `POST /api/v1/mcp/approver/callback`;
  - `POST /api/v1/mcp/executor/callback`.
- В callback-контракте поддержаны решения:
  - `approved`, `denied`, `expired`, `failed`, `applied`.
- Защита callback-контуров реализована shared-token политикой:
  - заголовок `X-Codex-MCP-Token` или `Authorization: Bearer ...`.
- В control-plane доработан lifecycle approval state:
  - добавлен `applied` в нормализованный decision/state flow;
  - реализованы допустимые переходы `requested -> approved|applied|denied|expired|failed` и `approved -> applied|failed`;
  - подтверждён idempotent-путь повторных callback-решений без повторного side effect.
- При callback-решении синхронизируется run wait-state:
  - снимается `waiting_mcp`, возобновляется timeout-guard логика.

## Проверки
- `make gen-openapi-go` — passed.
- `make gen-openapi-ts` — passed.
- `go test ./services/external/api-gateway/internal/...` — passed.
- `go test ./services/internal/control-plane/internal/domain/mcp` — passed.
