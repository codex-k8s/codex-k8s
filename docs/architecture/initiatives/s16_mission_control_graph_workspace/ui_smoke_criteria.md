---
doc_id: UI-SMOKE-S16-CK8S-0001
type: readiness-handover
title: "Sprint S16 Day 7 — UI smoke criteria for Mission Control graph workspace (Issue #546)"
status: in-review
owner_role: Dev
created_at: 2026-03-19
updated_at: 2026-03-19
related_issues: [519, 537, 545, 546, 547]
related_prs: []
related_adrs: ["ADR-0016"]
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-19-issue-546-ui-smoke"
---

# UI Smoke Criteria: Mission Control graph workspace

## TL;DR
- Документ фиксирует минимальный UI-level evidence bundle, который Wave `S16-E06` должен использовать перед handover в `run:qa`.
- Smoke ориентирован на graph-first workspace, continuity visibility, read-only launch preview и transport-driven list projection без локальной policy logic.
- Критерии предполагают candidate runtime после rollout order `migrations -> control-plane -> worker -> api-gateway -> web-console`.

## Scope

### Что обязательно покрыть
- fullscreen Mission Control workspace в `web-console`;
- fixed Wave 1 filters `open_only` и `assigned_to_me_or_unassigned`;
- active-state presets `working|waiting|blocked|review|recent_critical_updates|all_active`;
- root groups / nodes `discussion|work_item|run|pull_request`;
- secondary/dimmed visibility только как graph integrity контекст;
- right drawer с details, linked artifacts, activity и node/workspace watermarks;
- read-only launch preview, работающий только через preview route;
- list projection как fallback той же typed выборки, а не отдельная локальная read model.

### Что не является acceptance для этой волны
- фактическое выполнение `stage.next_step.execute` из graph workspace;
- provider-native write actions, review/merge flows, voice/STT и `agent` nodes;
- локальные эвристики next-step policy поверх transport payload.

## Обязательные smoke-сценарии

### SC-01. Workspace shell и fixed filters
1. Открыть Mission Control route по умолчанию.
2. Убедиться, что primary presentation = `graph`, а не board/dashboard-first fallback.
3. Проверить, что toolbar показывает fixed filters `open_only` и `assigned_to_me_or_unassigned` как read-only chips.
4. Переключить каждый `state_preset` и убедиться, что route/query меняются только по `view/filter/q/node_kind/node_id`.

Ожидаемый результат:
- graph workspace открывается без legacy board semantics;
- fixed filters отображаются, но не редактируются пользователем;
- preset toggle не ломает deep-link и не подменяет snapshot локальной фильтрацией.

### SC-02. Multi-root graph и secondary/dimmed semantics
1. Загрузить candidate snapshot с несколькими инициативами.
2. Проверить, что canvas показывает несколько root groups одновременно.
3. Найти хотя бы один `secondary_dimmed` node и убедиться, что он визуально ослаблен, но остаётся в graph только ради continuity path или recent closed context.
4. Проверить, что `run` отображается отдельным node kind, а `agent` не появляется на canvas.

Ожидаемый результат:
- multi-root layout сохраняется;
- dimmed node не маскируется под primary и не исчезает полностью;
- `run` node виден как самостоятельный шаг continuity chain.

### SC-03. Drawer, linked artifacts и persisted continuity gaps
1. Открыть `work_item`, `run` и `pull_request` nodes из одного root group.
2. Проверить drawer sections:
   - overview;
   - continuity gaps;
   - watermarks;
   - linked artifacts / adjacent nodes;
   - activity / chat.
3. Если у узла есть persisted gaps `missing_pull_request` или `missing_follow_up_issue`, убедиться, что они отображаются как persisted signals, а не как transient UI hints.
4. Перейти по linked artifact chips и убедиться, что deep-link остаётся внутри typed node navigation.

Ожидаемый результат:
- drawer показывает typed payload без локального реконструирования доменной модели;
- linked PR/follow-up issue continuity читается явно;
- persisted gaps остаются видимыми до фактического закрытия в transport payload.

### SC-04. Read-only launch preview
1. В drawer открыть launch surface `preview_next_stage`.
2. Убедиться, что preview открывается в отдельном read-only dialog.
3. Проверить наличие:
   - target label;
   - removed/added/final labels;
   - approval requirement;
   - resolved/remaining gap ids;
   - resulting node refs / provider redirects.
4. Убедиться, что dialog не содержит кнопки фактического submit/execute.

Ожидаемый результат:
- preview не создаёт command и не мутирует platform/provider state;
- label diff и continuity effect читаются напрямую из preview response;
- при `blocking_reason` UI показывает блокировку, но не делает локальных выводов о workaround.

### SC-05. List projection parity
1. Переключиться в `list`.
2. Проверить, что list использует те же nodes/snapshot, что и graph.
3. Сравнить для одного узла kind, continuity status, visibility tier, root id и last activity.
4. Вернуться в graph и убедиться, что выбранный node deep-link сохраняется.

Ожидаемый результат:
- list не использует отдельную dashboard-model;
- detail navigation и route continuity одинаковы для `graph` и `list`.

### SC-06. Watermarks и degraded/stale handling
1. Загрузить snapshot с workspace/node watermarks.
2. Проверить, что watermarks видны в canvas и drawer.
3. При realtime `invalidate|stale|degraded|resync_required` убедиться, что UI показывает transport-driven alerts, а не скрывает проблему.
4. Проверить, что degraded/stale режим не скрывает continuity gaps и linked artifacts.

Ожидаемый результат:
- freshness и coverage signals читаются как отдельные watermark surfaces;
- realtime alerts не подменяют continuity semantics локальной логикой;
- пользователь всё ещё видит проблему и может принудительно обновить snapshot.

## Evidence bundle для Wave S16-E06
- Скриншот graph workspace с минимум двумя root groups.
- Скриншот drawer для узла с persisted continuity gap.
- Скриншот read-only launch preview с label diff и continuity effect.
- Скриншот list projection для того же snapshot.
- Краткая запись candidate namespace/build ref и абсолютной даты проверки.

## Runtime checks для smoke
- UI:
  - открыть candidate `web-console` через browser;
  - сохранить route query после переключения preset/view и после открытия drawer node.
- Logs:
  - `kubectl -n <candidate-namespace> logs deploy/codex-k8s-api-gateway --tail=200 | rg 'mission_control.api.workspace|mission_control.api.node_details|mission_control.api.launch_preview'`
  - `kubectl -n <candidate-namespace> logs deploy/codex-k8s-control-plane --tail=200 | rg 'mission_control.workspace.(snapshot_loaded|preview_generated|preview_blocked|gap_detected|watermark_updated)'`

## Failure signals
- UI возвращается к dashboard/board-first представлению как primary path.
- `agent` node снова появляется на canvas.
- linked PR/follow-up issue не видны без локального догадай-слоя.
- preview dialog позволяет фактически submit action или скрывает `blocking_reason`.
- list projection расходится с graph snapshot по node status/visibility/root lineage.

## Связанные документы
- `docs/architecture/initiatives/s16_mission_control_graph_workspace/design_doc.md`
- `docs/architecture/initiatives/s16_mission_control_graph_workspace/api_contract.md`
- `docs/architecture/initiatives/s16_mission_control_graph_workspace/ui_smoke_criteria.md`
- `docs/delivery/epics/s16/epic-s16-day6-mission-control-graph-workspace-plan.md`
- `docs/delivery/sprints/s16/sprint_s16_mission_control_graph_workspace.md`
