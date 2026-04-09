---
doc_id: ALT-0009
type: alternatives
title: "Unified owner feedback loop — live wait and channel ownership trade-offs"
status: in-review
owner_role: SA
created_at: 2026-03-26
updated_at: 2026-03-26
related_issues: [541, 554, 557, 559, 568]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-26-issue-559-arch"
---

# Alternatives & Trade-offs: Unified owner feedback loop

## TL;DR
- Рассмотрели: detached resume-first path / live wait primary with platform-owned truth / dedicated owner-feedback service now.
- Рекомендуем: live wait primary with platform-owned truth and thin channel surfaces.
- Почему: это лучший баланс между same-session trust, channel parity, bounded contexts и скоростью handover в `run:design`.

## Контекст
- Какие ограничения/требования влияют:
  - same live pod / same `codex` session как primary happy-path;
  - max timeout/TTL built-in `kodex` MCP wait path не ниже owner wait window;
  - snapshot-resume только как recovery fallback;
  - Telegram inbox и staff-console fallback поверх одного persisted backend truth;
  - visibility для overdue / expired / manual-fallback states;
  - `run:self-improve` исключён из owner-facing contract.
- Что нельзя нарушить:
  - thin-edge роль `api-gateway`;
  - bounded role `worker` как async/reconcile contour;
  - external nature `telegram-interaction-adapter`;
  - markdown-only scope на `run:arch`.

## Вариант A: Detached resume-first path и channel-owned projections
- Описание:
  - считать нормальным path, где live wait короткий, а ответ owner в основном поднимает recovery/detached continuation; Telegram и staff-console держат свои почти-автономные projections.
- Плюсы:
  - ниже требования к runtime retention;
  - проще стартовые transport contracts.
- Минусы:
  - ломает same-session baseline;
  - канал начинает диктовать lifecycle semantics.
- Риски:
  - split-brain между Telegram и staff-console;
  - hidden downgrade в resume-first execution model;
  - потеря доверия owner к pending wait.
- Стоимость/сложность:
  - низкая на старте, высокая на исправление drift.

## Вариант B: Live wait primary + platform-owned truth + thin channel surfaces (recommended)
- Описание:
  - `control-plane` владеет request truth и continuation policy;
  - `worker` владеет dispatch/reconcile/lease side effects;
  - `agent-runner` удерживает live session и recovery snapshot;
  - Telegram и staff-console materialize только surfaces поверх общего persisted contract.
- Плюсы:
  - удерживает same-session happy-path;
  - даёт один owner для deadlines, parity и degraded states;
  - сохраняет bounded contexts Sprint S10/S11.
- Минусы:
  - требует аккуратного design-stage контракта для typed actions, projections и rollout;
  - увеличивает domain load `control-plane`.
- Риски:
  - нужен хороший mixed-version rollout для long-lived waits.
- Стоимость/сложность:
  - средняя.

## Вариант C: Dedicated owner-feedback coordinator service уже сейчас
- Описание:
  - выделить новый внутренний сервис, который владеет request truth, projections и continuation orchestration.
- Плюсы:
  - отдельный bounded context;
  - чистый future scaling path.
- Минусы:
  - новый DB owner и rollout contour до фиксации design contracts;
  - задержка doc-stage цепочки `arch -> design -> plan`.
- Риски:
  - premature topology lock-in;
  - рост coordination cost.
- Стоимость/сложность:
  - высокая.

## Сравнение
| Критерий | A | B | C |
|---|---:|---:|---:|
| Same-session trust | 1 | 5 | 4 |
| Channel parity | 2 | 5 | 4 |
| Скорость handover в `run:design` | 4 | 4 | 1 |
| Архитектурная консистентность | 1 | 5 | 3 |
| Операционная простота MVP | 3 | 4 | 2 |
| Риск premature topology lock-in | 4 | 5 | 1 |

## Рекомендация
- Выбор: **Вариант B**.
- Обоснование:
  - это единственный вариант, который одновременно сохраняет same-session primary happy-path, max timeout/TTL baseline и единый persisted truth для Telegram/staff-console;
  - он не требует нового service split и лучше укладывается в уже утверждённые S10/S11 boundaries;
  - он оставляет design stage достаточно сфокусированным: детализировать typed contracts/data model, а не спорить об owner-сервисе заново.
- Что теряем:
  - простоту короткого resume-first path из варианта A;
  - изоляцию отдельного сервиса из варианта C.
- Что выигрываем:
  - воспроизводимый wait/continuation contract, channel parity и более короткий путь к `run:design`.

## Нужен апрув от Owner
- [ ] Выбор варианта B.
- [ ] Подтверждение, что detached resume-run остаётся только recovery fallback, а не допустимым equal happy-path.
