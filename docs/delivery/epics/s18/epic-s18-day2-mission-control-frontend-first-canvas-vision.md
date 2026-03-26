---
doc_id: EPC-CK8S-S18-D2-MISSION-CONTROL-FRONTEND
type: epic
title: "Epic S18 Day 2: Vision для frontend-first Mission Control canvas UX на fake data (Issues #565/#567)"
status: in-review
owner_role: PM
created_at: 2026-03-26
updated_at: 2026-03-26
related_issues: [470, 480, 522, 523, 524, 525, 561, 562, 563, 565, 567]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-26-issue-565-vision"
---

# Epic S18 Day 2: Vision для frontend-first Mission Control canvas UX на fake data (Issues #565/#567)

## TL;DR
- Для Issue `#565` сформирован vision-package: mission, north star, persona outcomes, KPI/guardrails, wave boundaries и product principles для Sprint S18.
- Mission Control зафиксирован как owner-approved canvas-first workspace на fake data: сначала нужно утвердить UX свободного canvas для 2-3 инициатив, и только потом запускать backend rebuild в issue `#563`.
- Day1 baseline из Issue `#562` сохранён без reopening: fullscreen свободный canvas, taxonomy `Issue` / `PR` / `Run`, compact nodes, explicit relations, side panel/drawer, toolbar/controls и workflow editor UX на fake data остаются обязательными.
- Workflow editor остаётся policy-shaping UX, а не prompt editor и не live provider mutation path: repo-seed prompts сохраняются source of truth, workflow logic допускается только как deterministic generated `workflow-policy block`.
- Создана follow-up issue `#567` для stage `run:prd` без trigger-лейбла; PRD должен формализовать user stories, FR/AC/NFR, scenario matrix и expected evidence без разрыва цепочки `prd -> arch -> design -> plan -> dev`.

## Priority
- `P0`.

## Vision charter

### Mission statement
Сделать Mission Control owner-approved canvas-first workspace на fake data, чтобы owner / product lead и execution lead могли за одну сессию увидеть состояние 2-3 инициатив, понять связи `Issue` / `PR` / `Run`, безопасно исследовать workflow и принять решение о следующем шаге до запуска backend rebuild.

### Цели и ожидаемые результаты
1. Зафиксировать свободный fullscreen canvas как основной Mission Control UX вместо lane/column shell и подтвердить, что он остаётся понятным для 2-3 инициатив одновременно.
2. Сделать compact nodes, explicit relations, side panel/drawer и toolbar/controls основными surfaces для чтения состояния, фокусировки и навигации по инициативам.
3. Подтвердить workflow editor UX на fake data как часть core product contour без превращения его в prompt editor, live mutation path или backend-first dependency.
4. Удержать sequencing Sprint S18: `run:dev` реализует только isolated `web-console` prototype, а backend rebuild `#563` стартует только после owner approval нового UX baseline.

### Пользователи и стейкхолдеры
- Основные пользователи:
  - owner / product lead, которому нужно быстро понять состояние нескольких инициатив и принять безопасное product/delivery решение без GitHub-first detour;
  - execution lead / operator, которому нужен единый canvas для чтения связей `Issue` / `PR` / `Run`, просмотра timeline/details и переключения фокуса между инициативами;
  - future delivery team, которой нужен owner-approved UX baseline до старта архитектурного handover и backend rebuild.
- Стейкхолдеры:
  - `services/staff/web-console` как primary UX-контур isolated prototype;
  - `services/internal/control-plane`, `services/jobs/worker` и `services/external/api-gateway` как будущие потребители product contract после owner approval UX;
  - Sprint S13 waves `#524` / `#525`, которые остаются зависимыми от утверждения нового frontend baseline;
  - Owner как финальный апрувер направления.

### Persona outcomes
- Owner / product lead:
  - видит 2-3 инициативы на одном canvas без lane/column scaffolding;
  - понимает, где `Issue`, `PR` и `Run` находятся в текущем delivery контуре;
  - открывает нужные details/timeline и выбирает следующий безопасный шаг без догадок о скрытой иерархии.
- Execution lead / operator:
  - быстро переключает фокус между инициативами, не теряя связи между node и timeline;
  - использует drawer и toolbar как primary navigation, а не как secondary fallback;
  - видит workflow semantics и policy preview без обращения к prompt internals.
- Future implementation team:
  - получает owner-approved UX baseline, который можно переводить в PRD/architecture/design без возврата к rejected S16 assumptions;
  - не начинает backend rebuild, пока UX contour не зафиксирован на fake data.

### Продуктовые принципы и ограничения
- Сначала утверждается UX baseline, потом запускается backend rebuild.
- Canvas first: lane/column shell и обязательная `root-group/column/stack` иерархия не возвращаются без нового owner-решения.
- Wave 1 node language остаётся минимальной: `Issue`, `PR`, `Run`.
- Workflow editor остаётся частью продукта, но его задача в Sprint S18 показать безопасный policy UX на fake data, а не редактировать промпты и не мутировать provider state.
- Platform-safe actions only: live GitHub/provider mutation path в этом спринте не допускается.
- Repo-seed prompts остаются source of truth; workflow behavior допускается только как deterministic generated `workflow-policy block`.
- В рамках `run:vision` разрешены только markdown-изменения.

## Scope boundaries

### MVP scope
- Isolated `web-console` prototype на fake data.
- Fullscreen свободный canvas для 2-3 инициатив одновременно.
- Compact nodes и явные связи для `Issue`, `PR`, `Run`.
- Side panel/drawer для details, timeline и action context.
- Toolbar/controls для переключения инициатив, фильтров, density/focus и workflow interaction mode.
- Workflow editor UX на fake data с deterministic policy preview и platform-safe actions only.
- Явный guardrail, что `run:dev` в Sprint S18 ограничен isolated prototype и не открывает обязательную late-stage цепочку.

### Post-MVP / deferred scope
- Backend rebuild и transport/data foundation wave в issue `#563`.
- Live GitHub/provider sync и любой live mutation path.
- DB prompt editor или иной free-form prompt lifecycle.
- Release-safety cockpit, final readiness UI и связанные execution waves `#524` / `#525`.
- Автоматический переход `run:dev -> run:qa -> run:release -> run:postdeploy -> run:ops` внутри этой инициативы.

### Sequencing and locked baseline
- Active vision stage в Issue `#565` допустим только как продолжение locked intake baseline Issue `#562` и rethink sequencing Issue `#561`.
- Issue `#563` остаётся отдельной backend-задачей и не может стать частью core MVP scope Sprint S18 до owner approval UX.
- Если следующий stage начнёт требовать lane/column shell, расширенную taxonomy beyond `Issue` / `PR` / `Run` или prompt-editing semantics, stage должен быть остановлен до нового owner-решения.

## Candidate product waves

| Wave | Фокус | Exit signal |
|---|---|---|
| Wave 1 | Sprint S18 frontend-first contour: vision -> prd -> arch -> design -> plan -> dev и isolated fake-data prototype в `web-console` | Owner принимает новый canvas-first UX baseline и разрешает backend handover |
| Wave 2 | Backend rebuild в issue `#563`: data/transport/sync foundation, которая переводит approved UX в runtime contracts | Backend строится уже под утверждённый UX, а не навязывает ему старую модель |
| Wave 3 | Later-wave live integrations и operational hardening: release-safety cockpit, richer diagnostics, unblock соседних UI/readiness waves | Дополнительные потоки развиваются без reopening core UX baseline |

## Success metrics

### North Star
| ID | Метрика | Определение | Источник | Целевое значение |
|---|---|---|---|---|
| `NSM-565-01` | Canvas decision-readiness rate | Доля pilot walkthrough-сценариев, в которых owner, используя только Mission Control prototype, корректно объясняет состояние и безопасный следующий шаг минимум по двум инициативам в одной сессии | scripted UX reviews + owner acceptance notes | `>= 80%` до approval backend rebuild |

### Supporting metrics
| ID | Метрика | Определение/формула | Источник | Цель |
|---|---|---|---|---|
| `PM-565-01` | Orientation time p75 | p75 времени от открытия prototype до первого корректного объяснения состояния выбранной инициативы | pilot session timing + review checklist | `<= 90 секунд` |
| `UX-565-01` | Multi-initiative readability rate | Доля pilot-сценариев, где пользователь удерживает контекст 2-3 инициатив без возврата к lane/column shell и без потери связей `Issue` / `PR` / `Run` | walkthrough scorecards | `>= 85%` |
| `UX-565-02` | Details access efficiency | Доля core details/timeline действий, которые пользователь достигает не более чем за 2 взаимодействия от canvas | prototype interaction review | `>= 90%` |
| `WF-565-01` | Workflow policy clarity rate | Доля scripted workflow-сценариев, где пользователь правильно предсказывает policy result на fake data без prompt editor и без live mutation path | workflow walkthrough evidence | `>= 80%` |
| `DEL-565-01` | Prototype-scope compliance | Доля принятых scope-решений до owner approval UX, которые остаются внутри isolated fake-data prototype и не требуют раннего backend rebuild | stage docs + owner review log | `100%` |

### Guardrails (ранние сигналы)
- `GR-565-01`: если `UX-565-01 < 75%`, следующий stage должен упрощать canvas density и relation semantics, а не расширять taxonomy.
- `GR-565-02`: если для понимания prototype снова требуется lane/column shell или nested root-group модель, stage переводится в `need:input` до нового owner-решения.
- `GR-565-03`: если workflow editor требует DB prompt editor, raw prompt diff или live provider mutation, stage блокируется до пересмотра scope.
- `GR-565-04`: если `DEL-565-01 < 100%`, `run:dev` не может быть трактован как isolated fake-data prototype и не должен запускаться до возврата в product guardrails.
- `GR-565-05`: если соседние waves `#524` / `#525` или issue `#470` начинают фиксировать финальный cockpit UI до approval Sprint S18, требуется explicit owner decision по sequencing.

## Risks and Product Assumptions
| Тип | ID | Описание | Митигирующее действие | Статус |
|---|---|---|---|---|
| risk | `RSK-565-01` | Свободный canvas может превратиться в визуальный шум и потерять читаемость на 2-3 инициативах | Держать compact nodes, explicit relations и measurable readability guardrails уже на vision/PRD | open |
| risk | `RSK-565-02` | Workflow editor может расползтись в prompt editor или pseudo-automation, не подтверждённую product scope | Явно фиксировать policy-only UX и repo-seed prompts как source of truth | open |
| risk | `RSK-565-03` | Fake-data prototype может дать слишком абстрактный сигнал и затруднить handover в backend rebuild `#563` | В `run:prd` зафиксировать expected evidence и boundary между UX baseline и backend foundation | open |
| risk | `RSK-565-04` | Rejected S16 shell и связанные readiness waves могут снова стать неявным source of truth | Сохранять rethink sequencing `#561` как locked baseline во всех следующих stage | open |
| assumption | `ASM-565-01` | Owner может валидировать новый Mission Control UX на fake data до наличия нового backend foundation | Проверить в PRD expected evidence и acceptance walkthroughs | accepted |
| assumption | `ASM-565-02` | Минимальная taxonomy `Issue` / `PR` / `Run` достаточна для первой волны UX и не требует возврата к старой модели `discussion/work_item/run/pull_request` | Подтвердить user stories и edge cases на `run:prd` | accepted |
| assumption | `ASM-565-03` | Side panel/drawer и toolbar/controls достаточно, чтобы держать details, timeline и workflow control без lane/column shell | Удержать как explicit design baseline на `run:prd` и `run:design` | accepted |

## Readiness criteria для `run:prd`
- [x] Mission, north star и persona outcomes сформулированы для owner/product lead, execution lead и future implementation handover.
- [x] KPI/success metrics и guardrails определены как измеримые сигналы для canvas comprehension, multi-initiative readability, details access, workflow clarity и scope safety.
- [x] Locked Day1 baseline сохранён явно: fullscreen свободный canvas, taxonomy `Issue` / `PR` / `Run`, compact nodes, explicit relations, side panel/drawer, toolbar/controls и workflow editor UX на fake data.
- [x] Core MVP и deferred scope разделены явно; backend rebuild `#563`, live provider sync, DB prompt editor и late-stage auto-continue не входят в blocking scope Sprint S18.
- [x] Создана отдельная issue следующего этапа `run:prd` (`#567`) без trigger-лейбла.

## Acceptance criteria (Issue #565)
- [x] Mission, north star и product principles для frontend-first Mission Control canvas UX сформулированы явно.
- [x] KPI/success metrics и guardrails зафиксированы для comprehension, readability, details access, workflow clarity и prototype-scope compliance.
- [x] Persona outcomes, MVP/Post-MVP границы, риски и assumptions описаны без reopening Day1 baseline и без перехода в implementation details.
- [x] Явно сохранены как обязательный baseline: fullscreen свободный canvas, taxonomy `Issue` / `PR` / `Run`, compact nodes, explicit relations, side panel/drawer, toolbar/controls, workflow editor UX на fake data, platform-safe actions only, repo-seed prompts как source of truth и `run:dev` как isolated prototype.
- [x] Подготовлен handover в `run:prd` и создана follow-up issue `#567` без trigger-лейбла.

## Handover в следующий этап
- Следующий stage: `run:prd`.
- Follow-up issue: `#567`.
- Trigger-лейбл `run:prd` на issue `#567` ставит Owner.
- Обязательное continuity-требование для `#567`:
  - в конце PRD stage агент обязан создать issue для `run:arch` без trigger-лейбла;
  - в body этой issue нужно явно повторить требование продолжить цепочку `arch -> design -> plan -> dev` без разрывов.
- На `run:prd` нельзя потерять следующие решения vision:
  - Sprint S18 остаётся frontend-first fake-data flow, а backend rebuild идёт только через issue `#563` после owner approval;
  - fullscreen свободный canvas без lane/column shell обязателен;
  - Wave 1 nodes = `Issue`, `PR`, `Run`;
  - compact nodes, explicit relations, side panel/drawer и toolbar/controls обязательны;
  - workflow editor входит в core scope Sprint S18, но остаётся fake-data UX с deterministic `workflow-policy block`, без prompt editor и live mutation path;
  - `run:dev` ограничен isolated `web-console` prototype;
  - автоматический late-stage flow после `run:dev` внутри этой инициативы не включается;
  - waves `#524` / `#525` остаются заблокированными до owner approval Sprint S18.

## Связанные документы
- `docs/delivery/sprints/s18/sprint_s18_mission_control_frontend_first_canvas_fake_data.md`
- `docs/delivery/epics/s18/epic_s18.md`
- `docs/delivery/epics/s18/epic-s18-day1-mission-control-frontend-first-canvas-intake.md`
- `docs/delivery/traceability/s18_mission_control_frontend_first_canvas_history.md`
- `docs/delivery/issue_map.md`
- `docs/delivery/delivery_plan.md`
- `docs/product/requirements_machine_driven.md`
- `docs/product/agents_operating_model.md`
- `docs/product/labels_and_trigger_policy.md`
- `docs/product/stage_process_model.md`
- `docs/architecture/prompt_templates_policy.md`
- `docs/research/src_idea-machine_driven_company_requirements.md`
- `services/internal/control-plane/README.md`
- `services/jobs/worker/README.md`
- `services/external/api-gateway/README.md`
- `services/staff/web-console/README.md`
