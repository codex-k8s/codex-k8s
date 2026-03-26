---
doc_id: EPC-CK8S-0016
type: epic
title: "Epic Catalog: Sprint S16 (Mission Control graph workspace and continuity control plane)"
status: superseded
owner_role: PM
created_at: 2026-03-15
updated_at: 2026-03-25
related_issues: [480, 490, 492, 496, 510, 516, 519, 537, 542, 543, 544, 545, 546, 547, 561, 562, 563]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-15-issue-492-intake"
---

# Epic Catalog: Sprint S16 (historical superseded Mission Control baseline)

## TL;DR
- 2026-03-25 issue `#561` зафиксировала, что Sprint S16 больше не является текущим source of truth для Mission Control.
- Day1-Day6 package и issues `#542..#547` сохранены только как historical evidence отклонённого baseline.
- Актуальный Mission Control reset перенесён в `#562` (frontend-first UX на fake data) и `#563` (backend rebuild после owner approval UX).
- Новый agreed baseline: fullscreen canvas без lane/column shell, node taxonomy `Issue/PR/Run`, repo-seed prompts без DB prompt editor и `stale/freshness` только как доказанный lag provider mirror/reconcile path.

## Superseded Scope
- superseded UX shell: lane/column workspace и обязательная root-group hierarchy;
- superseded Wave 1 taxonomy: `discussion/work_item/run/pull_request`;
- superseded execution path: `#542..#547` как следующий обязательный `run:dev` contour;
- superseded semantics: старые freshness/watermark UX как центральный экранный статус.

## Stage roadmap
- Day 1 (Intake): `docs/delivery/epics/s16/epic-s16-day1-mission-control-graph-workspace-intake.md` (Issue `#492`).
- Day 2 (Vision): `docs/delivery/epics/s16/epic-s16-day2-mission-control-graph-workspace-vision.md` (Issue `#496`).
- Day 3 (PRD): `docs/delivery/epics/s16/epic-s16-day3-mission-control-graph-workspace-prd.md` + `docs/delivery/epics/s16/prd-s16-day3-mission-control-graph-workspace.md` (Issue `#510`); user stories, FR/AC/NFR, scenario matrix и expected evidence уже зафиксированы для graph workspace, continuity graph и foundation inventory.
- Day 4 (Architecture): `docs/delivery/epics/s16/epic-s16-day4-mission-control-graph-workspace-arch.md` + architecture package in `docs/architecture/initiatives/s16_mission_control_graph_workspace/` (Issue `#516`); зафиксированы `control-plane`-owned graph truth, bounded `worker` inventory foundation, hybrid truth lifecycle и continuity-gap ownership.
- Day 5 (Design): `docs/delivery/epics/s16/epic-s16-day5-mission-control-graph-workspace-design.md` + `docs/architecture/initiatives/s16_mission_control_graph_workspace/{design_doc.md,api_contract.md,data_model.md,migrations_policy.md}` (Issue `#519`); зафиксированы graph-first UX, typed snapshot/node/preview contracts, continuity-gap storage и rollout notes.
- Day 6 (Plan): `docs/delivery/epics/s16/epic-s16-day6-mission-control-graph-workspace-plan.md` (Issue `#537`); execution package зафиксирован через issues `#542..#547`.

## Delivery-governance rules
- Sprint S16 стартует только полным doc-stage контуром `intake -> vision -> prd -> arch -> design -> plan`.
- Каждый stage обязан создавать следующую follow-up issue без trigger-лейбла; trigger следующего запуска остаётся owner-managed.
- До `run:plan` Sprint S16 не создаёт implementation issues и не открывает code/runtime changes; после `run:plan` execution backlog запускается только по owner-managed waves `#542..#547`.
- Sprint S16 обязан сохранить three-way continuity:
  - owner redesign request из `#490`;
  - inventory foundation stream из `#480`;
  - existing execution/reference baseline из Sprint S9.
- Voice/STT, dashboard orchestrator agent и richer provider enrichment считаются later-wave scope и не могут становиться blocking condition для core Wave 1.
- Human review, merge и provider-native collaboration остаются в GitHub UI; dashboard не подменяет provider semantics.
