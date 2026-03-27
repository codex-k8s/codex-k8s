---
doc_id: ARC-S18-0001
type: architecture-design
title: "Sprint S18 Day 4 — Frontend-first Mission Control canvas architecture (Issue #571)"
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

# Sprint S18 Day 4 — Frontend-first Mission Control canvas architecture

## TL;DR
- Sprint S18 сохраняет Mission Control Wave 1 как `services/staff/web-console`-owned isolated prototype на fake data.
- `api-gateway`, `control-plane`, `worker` и `PostgreSQL` не получают новую Mission Control source-of-truth responsibility в этом спринте: они остаются текущими platform boundaries и explicit handover targets для backend rebuild `#563`.
- Repo-seed prompts остаются source of truth; workflow editor может показывать только deterministic `workflow-policy block`; DB prompt editor и live provider mutation path не вводятся.
- Handover после `run:arch` идёт в issue `#573` (`run:design`), а не в backend implementation stream; backend rebuild продолжает жить отдельным flow в issue `#563`.

## Контекст и входные артефакты
- Delivery-цепочка: `#562 (intake) -> #565 (vision) -> #567 (prd) -> #571 (arch)`.
- Reset baseline:
  - rethink `#561` перевёл Sprint S16 в superseded state;
  - source discussion `#480` остаётся историческим драйвером gap между Mission Control и owner expectations;
  - backend rebuild вынесен в отдельный follow-up `#563`.
- Source of truth для текущего stage:
  - `docs/delivery/epics/s18/epic-s18-day3-mission-control-frontend-first-canvas-prd.md`
  - `docs/delivery/epics/s18/prd-s18-day3-mission-control-frontend-first-canvas.md`
  - `docs/delivery/sprints/s18/sprint_s18_mission_control_frontend_first_canvas_fake_data.md`
  - `docs/delivery/traceability/s18_mission_control_frontend_first_canvas_history.md`
  - `docs/product/requirements_machine_driven.md`
  - `docs/product/agents_operating_model.md`
  - `docs/product/labels_and_trigger_policy.md`
  - `docs/product/stage_process_model.md`
  - `docs/architecture/c4_context.md`
  - `docs/architecture/c4_container.md`
  - `docs/architecture/agent_runtime_rbac.md`
  - `docs/architecture/prompt_templates_policy.md`
  - `services/staff/web-console/README.md`
  - `services/external/api-gateway/README.md`
  - `services/internal/control-plane/README.md`
  - `services/jobs/worker/README.md`
  - `services/jobs/agent-runner/README.md`

## Цели архитектурного этапа
- Превратить Day3 product contract в проверяемые service boundaries и ownership split для isolated fake-data prototype.
- Формально отделить текущий UI-owned prototype от будущего backend rebuild `#563`, чтобы Sprint S18 не зависел от недостроенного provider/data foundation.
- Зафиксировать, как сохраняются locked baselines Sprint S18 без возврата к lane/column shell, live mutation path, DB prompt editor или старой S16 taxonomy.
- Подготовить handover в `run:design` с явным списком contract/data/state вопросов, которые ещё предстоит детализировать.

## Non-goals
- Не выбираем точные HTTP/gRPC DTO, поля БД и миграции для backend rebuild.
- Не проектируем live GitHub/provider sync, persisted mirror, DB workflow definitions или new source-of-truth aggregate внутри Sprint S18.
- Не создаём новый deployable service и не меняем текущую production/runtime topology.
- Не вводим DB prompt editor, free-form prompt lifecycle или любую live provider mutation path.
- Не фиксируем release-safety cockpit, waves `#524/#525` и другие later-wave направления как blocking requirement для Sprint S18 `run:dev`.

## Неподвижные guardrails из PRD
- Mission Control остаётся fullscreen свободным canvas-first workspace, а не lane/column shell.
- Wave 1 taxonomy ограничена `Issue`, `PR`, `Run`.
- Compact nodes, explicit relations, side panel/drawer и toolbar/controls обязательны.
- Workflow editor остаётся fake-data UX и может показывать только deterministic `workflow-policy block`.
- Platform-safe actions only: prototype не мутирует GitHub/provider state.
- Repo-seed prompts остаются source of truth; DB prompt editor и free-form prompt storage не вводятся.
- `run:dev` в рамках Sprint S18 ограничен isolated `web-console` prototype и не запускает обязательную late-stage цепочку автоматически.

## Source-of-truth split

| Concern | Текущий owner в Sprint S18 | Deferred / downstream owner | Архитектурное правило |
|---|---|---|---|
| Canvas composition, node density, relation routing, drawer/toolbar state, workflow preview mode | `services/staff/web-console` | `run:design` детализация, затем `run:dev` | Это локальный UI/view-state и fake-data projection, а не platform-wide canonical truth |
| Fake-data scenario catalog (`Issue` / `PR` / `Run` bundles, relation samples, safe actions) | `services/staff/web-console` | заменяется backend read model только в отдельном flow `#563` | Все identifiers и relations в Sprint S18 имеют статус prototype fixtures и не объявляются persisted domain model |
| Prompt wording и workflow-policy copy | repo seeds + `docs/architecture/prompt_templates_policy.md` | позднее structured workflow policy layer в `#563` | `web-console` может только рендерить preview, но не становится owner prompt lifecycle |
| Live GitHub/provider state, review labels, true stale/freshness semantics | deferred to `#563` (`worker` mirror + `control-plane` policy`) | `#563` | Sprint S18 может только показывать safe placeholders или deep links; freshness не трактуется как возраст проекции |
| Platform-safe action semantics, approvals и stage policy | текущий platform baseline (`control-plane` + docs policy) | тот же owner после backend rebuild | Prototype не добавляет новую доменную policy и не обходит existing approval model |
| Persisted workflow definitions/instances и canonical `Issue/PR/Run` truth | deferred to `#563` (`control-plane` + `PostgreSQL`) | `#563` | Архитектурный пакет Sprint S18 явно запрещает превращать fake-data model в временную "почти-каноническую" БД-схему |

## Service boundaries for Sprint S18

| Service / layer | Роль в Sprint S18 | Что запрещено в Sprint S18 |
|---|---|---|
| `services/staff/web-console` | Единственный активный owner fake-data prototype: scenario model, canvas projection, local filters/focus, drawer, toolbar и workflow preview UX | Вычислять live GitHub truth, вводить DB prompt editor, требовать новые backend endpoints как prerequisite, мутировать provider state |
| `services/external/api-gateway` | Сохраняет existing thin-edge boundary для auth/session/static delivery и будущих typed transport seams | Встраивать Mission Control доменную логику, fake-data orchestration, workflow policy interpretation или postgres ownership |
| `services/internal/control-plane` | Остаётся owner platform-wide stage/policy baseline и future handover target для `#563` | Становиться hidden prerequisite Sprint S18, хранить fake-data scenario catalog или prematurely вводить canonical Mission Control aggregate для prototype |
| `services/jobs/worker` | Не участвует в Wave 1 runtime path; резервируется как future owner provider mirror/reconcile в `#563` | Синхронизировать fake-data prototype, вычислять canonical `Issue/PR/Run` truth или вводить stale/freshness semantics раньше backend rebuild |
| `services/jobs/agent-runner` / repo seeds | Остаются source of truth для prompt body и workflow-policy wording | Становиться prompt editor surface или runtime data source для canvas |
| `PostgreSQL` | Сохраняет текущее platform state; новые Mission Control persisted structures не вводятся на Sprint S18 Day4 | Использоваться как "временное" хранилище fake-data prototype ради ускорения `run:dev` |

## Chosen architecture slice

### Layer 1: Isolated prototype runtime
- Прототип живёт целиком в `web-console`.
- Состояние прототипа ограничено:
  - fake-data scenario catalog;
  - canvas selection/focus/filter state;
  - drawer open/details/timeline state;
  - workflow preview editing state;
  - safe action affordances.
- Этот слой не требует новых staff/private API, новых DB сущностей и новых background jobs.

### Layer 2: Reused platform policy baseline
- Repo-seed prompts и `prompt_templates_policy` дают канонический текстовый baseline для workflow-policy preview.
- Existing stage policy, approval semantics и runtime guardrails остаются внешним ограничением для prototype.
- `web-console` only mirrors these rules in UX language; он не становится source of truth для policy.

### Layer 3: Deferred backend seam
- `#563` остаётся единственным downstream flow, которому разрешено:
  - вводить persisted `Issue/PR/Run` read model;
  - проектировать provider mirror и reconcile;
  - формализовать structured workflow definitions/instances;
  - вернуть true stale/freshness semantics как lag provider mirror/reconcile path.
- Sprint S18 architecture package сохраняет этот seam документно, но не активирует runtime/code path.

## Alternatives summary
- Вариант A: начать backend-first rebuild сразу, переиспользуя части старой S16 модели.
  - Отклонён: смешивает rejected S16 assumptions с новым UX baseline и делает UI зависимым от ещё неутверждённой data model.
- Вариант B: собрать frontend prototype поверх текущих API и частично допустить live sync/mutation.
  - Отклонён: создаёт ложное ощущение canonical data path и ломает product guardrail `platform-safe actions only`.
- Вариант C (выбран): оставить Wave 1 полностью isolated в `web-console`, а backend rebuild держать отдельным flow через documented handover seam.
  - Выбран: лучше всего сохраняет baseline Sprint S18, thin-edge границы и проверяемую continuity `arch -> design -> plan -> dev`.

## Handover rules to backend rebuild `#563`
- Backend rebuild не может использовать fake-data identifiers или локальные relations как канонический persisted source.
- Любой будущий backend contract обязан стартовать от approved taxonomy `Issue` / `PR` / `Run`, а не от superseded S16 model.
- `stale/freshness` допускается только как доказанный lag provider mirror/reconcile path; возраст UI projection сам по себе не считается stale signal.
- Workflow policy может быть оцифрована только как structured policy layer, которая порождает deterministic `workflow-policy block`; prompt editor semantics не допускаются.
- Если backend rebuild требует вернуть lane/column shell, live mutation path или DB prompt editor, это считается scope violation и требует нового owner-решения.

## Architecture quality gates for `run:design`

| Gate | Что проверяем | Почему это обязательно |
|---|---|---|
| `QG-S18-A1` Isolation integrity | Prototype остаётся self-contained в `web-console` и не требует новых backend/runtime prerequisites | Иначе Sprint S18 потеряет frontend-first смысл |
| `QG-S18-A2` Boundary integrity | `api-gateway`, `control-plane`, `worker` остаются thin/deferred owners, а не hidden implementation path | Иначе `run:design` переоткроет архитектурные границы вместо детализации контрактов |
| `QG-S18-A3` Baseline fidelity | Fullscreen canvas, taxonomy `Issue/PR/Run`, compact nodes, explicit relations, drawer, toolbar и workflow preview сохранены без drift | Иначе design-stage вернётся к product rethink |
| `QG-S18-A4` Prompt-policy discipline | Repo seeds остаются source of truth, workflow preview не превращается в prompt editor | Иначе нарушается принятый prompt policy baseline |
| `QG-S18-A5` Deferred-scope discipline | `#563`, live sync, DB prompt editor, release-safety cockpit и waves `#524/#525` остаются non-blocking deferred scope | Иначе Sprint S18 снова смешает UX и backend foundation |

## Runtime impact и миграционный контур
- На этапе `run:arch` код, runtime, БД-схема и Kubernetes manifests не менялись.
- Для Sprint S18 `run:dev` допустим только isolated frontend prototype в `web-console` без новых transport/data migrations.
- Если после owner approval будет запущен backend rebuild `#563`, обязательный rollout order остаётся:
  - `schema -> control-plane -> worker -> api-gateway -> web-console`.
- Design stage `#573` должен отдельно зафиксировать:
  - fake-data state slices и interaction contracts, которые нужны `run:dev`;
  - explicit seams, которые позже заменятся backend truth без переоткрытия UX baseline;
  - migration notes, подтверждающие, что Sprint S18 `run:dev` остаётся без DB/runtime migration scope.

## Context7 и внешний baseline
- Context7 lookup на этапе `run:arch` не выполнялся: новые библиотеки и vendor integrations в scope отсутствуют.
- Локально проверен non-interactive GitHub CLI path для handover артефактов и PR automation:
  - `gh issue create --help`
  - `gh pr create --help`
  - `gh pr edit --help`

## Handover в `run:design`
- Следующий этап: `run:design`.
- Follow-up issue: `#573`.
- Trigger-лейбл на новую issue ставит Owner после review architecture package.
- На design-stage обязательно выпустить:
  - `design_doc.md` для canvas interaction model, drawer/toolbar behavior и workflow preview mode;
  - `api_contract.md` для explicit frontend seams, даже если Sprint S18 `run:dev` не требует новых live endpoints;
  - `data_model.md` для fake-data scenario model и documented replacement seam к backend rebuild `#563`;
  - `migrations_policy.md` с явным утверждением, что Sprint S18 prototype не создаёт обязательных runtime migrations, а backend migration scope остаётся за `#563`.
