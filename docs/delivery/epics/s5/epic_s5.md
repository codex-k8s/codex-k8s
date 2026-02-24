---
doc_id: EPC-CK8S-0005
type: epic
title: "Epic Catalog: Sprint S5 (Stage entry and label UX orchestration)"
status: planned
owner_role: EM
created_at: 2026-02-24
updated_at: 2026-02-24
related_issues: [154]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-24-issue-154-epic-catalog"
---

# Epic Catalog: Sprint S5 (Stage entry and label UX orchestration)

## TL;DR
- Sprint S5 фокусируется на управляемом UX запуска stage-процессов и снижении ручных ошибок в label-flow.
- Базовый deliverable: Day1 execution package для launch profiles и deterministic next-step actions.

## Контекст
- Product source of truth: `docs/product/requirements_machine_driven.md` (FR-053, FR-054).
- Stage policy source of truth: `docs/product/labels_and_trigger_policy.md`, `docs/product/stage_process_model.md`.
- Delivery process source of truth: `docs/delivery/development_process_requirements.md`.

## Эпики Sprint S5
- Day 1: `docs/delivery/epics/s5/epic-s5-day1-launch-profiles-and-stage-launcher-ux.md`

## Критерии успеха Sprint S5 (выжимка)
- [ ] Launch profiles покрывают минимум три сценария (`quick-fix`, `feature`, `new-service`) и имеют понятные guardrails.
- [ ] Service-message next-step actions дают рабочий primary + fallback path.
- [ ] Для Owner устранён ручной “по памяти” выбор порядка label-переходов.
