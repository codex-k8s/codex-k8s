---
doc_id: ARC-C4N-S18-0001
type: c4-container
title: "Sprint S18 Day 4 — C4 Container overlay for frontend-first Mission Control canvas"
status: in-review
owner_role: SA
created_at: 2026-03-27
updated_at: 2026-03-27
related_issues: [480, 561, 562, 563, 565, 567, 571, 573]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-27-issue-571-arch"
---

# C4 Container: Sprint S18 Day 4 frontend-first Mission Control canvas

## TL;DR
- В Wave 1 активным owner prototype behavior является только `web-console`.
- `api-gateway`, `control-plane`, `worker` и `PostgreSQL` остаются существующими platform containers, но не становятся новой Mission Control truth-path для Sprint S18.
- `agent-runner` важен как источник repo-seed prompt policy, однако не участвует в runtime state prototype и не превращается в data source canvas.

## Диаграмма (Mermaid C4Container)
```mermaid
C4Container
title Sprint S18 Day4 - Frontend-first Mission Control canvas container overlay

Person(owner, "Owner / operator", "Использует canvas prototype для safe next-step decisions")
System_Ext(github, "GitHub", "Provider system; only safe deep links in Sprint S18")
System_Ext(k8s, "Kubernetes", "Runtime for platform and future execution flows")

System_Boundary(b0, "kodex") {
  Container(web, "Web Console", "Vue 3", "Owns fake-data scenario catalog, canvas projection, drawer/toolbar state and workflow preview UX")
  Container(gw, "API Gateway", "Go HTTP", "Existing auth/session/static delivery and future thin-edge transport seam")
  Container(cp, "Control Plane", "Go", "Existing platform policy source of truth; future owner for persisted Mission Control semantics after #563")
  Container(worker, "Worker", "Go", "Future provider mirror/reconcile owner in #563; not in current prototype runtime path")
  Container(runner, "Agent Runner / repo seeds", "Go job + embedded markdown seeds", "Canonical prompt/policy wording source; not a canvas state owner")
  ContainerDb(db, "PostgreSQL", "PostgreSQL", "Current platform state; new Mission Control persisted truth deferred to #563")
}

Rel(owner, web, "Uses", "HTTPS")
Rel(web, gw, "Uses existing auth/session shell only", "HTTPS")
Rel(web, github, "Opens safe provider links only", "HTTPS")
Rel(gw, cp, "Existing staff/private flows only; no new Mission Control domain path in Sprint S18", "gRPC")
Rel(cp, db, "Reads/Writes existing platform state", "SQL")
Rel(worker, db, "Future mirror/reconcile path after #563", "SQL")
Rel(worker, github, "Future webhook/reconcile sync after #563", "HTTPS")
Rel(runner, cp, "Shares repo-seed prompt policy baseline", "Docs/runtime")
Rel(cp, k8s, "Existing platform orchestration", "Kubernetes API")
```

## Container responsibilities in Sprint S18

| Container | Роль |
|---|---|
| `web-console` | Единственный owner fake-data scenario model, canvas projection, local UI state и workflow preview UX |
| `api-gateway` | Thin-edge adapter без новой Mission Control бизнес-логики |
| `control-plane` | Existing platform policy source и deferred owner для будущего backend rebuild |
| `worker` | Deferred provider mirror/reconcile executor только после старта `#563` |
| `agent-runner` / repo seeds | Source of truth для prompt wording и workflow-policy text, но не для UI state |
| `postgres` | Existing platform storage; Sprint S18 не вводит в него новые Mission Control prototype structures |

## Runtime и data boundaries
- `web-console` не вычисляет live provider truth и не трактует fake data как canonical persisted model.
- `api-gateway` не принимает решений о relation semantics, workflow policy или safe action matrix для prototype.
- `control-plane` не должен становиться скрытой обязательной зависимостью для Sprint S18 `run:dev`.
- `worker` не участвует в fake-data prototype и не вычисляет stale/freshness раньше backend rebuild.
- `agent-runner` не владеет canvas state, даже если workflow preview использует wording из repo seeds.

## Continuity after `run:arch`
- Design package в issue `#573` должен описать explicit UI/state contracts и documented replacement seam к backend rebuild `#563`, не меняя этот ownership split.
- Любой downstream execution stream Sprint S18 обязан потреблять approved boundaries из этого container overlay, а не возвращать S16-style hidden backend dependency.
