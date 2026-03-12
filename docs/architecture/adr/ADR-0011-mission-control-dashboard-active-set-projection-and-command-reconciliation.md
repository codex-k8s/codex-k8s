---
doc_id: ADR-0011
type: adr
title: "Mission Control Dashboard: control-plane-owned active-set projection and command reconciliation"
status: accepted
owner_role: SA
created_at: 2026-03-12
updated_at: 2026-03-12
related_issues: [333, 335, 337, 340, 351]
related_prs: []
supersedes: []
superseded_by: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-12-issue-340-arch"
---

# ADR-0011: Mission Control Dashboard — control-plane-owned active-set projection and command reconciliation

## TL;DR
- Контекст: PRD Sprint S9 требует единый control-plane UX для active set, но без потери GitHub-first review model и webhook-driven orchestration.
- Решение: выбираем persisted active-set projection, relation graph и command ledger под ownership `control-plane`, а outbound provider sync/retries/reconciliation оставляем в `worker`.
- Последствия: dashboard получает единый source-of-truth и predictable degraded mode, но появляется обязательная design-работа по projection schema, command states и reconciliation evidence.

## Контекст
- Проблема:
  - client-side композиция из GitHub/runtime ответов приведёт к split-brain и не даст доказуемого dedupe после webhook echo;
  - отдельный dashboard-сервис сейчас преждевременно добавит новый consistency contour и усложнит stage handover.
- Ограничения:
  - `run:arch` остаётся markdown-only;
  - thin-edge для `api-gateway` и `web-console` обязателен;
  - GitHub-first MVP, external human review и active-set default нельзя нарушать.
- Связанные требования:
  - `FR-337-01..FR-337-14`;
  - `NFR-337-02`, `NFR-337-03`, `NFR-337-05`, `NFR-337-06`, `NFR-337-07`;
  - `FR-003`, `FR-004`, `FR-025`, `FR-028`, `FR-033`.
- Что ломается без решения:
  - неясный owner для active-set projection и command lifecycle;
  - недоказуемая корректность command -> webhook echo reconciliation;
  - невозможность спроектировать degraded mode без отдельного client-only state machine.

## Decision Drivers
- Детерминизм command/reconciliation lifecycle.
- Сохранение thin-edge и bounded-context границ.
- Возможность degraded mode без realtime hard dependency.
- Отсутствие premature microservice split и library lock-in.
- Прозрачная traceability для `discussion -> formal task` и provider sync.

## Рассмотренные варианты

### Вариант A: Client-side composition + minimal backend state
- Плюсы:
  - быстрый старт реализации;
  - минимум новых persisted моделей.
- Минусы:
  - frontend становится неявным владельцем active-set semantics;
  - сложно доказать dedupe/correlation correctness.
- Риски:
  - split-brain между provider state, runtime state и локальным UI state;
  - degraded mode быстро превращается в набор ad-hoc polling hacks.
- Стоимость внедрения:
  - низкая на старте, высокая на стабилизации.
- Эксплуатация:
  - трудный аудит и сложный postmortem для reconciliation bugs.

### Вариант B: Отдельный Mission Control service/read-model service уже сейчас
- Плюсы:
  - сильная изоляция dashboard scope;
  - потенциально проще масштабировать read-model отдельно.
- Минусы:
  - новый сервис и новый ownership contour до уточнения design contracts;
  - дополнительная межсервисная консистентность уже на MVP.
- Риски:
  - over-engineering и удлинение delivery before value.
- Стоимость внедрения:
  - высокая.
- Эксплуатация:
  - больше runtime-составляющих и operational overhead.

### Вариант C (выбран): Control-plane-owned projection + worker reconciliation
- Плюсы:
  - единый persisted source-of-truth для dashboard;
  - webhook-driven model и current service boundaries сохраняются;
  - realtime остаётся delivery-ускорителем, а не отдельным truth layer.
- Минусы:
  - design-stage обязан подробно описать projection schema и command states;
  - `control-plane` получает дополнительную projection responsibility.
- Риски:
  - если projection scope разрастётся без guardrails, `control-plane` может стать перегруженным.
- Стоимость внедрения:
  - средняя.
- Эксплуатация:
  - предсказуемая при условии typed contracts и lease-aware reconciliation.

## Решение
Мы выбираем: **вариант C — control-plane-owned active-set projection + worker reconciliation**.

## Обоснование (Rationale)
- Этот вариант лучше всего соответствует уже утверждённой архитектуре платформы:
  - `control-plane` владеет доменной консистентностью и persisted state;
  - `worker` реализует async/retry/reconciliation;
  - `api-gateway` и `web-console` не получают доменное ownership.
- Он сохраняет GitHub-first и webhook-driven подход без появления нового service boundary до того, как зафиксированы typed contracts.
- Он естественно поддерживает snapshot-first / delta-second model, необходимую для degraded UX и active-set fallback.

## Последствия (Consequences)

### Позитивные
- Для active-set projection, relation graph и command state появляется явный владелец.
- Discussion formalization и provider-safe commands получают единый audit/reconcile path.
- Можно проектировать degraded mode и realtime как дополняющие, а не конкурирующие пути.

### Негативные / компромиссы
- Придётся детально проработать projection freshness, stale markers и reconciliation evidence на `run:design`.
- Выделение отдельного dashboard-сервиса откладывается, даже если позже появятся scale reasons.

### Технический долг
- Что откладываем:
  - отдельный read-model service;
  - decision по конкретной graph/realtime/STT библиотеке;
  - advanced analytics/ML ranking для active-set cards.
- Когда вернуться:
  - после `run:design` и первых MVP pilot measurements.

## План внедрения (минимально)
- Изменения в коде:
  - отсутствуют на `run:arch`.
- Изменения в инфраструктуре:
  - отсутствуют.
- Миграции/совместимость:
  - design-stage должен определить projection/timeline/command schema и rollout order.
- Наблюдаемость:
  - design-stage должен зафиксировать event set для freshness, command dedupe, degraded mode и voice isolation.

## План отката/замены
- Условия отката:
  - если `run:design` покажет, что projection scope невозможно удержать в `control-plane` без service split.
- Как откатываем:
  - ADR переводится в `superseded`, а dashboard projection выносится в отдельный read-model service с тем же PRD guardrail набором.

## Ссылки
- PRD:
  - `docs/delivery/epics/s9/prd-s9-day3-mission-control-dashboard.md`
- Architecture:
  - `docs/architecture/initiatives/s9_mission_control_dashboard/architecture.md`
- Alternatives:
  - `docs/architecture/alternatives/ALT-0003-mission-control-dashboard-projection-and-realtime-trade-offs.md`
- Related baseline:
  - `docs/architecture/api_contract.md`
  - `docs/architecture/data_model.md`
  - `docs/delivery/epics/s3/epic-s3-day19.5-realtime-event-bus-and-websocket-backplane.md`
