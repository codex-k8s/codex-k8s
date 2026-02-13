---
doc_id: EPC-CK8S-S2-D6
type: epic
title: "Epic S2 Day 6: Approval matrix, MCP control tools and audit hardening"
status: completed
owner_role: EM
created_at: 2026-02-10
updated_at: 2026-02-13
related_issues: [19]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-13-s2-day6"
---

# Epic S2 Day 6: Approval matrix, MCP control tools and audit hardening

## TL;DR
- Цель эпика: закрыть security/governance контур для MVP перед финальным regression gate.
- Ключевая ценность: привилегированные действия переходят на детерминированные MCP-инструменты с явным approval и полным audit-trail.
- MVP-результат (факт S2 Day6): реализованы policy-driven approvals для MCP control tools, persisted approval queue + wait-state governance + staff approvals UI/API + расширенный audit-контур.

## Priority
- `P0`.

## Scope
### In scope
- Утверждение и реализация platform policy matrix:
  - effective policy по связке `agent_key + run_label + runtime_mode`;
  - явное разделение `approval:none` / `approval:owner` / `approval:delegated`;
  - запрет обхода через прямые write-каналы для операций, отмеченных как privileged.
- MCP control tools (минимальный MVP-набор):
  - `secret.sync.github_k8s`: детерминированное создание/обновление секрета в GitHub и Kubernetes для выбранного окружения;
  - `database.lifecycle`: create/delete/describe database в выбранном окружении по policy;
  - `mcp_owner_feedback_request`: оперативный вопрос владельцу с 2-5 вариантами + `custom` ответ.
- Безопасность control tools:
  - автогенерация секрет-значений внутри инструмента без вывода в модель;
  - идемпотентность повторных вызовов;
  - dry-run/simulation режим для ревизии и диагностики.
- HTTP approver/executor contracts:
  - унифицированный контракт request/callback с обязательным `correlation_id`;
  - поддержка статусов `approved` / `denied` / `expired` / `failed`;
  - интеграция Telegram approver/executor как первый production adapter.
- UX feedback/approval (по референсу `telegram-executor` + `telegram-approver`):
  - `owner feedback` поддерживает не только текстовый `custom`, но и voice/STT вариант ответа;
  - при `deny` для privileged action поддерживается диктовка причины (voice/STT) на стороне адаптера.
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
- Полная линейка внешних адаптеров (Slack/Jira/Mattermost) и production Telegram adapter rollout.
- Полный self-service UI для управления policy packs (выносится в Sprint S3).

## Критерии приемки эпика
- Любая privileged операция без апрува отклоняется и логируется как `approval.denied` или `failed_precondition`.
- `secret.sync.github_k8s` не раскрывает секретный материал в логах/PR/comments/flow events.
- `database.lifecycle` корректно обрабатывает create/delete и повторные вызовы без дрейфа состояния.
- `mcp_owner_feedback_request` поддерживает вариантные ответы и корректно резюмируется в run context.
- Voice/STT ответы owner (feedback + deny reason) корректно принимаются через HTTP contract и фиксируются в audit без утечки секретов.
- В staff UI видны pending approvals, wait reason и итог апрува по каждому run.

## Фактический результат (выполнено)
- В `control-plane` добавлены Day6 доменные сущности и persistence:
  - `mcp_action_requests` (approval queue, state transitions, target/payload snapshots);
  - расширение `agent_sessions` полями wait-state (`wait_state`, `timeout_guard_disabled`, `last_heartbeat_at`).
- Реализованы MCP control tools в domain/service слое:
  - `secret.sync.github_k8s`;
  - `database.lifecycle`;
  - `mcp_owner_feedback_request`.
- Добавлен approval lifecycle:
  - создание pending request;
  - approve/deny/expired/failed/apply transitions;
  - идемпотентный re-use pending request по сигнатуре действия.
- Реализован wait-state governance:
  - перевод run/session в `waiting_mcp` при ожидании approval;
  - снятие wait-state после финального решения;
  - pause/resume отражаются в `flow_events`.
- Расширен аудит событий:
  - `approval.requested|approved|denied|expired|failed|applied`;
  - `run.wait.paused|run.wait.resumed`.
- Расширены transport-контракты:
  - gRPC: `ListPendingApprovals`, `ResolveApprovalDecision`;
  - OpenAPI/staff HTTP: список pending approvals и endpoint принятия решения.
- Staff UI получил Day6 drilldown:
  - список pending approvals;
  - approve/deny с reason;
  - отображение `wait_state`/`wait_reason` в run details.

## Перенос в следующий этап
- HTTP approver/executor адаптеры (включая `github.com/codex-k8s/telegram-approver` и `github.com/codex-k8s/telegram-executor`) остаются частью S3 delivery для production rollout.
