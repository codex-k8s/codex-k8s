---
doc_id: SPR-CK8S-0006
type: sprint-plan
title: "Sprint S6: Agents configuration and prompt templates lifecycle (Issue #184)"
status: in-progress
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

# Sprint S6: Agents configuration and prompt templates lifecycle (Issue #184)

## TL;DR
- Цель спринта: перевести раздел `Configuration -> Agents` из scaffold в управляемый контур продукта с реальными данными, lifecycle шаблонов промптов и аудитом изменений.
- Базовый As-Is разрыв зафиксирован intake-этапом: UI работает на mock-данных, а staff OpenAPI пока не покрывает `agents/prompt-templates/audit`.
- На текущий момент цепочка `intake -> vision -> prd` завершена документарно и передана в `run:arch`.

## Scope спринта
### In scope
- Intake -> Vision -> PRD -> Architecture -> Design -> Plan -> Dev -> QA -> Release -> Postdeploy -> Ops.
- Отдельный контур `run:doc-audit` после `run:dev` для проверки трассируемости `issue -> docs -> implementation`.
- Формирование последовательных epics и GitHub issues для каждого stage без пропуска этапов.

### Out of scope
- Внедрение новых ролей агентов вне утвержденного system roster.
- Изменение базовой taxonomy labels (`run:*`, `state:*`, `need:*`) вне требований инициативы.
- Изменение Kubernetes-only и webhook-driven ограничений платформы.

## План этапов и handover

| Stage | Основной артефакт | Целевая роль | Правило выхода |
|---|---|---|---|
| Intake (`#184`) | Problem/Brief/Scope/Constraints + acceptance baseline | `pm` | Owner review intake-пакета и создана issue следующего этапа |
| Vision (`#185`) | Project charter + success metrics + риск-рамка | `pm` + `em` | Зафиксирован vision baseline и создана issue PRD |
| PRD (`#187`) | PRD + user stories + NFR draft | `pm` + `sa` | Подтверждены AC/NFR и создана issue Architecture |
| Architecture (`#189`) | C4 + ADR + boundaries | `sa` | Подтверждены границы и создана issue Design |
| Design | API/data model/design package | `sa` + `qa` | Подтвержден design пакет и создана issue Plan |
| Plan | Delivery plan + epics + DoD | `em` + `km` | Подготовлен execution package и issue Dev |
| Dev | Реализация + PR + docs sync | `dev` | PR готов, review gate пройден |
| Doc Audit | Аудит docs/traceability/checklists | `km` + `reviewer` | Закрыт drift и сформирован post-dev improvement backlog |

## Quality gates (S6 governance)

| Gate | Что проверяем | Статус |
|---|---|---|
| QG-S6-01 Intake completeness | Problem/Brief/Scope/Constraints и AC зафиксированы с анализом фактического As-Is | passed (Issue #184) |
| QG-S6-02 Stage continuity (vision) | Для следующего этапа после intake создана issue с обязательной инструкцией | passed (`#185`) |
| QG-S6-03 Vision baseline | Mission/KPI, границы MVP/Post-MVP и риск-рамка зафиксированы для входа в PRD | passed (`#185`) |
| QG-S6-04 PRD completeness | Подготовлен PRD-пакет с FR/AC/NFR-draft и user stories | passed (`#187`) |
| QG-S6-05 Traceability | Обновлены `issue_map`, `requirements_traceability`, sprint/epic документы | passed |
| QG-S6-06 Policy compliance | Изменения ограничены markdown без нарушения stage/label policy | passed |

## Stage acceptance progress (Intake -> Vision -> PRD)
- [x] Подтверждено, что текущий UI раздел `Agents` работает как scaffold и не подключен к backend (`#184`).
- [x] Зафиксирован продуктовый масштаб инициативы: настройки агентов + prompt templates + audit/history (`#184`).
- [x] Зафиксированы vision baseline-решения по mission/KPI и границам MVP/Post-MVP (`#185`).
- [x] Утвержден PRD-документ с FR/AC и NFR-draft для handover в архитектуру (`#187`).
- [x] Создана issue следующего этапа `run:arch` с обязательной инструкцией создать issue `run:design` (`#189`).

## Риски и допущения
- Риск: смешение scope между настройками агентов, prompt policy и runtime observability может размыть MVP-инкремент.
- Риск: отсутствие typed API contract для `agents/templates/audit` приведет к расхождению UI и backend.
- Риск: без явной стратегии version-locking возможны конфликтующие правки шаблонов.
- Допущение: существующая модель БД (`agents`, `agent_policies`, `prompt_templates`, `agent_sessions`, `flow_events`) остается базой для архитектурной проработки.

## Handover в следующий этап
- Следующий stage после PRD: `run:arch`.
- Follow-up issue следующего этапа: `#189`.
- Обязательный выход `run:arch`: создать issue для `run:design` с цепочкой ссылок `#184`, `#185`, `#187`, `#189` и явной инструкцией создать следующий stage issue после design (`run:plan`).
