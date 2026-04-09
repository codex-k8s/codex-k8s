---
doc_id: ARC-C4N-S13-0001
type: c4-container
title: "Sprint S13 Day 4 — C4 Container overlay for Quality Governance System"
status: in-review
owner_role: SA
created_at: 2026-03-15
updated_at: 2026-03-15
related_issues: [466, 469, 470, 471, 476, 484, 488, 494]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-15-issue-484-arch"
---

# C4 Container: Sprint S13 Day 4 Quality Governance System

## TL;DR
- Container baseline платформы не меняется: capability реализуется внутри существующих `agent-runner`, `control-plane`, `worker`, `api-gateway`, `web-console`, `postgres`.
- Day4 фиксирует ownership split для draft/evidence handoff, canonical governance aggregate, asynchronous reconciliation и typed visibility/decision projections.

## Диаграмма (Mermaid C4Container)
```mermaid
C4Container
title Sprint S13 Day4 - Quality Governance System container overlay

Person(owner, "Owner / reviewer", "Ждёт прозрачный governance state и следующий шаг")
Person(operator, "Platform operator", "Диагностирует gaps, readiness и outcomes policy-aware re-evaluation")
System_Ext(github, "GitHub", "Issues, PRs, reviews, labels and publication signals")

System_Boundary(b0, "kodex") {
  Container(runner, "Agent Runner / agent pod", "Codex CLI job", "Emits draft/evidence signals, semantic wave hints and verification outputs; never owns governance policy")
  Container(cp, "Control Plane", "Go", "Owns canonical governance aggregate, publication gate, risk/evidence/verification/waiver decisions and typed projections")
  Container(worker, "Worker", "Go", "Runs asynchronous sweeps, feedback ingestion, stale-gate escalation and submits reconciliation findings for policy-aware re-evaluation")
  Container(gw, "API Gateway", "Go HTTP", "Thin-edge ingress for GitHub/staff requests and future typed decision commands")
  Container(web, "Web Console", "Vue 3", "Displays typed governance projections and future structured decision surfaces")
  ContainerDb(db, "PostgreSQL", "Postgres + JSONB + pgvector", "Persisted governance aggregate, wave lineage, audit evidence and reconciliation state")
}

Rel(owner, web, "Reads governance projections and next steps", "HTTPS")
Rel(operator, web, "Uses diagnostics and readiness views", "HTTPS")
Rel(runner, cp, "Reports draft/evidence/verification signals", "Internal callbacks")
Rel(github, gw, "Sends review, label and publication webhooks", "HTTPS")
Rel(gw, cp, "Routes typed ingress without policy logic", "gRPC")
Rel(cp, db, "Reads/Writes governance aggregate, projections and audit state")
Rel(worker, db, "Reads/Writes reconciliation queues, sweep snapshots and feedback evidence")
Rel(worker, cp, "Submits reconciliation findings and requests policy-aware re-evaluation / escalation", "gRPC")
Rel(web, gw, "Reads staff/private projections and future command APIs", "HTTPS")
Rel(cp, github, "Updates linked service messages / label-aware status", "HTTPS")
```

## Container responsibilities in Quality Governance System

| Container | Role |
|---|---|
| `agent-runner` | Видит локальный run context и передаёт draft/evidence/verification signals без ownership policy semantics |
| `control-plane` | Единственный owner canonical governance aggregate, publication gate, waiver/residual-risk decisions и typed visibility contract |
| `worker` | Выполняет sweeps, stale detection и postdeploy feedback rollups, пишет только reconciliation/evidence state и передаёт findings в `control-plane` для late reclassification / gap closure |
| `api-gateway` | Отдаёт thin-edge ingress/transport surface для GitHub webhook и staff/private команд |
| `web-console` | Показывает typed projections и operator-facing next-step guidance без локальной бизнес-логики |
| `postgres` | Единая persisted coordination layer между pod для governance state, wave lineage и audit evidence |

## Runtime и data boundaries
- `agent-runner` не хранит source-of-truth governance state внутри pod.
- `api-gateway` и `web-console` не вычисляют risk tier, evidence completeness, waiver rules или publication admissibility самостоятельно.
- `worker` не закрывает gates и не создаёт policy semantics без решения `control-plane`.
- GitHub labels/comments остаются внешними publication/review surfaces и не заменяют canonical aggregate в PostgreSQL.

## Continuity after `run:arch`
- Design package в Issue `#494` должен описать typed handoff/projection/decision contracts и migration policy, не меняя этот container ownership split.
- Любой downstream runtime/UI stream Sprint S14 (`#470`) обязан потреблять готовые typed surfaces из `control-plane`, а не переносить ownership в `web-console` или отдельный temporary service.
