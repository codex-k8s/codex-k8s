---
doc_id: SPR-CK8S-0006
type: sprint-plan
title: "Sprint S6: Agents configuration and prompt templates lifecycle (Issue #184)"
status: in-progress
owner_role: PM
created_at: 2026-02-25
updated_at: 2026-02-27
related_issues: [184, 185, 187, 189, 195, 197, 199, 201, 216]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-25-issue-187-prd"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-25
---

# Sprint S6: Agents configuration and prompt templates lifecycle (Issue #184)

## TL;DR
- Цель спринта: перевести раздел `Configuration -> Agents` из scaffold в управляемый контур продукта с реальными данными, lifecycle шаблонов промптов и аудитом изменений.
- Базовый As-Is разрыв зафиксирован intake-этапом: UI работает на mock-данных, а staff OpenAPI пока не покрывает `agents/prompt-templates/audit`.
- На текущий момент chain continuity доведён до `run:release`: `#184 -> #185 -> #187 -> #189 -> #195 -> #197 -> #199 -> #201 -> #216`.

## Scope спринта
### In scope
- Intake -> Vision -> PRD -> Architecture -> Design -> Plan -> Dev -> QA -> Release -> Postdeploy -> Ops.
- Отдельный контур `run:doc-audit` после `run:dev` для проверки трассируемости `issue -> docs -> implementation`.
- Формирование последовательных epics и GitHub issues для каждого stage без пропуска этапов.
- Follow-up issue создаются без `run:*`-лейблов; trigger-лейбл на запуск следующего stage ставит Owner после review.

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
| Design (`#195`) | API/data model/design package | `sa` + `qa` | Подтвержден design пакет и создана issue Plan |
| Plan (`#197`) | Delivery plan + epics + DoD | `em` + `km` | Подготовлен execution package и issue Dev |
| Dev (`#199`) | Реализация + PR + docs sync | `dev` | PR `#202` готов, создана issue `run:qa` |
| QA (`#201`) | Acceptance/regression evidence + readiness decision | `qa` | QA gate пройден, создана issue `#216` для `run:release` |
| Release (`#216`) | Release plan + release notes + rollback plan | `em` + `sre` | Release gate пройден и создана issue `run:postdeploy` |
| Doc Audit | Аудит docs/traceability/checklists | `km` + `reviewer` | Закрыт drift и сформирован post-dev improvement backlog |

## Quality gates (S6 governance)

| Gate | Что проверяем | Статус |
|---|---|---|
| QG-S6-01 Intake completeness | Problem/Brief/Scope/Constraints и AC зафиксированы с анализом фактического As-Is | passed (Issue #184) |
| QG-S6-02 Stage continuity | Для каждого следующего этапа создана issue без trigger-лейбла с обязательной инструкцией continuity | passed (`#185` -> `#187` -> `#189` -> `#195` -> `#197` -> `#199`) |
| QG-S6-03 Vision baseline | Mission/KPI, границы MVP/Post-MVP и риск-рамка зафиксированы для входа в PRD | passed (`#185`) |
| QG-S6-04 PRD completeness | Подготовлен PRD-пакет с FR/AC/NFR-draft и user stories | passed (`#187`, PR `#190` merged) |
| QG-S6-05 Architecture handover | Архитектурный пакет, ADR/альтернативы и design continuity issue сформированы | passed (`#189` -> `#195`) |
| QG-S6-06 Traceability | Обновлены `issue_map`, `requirements_traceability`, sprint/epic документы | passed |
| QG-S6-07 Policy compliance | Изменения ограничены markdown без нарушения stage/label policy | passed |
| QG-S6-08 Design handover | Подготовлен design package и создан follow-up issue `run:plan` | passed (`#195` -> `#197`) |
| QG-S6-09 Plan handover | Подготовлен execution package `run:dev` (quality-gates + DoD + risks) и создана issue реализации | passed (`#197` -> `#199`) |
| QG-S6-10 QA readiness | Выполнены QA artifacts + regression evidence + решение readiness | passed (`#201` -> `#216`) |

## Stage acceptance progress (Intake -> Vision -> PRD -> Arch -> Design -> Plan)
- [x] Подтверждено, что текущий UI раздел `Agents` работает как scaffold и не подключен к backend (`#184`).
- [x] Зафиксирован продуктовый масштаб инициативы: настройки агентов + prompt templates + audit/history (`#184`).
- [x] Зафиксированы vision baseline-решения по mission/KPI и границам MVP/Post-MVP (`#185`).
- [x] Утвержден PRD-документ с FR/AC и NFR-draft для handover в архитектуру (`#187`).
- [x] Создана issue `#189` для stage `run:arch` без trigger-лейбла, с обязательной инструкцией создать issue `run:design`.
- [x] Подготовлен архитектурный пакет (`#189`) и создана issue `#195` для stage `run:design`.
- [x] Подготовлен design package (`#195`) и создана issue `#197` для stage `run:plan`.
- [x] Подготовлен execution package (`#197`) и создана issue `#199` для stage `run:dev`.
- [x] Реализован `run:dev` пакет (`#199`) и создана issue `#201` для stage `run:qa`.
- [x] Выполнен `run:qa` пакет (`#201`) и создана issue `#216` для stage `run:release`.

## Риски и допущения
- Риск: смешение scope между настройками агентов, prompt policy и runtime observability может размыть MVP-инкремент.
- Риск: отсутствие typed API contract для `agents/templates/audit` приведет к расхождению UI и backend.
- Риск: без явной стратегии version-locking возможны конфликтующие правки шаблонов.
- Допущение: существующая модель БД (`agents`, `agent_policies`, `prompt_templates`, `agent_sessions`, `flow_events`) остается базой для архитектурной проработки.

## Handover в следующий этап
- Актуальная цепочка stage-issues: `#184 (intake) -> #185 (vision) -> #187 (prd) -> #189 (arch) -> #195 (design) -> #197 (plan) -> #199 (dev) -> #201 (qa) -> #216 (release)`.
- Для issue `#216` зафиксировано обязательное правило: после `run:release` создать issue следующего этапа `run:postdeploy`.
