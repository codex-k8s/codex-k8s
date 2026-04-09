---
doc_id: PRD-CK8S-S10-D3-MCP-INTERACTIONS
type: prd
title: "Built-in MCP user interactions — PRD Sprint S10 Day 3"
status: in-review
owner_role: PM
created_at: 2026-03-12
updated_at: 2026-03-13
related_issues: [360, 378, 383, 385]
related_prs: []
related_docsets:
  - docs/delivery/sprints/s10/sprint_s10_mcp_user_interactions.md
  - docs/delivery/epics/s10/epic_s10.md
  - docs/delivery/issue_map.md
  - docs/delivery/requirements_traceability.md
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-03-12-issue-383-prd"
  approved_by: "ai-da-stas"
  approved_at: 2026-03-13
---

# PRD: Built-in MCP user interactions

## TL;DR
- Что строим: channel-neutral built-in interaction contract для `user.notify` и `user.decision.request` внутри `kodex`.
- Для кого: owner/product lead, конечный пользователь задачи, platform operator и будущий adapter stream.
- Почему: сегодня user-facing next steps и decision points либо теряются в GitHub comments, либо рискуют смешаться с approval flow.
- MVP: actionable `user.notify`, typed `user.decision.request`, wait-state discipline, typed response semantics, platform-owned correlation/audit/retry expectations и adapter-neutral boundaries.
- Критерии успеха: пользователь получает понятный next step или даёт валидный typed ответ без GitHub-comment detour, а platform lifecycle остаётся воспроизводимым и не смешивается с approval semantics.

## Проблема и цель
- Problem statement:
  - built-in MCP baseline платформы пока не даёт канонического user-facing path для понятного completion signal и запроса решения пользователя;
  - без отдельного interaction-domain команда будет либо пользоваться GitHub comments как ad-hoc каналом, либо ошибочно расширять approval flow под задачи user interaction;
  - без typed response semantics, wait-state policy и correlation expectations следующий этап быстро уйдёт в Telegram-first компромисс или transport-first дизайн без общего продукта.
- Цели:
  - зафиксировать core-MVP contract для `user.notify` и `user.decision.request`;
  - описать пользовательские сценарии, FR/AC/NFR и edge cases до architecture/design этапов;
  - сохранить separation between user interaction flow and approval/control flow;
  - подготовить architecture handover с явными product invariants для typed responses, wait-state lifecycle и adapter-neutral semantics.
- Почему сейчас:
  - Sprint S10 уже подтвердил отдельный product stream по built-in MCP user interactions на intake и vision этапах;
  - без PRD stage architecture будет спорить о transport/storage/runtime деталях без общего baseline требований;
  - сейчас лучшее окно отделить core platform contract от Telegram/adapters и richer conversation scope, пока реализация ещё не началась.

## Пользователи / Персоны

| Persona | Основная задача | Что считает успехом |
|---|---|---|
| Owner / product lead | Получить понятный итог run или быстро принять решение без изучения issue comments и label mechanics | `user.notify` даёт ясный completion/next step, а `user.decision.request` быстро возвращает валидный ответ |
| Конечный пользователь задачи | Ответить агенту кнопкой или свободным текстом без погружения в GitHub workflow | response path прост, понятен и не требует ручного разбора технических деталей |
| Platform operator | Держать platform-owned lifecycle, audit и correlation под контролем | interaction flow воспроизводим, наблюдаем и не смешивается с approval flow |
| Future adapter stream | Подключить Telegram или другой канал без переопределения core semantics | outbound/inbound contract остаётся channel-neutral, а channel-specific UX живёт поверх core contract |

## Scope boundaries
### In scope
- `user.notify` как built-in tool для:
  - completion signal;
  - actionable next step;
  - status/result update, который понятен человеку без дополнительного перехода в GitHub comments.
- `user.decision.request` как built-in tool для:
  - 2-5 предопределённых вариантов ответа;
  - optional free-text, когда фиксированных вариантов недостаточно;
  - typed user response, пригодного для дальнейшего platform processing.
- Product semantics interaction-domain:
  - typed meaning для `interaction_id`, `correlation_id`, `response_kind`, `selected_option`, `free_text`, `request_status`;
  - явная validity discipline для open/answered/expired/rejected response paths;
  - wait-state разрешён только для response-required сценариев.
- Platform-owned behavior expectations:
  - callback validity, retries/idempotency, correlation и audit evidence;
  - separation from approval/control flow;
  - channel-neutral adapter model и guardrails для будущих integrations.

### Out of scope
- Новый runtime server block вместо расширения built-in `kodex`.
- Telegram-specific payloads, UI affordances, delivery preferences, reminders и escalation logic.
- Voice/STT, richer multi-turn conversation threads и long-lived conversational UX.
- Storage schema, transport protocol, queueing/runtime implementation details до `run:arch` и `run:design`.
- Любой redesign approval flow под видом interaction scope.

## User stories и wave priorities

| Story ID | История | Wave | Приоритет |
|---|---|---|---|
| `S10-US-01` | Как owner/product lead, я хочу получить actionable `user.notify`, чтобы понимать итог run и следующий шаг без ручного follow-up в issue comments | Wave 1 | `P0` |
| `S10-US-02` | Как конечный пользователь, я хочу ответить на `user.decision.request` кнопкой из списка, чтобы быстро продолжить работу агента | Wave 1 | `P0` |
| `S10-US-03` | Как конечный пользователь, я хочу при необходимости отправить свободный текст как typed response, чтобы не быть ограниченным только предустановленными опциями | Wave 1 | `P0` |
| `S10-US-04` | Как platform operator, я хочу видеть детерминированный wait-state и audit/correlation lifecycle, чтобы retries и callback replays не ломали execution | Wave 2 | `P0` |
| `S10-US-05` | Как owner будущего adapter stream, я хочу подключать Telegram и другие каналы поверх channel-neutral contract, чтобы не зашивать vendor-specific semantics в core MVP | Wave 2 | `P0` |
| `S10-US-06` | Как продуктовая команда, я хочу оставить richer conversations, delivery preferences и voice/STT вне core MVP, чтобы не блокировать базовую capability | Wave 3 | `P1` |

### Wave sequencing
- Wave 1 (`core MVP`, `P0`):
  - actionable `user.notify`;
  - `user.decision.request` с typed option/free-text response;
  - response validation semantics;
  - separation from approval flow.
- Wave 2 (`platform evidence`, `P0`):
  - wait-state lifecycle;
  - callback validity/replay expectations;
  - retries/idempotency/audit/correlation guardrails;
  - adapter-ready channel-neutral semantics.
- Wave 3 (`deferred`, `P1`):
  - Telegram/adapters;
  - reminder/delivery preferences;
  - richer threads;
  - voice/STT.

## Functional Requirements

| ID | Требование |
|---|---|
| `FR-383-01` | Built-in MCP surface должна включать `user.notify` как канонический non-blocking path для completion signal, result update и понятного next step. |
| `FR-383-02` | `user.notify` не переводит run в paused/waiting state и не требует обязательного ответа пользователя для завершения своего сценария. |
| `FR-383-03` | Built-in MCP surface должна включать `user.decision.request` как typed request/response path для ситуаций, где без решения пользователя run не может безопасно продолжаться. |
| `FR-383-04` | `user.decision.request` должен поддерживать 2-5 вариантов ответа и optional free-text path без потери typed semantics результата. |
| `FR-383-05` | Product contract должен явно определять минимальные meaning-поля interaction-domain: `interaction_id`, `correlation_id`, `response_kind`, `selected_option`, `free_text`, `request_status`; точное transport/storage представление выбирается позже. |
| `FR-383-06` | Платформа должна различать valid, stale, duplicate, invalid и rejected response paths и не допускать повторного logical completion одного и того же decision point. |
| `FR-383-07` | Wait-state допускается только для response-required сценариев `user.decision.request`; notification-only path остаётся non-blocking. |
| `FR-383-08` | Interaction-domain и approval/control-domain остаются раздельными: approval-only semantics, storage path и UX не используются как shortcut для user interaction flows. |
| `FR-383-09` | Delivery, retries/idempotency, callback correlation, audit evidence и outcome classification принадлежат platform domain и не могут делегироваться channel-specific adapters. |
| `FR-383-10` | Outbound/inbound interaction contract остаётся channel-neutral: adapters обязаны соблюдать core semantics `response_kind`, `selected_option`, `free_text`, correlation и outcome classification. |
| `FR-383-11` | Telegram и другие adapters остаются отдельным follow-up stream после фиксации core platform contract и не блокируют core Sprint S10 MVP. |
| `FR-383-12` | Для core interaction flows должны быть определены expected product evidence и telemetry, достаточные для acceptance walkthrough и architecture handover. |

## Acceptance Criteria (Given/When/Then)

### `AC-383-01` Actionable `user.notify`
- Given агент завершил работу или должен передать понятный следующий шаг человеку,
- When он использует `user.notify`,
- Then пользователь получает completion/result signal, не попадает в wait-state и понимает, нужно ли ему что-то делать дальше.
- Expected evidence: acceptance walkthrough + telemetry для отправки actionable notification и последующего отсутствия forced comment fallback.

### `AC-383-02` Typed option response для `user.decision.request`
- Given run не может продолжаться без решения пользователя,
- When агент отправляет `user.decision.request` с вариантами ответа,
- Then пользователь может выбрать один из вариантов, а платформа получает валидный typed response с корректной correlation semantics.
- Expected evidence: interaction audit trail `request sent -> response accepted -> wait-state exited`.

### `AC-383-03` Optional free-text response
- Given фиксированных вариантов ответа недостаточно,
- When `user.decision.request` разрешает optional free-text,
- Then пользователь может отправить typed free-text response без обхода в GitHub comments, а semantics ответа остаются понятными downstream flow.
- Expected evidence: acceptance scenario с `response_kind=free_text` и валидным выходом из wait-state.

### `AC-383-04` Late/duplicate/invalid response safety
- Given у decision request есть validity window и один logical response path,
- When приходит late, duplicate, foreign или syntactically invalid response,
- Then платформа не совершает повторный state transition, не теряет correlation и оставляет audit evidence причины отклонения.
- Expected evidence: acceptance matrix по stale/duplicate/invalid callback сценариям.

### `AC-383-05` Separation from approval flow
- Given в системе уже существует approval/control flow,
- When агент решает уведомить пользователя или запросить его решение,
- Then interaction scenario проходит через отдельный product path и не использует approval-only semantics, storage path или UX shortcuts.
- Expected evidence: policy review checklist + explicit negative scenario “approval flow is not used”.

### `AC-383-06` Channel-neutral adapter semantics
- Given core MVP реализуется до Telegram и других каналов,
- When команда рассматривает adapter integrations,
- Then core interaction contract сохраняет channel-neutral meaning, а channel-specific affordances остаются follow-up scope.
- Expected evidence: PRD handover package и deferred-scope decision record для Telegram/adapters.

### `AC-383-07` Wait-state discipline и evidence
- Given run переведён в ожидание из-за `user.decision.request`,
- When пользователь даёт валидный ответ или request становится invalid/expired,
- Then система оставляет понятный exit path из wait-state и достаточный audit evidence для разбора сценария.
- Expected evidence: lifecycle walkthrough `wait entered -> response accepted/rejected -> wait exited`.

## Edge cases и non-happy paths

| ID | Сценарий | Ожидаемое поведение | Evidence |
|---|---|---|---|
| `EC-383-01` | `user.notify` отправлен без явного “что дальше” | Сообщение всё равно классифицируется как completion/result/next-step и не создаёт wait-state | actionable-notify review |
| `EC-383-02` | Для `user.decision.request` пришёл ответ после того, как request уже закрыт или истёк | Ответ не меняет state, а аудит фиксирует stale/expired rejection | stale-response scenario |
| `EC-383-03` | Приходит duplicate callback/retry того же logical response | Logical completion остаётся единственной, duplicate path завершается safe no-op с audit evidence | replay-idempotency scenario |
| `EC-383-04` | Пользователь отправляет free-text, хотя request допускает только варианты из списка | Ответ отклоняется как invalid, wait-state не завершается ошибочно | invalid-response-kind scenario |
| `EC-383-05` | Агент или адаптер пытается использовать approval semantics вместо interaction path | Сценарий считается policy violation и не входит в допустимый happy path | separation-negative scenario |
| `EC-383-06` | Adapter-specific UX требует дополнительного поля, которого нет в core contract | Дополнительный affordance не меняет core meaning и уходит в follow-up stream/architecture discussion | adapter-extension scenario |
| `EC-383-07` | Пользователь вынужден уйти в GitHub comment, потому что response path не сработал | Это считается fallback и метрика GitHub-comment fallback rate ухудшается; core MVP не должен считать это нормой | fallback-to-comments scenario |

## Non-Goals
- Telegram-first MVP.
- Универсальный чат/мессенджер с длинными тредами.
- Замена approval flow и owner feedback path.
- Выбор конкретного transport/storage/runtime решения в рамках PRD.
- Голосовой UX, STT и reminder/delivery policy как blocking scope.

## NFR draft для handover в architecture

| ID | Требование | Как измеряем / проверяем |
|---|---|---|
| `NFR-383-01` | Actionable interaction completion rate для core scenarios должен целиться в `>= 85%` | product telemetry + acceptance walkthrough против `NSM-378-01` |
| `NFR-383-02` | p75 времени от отправки `user.decision.request` до получения валидного typed response должен целиться в `<= 15 минут` | timestamps interaction lifecycle + `PM-378-02` |
| `NFR-383-03` | Separation correctness между interaction flow и approval flow должна оставаться `100%` | policy review + negative acceptance scenarios против `GOV-378-01` |
| `NFR-383-04` | Correlation/replay correctness должна целиться в `>= 99.5%` без duplicate logical completion | retry/callback audit + `REL-378-01` |
| `NFR-383-05` | `user.notify` не должен переводить run в wait-state ни в одном happy-path сценарии | acceptance matrix по notify-only flows |
| `NFR-383-06` | Каждый interaction request/response path должен иметь audit trail с actor, correlation, request status и outcome classification | architecture/design review + audit evidence |
| `NFR-383-07` | Channel-specific affordances не должны менять core semantics `response_kind`, `selected_option`, `free_text` и outcome classification | adapter contract review на `run:arch`/`run:design` |

## Analytics и product evidence
- События:
  - `interaction_notify_sent`
  - `interaction_request_sent`
  - `interaction_wait_entered`
  - `interaction_response_received`
  - `interaction_response_accepted`
  - `interaction_response_rejected`
  - `interaction_wait_exited`
  - `interaction_fallback_to_comment`
- Метрики:
  - `NSM-378-01` Actionable interaction completion rate
  - `PM-378-01` Next-step clarity rate
  - `PM-378-02` Decision turnaround p75
  - `GOV-378-01` Interaction vs approval separation correctness
  - `REL-378-01` Correlation and replay correctness
  - `UX-378-01` GitHub-comment fallback rate
- Expected evidence:
  - acceptance walkthrough по notify/decision/free-text flows;
  - negative scenarios по stale/duplicate/invalid callbacks;
  - явный review trace о том, что approval flow не используется как shortcut.

## Риски и допущения
| Type | ID | Описание | Митигирующее действие | Статус |
|---|---|---|---|---|
| risk | `RSK-383-01` | Инициатива может расползтись в Telegram-first UX до фиксации core contract | Сохранять adapters как Wave 3/follow-up stream и проверять scope на каждом handover | open |
| risk | `RSK-383-02` | Typed response semantics окажутся слишком размытыми и создадут неоднозначность downstream processing | Зафиксировать minimal meaning-fields и edge cases уже в PRD, а конкретику владения перенести в `run:arch` | open |
| risk | `RSK-383-03` | Wait-state, retries и correlation могут быть ошибочно распределены между агентом, адаптером и платформой | Сохранить platform-owned lifecycle как non-negotiable guardrail в handover issue `#385` | open |
| risk | `RSK-383-04` | Notify path станет информационным шумом вместо actionable completion signal | Держать next-step clarity rate и fallback rate как quality-gates | open |
| assumption | `ASM-383-01` | Existing built-in `kodex` surface можно расширить без отдельного runtime server block | accepted |
| assumption | `ASM-383-02` | Пользователю достаточно typed options/free-text path для core MVP без richer conversation threads | accepted |
| assumption | `ASM-383-03` | Channel-neutral semantics можно сохранить, даже если первый adapter окажется Telegram | accepted |

## Открытые вопросы для `run:arch`
- Как разложить ownership между `control-plane`, `worker`, `api-gateway` и built-in MCP surface, не потеряв product invariants?
- Где проходит граница между callback ingestion, lifecycle orchestration и adapter-specific transformation?
- Как сохранить typed semantics для response validation и replay safety без premature lock-in на storage/transport details?
- Какие architectural alternatives нужно зафиксировать отдельно, чтобы `run:design` не переоткрывал продуктовые решения Day3?

## Handover в `run:arch`
- Follow-up issue: `#385`.
- На архитектурном этапе нельзя потерять:
  - built-in server `kodex` как единственную core точку расширения;
  - `user.notify` как non-blocking path;
  - `user.decision.request` как единственный core wait-state path;
  - separation between interaction flow and approval/control flow;
  - channel-neutral meaning-fields и deferred scope для Telegram/adapters, voice/STT, richer threads.
- Архитектурный этап обязан определить:
  - service boundaries и ownership matrix;
  - callback lifecycle и wait-state path;
  - audit/correlation responsibilities;
  - список ADR/alternatives и issue для `run:design`.
