---
doc_id: ARC-C4C-S10-0001
type: c4-context
title: "Sprint S10 Day 4 — C4 Context overlay for built-in MCP user interactions"
status: in-review
owner_role: SA
created_at: 2026-03-12
updated_at: 2026-03-13
related_issues: [360, 378, 383, 385, 387]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-03-12-issue-385-arch"
  approved_by: "ai-da-stas"
  approved_at: 2026-03-13
---

# C4 Context: Sprint S10 Day 4 built-in MCP user interactions

## TL;DR
- Built-in MCP user interactions остаются capability slice внутри `kodex`, а не отдельной внешней системы.
- Human-facing delivery и responses идут через channel-neutral adapter layer; GitHub comments остаются fallback/context channel, а не primary interaction path.

## Диаграмма (Mermaid C4Context)
```mermaid
C4Context
title Sprint S10 Day4 - Built-in MCP user interactions context overlay

Person(agent, "System agent", "Вызывает built-in tools через MCP")
Person(owner, "Owner / Product lead", "Получает actionable notify и отвечает на decision request")
Person(user, "End user / requester", "Даёт typed option или free-text response")

System(system, "kodex interaction slice", "Channel-neutral user interaction capability поверх built-in server kodex")

System_Ext(adapters, "Interaction adapters", "Telegram/Slack/Web/HTTP adapters, future channel-specific UX")
System_Ext(github, "GitHub", "Issue/PR context and fallback links")
System_Ext(k8s, "Kubernetes", "Agent runtime and background execution")

Rel(agent, system, "Calls user.notify / user.decision.request", "MCP StreamableHTTP")
Rel(owner, adapters, "Receives notifications / sends decisions", "Adapter UX")
Rel(user, adapters, "Receives notifications / sends decisions", "Adapter UX")
Rel(system, adapters, "Dispatches delivery / receives callbacks", "HTTPS callback contracts")
Rel(system, github, "Reads issue/PR context and deep-links", "HTTPS")
Rel(system, k8s, "Runs agent/worker workloads", "Kubernetes API")
```

## Пояснения
- Interaction slice не заменяет approval flow и не делает GitHub primary response channel для core MVP.
- Human response path всегда проходит через adapter layer и возвращается в platform domain как typed callback.
- Kubernetes остаётся runtime substrate для agent/worker execution, но не владельцем interaction semantics.

## Внешние зависимости
- Interaction adapters: replaceable channel integrations без vendor lock-in в core contract.
- GitHub: issue/PR context, deep-links и fallback evidence, но не primary state machine user interactions.
- Kubernetes: runtime для agent-runner и worker background loops.
