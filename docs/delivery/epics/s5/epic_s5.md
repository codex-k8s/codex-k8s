---
doc_id: EPC-CK8S-0005
type: epic
title: "Epic Catalog: Sprint S5 (Stage entry and label UX orchestration)"
status: in-progress
owner_role: EM
created_at: 2026-02-24
updated_at: 2026-02-25
related_issues: [154, 155]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-24-issue-155-epic-catalog"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-25
---

# Epic Catalog: Sprint S5 (Stage entry and label UX orchestration)

## TL;DR
- Sprint S5 фокусируется на управляемом UX запуска stage-процессов и снижении ручных ошибок в label-flow.
- Базовый deliverable: Day1 execution package для launch profiles и deterministic next-step actions.
- Текущий статус Day1: для Issue #155 подготовлен `run:plan` пакет перед `run:dev` (quality-gates, критерии завершения, блокеры/риски/owner decisions).

## Контекст
- Product source of truth: `docs/product/requirements_machine_driven.md` (FR-053, FR-054).
- Stage policy source of truth: `docs/product/labels_and_trigger_policy.md`, `docs/product/stage_process_model.md`.
- Delivery process source of truth: `docs/delivery/development_process_requirements.md`.

## Эпики Sprint S5
- Day 1: `docs/delivery/epics/s5/epic-s5-day1-launch-profiles-and-stage-launcher-ux.md`
- Day 1 PRD: `docs/delivery/epics/s5/prd-s5-day1-launch-profiles-and-stage-launcher-ux.md`
- Day 1 ADR: `docs/architecture/adr/ADR-0008-profile-driven-stage-launch-and-next-step-contract.md`

## Delivery-governance пакет для Issue #155 (`run:plan`)

| Контур | Содержание | Статус |
|---|---|---|
| План исполнения | Декомпозиция I1..I5 (`P0/P1`) и role handover (`dev/qa/sre/km`) | ready |
| Quality-gates | QG-01..QG-05 (planning, contract, governance, traceability, review readiness) | QG-01..QG-05 passed |
| Acceptance | AC-01..AC-06 и `run:plan` acceptance criteria в Sprint S5 plan | ready |
| Traceability | Синхронизация `issue_map` и `requirements_traceability` под FR-053/FR-054 | ready |

## Blockers, risks и owner decisions
- Blockers: `BLK-155-01`, `BLK-155-02` закрыты после Owner review в PR #166.
- Risks: `RSK-155-01` (comment overload), `RSK-155-02` (manual fallback без pre-check).
- Owner decisions: `OD-155-01..OD-155-03` утверждены (fast-track policy с guardrails, ambiguity hard-stop, dual review-gate).

## Критерии успеха Sprint S5 (выжимка)
- [x] Launch profiles покрывают минимум три сценария (`quick-fix`, `feature`, `new-service`) и имеют понятные guardrails.
- [x] Service-message next-step actions дают рабочий primary + fallback path.
- [x] Для Owner устранён ручной “по памяти” выбор порядка label-переходов.
- [x] Подготовлен owner-facing пакет quality-gates и критериев завершения перед `run:dev`.
- [x] Owner approval на запуск `run:dev` по Issue #155.
