---
doc_id: ARC-C4N-CK8S-0001
type: c4-container
title: "codex-k8s — C4 Container"
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

# C4 Container: codex-k8s

## TL;DR
- Основные контейнеры: `web-console`, `api-gateway`, `control-plane`, `worker`, `postgres`.
- Технологии: Vue3, Go, PostgreSQL (`JSONB` + `pgvector`).
- Потоки данных: webhook и UI/label запросы -> stage orchestration -> DB sync/audit -> k8s/repo actions -> PR/feedback.

## Диаграмма (Mermaid C4Container)
```mermaid
C4Container
title codex-k8s - Container Diagram

Person(owner, "Owner/Admin", "Управляет платформой")
System_Ext(github, "GitHub", "OAuth, API, webhooks")
System_Ext(k8s, "Kubernetes", "Cluster API")
System_Ext(openai, "OpenAI API", "LLM")
System_Ext(approverexec, "HTTP Approver/Executor integrations", "Approval and feedback (Telegram/Slack/etc)")

System_Boundary(b0, "codex-k8s") {
  Container(web, "Web Console", "Vue3", "UI для настроек, агентов, сессий и запусков")
  Container(gw, "API Gateway", "Go HTTP", "Webhook ingress и публичный API слой")
  Container(cp, "Control Plane", "Go", "Доменная логика и state orchestration")
  Container(worker, "Worker", "Go", "Jobs/reconciliation/rotation/indexing")
  ContainerDb(db, "PostgreSQL", "Postgres + JSONB + pgvector", "Состояние, аудит, doc chunks")
}

Rel(owner, web, "Uses", "HTTPS")
Rel(web, gw, "Calls", "HTTPS")
Rel(github, gw, "Sends webhooks", "HTTPS")
Rel(gw, cp, "Calls", "gRPC")
Rel(cp, db, "Reads/Writes", "SQL")
Rel(worker, db, "Reads/Writes", "SQL")
Rel(cp, github, "Calls API", "HTTPS")
Rel(worker, github, "Calls API", "HTTPS")
Rel(cp, k8s, "Manages resources", "K8s API")
Rel(worker, k8s, "Executes reconciliations", "K8s API")
Rel(cp, openai, "Calls models", "HTTPS")
Rel(cp, approverexec, "Requests approvals and gets callbacks", "HTTPS")
```

## Контейнеры (описание)

### Web Console

* Ответственность: UI для пользователей, проектов, агентов, сессий, документов и журналов.
* Деплой: `services/staff/web-console`.
* Риски: рассинхрон прав в UI при кешировании.

### API Gateway

* Ответственность: webhook validation, auth, routing, edge policies.
* Контракты: OpenAPI для внешнего API.
* Ограничения: без бизнес-логики orchestration и без прямых postgres-репозиториев.

### Control Plane

* Ответственность: доменные use-cases, stage transitions, label policies, prompt template resolution, policy checks.
* Контракты: внутренние service APIs + provider interfaces.
* Ограничения: нет vendor-specific логики в домене.

### Worker

* Ответственность: long-running jobs, retries, reconciliation, rotation, indexing, lifecycle issue/run namespaces.
* Дополнительно: learning-mode post-PR explanations (file/line comments + summary).
* Ограничения: идемпотентность и запись статуса в БД обязательны.

### DB

* Схема/миграции: goose migrations.
* Топология MVP: один PostgreSQL cluster с отдельным логическим контуром для `flow_events`, `agent_sessions`, `token_usage`, `links` и `doc_chunks`.
* Read replica MVP: минимум одна asynchronous streaming replica.
* Эволюция без миграций приложения: переход к 2+ replica и sync/quorum режимам при необходимости.
* Резервирование/бэкап: staging backup baseline обязателен.

## Решения Owner

* Для audit/log/chunks на MVP выделяется отдельный логический БД-контур в рамках PostgreSQL.
* Read replica до production: минимум одна асинхронная streaming replica с архитектурным заделом на масштабирование без изменений приложения.

## Апрув

* request_id: owner-2026-02-06-mvp
* Решение: approved
* Комментарий: Контейнерная архитектура MVP утверждена.
