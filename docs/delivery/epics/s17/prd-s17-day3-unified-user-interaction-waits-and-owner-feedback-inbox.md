---
doc_id: PRD-CK8S-S17-D3-OWNER-FEEDBACK-WAITS
type: prd
title: "Unified long-lived owner feedback waits and inbox — PRD Sprint S17 Day 3"
status: in-review
owner_role: PM
created_at: 2026-03-25
updated_at: 2026-03-25
related_issues: [360, 361, 458, 532, 540, 541, 554, 557, 559]
related_prs: []
related_docsets:
  - docs/delivery/sprints/s17/sprint_s17_unified_user_interaction_waits_and_owner_feedback_inbox.md
  - docs/delivery/epics/s17/epic_s17.md
  - docs/delivery/issue_map.md
  - docs/delivery/requirements_traceability.md
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-25-issue-557-prd"
---

# PRD: Unified long-lived owner feedback waits and inbox

## TL;DR
- Что строим: unified owner feedback loop и long-lived wait contract, где pending request живёт в одном platform-owned контуре, а owner отвечает в Telegram или staff-console.
- Для кого: owner / product lead, live runtime path агента и staff/operator fallback.
- Почему: текущие built-in user interactions и Telegram path уже существуют, но ещё не дают проверяемую гарантию, что агент действительно ждёт пользователя в той же задаче и прозрачно проходит lifecycle `delivery -> wait -> response -> continuation`.
- MVP: same-session continuation как primary happy-path, max timeout/TTL built-in `kodex` MCP wait path не ниже owner wait window, snapshot-resume только как recovery fallback, dual-channel inbox, delivery-before-wait lifecycle transparency и deterministic text/voice/callback binding.
- Критерии успеха: owner видит pending request, отвечает из разрешённого канала без GitHub-comment detour, система сохраняет long wait `>=24h`, не путает delivery с user response и либо продолжает ту же live session, либо явно классифицирует recovery fallback.

## Проблема и цель
- Problem statement:
  - built-in interaction tools Sprint S10 и Telegram adapter Sprint S11 ещё не сведены в единый owner-facing contract, поэтому ожидание пользователя выглядит как набор разрозненных transport/runtime paths;
  - без отдельного PRD-stage команда может нормализовать synthetic resume, короткий tool timeout или Telegram-first UX как фактический happy-path вместо same-session waiting baseline;
  - отсутствие явной scenario/evidence matrix создаёт риск, что architecture/design зафиксируют transport и storage детали раньше, чем будут закрыты продуктовые решения по lifecycle transparency, inbox parity и recovery semantics.
- Цели:
  - формализовать product contract для unified owner feedback loop поверх channel-neutral persisted backend;
  - закрепить same-session continuation как primary happy-path, а snapshot-resume как recovery-only fallback;
  - определить user stories, FR/AC/NFR, scenario matrix и expected evidence для owner inbox, runtime continuity и operator visibility;
  - передать в `run:arch` набор продуктовых инвариантов без reopening Day1/Day2 baseline.
- Почему сейчас:
  - Issue `#541` и Issue `#554` уже закрепили problem framing, mission, KPI/guardrails и locked baseline;
  - launch profile `new-service` требует обязательный `run:prd` перед `run:arch`;
  - без PRD-stage следующий этап будет спорить о lifetime, ownership и recovery semantics без проверяемого продукта.

## Зафиксированные продуктовые решения
- `D-557-01`: unified owner feedback loop остаётся platform capability поверх Sprint S10/S11 baseline, а не Telegram-first или GitHub-comment-driven workaround.
- `D-557-02`: same live pod / same `codex` session остаётся primary happy-path для response-required сценариев.
- `D-557-03`: built-in `kodex` MCP wait path обязан иметь effective max timeout/TTL не ниже owner wait window, чтобы happy-path оставался реальным live wait, а не synthetic resume.
- `D-557-04`: persisted session snapshot и resume разрешены только как recovery/degradation fallback при потере live runtime.
- `D-557-05`: обязательный lifecycle для core сценариев: `created -> delivery pending -> delivery accepted -> waiting -> response -> continuation`.
- `D-557-06`: delivery accepted и waiting for user response остаются разными lifecycle состояниями; “сообщение доставлено” не означает “пользователь уже ответил”.
- `D-557-07`: Telegram pending inbox и staff-console fallback обязаны работать поверх одного persisted backend contract; канал не становится владельцем семантики.
- `D-557-08`: text reply, inline callback и voice reply должны детерминированно связываться с исходным request и не создавать duplicate logical completion.
- `D-557-09`: overdue / expired / manual-fallback сценарии обязаны быть product-visible для owner/operator, а не hidden operator-only path.
- `D-557-10`: `run:self-improve` остаётся вне owner-facing human-wait contract.
- `D-557-11`: дополнительные каналы, advanced reminders/escalations, attachments, multi-party routing, richer conversation UX и detached resume-run как равноправный happy-path остаются deferred scope.

## Scope boundaries
### In scope
- Unified owner feedback loop для response-required сценариев всех stage/run-типов, кроме `run:self-improve`.
- Pending inbox semantics:
  - owner видит pending request в Telegram и staff-console fallback без потери контекста;
  - ответ из разрешённого канала продолжает исходный run через общий persisted request state.
- Wait/continuation contract:
  - same-session continuation как primary happy-path;
  - long human-wait target `>=24h`;
  - max timeout/TTL built-in `kodex` MCP wait path не ниже owner wait window;
  - snapshot-resume как recovery-only fallback;
  - явная классификация overdue / expired / manual-fallback / recovery paths.
- Deterministic response binding:
  - text reply;
  - inline callback;
  - voice reply;
  - запрет duplicate logical completion одного request.
- Expected evidence и telemetry для owner feedback loop, пригодные для architecture/design handover и acceptance walkthrough.

### Out of scope
- Кодовая реализация, storage/schema/runtime lock-in и transport/API детали до `run:arch` и `run:design`.
- Попытка сделать detached resume-run равноправным default UX.
- Дополнительные каналы кроме Telegram и staff-console fallback.
- Advanced reminders, escalations, attachments, multi-party routing и richer conversation platform.
- Расширение owner-facing human-wait contract на `run:self-improve`.

## Пользователи / персоны

| Persona | Основная работа | Что считает успехом |
|---|---|---|
| Owner / product lead | Получить pending request, понять, что агент действительно ждёт ответ, и ответить из удобного канала | Виден контекст запроса, lifecycle не выглядит “зависшим”, а ответ приводит к понятному continuation outcome |
| Live runtime path агента | Сохранить ту же задачу и ту же session как основной путь до ответа пользователя | Wait-state остаётся live, tool timeout/TTL не обрывает happy-path раньше owner wait window |
| Staff/operator | Диагностировать pending, overdue, expired и manual-fallback случаи без потери продуктового смысла | Staff-console показывает те же request/lifecycle states и recovery/fallback classification, что и owner-facing path |

## User stories и wave priorities

| Story ID | История | Wave | Приоритет |
|---|---|---|---|
| `S17-US-01` | Как owner / product lead, я хочу видеть pending request в Telegram или staff-console с достаточным контекстом, чтобы ответить без GitHub-comment detour | Wave 1 | `P0` |
| `S17-US-02` | Как owner / product lead, я хочу быть уверенным, что агент реально ждёт мой ответ в той же задаче, а не эмулирует continuation после таймаута | Wave 1 | `P0` |
| `S17-US-03` | Как runtime path агента, я хочу сохранять same-session continuation как primary happy-path, чтобы не терять контекст и не создавать лишний recovery path | Wave 1 | `P0` |
| `S17-US-04` | Как staff/operator, я хочу видеть те же pending requests и lifecycle states в staff-console, чтобы fallback не терял продуктовую правду | Wave 1 | `P0` |
| `S17-US-05` | Как owner / operator, я хочу, чтобы text, voice и callback responses детерминированно связывались с исходным request, чтобы continuation не становился ambiguous | Wave 2 | `P0` |
| `S17-US-06` | Как owner / operator, я хочу видеть overdue, expired, manual-fallback и recovery scenarios как явные состояния, а не “тихо пропавший run” | Wave 2 | `P0` |
| `S17-US-07` | Как продуктовая команда, я хочу оставить дополнительные каналы, reminders, attachments, multi-party routing и detached resume-run за пределами core MVP, чтобы не размывать базовый контракт | Wave 3 | `P1` |

### Wave sequencing
- Wave 1 (`core MVP`, `P0`):
  - owner inbox через Telegram + staff-console fallback;
  - same-session continuation как primary happy-path;
  - delivery-before-wait lifecycle;
  - max timeout/TTL baseline built-in `kodex` MCP wait path.
- Wave 2 (`platform evidence`, `P0`):
  - deterministic text/voice/callback binding;
  - overdue / expired / manual-fallback visibility;
  - recovery-only snapshot-resume classification;
  - acceptance evidence и telemetry.
- Wave 3 (`deferred`, `P1`):
  - дополнительные каналы;
  - advanced reminders/escalations;
  - attachments;
  - multi-party routing;
  - richer conversation UX;
  - detached resume-run как отдельный продуктовый поток, если owner когда-либо решит его переоткрыть.

## Functional Requirements

| ID | Требование |
|---|---|
| `FR-557-01` | Платформа должна предоставлять unified owner feedback loop как channel-neutral capability поверх existing built-in interaction tools и Telegram/staff surfaces. |
| `FR-557-02` | Для response-required сценариев owner должен видеть pending request с достаточным контекстом и разрешённым path ответа без обязательного ухода в GitHub comments. |
| `FR-557-03` | Telegram pending inbox и staff-console fallback должны отображать один и тот же persisted request state и не расходиться по lifecycle semantics. |
| `FR-557-04` | Same live pod / same `codex` session должны оставаться primary happy-path для continuation response-required run-сценариев. |
| `FR-557-05` | Effective timeout/TTL built-in `kodex` MCP wait path должен быть не ниже owner wait window и не должен обрывать happy-path раньше long human-wait target. |
| `FR-557-06` | Persisted session snapshot и resume могут использоваться только как recovery/degradation fallback и должны классифицироваться отдельно от happy-path same-session continuation. |
| `FR-557-07` | Product contract должен явно фиксировать lifecycle `created -> delivery pending -> delivery accepted -> waiting -> response -> continuation` и distinction между delivery accepted и user response received. |
| `FR-557-08` | Overdue, expired, manual-fallback и recovery scenarios должны быть видимыми состояниями для owner/operator, а не скрытыми runtime side effects. |
| `FR-557-09` | Text reply, inline callback и voice reply должны детерминированно связываться с исходным request без duplicate logical completion и без ambiguous continuation path. |
| `FR-557-10` | Telegram и staff-console остаются UX-поверхностями platform-owned semantics; канал не может переопределять persisted request truth, lifecycle и continuation classification. |
| `FR-557-11` | GitHub comment fallback не считается нормальным happy-path и должен трактоваться как деградационный сценарий с явной классификацией. |
| `FR-557-12` | `run:self-improve` остаётся исключением: owner-facing long-lived wait contract не применяется к этому run-типу. |
| `FR-557-13` | Для core owner feedback flows должны быть определены expected evidence, product telemetry и quality signals, достаточные для acceptance walkthrough и architecture handover. |
| `FR-557-14` | Дополнительные каналы, reminders/escalations, attachments, multi-party routing, richer conversation UX и detached resume-run как равноправный happy-path не должны блокировать core MVP Sprint S17. |
| `FR-557-15` | Long human-wait target `>=24h` должен одновременно отражаться в interaction lifecycle, wait-state semantics, runtime lifetime expectations и acceptance evidence, а не быть только UI-обещанием. |

## Acceptance Criteria (Given/When/Then)

### `AC-557-01` Telegram inbox path
- Given owner получает response-required request через Telegram,
- When он отвечает inline option, text reply или voice reply,
- Then платформа принимает ответ как часть unified owner feedback loop, связывает его с исходным request и продолжает run через общий persisted contract.
- Expected evidence: acceptance walkthrough по Telegram path + audit trail `delivery accepted -> waiting -> response received -> continuation`.

### `AC-557-02` Staff-console fallback parity
- Given owner не использует Telegram или оператору нужен fallback,
- When тот же pending request открывается в staff-console,
- Then staff surface показывает тот же request state, те же lifecycle markers и допустимый path ответа/диагностики без потери продуктового контекста.
- Expected evidence: scenario walkthrough по staff fallback с тем же request identifier и same lifecycle projection.

### `AC-557-03` Delivery-before-wait lifecycle
- Given система отправляет owner-facing request,
- When request доставлен, но пользователь ещё не ответил,
- Then lifecycle явно проходит через `delivery pending -> delivery accepted -> waiting`, не подменяя delivery фактом пользовательского ответа.
- Expected evidence: lifecycle trace с отдельными событиями delivery acceptance и wait entry.

### `AC-557-04` Same-session happy-path и TTL baseline
- Given response-required scenario должен ждать owner до long human-wait window,
- When built-in `kodex` MCP wait path остаётся активным,
- Then live session остаётся primary happy-path, а effective max timeout/TTL не короче owner wait window и не вынуждает system normalise synthetic resume как основной UX.
- Expected evidence: runtime/session evidence, подтверждающий max timeout/TTL baseline и same-session continuation classification.

### `AC-557-05` Recovery-only snapshot resume
- Given live runtime был потерян до получения ответа пользователя,
- When continuation всё же восстанавливается через persisted snapshot,
- Then сценарий явно классифицируется как recovery fallback, а не как нормальный happy-path same-session continuation.
- Expected evidence: acceptance scenario с recovery classification и явным разграничением happy-path vs fallback.

### `AC-557-06` Deterministic reply binding
- Given request допускает text, voice или callback reply,
- When пользователь отвечает любым из разрешённых способов,
- Then платформа детерминированно связывает ответ с исходным request и не допускает duplicate logical completion.
- Expected evidence: scenario matrix по text/voice/callback reply binding и negative cases по duplicate/stale responses.

### `AC-557-07` Transparency for overdue and fallback states
- Given owner не ответил вовремя или требуется manual fallback,
- When request становится overdue, expired или manual-fallback,
- Then owner/operator видят явный статус, причину и continuation classification вместо “тихо зависшего run”.
- Expected evidence: acceptance walkthrough по overdue / expired / manual-fallback states и operator visibility review.

### `AC-557-08` Deferred scope discipline
- Given дополнительные каналы, reminders, attachments, multi-party routing, richer conversation UX или detached resume-run недоступны,
- When оценивается готовность core MVP,
- Then unified owner feedback loop остаётся готовым к следующему stage без нормализации этих contours как blocking scope.
- Expected evidence: scope review checklist и wave-based readiness note.

## Scenario matrix

| ID | Сценарий | Обязательное поведение | Expected evidence |
|---|---|---|---|
| `SC-557-01` | Owner отвечает на pending request в Telegram | Pending request имеет достаточный контекст, ответ связывается с тем же request, continuation не требует GitHub detour | Telegram walkthrough + lifecycle audit |
| `SC-557-02` | Owner или operator использует staff-console fallback | Виден тот же request state и lifecycle, что и в Telegram path; fallback не создаёт второй source of truth | Staff-console walkthrough + parity review |
| `SC-557-03` | Request доставлен, но owner отвечает спустя часы | Run остаётся в `waiting`, а long human-wait budget и lifecycle transparency сохраняются до ответа или явного fallback outcome | Lifecycle timeline review + wait window evidence |
| `SC-557-04` | Same-session continuation проходит как happy-path | Ответ owner продолжает ту же live session и не требует synthetic resume-path | Run/session evidence + continuation classification |
| `SC-557-05` | Live session потеряна во время wait-state | Snapshot-resume срабатывает как recovery-only path и явно помечается как fallback | Recovery scenario review |
| `SC-557-06` | Text/voice/callback replies приходят по одному request | Все reply types связываются с исходным request детерминированно и не создают duplicate logical completion | Binding matrix + negative duplicate case |
| `SC-557-07` | Request становится overdue / expired / manual-fallback | Owner/operator видят явный статус, причину и допустимый дальнейший action | Transparency walkthrough + operator review |
| `SC-557-08` | `run:self-improve` использует user-facing interaction semantics | Сценарий считается out of scope и не включается в owner-facing wait contract | Scope review note |

## Edge cases и non-happy paths

| ID | Сценарий | Ожидаемое поведение | Evidence |
|---|---|---|---|
| `EC-557-01` | Ответ пришёл после закрытия или истечения request | Система не меняет logical state и фиксирует expired/stale rejection | stale-response scenario |
| `EC-557-02` | Пришёл duplicate callback или повторная доставка того же ответа | Logical completion остаётся единственной, duplicate path завершается safe no-op с audit evidence | replay-idempotency scenario |
| `EC-557-03` | Owner ответил в одном канале, а другой канал ещё показывает pending request | Persisted backend truth синхронизирует статус, второй канал не становится расходящимся source of truth | dual-channel consistency review |
| `EC-557-04` | Tool timeout/TTL короче owner wait window | Это считается blocking policy violation, а не допустимым happy-path поведением | timeout-baseline check |
| `EC-557-05` | GitHub comment используется как единственный ответный path | Сценарий классифицируется как degraded fallback и ухудшает fallback metrics | GitHub-fallback scenario |
| `EC-557-06` | Voice/text reply нельзя однозначно сопоставить request | Ответ не должен приводить к continuation без явной deterministic binding semantics | ambiguous-binding review |

## Non-Goals
- Делать Telegram источником platform semantics.
- Возвращать detached resume-run как primary happy-path.
- Проектировать generalized conversation platform в рамках Sprint S17 Day3.
- Фиксировать design/schema/API/runtime topology до `run:arch` и `run:design`.
- Расширять owner-facing wait contract на `run:self-improve`.

## NFR draft для handover в architecture

| ID | Требование | Как измеряем / проверяем |
|---|---|---|
| `NFR-557-01` | `PM-554-01` Delivery-before-wait consistency должна оставаться `100%` | lifecycle audit + acceptance walkthrough |
| `NFR-557-02` | `REL-554-03` Live MCP wait coverage должна оставаться `100%` | runtime config/session evidence по effective timeout/TTL built-in `kodex` path |
| `NFR-557-03` | `REL-554-01` Same-session continuation rate должна целиться в `>= 80%` | run/session audit + continuation classification |
| `NFR-557-04` | `UX-554-01` Owner inbox visibility rate должна целиться в `>= 99%` | inbox projection audit + scenario review |
| `NFR-557-05` | `REL-554-02` Deterministic reply binding correctness должна целиться в `>= 99.5%` | callback/text/voice audit + incident review |
| `NFR-557-06` | `OPS-554-01` Overdue/manual-fallback transparency должна оставаться `100%` | lifecycle visibility review + operator walkthrough |
| `NFR-557-07` | Recovery fallback должен быть классифицирован отдельно от happy-path в `100%` recovery scenarios | run/session evidence + negative acceptance review |
| `NFR-557-08` | `run:self-improve` exclusion correctness должна оставаться `100%` | scope/policy review на architecture/design handover |

## Analytics и product evidence
- События:
  - `owner_feedback_request_created`
  - `owner_feedback_delivery_pending`
  - `owner_feedback_delivery_accepted`
  - `owner_feedback_wait_entered`
  - `owner_feedback_response_received`
  - `owner_feedback_same_session_continued`
  - `owner_feedback_recovery_resume_started`
  - `owner_feedback_recovery_resume_completed`
  - `owner_feedback_status_changed`
  - `owner_feedback_manual_fallback_opened`
  - `owner_feedback_github_fallback_used`
- Метрики:
  - `NSM-554-01` Deterministic owner feedback completion rate
  - `PM-554-01` Delivery-before-wait consistency
  - `PM-554-02` Response turnaround p75
  - `REL-554-01` Same-session continuation rate
  - `UX-554-01` Owner inbox visibility rate
  - `REL-554-02` Deterministic reply binding correctness
  - `REL-554-03` Live MCP wait coverage
  - `OPS-554-01` Overdue/manual-fallback transparency
- Expected evidence:
  - acceptance walkthrough по Telegram path и staff-console fallback;
  - lifecycle trace `delivery pending -> delivery accepted -> waiting -> response -> continuation`;
  - explicit recovery walkthrough для snapshot-resume fallback;
  - scenario matrix по text/voice/callback binding и negative cases по stale/duplicate replies;
  - scope review note, подтверждающий deferred later-wave contours.

## Риски и допущения

| Type | ID | Описание | Митигирующее действие | Статус |
|---|---|---|---|---|
| risk | `RSK-557-01` | Same-session happy-path может быть подменён operational shortcut через shorter timeout или detached resume | Сохранять max timeout/TTL baseline и отдельную классификацию recovery path как blocking constraint | open |
| risk | `RSK-557-02` | Telegram и staff-console могут начать расходиться по lifecycle semantics и context payload | Держать единый persisted backend contract и parity scenario matrix уже на PRD | open |
| risk | `RSK-557-03` | Voice/text/callback binding останется неоднозначным и создаст duplicate logical continuation | Делать deterministic binding blocking requirement до architecture/design | open |
| risk | `RSK-557-04` | Overdue/manual-fallback path останется hidden в operator tooling и сломает доверие owner | Держать transparency как обязательный acceptance signal, а не опциональную ops-заметку | open |
| assumption | `ASM-557-01` | Telegram pending inbox и staff-console fallback достаточно для core MVP без дополнительных каналов | accepted |
| assumption | `ASM-557-02` | Большинство pilot-сценариев можно покрыть same-session continuation, а snapshot-resume останется исключением | accepted |
| assumption | `ASM-557-03` | Owner-facing context можно подать без GitHub-comment detour и без richer conversation UX | accepted |

## Открытые вопросы для `run:arch`
- Как разделить ownership между `control-plane`, `worker`, `agent-runner`, `api-gateway`, `staff web-console` и `telegram-interaction-adapter`, не потеряв product invariants Day1/Day2/Day3?
- Где проходит граница между live wait lifetime, delivery/retry orchestration, persisted request truth и response ingestion/correlation?
- Как обеспечить max timeout/TTL baseline built-in `kodex` MCP wait path и не нормализовать recovery resume как основной UX?
- Какие архитектурные alternatives и ADR нужны, чтобы `run:design` не переоткрывал решения о same-session happy-path, recovery fallback и dual-channel semantics?

## Handover в `run:arch`
- Follow-up issue: `#559`.
- На архитектурном этапе нельзя потерять:
  - unified owner feedback loop как platform capability поверх Sprint S10/S11 baseline;
  - same live pod / same `codex` session как primary happy-path;
  - built-in `kodex` MCP wait path с max timeout/TTL не ниже owner wait window;
  - snapshot-resume только как recovery fallback;
  - lifecycle `created -> delivery pending -> delivery accepted -> waiting -> response -> continuation`;
  - Telegram pending inbox + staff-console fallback поверх одного persisted backend contract;
  - deterministic text/voice/callback binding;
  - overdue / expired / manual-fallback visibility;
  - `run:self-improve` exclusion;
  - deferred scope для additional channels, reminders/escalations, attachments, multi-party routing, richer conversation UX и detached resume-run.
- Архитектурный этап обязан определить:
  - service boundaries и ownership matrix;
  - lifecycle ownership для live wait, delivery, response ingestion и recovery fallback;
  - audit/correlation responsibilities;
  - список ADR/alternatives;
  - issue для `run:design` без trigger-лейбла с continuity-требованием `design -> plan -> dev`.

## Связанные документы
- `docs/delivery/sprints/s17/sprint_s17_unified_user_interaction_waits_and_owner_feedback_inbox.md`
- `docs/delivery/epics/s17/epic_s17.md`
- `docs/delivery/epics/s17/epic-s17-day1-unified-user-interaction-waits-and-owner-feedback-inbox-intake.md`
- `docs/delivery/epics/s17/epic-s17-day2-unified-user-interaction-waits-and-owner-feedback-inbox-vision.md`
- `docs/delivery/traceability/s17_unified_user_interaction_waits_and_owner_feedback_inbox_history.md`
- `docs/delivery/issue_map.md`
- `docs/delivery/delivery_plan.md`
- `docs/product/requirements_machine_driven.md`
- `docs/product/agents_operating_model.md`
- `docs/product/labels_and_trigger_policy.md`
- `docs/product/stage_process_model.md`
- `docs/delivery/sprints/s10/sprint_s10_mcp_user_interactions.md`
- `docs/delivery/sprints/s11/sprint_s11_telegram_user_interaction_adapter.md`
- `docs/architecture/mcp_approval_and_audit_flow.md`
- `docs/architecture/prompt_templates_policy.md`
- `docs/research/src_idea-machine_driven_company_requirements.md`
- `services/internal/control-plane/README.md`
- `services/jobs/agent-runner/README.md`
- `services/jobs/worker/README.md`
- `services/external/api-gateway/README.md`
- `services/external/telegram-interaction-adapter/README.md`
- `services/staff/web-console/README.md`
