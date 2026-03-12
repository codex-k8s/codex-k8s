---
doc_id: EPC-CK8S-S10-D5-MCP-INTERACTIONS
type: epic
title: "Epic S10 Day 5: Design для built-in MCP user interactions (Issue #387)"
status: in-review
owner_role: SA
created_at: 2026-03-12
updated_at: 2026-03-12
related_issues: [360, 378, 383, 385, 387, 389]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-12-issue-387-design-epic"
---

# Epic S10 Day 5: Design для built-in MCP user interactions (Issue #387)

## TL;DR
- Подготовлен design package Sprint S10 для перехода в `run:plan`:
  - `design_doc`,
  - `api_contract`,
  - `data_model`,
  - `migrations_policy`.
- Зафиксированы typed decisions для built-in tools, adapter callbacks, interaction aggregate, wait-state taxonomy и resume contract.
- Подготовлен handover в `run:plan` через follow-up issue `#389` без trigger-лейбла.

## Контекст
- Stage continuity: `#360 -> #378 -> #383 -> #385 -> #387 -> #389`.
- Входной baseline: architecture package Day4 (`ADR-0012`, `ALT-0004`, ownership matrix Sprint S10).
- Scope Day5: только markdown-изменения, без runtime/code/migration execution.

## Артефакты Day5
- `docs/architecture/initiatives/s10_mcp_user_interactions/design_doc.md`
- `docs/architecture/initiatives/s10_mcp_user_interactions/api_contract.md`
- `docs/architecture/initiatives/s10_mcp_user_interactions/data_model.md`
- `docs/architecture/initiatives/s10_mcp_user_interactions/migrations_policy.md`

## Ключевые design-решения
- `user.notify` остаётся non-blocking contract и не создаёт wait-state.
- `user.decision.request` использует existing runtime status `waiting_mcp`, но business meaning выражается через `wait_reason=interaction_response`.
- Interaction-domain получает отдельные persisted сущности: aggregate, delivery attempts, callback evidence и typed response records.
- `agent_sessions` переиспользуется как resume snapshot storage; interaction result попадает в deterministic resume payload, а не в approval tables.
- `api-gateway` получает отдельный callback family `/api/v1/mcp/interactions/callback`; approval callbacks остаются отдельным контуром.

## Context7 verification
- Выполнена попытка использовать Context7 для `kin-openapi` и `goose`.
- Результат: `Monthly quota exceeded`.
- Новые внешние зависимости в Day5 не требуются.

## Runtime и migration impact
- Runtime impact на этапе Day5: отсутствует.
- Migration impact для `run:dev`:
  - новые interaction tables;
  - additive wait-linkage в `agent_runs`;
  - backfill `wait_reason: mcp -> approval_pending`.
- Rollout order остаётся обязательным:
  `migrations -> control-plane -> worker -> api-gateway -> adapters`.

## Acceptance criteria status (Issue #387)
- [x] Подготовлен design package (`design_doc`, `api_contract`, `data_model`, `migrations_policy`).
- [x] Зафиксированы typed tool/callback contracts, interaction lifecycle, retries/idempotency и resume semantics.
- [x] Зафиксированы data-model и migration/rollback boundaries без reuse approval flow как source-of-truth.
- [x] Синхронизированы sprint/epic/traceability документы.
- [x] Подготовлена следующая stage-issue `#389` для `run:plan` без trigger-лейбла.

## Handover в `run:plan`
- Следующий этап: `run:plan`.
- Follow-up issue: `#389` (без trigger-лейбла; trigger ставит Owner).
- Обязательная декомпозиция в `run:plan`:
  - execution waves по `control-plane`, `worker`, `api-gateway`, quality/observability;
  - sequencing и DoR/DoD для rollout `migrations -> control-plane -> worker -> api-gateway`;
  - test and replay-safety gates перед `run:dev`.
