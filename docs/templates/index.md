---
doc_id: IDX-CK8S-TEMPLATES-0001
type: template-index
title: "Template Catalog Index"
status: active
owner_role: KM
created_at: 2026-03-11
updated_at: 2026-03-11
related_issues: [320]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-11-template-index"
---

# Template Catalog

## TL;DR
- `docs/templates/` содержит только канонические шаблоны документов.
- Для навигации по фактической проектной документации используется `docs/index.md`, а не этот каталог.

## Шаблоны по ролям
- PM: `problem.md`, `brief.md`, `scope_mvp.md`, `constraints.md`, `project_charter.md`, `success_metrics.md`, `prd.md`, `nfr.md`, `user_story.md`
- EM/KM: `delivery_plan.md`, `epic.md`, `definition_of_done.md`, `issue_map.md`, `roadmap.md`, `docset_issue.md`, `docset_pr.md`
- SA: `c4_context.md`, `c4_container.md`, `adr.md`, `alternatives.md`, `design_doc.md`, `api_contract.md`, `data_model.md`, `migrations_policy.md`
- QA: `test_strategy.md`, `test_plan.md`, `test_matrix.md`, `regression_checklist.md`, `postdeploy_review.md`
- SRE: `release_plan.md`, `release_notes.md`, `rollback_plan.md`, `runbook.md`, `monitoring.md`, `alerts.md`, `slo.md`, `incident_playbook.md`, `incident_postmortem.md`

## Правила применения
- Шаблон остаётся шаблоном: фактический документ создаётся в доменном каталоге (`docs/product`, `docs/architecture`, `docs/delivery`, `docs/ops`).
- Если template-path меняется, в том же PR синхронизируются `services.yaml/spec.roleDocTemplates` и role-aware prompt contracts.
- Project indexes, handover-пакеты и sprint artifacts не хранятся в `docs/templates/`.

## Смежные индексы
- Корневой индекс: `docs/index.md`
- Product index: `docs/product/README.md`
- Architecture index: `docs/architecture/README.md`
- Delivery index: `docs/delivery/README.md`
- Ops index: `docs/ops/README.md`
