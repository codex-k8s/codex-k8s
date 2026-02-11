---
doc_id: ARC-C4C-CK8S-0001
type: c4-context
title: "codex-k8s — C4 Context"
status: draft
owner_role: SA
created_at: 2026-02-06
updated_at: 2026-02-11
related_issues: [1]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-06-mvp"
---

# C4 Context: codex-k8s

## TL;DR
- Система в контуре: `codex-k8s` control-plane.
- Пользователи: Owner/Admin, Project Maintainer, системные и custom-агенты.
- Внешние зависимости: GitHub API/Webhooks, Kubernetes API, OpenAI API, HTTP approver/executor интеграции (Telegram как первый адаптер).

## Диаграмма (Mermaid C4Context)
```mermaid
C4Context
title codex-k8s - System Context

Person(owner, "Owner/Admin", "Управляет платформой, правами, проектами")
Person(maintainer, "Project Maintainer", "Работает с проектами и агентными запусками")
Person(agent, "System/Custom Agent", "Выполняет stage задачи и ревизии")
System(system, "codex-k8s", "Webhook-driven control-plane для AI процессов в Kubernetes")

System_Ext(github, "GitHub", "Repo API, OAuth, webhooks")
System_Ext(k8s, "Kubernetes cluster", "Runtime для platform/services/agents")
System_Ext(openai, "OpenAI API", "LLM provider")
System_Ext(approverexec, "HTTP Approver/Executor integrations", "Approval/feedback channel (Telegram/Slack/etc)")

Rel(owner, system, "Uses", "HTTPS UI/API")
Rel(maintainer, system, "Uses", "HTTPS UI/API")
Rel(agent, system, "Executes assigned stages", "HTTPS/gRPC")
Rel(github, system, "Sends webhooks", "HTTPS")
Rel(system, github, "Calls API", "HTTPS")
Rel(system, k8s, "Manages workloads", "Kubernetes API")
Rel(system, openai, "Calls models", "HTTPS")
Rel(system, approverexec, "Requests approvals/feedback", "HTTPS callback")
```

## Пояснения

- Основные взаимодействия: webhook ingest -> domain orchestration -> k8s/repo actions -> audit/state in Postgres.
- Границы ответственности: `codex-k8s` управляет процессами и состоянием, но не заменяет GitHub и Kubernetes как системы-источники соответствующих фактов.
- Stage orchestration в продуктовой модели определяется label taxonomy (`run:*`, `state:*`, `need:*`) и policy апрувов.

## Внешние зависимости

- GitHub: OAuth, repo/webhook operations, fine-grained tokens/service tokens.
- Kubernetes: runtime для сервисов платформы и агентных pod/namespace lifecycle.
- OpenAI: модельные вызовы и токены использования.
- HTTP approver/executor интеграции: канал апрува и уточнений для stage переходов (Telegram/Slack/Mattermost/Jira adapters).

## Решения Owner

- Отдельный provider для enterprise GitHub/GHE на этапе MVP не требуется.
- Production OpenAI account подключается сразу.

## Апрув

- request_id: owner-2026-02-06-mvp
- Решение: approved
- Комментарий: Внешние зависимости на MVP утверждены.
