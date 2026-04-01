---
doc_id: EPC-CK8S-S18-D5-MISSION-CONTROL-CANVAS
type: epic
title: "Epic S18 Day 5: Design для frontend-first Mission Control canvas и workflow UX на fake data (Issues #573/#579)"
status: in-review
owner_role: SA
created_at: 2026-04-01
updated_at: 2026-04-01
related_issues: [480, 561, 562, 563, 565, 567, 571, 573, 579]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-04-01-issue-573-design-epic"
---

# Epic S18 Day 5: Design для frontend-first Mission Control canvas и workflow UX на fake data (Issues #573/#579)

## TL;DR
- Подготовлен полный Day5 design package Sprint S18: `design_doc`, `api_contract`, `data_model`, `migrations_policy`.
- Зафиксированы implementation-ready contracts для frontend-only prototype: route shell, feature-local fake-data source, compact canvas nodes, explicit relations, drawer surfaces и deterministic workflow policy preview.
- Сохранены guardrails Sprint S18: fullscreen свободный canvas, taxonomy `Issue` / `PR` / `Run`, platform-safe actions only, repo-seed prompts как source of truth и отсутствие backend/API/DB changes.
- Создана follow-up issue `#579` для stage `run:plan` без trigger-лейбла.

## Priority
- `P0`.

## Контекст
- Intake baseline: `#562` (`docs/delivery/epics/s18/epic-s18-day1-mission-control-frontend-first-canvas-intake.md`).
- Vision baseline: `#565` (`docs/delivery/epics/s18/epic-s18-day2-mission-control-frontend-first-canvas-vision.md`).
- PRD baseline: `#567` (`docs/delivery/epics/s18/epic-s18-day3-mission-control-frontend-first-canvas-prd.md`, `docs/delivery/epics/s18/prd-s18-day3-mission-control-frontend-first-canvas.md`).
- Architecture baseline: `#571` (`docs/delivery/epics/s18/epic-s18-day4-mission-control-frontend-first-canvas-arch.md` + architecture package).
- Текущий этап: `run:design` в Issue `#573`.
- Scope этапа: только markdown-изменения.

## Design package
- `docs/architecture/initiatives/s18_mission_control_frontend_first_canvas/README.md`
- `docs/architecture/initiatives/s18_mission_control_frontend_first_canvas/design_doc.md`
- `docs/architecture/initiatives/s18_mission_control_frontend_first_canvas/api_contract.md`
- `docs/architecture/initiatives/s18_mission_control_frontend_first_canvas/data_model.md`
- `docs/architecture/initiatives/s18_mission_control_frontend_first_canvas/migrations_policy.md`
- `docs/delivery/traceability/s18_mission_control_frontend_first_canvas_history.md`

## Ключевые design-решения
- Route `MissionControlPage.vue` остаётся entry point, но data/state path переключается на explicit prototype submodule вместо current API/realtime implementation.
- Sprint S18 выбирает feature-local async-friendly source contract (`MissionControlPrototypeSource`) и fixture catalog как единственный data source для `run:dev`.
- Fullscreen canvas остаётся единственной primary surface:
  - no graph/list toggle;
  - no freshness/realtime chips;
  - no lane/column shell.
- Workflow editor остаётся local policy-preview UX:
  - structured toggles only;
  - generated `workflow-policy block`;
  - repo-seed source refs;
  - no free-form prompt editing and no provider mutation.
- Data model фиксирует только bundle-local scenario/node/relation/preset/ui-state entities; DB/runtime migrations explicitly отсутствуют.

## Context7 и внешний baseline
- Context7 lookup на Day5 не выполнялся: новые библиотеки и vendor integrations в scope отсутствуют.
- Новые внешние зависимости на этапе `run:design` не требуются.
- Follow-up issue `#579` создана напрямую через `gh issue create`; stage не ограничился только help/preview commands.

## Acceptance Criteria (Issue #573)
- [x] Подготовлен design package (`design_doc`, `api_contract`, `data_model`, `migrations_policy`).
- [x] Зафиксированы implementation-ready frontend contracts для canvas shell, fake-data source, drawer/workflow surfaces и prompt-source evidence.
- [x] Сохранены Sprint S18 guardrails: fullscreen canvas, taxonomy `Issue` / `PR` / `Run`, explicit relations, platform-safe actions only, repo-seed prompts как source of truth.
- [x] Backend rebuild `#563`, live provider sync, DB prompt editor, release-safety cockpit и waves `#524/#525` удержаны в deferred/later-wave scope.
- [x] Подготовлена follow-up issue `#579` для stage `run:plan`.

## Quality gates
| Gate | Что проверяем | Статус |
|---|---|---|
| `QG-S18-D5-01` Design completeness | Есть package `design_doc + api_contract + data_model + migrations_policy` | passed |
| `QG-S18-D5-02` Frontend isolation | Prototype остаётся bundle-local и не требует backend/API/schema changes | passed |
| `QG-S18-D5-03` Baseline fidelity | Fullscreen canvas, taxonomy `Issue/PR/Run`, compact nodes, explicit relations, drawer, toolbar и workflow preview сохранены без drift | passed |
| `QG-S18-D5-04` Prompt-policy discipline | Workflow preview остаётся deterministic generated block с repo-seed refs, без prompt editor semantics | passed |
| `QG-S18-D5-05` Deferred-scope discipline | `#563`, live sync, DB prompt editor, release-safety cockpit и waves `#524/#525` не стали blocking scope | passed |
| `QG-S18-D5-06` Stage continuity | Создана issue `#579` на `run:plan` без trigger-лейбла и с continuity `plan -> dev` | passed |

## Handover в `run:plan`
- Следующий этап: `run:plan`.
- Follow-up issue: `#579`.
- Trigger-лейбл на новую issue ставит Owner после review design package.
- На plan-stage обязательно:
  - разложить execution waves минимум на `route shell + prototype source -> canvas/drawer surfaces -> workflow preview and prompt-source evidence -> acceptance/demo evidence -> run:dev handover`;
  - зафиксировать quality gates, DoR/DoD и owner dependencies для frontend-only prototype;
  - продолжить цепочку `plan -> dev` без разрывов и без подмены scope на backend rebuild `#563`.
