---
doc_id: CST-CK8S-0001
type: constraints
title: "codex-k8s — Constraints"
status: draft
owner_role: PM
created_at: 2026-02-06
updated_at: 2026-02-06
related_issues: [1]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-06-mvp"
---

# Constraints: codex-k8s

## TL;DR
Критические ограничения: Kubernetes-only, webhook-driven продуктовые процессы, PostgreSQL (`JSONB` + `pgvector`), GitHub OAuth без self-signup, staging bootstrap по SSH root на Ubuntu 24.04.

## Бизнес-ограничения
- Сроки: нужен ранний staging для ручных тестов до полной функциональной готовности.
- Бюджет: инфраструктура MVP на одном сервере/staging-кластере.
- Юр./комплаенс: доступы по email-match и матрице прав, без публичной регистрации.

## Технические ограничения
- Платформы/ОС: целевой сервер bootstrap — Ubuntu 24.04.
- Языки/фреймворки: backend Go, frontend Vue3.
- Инфраструктура: только Kubernetes API (без альтернативных оркестраторов).
- Ограничения по данным: `JSONB` для гибких payload, `pgvector` для chunk search, шифрование repo токенов.
- Размер эмбеддинга для `doc_chunks.embedding`: `vector(3072)`.
- Event outbox table на MVP не вводим; достаточно `agent_runs` + `flow_events`.
- Learning mode должен работать как feature toggle на уровне пользователя/проекта и не ломать стандартный pipeline.
- Learning mode default управляется из `bootstrap/host/config.env` (в шаблоне default включён; пустое значение трактуется как выключено).
- Staff API использует short-lived JWT (через API gateway), cookie-session не используется как основной runtime-механизм.
- В первой поставке public API ограничен webhook ingress (`/api/v1/webhooks/github`).
- Отдельный provider для GitHub Enterprise/GHE на MVP не требуется.
- Подключение production OpenAI account допускается сразу.

## Операционные ограничения
- SLO/SLA: staging ориентирован на функциональные ручные тесты, не на production SLA.
- Поддержка 24/7: не требуется на этапе MVP.
- Storage профиль MVP: `k3s local-path`, Longhorn откладывается на следующий этап.
- Read replica для MVP: минимум одна асинхронная streaming replica с заделом на переход к 2+ replica и sync/quorum без изменений приложения.
- Режим runner:
  - локальные запуски: 1 persistent runner (long polling);
  - staging/ai-staging/prod при наличии домена: autoscaled runner set.
- Ограничения по деплою:
  - staging deploy: автоматический workflow на push в `main`;
  - production deploy: отдельный workflow с ручным запуском и approval gate;
  - bootstrap первой итерации настраивает только staging runner/pipeline.

## Security/Privacy ограничения
- Доступы: GitHub OAuth + внутренняя RBAC матрица по проектам.
- Секреты: платформенные из env; внутренние генерируются bootstrap-скриптом.
- PII/персональные данные: минимум (email и аудит), без утечки в логи.
- Обучающие комментарии не должны раскрывать секреты, внутренние токены и чувствительные данные.

## Неизменяемые решения (если уже есть)
- ADR-0001: Kubernetes-only orchestration.
- ADR-0002: webhook-driven execution + отдельные deploy workflows платформы.
- ADR-0003: PostgreSQL (`JSONB` + `pgvector`) как state and sync backend.
- ADR-0004: repository provider interface.

## Апрув
- request_id: owner-2026-02-06-mvp
- Решение: approved
- Комментарий: Ограничения MVP зафиксированы Owner.
