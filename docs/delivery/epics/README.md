---
doc_id: EPC-CK8S-INDEX-0001
type: epic-index
title: "Epic Index (grouped by sprint)"
status: active
owner_role: EM
created_at: 2026-02-24
updated_at: 2026-03-26
related_issues: [112, 154, 184, 185, 187, 189, 195, 197, 199, 201, 212, 218, 220, 222, 238, 241, 243, 244, 245, 246, 247, 248, 249, 250, 251, 252, 253, 254, 255, 256, 257, 258, 259, 260, 216, 262, 263, 265, 281, 282, 320, 327, 333, 335, 337, 340, 351, 360, 361, 363, 366, 369, 370, 371, 372, 373, 374, 375, 378, 383, 385, 387, 389, 391, 392, 393, 394, 395, 413, 416, 418, 444, 447, 448, 452, 454, 469, 471, 476, 480, 484, 490, 492, 494, 496, 510, 537, 541, 542, 543, 544, 545, 546, 547, 554, 557, 559, 561, 562, 563, 565]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-24-epic-index"
---

# Epic Index

## TL;DR
- Все day-эпики сгруппированы по папкам спринтов `s<номер>`.
- Каталог эпиков каждого спринта размещён в той же папке.
- Это устраняет смешение day-эпиков разных спринтов в одном каталоге.
- Epic index хранит только структуру каталогов и ссылки на sprint/day-epic документы.
- Historical issue-evidence размещается в `docs/delivery/traceability/s<номер>_*.md` и не дублируется в этом индексе.

## Структура

| Папка | Содержимое | Основной документ |
|---|---|---|
| `docs/delivery/epics/s1/` | Day0..Day7 + каталог Sprint S1 | `docs/delivery/epics/s1/epic_s1.md` |
| `docs/delivery/epics/s2/` | Day0..Day7 (+ Day3.5/Day4.5) + каталог Sprint S2 | `docs/delivery/epics/s2/epic_s2.md` |
| `docs/delivery/epics/s3/` | Day1..Day20 + каталог Sprint S3 | `docs/delivery/epics/s3/epic_s3.md` |
| `docs/delivery/epics/s4/` | Day1 + каталог Sprint S4 | `docs/delivery/epics/s4/epic_s4.md` |
| `docs/delivery/epics/s5/` | Day1 + каталог Sprint S5 | `docs/delivery/epics/s5/epic_s5.md` |
| `docs/delivery/epics/s6/` | Day1..Day11 (intake + vision + PRD package + architecture + design + plan + release closeout + postdeploy review + ops closeout) + каталог Sprint S6 | `docs/delivery/epics/s6/epic_s6.md` |
| `docs/delivery/epics/s7/` | Day1 intake + Day2 vision + Day3 PRD package (`epic + prd`) + Day4 architecture package (`epic + C4 overlays + ADR/alternatives`) + Day5 design package (`epic + design_doc/api_contract/data_model/migrations_policy`) + Day6 plan package (`epic-s7-day6-mvp-readiness-plan.md`) c execution issues `#243..#260` по `S7-E01..S7-E18`, включая MVP policy `repo-only` по prompt templates | `docs/delivery/epics/s7/epic_s7.md` |
| `docs/delivery/epics/s8/` | Day1 plan + Day2/Day3 onboarding эпики + Day4 documentation IA execution/result (`docs/index.md`, domain `README.md`, migration-map, repo-local refs remediation) + каталог Sprint S8 (Go refactoring + repository onboarding automation) + execution backlog `S8-E01..S8-E09` | `docs/delivery/epics/s8/epic_s8.md` |
| `docs/delivery/epics/s9/` | Day1 intake + Day2 vision + Day3 PRD package (`epic + prd`) + Day4 architecture package (`epic + C4 overlays + ADR/alternatives`) + Day5 design package (`epic + design_doc/api_contract/data_model/migrations_policy`) + Day6 plan package (`epic-s9-day6-mission-control-dashboard-plan.md`) с execution issues `#369..#375` для Mission Control Dashboard и каталога Sprint S9 (control-plane UX, active-set dashboard, discussion-first flow, realtime/provider reconciliation) | `docs/delivery/epics/s9/epic_s9.md` |
| `docs/delivery/epics/s10/` | Day1 intake + Day2 vision + Day3 PRD package (`epic + prd`) + Day4 architecture package (`epic + C4 overlays + ADR/alternatives`) + Day5 design package + Day6 plan package (`epic-s10-day6-mcp-user-interactions-plan.md`) по built-in MCP user interactions; execution backlog разложен на issues `#391..#395` для `control-plane`, `worker`, `api-gateway`, `agent-runner` и observability | `docs/delivery/epics/s10/epic_s10.md` |
| `docs/delivery/epics/s11/` | Day1 intake + Day2 vision + Day3 PRD package (`epic + prd`) + Day4 architecture package (`epic + C4 overlays + ADR/alternatives`) для Telegram-адаптера как первого внешнего channel path; continuity issue `#444` 2026-03-14 закрыта как `state:superseded`, architecture stage выполнен в `#452`, а issue `#454` создана для `run:design` | `docs/delivery/epics/s11/epic_s11.md` |
| `docs/delivery/epics/s12/` | Day1 intake + Day2 vision + Day3 PRD package (`epic + prd`) для GitHub API rate-limit resilience; continuity после PRD переведена в architecture issue `#418`, а следующие day-эпики создаются последовательно после owner review | `docs/delivery/epics/s12/epic_s12.md` |
| `docs/delivery/epics/s13/` | Day1 intake + Day2 vision + Day3 PRD package (`epic + prd`) + Day4 architecture package (`epic + C4 overlays + ADR/alternatives`) для `Quality Governance System`; architecture stage в issue `#484` закрепил canonical governance ownership, publication discipline и создал issue `#494` для `run:design` | `docs/delivery/epics/s13/epic_s13.md` |
| `docs/delivery/epics/s16/` | Historical superseded package Sprint S16: 2026-03-25 issue `#561` зафиксировала, что baseline `discussion/work_item/run/pull_request`, lane/column shell и execution handover `#542..#547` больше не являются текущим Mission Control source of truth. Эти day-эпики и PRD сохранены как evidence отклонённого baseline; активный reset path вынесен в issues `#562` (frontend-first UX) и `#563` (backend rebuild после approval UX). | `docs/delivery/epics/s16/epic_s16.md` |
| `docs/delivery/epics/s17/` | Day1 intake + Day2 vision + Day3 PRD package (`epic + prd`) для unified long-lived user interaction waits и owner feedback inbox; issue `#541` закрепила same live session как primary happy-path, 24h long wait, delivery-before-wait lifecycle, Telegram pending inbox, staff-console fallback и persisted text/voice binding, issue `#554` оформила mission/KPI/guardrails, а issue `#557` зафиксировала user stories/FR/AC/NFR, scenario matrix, expected evidence и создала issue `#559` для `run:arch` | `docs/delivery/epics/s17/epic_s17.md` |
| `docs/delivery/epics/s18/` | Day1 intake для frontend-first Mission Control reset на fake data: issue `#562` зафиксировала fullscreen canvas, taxonomy `Issue/PR/Run`, compact nodes, workflow editor UX, platform-safe actions only и isolated `web-console` prototype как цель `run:dev`; continuity issue `#565` создана для `run:vision`, а backend rebuild остаётся отдельной задачей `#563` после owner approval UX. | `docs/delivery/epics/s18/epic_s18.md` |

## Проверка консистентности
- Для каждого `s<номер>` должны существовать:
  - `epic_s<номер>.md`;
  - минимум один `epic-s<номер>-day*.md`.
- Ссылки на day-эпики не должны указывать на корень `docs/delivery/epics/` без `s<номер>/`.
- Исторические обновления по issue не дублируются в epic index и выносятся в `docs/delivery/traceability/s<номер>_*.md`.
