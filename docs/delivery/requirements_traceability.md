---
doc_id: TRT-CK8S-0001
type: requirements-traceability
title: "Requirements Traceability Matrix"
status: draft
owner_role: EM
created_at: 2026-02-06
updated_at: 2026-02-06
related_issues: [1]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-06-mvp"
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
| FR-012 | Жизненный цикл run/pod/namespace в БД + UI | `docs/product/requirements_machine_driven.md`, `docs/architecture/c4_container.md`, `docs/architecture/data_model.md` | covered |
| FR-013 | Многоподовость + split service/job zones | `docs/product/requirements_machine_driven.md`, `AGENTS.md`, `docs/design-guidelines/common/project_architecture.md` | covered |
| FR-014 | Слоты через БД | `docs/product/requirements_machine_driven.md`, `docs/architecture/data_model.md` | covered |
| FR-015 | Шаблоны документов в БД + markdown editor | `docs/product/requirements_machine_driven.md`, `docs/architecture/data_model.md`, `docs/architecture/api_contract.md` | covered |
| FR-016 | Bootstrap 2 режима (existing k8s / k3s install) | `docs/product/requirements_machine_driven.md`, `docs/delivery/delivery_plan.md`, `docs/product/brief.md` | covered |
| FR-017 | Project RBAC read/read_write/admin | `docs/product/requirements_machine_driven.md`, `docs/product/constraints.md`, `docs/architecture/data_model.md` | covered |
| FR-018 | No self-signup, email matching | `docs/product/requirements_machine_driven.md`, `docs/product/constraints.md`, `docs/architecture/data_model.md` | covered |
| FR-019 | Добавление пользователей через staff UI | `docs/product/requirements_machine_driven.md`, `docs/architecture/api_contract.md`, `docs/architecture/data_model.md` | covered |
| FR-020 | Multi-repo per project + per-repo services.yaml | `docs/product/requirements_machine_driven.md`, `docs/architecture/data_model.md`, `docs/product/brief.md` | covered |
| FR-021 | Repo token per repository + future Vault/JWT path | `docs/product/requirements_machine_driven.md`, `docs/architecture/data_model.md`, `docs/delivery/roadmap.md` | covered |
| FR-022 | codex-k8s как проект с monorepo services.yaml | `docs/product/requirements_machine_driven.md`, `README.md` | covered |
| FR-023 | Learning mode + educational PR comments | `docs/product/requirements_machine_driven.md`, `docs/product/brief.md`, `docs/architecture/api_contract.md`, `docs/delivery/delivery_plan.md`, `docs/architecture/data_model.md` | covered |
| FR-024 | CODEXK8S_ prefix для env/secrets/CI vars | `docs/product/requirements_machine_driven.md`, `AGENTS.md` | covered |
| FR-025 | MVP public API: only webhook ingress | `docs/product/requirements_machine_driven.md`, `docs/product/constraints.md`, `docs/architecture/api_contract.md` | covered |
| NFR-001 | Security baseline | `docs/product/requirements_machine_driven.md`, `docs/product/constraints.md`, `AGENTS.md` | covered |
| NFR-002 | Multi-pod consistency | `docs/product/requirements_machine_driven.md`, `docs/architecture/c4_container.md`, `docs/architecture/data_model.md` | covered |
| NFR-003 | No event outbox on MVP | `docs/product/requirements_machine_driven.md`, `docs/architecture/data_model.md`, `docs/product/constraints.md` | covered |
| NFR-004 | Embedding vector(3072) | `docs/product/requirements_machine_driven.md`, `docs/architecture/data_model.md`, `docs/product/constraints.md` | covered |
| NFR-005 | Read-replica baseline | `docs/product/requirements_machine_driven.md`, `docs/architecture/c4_container.md`, `docs/product/constraints.md` | covered |
| NFR-006 | One-command staging bootstrap via SSH | `docs/product/requirements_machine_driven.md`, `docs/delivery/delivery_plan.md`, `docs/product/brief.md` | covered |
| NFR-007 | CI/CD model (main->staging, prod gated) | `docs/product/requirements_machine_driven.md`, `docs/product/brief.md`, `docs/product/constraints.md`, `docs/delivery/delivery_plan.md` | covered |
| NFR-008 | MVP storage profile local-path | `docs/product/requirements_machine_driven.md`, `docs/product/constraints.md`, `docs/delivery/delivery_plan.md` | covered |

## Правило актуализации
- Любое новое требование сначала добавляется в `docs/product/requirements_machine_driven.md`, затем отражается в этой матрице.
- Если строка в матрице теряет ссылку на целевой документ, статус меняется на `gap` до устранения.
