---
doc_id: TRH-CK8S-S16-0001
type: traceability-history
title: "Sprint S16 Traceability History"
status: in-review
owner_role: KM
created_at: 2026-03-15
updated_at: 2026-03-16
related_issues: [480, 490, 492, 496, 510, 516, 519, 537]
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

## Актуализация по Issue #510 (`run:prd`, 2026-03-16)
- Подготовлен PRD package:
  - `docs/delivery/epics/s16/epic-s16-day3-mission-control-graph-workspace-prd.md`;
  - `docs/delivery/epics/s16/prd-s16-day3-mission-control-graph-workspace.md`;
  - обновлены `docs/delivery/sprints/s16/sprint_s16_mission_control_graph_workspace.md`, `docs/delivery/epics/s16/epic_s16.md`, `docs/delivery/delivery_plan.md`, `docs/delivery/issue_map.md` и `docs/delivery/sprints/README.md`.
- Зафиксированы:
  - user stories, FR/AC/NFR, scenario matrix и expected evidence для fullscreen graph workspace, filtered multi-root continuity, inventory-backed foundation, typed metadata/watermarks, platform-canonical launch params и platform-safe inline actions;
  - locked baseline по issue `#480`, exact Wave 1 filters `open_only`, `assigned_to_me_or_unassigned`, `active-state presets`, secondary/dimmed semantics только для graph integrity и Wave 1 nodes `discussion`, `work_item`, `run`, `pull_request`;
  - explicit continuity contract: stage through `run:dev` считается complete только при наличии `PR + linked follow-up issue`, а отсутствие любого из этих артефактов трактуется как continuity gap;
  - deferred boundary для voice/STT, dashboard orchestrator agent, отдельной `agent` node taxonomy, full-history/archive и richer provider enrichment.
- Через `gh issue create` создана follow-up issue `#516` для stage `run:arch`; в её body сохранено continuity-требование продолжить цепочку `arch -> design -> plan -> dev`.
- Выполнены markdown-only проверки: traceability sync и `git diff --check`; kubectl/logs/БД-запросы не выполнялись, потому что stage ограничен documentation-only scope.
- Root FR/NFR matrix в `docs/delivery/requirements_traceability.md` не менялась: PRD package формализует stage-specific contract Sprint S16 и handover в architecture, но не меняет repo-wide baseline `docs/product/requirements_machine_driven.md`.

## Актуализация по Issue #516 (`run:arch`, 2026-03-16)
- Подготовлен architecture package:
  - `docs/architecture/initiatives/s16_mission_control_graph_workspace/README.md`;
  - `docs/architecture/initiatives/s16_mission_control_graph_workspace/architecture.md`;
  - `docs/architecture/initiatives/s16_mission_control_graph_workspace/c4_context.md`;
  - `docs/architecture/initiatives/s16_mission_control_graph_workspace/c4_container.md`;
  - `docs/architecture/adr/ADR-0016-mission-control-graph-workspace-hybrid-truth-and-continuity-ownership.md`;
  - `docs/architecture/alternatives/ALT-0008-mission-control-graph-workspace-hybrid-truth-boundaries.md`;
  - `docs/delivery/epics/s16/epic-s16-day4-mission-control-graph-workspace-arch.md`.
- Зафиксированы:
  - `control-plane` как canonical owner graph truth, continuity state, typed metadata/watermarks, launch surfaces и hybrid truth merge policy;
  - `worker` как owner bounded provider inventory freshness, recent-closed-history backfill, enrichment/reconcile execution и lifecycle tasks без ownership graph semantics;
  - явный hybrid truth lifecycle `provider mirror -> graph truth -> workspace projection`;
  - persisted continuity gaps и rule `PR + linked follow-up issue` как domain constructs, а не только traceability convention;
  - сохранение locked baselines: issue `#480`, exact Wave 1 filters `open_only`, `assigned_to_me_or_unassigned`, `active-state presets`, secondary/dimmed semantics только для graph integrity, nodes `discussion`, `work_item`, `run`, `pull_request`, typed metadata/watermarks и platform-canonical launch params;
  - deferred boundary для voice/STT, dashboard orchestrator agent, отдельной `agent` taxonomy, full-history/archive и richer provider enrichment.
- Через `gh issue create` создана follow-up issue `#519` для stage `run:design`; в её body сохранено continuity-требование продолжить цепочку `design -> plan -> dev`.
- Выполнены markdown-only проверки: traceability sync, `git diff --check`, локальная проверка `gh issue create --help`, `gh pr create --help`, `gh pr edit --help`; kubectl/logs/БД-запросы не выполнялись, потому что stage ограничен documentation-only scope.
- Root FR/NFR matrix в `docs/delivery/requirements_traceability.md` не менялась: architecture package формализует ownership boundaries и handover в design, но не меняет repo-wide baseline `docs/product/requirements_machine_driven.md`.

## Актуализация по Issue #519 (`run:design`, 2026-03-16)
- Подготовлен design package:
  - `docs/architecture/initiatives/s16_mission_control_graph_workspace/design_doc.md`;
  - `docs/architecture/initiatives/s16_mission_control_graph_workspace/api_contract.md`;
  - `docs/architecture/initiatives/s16_mission_control_graph_workspace/data_model.md`;
  - `docs/architecture/initiatives/s16_mission_control_graph_workspace/migrations_policy.md`;
  - `docs/delivery/epics/s16/epic-s16-day5-mission-control-graph-workspace-design.md`.
- Зафиксированы:
  - graph-first interaction model поверх Day4 ownership split без нового deployable сервиса;
  - typed transport baseline `workspace -> node details -> activity -> launch preview -> existing command ledger`, где Sprint S9 dashboard contract переводится в superseded state без отдельного parallel namespace;
  - reuse existing Mission Control command path для `stage.next_step.execute`, а preview выносится в явный read-only contract с continuity effect;
  - persisted continuity gaps и workspace watermarks как отдельные domain constructs `control-plane`;
  - run nodes как обязательный Wave 1 canvas kind вместо `agent` nodes, которые остаются только migration residue до cleanup;
  - rollout path `expand schema -> shadow backfill -> read switch -> preview exposure -> cleanup last`, сохранив order `migrations -> control-plane -> worker -> api-gateway -> web-console`.
- Через `gh issue create` создана follow-up issue `#537` для stage `run:plan`; в её body сохранено continuity-требование продолжить цепочку `plan -> dev`.
- Выполнены markdown-only проверки: traceability sync, `git diff --check`, локальная проверка `gh issue create --help`, `gh pr create --help`, `gh pr edit --help`; kubectl/logs/БД-запросы не выполнялись, потому что stage ограничен documentation-only scope.
- Root FR/NFR matrix в `docs/delivery/requirements_traceability.md` не менялась: design package детализирует existing Sprint S16 baseline и handover в plan, но не меняет repo-wide baseline `docs/product/requirements_machine_driven.md`.
