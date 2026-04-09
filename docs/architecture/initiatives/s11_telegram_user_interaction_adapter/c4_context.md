---
doc_id: ARC-C4C-S11-0001
type: c4-context
title: "Sprint S11 Day 4 — C4 Context overlay for Telegram user interaction adapter"
status: approved
owner_role: SA
created_at: 2026-03-14
updated_at: 2026-03-14
related_issues: [361, 444, 447, 448, 452, 454, 456, 458]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-03-14-issue-452-arch"
  approved_by: "ai-da-stas"
  approved_at: 2026-03-14
---

# C4 Context: Sprint S11 Day 4 Telegram user interaction adapter

## TL;DR
- Telegram-адаптер остаётся первым внешним channel-specific path поверх platform-owned interaction slice, а не новым source-of-truth для interaction semantics.
- Raw Telegram traffic завершается во внешнем Telegram adapter contour; `kodex` получает только normalized callbacks и сохраняет channel-neutral meaning outcome.

## Диаграмма (Mermaid C4Context)
```mermaid
C4Context
title Sprint S11 Day4 - Telegram user interaction adapter context overlay

Person(agent, "System agent", "Вызывает built-in interaction tools через MCP")
Person(user, "End user / requester", "Получает notify и отвечает в Telegram")
Person(owner, "Owner / Product lead", "Получает decision request и operator-visible fallback signals")

System(system, "kodex Telegram interaction slice", "Platform-owned interaction lifecycle with first external Telegram channel")

System_Ext(tgadapter, "Telegram adapter contour", "Channel-specific rendering, Bot API mediation, raw webhook handling")
System_Ext(telegram, "Telegram Bot API", "Bot methods, callback queries and webhook delivery")
System_Ext(github, "GitHub", "Issue/PR context and fallback links")
System_Ext(k8s, "Kubernetes", "Runtime for platform services and background jobs")

Rel(agent, system, "Calls user.notify / user.decision.request", "MCP StreamableHTTP")
Rel(user, tgadapter, "Reads notifications / presses inline buttons / sends free text", "Telegram client UX")
Rel(owner, tgadapter, "Receives actionable requests and fallback updates", "Telegram client UX")
Rel(system, tgadapter, "Dispatches channel-neutral deliveries / receives normalized callbacks", "HTTPS callback contracts")
Rel(tgadapter, telegram, "Uses Bot API and receives webhooks", "HTTPS")
Rel(system, github, "Reads issue/PR context and deep-links", "HTTPS")
Rel(system, k8s, "Runs agent and worker workloads", "Kubernetes API")
```

## Пояснения
- `kodex` владеет interaction aggregate, audit/correlation и semantic classification; Telegram-specific transport detail остаётся во внешнем adapter contour.
- Telegram adapter contour может материализоваться отдельным runtime/service, но для core architecture он рассматривается как replaceable external adapter layer.
- GitHub остаётся fallback/context channel для ссылок и operator workflow, но не primary response path для core S11 flows.

## Внешние зависимости
- Telegram Bot API: first external channel delivery/update surface с webhook/polling constraints и callback query UX expectations.
- Telegram adapter contour: внешний channel-specific слой, который связывает Bot API с platform callback contract.
- GitHub: issue/PR context, deep-links и secondary/manual fallback evidence.
- Kubernetes: runtime substrate для `control-plane`, `worker`, `api-gateway` и agent pods.
