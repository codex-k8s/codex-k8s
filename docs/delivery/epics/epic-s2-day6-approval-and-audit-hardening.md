---
doc_id: EPC-CK8S-S2-D6
type: epic
title: "Epic S2 Day 6: Trigger approvals and audit hardening"
status: planned
owner_role: EM
created_at: 2026-02-10
updated_at: 2026-02-12
related_issues: []
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: ""
---

# Epic S2 Day 6: Trigger approvals and audit hardening

## TL;DR
- Цель эпика: ужесточить правила запуска (trigger labels) и сделать аудит достаточным для ежедневного dogfooding.
- Ключевая ценность: снижение риска несанкционированных запусков и прозрачность действий.
- MVP-результат: политика “кто может запускать” формализована и проверяется, события неизменно пишутся в audit/flow_events.

## Priority
- `P1`.

## Scope
### In scope
- Явная политика authorizer:
  - для агент-инициированных trigger/deploy labels (`run:*`) обязателен owner approval;
  - `state:*` и `need:*` могут применяться автоматически по policy.
- Ужесточение MCP-first policy:
  - верифицировать, что привилегированные write-операции доступны только через MCP approver/executor маршрут;
  - запретить обход policy через альтернативные каналы выполнения.
- Ввести платформенно-управляемую матрицу MCP policy:
  - effective policy для `agent_key + run label`;
  - явное разделение scope/action по сущностям Kubernetes и GitHub;
  - поддержка composite tools (например GitHub+Kubernetes secret sync) как отдельного класса policy.
- Единообразные audit-события по ключевым действиям (label, approval, run request, job, pr).
- Timeout guard policy:
  - при `wait_state=mcp` pod/run не завершается по timeout;
  - pause/resume таймера фиксируется в аудит-событиях.
- Документация политики и минимальные тесты.

### Out of scope
- Полная унификация всех внешних интеграций approver/executor (Slack/Jira/etc.) сверх референсных контрактов.

## Критерии приемки эпика
- Любая попытка запуска без прав отклоняется, логируется и видна в UI/логах.
- Для привилегированных runtime-действий подтвержден MCP-only путь и покрыт тестами.
