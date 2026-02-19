---
doc_id: ADR-0002
type: adr
title: "Webhook-driven execution with Kubernetes deploy jobs"
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
  request_id: "owner-2026-02-19-full-docset"
  approved_by: "ai-da-stas"
  approved_at: 2026-02-19
---

# ADR-0002: Webhook-driven execution with Kubernetes deploy jobs

## TL;DR
- Контекст: продуктовые процессы должны запускаться по webhooks, без workflow-first модели.
- Решение: orchestration доменных/агентных процессов только webhook-driven; deploy самой платформы выполняется через Kubernetes jobs под управлением control-plane.
- Последствия: сохраняем требование по продукту и исключаем зависимость от GitHub Actions workflows.

## Контекст
- Проблема/драйвер: нужен быстрый и управляемый deploy `codex-k8s` в production/prod.
- Ограничения: одновременно есть требование “никаких воркфлоу” для продуктовых процессов.
- Что “ломается” без решения: либо медленный ручной deploy, либо нарушение архитектурного принципа webhook-first.

## Decision Drivers (что важно)
- Скорость вывода в production.
- Чёткое разделение платформенного CI/CD и продуктовой оркестрации.
- Аудит и воспроизводимость.

## Рассмотренные варианты
### Вариант A: полностью без GitHub Actions
- Плюсы: максимально строгий webhook-only подход.
- Минусы: дорогой запуск MVP, больше ручных операций.

### Вариант B: webhook-driven продукт + отдельные deploy workflows платформы
- Плюсы: быстрый production deploy, прозрачный путь push->deploy.
- Минусы: сохраняется часть workflow инфраструктуры и ARC.

### Вариант C: workflow-first для всего
- Плюсы: простая унификация.
- Минусы: противоречит базовому продукт-требованию.

## Решение
Мы выбираем: **Вариант A (workflow-free deploy через control-plane + Kubernetes jobs)**.

## Обоснование (Rationale)
Платформенный CI/CD и продуктовая оркестрация должны использовать единый Kubernetes-native контур.
Это убирает ARC/GitHub Actions из критического пути и делает self-deploy воспроизводимым внутри платформы.

## Последствия (Consequences)
### Позитивные
- production можно поднимать и обновлять автоматически после push в `main` через внутренние job.
- продуктовые run-процессы остаются webhook-driven внутри `codex-k8s`.

### Негативные / компромиссы
- Нужно поддерживать внутренние build/deploy job и их наблюдаемость (логи/статусы).

### Технический долг
- В будущем можно выделить отдельный internal deploy-controller, если нагрузка на control-plane вырастет.

## План внедрения (минимально)
- Добавить манифесты Kubernetes job для build/deploy/codegen check.
- Вынести управление GitHub secrets/vars/webhooks и Kubernetes ресурсами в Go-код (`codex-bootstrap` + control-plane), без shell-first orchestration.
- Bootstrap-скрипт оставить только для первичной подготовки хоста.

## План отката/замены
- Условия отката: нестабильность внутренних build/deploy job.
- Как откатываем: аварийная команда `codex-bootstrap` с принудительным redeploy и (опционально) полной очисткой.

## Ссылки
- Brief: `docs/product/brief.md`
- Delivery Plan: `docs/delivery/delivery_plan.md`

## Апрув
- request_id: N/A
- Решение: approved
- Комментарий:
