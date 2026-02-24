---
doc_id: TRT-CK8S-0001
type: requirements-traceability
title: "Requirements Traceability Matrix"
status: active
owner_role: EM
created_at: 2026-02-06
updated_at: 2026-02-24
related_issues: [1, 19, 74, 90, 100, 112, 125]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-19-full-docset"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-19
---

# Requirements Traceability Matrix

## TL;DR
- Матрица показывает, где каждый FR/NFR зафиксирован в текущей документации.
- Source of truth требований: `docs/product/requirements_machine_driven.md`.

## Матрица

| ID | Кратко | Основные документы | Статус |
|---|---|---|---|
| FR-001 | Kubernetes-only через SDK | `docs/product/requirements_machine_driven.md`, `docs/product/constraints.md`, `docs/architecture/adr/ADR-0001-kubernetes-only.md` | covered |
| FR-002 | Repository provider interface | `docs/product/requirements_machine_driven.md`, `docs/architecture/adr/ADR-0004-repository-provider-interface.md`, `docs/architecture/c4_container.md` | covered |
| FR-003 | Webhook-driven процессы | `docs/product/requirements_machine_driven.md`, `docs/architecture/adr/ADR-0002-webhook-driven-and-deploy-workflows.md`, `docs/architecture/api_contract.md` | covered |
| FR-004 | PostgreSQL + JSONB + pgvector | `docs/product/requirements_machine_driven.md`, `docs/product/constraints.md`, `docs/architecture/adr/ADR-0003-postgres-jsonb-pgvector.md`, `docs/architecture/data_model.md` | covered |
| FR-005 | Платформа и БД в Kubernetes | `docs/product/requirements_machine_driven.md`, `docs/architecture/c4_container.md`, `docs/delivery/delivery_plan.md` | covered |
| FR-006 | MCP service tools в Go | `docs/product/requirements_machine_driven.md`, `docs/product/brief.md`, `docs/design-guidelines/AGENTS.md` | covered |
| FR-007 | GitHub OAuth для staff UI | `docs/product/requirements_machine_driven.md`, `docs/architecture/c4_context.md`, `docs/architecture/api_contract.md` | covered |
| FR-008 | Настройки в БД, deploy secrets из env | `docs/product/requirements_machine_driven.md`, `docs/product/constraints.md`, `AGENTS.md` | covered |
| FR-009 | Агенты/сессии/журналы в БД + UI | `docs/product/requirements_machine_driven.md`, `docs/architecture/data_model.md`, `docs/architecture/c4_container.md` | covered |
| FR-010 | Фиксированный roster агентов + задел на расширение | `docs/product/requirements_machine_driven.md`, `docs/architecture/data_model.md`, `docs/delivery/roadmap.md` | covered |
| FR-011 | Агентные токены: генерация/ротация/шифрование | `docs/product/requirements_machine_driven.md`, `docs/architecture/data_model.md`, `docs/product/constraints.md` | covered |
| FR-012 | Жизненный цикл run/pod/namespace в БД + UI | `docs/product/requirements_machine_driven.md`, `docs/architecture/c4_container.md`, `docs/architecture/data_model.md`, `docs/architecture/agent_runtime_rbac.md`, `docs/architecture/adr/ADR-0005-run-namespace-ttl-and-revise-reuse.md`, `docs/delivery/epics/s2/epic-s2-day3-per-issue-namespace-and-rbac.md`, `docs/delivery/epics/s3/epic-s3-day19.7-run-namespace-ttl-and-revise-reuse.md` | covered |
| FR-013 | Многоподовость + split service/job zones | `docs/product/requirements_machine_driven.md`, `AGENTS.md`, `docs/design-guidelines/common/project_architecture.md` | covered |
| FR-014 | Слоты через БД | `docs/product/requirements_machine_driven.md`, `docs/architecture/data_model.md` | covered |
| FR-015 | Шаблоны документов в БД + markdown editor | `docs/product/requirements_machine_driven.md`, `docs/architecture/data_model.md`, `docs/architecture/api_contract.md` | covered |
| FR-016 | Bootstrap 2 режима (existing k8s / k3s install) | `docs/product/requirements_machine_driven.md`, `docs/delivery/delivery_plan.md`, `docs/product/brief.md` | covered |
| FR-017 | Project RBAC read/read_write/admin | `docs/product/requirements_machine_driven.md`, `docs/product/constraints.md`, `docs/architecture/data_model.md` | covered |
| FR-018 | No self-signup, email matching | `docs/product/requirements_machine_driven.md`, `docs/product/constraints.md`, `docs/architecture/data_model.md` | covered |
| FR-019 | Добавление пользователей через staff UI | `docs/product/requirements_machine_driven.md`, `docs/architecture/api_contract.md`, `docs/architecture/data_model.md` | covered |
| FR-020 | Multi-repo per project + per-repo services.yaml | `docs/product/requirements_machine_driven.md`, `docs/architecture/data_model.md`, `docs/architecture/multi_repo_mode_design.md`, `docs/architecture/adr/ADR-0007-multi-repo-composition-and-docs-federation.md`, `docs/product/brief.md`, `docs/delivery/sprints/s4/sprint_s4_multi_repo_federation.md`, `docs/delivery/epics/s4/epic_s4.md`, `docs/delivery/epics/s4/epic-s4-day1-multi-repo-composition-and-docs-federation.md` | covered |
| FR-021 | Repo token per repository + future Vault/JWT path | `docs/product/requirements_machine_driven.md`, `docs/architecture/data_model.md`, `docs/delivery/roadmap.md` | covered |
| FR-022 | codex-k8s как проект с monorepo services.yaml | `docs/product/requirements_machine_driven.md`, `README.md` | covered |
| FR-023 | Learning mode + educational PR comments | `docs/product/requirements_machine_driven.md`, `docs/product/brief.md`, `docs/architecture/api_contract.md`, `docs/delivery/delivery_plan.md`, `docs/architecture/data_model.md` | covered |
| FR-024 | CODEXK8S_ prefix для env/secrets/CI vars | `docs/product/requirements_machine_driven.md`, `AGENTS.md` | covered |
| FR-025 | MVP public API: only webhook ingress | `docs/product/requirements_machine_driven.md`, `docs/product/constraints.md`, `docs/architecture/api_contract.md` | covered |
| FR-026 | Канонический каталог лейблов run/state/need | `docs/product/requirements_machine_driven.md`, `docs/product/labels_and_trigger_policy.md`, `docs/product/stage_process_model.md`, `docs/delivery/e2e_mvp_master_plan.md` | covered |
| FR-027 | Approval policy для trigger/deploy labels | `docs/product/requirements_machine_driven.md`, `docs/product/labels_and_trigger_policy.md`, `docs/architecture/mcp_approval_and_audit_flow.md` | covered |
| FR-028 | Stage process model с revise/rethink | `docs/product/requirements_machine_driven.md`, `docs/product/stage_process_model.md`, `docs/delivery/sprints/s2/sprint_s2_dogfooding.md` | covered |
| FR-029 | Базовый штат агентов (включая `dev` и `reviewer`) + custom роли проекта | `docs/product/requirements_machine_driven.md`, `docs/product/agents_operating_model.md`, `docs/architecture/data_model.md`, `docs/architecture/agent_runtime_rbac.md` | covered |
| FR-030 | Prompt templates policy: seed + DB override | `docs/product/requirements_machine_driven.md`, `docs/product/agents_operating_model.md`, `docs/architecture/prompt_templates_policy.md`, `docs/delivery/epics/s2/epic-s2-day4-agent-job-and-pr-flow.md` | covered |
| FR-031 | Mixed runtime mode full-env/code-only | `docs/product/requirements_machine_driven.md`, `docs/product/agents_operating_model.md`, `docs/architecture/agent_runtime_rbac.md`, `docs/delivery/epics/s2/epic-s2-day3-per-issue-namespace-and-rbac.md` | covered |
| FR-032 | Обязательные audit сущности agent_sessions/token_usage/links | `docs/product/requirements_machine_driven.md`, `docs/architecture/data_model.md`, `docs/architecture/mcp_approval_and_audit_flow.md` | covered |
| FR-033 | Traceability для stage pipeline | `docs/product/requirements_machine_driven.md`, `docs/delivery/issue_map.md`, `docs/delivery/requirements_traceability.md`, `docs/delivery/sprints/README.md`, `docs/delivery/epics/README.md` | covered |
| FR-034 | Контекстный рендер prompt templates | `docs/product/requirements_machine_driven.md`, `docs/architecture/prompt_templates_policy.md`, `docs/product/agents_operating_model.md`, `docs/delivery/epics/s2/epic-s2-day3.5-mcp-github-k8s-and-prompt-context.md` | covered |
| FR-035 | Локали prompt templates и fallback по locale | `docs/product/requirements_machine_driven.md`, `docs/architecture/prompt_templates_policy.md`, `docs/product/constraints.md` | covered |
| FR-036 | Сохранение/возобновление codex-cli session JSON | `docs/product/requirements_machine_driven.md`, `docs/architecture/data_model.md`, `docs/architecture/agent_runtime_rbac.md`, `docs/delivery/epics/s2/epic-s2-day4-agent-job-and-pr-flow.md` | covered |
| FR-037 | `agent` как центр настроек и политик выполнения | `docs/product/requirements_machine_driven.md`, `docs/architecture/data_model.md`, `docs/product/agents_operating_model.md` | covered |
| FR-038 | Contract-first OpenAPI + backend/frontend codegen | `docs/product/requirements_machine_driven.md`, `docs/architecture/api_contract.md`, `docs/delivery/sprints/s2/sprint_s2_dogfooding.md`, `docs/delivery/epics/s2/epic-s2-day1-migrations-and-schema-ownership.md` | covered |
| FR-039 | Универсальные HTTP-контракты approver/executor через MCP | `docs/product/requirements_machine_driven.md`, `docs/architecture/mcp_approval_and_audit_flow.md`, `docs/architecture/c4_context.md`, `docs/delivery/epics/s2/epic-s2-day3.5-mcp-github-k8s-and-prompt-context.md` | covered |
| FR-040 | Staff UI runtime debug: jobs/logs/wait queue | `docs/product/requirements_machine_driven.md`, `docs/architecture/api_contract.md`, `docs/delivery/epics/s3/epic-s3-day2-staff-runtime-debug-console.md` | covered |
| FR-041 | MCP control tools: secret sync + db lifecycle + owner feedback | `docs/product/requirements_machine_driven.md`, `docs/architecture/mcp_approval_and_audit_flow.md`, `docs/delivery/epics/s2/epic-s2-day6-approval-and-audit-hardening.md`, `docs/delivery/epics/s3/epic-s3-day3-mcp-deterministic-secret-sync.md`, `docs/delivery/epics/s3/epic-s3-day4-mcp-database-lifecycle.md`, `docs/delivery/epics/s3/epic-s3-day5-feedback-and-approver-interfaces.md` | covered |
| FR-042 | Approval matrix для MCP control tools | `docs/product/requirements_machine_driven.md`, `docs/product/labels_and_trigger_policy.md`, `docs/architecture/mcp_approval_and_audit_flow.md`, `docs/delivery/epics/s2/epic-s2-day6-approval-and-audit-hardening.md` | covered |
| FR-043 | `run:self-improve` trigger и диагностика | `docs/product/requirements_machine_driven.md`, `docs/product/labels_and_trigger_policy.md`, `docs/product/stage_process_model.md`, `docs/delivery/epics/s3/epic-s3-day6-self-improve-ingestion-and-diagnostics.md`, `docs/delivery/epics/s3/epic-s3-day8-agent-toolchain-auto-extension.md` | covered |
| FR-044 | `run:self-improve` updater + PR flow | `docs/product/requirements_machine_driven.md`, `docs/architecture/prompt_templates_policy.md`, `docs/delivery/epics/s3/epic-s3-day7-self-improve-updater-and-pr-flow.md`, `docs/delivery/epics/s3/epic-s3-day8-agent-toolchain-auto-extension.md` | covered |
| FR-045 | Full stage-flow activation `run:intake..run:ops` + revise/rethink | `docs/product/requirements_machine_driven.md`, `docs/product/stage_process_model.md`, `docs/delivery/epics/s3/epic-s3-day1-full-stage-and-label-activation.md` | covered |
| FR-046 | Post-MVP roadmap направлений | `docs/product/requirements_machine_driven.md`, `docs/product/brief.md`, `docs/delivery/roadmap.md` | covered |
| FR-047 | Docset import + safe sync в проекты | `docs/product/requirements_machine_driven.md`, `docs/delivery/epics/s3/epic-s3-day12-docset-import-and-safe-sync.md` | covered |
| FR-048 | Unified config/secrets governance + sync в GitHub/K8s | `docs/product/requirements_machine_driven.md`, `docs/delivery/epics/s3/epic-s3-day13-config-and-credentials-governance.md` | covered |
| FR-049 | Repo onboarding preflight (GitHub ops + domain resolution) | `docs/product/requirements_machine_driven.md`, `docs/delivery/epics/s3/epic-s3-day14-repository-onboarding-preflight.md` | covered |
| FR-050 | Prompt context docs tree + role-aware capabilities | `docs/product/requirements_machine_driven.md`, `docs/architecture/prompt_templates_policy.md`, `docs/delivery/epics/s3/epic-s3-day15-mvp-closeout-and-handover.md` | covered |
| FR-051 | GitHub run service messages v2 + slot URL for full-env | `docs/product/requirements_machine_driven.md`, `docs/architecture/api_contract.md`, `docs/delivery/epics/s3/epic-s3-day15-mvp-closeout-and-handover.md` | covered |
| FR-052 | Review-driven revise resolver + stage-aware next-step action cards | `docs/product/requirements_machine_driven.md`, `docs/product/labels_and_trigger_policy.md`, `docs/product/stage_process_model.md`, `docs/architecture/mcp_approval_and_audit_flow.md`, `docs/architecture/adr/ADR-0006-review-driven-revise-and-next-step-ux.md`, `docs/delivery/e2e_mvp_master_plan.md` (вкл. B.1 smoke для Issue #125) | covered |
| NFR-001 | Security baseline | `docs/product/requirements_machine_driven.md`, `docs/product/constraints.md`, `AGENTS.md` | covered |
| NFR-002 | Multi-pod consistency | `docs/product/requirements_machine_driven.md`, `docs/architecture/c4_container.md`, `docs/architecture/data_model.md` | covered |
| NFR-003 | No event outbox on MVP | `docs/product/requirements_machine_driven.md`, `docs/architecture/data_model.md`, `docs/product/constraints.md` | covered |
| NFR-004 | Embedding vector(3072) | `docs/product/requirements_machine_driven.md`, `docs/architecture/data_model.md`, `docs/product/constraints.md` | covered |
| NFR-005 | Read-replica baseline | `docs/product/requirements_machine_driven.md`, `docs/architecture/c4_container.md`, `docs/product/constraints.md` | covered |
| NFR-006 | One-command production bootstrap via SSH | `docs/product/requirements_machine_driven.md`, `docs/delivery/delivery_plan.md`, `docs/product/brief.md` | covered |
| NFR-007 | CI/CD model (main->production webhook-driven self-deploy) | `docs/product/requirements_machine_driven.md`, `docs/product/brief.md`, `docs/product/constraints.md`, `docs/delivery/delivery_plan.md` | covered |
| NFR-008 | MVP storage profile local-path | `docs/product/requirements_machine_driven.md`, `docs/product/constraints.md`, `docs/delivery/delivery_plan.md` | covered |
| NFR-009 | Управляемые лимиты параллелизма agent-runs | `docs/product/requirements_machine_driven.md`, `docs/product/agents_operating_model.md`, `docs/architecture/agent_runtime_rbac.md`, `docs/delivery/epics/s2/epic-s2-day3-per-issue-namespace-and-rbac.md`, `docs/delivery/epics/s3/epic-s3-day19.7-run-namespace-ttl-and-revise-reuse.md` | covered |
| NFR-010 | Полная audit-трассировка stage/label действий | `docs/product/requirements_machine_driven.md`, `docs/architecture/mcp_approval_and_audit_flow.md`, `docs/architecture/data_model.md`, `docs/delivery/epics/s2/epic-s2-day3.5-mcp-github-k8s-and-prompt-context.md` | covered |
| NFR-011 | Labels-as-vars в runtime orchestration | `docs/product/requirements_machine_driven.md`, `docs/product/labels_and_trigger_policy.md`, `docs/delivery/epics/s2/epic-s2-day2-issue-label-triggers-run-dev.md` | covered |
| NFR-012 | Запрет timeout-kill при ожидании MCP | `docs/product/requirements_machine_driven.md`, `docs/architecture/mcp_approval_and_audit_flow.md`, `docs/architecture/agent_runtime_rbac.md` | covered |
| NFR-013 | Надёжное хранение resumable session snapshot | `docs/product/requirements_machine_driven.md`, `docs/architecture/data_model.md`, `docs/product/constraints.md`, `docs/delivery/epics/s2/epic-s2-day4-agent-job-and-pr-flow.md` | covered |
| NFR-014 | Воспроизводимый OpenAPI codegen в CI | `docs/product/requirements_machine_driven.md`, `docs/architecture/api_contract.md`, `docs/delivery/epics/s2/epic-s2-day1-migrations-and-schema-ownership.md`, `deploy/base/codex-k8s/codegen-check-job.yaml.tpl` | covered |
| NFR-015 | Операционная latency для runtime debug UI | `docs/product/requirements_machine_driven.md`, `docs/architecture/api_contract.md`, `docs/delivery/epics/s3/epic-s3-day2-staff-runtime-debug-console.md` | covered |
| NFR-016 | Idempotent и secret-safe поведение MCP control tools | `docs/product/requirements_machine_driven.md`, `docs/architecture/mcp_approval_and_audit_flow.md`, `docs/delivery/epics/s2/epic-s2-day6-approval-and-audit-hardening.md`, `docs/delivery/epics/s3/epic-s3-day3-mcp-deterministic-secret-sync.md`, `docs/delivery/epics/s3/epic-s3-day4-mcp-database-lifecycle.md` | covered |
| NFR-017 | Воспроизводимость self-improve цикла | `docs/product/requirements_machine_driven.md`, `docs/architecture/data_model.md`, `docs/delivery/epics/s3/epic-s3-day6-self-improve-ingestion-and-diagnostics.md`, `docs/delivery/epics/s3/epic-s3-day7-self-improve-updater-and-pr-flow.md`, `docs/delivery/epics/s3/epic-s3-day8-agent-toolchain-auto-extension.md` | covered |
| NFR-018 | Консистентность переходов full stage-flow | `docs/product/requirements_machine_driven.md`, `docs/product/stage_process_model.md`, `docs/delivery/epics/s3/epic-s3-day1-full-stage-and-label-activation.md`, `docs/delivery/epics/s3/epic-s3-day20-e2e-regression-and-mvp-closeout.md`, `docs/delivery/e2e_mvp_master_plan.md` | covered |

## Правило актуализации
- Любое новое требование сначала добавляется в `docs/product/requirements_machine_driven.md`, затем отражается в этой матрице.
- Если строка в матрице теряет ссылку на целевой документ, статус меняется на `gap` до устранения.
