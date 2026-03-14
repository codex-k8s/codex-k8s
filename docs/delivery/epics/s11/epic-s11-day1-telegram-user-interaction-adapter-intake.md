---
doc_id: EPC-CK8S-S11-D1-TELEGRAM-ADAPTER
type: epic
title: "Epic S11 Day 1: Intake для Telegram-адаптера взаимодействия с пользователем (Issue #361)"
status: in-review
owner_role: PM
created_at: 2026-03-14
updated_at: 2026-03-14
related_issues: [361, 444]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-14-issue-361-intake"
---

# Epic S11 Day 1: Intake для Telegram-адаптера взаимодействия с пользователем (Issue #361)

## TL;DR
- После platform-side intake в Issue `#360` Telegram выделяется в отдельный Sprint S11 как первый реальный внешний канал взаимодействия с пользователем.
- Intake фиксирует, что Telegram не должен стартовать параллельно core interaction stream: сначала стабилизируется platform contract, затем поверх него строится channel-specific adapter path.
- MVP Telegram-потока ограничен сценариями `user.notify`, `user.decision.request`, inline callbacks и optional free-text reply, а voice/STT, reminders и richer conversation flows выносятся из core wave.
- Через Context7 по `/mymmrac/telego` и `go list -m -json github.com/mymmrac/telego@latest` подтверждено, что `github.com/mymmrac/telego v1.7.0` покрывает webhook mode, inline keyboards и callback query handling; библиотека зафиксирована в каталоге зависимостей как planned baseline для следующей стадии, но не product contract.
- Continuity issue `#444` подготовлена для stage `run:vision` с явным S10 readiness gate.

## Контекст
- Issue `#334` зафиксировала двухшаговую последовательность: сначала platform-side interaction-domain в core платформы, затем отдельный channel integrator для Telegram.
- Issue `#360` уже открыла Sprint S10 и закрепила core guardrails:
  - built-in `user.notify` и `user.decision.request` как platform baseline;
  - channel-neutral interaction-domain;
  - separation from approval flow;
  - Telegram как отдельный последовательный follow-up stream.
- Reference repositories `telegram-approver` и `telegram-executor` полезны как UX/stack ориентир, но не могут считаться source of truth для границ `codex-k8s`.
- Проверка через Context7 на `2026-03-14` подтвердила, что `telego` поддерживает webhook ingestion, inline buttons и callback queries, поэтому библиотека подходит как pragmatic reference baseline для дальнейшей проработки.
- На `2026-03-14` prerequisite для следующего stage уже проверяем: Issue `#387` и Issue `#389` закрыты, а значит S10 design/plan package уже зафиксировал typed interaction contract для Telegram follow-up stream.

## Рекомендованный launch profile
- Базовый launch profile: `new-service`.
- Обязательная эскалация до полного doc-stage контура:
  - `vision` обязателен, потому что Telegram становится первым channel-specific user-facing experience с отдельными UX/KPI/operability guardrails;
  - `arch` обязателен, потому что scope затрагивает service boundaries, callback security/correlation, операционные ограничения и, вероятно, отдельный adapter contour.
- Целевая stage-цепочка:
  `run:intake -> run:vision -> run:prd -> run:arch -> run:design -> run:plan`.

## Problem Statement
### As-Is
- У платформы пока нет отдельного формализованного Telegram stream поверх interaction contract: есть только общая постановка в Issue `#361` и reference repositories вне текущего монорепозитория.
- Если Telegram стартует раньше стабилизации platform-core contracts Sprint S10, adapter почти неизбежно привяжется к временным callback/wait assumptions.
- Сейчас не зафиксированы продуктовые границы первого канала:
  - какой именно UX считается достаточным для MVP;
  - какие callback/free-text сценарии входят в базовый scope;
  - какие риски по security, correlation и operability должны стать обязательными guardrails.

### To-Be
- Telegram описан как первый channel-specific adapter stream поверх platform-owned interaction contract, а не как самостоятельный источник core semantics.
- MVP Telegram-потока покрывает базовую доставку, inline decision flow, callback handling и optional free-text reply без расширения в сложный conversational product.
- Следующие stage'ы получают ясную продуктовую рамку по UX, зависимостям, ограничениям и handover decisions, не смешивая core interaction-domain и Telegram-specific affordances.

## Brief
- **Проблема:** после запуска platform interaction-domain пользователю всё ещё нужен реальный канал доставки и ответа; без отдельного Telegram stream core capability останется без проверяемого adapter path.
- **Для кого:** для конечного пользователя, который отвечает агенту вне GitHub; для owner/product lead, который ждёт предсказуемый канал принятия решения; для platform operator, который должен поддерживать первый channel adapter без потери audit/correlation discipline.
- **Предлагаемое решение:** выделить Telegram в отдельный Sprint S11 и пройти полный doc-stage контур до implementation-ready execution package.
- **Почему сейчас:** sequencing уже подтверждён Owner, а reference stack понятен; лучше зафиксировать продуктовые границы до PRD/architecture, чем потом вычищать Telegram-first предположения из core contracts.
- **Что считаем успехом:** intake-пакет фиксирует Telegram как отдельный последовательный stream, описывает MVP scope и создаёт handover в `run:vision` без потери platform prerequisites.
- **Что не делаем на этой стадии:** не фиксируем schema/API/runtime topology, не обещаем voice/STT и не копируем reference repositories как готовую реализацию.

## MVP Scope
### In scope
- Доставка `user.notify` в Telegram как первого user-facing notification path.
- Доставка `user.decision.request` с 2-5 inline options.
- Приём callback-ответов по кнопкам.
- Optional free-text reply как fallback/дополнение к кнопкам.
- Базовая webhook/callback security, idempotency, correlation и operability рамка для Telegram adapter.
- Handover в `run:vision` через continuity issue `#444`.

### Out of scope для core wave
- Voice input и STT.
- Rich conversation threads, reminders и расширенные conversational UX flows.
- Multi-chat routing policy, multi-user assignment и дополнительные каналы помимо Telegram.
- Любая попытка переопределить platform interaction-domain через channel-specific требования Telegram.
- Детализированные schema, migrations, HTTP/gRPC contracts и rollout/rollback решения до следующих stage'ов.

## Constraints
- Telegram stream может стартовать только как follow-up после platform-core stream Sprint S10 и должен сохранять зависимость от typed interaction contract.
- Проверяемый gate для `#444`: Issue `#389` остаётся закрытой и продолжает ссылаться на design package Issue `#387` как на effective S10 baseline по typed interaction contract.
- Approval flow и user interaction flow остаются раздельными доменами даже в Telegram UX.
- Контракты должны оставаться typed и adapter-friendly; Telegram-specific transport не должен становиться source of truth для platform semantics.
- Dispatch, retries, idempotency, audit и correlation остаются responsibility platform domain; Telegram adapter потребляет и материализует эти semantics.
- Intake stage остаётся markdown-only и не фиксирует implementation details раньше `run:prd` / `run:arch`.

## Product principles
- Первый внешний канал важен как доказательство жизнеспособности platform interaction contract, а не как Telegram-only feature ради самой интеграции.
- Channel-specific UX может улучшать delivery experience, но не должен ломать platform-owned audit, wait-state и correlation semantics.
- Inline callbacks и free-text должны решать базовую потребность пользователя быстро ответить агенту без ухода в сложный conversational продукт.
- Sequencing важнее скорости: лучше отложить старт Telegram, чем закрепить временный core contract.
- Reference stack полезен как ускоритель, но продуктовый контракт должен рождаться из stage-проработки внутри `codex-k8s`.

## Candidate product waves

| Wave | Фокус | Exit signal |
|---|---|---|
| Wave 1 | Core Telegram MVP: notify, decision request, inline callbacks, optional free-text, webhook/callback baseline | Пользователь может получать и подтверждать основные agent prompts в Telegram без GitHub fallback по умолчанию |
| Wave 2 | Operability and hardening: delivery observability, callback safety, routing policy clarifications, fallback UX | Adapter path становится platform-safe и пригодным для регулярной эксплуатации |
| Wave 3 | Deferred expansion: voice/STT, reminders, richer threads, extra channels | Дополнительные каналы/модальности строятся поверх уже подтверждённого Telegram/core baseline |

## Acceptance Criteria (Intake stage)
- [x] Проблема зафиксирована как отсутствие отдельного channel-specific adapter stream, а не как локальный transport task.
- [x] Подтверждено, что Telegram идёт после Sprint S10 core stream и не смешивается с ним.
- [x] Явно определён MVP baseline: notify, decision request, inline callbacks, optional free-text, базовая webhook/callback рамка.
- [x] Зафиксированы обязательные границы: typed platform contract, separation from approval flow, deferred scope для voice/STT и richer conversations.
- [x] Reference repositories и `telego` отмечены как baseline, но без требования копировать реализацию 1-в-1; planned dependency baseline каталогизирован отдельно от product contract.
- [x] Подготовлена continuity issue `#444` для stage `run:vision`.

## Декомпозиция по этапам (до plan)

| Stage | Фокус | Выходной артефакт |
|---|---|---|
| Vision | Mission, persona outcomes, KPI/guardrails и MVP/Post-MVP рамка для Telegram | charter + success metrics + risk frame |
| PRD | User stories, FR/AC/NFR, channel-specific edge cases и expected evidence | PRD + user stories + NFR |
| Arch | Service boundaries, adapter ownership, callback security/correlation lifecycle | architecture package + ADR/alternatives |
| Design | API/data/webhook/runtime contracts, rollout/rollback notes и operability model | design package + API/data model |
| Plan | Execution waves, quality-gates, implementation issues и split core vs deferred scope | execution package + linked issues |

## Risks and Product Assumptions
- Риск: Telegram-first решения начнут диктовать platform-core semantics и приведут к лишнему coupling.
- Риск: scope расползётся в voice/STT и advanced conversation flows раньше фиксации базовой ценности канала.
- Риск: прямое копирование reference repositories принесёт в `codex-k8s` чужие service boundaries и governance assumptions.
- Допущение: closed-plan baseline Sprint S10 (`#389`) продолжит ссылаться на design package `#387` и не будет переоткрыт/суперседирован без явной новой continuity-точки для Telegram stream.
- Допущение: webhook + inline callbacks + optional free-text достаточно, чтобы доказать ценность первого внешнего канала без richer conversational scope.

## Stage Handover Instructions
- Следующий этап: `run:vision`.
- Созданная issue следующего этапа: `#444`.
- Проверяемый prerequisite для `#444`: Issue `#389` закрыта и остаётся актуальным S10 handover в `run:dev`, а design package Issue `#387` остаётся source-of-truth для typed interaction contract.
- На `2026-03-14` prerequisite выполнен: `#387` closed, `#389` closed.
- На stage `run:vision` обязательно сохранить и не размыть следующие решения intake:
  - Telegram остаётся зависимым stream после Sprint S10 core contracts;
  - MVP ограничен notify/decision/callback/free-text path;
  - approval flow и user interaction flow не смешиваются;
  - `telego` используется только как pragmatic reference baseline;
  - voice/STT, reminders и richer conversations остаются deferred scope.
- После завершения vision stage должна быть создана новая issue для `run:prd` без trigger-лейбла.
