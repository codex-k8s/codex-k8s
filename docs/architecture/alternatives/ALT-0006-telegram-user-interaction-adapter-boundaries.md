---
doc_id: ALT-0006
type: alternatives
title: "Telegram user interaction adapter — lifecycle ownership and boundary trade-offs"
status: in-review
owner_role: SA
created_at: 2026-03-14
updated_at: 2026-03-14
related_issues: [361, 444, 447, 448, 452, 454]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-14-issue-452-arch"
---

# Alternatives & Trade-offs: Telegram user interaction adapter

## TL;DR
- Рассмотрели: Telegram-first transport ownership, новый internal Telegram service, platform-owned semantics with external adapter contour.
- Рекомендуем: platform-owned semantics с `control-plane` owner, `worker` side effects, thin callback bridge в `api-gateway` и внешним Telegram adapter contour.
- Почему: лучший баланс между Sprint S10/S11 guardrails, thin-edge consistency, replay safety, multi-channel trajectory и умеренным delivery overhead.

## Контекст
- PRD требует:
  - первый внешний Telegram channel path поверх Sprint S10 interaction contract;
  - inline callbacks и optional free-text без Telegram-first drift;
  - callback/webhook security, duplicate/replay/expired safety и operator visibility;
  - separation from approval flow.
- Нельзя нарушить:
  - `api-gateway` как thin-edge;
  - platform-owned semantics и wait-state lifecycle;
  - markdown-only scope `run:arch`;
  - continuity `arch -> design -> plan -> dev`.

## Вариант A: Telegram-first transport ownership
- Описание:
  - raw Telegram webhook и semantic callback handling живут в adapter/gateway path, а callback payload несёт слишком много business meaning.
- Плюсы:
  - быстрый путь к working prototype;
  - меньше platform-side abstractions на старте.
- Минусы:
  - semantic classification оказывается на transport boundary;
  - `api-gateway` и adapter contour получают лишнюю доменную ответственность.
- Риски:
  - callback replay safety и operator visibility становятся transport-local;
  - channel-neutral contract размывается Telegram fields и callback data.
- Стоимость/сложность:
  - низкий initial cost, высокий semantic debt.

## Вариант B: Новый internal Telegram service уже сейчас
- Описание:
  - выделить отдельный внутренний сервис для Telegram delivery, callback lifecycle и operator visibility.
- Плюсы:
  - сильная изоляция channel-specific bounded context;
  - отдельный scale path для Telegram workload.
- Минусы:
  - новый DB owner и новый rollout contour до design-stage;
  - больше coordination между `control-plane`, `worker` и новым сервисом.
- Риски:
  - premature architecture split и затянутый delivery.
- Стоимость/сложность:
  - высокая.

## Вариант C: Platform-owned semantics + external adapter contour (recommended)
- Описание:
  - `control-plane` владеет semantic lifecycle, correlation и operator visibility;
  - `worker` владеет dispatch/retry/expiry и post-callback edit/follow-up actions;
  - `api-gateway` принимает normalized callbacks с platform auth;
  - raw Telegram webhook, secret-token verification и callback query acknowledgement остаются во внешнем adapter contour.
- Плюсы:
  - channel-neutral contract Sprint S10 остаётся source-of-truth;
  - thin-edge и replay safety сохраняются;
  - edit-vs-follow-up policy остаётся platform-owned и operator-visible.
- Минусы:
  - design-stage должен отдельно детализировать callback handle/token rules и data model provider refs;
  - внешний adapter contour добавляет rollout coordination.
- Риски:
  - при росте scope может понадобиться future service split.
- Стоимость/сложность:
  - средняя.

## Сравнение
| Критерий | A | B | C |
|---|---:|---:|---:|
| Соответствие Sprint S10/S11 guardrails | 1 | 4 | 5 |
| Thin-edge consistency | 1 | 4 | 5 |
| Replay/expiry safety | 2 | 4 | 5 |
| Multi-channel reuse | 1 | 4 | 5 |
| Delivery overhead | 5 | 1 | 4 |
| Operator visibility clarity | 2 | 4 | 5 |

## Рекомендация
- Выбор: **вариант C**.
- Обоснование:
  - raw Telegram specifics остаются вне core bounded context;
  - semantic lifecycle, audit/correlation и manual fallback остаются едиными в платформе;
  - design-stage получает достаточно чёткую ownership-модель без premature service split.
- Что теряем:
  - самый короткий path к transport-first prototype и отдельный scale contour уже на Day4.
- Что выигрываем:
  - устойчивую platform-owned semantics, operator-safe lifecycle и future multi-channel trajectory.

## Нужен апрув от Owner
- [x] Выбор варианта C.
- [x] Разрешение компромисса: raw Telegram webhook/auth остаётся во внешнем adapter contour, а callback payload direction фиксируется как opaque/server-side lookup strategy.
