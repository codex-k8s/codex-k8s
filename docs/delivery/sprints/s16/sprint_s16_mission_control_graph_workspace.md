---
doc_id: SPR-CK8S-0016
type: sprint-plan
title: "Sprint S16: Mission Control graph workspace and continuity control plane (Issues #492/#516/#519/#537)"
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

# Sprint S16: Mission Control graph workspace and continuity control plane (historical superseded baseline)

## TL;DR
- 2026-03-25 issue `#561` перевела Sprint S16 в historical superseded baseline по owner decision из discussion `#480`.
- Day1-Day6 и handover `#542..#547` сохраняются как historical evidence и больше не являются текущим source of truth для Mission Control UX, data model и sequencing.
- Актуальный reset path: `#561` (`run:rethink`) -> `#562` (frontend-first sprint на fake data) -> `#563` (backend rebuild после owner approval UX).
- Новый baseline after rethink: fullscreen свободный canvas, node taxonomy `Issue/PR/Run`, workflow editor/policy как часть нового UX-направления, repo-seed prompts + deterministic `workflow-policy block`, `stale/freshness` только как доказанный lag provider mirror/reconcile path.

## Статус после rethink
- Больше не считать source of truth:
  - lane/column shell и обязательную иерархию `root-group -> column -> stack`;
  - Wave 1 taxonomy `discussion/work_item/run/pull_request`;
  - rollout `#542..#547` как обязательный следующий execution path;
  - `#547` как readiness gate перед `run:qa`;
  - трактовку `stale/freshness` как возраста проекции.
- Historical evidence, которое сохраняется:
  - Day1-Day6 chain `#492 -> #496 -> #510 -> #516 -> #519 -> #537`;
  - reasoning по ownership split, continuity и scope boundaries;
  - ссылки на отклонённый execution backlog `#542..#547`.
- Текущий sequencing после doc-reset:
  - `#562` вести как frontend-first flow `intake -> vision -> prd -> arch -> design -> plan -> dev`;
  - `#563` запускать только после owner approval результата `#562`;
  - `#522` и `#523` можно продолжать независимо;
  - `#524` и `#525` не стартовать до approval `#562`;
  - `#470` продолжать только в части `release safety`, `observability contract` и stop/rollback criteria без фиксации финального cockpit UI.
- Дополнительный reset guardrail:
  - workflow editor не выпадает из нового baseline: он должен проектироваться вместе с canvas UX и policy авто-ревью/follow-up propagation, но без превращения в DB prompt editor.

## Scope спринта
### In scope
- Полная doc-stage цепочка `intake -> vision -> prd -> arch -> design -> plan` для инициативы Mission Control graph workspace.
- Формализация продуктовой модели для:
  - fullscreen canvas + detached top toolbar + right drawer/chat;
  - filtered multi-root graph workspace;
  - hybrid truth matrix platform/GitHub;
  - inventory-backed provider foundation;
  - typed metadata/watermarks/launch params;
  - continuity rule `stage artifact = PR + linked follow-up issue`.
- Создание последовательных follow-up issue без `run:*`-лейблов; после `run:plan` Owner отдельно запускает execution stage.

### Out of scope
- Кодовая реализация до завершения и утверждения `run:plan`.
- Voice/STT как blocking scope для core Wave 1.
- Подмена GitHub review/merge/provider-native collaboration дашбордом.
- Попытка использовать GitHub Projects / Issue Type / Relationships как primary graph source of truth.
- Live-fetch-only dashboard без persisted provider mirror.

## Рекомендованный launch profile
- Базовый launch profile: `new-service`.
- Обоснование:
  - инициатива меняет product contour Mission Control и затрагивает несколько bounded contexts;
  - нужны обязательные `vision`, `arch` и `design`, чтобы зафиксировать product truth matrix и ownership boundaries до implementation;
  - сокращённые траектории не удержат continuity contract и cross-service impact.
- Целевая continuity-цепочка:
  `#492 (intake) -> #496 (vision) -> #510 (prd) -> #516 (arch) -> #519 (design) -> #537 (plan) -> #542..#547 (dev waves) -> qa -> release -> postdeploy -> ops`.

## Intake baseline, зафиксированный на Day 1

### Product shape
- Mission Control должен стать fullscreen graph workspace/control plane, а не улучшенной dashboard-страницей Sprint S9.
- Workspace по умолчанию multi-root: показывает все сущности, прошедшие точные Wave 1 filters `open_only`, `assigned_to_me_or_unassigned` и active-state presets, а не только одну выбранную инициативу и не «весь мир».
- Graph layout для каждой инициативы идёт слева направо: discussion/root слева, runs и downstream artifacts справа.
- Узлы, нужные для graph integrity, но не прошедшие основной фильтр, остаются secondary/dimmed, а не исчезают полностью.

### Truth matrix and continuity
- Platform canonical:
  - operational graph and relations;
  - run nodes and produced artifacts;
  - launch params;
  - dashboard metadata;
  - sync state;
  - platform-generated watermarks.
- GitHub canonical:
  - issue/pr/comment/review state;
  - provider-native development links.
- Каждый stage до `run:dev` включительно обязан завершаться PR/markdown package и linked follow-up issue для следующего шага.

### Wave 1 baseline
- Node types: `discussion`, `work_item`, `run`, `pull_request`.
- `agent` не становится canvas node в первой волне.
- Comments/chat/summaries остаются drawer/timeline entities.
- Inventory-backed provider mirror из `#480` входит в core foundation с coverage contract `all open Issues/PR + bounded recent closed history`, но сам по себе не считается финальным продуктовым результатом.
- Voice/STT и dashboard orchestrator agent остаются later-wave path.

## Vision baseline, зафиксированный на Day 2

### Mission and outcomes
- Mission Control подтверждён как primary multi-root graph workspace/control plane, а не как board/list refresh Sprint S9.
- Workspace должен помогать пользователю быстро понять continuity по нескольким инициативам сразу: от discussion/work item до run, PR и follow-up issue.
- Граница между core Wave 1 и later waves зафиксирована явно: core ценность достигается без отдельной `agent` node taxonomy и без voice/STT path.

### Personas and product guardrails
- Owner / product lead должен видеть situational awareness по нескольким инициативам и запускать следующий безопасный шаг без ручного GitHub label hunting.
- Delivery operator / engineer / manager должен получать единый control plane для run context, launch params и downstream artifacts.
- Discussion-driven workflow остаётся first-class, но не становится единственным входом: stage-issue можно создавать и связывать напрямую.
- Human review, merge и provider-native collaboration остаются в GitHub UI; dashboard не подменяет provider semantics.

### Success framing
- Vision зафиксировал измеримую рамку успеха:
  - graph workspace adoption;
  - next-step clarity;
  - inventory-backed coverage;
  - hybrid truth merge correctness;
  - continuity completeness по правилу `PR + follow-up issue`.
- Day3 PRD package `#510` уже превратил vision-рамку в user stories, FR/AC/NFR, scenario matrix и expected evidence; следующий stage должен удержать этот продуктовый контракт в architecture package `#516`.

## План этапов и handover

| Stage | Основной артефакт | Целевая роль | Правило выхода |
|---|---|---|---|
| Intake (`#492`) | Problem/Brief/Scope/Constraints + intake AC | `pm` | Owner review intake-пакета и создана issue следующего этапа |
| Vision (`#496`) | Mission, north star, persona outcomes, KPI/guardrails, wave framing | `pm` | Зафиксирован vision baseline и создана issue для `run:prd` |
| PRD (`#510`) | User stories, FR/AC/NFR, scenario matrix, expected evidence | `pm` + `sa` | Подтверждён PRD package и создана issue `#516` для `run:arch` |
| Architecture (`#516`) | Ownership matrix, graph truth model, provider mirror/service boundaries | `sa` | Подтверждены сервисные границы и создана issue `#519` для `run:design` |
| Design (`#519`) | Typed API/data/UI contracts, metadata/watermark design, rollout notes | `sa` + `qa` | Подготовлен implementation-ready design package и создана issue `#537` для `run:plan` |
| Plan (`#537`) | Delivery waves, execution decomposition, DoR/DoD, quality-gates | `em` + `km` | Сформирован execution package и созданы owner-managed issues `#542..#547` для `run:dev` |

## Guardrails спринта
- Sprint S16 расширяет Mission Control поверх existing baselines Sprint S9/S12/issue `#480`, а не игнорирует их.
- Dashboard не создаёт обходов label/stage policy, audit trail, owner approvals и provider review semantics.
- Hybrid truth matrix должна оставаться typed и explicit; markdown scraping и LLM-generated watermarks не допускаются как canonical source.
- Low-fidelity live-fetch UI без persisted mirror не считается допустимым shortcut.
- Voice/orchestrator path не имеет права блокировать core Wave 1.

## Handover
- Day1/Day6 package:
  - `docs/delivery/sprints/s16/sprint_s16_mission_control_graph_workspace.md`;
  - `docs/delivery/epics/s16/epic_s16.md`;
  - `docs/delivery/epics/s16/epic-s16-day1-mission-control-graph-workspace-intake.md`;
  - `docs/delivery/epics/s16/epic-s16-day2-mission-control-graph-workspace-vision.md`;
  - `docs/delivery/epics/s16/epic-s16-day3-mission-control-graph-workspace-prd.md`;
  - `docs/delivery/epics/s16/prd-s16-day3-mission-control-graph-workspace.md`;
  - `docs/delivery/epics/s16/epic-s16-day4-mission-control-graph-workspace-arch.md`;
  - `docs/delivery/epics/s16/epic-s16-day5-mission-control-graph-workspace-design.md`;
  - `docs/delivery/epics/s16/epic-s16-day6-mission-control-graph-workspace-plan.md`;
  - `docs/architecture/initiatives/s16_mission_control_graph_workspace/README.md`;
  - `docs/architecture/initiatives/s16_mission_control_graph_workspace/architecture.md`;
  - `docs/architecture/initiatives/s16_mission_control_graph_workspace/design_doc.md`;
  - `docs/architecture/initiatives/s16_mission_control_graph_workspace/api_contract.md`;
  - `docs/architecture/initiatives/s16_mission_control_graph_workspace/data_model.md`;
  - `docs/architecture/initiatives/s16_mission_control_graph_workspace/migrations_policy.md`;
  - `docs/delivery/traceability/s16_mission_control_graph_workspace_history.md`.
- Текущий stage в review: `run:plan` в Issue `#537`.
- Следующий operational stage: `run:dev` через owner-managed issues `#542..#547`.
- Execution backlog:
  - `#542` (`S16-E01`) — schema foundation и bounded graph backfill.
  - `#543` (`S16-E02`) — `control-plane` graph truth, continuity projections и read-only launch preview.
  - `#544` (`S16-E03`) — `worker` reconcile, freshness и bounded recent-closed-history execution.
  - `#545` (`S16-E04`) — typed transport surfaces и launch preview exposure.
  - `#546` (`S16-E05`) — `web-console` graph workspace и continuity UX.
  - `#547` (`S16-E06`) — observability, rollout gate и readiness evidence.
- На `run:dev` нельзя потерять следующие решения intake + vision + PRD + architecture + design + plan:
  - Sprint S16 = полный redesign Mission Control в primary multi-root graph workspace/control plane;
  - `#480` = mandatory foundation stream с coverage contract `all open Issues/PR + bounded recent closed history`;
  - multi-root filtered workspace = default baseline;
  - Wave 1 filters = `open_only`, `assigned_to_me_or_unassigned`, active-state presets;
  - secondary/dimmed handling используется только для graph integrity;
  - Wave 1 nodes = `discussion/work_item/run/pull_request`, без отдельной `agent` node taxonomy;
  - hybrid truth matrix остаётся typed и explicit;
  - typed metadata, platform-generated watermarks и platform-canonical launch params обязательны;
  - platform-safe inline actions ограничены context/drawer, inspect run context, launch next allowed stage и open linked PR/follow-up issue;
  - отсутствие linked PR или linked follow-up issue считается continuity gap, а не допустимым частичным результатом stage;
  - human review/merge/provider-native collaboration остаются в GitHub UI;
  - voice/STT, dashboard orchestrator agent, отдельная `agent` node taxonomy, full-history/archive и richer provider enrichment остаются later-wave scope;
  - stage continuity до `run:dev` = `PR + linked follow-up issue`;
  - `control-plane` остаётся owner graph truth, continuity state и launch surfaces, а `worker` ограничен background/reconcile execution для provider mirror и lifecycle tasks;
  - architecture stage зафиксировал hybrid truth lifecycle `provider mirror -> graph truth -> workspace projection`, typed watermarks и continuity gaps как доменный контур `control-plane`;
  - design stage зафиксировал graph-first transport/data contracts, persisted continuity gaps/workspace watermarks, run nodes вместо `agent` nodes и rollout path `migrations -> control-plane -> worker -> api-gateway -> web-console`;
  - plan stage зафиксировал waves `#542 -> #543 -> #544 -> #545 -> #546 -> #547`, DoR/DoD, quality-gates и запрет массового параллельного старта execution backlog;
  - handover в `run:qa` не допускается до закрытия `#547` и подтверждённого readiness/observability evidence;
  - Sprint S16 не возвращается к Sprint S9 dashboard-first модели и не вводит voice/STT или отдельную `agent` taxonomy в core Wave 1.
