---
doc_id: EPC-CK8S-S18-D3-MISSION-CONTROL-FRONTEND
type: epic
title: "Epic S18 Day 3: PRD для frontend-first Mission Control canvas и workflow UX на fake data (Issues #567/#571)"
status: in-review
owner_role: PM
created_at: 2026-03-27
updated_at: 2026-03-27
related_issues: [470, 480, 522, 523, 524, 525, 561, 562, 563, 565, 567, 571]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-27-issue-567-prd"
---

# Epic S18 Day 3: PRD для frontend-first Mission Control canvas и workflow UX на fake data (Issues #567/#571)

## TL;DR
- Подготовлен PRD-пакет Sprint S18 для frontend-first Mission Control canvas UX: `epic-s18-day3-mission-control-frontend-first-canvas-prd.md` и `prd-s18-day3-mission-control-frontend-first-canvas.md`.
- Зафиксированы user stories, FR/AC/NFR, scenario matrix, edge cases и expected evidence для owner/product lead walkthrough, operator navigation и workflow authoring UX на fake data.
- Принято продуктовое решение: Sprint S18 остаётся isolated fake-data prototype в `web-console`; fullscreen свободный canvas, taxonomy `Issue` / `PR` / `Run`, compact nodes, explicit relations, side panel/drawer, toolbar/controls, workflow editor UX и repo-seed prompts как source of truth являются locked baseline.
- Backend rebuild `#563`, live GitHub/provider sync, DB prompt editor, release-safety cockpit и waves `#524` / `#525` остаются отдельным или later-wave scope и не блокируют core MVP Sprint S18.
- Создана follow-up issue `#571` для stage `run:arch` без trigger-лейбла.

## Priority
- `P0`.

## Контекст
- Intake baseline: `#562` (`docs/delivery/epics/s18/epic-s18-day1-mission-control-frontend-first-canvas-intake.md`).
- Vision baseline: `#565` (`docs/delivery/epics/s18/epic-s18-day2-mission-control-frontend-first-canvas-vision.md`).
- Текущий этап: `run:prd` в Issue `#567`.
- Следующий этап: `run:arch` в Issue `#571`.
- Входной product contract:
  - `docs/delivery/sprints/s18/sprint_s18_mission_control_frontend_first_canvas_fake_data.md`;
  - `docs/product/requirements_machine_driven.md`;
  - `docs/product/agents_operating_model.md`;
  - `docs/product/labels_and_trigger_policy.md`;
  - `docs/product/stage_process_model.md`;
  - `docs/architecture/prompt_templates_policy.md`.

## Scope
### In scope
- Формализация user stories, FR/AC/NFR, scenario matrix и expected evidence для frontend-first Mission Control canvas и workflow UX на fake data.
- Приоритизация волн `core canvas comprehension -> workflow policy preview -> deferred backend/live integrations`.
- Фиксация продуктовых guardrails для fullscreen свободного canvas, taxonomy `Issue` / `PR` / `Run`, compact nodes, explicit relations, side panel/drawer, toolbar/controls, workflow editor UX, platform-safe actions only и repo-seed prompts как source of truth.
- Явный handover в `run:arch` с перечнем продуктовых инвариантов, которые нельзя потерять при переводе в architecture package.
- Синхронизация traceability (`issue_map`, `delivery_plan`, sprint/epic docs, history package, indexes).

### Out of scope
- Кодовая реализация, state/data/API contracts и architecture lock-in до `run:arch` / `run:design`.
- Backend rebuild `#563` и live provider sync как blocking requirement текущего stage.
- DB prompt editor, live GitHub mutation path и release-safety cockpit.
- Автоматический late-stage transition `run:dev -> run:qa -> run:release -> run:postdeploy -> run:ops` внутри Sprint S18.

## PRD package
- `docs/delivery/epics/s18/epic-s18-day3-mission-control-frontend-first-canvas-prd.md`
- `docs/delivery/epics/s18/prd-s18-day3-mission-control-frontend-first-canvas.md`
- `docs/delivery/traceability/s18_mission_control_frontend_first_canvas_history.md`

## Wave priorities

| Wave | Приоритет | Scope | Exit signal |
|---|---|---|---|
| Wave 1 | `P0` | Isolated fake-data prototype в `web-console`: fullscreen canvas, taxonomy `Issue` / `PR` / `Run`, compact nodes, explicit relations, side panel/drawer, toolbar/controls, workflow policy preview | Owner может за одну сессию объяснить состояние и safe next step по 2-3 инициативам без lane/column shell и без live provider path |
| Wave 2 | `P0` | Workflow authoring UX и operator details/timeline flow: policy preview, relation semantics, missing-link transparency, action safety evidence | Workflow semantics понятны без prompt editor и без live mutation path, а operator не теряет контекст при переключении инициатив |
| Wave 3 | `P1` (deferred) | Backend rebuild `#563`, live GitHub/provider sync, DB prompt editor, release-safety cockpit, waves `#524` / `#525` | Эти направления двигаются только после owner-approved architecture/design без reopening core Sprint S18 baseline |

## Acceptance criteria (Issue #567)
- [x] Подготовлен PRD-артефакт frontend-first Mission Control canvas UX и синхронизирован в traceability-документах.
- [x] Для core flows зафиксированы user stories, FR/AC/NFR, scenario matrix, edge cases и expected evidence.
- [x] Явно сохранены locked baselines: fullscreen свободный canvas, taxonomy `Issue` / `PR` / `Run`, compact nodes, explicit relations, side panel/drawer, toolbar/controls, workflow editor UX на fake data, platform-safe actions only, repo-seed prompts как source of truth и `run:dev` как isolated `web-console` prototype.
- [x] Wave priorities сформулированы без смешения core MVP и deferred backend/live integration scope.
- [x] Создана follow-up issue `#571` для stage `run:arch` без trigger-лейбла.

## Quality gates

| Gate | Что проверяем | Статус |
|---|---|---|
| QG-S18-D3-01 PRD completeness | User stories, FR/AC/NFR, scenario matrix и expected evidence покрывают owner, operator и workflow-authoring paths | passed |
| QG-S18-D3-02 Locked baseline preserved | Fullscreen canvas, taxonomy `Issue` / `PR` / `Run`, compact nodes, explicit relations, side panel/drawer, toolbar/controls и fake-data workflow UX сохранены без reopening | passed |
| QG-S18-D3-03 Deferred-scope discipline | Backend rebuild `#563`, live provider sync, DB prompt editor, release-safety cockpit и waves `#524` / `#525` не смешаны с core MVP Sprint S18 | passed |
| QG-S18-D3-04 Stage continuity | Создана issue `#571` для `run:arch` без trigger-лейбла и с continuity-требованием `arch -> design -> plan -> dev` | passed |
| QG-S18-D3-05 Policy compliance | Изменены только markdown-артефакты | passed |

## Handover в `run:arch`
- Следующий этап: `run:arch`.
- Follow-up issue: `#571`.
- Trigger-лейбл `run:arch` на issue `#571` ставит Owner после review PRD-пакета.
- Обязательные выходы архитектурного этапа:
  - service boundaries и ownership split для isolated fake-data prototype в `web-console` и handover boundary к backend rebuild `#563`;
  - architecture alternatives по relation semantics, fake-data isolation, workflow policy preview и future backend sync boundary;
  - фиксация, как сохраняются locked baselines Sprint S18 без возврата к lane/column shell, live mutation path или prompt editor semantics;
  - отдельная issue на `run:design` без trigger-лейбла после завершения `run:arch` с явным continuity-требованием продолжить цепочку `design -> plan -> dev`.

## Открытые риски и допущения

| Type | ID | Описание | Статус |
|---|---|---|---|
| risk | `RSK-567-01` | Свободный canvas может остаться visually noisy и потерять понятность для 2-3 инициатив | open |
| risk | `RSK-567-02` | Workflow editor может снова расползтись в prompt editor или live mutation path | open |
| risk | `RSK-567-03` | Fake-data prototype может не дать достаточно ясного handover signal для backend rebuild `#563`, если expected evidence окажется слишком абстрактным | open |
| assumption | `ASM-567-01` | Owner может валидировать новый Mission Control UX на fake data до появления нового backend foundation | accepted |
| assumption | `ASM-567-02` | Taxonomy `Issue` / `PR` / `Run` достаточна для core wave Sprint S18 и не требует возврата к старой S16 taxonomy | accepted |
