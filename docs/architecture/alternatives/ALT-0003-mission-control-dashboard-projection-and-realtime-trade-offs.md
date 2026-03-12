---
doc_id: ALT-0003
type: alternatives
title: "Mission Control Dashboard — projection and realtime trade-offs"
status: in-review
owner_role: SA
created_at: 2026-03-12
updated_at: 2026-03-12
related_issues: [333, 335, 337, 340, 351]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-12-issue-340-arch"
---

# Alternatives & Trade-offs: Mission Control Dashboard

## TL;DR
- Рассмотрели: client-side composition / отдельный dashboard-service / control-plane-owned projection + worker reconciliation.
- Рекомендуем: control-plane-owned projection + worker reconciliation.
- Почему: лучший баланс между GitHub-first MVP, thin-edge boundaries, webhook-driven orchestration и degraded usability.

## Контекст
- PRD требует:
  - active-set default;
  - discussion-first formalization;
  - provider-safe commands;
  - webhook echo dedupe;
  - degraded mode без realtime hard dependency.
- Нельзя нарушить:
  - thin-edge для `api-gateway` и `web-console`;
  - GitHub-first provider model и external human review;
  - markdown-only scope для `run:arch`.

## Вариант A: Client-side composition of state
- Описание:
  - web-console сам собирает dashboard state из provider/runtime APIs и локального UI state.
- Плюсы:
  - быстрый старт;
  - минимум новых backend моделей.
- Минусы:
  - доменные правила утекут в frontend;
  - сложно поддержать idempotent command/reconciliation lifecycle.
- Риски:
  - split-brain после webhook echo;
  - degraded mode станет неуправляемым.
- Стоимость/сложность:
  - низкий initial cost, высокий риск rework.

## Вариант B: Отдельный dashboard microservice/read-model service
- Описание:
  - создать новый сервис только для dashboard projections, commands и timeline.
- Плюсы:
  - изоляция dashboard domain;
  - потенциал отдельного масштабирования.
- Минусы:
  - ещё один consistency boundary до design-stage;
  - выше delivery overhead на MVP.
- Риски:
  - преждевременная архитектурная сложность;
  - увеличение integration surface.
- Стоимость/сложность:
  - высокая.

## Вариант C: Control-plane-owned projection + worker reconciliation (recommended)
- Описание:
  - `control-plane` владеет projection/relations/commands;
  - `worker` исполняет provider sync, retries и reconciliation;
  - `api-gateway`/`web-console` дают typed transport и UX.
- Плюсы:
  - сохраняет текущие bounded contexts;
  - позволяет snapshot-first / delta-second model;
  - упрощает audit и traceability.
- Минусы:
  - design-stage обязан строго ограничить projection scope;
  - нужно явно описать stale/fallback behavior.
- Риски:
  - перегрузка `control-plane`, если active-set scope начнёт расти без guardrails.
- Стоимость/сложность:
  - средняя.

## Сравнение
| Критерий | A | B | C |
|---|---:|---:|---:|
| Скорость старта | 5 | 2 | 4 |
| Риск split-brain | 1 | 3 | 5 |
| Соответствие thin-edge | 1 | 4 | 5 |
| Управляемость degraded mode | 2 | 4 | 5 |
| Delivery overhead | 4 | 1 | 4 |
| Auditability | 2 | 4 | 5 |

## Рекомендация
- Выбор: **вариант C**.
- Обоснование:
  - лучше всего согласуется с существующим ownership платформы;
  - не требует premature service split или frontend domain logic;
  - сохраняет realtime как delivery optimization, а не source-of-truth.
- Что теряем:
  - возможность немедленно изолировать dashboard в отдельный сервис.
- Что выигрываем:
  - проверяемую консистентность projections/commands и более короткий путь к `run:design`.

## Нужен апрув от Owner
- [x] Выбор варианта C.
- [x] Разрешение компромисса: отдельный dashboard-service откладывается до появления измеримых scale причин.
