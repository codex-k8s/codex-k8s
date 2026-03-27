---
doc_id: EPC-CK8S-S18-D4-MISSION-CONTROL-FRONTEND
type: epic
title: "Epic S18 Day 4: Architecture для frontend-first Mission Control canvas и workflow UX на fake data (Issues #571/#573)"
status: in-review
owner_role: SA
created_at: 2026-03-27
updated_at: 2026-03-27
related_issues: [480, 561, 562, 563, 565, 567, 571, 573]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-27-issue-571-arch"
---

# Epic S18 Day 4: Architecture для frontend-first Mission Control canvas и workflow UX на fake data (Issues #571/#573)

## TL;DR
- Подготовлен architecture package Sprint S18: `README`, `architecture`, `c4_context`, `c4_container`, ADR и alternatives.
- Зафиксировано решение Day4: Sprint S18 остаётся isolated fake-data prototype в `web-console`; `api-gateway`, `control-plane`, `worker` и `PostgreSQL` не становятся новым Mission Control truth-path до запуска отдельного backend rebuild `#563`.
- Repo-seed prompts сохранены как source of truth, workflow editor остаётся policy-preview UX с deterministic `workflow-policy block`, live provider mutation path и DB prompt editor не вводятся.
- Создана follow-up issue `#573` для stage `run:design` без trigger-лейбла.

## Priority
- `P0`.

## Контекст
- Intake baseline: `#562` (`docs/delivery/epics/s18/epic-s18-day1-mission-control-frontend-first-canvas-intake.md`).
- Vision baseline: `#565` (`docs/delivery/epics/s18/epic-s18-day2-mission-control-frontend-first-canvas-vision.md`).
- PRD baseline: `#567` (`docs/delivery/epics/s18/epic-s18-day3-mission-control-frontend-first-canvas-prd.md`, `docs/delivery/epics/s18/prd-s18-day3-mission-control-frontend-first-canvas.md`).
- Текущий этап: `run:arch` в Issue `#571`.
- Scope этапа: только markdown-изменения.

## Architecture package
- `docs/architecture/initiatives/s18_mission_control_frontend_first_canvas/README.md`
- `docs/architecture/initiatives/s18_mission_control_frontend_first_canvas/architecture.md`
- `docs/architecture/initiatives/s18_mission_control_frontend_first_canvas/c4_context.md`
- `docs/architecture/initiatives/s18_mission_control_frontend_first_canvas/c4_container.md`
- `docs/architecture/adr/ADR-0018-mission-control-frontend-first-prototype-and-backend-handover-boundary.md`
- `docs/architecture/alternatives/ALT-0010-mission-control-frontend-first-prototype-boundaries.md`
- `docs/delivery/traceability/s18_mission_control_frontend_first_canvas_history.md`

## Ключевые решения Stage
- `web-console` становится единственным owner fake-data scenario catalog, canvas projection, drawer/toolbar state и workflow preview UX для Sprint S18.
- `api-gateway` остаётся thin-edge boundary без новой Mission Control доменной логики и без прямого участия в prototype state.
- `control-plane` и `worker` закреплены как deferred handover owners для backend rebuild `#563`, а не как hidden prerequisite Sprint S18.
- Repo-seed prompts и `prompt_templates_policy` остаются source of truth; workflow preview разрешён только как deterministic `workflow-policy block`.
- Backend rebuild `#563`, live provider sync, DB prompt editor, release-safety cockpit и waves `#524/#525` остаются deferred/later-wave scope.

## Context7 и внешний baseline
- Context7 lookup на Day4 не выполнялся: новые библиотеки и vendor integrations в scope отсутствуют.
- Локально проверены `gh issue create --help`, `gh pr create --help`, `gh pr edit --help` для non-interactive issue/PR automation.
- Новые внешние зависимости на этапе `run:arch` не требуются.

## Acceptance Criteria (Issue #571)
- [x] Подготовлен architecture package с service boundaries, ownership split, C4 overlays, ADR и alternatives для Sprint S18.
- [x] Явно отделены isolated fake-data prototype и future backend rebuild `#563`.
- [x] Locked baseline Sprint S18 сохранён: fullscreen canvas, taxonomy `Issue` / `PR` / `Run`, compact nodes, explicit relations, drawer, toolbar, workflow preview, platform-safe actions only, repo-seed prompts как source of truth.
- [x] Deferred/later-wave scope не смешан с core architecture package.
- [x] Подготовлена follow-up issue `#573` для stage `run:design` без trigger-лейбла.

## Quality gates
| Gate | Что проверяем | Статус |
|---|---|---|
| `QG-S18-D4-01` Architecture completeness | Есть package `README + architecture + C4 + ADR + alternatives` | passed |
| `QG-S18-D4-02` Boundary integrity | Ownership split `web-console` / thin-edge / deferred backend owners выражен явно | passed |
| `QG-S18-D4-03` Baseline fidelity | Locked Sprint S18 baseline сохранён без reopening S16 assumptions | passed |
| `QG-S18-D4-04` Deferred-scope discipline | `#563`, live sync, DB prompt editor и waves `#524/#525` не стали blocking requirement | passed |
| `QG-S18-D4-05` Stage continuity | Создана issue `#573` на `run:design` без trigger-лейбла | passed |

## Handover в `run:design`
- Следующий этап: `run:design`.
- Follow-up issue: `#573`.
- Trigger-лейбл на новую issue ставит Owner после review architecture package.
- На design-stage обязательно:
  - выпустить `design_doc`, `api_contract`, `data_model`, `migrations_policy`;
  - зафиксировать fake-data state slices, canvas interaction rules, drawer/toolbar behavior и workflow preview contracts;
  - описать replacement seam к backend rebuild `#563` без reopening Wave 1 baseline;
  - сохранить continuity `design -> plan -> dev` без разрывов.
