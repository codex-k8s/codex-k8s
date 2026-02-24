---
doc_id: DSG-CK8S-0119-IDX
type: design-index
title: "Issue #119 — Design Package Index"
status: draft
owner_role: SA
created_at: 2026-02-24
updated_at: 2026-02-24
related_issues: [119, 118]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-02-24-issue-119-design"
---

# Design package index: Issue #119

## Назначение
- Единая точка входа в design-пакет по Issue #119.
- Фиксирует структуру артефактов и порядок чтения для проверки Owner.

## Состав design-пакета
- `docs/delivery/design/issue-119/design_doc.md` — сервисные границы, сценарии A+B, инварианты и риски.
- `docs/delivery/design/issue-119/api_contract.md` — контрактные правила по labels/transitions без изменения API.
- `docs/delivery/design/issue-119/data_model.md` — evidence schema и используемые сущности данных.
- `docs/delivery/design/issue-119/migration_policy.md` — подтверждение отсутствия миграций и process-only impact.
- `docs/delivery/design/issue-119/traceability_matrix.md` — матрица `требование -> сценарий -> артефакт -> evidence`.

## Порядок review
1. `design_doc.md` (границы и инварианты).
2. `api_contract.md` + `data_model.md` (контракты и данные).
3. `migration_policy.md` (migration/runtime impact).
4. `traceability_matrix.md` (проверка полноты связей и AC coverage).

## Связанные документы
- `docs/delivery/e2e_mvp_master_plan.md` (срез A1/A2/A3/B1/B2/B3).
- `docs/delivery/issue_map.md` (трассировка issue/doc).
- `docs/delivery/requirements_traceability.md` (FR/NFR coverage).
- `docs/delivery/prd/issue-119/prd.md`
- `docs/delivery/prd/issue-119/nfr.md`
- `docs/delivery/prd/issue-119/user_story.md`
