---
doc_id: EPC-CK8S-0002
type: epic
title: "Epic Catalog: Sprint S2 (Dogfooding via Issues)"
status: active
owner_role: EM
created_at: 2026-02-10
updated_at: 2026-02-13
related_issues: [19]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-11-s2-progress"
---

# Epic Catalog: Sprint S2 (Dogfooding via Issues)

## TL;DR
- Цель Sprint S2: довести `codex-k8s` до режима dogfooding, где разработка `codex-k8s` запускается через GitHub Issue + лейблы `run:dev` и `run:dev:revise`, а агент работает в отдельном namespace со стеком и завершает цикл созданием PR.
- Первый приоритет: исправить архитектурное отклонение (thin-edge в `external/api-gateway`, домен и БД ownership в `internal/control-plane`).
- Второй приоритет: зафиксировать contract-first OpenAPI для external/staff API и только после этого расширять issue-driven run pipeline (webhook issue label -> run request -> namespace -> agent job -> PR).
- Завершающий приоритет S2: подготовить governance baseline для Sprint S3 (approval matrix, MCP control tools, regression gate).

## Контекст
- Source of truth требований: `docs/product/requirements_machine_driven.md`.
- Процессная продуктовая модель: `docs/product/agents_operating_model.md`, `docs/product/labels_and_trigger_policy.md`, `docs/product/stage_process_model.md`.
- Source of truth инженерных ограничений: `docs/design-guidelines/**` (особенно `common/project_architecture.md`, `go/services_design_requirements.md`).

## Эпики Sprint S2 (план)
- Day 0: `docs/delivery/epics/epic-s2-day0-control-plane-extraction.md`
- Day 1: `docs/delivery/epics/epic-s2-day1-migrations-and-schema-ownership.md` (включая OpenAPI contract-first baseline)
- Day 2: `docs/delivery/epics/epic-s2-day2-issue-label-triggers-run-dev.md`
- Day 3: `docs/delivery/epics/epic-s2-day3-per-issue-namespace-and-rbac.md`
- Day 3.5: `docs/delivery/epics/epic-s2-day3.5-mcp-github-k8s-and-prompt-context.md`
- Day 4: `docs/delivery/epics/epic-s2-day4-agent-job-and-pr-flow.md`
- Day 4.5: `docs/delivery/epics/epic-s2-day4.5-pgx-db-models-and-repository-refactor.md`
- Day 5: `docs/delivery/epics/epic-s2-day5-staff-ui-dogfooding-observability.md`
- Day 6: `docs/delivery/epics/epic-s2-day6-approval-and-audit-hardening.md`
- Day 7: `docs/delivery/epics/epic-s2-day7-dogfooding-regression-gate.md`

## Текущий прогресс
- Day 0: completed + approved.
- Day 1: completed + approved (OpenAPI rollout + codegen baseline внедрены).
- Day 2: completed + approved.
- Day 3: completed + approved.
- Day 3.5: completed (MCP-first tool layer + prompt context assembler готовы как dependency для Day4).
- Day 4: completed (agent-runner runtime, session persistence/resume, split access model и PR-flow через MCP).
- Day 4.5: planned (pgx + db-model rollout).
- Day 5..7: planned.
- Day 6: scope расширен до approval matrix + MCP control tools (secret/db/feedback) + audit hardening.
- Day 7: scope расширен до MVP readiness regression gate + Sprint S3 kickoff package.

## Критерий успеха Sprint S2 (выжимка)
- Один Issue с лейблом `run:dev` приводит к запуску агентного Job в отдельном namespace и к созданию PR.
- Один Issue с лейблом `run:dev:revise` запускает цикл ревизии и обновляет PR.
- `external/api-gateway` остаётся thin-edge и не содержит доменной логики/репозиториев.
- Sprint S3 стартует без P0-блокеров и с зафиксированным планом `Day1..Day12`.
