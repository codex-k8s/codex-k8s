---
doc_id: ARC-MCP-CK8S-0001
type: mcp-approval-flow
title: "codex-k8s — MCP Approval and Audit Flow"
status: draft
owner_role: SA
created_at: 2026-02-11
updated_at: 2026-02-12
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
- HTTP approver/executor поддерживаются как стандартные контракты интеграции; Telegram — первая реализация, но не единственная.
- В `codex-k8s` сохраняется двухслойная модель MCP: встроенные Go-ручки платформы + внешний декларативный слой (`github.com/codex-k8s/yaml-mcp-server`).

## Политика апрувов

### Обязательный approval gate
- Применяется к агент-инициированным `run:*` label operations.
- Решение принимает Owner (или делегированный approver policy).
- До апрува действие остаётся в состоянии `pending approval`.
- Любые привилегированные runtime-действия (`apply/delete`, rollout/restart, deploy management) допускаются только через MCP-ручки с approver policy.

### Без обязательного approval gate
- `state:*` и `need:*` можно применять автоматически по policy.
- Не допускается их использование как скрытых trigger/deploy сигналов.

## Последовательность (высокоуровнево)

1. Агент формирует `label apply request`.
2. Запрос фиксируется в audit (`approval.requested`).
3. Owner принимает `approve/deny`.
4. При `approve` применяется label и создаётся `approval.approved` + `label.applied`.
5. При `deny` создаётся `approval.denied`; workflow не запускается.

## Базовый режим S2 Day4+

- Начиная с Day4, для agent pod действует split access model:
  - прямые Kubernetes credentials не выдаются;
  - governance-операции GitHub/Kubernetes (issue/pr/comments/labels/runtime write-path) выполняются через MCP approver/executor ручки;
  - в pod выдаётся только отдельный минимальный Git bot-token для git transport path (clone/fetch/commit/push в рабочую ветку).
- Day6 расширяет и ужесточает policy (матрица апрувов, единообразные события, тесты отказоустойчивости), но не переводит систему с direct-path, так как direct-path не является базовым режимом.

## Политики доступа к MCP (roadmap Day6)

- Политики MCP управляются через платформенные данные (а не хардкодом в prompt):
  - baseline policy по `agent_key`;
  - уточнение policy по `run:*` label/типу задачи;
  - финальная effective policy на запуск = merge(`agent policy`, `label policy`, `project overrides`).
- Для каждого инструмента/ресурса фиксируются:
  - scope (`namespace`, `cluster`, `repository`);
  - allowed actions (`read`, `write`, `approve-required`);
  - actor constraints (каким ролям и при каких label доступно).
- В `flow_events` сохраняется snapshot effective policy (ключ policy + источник), чтобы audit был воспроизводимым.

### Комбинированные ручки
- В roadmap закладываются composite MCP-ручки для атомарных операций между системами:
  - пример: `secret.sync.github_to_k8s` (создание/обновление секрета одновременно в GitHub и Kubernetes);
  - composite-ручки имеют отдельный approval профиль и отдельные события аудита.

## HTTP-контракты интеграций approver/executor

- Платформа поддерживает внешний расширяемый слой MCP (например, `github.com/codex-k8s/yaml-mcp-server`) с универсальными HTTP-интеграциями.
- `github.com/codex-k8s/telegram-approver` и `github.com/codex-k8s/telegram-executor` считаются референсными адаптерами этого контракта.
- Требование к контрактам:
  - async режим с callback обязателен для долгих операций;
  - единый `correlation_id` проходит от запроса до callback;
  - решение/результат фиксируется в `flow_events` и связывается с `agent_sessions`.
- Это позволяет добавлять Slack/Mattermost/Jira и другие адаптеры без изменений core-кода `codex-k8s`.

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

## Интеграция с traceability

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
