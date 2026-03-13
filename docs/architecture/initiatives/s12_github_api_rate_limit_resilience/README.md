---
doc_id: IDX-CK8S-ARCH-S12-0001
type: initiative-index
title: "Initiative Package: s12_github_api_rate_limit_resilience"
status: approved
owner_role: SA
created_at: 2026-03-13
updated_at: 2026-03-13
related_issues: [366, 413, 416, 418, 420, 423, 425, 426, 427, 428, 429, 430, 431]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-03-13-issue-420-design-package"
  approved_by: "ai-da-stas"
  approved_at: 2026-03-13
---

# s12_github_api_rate_limit_resilience

## TL;DR
- Пакет объединяет Day4 architecture и Day5 design артефакты Sprint S12 для GitHub API rate-limit resilience.
- Внутри зафиксированы C4 overlays, ownership split для `control-plane` / `worker` / `agent-runner` / `api-gateway` / `web-console`, lifecycle `detect -> classify -> wait -> resume/manual action`, ADR/alternatives и implementation-ready contracts для wait-state, transport, data model и rollout.
- Plan-stage в Issue `#423` завершил документный контур и декомпозировал execution waves `#425..#431`; этот пакет остаётся source-of-truth для всех implementation streams.

## Содержимое
- `docs/architecture/initiatives/s12_github_api_rate_limit_resilience/README.md`
- `docs/architecture/initiatives/s12_github_api_rate_limit_resilience/architecture.md`
- `docs/architecture/initiatives/s12_github_api_rate_limit_resilience/c4_context.md`
- `docs/architecture/initiatives/s12_github_api_rate_limit_resilience/c4_container.md`
- `docs/architecture/initiatives/s12_github_api_rate_limit_resilience/design_doc.md`
- `docs/architecture/initiatives/s12_github_api_rate_limit_resilience/api_contract.md`
- `docs/architecture/initiatives/s12_github_api_rate_limit_resilience/data_model.md`
- `docs/architecture/initiatives/s12_github_api_rate_limit_resilience/migrations_policy.md`

## Связанные source-of-truth документы
- `docs/architecture/api_contract.md`
- `docs/architecture/data_model.md`
- `docs/architecture/agent_runtime_rbac.md`
- `docs/architecture/mcp_approval_and_audit_flow.md`
- `docs/architecture/adr/ADR-0013-github-rate-limit-controlled-wait-ownership.md`
- `docs/architecture/alternatives/ALT-0005-github-rate-limit-wait-state-boundaries.md`
- `docs/delivery/epics/s12/epic-s12-day4-github-api-rate-limit-arch.md`
- `docs/delivery/epics/s12/epic-s12-day5-github-api-rate-limit-design.md`
- `docs/delivery/epics/s12/epic-s12-day3-github-api-rate-limit-prd.md`
- `docs/delivery/epics/s12/prd-s12-day3-github-api-rate-limit-resilience.md`
- `docs/delivery/epics/s12/epic-s12-day6-github-api-rate-limit-plan.md`

## Continuity after `run:plan`
- Документный контур `intake -> vision -> prd -> arch -> design -> plan` согласован и завершён.
- Owner-managed execution waves `#425..#431` обязаны сохранять этот пакет как архитектурный baseline без drift между delivery и architecture слоями.
- Обязательный sequencing для implementation handover: `#425 -> #426 -> #427 -> #428 -> #429 -> #430 -> #431`.
