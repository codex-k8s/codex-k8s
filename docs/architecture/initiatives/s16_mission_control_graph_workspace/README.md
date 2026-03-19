---
doc_id: IDX-CK8S-ARCH-S16-0001
type: initiative-index
title: "Initiative Package: s16_mission_control_graph_workspace"
status: in-review
owner_role: SA
created_at: 2026-03-16
updated_at: 2026-03-19
related_issues: [480, 490, 492, 496, 510, 516, 519, 537, 546, 547]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-16-issue-516-arch"
---

# s16_mission_control_graph_workspace

## TL;DR
- Пакет объединяет Day4-Day5 architecture/design артефакты Sprint S16 для Mission Control graph workspace и continuity control plane.
- Внутри зафиксированы ownership split для `control-plane` / `worker` / `api-gateway` / `web-console` / `agent-runner`, graph-first interaction model, typed transport/data/migration contracts и rollout notes.
- Следующий обязательный этап после review этого пакета: owner-managed issue `#537` для `run:plan`, где baseline Day5 должен быть разложен в execution waves без reopening product scope.

## Содержимое
- `docs/architecture/initiatives/s16_mission_control_graph_workspace/README.md`
- `docs/architecture/initiatives/s16_mission_control_graph_workspace/architecture.md`
- `docs/architecture/initiatives/s16_mission_control_graph_workspace/c4_context.md`
- `docs/architecture/initiatives/s16_mission_control_graph_workspace/c4_container.md`
- `docs/architecture/initiatives/s16_mission_control_graph_workspace/design_doc.md`
- `docs/architecture/initiatives/s16_mission_control_graph_workspace/api_contract.md`
- `docs/architecture/initiatives/s16_mission_control_graph_workspace/data_model.md`
- `docs/architecture/initiatives/s16_mission_control_graph_workspace/migrations_policy.md`
- `docs/architecture/initiatives/s16_mission_control_graph_workspace/ui_smoke_criteria.md`

## Связанные source-of-truth документы
- `docs/architecture/adr/ADR-0016-mission-control-graph-workspace-hybrid-truth-and-continuity-ownership.md`
- `docs/architecture/alternatives/ALT-0008-mission-control-graph-workspace-hybrid-truth-boundaries.md`
- `docs/delivery/epics/s16/epic-s16-day4-mission-control-graph-workspace-arch.md`
- `docs/delivery/epics/s16/epic-s16-day5-mission-control-graph-workspace-design.md`
- `docs/delivery/epics/s16/epic-s16-day3-mission-control-graph-workspace-prd.md`
- `docs/delivery/epics/s16/prd-s16-day3-mission-control-graph-workspace.md`
- `docs/delivery/sprints/s16/sprint_s16_mission_control_graph_workspace.md`
- `docs/delivery/traceability/s16_mission_control_graph_workspace_history.md`
- `docs/architecture/initiatives/s9_mission_control_dashboard/README.md`
- `docs/architecture/initiatives/s16_mission_control_graph_workspace/ui_smoke_criteria.md`

## Continuity after `run:design`
- Документный контур `intake -> vision -> prd -> arch -> design` согласован и доведён до review-ready design package.
- Day5 зафиксировал:
  - graph-first workspace поверх Day4 ownership split без нового deployable сервиса;
  - сохранение одного command ledger для platform-safe actions и отдельного read-only launch preview;
  - persisted continuity gaps и workspace watermarks как domain constructs `control-plane`;
  - эволюционный rollout path `schema/backfill -> control-plane -> worker -> api-gateway -> web-console`;
  - inventory foundation issue `#480`, exact Wave 1 filters `open_only`, `assigned_to_me_or_unassigned`, `active-state presets`, secondary/dimmed semantics только для graph integrity, nodes `discussion/work_item/run/pull_request`, platform-canonical launch params и continuity rule `PR + linked follow-up issue`.
- Sprint S16 Day6 остаётся downstream `run:plan` stage и может наследовать только зафиксированные transport/data/migration contracts Day5.
- Wave `S16-E05` добавляет UI-level smoke criteria в `ui_smoke_criteria.md`, чтобы readiness gate `#547` проверял graph workspace, continuity visibility и read-only preview по одинаковому evidence contract.
