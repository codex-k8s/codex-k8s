---
doc_id: ARC-C4C-S13-0001
type: c4-context
title: "Sprint S13 Day 4 — C4 Context overlay for Quality Governance System"
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

# C4 Context: Sprint S13 Day 4 Quality Governance System

## TL;DR
- `Quality Governance System` остаётся capability slice внутри `codex-k8s`, а не отдельной внешней governance-платформой.
- Owner/reviewer, delivery roles и platform operator получают разные visibility surfaces, но единый source-of-truth для policy semantics живёт внутри platform domain.

## Диаграмма (Mermaid C4Context)
```mermaid
C4Context
title Sprint S13 Day4 - Quality Governance System context overlay

Person(owner, "Owner / reviewer", "Принимает go/no-go, waiver и residual-risk decisions")
Person(delivery, "Delivery role / agent", "Готовит working draft, semantic waves и evidence")
Person(operator, "Platform operator", "Диагностирует governance gaps и release readiness")

System(system, "codex-k8s quality governance slice", "Canonical change-governance capability для agent-scale delivery")

System_Ext(github, "GitHub", "Issues, PRs, reviews, labels, webhooks и review/publication surfaces")
System_Ext(k8s, "Kubernetes", "Runtime substrate для agent-runner и worker")
System_Ext(staff, "Staff UI", "Typed visibility surface для governance state и operator actions")

Rel(delivery, system, "Submits draft/evidence signals and semantic wave intent", "Internal callbacks / gRPC")
Rel(owner, system, "Reads decision surface and records review/waiver outcomes", "GitHub service-comment + staff UI")
Rel(operator, staff, "Uses governance dashboards, readiness views and gap diagnostics", "HTTPS")
Rel(system, github, "Consumes review/publication events and updates linked messages", "HTTPS / webhooks")
Rel(system, k8s, "Runs agent and background reconciliation workloads", "Kubernetes API")
Rel(system, staff, "Publishes typed projections and future structured commands", "Staff/private API")
```

## Пояснения
- GitHub остаётся внешним source of review/publication events, но не source-of-truth для canonical governance semantics.
- Staff UI и GitHub service-comments остаются surfaces одного и того же typed projection.
- Kubernetes обеспечивает runtime только для agent/worker execution и не владеет risk/evidence/waiver semantics.

## Внешние зависимости
- GitHub: review, labels, PR surfaces и webhook evidence.
- Kubernetes: runtime для `agent-runner` и `worker`.
- Staff UI/API: operator/owner visibility surface, но не место для доменных решений.

## Continuity after `run:arch`
- Issue `#494` (`run:design`) должен сохранить этот context overlay как baseline для typed transport/data contracts.
- Sprint S14 (`#470`) не может превращать GitHub, Staff UI или Kubernetes в source-of-truth для policy semantics: этот инвариант остаётся внутри platform domain.
