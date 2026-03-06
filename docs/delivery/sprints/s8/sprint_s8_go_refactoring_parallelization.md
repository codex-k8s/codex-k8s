---
doc_id: SPR-CK8S-0008
type: sprint-plan
title: "Sprint S8: Go refactoring parallelization + repository onboarding automation"
status: in-progress
owner_role: EM
created_at: 2026-02-27
updated_at: 2026-03-06
related_issues: [223, 225, 226, 227, 228, 229, 230, 281, 282]
related_prs: [231]
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-27-issue-223-plan-revise"
---

# Sprint S8: Go refactoring parallelization + repository onboarding automation

## TL;DR
- Sprint S8 стартовал как отдельный поток Go-рефакторинга, изолированный от Sprint S7.
- В текущей ревизии Sprint S8 расширен ещё двумя P0 onboarding-потоками: инициализация пустого репозитория и adoption существующего репозитория с кодом.
- Цель спринта: параллельно закрыть инженерный долг по Go-слою и убрать ручной bootstrap при подключении новых проектных репозиториев.

## Stage roadmap
- Day 1 (Plan): `docs/delivery/epics/s8/epic-s8-day1-go-refactoring-plan.md` (Issue `#223`).
- Day 2 (Execution): `docs/delivery/epics/s8/epic-s8-day2-empty-repository-initialization.md` (Issue `#281`).
- Day 3 (Execution): `docs/delivery/epics/s8/epic-s8-day3-existing-repository-adoption.md` (Issue `#282`).
- Day 2+ (Execution): `run:dev -> run:qa -> run:release` для задач `#225..#230`, `#281`, `#282`.

## Handover
- Next stage: `run:dev` по задачам `#225..#230`, `#281`, `#282`.
- Гейт перехода: review/approve plan-артефакта Sprint S8 Owner'ом.
