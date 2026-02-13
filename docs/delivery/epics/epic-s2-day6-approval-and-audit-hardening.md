---
doc_id: EPC-CK8S-S2-D6
type: epic
title: "Epic S2 Day 6: Approval matrix, MCP control tools and audit hardening"
status: planned
owner_role: EM
created_at: 2026-02-10
updated_at: 2026-02-13
related_issues: [19]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: ""
---

# Epic S2 Day 6: Approval matrix, MCP control tools and audit hardening

## TL;DR
- Цель эпика: закрыть security/governance контур для MVP перед финальным regression gate.
- Ключевая ценность: привилегированные действия переходят на детерминированные MCP-инструменты с явным approval и полным audit-trail.
- MVP-результат: готова policy-матрица, минимальные control tools (secrets/db/feedback), HTTP approver contracts и Telegram adapter baseline.

## Priority
- `P0`.

## Scope
### In scope
- Утверждение и реализация platform policy matrix:
  - effective policy по связке `agent_key + run_label + runtime_mode`;
  - явное разделение `approval:none` / `approval:owner` / `approval:delegated`;
  - запрет обхода через прямые write-каналы для операций, отмеченных как privileged.
- MCP control tools (минимальный MVP-набор):
  - `mcp_secret_sync_env`: детерминированное создание/обновление секрета в GitHub и Kubernetes для выбранного окружения;
  - `mcp_database_lifecycle`: create/delete database в выбранном окружении по policy;
  - `mcp_owner_feedback_request`: оперативный вопрос владельцу с 2-5 вариантами + `custom` ответ.
- Безопасность control tools:
  - автогенерация секрет-значений внутри инструмента без вывода в модель;
  - идемпотентность повторных вызовов;
  - dry-run/simulation режим для ревизии и диагностики.
- HTTP approver/executor contracts:
  - унифицированный контракт request/callback с обязательным `correlation_id`;
  - поддержка статусов `approved` / `denied` / `expired` / `failed`;
  - Telegram approver/executor baseline как первый production adapter.
- Wait-state governance:
  - `waiting_mcp` и `waiting_owner_review` отражаются в БД/аудите;
  - timeout для `waiting_mcp` всегда paused;
  - restart/resume без потери контекста approval-запросов.
- Observability/аудит:
  - унифицированные события `approval.*`, `mcp.tool.*`, `run.wait.*`;
  - отдельный drilldown в staff UI по pending approvals и wait reasons;
  - traceability `issue/pr <-> run <-> approval_request` в `links`.
- Документация и тесты:
  - обновление product/architecture/delivery документов;
  - интеграционные тесты deny/approve/timeout для MCP control tools.

### Out of scope
- Полная линейка внешних адаптеров (Slack/Jira/Mattermost) сверх Telegram baseline.
- Полный self-service UI для управления policy packs (выносится в Sprint S3).

## Критерии приемки эпика
- Любая privileged операция без апрува отклоняется и логируется как `approval.denied` или `failed_precondition`.
- `mcp_secret_sync_env` не раскрывает секретный материал в логах/PR/comments/flow events.
- `mcp_database_lifecycle` корректно обрабатывает create/delete и повторные вызовы без дрейфа состояния.
- `mcp_owner_feedback_request` поддерживает вариантные ответы и корректно резюмируется в run context.
- В staff UI видны pending approvals, wait reason и итог апрува по каждому run.
