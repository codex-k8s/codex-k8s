---
doc_id: EPC-CK8S-S18-D1-MISSION-CONTROL-FRONTEND
type: epic
title: "Epic S18 Day 1: Intake для frontend-first Mission Control canvas UX на fake data (Issue #562)"
status: in-review
owner_role: PM
created_at: 2026-03-26
updated_at: 2026-03-26
related_issues: [470, 480, 522, 523, 524, 525, 561, 562, 563, 565]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-26-issue-562-intake"
---

# Epic S18 Day 1: Intake для frontend-first Mission Control canvas UX на fake data (Issue #562)

## TL;DR
- Owner отверг продуктовый результат Sprint S16 и потребовал controlled reset: сначала показать и утвердить новый Mission Control UX на fake data, а уже потом возвращаться к backend rebuild.
- Intake фиксирует Sprint S18 как отдельную frontend-first инициативу: нужен isolated `web-console` prototype, который исследует canvas, density, interaction model и workflow editor UX без давления старого backend/data-model baseline.
- Day1-решение: рекомендованный путь = frontend-first fake-data sprint, затем отдельный backend sprint `#563` после owner approval UX.
- Зафиксированы обязательные baselines: fullscreen свободный canvas, Wave 1 taxonomy `Issue` / `PR` / `Run`, compact nodes для 2-3 инициатив, явные relations, side panel/drawer, toolbar/controls, workflow editor UX на fake data и platform-safe actions only.
- Prompt policy не меняется: repo-seed prompts остаются каноничными, workflow behavior допускается только через deterministic generated `workflow-policy block`, без DB prompt editor.
- Подготовлена continuity issue `#565` для stage `run:vision`.

## Контекст
- Sprint S9 уже дал первый Mission Control baseline как dashboard/control-plane, а issue `#480` затем зафиксировала gap между platform evidence и ожидаемым owner workspace.
- Sprint S16 попытался решить этот gap через новый graph workspace, но product baseline оказался отвергнут: lane/column shell, taxonomy `discussion/work_item/run/pull_request`, старые freshness semantics и backend-first sequencing больше не считаются актуальными.
- Issue `#561` перевела Sprint S16 в historical superseded state и зафиксировала новый reset path:
  - frontend-first sprint `#562` на fake data для утверждения UX;
  - отдельный backend rebuild `#563` только после owner approval нового UX;
  - `#524` и `#525` не стартуют до approval frontend baseline.
- Текущие code points всё ещё отражают старую траекторию:
  - `services/staff/web-console/src/pages/operations/MissionControlPage.vue`;
  - `services/staff/web-console/src/features/mission-control/MissionControlRootGroupLane.vue`;
  - `services/staff/web-console/src/features/mission-control/lib.ts`.
- Значит проблема теперь не “ещё один UI-refactor”, а продуктовый reset: сначала подтвердить правильную interaction model, потом под неё перестраивать backend и transport.

## Рекомендованный launch profile
- Базовый launch profile: `new-service`.
- Причины:
  - инициатива меняет product contour Mission Control и влияет на sequencing соседних delivery потоков;
  - нужен полный doc-stage contour, чтобы UX не был снова захвачен старой data/model semantics;
  - архитектурный handover в backend rebuild `#563` должен опираться на owner-approved UX, а не на промежуточные догадки.
- Целевая stage-цепочка:
  `run:intake -> run:vision -> run:prd -> run:arch -> run:design -> run:plan -> run:dev`.

## Problem Statement
### As-Is
- Текущий Mission Control в коде и исторических docs всё ещё тянет за собой старый S16 shell:
  - lane/column layout вместо свободного canvas;
  - избыточную taxonomy `discussion/work_item/run/pull_request`;
  - старую трактовку freshness/watermark как центрального UX-сигнала.
- Если продолжать backend-first rebuild до утверждения UX, новый provider/workflow backend снова будет проектироваться под отвергнутую interaction model.
- Workflow editor пока не подтверждён как продуктовый контур: непонятно, как именно должны выглядеть stage ordering, auto-review policy, follow-up propagation, relation/type rules и watermark rules в новом canvas-first UX.
- Live GitHub/provider integration слишком рано давит на UX-исследование: вместо intentional prototype команда рискует снова “подогнать” экран под существующие API и старую data model.

### To-Be
- Mission Control сначала проходит отдельный frontend-first цикл как product prototype на fake data.
- Prototype должен позволить быстро валидировать:
  - fullscreen canvas без lane/column shell;
  - компактную и читаемую node language `Issue` / `PR` / `Run`;
  - одновременную работу с 2-3 инициативами;
  - explicit relations вместо визуальной имитации колонок;
  - side panel/drawer и toolbar/controls как первичные surfaces;
  - workflow editor UX на fake data без live mutation path.
- Только после owner approval этого UX запускается отдельный backend sprint `#563`, который перестраивает data/transport/runtime контуры под уже утверждённую interaction model.

## Brief
- **Проблема:** Mission Control продолжает тащить за собой отвергнутый S16 baseline, а backend pressure мешает сначала утвердить правильный UX.
- **Для кого:** для Owner и delivery/operator пользователей, которым нужен понятный fullscreen workspace для нескольких инициатив; для команды, которой нужен подтверждённый UX baseline до backend rebuild.
- **Предлагаемое решение:** выделить Sprint S18 как отдельный frontend-first flow с fake-data prototype и жесткими guardrails по scope.
- **Почему сейчас:** после doc-reset `#561` нельзя возвращаться к incremental polishing старой модели; новый UX должен стать новым source of truth до старта `#563`.
- **Что считаем успехом:** intake-пакет фиксирует baseline, sequencing и ограничения, а следующая stage issue `#565` переводит тему в vision без drift.
- **Что не делаем на этой стадии:** не проектируем backend rebuild, не открываем live GitHub mutation path, не вводим DB prompt editor и не обещаем автоматический late-stage flow после prototype.

## Candidate sequencing options

| Вариант | Краткое описание | Плюсы | Риски | Intake-решение |
|---|---|---|---|---|
| A. Incremental polish старого S16 shell | Сохранить lane/column и старую taxonomy, локально улучшая вид и плотность | Быстрый старт, минимум сопротивления текущему коду | Закрепляет уже отвергнутый baseline и переносит проблему в backend | Явно не принимается |
| B. Frontend-first fake-data sprint, потом backend rebuild | Сначала утвердить canvas-first UX на isolated prototype, потом запускать `#563` | Правильный порядок для product reset, снижает риск backend-first drift | Требует жёстко удерживать scope и не подменять vision/PRD ранней реализацией | Рекомендуемый baseline Sprint S18 |
| C. Backend-first rebuild, затем UI polish | Сначала перестроить inventory/workflow backend, а UX уточнить позже | Ранний старт foundation stream | Снова подчиняет UX старой или случайной data-model и повторяет проблему S16 | Явно не принимается |

## MVP Scope
### In scope
- Полный doc-stage contour Sprint S18 до `run:dev`.
- Frontend-first product baseline для:
  - fullscreen свободного canvas;
  - node taxonomy `Issue`, `PR`, `Run`;
  - compact nodes и явных node-to-node relations;
  - side panel/drawer и toolbar/controls;
  - workflow editor UX на fake data.
- Guardrail, что `run:dev` реализует только isolated `web-console` prototype на fake data.
- Platform-safe actions only: никаких live GitHub/provider mutations в этом спринте.
- Repo-seed prompts как source of truth; workflow logic только как deterministic generated `workflow-policy block`.
- Handover в `run:vision` через continuity issue `#565`.

### Out of scope для core wave
- Backend rebuild, provider mirror redesign и schema/data foundation waves: это issue `#563`.
- DB prompt editor или любой free-form prompt storage lifecycle.
- Возврат к старой taxonomy `discussion/work_item/run/pull_request`.
- Final release-safety cockpit UI, readiness gate Sprint S13 и решения по waves `#524` / `#525`.
- Live GitHub/provider sync как обязательная основа prototype.
- Автоматическая обязательная late-stage цепочка после `run:dev` внутри этого спринта.

## Constraints
- Sprint S18 обязан сохранять решения rethink `#561`:
  - старый S16 baseline остаётся только historical evidence;
  - frontend-first sequencing обязателен;
  - backend rebuild запускается только после owner approval UX.
- Prototype не должен зависеть от текущего Mission Control API как source of truth.
- Workflow editor нельзя превращать в prompt editor.
- Prompt behavior не уходит в БД: repo seeds остаются каноничными.
- До `run:dev` stage остаётся markdown-only и не фиксирует premature implementation detail вне product/delivery пакета.

## Product principles
- UX baseline сначала, backend rebuild потом.
- Canvas first, а не “спасти lane/column shell”.
- Минимальная node language важнее наследования старой taxonomy.
- Relation semantics должны быть явными и визуально читаемыми без колонок.
- Workflow interaction сначала доказывается на fake data и только потом получает backend foundation.

## Candidate product waves

| Wave | Фокус | Exit signal |
|---|---|---|
| Wave 1 | Frontend-first canvas baseline, compact nodes, explicit relations, side panel/drawer, toolbar/controls, workflow editor UX на fake data | Owner принимает новый UX contour и разрешает перейти к backend rebuild |
| Wave 2 | Backend sprint `#563`: inventory/workflow data model, transport, sync/reconcile и cleanup obsolete S16 assumptions | Backend строится уже под утверждённый UX, а не наоборот |
| Wave 3 | Later-wave production hardening: release-safety cockpit, richer diagnostics, cross-repo/global views и прочие расширения | Дополнительные потоки развиваются без reopening core UX baseline |

## Acceptance Criteria (Intake stage)
- [x] Проблема зафиксирована как отдельный product reset, а не как локальный UI-refactor существующего Mission Control.
- [x] Сравнены минимум три sequencing options, и frontend-first fake-data sprint с отдельным backend rebuild выбран как рекомендуемый baseline.
- [x] Явно зафиксированы обязательные Day1 baselines: fullscreen свободный canvas, taxonomy `Issue` / `PR` / `Run`, compact nodes, explicit relations, side panel/drawer, toolbar/controls и workflow editor UX на fake data.
- [x] Явно зафиксировано, что `run:dev` в этом спринте ограничен isolated fake-data prototype и не открывает автоматический `qa/release/postdeploy/ops` path.
- [x] Prompt policy зафиксирована без drift: repo-seed prompts остаются source of truth, DB prompt editor не вводится, workflow logic допускается только как structured policy layer.
- [x] Подготовлена continuity issue `#565` для stage `run:vision` без trigger-лейбла.

## Декомпозиция по этапам

| Stage | Фокус | Выходной артефакт |
|---|---|---|
| Vision | Mission, north star, persona outcomes, KPI/guardrails и wave boundaries для нового canvas-first UX | charter + success metrics + risk frame |
| PRD | User stories, FR/AC/NFR, scenario matrix и expected evidence для fake-data prototype | PRD + user stories + NFR |
| Arch | Prototype isolation, UX/backend boundary и future handover в issue `#563` | architecture package + alternatives |
| Design | UI/interaction/state contracts, workflow editor behavior и feedback loop notes | design package |
| Plan | Delivery waves, owner feedback loops, DoR/DoD и execution package для isolated prototype | execution package |
| Dev | Реализация isolated `web-console` prototype на fake data | PR с prototype code + docs update |

## Risks and Product Assumptions
- Риск: команда попробует “для удобства” частично сохранить старую lane/column модель и тем самым вернёт rejected baseline.
- Риск: backend concerns снова начнут диктовать UX раньше owner approval и размоют controlled reset.
- Риск: workflow editor попытаются превратить в live mutation path или prompt editor.
- Риск: соседние потоки `#524` / `#525` начнут фиксировать UI до завершения Sprint S18.
- Допущение: isolated fake-data prototype даст Owner достаточно сигнала, чтобы принять или скорректировать UX до старта `#563`.
- Допущение: repo-seed prompt policy и structured workflow-policy layer достаточно, чтобы валидировать workflow UX без DB prompt editor.

## Stage Handover Instructions
- Следующий этап: `run:vision`.
- Созданная issue следующего этапа: `#565`.
- На stage `run:vision` обязательно сохранить и не переоткрывать без нового owner-решения следующие intake-decisions:
  - Sprint S18 остаётся frontend-first flow на fake data, а backend rebuild идёт только после owner approval;
  - fullscreen свободный canvas без lane/column shell обязателен;
  - Wave 1 nodes = `Issue`, `PR`, `Run`;
  - compact nodes, explicit relations, side panel/drawer и toolbar/controls обязательны;
  - workflow editor входит в core scope Sprint S18, но остаётся fake-data UX без live mutation path;
  - `run:dev` ограничен isolated `web-console` prototype;
  - repo-seed prompts остаются каноничными, DB prompt editor не появляется;
  - автоматический late-stage flow после `run:dev` внутри этой инициативы не включается.
- После завершения vision stage должна быть создана новая issue для `run:prd` без trigger-лейбла.
