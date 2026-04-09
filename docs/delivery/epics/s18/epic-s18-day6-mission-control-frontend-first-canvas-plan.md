---
doc_id: EPC-CK8S-S18-D6-MISSION-CONTROL-CANVAS
type: epic
title: "Epic S18 Day 6: Plan для frontend-first Mission Control canvas prototype и handover в run:dev (Issues #579/#581)"
status: in-review
owner_role: EM
created_at: 2026-04-01
updated_at: 2026-04-01
related_issues: [480, 561, 562, 563, 565, 567, 571, 573, 579, 581]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-04-01-issue-579-plan-epic"
---

# Epic S18 Day 6: Plan для frontend-first Mission Control canvas prototype и handover в `run:dev`

## TL;DR
- Подготовлен execution package Sprint S18 для перехода в `run:dev` по isolated Mission Control canvas prototype в `web-console`.
- Реализация остаётся в одном owner-managed implementation issue `#581`, но внутри неё зафиксированы четыре последовательные execution waves.
- Зафиксированы quality gates, DoR/DoD, blockers, risks и owner decisions, чтобы `run:dev` не смешал frontend-only prototype с backend rebuild `#563`.
- Сохранена continuity `plan -> dev` без разрывов: follow-up issue `#581` создана без trigger-лейбла.

## Контекст
- Stage continuity: `#562 -> #565 -> #567 -> #571 -> #573 -> #579`.
- Входной baseline: design package Day5
  - `docs/architecture/initiatives/s18_mission_control_frontend_first_canvas/design_doc.md`
  - `docs/architecture/initiatives/s18_mission_control_frontend_first_canvas/api_contract.md`
  - `docs/architecture/initiatives/s18_mission_control_frontend_first_canvas/data_model.md`
  - `docs/architecture/initiatives/s18_mission_control_frontend_first_canvas/migrations_policy.md`
- Scope текущего stage: только markdown-изменения и handover backlog.
- Правило continuity: trigger-лейбл `run:dev` на issue `#581` ставит только Owner после review plan package.

## Execution package (`#581`)

| Wave | Scope | Priority | Результат |
|---|---|---|---|
| `Wave 1` | Route shell и feature-local prototype source/store | P0 | `MissionControlPage.vue` переключён на explicit prototype path без current API/realtime branch |
| `Wave 2` | Fullscreen canvas composition, compact nodes, explicit relations, toolbar и drawer surfaces | P0 | Owner-ready canvas baseline для `1..3` инициатив без lane/column shell |
| `Wave 3` | Workflow policy preview, prompt-source evidence и platform-safe action affordances | P0 | Deterministic generated `workflow-policy block` с repo-seed refs, без mutation/editor semantics |
| `Wave 4` | Acceptance/demo evidence, traceability sync и owner-ready PR package | P0 | PR с implementation evidence, checks и continuity handover следующего stage |

## Sequencing constraints
- `Wave 1` обязательна первой: пока route не переключён на prototype source/store, любые UI правки рискуют остаться связанными с superseded API/realtime path.
- `Wave 2` запускается только поверх готового prototype source/store и не должна возвращать graph/list toggle, freshness chips или lane/column shell.
- `Wave 3` не начинается как отдельный prompt-editing stream: workflow preview остаётся продолжением UI prototype и использует только structured toggles + deterministic generated block.
- `Wave 4` закрывает acceptance evidence и traceability только после стабилизации первых трёх волн; без owner-ready PR package Sprint S18 не считается готовым к следующему stage decision.
- Во всех waves запрещено расширять scope на backend rebuild `#563`, live provider sync, DB prompt editor, release-safety cockpit и waves `#524/#525`.

## Quality gates (`run:plan`)

| Gate | Что проверяем | Статус |
|---|---|---|
| `QG-S18-D6-01` | Создан owner-managed implementation backlog `#581` без trigger-лейбла | passed |
| `QG-S18-D6-02` | Внутри `#581` зафиксированы последовательные execution waves `Wave 1..Wave 4` | passed |
| `QG-S18-D6-03` | Frontend-only boundary сохранён: без backend/API/DB/runtime migrations и без скрытого расширения в `#563` | passed |
| `QG-S18-D6-04` | Locked baseline Sprint S18 сохранён: fullscreen canvas, taxonomy `Issue/PR/Run`, compact nodes, explicit relations, drawer, toolbar | passed |
| `QG-S18-D6-05` | Workflow preview удержан как deterministic generated block с repo-seed refs и platform-safe actions only | passed |
| `QG-S18-D6-06` | Требования к acceptance/demo evidence и traceability для `run:dev` зафиксированы явно | passed |
| `QG-S18-D6-07` | Continuity `plan -> dev` сохранена: issue `#581` создана и привязана к S18 traceability | passed |
| `QG-S18-D6-08` | Scope этапа ограничен markdown-only изменениями | passed |

## Definition of Ready / Done для перехода в `run:dev`

### Definition of Ready (`run:dev` launch)
- [x] Design package Day5 (`#573`) подтверждён как source of truth.
- [x] Для implementation stage создана отдельная follow-up issue `#581`.
- [x] Зафиксированы execution waves `Wave 1..Wave 4`, sequencing constraints и quality gates.
- [x] Deferred boundaries `#563`, `#524`, `#525`, live sync и DB prompt editor сохранены явно.
- [x] Trigger-лейбл `run:dev` на issue `#581` не выставлен автоматически.

### Definition of Done (`run:plan` stage)
- [x] Выпущен plan-эпик Day6 с execution package, quality gates и DoR/DoD.
- [x] Обновлены `delivery_plan`, sprint/epic каталоги, `issue_map`, history package и индексы для Sprint S18.
- [x] Подготовлен owner-facing handover в `run:dev` через issue `#581`.
- [x] Зафиксированы blockers, risks и owner decisions для следующего stage.

## Self-check (common checklist)
- Проверен scope изменений: только markdown-документы (`*.md`), без кода, YAML/JSON, Dockerfile и shell-скриптов.
- Проверена консистентность stage-policy: `run:dev` остаётся единственным кодовым этапом, trigger-лейбл на `#581` не выставлялся.
- Проверена traceability-синхронизация для Sprint S18: `delivery_plan`, sprint/epic docs, `issue_map`, history package, sprint/epic indexes.
- Новые внешние зависимости не выбирались; Context7 и catalog внешних зависимостей не требовались.
- Секреты, токены и environment values в артефактах не публиковались.

## Blockers, risks, owner decisions

| Тип | ID | Описание | Статус |
|---|---|---|---|
| blocker | `BLK-S18-D6-01` | До старта `run:dev` нужен owner review/approval plan package по Issue `#579` | open |
| blocker | `BLK-S18-D6-02` | Trigger-лейбл `run:dev` на issue `#581` должен выставить Owner после review plan package | open |
| risk | `RSK-S18-D6-01` | Если `Wave 1` оставит route на current API/realtime path, весь prototype унаследует superseded S16/S9 semantics | monitoring |
| risk | `RSK-S18-D6-02` | Если `Wave 2` попытается вернуть lane/column shell или graph/list toggle, locked baseline Sprint S18 будет нарушен | monitoring |
| risk | `RSK-S18-D6-03` | Если `Wave 3` разрастётся до prompt editor или live mutation path, Sprint S18 потеряет frontend-only boundary и потребует нового owner decision | monitoring |
| risk | `RSK-S18-D6-04` | Если `Wave 4` не подготовит owner-ready demo evidence, acceptance результата Sprint S18 затянется и заблокирует следующий stage decision | monitoring |
| owner-decision | `OD-S18-D6-01` | Реализация Sprint S18 идёт через одну implementation issue `#581` с внутренними waves, а не через набор параллельных sub-issues | accepted |
| owner-decision | `OD-S18-D6-02` | Backend rebuild `#563` остаётся отдельным follow-up и не входит в execution package Day6 | accepted |
| owner-decision | `OD-S18-D6-03` | После `run:dev` обязательная late-stage цепочка внутри Sprint S18 не запускается автоматически; следующий шаг остаётся owner-managed | accepted |

## Tooling validation
- Локально проверены команды для GitHub flow:
  - `gh issue view 579 --json number,title,body,url`
  - `gh issue create --help`
  - `gh pr create --help`
  - `gh pr edit --help`
- Через `gh issue create` оформлена handover issue `#581`.
- Kubectl/logs/БД-запросы не выполнялись, потому что stage ограничен documentation-only scope и не требовал runtime-debug.

## Acceptance Criteria (Issue #579)
- [x] Подготовлен execution-ready plan package для Sprint S18 frontend-first Mission Control canvas prototype.
- [x] Зафиксированы execution waves, sequencing, blockers, quality gates и DoR/DoD без пересмотра Day1-Day5 baseline.
- [x] Сохранены guardrails: fullscreen canvas, taxonomy `Issue` / `PR` / `Run`, compact nodes, explicit relations, side panel/drawer, toolbar/controls, workflow preview на fake data, platform-safe actions only и repo-seed prompts как source of truth.
- [x] Backend rebuild `#563`, live provider sync, DB prompt editor, release-safety cockpit и waves `#524` / `#525` удержаны в deferred/later-wave scope.
- [x] Создана follow-up issue `#581` для `run:dev`, и в её body явно сохранено continuity-требование продолжить цепочку `dev` без разрывов.

## Handover в `run:dev`
- Следующий operational stage: `run:dev`.
- Follow-up issue: `#581`.
- Trigger-лейбл `run:dev` на `#581` ставит Owner после review plan package.
- Для `run:dev` обязательны:
  - реализация строго по waves `Wave 1 -> Wave 2 -> Wave 3 -> Wave 4`;
  - PR с checks, demo evidence и traceability sync;
  - сохранение frontend-only boundary без reopening scope `#563`;
  - continuity следующего stage без разрыва цепочки `dev`.
