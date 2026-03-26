---
doc_id: ARC-C4C-S17-0001
type: c4-context
title: "Sprint S17 Day 4 — C4 Context overlay for unified owner feedback loop"
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

# C4 Context: Sprint S17 Day 4 unified owner feedback loop

## TL;DR
- Unified owner feedback loop остаётся platform capability внутри `codex-k8s`, а не Telegram-first contour и не отдельная внешняя система.
- Staff-console fallback входит в platform-owned slice, а Telegram остаётся первым внешним channel path поверх того же persisted truth.

## Диаграмма (Mermaid C4Context)
```mermaid
C4Context
title Sprint S17 Day4 - Unified owner feedback loop context overlay

Person(agent, "System agent", "Вызывает built-in owner feedback wait path и продолжает работу после ответа")
Person(owner, "Owner / Product lead", "Получает pending request и отвечает в Telegram или staff-console")
Person(operator, "Staff / operator", "Отслеживает overdue / expired / manual-fallback состояния")

System(system, "codex-k8s owner feedback continuity slice", "Platform-owned live wait, persisted request truth and continuation capability including staff-console fallback")

System_Ext(tgadapter, "Telegram interaction adapter", "External Telegram delivery, raw webhook handling and reply normalization")
System_Ext(github, "GitHub", "Issue/PR context, service messages and degraded fallback links")
System_Ext(k8s, "Kubernetes", "Runtime for agent pods, control-plane and worker execution")

Rel(agent, system, "Calls built-in owner feedback wait path", "MCP StreamableHTTP")
Rel(owner, tgadapter, "Reads pending request and replies in Telegram", "Telegram UX")
Rel(owner, system, "Uses staff-console fallback", "HTTPS")
Rel(operator, system, "Reviews canonical degraded states", "HTTPS")
Rel(system, tgadapter, "Dispatches deliveries and receives normalized replies", "HTTPS")
Rel(system, github, "Reads issue context and publishes service messages", "HTTPS")
Rel(system, k8s, "Runs live waits and background reconciliation", "Kubernetes API")
```

## Пояснения
- Staff-console fallback materializes как часть platform-owned system, а не как второй внешний канал с собственной семантикой.
- Telegram adapter остаётся external transport contour: raw transport/webhook detail не переопределяет domain meaning request states.
- GitHub остаётся context and fallback channel для ссылок, service messages и degraded-path коммуникации, но не primary accepted-response path.

## Внешние зависимости
- Telegram interaction adapter: первый внешний owner-facing channel для pending inbox и reply normalization.
- GitHub: context, traceability и service messages.
- Kubernetes: runtime substrate для same-session wait, background jobs и lease retention.
