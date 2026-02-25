---
doc_id: EPC-CK8S-S6-D4
type: epic
title: "Epic S6 Day 4: Architecture для lifecycle управления агентами и шаблонами промптов (Issue #189)"
status: in-review
owner_role: SA
created_at: 2026-02-25
updated_at: 2026-02-25
related_issues: [184, 185, 187, 189, 195]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-25-issue-189-arch"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-25
---

# Epic S6 Day 4: Architecture для lifecycle управления агентами и шаблонами промптов (Issue #189)

## TL;DR
- Подготовлен архитектурный пакет для доменов `agents settings`, `prompt templates lifecycle`, `audit/history`.
- Зафиксированы границы сервисов, C4 container view, риски и mitigation.
- Подготовлен handover-пакет в `run:design`.

## Контекст
Продолжение цепочки S6: #184 (intake) -> #185 (vision) -> #187 (prd) -> #189 (arch).
PRD-пакет Day3 зафиксирован в Issue #187; PR #190 с PRD-артефактами смержен в `main`.

## Основные артефакты
- Архитектурный дизайн: `docs/architecture/agents_prompt_templates_lifecycle_design.md`.
- ADR: `docs/architecture/adr/ADR-0009-prompt-templates-lifecycle-and-audit.md`.
- Альтернативы: `docs/architecture/alternatives/ALT-0001-agents-prompt-templates-lifecycle.md`.

## Границы и ownership
- `api-gateway` — thin-edge validation/auth/routing.
- `control-plane` — доменная логика, data ownership и audit событий.
- `worker` — асинхронные и идемпотентные фоновые задачи.
- `web-console` — UX, без бизнес-логики.

## Риски и mitigation (архитектурный baseline)
- Конфликт параллельных правок: optimistic concurrency + `conflict` ошибки.
- Неполный audit trail: транзакция `domain write + flow_event`.
- Большие diff: server-side diff + лимиты размера/кэш.
- Drift между seed и overrides: indicator source + checksum.

## Migration и runtime impact
- На этапе `run:arch` runtime и схема БД не менялись (markdown-only change set).
- В handover для `run:design` зафиксирован обязательный миграционный пакет для `prompt_templates` (поля версий/состояний/индексов, rollback-подход).
- Для последующего `run:dev` закреплён rollout order: `migrations -> internal services -> edge services -> frontend`.

## Dependency baseline (Context7)
- Проверен `kin-openapi` (`/getkin/kin-openapi`): текущий стек достаточен для request/response validation в contract-first контуре.
- Проверен `monaco-editor` (`/microsoft/monaco-editor`): встроенный `DiffEditor` покрывает шаблонный diff use-case.
- Новые внешние библиотеки на этом этапе не требуются.

## Handover в `run:design`
- OpenAPI контракты staff API (agents/templates/audit).
- gRPC контракты для `api-gateway -> control-plane` + typed DTO/casters.
- Изменения data model и миграции (если добавляются поля/таблицы).
- UI flow и state-management для diff/preview/history.
- План observability и тестирования.

## Следующий этап
- Создан follow-up issue для `run:design`: `#195`.
- В issue `#195` зафиксирована обязательная инструкция после `run:design` создать issue следующего этапа `run:plan`.
