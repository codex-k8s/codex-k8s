---
doc_id: TRT-CK8S-0001
type: requirements-traceability
title: "Requirements Traceability Matrix"
status: active
owner_role: EM
created_at: 2026-02-06
updated_at: 2026-03-11
related_issues: [1, 19, 74, 90, 100, 112, 154, 155, 159, 165, 170, 171, 175, 184, 185, 187, 189, 195, 197, 199, 201, 210, 212, 218, 220, 222, 223, 225, 226, 227, 228, 229, 230, 238, 241, 243, 244, 245, 246, 247, 248, 249, 250, 251, 252, 253, 254, 255, 256, 257, 258, 259, 260, 216, 262, 263, 265, 281, 282]
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
| FR-009 | Сессии/журналы в БД + runtime debug UI; prompt templates вне staff UI | `docs/product/requirements_machine_driven.md`, `docs/product/agents_operating_model.md`, `docs/architecture/data_model.md`, `docs/architecture/c4_container.md` | covered |
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
| FR-026 | Канонический каталог лейблов run/state/need + PR trigger `need:reviewer` для pre-review | `docs/product/requirements_machine_driven.md`, `docs/product/labels_and_trigger_policy.md`, `docs/product/stage_process_model.md`, `docs/product/agents_operating_model.md`, `docs/delivery/e2e_mvp_master_plan.md` | covered |
| FR-027 | Approval policy для trigger/deploy labels | `docs/product/requirements_machine_driven.md`, `docs/product/labels_and_trigger_policy.md`, `docs/architecture/mcp_approval_and_audit_flow.md` | covered |
| FR-028 | Stage process model с revise/rethink | `docs/product/requirements_machine_driven.md`, `docs/product/stage_process_model.md`, `docs/delivery/sprints/s2/sprint_s2_dogfooding.md` | covered |
| FR-029 | Базовый штат системных агентов; custom-agent factory post-MVP | `docs/product/requirements_machine_driven.md`, `docs/product/agents_operating_model.md`, `docs/architecture/data_model.md`, `docs/architecture/agent_runtime_rbac.md`, `docs/delivery/roadmap.md` | covered |
| FR-030 | Prompt templates policy: repo-only role-specific seeds | `docs/product/requirements_machine_driven.md`, `docs/product/agents_operating_model.md`, `docs/architecture/prompt_templates_policy.md`, `docs/delivery/epics/s2/epic-s2-day4-agent-job-and-pr-flow.md`, `services.yaml`, `services/jobs/agent-runner/internal/runner/promptseeds/README.md`, `services/jobs/agent-runner/internal/runner/promptseeds/*.md`, `services/jobs/agent-runner/internal/runner/helpers_prompt_doc_stage_seeds_test.go`, `services/jobs/agent-runner/internal/runner/templates/prompt_envelope.tmpl`, `services/jobs/agent-runner/internal/runner/templates/prompt_blocks/*.tmpl`, `services/jobs/worker/internal/domain/worker/agent_job_context_test.go` | covered |
| FR-031 | Mixed runtime mode full-env/code-only | `docs/product/requirements_machine_driven.md`, `docs/product/agents_operating_model.md`, `docs/architecture/agent_runtime_rbac.md`, `docs/delivery/epics/s2/epic-s2-day3-per-issue-namespace-and-rbac.md` | covered |
| FR-032 | Обязательные audit сущности agent_sessions/token_usage/links | `docs/product/requirements_machine_driven.md`, `docs/architecture/data_model.md`, `docs/architecture/mcp_approval_and_audit_flow.md` | covered |
| FR-033 | Traceability для stage pipeline | `docs/product/requirements_machine_driven.md`, `docs/delivery/issue_map.md`, `docs/delivery/requirements_traceability.md`, `docs/delivery/sprints/README.md`, `docs/delivery/epics/README.md`, `docs/templates/prd.md` | covered |
| FR-034 | Контекстный рендер prompt templates | `docs/product/requirements_machine_driven.md`, `docs/architecture/prompt_templates_policy.md`, `docs/product/agents_operating_model.md`, `docs/delivery/epics/s2/epic-s2-day3.5-mcp-github-k8s-and-prompt-context.md` | covered |
| FR-035 | Локали prompt templates и fallback по platform default locale | `docs/product/requirements_machine_driven.md`, `docs/architecture/prompt_templates_policy.md`, `docs/product/constraints.md`, `docs/architecture/api_contract.md`, `services/jobs/worker/internal/app/config.go` | covered |
| FR-036 | Сохранение/возобновление codex-cli session JSON | `docs/product/requirements_machine_driven.md`, `docs/architecture/data_model.md`, `docs/architecture/agent_runtime_rbac.md`, `docs/delivery/epics/s2/epic-s2-day4-agent-job-and-pr-flow.md` | covered |
| FR-037 | `agent` как реестр системных профилей, без settings UI в MVP | `docs/product/requirements_machine_driven.md`, `docs/architecture/data_model.md`, `docs/product/agents_operating_model.md` | covered |
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
| FR-052 | Review-driven revise resolver + stage-aware next-step action cards | `docs/product/requirements_machine_driven.md`, `docs/product/labels_and_trigger_policy.md`, `docs/product/stage_process_model.md`, `docs/architecture/mcp_approval_and_audit_flow.md`, `docs/architecture/adr/ADR-0006-review-driven-revise-and-next-step-ux.md` | covered |
| FR-053 | Launch profiles для разных типов инициатив (`quick-fix`, `feature`, `new-service`) | `docs/product/requirements_machine_driven.md`, `docs/product/stage_process_model.md`, `docs/product/labels_and_trigger_policy.md`, `docs/architecture/api_contract.md`, `docs/architecture/adr/ADR-0008-profile-driven-stage-launch-and-next-step-contract.md`, `docs/delivery/sprints/s5/sprint_s5_stage_entry_and_label_ux.md`, `docs/delivery/epics/s5/epic_s5.md`, `docs/delivery/epics/s5/epic-s5-day1-launch-profiles-and-stage-launcher-ux.md`, `docs/delivery/epics/s5/prd-s5-day1-launch-profiles-and-stage-launcher-ux.md`, `docs/delivery/epics/s5/epic-s5-day2-launch-profiles-dev-execution.md`, `docs/delivery/issue_map.md` | covered |
| FR-054 | Next-step actions: primary deep-link + fallback-команда | `docs/product/requirements_machine_driven.md`, `docs/product/labels_and_trigger_policy.md`, `docs/product/stage_process_model.md`, `docs/architecture/api_contract.md`, `docs/architecture/adr/ADR-0008-profile-driven-stage-launch-and-next-step-contract.md`, `docs/delivery/sprints/s5/sprint_s5_stage_entry_and_label_ux.md`, `docs/delivery/epics/s5/epic_s5.md`, `docs/delivery/epics/s5/epic-s5-day1-launch-profiles-and-stage-launcher-ux.md`, `docs/delivery/epics/s5/prd-s5-day1-launch-profiles-and-stage-launcher-ux.md`, `docs/delivery/epics/s5/epic-s5-day2-launch-profiles-dev-execution.md`, `docs/delivery/issue_map.md` | covered |
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

## Актуализация по Issue #184 (`run:intake`, 2026-02-25)
- Для FR-009/FR-030/FR-032/FR-033/FR-038 добавлен intake traceability пакет Sprint S6:
  `docs/delivery/sprints/s6/sprint_s6_agents_prompt_management.md`,
  `docs/delivery/epics/s6/epic_s6.md`,
  `docs/delivery/epics/s6/epic-s6-day1-agents-prompts-intake.md`.
- Создана stage-continuity issue `#185` для stage `run:vision` без trigger-лейбла (ставит Owner) с обязательной инструкцией сформировать issue следующего этапа (`run:prd`), чтобы сохранить последовательную декомпозицию до `run:doc-audit`.
- Зафиксировано продуктовое расхождение As-Is: UI-раздел `Agents` и prompt templates находится в scaffold-состоянии, при этом contract-first staff API пока не содержит endpoint-ов для agents/templates/audit lifecycle.
- Зафиксирован stage-handover baseline для полного цикла до `run:doc-audit` и обязательное правило создания follow-up issue на каждом следующем stage до `run:plan` включительно.

## Актуализация по Issue #187 (`run:prd`, 2026-02-25)
- Для FR-009/FR-015/FR-030/FR-033/FR-038 и NFR-010/NFR-015/NFR-018 добавлен PRD traceability пакет Sprint S6:
  `docs/delivery/epics/s6/epic-s6-day3-agents-prompts-prd.md`,
  `docs/delivery/epics/s6/prd-s6-day3-agents-prompts-lifecycle.md`,
  `docs/delivery/epics/s6/epic_s6.md`,
  `docs/delivery/sprints/s6/sprint_s6_agents_prompt_management.md`.
- Формализованы требования и критерии приемки для контуров `agents settings`, `prompt templates lifecycle`, `history/audit` в формате FR/AC/NFR-draft.
- Подтверждена трассируемость stage-цепочки `#184 -> #185 -> #187` и создана follow-up issue `#189` для stage `run:arch` без trigger-лейбла (ставит Owner) с обязательной инструкцией создать issue `run:design` по завершении архитектурного этапа.
- Зафиксирован policy-safe scope этапа: markdown-only изменения без обновления code/runtime артефактов.

## Актуализация по Issue #189 (`run:arch`, 2026-02-25)
- Архитектурный пакет для lifecycle управления агентами и шаблонами промптов зафиксирован в:
  `docs/architecture/agents_prompt_templates_lifecycle_design.md`,
  `docs/architecture/adr/ADR-0009-prompt-templates-lifecycle-and-audit.md`,
  `docs/architecture/alternatives/ALT-0001-agents-prompt-templates-lifecycle.md`.
- Трассируемость PRD-артефактов S6 Day3 зафиксирована через Issue `#187` и PR `#190` (merged).
- Handover в `run:design` включает обязательные артефакты OpenAPI, data model/migrations и UI flow для `agents/templates/audit`, а также migration/runtime impact.
- По итогам `run:arch` создана follow-up issue `#195` для stage `run:design` с обязательной инструкцией после завершения stage создать issue следующего этапа `run:plan`.
- Через Context7 подтверждено, что для design-этапа не требуется новая внешняя библиотека:
  достаточно текущего стека `kin-openapi` (валидация контрактов) и `monaco-editor` (DiffEditor).

## Актуализация по Issue #195 (`run:design`, 2026-02-25)
- Подготовлен полный design package для `agents/templates/audit`:
  `docs/architecture/agents_prompt_templates_lifecycle_design_doc.md`,
  `docs/architecture/agents_prompt_templates_lifecycle_api_contract.md`,
  `docs/architecture/agents_prompt_templates_lifecycle_data_model.md`,
  `docs/architecture/agents_prompt_templates_lifecycle_migrations_policy.md`.
- Зафиксированы typed transport boundaries (staff HTTP + internal gRPC), error/validator/concurrency contract и UI flow для list/details/diff/preview/history.
- Обновлены артефакты Sprint S6 Day5:
  `docs/delivery/epics/s6/epic-s6-day5-agents-prompts-design.md`,
  `docs/delivery/epics/s6/epic_s6.md`,
  `docs/delivery/sprints/s6/sprint_s6_agents_prompt_management.md`.
- Через Context7 подтверждён dependency baseline для реализации без новых библиотек:
  `kin-openapi` (`/getkin/kin-openapi`) и `monaco-editor` (`/microsoft/monaco-editor`).
- Создана follow-up issue `#197` для stage `run:plan` с обязательной инструкцией после `run:plan` создать issue `run:dev`.

## Актуализация по Issue #197 (`run:plan`, 2026-02-25)
- Для FR-033/FR-038 и NFR-010/NFR-018 добавлен execution-governance пакет Sprint S6 Day6:
  `docs/delivery/epics/s6/epic-s6-day6-agents-prompts-plan.md`,
  `docs/delivery/epics/s6/epic_s6.md`,
  `docs/delivery/sprints/s6/sprint_s6_agents_prompt_management.md`,
  `docs/delivery/delivery_plan.md`.
- Зафиксирована декомпозиция `run:dev` по потокам W1..W7 с quality-gates QG-S6-D6-01..QG-S6-D6-07 и DoR/DoD-критериями перехода в `run:qa`.
- Сформирован реестр blockers/risks/owner decisions для handover в реализацию без выхода за архитектурные границы Day5 design package.
- Создана follow-up issue `#199` для stage `run:dev` без trigger-лейбла с обязательной continuity-инструкцией создать issue `run:qa` после завершения реализации.
- Через Context7 (`/websites/cli_github_manual`) подтверждён актуальный синтаксис `gh issue/pr` команд для fallback/PR-flow; новые внешние зависимости не требуются.

## Актуализация по Issue #199 (`run:dev`, 2026-02-25)
- Для FR-009/FR-015/FR-030/FR-033/FR-038 и NFR-010/NFR-015/NFR-018 реализован execution-пакет:
  - contract-first расширение `services/external/api-gateway/api/server/api.yaml` и `proto/codexk8s/controlplane/v1/controlplane.proto`;
  - доменные use-cases/control-plane transport для `agents/templates/audit`;
  - миграция `prompt_templates` + `agents.settings/settings_version`;
  - frontend `Agents` переведён с scaffold на typed API flow (list/details/settings/diff/preview/history).
- Реализация оформлена в `GitHub PR #202` с синхронным обновлением contract/codegen/docs артефактов.
- Собрано regression evidence:
  - `go test ./services/internal/control-plane/...`
  - `go test ./services/external/api-gateway/...`
  - `npm run build` (`services/staff/web-console`)
  - `make lint-go`
  - `make dupl-go` (зафиксированы pre-existing дубли вне scope текущих правок).
- Создана follow-up issue `#201` для stage `run:qa` с обязательной continuity-инструкцией по созданию issue `run:release` после завершения QA.

## Актуализация по Issue #262 (`run:release`, 2026-03-02)
- Для FR-028/FR-033/FR-045 и NFR-007/NFR-010/NFR-018 зафиксирован release closeout пакет Sprint S6:
  `docs/delivery/epics/s6/epic-s6-day9-release-closeout.md`,
  `docs/delivery/epics/s6/epic_s6.md`,
  `docs/delivery/sprints/s6/sprint_s6_agents_prompt_management.md`,
  `docs/delivery/delivery_plan.md`,
  `docs/delivery/issue_map.md`.
- Подтверждена release continuity цепочка:
  `#199 -> #201 -> #216 -> #262`, сформирован handover в `run:postdeploy` через issue `#263`.
- Зафиксированы release quality-gates, DoD, release notes и rollback/mitigation план без расширения scope за пределы markdown-only policy.
- Через Context7 (`/websites/cli_github_manual`) подтверждён актуальный синтаксис `gh issue/pr` команд для PR-flow и label-transition fallback.

## Актуализация по Issue #263 (`run:postdeploy`, 2026-03-02)
- Для FR-028/FR-033/FR-045 и NFR-001/NFR-003/NFR-010/NFR-018 оформлен postdeploy evidence пакет Sprint S6:
  `docs/delivery/epics/s6/epic-s6-day10-postdeploy-review.md`,
  `docs/ops/s6_postdeploy_ops_handover.md`,
  `docs/ops/production_runbook.md`,
  `docs/delivery/epics/s6/epic_s6.md`,
  `docs/delivery/sprints/s6/sprint_s6_agents_prompt_management.md`,
  `docs/delivery/delivery_plan.md`,
  `docs/delivery/issue_map.md`.
- Подтверждена stage continuity цепочка:
  `#199 -> #201 -> #216 -> #262 -> #263`, подготовлен handover в `run:ops` через issue `#265`.
- Зафиксированы runtime проверки postdeploy (health/logs/events/services/jobs), residual operational risks и action items для следующего этапа.
- Через Context7 подтверждены актуальные рекомендации:
  - Kubernetes probe semantics (`startupProbe` для slow-start): `/websites/kubernetes_io`;
  - anti-noise alerting (`for`, `keep_firing_for`) и user-impact paging: `/prometheus/docs`, `/websites/prometheus_io`.

## Актуализация по Issue #265 (`run:ops`, 2026-03-02)
- Для FR-028/FR-033/FR-045 и NFR-001/NFR-007/NFR-010/NFR-018 оформлен ops closeout пакет Sprint S6:
  `docs/delivery/epics/s6/epic-s6-day11-ops-operational-closeout.md`,
  `docs/ops/s6_ops_operational_baseline.md`,
  `docs/ops/production_runbook.md`,
  `docs/ops/s6_postdeploy_ops_handover.md`,
  `docs/delivery/epics/s6/epic_s6.md`,
  `docs/delivery/sprints/s6/sprint_s6_agents_prompt_management.md`,
  `docs/delivery/delivery_plan.md`,
  `docs/delivery/issue_map.md`.
- Подтверждена stage continuity цепочка:
  `#199 -> #201 -> #216 -> #262 -> #263 -> #265`; зафиксирован handover на следующий контур `run:doc-audit`.
- Зафиксированы формальные эксплуатационные решения:
  runbook triage baseline, monitoring/alert thresholds, SLO burn-rate policy и rollback readiness критерии.
- Через Context7 подтверждены актуальные рекомендации:
  - Kubernetes probes и handoff `startupProbe -> liveness/readiness`: `/websites/kubernetes_io`;
  - Prometheus alerting anti-noise (`for`, `keep_firing_for`) и rule baseline: `/prometheus/docs`, `/websites/prometheus_io`.

## Актуализация по Issue #155 (`run:plan`, 2026-02-25)
- Для FR-053/FR-054 добавлены execution-governance артефакты Sprint S5 (`epic_s5.md`, обновлённый sprint-plan, issue-map sync).
- Зафиксированы quality-gates QG-01..QG-05 и критерии завершения handover в `run:dev`; QG-05 закрыт после Owner review в PR #166.
- Реестр `BLK-155-*`, `RSK-155-*`, `OD-155-*` синхронизирован между `sprint_s5`, `epic_s5` и Day1 epic; `BLK-155-01..02` закрыты, `OD-155-01..03` утверждены (2026-02-25).

## Актуализация по Issue #170 (`run:plan`, 2026-02-25)
- Добавлен Day2 execution-артефакт `docs/delivery/epics/s5/epic-s5-day2-launch-profiles-dev-execution.md` для single-epic реализации FR-053/FR-054.
- Зафиксированы quality-gates QG-D2-01..QG-D2-05 и DoD-пакет для handover в `run:dev`.
- Создана implementation issue #171; связь `#170 -> #171 -> FR-053/FR-054` синхронизирована в `issue_map` и Sprint S5 docs.

## Актуализация по Issue #175 (`run:dev`, 2026-02-25)
- Для FR-026/FR-027 зафиксировано исключение в label policy: `need:reviewer` на PR (`pull_request:labeled`) запускает reviewer-run для ручного pre-review.
- Обновлены связные документы: `README.md`, `docs/product/{requirements_machine_driven,labels_and_trigger_policy,agents_operating_model,stage_process_model}.md`, `docs/architecture/api_contract.md`.

## Актуализация по Issue #171 (`run:dev`, 2026-02-25)
- В `runstatus` реализована typed next-step action matrix:
  `action_kind`, `target_label`, `display_variant`, `url`.
- Добавлен deterministic profile resolver для next-step матрицы (baseline `quick-fix|feature|new-service`) и preview/execute path через staff API.
- Для ambiguity/not-resolved review-stage сценариев добавлен hard-stop remediation:
  автоматическая постановка `need:input` через runstatus/webhook path до публикации warning comment.
- Проверка изменений зафиксирована unit-пакетом:
  `go test ./services/internal/control-plane/internal/domain/runstatus ./services/internal/control-plane/internal/domain/webhook`.

## Актуализация по Issue #212 (`run:intake`, 2026-02-27)
- Для FR-026/FR-028/FR-033/FR-036/FR-040/FR-043/FR-045 и NFR-010/NFR-013/NFR-017/NFR-018 добавлен Sprint S7 intake traceability пакет:
  `docs/delivery/sprints/s7/sprint_s7_mvp_readiness_gap_closure.md`,
  `docs/delivery/epics/s7/epic_s7.md`,
  `docs/delivery/epics/s7/epic-s7-day1-mvp-readiness-intake.md`,
  `docs/delivery/development_process_requirements.md`,
  `docs/product/labels_and_trigger_policy.md`,
  `docs/product/stage_process_model.md`.
- Intake зафиксировал фактические MVP gaps:
  - `comingSoon`/scaffold контур в staff UI (`navigation.ts` + профильные TODO-страницы);
  - S6 dependency-chain: `#199/#201` закрыты, release closeout выполнен в Issue `#262`, активный continuity-блокер перенесён в `#263` (`run:postdeploy`);
  - отсутствие подтверждённого run-evidence для `run:doc-audit` в текущем delivery-цикле.
- Для всех открытых owner-замечаний PR #213 выставлен статус `fix_required`; замечания сгруппированы по приоритету `behavior/data -> quality/style`.
- В backlog S7 добавлены 18 candidate execution-эпиков (`S7-E01..S7-E18`) с owner-aligned handover в `run:vision`:
  rebase/mainline hygiene, UI cleanup (navigation/sections/filter), agents de-scope + repo-only prompt policy для MVP, runs/deploy UX, `mode:discussion` reliability, late-stage `run:<stage>:revise` coverage, QA DNS acceptance-policy, `run:intake:revise` status consistency, `run:self-improve` session reliability, финальный readiness gate.
- Для стандартизации качества backlog зафиксировано требование PMO из Issue `#210`:
  формулировка задач в формате user story и обязательный блок edge cases для QA-ready acceptance.
- Для процессного governance добавлен единый стандарт:
  - заголовков и body для Issue/PR по stage/role;
  - информационной архитектуры проектной документации (каталоги `product/architecture/delivery/ops/templates`);
  - ролевой матрицы обязательных шаблонов документации.

## Актуализация по Issue #223 (`run:plan`, 2026-02-27)
- Для FR-002/FR-004/FR-033 и NFR-002/NFR-010/NFR-018 добавлен execution-governance пакет Sprint S8 Day1:
  `docs/delivery/epics/s8/epic-s8-day1-go-refactoring-plan.md`,
  `docs/delivery/epics/s8/epic_s8.md`,
  `docs/delivery/sprints/s8/sprint_s8_go_refactoring_parallelization.md`,
  `docs/delivery/delivery_plan.md`.
- Выполнен plan-аудит Go-кода по сервисам и библиотекам (oversized files, duplicate hotspots, database access alignment).
- Создан параллельный implementation backlog `#225..#230`:
  - `#225` control-plane decomposition;
  - `#226` api-gateway transport cleanup;
  - `#227` worker decomposition;
  - `#228` agent-runner helper normalization;
  - `#229` shared libs pgx/servicescfg alignment;
  - `#230` cross-service hygiene closure.
- Через Context7 (`/websites/cli_github_manual`) подтвержден актуальный CLI-синтаксис `gh issue create`/`gh pr create`/`gh pr edit`; новые внешние зависимости не добавлялись.

## Актуализация по Issue #281 (`run:dev`, planned 2026-03-06)
- Для FR-020/FR-033/FR-049/FR-050 и NFR-010/NFR-018 добавлен execution backlog Sprint S8 Day2:
  `docs/delivery/epics/s8/epic-s8-day2-empty-repository-initialization.md`,
  `docs/delivery/epics/s8/epic_s8.md`,
  `docs/delivery/sprints/s8/sprint_s8_go_refactoring_parallelization.md`,
  `docs/delivery/delivery_plan.md`,
  `docs/delivery/issue_map.md`.
- Зафиксирован отдельный onboarding path для пустого GitHub-репозитория:
  deterministic detection -> bootstrap bundle -> direct initial commit в default branch -> onboarding summary issue.
- В traceability baseline закреплено, что empty-repo onboarding:
  - не использует PR как обязательный init-механизм до появления первой ветки/коммита;
  - обязан создавать typed `services.yaml` и docs scaffold как стартовую точку для stage-flow;
  - обязан быть идемпотентным и сохранять audit trail по созданным файлам и bootstrap SHA.
- Источники требований и решений:
  - FR-020 (per-repo `services.yaml`);
  - FR-049 (repository onboarding preflight);
  - FR-050 (docs tree в prompt context);
  - ADR-0007 (multi-repo composition/docs federation как будущий режим).

## Актуализация по Issue #282 (`run:dev`, planned 2026-03-06)
- Для FR-002/FR-020/FR-033/FR-049/FR-050 и NFR-010/NFR-018 добавлен execution backlog Sprint S8 Day3:
  `docs/delivery/epics/s8/epic-s8-day3-existing-repository-adoption.md`,
  `docs/delivery/epics/s8/epic_s8.md`,
  `docs/delivery/sprints/s8/sprint_s8_go_refactoring_parallelization.md`,
  `docs/delivery/delivery_plan.md`,
  `docs/delivery/issue_map.md`.
- Зафиксирован целевой adoption path для существующего репозитория с кодом:
  repository classification -> deterministic scan report -> dedicated onboarding-task -> PR с `services.yaml` и docs baseline.
- В traceability baseline закреплены ограничения безопасности и качества:
  - onboarding-task имеет ограниченный writable scope и не должна делать произвольный рефакторинг приложения;
  - adoption выполняется только через reviewable PR;
  - rerun должен быть идемпотентным на уровне onboarding branch/PR для одного и того же `ref`.
- Источники требований и решений:
  - FR-002 (provider abstraction);
  - FR-020 (per-repo `services.yaml`);
  - FR-049 (repository onboarding preflight);
  - FR-050 (role-aware docs context);
  - ADR-0007 и S4 Day1 (multi-repo compose/docs federation).

## Актуализация по Issue #227 (`run:dev`, 2026-02-28)
- Для FR-033 и NFR-018 выполнена декомпозиция worker orchestration-сервиса без изменения продуктового поведения:
  `services/jobs/worker/internal/domain/worker/service.go` разделён на
  `service_queue_cleanup.go`, `service_queue_lifecycle.go`, `service_queue_dispatch.go`, `service_queue_finalize.go`.
- Для сокращения повторов namespace-resolution добавлен package-level helper:
  `services/jobs/worker/internal/domain/worker/service_queue_helpers.go` (`applyPreparedNamespace`),
  и обновлён recovery-путь в `job_not_found_recovery.go`.
- Поведенческое покрытие сохранено: пройдены проверки
  `go test ./services/jobs/worker/internal/domain/worker/...` и `go test ./services/jobs/worker/...`.
- По checklist gate выполнен `make lint-go`.
- `make dupl-go` зафиксировал pre-existing дубли вне scope текущего issue (в `control-plane` и `api-gateway`);
  для изменённого набора файлов `worker` локальная проверка `dupl` не выявила новых дублей.
- Трассируемость синхронизирована с `docs/delivery/issue_map.md` (добавлена строка по Issue `#227`).

## Актуализация по Issue #229 (`run:dev`, 2026-02-28)
- Для FR-004/FR-033 и NFR-018 выполнено выравнивание shared Go-библиотек в bounded scope `S8-E05`:
  `libs/go/postgres` и `libs/go/servicescfg`.
- В `libs/go/postgres` закреплён pgx-native baseline:
  - `OpenPGXPool` остаётся основным API для нового кода;
  - `Open` переведён в explicit compatibility-wrapper `OpenSQLDB` с `Deprecated`-пометкой;
  - добавлен unit coverage (`db_test.go`) для normalization/DSN helper-функций.
- В `libs/go/servicescfg` выполнена модульная декомпозиция без изменения поведения:
  `load.go` разделён на тематические файлы `load_namespace.go`, `load_validation.go`,
  `load_components.go`, `load_context.go`, `load_imports.go`, `load_helpers.go`.
- Релевантный дизайн-гайд обновлён:
  `docs/design-guidelines/go/infrastructure_integration_requirements.md` теперь явно фиксирует правило
  `pgxpool` по умолчанию и `database/sql` только как compatibility-path.
- Проверки по изменённому scope:
  `go test ./libs/go/servicescfg ./libs/go/postgres/...`, `go test ./...`, `make lint-go`.
- `make dupl-go` фиксирует pre-existing дубли вне scope текущего issue (в `control-plane` и `api-gateway`).
- Трассируемость синхронизирована с `docs/delivery/issue_map.md` (добавлена строка по Issue `#229`).

## Актуализация по Issue #230 (`run:dev`, 2026-02-28)
- Для FR-002/FR-004/FR-033 и NFR-002/NFR-010/NFR-018 выполнен финальный consolidating stream `S8-E06`:
  cross-service hygiene closure и residual debt report.
- Удалены low-risk дубли в `control-plane`/`worker`/`libs`:
  - вынесен общий helper `libs/go/registry/image_ref.go` и удалено дублирование `extractRegistryRepositoryPath/splitImageRef`;
  - добавлен общий helper ожидания job с логами ошибок (`waitForJobCompletionWithFailureLogs`) для `runtimedeploy` build/repo-sync path;
  - унифицирован gRPC mapping `RepositoryBinding`/`ConfigEntry` через package-level helper-caster'ы;
  - конструктор `staff.Service` переведён на `staff.Dependencies` для устранения сигнатурной дубликации.
- `tools/lint/dupl-baseline.txt` синхронизирован с текущим кодом:
  baseline сокращён с `62` до `43` строк, удалены устаревшие записи и зафиксированы только актуальные residual duplicates.
- Подготовлен consolidated отчёт:
  `docs/delivery/epics/s8/epic-s8-e06-go-hygiene-closure-report.md`
  (self-check по `common/go` чек-листам, residual debt backlog с приоритетами и owner-decision предложениями).
- Проверки по изменённому scope:
  `make dupl-go`, `make lint-go`, `go test ./services/internal/control-plane/...`,
  `go test ./services/jobs/worker/...`, `go test ./libs/go/registry/...`.
- Трассируемость синхронизирована с `docs/delivery/issue_map.md` (добавлены строки по `#226`, `#228`, `#230`).

## Актуализация по Issue #218 (`run:vision`, 2026-02-27)
- Для FR-026/FR-028/FR-033/FR-045/FR-052/FR-053/FR-054 и NFR-010/NFR-018 добавлен vision traceability пакет Sprint S7:
  `docs/delivery/epics/s7/epic-s7-day2-mvp-readiness-vision.md`,
  `docs/delivery/epics/s7/epic_s7.md`,
  `docs/delivery/sprints/s7/sprint_s7_mvp_readiness_gap_closure.md`,
  `docs/delivery/delivery_plan.md`,
  `docs/delivery/issue_map.md`.
- Vision-stage формализовал measurable KPI для всех execution-потоков `S7-E01..S7-E18` и зафиксировал baseline по каждому потоку:
  `user story + acceptance criteria + edge cases + expected evidence`.
- В vision baseline добавлена owner policy для MVP: custom agents/prompt lifecycle вынесены в post-MVP, prompt templates обслуживаются по repo workflow.
- Введено обязательное governance-правило decomposition parity перед `run:dev`:
  `approved_execution_epics_count == created_run_dev_issues_count` (coverage ratio = `1.0`).
- Для stage continuity создана follow-up issue `#220` (`run:prd`) без trigger-лейбла; в issue передан обязательный шаблон создания следующей stage-задачи (`run:arch`).
- Scope этапа сохранён policy-safe: markdown-only изменения без модификации code/runtime артефактов.

## Актуализация по Issue #220 (`run:prd`, 2026-02-27)
- Для FR-026/FR-028/FR-033/FR-045/FR-052/FR-053/FR-054 и NFR-010/NFR-018 добавлен PRD traceability пакет Sprint S7:
  `docs/delivery/epics/s7/epic-s7-day3-mvp-readiness-prd.md`,
  `docs/delivery/epics/s7/prd-s7-day3-mvp-readiness-gap-closure.md`,
  `docs/delivery/epics/s7/epic_s7.md`,
  `docs/delivery/sprints/s7/sprint_s7_mvp_readiness_gap_closure.md`,
  `docs/delivery/delivery_plan.md`,
  `docs/delivery/issue_map.md`.
- PRD-stage формализовал stream-level execution contract для `S7-E01..S7-E18`:
  `user story + FR + AC + NFR + edge cases + expected evidence + dependencies`.
- Зафиксированы deterministic sequencing и dependency graph для перехода `run:prd -> run:arch -> run:design -> run:plan`.
- В PRD явным контуром зафиксирован `repo-only` policy для prompt templates на MVP и de-scope custom agents/prompt lifecycle.
- Подтверждено governance-правило decomposition parity перед `run:dev`:
  `approved_execution_epics_count == created_run_dev_issues_count` (coverage ratio = `1.0`, блокировка при mismatch).
- Для stage continuity создана follow-up issue `#222` (`run:arch`) без trigger-лейбла; в handover переданы PRD-пакет, sequencing-ограничения и parity-gate правила.
- Scope этапа сохранён policy-safe: markdown-only изменения без модификации code/runtime артефактов.

## Актуализация по Issue #222 (`run:arch`, 2026-03-02)
- Для FR-026/FR-028/FR-033/FR-053/FR-054 и NFR-010/NFR-018 добавлен architecture traceability пакет Sprint S7:
  `docs/delivery/epics/s7/epic-s7-day4-mvp-readiness-arch.md`,
  `docs/architecture/s7_mvp_readiness_gap_closure_architecture.md`,
  `docs/architecture/c4_context_s7_mvp_readiness_gap_closure.md`,
  `docs/architecture/c4_container_s7_mvp_readiness_gap_closure.md`,
  `docs/architecture/adr/ADR-0010-s7-mvp-readiness-stream-boundaries-and-parity-gate.md`,
  `docs/architecture/alternatives/ALT-0002-s7-mvp-readiness-stream-architecture.md`,
  `docs/delivery/epics/s7/epic_s7.md`,
  `docs/delivery/sprints/s7/sprint_s7_mvp_readiness_gap_closure.md`,
  `docs/delivery/delivery_plan.md`,
  `docs/delivery/issue_map.md`.
- На architecture-stage зафиксированы:
  - ownership matrix и сервисные границы по `S7-E01..S7-E18`;
  - deterministic wave-sequencing для перехода `run:arch -> run:design -> run:plan`;
  - parity-gate перед `run:dev`: `approved_execution_epics_count == created_run_dev_issues_count`.
- Для stage continuity создана follow-up issue `#238` (`run:design`) без trigger-лейбла с обязательным handover на подготовку `design_doc/api_contract/data_model/migrations_policy`.
- Через Context7 подтверждён baseline для инструментов stage-handover и C4-артефактов:
  `/websites/cli_github_manual` (актуальный `gh issue/pr` синтаксис) и `/mermaid-js/mermaid` (валидный C4 синтаксис).
- Scope этапа сохранён policy-safe: markdown-only изменения без модификации code/runtime артефактов.

## Актуализация по Issue #238 (`run:design`, 2026-03-02)
- Для FR-026/FR-028/FR-033/FR-053/FR-054 и NFR-010/NFR-018 добавлен design traceability пакет Sprint S7:
  `docs/delivery/epics/s7/epic-s7-day5-mvp-readiness-design.md`,
  `docs/architecture/s7_mvp_readiness_gap_closure_design_doc.md`,
  `docs/architecture/s7_mvp_readiness_gap_closure_api_contract.md`,
  `docs/architecture/s7_mvp_readiness_gap_closure_data_model.md`,
  `docs/architecture/s7_mvp_readiness_gap_closure_migrations_policy.md`,
  `docs/delivery/epics/s7/epic_s7.md`,
  `docs/delivery/sprints/s7/sprint_s7_mvp_readiness_gap_closure.md`,
  `docs/delivery/delivery_plan.md`,
  `docs/delivery/issue_map.md`.
- На design-stage зафиксированы typed contract decisions для потоков:
  `S7-E06`, `S7-E07`, `S7-E09`, `S7-E10`, `S7-E13`, `S7-E16`, `S7-E17`.
- Зафиксированы persisted-state изменения и migration/rollback политика:
  `runtime_deploy_tasks`, `agent_runs`, `agent_sessions` (+ flow-events payload hardening).
- Через Context7 подтверждён dependency baseline и актуальная документация:
  `/getkin/kin-openapi` (OpenAPI request/response validation path),
  `/microsoft/monaco-editor` (DiffEditor API `createDiffEditor`/`setModel`).
- Новые внешние зависимости не добавлялись; каталог зависимостей не требует обновления.
- Для stage continuity создана follow-up issue `#241` (`run:plan`) без trigger-лейбла.
- Scope этапа сохранён policy-safe: markdown-only изменения без модификации code/runtime артефактов.

## Актуализация по Issue #241 (`run:plan`, 2026-03-02)
- Для FR-026/FR-028/FR-033/FR-053/FR-054 и NFR-010/NFR-018 добавлен plan traceability пакет Sprint S7:
  `docs/delivery/epics/s7/epic-s7-day6-mvp-readiness-plan.md`,
  `docs/delivery/epics/s7/epic_s7.md`,
  `docs/delivery/sprints/s7/sprint_s7_mvp_readiness_gap_closure.md`,
  `docs/delivery/delivery_plan.md`,
  `docs/delivery/issue_map.md`.
- На plan-stage сформирован execution package для `S7-E01..S7-E18` с wave-sequencing и quality-gates перед `run:dev`.
- По owner-уточнению в Issue `#241` создана отдельная implementation issue на каждый поток:
  `#243..#260` (один issue на `S7-E01..S7-E18`), без trigger-лейблов `run:*`.
- Подтверждено decomposition parity-правило перед входом в `run:dev`:
  `approved_execution_epics_count == created_run_dev_issues_count` (`18 == 18`).
- Через Context7 (`/websites/cli_github_manual`) подтверждён актуальный неинтерактивный синтаксис `gh issue create` / `gh pr create` / `gh pr edit`; новые внешние зависимости не добавлялись.
- Scope этапа сохранён policy-safe: markdown-only изменения без модификации code/runtime артефактов.

## Актуализация по Issue #243 (`run:dev`, 2026-03-02)
- Для FR-026/FR-028/FR-033 и NFR-010/NFR-018 реализован foundation stream `S7-E01`:
  зафиксирован единый deterministic rebase/mainline процесс для revise-итераций в `run:dev`.
- Обновлён process source-of-truth:
  `docs/delivery/development_process_requirements.md` закрепил обязательный порядок
  `git fetch -> git rebase origin/main -> conflict-marker check -> checks -> git push --force-with-lease`,
  запрет `git merge origin/main` для revise-веток и обязательный PR rebase-checklist.
- Актуализирована traceability-матрица Sprint S7:
  `docs/delivery/issue_map.md` (выделен отдельный статус issue `#243`, остаток backlog перенесён в диапазон `#245..#260` после закрытия `#244`).
- Через Context7 (`/websites/git-scm`) подтверждены актуальные команды `git rebase --continue|--abort`
  и безопасный push-path `git push --force-with-lease` для rebased PR-веток.
- Новые внешние зависимости не добавлялись; изменения ограничены markdown/process governance контуром.

## Актуализация по Issue #244 (`run:dev`, 2026-03-05)
- Для FR-026/FR-028/FR-033 и NFR-010/NFR-018 реализован stream `S7-E02`:
  из staff sidebar удалены non-MVP navigation entries (`governance`, `admin`, `configuration/docs`, `configuration/mcp-tools`, `configuration/agents`, `configuration/config-entries`).
- В `services/staff/web-console/src/router/routes.ts` удалены связанные non-MVP маршруты и добавлен fallback redirect на `projects` для stale deep-links,
  чтобы после cleanup не возникало broken transitions.
- Удалён связанный dead code:
  страницы `pages/governance/*`, `pages/admin/*`, `pages/configuration/{DocsKnowledgePage,McpToolsPage}.vue`,
  UI-контур `config-entries` (страница + feature-слой) и platform-tokens scaffold в `System settings`.
- Актуализирована Sprint S7 traceability:
  `docs/delivery/issue_map.md`, `docs/delivery/sprints/s7/sprint_s7_mvp_readiness_gap_closure.md`, `docs/delivery/epics/s7/epic_s7.md`.
- По owner request оформлен follow-up issue `#274` (`S7-E19`) на backend cleanup Agents/Configs/Secrets,
  синхронизированы sprint/epic/plan документы.
- Проверки по scope:
  `npm --prefix services/staff/web-console run build`, route inventory search (`rg -n` по удалённым route names),
  smoke-check MVP навигации (`projects`, `project-repositories`, `project-members`, `runs`, `runtime-deploy/tasks`,
  `wait-queue`, `approvals`, `system-settings`, `users`) и отсутствие broken links на удалённых
  `/runtime-deploy/images` и `/running-jobs`.

## Актуализация по Issue #245 (`run:dev`, 2026-03-05)
- Для FR-026/FR-028/FR-033 и NFR-010/NFR-018 реализован stream `S7-E03`:
  удалён глобальный filter-entry в app shell вместе с зависимым UI summary/reset контуром.
- В `services/staff/web-console/src/features/ui-context/store.ts` удалено глобальное состояние `env/namespace`
  и связанный cookie-persistence path; сохранён только selected project context, нужный для MVP-навигации.
- В `services/staff/web-console/src/pages/operations/RuntimeDeployTasksPage.vue`
  загрузка списка отвязана от `uiContext.env`, чтобы глобальный фильтр больше не влиял на list/load поведение.
- Удалён связанный неиспользуемый компонент `services/staff/web-console/src/shared/ui/AdminClusterContextBar.vue`,
  очищены i18n-ключи глобального фильтра в `services/staff/web-console/src/i18n/messages/{en,ru}.ts`.
- Актуализирована Sprint S7 traceability:
  `docs/delivery/issue_map.md`, `docs/delivery/sprints/s7/sprint_s7_mvp_readiness_gap_closure.md`, `docs/delivery/epics/s7/epic_s7.md`;
  remaining backlog нормализован как `#246..#260` + post-plan `#274`.
- Проверки по scope:
  cleanup-поиск `rg -n` по `Global filter`/`uiContext.env`/`uiContext.namespace`,
  `npm --prefix services/staff/web-console run build`.

## Актуализация по Issue #246 (`run:dev`, 2026-03-09)
- Для FR-E04-1/FR-E04-2 и NFR-E04-1 stream `S7-E04` финализирован без нового redirect-кода:
  owner-review подтвердил, что после удаления UI-контура `runtime-deploy/images`
  отдельный redirect для `/runtime-deploy/images*` не нужен.
- В `services/staff/web-console/src/router/routes.ts` сохраняется только общий fallback
  `/:pathMatch(.*)* -> projects`; stale URL `/runtime-deploy/images*` попадает в него
  после cleanup `#244`, поэтому удалённый раздел не возвращается в MVP navigation.
- Нормализована Sprint S7 traceability:
  `docs/delivery/issue_map.md`, `docs/delivery/sprints/s7/sprint_s7_mvp_readiness_gap_closure.md`, `docs/delivery/epics/s7/epic_s7.md`;
  remaining backlog нормализован как `#247..#260` + post-plan `#274`.
- Через Context7 (`/vuejs/router`) подтверждён актуальный синтаксис Vue Router 4 для catch-all path
  `:pathMatch(.*)*`, которым закрываются удалённые пути без отдельного route record.
- Проверки по scope:
  `rg -n "pathMatch" services/staff/web-console/src/router/routes.ts`,
  `npm --prefix services/staff/web-console run build`.

## Актуализация по Issues #247 / #248 / #249 (`run:dev`, 2026-03-09)
- Для FR-009/FR-029/FR-030/FR-035/FR-037 и NFR-010 выполнен combined cleanup/doco-sync pass по потокам `S7-E05`, `S7-E06`, `S7-E07`.
- Подтверждено фактическое MVP-состояние после ранее выполненных `#244` и `#274`:
  - UI-раздел `Agents` больше не входит в MVP navigation;
  - runtime mode/locale agent settings не редактируются через staff UI/API;
  - prompt templates работают только по repo-seed policy без selector `repo|db`.
- В коде удалены остаточные stale references:
  - agent-related i18n scaffold keys;
  - тест `resolvePathUnescaped`, привязанный к удаленному `/staff/prompt-templates/*` path;
  - мертвые proto messages и HTTP DTO/caster модели старого `Agents/PromptTemplates` staff API.
- В `worker` добавлен unit-test, фиксирующий инвариант:
  - `PromptTemplateSource == repo_seed`;
  - `PromptTemplateLocale` берется из platform default worker config.
- Обновлены source-of-truth документы:
  `docs/product/agents_operating_model.md`,
  `docs/product/requirements_machine_driven.md`,
  `docs/product/constraints.md`,
  `docs/product/brief.md`,
  `docs/architecture/prompt_templates_policy.md`,
  `docs/architecture/api_contract.md`,
  `docs/architecture/data_model.md`,
  а также Sprint S7 traceability docs.
- Combined cleanup `#247/#248/#249` вместе с backend cleanup `#274` также поглотили standalone streams `S7-E08/#250` и `S7-E15/#257`: отдельные `run:dev` для них больше не требуются.
- Remaining standalone backlog Sprint S7 после этой актуализации нормализован как `#251..#256`, `#258..#260`.

## Актуализация по Issue #251 (`run:dev`, 2026-03-10)
- Для FR-012/FR-040 и NFR-010 реализован stream `S7-E09`:
  в `services/staff/web-console/src/pages/RunsPage.vue` удалена колонка `run type`,
  чтобы список запусков показывал только релевантные поля операционной диагностики.
- В `services/staff/web-console/src/pages/RunDetailsPage.vue`
  delete namespace action больше не зависит от `job_exists`;
  UI показывает action при наличии известного namespace и переиспользует существующий typed endpoint
  `DELETE /api/v1/staff/runs/{run_id}/namespace`, который уже идемпотентно обрабатывает повторный delete (`already_deleted=true`).
- В `services/staff/web-console/src/pages/operations/WaitQueuePage.vue` и
  `services/staff/web-console/src/i18n/messages/{en,ru}.ts`
  пользовательские подписи нормализованы как `trigger kind` / `вид триггера`,
  чтобы в MVP UI не оставался термин `run type`.
- Актуализирована Sprint S7 traceability:
  `docs/delivery/issue_map.md`, `docs/delivery/sprints/s7/sprint_s7_mvp_readiness_gap_closure.md`, `docs/delivery/epics/s7/epic_s7.md`;
  remaining backlog нормализован как `#252..#256`, `#258..#260`.
- Проверки по scope:
  cleanup-поиск `rg -n "Run type|Тип запуска" services/staff/web-console/src`,
  `npm --prefix services/staff/web-console run build`.

## Актуализация по Issue #252 (`run:dev`, 2026-03-10)
- Для FR-012/FR-038/FR-040 и NFR-010/NFR-018 реализован stream `S7-E10`:
  contract-first расширены `services/external/api-gateway/api/server/api.yaml` и
  `proto/codexk8s/controlplane/v1/controlplane.proto`, добавлены typed HTTP/gRPC actions
  `cancel/stop` и синхронно обновлены backend/frontend codegen артефакты.
- В control-plane расширена persisted-модель `runtime_deploy_tasks`:
  добавлены control/audit поля `cancel_requested_*`, `stop_requested_*`,
  `terminal_status_source`, `terminal_event_seq`, новая миграция
  `services/internal/control-plane/cmd/cli/migrations/20260310110000_day26_runtime_deploy_task_controls.sql`
  и идемпотентный repository path `RequestAction`.
- В доменном слое `runtimedeploy` и staff use-case добавлены guardrails:
  повторный `cancel/stop` возвращает идемпотентный результат,
  `stop` разрешён только для `running` задачи с активным lease,
  операторские actions пишут audit events в `flow_events`,
  а reconcile-loop быстрее прерывает текущий deploy flow после terminal cancel.
- В staff UI на странице деталей deploy task добавлены кнопки `cancel/stop`,
  confirm dialog с optional reason, success/error feedback, отображение новых audit-полей
  и обработка `failed_precondition` для операторского сценария.
- Актуализированы traceability документы Sprint S7:
  `docs/delivery/issue_map.md`,
  `docs/delivery/sprints/s7/sprint_s7_mvp_readiness_gap_closure.md`,
  `docs/delivery/epics/s7/epic_s7.md`;
  remaining backlog нормализован как `#254..#256`, `#258..#260`.
- Проверки по scope:
  `make gen-proto-go`,
  `make gen-openapi`,
  `go test ./services/internal/control-plane/... ./services/external/api-gateway/...`,
  `npm --prefix services/staff/web-console run build`.

## Актуализация по Issue #255 (`run:dev`, 2026-03-11)
- Для FR-026/FR-028/FR-030/FR-033/FR-052 и NFR-010/NFR-011/NFR-018 реализован stream `S7-E13`:
  revise-loop для недостающих late-stage labels `run:doc-audit|qa|release|postdeploy|ops|self-improve:revise`
  доведён до фактического рабочего состояния в code path.
- В label/runtime source-of-truth добавлены отсутствовавшие revise-элементы:
  typed trigger kinds `doc_audit_revise`, `qa_revise`, `release_revise`, `postdeploy_revise`, `ops_revise`, `self_improve_revise` в `libs/go/domain/webhook`,
  env labels `CODEXK8S_RUN_*_REVISE_LABEL` в `services/internal/control-plane/internal/app/config.go`,
  runtime defaults в `services/internal/control-plane/internal/domain/runtimedeploy/service_defaults.go`,
  а также revise-поля в `TriggerLabels` / next-step label catalog path.
- В control-plane закрыт resolver/runtime gap:
  `services/internal/control-plane/internal/domain/webhook/pull_request_review_resolver.go`
  теперь резолвит PR review `changes_requested` из `run:doc-audit|qa|release|postdeploy|ops|self-improve`
  в детерминированный trigger `run:<stage>:revise`;
  `resolveRunAgentKey` направляет такой run в корректную stage-role;
  `services/internal/control-plane/internal/domain/runstatus/next_step_actions.go`
  строит next-step матрицу для этих revise trigger kinds без потери stage context.
- В `agent-runner` multi-stage revise синхронизирован с policy исполнения:
  `prompt_seed_mapping.go` сохраняет stage mapping для всех новых revise seed'ов,
  `helpers_prompt_docs.go` оставляет full-env prompt docs env = `ai` для `ops_revise`,
  `write_scope_policy.go` сохраняет markdown-only scope для doc-stage revise и restricted scope для `self_improve_revise`.
- Добавлены unit-тесты на ключевые сценарии:
  PR review с label `run:<stage>` создаёт run с trigger `run:<stage>:revise` и корректным агентом для late-stage revise-loop,
  `resolveRunAgentKey` для новых revise trigger kinds,
  mapping next-step stage descriptor,
  runner prompt/docs/write-scope policy.
- Через Context7 (`/caarlos0/env`) перепроверен актуальный baseline для env-tag конфигурации `env`/`envDefault`;
  новые внешние зависимости не добавлялись.

## Актуализация по Issue #256 (`run:dev`, 2026-03-11)
- Для FR-028/FR-033 и NFR-010 реализован stream `S7-E14`:
  `docs/ops/production_runbook.md` закрепил QA policy, по которой новые и изменённые HTTP-ручки
  проверяются через Kubernetes service DNS path (`<service>.<namespace>.svc.cluster.local`),
  а browser/OAuth flow больше не считается единственным acceptance gate.
- Синхронно обновлены QA role templates:
  `docs/templates/test_strategy.md`,
  `docs/templates/test_plan.md`,
  `docs/templates/test_matrix.md`,
  `docs/templates/regression_checklist.md`.
  Теперь они требуют DNS evidence bundle:
  namespace, service FQDN, точную команду, HTTP status, excerpt ответа и `kubectl`-диагностику при fail.
- Full-env verification выполнен в namespace `codex-k8s-dev-1`:
  `kubectl config view --minify -o jsonpath='{..namespace}'`,
  `kubectl get svc -o wide`,
  `getent hosts codex-k8s.codex-k8s-dev-1.svc.cluster.local`,
  `curl -sS ... /healthz` (`200`),
  `curl -sS ... /api/v1/auth/me` (`401`),
  `curl -sS -X POST ... /api/v1/webhooks/github` (`400`).
- Актуализированы traceability документы Sprint S7:
  `docs/delivery/issue_map.md`,
  `docs/delivery/delivery_plan.md`,
  `docs/delivery/sprints/s7/sprint_s7_mvp_readiness_gap_closure.md`,
  `docs/delivery/epics/s7/epic_s7.md`;
  remaining backlog нормализован как `#254`, `#258..#260`.
- Проверки по scope:
  `kubectl config view --minify -o jsonpath='{..namespace}'`,
  `kubectl get svc -o wide`,
  `getent hosts codex-k8s.codex-k8s-dev-1.svc.cluster.local`,
  `curl -sS -o /tmp/codex-health.out -D /tmp/codex-health.headers -w '%{http_code}\n' http://codex-k8s.codex-k8s-dev-1.svc.cluster.local/healthz`,
  `curl -sS -o /tmp/codex-authme.out -D /tmp/codex-authme.headers -w '%{http_code}\n' http://codex-k8s.codex-k8s-dev-1.svc.cluster.local/api/v1/auth/me`,
  `curl -sS -o /tmp/codex-webhook.out -D /tmp/codex-webhook.headers -w '%{http_code}\n' -X POST http://codex-k8s.codex-k8s-dev-1.svc.cluster.local/api/v1/webhooks/github`.

## Актуализация по Issue #258 (`run:dev`, 2026-03-11)
- Для FR-026/FR-028/FR-033 и NFR-010/NFR-018 реализован stream `S7-E16`:
  в `services/internal/control-plane/internal/domain/runstatus`
  нормализована логика выбора и слияния run-status comment state,
  чтобы поздний duplicate/stale update не мог понизить фактически успешный terminal status.
- Добавлен приоритет terminal-state для service-message:
  `succeeded > failed > running/pending`, а при равном status rank
  сохраняется более поздняя lifecycle phase и только затем fallback к `comment_id`.
- Dedupe run-status comments теперь выбирает канонический comment по state precedence,
  а не только по максимальному GitHub `comment_id`,
  что устраняет false-failed при дублирующих terminal updates.
- Добавлено регрессионное unit-покрытие для:
  - merge terminal states (`succeeded` не понижается до `failed`);
  - защиты terminal state от позднего non-terminal update;
  - выбора canonical comment при duplicate terminal comments.
- Актуализирована Sprint S7 traceability:
  `docs/delivery/issue_map.md`,
  `docs/delivery/sprints/s7/sprint_s7_mvp_readiness_gap_closure.md`,
  `docs/delivery/epics/s7/epic_s7.md`;
  remaining backlog нормализован как `#254`, `#255`, `#259..#260`.
- Проверки по scope:
  `go test ./services/internal/control-plane/internal/domain/runstatus/...`.

## Актуализация по Issue #259 (`run:dev`, 2026-03-11)
- Для FR-009/FR-036 и NFR-013/NFR-017 реализован stream `S7-E17`:
  persisted snapshot в `agent_sessions` переведён на versioned CAS-like persistence
  с полями `snapshot_version`, `snapshot_checksum`, `snapshot_updated_at`.
- В `services/internal/control-plane/internal/repository/postgres/agentsession`
  write-path больше не перетирает последний non-empty `codex_cli_session_json` пустым retry/update,
  одинаковый replay становится идемпотентным, а stale rewrite возвращает conflict с `actual_snapshot_version`.
- Внутренний gRPC callback `UpsertAgentSession`/`GetLatestAgentSession`
  расширен version/checksum metadata; conflict возвращается через typed gRPC status details,
  а `agent-runner` держит `snapshot_version` текущего run и публикует enriched `run.agent.session.saved`.
- Добавлены migration/runtime артефакты:
  `services/internal/control-plane/cmd/cli/migrations/20260311143000_day28_agent_session_snapshot_versioning.sql`,
  `libs/go/domain/agent/session_snapshot_checksum.go`,
  обновлённые `proto/codexk8s/controlplane/v1/controlplane.proto` и generated Go contracts.
- Актуализированы traceability документы Sprint S7:
  `docs/delivery/issue_map.md`,
  `docs/delivery/delivery_plan.md`,
  `docs/delivery/sprints/s7/sprint_s7_mvp_readiness_gap_closure.md`,
  `docs/delivery/epics/s7/epic_s7.md`;
  remaining backlog нормализован как `#254`, `#255`, `#260`.
- Проверки по scope:
  `make gen-proto-go`,
  `go test ./libs/go/domain/agent ./services/internal/control-plane/... ./services/jobs/agent-runner/...`,
  `make lint-go`,
  `make dupl-go`.

## Актуализация по Issue #260 (`run:dev`, 2026-03-11)
- Для FR-028/FR-033/FR-045 и NFR-010 реализован stream `S7-E18`:
  documentation governance runtime-проекция в `services.yaml`
  синхронизирована с `docs/delivery/development_process_requirements.md`.
- `services.yaml/spec.roleDocTemplates` теперь соответствует role-template matrix:
  - `em` получает полный release/delivery template set;
  - `dev` ограничен `user_story.md` + `definition_of_done.md`;
  - `qa` дополнен `postdeploy_review.md`;
  - `sre` переведён на ops/incident templates без drift;
  - `km` получает delivery/roadmap/docset traceability templates;
  - `reviewer` больше не получает лишние doc templates.
- `services.yaml/spec.projectDocs` теперь гарантирует доступ runtime prompt к
  `docs/delivery` для `dev/qa/sre` и к `docs/ops` для `qa`,
  чтобы role-aware prompt envelope ссылался на релевантные source-of-truth каталоги,
  а не на устаревший урезанный набор.
- В `docs/delivery/development_process_requirements.md`
  явно закреплено, что `services.yaml/spec.roleDocTemplates` и `services.yaml/spec.projectDocs`
  обязаны оставаться синхронными с role-template matrix и doc IA;
  тот же governance contract продублирован в
  `services/jobs/agent-runner/internal/runner/promptseeds/README.md`.
- Добавлены regression-тесты agent-runner,
  которые проверяют реальный repository `services.yaml` на соответствие governance-матрице
  и наличие delivery/ops doc coverage для нужных ролей.
- Актуализированы traceability документы Sprint S7:
  `docs/delivery/issue_map.md`,
  `docs/delivery/delivery_plan.md`,
  `docs/delivery/sprints/s7/sprint_s7_mvp_readiness_gap_closure.md`,
  `docs/delivery/epics/s7/epic_s7.md`;
  remaining standalone backlog нормализован как `#254`.
- Проверки по scope:
  `go test ./services/jobs/agent-runner/internal/runner/...`,
  `go test ./services/jobs/agent-runner/...`.

## Актуализация по Issue #274 (`run:dev`, 2026-03-05)
- Для FR-026/FR-028/FR-033 и NFR-010/NFR-018 реализован stream `S7-E19`:
  выполнен backend cleanup non-MVP контуров `agents`, `prompt templates`, `config entries`,
  `runtime-deploy registry images` и `running jobs`.
- В OpenAPI удалены staff endpoint-ы:
  `/staff/agents*`, `/staff/prompt-templates*`, `/staff/config-entries*`,
  `/staff/runs/jobs`, `/staff/runtime-deploy/images*`; синхронно обновлены backend/frontend codegen артефакты.
- В `control-plane` удалены соответствующие non-MVP RPC из `ControlPlaneService` и
  runtime/domain-реализации staff use-cases (transport + domain + app wiring).
- В control-plane удалены неиспользуемые SQL/repository слои:
  `internal/repository/postgres/prompttemplate/*`,
  `internal/repository/postgres/configentry/*`,
  staff-only части `internal/repository/postgres/agent/*` (list/get/update settings),
  а также связанные domain types/repository contracts.
- Добавлена миграция
  `services/internal/control-plane/cmd/cli/migrations/20260305170000_day27_staff_non_mvp_schema_cleanup.sql`
  для удаления non-MVP схемы:
  `drop table prompt_templates`, `drop table config_entries`,
  `alter table agents drop columns settings/settings_version`.
- В `api-gateway` удалены связанные transport handlers/call-builders;
  для `web-console` очищен dead scaffold (`features/agents`, `configuration/Agents*`).
- Обновлены архитектурные и delivery-артефакты трассировки:
  `docs/architecture/api_contract.md`, `docs/delivery/issue_map.md`, `docs/delivery/requirements_traceability.md`.
- Проверки по scope:
  `go test ./services/external/api-gateway/... ./services/internal/control-plane/...`,
  `go mod tidy`,
  `make lint-go`, `make dupl-go`.

## Актуализация по Issue #225 (`run:dev`, 2026-02-28)
- Для FR-002/FR-033 и NFR-002/NFR-010/NFR-018 выполнен рефакторинг bounded scope `S8-E01`:
  декомпозированы oversized-файлы `webhook/service.go`, `staff/service_methods.go`, `transport/grpc/server.go`
  в тематические smaller units без изменения API/proto/OpenAPI контрактов.
- По правилам размещения кода вынесены helper- и methods-блоки в отдельные файлы:
  - `services/internal/control-plane/internal/domain/webhook/service_helpers.go`;
  - `services/internal/control-plane/internal/domain/staff/service_config_entries.go`;
  - `services/internal/control-plane/internal/domain/staff/service_repository_management.go`;
  - `services/internal/control-plane/internal/domain/staff/service_repository_management_types.go`;
  - `services/internal/control-plane/internal/transport/grpc/server_staff_methods.go`;
  - `services/internal/control-plane/internal/transport/grpc/server_runtime_methods.go`.
- Устранён ad-hoc payload в `RunRepositoryPreflight`: локальный `[]struct{...}` заменён на typed модель `servicesYAMLPreflightEnvSlot`.
- Сохранён единый подход к error mapping на транспортной границе:
  gRPC-преобразование ошибок продолжает выполняться через `toStatus`, без локальных межслойных трансляторов в handlers/domain.
- Проверки по изменённому scope:
  - `go test ./services/internal/control-plane/internal/domain/webhook ./services/internal/control-plane/internal/domain/staff ./services/internal/control-plane/internal/transport/grpc`;
  - `go test ./services/internal/control-plane/...`;
  - `make lint-go` (pass);
  - `make dupl-go` (обнаруживает pre-existing дубли в репозитории, включая исторически повторяющиеся блоки вне scope `#225`).
- Через Context7 (`/grpc/grpc-go`) подтверждена актуальная рекомендация по возврату gRPC-ошибок через `status.Error` и сохранению уже типизированных status-кодов.
