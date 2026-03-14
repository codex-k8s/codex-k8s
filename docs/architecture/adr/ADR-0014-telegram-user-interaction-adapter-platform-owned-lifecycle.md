---
doc_id: ADR-0014
type: adr
title: "Telegram user interaction adapter: platform-owned lifecycle with external adapter contour"
status: proposed
owner_role: SA
created_at: 2026-03-14
updated_at: 2026-03-14
related_issues: [361, 444, 447, 448, 452, 454]
related_prs: []
supersedes: []
superseded_by: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-14-issue-452-arch"
---

# ADR-0014: Telegram user interaction adapter — platform-owned lifecycle with external adapter contour

## TL;DR
- Контекст: Sprint S11 требует первый внешний Telegram channel path поверх typed interaction contract Sprint S10 без Telegram-first drift и без смешения с approval flow.
- Решение: выбираем platform-owned lifecycle, где `control-plane` владеет semantics и correlation, `worker` исполняет delivery/retry/edit/follow-up side effects, `api-gateway` остаётся thin bridge для normalized callbacks, а raw Telegram transport/webhooks живут во внешнем adapter contour.
- Последствия: сохраняются channel-neutral semantics, thin-edge и future multi-channel path, но design-stage обязан конкретизировать callback handle/token model, data model и rollout notes.

## Контекст
- Проблема:
  - если Telegram adapter contour или `api-gateway` станут владельцами semantic callback outcome, платформа потеряет единый source-of-truth для correlation, replay safety и wait-state lifecycle;
  - если raw Telegram webhook сразу терминируется внутри core platform transport, Telegram-specific constraints начнут диктовать форму core contracts;
  - если выделить новый внутренний Telegram service уже сейчас, delivery contour получит premature DB owner и лишний consistency boundary до фиксации design contracts.
- Ограничения:
  - `run:arch` остаётся markdown-only;
  - Sprint S10 typed interaction contract обязателен как platform-owned baseline;
  - approval flow и interaction flow нельзя смешивать;
  - `api-gateway` должен остаться thin-edge.
- Связанные требования:
  - PRD `docs/delivery/epics/s11/prd-s11-day3-telegram-user-interaction-adapter.md`;
  - `FR-003`, `FR-025`, `FR-039`;
  - `NFR-010`, `NFR-016`, `NFR-018`.
- Что ломается без решения:
  - размывается ownership callback/webhook lifecycle;
  - становится недоказуемой channel-neutral semantics Sprint S10;
  - `run:design` переоткрывает архитектурные компромиссы вместо детализации contracts/data.

## Decision Drivers (что важно)
- Channel-neutral platform semantics поверх первого внешнего Telegram channel.
- Thin-edge boundary для `api-gateway`.
- Platform-owned replay safety, correlation и operator visibility.
- Возможность future multi-channel expansion без Telegram-first domain model.
- Отсутствие premature service split и лишнего rollout contour.

## Рассмотренные варианты
### Вариант A: Telegram-first adapter/gateway path с semantic state на transport boundary
- Плюсы:
  - короткий initial path до working webhook flow;
  - меньше platform contracts на старте.
- Минусы:
  - semantic classification переезжает в adapter/gateway layer;
  - raw Telegram payload начинает формировать core model.
- Риски:
  - thin-edge нарушается;
  - duplicate/replay/expired handling становится adapter-local detail;
  - approval-like Telegram flows начинают смешиваться с user interactions.
- Стоимость внедрения:
  - низкая на старте, высокая при исправлении semantic drift.
- Эксплуатация:
  - operator visibility зависит от adapter-local logs и provider refs.

### Вариант B: Новый внутренний Telegram-specific service уже на Day4
- Плюсы:
  - сильная изоляция channel-specific bounded context;
  - отдельный scale path для Telegram delivery.
- Минусы:
  - новый DB owner и новый rollout contour до design-stage;
  - выше coordination cost между `control-plane`, `worker` и новым сервисом.
- Риски:
  - premature split и затягивание delivery.
- Стоимость внедрения:
  - высокая.
- Эксплуатация:
  - ещё один runtime contour и отдельная consistency surface.

### Вариант C (выбран): Platform-owned semantics + worker side effects + thin callback bridge + external adapter contour
- Плюсы:
  - сохраняет Sprint S10 contract как единственный semantic baseline;
  - удерживает raw Telegram transport вне core platform bounded context;
  - оставляет `worker` владельцем async delivery/retry/edit/follow-up side effects;
  - позволяет сделать callback payload opaque и сохранить channel-neutral meaning response.
- Минусы:
  - design-stage обязан подробно описать callback handle/token model и persistence of provider refs;
  - `control-plane` получает дополнительную доменную ответственность.
- Риски:
  - при росте scope может понадобиться future split adapter orchestration.
- Стоимость внедрения:
  - средняя.
- Эксплуатация:
  - предсказуемая при условии typed contracts и единых audit/correlation rules.

## Решение
Мы выбираем: **вариант C — platform-owned semantics + worker side effects + thin callback bridge + external adapter contour**.

## Обоснование (Rationale)
- Этот вариант лучше всего удерживает balance между PRD guardrails Sprint S11 и уже утверждённым S10 interaction baseline.
- Raw Telegram webhook authenticity и callback UX остаются там, где им место: в channel-specific adapter contour, а не в core domain.
- `control-plane` остаётся единственным semantic owner, поэтому duplicate/replay/expired classification и operator visibility не расползаются по нескольким слоям.
- `worker` остаётся естественным владельцем retries, expiry и post-callback UX continuation, не перегружая callback ingress path.

## Последствия (Consequences)
### Позитивные
- У platform появляется единый owner для interaction semantics, wait-state transitions и audit/correlation.
- Telegram transport сохраняется replaceable, а callback payload direction остаётся opaque/server-side.
- Post-callback UX decisions (`edit` vs `follow-up notify`) переходят в async platform-owned path и становятся operator-visible.

### Негативные / компромиссы
- Design-stage должен определить exact callback handle/token format, DTO family и data model для provider message refs.
- Telegram adapter contour остаётся внешним logical boundary, поэтому rollout/rollback notes должны покрывать независимую эволюцию adapter side.

### Технический долг
- Что откладываем:
  - выделение отдельного internal Telegram orchestration service;
  - concrete SDK/runtime choice и deployment topology Telegram adapter contour;
  - richer conversation flows и multi-channel orchestration.
- Когда вернуться:
  - после `run:design` и первых MVP measurements по delivery/replay/operator visibility.

## План внедрения (минимально)
- Изменения в коде:
  - отсутствуют на `run:arch`.
- Изменения в инфраструктуре:
  - отсутствуют.
- Миграции/совместимость:
  - design-stage должен определить interaction/delivery/callback persistence и rollout order `migrations -> control-plane -> worker -> api-gateway -> adapter`.
- Наблюдаемость:
  - design-stage должен зафиксировать event set для dispatch, callback classification, edit/follow-up continuation и manual fallback signals.

## План отката/замены
- Условия отката:
  - если `run:design` покажет, что external adapter contour или chosen ownership не удерживают operator visibility, latency или bounded-context integrity.
- Как откатываем:
  - ADR переводится в `superseded`, а инициативе выбирается либо отдельный internal service, либо другой transport boundary при сохранении Sprint S10 semantics.

## Ссылки
- PRD:
  - `docs/delivery/epics/s11/prd-s11-day3-telegram-user-interaction-adapter.md`
- Architecture:
  - `docs/architecture/initiatives/s11_telegram_user_interaction_adapter/architecture.md`
- Alternatives:
  - `docs/architecture/alternatives/ALT-0006-telegram-user-interaction-adapter-boundaries.md`
- Related baseline:
  - `docs/architecture/initiatives/s10_mcp_user_interactions/design_doc.md`
  - `docs/architecture/mcp_approval_and_audit_flow.md`
  - official Telegram Bot API docs (`Getting updates`, `setWebhook`, callback behaviour), reviewed 2026-03-14
