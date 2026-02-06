---
doc_id: MAP-CK8S-0001
type: issue-map
title: "Issue ↔ Docs Map"
status: draft
owner_role: KM
created_at: 2026-02-06
updated_at: 2026-02-06
---

# Issue ↔ Docs Map

## TL;DR
Матрица трассируемости: Issue/PR ↔ документы ↔ релизы.

## Матрица
| Issue/PR | DocSet | PRD | Design | ADRs | Test Plan | Release Notes | Postdeploy | Status |
|---|---|---|---|---|---|---|---|---|
| #1 | `docs/_docset/issues/issue-0001-codex-k8s-bootstrap.md` | `docs/product/requirements_machine_driven.md` + `docs/product/brief.md` + `docs/product/constraints.md` | `docs/architecture/c4_context.md`, `docs/architecture/c4_container.md`, `docs/architecture/data_model.md`, `docs/architecture/api_contract.md` | `ADR-0001..0004` | learning mode smoke TBD | TBD | TBD | in-progress |

## Требования и трассировка
- Source of truth требований: `docs/product/requirements_machine_driven.md`.
- Матрица трассируемости: `docs/delivery/requirements_traceability.md`.

## Sprint S1 артефакты
- Process requirements: `docs/delivery/development_process_requirements.md`.
- Sprint plan: `docs/delivery/sprint_s1_day0_day7.md`.
- Epic catalog: `docs/delivery/epic.md`.
- Epic docs: `docs/delivery/epics/epic-day0-bootstrap-baseline.md`, `docs/delivery/epics/epic-day1-webhook-idempotency.md`, `docs/delivery/epics/epic-day2-worker-slots-k8s.md`, `docs/delivery/epics/epic-day3-auth-rbac-ui.md`, `docs/delivery/epics/epic-day4-repository-provider.md`, `docs/delivery/epics/epic-day5-learning-mode.md`, `docs/delivery/epics/epic-day6-hardening-observability.md`, `docs/delivery/epics/epic-day7-stabilization-gate.md`.

## Правила
- Если нет обязательного документа — статус `blocked`.
- Ссылки должны быть кликабельны.
- Матрица обновляется при каждом переходе этапа Delivery Plan.
