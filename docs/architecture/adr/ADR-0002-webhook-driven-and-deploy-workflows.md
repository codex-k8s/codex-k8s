---
doc_id: ADR-0002
type: adr
title: "Webhook-driven execution with dedicated deploy workflows"
status: accepted
owner_role: SA
created_at: 2026-02-06
updated_at: 2026-02-06
related_issues: [1]
related_prs: []
supersedes: []
superseded_by: []
approvals:
  required: ["Owner"]
  status: approved
  request_id: ""
---

# ADR-0002: Webhook-driven execution with dedicated deploy workflows

## TL;DR
- Контекст: продуктовые процессы должны запускаться по webhooks, без workflow-first модели.
- Решение: orchestration доменных/агентных процессов только webhook-driven; для deploy самой платформы допускаем отдельные GitHub Actions workflows (staging first).
- Последствия: сохраняем требование по продукту и получаем практичный CI/CD путь для платформы.

## Контекст
- Проблема/драйвер: нужен быстрый и управляемый deploy `codex-k8s` в staging/prod.
- Ограничения: одновременно есть требование “никаких воркфлоу” для продуктовых процессов.
- Что “ломается” без решения: либо медленный ручной deploy, либо нарушение архитектурного принципа webhook-first.

## Decision Drivers (что важно)
- Скорость вывода в staging.
- Чёткое разделение платформенного CI/CD и продуктовой оркестрации.
- Аудит и воспроизводимость.

## Рассмотренные варианты
### Вариант A: полностью без GitHub Actions
- Плюсы: максимально строгий webhook-only подход.
- Минусы: дорогой запуск MVP, больше ручных операций.

### Вариант B: webhook-driven продукт + отдельные deploy workflows платформы
- Плюсы: быстрый staging deploy, прозрачный путь push->deploy.
- Минусы: сохраняется часть workflow инфраструктуры.

### Вариант C: workflow-first для всего
- Плюсы: простая унификация.
- Минусы: противоречит базовому продукт-требованию.

## Решение
Мы выбираем: **Вариант B**.

## Обоснование (Rationale)
Платформенный CI/CD и продуктовая оркестрация решают разные задачи; их можно разделить без противоречия архитектуре.

## Последствия (Consequences)
### Позитивные
- staging можно поднимать и обновлять автоматически после push в `main`.
- продуктовые run-процессы остаются webhook-driven внутри `codex-k8s`.

### Негативные / компромиссы
- Нужно поддерживать self-hosted runner в Kubernetes.

### Технический долг
- В будущем можно заменить deploy workflows на встроенный deploy controller, если это даст выгоду.

## План внедрения (минимально)
- Добавить `ai_staging_deploy` workflow для `codex-k8s`.
- Production deploy workflow оставить как следующий этап после стабилизации staging.
- Bootstrap-скрипт должен:
  - запросить GitHub fine-grained token;
  - создать/настроить runner secret;
  - развернуть ARC/runner scale set в k8s;
  - подготовить переменные/секреты репозитория.

## План отката/замены
- Условия отката: нестабильность runner/Actions или переход на встроенный deploy controller.
- Как откатываем: manual deploy script + отключение workflows.

## Ссылки
- Brief: `docs/product/brief.md`
- Delivery Plan: `docs/delivery/delivery_plan.md`

## Апрув
- request_id: N/A
- Решение: approved
- Комментарий:
