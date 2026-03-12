---
doc_id: ARC-C4C-S9-0001
type: c4-context
title: "Sprint S9 Day 4 — C4 Context overlay for Mission Control Dashboard"
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

# C4 Context: Sprint S9 Day 4 Mission Control Dashboard

## TL;DR
- Mission Control Dashboard остаётся product slice внутри `codex-k8s`, а не отдельной внешней системой.
- GitHub остаётся provider source-of-truth для review/collaboration, Kubernetes остаётся источником runtime состояния, voice intake — только optional candidate stream.

## Диаграмма (Mermaid C4Context)
```mermaid
C4Context
title Sprint S9 Day4 - Mission Control Dashboard context overlay

Person(owner, "Owner / Product lead", "Получает situational awareness по active set и выбирает следующий шаг")
Person(operator, "Engineer / Operator", "Управляет work items, PR, агентами и sync state")
Person(discussion, "Discussion-first user", "Начинает с discussion и формализует её в task")

System(system, "codex-k8s Mission Control Dashboard slice", "Active-set control plane для work items, discussion, PR и agents")

System_Ext(github, "GitHub", "Issues, PR, comments, reviews, webhooks")
System_Ext(k8s, "Kubernetes", "Agent/runtime state")
System_Ext(voice, "Optional voice intake provider", "Candidate transcript source, Wave 3 only")

Rel(owner, system, "Uses", "HTTPS UI")
Rel(operator, system, "Uses", "HTTPS UI")
Rel(discussion, system, "Uses", "HTTPS UI")
Rel(system, github, "Reads provider state / sends provider-safe commands", "HTTPS")
Rel(github, system, "Sends webhooks / provider echoes", "HTTPS")
Rel(system, k8s, "Reads runtime and agent state", "Kubernetes API")
Rel(voice, system, "Provides optional candidate input", "HTTPS")
```

## Пояснения
- Mission Control Dashboard не заменяет GitHub как место финального human review и merge decision.
- Runtime/agent состояние приходит из текущих внутренних контуров платформы и агрегируется в active-set projection.
- Optional voice provider не становится обязательной зависимостью для core MVP.
