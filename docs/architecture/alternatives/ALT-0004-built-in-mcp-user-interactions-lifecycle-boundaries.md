---
doc_id: ALT-0004
type: alternatives
title: "Built-in MCP user interactions — lifecycle ownership and adapter boundary trade-offs"
status: in-review
owner_role: SA
created_at: 2026-03-12
updated_at: 2026-03-12
related_issues: [360, 378, 383, 385, 387]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-12-issue-385-arch"
---

# Alternatives & Trade-offs: built-in MCP user interactions

## TL;DR
- Рассмотрели: reuse approval flow, отдельный interaction-service сейчас, control-plane-owned interaction lifecycle с worker dispatch.
- Рекомендуем: control-plane-owned interaction lifecycle с `worker` dispatch/retries и thin-edge callback ingress в `api-gateway`.
- Почему: лучший баланс между product guardrails Day3, thin-edge architecture, platform-owned audit/replay safety и отсутствием premature service split.

## Контекст
- PRD требует:
  - `user.notify` как non-blocking completion/next-step path;
  - `user.decision.request` как typed wait-state interaction;
  - separation from approval flow;
  - channel-neutral adapter model;
  - platform-owned retry/idempotency/correlation/audit.
- Нельзя нарушить:
  - built-in server `codex_k8s` как единственную core точку расширения;
  - thin-edge для `api-gateway`;
  - markdown-only scope для `run:arch`.

## Вариант A: Reuse approval flow and `owner.feedback.request`
- Описание:
  - использовать approval tables, approval state vocabulary и текущий owner-feedback callback path как основу для user interactions.
- Плюсы:
  - быстрый старт;
  - reuse уже существующих callback/auth patterns.
- Минусы:
  - business meaning approval и user response разные;
  - обычные user interactions начинают зависеть от approval vocabulary.
- Риски:
  - потеря separation from approval flow;
  - сложный rework при появлении adapters.
- Стоимость/сложность:
  - низкий initial cost, высокий semantic debt.

## Вариант B: Новый interaction-service/read-model service уже на MVP
- Описание:
  - выделить отдельный сервис для interaction aggregate, dispatch и callbacks.
- Плюсы:
  - сильная изоляция bounded context;
  - возможный independent scaling path.
- Минусы:
  - новый DB owner и новый rollout contour ещё до design-stage;
  - выше delivery overhead и больше inter-service consistency work.
- Риски:
  - premature split и затяжка delivery.
- Стоимость/сложность:
  - высокая.

## Вариант C: Control-plane-owned interaction lifecycle + worker dispatch (recommended)
- Описание:
  - `control-plane` владеет interaction aggregate, validation, wait-state и resume;
  - `worker` исполняет outbound dispatch, retries и expiry loops;
  - `api-gateway` принимает callbacks как thin-edge ingress;
  - adapters остаются replaceable transport layer.
- Плюсы:
  - сохраняет текущие bounded contexts;
  - поддерживает platform-owned replay safety и audit;
  - не требует нового runtime server block или отдельного interaction-service.
- Минусы:
  - design-stage обязан строго зафиксировать shared-vs-isolated wait-state semantics;
  - есть риск разрастания `control-plane` без guardrails.
- Риски:
  - при росте scope может потребоваться future service split.
- Стоимость/сложность:
  - средняя.

## Сравнение
| Критерий | A | B | C |
|---|---:|---:|---:|
| Соответствие PRD guardrails | 1 | 4 | 5 |
| Thin-edge consistency | 2 | 4 | 5 |
| Скорость перехода в design/dev | 4 | 1 | 4 |
| Replay/audit correctness | 2 | 4 | 5 |
| Adapter neutrality | 2 | 4 | 5 |
| Delivery overhead | 5 | 1 | 4 |

## Рекомендация
- Выбор: **вариант C**.
- Обоснование:
  - лучше всего согласуется с текущей архитектурой платформы;
  - не ломает separation between interaction flow and approval flow;
  - позволяет отложить service split до появления реальных scale signals.
- Что теряем:
  - возможность сразу масштабировать interaction-domain как отдельный сервис.
- Что выигрываем:
  - проверяемую ownership-модель для Day5 design без premature infrastructure overhead.

## Нужен апрув от Owner
- [x] Выбор варианта C.
- [x] Разрешение компромисса: отдельный interaction-service откладывается до появления измеримых scale/throughput причин.
