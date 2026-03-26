---
doc_id: ARC-C4N-S16-0001
type: c4-container
title: "Sprint S16 Day 4 — C4 Container overlay for Mission Control graph workspace"
status: superseded
owner_role: SA
created_at: 2026-03-16
updated_at: 2026-03-25
related_issues: [480, 490, 492, 496, 510, 516, 519, 561, 562, 563]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-16-issue-516-arch"
---

# C4 Container: Sprint S16 Day 4 Mission Control graph workspace

## TL;DR
- 2026-03-25 issue `#561` перевела этот C4 container overlay в historical superseded state.
- Контейнерные границы и ownership из этого файла больше не являются текущим Mission Control source of truth; они сохранены только как evidence отклонённого S16 baseline.

## Диаграмма (Mermaid C4Container)
```mermaid
C4Container
title Sprint S16 Day4 - Mission Control graph workspace container overlay

Person(owner, "Owner / operator", "Использует graph workspace для continuity and next-step decisions")
System_Ext(github, "GitHub", "Issues, PRs, comments, reviews, labels")
System_Ext(k8s, "Kubernetes", "Agent and background runtime")

System_Boundary(b0, "codex-k8s") {
  Container(runner, "Agent Runner / agent pod", "Codex CLI job", "Emits run lineage, produced artifacts and launch params; never owns graph truth")
  Container(cp, "Control Plane", "Go", "Owns graph truth, node classification, continuity state, metadata/watermarks and next-step policy")
  Container(worker, "Worker", "Go", "Runs bounded inventory sync, recent-closed-history backfill, enrichment/reconcile jobs and lifecycle tasks")
  Container(gw, "API Gateway", "Go HTTP", "Thin-edge GitHub/staff ingress and typed command routing")
  Container(web, "Web Console", "Vue 3", "Graph canvas, toolbar, drawer, list fallback and visibility surfaces")
  ContainerDb(db, "PostgreSQL", "Postgres + JSONB + pgvector", "Persisted graph truth, provider mirror, watermarks, continuity gaps and audit state")
}

Rel(owner, web, "Uses", "HTTPS")
Rel(web, gw, "Reads graph projections and next-step surfaces", "HTTPS")
Rel(github, gw, "Sends webhook events", "HTTPS")
Rel(gw, cp, "Routes typed ingress without graph logic", "gRPC")
Rel(cp, db, "Reads/Writes graph truth, continuity state and launch surfaces")
Rel(worker, db, "Reads/Writes provider mirror state, freshness cursors and background-task evidence")
Rel(worker, cp, "Requests policy-aware graph re-evaluation / continuity refresh", "gRPC")
Rel(runner, cp, "Reports run outputs, lineage hints and launch params", "Internal callbacks")
Rel(worker, github, "Executes provider sync and bounded backfill", "HTTPS")
Rel(cp, k8s, "Reads runtime lineage context", "Kubernetes API")
```

## Container responsibilities in Mission Control graph workspace

| Container | Role |
|---|---|
| `agent-runner` | Передаёт run context, produced artifacts и launch params без ownership graph semantics |
| `control-plane` | Единственный owner graph truth, continuity gaps, typed metadata/watermarks и next-step eligibility |
| `worker` | Выполняет bounded inventory sync, recent-closed-history backfill, enrichment/reconcile jobs и lifecycle tasks без ownership graph truth |
| `api-gateway` | Thin-edge ingress для GitHub webhook и staff/private actions |
| `web-console` | Показывает canvas, drawer и list fallback на основе typed projections |
| `postgres` | Единая persisted coordination layer для graph truth, provider mirror, watermarks и audit state |

## Runtime и data boundaries
- `web-console` не строит canonical graph и не рассчитывает allowed next steps локально.
- `api-gateway` не принимает решений о node classification, continuity completeness или hybrid truth merge.
- `worker` не изменяет canonical node kinds и не закрывает continuity gaps без решения `control-plane`.
- `agent-runner` не становится owner relations только потому, что первым видел run outcome.

## Continuity after `run:arch`
- Design package в Issue `#519` должен описать typed snapshot/detail/launch contracts и migration policy, не меняя этот container ownership split.
- Любой downstream execution stream Sprint S16 обязан потреблять готовые typed surfaces из `control-plane`, а не переносить graph truth в UI или отдельный temporary service.
