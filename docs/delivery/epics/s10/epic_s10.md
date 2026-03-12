---
doc_id: EPC-CK8S-0010
type: epic
title: "Epic Catalog: Sprint S10 (Built-in MCP user interactions and channel-neutral adapter contracts)"
status: in-review
owner_role: PM
created_at: 2026-03-12
updated_at: 2026-03-12
related_issues: [360, 378]
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
- Day2 vision (`#378`) должен закрепить mission, persona outcomes, KPI и handover в PRD issue.
- До `run:plan` Sprint S10 остаётся markdown-only контуром: implementation issues, кодовые правки и выбор конкретных library/runtime деталей откладываются до подтверждённых product и architecture decisions.

## Stage roadmap
- Day 1 (Intake): `docs/delivery/epics/s10/epic-s10-day1-mcp-user-interactions-intake.md` (Issue `#360`).
- Day 2 (Vision): continuity issue `#378`; должна сформировать mission, KPI, guardrails и persona outcomes.
- Day 3 (PRD): создаётся после Owner review vision stage; должна зафиксировать user stories, FR/AC/NFR, edge cases и expected evidence.
- Day 4 (Architecture): создаётся после PRD review; должна закрепить boundaries между `control-plane`, `worker`, transport edges и future channel adapters.
- Day 5 (Design): создаётся после architecture review; должна оформить implementation-ready API/data/wait-state contracts.
- Day 6 (Plan): создаётся после design review; должна сформировать execution waves, quality-gates и отдельные implementation issues.

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
- `S10-E05` не считается blocking scope для core Sprint S10: Telegram и другие adapters стартуют только после фиксации channel-neutral platform contract.
- Approval flow redesign не входит в Sprint S10 и не может использоваться как shortcut для user interaction scope.
