---
doc_id: ADR-0012
type: adr
title: "Built-in MCP user interactions: control-plane-owned lifecycle with worker dispatch and thin-edge callbacks"
status: accepted
owner_role: SA
created_at: 2026-03-12
updated_at: 2026-03-13
related_issues: [360, 378, 383, 385, 387]
related_prs: []
supersedes: []
superseded_by: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-03-12-issue-385-arch"
  approved_by: "ai-da-stas"
  approved_at: 2026-03-13
---

# ADR-0012: Built-in MCP user interactions — control-plane-owned lifecycle with worker dispatch and thin-edge callbacks

## TL;DR
- Контекст: PRD Sprint S10 требует channel-neutral user interaction path для `user.notify` и `user.decision.request`, но без смешения с approval flow.
- Решение: выбираем control-plane-owned interaction aggregate и wait-state lifecycle, оставляя outbound dispatch/retries/expiry в `worker`, а callback ingress/auth на thin-edge `api-gateway`.
- Последствия: появляется явный доменный owner для interaction semantics и replay safety, но design-stage обязан отдельно описать DTO/schema/migration details и shared-vs-isolated wait-state infrastructure.

## Контекст
- Проблема:
  - reuse approval flow (`owner.feedback.request`, approval states, approval records) сломает separation between interaction-domain and control-domain;
  - если agent pod или adapters будут владеть callback lifecycle, платформа потеряет reproducible audit/correlation и safe retries.
- Ограничения:
  - `run:arch` остаётся markdown-only;
  - built-in server `kodex` нельзя заменять новым runtime server block;
  - `api-gateway` и adapters обязаны остаться thin transport layers без business-state ownership.
- Связанные требования:
  - PRD `docs/delivery/epics/s10/prd-s10-day3-mcp-user-interactions.md`;
  - `FR-003`, `FR-004`, `FR-006`, `FR-025`, `FR-028`;
  - `NFR-010`, `NFR-012`, `NFR-016`, `NFR-018`.
- Что ломается без решения:
  - неясный владелец interaction aggregate, wait-state и callback idempotency;
  - высокий риск повторно смешать approval flow и user interaction flow;
  - design-stage не сможет доказать replay safety и adapter neutrality.

## Decision Drivers
- Separation between interaction flow and approval flow.
- Platform-owned wait-state, audit/correlation and replay safety.
- Сохранение thin-edge для `api-gateway` и adapters.
- Отсутствие premature microservice split.
- Channel-neutral contract без Telegram-first lock-in.

## Рассмотренные варианты

### Вариант A: Reuse approval flow как основу для user interactions
- Плюсы:
  - быстрый старт за счёт существующих callback endpoints и approval records;
  - минимум новых доменных понятий на старте.
- Минусы:
  - approval semantics (`approved/denied/applied`) не совпадают с interaction semantics;
  - обычный user response начинает зависеть от control/approval vocabulary.
- Риски:
  - product guardrail "interaction != approval" станет недоказуемым;
  - будущие adapters будут наследовать неверную бизнес-модель.
- Стоимость внедрения:
  - низкая на старте, высокая при исправлении semantic drift.
- Эксплуатация:
  - аудит и postmortem будут путать approval events и user responses.

### Вариант B: Новый interaction-service уже сейчас
- Плюсы:
  - сильная изоляция interaction-domain;
  - потенциально проще масштабировать delivery отдельно.
- Минусы:
  - новый service boundary и новый DB owner до фиксации design contracts;
  - больше orchestration и rollout overhead на MVP.
- Риски:
  - premature architecture split и усложнение delivery path.
- Стоимость внедрения:
  - высокая.
- Эксплуатация:
  - ещё один runtime contour с межсервисной консистентностью.

### Вариант C (выбран): Control-plane-owned interaction lifecycle + worker dispatch + thin-edge callbacks
- Плюсы:
  - сохраняет current service boundaries платформы;
  - даёт явный owner для interaction aggregate, wait-state и callback validation;
  - оставляет adapters channel-neutral extensions, а не core owners.
- Минусы:
  - design-stage обязан отдельно описать persisted model и wait-state migration notes;
  - `control-plane` получает дополнительную доменную ответственность.
- Риски:
  - если interaction scope начнёт расти без guardrails, `control-plane` можно перегрузить.
- Стоимость внедрения:
  - средняя.
- Эксплуатация:
  - предсказуемая при условии typed contracts и lease-aware reconciliation.

## Решение
Мы выбираем: **вариант C — control-plane-owned interaction lifecycle + worker dispatch + thin-edge callbacks**.

## Обоснование (Rationale)
- Этот вариант лучше всего соответствует уже утверждённой архитектуре платформы:
  - `control-plane` владеет built-in MCP surface, state transitions и audit;
  - `worker` реализует async/retry/expiry loops;
  - `api-gateway` остаётся edge transport boundary.
- Он сохраняет product guardrail о separation from approval flow и не требует нового runtime service до появления измеримых scale signals.
- Он естественно поддерживает channel-neutral adapter model: adapters получают typed envelopes, но не становятся owners core semantics.

## Последствия (Consequences)

### Позитивные
- Для interaction aggregate, response validation и wait-state resume появляется явный владелец.
- Approval/control domain остаётся отдельным bounded context, даже если часть transport plumbing переиспользуется.
- Worker retry/expiry loops и callback replay handling можно проектировать как platform-owned evidence path.

### Негативные / компромиссы
- На `run:design` придётся отдельно зафиксировать interaction DTO/schema и boundary with existing wait-state taxonomy.
- Выделение нового interaction-service откладывается, даже если позже появятся throughput/scale причины.

### Технический долг
- Что откладываем:
  - отдельный interaction/read-model service;
  - выбор конкретных adapter SDK/protocol libraries;
  - richer conversation features, reminders и voice flows.
- Когда вернуться:
  - после design-stage и первых MVP measurements.

## План внедрения (минимально)
- Изменения в коде:
  - отсутствуют на `run:arch`.
- Изменения в инфраструктуре:
  - отсутствуют.
- Миграции/совместимость:
  - design-stage должен определить interaction records, delivery attempts, callback evidence и wait-state rollout path.
- Наблюдаемость:
  - design-stage должен зафиксировать event set для dispatch, retry, expiry, accepted/rejected response и resume.

## План отката/замены
- Условия отката:
  - если `run:design` покажет, что interaction-domain не удерживается в `control-plane` без нарушения SLO или bounded-context integrity.
- Как откатываем:
  - ADR переводится в `superseded`, а interaction lifecycle выносится в отдельный сервис с теми же PRD guardrails.

## Ссылки
- PRD:
  - `docs/delivery/epics/s10/prd-s10-day3-mcp-user-interactions.md`
- Architecture:
  - `docs/architecture/initiatives/s10_mcp_user_interactions/architecture.md`
- Alternatives:
  - `docs/architecture/alternatives/ALT-0004-built-in-mcp-user-interactions-lifecycle-boundaries.md`
- Related baseline:
  - `docs/architecture/api_contract.md`
  - `docs/architecture/data_model.md`
  - `docs/architecture/mcp_approval_and_audit_flow.md`
