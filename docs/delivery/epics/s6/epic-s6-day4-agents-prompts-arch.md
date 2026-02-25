---
doc_id: EPC-CK8S-S6-D4
type: epic
title: "Epic S6 Day 4: Architecture для lifecycle управления агентами и шаблонами промптов (Issue #189)"
status: in-progress
owner_role: SA
created_at: 2026-02-25
updated_at: 2026-02-25
related_issues: [184, 185, 187, 189]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-25-issue-189-arch"
---

# Epic S6 Day 4: Architecture для lifecycle управления агентами и шаблонами промптов (Issue #189)

## TL;DR
- Подготовлен архитектурный пакет для доменов `agents settings`, `prompt templates lifecycle`, `audit/history`.
- Зафиксированы границы сервисов, C4 container view, риски и mitigation.
- Подготовлен handover-пакет в `run:design`.

## Контекст
Продолжение цепочки S6: #184 (intake) -> #185 (vision) -> #187 (prd) -> #189 (arch).
PRD-пакет Day3 оформлен в PR #190 (ожидает merge) и служит входом для архитектурных решений.

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

## Handover в `run:design`
- OpenAPI контракты staff API (agents/templates/audit).
- gRPC контракты для `api-gateway -> control-plane` + typed DTO/casters.
- Изменения data model и миграции (если добавляются поля/таблицы).
- UI flow и state-management для diff/preview/history.
- План observability и тестирования.

## Следующий этап
- `run:design` (создать follow-up issue после завершения архитектуры).
