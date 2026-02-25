---
doc_id: EPC-CK8S-S6-D3
type: epic
title: "Epic S6 Day 3: PRD для lifecycle управления агентами и шаблонами промптов (Issue #187)"
status: in-review
owner_role: PM
created_at: 2026-02-25
updated_at: 2026-02-25
related_issues: [184, 185, 187, 189]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-25-issue-187-prd"
---

# Epic S6 Day 3: PRD для lifecycle управления агентами и шаблонами промптов (Issue #187)

## TL;DR
- На PRD-этапе зафиксированы детальные функциональные требования и acceptance criteria для контуров `agents settings`, `prompt templates lifecycle`, `audit/history`.
- Подготовлен отдельный PRD-артефакт: `docs/delivery/epics/s6/prd-s6-day3-agents-prompts-lifecycle.md`.
- Сформирован handover в stage `run:arch` через follow-up issue `#189` без trigger-лейбла (лейбл ставит Owner) с обязательной инструкцией создать issue для `run:design` после завершения архитектурного этапа.

## Priority
- `P0`.

## Контекст
- Intake baseline зафиксирован в Issue `#184` и Day1-эпике (`epic-s6-day1-agents-prompts-intake.md`).
- Vision baseline зафиксирован в Issue `#185`: mission, KPI, границы MVP/Post-MVP и риск-рамка.
- Текущий этап `run:prd` для Issue `#187` формализует требования и критерии приемки перед архитектурной проработкой.

## Scope
### In scope
- Детализированные FR для staff-сценариев управления агентами и шаблонами промптов.
- Acceptance criteria в формате Given/When/Then.
- NFR-draft для handover в `run:arch`.
- Трассируемый переход `#184 -> #185 -> #187 -> #189`.

### Out of scope
- Реализация UI/backend кода.
- Финальная архитектурная декомпозиция и ADR на этом этапе.
- Изменение базовой taxonomy labels (`run:*`, `state:*`, `need:*`).

## PRD package
- Канонический PRD-документ Day3:
  - `docs/delivery/epics/s6/prd-s6-day3-agents-prompts-lifecycle.md`
- Содержание пакета:
  - 12 функциональных требований (`FR-187-*`);
  - 8 acceptance сценариев (`AC-01..AC-08`);
  - 6 NFR-draft требований (`NFR-187-*`);
  - user stories и архитектурные вопросы для следующего этапа.

## Acceptance criteria (Day3 stage)
- [x] Утвержден PRD-документ с формализованными FR/AC и пользовательскими сценариями.
- [x] Зафиксирован NFR-draft, достаточный для архитектурной проработки `run:arch`.
- [x] Подтверждена трассируемость `#184 -> #185 -> #187` в `issue_map` и `requirements_traceability`.
- [x] Подготовлен handover-пакет в `run:arch`.
- [x] Создана отдельная follow-up issue `#189` для следующего этапа без trigger-лейбла (лейбл ставит Owner).

## NFR-draft summary
| Категория | Черновое требование |
|---|---|
| Security | Все mutation-операции по агентам и шаблонам проходят через проектный RBAC без обходных путей. |
| Auditability | Для каждой mutation-операции фиксируются `actor`, `correlation_id`, `issue/pr context`, причина изменения. |
| Observability | Есть метрики и журнал ошибок по операциям `publish`, `rollback`, `effective preview`, `diff`. |
| UX responsiveness | Целевые бюджеты p95: list/details <= 2s, diff <= 2s, effective preview <= 3s. |
| Reliability | Операции publish/rollback должны быть идемпотентны и безопасны при повторных запросах. |
| Localization | Поддерживается `ru/en` и детерминированный locale fallback без потери трассируемости источника шаблона. |

## Quality gates (handover to architecture)
| Gate | Что проверяем | Статус |
|---|---|---|
| QG-S6-D3-01 PRD completeness | FR/AC/NFR покрывают scope Day3 без смешения с реализацией | passed |
| QG-S6-D3-02 Stage continuity | Создана issue следующего этапа с обязательной инструкцией о следующем handover | passed (`#189`) |
| QG-S6-D3-03 Traceability | Обновлены sprint/epic, issue_map и requirements_traceability | passed |
| QG-S6-D3-04 Policy compliance | Изменены только markdown-артефакты | passed |

## Handover в `run:arch`
- Следующий этап: `run:arch`.
- Follow-up issue: `#189`.
- Trigger-лейбл `run:arch` на issue `#189` ставит Owner.
- Обязательные выходы архитектурного этапа:
  - сервисные границы и ownership по контурам `agents/templates/audit`;
  - C4/ADR решения по transport/data/policy boundaries;
  - mitigation-план по рискам конкурентного редактирования, audit полноты и latency budgets;
  - отдельная issue на stage `run:design` по завершении `run:arch` (без trigger-лейбла при создании).

## Трассируемость
- Intake: `#184` -> `docs/delivery/epics/s6/epic-s6-day1-agents-prompts-intake.md`.
- Vision: `#185` (handover baseline для PRD).
- PRD: `#187` -> `docs/delivery/epics/s6/prd-s6-day3-agents-prompts-lifecycle.md`.
- Architecture (next): `#189`.

## Открытые риски и допущения
- Риск: несогласованность версии/источника шаблона при параллельных изменениях без явной strategy locking.
- Риск: неполная audit-трасса при частичных отказах publish/rollback.
- Риск: рост latency effective preview при сложной цепочке fallback `project -> global -> seed`.
- Допущение: существующая доменная модель (`agents`, `agent_policies`, `prompt_templates`, `agent_sessions`, `flow_events`) покрывает стартовый архитектурный baseline.
