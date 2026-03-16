---
doc_id: IDX-CK8S-ARCH-S16-0001
type: initiative-index
title: "Initiative Package: s16_mission_control_graph_workspace"
status: in-review
owner_role: SA
created_at: 2026-03-16
updated_at: 2026-03-16
related_issues: [480, 490, 492, 496, 510, 516, 519]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-16-issue-516-arch"
---

# s16_mission_control_graph_workspace

## TL;DR
- Пакет объединяет Day4 architecture-артефакты Sprint S16 для Mission Control graph workspace и continuity control plane.
- Внутри зафиксированы ownership split для `control-plane` / `worker` / `api-gateway` / `web-console` / `agent-runner`, hybrid truth model между platform graph state и GitHub inventory mirror, а также Day4 ADR/alternatives по continuity surfaces и bounded provider foundation.
- Следующий обязательный этап после review этого пакета: owner-managed issue `#519` для `run:design`, где baseline Day4 должен быть переведён в typed transport/data/migration contracts без reopening product scope.

## Содержимое
- `docs/architecture/initiatives/s16_mission_control_graph_workspace/README.md`
- `docs/architecture/initiatives/s16_mission_control_graph_workspace/architecture.md`
- `docs/architecture/initiatives/s16_mission_control_graph_workspace/c4_context.md`
- `docs/architecture/initiatives/s16_mission_control_graph_workspace/c4_container.md`

## Связанные source-of-truth документы
- `docs/architecture/adr/ADR-0016-mission-control-graph-workspace-hybrid-truth-and-continuity-ownership.md`
- `docs/architecture/alternatives/ALT-0008-mission-control-graph-workspace-hybrid-truth-boundaries.md`
- `docs/delivery/epics/s16/epic-s16-day4-mission-control-graph-workspace-arch.md`
- `docs/delivery/epics/s16/epic-s16-day3-mission-control-graph-workspace-prd.md`
- `docs/delivery/epics/s16/prd-s16-day3-mission-control-graph-workspace.md`
- `docs/delivery/sprints/s16/sprint_s16_mission_control_graph_workspace.md`
- `docs/delivery/traceability/s16_mission_control_graph_workspace_history.md`
- `docs/architecture/initiatives/s9_mission_control_dashboard/README.md`

## Continuity after `run:arch`
- Документный контур `intake -> vision -> prd -> arch` согласован и доведён до review-ready architecture package.
- Day4 зафиксировал:
  - `control-plane` как owner canonical graph truth, continuity state, typed metadata/watermarks и launch surfaces;
  - `worker` как owner bounded inventory freshness, enrichment/reconcile execution и lifecycle/background tasks без ownership graph semantics;
  - `api-gateway` и `web-console` как thin-edge/visibility surfaces без локального hybrid truth merge;
  - inventory foundation issue `#480`, exact Wave 1 filters `open_only`, `assigned_to_me_or_unassigned`, `active-state presets`, secondary/dimmed semantics только для graph integrity, nodes `discussion/work_item/run/pull_request`, platform-canonical launch params и continuity rule `PR + linked follow-up issue`.
- Sprint S16 Day5 (`#519`) остаётся downstream design stage и может наследовать только зафиксированные ownership boundaries, hybrid truth policy и deferred-scope guardrails Day4.
