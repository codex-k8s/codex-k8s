---
doc_id: ARC-C4C-S12-0001
type: c4-context
title: "Sprint S12 Day 4 — C4 Context overlay for GitHub API rate-limit resilience"
status: approved
owner_role: SA
created_at: 2026-03-13
updated_at: 2026-03-13
related_issues: [366, 413, 416, 418, 420, 423, 425, 426, 427, 428, 429, 430, 431]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-03-13-issue-418-arch"
  approved_by: "ai-da-stas"
  approved_at: 2026-03-13
---

# C4 Context: Sprint S12 Day 4 GitHub API rate-limit resilience

## TL;DR
- GitHub API rate-limit resilience остаётся capability slice внутри `codex-k8s`, а не отдельной внешней quota-management системы.
- Owner/reviewer, platform operator и агент получают разные visibility semantics, но единый source-of-truth остаётся внутри platform domain.

## Диаграмма (Mermaid C4Context)
```mermaid
C4Context
title Sprint S12 Day4 - GitHub API rate-limit resilience context overlay

Person(owner, "Owner / reviewer", "Ждёт завершения stage и видит controlled wait вместо ложного failed")
Person(operator, "Platform operator", "Диагностирует blocked contour и следующий шаг")
Person(agent, "System agent", "Должен handoff rate-limit signal без infinite local retries")

System(system, "codex-k8s rate-limit resilience slice", "Contour-aware controlled wait capability для GitHub API")

System_Ext(github, "GitHub API", "Primary and secondary rate-limit signals, repo and issue/PR operations")
System_Ext(k8s, "Kubernetes", "Runtime substrate for agent-runner and worker")
System_Ext(staff, "Staff UI", "Visibility surface for waits, hints and manual actions")

Rel(agent, system, "Hands off raw evidence and resumes only after platform signal", "MCP/internal callbacks")
Rel(owner, system, "Reads wait-state, contour and next-step guidance", "GitHub service-comment + staff UI")
Rel(operator, staff, "Uses wait-state views and diagnostics", "HTTPS")
Rel(system, github, "Calls API and classifies provider signals", "HTTPS")
Rel(system, k8s, "Runs agent and background reconciliation", "Kubernetes API")
Rel(system, staff, "Publishes typed wait projections", "Staff/private API")
```

## Пояснения
- GitHub остаётся внешним source of provider signals, но не source-of-truth для user-facing wait semantics.
- Staff UI и GitHub comments остаются surfaces одного и того же typed projection.
- Kubernetes обеспечивает runtime только для agent/worker execution и не владеет rate-limit semantics.

## Внешние зависимости
- GitHub API: rate-limit headers/signals и affected operations.
- Kubernetes: runtime для `agent-runner` и `worker`.
- Staff UI/API: операторская и owner visibility surface, но не место для бизнес-решений.

## Continuity after `run:plan`
- Plan package Issue `#423` зафиксировал, что этот context overlay остаётся неизменным baseline для execution waves `#425..#431`.
- Ни одна implementation wave не получает права превращать GitHub API, Kubernetes или Staff UI в source-of-truth для wait semantics: этот инвариант остаётся внутри platform domain.
