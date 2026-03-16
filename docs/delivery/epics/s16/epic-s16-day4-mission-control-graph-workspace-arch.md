---
doc_id: EPC-CK8S-S16-D4-MISSION-CONTROL-GRAPH
type: epic
title: "Epic S16 Day 4: Architecture для Mission Control graph workspace и hybrid truth ownership (Issues #516/#519)"
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

# Epic S16 Day 4: Architecture для Mission Control graph workspace и hybrid truth ownership (Issues #516/#519)

## TL;DR
- Подготовлен architecture package Sprint S16 для Mission Control graph workspace: architecture decomposition, C4 overlays, ADR-0016 и alternatives по hybrid truth ownership, continuity state и bounded provider foundation.
- Зафиксирован ownership split для graph truth, provider inventory foundation, next-step surfaces, typed watermarks и continuity gaps.
- Подготовлен handover в `run:design` без premature transport/schema lock-in.

## Priority
- `P0`.

## Контекст
- Intake baseline: `#492` (`docs/delivery/epics/s16/epic-s16-day1-mission-control-graph-workspace-intake.md`).
- Vision baseline: `#496` (`docs/delivery/epics/s16/epic-s16-day2-mission-control-graph-workspace-vision.md`).
- PRD baseline: `#510` (`docs/delivery/epics/s16/epic-s16-day3-mission-control-graph-workspace-prd.md`, `docs/delivery/epics/s16/prd-s16-day3-mission-control-graph-workspace.md`).
- Текущий этап: `run:arch` в Issue `#516`.
- Scope этапа: только markdown-изменения.

## Architecture package
- `docs/architecture/initiatives/s16_mission_control_graph_workspace/README.md`
- `docs/architecture/initiatives/s16_mission_control_graph_workspace/architecture.md`
- `docs/architecture/initiatives/s16_mission_control_graph_workspace/c4_context.md`
- `docs/architecture/initiatives/s16_mission_control_graph_workspace/c4_container.md`
- `docs/architecture/adr/ADR-0016-mission-control-graph-workspace-hybrid-truth-and-continuity-ownership.md`
- `docs/architecture/alternatives/ALT-0008-mission-control-graph-workspace-hybrid-truth-boundaries.md`
- `docs/delivery/traceability/s16_mission_control_graph_workspace_history.md`

## Ключевые решения Stage
- `control-plane` становится единственным owner canonical graph truth, node classification `discussion/work_item/run/pull_request`, continuity gaps, typed metadata/watermarks и platform-canonical launch surfaces.
- `worker` закреплён за bounded provider inventory foundation: mirror freshness, recent-closed-history backfill, enrichment/reconcile jobs и lifecycle/background tasks; hybrid truth merge и next-step semantics не живут в worker.
- `api-gateway` и `web-console` остаются thin transport/visibility surfaces и не вычисляют graph truth, allowed next step или continuity completeness локально.
- Day4 закрепляет hybrid truth lifecycle `provider mirror -> graph truth -> workspace projection` и boundary core Wave 1 vs deferred contours.

## Context7 и внешний baseline
- Context7 lookup на Day4 не выполнялся: новые библиотеки и vendor integrations в scope отсутствуют.
- Локально проверены `gh issue create --help`, `gh pr create --help`, `gh pr edit --help` для non-interactive issue/PR automation.
- Новые внешние зависимости на этапе `run:arch` не требуются.

## Acceptance Criteria (Issue #516)
- [x] Подготовлен architecture package с service boundaries, ownership, C4 overlays, ADR и alternatives для Mission Control graph workspace.
- [x] Для core flows определены owner-сервисы и границы ответственности: graph truth, provider foundation, continuity gaps, typed metadata/watermarks, platform-canonical launch params и next-step surfaces.
- [x] Зафиксированы architecture-level trade-offs по hybrid truth merge, bounded provider foundation и continuity graph без premature transport/storage lock-in.
- [x] Обновлены delivery/traceability документы и package indexes.
- [x] Подготовлена follow-up issue `#519` для stage `run:design` без trigger-лейбла.

## Quality gates
| Gate | Что проверяем | Статус |
|---|---|---|
| `QG-S16-D4-01` Architecture completeness | Есть package `architecture + C4 + ADR + alternatives` | passed |
| `QG-S16-D4-02` Boundary integrity | Ownership за `control-plane` / `worker` / thin surfaces выражен явно | passed |
| `QG-S16-D4-03` Hybrid truth integrity | `provider mirror -> graph truth -> projection` описаны как разные typed layers | passed |
| `QG-S16-D4-04` Wave 1 discipline | Exact filters/nodes, dimmed semantics и deferred contours сохранены без scope drift | passed |
| `QG-S16-D4-05` Stage continuity | Подготовлена issue `#519` на `run:design` без trigger-лейбла | passed |

## Handover в `run:design`
- Следующий этап: `run:design`.
- Follow-up issue: `#519`.
- Trigger-лейбл на новую issue ставит Owner после review architecture package.
- На design-stage обязательно:
  - выпустить `design_doc`, `api_contract`, `data_model`, `migrations_policy`;
  - определить typed contracts для graph snapshot, node details, launch preview/launch, continuity gaps и watermarks;
  - зафиксировать canonical graph data model, mirror references, rollout/backfill/rollback notes и sequencing `control-plane -> worker -> api-gateway -> web-console`.
