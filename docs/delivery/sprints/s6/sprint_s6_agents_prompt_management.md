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
  request_id: "owner-2026-02-25-issue-184-intake"
---

# Sprint S6: Agents configuration and prompt templates lifecycle (Issue #184)

## TL;DR
- Цель спринта: перевести раздел `Configuration -> Agents` из scaffold в управляемый контур продукта с реальными данными, edit-flow шаблонов промптов и аудитом изменений.
- Стартовая точка: в UI сейчас mock-данные и TODO на интеграцию; в staff OpenAPI отсутствуют endpoint-ы `agents/prompt-templates`, хотя в data model уже есть `agents`, `agent_policies`, `prompt_templates`, `agent_sessions`, `flow_events`.
- Траектория выбрана как `new-service` (полный документный цикл до реализации), потому что инициатива затрагивает продуктовую модель, API-контракты, RBAC/audit и жизненный цикл stage/label policy.

## Scope спринта
### In scope
- Intake -> Vision -> PRD -> Architecture -> Design -> Plan -> Dev -> QA -> Release -> Postdeploy -> Ops.
- Отдельный контур `run:doc-audit` после `run:dev` для проверки трассируемости `issue -> docs -> implementation`.
- Формирование последовательных epics и GitHub issues для каждого stage (без пропуска этапов).

### Out of scope
- Внедрение новых ролей агентов вне утвержденного system roster.
- Изменение базовой taxonomy labels (`run:*`, `state:*`, `need:*`) вне конкретных требований этой инициативы.
- Изменение Kubernetes-only и webhook-driven ограничений платформы.

## План этапов и handover

| Stage | Основной артефакт | Целевая роль | Правило выхода |
|---|---|---|---|
| Intake (`#184`) | Problem/Brief/Scope/Constraints + acceptance baseline | `pm` | Owner review intake-пакета и создана issue следующего этапа |
| Vision | Project charter + success metrics + риск-рамка | `pm` + `em` | Зафиксирован vision baseline и создана issue PRD |
| PRD | PRD + user stories + NFR draft | `pm` + `sa` | Подтверждены AC/NFR и создана issue Architecture |
| Architecture | C4 + ADR + boundaries | `sa` | Подтверждены границы и создана issue Design |
| Design | API/data model/design package | `sa` + `qa` | Подтвержден design пакет и создана issue Plan |
| Plan | Delivery plan + epics + DoD | `em` + `km` | Подготовлен execution package и issue Dev |
| Dev | Реализация + PR + docs sync | `dev` | PR готов, review gate пройден |
| Doc Audit | Аудит docs/traceability/checklists | `km` + `reviewer` | Закрыт drift и сформирован post-dev improvement backlog |

## Quality gates (S6 governance)

| Gate | Что проверяем | Статус |
|---|---|---|
| QG-S6-01 Intake completeness | Problem/Brief/Scope/Constraints и AC зафиксированы с анализом фактического As-Is | passed (Issue #184) |
| QG-S6-02 Stage continuity | Для следующего этапа создана отдельная issue с инструкцией создать issue следующего stage | passed (`#185`) |
| QG-S6-03 Traceability | Обновлены `issue_map`, `requirements_traceability`, sprint/epic индексы | in-progress |
| QG-S6-04 Policy compliance | Изменения ограничены markdown, без нарушения stage/label policy | in-progress |

## Stage acceptance baseline (Intake -> Vision)
- [x] Подтверждено, что текущий UI раздел `Agents` работает как scaffold и не подключен к backend.
- [x] Зафиксирован продуктовый масштаб инициативы: настройки агентов + prompt templates + audit/history.
- [x] Выбрана траектория `new-service` (полный pipeline, без fast-track).
- [x] Создана issue на `run:vision` с обязательной инструкцией создать issue на `run:prd` (`#185`).

## Риски и допущения
- Риск: смешение scope между настройками агентов, prompt policy и runtime observability может размыть MVP-инкремент.
- Риск: отсутствие typed API contract для `agents/templates/audit` приведет к расхождению UI и backend.
- Риск: без явной матрицы RBAC/audit есть вероятность появления небезопасного edit-flow шаблонов.
- Допущение: существующая модель БД (`agents`, `agent_policies`, `prompt_templates`, `agent_sessions`, `flow_events`) остается базой для проектирования API и UI.

## Handover в следующий этап
- Следующий stage после intake: `run:vision`.
- Follow-up issue следующего этапа: `#185`.
- Отдельная GitHub issue создается в рамках текущего intake-run и должна включать явный пункт: по завершению vision создать issue для `run:prd` с аналогичной цепочкой.
