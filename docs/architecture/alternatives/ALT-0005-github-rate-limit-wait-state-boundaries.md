---
doc_id: ALT-0005
type: alternatives
title: "GitHub API rate-limit resilience — wait-state ownership and contour boundary trade-offs"
status: in-review
owner_role: SA
created_at: 2026-03-13
updated_at: 2026-03-13
related_issues: [366, 413, 416, 418, 420]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-13-issue-418-arch"
---

# Alternatives & Trade-offs: GitHub API rate-limit resilience

## TL;DR
- Рассмотрели: локальный retry/backoff per service, новый quota-orchestrator сервис, `control-plane`-owned wait semantics с `worker` resume orchestration.
- Рекомендуем: `control-plane`-owned wait semantics с `worker` resume orchestration и `agent-runner` handoff-only behavior.
- Почему: лучший баланс между PRD guardrails, thin-edge architecture, multi-pod audit consistency и ограничением Day4 на markdown-only stage.

## Контекст
- PRD требует:
  - controlled wait вместо ложного `failed`;
  - split `platform PAT` vs `agent bot-token`;
  - provider-driven uncertainty для secondary limits;
  - hard-failure separation;
  - no infinite local retries.
- Нельзя нарушить:
  - GitHub-first baseline;
  - thin-edge роль `api-gateway` и `web-console`;
  - существующий bounded context `control-plane` / `worker` / `agent-runner`.

## Вариант A: Локальный retry/backoff внутри каждого участника
- Описание:
  - каждый сервис и agent pod сам решает, когда ждать и когда повторять GitHub вызов.
- Плюсы:
  - быстрое внедрение;
  - минимум новых persisted concepts.
- Минусы:
  - нет единого source-of-truth для wait semantics;
  - visibility и contour attribution становятся случайными.
- Риски:
  - agent path продолжит local retry loop;
  - вторичные лимиты будут трактоваться по-разному в разных местах.
- Стоимость/сложность:
  - низкая initial cost, высокий operational debt.

## Вариант B: Новый quota-orchestrator сервис уже сейчас
- Описание:
  - вынести detect/classify/wait/resume в отдельный сервис и отдельный DB owner.
- Плюсы:
  - сильная изоляция bounded context;
  - future-friendly scaling path.
- Минусы:
  - новый runtime contour ещё до design-stage;
  - выше delivery overhead и coordination cost.
- Риски:
  - premature split;
  - затяжка Sprint S12 до решения core PRD outcomes.
- Стоимость/сложность:
  - высокая.

## Вариант C: `control-plane` owns wait semantics, `worker` owns resume orchestration (recommended)
- Описание:
  - `control-plane` классифицирует raw evidence, ведёт wait aggregate и формирует visibility contract;
  - `worker` исполняет wake-up scheduling и finite auto-resume attempts;
  - `agent-runner` только передаёт raw evidence и не retry'ит локально после handoff;
  - `api-gateway` и `web-console` читают typed projection.
- Плюсы:
  - сохраняет текущие service boundaries;
  - даёт единый owner для contour attribution и recovery hints;
  - минимизирует риск false-failed и endless retries.
- Минусы:
  - design-stage должен точно описать finite auto-resume policy;
  - `control-plane` получает дополнительную доменную нагрузку.
- Риски:
  - при росте scope возможен future split.
- Стоимость/сложность:
  - средняя.

## Сравнение
| Критерий | A | B | C |
|---|---:|---:|---:|
| Соответствие PRD guardrails | 1 | 4 | 5 |
| Contour fidelity | 1 | 4 | 5 |
| No-local-retry discipline | 1 | 4 | 5 |
| Thin-edge consistency | 2 | 4 | 5 |
| Delivery speed into design | 4 | 1 | 4 |
| Operational clarity | 1 | 4 | 5 |

## Рекомендация
- Выбор: **вариант C**.
- Обоснование:
  - лучше всего согласуется с текущей платформенной архитектурой и PRD Sprint S12;
  - позволяет централизовать semantics без нового сервиса;
  - удерживает `agent-runner` в роли source emitter, а не владельца бизнес-решений.
- Что теряем:
  - immediate separate scaling contour.
- Что выигрываем:
  - единый owner для controlled wait и более надёжный переход в `run:design`.

## Нужен апрув от Owner
- [ ] Выбор варианта C.
- [ ] Подтверждение компромисса: отдельный quota-orchestrator service не нужен до появления измеримых scale signals.
