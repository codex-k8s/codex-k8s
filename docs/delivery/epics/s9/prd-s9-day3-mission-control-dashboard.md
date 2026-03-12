---
doc_id: PRD-CK8S-S9-D3-MISSION-CONTROL
type: prd
title: "Mission Control Dashboard — PRD Sprint S9 Day 3"
status: in-review
owner_role: PM
created_at: 2026-03-12
updated_at: 2026-03-12
related_issues: [333, 335, 337, 340]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-12-issue-337-prd"
---

# PRD: Mission Control Dashboard

## TL;DR
- Что строим: Mission Control Dashboard как основной control-plane UX staff console для active work items, discussion, PR и агентов.
- Для кого: owner/product lead, engineer/operator и discussion-first пользователь.
- Почему: текущий baseline платформы уже умеет отдельные runtime/debug потоки, но не даёт единый operational screen и безопасный discussion-first workflow.
- MVP: active-set dashboard shell, typed entity/relation model, side panel, filters/list fallback, discussion formalization, provider-safe commands и reconciliation guardrails.
- Критерии успеха: пользователь быстро получает situational awareness, формализует discussion в task без потери traceability, а console-initiated actions не создают дубли после webhook echo.

## Проблема и цель
- Problem statement:
  - staff console пока остаётся набором разрозненных operational screen'ов, а не control plane;
  - пользователь по-прежнему собирает картину состояния через GitHub UI, отдельные runtime-экраны и комментарии;
  - discussion-first путь `идея -> discussion -> formalized task -> stage launch` не оформлен как штатный продуктовый сценарий;
  - без явного command/reconciliation path консольные действия рискуют порождать повторные task/run/comment сущности.
- Цели:
  - сделать Mission Control Dashboard основной landing page staff console для active set;
  - дать пользователю единый экран, который за 5-10 секунд отвечает на вопрос «что важно сейчас и что делать дальше»;
  - закрепить provider-safe operational model: console управляет работой, но не заменяет provider UI для human review;
  - сохранить traceability, label policy и webhook-driven orchestration.
- Почему сейчас:
  - Sprint S9 уже подтвердил отдельную инициативу Mission Control Dashboard на intake/vision этапах;
  - без PRD следующий stage смешает product scope с частными архитектурными решениями;
  - сейчас есть лучшее окно зафиксировать проверяемые критерии до архитектуры и design.

## Зафиксированные продуктовые решения
- `D-337-01`: Mission Control Dashboard остаётся control plane для active work, а не общим redesign staff console.
- `D-337-02`: GitHub-first MVP сохраняется; GitLab parity остаётся future-compatible направлением вне первой волны.
- `D-337-03`: Human review, merge decision и provider-specific collaboration остаются во внешнем provider UI.
- `D-337-04`: Active-set default и list fallback обязательны уже в core MVP; полное архивное представление не может быть дефолтным режимом.
- `D-337-05`: Любое действие из dashboard обязано идти по typed command path с audit, sync state и reconciliation.
- `D-337-06`: Realtime является first-class UX, но при деградации должен существовать безопасный fallback без потери core usability.
- `D-337-07`: Voice intake не блокирует core MVP и остаётся conditional stream до отдельного product go/no-go.

## Scope boundaries
### In scope
- Active-set dashboard shell:
  - default landing page;
  - summary strip;
  - primary board/canvas view;
  - list view как fallback для large active set и degraded mode;
  - правая side panel для details/timeline/comments/actions.
- Typed сущности и связи:
  - Work Item;
  - Discussion;
  - Pull Request;
  - Agent;
  - Comment/Timeline projection;
  - Relation.
- Core product flows:
  - active-set situational awareness;
  - create discussion/task;
  - discussion -> formalize as task;
  - open context and continue stage;
  - observe provider/internal state and sync status;
  - provider-safe commands + webhook echo reconciliation.
- Product evidence и analytics для success metrics из vision.

### Out of scope
- Полный redesign остальных разделов staff console.
- Выбор конкретной graph/realtime/STT библиотеки и implementation lock-in до `run:arch`.
- Перенос human review или merge controls в staff console.
- GitLab parity в первой волне.
- Voice intake как обязательная часть core MVP.

## Пользователи / персоны

| Persona | Основная работа | Что считает успехом |
|---|---|---|
| Owner / product lead | Понять active set, определить приоритет и запустить следующий безопасный шаг | Control-plane экран быстро показывает, где блокировки, кто занят и что нужно делать дальше |
| Engineer / operator | Управлять work items, PR, агентами и ожиданиями без хаотичной навигации по разным экранам | Все активные сущности и разрешённые действия видны из одного места, включая provider state и sync status |
| Discussion-first пользователь | Начать с идеи/обсуждения, затем формализовать задачу и запустить stage | Discussion можно быстро превратить в formal task без потери контекста и без ручной orchestration через label UX |

## User stories и wave priorities

| Story ID | История | Wave | Приоритет |
|---|---|---|---|
| `S9-US-01` | Как owner/product lead, я хочу открыть dashboard и сразу увидеть active set, блокировки и ближайшие действия, чтобы быстро принять решение о следующем шаге | Wave 1 | `P0` |
| `S9-US-02` | Как engineer/operator, я хочу видеть связанные work items, PR, agents и sync state в одном интерфейсе, чтобы не переключаться между GitHub и разрозненными staff pages | Wave 1 | `P0` |
| `S9-US-03` | Как discussion-first пользователь, я хочу создать discussion и затем формализовать её в task, чтобы пройти путь от идеи к stage launch без потери истории | Wave 2 | `P0` |
| `S9-US-04` | Как operator, я хочу инициировать безопасные действия из console и видеть их статус до reconciliation, чтобы доверять dashboard как control plane | Wave 2 | `P0` |
| `S9-US-05` | Как пользователь voice-first сценария, я хочу при наличии policy-safe voice intake получить draft discussion/task, чтобы ускорить старт работы без блокировки core MVP | Wave 3 | `P1` |

### Wave sequencing
- Wave 1 (`pilot`, `P0`):
  - active-set dashboard shell;
  - typed entity/relation baseline;
  - side panel;
  - search/filters;
  - list fallback;
  - provider context visibility.
- Wave 2 (`MVP release`, `P0`):
  - discussion formalization;
  - comments/timeline projection;
  - command status and reconciliation;
  - dedupe/correlation;
  - degraded realtime fallback.
- Wave 3 (`conditional`, `P1`):
  - guided voice intake;
  - transcript -> draft structuring;
  - explicit confirmation before creating discussion/task.

## Functional Requirements

| ID | Требование |
|---|---|
| `FR-337-01` | Mission Control Dashboard должен быть default landing page staff console для active work items, discussion, PR и агентов. |
| `FR-337-02` | Dashboard должен работать от typed entity/relation model (`Work Item`, `Discussion`, `Pull Request`, `Agent`, `Comment/Timeline projection`, `Relation`), а не от набора разрозненных списков. |
| `FR-337-03` | Пользователь должен видеть active set по состояниям `working`, `waiting`, `blocked`, `review` и `recent critical updates`, а не полный архив по умолчанию. |
| `FR-337-04` | Dashboard должен поддерживать search, filters и list fallback, чтобы high-volume сценарии не ломали usability primary view. |
| `FR-337-05` | Выбор сущности должен открывать side panel с деталями, timeline/comments, связанными объектами, текущими блокерами и разрешёнными next actions. |
| `FR-337-06` | Discussion-first flow обязан быть first-class сценарием: создать discussion, продолжить discussion и формализовать её в task с сохранением relation и истории. |
| `FR-337-07` | Dashboard должен показывать как внутреннее состояние платформы, так и внешний provider state по task/PR/comment потокам. |
| `FR-337-08` | Действия из console ограничиваются provider-safe и policy-safe набором: создать task/discussion, продолжить stage, открыть provider context, продолжить comment/discussion flow, формализовать discussion. |
| `FR-337-09` | Каждое console-инициированное действие должно иметь явный lifecycle статусов (`queued`, `pending_sync`, `reconciled`, `failed`) на уровне UX и аудита. |
| `FR-337-10` | Webhook echo и повторные delivery не должны создавать дубли task/run/comment сущностей или повторные запуски flows для одного business intent. |
| `FR-337-11` | Realtime updates являются основной моделью обновления active set, но при деградации dashboard обязан оставаться usable через explicit refresh и list fallback. |
| `FR-337-12` | Dashboard не должен создавать обходы label policy, owner approval и review/revise loop; при невозможности действия пользователь видит блокировку, а не скрытый side effect. |
| `FR-337-13` | GitHub-first provider model и external human review сохраняются в MVP; dashboard не заменяет provider UI для review и merge decision. |
| `FR-337-14` | Voice intake, если будет включён в следующих волнах, должен вести только к draft discussion/task flow и не может блокировать core dashboard MVP. |

## Acceptance Criteria (Given/When/Then)

### `AC-337-01` Active-set landing и situational awareness
- Given пользователь открывает staff console в активном рабочем цикле,
- When загружается Mission Control Dashboard,
- Then он видит active set, summary strip, приоритетные блокировки и ближайшие действия без перехода на другие страницы.
- Expected evidence: product acceptance walkthrough + telemetry для `dashboard_opened`, `entity_opened`, `first_action_taken`.

### `AC-337-02` Entity context и side panel
- Given в active set есть work item, discussion, PR или agent,
- When пользователь выбирает сущность,
- Then side panel показывает детали, relations, timeline/comments, sync state и допустимые next actions.
- Expected evidence: acceptance walkthrough с coverage сущностей `Work Item`, `Discussion`, `PR`, `Agent`.

### `AC-337-03` Discussion -> formal task
- Given пользователь начал работу с discussion,
- When он выбирает `Formalize as task`,
- Then создаётся ровно одна связанная formal task, discussion сохраняет историю и relation, а дальнейший stage launch опирается на formal task.
- Expected evidence: flow evidence `discussion_created -> formalized -> linked_task_visible`.

### `AC-337-04` Provider-safe actions и sync status
- Given пользователь инициирует разрешённое действие из dashboard,
- When команда отправляется во внутренний command path и далее во внешний provider,
- Then пользователь видит sync status до reconciliation и понимает, завершено действие, ожидает подтверждения или завершилось ошибкой.
- Expected evidence: audit trail + UX state capture для `action_invoked`, `command_state_changed`, `sync_reconciled`.

### `AC-337-05` Webhook echo dedupe и reconciliation
- Given то же действие инициировано из dashboard и затем приходит webhook echo или повторная доставка,
- When система обрабатывает provider callbacks,
- Then повторные task/run/comment сущности не создаются, а активная карточка показывает единый reconciled result.
- Expected evidence: acceptance matrix по duplicate delivery и business-intent dedupe.

### `AC-337-06` High-volume fallback и degraded realtime mode
- Given active set вырос или realtime временно деградировал,
- When primary board/canvas view становится менее удобным,
- Then пользователь может переключиться на list fallback, применить filters/search и продолжить работу без потери core actions.
- Expected evidence: сценарии `20 / 100 / 400+ entities` и `realtime_degraded`.

### `AC-337-07` Voice stream не блокирует core MVP
- Given в окружении нет готовности по AI policy, ROI или operational readiness для voice,
- When команда принимает решение по MVP release,
- Then Mission Control Dashboard остаётся релизуемым без voice intake, а voice остаётся отдельным candidate stream.
- Expected evidence: release decision package и wave-based scope check.

## Edge cases и non-happy paths

| ID | Сценарий | Ожидаемое поведение | Evidence |
|---|---|---|---|
| `EC-337-01` | Один и тот же webhook delivery приходит повторно после console command | Повторная обработка не создаёт новую task/run/comment запись; audit отражает dedupe | duplicate-delivery acceptance scenario |
| `EC-337-02` | Пользователь изменяет issue/PR только во внешнем provider UI | Dashboard показывает внешний апдейт как synced external change без ложной локальной команды | provider-sync scenario |
| `EC-337-03` | Пользователь или система дважды пытается formalize одну discussion | Создаётся одна formal task; повторная попытка завершается safe no-op или понятным blocked state | discussion-formalization race scenario |
| `EC-337-04` | Realtime канал недоступен или отстаёт | UI показывает degraded/stale state, предлагает explicit refresh и list fallback | degraded-mode scenario |
| `EC-337-05` | Active set превышает usability board/canvas view | Search/filters/list fallback остаются приоритетным способом работы; полный архив не становится default | large-active-set scenario |
| `EC-337-06` | Provider ответил ошибкой или подтверждение задерживается | Пользователь видит `failed` или `pending_sync`, может безопасно понять состояние без скрытого повторного запуска | command-status scenario |
| `EC-337-07` | Voice policy выключена или STT недоступен | Core MVP flow остаётся полностью доступным через text/discussion path | voice-disabled scenario |

## Non-Goals
- Полный исторический graph-view по умолчанию.
- Visual builder произвольных dashboard layouts.
- Замена provider UI для human review и merge decision.
- GitLab parity в первой волне.
- Voice-first обязательный onboarding в core MVP.

## NFR draft для handover в architecture

| ID | Требование | Как измеряем / проверяем |
|---|---|---|
| `NFR-337-01` | p75 времени от открытия dashboard до первого entity-focused действия должен быть `<= 60 секунд` | analytics события `dashboard_opened` и `first_action_taken`, pilot evidence |
| `NFR-337-02` | p95 показа начального active-set snapshot после открытия dashboard должен быть достаточно низким для operational use и целится в `<= 5 секунд` | runtime/product pilot measurement |
| `NFR-337-03` | Доля console-initiated команд без дублей после webhook echo должна быть `>= 99.5%` | reconciliation audit + acceptance matrix |
| `NFR-337-04` | UX должен оставаться usable на active set `20..400+` сущностей за счёт active-set default, filters и list fallback | scenario-based acceptance evidence |
| `NFR-337-05` | Каждое действие из dashboard должно оставлять audit trail с `actor`, `source`, `correlation_id` и lifecycle status | audit evidence + traceability review |
| `NFR-337-06` | Потеря realtime или voice capability не должна блокировать core read/write flows dashboard | degraded-mode acceptance scenarios |
| `NFR-337-07` | Dashboard не должен обходить owner approval, label policy и external human review baseline | policy review + architecture gate |

## Analytics и product evidence
- События:
  - `dashboard_opened`
  - `entity_opened`
  - `first_action_taken`
  - `discussion_created`
  - `discussion_formalized`
  - `action_invoked`
  - `command_state_changed`
  - `sync_reconciled`
  - `realtime_degraded`
  - `voice_intake_started`
  - `voice_intake_confirmed`
- Метрики:
  - `NSM-335-01` Control-plane adoption rate
  - `PM-335-01` Situational awareness latency
  - `PM-335-02` Discussion-to-task lead time
  - `OPS-335-01` Console-start coverage
  - `REL-335-01` Reconciliation correctness

## Handover в `run:arch`
- Follow-up issue: `#340`.
- На архитектурном этапе нельзя потерять:
  - active-set default и list fallback как обязательную часть product contract;
  - external human review и GitHub-first provider model;
  - typed command path с явным sync/reconciliation lifecycle;
  - discussion -> formal task relation как first-class сценарий;
  - conditional статус voice stream без блокировки core MVP.
- Архитектурный этап обязан определить:
  - owner-сервисы для read model, typed commands, provider sync, reconciliation и comments/timeline projection;
  - варианты realtime transport и degraded fallback path;
  - где и как живут relation model, command status и active-set projection;
  - какие design-артефакты и issue нужны для `run:design`.

