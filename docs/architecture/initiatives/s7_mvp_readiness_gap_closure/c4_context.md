---
doc_id: ARC-C4C-S7-0001
type: c4-context
title: "Sprint S7 Day 4 — C4 Context overlay for MVP readiness streams"
status: in-review
owner_role: SA
created_at: 2026-03-02
updated_at: 2026-03-02
related_issues: [220, 222, 238]
related_prs: []
approvals:
  required: ["Owner"]
  status: pending
  request_id: "owner-2026-03-02-issue-222-arch"
---

# C4 Context: Sprint S7 Day 4 MVP readiness streams

## TL;DR
- Система в контуре: `kodex` stage/governance runtime.
- Фокус overlay: потоки `S7-E01..S7-E18`, которые закрывают MVP readiness gaps.
- Ключевые внешние системы: GitHub, Kubernetes, OpenAI.

## Диаграмма (Mermaid C4Context)
```mermaid
C4Context
title Sprint S7 Day4 - MVP readiness architecture context

Person(owner, "Owner", "Финальный апрув stage/transition и go-no-go")
Person(pm_em, "PM/EM", "Управление backlog, parity-gate и continuity")
Person(dev_qa_sre_km, "Dev/QA/SRE/KM", "Исполнение, QA evidence, ops и документация")

System(system, "kodex", "Webhook-driven orchestration и stage process")

System_Ext(github, "GitHub", "Issues/PR/labels/reviews")
System_Ext(k8s, "Kubernetes", "Runtime execution environment")
System_Ext(openai, "OpenAI", "LLM execution provider")

Rel(owner, system, "Утверждает переходы и итоговые артефакты", "UI/Issue/PR")
Rel(pm_em, system, "Фиксирует parity-gate и traceability", "Docs/Labels")
Rel(dev_qa_sre_km, system, "Исполняет streams S7-E01..S7-E18", "run:* stages")
Rel(system, github, "Читает/обновляет issue/pr context", "API/Webhooks")
Rel(system, k8s, "Управляет runtime jobs/namespaces", "Kubernetes API")
Rel(system, openai, "Модельные вызовы для agent runs", "HTTPS")
```

## Пояснения
- S7 Day4 не вводит новых внешних интеграций; фокус только на декомпозиции границ и ownership.
- Потоки, требующие runtime-изменений, остаются в `internal/jobs`; edge и UI должны оставаться thin adapters.
- Governance-потоки (`S7-E01`, `S7-E12`, `S7-E18`) закрепляют обязательные quality-gates перед `run:dev`.

## Открытые вопросы для run:design
- Нужен ли отдельный transport endpoint для `runtime deploy cancel/stop` (`S7-E10`) или достаточно расширить существующий typed action contract?
- Какие persisted состояния будут изменяться для потоков reliability (`S7-E16`, `S7-E17`) и требуют ли отдельной миграции?
- Как формально валидировать parity-gate в automation без нарушения текущего stage process model?
