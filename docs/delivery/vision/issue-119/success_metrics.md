---
doc_id: MET-CK8S-0119
type: success-metrics
title: "Issue #119 — Метрики успеха E2E A+B"
status: draft
owner_role: PM
created_at: 2026-02-24
updated_at: 2026-02-24
related_issues: [119]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-24-issue-119-vision"
---

# Метрики успеха: Issue #119 — E2E A+B

## TL;DR
- North Star Metric: 100% прохождение A+B набора без P0/P1 регрессий.
- Поддерживающие метрики: корректность label/state transitions, успешность review-driven revise, полнота evidence.

## Контекст
Метрики должны подтвердить, что критический e2e-контур stage lifecycle и review-driven revise воспроизводим и безопасен для MVP gate.

## Метрики
### 1) North Star
- Название: Pass rate A+B (A1–A3, B1–B3)
- Определение: доля сценариев A/B, завершённых без P0/P1 регрессий и с ожидаемыми переходами.
- Формула: успешные сценарии / 6 сценариев.
- Источник: evidence bundle (Issue #118) + flow_events/agent_runs.
- Частота: по завершению прогона.
- Цель/Target: 100%.

### 2) Supporting metrics
- Label Transition Accuracy
  - Определение: доля ожидаемых `run:*`/`state:*`/`need:*` переходов, зафиксированных в `flow_events`.
  - Источник: flow_events export + evidence bundle.
  - Target: 100%.
- Review-driven revise success
  - Определение: доля B1/B3 сценариев, где revise-run стартовал корректно и завершился без ambiguity.
  - Источник: PR review events + run status.
  - Target: 100%.
- Ambiguity handling correctness
  - Определение: B2 сценарий приводит к `need:input` и отсутствию revise-run.
  - Источник: flow_events + Issue comments.
  - Target: 100%.
- Evidence completeness
  - Определение: наличие всех обязательных артефактов evidence (run_id, label transitions, логи, SQL/срезы).
  - Источник: Issue #118.
  - Target: 100%.

## Сигналы раннего предупреждения (Guardrails)
- Любая P0/P1 регрессия на A/B сценариях.
- Отсутствие audit-событий для run/label переходов.
- Несоответствие между issue/PR labels и flow_events.

## План сбора данных
- Источники: `flow_events`, `agent_runs`, `mcp_action_requests`, issue/PR comments.
- Место хранения: evidence bundle в Issue #118 + при необходимости выгрузки из БД.
- Ответственный: PM + QA (с подтверждением Owner).

## Апрув
- Запрошен: 2026-02-24 (request_id: owner-2026-02-24-issue-119-vision)
- Решение: pending
- Комментарий:
