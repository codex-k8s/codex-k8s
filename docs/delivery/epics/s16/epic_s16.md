---
doc_id: EPC-CK8S-0016
type: epic
title: "Epic Catalog: Sprint S16 (Mission Control graph workspace and continuity control plane)"
status: in-review
owner_role: PM
created_at: 2026-03-15
updated_at: 2026-03-15
related_issues: [480, 490, 492, 496]
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
- Issue `#480` не исчезает, а становится обязательным foundation layer внутри S16: persisted GitHub inventory mirror и bounded reconcile с coverage contract `all open Issues/PR + bounded recent closed history` должны питать новый workspace.
- Создана continuity issue `#496` для stage `run:vision`; до `run:plan` Sprint S16 остаётся markdown-only и не создаёт implementation issues.

## Stage roadmap
- Day 1 (Intake): `docs/delivery/epics/s16/epic-s16-day1-mission-control-graph-workspace-intake.md` (Issue `#492`).
- Day 2 (Vision): Issue `#496`; ожидается mission, north star, persona outcomes, KPI/guardrails и wave framing без reopening intake decisions.
- Day 3 (PRD): `TBD`; ожидается user stories, FR/AC/NFR и scenario matrix для graph workspace, continuity graph и foundation inventory.
- Day 4 (Architecture): `TBD`; ожидается ownership split по graph truth matrix, provider mirror, run continuity и service boundaries.
- Day 5 (Design): `TBD`; ожидается implementation-ready design package по typed metadata, graph surfaces, transport/data contracts и rollout notes.
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
