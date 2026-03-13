---
doc_id: EPC-CK8S-INDEX-0001
type: epic-index
title: "Epic Index (grouped by sprint)"
status: active
owner_role: EM
created_at: 2026-02-24
updated_at: 2026-03-13
related_issues: [112, 154, 184, 185, 187, 189, 195, 197, 199, 201, 212, 218, 220, 222, 238, 241, 243, 244, 245, 246, 247, 248, 249, 250, 251, 252, 253, 254, 255, 256, 257, 258, 259, 260, 216, 262, 263, 265, 281, 282, 320, 327, 333, 335, 337, 340, 351, 360, 363, 369, 370, 371, 372, 373, 374, 375, 378, 383, 385, 387, 389, 391, 392, 393, 394, 395]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-24-epic-index"
---

# Epic Index

## TL;DR
- Все day-эпики сгруппированы по папкам спринтов (`s1`, `s2`, `s3`, `s4`, `s5`, `s6`, `s7`, `s8`, `s9`, `s10`).
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

## Проверка консистентности
- Для каждого `s<номер>` должны существовать:
  - `epic_s<номер>.md`;
  - минимум один `epic-s<номер>-day*.md`.
- Ссылки на day-эпики не должны указывать на корень `docs/delivery/epics/` без `s<номер>/`.
- Исторические обновления по issue не дублируются в epic index и выносятся в `docs/delivery/traceability/s<номер>_*.md`.
