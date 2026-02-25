---
doc_id: EPC-CK8S-0005
type: epic
title: "Epic Catalog: Sprint S5 (Stage entry and label UX orchestration)"
status: in-progress
owner_role: EM
created_at: 2026-02-24
updated_at: 2026-02-25
related_issues: [154, 155, 170, 171]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-25-issue-170-epic-catalog"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-25
---

# Epic Catalog: Sprint S5 (Stage entry and label UX orchestration)

## TL;DR
- Sprint S5 фокусируется на управляемом UX запуска stage-процессов и снижении ручных ошибок в label-flow.
- Базовые deliverables: Day1 vision/prd/adr package + Day2 single-epic execution package для реализации.
- Текущий статус: Issue #155 закрыл Day1 governance baseline, Issue #170 зафиксировал Day2 handover, Issue #171 находится в `run:dev` реализации.

## Контекст
- Product source of truth: `docs/product/requirements_machine_driven.md` (FR-053, FR-054).
- Stage policy source of truth: `docs/product/labels_and_trigger_policy.md`, `docs/product/stage_process_model.md`.
- Delivery process source of truth: `docs/delivery/development_process_requirements.md`.

## Эпики Sprint S5
- Day 1: `docs/delivery/epics/s5/epic-s5-day1-launch-profiles-and-stage-launcher-ux.md`
- Day 1 PRD: `docs/delivery/epics/s5/prd-s5-day1-launch-profiles-and-stage-launcher-ux.md`
- Day 1 ADR: `docs/architecture/adr/ADR-0008-profile-driven-stage-launch-and-next-step-contract.md`
- Day 2: `docs/delivery/epics/s5/epic-s5-day2-launch-profiles-dev-execution.md`

## Delivery-governance пакет для Issue #155 (`run:plan`)

| Контур | Содержание | Статус |
|---|---|---|
| План исполнения | Декомпозиция I1..I5 (`P0/P1`) и role handover (`dev/qa/sre/km`) | ready |
| Quality-gates | QG-01..QG-05 (planning, contract, governance, traceability, review readiness) | QG-01..QG-05 passed |
| Acceptance | AC-01..AC-06 и `run:plan` acceptance criteria в Sprint S5 plan | ready |
| Traceability | Синхронизация `issue_map` и `requirements_traceability` под FR-053/FR-054 | ready |

## Delivery-governance пакет для Issue #170 (`run:plan`)

| Контур | Содержание | Статус |
|---|---|---|
| План исполнения | Single-epic execution модель: один epic + implementation issue #171 | ready |
| Quality-gates | QG-D2-01..QG-D2-05 (planning, contract, governance, traceability, readiness) | passed |
| Acceptance | AC-01..AC-06 + AC-D2-01..AC-D2-03 для single-epic handover | ready |
| Traceability | Day2 epic sync в `issue_map` и `requirements_traceability` | ready |

## Blockers, risks и owner decisions
- Blockers: `BLK-155-01`, `BLK-155-02` закрыты после Owner review в PR #166.
- Risks: `RSK-155-01` (comment overload), `RSK-155-02` (manual fallback без pre-check).
- Owner decisions: `OD-155-01..OD-155-03` утверждены (fast-track policy с guardrails, ambiguity hard-stop, dual review-gate).
- Day2 blockers: `BLK-170-01`, `BLK-170-02` закрыты (single-epic формат и обязательный pre-check закреплены).
- Day2 risks: `RSK-170-01` (review throughput), `RSK-170-02` (fallback/policy drift).
- Day2 owner decisions: `OD-170-01..OD-170-02` утверждены (one epic/one issue + статус согласовано для связанных docs).

## Критерии успеха Sprint S5 (выжимка)
- [x] Launch profiles покрывают минимум три сценария (`quick-fix`, `feature`, `new-service`) и имеют понятные guardrails.
- [x] Service-message next-step actions дают рабочий primary + fallback path.
- [x] Для Owner устранён ручной “по памяти” выбор порядка label-переходов.
- [x] Подготовлен owner-facing пакет quality-gates и критериев завершения перед `run:dev`.
- [x] Owner approval на запуск `run:dev` по Issue #155.
- [x] Подготовлен Day2 single-epic execution package по Issue #170.
- [x] Создана implementation issue #171 для реализации в одном delivery-контуре.
