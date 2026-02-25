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
  status: pending
  request_id: "owner-2026-02-24-issue-155-epic-catalog"
---

# Epic Catalog: Sprint S5 (Stage entry and label UX orchestration)

## TL;DR
- Sprint S5 фокусируется на управляемом UX запуска stage-процессов и снижении ручных ошибок в label-flow.
- Базовый deliverable: Day1 execution package для launch profiles и deterministic next-step actions.
- Текущий статус Day1: vision/prd пакет сформирован в Issue #155 и передан в Owner review.

## Контекст
- Product source of truth: `docs/product/requirements_machine_driven.md` (FR-053, FR-054).
- Stage policy source of truth: `docs/product/labels_and_trigger_policy.md`, `docs/product/stage_process_model.md`.
- Delivery process source of truth: `docs/delivery/development_process_requirements.md`.

## Эпики Sprint S5
- Day 1: `docs/delivery/epics/s5/epic-s5-day1-launch-profiles-and-stage-launcher-ux.md`
- Day 1 PRD: `docs/delivery/epics/s5/prd-s5-day1-launch-profiles-and-stage-launcher-ux.md`

## Критерии успеха Sprint S5 (выжимка)
- [x] Launch profiles покрывают минимум три сценария (`quick-fix`, `feature`, `new-service`) и имеют понятные guardrails.
- [x] Service-message next-step actions дают рабочий primary + fallback path.
- [x] Для Owner устранён ручной “по памяти” выбор порядка label-переходов.
- [ ] Owner approval на запуск `run:dev` по Issue #155.
