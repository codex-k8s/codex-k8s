---
doc_id: IDX-CK8S-DOCS-0001
type: docs-index
title: "codex-k8s Documentation Index"
status: in-review
owner_role: KM
created_at: 2026-03-11
updated_at: 2026-03-11
related_issues: [318, 320]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-11-docs-index"
---

# Documentation Index

## TL;DR
- `docs/index.md` — канонический корневой навигатор по проектной документации.
- Source-of-truth документы остаются в доменах `product`, `architecture`, `delivery`, `ops`.
- Инициативные и handover-пакеты размещаются во вложенных доменных подпапках, а `docs/templates/` используется только как каталог шаблонов.

## Доменные каталоги

| Каталог | Назначение | Канонический индекс |
|---|---|---|
| `docs/product/` | Требования, ограничения, роли, label/stage policy | `docs/product/README.md` |
| `docs/architecture/` | C4, ADR, API/data model, prompt/runtime policy, инициативные архитектурные пакеты | `docs/architecture/README.md` |
| `docs/delivery/` | Delivery plan, sprint/epic docs, traceability, process requirements, migration-map | `docs/delivery/README.md` |
| `docs/ops/` | Production runbook и эксплуатационные handover-артефакты | `docs/ops/README.md` |
| `docs/templates/` | Канонические шаблоны документов по ролям/stage | `docs/templates/index.md` |

## Быстрый маршрут
- Если нужен продуктовый source of truth: `docs/product/requirements_machine_driven.md`, `docs/product/constraints.md`, `docs/product/stage_process_model.md`.
- Если нужен архитектурный baseline: `docs/architecture/c4_context.md`, `docs/architecture/c4_container.md`, `docs/architecture/api_contract.md`, `docs/architecture/data_model.md`.
- Если нужен delivery/process baseline: `docs/delivery/development_process_requirements.md`, `docs/delivery/delivery_plan.md`, `docs/delivery/issue_map.md`, `docs/delivery/requirements_traceability.md`.
- Если нужен ops baseline: `docs/ops/production_runbook.md`.
- Если нужен шаблон артефакта: `docs/templates/index.md`.

## Специализированные каталоги
- Архитектурные initiative/stage-specific пакеты хранятся в `docs/architecture/initiatives/`.
- Эксплуатационные handover-пакеты хранятся в `docs/ops/handovers/`.
- Delivery day-эпики и sprint планы хранятся в `docs/delivery/epics/` и `docs/delivery/sprints/`.

## Governance
- Перенос документов выполняется только с migration-map: `docs/delivery/documentation_ia_migration_map.md`.
- При изменении doc-path обязательно синхронизируются `services.yaml`, traceability-документы и открытые GitHub issues.
- Branch-specific blob links для документов не считаются канонической навигацией и должны быть заменены на repo-local path refs или стабильные issue/PR ссылки.
