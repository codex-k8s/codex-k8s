---
doc_id: API-S18-MISSION-CONTROL-0001
type: api-contract
title: "Mission Control frontend-first canvas prototype — API contract Sprint S18 Day 5"
status: in-review
owner_role: SA
created_at: 2026-04-01
updated_at: 2026-04-01
related_issues: [480, 561, 562, 563, 565, 567, 571, 573, 579]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-04-01-issue-573-api-contract"
---

# API Contract: Mission Control frontend-first canvas prototype

## TL;DR
- Тип контракта Sprint S18: feature-local TypeScript source contract, а не HTTP/gRPC transport.
- Аутентификация: не требуется, потому что источник данных bundle-local.
- Версионирование: через git review и repo-local types, без OpenAPI/codegen drift.
- Основные операции: load catalog, load scenario, resolve node details, generate workflow preview, reset local state.

## Спецификации (source of truth)
- OpenAPI (не меняется в Sprint S18): `services/external/api-gateway/api/server/api.yaml`
- gRPC proto (не меняется в Sprint S18): `proto/codexk8s/controlplane/v1/controlplane.proto`
- Feature-local source contract для `run:dev`:
  - `src/features/mission-control/prototype/source.ts`
  - `src/features/mission-control/prototype/types.ts`
  - `src/features/mission-control/prototype/fixtures.ts`
- Design source of truth:
  - `docs/architecture/initiatives/s18_mission_control_frontend_first_canvas/design_doc.md`
  - `docs/architecture/initiatives/s18_mission_control_frontend_first_canvas/data_model.md`

## Endpoints / Methods (кратко)
| Operation | Method/Topic | Path/Name | Auth | Idempotency | Notes |
|---|---|---|---|---|---|
| Load scenario catalog | local call | `MissionControlPrototypeSource.loadCatalog()` | n/a | yes | Returns walkthrough scenarios and available workflow presets |
| Load canvas scenario | local call | `MissionControlPrototypeSource.loadScenario({ scenarioId })` | n/a | yes | Returns nodes, relations, initiative groups and initial drawer state |
| Get node details | local call | `MissionControlPrototypeSource.getNodeDetails({ scenarioId, nodeId })` | n/a | yes | Returns drawer payload, timeline and safe actions |
| Generate workflow preview | local call | `MissionControlPrototypeSource.generateWorkflowPreview({ scenarioId, presetId, draft })` | n/a | yes | Deterministic preview only; no side effects |
| Reset local workspace state | store action | `MissionControlPrototypeStore.resetScenarioState()` | n/a | yes | Restores viewport, focus and workflow draft to fixture baseline |

## Модель ошибок
- Error codes / message keys:
  - `missionControlPrototype.scenarioNotFound`
  - `missionControlPrototype.nodeNotFound`
  - `missionControlPrototype.workflowPresetNotFound`
  - `missionControlPrototype.invalidWorkflowDraft`
- Retries:
  - not required; source is bundle-local and deterministic.
- Rate limits:
  - not applicable.

## Контракты данных (DTO)
- `MissionControlPrototypeCatalogItem`
  - `scenario_id`, `title`, `summary`, `initiative_count`, `node_count`, `default_focus_initiative_id`.
- `MissionControlCanvasScenario`
  - `scenario_id`, `initiatives[]`, `nodes[]`, `relations[]`, `default_viewport`, `workflow_presets[]`, `source_refs[]`.
- `MissionControlCanvasNode`
  - `node_id`, `node_kind`, `initiative_id`, `title`, `state`, `stage_label`, `layout`, `badges[]`, `safe_actions[]`.
- `MissionControlNodeDetails`
  - `node_id`, `overview_markdown`, `timeline_items[]`, `related_nodes[]`, `safe_actions[]`, `workflow_preset_ids[]`.
- `MissionWorkflowDraft`
  - `stage_sequence_variant`, `auto_review_policy`, `follow_up_policy`, `safe_action_profile`.
- `MissionWorkflowPreviewResult`
  - `generated_block_markdown`, `source_refs[]`, `change_explanations[]`, `warnings[]`.
- Validation rules:
  - only `Issue`, `PR`, `Run` are allowed node kinds;
  - relation endpoints must reference existing node ids within the same scenario;
  - every workflow preview result must carry at least one `source_ref`.

## Replacement seam to backend rebuild `#563`
| Prototype contract | Future owner in `#563` | Replacement rule |
|---|---|---|
| `loadCatalog()` | `web-console` + backend read model | Browser stops reading static fixture catalog and requests approved scenario/workspace data |
| `loadScenario()` | `control-plane` read model, possibly fed by `worker` mirror | Response may become transport-backed, but route-level canvas semantics stay the same |
| `getNodeDetails()` | `control-plane` + typed staff/private transport | Drawer payload becomes provider-backed; fake-data detail ids disappear |
| `generateWorkflowPreview()` | structured workflow policy layer from `#563` | Generated block remains deterministic and read-only; free-form prompt editing stays forbidden |

## Backward compatibility
- No backward compatibility guarantee is required inside Sprint S18 prototype; the repository is early-stage and the route is intentionally being re-baselined.
- Current generated `MissionControl*` DTO from OpenAPI stay outside Sprint S18 implementation surface and must not be retrofitted as temporary fake-data types.
- Future backend work in `#563` may replace the source implementation, but must preserve:
  - fullscreen canvas as primary route;
  - taxonomy `Issue` / `PR` / `Run`;
  - explicit relations;
  - drawer/workflow preview behavior.

## Наблюдаемость
- Логи:
  - no new backend logs or API audit surface.
- Метрики:
  - not required in Sprint S18.
- Трейсы:
  - not required; there is no remote request path.

## Открытые вопросы
- Не требуется.

## Апрув
- request_id: `owner-2026-04-01-issue-573-api-contract`
- Решение: pending
- Комментарий: требуется owner review локального source contract и explicit replacement seam к `#563`.
