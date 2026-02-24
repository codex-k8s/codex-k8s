---
doc_id: DSG-CK8S-0119
type: design
title: "Issue #119 — Design: E2E A+B core lifecycle and review-revise regression"
status: draft
owner_role: SA
created_at: 2026-02-24
updated_at: 2026-02-24
related_issues: [119, 118]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-24-issue-119-design"
---

# Design: Issue #119 — E2E A+B

## TL;DR
- Цель design-этапа: зафиксировать сервисные границы, контрактные инварианты и модель evidence для сценариев A1/A2/A3 и B1/B2/B3.
- Изменений API/схемы БД не требуется: используется текущий stage/label/runtime контур.
- Основной результат: единый design-пакет для воспроизводимой проверки и handover evidence в Issue #118.

## Структура design-пакета
- Индекс: `docs/delivery/design/issue-119/README.md`
- Формальная матрица трассируемости: `docs/delivery/design/issue-119/traceability_matrix.md`
- Контрактные и migration/data артефакты:
  - `docs/delivery/design/issue-119/api_contract.md`
  - `docs/delivery/design/issue-119/data_model.md`
  - `docs/delivery/design/issue-119/migration_policy.md`

## Scope
- In scope:
  - архитектурная декомпозиция сценариев A+B по слоям `external -> internal -> jobs -> github`;
  - контрактные инварианты review-driven revise;
  - структура evidence и правила трассируемости.
- Out of scope:
  - изменения кода, миграций, runtime policy, OpenAPI/proto;
  - сценарии C/D/E/F из master plan.

## Сервисные границы

| Слой | Ответственность в issue #119 | Что проверяем |
|---|---|---|
| `services/external/api-gateway` | thin-edge webhook ingress/callback routing | нет доменной логики на edge |
| `services/internal/control-plane` | stage resolver, label policy, audit events, run status | корректный revise path для B1/B3, ambiguity path для B2 |
| `services/jobs/worker` | run orchestration, transitions, wait states | запуск/не запуск revise согласно policy |
| `services/jobs/agent-runner` | исполнение role prompt и публикация progress/evidence | соблюдение markdown-only и status cadence |
| GitHub (issue/pr labels) | операционный источник состояния review gate | ожидаемые `run:*`/`state:*`/`need:*` transitions |

## Дизайн сценариев A+B

### A1-A3 (core lifecycle)
1. `run:intake -> run:vision -> run:prd -> run:arch -> run:design -> run:plan`.
2. `run:dev` создает/обновляет PR, затем `run:dev:revise` для итерации.
3. `run:qa -> run:release -> run:postdeploy -> run:ops`.

Инварианты:
- на issue одновременно только один trigger `run:*`;
- после артефактного шага ставится `state:in-review`;
- все переходы фиксируются в `flow_events`.

### B1-B3 (review-driven revise)
1. B1: при `changes_requested` и однозначном stage запускается `run:<stage>:revise`.
2. B2: при ambiguity stage revise-run не стартует, выставляется `need:input`.
3. B3: профиль `[ai-model-*]`/`[ai-reasoning-*]` резолвится sticky по policy chain.

Инварианты:
- stage resolver детерминирован (`PR -> Issue -> run context -> flow_events`);
- ambiguity не приводит к silent fallback;
- коммуникация и feedback остаются на русском языке.

## Матрица трассируемости (scenario -> требования -> evidence)

| Scenario | Требования | Архитектурный источник | Evidence target |
|---|---|---|---|
| A1 | FR-028, FR-033 | `docs/product/stage_process_model.md` | `Issue #118` + transitions в `flow_events` |
| A2 | FR-033, FR-052 | `docs/architecture/adr/ADR-0006-review-driven-revise-and-next-step-ux.md` | PR #120 + run/service comments |
| A3 | FR-028, NFR-018 | `docs/product/stage_process_model.md` | цепочка stage completion без P0/P1 |
| B1 | FR-052 | `docs/architecture/mcp_approval_and_audit_flow.md` | факт запуска `run:<stage>:revise` |
| B2 | FR-052, NFR-018 | `docs/product/labels_and_trigger_policy.md` | `need:input` + отсутствие revise-run |
| B3 | FR-035, FR-052 | `docs/architecture/adr/ADR-0006-review-driven-revise-and-next-step-ux.md` | сохранение model/reasoning profile |

## Trade-offs
- Выбор: не добавлять новые контракты/схему БД в рамках `run:design`.
- Плюс: минимальный риск регрессии перед MVP gate.
- Минус: повторное использование существующих audit-событий без нового специализированного event type для issue #119.

## Риски и меры
- Риск: рассинхрон labels между PR/Issue.
  - Мера: проверка B2 как обязательный fail-safe.
- Риск: неполный evidence bundle.
  - Мера: обязательные поля evidence фиксированы в `data_model.md`.
- Риск: ручные обходы pipeline.
  - Мера: traceability обновляется синхронно в `issue_map` и master plan.

## Влияние на runtime и миграции
- Runtime: только использование существующего orchestration path.
- Data schema: без изменений.
- Миграции: не требуются.

## План верификации и handover
1. Сопоставить A1..B3 из `docs/delivery/e2e_mvp_master_plan.md` с фактическими transitions.
2. Для каждого сценария зафиксировать `run_id`, `correlation_id`, `expected vs actual`.
3. Публиковать итоговый evidence bundle в Issue #118.
4. Проверить синхронизацию ссылок в `docs/delivery/issue_map.md` и `traceability_matrix.md`.

## Acceptance Criteria (design stage)
- [ ] Зафиксированы сервисные границы и ответственность по A+B.
- [ ] Зафиксированы контрактные инварианты B1/B2/B3.
- [ ] Зафиксирован формат evidence и traceability links для Issue #118.
- [ ] Подтверждено отсутствие необходимости API/DB миграций.

## Матрица трассируемости
- Детальная матрица `требование -> сценарий -> артефакт -> evidence` вынесена в:
  `docs/delivery/design/issue-119/traceability_matrix.md`.

## Связанные артефакты
- `docs/delivery/e2e_mvp_master_plan.md`
- `docs/delivery/prd/issue-119/prd.md`
- `docs/architecture/adr/ADR-0006-review-driven-revise-and-next-step-ux.md`
- `docs/architecture/mcp_approval_and_audit_flow.md`

## Апрув
- request_id: owner-2026-02-24-issue-119-design
- Решение: pending
- Комментарий:
