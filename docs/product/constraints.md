---
doc_id: CST-CK8S-0001
type: constraints
title: "codex-k8s — Constraints"
status: active
owner_role: PM
created_at: 2026-02-06
updated_at: 2026-02-15
related_issues: [1, 19]
related_prs: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: "owner-2026-02-12-project-docs-approval"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-12
---

# Constraints: codex-k8s

## TL;DR
Критические ограничения: Kubernetes-only, webhook-driven продуктовые процессы, PostgreSQL (`JSONB` + `pgvector`), GitHub OAuth без self-signup, staging bootstrap по SSH root на Ubuntu 24.04.

## Source of truth
- Канонический список требований и решений Owner: `docs/product/requirements_machine_driven.md`.
- Этот документ фиксирует ограничения и NFR-проекцию для реализации baseline требований.
- Процесс delivery и doc-governance: `docs/delivery/development_process_requirements.md`.

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
- В audit-контуре обязательны сущности `agent_sessions`, `token_usage`, `links` как часть трассировки agent lifecycle.
- Learning mode должен работать как feature toggle на уровне пользователя/проекта и не ломать стандартный pipeline.
- Learning mode default управляется из `bootstrap/host/config.env` (в шаблоне default включён; пустое значение трактуется как выключено).
- Staff API использует short-lived JWT (через API gateway), cookie-session не используется как основной runtime-механизм.
- В первой поставке public API ограничен webhook ingress (`/api/v1/webhooks/github`).
- Отдельный provider для GitHub Enterprise/GHE на MVP не требуется.
- Подключение production OpenAI account допускается сразу.
- Stage-процесс управления задачами фиксирован через label taxonomy `run:*` + `state:*` + `need:*`.
- Для MVP обязательна активация полного stage-каталога (`run:intake..run:ops`, `run:*:revise`, `run:rethink`) и `run:self-improve`.
- Базовый системный штат агентов включает `dev` и `reviewer` как обязательные роли review-контура: для всех `run:*` pre-review обязателен перед финальным Owner review.
- Шаблоны агентных промптов обязаны поддерживать схему: repo seed + DB override (`work` и `review`).
- Шаблоны промптов хранятся по локалям; выбор языка обязателен по цепочке `project locale -> system default locale -> en`.
- Для системных агентов обязательно наличие seed-шаблонов минимум для `ru` и `en`.
- Для external/staff HTTP API обязателен contract-first подход по OpenAPI (spec + runtime validation + codegen backend/frontend).
- В окружениях `ai-staging` и `prod` платформенные Kubernetes ресурсы помечаются label `app.kubernetes.io/part-of=codex-k8s` (канонический критерий для UI/guardrails и backend policy).
- В `ai` окружениях (ai-slots) при dogfooding платформа может разворачиваться без label `app.kubernetes.io/part-of=codex-k8s`, чтобы UI позволял тестировать действия над ресурсами самой платформы (в т.ч. destructive через dry-run) и не применял platform guardrails по label.
- Для будущего admin/cluster контура staff-консоли обязательны guardrails:
  - ресурсы, помеченные `app.kubernetes.io/part-of=codex-k8s`, нельзя удалять (UI и backend policy);
  - `ai-staging` и `prod` — строго view-only для ресурсов с `app.kubernetes.io/part-of=codex-k8s`;
  - ai-slots — destructive действия только dry-run (кнопки есть для dogfooding/debug, реальное действие не выполняется);
  - значения `Secret` по умолчанию не показывать (только метаданные); reveal/редактирование только как отдельное осознанное действие под RBAC и аудитом.
- Интеграции approver/executor должны реализовываться через универсальные HTTP-контракты MCP, без вендорной привязки к конкретному мессенджеру.
- Для MVP обязателен минимальный контур MCP control tools:
  - deterministic secret sync между GitHub и Kubernetes;
  - database create/delete по окружениям;
  - owner feedback handle с вариантами ответа + custom input.

## Операционные ограничения
- SLO/SLA: staging ориентирован на функциональные ручные тесты, не на production SLA.
- Поддержка 24/7: не требуется на этапе MVP.
- Storage профиль MVP: `k3s local-path`, Longhorn откладывается на следующий этап.
- Read replica для MVP: минимум одна асинхронная streaming replica с заделом на переход к 2+ replica и sync/quorum без изменений приложения.
- Режим runner:
  - локальные запуски: 1 persistent runner (long polling);
  - staging/ai-staging/prod при наличии домена: autoscaled runner set.
- Режимы агентного исполнения:
  - `full-env` и `code-only` используются совместно по роли;
  - для `full-env` запусков обязательна изоляция по namespace и cleanup policy.
- В `full-env` агент в границах своего namespace может выполнять runtime-диагностику (логи/метрики/DB/cache/exec в pod), но операции изменения окружения выполняются через MCP-инструменты с approver policy.
- Ограничения по жизненному циклу агента:
  - при ожидании ответа MCP (`wait_state=mcp`) pod/run не может быть завершён по timeout;
  - таймер timeout должен быть paused до завершения MCP ожидания;
  - `codex-cli` session JSON сохраняется для resumable восстановления run после паузы/перезапуска.
- Для self-improve режима:
  - изменения применяются только через PR и owner review;
  - вывод self-improve обязан содержать трассировку: какие логи/комментарии/артефакты привели к конкретному улучшению.
- Ограничения по деплою:
  - staging deploy: автоматический workflow на push в `main`;
  - production deploy: отдельный workflow с ручным запуском и approval gate;
  - bootstrap первой итерации настраивает только staging runner/pipeline.

## Security/Privacy ограничения
- Доступы: GitHub OAuth + внутренняя RBAC матрица по проектам.
- Trigger/deploy labels (`run:*`) при агент-инициации применяются только после апрува Owner.
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
