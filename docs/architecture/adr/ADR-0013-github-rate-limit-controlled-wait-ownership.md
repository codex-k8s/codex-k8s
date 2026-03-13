---
doc_id: ADR-0013
type: adr
title: "GitHub API rate-limit resilience: control-plane-owned controlled wait with worker resume orchestration"
status: accepted
owner_role: SA
created_at: 2026-03-13
updated_at: 2026-03-13
related_issues: [366, 413, 416, 418, 420, 423, 425, 426, 427, 428, 429, 430, 431]
related_prs: []
supersedes: []
superseded_by: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-03-13-issue-418-arch"
  approved_by: "ai-da-stas"
  approved_at: 2026-03-13
---

# ADR-0013: GitHub API rate-limit resilience — control-plane-owned controlled wait with worker resume orchestration

## TL;DR
- Контекст: Sprint S12 PRD требует controlled wait capability для recoverable GitHub rate-limit без смешения `platform PAT` и `agent bot-token`, без local retry-loop и без ложного countdown promises.
- Решение: выбираем `control-plane` как единственного владельца classification, controlled wait aggregate, contour attribution и recovery hints; `worker` исполняет time-based resume orchestration, а `agent-runner` только передаёт raw evidence и прекращает локальные retries.
- Последствия: появляется единый domain owner для wait-state semantics и visibility, но design-stage обязан отдельно описать transport/data contracts и finite auto-resume policy.

## Контекст
- Проблема:
  - сейчас ownership detect/classify/wait/resume не закреплён и может расползтись между `control-plane`, `worker`, `agent-runner`, `api-gateway` и UI;
  - agent-path особенно рискован: без явного owner локальный retry-loop может жить внутри pod и обходить product contract.
- Ограничения:
  - `run:arch` остаётся markdown-only;
  - GitHub остаётся единственным provider baseline;
  - `api-gateway` и `web-console` должны остаться thin adapters;
  - product contract обязан сохранять два user-facing contour и hard-failure separation.
- Связанные требования:
  - `docs/delivery/epics/s12/prd-s12-day3-github-api-rate-limit-resilience.md`
  - `FR-003`, `FR-012`, `FR-026`, `FR-028`, `FR-032`, `FR-033`
  - `NFR-010`, `NFR-012`, `NFR-018`
- Что ломается без решения:
  - false-failed и contour drift не будут воспроизводимо устранимы;
  - design-stage не сможет зафиксировать DTO/schema без повторного спора об owner-сервисе;
  - UI и service-comments снова станут собирать смысл из сырых логов.

## Decision Drivers
- Единый domain owner для classification и wait semantics.
- Сохранение split `platform PAT` vs `agent bot-token`.
- Запрет infinite local retries на agent path.
- Thin-edge consistency для `api-gateway` и `web-console`.
- Отсутствие premature quota-orchestrator service split.

## Рассмотренные варианты

### Вариант A: Локальный retry/backoff в каждом сервисе и agent pod
- Плюсы:
  - низкий initial cost;
  - минимум новых persisted concepts.
- Минусы:
  - rate-limit semantics фрагментируются по сервисам;
  - невозможно доказать единый wait-state UX.
- Риски:
  - agent pod продолжит brute-force retries;
  - owner/operator получат inconsistent signals.
- Стоимость внедрения:
  - низкая на старте, высокая по operational debt.
- Эксплуатация:
  - постфактум трудно объяснить, какой контур был заблокирован и почему.

### Вариант B: Новый quota-orchestrator сервис уже на Day4
- Плюсы:
  - отдельный bounded context;
  - потенциальный future scaling path.
- Минусы:
  - новый сервис и новый DB owner до design-stage;
  - лишний rollout contour и coordination overhead.
- Риски:
  - premature architecture split и задержка Sprint S12.
- Стоимость внедрения:
  - высокая.
- Эксплуатация:
  - ещё один cross-service consistency path.

### Вариант C (выбран): `control-plane` owns semantics, `worker` owns resume orchestration, `agent-runner` only hands off evidence
- Плюсы:
  - сохраняет текущие service boundaries;
  - даёт единый owner для wait-state, contour attribution и recovery hints;
  - удерживает agent path в platform-managed backpressure discipline.
- Минусы:
  - `control-plane` получает дополнительную доменную ответственность;
  - design-stage должен аккуратно зафиксировать finite auto-resume policy.
- Риски:
  - при росте scope может понадобиться later service split.
- Стоимость внедрения:
  - средняя.
- Эксплуатация:
  - предсказуемая при условии typed contracts и wait sweeps с audit evidence.

## Решение
Мы выбираем: **вариант C — `control-plane` owns semantics, `worker` owns resume orchestration, `agent-runner` only hands off evidence**.

## Обоснование (Rationale)
- Это лучший баланс между PRD guardrails и текущими service boundaries платформы:
  - `control-plane` уже владеет run/session lifecycle, audit и policy;
  - `worker` естественно подходит для time-based reconciliation;
  - `agent-runner` не должен становиться domain owner только потому, что первым увидел CLI error.
- Решение минимизирует риск false-failed и contour drift, не вводя новый сервис до появления фактических scale signals.
- Оно делает thin-edge `api-gateway`/`web-console` доказуемым: UI читает typed projection, а не придумывает бизнес-смысл.

## Последствия (Consequences)

### Позитивные
- Появляется единый owner для recoverable/hard-failure classification и controlled wait aggregate.
- Agent path принудительно переходит в platform-managed backpressure, а не в local retry loop.
- Visibility surfaces и audit остаются синхронизированы через один persisted source-of-truth.

### Негативные / компромиссы
- `control-plane` получает больше доменной логики вокруг GitHub provider signals.
- Design-stage обязан отдельно доказать finite auto-resume strategy, чтобы не превратить wait-state в бесконечную очередь.

### Технический долг
- Что откладываем:
  - отдельный quota-orchestrator service;
  - multi-provider budgeting;
  - predictive quota analytics.
- Когда вернуться:
  - после design-stage и первых production-like measurements по S12.

## План внедрения (минимально)
- Изменения в коде:
  - отсутствуют на `run:arch`.
- Изменения в инфраструктуре:
  - отсутствуют.
- Миграции/совместимость:
  - design-stage должен определить persisted wait aggregate, sweep state и backward-safe rollout.
- Наблюдаемость:
  - design-stage должен зафиксировать event set для detect, classify, wait entered, auto-resume attempted, escalated manual action.

## План отката/замены
- Условия отката:
  - если на `run:design` выяснится, что finite auto-resume policy и aggregate ownership не удерживаются в `control-plane` без нарушения bounded-context integrity.
- Как откатываем:
  - ADR переводится в `superseded`, а wait-state lifecycle выносится в отдельный orchestration service с теми же PRD guardrails.

## Continuity after `run:plan`
- Plan package Issue `#423` подтвердил это ADR как owner-baseline для execution waves `#425..#431`.
- Ни одна wave не может перенести ownership classification/recovery semantics из `control-plane` в `worker`, `agent-runner`, `api-gateway` или `web-console`.

## Ссылки
- PRD:
  - `docs/delivery/epics/s12/prd-s12-day3-github-api-rate-limit-resilience.md`
- Architecture:
  - `docs/architecture/initiatives/s12_github_api_rate_limit_resilience/architecture.md`
- Alternatives:
  - `docs/architecture/alternatives/ALT-0005-github-rate-limit-wait-state-boundaries.md`
- Related baseline:
  - `docs/architecture/api_contract.md`
  - `docs/architecture/data_model.md`
  - `docs/architecture/agent_runtime_rbac.md`
  - `docs/architecture/mcp_approval_and_audit_flow.md`
