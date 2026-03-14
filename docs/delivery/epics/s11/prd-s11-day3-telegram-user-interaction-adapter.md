---
doc_id: PRD-CK8S-S11-I448
type: prd
title: "Telegram-адаптер взаимодействия с пользователем — PRD Sprint S11 Day 3"
status: completed
owner_role: PM
created_at: 2026-03-14
updated_at: 2026-03-14
related_issues: [361, 444, 447, 448, 452]
related_prs: []
related_docsets:
  - docs/delivery/sprints/s11/sprint_s11_telegram_user_interaction_adapter.md
  - docs/delivery/epics/s11/epic_s11.md
  - docs/delivery/issue_map.md
  - docs/delivery/requirements_traceability.md
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-03-14-issue-448-prd"
  approved_by: "ai-da-stas"
  approved_at: 2026-03-14
---

# PRD: Telegram-адаптер взаимодействия с пользователем

## TL;DR
- Что строим: первый внешний channel-specific adapter для platform interaction contract Sprint S10, который доставляет `user.notify` и `user.decision.request` в Telegram и принимает typed user responses через inline callbacks и optional free-text.
- Для кого: конечный пользователь задачи, owner/product lead и platform operator.
- Почему: channel-neutral contract Sprint S10 уже зафиксирован, но без первого внешнего канала платформа не доказывает реальную user-facing ценность и быстрый response path вне GitHub comments.
- MVP: Telegram delivery для notify/decision flows, inline callback response, optional free-text fallback, webhook/callback safety expectations, operator visibility baseline и strict separation from approval flow.
- Критерии успеха: пользователь получает понятный next step или даёт валидный typed ответ в Telegram, а platform lifecycle остаётся audit-safe, channel-neutral и воспроизводимым.

## Проблема и цель
- Problem statement:
  - built-in interaction contract Sprint S10 уже определяет `user.notify` и `user.decision.request`, но у платформы ещё нет первого реального внешнего канала доставки и ответа пользователя;
  - без отдельного Telegram PRD следующий этап быстро уйдёт либо в transport-first webhook дизайн, либо в Telegram-first UX assumptions без проверяемого product contract;
  - без формализации callback/free-text semantics, webhook safety, duplicate/replay handling и operator fallback Telegram stream рискует нарушить platform-owned audit/correlation discipline.
- Цели:
  - зафиксировать core-MVP contract Telegram-адаптера как первого внешнего channel path поверх Sprint S10;
  - описать user stories, FR/AC/NFR и edge cases до architecture/design этапов;
  - сохранить separation between user interaction flow and approval/control flow;
  - подготовить architecture handover с явными product invariants для callback safety, typed outcomes, webhook authenticity expectations и adapter-layer boundaries.
- Почему сейчас:
  - Sprint S11 уже подтвердил на intake и vision этапах, что Telegram должен идти сразу после platform-core stream Sprint S10 и не переопределять его semantics;
  - без PRD stage architecture будет спорить о service boundaries и callback implementation details без product baseline;
  - official Telegram Bot API и reference SDK baseline уже позволяют зафиксировать обязательные продуктовые ограничения, не выбирая преждевременно transport/storage реализацию.

## Зафиксированные продуктовые решения
- `D-448-01`: Telegram остаётся первым channel-specific adapter поверх platform-owned interaction contract Sprint S10 и не может переопределять core semantics.
- `D-448-02`: Core MVP Telegram ограничен `user.notify`, `user.decision.request`, inline callbacks и optional free-text; voice/STT, reminders, rich threads и дополнительные каналы остаются deferred scope.
- `D-448-03`: Approval/control flow и user interaction flow остаются раздельными даже если оба eventually materialize через похожие Telegram affordances.
- `D-448-04`: Delivery, retries/idempotency, duplicate/replay classification, wait-state lifecycle и correlation/audit остаются platform-owned responsibilities, а Telegram adapter materializes готовые semantics.
- `D-448-05`: Callback UX должен оставаться typed and explainable: user selection/free-text нельзя считать валидным ответом без correlation-safe classification и callback acknowledgement path.
- `D-448-06`: Telegram-specific transport constraints и reference SDK baseline влияют на architecture choices, но не меняют product meaning `response_kind`, `selected_option`, `free_text`, `request_status` и `interaction outcome`.

## Scope boundaries
### In scope
- Telegram delivery path для `user.notify` как первого user-facing notification channel.
- Telegram delivery path для `user.decision.request` с 2-5 inline options.
- Callback response path по inline buttons.
- Optional free-text reply как controlled fallback или дополнительный typed answer path.
- Product semantics для:
  - `interaction_id`, `correlation_id`, `response_kind`, `selected_option`, `free_text`, `request_status`, `callback_status`;
  - valid, stale, duplicate, expired, invalid и rejected response outcomes;
  - operator/user fallback path, когда Telegram response path не дал завершить сценарий.
- Telegram-specific expectations, подтверждённые официальной Bot API documentation:
  - webhook и `getUpdates` являются взаимоисключающими режимами получения updates;
  - incoming updates хранятся на стороне Telegram до 24 часов;
  - после нажатия inline callback Telegram client ждёт `answerCallbackQuery`, поэтому acknowledgement path обязателен даже без отдельного текста пользователю;
  - webhook path должен поддерживать secret-token based authenticity expectations.

### Out of scope
- Новый channel-neutral contract вместо Sprint S10 baseline.
- Storage schema, transport protocol, queueing/runtime implementation details до `run:arch` и `run:design`.
- Voice/STT, rich multi-turn conversation threads, advanced reminders, multi-chat routing и дополнительные каналы.
- Превращение Telegram в owner-only approval UI или redesign approval flow под видом interaction scope.
- Прямое копирование reference repositories как готового service-boundary решения.

## Пользователи / персоны

| Persona | Основная задача | Что считает успехом |
|---|---|---|
| Конечный пользователь задачи | Получить уведомление или быстро ответить агенту в привычном мессенджере | Telegram path ясен, short-latency, не требует GitHub comment detour и не заставляет разбираться в внутренней механике платформы |
| Owner / product lead | Быстро принять решение или подтвердить следующий шаг run | `user.notify` и `user.decision.request` приходят в Telegram с понятным контекстом и возвращают валидный typed outcome |
| Platform operator | Сохранять audit/correlation discipline и наблюдаемость первого внешнего канала | Telegram delivery/callback path объясним, platform-owned semantics не теряются, duplicate/replay/expired paths диагностируются |

## User stories и wave priorities

| Story ID | История | Wave | Приоритет |
|---|---|---|---|
| `S11-US-01` | Как конечный пользователь, я хочу получать actionable `user.notify` в Telegram, чтобы понимать результат run и следующий шаг без перехода в GitHub | Wave 1 | `P0` |
| `S11-US-02` | Как конечный пользователь, я хочу отвечать на `user.decision.request` inline-кнопкой в Telegram, чтобы быстро продолжить работу агента | Wave 1 | `P0` |
| `S11-US-03` | Как конечный пользователь, я хочу при необходимости отправить свободный текст вместо выбора только из кнопок, чтобы не терять expressiveness в нестандартных сценариях | Wave 1 | `P0` |
| `S11-US-04` | Как owner/product lead, я хочу получать ясный decision context и typed result из Telegram, чтобы не угадывать смысл пользовательского ответа | Wave 2 | `P0` |
| `S11-US-05` | Как platform operator, я хочу безопасно обрабатывать duplicate/replay/expired callbacks и видеть correlation-safe outcome, чтобы Telegram path не ломал execution lifecycle | Wave 2 | `P0` |
| `S11-US-06` | Как продуктовая команда, я хочу оставить richer conversation UX, reminders и дополнительные каналы вне core MVP, чтобы не блокировать первый production-ready channel path | Wave 3 | `P1` |

### Wave sequencing
- Wave 1 (`core Telegram MVP`, `P0`):
  - actionable `user.notify` delivery;
  - `user.decision.request` с inline options;
  - optional free-text fallback;
  - non-GitHub response path.
- Wave 2 (`callback safety + evidence`, `P0`):
  - callback acknowledgement discipline;
  - duplicate/replay/expired handling expectations;
  - webhook authenticity expectations;
  - operator visibility, fallback clarity и audit evidence.
- Wave 3 (`deferred`, `P1`):
  - voice/STT;
  - reminders;
  - rich threads;
  - multi-chat routing;
  - дополнительные каналы.

## Functional Requirements

| ID | Требование |
|---|---|
| `FR-448-01` | Platform interaction contract Sprint S10 должен materialize в Telegram как первый внешний delivery path для `user.notify` без переопределения core semantics. |
| `FR-448-02` | Telegram adapter должен поддерживать `user.notify` как non-blocking completion/result/next-step path и не переводить сценарий в wait-state. |
| `FR-448-03` | Telegram adapter должен поддерживать `user.decision.request` как typed response path с 2-5 inline options. |
| `FR-448-04` | Для `user.decision.request` Telegram adapter должен поддерживать optional free-text response без потери typed semantics результата. |
| `FR-448-05` | Product contract должен явно различать callback outcomes `accepted`, `stale`, `duplicate`, `expired`, `invalid` и `rejected`, не допуская duplicate logical completion одного decision point. |
| `FR-448-06` | После нажатия inline callback платформа обязана иметь acknowledgement path через `answerCallbackQuery`; Telegram UX не может считаться завершённым без callback acknowledgement semantics. |
| `FR-448-07` | Delivery, retries/idempotency, callback classification, wait-state lifecycle и correlation/audit responsibilities остаются platform-owned и не делегируются Telegram transport logic как source of truth. |
| `FR-448-08` | Approval/control flow не может использоваться как shortcut для Telegram interaction scenarios, даже если UX surface внешне похож. |
| `FR-448-09` | Webhook path должен поддерживать product-level expectations по callback authenticity и secret-token validation без привязки PRD к конкретной реализации. |
| `FR-448-10` | Product contract должен учитывать, что Telegram webhook и polling (`getUpdates`) взаимоисключающи; core MVP ориентирован на webhook-compatible path как production baseline. |
| `FR-448-11` | В случаях, когда Telegram response path не дал валидного результата, система должна оставлять явный fallback path и evidence вместо silent loss. |
| `FR-448-12` | Telegram-specific transport constraints не должны менять channel-neutral meaning полей `response_kind`, `selected_option`, `free_text`, `request_status` и `interaction outcome`. |
| `FR-448-13` | Для core Telegram flows должны быть определены expected product evidence и telemetry, достаточные для acceptance walkthrough и architecture handover. |

## Acceptance Criteria (Given/When/Then)

### `AC-448-01` Actionable `user.notify` в Telegram
- Given агент завершил работу или должен передать человеку понятный следующий шаг,
- When он использует Telegram delivery path для `user.notify`,
- Then пользователь получает actionable notification в Telegram, не попадает в wait-state и понимает, что делать дальше без GitHub fallback по умолчанию.
- Expected evidence: acceptance walkthrough по notify-only path и telemetry отсутствия forced comment detour.

### `AC-448-02` Inline callback response для `user.decision.request`
- Given run не может безопасно продолжаться без решения пользователя,
- When агент отправляет `user.decision.request` в Telegram с inline options,
- Then пользователь может выбрать вариант кнопкой, а платформа получает валидный typed response с корректной correlation semantics и callback acknowledgement.
- Expected evidence: audit trail `request sent -> callback received -> callback acknowledged -> response accepted`.

### `AC-448-03` Optional free-text response
- Given fixed options недостаточны,
- When `user.decision.request` разрешает optional free-text,
- Then пользователь может отправить typed free-text response в Telegram без обхода в GitHub comments, а downstream flow получает понятный `response_kind`.
- Expected evidence: сценарий с `response_kind=free_text` и валидным выходом из wait-state.

### `AC-448-04` Duplicate/replay/expired callback safety
- Given у decision request есть validity window и один logical completion path,
- When приходит duplicate, replayed, stale или expired callback,
- Then платформа не совершает повторный state transition, не теряет correlation и оставляет audit evidence причины отклонения.
- Expected evidence: acceptance matrix по stale/duplicate/expired scenarios.

### `AC-448-05` Webhook authenticity expectations
- Given Telegram доставляет callback/update через webhook path,
- When система принимает inbound update,
- Then продуктовый contract предполагает secret-token based authenticity check и не считает raw inbound request автоматически trustworthy.
- Expected evidence: architecture handover checklist для webhook authenticity и negative scenario review.

### `AC-448-06` Separation from approval flow
- Given в системе уже существует approval/control flow,
- When агент отправляет Telegram notify или decision request,
- Then interaction scenario остаётся в отдельном product path и не использует approval-only semantics, storage path или UX shortcuts.
- Expected evidence: policy review checklist и negative scenario "approval flow is not used".

### `AC-448-07` Fallback clarity
- Given Telegram delivery или response path не дал валидного результата,
- When сценарий не может завершиться через Telegram,
- Then owner/operator получает явный fallback path и evidence вместо silent loss или неясного stuck-state.
- Expected evidence: acceptance scenario `delivery/callback failed -> fallback signaled`.

### `AC-448-08` Deferred scope discipline
- Given команда обсуждает voice/STT, reminders, rich threads, multi-chat routing или дополнительные каналы,
- When принимается решение по core MVP,
- Then эти темы остаются deferred scope и не блокируют release первого Telegram channel path.
- Expected evidence: PRD handover package и deferred-scope decision record.

## Edge cases и non-happy paths

| ID | Сценарий | Ожидаемое поведение | Evidence |
|---|---|---|---|
| `EC-448-01` | Пользователь нажимает callback button, но callback acknowledgement не отправлен вовремя | Сценарий считается broken UX path и должен фиксироваться как defect/evidence gap, а не как успешный response | callback-ack scenario |
| `EC-448-02` | Приходит duplicate/replay callback того же logical response | Logical completion остаётся единственной, duplicate path завершается safe no-op с audit evidence | replay-idempotency scenario |
| `EC-448-03` | Приходит callback после expiry/closure request | Ответ не меняет state и классифицируется как stale/expired | stale-response scenario |
| `EC-448-04` | Пользователь отправляет free-text, хотя request не разрешает такой path | Ответ отклоняется как invalid response kind и не закрывает wait-state ошибочно | invalid-free-text scenario |
| `EC-448-05` | Telegram delivery успешен, но user всё равно уходит в GitHub comment | Это считается measurable fallback и ухудшает `UX-447-01`; core MVP не должен считать такой путь нормой | fallback-to-comments scenario |
| `EC-448-06` | Transport constraints не позволяют безопасно передать весь context в callback payload | Product contract сохраняет typed outcome semantics, а техническая стратегия выносится на `run:arch` как отдельный design question | callback-payload-strategy scenario |
| `EC-448-07` | Telegram webhook доставляет update повторно или с задержкой | Platform path должен остаться correlation-safe и не делать duplicate logical completion | webhook-redelivery scenario |

## Non-Goals
- Telegram-first redesign core interaction domain.
- Универсальный conversational product с длинными тредами.
- Замена approval flow, owner feedback path или GitHub governance механизмов.
- Выбор конкретной DTO/schema/runtime реализации в рамках PRD.
- Voice/STT, reminders, multi-channel orchestration как blocking scope.

## NFR draft для handover в architecture

| ID | Требование | Как измеряем / проверяем |
|---|---|---|
| `NFR-448-01` | Actionable Telegram completion rate для eligible scenarios должна целиться в `>= 80%` | product telemetry + acceptance walkthrough против `NSM-447-01` |
| `NFR-448-02` | p75 времени от отправки `user.decision.request` в Telegram до валидного ответа пользователя должна целиться в `<= 10 минут` | interaction timestamps + `PM-447-01` |
| `NFR-448-03` | Delivery success rate для MVP Telegram scenarios должна целиться в `>= 98%` без ручного operator вмешательства | delivery audit + `REL-447-01` |
| `NFR-448-04` | Callback safety correctness по duplicate/replay/stale paths должна целиться в `>= 99.5%` | callback audit + `REL-447-02` |
| `NFR-448-05` | Separation correctness между Telegram interaction flow и approval flow должна оставаться `100%` | policy review + `GOV-447-01` |
| `NFR-448-06` | Telegram-eligible scenarios не должны уходить в GitHub fallback чаще, чем `20%` случаев | fallback telemetry + `UX-447-01` |
| `NFR-448-07` | Каждый request/response path должен иметь audit trail с actor, correlation, request status, callback status и outcome classification (`100%`) | architecture/design review + audit evidence |

## Analytics и product evidence
- События:
  - `telegram_notify_sent`
  - `telegram_request_sent`
  - `telegram_callback_received`
  - `telegram_callback_acknowledged`
  - `telegram_response_accepted`
  - `telegram_response_rejected`
  - `telegram_fallback_signaled`
  - `telegram_delivery_failed`
- Метрики:
  - `NSM-447-01` Telegram actionable completion rate
  - `PM-447-01` Decision turnaround p75
  - `UX-447-01` GitHub fallback rate
  - `REL-447-01` Delivery success rate
  - `REL-447-02` Callback safety correctness
  - `GOV-447-01` Platform semantics purity
- Expected evidence:
  - acceptance walkthrough по notify/decision/free-text flows;
  - negative matrix по stale/duplicate/invalid callbacks;
  - review trace по webhook authenticity expectations;
  - explicit review trace о том, что approval flow не используется как shortcut.

## Риски и допущения
| Type | ID | Описание | Митигирующее действие | Статус |
|---|---|---|---|---|
| risk | `RSK-448-01` | Telegram stream расползётся в channel-first UX и начнёт диктовать core semantics Sprint S10 | Сохранять adapter-layer framing и sequencing gate как non-negotiable guardrail | open |
| risk | `RSK-448-02` | Callback UX окажется удобным, но неоперабельным при duplicate/replay/expiry и webhook redelivery | Зафиксировать callback safety и operator visibility как blocking quality-gates уже на PRD | open |
| risk | `RSK-448-03` | Free-text path станет слишком свободным и размоет typed outcome semantics | Держать explicit `response_kind` и edge cases; ownership detail передать в `run:arch` | open |
| risk | `RSK-448-04` | Команда попытается решить transport limits за счёт изменения product semantics | Разделить product meaning и technical callback payload strategy, вынести реализацию на `run:arch`/`run:design` | open |
| assumption | `ASM-448-01` | Notify + decision request + inline callbacks + optional free-text достаточно, чтобы подтвердить ценность первого внешнего канала | accepted |
| assumption | `ASM-448-02` | Telegram webhook-based path подходит как production baseline первого адаптера без polling-first модели | accepted |
| assumption | `ASM-448-03` | Reference SDK `telego v1.7.0`, подтверждённый 2026-03-14, достаточно зрелый как implementation baseline без влияния на product contract | accepted |

## Открытые вопросы для `run:arch`
- Как разложить ownership между `control-plane`, `worker`, `api-gateway` и Telegram adapter contour без потери product invariants?
- Где проходит граница между callback ingestion, Telegram delivery, wait-state orchestration и fallback signaling?
- Как сохранить typed outcome semantics при transport constraints callback payload и webhook redelivery без premature storage lock-in?
- Какие architectural alternatives нужно зафиксировать отдельно, чтобы `run:design` не переоткрывал product decisions Day3?

## Handover в `run:arch`
- Follow-up issue: `#452`.
- На архитектурном этапе нельзя потерять:
  - Telegram как adapter-layer stream поверх Sprint S10 interaction contract;
  - `user.notify` как non-blocking path;
  - `user.decision.request` как typed response path с inline callbacks и optional free-text;
  - callback acknowledgement, duplicate/replay/expired handling и webhook authenticity expectations;
  - separation between interaction flow and approval/control flow;
  - deferred scope для voice/STT, reminders, rich threads и дополнительных каналов.
- Архитектурный этап обязан определить:
  - service boundaries и ownership matrix;
  - callback/webhook lifecycle и fallback path;
  - audit/correlation responsibilities;
  - ADR/alternatives backlog и issue для `run:design`, в которой будет повторено continuity-требование довести цепочку до `run:dev`.
