---
doc_id: EPC-CK8S-0010
type: epic
title: "Epic Catalog: Sprint S10 (Built-in MCP user interactions and channel-neutral adapter contracts)"
status: in-review
owner_role: PM
created_at: 2026-03-12
updated_at: 2026-03-13
related_issues: [360, 378, 383, 385, 387, 389, 391, 392, 393, 394, 395]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-03-12-issue-360-intake"
  approved_by: "ai-da-stas"
  approved_at: 2026-03-13
---

# Epic Catalog: Sprint S10 (Built-in MCP user interactions and channel-neutral adapter contracts)

## TL;DR
- Sprint S10 открывает отдельную продуктовую инициативу вокруг built-in MCP user interactions: платформа должна уметь безопасно уведомлять пользователя и запрашивать его решение без смешения с approval flow.
- Day1 intake (`#360`) фиксирует problem statement, MVP scope, guardrails и continuity в `run:vision`.
- Day2 vision (`#378`) фиксирует mission, persona outcomes, KPI/guardrails и handover в PRD issue `#383`.
- Day3 PRD (`#383`) фиксирует user stories, FR/AC/NFR, wait-state/correlation guardrails и handover в architecture issue `#385`.
- Day4 architecture (`#385`) фиксирует ownership split, interaction lifecycle и design continuity issue `#387`.
- Day5 design (`#387`) фиксирует implementation-ready contracts/data/migrations package и handover в plan issue `#389`.
- Day6 plan (`#389`) фиксирует execution package `#391..#395`, sequencing-waves, quality-gates и owner-managed handover в `run:dev`.
- Sprint S10 до `run:dev` остаётся markdown-only контуром: кодовые правки и library/runtime decisions начинаются только после owner review plan package.

## Stage roadmap
- Day 1 (Intake): `docs/delivery/epics/s10/epic-s10-day1-mcp-user-interactions-intake.md` (Issue `#360`).
- Day 2 (Vision): `docs/delivery/epics/s10/epic-s10-day2-mcp-user-interactions-vision.md` (Issue `#378`).
- Day 3 (PRD): `docs/delivery/epics/s10/epic-s10-day3-mcp-user-interactions-prd.md` + `docs/delivery/epics/s10/prd-s10-day3-mcp-user-interactions.md` (Issue `#383`); зафиксированы user stories, FR/AC/NFR, edge cases и expected evidence.
- Day 4 (Architecture): `docs/delivery/epics/s10/epic-s10-day4-mcp-user-interactions-arch.md` + architecture package in `docs/architecture/initiatives/s10_mcp_user_interactions/` (Issue `#385`).
- Day 5 (Design): `docs/delivery/epics/s10/epic-s10-day5-mcp-user-interactions-design.md` + design package in `docs/architecture/initiatives/s10_mcp_user_interactions/` (Issue `#387`).
- Day 6 (Plan): `docs/delivery/epics/s10/epic-s10-day6-mcp-user-interactions-plan.md` (Issue `#389`); execution package зафиксирован через issues `#391..#395`.

## Execution package (S10-E01..S10-E05)

| Stream | Implementation issue | Scope | Почему это отдельный поток |
|---|---|---|---|
| `S10-E01` | `#391` | `control-plane` foundation: additive schema, built-in tool orchestration, interaction aggregate, callback classification, wait linkage | Это source-of-truth поток для interaction domain; без него остальные streams теряют общий contract |
| `S10-E02` | `#392` | `worker` dispatch/retries/expiry/delivery-attempt ledger | Retry/expiry и enqueue resume должны быть изолированы от callback classification и edge-transport |
| `S10-E03` | `#393` | Contract-first `api-gateway` callback ingress и thin-edge bridge | Edge должен остаться thin-edge и не смешивать transport с platform domain |
| `S10-E04` | `#394` | Deterministic resume path в `agent-runner` | Resume handoff требует отдельной проверки, чтобы typed payload не растворился между `control-plane`, `worker` и runner |
| `S10-E05` | `#395` | Observability, quality gates, rollout/rollback readiness | Core interaction path нельзя передавать в `run:qa` без evidence по replay/idempotency/resume safety |

## Delivery-governance правила
- До `run:plan` Sprint S10 не создаёт implementation issues и не добавляет новые зависимости в репозиторий.
- Plan stage создал implementation issues `#391..#395` без trigger-лейблов; их запуск остаётся owner-managed.
- Каждый stage создаёт следующую issue без trigger-лейбла; запуск следующего stage остаётся Owner-managed.
- После PRD stage continuity issue `#385` становится единственной точкой входа в `run:arch` для этой инициативы.
- После architecture stage continuity issue `#387` становится единственной точкой входа в `run:design` для этой инициативы.
- После design stage continuity issue `#389` становится единственной точкой входа в `run:plan` для этой инициативы.
- Core `run:dev` sequencing фиксируется как `#391 -> #392 -> #393/#394 -> #395`; массовый параллельный старт execution issues запрещён.
- Channel-specific adapters, Telegram, reminders и voice/STT не считаются blocking scope для core Sprint S10 и стартуют отдельным follow-up контуром после стабилизации core interaction contract.
- Approval flow redesign не входит в Sprint S10 и не может использоваться как shortcut для user interaction scope.
