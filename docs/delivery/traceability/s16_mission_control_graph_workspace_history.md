---
doc_id: TRH-CK8S-S16-0001
type: traceability-history
title: "Sprint S16 Traceability History"
status: superseded
owner_role: KM
created_at: 2026-03-15
updated_at: 2026-03-25
related_issues: [480, 490, 492, 496, 510, 516, 519, 537, 542, 543, 544, 545, 546, 547, 561, 562, 563]
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
- С 2026-03-25 Sprint S16 хранится только как historical superseded baseline; активный reset path вынесен в issues `#561`, `#562`, `#563`.

## Актуализация по Issue #561 (`run:rethink`, 2026-03-25)
- Выполнен doc-reset Mission Control:
  - обновлены `docs/delivery/delivery_plan.md`, `docs/delivery/issue_map.md`, `docs/delivery/sprints/README.md`, `docs/delivery/epics/README.md`;
  - Sprint S16 и `docs/architecture/initiatives/s16_mission_control_graph_workspace/*` переведены в `status: superseded`;
  - Day1-Day6 и execution handover `#542..#547` зафиксированы только как historical evidence.
- Явно зафиксирован новый agreed baseline из discussion `#480`:
  - fullscreen свободный canvas без lane/column shell и без обязательной root-group модели;
  - минимальная node taxonomy Wave 1: `Issue`, `PR`, `Run`;
  - frontend-first sprint `#562` на fake data для утверждения UX;
  - workflow editor и workflow policy остаются частью нового направления, но сначала проходят frontend-first UX-валидацию на fake data;
  - backend rebuild вынесен в отдельный sprint `#563` после owner approval UX;
  - repo-seed prompt policy остаётся каноничной, а workflow behavior допускается только через deterministic generated `workflow-policy block`;
  - `stale/freshness` теперь означает только доказанный lag provider mirror/reconcile path.
- Зафиксирован порядок по соседнему backlog:
  - `#522` и `#523` можно продолжать независимо;
  - `#524` и `#525` не стартовать до approval `#562`;
  - `#470` продолжать только в части `release safety`, `observability contract` и stop/rollback criteria без финального cockpit UI.
- Issue `#547`, закрытая как not planned, сохранена только как historical superseded readiness handover и не является gate перед `run:qa`.

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

## Актуализация по Issue #537 (`run:plan`, 2026-03-16)
- Подготовлен plan package:
  - `docs/delivery/epics/s16/epic-s16-day6-mission-control-graph-workspace-plan.md`;
  - `docs/delivery/sprints/s16/sprint_s16_mission_control_graph_workspace.md`;
  - `docs/delivery/epics/s16/epic_s16.md`;
  - `docs/delivery/delivery_plan.md`;
  - `docs/delivery/issue_map.md`;
  - `docs/delivery/sprints/README.md`;
  - `docs/delivery/epics/README.md`.
- Зафиксированы:
  - execution package `S16-E01..S16-E06` с waves `schema/backfill foundation -> control-plane graph truth -> worker reconcile/freshness -> transport/preview -> web-console graph workspace -> readiness gate`;
  - owner-managed handover issues `#542`, `#543`, `#544`, `#545`, `#546`, `#547` без trigger-лейблов для перехода в `run:dev`;
  - явные DoR/DoD, quality-gates и rollout constraints `migrations -> control-plane -> worker -> api-gateway -> web-console -> readiness gate`;
  - сохранение design guardrails: issue `#480` остаётся foundation layer, exact Wave 1 filters/nodes не меняются, secondary/dimmed semantics работают только для graph integrity, launch preview остаётся read-only поверх existing command ledger, новый deployable сервис не появляется;
  - boundary относительно Sprint S9 удержан: dashboard-first model не возвращается, voice/STT, отдельная `agent` taxonomy и richer provider enrichment не входят в core execution package Day6.
- Созданы follow-up issues `#542`, `#543`, `#544`, `#545`, `#546`, `#547` для stage `run:dev` без trigger-лейблов.
- Для GitHub continuity повторно подтверждён non-interactive CLI flow локальными `gh issue create --help`, `gh pr create --help`, `gh pr edit --help`; через `gh issue create` оформлены handover issues `#542..#547`; kubectl/logs/БД-запросы не выполнялись, потому что stage ограничен documentation-only scope.
- Root FR/NFR matrix в `docs/delivery/requirements_traceability.md` не менялась по существу: plan stage зафиксировал execution decomposition и historical delta; в root-матрице синхронизирован related-issues index.
