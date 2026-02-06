---
doc_id: RDM-CK8S-0001
type: roadmap
title: "codex-k8s — Roadmap"
status: draft
owner_role: PM
created_at: 2026-02-06
updated_at: 2026-02-06
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-06-mvp"
---

# Roadmap: codex-k8s

## TL;DR
- Q1: foundation + core backend + bootstrap staging.
- Q2: hardened staging + richer UI + learning mode + docs/agent/session observability.
- Q3: production readiness + GitLab provider onboarding.
- Q4: extensibility for custom agents/process templates.

## Принципы приоритизации
- Сначала контроль рисков и deployability.
- Затем продуктовые возможности и масштабирование.
- Избегать расширений, которые ломают webhook-driven core.

## Roadmap (high-level)
| Период | Инициатива | Цель | Метрики | Статус |
|---|---|---|---|---|
| Q1 | MVP core + staging bootstrap | запустить рабочий staging и ручные тесты | one-command bootstrap, green deploy from main | planned |
| Q2 | UX и learning mode | контроль проектов/агентов/слотов в UI + обучающие explain-потоки | >=80% действий через UI и >=70% learner-runs с полезным feedback | planned |
| Q3 | Production readiness | безопасный production deploy | prod runbook + approval gates + rollback drills | planned |
| Q4 | Extensibility | задел на пользовательских агентов и процессы | configurable agent templates (phase-1) | planned |

## Backlog кандидатов
- Split control-plane по внутренним сервисам при росте нагрузки.
- Vault/KMS интеграция вместо хранения repo token material в БД.
- Расширенная политика workflow approvals.

## Риски roadmap
- Задержка из-за инфраструктурной автоматизации bootstrap.
- Недостаточная зрелость security baseline при раннем масштабировании.

## Апрув
- request_id: owner-2026-02-06-mvp
- Решение: approved
- Комментарий: Дорожная карта этапов MVP утверждена Owner.
