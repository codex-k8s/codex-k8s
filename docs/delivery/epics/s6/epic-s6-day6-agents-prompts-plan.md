---
doc_id: EPC-CK8S-S6-D6
type: epic
title: "Epic S6 Day 6: Plan для реализации lifecycle управления агентами и шаблонами промптов (Issues #197/#199)"
status: in-review
owner_role: EM
created_at: 2026-02-25
updated_at: 2026-02-25
related_issues: [184, 185, 187, 189, 195, 197, 199]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-25-issue-197-plan"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-25
---

# Epic S6 Day 6: Plan для реализации lifecycle управления агентами и шаблонами промптов (Issues #197/#199)

## TL;DR
- Day6 зафиксировал execution-package для перехода в `run:dev` на основе design-пакета Day5.
- Подготовлены quality-gates реализации, DoR/DoD и release-readiness критерии для `control-plane`, `api-gateway`, `web-console`, migrations/seed-sync.
- Создана follow-up issue `#199` для stage `run:dev` (без trigger-лейбла, лейбл ставит Owner).

## Контекст
- Stage continuity: `#184 -> #185 -> #187 -> #189 -> #195 -> #197 -> #199`.
- Входные артефакты:
  - `docs/architecture/agents_prompt_templates_lifecycle_design_doc.md`
  - `docs/architecture/agents_prompt_templates_lifecycle_api_contract.md`
  - `docs/architecture/agents_prompt_templates_lifecycle_data_model.md`
  - `docs/architecture/agents_prompt_templates_lifecycle_migrations_policy.md`
  - `docs/delivery/sprints/s6/sprint_s6_agents_prompt_management.md`
  - `docs/delivery/epics/s6/epic_s6.md`

## Scope
### In scope
- Декомпозиция `run:dev` на implementation-потоки и порядок поставки.
- Definition of Ready / Definition of Done для этапа реализации.
- Quality-gates по контрактам, миграциям, тестам, observability и release readiness.
- Подготовка handover issue `#199` и continuity-правила для следующего этапа `run:qa`.

### Out of scope
- Реализация кода и миграций в рамках `run:plan`.
- Изменение label taxonomy (`run:*`, `state:*`, `need:*`).
- Изменение архитектурных границ вне design-пакета Day5.

## Декомпозиция `run:dev` (execution roadmap)
| Поток | Priority | Инкременты | Критерий завершения |
|---|---|---|---|
| W1 Contract-first | `P0` | Обновление OpenAPI/proto, регенерация backend/frontend codegen, синхронизация typed DTO/casters | Контракты и codegen совпадают, drift отсутствует |
| W2 Data + migrations | `P0` | DDL/backfill/index hardening для `prompt_templates`, post-check инвариантов, seed bootstrap/sync `dry-run -> apply` | Active-version uniqueness подтверждён, fallback на embed seeds сохранён |
| W3 Control-plane domain | `P0` | Use-cases: update settings, create/activate version, preview/diff, history/audit; optimistic concurrency + idempotency | Доменные операции покрыты unit/integration тестами, error taxonomy соблюдён |
| W4 API-gateway edge | `P0` | HTTP handlers/validators/casters, RBAC/authn/authz checks, error mapping только на transport boundary | Thin-edge соблюден, контрактные ответы typed |
| W5 Web-console integration | `P1` | Интеграция typed API, state-flow list/details/diff/preview/history без mock данных | UI сценарии AC-01..AC-08 выполняются |
| W6 QA + observability | `P1` | Regression-пакет backend/frontend, метрики/логи/трейсы по lifecycle операциям | Есть evidence для handover в `run:qa` |
| W7 Rollout readiness | `P1` | Порядок выката `migrations -> internal -> edge -> frontend`, rollback/feature-flag готовность | Release gate подтверждён до перевода в `run:qa` |

## Quality-gates (`run:dev` entry/exit)
| Gate | Что проверяем | Критерий выхода |
|---|---|---|
| QG-S6-D6-01 Contract sync | OpenAPI/proto/codegen синхронизированы | Нет расхождений между спецификациями и transport-кодом |
| QG-S6-D6-02 Migration safety | DDL/backfill/index + post-check инвариантов | `prompt_templates` удовлетворяет active-version uniqueness |
| QG-S6-D6-03 Domain boundaries | `control-plane` владеет use-cases и БД, `api-gateway` остаётся thin-edge | Нет domain-логики в edge/frontend слоях |
| QG-S6-D6-04 UI readiness | Web-console работает на реальных typed API | Сценарии list/details/diff/preview/history проходят без mock |
| QG-S6-D6-05 QA evidence | Unit/integration/frontend regression собран | Evidence приложен в PR и готов для `run:qa` |
| QG-S6-D6-06 Release readiness | Rollout/rollback порядок и observability подтверждены | Возможен безопасный переход в `run:qa` |
| QG-S6-D6-07 Traceability continuity | `issue_map`, `requirements_traceability`, sprint/epic docs актуальны | Цепочка `#197 -> #199 -> run:qa issue` зафиксирована |

## Definition of Ready для `run:dev` (Issue #199)
- [x] Design package Day5 зафиксирован и покрывает API/data/migrations/UI.
- [x] Execution roadmap W1..W7 и приоритеты `P0/P1` сформированы.
- [x] Quality-gates QG-S6-D6-01..QG-S6-D6-07 утверждены в delivery-документации.
- [x] Handover issue `#199` создана без trigger-лейбла.
- [x] Зафиксировано continuity-правило: после `run:dev` создать issue `run:qa`.

## Definition of Done package для `run:dev`
- Контракты:
  - OpenAPI/proto обновлены, codegen артефакты синхронизированы.
- Данные:
  - миграции и backfill выполнены, post-check инвариантов пройден.
- Реализация:
  - доменные use-cases и edge/UI интеграция покрывают AC-01..AC-08 из PRD.
- Качество:
  - unit/integration/frontend tests зелёные, regression evidence зафиксирован.
- Наблюдаемость:
  - метрики/логи/трейсы для lifecycle операций доступны и привязаны к `correlation_id`.
- Governance:
  - `state:in-review` установлен на PR и Issue;
  - подготовлена follow-up issue `run:qa`.

## Blockers, риски и owner decisions
| Тип | ID | Описание | Статус |
|---|---|---|---|
| blocker | BLK-197-01 | Требуется Owner-постановка trigger-лейбла `run:dev` на issue `#199` | open |
| blocker | BLK-197-02 | Требуется подтверждение release window для migration-first rollout | open |
| risk | RSK-197-01 | Конфликт версий шаблонов при параллельных правках может увеличить `conflict` rate | monitoring |
| risk | RSK-197-02 | Несинхронный rollout OpenAPI/proto/frontend может привести к transport drift | monitoring |
| risk | RSK-197-03 | Рост latency `preview/diff` при больших шаблонах | monitoring |
| owner-decision | OD-197-01 | Применять rollout order `migrations -> internal -> edge -> frontend` без отклонений | proposed |
| owner-decision | OD-197-02 | Seed слой остаётся постоянным fallback; bootstrap `dry-run -> apply` не перезаписывает project overrides | proposed |
| owner-decision | OD-197-03 | Выход из `run:dev` допустим только при готовой issue `run:qa` и полном QA evidence | proposed |

## Context7 validation
- Подтвержден синтаксис `gh issue create/edit` и `gh pr create/edit` по `GitHub CLI manual` (`/websites/cli_github_manual`) для fallback-команд и PR-flow.
- Для реализации Day6 не требуется новая внешняя библиотека: dependency baseline остаётся `kin-openapi` + `monaco-editor` (зафиксировано в Day5 design-пакете).

## Handover в `run:dev`
- Follow-up issue: `#199`.
- Issue `#199` создана без trigger-лейбла; запуск stage выполняется Owner.
- Обязательное правило continuity:
  - после завершения `run:dev` создать issue для `run:qa`;
  - в issue `run:qa` включить summary regression evidence и явную инструкцию создать issue `run:release`.

## Связанные документы
- `docs/delivery/sprints/s6/sprint_s6_agents_prompt_management.md`
- `docs/delivery/epics/s6/epic_s6.md`
- `docs/delivery/issue_map.md`
- `docs/delivery/requirements_traceability.md`
- `docs/architecture/agents_prompt_templates_lifecycle_design_doc.md`
- `docs/architecture/agents_prompt_templates_lifecycle_api_contract.md`
- `docs/architecture/agents_prompt_templates_lifecycle_data_model.md`
- `docs/architecture/agents_prompt_templates_lifecycle_migrations_policy.md`
