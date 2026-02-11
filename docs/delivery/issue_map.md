---
doc_id: MAP-CK8S-0001
type: issue-map
title: "Issue ↔ Docs Map"
status: draft
owner_role: KM
created_at: 2026-02-06
updated_at: 2026-02-12
---

# Issue ↔ Docs Map

## TL;DR
Матрица трассируемости: Issue/PR ↔ документы ↔ релизы.

## Матрица
| Issue/PR | Traceability bundle | PRD | Design | ADRs | Test Plan | Release Notes | Postdeploy | Status |
|---|---|---|---|---|---|---|---|---|
| #1 | `docs/delivery/issue_map.md` + `docs/delivery/requirements_traceability.md` + `docs/delivery/sprint_s1_mvp_vertical_slice.md` + `docs/delivery/sprint_s2_dogfooding.md` | `docs/product/requirements_machine_driven.md` + `docs/product/brief.md` + `docs/product/constraints.md` + `docs/product/agents_operating_model.md` + `docs/product/labels_and_trigger_policy.md` + `docs/product/stage_process_model.md` | `docs/architecture/c4_context.md`, `docs/architecture/c4_container.md`, `docs/architecture/data_model.md`, `docs/architecture/api_contract.md`, `docs/architecture/agent_runtime_rbac.md`, `docs/architecture/mcp_approval_and_audit_flow.md`, `docs/architecture/prompt_templates_policy.md` | `ADR-0001..0004` | S2 dogfooding test plan pending | TBD | TBD | in-progress |

## Требования и трассировка
- Source of truth требований: `docs/product/requirements_machine_driven.md`.
- Матрица трассируемости: `docs/delivery/requirements_traceability.md`.

## Sprint S1 артефакты
- Process requirements: `docs/delivery/development_process_requirements.md`.
- Sprint plan: `docs/delivery/sprint_s1_mvp_vertical_slice.md`.
- Epic catalog: `docs/delivery/epic_s1.md`.
- Epic docs: `docs/delivery/epics/epic-s1-day0-bootstrap-baseline.md`, `docs/delivery/epics/epic-s1-day1-webhook-idempotency.md`, `docs/delivery/epics/epic-s1-day2-worker-slots-k8s.md`, `docs/delivery/epics/epic-s1-day3-auth-rbac-ui.md`, `docs/delivery/epics/epic-s1-day4-repository-provider.md`, `docs/delivery/epics/epic-s1-day5-learning-mode.md`, `docs/delivery/epics/epic-s1-day6-hardening-observability.md`, `docs/delivery/epics/epic-s1-day7-stabilization-gate.md`.

## Sprint S2 (план) артефакты
- Sprint plan: `docs/delivery/sprint_s2_dogfooding.md`.
- Epic catalog: `docs/delivery/epic_s2.md`.
- Epic docs: `docs/delivery/epics/epic-s2-day0-control-plane-extraction.md`, `docs/delivery/epics/epic-s2-day1-migrations-and-schema-ownership.md`, `docs/delivery/epics/epic-s2-day2-issue-label-triggers-run-dev.md`, `docs/delivery/epics/epic-s2-day3-per-issue-namespace-and-rbac.md`, `docs/delivery/epics/epic-s2-day4-agent-job-and-pr-flow.md`, `docs/delivery/epics/epic-s2-day5-staff-ui-dogfooding-observability.md`, `docs/delivery/epics/epic-s2-day6-approval-and-audit-hardening.md`, `docs/delivery/epics/epic-s2-day7-dogfooding-regression-gate.md`.
- Product process model docs: `docs/product/agents_operating_model.md`, `docs/product/labels_and_trigger_policy.md`, `docs/product/stage_process_model.md`.
- Day4 pre-implementation detail baseline: `docs/delivery/epics/epic-s2-day4-agent-job-and-pr-flow.md` (prompt/config/resume/session/PR flow + temporary decisions).
- Факт на 2026-02-11: Day0..Day3 приняты Owner (`approved`), активный фронт работ — Day4.

## Правила
- Если нет обязательного документа — статус `blocked`.
- Ссылки должны быть кликабельны.
- Матрица обновляется при каждом переходе этапа Delivery Plan.
