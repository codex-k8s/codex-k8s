---
doc_id: EPC-CK8S-S17-D1-INTERACTION-WAITS
type: epic
title: "Epic S17 Day 1: Intake для unified long-lived user interaction waits и owner feedback inbox (Issue #541)"
status: in-review
owner_role: PM
created_at: 2026-03-20
updated_at: 2026-03-20
related_issues: [360, 361, 458, 473, 532, 540, 541, 554]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-20-issue-541-intake"
---

# Epic S17 Day 1: Intake для unified long-lived user interaction waits и owner feedback inbox (Issue #541)

## TL;DR
- Платформа уже умеет вызывать `user.notify` и `user.decision.request`, а Sprint S11 уже довёл Telegram channel path до production-like baseline, но owner feedback loop всё ещё не даёт детерминированного контракта `пользователь ответил -> та же работа продолжилась`.
- Intake фиксирует Sprint S17 как отдельную cross-cutting инициативу: нужен единый long-lived human-wait contract для всех run-типов, кроме `run:self-improve`, с прозрачным lifecycle `delivery -> wait -> response -> continuation`.
- Day1-решение: primary happy-path = та же живая pod / та же `codex` session до ответа пользователя; persisted session snapshot остаётся только recovery fallback на случай потери live runtime.
- Зафиксированы обязательные baselines: long human wait не меньше 24 часов, явное разделение `delivery pending` и `waiting for user response`, Telegram pending inbox, staff-console fallback, persisted text/voice binding, deterministic continuation semantics и continuity issue `#554` для stage `run:vision`.

## Контекст
- Sprint S10 уже сформировал platform-side contract built-in user interactions:
  - `user.notify` как non-blocking notification path;
  - `user.decision.request` как typed wait-state interaction;
  - coarse runtime wait-state `waiting_mcp` и persisted resume path.
- Sprint S11 уже довёл Telegram channel path до implementation baseline:
  - Telegram adapter принимает raw webhook и normalizes text/voice replies;
  - platform semantics и correlation остаются в core platform;
  - появился owner-facing callback path, но не появился единый продуктовый контур long-lived wait/resume.
- Production/debug signals зафиксированы в issues `#532`, `#540` и в source issue `#541`:
  - interaction может быть создан успешно, но delivery в Telegram приходит позже, чем run формально уходит в wait-state;
  - inline button сохраняет callback/response, но живая discussion session не продолжает ту же работу;
  - вместо same-session continuation платформа поднимает отдельный `interaction-resume` run, который сам по себе может уйти в polling;
  - text/voice replies не всегда надёжно привязаны к конкретному pending interaction request;
  - в `services/jobs/agent-runner/internal/runner/templates/codex_config.toml.tmpl` уже сейчас зафиксирован `tool_timeout_sec = 180`, что конфликтует с owner-driven ожиданием ответа длительностью до суток.
- Текущий baseline документации уже содержит важные, но недостаточные части решения:
  - `docs/product/stage_process_model.md` фиксирует `waiting_mcp` и paused semantics;
  - `docs/architecture/mcp_approval_and_audit_flow.md` отделяет `user.notify` и `user.decision.request` от approval flow;
  - Sprint S10/S11 фиксируют typed interaction contract и Telegram-specific adapter contract.
- Проблема теперь не локальная transport-bug задача, а cross-cutting продуктовый gap между runtime execution model, channel UX и owner visibility.

## Рекомендованный launch profile
- Базовый launch profile: `new-service`.
- Причины:
  - инициатива меняет platform-wide execution contract для всех run-типов, кроме `run:self-improve`;
  - затрагиваются `control-plane`, `worker`, `agent-runner`, `api-gateway`, `telegram-interaction-adapter`, `web-console` и delivery observability;
  - требуется явно выбрать между конкурирующими execution models и зафиксировать cost/recovery trade-offs до PRD/architecture;
  - нужны обязательные `vision`, `arch` и `design`, иначе same-session continuity, long wait lifetime и fallback channel path снова разъедутся.
- Целевая stage-цепочка:
  `run:intake -> run:vision -> run:prd -> run:arch -> run:design -> run:plan`.

## Problem Statement
### As-Is
- Built-in feedback tools существуют, но не образуют единый deterministic contract для run-типов `discussion`, doc-stages, `run:dev`, `run:qa`, `run:release`, `run:ops` и review-driven flows.
- Платформа не гарантирует, что ответ пользователя по кнопке, тексту или голосу продолжит исходную живую session; detached resume-run фактически выступает как основной path, а не как recovery-only fallback.
- Delivery lifecycle не разделён достаточно явно:
  - `request created`;
  - `delivery pending`;
  - `delivery accepted`;
  - `waiting for user response`;
  - `response received`;
  - `continuation resumed`.
- Owner не получает first-class inbox contract:
  - Telegram не даёт устойчивый список pending requests как source of truth;
  - staff-console не описан как обязательный fallback, если Telegram path недоступен;
  - text/voice replies не имеют достаточно явного persisted binding к ожидающему request.
- Long human wait не оформлен как platform contract:
  - runtime, pod lifetime и cleanup semantics не зафиксированы как один продуктовый baseline;
  - `tool_timeout_sec = 180` противоречит ожиданию owner-driven response window до 24 часов.

### To-Be
- Платформа описывает owner feedback loop как first-class capability:
  - агент создаёт interaction request;
  - delivery сначала подтверждается как `accepted` или переводится в понятный degraded/fallback path;
  - только после этого run становится в `waiting for user response`;
  - пользователь отвечает через Telegram или staff-console;
  - исходная живая session продолжает работу либо, если live runtime утрачен, запускается canonical recovery continuation.
- Primary execution model = same live pod / same `codex` session until user response; persisted session snapshot нужен только для crash/eviction/redeploy recovery.
- Telegram pending inbox и staff-console fallback работают поверх одного persisted backend contract с общим idempotent interaction lifecycle.
- Text и voice replies имеют явный pending-input binding, не теряют контекст и не создают дублирующих ответов.

## Brief
- **Проблема:** current user interaction path умеет доставлять запросы и принимать callback'и, но не даёт единый и понятный owner-facing contract long wait + same-session continuation.
- **Для кого:** для owner / product lead, который отвечает агенту; для runtime/agent path, который должен ждать и продолжать ту же задачу; для operator/staff UX, который должен видеть pending, overdue и fallback states.
- **Предлагаемое решение:** выделить Sprint S17 как отдельный cross-cutting stream и зафиксировать product baseline по long-lived waits, continuation semantics и owner inbox до начала новой implementation wave.
- **Почему сейчас:** gap уже проявился в production-like flows, и дальнейшее расширение `discussion`, doc-stages, late delivery и review-driven paths небезопасно без единого human-wait contract.
- **Что считаем успехом:** stage-пакет зафиксировал execution model, long-wait baseline, inbox/fallback scope, lifecycle transparency и handover в `run:vision` без потери ключевых решений.
- **Что не делаем на этой стадии:** не фиксируем schema/API/runtime implementation detail, не выбираем rollout topology и не запускаем code/runtime changes вне markdown-only scope.

## Candidate execution models

| Вариант | Краткое описание | Плюсы | Риски | Intake-решение |
|---|---|---|---|---|
| A. Always-live pod/session until response | Run и pod остаются живыми всё время ожидания, а ответ пользователя всегда продолжает ту же session | Максимально простой пользовательский mental model; нет detach-gap на happy-path | Самый дорогой по ресурсам; плохо переживает eviction/redeploy без отдельного recovery contract | Не принят как единственная модель |
| B. Hybrid live-session happy-path + snapshot resume fallback | Live pod/session остаётся primary path; session snapshot используется только при потере live runtime | Сохраняет same-session UX и даёт recovery-path без product drift | Требует явного lifecycle contract для pause/resume/cleanup и канонического recovery continuation | Рекомендованный baseline Sprint S17 |
| C. Detached resume-run as primary model | Исходный run ждёт отдельно, а ответ пользователя всегда продолжает новый resume-run | Ниже ресурсная стоимость live pods, проще держать жёсткие pod TTL | Размывает continuity, создаёт ambiguity в UX, усложняет traceability и уже показывает слабый happy-path в production evidence | Явно не принимается как default UX |

## MVP Scope
### In scope
- Unified human-wait contract для всех run-типов, кроме `run:self-improve`.
- Сравнение execution models и фиксация recommended path `hybrid same live session + recovery fallback`.
- Long human-wait policy не меньше 24 часов:
  - interaction TTL;
  - run/session wait semantics;
  - pod/job lifetime expectations;
  - cleanup / reclaim / cancel policy;
  - policy для `tool_timeout_sec` или equivalent split timeout model.
- Delivery-before-wait consistency и явный lifecycle статусов `created -> delivery pending -> delivery accepted -> waiting -> response -> continuation`.
- Owner-facing inbox contract:
  - Telegram pending requests;
  - staff-console fallback при отсутствии/деградации Telegram path;
  - возможный dual-visibility mode без двойного принятия ответа.
- Persisted binding для `Ответить текстом` и `Ответить голосом`, включая retry-safe/idempotent response handling.
- Deterministic continuation semantics после inline button, text reply и voice reply.
- Handover в `run:vision` через continuity issue `#554`.

### Out of scope для core wave
- Кодовая реализация, schema migrations, deploy manifests и runtime rollouts до завершения `run:plan`.
- Редизайн approval flow под видом user interaction scope.
- Дополнительные каналы помимо Telegram и staff-console fallback.
- Advanced reminders, attachments, multi-party routing, richer conversation threads и generalized conversation platform.
- Любая попытка вернуть detached resume-run как основной happy-path без нового owner-решения.

## Constraints
- Sprint S17 обязан сохранять решения Sprint S10/S11 как baseline, а не проектироваться заново:
  - built-in tools остаются внутри existing `codex_k8s` MCP server;
  - interaction-domain остаётся отдельным от approval flow;
  - Telegram transport сохраняется channel-specific contour, а platform semantics остаются channel-neutral.
- `run:self-improve` остаётся явным исключением и не обязан поддерживать owner-facing human wait contract.
- Staff-console fallback и Telegram pending inbox должны разделять один persisted backend contract; канал не может становиться owner of semantics.
- Delivery lifecycle, wait-state semantics и continuation contract должны оставаться auditable и idempotent.
- Intake stage остаётся markdown-only и не закрепляет premature implementation details раньше `run:prd` / `run:arch`.

## Product principles
- Same-task continuity важнее дешёвого detached resume-path.
- Delivery acceptance должна предшествовать formal wait-state, иначе owner UX и audit расходятся.
- Channel-specific UX допустим только поверх одного platform-owned interaction contract.
- Long human wait является first-class product state, а не побочным эффектом transport timeout.
- Fallback через staff-console обязателен там, где Telegram channel unavailable или degraded.

## Candidate product waves

| Wave | Фокус | Exit signal |
|---|---|---|
| Wave 1 | Unified human-wait contract, hybrid same-session continuity, 24h wait baseline, Telegram pending inbox, staff-console fallback, text/voice binding, lifecycle visibility | Owner может получить запрос, ответить в Telegram или staff-console и увидеть deterministic continuation исходного run либо его canonical recovery continuation |
| Wave 2 | Overdue/expiry/manual fallback hardening, operator visibility, replay safety, cost-control guardrails и degraded channel handling | Pending/expired/manual-fallback flows становятся прозрачными и безопасными без потери same-session priority |
| Wave 3 | Дополнительные каналы, richer conversation UX, reminders/escalations, attachments и multi-party routing | Новые каналы/модальности добавляются без переопределения core wait/resume semantics |

## Acceptance Criteria (Intake stage)
- [x] Проблема зафиксирована как отдельный cross-cutting gap в human-wait contract, а не как локальный Telegram bug или timeout tweak.
- [x] Сравнены минимум три execution model options, и `hybrid live-session happy-path + snapshot recovery fallback` зафиксирован как recommended baseline.
- [x] Явно зафиксированы обязательные baselines: same live pod/codex session как primary happy-path, long human-wait target `>= 24h`, delivery-before-wait lifecycle и self-improve exclusion.
- [x] Явно зафиксирован owner-facing inbox scope: Telegram pending inbox + staff-console fallback поверх одного persisted backend contract.
- [x] Persisted text/voice binding и deterministic continuation после inline/text/voice reply включены в core Wave 1.
- [x] Подготовлена continuity issue `#554` для stage `run:vision` без trigger-лейбла.

## Декомпозиция по этапам (до plan)

| Stage | Фокус | Выходной артефакт |
|---|---|---|
| Vision | Mission, north star, persona outcomes, KPI/guardrails и wave boundaries для owner feedback loop | charter + success metrics + risk frame |
| PRD | User stories, FR/AC/NFR, edge cases, expected evidence и locked baselines для wait/resume/inbox contract | PRD + user stories + NFR |
| Arch | Ownership split, execution model, live-session vs recovery semantics, channel-neutral domain и lifetime policy | architecture package + ADR/alternatives |
| Design | API/data/run-state/UI/adapter contracts, rollout/rollback notes, observability lifecycle | design package + API/data model |
| Plan | Execution waves, quality-gates, DoR/DoD, implementation issues и owner-managed sequencing | execution package + linked issues |

## Risks and Product Assumptions
- Риск: resource cost long-lived live sessions окажется выше ожидаемого; поэтому hybrid model обязателен, а always-live нельзя принимать без recovery fallback.
- Риск: detached resume-run останется "удобным shortcut" для implementation и снова размоет same-session UX.
- Риск: Telegram-first UX начнёт диктовать core semantics и вытеснит staff-console fallback из обязательного scope.
- Риск: text/voice retries будут создавать duplicate responses, если binding и idempotency не оформить как first-class contract.
- Допущение: текущий platform baseline уже умеет хранить coarse session snapshot и interaction evidence, поэтому Day1 может зафиксировать hybrid approach без проектирования новой платформы с нуля.
- Допущение: owner value выше всего там, где ответ пользователя продолжает исходную задачу, а не создаёт новый малообъяснимый polling run.

## Stage Handover Instructions
- Следующий этап: `run:vision`.
- Созданная issue следующего этапа: `#554`.
- На stage `run:vision` обязательно сохранить и не размыть следующие intake-решения:
  - primary happy-path = same live pod / same `codex` session until user response;
  - persisted session snapshot используется только как recovery fallback;
  - long human-wait target не меньше 24 часов;
  - `delivery pending` и `waiting for user response` остаются разными lifecycle phases;
  - Telegram pending inbox и staff-console fallback обязательны как один owner-facing contour;
  - text/voice binding и deterministic continuation после inline/text/voice reply входят в core Wave 1;
  - `run:self-improve` остаётся исключением из human-wait contract;
  - detached resume-run не возвращается как основной UX без нового owner-решения.
- После завершения vision stage должна быть создана новая issue для `run:prd` без trigger-лейбла.
