---
doc_id: DM-S18-MISSION-CONTROL-0001
type: data-model
title: "Mission Control frontend-first canvas prototype — Data model Sprint S18 Day 5"
status: in-review
owner_role: SA
created_at: 2026-04-01
updated_at: 2026-04-01
related_issues: [480, 561, 562, 563, 565, 567, 571, 573, 579]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-04-01-issue-573-data-model"
---

# Data Model: Mission Control frontend-first canvas prototype

## TL;DR
- Ключевые сущности: `MissionCanvasScenario`, `MissionCanvasInitiative`, `MissionCanvasNode`, `MissionCanvasRelation`, `MissionDrawerRecord`, `MissionWorkflowPreset`, `MissionCanvasUIState`.
- Основные связи: scenario owns initiatives/nodes/presets; node owns drawer record and relation refs; workflow preset feeds read-only preview for selected node or initiative.
- Риски миграций: для Sprint S18 отсутствуют, потому что data model живёт только в frontend bundle и не объявляется временным backend source of truth.

## Сущности
### Entity: `MissionCanvasScenario`
- Назначение: верхнеуровневый walkthrough bundle для одного owner demo.
- Важные инварианты:
  - `scenario_id` уникален в пределах feature catalog;
  - scenario содержит только taxonomy `Issue`, `PR`, `Run`;
  - scenario хранит source refs на prompt/policy docs, но не хранит editable prompt text как source of truth.

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| `scenario_id` | `string` | no |  | unique | stable local ref |
| `title` | `string` | no |  |  | |
| `summary` | `string` | no |  |  | short walkthrough context |
| `initiative_ids` | `string[]` | no | `[]` | each id must exist in `MissionCanvasInitiative` | |
| `node_ids` | `string[]` | no | `[]` | each id must exist in `MissionCanvasNode` | |
| `workflow_preset_ids` | `string[]` | no | `[]` | each id must exist in `MissionWorkflowPreset` | |
| `source_refs` | `string[]` | no | `[]` | non-empty | repo seed / policy refs displayed in UI |

### Entity: `MissionCanvasInitiative`
- Назначение: presentational grouping for `1..3` initiatives on the same canvas.
- Важные инварианты:
  - initiative is not a node kind;
  - node ordering is advisory only and never turns into columns/lanes;
  - accent token is cosmetic and must not encode workflow meaning.

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| `initiative_id` | `string` | no |  | unique within scenario | |
| `label` | `string` | no |  |  | visible cluster label |
| `accent_token` | `string` | no |  |  | color/style ref |
| `focus_order` | `number` | no | `0` | `>= 0` | toolbar ordering |
| `node_ids` | `string[]` | no | `[]` | each id belongs to the same scenario | |

### Entity: `MissionCanvasNode`
- Назначение: compact canvas card for `Issue`, `PR` or `Run`.
- Важные инварианты:
  - `node_kind` belongs to closed set `Issue|PR|Run`;
  - node carries only compact summary data; full narrative stays in drawer record;
  - `layout` is scenario-local and not reused as a persisted backend coordinate system.

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| `node_id` | `string` | no |  | unique within scenario | local ref only |
| `node_kind` | `'Issue' | 'PR' | 'Run'` | no |  | closed enum | |
| `initiative_id` | `string` | no |  | must reference `MissionCanvasInitiative` | |
| `title` | `string` | no |  |  | compact label |
| `state` | `string` | no |  | check by presenter enum | working/review/etc. |
| `stage_label` | `string` | yes |  |  | optional stage chip |
| `layout_x` | `number` | no | `0` | freeform coordinate | |
| `layout_y` | `number` | no | `0` | freeform coordinate | |
| `badges` | `string[]` | no | `[]` | bounded list | attention hints only |
| `relation_ids` | `string[]` | no | `[]` | each id must exist in `MissionCanvasRelation` | |
| `detail_id` | `string` | no |  | must reference `MissionDrawerRecord` | |
| `safe_action_ids` | `string[]` | no | `[]` | local refs only | deep-links / preview actions |

### Entity: `MissionCanvasRelation`
- Назначение: explicit edge between two canvas nodes.
- Важные инварианты:
  - source and target nodes always belong to the same scenario;
  - relation kind comes from the closed set chosen for Sprint S18;
  - relations required for selected-node context remain visible even when search dims unrelated nodes.

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| `relation_id` | `string` | no |  | unique within scenario | |
| `relation_kind` | `'drives' | 'produces' | 'tracks' | 'blocks'` | no |  | closed enum | |
| `source_node_id` | `string` | no |  | must reference `MissionCanvasNode` | |
| `target_node_id` | `string` | no |  | must reference `MissionCanvasNode` | |
| `label` | `string` | no |  |  | visible edge text |
| `importance` | `'primary' | 'supporting'` | no | `'primary'` |  | used for dimming |

### Entity: `MissionDrawerRecord`
- Назначение: full-detail payload for the selected node.
- Важные инварианты:
  - one drawer record per node for Sprint S18;
  - timeline entries are fake-data evidence only;
  - safe actions are read-only or deep-link actions.

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| `detail_id` | `string` | no |  | unique within scenario | |
| `node_id` | `string` | no |  | must reference `MissionCanvasNode` | |
| `overview_markdown` | `string` | no |  |  | narrative summary |
| `timeline_items` | `object[]` | no | `[]` | typed by feature-local union | no provider realtime |
| `related_node_ids` | `string[]` | no | `[]` | existing node refs only | drawer shortcuts |
| `safe_actions` | `object[]` | no | `[]` | preview/deep-link only | no provider mutation |
| `workflow_preset_ids` | `string[]` | no | `[]` | must reference `MissionWorkflowPreset` | |

### Entity: `MissionWorkflowPreset`
- Назначение: structured baseline for workflow policy preview.
- Важные инварианты:
  - preset carries deterministic generation inputs, not free-form prompt body;
  - every preset includes at least one repo-seed or policy source ref;
  - generated block stays read-only in UI.

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| `preset_id` | `string` | no |  | unique within scenario | |
| `label` | `string` | no |  |  | visible preset name |
| `stage_sequence` | `string[]` | no | `[]` | ordered | fake-data workflow stages |
| `auto_review_policy` | `string` | no |  | closed enum in feature types | |
| `follow_up_policy` | `string` | no |  | closed enum in feature types | |
| `safe_action_profile` | `string` | no |  | closed enum in feature types | |
| `prompt_seed_refs` | `string[]` | no | `[]` | non-empty | repo seed / docs refs |
| `generated_block_template` | `string` | no |  |  | markdown preview template |
| `allowed_overrides` | `string[]` | no | `[]` | closed set | structured toggles only |

### Entity: `MissionCanvasUIState`
- Назначение: route-local interaction state inside Pinia.
- Важные инварианты:
  - store state remains UI-local and resets safely to fixture baseline;
  - no field is treated as persisted truth or synchronized backend state;
  - `selected_node_id` and `active_workflow_preset_id` always reference the active scenario.

| Field | Type | Nullable | Default | Constraints | Notes |
|---|---|---:|---|---|---|
| `active_scenario_id` | `string` | no |  | must exist in catalog | |
| `focused_initiative_id` | `string` | yes |  | same scenario only | |
| `selected_node_id` | `string` | yes |  | same scenario only | |
| `search_query` | `string` | no | `""` |  | |
| `zoom_level` | `number` | no | `1` | `> 0` | canvas viewport |
| `pan_x` | `number` | no | `0` |  | canvas viewport |
| `pan_y` | `number` | no | `0` |  | canvas viewport |
| `drawer_tab` | `'details' | 'timeline' | 'workflow'` | no | `'details'` | |
| `active_workflow_preset_id` | `string` | yes |  | same scenario only | |

## Связи
- `MissionCanvasScenario` 1:N `MissionCanvasInitiative`
- `MissionCanvasScenario` 1:N `MissionCanvasNode`
- `MissionCanvasScenario` 1:N `MissionCanvasRelation`
- `MissionCanvasScenario` 1:N `MissionWorkflowPreset`
- `MissionCanvasNode` 1:1 `MissionDrawerRecord`
- `MissionCanvasNode` 1:N `MissionCanvasRelation` (as source or target)
- `MissionDrawerRecord` N:M `MissionWorkflowPreset` via `workflow_preset_ids`
- `MissionCanvasUIState` references one active scenario and optional initiative/node/preset within that scenario

## Индексы и запросы (критичные)
- Query: load active scenario by id
  - Index / structure:
    - `scenarioById: Map<string, MissionCanvasScenario>`
- Query: resolve selected node and drawer payload
  - Index / structure:
    - `nodesById: Map<string, MissionCanvasNode>`
    - `drawerByNodeId: Map<string, MissionDrawerRecord>`
- Query: highlight related nodes and edges
  - Index / structure:
    - `relationsByNodeId: Map<string, MissionCanvasRelation[]>`
- Query: regenerate workflow preview from selected preset
  - Index / structure:
    - `workflowPresetById: Map<string, MissionWorkflowPreset>`
- Query: client-side search over the active scenario
  - Index / structure:
    - precomputed `searchTokensByNodeId`
- Оценка нагрузки:
  - bounded owner-demo dataset only; no remote pagination or persistence path is required in Sprint S18.

## Политика хранения данных
- Retention:
  - data lives only for the browser session / page reload.
- Архивирование:
  - not required.
- PII/комплаенс:
  - no PII or secret material is expected in the fixture catalog;
  - source refs point to repo-local docs or prompt seed paths only.

## Миграции (ссылка)
- См. `docs/architecture/initiatives/s18_mission_control_frontend_first_canvas/migrations_policy.md`.
- Sprint S18 does not create DB migrations; the future persisted model belongs to `#563`.

## Открытые вопросы
- Нужно ли plan-stage принудительно требовать общий `searchTokensByNodeId` helper, или достаточно простого presenter-level filtering на bounded dataset?

## Апрув
- request_id: `owner-2026-04-01-issue-573-data-model`
- Решение: pending
- Комментарий: требуется owner review feature-local scenario model и documented replacement seam к backend rebuild.
