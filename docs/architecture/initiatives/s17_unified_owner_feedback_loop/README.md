---
doc_id: IDX-CK8S-ARCH-S17-0001
type: initiative-index
title: "Initiative Package: s17_unified_owner_feedback_loop"
status: in-review
owner_role: SA
created_at: 2026-03-26
updated_at: 2026-03-26
related_issues: [541, 554, 557, 559, 568]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-26-issue-559-arch"
---

# s17_unified_owner_feedback_loop

## TL;DR
- Пакет фиксирует Day4 architecture baseline Sprint S17 для unified owner feedback loop, same-session continuation и long-lived live wait contract.
- Внутри закреплены C4 overlays, service boundaries, ownership split, live-wait lifetime policy, recovery-only snapshot-resume boundary и channel-neutral truth для Telegram + staff-console.
- Follow-up issue `#568` открыта для `run:design`; следующий stage должен детализировать API/data/runtime contracts без reopening Day1-Day4 baseline.

## Содержимое
- `docs/architecture/initiatives/s17_unified_owner_feedback_loop/README.md`
- `docs/architecture/initiatives/s17_unified_owner_feedback_loop/architecture.md`
- `docs/architecture/initiatives/s17_unified_owner_feedback_loop/c4_context.md`
- `docs/architecture/initiatives/s17_unified_owner_feedback_loop/c4_container.md`

## Связанные source-of-truth документы
- `docs/architecture/api_contract.md`
- `docs/architecture/data_model.md`
- `docs/architecture/agent_runtime_rbac.md`
- `docs/architecture/mcp_approval_and_audit_flow.md`
- `docs/architecture/prompt_templates_policy.md`
- `docs/architecture/adr/ADR-0017-unified-owner-feedback-loop-live-wait-primary-platform-owned-continuation.md`
- `docs/architecture/alternatives/ALT-0009-unified-owner-feedback-loop-live-wait-and-channel-ownership.md`
- `docs/delivery/epics/s17/epic-s17-day4-unified-user-interaction-waits-and-owner-feedback-inbox-arch.md`
- `docs/delivery/epics/s17/prd-s17-day3-unified-user-interaction-waits-and-owner-feedback-inbox.md`
- `docs/delivery/sprints/s17/sprint_s17_unified_user_interaction_waits_and_owner_feedback_inbox.md`

## Continuity after `run:arch`
- Delivery-цепочка Sprint S17 остаётся последовательной: `#541 -> #554 -> #557 -> #559 -> #568 -> plan issue -> dev`.
- Owner-managed следующий этап: issue `#568` для `run:design` без trigger-лейбла.
- Design stage обязан сохранить:
  - same live pod / same `codex` session как primary happy-path;
  - max timeout/TTL built-in `codex_k8s` MCP wait path не ниже owner wait window;
  - snapshot-resume только как recovery fallback;
  - Telegram inbox и staff-console fallback поверх одного persisted backend truth;
  - deterministic text/voice/callback binding и visibility для overdue / expired / manual-fallback states.
