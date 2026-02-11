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
- Перевод временного Day4 runtime-path на policy-driven control:
  - ограничить/отключить прямые write-операции агента через `kubectl`/`gh`, где требуется approval;
  - включить обязательный MCP approver/executor маршрут для привилегированных действий.
- Единообразные audit-события по ключевым действиям (label, approval, run request, job, pr).
- Timeout guard policy:
  - при `wait_state=mcp` pod/run не завершается по timeout;
  - pause/resume таймера фиксируется в аудит-событиях.
- Документация политики и минимальные тесты.

### Dependency from Day4
- Day4 допускает временный direct runtime путь (агентный контейнер с `gh`/`kubectl` для автономного цикла).
- Day6 обязан формально сузить этот путь и закрепить, какие операции остаются direct (read-only/diagnostic), а какие только через MCP approval-flow.

### Out of scope
- Полная унификация всех внешних интеграций approver/executor (Slack/Jira/etc.) сверх референсных контрактов.

## Критерии приемки эпика
- Любая попытка запуска без прав отклоняется, логируется и видна в UI/логах.
- Для привилегированных runtime-действий direct path закрыт policy-ограничениями и подтверждён тестами.
