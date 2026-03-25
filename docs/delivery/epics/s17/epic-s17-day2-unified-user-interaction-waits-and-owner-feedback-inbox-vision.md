---
doc_id: EPC-CK8S-S17-D2-OWNER-FEEDBACK-WAITS
type: epic
title: "Epic S17 Day 2: Vision для unified long-lived user interaction waits и owner feedback inbox (Issues #554/#557)"
status: in-review
owner_role: PM
created_at: 2026-03-25
updated_at: 2026-03-25
related_issues: [360, 361, 458, 532, 540, 541, 554, 557]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-25-issue-554-vision"
---

# Epic S17 Day 2: Vision для unified long-lived user interaction waits и owner feedback inbox (Issues #554/#557)

## TL;DR
- Для Issue `#554` сформирован vision-package: mission, north star, persona outcomes, KPI/guardrails, wave boundaries и product principles для Sprint S17.
- Unified owner feedback loop зафиксирован как platform capability: owner отвечает в Telegram или staff-console, видит pending request и получает предсказуемое продолжение той же задачи без GitHub-comment detour и channel drift.
- Day1 baseline из Issue `#541` сохранён без reopening: same live pod / same `codex` session остаётся primary happy-path, snapshot-resume используется только как recovery fallback, long human-wait target `>=24h`, delivery-before-wait lifecycle, Telegram pending inbox, staff-console fallback, deterministic text/voice binding и `run:self-improve` exclusion остаются обязательными.
- Для built-in interaction tools внутри existing MCP server `codex_k8s` happy-path трактуется как реальное live ожидание ответа tool в той же session: effective timeout/TTL должен быть максимальным и не короче owner wait window, а resume с подложенным tool result остаётся только recovery/degradation path.
- Создана follow-up issue `#557` для stage `run:prd` без trigger-лейбла; PRD должен формализовать user stories, FR/AC/NFR, scenario matrix и expected evidence без разрыва цепочки `prd -> arch -> design -> plan -> dev`.

## Priority
- `P0`.

## Vision charter

### Mission statement
Сделать unified long-lived owner feedback loop базовой platform capability, чтобы owner / product lead мог получить pending request в Telegram или staff-console, ответить в удобном канале и быть уверенным, что агент действительно ждёт его ответ и продолжит ту же задачу детерминированно, без потери контекста и без channel-specific drift.

### Цели и ожидаемые результаты
1. Превратить текущие built-in interaction tools и существующий Telegram path в единый owner feedback loop с понятным inbox и lifecycle `delivery -> wait -> response -> continuation`.
2. Закрепить same-session continuation как primary happy-path: live pod / live `codex` session и долгоживущий вызов built-in `codex_k8s` interaction tool ждут пользователя, а snapshot-resume остаётся только failure-recovery механизмом.
3. Дать owner и operator явную прозрачность pending / overdue / expired / manual-fallback состояний, чтобы ожидание пользователя не выглядело как “run завис” или “агент пропал”.
4. Сохранить channel-neutral persisted backend contract: Telegram и staff-console работают как разные UX-поверхности одного и того же platform-owned wait/response state.

### Пользователи и стейкхолдеры
- Основные пользователи:
  - owner / product lead, которому нужен понятный pending inbox, удобный канал ответа и уверенность, что агент реально ждёт его решение;
  - execution/runtime path, которому нужен детерминированный переход `delivery accepted -> waiting -> response -> same-session continuation`, а recovery resume используется только при потере live runtime;
  - staff/operator, которому нужна visibility по pending requests, overdue/fallback cases и доступ к staff-console как operational fallback без потери продуктового контекста.
- Стейкхолдеры:
  - `services/internal/control-plane` как будущий owner wait-state semantics, persisted interaction state и lifecycle truth;
  - `services/jobs/worker` как будущий owner delivery/retry/wait timeout и overdue/manual-fallback processing;
  - `services/jobs/agent-runner` как runtime-контур live-session continuity и deterministic resume payload;
  - `services/staff/web-console` как staff-facing inbox/fallback surface;
  - `services/external/telegram-interaction-adapter` как первый внешний канал ответа поверх platform-owned semantics;
  - `services/external/api-gateway` как edge-контур HTTP transport/callback policy.
- Владелец решения: Owner.

### Продуктовые принципы и ограничения
- Sprint S17 наследует channel-neutral interaction baseline Sprint S10 и первый внешний Telegram path Sprint S11, но не переопределяет их product semantics; инициатива объединяет их в long-lived wait/continuation contract.
- Same live pod / same `codex` session остаётся primary happy-path; detached resume-run не возвращается как равноправный default UX без нового owner-решения.
- Persisted session snapshot используется только как recovery fallback при потере live runtime, а не как основной путь продолжения.
- Existing built-in MCP server `codex_k8s` остаётся core точкой для owner-facing interaction tools; на happy-path effective timeout/TTL этого MCP path должен быть максимальным и не ниже long human-wait target, чтобы агент ждал реальный ответ tool, а не заменял ожидание synthetic resume-потоком.
- Delivery acceptance и waiting for user response остаются разными lifecycle состояниями; “сообщение отправлено” не означает “пользователь уже ответил”.
- Telegram inbox и staff-console fallback обязаны работать поверх одного persisted backend contract; канал не может стать owner of semantics.
- Text reply, inline callback и voice reply должны связываться с исходным request детерминированно и не создавать ambiguous continuation path.
- `run:self-improve` остаётся вне owner-facing human-wait contract.
- В рамках `run:vision` разрешены только markdown-изменения.

## Scope boundaries

### MVP scope
- Unified owner feedback loop для response-required run-сценариев всех stage/run-типов, кроме `run:self-improve`.
- Pending inbox semantics:
  - owner видит pending request в Telegram и/или staff-console;
  - ответ из любого разрешённого канала продолжает исходный run без потери product context.
- Wait/continuation contract:
  - lifecycle `created -> delivery pending -> delivery accepted -> waiting -> response -> continuation`;
  - same-session continuation как primary path;
  - snapshot-resume как recovery fallback;
  - max timeout/TTL для built-in `codex_k8s` MCP wait path не ниже owner wait window, чтобы tool-call оставался живым до ответа пользователя;
  - visibility по overdue / expired / manual-fallback состояниям.
- Dual-channel semantics:
  - Telegram как первый внешний response path;
  - staff-console как обязательный fallback и operational visibility surface;
  - единый persisted backend contract для request state, response binding и continuation outcome.
- Deterministic text/voice reply binding и handover в `run:prd` через issue `#557`.

### Post-MVP / deferred scope
- Дополнительные каналы помимо Telegram и staff-console.
- Advanced reminders, escalations и routing policy.
- Attachments, multi-party routing и richer conversation threads.
- Любые attempts сделать detached resume-run равноправным happy-path или заменить same-session model.
- Generalized conversation platform вне bounded owner feedback loop.

### Sequencing and locked baseline
- Active vision stage в Issue `#554` допустим только как продолжение locked intake baseline Issue `#541`; reopening baseline требует нового owner-решения.
- Sprint S10 остаётся effective baseline для built-in interaction semantics, а Sprint S11 остаётся effective baseline для Telegram channel path; Sprint S17 не подменяет их и не возвращает approval semantics в user feedback domain.
- Если следующий stage начинает требовать Telegram-first contract, richer conversation UX или detached resume-run как default flow, stage должен быть остановлен до нового owner-решения.

## Success metrics

### North Star
| ID | Метрика | Определение | Источник | Целевое значение |
|---|---|---|---|---|
| `NSM-554-01` | Deterministic owner feedback completion rate | Доля owner feedback requests, которые проходят путь `delivery accepted -> valid user response -> continuation resumed` через Telegram или staff-console без потери контекста, ручного GitHub fallback и без неопределённости по исходному run | interaction audit events + `flow_events` + run/continuation linkage review | `>= 85%` в pilot-сценариях MVP |

### Supporting metrics
| ID | Метрика | Определение/формула | Источник | Цель |
|---|---|---|---|---|
| `PM-554-01` | Delivery-before-wait consistency | Доля response-required сценариев, где run сначала достигает `delivery accepted`, и только после этого переходит в `waiting for user response` | lifecycle audit + state transition review | `100%` |
| `PM-554-02` | Response turnaround p75 | p75 времени от `delivery accepted` до получения валидного owner response | interaction timestamps + callback audit | `<= 30 минут` |
| `REL-554-01` | Same-session continuation rate | Доля валидных ответов пользователя, которые продолжают исходный run в той же live session без перехода в recovery fallback | run/session audit + continuation classification | `>= 80%` |
| `UX-554-01` | Owner inbox visibility rate | Доля pending requests, доступных owner через Telegram и/или staff-console с достаточным контекстом для ответа без GitHub detour | inbox projection audit + UX sample review | `>= 99%` |
| `REL-554-02` | Deterministic reply binding correctness | Доля text/voice/callback replies, сопоставленных с исходным request без duplicate logical response и без ambiguous continuation | callback audit + incident review | `>= 99.5%` |
| `REL-554-03` | Live MCP wait coverage | Доля response-required сценариев, где effective timeout/TTL built-in `codex_k8s` MCP path не короче owner wait window и happy-path не требует synthetic resume с подложенным tool result | runtime config audit + run/session evidence | `100%` |
| `OPS-554-01` | Overdue/manual-fallback transparency | Доля overdue / expired / manual-fallback сценариев, где owner/operator видят явный статус и причину, а не “тихий зависший run” | lifecycle audit + operator review | `100%` |

### Guardrails (ранние сигналы)
- `GR-554-01`: если `PM-554-01 < 100%`, следующий stage обязан приоритизировать delivery/wait-state separation выше любого расширения каналов или UX.
- `GR-554-02`: если `REL-554-01 < 65%`, `run:prd` и `run:arch` не могут трактовать same-session path как operationally healthy baseline без отдельного owner-решения.
- `GR-554-03`: если `UX-554-01 < 95%`, дальнейшее расширение Telegram/staff UX блокируется до устранения inbox visibility gaps.
- `GR-554-04`: если `REL-554-02 < 99%`, следующие stage обязаны приоритизировать response binding/correlation semantics выше richer conversation affordances.
- `GR-554-05`: если `OPS-554-01 < 100%`, overdue / manual-fallback path нельзя оставлять только в operator logs; visibility обязана стать blocking requirement.
- `GR-554-06`: если core AC начинают требовать дополнительные каналы, advanced reminders, attachments, multi-party routing или detached resume-run как равноправный happy-path, stage переводится в `need:input` до нового owner-решения.
- `GR-554-07`: если `REL-554-03 < 100%` или effective timeout/TTL built-in `codex_k8s` MCP path остаётся короче long human-wait budget, следующий stage обязан сначала закрыть timeout/TTL baseline, а не нормализовать detached resume как happy-path.

## Risks and Product Assumptions
| Тип | ID | Описание | Митигирующее действие | Статус |
|---|---|---|---|---|
| risk | `RSK-554-01` | Scope может расползтись в Telegram-first или generalized conversation initiative вместо deterministic owner feedback loop | Держать dual-channel contract и later-wave scope явно отделёнными уже на vision/PRD | open |
| risk | `RSK-554-02` | 24h same-session expectation может привести к скрытым runtime/cost компромиссам и неявному возврату detached resume как default path | Сохранить same-session как product baseline, а recovery fallback как отдельный деградационный path с прозрачной диагностикой | open |
| risk | `RSK-554-03` | Staff-console fallback может стать вторичным каналом без parity visibility и потерять operational usefulness | Зафиксировать inbox visibility и overdue/manual-fallback transparency как guardrails | open |
| risk | `RSK-554-04` | Voice/text replies могут давать ambiguous binding и нарушать deterministic continuation | Тащить deterministic reply binding как blocking contract в PRD/architecture/design | open |
| risk | `RSK-554-05` | Короткий timeout/TTL built-in `codex_k8s` MCP path может разрушить same-session happy-path и вынудить platform симулировать continuation через resume с подложенным tool result | Зафиксировать max MCP timeout/TTL как blocking baseline уже на `run:prd` и не принимать recovery resume как happy-path substitute | open |
| assumption | `ASM-554-01` | Existing built-in interaction tools и Telegram channel можно объединить в единый owner feedback contract без reopening core interaction semantics Sprint S10 | Проверить FR/AC/NFR и lifecycle contract на `run:prd` | accepted |
| assumption | `ASM-554-02` | Основная пользовательская ценность достигается через предсказуемое ожидание и continuation, а не через richer conversation UX | Подтвердить user stories и success metrics на `run:prd` | accepted |
| assumption | `ASM-554-03` | Live-session happy-path сможет покрыть большинство pilot-сценариев, а snapshot-resume останется исключительным recovery path | Проверить expected evidence и NFR baseline на `run:prd` / `run:arch` | accepted |

## Readiness criteria для `run:prd`
- [x] Mission, north star и persona outcomes сформулированы для owner/product lead, same-session runtime path и staff/operator fallback path.
- [x] KPI/success metrics и guardrails определены как измеримые product/operational сигналы.
- [x] Locked Day1 baseline сохранён явно: same live pod / same `codex` session как primary happy-path, snapshot-resume только как recovery fallback, long human-wait target `>=24h`, delivery-before-wait lifecycle, Telegram pending inbox, staff-console fallback, deterministic text/voice binding и `run:self-improve` exclusion.
- [x] Happy-path явно требует max timeout/TTL для built-in `codex_k8s` MCP wait path не ниже owner wait window; synthetic resume с подложенным tool result не нормализован как основной UX.
- [x] Core MVP и later-wave scope разделены явно; дополнительные каналы, reminders/escalations, attachments, multi-party routing, richer conversation UX и detached resume-run как равноправный happy-path не входят в blocking scope.
- [x] Создана отдельная issue следующего этапа `run:prd` (`#557`) без trigger-лейбла.

## Acceptance criteria (Issue #554)
- [x] Mission, north star и продуктовые принципы для unified long-lived owner feedback loop сформулированы явно.
- [x] KPI/success metrics и guardrails зафиксированы для delivery-before-wait consistency, response turnaround, same-session continuation, inbox visibility, deterministic reply binding и overdue/manual-fallback transparency.
- [x] Персоны, MVP/Post-MVP границы, риски и assumptions описаны без reopening Day1 baseline и без перехода в implementation details.
- [x] Явно сохранены как обязательный baseline: same-session happy-path, snapshot-resume recovery-only fallback, `>=24h` wait target, delivery-before-wait lifecycle, Telegram pending inbox + staff-console fallback, deterministic text/voice binding и `run:self-improve` exclusion.
- [x] Зафиксировано, что для built-in `codex_k8s` MCP path happy-path требует максимальный timeout/TTL не ниже owner wait window; resume с подложенным tool result остаётся только recovery/degradation path.
- [x] Подготовлен handover в `run:prd` и создана follow-up issue `#557` без trigger-лейбла.

## Handover в следующий этап
- Следующий stage: `run:prd`.
- Follow-up issue: `#557`.
- Trigger-лейбл `run:prd` на issue `#557` ставит Owner.
- Обязательное continuity-требование для `#557`:
  - в конце PRD stage агент обязан создать issue для `run:arch` без trigger-лейбла;
  - в body этой issue нужно явно повторить требование продолжить цепочку `arch -> design -> plan -> dev` без разрывов.
- На `run:prd` нельзя потерять следующие решения vision:
  - unified owner feedback loop остаётся platform capability поверх channel-neutral interaction baseline Sprint S10 и Telegram path Sprint S11;
  - same live pod / same `codex` session остаётся primary happy-path;
  - built-in `codex_k8s` MCP wait path обязан использовать максимальный timeout/TTL не ниже long human-wait target, чтобы happy-path был реальным live wait, а не hidden resume substitute;
  - snapshot-resume остаётся recovery-only fallback;
  - long human-wait target `>=24h` и lifecycle `created -> delivery pending -> delivery accepted -> waiting -> response -> continuation` обязательны;
  - Telegram pending inbox и staff-console fallback остаются dual-channel поверх одного persisted backend contract;
  - text/voice reply binding остаётся deterministic;
  - overdue / expired / manual-fallback scenarios должны быть product-visible, а не hidden operator-only path;
  - `run:self-improve` остаётся вне human-wait contract;
  - дополнительные каналы, reminders/escalations, attachments, multi-party routing, richer conversation UX и detached resume-run как равноправный happy-path остаются later-wave scope.

## Связанные документы
- `docs/delivery/sprints/s17/sprint_s17_unified_user_interaction_waits_and_owner_feedback_inbox.md`
- `docs/delivery/epics/s17/epic_s17.md`
- `docs/delivery/epics/s17/epic-s17-day1-unified-user-interaction-waits-and-owner-feedback-inbox-intake.md`
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
