---
doc_id: EPC-CK8S-S6-D1
type: epic
title: "Epic S6 Day 1: Intake для раздела управления агентами и шаблонами промптов (Issue #184)"
status: completed
owner_role: PM
created_at: 2026-02-25
updated_at: 2026-02-25
related_issues: [184, 185]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-25-issue-184-intake"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-25
---

# Epic S6 Day 1: Intake для раздела управления агентами и шаблонами промптов (Issue #184)

## TL;DR
- В `services/staff/web-console/src/pages/configuration/AgentsPage.vue` и `AgentDetailsPage.vue` раздел `Agents` реализован как scaffold: mock-данные, локальные шаблоны и TODO на backend-интеграцию.
- В staff OpenAPI (`services/external/api-gateway/api/server/api.yaml`) сейчас нет endpoint-ов для списка/настроек агентов, prompt templates, diff/effective preview и audit history.
- При этом архитектурный контур уже предполагает эти сущности (`agents`, `agent_policies`, `prompt_templates`, `agent_sessions`, `flow_events`) и policy requirements (locale fallback, audit trail, role-specific templates).

## Problem Statement
### As-Is
- UI показывает демонстрационный список агентов и mock editor шаблонов, сохранение не вызывает backend.
- История изменений и audit вкладка в деталях агента также являются заглушками.
- Contract-first контур staff API не содержит операций для управления агентами и шаблонами.

### To-Be
- Раздел `Agents` работает на реальных данных проекта/платформы.
- Для шаблонов промптов доступен полный lifecycle: список, просмотр, редактирование, diff, effective preview, версионирование, audit history.
- Все изменения проходят через typed API/DTO, policy guardrails и аудит в `flow_events`/`agent_sessions`.

## Brief
- Бизнес-ценность: убрать ручное управление prompt seeds в коде и дать Owner управляемый staff контур настройки агентов.
- Пользовательская ценность: предсказуемое поведение агентов по ролям/локалям, прозрачная история изменений и безопасный rollout.
- Техническая ценность: согласовать UI, OpenAPI и доменную модель без дрейфа между слоями.

## MVP Scope
### In scope (MVP)
- Agents registry в staff UI/API (list/details/settings baseline).
- Prompt templates lifecycle (`work/revise`, `ru/en`, diff/effective preview, active version marker).
- Audit/history для изменений шаблонов и параметров агентов.
- Явная трассируемость `issue -> docs -> stage artifacts -> implementation issues`.

### Out of scope (post-MVP)
- Конструктор новых custom-ролей с произвольными capability packs.
- Автоматический ML-based quality scoring шаблонов.
- Полная self-service marketplace модель prompt templates.

## Constraints
- Kubernetes-only, webhook-driven, PostgreSQL (`JSONB` + `pgvector`) и текущая stage/label policy остаются без изменений.
- Для external/staff API обязателен contract-first OpenAPI и typed DTO/casters.
- Политика prompt templates (`project override -> global override -> repo seed`, locale fallback) должна сохраняться.
- Trigger `run:*` и review gate работают по действующей label policy без обходов.

## Acceptance Criteria (Intake stage)
- [x] Зафиксирован подтвержденный разрыв между текущим UI scaffold и отсутствующими backend/API контрактами.
- [x] Определен продуктовый scope MVP для раздела `Agents` (settings + templates + history/audit).
- [x] Определены ограничения и неподвижные правила, которые нельзя нарушать на следующих stage.
- [x] Сформирован plan-handover для полного цикла `vision -> prd -> arch -> design -> plan -> dev -> doc-audit`.
- [x] Создана отдельная issue на следующий stage (`run:vision`) с инструкцией создать issue на следующий этап (`run:prd`) — `#185`.

## Декомпозиция по этапам (до doc-audit)

| Stage | Фокус | Выходной артефакт |
|---|---|---|
| Vision | Продуктовое видение и критерии успеха контуров `Agents`/`Templates` | charter + success metrics + risk register |
| PRD | Формализация функциональных сценариев и NFR | PRD + user stories + NFR |
| Arch | Границы сервисов и ownership data/API | C4 + ADR + architecture decisions |
| Design | Контракты API и модель данных для реализации | design package + API/data model |
| Plan | Delivery-план, эпики и implementation issues | execution package + linked issues |
| Dev | Реализация контуров UI/API/domain + PR | code + tests + docs sync |
| Doc-Audit | Проверка соответствия реализации документам | audit report + remediation backlog |

## Risks and Product Assumptions
- Риск: без staged-декомпозиции инициатива станет слишком широкой и потеряет управляемый DoD.
- Риск: если начать dev без vision/prd, появится drift в AC между UI, API и domain policy.
- Риск: history/audit может быть реализован частично и не покрыть регуляторный/операционный контур.
- Допущение: текущие сущности БД достаточны как стартовый baseline, а изменения схемы будут фиксироваться на стадиях arch/design.

## Stage Handover Instructions
- Следующий этап: `run:vision`.
- Созданная issue следующего этапа: `#185`.
- Обязательный артефакт следующего этапа: создать issue для `run:prd` с ссылкой на vision issue и с повторением требования “создать issue для следующего stage”.
- В конце каждой стадии до `run:plan` включительно должна создаваться новая issue для следующей стадии, чтобы сформировать последовательный backlog эпиков и implementation issues перед стартом `run:dev`.
