---
doc_id: IDX-CK8S-ARCH-S10-0001
type: initiative-index
title: "Initiative Package: s10_mcp_user_interactions"
status: in-review
owner_role: SA
created_at: 2026-03-12
updated_at: 2026-03-13
related_issues: [360, 378, 383, 385, 387, 389]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-03-12-issue-385-arch"
  approved_by: "ai-da-stas"
  approved_at: 2026-03-13
---

# s10_mcp_user_interactions

## TL;DR
- Пакет объединяет Day4 architecture и Day5 design артефакты Sprint S10 для built-in MCP user interactions.
- Внутри зафиксированы C4 overlays, service boundaries, lifecycle ownership, typed tool/callback contracts, interaction data model, wait-state taxonomy и rollout/migration policy.

## Содержимое
- `docs/architecture/initiatives/s10_mcp_user_interactions/README.md`
- `docs/architecture/initiatives/s10_mcp_user_interactions/architecture.md`
- `docs/architecture/initiatives/s10_mcp_user_interactions/c4_context.md`
- `docs/architecture/initiatives/s10_mcp_user_interactions/c4_container.md`
- `docs/architecture/initiatives/s10_mcp_user_interactions/design_doc.md`
- `docs/architecture/initiatives/s10_mcp_user_interactions/api_contract.md`
- `docs/architecture/initiatives/s10_mcp_user_interactions/data_model.md`
- `docs/architecture/initiatives/s10_mcp_user_interactions/migrations_policy.md`

## Связанные source-of-truth документы
- `docs/architecture/api_contract.md`
- `docs/architecture/data_model.md`
- `docs/architecture/mcp_approval_and_audit_flow.md`
- `docs/architecture/adr/ADR-0012-built-in-mcp-user-interactions-control-plane-owned-lifecycle.md`
- `docs/architecture/alternatives/ALT-0004-built-in-mcp-user-interactions-lifecycle-boundaries.md`
- `docs/delivery/epics/s10/epic-s10-day4-mcp-user-interactions-arch.md`
- `docs/delivery/epics/s10/epic-s10-day5-mcp-user-interactions-design.md`
- `docs/delivery/epics/s10/prd-s10-day3-mcp-user-interactions.md`
