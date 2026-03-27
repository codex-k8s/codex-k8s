---
doc_id: ARC-C4N-S17-0001
type: c4-container
title: "Sprint S17 Day 4 — C4 Container overlay for unified owner feedback loop"
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

# C4 Container: Sprint S17 Day 4 unified owner feedback loop

## TL;DR
- Container baseline платформы не меняется: owner feedback continuity раскладывается на существующие `agent-runner`, `api-gateway`, `control-plane`, `worker`, `web-console`, `postgres` и внешний `telegram-interaction-adapter`.
- Day4 фиксирует только ownership split для live wait execution, persisted truth, callback ingress, dual-channel projections и degraded-state reconciliation.

## Диаграмма (Mermaid C4Container)
```mermaid
C4Container
title Sprint S17 Day4 - Unified owner feedback loop container overlay

Person(owner, "Owner / Product lead", "Получает pending request и отвечает через Telegram или staff-console")
Person(operator, "Staff / operator", "Наблюдает overdue / expired / manual-fallback states")
Person(agent, "Agent pod", "Выполняет задачу и ждёт owner reply")
System_Ext(tgadapter, "Telegram interaction adapter", "External Telegram delivery, webhook and reply normalization")
System_Ext(github, "GitHub", "Issue/PR context and service messages")

System_Boundary(b0, "codex-k8s") {
  Container(runner, "Agent Runner / agent pod", "Codex CLI job", "Keeps live same-session wait, heartbeats session and restores snapshot only for recovery")
  Container(gw, "API Gateway", "Go HTTP", "Thin-edge ingress for staff fallback actions and normalized adapter callbacks")
  Container(cp, "Control Plane", "Go", "Owns feedback request aggregate, lifecycle, deadlines, binding and continuation policy")
  Container(worker, "Worker", "Go", "Dispatch, retries, wait lease keepalive, overdue/expired/manual-fallback reconciliation")
  Container(web, "Staff Web Console", "Vue 3 + TypeScript", "Fallback inbox and operator visibility surface over typed staff API")
  ContainerDb(db, "PostgreSQL", "Postgres + JSONB + pgvector", "Canonical request truth, session wait markers, projections and audit")
}

Rel(agent, runner, "Executes task in live runtime", "Kubernetes")
Rel(runner, cp, "Calls built-in wait path and sends heartbeat/snapshot metadata", "MCP StreamableHTTP / internal callbacks")
Rel(cp, db, "Reads/Writes request truth, deadlines, continuation and audit")
Rel(worker, db, "Reads/Writes dispatch queue, leases, timers and degraded-state evidence")
Rel(worker, tgadapter, "Sends delivery and fallback notifications", "HTTPS")
Rel(owner, tgadapter, "Reads pending request / replies", "Telegram UX")
Rel(owner, web, "Uses fallback inbox and response actions", "HTTPS")
Rel(operator, web, "Reviews visibility and fallback state", "HTTPS")
Rel(web, gw, "Typed staff API", "HTTPS")
Rel(tgadapter, gw, "Normalized callbacks", "HTTPS")
Rel(gw, cp, "Thin bridge for typed actions and replies", "gRPC")
Rel(cp, github, "Reads context and writes service messages", "HTTPS")
```

## Container responsibilities in unified owner feedback loop

| Container | Role |
|---|---|
| `agent-runner` | Удерживает live same-session execution, heartbeat и snapshot capture; не владеет request truth и не может сам назначить detached resume как happy-path |
| `api-gateway` | Auth/RBAC/schema validation и typed bridge для staff fallback actions и normalized adapter callbacks |
| `control-plane` | Feedback request aggregate owner, wait/deadline policy, deterministic binding, accepted-response winner и continuation classification |
| `worker` | Delivery, retries, lease keepalive, overdue/expired/manual-fallback reconciliation и background notifications |
| `staff web-console` | Platform-owned fallback inbox и operator visibility surface без собственного lifecycle source of truth |
| `postgres` | Единственная persisted coordination layer для request truth, session waits, projections и audit evidence |
| `telegram-interaction-adapter` | Channel-specific delivery, raw webhook handling, voice/text normalization и provider refs без platform semantics |

## Runtime и data boundaries
- `agent-runner` не хранит source-of-truth request lifecycle внутри pod и не может завершить request без `control-plane`.
- `api-gateway` не выбирает semantic winner ответа и не materializes local wait-state policy.
- `staff web-console` не поддерживает separate inbox state и не обходит typed API/control-plane decisions.
- `telegram-interaction-adapter` не владеет overdue/expired/manual-fallback semantics и не определяет accepted-response outcome.
- `worker` не определяет business meaning reply types; он исполняет only background dispatch/reconcile side effects.

## Handover note for `run:design`
- Зафиксировать exact typed contracts для staff fallback actions, Telegram callbacks и built-in wait path.
- Определить persistence/read-model boundaries для projections, degraded-state visibility и recovery linkage без изменения этого ownership split.
