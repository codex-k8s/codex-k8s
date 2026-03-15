---
doc_id: TRH-CK8S-S16-0001
type: traceability-history
title: "Sprint S16 Traceability History"
status: in-review
owner_role: KM
created_at: 2026-03-15
updated_at: 2026-03-16
related_issues: [480, 490, 492, 496, 510]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-15-traceability-s16-history"
---

# Sprint S16 Traceability History

## TL;DR
- Этот файл хранит historical delta для Sprint S16.
- Текущая master-карта связей остаётся в `docs/delivery/issue_map.md`.
- Текущее покрытие FR/NFR остаётся в `docs/delivery/requirements_traceability.md`.

## Актуализация по Issue #492 (`run:intake`, 2026-03-15)
- Подготовлен intake package:
  - `docs/delivery/sprints/s16/sprint_s16_mission_control_graph_workspace.md`;
  - `docs/delivery/epics/s16/epic_s16.md`;
  - `docs/delivery/epics/s16/epic-s16-day1-mission-control-graph-workspace-intake.md`.
- Зафиксированы:
  - Sprint S16 как полный redesign Mission Control в graph workspace/control plane, а не как incremental tuning Sprint S9 dashboard;
  - поглощение issue `#480` как mandatory foundation layer для persisted provider mirror, bounded reconcile и coverage contract `all open Issues/PR + bounded recent closed history`;
  - hybrid truth matrix между platform state и GitHub state;
  - filtered multi-root workspace с точными Wave 1 filters `open_only`, `assigned_to_me_or_unassigned`, active-state presets, left-to-right graph layout и secondary/dimmed semantics только для связующих узлов;
  - Wave 1 node set `discussion`, `work_item`, `run`, `pull_request`, без `agent` node;
  - typed metadata contract, platform-generated watermarks и platform-canonical launch params;
  - continuity rule: каждый stage до `run:dev` включительно обязан завершаться `PR + linked follow-up issue`.
- Создана continuity issue `#496` для stage `run:vision` без trigger-лейбла.
- Локально перепроверены `gh issue create --help`, `gh pr create --help` и `gh pr edit --help` для non-interactive continuity issue / PR flow.
- Root FR/NFR matrix в `docs/delivery/requirements_traceability.md` не менялась по существу: intake stage формализует problem/scope/handover и historical delta, а не добавляет новые канонические требования.

## Актуализация по Issue #496 (`run:vision`, 2026-03-15)
- Подготовлен vision package:
  - `docs/delivery/epics/s16/epic-s16-day2-mission-control-graph-workspace-vision.md`;
  - обновлены `docs/delivery/sprints/s16/sprint_s16_mission_control_graph_workspace.md`, `docs/delivery/epics/s16/epic_s16.md`, `docs/delivery/delivery_plan.md` и `docs/delivery/issue_map.md`.
- Зафиксированы:
  - Mission Control как primary multi-root graph workspace/control plane для continuity `discussion/work_item -> run -> pull_request/follow-up issue -> next run`;
  - mission, north star, persona outcomes, KPI/guardrails и wave boundaries без reopening Day1 baseline;
  - non-negotiable baseline по issue `#480`: persisted provider mirror и coverage contract `all open Issues/PR + bounded recent closed history`;
  - exact Wave 1 filters `open_only`, `assigned_to_me_or_unassigned`, `active-state presets`, secondary/dimmed semantics только для graph integrity и Wave 1 nodes `discussion`, `work_item`, `run`, `pull_request`;
  - typed metadata/watermark contract, platform-canonical launch params и continuity rule `PR + linked follow-up issue`;
  - service-boundary guardrail для следующего stage: `control-plane` остаётся owner graph truth, continuity state и launch surfaces, а `worker` ограничен background/reconcile execution для foundation inventory и lifecycle tasks;
  - later-wave boundary: voice/STT, dashboard orchestrator agent и отдельная `agent` node taxonomy не блокируют core Wave 1.
- Через `gh issue create` создана follow-up issue `#510` для stage `run:prd`; в её body сохранено continuity-требование продолжить цепочку `arch -> design -> plan -> dev` после PRD.
- Выполнены markdown-only проверки: traceability sync, `git diff --check`, локальная проверка `gh issue create --help`; kubectl/logs/БД-запросы не выполнялись, потому что stage ограничен documentation-only scope.
- Root FR/NFR matrix в `docs/delivery/requirements_traceability.md` не менялась: vision stage уточнил product baseline и handover, но не добавлял новые канонические FR/NFR в `docs/product/requirements_machine_driven.md`.
