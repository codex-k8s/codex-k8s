---
doc_id: ARC-MCP-CK8S-0001
type: mcp-approval-flow
title: "codex-k8s — MCP Approval and Audit Flow"
status: draft
owner_role: SA
created_at: 2026-02-11
updated_at: 2026-02-13
related_issues: [1, 19]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: ""
---

# MCP Approval and Audit Flow

## TL;DR
- MCP в MVP baseline используется для GitHub label-операций и control tools (secret sync, database lifecycle, owner feedback).
- GitHub issue/PR/comments и Kubernetes runtime-операции выполняются агентом напрямую через `gh`/`kubectl`.
- Approval gate в MCP управляется policy matrix: для label-инструментов возможен `approval:none`, для privileged control tools — `approval:required`.
- Все действия логируются в единый audit-контур (`flow_events`, `agent_sessions`, `links`, `token_usage`).
- HTTP approver/executor поддерживаются как стандартные контракты интеграции; Telegram зафиксирован как приоритетный adapter path для следующего этапа.
- В `codex-k8s` сохраняется двухслойная модель MCP: встроенные Go-ручки платформы + внешний декларативный слой (`github.com/codex-k8s/yaml-mcp-server`).

## Политика апрувов

### Baseline (текущий этап)
- Для MCP label-инструментов (`github_labels_list|add|remove|transition`) используется `approval:none`.
- Label transitions всё равно проходят через control-plane MCP, чтобы сохранять единый audit-контур.
- Для control tools (`secret.sync.github_k8s`, `database.lifecycle`, `owner.feedback.request`) включается approver gate по policy.
- Для `secret.sync.github_k8s` действует idempotency-key и retry-safe replay без повторного side effect.
- Для `database.lifecycle`:
  - `create/delete` идут через approval flow;
  - `describe` выполняется как read-only action без side effects;
  - `delete` требует явного `confirm_delete=true`;
  - ownership-check выполняется по таблице `project_databases`;
  - окружения ограничиваются allowlist (`CODEXK8S_PROJECT_DB_LIFECYCLE_ALLOWED_ENVS`, fallback `dev,staging,ai-staging,prod`).

### Planned (следующие этапы)
- Для части label/runtime/secret инструментов будет включаться обязательный approver gate.
- Решение будет принимать Owner или делегированный approver policy.
- До апрува действие остаётся в состоянии `pending approval`.

## Последовательность (высокоуровнево)

1. Агент формирует MCP request (label/control tool).
2. Запрос фиксируется в audit (`approval.requested`).
3. Owner принимает `approve/deny`.
4. При `approve` выполняется действие и создаётся `approval.approved` + `mcp.tool.applied`.
5. При `deny` создаётся `approval.denied`; действие не выполняется.

## Базовый режим S2 Day4+

- Начиная с Day4, для agent pod действует split access model:
  - в pod выдаётся отдельный Git bot-token (`CODEXK8S_GIT_BOT_TOKEN`) для `gh/git` операций;
  - control-plane MCP инструменты используют bot-token из `platform_github_tokens.bot_token_encrypted`;
  - для `full-env` формируется namespaced `KUBECONFIG` и разрешён direct `kubectl` в рамках namespace;
  - MCP остаётся для label operations и policy-аудита transitions.
- `repositories.token_encrypted` в этом режиме не используется MCP runtime-контуром
  и остаётся в domain-path управления репозиториями (staff/project management).
- Day6+ расширяет policy: approver matrix, secret-management инструменты через MCP, единообразные события и отказоустойчивость.
- Day6+ также включает контур `run:self-improve`, где MCP используется для traceable transitions и owner feedback loops.
- В Day3 добавлен deterministic secret materialization для `secret.sync.github_k8s` (policy-driven generation + idempotent apply/replay).

## Политики доступа к MCP (S2 Day6 baseline + roadmap)

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
  - пример: `secret.sync.github_k8s` (создание/обновление секрета одновременно в GitHub и Kubernetes);
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
- `mcp.tool.requested`
- `mcp.tool.applied`
- `mcp.tool.failed`
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
