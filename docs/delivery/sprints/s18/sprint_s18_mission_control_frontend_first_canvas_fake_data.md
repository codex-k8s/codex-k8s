---
doc_id: SPR-CK8S-0018
type: sprint-plan
title: "Sprint S18: Frontend-first Mission Control canvas UX on fake data (Issue #562)"
status: in-review
owner_role: PM
created_at: 2026-03-26
updated_at: 2026-03-27
related_issues: [470, 480, 522, 523, 524, 525, 561, 562, 563, 565, 567, 571, 573]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-26-issue-562-intake"
---

# Sprint S18: Frontend-first Mission Control canvas UX on fake data (Issue #562)

## TL;DR
- Цель спринта: сначала утвердить новый Mission Control UX как isolated `web-console` prototype на fake data, и только потом запускать backend rebuild в issue `#563`.
- Intake stage в issue `#562` зафиксировал controlled reset после doc-reset `#561`: fullscreen свободный canvas, минимальная taxonomy `Issue` / `PR` / `Run`, compact nodes, side panel/drawer, toolbar/controls и workflow editor UX как часть frontend-first контура.
- Vision stage в issue `#565` закрепил mission, north star, persona outcomes, KPI/guardrails и wave boundaries для canvas-first UX без reopening Day1 baseline.
- PRD stage в issue `#567` формализовал user stories, FR/AC/NFR, scenario matrix и expected evidence для owner/product lead path, operator path и workflow policy preview; continuity issue `#571` создана для `run:arch`.
- Architecture stage в issue `#571` закрепил `web-console` как owner isolated fake-data prototype, отделил Sprint S18 от backend rebuild `#563` и создал continuity issue `#573` для `run:design`.
- Sprint S18 намеренно не продолжает старый S16 baseline `lane/column` и не пытается “подкрасить” отклонённый graph shell.
- Prompt policy не меняется: repo-seed prompts остаются каноничными, workflow logic допускается только как deterministic generated `workflow-policy block`, без DB prompt editor.
- Backend follow-up `#563` остаётся отдельной задачей после owner approval результата Sprint S18, а следующий owner-managed handover идёт через issue `#573` на `run:design`.

## Scope спринта
### In scope
- Полная doc-stage цепочка `intake -> vision -> prd -> arch -> design -> plan -> dev` для frontend-first Mission Control UX.
- Формализация и затем реализация isolated fake-data prototype в `services/staff/web-console/`.
- Новый Wave 1 UX baseline:
  - fullscreen свободный canvas без lane/column shell;
  - минимальная node taxonomy `Issue`, `PR`, `Run`;
  - compact-but-readable nodes для 2-3 инициатив одновременно;
  - явные связи node-to-node вместо визуальной имитации колонок;
  - side panel/drawer для details, timeline и actions;
  - toolbar/controls для переключения инициатив, фильтров и workflow interaction;
  - workflow editor UX на fake data с platform-safe actions only.
- Явный guardrail, что `run:dev` в рамках Sprint S18 делает только isolated prototype и не открывает обязательную late-stage цепочку `qa -> release -> postdeploy -> ops`.

### Out of scope
- Backend rebuild, provider mirror redesign и transport re-foundation: это отдельный Sprint/Issue `#563`.
- Live GitHub mutation path, live provider sync и попытка использовать текущий Mission Control API как source of truth для prototype.
- DB prompt editor или любой free-form prompt storage.
- Финальный release-safety cockpit UI, readiness gate Sprint S13 и соседние execution waves `#524` / `#525`.
- Автоматический handover в `run:qa` после завершения `run:dev` в этой инициативе.

## Рекомендованный launch profile
- Базовый launch profile: `new-service`.
- Обоснование:
  - инициатива меняет product contour Mission Control и затрагивает continuity между несколькими stage;
  - frontend-first scope не отменяет того, что без `vision`, `arch` и `design` UX снова начнёт подчиняться старой data/model semantics;
  - нужен отдельный owner-reviewed baseline до старта backend rebuild и до разблокировки зависимых UI/release потоков.
- Целевая continuity-цепочка:
  `#562 (intake) -> #565 (vision) -> #567 (prd) -> architecture issue -> design issue -> plan issue -> dev prototype`.

## План этапов и handover

| Stage | Основной артефакт | Целевая роль | Правило выхода |
|---|---|---|---|
| Intake (`#562`) | Problem/Brief/Scope/Constraints + intake AC | `pm` | Зафиксирован frontend-first baseline и создана issue `#565` для `run:vision` |
| Vision (`#565`) | Mission, north star, persona outcomes, KPI/guardrails, wave boundaries | `pm` | Подтверждён vision baseline и создана issue `#567` для `run:prd` |
| PRD (`#567`) | User stories, FR/AC/NFR, scenario matrix и expected evidence | `pm` + `sa` | Зафиксирован product contract и создана issue `#571` для `run:arch` |
| Architecture (`#571`) | Prototype isolation, ownership split, future handover в backend rebuild | `sa` | Подтверждены архитектурные границы и создана issue `#573` для `run:design` |
| Design (`#573`) | UI/data/interaction/design package для fake-data prototype | `sa` + `qa` | Подготовлен implementation-ready design package и создана issue для `run:plan` |
| Plan (TBD) | Delivery waves, feedback loops, DoR/DoD, owner-managed `run:dev` handover | `em` + `km` | Сформирован execution package для isolated prototype |
| Development (TBD) | Isolated `web-console` prototype на fake data | `dev` | Открыт PR с prototype-реализацией; дальнейшие backend/late-stage задачи остаются отдельными initiative flows |

## Guardrails спринта
- Fullscreen canvas обязателен; возврат к lane/column shell или обязательной `root-group/column/stack` иерархии не допускается без нового owner-решения.
- Wave 1 node taxonomy остаётся минимальной: `Issue`, `PR`, `Run`.
- Workflow editor остаётся частью Mission Control UX, но работает только на fake data и не становится live mutation path.
- Platform-safe actions only: этот sprint не должен напрямую мутировать GitHub/provider state.
- Repo-seed prompts остаются source of truth; workflow logic допускается только как structured policy layer, без DB prompt editor.
- Sprint S18 не закрывает backend foundation: issue `#563` стартует только после owner approval результата этого frontend-first спринта.
- Sequencing из rethink `#561` сохраняется:
  - `#522` и `#523` можно продолжать независимо;
  - `#524` и `#525` нельзя запускать до owner approval нового frontend baseline;
  - `#470` нельзя использовать для фиксации финального cockpit UI до завершения Sprint S18.

## Handover
- Текущий stage in-review: `run:arch` в issue `#571`.
- Актуальный package:
  - `docs/delivery/sprints/s18/sprint_s18_mission_control_frontend_first_canvas_fake_data.md`;
  - `docs/delivery/epics/s18/epic_s18.md`;
  - `docs/delivery/epics/s18/epic-s18-day1-mission-control-frontend-first-canvas-intake.md`;
  - `docs/delivery/epics/s18/epic-s18-day2-mission-control-frontend-first-canvas-vision.md`;
  - `docs/delivery/epics/s18/epic-s18-day3-mission-control-frontend-first-canvas-prd.md`;
  - `docs/delivery/epics/s18/epic-s18-day4-mission-control-frontend-first-canvas-arch.md`;
  - `docs/delivery/epics/s18/prd-s18-day3-mission-control-frontend-first-canvas.md`;
  - `docs/architecture/initiatives/s18_mission_control_frontend_first_canvas/README.md`;
  - `docs/architecture/initiatives/s18_mission_control_frontend_first_canvas/architecture.md`;
  - `docs/architecture/initiatives/s18_mission_control_frontend_first_canvas/c4_context.md`;
  - `docs/architecture/initiatives/s18_mission_control_frontend_first_canvas/c4_container.md`;
  - `docs/architecture/adr/ADR-0018-mission-control-frontend-first-prototype-and-backend-handover-boundary.md`;
  - `docs/architecture/alternatives/ALT-0010-mission-control-frontend-first-prototype-boundaries.md`;
  - `docs/delivery/traceability/s18_mission_control_frontend_first_canvas_history.md`.
- Следующий stage: `run:design` через issue `#573`.
- До завершения следующего stage нельзя потерять следующие Day1/Day2/Day3 decisions:
  - сначала утверждается UX baseline на fake data, потом backend rebuild;
  - fullscreen свободный canvas без lane/column shell;
  - Wave 1 taxonomy `Issue`, `PR`, `Run`;
  - compact nodes, explicit relations, side panel/drawer и toolbar/controls обязательны;
  - workflow editor UX входит в scope Sprint S18, но работает только на fake data и с platform-safe actions;
  - owner/product lead path, operator path и workflow policy preview должны оставаться core PRD contract;
  - north star инициативы = owner-ready canvas walkthrough для 2-3 инициатив без возврата к lane/column shell;
  - `run:dev` внутри этого спринта ограничен isolated `web-console` prototype;
  - repo-seed prompts остаются каноничными, free-form DB prompt storage не вводится;
  - после `run:dev` обязательная late-stage цепочка внутри этой инициативы не запускается автоматически.
- Trigger-лейбл для issue `#573` не ставится автоматически и остаётся owner-managed переходом после review architecture package.
