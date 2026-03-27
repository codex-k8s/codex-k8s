---
doc_id: ARC-C4C-S18-0001
type: c4-context
title: "Sprint S18 Day 4 — C4 Context overlay for frontend-first Mission Control canvas"
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

# C4 Context: Sprint S18 Day 4 frontend-first Mission Control canvas

## TL;DR
- Sprint S18 моделирует Mission Control как isolated fake-data UX slice внутри `codex-k8s`, а не как live GitHub/control-plane read model.
- Пользователи взаимодействуют с owner-approved canvas UX, тогда как GitHub и backend rebuild `#563` остаются внешними или downstream dependencies, но не текущим runtime prerequisite.
- Диаграмма подчёркивает separation: текущий prototype доказывает UX и safe action semantics, а persisted provider/data truth остаётся отдельной будущей задачей.

## Диаграмма (Mermaid C4Context)
```mermaid
C4Context
title Sprint S18 Day4 - Frontend-first Mission Control canvas context overlay

Person(owner, "Owner / product lead", "Смотрит 2-3 инициативы и выбирает следующий безопасный шаг")
Person(operator, "Execution lead / operator", "Переключает инициативы, читает связи и детали")
Person(team, "Future design/dev team", "Берёт approved UX baseline в `run:design -> run:dev`")

System(system, "Sprint S18 Mission Control prototype slice", "Isolated fake-data canvas UX in `web-console` with explicit backend handover seam")

System_Ext(github, "GitHub", "Provider system; в Sprint S18 доступны только safe deep links, без live mutation path")
System_Ext(k8s, "Kubernetes", "Runtime substrate платформы; не является source of truth для prototype data")
System_Ext(backend, "Mission Control backend rebuild #563", "Будущий persisted provider/data foundation и workflow policy implementation")

Rel(owner, system, "Uses", "HTTPS UI")
Rel(operator, system, "Uses", "HTTPS UI")
Rel(team, system, "Consumes reviewed package and handover rules", "Docs / PR")
Rel(system, github, "Opens provider-safe links only", "HTTPS")
Rel(system, k8s, "Runs inside existing platform runtime", "Kubernetes")
Rel(system, backend, "Hands over approved UX baseline and boundaries", "Documentation contract")
```

## Пояснения
- GitHub не выступает live data source для prototype runtime path: Sprint S18 доказывает UX, а не backend truth.
- Kubernetes обеспечивает окружение платформы, но не хранит fake-data scenario state как доменную истину.
- Backend rebuild `#563` показан как downstream system, потому что именно он позже должен принять persisted provider/data ownership без reopening UX baseline.

## Внешние зависимости
- GitHub: только provider-safe deep links и исторический продуктовый контекст.
- Kubernetes: runtime substrate для текущей платформы и будущих execution flows.
- Backend rebuild `#563`: будущий consumer approved UX baseline и отдельный owner data/model work.

## Continuity after `run:arch`
- Issue `#573` (`run:design`) должен сохранить этот context overlay как baseline для UI/state/contract design package.
- Live provider sync, DB prompt editor и release-safety cockpit остаются вне core context Sprint S18.
