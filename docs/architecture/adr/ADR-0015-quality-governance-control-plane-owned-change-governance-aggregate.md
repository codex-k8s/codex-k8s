---
doc_id: ADR-0015
type: adr
title: "Quality Governance System: control-plane-owned change-governance aggregate with worker reconciliation"
status: accepted
owner_role: SA
created_at: 2026-03-15
updated_at: 2026-03-15
related_issues: [466, 469, 470, 471, 476, 484, 488, 494]
related_prs: []
supersedes: []
superseded_by: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-15-issue-484-arch"
---

# ADR-0015: Quality Governance System — control-plane-owned change-governance aggregate with worker reconciliation

## TL;DR
- Контекст: Sprint S13 PRD требует единый product contract для explicit risk tier, separate evidence/verification/waiver constructs, hidden `internal working draft`, semantic-wave publication discipline и downstream boundary `Sprint S13 -> Sprint S14`.
- Решение: выбираем `control-plane` как единственного владельца canonical change-governance aggregate и publication gate; `worker` исполняет asynchronous sweeps/reclassification, а `agent-runner` и UI surfaces только поставляют signals/commands поверх typed contracts.
- Последствия: появляется один domain owner для policy semantics и audit, но design-stage обязан отдельно зафиксировать transport/data contracts, rollout/backfill notes и operator surfaces.

## Контекст
- Проблема:
  - без закреплённого owner риск/evidence/waiver/publication semantics расползутся между `agent-runner`, `api-gateway`, `web-console`, GitHub comments и background jobs;
  - publication policy `internal working draft -> semantic wave map -> published waves` требует централизованного gate, иначе raw draft и semantically mixed bundles снова будут попадать в review stream;
  - Sprint S14 (`#470`) может преждевременно переоткрыть policy semantics под видом runtime/UI constraints, если Day4 не зафиксирует canonical owner.
- Ограничения:
  - `run:arch` остаётся markdown-only;
  - `api-gateway` и `web-console` должны оставаться thin adapters;
  - `high/critical` changes не допускают silent waivers;
  - не вводим новый runtime service до design-stage.
- Связанные требования:
  - `docs/delivery/epics/s13/prd-s13-day3-quality-governance-system.md`
  - `FR-026`, `FR-028`, `FR-033`, `FR-053`, `FR-054`
  - `NFR-010`, `NFR-018`
- Что ломается без решения:
  - owner/reviewer/operator будут видеть разные версии governance state;
  - design-stage снова вернётся к спору об owner-сервисе вместо typed contracts;
  - boundary `Sprint S13 -> Sprint S14` потеряет силу и превратится в implementation-first drift.

## Decision Drivers
- Единый domain owner для policy semantics и audit trail.
- Сохранение separate constructs `risk tier / evidence completeness / verification minimum / waiver state / publication state`.
- Запрет raw draft publication и silent waivers для `high/critical`.
- Thin-edge consistency для `api-gateway` и `web-console`.
- Отсутствие premature service split до появления доказанных scale signals.

## Рассмотренные варианты

### Вариант A: GitHub-native/docs-only governance state
- Плюсы:
  - минимальный initial cost;
  - опора на уже существующие issue/PR surfaces.
- Минусы:
  - canonical state размазывается по комментариям, labels и narrative notes;
  - publication discipline и proportionality нельзя доказать автоматически даже на уровне design.
- Риски:
  - raw draft leakage;
  - inconsistent owner/operator visibility;
  - S14 runtime/UI stream начнёт подменять policy semantics.
- Стоимость внедрения:
  - низкая на старте, высокая по operational debt.
- Эксплуатация:
  - трудно восстановить, какой governance state считался каноническим в конкретный момент.

### Вариант B: Новый dedicated quality-governance service уже на Day4
- Плюсы:
  - отдельный bounded context;
  - потенциальный future scaling path.
- Минусы:
  - новый service boundary и новый DB owner до фиксации typed contracts;
  - лишний rollout contour и coordination overhead внутри Sprint S13.
- Риски:
  - задержка `run:design` и `run:plan`;
  - появится лишний integration debt между `control-plane`, `worker` и новым сервисом.
- Стоимость внедрения:
  - высокая.
- Эксплуатация:
  - ещё один cross-service consistency path без доказанной необходимости.

### Вариант C (выбран): `control-plane` owns aggregate, `worker` owns asynchronous reconciliation, thin surfaces consume typed projections
- Плюсы:
  - сохраняет текущие service boundaries платформы;
  - даёт единый owner для risk/evidence/waiver/publication semantics;
  - удерживает `agent-runner`, `api-gateway` и `web-console` в роли adapters, а не policy owners.
- Минусы:
  - `control-plane` получает дополнительную доменную ответственность;
  - design-stage должен аккуратно описать rollout/backfill и operator surfaces.
- Риски:
  - при росте throughput может понадобиться future extraction.
- Стоимость внедрения:
  - средняя.
- Эксплуатация:
  - предсказуемая при условии typed contracts и audit-safe reconciliation.

## Решение
Мы выбираем: **вариант C — `control-plane` owns canonical governance aggregate, `worker` owns asynchronous reconciliation, thin surfaces consume typed projections**.

## Обоснование (Rationale)
- Вариант C лучше всего соответствует текущим архитектурным границам платформы:
  - `control-plane` уже владеет run/session lifecycle, policy, label transitions и audit;
  - `worker` естественно подходит для sweeps, gap reconciliation и late reclassification;
  - `agent-runner` не должен становиться domain owner только потому, что первым увидел draft/evidence.
- Решение сохраняет PRD guardrails без premature service split и делает boundary `Sprint S13 -> Sprint S14` проверяемой: runtime/UI stream будет обязан потреблять typed surfaces, а не придумывать новую policy semantics.
- Thin-edge для `api-gateway` и `web-console` остаётся доказуемым: они читают/передают typed projections и commands, но не вычисляют governance state самостоятельно.

## Последствия (Consequences)

### Позитивные
- Появляется единый owner canonical governance semantics и audit trail.
- Publication gate `working draft -> semantic waves -> published waves` можно проектировать как один domain lifecycle.
- Asynchronous feedback и late reclassification получают свой owner без переноса policy logic в UI или agent pod.

### Негативные / компромиссы
- `control-plane` получает больше доменной логики вокруг change governance.
- Design-stage должен отдельно доказать typed projections, rollout/backfill notes и proportional low-risk path, чтобы не превратить aggregate в бюрократический монолит.

### Технический долг
- Что откладываем:
  - отдельный quality-governance service;
  - advanced runtime/UI automation Sprint S14;
  - service-specific tuning и predictive governance analytics.
- Когда вернуться:
  - после `run:design` / `run:plan`, если появятся явные scale или ownership breakdown signals.

## План внедрения (минимально)
- Изменения в коде:
  - отсутствуют на `run:arch`.
- Изменения в инфраструктуре:
  - отсутствуют.
- Миграции/совместимость:
  - `run:design` должен определить canonical aggregate, wave lineage, audit linkage и backfill strategy.
- Наблюдаемость:
  - `run:design` должен зафиксировать event set для publication, completeness evaluation, waiver, release-ready, late reclassification и governance-gap reconciliation.

## План отката/замены
- Условия отката:
  - если на `run:design` выяснится, что canonical aggregate не удерживается в `control-plane` без нарушения bounded-context integrity или unacceptable coupling.
- Как откатываем:
  - ADR переводится в `superseded`, а governance aggregate выносится в отдельный orchestration service с сохранением тех же PRD guardrails.

## Ссылки
- PRD:
  - `docs/delivery/epics/s13/prd-s13-day3-quality-governance-system.md`
- Architecture:
  - `docs/architecture/initiatives/s13_quality_governance_system/architecture.md`
- Alternatives:
  - `docs/architecture/alternatives/ALT-0007-quality-governance-boundaries.md`
- Related baseline:
  - `docs/architecture/api_contract.md`
  - `docs/architecture/data_model.md`
  - `docs/architecture/agent_runtime_rbac.md`
  - `docs/architecture/mcp_approval_and_audit_flow.md`
