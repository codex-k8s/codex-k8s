---
doc_id: IDX-CK8S-DELIVERY-0001
type: domain-index
title: "Delivery Documentation Index"
status: in-review
owner_role: EM
created_at: 2026-03-11
updated_at: 2026-03-11
related_issues: [320]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-11-delivery-index"
---

# Delivery Documentation

## TL;DR
- `docs/delivery/` хранит process requirements, delivery plan, sprint/epic execution и traceability.
- Здесь же лежит migration-map для рефакторинга docs IA.

## Канонические документы
- `docs/delivery/development_process_requirements.md`
- `docs/delivery/delivery_plan.md`
- `docs/delivery/issue_map.md`
- `docs/delivery/requirements_traceability.md`
- `docs/delivery/e2e_mvp_master_plan.md`
- `docs/delivery/documentation_ia_migration_map.md`

## Индексы и каталоги
- `docs/delivery/sprints/README.md`
- `docs/delivery/epics/README.md`
- `docs/delivery/sprints/`
- `docs/delivery/epics/`

## Правила
- Любой перенос документов синхронно отражается в `issue_map`, `requirements_traceability`, relevant sprint/epic index и `services.yaml`.
- Delivery docs отвечают за sequencing, quality-gates и traceability, но не дублируют product/architecture source of truth.
