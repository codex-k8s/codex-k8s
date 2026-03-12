---
doc_id: ARC-C4N-S9-0001
type: c4-container
title: "Sprint S9 Day 4 — C4 Container overlay for Mission Control Dashboard"
status: in-review
owner_role: SA
created_at: 2026-03-12
updated_at: 2026-03-12
related_issues: [333, 335, 337, 340, 351]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-12-issue-340-arch"
---

# C4 Container: Sprint S9 Day 4 Mission Control Dashboard

## TL;DR
- Container baseline не меняется: Mission Control Dashboard реализуется внутри существующих `web-console`, `api-gateway`, `control-plane`, `worker`, `postgres`.
- Новая Day4-фиксация касается только ownership-map для projections, commands, provider sync и realtime fallback.

## Диаграмма (Mermaid C4Container)
```mermaid
C4Container
title Sprint S9 Day4 - Mission Control Dashboard container overlay

Person(owner, "Owner / Operator", "Работает через Mission Control Dashboard")
System_Ext(github, "GitHub", "Issues, PR, comments, reviews")
System_Ext(k8s, "Kubernetes", "Agents and runtime state")
System_Ext(voice, "Optional voice intake provider", "Candidate stream only")

System_Boundary(b0, "codex-k8s") {
  Container(web, "Web Console", "Vue3", "Mission Control workspace, side panel, degraded/list fallback UX")
  Container(gw, "API Gateway", "Go HTTP + WS", "Thin-edge staff transport, auth, snapshot/delta delivery")
  Container(cp, "Control Plane", "Go", "Projection owner, relation graph, command lifecycle, policy and snapshot semantics")
  Container(worker, "Worker", "Go", "Outbound provider sync, retries, reconciliation and projection refresh")
  ContainerDb(db, "PostgreSQL", "Postgres + JSONB + pgvector", "Projection state, commands, relations, flow events, realtime event log")
}

Rel(owner, web, "Uses", "HTTPS")
Rel(web, gw, "Reads snapshots, sends commands, receives deltas", "HTTPS + WebSocket")
Rel(gw, cp, "Typed staff transport", "gRPC")
Rel(cp, db, "Reads/Writes projections, relations, command states")
Rel(worker, db, "Reads/Writes reconcile queue and projection refresh state")
Rel(worker, github, "Executes provider-safe mutations and sync", "HTTPS")
Rel(github, gw, "Sends webhook events", "HTTPS")
Rel(cp, k8s, "Reads runtime/agent context", "Kubernetes API")
Rel(voice, gw, "Optional callback/upload path", "HTTPS")
```

## Container responsibilities in Mission Control Dashboard

| Container | Role |
|---|---|
| `web-console` | Presentation-only workspace, filters, board/list toggle, stale/degraded indicators |
| `api-gateway` | Staff auth, typed HTTP/WS transport, webhook normalization boundary |
| `control-plane` | Active-set projection owner, relation graph, command admission, lifecycle and policy |
| `worker` | Provider sync/retries, webhook echo reconciliation helpers, projection refresh jobs |
| `postgres` | Единственный persisted coordination layer для projections, commands, relations и realtime event log |

## Runtime и data boundaries
- `web-console` не является каноническим владельцем relation graph или command state.
- `api-gateway` не принимает решений о dedupe, policy или active-set membership.
- `worker` не публикует пользовательские UX-решения: он исполняет и фиксирует reconciliation результат.
- `postgres` остаётся единственной точкой синхронизации между pod; отдельный broker/service для dashboard на Day4 не вводится.

## Handover note for `run:design`
- Уточнить, какие mission-control endpoints и realtime topics reuse current staff transport, а какие потребуют новый typed namespace.
- Зафиксировать точную projection persistence model без нарушения DB ownership `control-plane`.
