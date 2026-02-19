---
doc_id: SPR-CK8S-0002
type: sprint-plan
title: "Sprint S2: Dogfooding via Issue labels (run:dev / run:dev:revise)"
status: completed
owner_role: EM
created_at: 2026-02-10
updated_at: 2026-02-13
related_issues: [19]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-19-full-docset"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-19
---

# Sprint S2: Dogfooding via Issue labels (run:dev / run:dev:revise)

## TL;DR
- Спринт доводит платформу до режима dogfooding: разработка запускается от GitHub Issue лейблов и завершается PR.
- Сначала: архитектурное выравнивание (thin-edge gateway + домен в control-plane).
- До расширения внешнего транспорта: contract-first OpenAPI (api-gateway + web-console client codegen).
- Потом: label-driven orchestration + отдельные namespaces + MCP tool layer + agent job + PR flow + UI наблюдение.
- Финал S2 (Day6/Day7): approval matrix, MCP control tools baseline (secret/db/feedback), audit hardening и regression gate перед Sprint S3.

## План эпиков по дням

| День | Эпик | Priority | Документ | Статус |
|---|---|---|---|---|
| Day 0 | Control-plane extraction + thin-edge gateway | P0 | `docs/delivery/epics/epic-s2-day0-control-plane-extraction.md` | completed |
| Day 1 | Migrations/schema ownership + OpenAPI contract-first baseline | P0 | `docs/delivery/epics/epic-s2-day1-migrations-and-schema-ownership.md` | completed |
| Day 2 | Issue label triggers: `run:dev`, `run:dev:revise` | P0 | `docs/delivery/epics/epic-s2-day2-issue-label-triggers-run-dev.md` | completed |
| Day 3 | Per-issue namespace + RBAC/resource policy baseline | P0 | `docs/delivery/epics/epic-s2-day3-per-issue-namespace-and-rbac.md` | completed |
| Day 3.5 | MCP GitHub/K8s tools + prompt context assembler | P0 | `docs/delivery/epics/epic-s2-day3.5-mcp-github-k8s-and-prompt-context.md` | completed |
| Day 4 | Agent job image + git/PR flow via MCP | P0 | `docs/delivery/epics/epic-s2-day4-agent-job-and-pr-flow.md` | completed |
| Day 4.5 | PGX + db-model rollout for repository layer | P0 | `docs/delivery/epics/epic-s2-day4.5-pgx-db-models-and-repository-refactor.md` | completed |
| Day 5 | Staff UI: dogfooding visibility + drilldowns | P1 | `docs/delivery/epics/epic-s2-day5-staff-ui-dogfooding-observability.md` | completed |
| Day 6 | Approval matrix + MCP control tools + audit hardening | P0 | `docs/delivery/epics/epic-s2-day6-approval-and-audit-hardening.md` | completed |
| Day 7 | MVP readiness regression gate + Sprint S3 kickoff package | P0 | `docs/delivery/epics/epic-s2-day7-dogfooding-regression-gate.md` | completed |

## Daily gate (обязательно)
- Планирование/DoR на день выполнены.
- Изменения дня задеплоены на production.
- Пройден ручной smoke (минимум: webhook -> run -> worker -> k8s -> UI).
- Если менялись API/data model/RBAC/webhook процессы: документация обновлена синхронно.

## Scope labels для S2
- Активные trigger labels в исполнении: `run:dev`, `run:dev:revise`.
- Каталог `run:*` фиксируется полностью в документации и GitHub vars, остальные trigger labels остаются `planned`.
- `state:*` и `need:*` используются для служебной оркестрации и блокировок, без прямого запуска deploy.
- На Day7 фиксируется план перехода к Sprint S3, где включается `run:self-improve` и полный stage-контур.

## Фактический прогресс (2026-02-13)
- Day 0: completed + approved.
- Day 1: completed + approved.
- Day 2: completed + approved.
- Day 3: completed + approved.
- Day 3.5: completed (ready for Day4 integration).
- Day 4: completed (agent-runner runtime + resume/session persistence + PR flow via MCP baseline).
- Day 4.5: completed (PGX/db-model rollout и repository refactor baseline).
- Day 5: completed (staff observability/drilldown + namespace lifecycle controls).
- Day 6: completed (approval matrix, MCP control tools baseline, wait-state/audit hardening, staff approvals UI/API).
- Day 7: completed (regression gate + go/no-go + Sprint S3 kickoff package, см. `docs/delivery/regression_s2_gate.md`).
- Sprint S2 закрыт без открытых `P0` блокеров.
