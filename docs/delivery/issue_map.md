---
doc_id: MAP-CK8S-0001
type: issue-map
title: "Issue ↔ Docs Map"
status: active
owner_role: KM
created_at: 2026-02-06
updated_at: 2026-02-20
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-19-full-docset"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-19
---

# Issue ↔ Docs Map

## TL;DR
Матрица трассируемости: Issue/PR ↔ документы ↔ релизы.

## Матрица
| Issue/PR | Traceability bundle | PRD | Design | ADRs | Test Plan | Release Notes | Postdeploy | Status |
|---|---|---|---|---|---|---|---|---|
| #1 | `docs/delivery/issue_map.md` + `docs/delivery/requirements_traceability.md` + `docs/delivery/sprint_s1_mvp_vertical_slice.md` + `docs/delivery/sprint_s2_dogfooding.md` | `docs/product/requirements_machine_driven.md` + `docs/product/brief.md` + `docs/product/constraints.md` + `docs/product/agents_operating_model.md` + `docs/product/labels_and_trigger_policy.md` + `docs/product/stage_process_model.md` | `docs/architecture/c4_context.md`, `docs/architecture/c4_container.md`, `docs/architecture/data_model.md`, `docs/architecture/api_contract.md`, `docs/architecture/agent_runtime_rbac.md`, `docs/architecture/mcp_approval_and_audit_flow.md`, `docs/architecture/prompt_templates_policy.md` | `ADR-0001..0004` | S2 dogfooding test plan pending | TBD | TBD | in-progress |
| #19 | `docs/delivery/issue_map.md` + `docs/delivery/requirements_traceability.md` + `docs/delivery/sprint_s2_dogfooding.md` + `docs/delivery/sprint_s3_mvp_completion.md` | `docs/product/requirements_machine_driven.md` + `docs/product/brief.md` + `docs/product/constraints.md` + `docs/product/agents_operating_model.md` + `docs/product/labels_and_trigger_policy.md` + `docs/product/stage_process_model.md` | `docs/architecture/api_contract.md`, `docs/architecture/data_model.md`, `docs/architecture/mcp_approval_and_audit_flow.md`, `docs/architecture/prompt_templates_policy.md` | `ADR-0001..0004` | S2 Day7 regression gate completed (`docs/delivery/regression_s2_gate.md`); S3 Day15 completed (`docs/delivery/epics/epic-s3-day15-mvp-closeout-and-handover.md`) | TBD | TBD | in-progress |
| #29 | `docs/delivery/issue_map.md` + project docset metadata alignment (`status/approvals`) | `docs/product/requirements_machine_driven.md` + `docs/product/brief.md` + `docs/product/constraints.md` + `docs/product/agents_operating_model.md` + `docs/product/labels_and_trigger_policy.md` + `docs/product/stage_process_model.md` | `docs/architecture/c4_context.md`, `docs/architecture/c4_container.md`, `docs/architecture/data_model.md`, `docs/architecture/api_contract.md`, `docs/architecture/agent_runtime_rbac.md`, `docs/architecture/mcp_approval_and_audit_flow.md`, `docs/architecture/prompt_templates_policy.md` | `ADR-0001..0004` | n/a (документационные правки) | n/a | n/a | completed |
| #45 | `docs/delivery/issue_map.md` + `docs/delivery/sprint_s3_mvp_completion.md` + `docs/delivery/epic_s3.md` + `docs/delivery/epics/epic-s3-day16-grpc-transport-boundary-hardening.md` | `docs/product/requirements_machine_driven.md` + `docs/product/labels_and_trigger_policy.md` + `docs/product/stage_process_model.md` | `docs/architecture/c4_container.md`, `docs/architecture/api_contract.md` + `docs/design-guidelines/go/services_design_requirements.md` + `docs/design-guidelines/go/grpc.md` | `ADR-0001..0004` | Day16 реализован: transport -> service boundary enforced, прямые repository вызовы удалены из gRPC handlers; проверка `go test ./services/internal/control-plane/...` | TBD | TBD | completed |
| #57 | `docs/delivery/issue_map.md` + `docs/delivery/requirements_traceability.md` + `docs/product/agents_operating_model.md` | `docs/product/requirements_machine_driven.md` + `docs/product/agents_operating_model.md` + `docs/product/labels_and_trigger_policy.md` | `docs/architecture/api_contract.md`, `docs/architecture/mcp_approval_and_audit_flow.md`, `docs/architecture/prompt_templates_policy.md` | `ADR-0001..0004` | MCP progress-feedback tool `run_status_report` + prompt policy update (cadence 5-7 tool calls, locale-bound communication); проверка `go test` по `mcp`/`agent-runner` пакетам | TBD | TBD | in-progress |
| #74 | `docs/delivery/issue_map.md` + `docs/delivery/requirements_traceability.md` + `docs/architecture/adr/ADR-0005-run-namespace-ttl-and-revise-reuse.md` | `docs/product/requirements_machine_driven.md` + `docs/product/labels_and_trigger_policy.md` + `docs/product/agents_operating_model.md` | `docs/architecture/agent_runtime_rbac.md`, `docs/architecture/adr/ADR-0005-run-namespace-ttl-and-revise-reuse.md` | `ADR-0005` | Design-only: зафиксированы role-based TTL policy для run namespace, reuse/lease extension на `run:*:revise`, режим `run:debug` как manual-retention; реализация и runtime-проверки выполняются следующим `run:dev` циклом | TBD | TBD | in-progress |

## Требования и трассировка
- Source of truth требований: `docs/product/requirements_machine_driven.md`.
- Матрица трассируемости: `docs/delivery/requirements_traceability.md`.

## Sprint S1 артефакты
- Process requirements: `docs/delivery/development_process_requirements.md`.
- Sprint plan: `docs/delivery/sprint_s1_mvp_vertical_slice.md`.
- Epic catalog: `docs/delivery/epic_s1.md`.
- Epic docs: `docs/delivery/epics/epic-s1-day0-bootstrap-baseline.md`, `docs/delivery/epics/epic-s1-day1-webhook-idempotency.md`, `docs/delivery/epics/epic-s1-day2-worker-slots-k8s.md`, `docs/delivery/epics/epic-s1-day3-auth-rbac-ui.md`, `docs/delivery/epics/epic-s1-day4-repository-provider.md`, `docs/delivery/epics/epic-s1-day5-learning-mode.md`, `docs/delivery/epics/epic-s1-day6-hardening-observability.md`, `docs/delivery/epics/epic-s1-day7-stabilization-gate.md`.

## Sprint S2 артефакты
- Sprint plan: `docs/delivery/sprint_s2_dogfooding.md`.
- Epic catalog: `docs/delivery/epic_s2.md`.
- Epic docs: `docs/delivery/epics/epic-s2-day0-control-plane-extraction.md`, `docs/delivery/epics/epic-s2-day1-migrations-and-schema-ownership.md`, `docs/delivery/epics/epic-s2-day2-issue-label-triggers-run-dev.md`, `docs/delivery/epics/epic-s2-day3-per-issue-namespace-and-rbac.md`, `docs/delivery/epics/epic-s2-day3.5-mcp-github-k8s-and-prompt-context.md`, `docs/delivery/epics/epic-s2-day4-agent-job-and-pr-flow.md`, `docs/delivery/epics/epic-s2-day4.5-pgx-db-models-and-repository-refactor.md`, `docs/delivery/epics/epic-s2-day5-staff-ui-dogfooding-observability.md`, `docs/delivery/epics/epic-s2-day6-approval-and-audit-hardening.md`, `docs/delivery/epics/epic-s2-day7-dogfooding-regression-gate.md`.
- Product process model docs: `docs/product/agents_operating_model.md`, `docs/product/labels_and_trigger_policy.md`, `docs/product/stage_process_model.md`.
- Day4 implementation (completed): `docs/delivery/epics/epic-s2-day4-agent-job-and-pr-flow.md` (agent-runner runtime, session persistence/resume, PR flow via MCP governance path).
- Day4.5 implementation (completed): `docs/delivery/epics/epic-s2-day4.5-pgx-db-models-and-repository-refactor.md` (typed db-model/caster rollout в repository слое).
- Day5 implementation (completed): `docs/delivery/epics/epic-s2-day5-staff-ui-dogfooding-observability.md` (runs observability, drilldown и namespace lifecycle controls).
- Day3.5 dependency: `docs/delivery/epics/epic-s2-day3.5-mcp-github-k8s-and-prompt-context.md` (MCP tool layer + prompt context assembler).
- Факт на 2026-02-13: Day0..Day7 выполнены; Sprint S2 закрыт, Sprint S3 подготовлен к старту.

## Sprint S3 артефакты
- Sprint plan: `docs/delivery/sprint_s3_mvp_completion.md`.
- Epic catalog: `docs/delivery/epic_s3.md`.
- Epic docs: `docs/delivery/epics/epic-s3-day1-full-stage-and-label-activation.md`, `docs/delivery/epics/epic-s3-day2-staff-runtime-debug-console.md`, `docs/delivery/epics/epic-s3-day3-mcp-deterministic-secret-sync.md`, `docs/delivery/epics/epic-s3-day4-mcp-database-lifecycle.md`, `docs/delivery/epics/epic-s3-day5-feedback-and-approver-interfaces.md`, `docs/delivery/epics/epic-s3-day6-self-improve-ingestion-and-diagnostics.md`, `docs/delivery/epics/epic-s3-day7-self-improve-updater-and-pr-flow.md`, `docs/delivery/epics/epic-s3-day8-agent-toolchain-auto-extension.md`, `docs/delivery/epics/epic-s3-day9-declarative-full-env-deploy-and-runtime-parity.md`, `docs/delivery/epics/epic-s3-day10-staff-console-vuetify-redesign.md`, `docs/delivery/epics/epic-s3-day11-full-env-slots-and-subdomains.md`, `docs/delivery/epics/epic-s3-day12-docset-import-and-safe-sync.md`, `docs/delivery/epics/epic-s3-day13-config-and-credentials-governance.md`, `docs/delivery/epics/epic-s3-day14-repository-onboarding-preflight.md`, `docs/delivery/epics/epic-s3-day15-mvp-closeout-and-handover.md`, `docs/delivery/epics/epic-s3-day16-grpc-transport-boundary-hardening.md`, `docs/delivery/epics/epic-s3-day17-environment-scoped-secret-overrides-and-oauth-callbacks.md`, `docs/delivery/epics/epic-s3-day18-runtime-error-journal-and-staff-alert-center.md`, `docs/delivery/epics/epic-s3-day19-frontend-manual-qa-hardening-loop.md`, `docs/delivery/epics/epic-s3-day19.5-realtime-event-bus-and-websocket-backplane.md`, `docs/delivery/epics/epic-s3-day19.6-staff-realtime-subscriptions-and-ui.md`, `docs/delivery/epics/epic-s3-day20-e2e-regression-and-mvp-closeout.md`.
- Ключевые deliverables: полный stage label coverage, staff debug observability, MCP control tools, `run:self-improve`.
- Факт на 2026-02-19: Day1..Day16 и Day18 выполнены; Day9 закрыт как completed (full e2e часть вынесена в Day20). Sprint S3 в статусе in-progress: в очереди Day17/Day19/Day19.5/Day19.6 (core-flow + frontend hardening + realtime), финальный Day20 — full e2e regression + MVP closeout.

## Правила
- Если нет обязательного документа — статус `blocked`.
- Ссылки должны быть кликабельны.
- Матрица обновляется при каждом переходе этапа Delivery Plan.
