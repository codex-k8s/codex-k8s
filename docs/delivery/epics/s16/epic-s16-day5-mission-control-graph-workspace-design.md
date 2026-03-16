---
doc_id: EPC-CK8S-S16-D5-MISSION-CONTROL-GRAPH
type: epic
title: "Epic S16 Day 5: Design для Mission Control graph workspace и rollout policy (Issues #519/#537)"
status: in-review
owner_role: SA
created_at: 2026-03-16
updated_at: 2026-03-16
related_issues: [480, 490, 492, 496, 510, 516, 519, 537]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-16-issue-519-design"
---

# Epic S16 Day 5: Design для Mission Control graph workspace и rollout policy (Issues #519/#537)

## TL;DR
- Подготовлен design package Sprint S16 для Mission Control graph workspace: detailed design, typed API contract, data model и migrations policy.
- Зафиксирован implementation-ready переход от Sprint S9 dashboard foundation к graph-first workspace без reopening Day4 ownership.
- Подготовлен handover в `run:plan` без premature code/schema changes.

## Priority
- `P0`.

## Контекст
- Intake baseline: `#492` (`docs/delivery/epics/s16/epic-s16-day1-mission-control-graph-workspace-intake.md`).
- Vision baseline: `#496` (`docs/delivery/epics/s16/epic-s16-day2-mission-control-graph-workspace-vision.md`).
- PRD baseline: `#510` (`docs/delivery/epics/s16/prd-s16-day3-mission-control-graph-workspace.md`).
- Architecture baseline: `#516` (`docs/architecture/initiatives/s16_mission_control_graph_workspace/architecture.md`).
- Текущий этап: `run:design` в Issue `#519`.
- Scope этапа: только markdown-изменения.

## Design package
- `docs/architecture/initiatives/s16_mission_control_graph_workspace/design_doc.md`
- `docs/architecture/initiatives/s16_mission_control_graph_workspace/api_contract.md`
- `docs/architecture/initiatives/s16_mission_control_graph_workspace/data_model.md`
- `docs/architecture/initiatives/s16_mission_control_graph_workspace/migrations_policy.md`
- `docs/delivery/traceability/s16_mission_control_graph_workspace_history.md`

## Ключевые решения Stage
- Graph-first workspace заменяет board/list-first S9 read model внутри того же Mission Control bounded context.
- Existing `mission_control_commands` ledger сохраняется; graph workspace не получает второй write-path, а использует preview + existing `stage.next_step.execute`.
- Persisted continuity gaps и workspace watermarks становятся отдельными доменными артефактами `control-plane`.
- Wave 1 transport emitting `discussion/work_item/run/pull_request`; `agent` node уходит из canvas в run metadata.
- Rollout и migration path остаются эволюционными: additive schema + shadow backfill + read switch + cleanup last.

## Context7 и внешний baseline
- Context7 lookup на Day5 не выполнялся: новые библиотеки и vendor integrations в scope отсутствуют.
- Локально проверены `gh issue create --help`, `gh pr create --help`, `gh pr edit --help` для non-interactive issue/PR automation.
- Новые внешние зависимости на этапе `run:design` не требуются.

## Acceptance Criteria (Issue #519)
- [x] Подготовлен design package с артефактами `design_doc`, `api_contract`, `data_model`, `migrations_policy`.
- [x] Typed contracts сохраняют locked baseline: `#480`, exact Wave 1 filters/nodes, secondary/dimmed semantics only for graph integrity, typed metadata/watermarks, platform-canonical launch params и rule `PR + linked follow-up issue`.
- [x] Зафиксированы data ownership, rollout sequence и migration/backfill/rollback notes без нового deployable сервиса и без переноса graph truth из `control-plane`.
- [x] Voice/STT, dashboard orchestrator agent, отдельная `agent` taxonomy, full-history/archive и richer provider enrichment оставлены deferred scope.
- [x] Подготовлена follow-up issue `#537` для stage `run:plan` без trigger-лейбла.

## Quality gates
| Gate | Что проверяем | Статус |
|---|---|---|
| `QG-S16-D5-01` Design completeness | Есть package `design_doc + api_contract + data_model + migrations_policy` | passed |
| `QG-S16-D5-02` Boundary integrity | Day4 ownership split сохранён без local graph logic outside `control-plane` | passed |
| `QG-S16-D5-03` Contract readiness | Transport/data/migration decisions достаточно конкретны для `run:plan` decomposition | passed |
| `QG-S16-D5-04` Wave 1 discipline | Exact filters/nodes, bounded foundation и deferred contours сохранены без scope drift | passed |
| `QG-S16-D5-05` Stage continuity | Подготовлена follow-up issue `#537` на `run:plan` без trigger-лейбла | passed |

## Handover в `run:plan`
- Следующий этап: `run:plan`.
- Follow-up issue: `#537`.
- Trigger-лейбл на новую issue ставит Owner после review design package.
- На plan-stage обязательно:
  - декомпозировать rollout минимум на waves `schema/backfill -> control-plane graph truth -> worker reconcile -> transport -> web-console -> readiness gate`;
  - разложить отдельные quality gates для graph parity, continuity gaps, launch preview и provider boundary;
  - сохранить порядок rollout `migrations -> control-plane -> worker -> api-gateway -> web-console`.
