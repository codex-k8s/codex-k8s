---
doc_id: ARC-MCP-CK8S-0001
type: mcp-approval-flow
title: "codex-k8s — MCP Approval and Audit Flow"
status: draft
owner_role: SA
created_at: 2026-02-11
updated_at: 2026-02-11
related_issues: [1]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: ""
---

# MCP Approval and Audit Flow

## TL;DR
- Любые trigger/deploy действия, инициированные агентом, проходят approval gate.
- Для `run:*` применяется owner approval перед фактическим label apply.
- Все действия логируются в единый audit-контур (`flow_events`, `agent_sessions`, `links`, `token_usage`).

## Политика апрувов

### Обязательный approval gate
- Применяется к агент-инициированным `run:*` label operations.
- Решение принимает Owner (или делегированный approver policy).
- До апрува действие остаётся в состоянии `pending approval`.

### Без обязательного approval gate
- `state:*` и `need:*` можно применять автоматически по policy.
- Не допускается их использование как скрытых trigger/deploy сигналов.

## Последовательность (высокоуровнево)

1. Агент формирует `label apply request`.
2. Запрос фиксируется в audit (`approval.requested`).
3. Owner принимает `approve/deny`.
4. При `approve` применяется label и создаётся `approval.approved` + `label.applied`.
5. При `deny` создаётся `approval.denied`; workflow не запускается.

## Timeout поведение во время MCP ожидания

- Когда run/session находится в `wait_state=mcp`, timeout-kill не применяется.
- Таймер run переводится в paused state до получения ответа MCP/approval callback.
- После получения ответа таймер возобновляется с оставшимся временем.
- Смена wait-state и pause/resume таймера фиксируется в `flow_events`.

## Обязательные audit-поля

- `correlation_id`
- `actor_type` и `actor_id`
- `event_type`
- `approval_state` (если применимо)
- `payload` (label, issue/pr/run refs, reason)
- timestamp

## События минимального набора

- `label.requested`
- `approval.requested`
- `approval.approved`
- `approval.denied`
- `label.applied`
- `label.rejected`
- `run.enqueued`
- `run.started`
- `run.finished`
- `run.wait.paused`
- `run.wait.resumed`

## Интеграция с DocSet/traceability

- Для каждого `run:*` этапа связываются:
  - Issue/PR,
  - run record,
  - документы этапа.
- Связи пишутся в `links` и отражаются в `docs/delivery/issue_map.md`.

## Связанные документы
- `docs/product/labels_and_trigger_policy.md`
- `docs/product/stage_process_model.md`
- `docs/architecture/data_model.md`
- `docs/delivery/requirements_traceability.md`
