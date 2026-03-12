---
doc_id: IDX-CK8S-ARCH-S10-0001
type: initiative-index
title: "Initiative Package: s10_mcp_user_interactions"
status: in-review
owner_role: SA
created_at: 2026-03-12
updated_at: 2026-03-12
related_issues: [360, 378, 383, 385, 387]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-12-issue-385-arch"
---

# s10_mcp_user_interactions

## TL;DR
- Пакет объединяет Day4 architecture-артефакты Sprint S10 для built-in MCP user interactions.
- Внутри зафиксированы C4 overlays, service boundaries, lifecycle ownership, ADR и alternatives по separation between interaction flow и approval flow.

## Содержимое
- `docs/architecture/initiatives/s10_mcp_user_interactions/README.md`
- `docs/architecture/initiatives/s10_mcp_user_interactions/architecture.md`
- `docs/architecture/initiatives/s10_mcp_user_interactions/c4_context.md`
- `docs/architecture/initiatives/s10_mcp_user_interactions/c4_container.md`

## Связанные source-of-truth документы
- `docs/architecture/api_contract.md`
- `docs/architecture/data_model.md`
- `docs/architecture/mcp_approval_and_audit_flow.md`
- `docs/architecture/adr/ADR-0012-built-in-mcp-user-interactions-control-plane-owned-lifecycle.md`
- `docs/architecture/alternatives/ALT-0004-built-in-mcp-user-interactions-lifecycle-boundaries.md`
- `docs/delivery/epics/s10/epic-s10-day4-mcp-user-interactions-arch.md`
- `docs/delivery/epics/s10/prd-s10-day3-mcp-user-interactions.md`
