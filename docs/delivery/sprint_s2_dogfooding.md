---
doc_id: SPR-CK8S-0002
type: sprint-plan
title: "Sprint S2: Dogfooding via Issue labels (run:dev / run:dev:revise)"
status: active
owner_role: EM
created_at: 2026-02-10
updated_at: 2026-02-11
related_issues: []
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-11-s2-progress"
---

# Sprint S2: Dogfooding via Issue labels (run:dev / run:dev:revise)

## TL;DR
- Спринт доводит платформу до режима dogfooding: разработка запускается от GitHub Issue лейблов и завершается PR.
- Сначала: архитектурное выравнивание (thin-edge gateway + домен в control-plane).
- До расширения внешнего транспорта: contract-first OpenAPI (api-gateway + web-console client codegen).
- Потом: label-driven orchestration + отдельные namespaces + agent job + PR flow + UI наблюдение.
- Параллельно: фиксируется канонический каталог `run:*`, `state:*`, `need:*` и policy апрувов для trigger/deploy действий.

## План эпиков по дням

| День | Эпик | Priority | Документ | Статус |
|---|---|---|---|---|
| Day 0 | Control-plane extraction + thin-edge gateway | P0 | `docs/delivery/epics/epic-s2-day0-control-plane-extraction.md` | completed |
| Day 1 | Migrations/schema ownership + OpenAPI contract-first baseline | P0 | `docs/delivery/epics/epic-s2-day1-migrations-and-schema-ownership.md` | completed |
| Day 2 | Issue label triggers: `run:dev`, `run:dev:revise` | P0 | `docs/delivery/epics/epic-s2-day2-issue-label-triggers-run-dev.md` | active |
| Day 3 | Per-issue namespace + RBAC/resource policy baseline | P0 | `docs/delivery/epics/epic-s2-day3-per-issue-namespace-and-rbac.md` | planned |
| Day 4 | Agent job image + git/gh PR flow | P0 | `docs/delivery/epics/epic-s2-day4-agent-job-and-pr-flow.md` | planned |
| Day 5 | Staff UI: dogfooding visibility + drilldowns | P1 | `docs/delivery/epics/epic-s2-day5-staff-ui-dogfooding-observability.md` | planned |
| Day 6 | Approvals/audit hardening for trigger actions | P1 | `docs/delivery/epics/epic-s2-day6-approval-and-audit-hardening.md` | planned |
| Day 7 | Regression gate for dogfooding end-to-end | P0 | `docs/delivery/epics/epic-s2-day7-dogfooding-regression-gate.md` | planned |

## Daily gate (обязательно)
- Планирование/DoR на день выполнены.
- Изменения дня задеплоены на staging.
- Пройден ручной smoke (минимум: webhook -> run -> worker -> k8s -> UI).
- Если менялись API/data model/RBAC/webhook процессы: документация обновлена синхронно.

## Scope labels для S2
- Активные trigger labels в исполнении: `run:dev`, `run:dev:revise`.
- Каталог `run:*` фиксируется полностью в документации и GitHub vars, остальные trigger labels остаются `planned`.
- `state:*` и `need:*` используются для служебной оркестрации и блокировок, без прямого запуска deploy.
