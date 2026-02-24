---
doc_id: DSG-CK8S-0119-TRC
type: design-traceability
title: "Issue #119 — Design Traceability Matrix (A+B)"
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

# Traceability matrix: Issue #119

## TL;DR
- Матрица связывает требования, сценарии A+B, design-артефакты и ожидаемое evidence.
- Назначение: убрать неоднозначность проверки и дать Owner проверяемый маршрут `source -> design -> evidence`.

## Матрица трассируемости

| Requirement | Сценарии #119 | Design-артефакты | Evidence / Verification |
|---|---|---|---|
| FR-026 (`run/state/need` taxonomy) | A1, A2, A3, B1, B2 | `design_doc.md`, `api_contract.md` | transitions в `flow_events`, проверка label-path в `docs/delivery/e2e_mvp_master_plan.md` |
| FR-033 (traceability для stage pipeline) | A1, A2, A3, B1, B2, B3 | `README.md`, `traceability_matrix.md`, `design_doc.md` | ссылки в `docs/delivery/issue_map.md` + handover evidence в Issue #118 |
| FR-052 (review-driven revise resolver) | B1, B2, B3 | `design_doc.md`, `api_contract.md` | B1: запуск `run:<stage>:revise`; B2: `need:input`; B3: sticky profile фиксация |
| NFR-010 (audit completeness) | A2, B1, B2, B3 | `data_model.md`, `api_contract.md` | наличие `correlation_id`, transitions и service-comment ссылок |
| NFR-018 (консистентность stage transitions) | A1, A2, A3, B1, B2 | `design_doc.md`, `migration_policy.md` | детерминированный resolver, отсутствие silent fallback, B2 fail-safe |

## Покрытие AC из user story

| AC (user_story.md) | Сценарии | Где зафиксировано |
|---|---|---|
| AC-1: lifecycle pass | A1, A2, A3 | `design_doc.md`, `docs/delivery/e2e_mvp_master_plan.md` |
| AC-2: revise + sticky profile | B1, B3 | `design_doc.md`, `api_contract.md` |
| AC-3: ambiguity guard | B2 | `design_doc.md`, `api_contract.md` |
| AC-4: evidence handover | A+B итог | `data_model.md`, `issue_map.md`, Issue #118 evidence comment |

## Правило обновления
- Любая правка сценариев A+B в issue #119 должна одновременно обновлять:
  - этот файл;
  - `docs/delivery/issue_map.md`;
  - `docs/delivery/e2e_mvp_master_plan.md` (срез #119).
