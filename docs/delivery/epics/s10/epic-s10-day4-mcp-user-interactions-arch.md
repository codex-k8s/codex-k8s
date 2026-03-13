---
doc_id: EPC-CK8S-S10-D4-MCP-INTERACTIONS
type: epic
title: "Epic S10 Day 4: Architecture для built-in MCP user interactions (Issue #385)"
status: in-review
owner_role: SA
created_at: 2026-03-12
updated_at: 2026-03-13
related_issues: [360, 378, 383, 385, 387]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-03-12-issue-385-arch"
  approved_by: "ai-da-stas"
  approved_at: 2026-03-13
---

# Epic S10 Day 4: Architecture для built-in MCP user interactions (Issue #385)

## TL;DR
- Подготовлен architecture package Sprint S10 для built-in MCP user interactions: architecture decomposition, C4 overlays, ADR-0012 и alternatives по lifecycle ownership и adapter boundaries.
- Зафиксирован ownership split для built-in tool invocation, interaction aggregate, wait-state transitions, outbound dispatch/retries, callback ingestion и approval-flow separation.
- Подготовлен handover в `run:design` без premature transport/schema lock-in.

## Priority
- `P0`.

## Контекст
- Intake baseline: `#360` (`docs/delivery/epics/s10/epic-s10-day1-mcp-user-interactions-intake.md`).
- Vision baseline: `#378` (`docs/delivery/epics/s10/epic-s10-day2-mcp-user-interactions-vision.md`).
- PRD baseline: `#383` (`docs/delivery/epics/s10/epic-s10-day3-mcp-user-interactions-prd.md`, `docs/delivery/epics/s10/prd-s10-day3-mcp-user-interactions.md`).
- Текущий этап: `run:arch` в Issue `#385`.
- Scope этапа: только markdown-изменения.

## Architecture package
- `docs/architecture/initiatives/s10_mcp_user_interactions/README.md`
- `docs/architecture/initiatives/s10_mcp_user_interactions/architecture.md`
- `docs/architecture/initiatives/s10_mcp_user_interactions/c4_context.md`
- `docs/architecture/initiatives/s10_mcp_user_interactions/c4_container.md`
- `docs/architecture/adr/ADR-0012-built-in-mcp-user-interactions-control-plane-owned-lifecycle.md`
- `docs/architecture/alternatives/ALT-0004-built-in-mcp-user-interactions-lifecycle-boundaries.md`
- `docs/delivery/traceability/s10_mcp_user_interactions_history.md`

## Ключевые решения Stage
- `control-plane` остаётся владельцем interaction aggregate, typed validation, audit/correlation и wait-state transitions.
- `worker` закреплён за outbound dispatch, retries и timeout/expiry reconciliation; `api-gateway` не принимает решений о accepted/rejected business outcome.
- Approval flow и user interaction flow разделены как разные bounded contexts; approval vocabulary не используется как primary model для user responses.
- Built-in `codex_k8s` остаётся единственной core точкой расширения; adapters остаются replaceable transport layers без Telegram-first lock-in.

## Context7 верификация
- Выполнена попытка использовать Context7 для Mermaid/C4 documentation.
- Результат: `Monthly quota exceeded`.
- Для пакета использованы существующие Mermaid/C4 conventions репозитория; новых внешних зависимостей не требуется.

## Acceptance Criteria (Issue #385)
- [x] Подготовлен architecture package с service boundaries, ownership, C4 overlays, ADR и alternatives для built-in MCP user interactions.
- [x] Для core flows определены owner-сервисы и границы ответственности: built-in tool invocation, interaction aggregate, wait-state transitions, callback ingestion, retries/expiry и adapter isolation.
- [x] Зафиксированы architecture-level trade-offs по lifecycle ownership и approval separation без premature transport/storage lock-in.
- [x] Обновлены delivery/traceability документы и package indexes.
- [x] Подготовлена follow-up issue `#387` для stage `run:design` без trigger-лейбла.

## Quality gates
| Gate | Что проверяем | Статус |
|---|---|---|
| `QG-S10-D4-01` Architecture completeness | Есть package `architecture + C4 + ADR + alternatives` | passed |
| `QG-S10-D4-02` Boundary integrity | Thin-edge сохранён, ownership за `control-plane`/`worker`/`api-gateway` зафиксирован явно | passed |
| `QG-S10-D4-03` Approval separation | Interaction-domain отделён от approval/control domain на уровне semantics и source-of-truth | passed |
| `QG-S10-D4-04` Adapter neutrality | Core contract не привязан к Telegram-first UX и не требует vendor-specific semantics | passed |
| `QG-S10-D4-05` Stage continuity | Подготовлена issue `#387` на `run:design` без trigger-лейбла | passed |

## Handover в `run:design`
- Следующий этап: `run:design`.
- Follow-up issue: `#387`.
- Trigger-лейбл на новую issue ставит Owner после review architecture package.
- На design-stage обязательно:
  - выпустить `design_doc`, `api_contract`, `data_model`, `migrations_policy`;
  - определить typed tool/callback contracts, interaction status model и wait-state/resume semantics;
  - зафиксировать persistence strategy для interaction records и rollout/rollback notes.
