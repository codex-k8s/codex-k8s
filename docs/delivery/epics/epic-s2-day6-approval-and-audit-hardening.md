---
doc_id: EPC-CK8S-S2-D6
type: epic
title: "Epic S2 Day 6: Trigger approvals and audit hardening"
status: planned
owner_role: EM
created_at: 2026-02-10
updated_at: 2026-02-10
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
- Явная политика authorizer для trigger-лейблов и run requests.
- Единообразные flow_events по ключевым действиям (label, run request, job, pr).
- Документация политики и минимальные тесты.

### Out of scope
- Полная интеграция внешнего MCP-слоя (yaml-mcp-server) как обязательного пути (это отдельный этап).

## Критерии приемки эпика
- Любая попытка запуска без прав отклоняется, логируется и видна в UI/логах.

