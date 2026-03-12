---
doc_id: EPC-CK8S-0010
type: epic
title: "Epic Catalog: Sprint S10 (Built-in MCP user interactions and channel-neutral adapter contracts)"
status: in-review
owner_role: PM
created_at: 2026-03-12
updated_at: 2026-03-12
related_issues: [360, 378, 383, 385, 387, 389]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-12-issue-360-intake"
---

# Epic Catalog: Sprint S10 (Built-in MCP user interactions and channel-neutral adapter contracts)

## TL;DR
- Sprint S10 открывает отдельную продуктовую инициативу вокруг built-in MCP user interactions: платформа должна уметь безопасно уведомлять пользователя и запрашивать его решение без смешения с approval flow.
- Day1 intake (`#360`) фиксирует problem statement, MVP scope, guardrails и continuity в `run:vision`.
- Day2 vision (`#378`) фиксирует mission, persona outcomes, KPI/guardrails и handover в PRD issue `#383`.
- Day3 PRD (`#383`) фиксирует user stories, FR/AC/NFR, wait-state/correlation guardrails и handover в architecture issue `#385`.
- Day4 architecture (`#385`) фиксирует ownership split, interaction lifecycle и design continuity issue `#387`.
- Day5 design (`#387`) фиксирует implementation-ready contracts/data/migrations package и handover в plan issue `#389`.
- До `run:plan` Sprint S10 остаётся markdown-only контуром: implementation issues, кодовые правки и выбор конкретных library/runtime деталей откладываются до подтверждённых product и architecture decisions.

## Stage roadmap
- Day 1 (Intake): `docs/delivery/epics/s10/epic-s10-day1-mcp-user-interactions-intake.md` (Issue `#360`).
- Day 2 (Vision): `docs/delivery/epics/s10/epic-s10-day2-mcp-user-interactions-vision.md` (Issue `#378`).
- Day 3 (PRD): `docs/delivery/epics/s10/epic-s10-day3-mcp-user-interactions-prd.md` + `docs/delivery/epics/s10/prd-s10-day3-mcp-user-interactions.md` (Issue `#383`); зафиксированы user stories, FR/AC/NFR, edge cases и expected evidence.
- Day 4 (Architecture): `docs/delivery/epics/s10/epic-s10-day4-mcp-user-interactions-arch.md` + architecture package in `docs/architecture/initiatives/s10_mcp_user_interactions/` (Issue `#385`).
- Day 5 (Design): `docs/delivery/epics/s10/epic-s10-day5-mcp-user-interactions-design.md` + design package in `docs/architecture/initiatives/s10_mcp_user_interactions/` (Issue `#387`).
- Day 6 (Plan): continuity issue `#389`; должна сформировать execution waves, quality-gates и отдельные implementation issues.

## Candidate product streams

| Epic ID | Scope | Почему это отдельный поток |
|---|---|---|
| `S10-E01` | Built-in tools baseline: `user.notify` и `user.decision.request` | Это минимальный пользовательский contract, который должен быть понятен агенту, платформе и адаптерам |
| `S10-E02` | Interaction-domain model: request, response, correlation, response_kind, selected_option, free_text, wait-state semantics | Без typed interaction model инициатива скатится в ad-hoc payloads и ручную интерпретацию callback data |
| `S10-E03` | Channel-neutral outbound/inbound contracts для interaction adapters | Нужно сразу отделить platform core от Telegram и других каналов, чтобы не зашить vendor-specific semantics в основу |
| `S10-E04` | Runtime/audit/retry/idempotency lifecycle для user interactions | Wait-state, retries и callbacks должны быть безопасны и воспроизводимы на уровне платформы, а не внутри agent pod |
| `S10-E05` | Telegram adapter и другие channel-specific follow-up streams | Высокая ценность, но отдельная последовательная инициатива после стабилизации core interaction contract |

## Delivery-governance правила
- До `run:plan` Sprint S10 не создаёт implementation issues и не добавляет новые зависимости в репозиторий.
- Каждый stage создаёт следующую issue без trigger-лейбла; запуск следующего stage остаётся Owner-managed.
- После PRD stage continuity issue `#385` становится единственной точкой входа в `run:arch` для этой инициативы.
- После architecture stage continuity issue `#387` становится единственной точкой входа в `run:design` для этой инициативы.
- После design stage continuity issue `#389` становится единственной точкой входа в `run:plan` для этой инициативы.
- `S10-E05` не считается blocking scope для core Sprint S10: Telegram и другие adapters стартуют только после фиксации channel-neutral platform contract.
- Approval flow redesign не входит в Sprint S10 и не может использоваться как shortcut для user interaction scope.
