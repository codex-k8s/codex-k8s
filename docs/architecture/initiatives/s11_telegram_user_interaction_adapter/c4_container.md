---
doc_id: ARC-C4N-S11-0001
type: c4-container
title: "Sprint S11 Day 4 — C4 Container overlay for Telegram user interaction adapter"
status: approved
owner_role: SA
created_at: 2026-03-14
updated_at: 2026-03-14
related_issues: [361, 444, 447, 448, 452, 454, 456, 458]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-03-14-issue-452-arch"
  approved_by: "ai-da-stas"
  approved_at: 2026-03-14
---

# C4 Container: Sprint S11 Day 4 Telegram user interaction adapter

## TL;DR
- Container baseline платформы не меняется: Telegram channel path раскладывается на существующие `agent-runner`, `api-gateway`, `control-plane`, `worker`, `postgres` и внешний Telegram adapter contour.
- Day4 фиксирует только ownership split для outbound delivery, raw webhook/auth boundary, normalized callback ingress, semantic classification и post-callback UX continuation.

## Диаграмма (Mermaid C4Container)
```mermaid
C4Container
title Sprint S11 Day4 - Telegram user interaction adapter container overlay

Person(user, "Telegram user", "Получает notify и отвечает кнопкой или free text")
Person(agent, "Agent pod", "Вызывает built-in interaction tools")
System_Ext(tgadapter, "Telegram adapter contour", "Channel-specific rendering, raw webhook handling, Bot API mediation")
System_Ext(telegram, "Telegram Bot API", "Bot transport and webhook delivery")
System_Ext(github, "GitHub", "Issue/PR links and fallback context")

System_Boundary(b0, "kodex") {
  Container(runner, "Agent Runner / agent pod", "Codex CLI job", "Calls built-in interaction tools and resumes run")
  Container(gw, "API Gateway", "Go HTTP", "Validates adapter callback auth/schema and bridges to control-plane")
  Container(cp, "Control Plane", "Go", "Owns interaction aggregate, semantic classification, wait-state and operator visibility")
  Container(worker, "Worker", "Go", "Dispatch, retries, expiry and post-callback edit/follow-up jobs")
  ContainerDb(db, "PostgreSQL", "Postgres + JSONB + pgvector", "Interaction state, delivery attempts, callback evidence, flow events")
}

Rel(agent, runner, "Executes work inside runtime", "Kubernetes")
Rel(runner, cp, "Calls user.notify / user.decision.request", "MCP StreamableHTTP")
Rel(cp, db, "Reads/Writes interaction state, correlation and wait markers")
Rel(worker, db, "Reads/Writes dispatch queue, attempts and follow-up state")
Rel(worker, tgadapter, "Sends delivery and follow-up commands", "HTTPS")
Rel(tgadapter, telegram, "Uses Bot API and receives webhooks", "HTTPS")
Rel(user, tgadapter, "Reads message / clicks buttons / sends free text", "Telegram UX")
Rel(tgadapter, gw, "Sends normalized callbacks and delivery receipts", "HTTPS")
Rel(gw, cp, "Thin bridge for typed callback transport", "gRPC")
Rel(cp, github, "Reads issue/PR links and fallback context", "HTTPS")
```

## Container responsibilities in Telegram user interactions

| Container | Role |
|---|---|
| `agent-runner` | Использует только built-in interaction tools и deterministic resume path; не владеет chat ids, webhook payloads и callback lifecycle |
| `api-gateway` | Platform callback auth, schema validation, typed transport normalization и gRPC bridge в `control-plane` |
| `control-plane` | Interaction aggregate owner, semantic classification, wait-state transitions, audit/correlation, operator visibility |
| `worker` | Outbound dispatch, retries, expiry scans и post-callback edit/follow-up continuation |
| `postgres` | Единственная persisted coordination layer для interaction lifecycle и delivery evidence |
| Telegram adapter contour | Channel-specific rendering, raw Telegram webhook verification, callback query acknowledgement и provider message refs |

## Runtime и data boundaries
- Raw Telegram webhooks не терминируются внутри `api-gateway`; этот boundary остаётся во внешнем Telegram adapter contour.
- `control-plane` не вызывает Telegram Bot API напрямую и не хранит Bot API payloads как primary model.
- `worker` выполняет edit/follow-up after semantic decision, но не решает сам, какой callback logical winner.
- `agent-runner` не хранит source-of-truth interaction state в pod и не может напрямую обращаться к Telegram transport.

## Handover note for `run:design`
- Зафиксировать exact outbound/inbound DTO, callback handle model и persistence model для provider message refs и follow-up actions.
- Уточнить rollout/rollback order, если Telegram adapter contour выкатывается независимо от core platform services.
