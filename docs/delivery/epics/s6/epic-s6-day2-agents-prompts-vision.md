---
doc_id: EPC-CK8S-S6-D2
type: epic
title: "Epic S6 Day 2: Vision для lifecycle управления агентами и шаблонами промптов (Issues #185/#187)"
status: in-review
owner_role: PM
created_at: 2026-02-25
updated_at: 2026-02-25
related_issues: [184, 185, 187]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-25-issue-185-vision"
---

# Epic S6 Day 2: Vision для lifecycle управления агентами и шаблонами промптов (Issues #185/#187)

## TL;DR
- Для Issue #185 сформирован vision-пакет: mission, целевые outcomes, метрики успеха, MVP/Post-MVP границы, риск-рамка.
- Зафиксирован критерий готовности к `run:prd`: PRD должен детализировать FR/AC/NFR для трёх контуров (`agent settings`, `prompt templates lifecycle`, `audit/history`) без расширения baseline label taxonomy.
- Создана follow-up issue на следующий этап `run:prd`: #187 (без trigger-лейбла; лейбл ставит Owner) с обязательной инструкцией создать issue на `run:arch` после завершения PRD.

## Priority
- `P0`.

## Vision charter

### Mission statement
Сделать управление агентами и шаблонами промптов управляемым продуктовым контуром в staff UI/API вместо scaffold-подхода, чтобы Owner и platform-команда могли безопасно и предсказуемо управлять поведением агентов без ручных правок seed-файлов в коде.

### Цели и ожидаемые результаты
1. Превратить `Configuration -> Agents` из mock/scaffold в контрактный контур с typed API, lifecycle операций и аудитом изменений.
2. Обеспечить детерминированный lifecycle prompt templates (`work/revise`, locale-aware, diff/effective preview, versioning/activation) с прозрачным rollback-path.
3. Закрепить операционную наблюдаемость изменений (кто/что/когда/почему) через audit trail и traceability до delivery-артефактов.

### Пользователи и стейкхолдеры
- Основные пользователи: Owner, platform admins, PM/EM/KM (governance), Dev/QA/SRE (операционный контур run-циклов).
- Стейкхолдеры: команда control-plane, staff web-console, reviewer и security/governance контур.
- Владелец решения: Owner.

### Продуктовые принципы и ограничения
- Без изменения неподвижных рамок: Kubernetes-only, webhook-driven orchestration, PostgreSQL (`JSONB` + `pgvector`), MCP policy/audit.
- Для staff/external API сохраняется contract-first OpenAPI.
- Любой change-flow по agents/templates проходит через typed contracts и аудит (`flow_events`, `agent_sessions`, `links`).
- В рамках `run:vision` допускаются только markdown-изменения.

## Scope boundaries

### MVP scope
- Agents settings baseline:
  - список и карточка агента;
  - управляемые параметры policy/runtime в рамках утверждённых системных ролей.
- Prompt templates lifecycle:
  - list/view/edit;
  - diff текущей и новой версии;
  - effective preview с учетом locale fallback и policy resolve;
  - activation marker и rollback к предыдущей версии.
- Audit/history:
  - история изменений шаблонов и настроек агента;
  - корреляция изменений с run/session evidence.

### Post-MVP scope (не в текущем цикле)
- Конструктор custom-ролей с произвольными capability packs.
- Автоматический quality scoring prompt templates на базе ML.
- Marketplace-шаблоны и межпроектный self-service exchange.

## Success metrics

### North Star
| ID | Метрика | Определение | Источник | Целевое значение |
|---|---|---|---|---|
| NSM-01 | Managed prompt adoption | Доля agent-runs, где `template.source` = `project_override` или `global_override` (без fallback на repo seed) | `agent_sessions.session_json` + `flow_events` | >= 80% в течение 30 дней после MVP rollout |

### Supporting metrics
| ID | Метрика | Определение/формула | Источник | Цель |
|---|---|---|---|---|
| PM-01 | Template change lead time | P50 времени от запроса изменения шаблона до активации новой версии | `flow_events` + `prompt_templates` history | <= 30 минут |
| PM-02 | Preview correctness rate | Доля изменений, принятых без rollback в первые 48 часов после публикации | `prompt_templates` history + incidents register | >= 95% |
| OPS-01 | Audit completeness | Доля изменений с полным набором `actor`, `correlation_id`, `source`, `diff summary` | `flow_events` + `links` | 100% |
| OPS-02 | Effective preview latency | p95 времени ответа endpoint preview на staff API | API metrics/logs | <= 2 секунды |
| GOV-01 | Traceability freshness | Доля stage-изменений, отражённых в `issue_map` и `requirements_traceability` в день изменения | docs updates + PR evidence | 100% |

### Guardrails (ранние сигналы)
- GR-01: если при `run:dev` не сформирован typed contract для `agents/templates/audit`, старт реализации блокируется.
- GR-02: если audit completeness < 100%, stage не переводится в `done`.
- GR-03: если fallback на repo seed > 40% после двух спринтов, требуется rethink scope и корректировка rollout-плана.

## Риски и продуктовые допущения
| Тип | ID | Описание | Митигирующее действие | Статус |
|---|---|---|---|---|
| risk | RSK-185-01 | Размывание MVP из-за попытки одновременно закрыть settings/templates/audit + расширенные UX-фичи | Жёстко отделить Post-MVP scope и зафиксировать стоп-факторы в PRD | open |
| risk | RSK-185-02 | Drift между UI и backend при отсутствии contract-first детализации до `run:dev` | Вынести OpenAPI/DTO требования в PRD + Architecture/Design gates | open |
| risk | RSK-185-03 | Недостаточная операционная прозрачность при частичном аудите изменений | Ввести обязательный KPI `OPS-01=100%` как gate для этапов после `run:dev` | open |
| assumption | ASM-185-01 | Текущие сущности `agents`, `agent_policies`, `prompt_templates`, `agent_sessions`, `flow_events` достаточны как baseline для PRD | Проверить на `run:arch` и зафиксировать миграционные гэпы только при доказанной необходимости | accepted |
| assumption | ASM-185-02 | Existing role roster покрывает MVP без внедрения custom-role constructor | Ограничить scope PRD системными ролями и policy-моделью baseline | accepted |

## Readiness criteria для `run:prd`
- [x] Mission и expected value сформулированы по трём контурам (`settings`, `templates lifecycle`, `audit/history`).
- [x] Метрики успеха и guardrails формализованы (product + operational).
- [x] Границы MVP/Post-MVP и стоп-факторы зафиксированы.
- [x] Риски и допущения оформлены для handover в PRD.
- [x] Создана отдельная issue следующего этапа `run:prd` (#187) без trigger-лейбла.

## Acceptance criteria (Issue #185)
- [x] Утвержден vision-документ с четкой проблемой, целевой аудиторией и expected value.
- [x] Утверждены измеримые метрики успеха (product + operational).
- [x] Зафиксированы границы MVP и ключевые риски/допущения для PRD.
- [x] Подготовлен handover-пакет в `run:prd`.

## Handover в следующий этап
- Следующий stage: `run:prd`.
- Follow-up issue: #187.
- Trigger-лейбл `run:prd` на issue #187 ставит Owner.
- Обязательное условие для #187: в конце PRD-stage создать issue для stage `run:arch` без trigger-лейбла с ссылками на #184, #185 и #187, а также с инструкцией создать issue следующего этапа (`run:design`) после закрытия `run:arch`.

## Связанные документы
- `docs/delivery/sprints/s6/sprint_s6_agents_prompt_management.md`
- `docs/delivery/epics/s6/epic_s6.md`
- `docs/delivery/epics/s6/epic-s6-day1-agents-prompts-intake.md`
- `docs/delivery/issue_map.md`
- `docs/delivery/requirements_traceability.md`
- `docs/product/requirements_machine_driven.md`
- `docs/architecture/data_model.md`
- `docs/architecture/prompt_templates_policy.md`
- `docs/architecture/api_contract.md`
