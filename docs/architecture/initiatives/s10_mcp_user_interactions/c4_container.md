---
doc_id: ARC-C4N-S10-0001
type: c4-container
title: "Sprint S10 Day 4 — C4 Container overlay for built-in MCP user interactions"
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

# C4 Container: Sprint S10 Day 4 built-in MCP user interactions

## TL;DR
- Container baseline не меняется: built-in MCP user interactions реализуются внутри существующих `agent-runner`, `api-gateway`, `control-plane`, `worker`, `postgres`.
- Новая Day4-фиксация касается только ownership split для interaction state, callback ingress, retries/expiry и wait-state resume.

## Диаграмма (Mermaid C4Container)
```mermaid
C4Container
title Sprint S10 Day4 - Built-in MCP user interactions container overlay

Person(user, "User", "Получает notify и отвечает на decision request")
Person(agent, "Agent pod", "Вызывает built-in interaction tools")
System_Ext(adapters, "Interaction adapters", "Telegram/Slack/Web/HTTP adapters")
System_Ext(github, "GitHub", "Issue/PR context and deep-links")

System_Boundary(b0, "codex-k8s") {
  Container(runner, "Agent Runner / agent pod", "Codex CLI job", "Calls built-in tools and resumes run after typed response")
  Container(gw, "API Gateway", "Go HTTP", "Thin-edge callback ingress, adapter auth, payload normalization")
  Container(cp, "Control Plane", "Go", "Owns interaction aggregate, wait-state transitions, validation, audit and tool semantics")
  Container(worker, "Worker", "Go", "Dispatch, retries, timeout/expiry reconciliation and delivery attempts")
  ContainerDb(db, "PostgreSQL", "Postgres + JSONB + pgvector", "Interaction records, delivery attempts, run/session waits, flow events")
}

Rel(agent, runner, "Executes work inside runtime", "Kubernetes")
Rel(runner, cp, "Calls user.notify / user.decision.request", "MCP StreamableHTTP")
Rel(cp, db, "Reads/Writes interaction state, audit, run/session wait markers")
Rel(worker, db, "Reads/Writes dispatch queue, attempt records and expiry state")
Rel(worker, adapters, "Sends delivery payloads and retries", "HTTPS")
Rel(user, adapters, "Reads messages and sends typed response", "Adapter UX")
Rel(adapters, gw, "Sends callback / delivery ack", "HTTPS")
Rel(gw, cp, "Normalized callback transport", "gRPC")
Rel(cp, github, "Reads issue/PR context and deep-links", "HTTPS")
```

## Container responsibilities in built-in MCP user interactions

| Container | Role |
|---|---|
| `agent-runner` | Использует только built-in MCP tools; не владеет callback lifecycle и persisted interaction state |
| `api-gateway` | Callback ingress, adapter auth, typed transport normalization |
| `control-plane` | Interaction aggregate owner, state transitions, validation, wait-state pause/resume, audit/correlation |
| `worker` | Outbound delivery, retries, expiry scan, attempt-level reconciliation |
| `postgres` | Единственная persisted coordination layer между pod для interaction lifecycle |

## Runtime и data boundaries
- `agent-runner` не хранит source-of-truth interaction state внутри pod.
- `api-gateway` не принимает решений о accepted/rejected response outcome и idempotency.
- `worker` не решает business semantics `response_kind` и не завершает run напрямую без `control-plane`.
- `postgres` остаётся единственной точкой синхронизации; отдельный broker/service для interaction lifecycle на Day4 не вводится.

## Handover note for `run:design`
- Уточнить, какой callback contract reuse текущий MCP surface, а какой вводит новый typed family внутри `api-gateway`.
- Зафиксировать точную persisted model без нарушения schema ownership `control-plane`.
