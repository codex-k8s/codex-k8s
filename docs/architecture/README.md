---
doc_id: IDX-CK8S-ARCH-0001
type: domain-index
title: "Architecture Documentation Index"
status: in-review
owner_role: SA
created_at: 2026-03-11
updated_at: 2026-03-11
related_issues: [320]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-11-architecture-index"
---

# Architecture Documentation

## TL;DR
- На корне `docs/architecture/` остаются только канонические архитектурные source-of-truth документы и общие каталоги `adr/`, `alternatives/`, `initiatives/`.
- Исторические stage-specific пакеты вынесены в `docs/architecture/initiatives/`.

## Канонические документы
- `docs/architecture/c4_context.md`
- `docs/architecture/c4_container.md`
- `docs/architecture/api_contract.md`
- `docs/architecture/data_model.md`
- `docs/architecture/agent_runtime_rbac.md`
- `docs/architecture/mcp_approval_and_audit_flow.md`
- `docs/architecture/prompt_templates_policy.md`
- `docs/architecture/multi_repo_mode_design.md`

## Вложенные каталоги
- `docs/architecture/adr/` — архитектурные решения.
- `docs/architecture/alternatives/` — варианты решений и зафиксированные альтернативы.
- `docs/architecture/initiatives/` — исторические initiative/stage-specific пакеты с собственными индексами.

## Навигация по пакетам
- Lifecycle agents/prompt templates: `docs/architecture/initiatives/agents_prompt_templates_lifecycle/README.md`
- Sprint S7 MVP readiness package: `docs/architecture/initiatives/s7_mvp_readiness_gap_closure/README.md`
