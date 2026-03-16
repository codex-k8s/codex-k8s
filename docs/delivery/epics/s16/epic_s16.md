---
doc_id: EPC-CK8S-0016
type: epic
title: "Epic Catalog: Sprint S16 (Mission Control graph workspace and continuity control plane)"
status: in-review
owner_role: PM
created_at: 2026-03-15
updated_at: 2026-03-16
related_issues: [480, 490, 492, 496, 510, 516, 519]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-15-issue-492-intake"
---

# Epic Catalog: Sprint S16 (Mission Control graph workspace and continuity control plane)

## TL;DR
- Sprint S16 открывает новый product stream вокруг полного redesign Mission Control: staff console должна стать graph workspace/control plane, а не оставаться board/list слоем над частичной platform evidence.
- Day1 intake (`#492`) зафиксировал fullscreen workspace, hybrid truth matrix, filtered multi-root graph с точными Wave 1 filters `open_only`, `assigned_to_me_or_unassigned` и active-state presets, закрытый Wave 1 node set `discussion/work_item/run/pull_request`, platform-canonical metadata/watermarks/launch params и continuity rule `PR + follow-up issue`.
- Day2 vision (`#496`) зафиксировал mission, north star, persona outcomes, KPI/guardrails и wave boundaries для primary multi-root graph workspace/control plane, не переоткрывая intake baseline.
- Issue `#480` не исчезает, а становится обязательным foundation layer внутри S16: persisted GitHub inventory mirror и bounded reconcile с coverage contract `all open Issues/PR + bounded recent closed history` должны питать новый workspace.
- Day3 PRD (`#510`) зафиксировал user stories, FR/AC/NFR, scenario matrix, expected evidence и boundary core Wave 1 vs deferred contours; создана follow-up issue `#516` для `run:arch`.
- Day4 Architecture (`#516`) зафиксировал ownership split, hybrid truth lifecycle `provider mirror -> graph truth -> workspace projection`, continuity gaps и bounded provider foundation; создана follow-up issue `#519` для `run:design`.
- До `run:plan` Sprint S16 остаётся markdown-only и не создаёт implementation issues.

## Stage roadmap
- Day 1 (Intake): `docs/delivery/epics/s16/epic-s16-day1-mission-control-graph-workspace-intake.md` (Issue `#492`).
- Day 2 (Vision): `docs/delivery/epics/s16/epic-s16-day2-mission-control-graph-workspace-vision.md` (Issue `#496`).
- Day 3 (PRD): `docs/delivery/epics/s16/epic-s16-day3-mission-control-graph-workspace-prd.md` + `docs/delivery/epics/s16/prd-s16-day3-mission-control-graph-workspace.md` (Issue `#510`); user stories, FR/AC/NFR, scenario matrix и expected evidence уже зафиксированы для graph workspace, continuity graph и foundation inventory.
- Day 4 (Architecture): `docs/delivery/epics/s16/epic-s16-day4-mission-control-graph-workspace-arch.md` + architecture package in `docs/architecture/initiatives/s16_mission_control_graph_workspace/` (Issue `#516`); зафиксированы `control-plane`-owned graph truth, bounded `worker` inventory foundation, hybrid truth lifecycle и continuity-gap ownership.
- Day 5 (Design): Issue `#519`; ожидается implementation-ready design package по typed metadata, graph surfaces, transport/data contracts и rollout notes.
- Day 6 (Plan): `TBD`; ожидается execution package с delivery waves, quality-gates и owner-managed handover в `run:dev`.

## Delivery-governance rules
- Sprint S16 стартует только полным doc-stage контуром `intake -> vision -> prd -> arch -> design -> plan`.
- Каждый stage обязан создавать следующую follow-up issue без trigger-лейбла; trigger следующего запуска остаётся owner-managed.
- До `run:plan` Sprint S16 не создаёт implementation issues и не открывает code/runtime changes.
- Sprint S16 обязан сохранить three-way continuity:
  - owner redesign request из `#490`;
  - inventory foundation stream из `#480`;
  - existing execution/reference baseline из Sprint S9.
- Voice/STT, dashboard orchestrator agent и richer provider enrichment считаются later-wave scope и не могут становиться blocking condition для core Wave 1.
- Human review, merge и provider-native collaboration остаются в GitHub UI; dashboard не подменяет provider semantics.
