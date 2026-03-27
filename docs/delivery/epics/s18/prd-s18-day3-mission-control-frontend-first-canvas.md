---
doc_id: PRD-CK8S-S18-D3-MISSION-CONTROL-FRONTEND
type: prd
title: "Frontend-first Mission Control canvas and workflow UX on fake data — PRD Sprint S18 Day 3"
status: in-review
owner_role: PM
created_at: 2026-03-27
updated_at: 2026-03-27
related_issues: [470, 480, 522, 523, 524, 525, 561, 562, 563, 565, 567, 571]
related_prs: []
related_docsets:
  - docs/delivery/sprints/s18/sprint_s18_mission_control_frontend_first_canvas_fake_data.md
  - docs/delivery/epics/s18/epic_s18.md
  - docs/delivery/issue_map.md
  - docs/delivery/requirements_traceability.md
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-27-issue-567-prd"
---

# PRD: Frontend-first Mission Control canvas and workflow UX on fake data

## TL;DR
- Что строим: isolated `web-console` prototype на fake data, который делает Mission Control fullscreen canvas-first workspace для 2-3 инициатив и включает workflow policy preview без live mutation path.
- Для кого: owner / product lead, execution lead / operator и workflow authoring path внутри Sprint S18.
- Почему: после rethink `#561` и locked baseline `#562/#565` нужен проверяемый PRD, чтобы не смешать новый UX с rejected S16 shell, backend rebuild `#563` и live provider semantics.
- MVP: fullscreen свободный canvas, taxonomy `Issue` / `PR` / `Run`, compact nodes, explicit relations, side panel/drawer, toolbar/controls, workflow editor UX на fake data, platform-safe actions only и repo-seed prompts как source of truth.
- Критерии успеха: owner понимает состояние и следующий безопасный шаг минимум по двум инициативам за одну сессию, operator не теряет details/timeline при переключении фокуса, workflow policy preview читается без prompt editor и без live provider path.

## Проблема и цель
- Problem statement:
  - текущий Mission Control после Sprint S16 остаётся связанным с rejected lane/column shell и не должен снова становиться неявным source of truth для нового UX;
  - без отдельного PRD-stage команда рискует смешать core canvas comprehension, workflow policy preview, live provider mutation path и backend rebuild `#563` в один размытый scope;
  - если Sprint S18 не получит проверяемые user stories, FR/AC/NFR и expected evidence, следующий `run:arch` начнёт спорить об architecture/design без продуктового контракта.
- Цели:
  - закрепить Sprint S18 как frontend-first Mission Control UX baseline на fake data;
  - формализовать, как owner/product lead, operator и workflow author используют fullscreen canvas, drawer и toolbar без lane/column shell;
  - зафиксировать workflow editor как policy-shaping UX на fake data, а не как prompt editor и не как live mutation path;
  - отделить core Wave 1 Sprint S18 от backend rebuild `#563` и других deferred/later-wave направлений.
- Почему сейчас:
  - intake `#562` и vision `#565` уже закрепили sequencing, north star, guardrails и locked baseline;
  - launch profile `new-service` требует обязательный `run:prd` перед `run:arch`;
  - без PRD-stage новый UX снова рискует быть подчинён старой S16 модели или случайным backend assumptions.

## Зафиксированные продуктовые решения
- `D-567-01`: Sprint S18 остаётся frontend-first flow на fake data; backend rebuild `#563` запускается только после owner approval результата этого UX contour.
- `D-567-02`: Mission Control для Sprint S18 остаётся fullscreen свободным canvas-first workspace, а не lane/column shell и не board/list refresh.
- `D-567-03`: Wave 1 taxonomy остаётся минимальной: `Issue`, `PR`, `Run`.
- `D-567-04`: Compact nodes, explicit relations, side panel/drawer и toolbar/controls являются обязательными surfaces core prototype.
- `D-567-05`: Workflow editor остаётся частью Mission Control UX, но работает только на fake data и показывает deterministic `workflow-policy block`.
- `D-567-06`: Repo-seed prompts остаются source of truth; DB prompt editor и free-form prompt storage не входят в Sprint S18.
- `D-567-07`: Platform-safe actions only: prototype не мутирует live GitHub/provider state и не требует live provider sync как блокирующий foundation.
- `D-567-08`: `run:dev` в рамках Sprint S18 ограничен isolated `web-console` prototype и не открывает обязательную цепочку `qa -> release -> postdeploy -> ops`.
- `D-567-09`: Backend rebuild `#563`, release-safety cockpit, waves `#524/#525` и related readiness contours остаются отдельным или later-wave scope.
- `D-567-10`: Если следующий stage потребует вернуть S16 taxonomy, lane/column shell, live mutation path или prompt editor semantics, такой drift считается blocking scope violation.

## Scope boundaries
### In scope
- Isolated Mission Control prototype в `services/staff/web-console/` на fake data.
- Fullscreen свободный canvas для 2-3 инициатив одновременно.
- Minimal taxonomy `Issue`, `PR`, `Run`.
- Compact nodes и явные node-to-node relations.
- Side panel/drawer для details, timeline, relations и safe action context.
- Toolbar/controls для выбора инициатив, фильтров, density/focus и workflow interaction mode.
- Workflow editor UX на fake data с deterministic policy preview и platform-safe actions only.
- Product evidence, достаточный для handover в `run:arch`, `run:design`, `run:plan` и затем `run:dev`.

### Out of scope
- Backend rebuild `#563`, inventory mirror redesign и live provider sync.
- DB prompt editor, prompt storage lifecycle и raw prompt authoring UX.
- Live GitHub/provider mutation path.
- Финальный release-safety cockpit и readiness waves `#524` / `#525`.
- Обязательный automatic continuation в `run:qa -> run:release -> run:postdeploy -> run:ops` внутри Sprint S18.

## Пользователи / персоны

| Persona | Основная работа | Что считает успехом |
|---|---|---|
| Owner / product lead | За одну сессию понять состояние 2-3 инициатив и выбрать следующий безопасный шаг | Canvas даёт ясную картину без lane/column shell и без ручного GitHub detour |
| Execution lead / operator | Переключаться между инициативами, читать связи `Issue` / `PR` / `Run`, открывать details/timeline и не терять контекст | Drawer и toolbar остаются primary navigation surfaces, relation semantics понятны, missing links видны явно |
| Workflow author | Исследовать workflow semantics и policy preview без погружения в prompt internals и без live mutation path | Workflow preview читается как продуктовый UX, а не как prompt editor или backend console |

## User stories и wave priorities

| Story ID | История | Wave | Приоритет |
|---|---|---|---|
| `S18-US-01` | Как owner / product lead, я хочу открыть Mission Control и увидеть 2-3 инициативы на одном fullscreen canvas, чтобы быстро понять состояние и safe next step без lane/column shell | Wave 1 | `P0` |
| `S18-US-02` | Как owner / product lead, я хочу открыть details и timeline выбранной инициативы через drawer, чтобы не терять контекст canvas | Wave 1 | `P0` |
| `S18-US-03` | Как execution lead / operator, я хочу переключать инициативы и читать связи `Issue` / `PR` / `Run`, чтобы видеть continuity без скрытой иерархии | Wave 1 | `P0` |
| `S18-US-04` | Как execution lead / operator, я хочу, чтобы missing relation или ambiguous next step были видны явно, а не терялись в UI | Wave 2 | `P0` |
| `S18-US-05` | Как workflow author, я хочу preview policy semantics на fake data, чтобы проверить UX без prompt editor и без live mutation path | Wave 2 | `P0` |
| `S18-US-06` | Как продуктовая команда, я хочу сохранить prototype внутри isolated `web-console` scope, чтобы не запускать backend rebuild `#563` раньше owner approval | Wave 1 | `P0` |
| `S18-US-07` | Как команда delivery, я хочу оставить backend rebuild, live provider sync, DB prompt editor и late-stage flow за пределами core MVP, чтобы не размывать Sprint S18 | Wave 3 | `P1` |

### Wave sequencing
- Wave 1 (`P0`):
  - fullscreen canvas;
  - taxonomy `Issue` / `PR` / `Run`;
  - compact nodes и explicit relations;
  - side panel/drawer и toolbar/controls;
  - isolated fake-data prototype.
- Wave 2 (`P0`):
  - workflow policy preview;
  - details/timeline continuity;
  - missing-link transparency;
  - expected evidence по operator/workflow paths.
- Wave 3 (`P1`, deferred):
  - backend rebuild `#563`;
  - live provider sync/mutation path;
  - DB prompt editor;
  - release-safety cockpit;
  - waves `#524/#525`;
  - late-stage flow после `run:dev`.

## Functional Requirements

| ID | Требование |
|---|---|
| `FR-567-01` | Mission Control должен быть fullscreen свободным canvas-first workspace для Sprint S18, а не lane/column shell или board/list-only refresh. |
| `FR-567-02` | Prototype должен показывать 2-3 инициативы одновременно через минимальную taxonomy `Issue`, `PR`, `Run`. |
| `FR-567-03` | Node language должна оставаться compact-but-readable и использовать explicit relations вместо визуальной имитации колонок или вложенных lane-групп. |
| `FR-567-04` | Side panel/drawer должен быть обязательной surface для details, timeline, relations и action context выбранной инициативы. |
| `FR-567-05` | Toolbar/controls должны позволять переключать инициативы, управлять focus/density и входить в workflow interaction mode без выхода из canvas. |
| `FR-567-06` | Workflow editor должен работать только на fake data и показывать deterministic policy preview без live provider mutation path. |
| `FR-567-07` | Repo-seed prompts должны оставаться source of truth для workflow semantics; DB prompt editor и raw prompt authoring не допускаются в core Sprint S18. |
| `FR-567-08` | Prototype должен использовать только platform-safe actions и не делать live GitHub/provider mutations в рамках Sprint S18. |
| `FR-567-09` | `run:dev` для Sprint S18 должен быть ограничен isolated `web-console` prototype и не обязан автоматически открывать late-stage delivery chain. |
| `FR-567-10` | Backend rebuild `#563`, live provider sync, release-safety cockpit и waves `#524/#525` должны оставаться вне core MVP и не блокировать Day3 package. |

## Acceptance Criteria (Given/When/Then)

### `AC-567-01` Fullscreen canvas и multi-initiative comprehension
- Given у пользователя есть 2-3 активные инициативы,
- When он открывает Mission Control prototype,
- Then он видит fullscreen свободный canvas без lane/column shell и может объяснить состояние минимум двух инициатив без ручного GitHub hunting.
- Expected evidence: owner walkthrough по 2-3 инициативам + product review note.

### `AC-567-02` Taxonomy и relation semantics
- Given на canvas присутствуют сущности инициативы,
- When пользователь читает node cards и связи между ними,
- Then он видит только taxonomy `Issue`, `PR`, `Run`, а relation semantics остаются явными и не скрываются за nested lane/group model.
- Expected evidence: canvas review с relation map и visual acceptance capture.

### `AC-567-03` Drawer/details continuity
- Given пользователь выбрал инициативу или node,
- When он открывает details,
- Then side panel/drawer показывает details, timeline и action context, не ломая общий контекст canvas.
- Expected evidence: operator walkthrough + interaction review.

### `AC-567-04` Toolbar и переключение инициатив
- Given пользователь работает с несколькими инициативами,
- When он меняет focus, density или выбранную инициативу,
- Then toolbar/controls позволяют сделать это без потери контекста drawer и без ухода в отдельный shell.
- Expected evidence: multi-initiative navigation walkthrough.

### `AC-567-05` Workflow policy preview на fake data
- Given workflow author открывает workflow interaction mode,
- When он меняет policy-facing input в prototype,
- Then он получает deterministic policy preview на fake data без prompt editor, raw prompt diff и live provider mutation.
- Expected evidence: workflow walkthrough + scope review note.

### `AC-567-06` Platform-safe actions and isolation
- Given пользователь инициирует действие из prototype,
- When действие требует внешнюю мутацию или backend dependency,
- Then prototype либо остаётся в fake-data semantics, либо явно показывает, что live path вне scope Sprint S18.
- Expected evidence: action-safety checklist + negative live-path scenario.

### `AC-567-07` Isolated prototype scope
- Given Sprint S18 дошёл до `run:dev`,
- When команда оценивает результат,
- Then реализуемый результат остаётся isolated `web-console` prototype и не требует автоматического handover в `run:qa -> run:release -> run:postdeploy -> run:ops`.
- Expected evidence: scope review + delivery handover note.

### `AC-567-08` Deferred scope does not block core MVP
- Given backend rebuild `#563`, live provider sync, DB prompt editor и release-safety cockpit отсутствуют,
- When оценивается готовность core Sprint S18,
- Then Mission Control остаётся валидным frontend-first prototype и эти направления не считаются blocking prerequisites.
- Expected evidence: wave-based release readiness checklist.

## Scenario matrix

| ID | Сценарий | Обязательное поведение | Expected evidence |
|---|---|---|---|
| `SC-567-01` | Owner смотрит 2-3 инициативы на одном canvas | Видит состояние и safe next step минимум по двум инициативам без lane/column shell | Owner walkthrough + review note |
| `SC-567-02` | Operator переключается между инициативами | Toolbar и drawer сохраняют continuity, а связи `Issue` / `PR` / `Run` остаются читаемыми | Operator navigation review |
| `SC-567-03` | Operator открывает details/timeline выбранной инициативы | Drawer показывает details, timeline и relation context без потери общего вида canvas | Drawer walkthrough |
| `SC-567-04` | Workflow author проверяет policy preview | Workflow mode остаётся fake-data UX и не превращается в prompt editor или live action console | Workflow UX walkthrough |
| `SC-567-05` | Инициатива имеет missing relation или ambiguous next step | Prototype явно показывает пробел relation semantics, не скрывая проблему | Missing-link review |
| `SC-567-06` | Пользователь пытается запустить live provider mutation | Сценарий классифицируется как out of scope и не становится частью core prototype behavior | Action safety review |
| `SC-567-07` | Команда оценивает dev handover | `run:dev` трактуется как isolated prototype без обязательного late-stage auto-continue | Delivery handover review |

## Edge cases и non-happy paths

| ID | Сценарий | Ожидаемое поведение | Evidence |
|---|---|---|---|
| `EC-567-01` | На canvas появляется потребность вернуть lane/column shell | Это считается blocking scope drift и требует нового owner-решения | scope-drift review |
| `EC-567-02` | Taxonomy `Issue` / `PR` / `Run` кажется недостаточной и кто-то предлагает вернуть S16 model | Drift фиксируется как out of scope до owner-решения, а core Sprint S18 не переоткрывает старую taxonomy | taxonomy review |
| `EC-567-03` | Workflow preview требует raw prompt editing | Сценарий блокируется: repo-seed prompts остаются source of truth, prompt editor не вводится | workflow policy review |
| `EC-567-04` | Пользователь пытается выполнить live mutation из prototype | Prototype показывает, что действие вне scope Sprint S18, и не нормализует live path как happy-path | action safety evidence |
| `EC-567-05` | Кто-то пытается считать `#563` или `#524/#525` blocking prerequisite для PRD acceptance | Сценарий классифицируется как deferred/later-wave drift и не блокирует core MVP | wave-boundary review |

## Non-Goals
- Возвращать lane/column shell или nested root-group модель в core Sprint S18.
- Делать backend rebuild `#563` частью current PRD scope.
- Вводить DB prompt editor или raw prompt authoring.
- Делать live GitHub/provider mutation path частью core prototype.
- Считать late-stage delivery flow обязательным продолжением Sprint S18 dev-result.

## NFR draft для handover в architecture

| ID | Требование | Как измеряем / проверяем |
|---|---|---|
| `NFR-567-01` | `PM-565-01` Orientation time p75 должен оставаться `<= 90 секунд` для owner walkthrough | pilot session timing + product review |
| `NFR-567-02` | `UX-565-01` Multi-initiative readability rate должна целиться в `>= 85%` | walkthrough scorecards |
| `NFR-567-03` | `UX-565-02` Details access efficiency должна оставаться `>= 90%` | interaction review |
| `NFR-567-04` | `WF-565-01` Workflow policy clarity rate должна целиться в `>= 80%` | workflow walkthrough evidence |
| `NFR-567-05` | `DEL-565-01` Prototype-scope compliance должна оставаться `100%` | stage docs + owner review log |
| `NFR-567-06` | Platform-safe action compliance должна оставаться `100%` для core Sprint S18 сценариев | action-safety checklist |
| `NFR-567-07` | Fake-data isolation должна оставаться `100%`: core prototype не зависит от live provider sync и не нормализует backend rebuild как prerequisite | architecture/design handover review |

## Analytics и product evidence
- События:
  - `mission_control_canvas_opened`
  - `mission_control_initiative_focused`
  - `mission_control_drawer_opened`
  - `mission_control_toolbar_control_used`
  - `mission_control_relation_gap_seen`
  - `mission_control_workflow_preview_opened`
  - `mission_control_workflow_preview_updated`
  - `mission_control_out_of_scope_live_action_attempted`
- Метрики:
  - `NSM-565-01` Canvas decision-readiness rate
  - `PM-565-01` Orientation time p75
  - `UX-565-01` Multi-initiative readability rate
  - `UX-565-02` Details access efficiency
  - `WF-565-01` Workflow policy clarity rate
  - `DEL-565-01` Prototype-scope compliance
- Expected evidence:
  - owner walkthrough по 2-3 инициативам;
  - operator navigation review для toolbar/drawer;
  - workflow walkthrough на fake data;
  - action-safety checklist по negative live-path scenarios;
  - scope review note, подтверждающий deferred/later-wave границы.

## Риски и допущения

| Type | ID | Описание | Митигирующее действие | Статус |
|---|---|---|---|---|
| risk | `RSK-567-01` | Свободный canvas может стать visually noisy и ухудшить decision-readiness | Удерживать compact nodes, explicit relations и readability metrics как blocking signals | open |
| risk | `RSK-567-02` | Workflow editor может снова стать proxy для prompt editing или live mutation path | Держать deterministic policy preview и repo-seed prompt baseline как обязательный guardrail | open |
| risk | `RSK-567-03` | Fake-data prototype может оказаться слишком абстрактным для backend handover `#563` | Зафиксировать expected evidence и handover requirements уже на Day3 | open |
| risk | `RSK-567-04` | S16/S13 legacy contours (`lane/column`, `#524/#525`) могут снова начать диктовать UX | Удерживать rethink `#561` и deferred wave boundaries как locked baseline | open |
| assumption | `ASM-567-01` | Owner может принять UX baseline на fake data до появления нового backend foundation | accepted |
| assumption | `ASM-567-02` | Taxonomy `Issue` / `PR` / `Run` достаточна для core wave Sprint S18 | accepted |
| assumption | `ASM-567-03` | Drawer и toolbar достаточно, чтобы держать details/timeline/workflow context без lane/column shell | accepted |

## Открытые вопросы для `run:arch`
- Как разделить boundaries между isolated fake-data prototype в `web-console` и будущим backend rebuild `#563`, не потеряв locked baseline Sprint S18?
- Какие architecture alternatives нужны для relation semantics, workflow policy preview и future live sync boundary, чтобы `run:design` не переоткрыл product decisions Day1-Day3?
- Какие service ownership decisions надо зафиксировать, чтобы fake-data prototype не стал случайным source of truth для backend contracts?
- Как в architecture package сохранить explicit boundary, что Sprint S18 не нормализует live provider path, DB prompt editor и late-stage auto-continue?

## Handover в `run:arch`
- Follow-up issue: `#571`.
- На архитектурном этапе нельзя потерять:
  - frontend-first sequencing `#562 -> #565 -> #567`, после которого backend rebuild `#563` остаётся отдельным follow-up;
  - fullscreen свободный canvas без lane/column shell;
  - taxonomy `Issue` / `PR` / `Run`;
  - compact nodes, explicit relations, side panel/drawer и toolbar/controls;
  - workflow editor как fake-data policy UX, а не prompt editor и не live mutation path;
  - repo-seed prompts как source of truth;
  - platform-safe actions only;
  - `run:dev` как isolated `web-console` prototype без обязательного automatic late-stage flow;
  - deferred статус backend rebuild `#563`, release-safety cockpit и waves `#524/#525`.
- Архитектурный этап обязан определить:
  - service boundaries и ownership split;
  - fake-data isolation boundary и future backend handover boundary;
  - architecture alternatives для workflow policy preview, relation semantics и future sync path;
  - issue для `run:design` без trigger-лейбла с continuity-требованием `design -> plan -> dev`.

## Связанные документы
- `docs/delivery/sprints/s18/sprint_s18_mission_control_frontend_first_canvas_fake_data.md`
- `docs/delivery/epics/s18/epic_s18.md`
- `docs/delivery/epics/s18/epic-s18-day1-mission-control-frontend-first-canvas-intake.md`
- `docs/delivery/epics/s18/epic-s18-day2-mission-control-frontend-first-canvas-vision.md`
- `docs/delivery/epics/s18/epic-s18-day3-mission-control-frontend-first-canvas-prd.md`
- `docs/delivery/traceability/s18_mission_control_frontend_first_canvas_history.md`
- `docs/delivery/issue_map.md`
- `docs/delivery/delivery_plan.md`
- `docs/product/requirements_machine_driven.md`
- `docs/product/agents_operating_model.md`
- `docs/product/labels_and_trigger_policy.md`
- `docs/product/stage_process_model.md`
- `docs/architecture/prompt_templates_policy.md`
- `docs/research/src_idea-machine_driven_company_requirements.md`
- `services/staff/web-console/README.md`

