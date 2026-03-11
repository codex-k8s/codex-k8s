---
doc_id: MAP-CK8S-DOCS-IA-0001
type: migration-map
title: "Documentation IA Migration Map"
status: in-review
owner_role: KM
created_at: 2026-03-11
updated_at: 2026-03-11
related_issues: [318, 320]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-11-docs-ia-migration-map"
---

# Documentation IA Migration Map

## TL;DR
- Migration-map фиксирует допустимые переносы в рамках Issue `#320`.
- Верхний доменный слой `docs/product|architecture|delivery|ops|templates` не меняется.
- Переносы выполняются только вместе с обновлением traceability, `services.yaml` и открытых GitHub issues.

## Migration Map

| Old path | New path | Owner role | Affected links/issues | Migration note |
|---|---|---|---|---|
| `docs/architecture/agents_prompt_templates_lifecycle_design.md` | `docs/architecture/initiatives/agents_prompt_templates_lifecycle/architecture.md` | SA | `docs/delivery/issue_map.md`, `docs/delivery/requirements_traceability.md`, Sprint S6 epics/issues `#189/#195/#197/#199/#201/#216/#262/#263/#265` | Архитектурный package moved under initiative slug; root architecture keeps only cross-initiative source-of-truth docs. |
| `docs/architecture/agents_prompt_templates_lifecycle_design_doc.md` | `docs/architecture/initiatives/agents_prompt_templates_lifecycle/design_doc.md` | SA | `docs/delivery/issue_map.md`, `docs/delivery/requirements_traceability.md`, Sprint S6 epics/issues `#195/#197/#199/#201/#216/#262/#263/#265` | Design-stage artifact grouped with same initiative package. |
| `docs/architecture/agents_prompt_templates_lifecycle_api_contract.md` | `docs/architecture/initiatives/agents_prompt_templates_lifecycle/api_contract.md` | SA | `docs/delivery/issue_map.md`, `docs/delivery/requirements_traceability.md`, Sprint S6 epics/issues `#195/#197/#199/#201/#216/#262/#263/#265` | Initiative-specific API contract moved out of architecture root. |
| `docs/architecture/agents_prompt_templates_lifecycle_data_model.md` | `docs/architecture/initiatives/agents_prompt_templates_lifecycle/data_model.md` | SA | `docs/delivery/issue_map.md`, `docs/delivery/requirements_traceability.md`, Sprint S6 epics/issues `#195/#197/#199/#201/#216/#262/#263/#265` | Initiative-specific data model moved out of architecture root. |
| `docs/architecture/agents_prompt_templates_lifecycle_migrations_policy.md` | `docs/architecture/initiatives/agents_prompt_templates_lifecycle/migrations_policy.md` | SA | `docs/delivery/issue_map.md`, `docs/delivery/requirements_traceability.md`, Sprint S6 epics/issues `#195/#197/#199/#201/#216/#262/#263/#265` | Initiative migration policy grouped with same package. |
| `docs/architecture/s7_mvp_readiness_gap_closure_architecture.md` | `docs/architecture/initiatives/s7_mvp_readiness_gap_closure/architecture.md` | SA | `docs/delivery/delivery_plan.md`, `docs/delivery/issue_map.md`, `docs/delivery/requirements_traceability.md`, Sprint S7 epics/issues `#222/#238/#241` | Sprint-specific architecture package moved under initiative slug. |
| `docs/architecture/c4_context_s7_mvp_readiness_gap_closure.md` | `docs/architecture/initiatives/s7_mvp_readiness_gap_closure/c4_context.md` | SA | `docs/delivery/delivery_plan.md`, `docs/delivery/requirements_traceability.md`, Sprint S7 epics/issues `#222/#238/#241` | C4 overlay moved next to the rest of S7 initiative package. |
| `docs/architecture/c4_container_s7_mvp_readiness_gap_closure.md` | `docs/architecture/initiatives/s7_mvp_readiness_gap_closure/c4_container.md` | SA | `docs/delivery/delivery_plan.md`, `docs/delivery/requirements_traceability.md`, Sprint S7 epics/issues `#222/#238/#241` | C4 overlay moved next to the rest of S7 initiative package. |
| `docs/architecture/s7_mvp_readiness_gap_closure_design_doc.md` | `docs/architecture/initiatives/s7_mvp_readiness_gap_closure/design_doc.md` | SA | `docs/delivery/delivery_plan.md`, `docs/delivery/issue_map.md`, `docs/delivery/requirements_traceability.md`, Sprint S7 epics/issues `#238/#241` | Design-stage artifact grouped with the S7 package. |
| `docs/architecture/s7_mvp_readiness_gap_closure_api_contract.md` | `docs/architecture/initiatives/s7_mvp_readiness_gap_closure/api_contract.md` | SA | `docs/delivery/delivery_plan.md`, `docs/delivery/issue_map.md`, `docs/delivery/requirements_traceability.md`, Sprint S7 epics/issues `#238/#241` | Initiative-specific API contract moved out of architecture root. |
| `docs/architecture/s7_mvp_readiness_gap_closure_data_model.md` | `docs/architecture/initiatives/s7_mvp_readiness_gap_closure/data_model.md` | SA | `docs/delivery/delivery_plan.md`, `docs/delivery/issue_map.md`, `docs/delivery/requirements_traceability.md`, Sprint S7 epics/issues `#238/#241` | Initiative-specific data model moved out of architecture root. |
| `docs/architecture/s7_mvp_readiness_gap_closure_migrations_policy.md` | `docs/architecture/initiatives/s7_mvp_readiness_gap_closure/migrations_policy.md` | SA | `docs/delivery/delivery_plan.md`, `docs/delivery/issue_map.md`, `docs/delivery/requirements_traceability.md`, Sprint S7 epics/issues `#238/#241` | Initiative migration policy grouped with the S7 package. |
| `docs/ops/s6_postdeploy_ops_handover.md` | `docs/ops/handovers/s6/postdeploy_ops_handover.md` | SRE | `docs/delivery/issue_map.md`, `docs/delivery/requirements_traceability.md`, `docs/ops/production_runbook.md`, Sprint S6 issues `#263/#265` | Historical handover moved under dedicated ops handovers catalog. |
| `docs/ops/s6_ops_operational_baseline.md` | `docs/ops/handovers/s6/operational_baseline.md` | SRE | `docs/delivery/issue_map.md`, `docs/delivery/requirements_traceability.md`, `docs/ops/production_runbook.md`, Sprint S6 issues `#263/#265` | Historical operational baseline moved under dedicated ops handovers catalog. |
| `docs/README.md` | `docs/index.md` | KM | Issues `#281`, `#282`, repo-local docs references, future onboarding scaffold | Legacy root docs README path is retired; canonical root navigation now lives in `docs/index.md`. |
| `docs/03_engineering/definition_of_done.md` | `docs/templates/definition_of_done.md` | EM | `docs/templates/user_story.md`, future bootstrap/template consumers | Legacy template reference replaced with existing template catalog path. |

## Update order
1. Add root/domain indexes and migration-map.
2. Move files into target domain subfolders.
3. Rewrite internal repo links and `services.yaml`.
4. Update affected GitHub issues after repo paths are stable.
5. Run repo-local doc-drift check and attach evidence to PR.
